package logger

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	xlog "xorm.io/xorm/log"
)

func TestConfigDefaultsValidationAndRegistration(t *testing.T) {
	_ = CloseAll()
	t.Cleanup(func() { _ = CloseAll() })
	Info("safe before initialization")
	config := Config{Filename: filepath.Join(t.TempDir(), "test.log")}
	if err := Init(&config); err != nil {
		t.Fatal(err)
	}
	if config.Level != "" || config.MaxSize != 0 {
		t.Fatal("mutated caller config")
	}
	if Logger() == nil || len(GetAll()) != 1 || GetAll()[0] != Logger() {
		t.Fatal("default logger not registered")
	}
	Info("hello")
	SyncAll()
	for _, c := range []Config{{Level: "bad"}, {Console: "bad"}, {Format: "xml"}, {MaxSize: -1}, {MaxBackups: -1}, {MaxAge: -1}} {
		if _, err := c.Build(); err == nil {
			t.Fatalf("accepted invalid config %+v", c)
		}
	}
	if err := CloseAll(); err != nil {
		t.Fatal(err)
	}
	if Logger() != nil || len(GetAll()) != 0 {
		t.Fatal("not cleared")
	}
	if err := CloseAll(); err != nil {
		t.Fatal(err)
	}
}

func TestRecoveryReusesNonEmptyStack(t *testing.T) {
	gin.SetMode(gin.TestMode)
	core, logs := observer.New(zapcore.DebugLevel)
	engine := gin.New()
	engine.Use(Recovery(zap.New(core)))
	engine.GET("/panic", func(c *gin.Context) { panic("boom") })
	for i := 0; i < 8; i++ {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/panic?token=secret-value", nil)
		request.Header.Set("Authorization", "secret-value")
		engine.ServeHTTP(response, request)
		if response.Code != 500 {
			t.Fatal(response.Code)
		}
	}
	if logs.Len() != 8 {
		t.Fatal(logs.Len())
	}
	for _, entry := range logs.All() {
		fields := entry.ContextMap()
		stack, ok := fields["stacktrace"].(string)
		if !ok || !strings.Contains(stack, "TestRecoveryReusesNonEmptyStack") {
			t.Fatalf("missing stack: %v", fields)
		}
		if fields["path"] != "/panic" {
			t.Fatal(fields)
		}
		if _, ok := fields["request"]; ok {
			t.Fatal("request headers logged")
		}
	}
}

func TestGinWriterAndSkipPaths(t *testing.T) {
	var output bytes.Buffer
	core, logs := observer.New(zap.InfoLevel)
	engine := gin.New()
	engine.Use(WithWriter(zap.New(core), &output, "/health"))
	engine.GET("/health", func(c *gin.Context) { c.Status(200) })
	engine.GET("/ok", func(c *gin.Context) { c.Status(200) })
	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/health", nil))
	if output.Len() != 0 {
		t.Fatal("skip path logged")
	}
	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/ok?token=secret-value", nil))
	if output.Len() == 0 || strings.Contains(output.String(), "secret-value") || logs.Len() != 0 {
		t.Fatalf("output=%s original=%d", output.String(), logs.Len())
	}
}

func TestAdaptersConcurrencyAndPayloadSafety(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	l := zap.New(core)
	x := NewXormLogger(l)
	m := NewMongoLogger(l)
	raw, _ := bson.Marshal(bson.M{"password": "secret-value"})
	m.commandStarted(context.Background(), &event.CommandStartedEvent{CommandName: "find", Command: raw})
	m.commandSucceeded(context.Background(), &event.CommandSucceededEvent{Reply: raw})
	x.AfterSQL(xlog.LogContext{SQL: "select ?", Args: []interface{}{int64(12345)}})
	for _, entry := range logs.All() {
		for _, field := range entry.Context {
			if strings.Contains(field.String, "secret-value") {
				t.Fatal("payload logged")
			}
		}
	}
	x.SetLevel(xlog.LOG_OFF)
	before := logs.Len()
	x.AfterSQL(xlog.LogContext{SQL: "select 1"})
	if logs.Len() != before {
		t.Fatal("LOG_OFF ignored")
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				x.ShowSQL(j%2 == 0)
				x.SetLevel(xlog.LOG_INFO)
				x.AfterSQL(xlog.LogContext{SQL: "select ?", Err: errors.New("test")})
				m.ShowMongo(j%2 == 0)
				_ = m.IsShowMongo()
			}
		}()
	}
	wg.Wait()
}

func TestDatabaseAdaptersOnlyAttachCallerToErrors(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	base := zap.New(core, zap.AddCaller())

	mongoLogger := NewMongoLogger(base)
	mongoLogger.commandSucceeded(context.Background(), &event.CommandSucceededEvent{CommandFinishedEvent: event.CommandFinishedEvent{CommandName: "ping", DatabaseName: "admin"}})
	mongoLogger.commandFailed(context.Background(), &event.CommandFailedEvent{CommandFinishedEvent: event.CommandFinishedEvent{CommandName: "find", DatabaseName: "test"}, Failure: errors.New("failed")})

	xormLogger := NewXormLogger(base)
	xormLogger.AfterSQL(xlog.LogContext{SQL: "SELECT 1"})
	xormLogger.AfterSQL(xlog.LogContext{SQL: "SELECT 1", Err: errors.New("failed")})

	entries := logs.All()
	if len(entries) != 4 {
		t.Fatalf("expected four entries, got %d", len(entries))
	}
	for _, index := range []int{0, 2} {
		if entries[index].Caller.Defined {
			t.Fatalf("success entry %d unexpectedly has caller %s", index, entries[index].Caller)
		}
	}
	for _, index := range []int{1, 3} {
		if !entries[index].Caller.Defined {
			t.Fatalf("error entry %d has no caller", index)
		}
	}
}

func TestLoggerCoreOnlyWritesCallerForErrors(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	log := zap.New(errorCallerCore{Core: core}, zap.AddCaller())
	log.Info("info")
	log.Warn("warn")
	log.Error("error")

	entries := logs.All()
	if len(entries) != 3 {
		t.Fatalf("expected three entries, got %d", len(entries))
	}
	if entries[0].Caller.Defined || entries[1].Caller.Defined {
		t.Fatal("caller was written below error level")
	}
	if !entries[2].Caller.Defined {
		t.Fatal("error caller was not written")
	}
}
