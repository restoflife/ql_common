package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"sync/atomic"
	"testing"

	"xorm.io/xorm"
	"xorm.io/xorm/core"
)

type testConnector struct{ state *txState }
type txState struct {
	commits, rollbacks     atomic.Int32
	commitErr, rollbackErr error
}

func (c testConnector) Connect(context.Context) (driver.Conn, error) {
	return &testConn{state: c.state}, nil
}
func (c testConnector) Driver() driver.Driver { return testDriver{c.state} }

type testDriver struct{ state *txState }

func (d testDriver) Open(string) (driver.Conn, error) { return &testConn{state: d.state}, nil }

type testConn struct{ state *txState }

func (c *testConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("not implemented") }
func (c *testConn) Close() error                        { return nil }
func (c *testConn) Begin() (driver.Tx, error)           { return &testTx{c.state}, nil }

type testTx struct{ state *txState }

func (t *testTx) Commit() error   { t.state.commits.Add(1); return t.state.commitErr }
func (t *testTx) Rollback() error { t.state.rollbacks.Add(1); return t.state.rollbackErr }

func setupEngine(t *testing.T, state *txState) {
	t.Helper()
	_ = ShutdownXormE()
	sqlDB := sql.OpenDB(testConnector{state})
	engine, err := xorm.NewEngineWithDB("mysql", "root@tcp(localhost:3306)/test", core.FromDB(sqlDB))
	if err != nil {
		_ = sqlDB.Close()
		t.Fatal(err)
	}
	group, err := xorm.NewEngineGroup(engine, []*xorm.Engine{})
	if err != nil {
		_ = engine.Close()
		t.Fatal(err)
	}
	if err := dbMgr.Init([]string{"test"}, func(string) (*xorm.EngineGroup, error) { return group, nil }, func(g *xorm.EngineGroup) error { return g.Close() }); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := ShutdownXormE(); err != nil {
			t.Error(err)
		}
	})
}

func TestTransactionCommitRollbackAndPanic(t *testing.T) {
	for _, kind := range []string{"commit", "error", "panic", "commit error", "rollback error"} {
		t.Run(kind, func(t *testing.T) {
			state := &txState{}
			failure := errors.New("callback failure")
			if kind == "commit error" {
				state.commitErr = errors.New("commit failure")
			}
			if kind == "rollback error" {
				state.rollbackErr = errors.New("rollback failure")
			}
			setupEngine(t, state)
			var err error
			recovered := false
			func() {
				defer func() {
					if p := recover(); p != nil {
						if p != "boom" {
							t.Errorf("unexpected panic %v", p)
						}
						recovered = true
					}
				}()
				err = Transaction(context.Background(), "test", func(*xorm.Session) error {
					if kind == "panic" {
						panic("boom")
					}
					if kind == "error" || kind == "rollback error" {
						return failure
					}
					return nil
				})
			}()
			switch kind {
			case "commit":
				if err != nil || state.commits.Load() != 1 || state.rollbacks.Load() != 0 {
					t.Fatalf("err=%v state=%+v", err, state)
				}
			case "error", "rollback error":
				if !errors.Is(err, failure) || state.rollbacks.Load() != 1 {
					t.Fatalf("err=%v", err)
				}
				if state.rollbackErr != nil && !errors.Is(err, state.rollbackErr) {
					t.Fatalf("rollback error lost: %v", err)
				}
			case "panic":
				if !recovered || state.rollbacks.Load() != 1 {
					t.Fatal("panic did not roll back")
				}
			case "commit error":
				if !errors.Is(err, state.commitErr) {
					t.Fatalf("commit error lost: %v", err)
				}
			}
		})
	}
}

func TestValidationAndRestart(t *testing.T) {
	if err := MustBootUpXORM(map[string]*XORMConfigLite{"bad": nil}, nil); err == nil {
		t.Fatal("nil config accepted")
	}
	if err := BootUpXORMContext(nil, nil, nil); err == nil {
		t.Fatal("nil context accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := NewSessionContext(ctx, "missing"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := Transaction(context.Background(), "missing", nil); err == nil {
		t.Fatal("nil callback accepted")
	}
	if _, err := NewSession("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	Close(nil)
	setupEngine(t, &txState{})
	if err := ShutdownXormE(); err != nil {
		t.Fatal(err)
	}
	if _, err := NewSession("test"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	setupEngine(t, &txState{})
}

func TestPublicEngineGroupUsesRegistry(t *testing.T) {
	setupEngine(t, &txState{})
	group, err := GetEngineGroup("test")
	if err != nil {
		t.Fatal(err)
	}
	registered, err := get("test")
	if err != nil || registered != group {
		t.Fatal("separate engine returned", err)
	}
	if err := ShutdownXormE(); err != nil {
		t.Fatal(err)
	}
	if _, err := GetEngineGroup("test"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
}

func TestInvalidPoolConfigurations(t *testing.T) {
	cases := map[string]*XORMConfigLite{
		"missing driver":    {Dsn: "dsn"},
		"missing DSN":       {Driver: "mysql"},
		"negative idle":     {Driver: "mysql", Dsn: "dsn", MaxIdle: -1},
		"idle exceeds open": {Driver: "mysql", Dsn: "dsn", MaxIdle: 2, MaxOpen: 1},
		"slave pool mismatch": {Driver: "mysql", Dsn: "dsn", Slave: []struct {
			Dsn string `toml:"dsn" yaml:"dsn" json:"dsn"`
		}{{Dsn: "slave"}}, SlavePools: []PoolConfig{{}, {}}},
	}
	for name, cfg := range cases {
		t.Run(name, func(t *testing.T) {
			if err := BootUpXORMContext(context.Background(), map[string]*XORMConfigLite{"test": cfg}, nil); !errors.Is(err, ErrInvalidDatabaseConfig) {
				t.Fatalf("expected ErrInvalidDatabaseConfig, got %v", err)
			}
		})
	}
}
