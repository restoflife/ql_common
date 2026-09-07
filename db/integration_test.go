//go:build integration

package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
	"xorm.io/xorm"
)

// Run only against a disposable database: this test creates and drops one table.
func TestIntegrationTransactions(t *testing.T) {
	dsn := os.Getenv("QL_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("QL_TEST_MYSQL_DSN not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	configs := map[string]*XORMConfigLite{"integration": {Driver: "mysql", Dsn: dsn, MaxLife: 60, MaxOpen: 4, MaxIdle: 2}}
	if err := BootUpXORMContext(ctx, configs, nil); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := ShutdownXormE(); err != nil {
			t.Error(err)
		}
	})
	table := fmt.Sprintf("ql_common_test_%d", time.Now().UnixNano())
	session, err := NewSessionContext(ctx, "integration")
	if err != nil {
		t.Fatal(err)
	}
	defer Close(session)
	if _, err := session.Exec("CREATE TABLE " + table + " (id BIGINT PRIMARY KEY)"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := session.Exec("DROP TABLE " + table); err != nil {
			t.Error(err)
		}
	}()
	if err := Transaction(ctx, "integration", func(s *xorm.Session) error { _, err := s.Exec("INSERT INTO " + table + " VALUES (1)"); return err }); err != nil {
		t.Fatal(err)
	}
	rollback := errors.New("rollback")
	err = Transaction(ctx, "integration", func(s *xorm.Session) error {
		if _, err := s.Exec("INSERT INTO " + table + " VALUES (2)"); err != nil {
			return err
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatal(err)
	}
	rows, err := session.QueryString("SELECT id FROM " + table)
	if err != nil || len(rows) != 1 || rows[0]["id"] != "1" {
		t.Fatalf("rows=%v err=%v", rows, err)
	}
	if err := BootUpXORMContext(ctx, configs, nil); !errors.Is(err, ErrDuplicate) {
		t.Fatal(err)
	}
}
