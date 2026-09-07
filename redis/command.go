package redis

import (
	"context"
	driver "github.com/redis/go-redis/v9"
)

// DoContext executes an arbitrary Redis command through the registered client.
// Example: DoContext(ctx, "game", "ZADD", "rank", 100, "123"). Pass arguments separately.
// Prefer typed helpers for standard commands. The caller handles the returned Redis value.
func DoContext(ctx context.Context, name string, args ...any) (any, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(args) == 0 {
		return nil, ErrCommandRequired
	}
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.Do(ctx, args...).Result()
}

// Do is the compatibility entry point; HTTP handlers should use DoContext.
func Do(name string, args ...any) (any, error) { return DoContext(context.Background(), name, args...) }

// Z is a sorted-set member, usable without importing the Redis driver separately.
type Z = driver.Z

// ZRevRangeContext returns members ordered by descending score (stop is inclusive).
func ZRevRangeContext(ctx context.Context, name, key string, start, stop int64) ([]string, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.ZRevRange(ctx, key, start, stop).Result()
}

// ZRevRangeWithScoresContext returns descending members together with their scores.
func ZRevRangeWithScoresContext(ctx context.Context, name, key string, start, stop int64) ([]Z, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.ZRevRangeWithScores(ctx, key, start, stop).Result()
}
