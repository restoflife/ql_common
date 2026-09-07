package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/restoflife/ql_common/internal/registry"
	"github.com/restoflife/ql_common/logger"
	"go.uber.org/zap"
)

var redisMgr registry.Registry[redis.UniversalClient]

// MustBootUpRedis retains the original API with a bounded startup timeout.
// Deprecated: Use BootUpRedisContext.
func MustBootUpRedis(configs map[string]*Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return BootUpRedisContext(ctx, configs)
}

// BootUpRedisContext initializes a batch atomically without background goroutines.
func BootUpRedisContext(ctx context.Context, configs map[string]*Config) error {
	if ctx == nil {
		return ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	names := make([]string, 0, len(configs))
	for name, c := range configs {
		if c == nil {
			return configError(name, "configuration is nil")
		}
		if c.MaxActiveConns < 0 || c.MaxIdleConns < 0 || c.ConnMaxIdleTime < 0 || c.DB < 0 || c.PoolSize < 0 || c.MinIdle < 0 || (c.PoolSize > 0 && c.MinIdle > c.PoolSize) {
			return configError(name, "invalid pool or database settings")
		}
		switch c.Mode {
		case "", "standalone":
			if c.Addr == "" {
				return configError(name, "address is required")
			}
		case SENTINEL:
			if c.MasterName == "" || len(c.Slaves) == 0 {
				return configError(name, "master and sentinel addresses are required")
			}
		case CLUSTER:
			if len(c.Slaves) == 0 || c.DB != 0 {
				return configError(name, "cluster addresses are required and DB must be zero")
			}
		default:
			return configError(name, fmt.Sprintf("unsupported mode %q", c.Mode))
		}
		for _, addr := range c.Slaves {
			if addr == "" {
				return configError(name, "node address is empty")
			}
		}
		names = append(names, name)
	}
	return redisMgr.Init(names, func(name string) (redis.UniversalClient, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		c := configs[name]
		var client redis.UniversalClient
		switch c.Mode {
		case SENTINEL:
			client = redis.NewFailoverClient(&redis.FailoverOptions{
				MasterName: c.MasterName, SentinelAddrs: c.Slaves, Password: c.Password,
				DB: c.DB, PoolSize: c.PoolSize, MinIdleConns: c.MinIdle, ContextTimeoutEnabled: true,
				MaxActiveConns: c.MaxActiveConns, MaxIdleConns: c.MaxIdleConns, ConnMaxIdleTime: c.ConnMaxIdleTime, TLSConfig: c.TLSConfig,
			})
		case CLUSTER:
			client = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs: c.Slaves, Password: c.Password, PoolSize: c.PoolSize, MinIdleConns: c.MinIdle, ContextTimeoutEnabled: true,
				MaxActiveConns: c.MaxActiveConns, MaxIdleConns: c.MaxIdleConns, ConnMaxIdleTime: c.ConnMaxIdleTime, TLSConfig: c.TLSConfig,
			})
		default:
			client = redis.NewClient(&redis.Options{
				Addr: c.Addr, Password: c.Password, DB: c.DB,
				PoolSize: c.PoolSize, MinIdleConns: c.MinIdle, ContextTimeoutEnabled: true,
				MaxActiveConns: c.MaxActiveConns, MaxIdleConns: c.MaxIdleConns, ConnMaxIdleTime: c.ConnMaxIdleTime, TLSConfig: c.TLSConfig,
			})
		}
		if err := client.Ping(ctx).Err(); err != nil {
			_ = client.Close()
			return nil, err
		}
		return client, nil
	}, func(c redis.UniversalClient) error { return c.Close() })
}

// GetRedis returns the driver client for context-aware commands and advanced APIs.
func GetRedis(name string) (redis.UniversalClient, error) { return redisMgr.Get(name) }

// Deprecated: Use ShutdownRedisE.
func ShutdownRedis() {
	if err := ShutdownRedisE(); err != nil {
		logger.Error("Redis shutdown failed", zap.Error(err))
	}
}

// ShutdownRedisE unregisters clients and reports close failures. Drain users first.
func ShutdownRedisE() error {
	return redisMgr.Close(func(c redis.UniversalClient) error { return c.Close() })
}

// String operations.

// Set provides the corresponding package operation.
// Deprecated: Use SetContext.
func Set(name, key string, value any, expiration time.Duration) error {
	return SetContext(context.Background(), name, key, value, expiration)
}

// SetNX provides the corresponding package operation.
// Deprecated: Use SetNXContext.
func SetNX(name, key string, value any, expiration time.Duration) (bool, error) {
	return SetNXContext(context.Background(), name, key, value, expiration)
}

// SetXX provides the corresponding package operation.
// Deprecated: Use SetXXContext.
func SetXX(name, key string, value any, expiration time.Duration) (bool, error) {
	return SetXXContext(context.Background(), name, key, value, expiration)
}

// Get provides the corresponding package operation.
// Deprecated: Use GetContext.
func Get(name, key string) (string, error) {
	return GetContext(context.Background(), name, key)
}

// GetDel provides the corresponding package operation.
// Deprecated: Use GetDelContext.
func GetDel(name, key string) (string, error) {
	return GetDelContext(context.Background(), name, key)
}

// GetEx provides the corresponding package operation.
// Deprecated: Use GetExContext.
func GetEx(name, key string, expiration time.Duration) (string, error) {
	return GetExContext(context.Background(), name, key, expiration)
}

// Incr provides the corresponding package operation.
// Deprecated: Use IncrContext.
func Incr(name, key string) (int64, error) {
	return IncrContext(context.Background(), name, key)
}

// IncrBy provides the corresponding package operation.
// Deprecated: Use IncrByContext.
func IncrBy(name, key string, value int64) (int64, error) {
	return IncrByContext(context.Background(), name, key, value)
}

// IncrByFloat provides the corresponding package operation.
// Deprecated: Use IncrByFloatContext.
func IncrByFloat(name, key string, value float64) (float64, error) {
	return IncrByFloatContext(context.Background(), name, key, value)
}

// Decr provides the corresponding package operation.
// Deprecated: Use DecrContext.
func Decr(name, key string) (int64, error) {
	return DecrContext(context.Background(), name, key)
}

// DecrBy provides the corresponding package operation.
// Deprecated: Use DecrByContext.
func DecrBy(name, key string, value int64) (int64, error) {
	return DecrByContext(context.Background(), name, key, value)
}

// Append provides the corresponding package operation.
// Deprecated: Use AppendContext.
func Append(name, key, value string) (int64, error) {
	return AppendContext(context.Background(), name, key, value)
}

// MGet provides the corresponding package operation.
// Deprecated: Use MGetContext.
func MGet(name string, keys ...string) ([]any, error) {
	return MGetContext(context.Background(), name, keys...)
}

// MSet provides the corresponding package operation.
// Deprecated: Use MSetContext.
func MSet(name string, values ...any) error {
	return MSetContext(context.Background(), name, values...)
}

// GetRange provides the corresponding package operation.
// Deprecated: Use GetRangeContext.
func GetRange(name, key string, start, end int64) (string, error) {
	return GetRangeContext(context.Background(), name, key, start, end)
}

// SetRange provides the corresponding package operation.
// Deprecated: Use SetRangeContext.
func SetRange(name, key string, offset int64, value string) (int64, error) {
	return SetRangeContext(context.Background(), name, key, offset, value)
}

// StrLen provides the corresponding package operation.
// Deprecated: Use StrLenContext.
func StrLen(name, key string) (int64, error) {
	return StrLenContext(context.Background(), name, key)
}

// Hash operations.

// HSet provides the corresponding package operation.
// Deprecated: Use HSetContext.
func HSet(name, key, field string, value any) error {
	return HSetContext(context.Background(), name, key, field, value)
}

// HSetNX provides the corresponding package operation.
// Deprecated: Use HSetNXContext.
func HSetNX(name, key, field string, value any) (bool, error) {
	return HSetNXContext(context.Background(), name, key, field, value)
}

// HGet provides the corresponding package operation.
// Deprecated: Use HGetContext.
func HGet(name, key, field string) (string, error) {
	return HGetContext(context.Background(), name, key, field)
}

// HGetAll provides the corresponding package operation.
// Deprecated: Use HGetAllContext.
func HGetAll(name, key string) (map[string]string, error) {
	return HGetAllContext(context.Background(), name, key)
}

// HMSet provides the corresponding package operation.
// Deprecated: Use HMSetContext.
func HMSet(name, key string, values any) error {
	return HMSetContext(context.Background(), name, key, values)
}

// HMGet provides the corresponding package operation.
// Deprecated: Use HMGetContext.
func HMGet(name, key string, fields ...string) ([]any, error) {
	return HMGetContext(context.Background(), name, key, fields...)
}

// HDel provides the corresponding package operation.
// Deprecated: Use HDelContext.
func HDel(name, key string, fields ...string) (int64, error) {
	return HDelContext(context.Background(), name, key, fields...)
}

// HExists provides the corresponding package operation.
// Deprecated: Use HExistsContext.
func HExists(name, key, field string) (bool, error) {
	return HExistsContext(context.Background(), name, key, field)
}

// HIncrBy provides the corresponding package operation.
// Deprecated: Use HIncrByContext.
func HIncrBy(name, key, field string, incr int64) (int64, error) {
	return HIncrByContext(context.Background(), name, key, field, incr)
}

// HIncrByFloat provides the corresponding package operation.
// Deprecated: Use HIncrByFloatContext.
func HIncrByFloat(name, key, field string, incr float64) (float64, error) {
	return HIncrByFloatContext(context.Background(), name, key, field, incr)
}

// HKeys provides the corresponding package operation.
// Deprecated: Use HKeysContext.
func HKeys(name, key string) ([]string, error) {
	return HKeysContext(context.Background(), name, key)
}

// HLen provides the corresponding package operation.
// Deprecated: Use HLenContext.
func HLen(name, key string) (int64, error) {
	return HLenContext(context.Background(), name, key)
}

// HVals provides the corresponding package operation.
// Deprecated: Use HValsContext.
func HVals(name, key string) ([]string, error) {
	return HValsContext(context.Background(), name, key)
}

// List operations.

// LPush provides the corresponding package operation.
// Deprecated: Use LPushContext.
func LPush(name, key string, values ...any) (int64, error) {
	return LPushContext(context.Background(), name, key, values...)
}

// RPush provides the corresponding package operation.
// Deprecated: Use RPushContext.
func RPush(name, key string, values ...any) (int64, error) {
	return RPushContext(context.Background(), name, key, values...)
}

// LPop provides the corresponding package operation.
// Deprecated: Use LPopContext.
func LPop(name, key string) (string, error) {
	return LPopContext(context.Background(), name, key)
}

// RPop provides the corresponding package operation.
// Deprecated: Use RPopContext.
func RPop(name, key string) (string, error) {
	return RPopContext(context.Background(), name, key)
}

// LRange provides the corresponding package operation.
// Deprecated: Use LRangeContext.
func LRange(name, key string, start, stop int64) ([]string, error) {
	return LRangeContext(context.Background(), name, key, start, stop)
}

// LLen provides the corresponding package operation.
// Deprecated: Use LLenContext.
func LLen(name, key string) (int64, error) {
	return LLenContext(context.Background(), name, key)
}

// LIndex provides the corresponding package operation.
// Deprecated: Use LIndexContext.
func LIndex(name, key string, index int64) (string, error) {
	return LIndexContext(context.Background(), name, key, index)
}

// LInsert provides the corresponding package operation.
// Deprecated: Use LInsertContext.
func LInsert(name, key, op string, pivot, value any) (int64, error) {
	return LInsertContext(context.Background(), name, key, op, pivot, value)
}

// LRem provides the corresponding package operation.
// Deprecated: Use LRemContext.
func LRem(name, key string, count int64, value any) (int64, error) {
	return LRemContext(context.Background(), name, key, count, value)
}

// LTrim provides the corresponding package operation.
// Deprecated: Use LTrimContext.
func LTrim(name, key string, start, stop int64) error {
	return LTrimContext(context.Background(), name, key, start, stop)
}

// LSet provides the corresponding package operation.
// Deprecated: Use LSetContext.
func LSet(name, key string, index int64, value any) error {
	return LSetContext(context.Background(), name, key, index, value)
}

// Set operations.

// SAdd provides the corresponding package operation.
// Deprecated: Use SAddContext.
func SAdd(name, key string, members ...any) (int64, error) {
	return SAddContext(context.Background(), name, key, members...)
}

// SRem provides the corresponding package operation.
// Deprecated: Use SRemContext.
func SRem(name, key string, members ...any) (int64, error) {
	return SRemContext(context.Background(), name, key, members...)
}

// SMembers provides the corresponding package operation.
// Deprecated: Use SMembersContext.
func SMembers(name, key string) ([]string, error) {
	return SMembersContext(context.Background(), name, key)
}

// SIsMember provides the corresponding package operation.
// Deprecated: Use SIsMemberContext.
func SIsMember(name, key string, member any) (bool, error) {
	return SIsMemberContext(context.Background(), name, key, member)
}

// SCard provides the corresponding package operation.
// Deprecated: Use SCardContext.
func SCard(name, key string) (int64, error) {
	return SCardContext(context.Background(), name, key)
}

// SPop provides the corresponding package operation.
// Deprecated: Use SPopContext.
func SPop(name, key string) (string, error) {
	return SPopContext(context.Background(), name, key)
}

// SUnion provides the corresponding package operation.
// Deprecated: Use SUnionContext.
func SUnion(name string, keys ...string) ([]string, error) {
	return SUnionContext(context.Background(), name, keys...)
}

// SInter provides the corresponding package operation.
// Deprecated: Use SInterContext.
func SInter(name string, keys ...string) ([]string, error) {
	return SInterContext(context.Background(), name, keys...)
}

// SDiff provides the corresponding package operation.
// Deprecated: Use SDiffContext.
func SDiff(name string, keys ...string) ([]string, error) {
	return SDiffContext(context.Background(), name, keys...)
}

// Sorted Set operations.

// ZAdd provides the corresponding package operation.
// Deprecated: Use ZAddContext.
func ZAdd(name, key string, members ...redis.Z) (int64, error) {
	return ZAddContext(context.Background(), name, key, members...)
}

// ZRem provides the corresponding package operation.
// Deprecated: Use ZRemContext.
func ZRem(name, key string, members ...any) (int64, error) {
	return ZRemContext(context.Background(), name, key, members...)
}

// ZRange provides the corresponding package operation.
// Deprecated: Use ZRangeContext.
func ZRange(name, key string, start, stop int64) ([]string, error) {
	return ZRangeContext(context.Background(), name, key, start, stop)
}

// ZRangeWithScores provides the corresponding package operation.
// Deprecated: Use ZRangeWithScoresContext.
func ZRangeWithScores(name, key string, start, stop int64) ([]redis.Z, error) {
	return ZRangeWithScoresContext(context.Background(), name, key, start, stop)
}

// ZRank provides the corresponding package operation.
// Deprecated: Use ZRankContext.
func ZRank(name, key, member string) (int64, error) {
	return ZRankContext(context.Background(), name, key, member)
}

// ZScore provides the corresponding package operation.
// Deprecated: Use ZScoreContext.
func ZScore(name, key, member string) (float64, error) {
	return ZScoreContext(context.Background(), name, key, member)
}

// ZIncrBy provides the corresponding package operation.
// Deprecated: Use ZIncrByContext.
func ZIncrBy(name, key, member string, increment float64) (float64, error) {
	return ZIncrByContext(context.Background(), name, key, member, increment)
}

// ZCard provides the corresponding package operation.
// Deprecated: Use ZCardContext.
func ZCard(name, key string) (int64, error) {
	return ZCardContext(context.Background(), name, key)
}

// ZCount provides the corresponding package operation.
// Deprecated: Use ZCountContext.
func ZCount(name, key, min, max string) (int64, error) {
	return ZCountContext(context.Background(), name, key, min, max)
}

// ZRemRangeByRank provides the corresponding package operation.
// Deprecated: Use ZRemRangeByRankContext.
func ZRemRangeByRank(name, key string, start, stop int64) (int64, error) {
	return ZRemRangeByRankContext(context.Background(), name, key, start, stop)
}

// ZRemRangeByScore provides the corresponding package operation.
// Deprecated: Use ZRemRangeByScoreContext.
func ZRemRangeByScore(name, key, min, max string) (int64, error) {
	return ZRemRangeByScoreContext(context.Background(), name, key, min, max)
}

// Key operations.

// Exists provides the corresponding package operation.
// Deprecated: Use ExistsContext.
func Exists(name string, keys ...string) (int64, error) {
	return ExistsContext(context.Background(), name, keys...)
}

// Del provides the corresponding package operation.
// Deprecated: Use DelContext.
func Del(name string, keys ...string) (int64, error) {
	return DelContext(context.Background(), name, keys...)
}

// Expire provides the corresponding package operation.
// Deprecated: Use ExpireContext.
func Expire(name, key string, expiration time.Duration) (bool, error) {
	return ExpireContext(context.Background(), name, key, expiration)
}

// ExpireAt provides the corresponding package operation.
// Deprecated: Use ExpireAtContext.
func ExpireAt(name, key string, tm time.Time) (bool, error) {
	return ExpireAtContext(context.Background(), name, key, tm)
}

// TTL provides the corresponding package operation.
// Deprecated: Use TTLContext.
func TTL(name, key string) (time.Duration, error) {
	return TTLContext(context.Background(), name, key)
}

// Persist provides the corresponding package operation.
// Deprecated: Use PersistContext.
func Persist(name, key string) (bool, error) {
	return PersistContext(context.Background(), name, key)
}

// Rename provides the corresponding package operation.
// Deprecated: Use RenameContext.
func Rename(name, key, newkey string) error {
	return RenameContext(context.Background(), name, key, newkey)
}

// RenameNX provides the corresponding package operation.
// Deprecated: Use RenameNXContext.
func RenameNX(name, key, newkey string) (bool, error) {
	return RenameNXContext(context.Background(), name, key, newkey)
}

// Type provides the corresponding package operation.
// Deprecated: Use TypeContext.
func Type(name, key string) (string, error) {
	return TypeContext(context.Background(), name, key)
}

// Keys provides the corresponding package operation.
// Deprecated: Use KeysContext.
func Keys(name, pattern string) ([]string, error) {
	return KeysContext(context.Background(), name, pattern)
}

// Scan provides the corresponding package operation.
// Deprecated: Use ScanContext.
func Scan(name string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return ScanContext(context.Background(), name, cursor, match, count)
}

// Pipeline provides the corresponding package operation.
// Deprecated: Use PipelineContext.
func Pipeline(name string, fn func(redis.Pipeliner) error) error {
	return PipelineContext(context.Background(), name, fn)
}

// TxPipeline provides the corresponding package operation.
// Deprecated: Use TxPipelineContext.
func TxPipeline(name string, fn func(redis.Pipeliner) error) error {
	return TxPipelineContext(context.Background(), name, fn)
}

// Eval provides the corresponding package operation.
// Deprecated: Use EvalContext.
func Eval(name string, script string, keys []string, args ...any) (any, error) {
	return EvalContext(context.Background(), name, script, keys, args...)
}

// EvalSha provides the corresponding package operation.
// Deprecated: Use EvalShaContext.
func EvalSha(name, sha1 string, keys []string, args ...any) (any, error) {
	return EvalShaContext(context.Background(), name, sha1, keys, args...)
}

// ScriptLoad provides the corresponding package operation.
// Deprecated: Use ScriptLoadContext.
func ScriptLoad(name, script string) (string, error) {
	return ScriptLoadContext(context.Background(), name, script)
}

// ScriptExists provides the corresponding package operation.
// Deprecated: Use ScriptExistsContext.
func ScriptExists(name string, hashes ...string) ([]bool, error) {
	return ScriptExistsContext(context.Background(), name, hashes...)
}

// HyperLogLog operations.

// PFAdd provides the corresponding package operation.
// Deprecated: Use PFAddContext.
func PFAdd(name, key string, elements ...any) (int64, error) {
	return PFAddContext(context.Background(), name, key, elements...)
}

// PFCount provides the corresponding package operation.
// Deprecated: Use PFCountContext.
func PFCount(name string, keys ...string) (int64, error) {
	return PFCountContext(context.Background(), name, keys...)
}

// PFMerge provides the corresponding package operation.
// Deprecated: Use PFMergeContext.
func PFMerge(name, dest string, keys ...string) error {
	return PFMergeContext(context.Background(), name, dest, keys...)
}

// Bitmap operations.

// SetBit provides the corresponding package operation.
// Deprecated: Use SetBitContext.
func SetBit(name, key string, offset int64, value int) (int64, error) {
	return SetBitContext(context.Background(), name, key, offset, value)
}

// GetBit provides the corresponding package operation.
// Deprecated: Use GetBitContext.
func GetBit(name, key string, offset int64) (int64, error) {
	return GetBitContext(context.Background(), name, key, offset)
}

// BitCount provides the corresponding package operation.
// Deprecated: Use BitCountContext.
func BitCount(name, key string) (int64, error) {
	return BitCountContext(context.Background(), name, key)
}
