package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"sync/atomic"
	"testing"

	"xorm.io/xorm"
	"xorm.io/xorm/core"
)

type reconnectConnector struct{ attempts *atomic.Int32 }

func (c reconnectConnector) Connect(context.Context) (driver.Conn, error) {
	return &reconnectConn{attempts: c.attempts}, nil
}
func (c reconnectConnector) Driver() driver.Driver { return reconnectDriver{attempts: c.attempts} }

type reconnectDriver struct{ attempts *atomic.Int32 }

func (d reconnectDriver) Open(string) (driver.Conn, error) {
	return &reconnectConn{attempts: d.attempts}, nil
}

type reconnectConn struct{ attempts *atomic.Int32 }

func (c *reconnectConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}
func (c *reconnectConn) Close() error              { return nil }
func (c *reconnectConn) Begin() (driver.Tx, error) { return reconnectTx{}, nil }
func (c *reconnectConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	if c.attempts.Add(1) == 1 {
		return nil, driver.ErrBadConn
	}
	return &reconnectRows{}, nil
}

type reconnectTx struct{}

func (reconnectTx) Commit() error   { return nil }
func (reconnectTx) Rollback() error { return nil }

type reconnectRows struct{ sent bool }

func (*reconnectRows) Columns() []string { return []string{"value"} }
func (*reconnectRows) Close() error      { return nil }
func (r *reconnectRows) Next(values []driver.Value) error {
	if r.sent {
		return io.EOF
	}
	r.sent = true
	values[0] = "ok"
	return nil
}

// TestQueryRetriesBadConnection provides the corresponding package operation.
func TestQueryRetriesBadConnection(t *testing.T) {
	_ = ShutdownXormE()
	attempts := &atomic.Int32{}
	sqlDB := sql.OpenDB(reconnectConnector{attempts: attempts})
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
	if err = dbMgr.Init([]string{"reconnect"}, func(string) (*xorm.EngineGroup, error) {
		return group, nil
	}, closeEngineGroup); err != nil {
		_ = group.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ShutdownXormE() })

	session, err := NewSessionContext(context.Background(), "reconnect")
	if err != nil {
		t.Fatal(err)
	}
	defer Close(session)
	rows, err := session.SQL("SELECT value").QueryString()
	if err != nil {
		t.Fatalf("the current query did not recover automatically: %v", err)
	}
	if attempts.Load() != 2 || len(rows) != 1 || rows[0]["value"] != "ok" {
		t.Fatalf("attempts=%d rows=%v", attempts.Load(), rows)
	}
}
