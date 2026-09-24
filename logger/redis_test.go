package logger

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestRedisHookLogsSafeCommandMetadata(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	hook := NewRedisHook(zap.New(core))

	set := redis.NewStatusCmd(context.Background(), "set", "session:42", "secret-token")
	if err := hook.ProcessHook(func(context.Context, redis.Cmder) error { return nil })(context.Background(), set); err != nil {
		t.Fatal(err)
	}
	auth := redis.NewStatusCmd(context.Background(), "auth", "user", "secret-password")
	if err := hook.ProcessHook(func(context.Context, redis.Cmder) error { return nil })(context.Background(), auth); err != nil {
		t.Fatal(err)
	}

	entries := logs.All()
	if len(entries) != 2 {
		t.Fatalf("expected two entries, got %d", len(entries))
	}
	if entries[0].Level != zap.InfoLevel || entries[0].Message != REDIS {
		t.Fatal(entries[0])
	}
	fields := entries[0].ContextMap()
	if fields["command"] != "SET" || fields["key"] != "session:42" {
		t.Fatalf("unexpected safe fields: %v", fields)
	}
	for _, entry := range entries {
		text := entry.Message
		for _, field := range entry.Context {
			text += field.String
		}
		if strings.Contains(text, "secret-token") || strings.Contains(text, "secret-password") {
			t.Fatalf("secret was logged: %s", text)
		}
	}
	if _, exists := entries[1].ContextMap()["key"]; exists {
		t.Fatal("AUTH argument was logged as a key")
	}
}

func TestRedisHookFailureIncludesCallerAndStack(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	hook := NewRedisHook(zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel)))
	want := errors.New("redis unavailable")
	cmd := redis.NewStringCmd(context.Background(), "get", "player:7")

	err := hook.ProcessHook(func(context.Context, redis.Cmder) error { return want })(context.Background(), cmd)
	if !errors.Is(err, want) {
		t.Fatal(err)
	}
	entry := logs.All()[0]
	if entry.Level != zap.ErrorLevel || !entry.Caller.Defined || entry.Stack == "" {
		t.Fatalf("missing error diagnostics: caller=%v stack=%q", entry.Caller, entry.Stack)
	}
	fields := entry.ContextMap()
	if _, exists := fields["caller"]; exists {
		t.Fatalf("unexpected custom caller field: %v", fields)
	}
	if fields["error"] != want.Error() {
		t.Fatalf("unexpected failure fields: %v", fields)
	}
}

func TestRedisHookPipelineLogsBoundedSafeSummary(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	hook := NewRedisHook(zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel)))
	cmds := []redis.Cmder{
		redis.NewStatusCmd(context.Background(), "set", "session:1", "secret-one"),
		redis.NewStringCmd(context.Background(), "get", "profile:1"),
		redis.NewStatusCmd(context.Background(), "auth", "secret-two"),
	}

	want := errors.New("pipeline failed")
	err := hook.ProcessPipelineHook(func(context.Context, []redis.Cmder) error { return want })(context.Background(), cmds)
	if !errors.Is(err, want) {
		t.Fatal(err)
	}
	entry := logs.All()[0]
	if entry.Level != zap.ErrorLevel || entry.Message != REDIS || entry.Stack == "" {
		t.Fatal(entry)
	}
	fields := entry.ContextMap()
	if fields["command_count"] != int64(3) {
		t.Fatalf("unexpected command count: %v", fields)
	}
	for _, field := range entry.Context {
		if strings.Contains(field.String, "secret-one") || strings.Contains(field.String, "secret-two") {
			t.Fatalf("pipeline secret was logged: %s", field.String)
		}
	}
}

func TestRedisHookSuccessUsesInfoLevel(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	hook := NewRedisHook(zap.New(core))
	cmd := redis.NewStringCmd(context.Background(), "get", "key")
	_ = hook.ProcessHook(func(context.Context, redis.Cmder) error { return nil })(context.Background(), cmd)
	if logs.Len() != 1 || logs.All()[0].Level != zap.InfoLevel {
		t.Fatalf("successful command was not emitted at Info level: %v", logs.All())
	}
}

func TestRedisHookSkipsDriverNegotiationCommands(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	hook := NewRedisHook(zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel)))
	want := errors.New("unsupported by old redis")

	hello := redis.NewStatusCmd(context.Background(), "hello", 3)
	if err := hook.ProcessHook(func(context.Context, redis.Cmder) error { return want })(context.Background(), hello); !errors.Is(err, want) {
		t.Fatal(err)
	}
	client := redis.NewStatusCmd(context.Background(), "client", "setinfo", "LIB-NAME", "go-redis")
	client.SetErr(want)
	if err := hook.ProcessPipelineHook(func(context.Context, []redis.Cmder) error { return want })(context.Background(), []redis.Cmder{client}); !errors.Is(err, want) {
		t.Fatal(err)
	}
	if logs.Len() != 0 {
		t.Fatalf("driver negotiation commands were logged: %v", logs.All())
	}
}

func TestRedisHookPipelineIgnoresNegotiationFailureForBusinessSuccess(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	hook := NewRedisHook(zap.New(core))
	want := errors.New("unsupported by old redis")
	client := redis.NewStatusCmd(context.Background(), "client", "setinfo", "LIB-NAME", "go-redis")
	client.SetErr(want)
	get := redis.NewStringCmd(context.Background(), "get", "profile:1")
	if err := hook.ProcessPipelineHook(func(context.Context, []redis.Cmder) error { return want })(context.Background(), []redis.Cmder{client, get}); !errors.Is(err, want) {
		t.Fatal(err)
	}
	entries := logs.All()
	if len(entries) != 1 || entries[0].Level != zap.InfoLevel {
		t.Fatalf("business command should be logged as successful: %v", entries)
	}
	fields := entries[0].ContextMap()
	if fields["command_count"] != int64(1) {
		t.Fatalf("internal command was not filtered: %v", fields)
	}
}
