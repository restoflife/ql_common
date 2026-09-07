package redis

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	driver "github.com/redis/go-redis/v9"
)

type contextKey struct{}
type testClient struct {
	driver.UniversalClient
	closed atomic.Int32
	seen   context.Context
}

func (c *testClient) Close() error { c.closed.Add(1); return nil }
func (c *testClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *driver.StatusCmd {
	c.seen = ctx
	result := driver.NewStatusCmd(ctx)
	result.SetVal("OK")
	return result
}

func TestContextForwardingAndRestart(t *testing.T) {
	_ = ShutdownRedisE()
	t.Cleanup(func() { _ = ShutdownRedisE() })
	c := &testClient{}
	add := func() error {
		return redisMgr.Init([]string{"test"}, func(string) (driver.UniversalClient, error) { return c, nil }, func(c driver.UniversalClient) error { return c.Close() })
	}
	if err := add(); err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(context.Background(), contextKey{}, "marker")
	if err := SetContext(ctx, "test", "key", "value", time.Second); err != nil {
		t.Fatal(err)
	}
	if c.seen != ctx {
		t.Fatal("caller context not forwarded")
	}
	if err := add(); !errors.Is(err, ErrDuplicate) {
		t.Fatal(err)
	}
	if err := ShutdownRedisE(); err != nil {
		t.Fatal(err)
	}
	if err := ShutdownRedisE(); err != nil {
		t.Fatal(err)
	}
	if c.closed.Load() != 1 {
		t.Fatal("closed repeatedly")
	}
	if _, err := GetRedis("test"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if err := add(); err != nil {
		t.Fatal("restart failed", err)
	}
}

func TestValidationAndCancellation(t *testing.T) {
	for _, c := range []*Config{nil, {}, {Mode: "typo", Addr: "localhost:6379"}, {Mode: SENTINEL}, {Mode: CLUSTER, Slaves: []string{"localhost:6379"}, DB: 1}, {Addr: "localhost:6379", PoolSize: -1}} {
		if err := MustBootUpRedis(map[string]*Config{"bad": c}); err == nil {
			t.Fatalf("accepted %+v", c)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := GetContext(ctx, "missing", "key"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := GetContext(nil, "missing", "key"); err == nil {
		t.Fatal("nil context accepted")
	}
	if err := PipelineContext(context.Background(), "missing", nil); err == nil {
		t.Fatal("nil callback accepted")
	}
	if err := BootUpRedisContext(nil, nil); err == nil {
		t.Fatal("nil context accepted")
	}
	if err := BootUpRedisContext(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
