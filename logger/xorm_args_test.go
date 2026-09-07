package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"strings"
	"testing"
	"time"
	xlog "xorm.io/xorm/log"
)

func TestXormLoggerBindsSQLArguments(t *testing.T) {
	core, entries := observer.New(zap.DebugLevel)
	logger := NewXormLogger(zap.New(core))
	logger.AfterSQL(xlog.LogContext{SQL: "SELECT * FROM player WHERE player_id=? AND name=?", Args: []interface{}{int64(12345), "a'b"}, ExecuteTime: time.Millisecond})
	if entries.Len() != 1 {
		t.Fatal("expected one SQL log")
	}
	fields := entries.All()[0].ContextMap()
	sql, _ := fields["sql"].(string)
	if !strings.Contains(sql, "player_id=12345") || !strings.Contains(sql, "name='a''b'") || strings.Contains(sql, "?") || fields["latency"] != "1ms" {
		t.Fatalf("arguments were not bound: %v", fields)
	}
	logger.ShowSQL(false)
	logger.AfterSQL(xlog.LogContext{SQL: "SELECT ?", Args: []interface{}{1}})
	if entries.Len() != 1 {
		t.Fatal("disabled logger wrote SQL")
	}
}
