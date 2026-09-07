package db

import (
	"context"
	"errors"
	"math"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/restoflife/ql_common/internal/registry"
	"github.com/restoflife/ql_common/logger"
	"go.uber.org/zap"
	"xorm.io/xorm"
)

var dbMgr registry.Registry[*xorm.EngineGroup]

// MustBootUpXORM is the compatibility entry point. MaxLife is measured in seconds.
// Deprecated: Use BootUpXORMContext.
func MustBootUpXORM(configs map[string]*XORMConfigLite, sqlLog *zap.Logger, opts ...Option) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return BootUpXORMContext(ctx, configs, sqlLog, opts...)
}

// BootUpXORMContext initializes a batch atomically. Schema synchronization side
// effects cannot be rolled back; the sync callback must not call boot/shutdown.
func BootUpXORMContext(ctx context.Context, configs map[string]*XORMConfigLite, sqlLog *zap.Logger, opts ...Option) error {
	if ctx == nil {
		return ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if sqlLog == nil {
		sqlLog = zap.NewNop()
	}
	opt := newOptions(opts...)
	names := make([]string, 0, len(configs))
	for name, c := range configs {
		if c == nil || c.Driver == "" || c.Dsn == "" {
			return databaseConfigError(name, "driver and DSN are required")
		}
		if c.MaxIdle < 0 || c.MaxOpen < 0 || c.MaxLife < 0 || int64(c.MaxLife) > math.MaxInt64/int64(time.Second) {
			return databaseConfigError(name, "invalid pool settings")
		}
		if c.MaxOpen > 0 && c.MaxIdle > c.MaxOpen {
			return databaseConfigError(name, "max_idle exceeds max_open")
		}
		for _, slave := range c.Slave {
			if slave.Dsn == "" {
				return databaseConfigError(name, "slave DSN is empty")
			}
		}
		if len(c.SlavePools) != 0 && len(c.SlavePools) != len(c.Slave) {
			return databaseConfigError(name, "slave pool count mismatch")
		}
		pools := append([]PoolConfig(nil), c.SlavePools...)
		if c.MasterPool != nil {
			pools = append(pools, *c.MasterPool)
		}
		for _, pool := range pools {
			if pool.MaxIdle < 0 || pool.MaxOpen < 0 || pool.Lifetime < 0 || (pool.MaxOpen > 0 && pool.MaxIdle > pool.MaxOpen) {
				return databaseConfigError(name, "invalid engine pool")
			}
		}
		names = append(names, name)
	}
	return dbMgr.Init(names, func(name string) (_ *xorm.EngineGroup, err error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		c := configs[name]
		engines := make([]*xorm.Engine, 0, 1+len(c.Slave))
		committed := false
		defer func() {
			if !committed {
				for _, engine := range engines {
					err = errors.Join(err, engine.Close())
				}
			}
		}()
		master, err := xorm.NewEngine(c.Driver, c.Dsn)
		if err != nil {
			return nil, err
		}
		engines = append(engines, master)
		for _, slave := range c.Slave {
			engine, err := xorm.NewEngine(c.Driver, slave.Dsn)
			if err != nil {
				return nil, err
			}
			engines = append(engines, engine)
		}
		group, err := xorm.NewEngineGroup(master, engines[1:])
		if err != nil {
			return nil, err
		}
		group.SetLogger(logger.NewXormLogger(sqlLog))
		group.ShowSQL(c.ShowSql)
		if c.MaxIdle > 0 {
			group.SetMaxIdleConns(c.MaxIdle)
		}
		if c.MaxOpen > 0 {
			group.SetMaxOpenConns(c.MaxOpen)
		}
		if c.MaxLife > 0 {
			group.SetConnMaxLifetime(time.Second * time.Duration(c.MaxLife))
		}
		apply := func(engine *xorm.Engine, pool PoolConfig) {
			engine.SetMaxIdleConns(pool.MaxIdle)
			engine.SetMaxOpenConns(pool.MaxOpen)
			engine.SetConnMaxLifetime(pool.Lifetime)
		}
		if c.MasterPool != nil {
			apply(master, *c.MasterPool)
		}
		for i, pool := range c.SlavePools {
			apply(engines[i+1], pool)
		}
		for _, engine := range engines {
			if err = engine.PingContext(ctx); err != nil {
				return nil, err
			}
		}
		sqlLog.WithOptions(zap.WithCaller(false)).Info("MySQL ping database succeeded", zap.String("name", name))
		if opt.sync != nil && c.Synchronization {
			if err = opt.sync(name, group); err != nil {
				return nil, err
			}
		}
		if err = ctx.Err(); err != nil {
			return nil, err
		}
		committed = true
		return group, nil
	}, closeEngineGroup)
}

// Transaction provides the corresponding package operation.
func Transaction(ctx context.Context, name string, fn func(*xorm.Session) error) (err error) {
	if fn == nil {
		return ErrNilTransactionCallback
	}
	session, err := NewSessionContext(ctx, name)
	if err != nil {
		return err
	}
	defer Close(session)

	if err = session.Begin(); err != nil {
		return err
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			_ = session.Rollback()
			panic(recovered)
		}
	}()
	if err = fn(session); err != nil {
		return errors.Join(err, session.Rollback())
	}

	return session.Commit()
}

// NewSessionContext provides the corresponding package operation.
func NewSessionContext(ctx context.Context, name string) (*xorm.Session, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if g, e := get(name); e == nil {
		return g.NewSession().Context(ctx), nil
	} else {
		return nil, e
	}
}

// NewSession provides the corresponding package operation.
// Deprecated: Use NewSessionContext.
func NewSession(name string) (*xorm.Session, error) {
	if g, e := get(name); e == nil {
		return g.NewSession(), nil
	} else {
		return nil, e
	}
}

// GetEngineGroup returns a shared group; requests must not close it.
func GetEngineGroup(name string) (*xorm.EngineGroup, error) { return dbMgr.Get(name) }

func get(name string) (*xorm.EngineGroup, error) { return dbMgr.Get(name) }

// Close provides the corresponding package operation.
func Close(session *xorm.Session) {
	if session == nil {
		return
	}
	if err := session.Close(); err != nil {
		return
	}
}

// ShutdownXorm provides the corresponding package operation.
// Deprecated: Use ShutdownXormE.
func ShutdownXorm() {
	if err := ShutdownXormE(); err != nil {
		logger.Error("XORM shutdown failed", zap.Error(err))
	}
}

// ShutdownXormE closes and unregisters every engine; call after draining requests.
func ShutdownXormE() error { return dbMgr.Close(closeEngineGroup) }

func closeEngineGroup(g *xorm.EngineGroup) error {
	result := g.Master().Close()
	for _, slave := range g.Slaves() {
		result = errors.Join(result, slave.Close())
	}
	return result
}

type syncFunc func(string, *xorm.EngineGroup) error

// Options provides the corresponding package operation.
type Options struct {
	sync syncFunc
}

// Option provides the corresponding package operation.
type Option func(*Options)

// SetSyncFunc provides the corresponding package operation.
func SetSyncFunc(f syncFunc) Option {
	return func(o *Options) {
		o.sync = f
	}
}

func newOptions(opts ...Option) Options {
	opt := Options{
		sync: nil,
	}
	for _, o := range opts {
		if o != nil {
			o(&opt)
		}
	}
	return opt
}
