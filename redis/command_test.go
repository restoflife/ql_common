package redis

import (
	"context"
	driver "github.com/redis/go-redis/v9"
	"reflect"
	"testing"
)

type commandClient struct {
	testClient
	args []any
}

func (c *commandClient) Do(ctx context.Context, args ...any) *driver.Cmd {
	c.seen = ctx
	c.args = args
	r := driver.NewCmd(ctx, args...)
	r.SetVal(int64(1))
	return r
}
func TestCustomCommandUsesRegistry(t *testing.T) {
	_ = ShutdownRedisE()
	t.Cleanup(func() { _ = ShutdownRedisE() })
	c := &commandClient{}
	if err := redisMgr.Init([]string{"game"}, func(string) (driver.UniversalClient, error) { return c, nil }, func(c driver.UniversalClient) error { return c.Close() }); err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(context.Background(), contextKey{}, "marker")
	args := []any{"ZADD", "rank", 100, "player"}
	got, err := DoContext(ctx, "game", args...)
	if err != nil || got != int64(1) || c.seen != ctx || !reflect.DeepEqual(c.args, args) {
		t.Fatal("command forwarding failed", err)
	}
	if _, err := DoContext(ctx, "game"); err == nil {
		t.Fatal("empty command")
	}
	if _, err := DoContext(nil, "game", args...); err == nil {
		t.Fatal("nil context")
	}
	if err := ShutdownRedisE(); err != nil {
		t.Fatal(err)
	}
	if _, err := DoContext(ctx, "game", args...); err == nil {
		t.Fatal("closed client returned")
	}
}
