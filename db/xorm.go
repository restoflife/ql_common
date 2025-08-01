package db

import (
	"context"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/restoflife/ql_common/logger"
	"go.uber.org/zap"
	"xorm.io/xorm"
)

// 存储所有的数据库引擎组（主从）
var dbMgr = map[string]*xorm.EngineGroup{}

// MustBootUpXORM 初始化并启动 XORM 引擎（可支持多个数据库配置）
func MustBootUpXORM(configs map[string]*XORMConfigLite, sqlLog *zap.Logger, opts ...Option) error {
	options := newOptions(opts...)

	for name, c := range configs {
		// 创建主库连接
		master, err := xorm.NewEngine(c.Driver, c.Dsn)
		if err != nil {
			return err
		}

		// 创建从库连接
		slaves := make([]*xorm.Engine, len(c.Slave))
		for i, s := range c.Slave {
			slave, x := xorm.NewEngine(c.Driver, s.Dsn)
			if x != nil {
				return x
			}
			slaves[i] = slave
		}

		// 创建主从引擎组
		db, err := xorm.NewEngineGroup(master, slaves)
		if err != nil {
			return err
		}

		// 设置 SQL 日志
		db.SetLogger(logger.NewXormLogger(sqlLog))
		db.ShowSQL(c.ShowSql)

		// 设置连接池参数
		if c.MaxIdle > 0 {
			db.SetMaxIdleConns(c.MaxIdle)
		}
		if c.MaxOpen > 0 {
			db.SetMaxOpenConns(c.MaxOpen)
		}
		if c.MaxLife > 0 {
			db.SetConnMaxLifetime(time.Millisecond * time.Duration(c.MaxLife))
		}

		// 测试连接
		if err = db.Ping(); err != nil {
			return err
		}

		// 防止重复加载相同名字的数据库连接
		if _, ok := dbMgr[name]; ok {
			return fmt.Errorf("database components loaded twice：[%s]", name)
		}

		// 同步数据库结构（如果设置了同步）
		if options.sync != nil && c.Synchronization {
			if err = options.sync(name, db); err != nil {
				return err
			}
		}

		// 保存引擎组
		dbMgr[name] = db
		sqlLog.Info("XORM连接成功", zap.String("name", name))
	}

	// 定时健康检查（每 5 小时 ping 一次）
	go func() {
		ticker := time.NewTicker(time.Hour * 5)
		for {
			select {
			case <-ticker.C:
				for _, v := range dbMgr {
					if err := v.Ping(); err != nil {
						sqlLog.Error("mysql ticker ping database fail", zap.Error(err))
						return
					}
				}
			}
		}
	}()

	return nil
}

// Transaction 封装事务操作逻辑
func Transaction(ctx context.Context, name string, fn func(*xorm.Session) error) (err error) {
	session, err := NewSessionContext(ctx, name)
	if err != nil {
		return err
	}
	defer Close(session)

	if err = session.Begin(); err != nil {
		return err
	}

	if err = fn(session); err != nil {
		_ = session.Rollback()
		return err
	}

	return session.Commit()
}

// NewSessionContext 获取一个绑定 context 的数据库会话（需手动释放）
func NewSessionContext(ctx context.Context, name string) (*xorm.Session, error) {
	if g, e := get(name); e == nil {
		return g.NewSession().Context(ctx), nil
	} else {
		return nil, e
	}
}

// NewSession 获取一个数据库会话（需手动释放）
func NewSession(name string) (*xorm.Session, error) {
	if g, e := get(name); e == nil {
		return g.NewSession(), nil
	} else {
		return nil, e
	}
}

// 获取对应数据库名称的引擎组
func get(name string) (*xorm.EngineGroup, error) {
	g, ok := dbMgr[name]
	if !ok {
		return nil, fmt.Errorf("database does not exist:[%s]", name)
	}
	return g, nil
}

// Close 关闭 XORM 会话
func Close(session *xorm.Session) {
	if err := session.Close(); err != nil {
		// 可以添加日志记录
		return
	}
}

// ShutdownXorm 应用退出时关闭所有数据库连接
func ShutdownXorm() {
	for _, v := range dbMgr {
		if err := v.Close(); err != nil {
			// 可以添加日志记录
			continue
		}
	}
}

// 同步函数类型（可用于同步数据库结构）
type syncFunc func(string, *xorm.EngineGroup) error

// Options 用于配置 BootUp 的可选参数
type Options struct {
	sync syncFunc
}

// Option 是对 Options 的函数式配置
type Option func(*Options)

// SetSyncFunc 设置同步函数
func SetSyncFunc(f syncFunc) Option {
	return func(o *Options) {
		o.sync = f
	}
}

// 解析所有 Option
func newOptions(opts ...Option) Options {
	opt := Options{
		sync: nil,
	}
	for _, o := range opts {
		o(&opt)
	}
	return opt
}
