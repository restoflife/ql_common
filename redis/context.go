package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// SetContext is Set with caller-owned cancellation and deadline.
func SetContext(ctx context.Context, name, key string, value any, expiration time.Duration) error {
	if ctx == nil {
		err := ErrNilContext
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	return client.Set(ctx, key, value, expiration).Err()
}

// SetNXContext is SetNX with caller-owned cancellation and deadline.
func SetNXContext(ctx context.Context, name, key string, value any, expiration time.Duration) (bool, error) {
	if ctx == nil {
		err := ErrNilContext
		return false, err
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return false, err
	}
	return client.SetNX(ctx, key, value, expiration).Result()
}

// SetXXContext is SetXX with caller-owned cancellation and deadline.
func SetXXContext(ctx context.Context, name, key string, value any, expiration time.Duration) (bool, error) {
	if ctx == nil {
		err := ErrNilContext
		return false, err
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return false, err
	}
	return client.SetXX(ctx, key, value, expiration).Result()
}

// GetContext is Get with caller-owned cancellation and deadline.
func GetContext(ctx context.Context, name, key string) (string, error) {
	if ctx == nil {
		err := ErrNilContext
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.Get(ctx, key).Result()
}

// GetDelContext is GetDel with caller-owned cancellation and deadline.
func GetDelContext(ctx context.Context, name, key string) (string, error) {
	if ctx == nil {
		err := ErrNilContext
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.GetDel(ctx, key).Result()
}

// GetExContext is GetEx with caller-owned cancellation and deadline.
func GetExContext(ctx context.Context, name, key string, expiration time.Duration) (string, error) {
	if ctx == nil {
		err := ErrNilContext
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.GetEx(ctx, key, expiration).Result()
}

// IncrContext is Incr with caller-owned cancellation and deadline.
func IncrContext(ctx context.Context, name, key string) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.Incr(ctx, key).Result()
}

// IncrByContext is IncrBy with caller-owned cancellation and deadline.
func IncrByContext(ctx context.Context, name, key string, value int64) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.IncrBy(ctx, key, value).Result()
}

// IncrByFloatContext is IncrByFloat with caller-owned cancellation and deadline.
func IncrByFloatContext(ctx context.Context, name, key string, value float64) (float64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.IncrByFloat(ctx, key, value).Result()
}

// DecrContext is Decr with caller-owned cancellation and deadline.
func DecrContext(ctx context.Context, name, key string) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.Decr(ctx, key).Result()
}

// DecrByContext is DecrBy with caller-owned cancellation and deadline.
func DecrByContext(ctx context.Context, name, key string, value int64) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.DecrBy(ctx, key, value).Result()
}

// AppendContext is Append with caller-owned cancellation and deadline.
func AppendContext(ctx context.Context, name, key, value string) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.Append(ctx, key, value).Result()
}

// MGetContext is MGet with caller-owned cancellation and deadline.
func MGetContext(ctx context.Context, name string, keys ...string) ([]any, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.MGet(ctx, keys...).Result()
}

// MSetContext is MSet with caller-owned cancellation and deadline.
func MSetContext(ctx context.Context, name string, values ...any) error {
	if ctx == nil {
		err := ErrNilContext
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	return client.MSet(ctx, values...).Err()
}

// GetRangeContext is GetRange with caller-owned cancellation and deadline.
func GetRangeContext(ctx context.Context, name, key string, start, end int64) (string, error) {
	if ctx == nil {
		err := ErrNilContext
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.GetRange(ctx, key, start, end).Result()
}

// SetRangeContext is SetRange with caller-owned cancellation and deadline.
func SetRangeContext(ctx context.Context, name, key string, offset int64, value string) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.SetRange(ctx, key, offset, value).Result()
}

// StrLenContext is StrLen with caller-owned cancellation and deadline.
func StrLenContext(ctx context.Context, name, key string) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.StrLen(ctx, key).Result()
}

// HSetContext is HSet with caller-owned cancellation and deadline.
func HSetContext(ctx context.Context, name, key, field string, value any) error {
	if ctx == nil {
		err := ErrNilContext
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	return client.HSet(ctx, key, field, value).Err()
}

// HSetNXContext is HSetNX with caller-owned cancellation and deadline.
func HSetNXContext(ctx context.Context, name, key, field string, value any) (bool, error) {
	if ctx == nil {
		err := ErrNilContext
		return false, err
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return false, err
	}
	return client.HSetNX(ctx, key, field, value).Result()
}

// HGetContext is HGet with caller-owned cancellation and deadline.
func HGetContext(ctx context.Context, name, key, field string) (string, error) {
	if ctx == nil {
		err := ErrNilContext
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.HGet(ctx, key, field).Result()
}

// HGetAllContext is HGetAll with caller-owned cancellation and deadline.
func HGetAllContext(ctx context.Context, name, key string) (map[string]string, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.HGetAll(ctx, key).Result()
}

// HMSetContext is HMSet with caller-owned cancellation and deadline.
func HMSetContext(ctx context.Context, name, key string, values any) error {
	if ctx == nil {
		err := ErrNilContext
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	return client.HMSet(ctx, key, values).Err()
}

// HMGetContext is HMGet with caller-owned cancellation and deadline.
func HMGetContext(ctx context.Context, name, key string, fields ...string) ([]any, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.HMGet(ctx, key, fields...).Result()
}

// HDelContext is HDel with caller-owned cancellation and deadline.
func HDelContext(ctx context.Context, name, key string, fields ...string) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.HDel(ctx, key, fields...).Result()
}

// HExistsContext is HExists with caller-owned cancellation and deadline.
func HExistsContext(ctx context.Context, name, key, field string) (bool, error) {
	if ctx == nil {
		err := ErrNilContext
		return false, err
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return false, err
	}
	return client.HExists(ctx, key, field).Result()
}

// HIncrByContext is HIncrBy with caller-owned cancellation and deadline.
func HIncrByContext(ctx context.Context, name, key, field string, incr int64) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.HIncrBy(ctx, key, field, incr).Result()
}

// HIncrByFloatContext is HIncrByFloat with caller-owned cancellation and deadline.
func HIncrByFloatContext(ctx context.Context, name, key, field string, incr float64) (float64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.HIncrByFloat(ctx, key, field, incr).Result()
}

// HKeysContext is HKeys with caller-owned cancellation and deadline.
func HKeysContext(ctx context.Context, name, key string) ([]string, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.HKeys(ctx, key).Result()
}

// HLenContext is HLen with caller-owned cancellation and deadline.
func HLenContext(ctx context.Context, name, key string) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.HLen(ctx, key).Result()
}

// HValsContext is HVals with caller-owned cancellation and deadline.
func HValsContext(ctx context.Context, name, key string) ([]string, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.HVals(ctx, key).Result()
}

// LPushContext is LPush with caller-owned cancellation and deadline.
func LPushContext(ctx context.Context, name, key string, values ...any) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.LPush(ctx, key, values...).Result()
}

// RPushContext is RPush with caller-owned cancellation and deadline.
func RPushContext(ctx context.Context, name, key string, values ...any) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.RPush(ctx, key, values...).Result()
}

// LPopContext is LPop with caller-owned cancellation and deadline.
func LPopContext(ctx context.Context, name, key string) (string, error) {
	if ctx == nil {
		err := ErrNilContext
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.LPop(ctx, key).Result()
}

// RPopContext is RPop with caller-owned cancellation and deadline.
func RPopContext(ctx context.Context, name, key string) (string, error) {
	if ctx == nil {
		err := ErrNilContext
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.RPop(ctx, key).Result()
}

// LRangeContext is LRange with caller-owned cancellation and deadline.
func LRangeContext(ctx context.Context, name, key string, start, stop int64) ([]string, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.LRange(ctx, key, start, stop).Result()
}

// LLenContext is LLen with caller-owned cancellation and deadline.
func LLenContext(ctx context.Context, name, key string) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.LLen(ctx, key).Result()
}

// LIndexContext is LIndex with caller-owned cancellation and deadline.
func LIndexContext(ctx context.Context, name, key string, index int64) (string, error) {
	if ctx == nil {
		err := ErrNilContext
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.LIndex(ctx, key, index).Result()
}

// LInsertContext is LInsert with caller-owned cancellation and deadline.
func LInsertContext(ctx context.Context, name, key, op string, pivot, value any) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.LInsert(ctx, key, op, pivot, value).Result()
}

// LRemContext is LRem with caller-owned cancellation and deadline.
func LRemContext(ctx context.Context, name, key string, count int64, value any) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.LRem(ctx, key, count, value).Result()
}

// LTrimContext is LTrim with caller-owned cancellation and deadline.
func LTrimContext(ctx context.Context, name, key string, start, stop int64) error {
	if ctx == nil {
		err := ErrNilContext
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	return client.LTrim(ctx, key, start, stop).Err()
}

// LSetContext is LSet with caller-owned cancellation and deadline.
func LSetContext(ctx context.Context, name, key string, index int64, value any) error {
	if ctx == nil {
		err := ErrNilContext
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	return client.LSet(ctx, key, index, value).Err()
}

// SAddContext is SAdd with caller-owned cancellation and deadline.
func SAddContext(ctx context.Context, name, key string, members ...any) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.SAdd(ctx, key, members...).Result()
}

// SRemContext is SRem with caller-owned cancellation and deadline.
func SRemContext(ctx context.Context, name, key string, members ...any) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.SRem(ctx, key, members...).Result()
}

// SMembersContext is SMembers with caller-owned cancellation and deadline.
func SMembersContext(ctx context.Context, name, key string) ([]string, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.SMembers(ctx, key).Result()
}

// SIsMemberContext is SIsMember with caller-owned cancellation and deadline.
func SIsMemberContext(ctx context.Context, name, key string, member any) (bool, error) {
	if ctx == nil {
		err := ErrNilContext
		return false, err
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return false, err
	}
	return client.SIsMember(ctx, key, member).Result()
}

// SCardContext is SCard with caller-owned cancellation and deadline.
func SCardContext(ctx context.Context, name, key string) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.SCard(ctx, key).Result()
}

// SPopContext is SPop with caller-owned cancellation and deadline.
func SPopContext(ctx context.Context, name, key string) (string, error) {
	if ctx == nil {
		err := ErrNilContext
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.SPop(ctx, key).Result()
}

// SUnionContext is SUnion with caller-owned cancellation and deadline.
func SUnionContext(ctx context.Context, name string, keys ...string) ([]string, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.SUnion(ctx, keys...).Result()
}

// SInterContext is SInter with caller-owned cancellation and deadline.
func SInterContext(ctx context.Context, name string, keys ...string) ([]string, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.SInter(ctx, keys...).Result()
}

// SDiffContext is SDiff with caller-owned cancellation and deadline.
func SDiffContext(ctx context.Context, name string, keys ...string) ([]string, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.SDiff(ctx, keys...).Result()
}

// ZAddContext is ZAdd with caller-owned cancellation and deadline.
func ZAddContext(ctx context.Context, name, key string, members ...redis.Z) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.ZAdd(ctx, key, members...).Result()
}

// ZRemContext is ZRem with caller-owned cancellation and deadline.
func ZRemContext(ctx context.Context, name, key string, members ...any) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.ZRem(ctx, key, members...).Result()
}

// ZRangeContext is ZRange with caller-owned cancellation and deadline.
func ZRangeContext(ctx context.Context, name, key string, start, stop int64) ([]string, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeWithScoresContext is ZRangeWithScores with caller-owned cancellation and deadline.
func ZRangeWithScoresContext(ctx context.Context, name, key string, start, stop int64) ([]redis.Z, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.ZRangeWithScores(ctx, key, start, stop).Result()
}

// ZRankContext is ZRank with caller-owned cancellation and deadline.
func ZRankContext(ctx context.Context, name, key, member string) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.ZRank(ctx, key, member).Result()
}

// ZScoreContext is ZScore with caller-owned cancellation and deadline.
func ZScoreContext(ctx context.Context, name, key, member string) (float64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.ZScore(ctx, key, member).Result()
}

// ZIncrByContext is ZIncrBy with caller-owned cancellation and deadline.
func ZIncrByContext(ctx context.Context, name, key, member string, increment float64) (float64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.ZIncrBy(ctx, key, increment, member).Result()
}

// ZCardContext is ZCard with caller-owned cancellation and deadline.
func ZCardContext(ctx context.Context, name, key string) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.ZCard(ctx, key).Result()
}

// ZCountContext is ZCount with caller-owned cancellation and deadline.
func ZCountContext(ctx context.Context, name, key, min, max string) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.ZCount(ctx, key, min, max).Result()
}

// ZRemRangeByRankContext is ZRemRangeByRank with caller-owned cancellation and deadline.
func ZRemRangeByRankContext(ctx context.Context, name, key string, start, stop int64) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.ZRemRangeByRank(ctx, key, start, stop).Result()
}

// ZRemRangeByScoreContext is ZRemRangeByScore with caller-owned cancellation and deadline.
func ZRemRangeByScoreContext(ctx context.Context, name, key, min, max string) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.ZRemRangeByScore(ctx, key, min, max).Result()
}

// ExistsContext is Exists with caller-owned cancellation and deadline.
func ExistsContext(ctx context.Context, name string, keys ...string) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.Exists(ctx, keys...).Result()
}

// DelContext is Del with caller-owned cancellation and deadline.
func DelContext(ctx context.Context, name string, keys ...string) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.Del(ctx, keys...).Result()
}

// ExpireContext is Expire with caller-owned cancellation and deadline.
func ExpireContext(ctx context.Context, name, key string, expiration time.Duration) (bool, error) {
	if ctx == nil {
		err := ErrNilContext
		return false, err
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return false, err
	}
	return client.Expire(ctx, key, expiration).Result()
}

// ExpireAtContext is ExpireAt with caller-owned cancellation and deadline.
func ExpireAtContext(ctx context.Context, name, key string, tm time.Time) (bool, error) {
	if ctx == nil {
		err := ErrNilContext
		return false, err
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return false, err
	}
	return client.ExpireAt(ctx, key, tm).Result()
}

// TTLContext is TTL with caller-owned cancellation and deadline.
func TTLContext(ctx context.Context, name, key string) (time.Duration, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.TTL(ctx, key).Result()
}

// PersistContext is Persist with caller-owned cancellation and deadline.
func PersistContext(ctx context.Context, name, key string) (bool, error) {
	if ctx == nil {
		err := ErrNilContext
		return false, err
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return false, err
	}
	return client.Persist(ctx, key).Result()
}

// RenameContext is Rename with caller-owned cancellation and deadline.
func RenameContext(ctx context.Context, name, key, newkey string) error {
	if ctx == nil {
		err := ErrNilContext
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	return client.Rename(ctx, key, newkey).Err()
}

// RenameNXContext is RenameNX with caller-owned cancellation and deadline.
func RenameNXContext(ctx context.Context, name, key, newkey string) (bool, error) {
	if ctx == nil {
		err := ErrNilContext
		return false, err
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return false, err
	}
	return client.RenameNX(ctx, key, newkey).Result()
}

// TypeContext is Type with caller-owned cancellation and deadline.
func TypeContext(ctx context.Context, name, key string) (string, error) {
	if ctx == nil {
		err := ErrNilContext
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.Type(ctx, key).Result()
}

// KeysContext is Keys with caller-owned cancellation and deadline.
func KeysContext(ctx context.Context, name, pattern string) ([]string, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.Keys(ctx, pattern).Result()
}

// ScanContext is Scan with caller-owned cancellation and deadline.
func ScanContext(ctx context.Context, name string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, 0, err
	}
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return nil, 0, err
	}
	return client.Scan(ctx, cursor, match, count).Result()
}

// PipelineContext is Pipeline with caller-owned cancellation and deadline.
func PipelineContext(ctx context.Context, name string, fn func(redis.Pipeliner) error) error {
	if fn == nil {
		return ErrNilPipeline
	}

	if ctx == nil {
		err := ErrNilContext
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	_, err = client.Pipelined(ctx, fn)
	return err
}

// TxPipelineContext is TxPipeline with caller-owned cancellation and deadline.
func TxPipelineContext(ctx context.Context, name string, fn func(redis.Pipeliner) error) error {
	if fn == nil {
		return ErrNilPipeline
	}

	if ctx == nil {
		err := ErrNilContext
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	_, err = client.TxPipelined(ctx, fn)
	return err
}

// EvalContext is Eval with caller-owned cancellation and deadline.
func EvalContext(ctx context.Context, name string, script string, keys []string, args ...any) (any, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.Eval(ctx, script, keys, args...).Result()
}

// EvalShaContext is EvalSha with caller-owned cancellation and deadline.
func EvalShaContext(ctx context.Context, name, sha1 string, keys []string, args ...any) (any, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.EvalSha(ctx, sha1, keys, args...).Result()
}

// ScriptLoadContext is ScriptLoad with caller-owned cancellation and deadline.
func ScriptLoadContext(ctx context.Context, name, script string) (string, error) {
	if ctx == nil {
		err := ErrNilContext
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.ScriptLoad(ctx, script).Result()
}

// ScriptExistsContext is ScriptExists with caller-owned cancellation and deadline.
func ScriptExistsContext(ctx context.Context, name string, hashes ...string) ([]bool, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.ScriptExists(ctx, hashes...).Result()
}

// PFAddContext is PFAdd with caller-owned cancellation and deadline.
func PFAddContext(ctx context.Context, name, key string, elements ...any) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.PFAdd(ctx, key, elements...).Result()
}

// PFCountContext is PFCount with caller-owned cancellation and deadline.
func PFCountContext(ctx context.Context, name string, keys ...string) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.PFCount(ctx, keys...).Result()
}

// PFMergeContext is PFMerge with caller-owned cancellation and deadline.
func PFMergeContext(ctx context.Context, name, dest string, keys ...string) error {
	if ctx == nil {
		err := ErrNilContext
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	return client.PFMerge(ctx, dest, keys...).Err()
}

// SetBitContext is SetBit with caller-owned cancellation and deadline.
func SetBitContext(ctx context.Context, name, key string, offset int64, value int) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.SetBit(ctx, key, offset, value).Result()
}

// GetBitContext is GetBit with caller-owned cancellation and deadline.
func GetBitContext(ctx context.Context, name, key string, offset int64) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.GetBit(ctx, key, offset).Result()
}

// BitCountContext is BitCount with caller-owned cancellation and deadline.
func BitCountContext(ctx context.Context, name, key string) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.BitCount(ctx, key, nil).Result()
}
