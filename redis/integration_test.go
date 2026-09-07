//go:build integration

package redis

import (
	"context"
	"errors"
	"fmt"
	driver "github.com/redis/go-redis/v9"
	"os"
	"testing"
	"time"
)

func TestIntegrationCommandsAndRestart(t *testing.T) {
	addr := os.Getenv("QL_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("QL_TEST_REDIS_ADDR not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	configs := map[string]*Config{"integration": {Addr: addr}}
	if err := BootUpRedisContext(ctx, configs); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := ShutdownRedisE(); err != nil {
			t.Error(err)
		}
	})
	key := fmt.Sprintf("ql_common_test:%d", time.Now().UnixNano())
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = DelContext(cleanup, "integration", key)
	}()
	if err := SetContext(ctx, "integration", key, "1", time.Minute); err != nil {
		t.Fatal(err)
	}
	if value, err := GetContext(ctx, "integration", key); err != nil || value != "1" {
		t.Fatalf("value=%q err=%v", value, err)
	}
	if err := PipelineContext(ctx, "integration", func(p driver.Pipeliner) error { p.Incr(ctx, key); return nil }); err != nil {
		t.Fatal(err)
	}
	result, err := EvalContext(ctx, "integration", "return redis.call('GET', KEYS[1])", []string{key})
	if err != nil || result != "2" {
		t.Fatalf("result=%v err=%v", result, err)
	}
	if err := BootUpRedisContext(ctx, configs); !errors.Is(err, ErrDuplicate) {
		t.Fatal(err)
	}
	if err := ShutdownRedisE(); err != nil {
		t.Fatal(err)
	}
	if err := ShutdownRedisE(); err != nil {
		t.Fatal(err)
	}
	if err := BootUpRedisContext(ctx, configs); err != nil {
		t.Fatal(err)
	}
	if value, err := GetContext(ctx, "integration", key); err != nil || value != "2" {
		t.Fatalf("value=%q err=%v", value, err)
	}
}
