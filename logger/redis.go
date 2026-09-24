package logger

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	redisKeyLimit             = 256
	redisPipelineSummaryLimit = 32
)

// RedisHook logs Redis commands without recording values or complete argument lists.
// Successful commands are logged at Info level and failures at Error level.
type RedisHook struct {
	logger *zap.Logger
	plain  *zap.Logger
}

type redisCommandSummary struct {
	Command string `json:"command"`
	Key     string `json:"key,omitempty"`
}

// NewRedisHook constructs a safe logging hook for go-redis clients.
func NewRedisHook(log *zap.Logger) *RedisHook {
	if log == nil {
		log = zap.NewNop()
	}
	return &RedisHook{
		logger: log,
		plain:  log.WithOptions(zap.WithCaller(false)),
	}
}

// DialHook leaves connection establishment unchanged. Command results are handled by
// ProcessHook and ProcessPipelineHook.
func (h *RedisHook) DialHook(next redis.DialHook) redis.DialHook { return next }

// ProcessHook logs one Redis command after it completes.
func (h *RedisHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		started := time.Now()
		err := next(ctx, cmd)
		if redisInternalCommand(cmd) {
			return err
		}
		fields := redisCommandFields(cmd, time.Since(started))
		if err != nil {
			fields = append(fields, zap.Error(err))
			h.logger.Error(REDIS, fields...)
			return err
		}
		h.plain.Info(REDIS, fields...)
		return nil
	}
}

// ProcessPipelineHook logs a bounded, value-free summary of a Redis pipeline.
func (h *RedisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		started := time.Now()
		err := next(ctx, cmds)
		visible, visibleErr := redisVisiblePipeline(cmds, err)
		if len(visible) == 0 {
			return err
		}
		fields := []zap.Field{
			zap.Int("command_count", len(visible)),
			zap.Any("commands", redisPipelineSummary(visible)),
			zap.String("latency", time.Since(started).String()),
		}
		if visibleErr != nil {
			fields = append(fields, zap.Error(visibleErr))
			h.logger.Error(REDIS, fields...)
			return err
		}
		h.plain.Info(REDIS, fields...)
		return err
	}
}

// HELLO and CLIENT are used by go-redis for protocol negotiation and capability
// detection. Older servers may reject them while the driver safely falls back, so
// they are not business command failures and are deliberately omitted from logs.
func redisInternalCommand(cmd redis.Cmder) bool {
	if cmd == nil {
		return false
	}
	switch strings.ToLower(cmd.Name()) {
	case "hello", "client":
		return true
	default:
		return false
	}
}

func redisVisiblePipeline(cmds []redis.Cmder, fallback error) ([]redis.Cmder, error) {
	visible := make([]redis.Cmder, 0, len(cmds))
	internalFailed := false
	var visibleErr error
	for _, cmd := range cmds {
		if redisInternalCommand(cmd) {
			internalFailed = internalFailed || cmd.Err() != nil
			continue
		}
		visible = append(visible, cmd)
		if visibleErr == nil {
			visibleErr = cmd.Err()
		}
	}
	if visibleErr == nil && fallback != nil && !internalFailed {
		visibleErr = fallback
	}
	return visible, visibleErr
}

func redisCommandFields(cmd redis.Cmder, latency time.Duration) []zap.Field {
	fields := []zap.Field{
		zap.String("command", redisCommandName(cmd)),
		zap.String("latency", latency.String()),
	}
	if key, ok := redisCommandKey(cmd); ok {
		fields = append(fields, zap.String("key", key))
	}
	return fields
}

func redisPipelineSummary(cmds []redis.Cmder) []redisCommandSummary {
	limit := len(cmds)
	if limit > redisPipelineSummaryLimit {
		limit = redisPipelineSummaryLimit
	}
	result := make([]redisCommandSummary, 0, limit)
	for _, cmd := range cmds[:limit] {
		summary := redisCommandSummary{Command: redisCommandName(cmd)}
		if key, ok := redisCommandKey(cmd); ok {
			summary.Key = key
		}
		result = append(result, summary)
	}
	return result
}

func redisCommandName(cmd redis.Cmder) string {
	if cmd == nil {
		return "unknown"
	}
	return strings.ToUpper(cmd.Name())
}

// Only commands whose second argument is unambiguously a key are listed here.
// Unknown commands intentionally omit the key instead of risking disclosure of a
// password, token, script, value, channel payload, or other command argument.
var redisFirstKeyCommands = map[string]struct{}{
	"append": {}, "bitcount": {}, "bitfield": {}, "decr": {}, "decrby": {},
	"del": {}, "dump": {}, "exists": {}, "expire": {}, "expireat": {},
	"expiretime": {}, "geoadd": {}, "geodist": {}, "geohash": {}, "geopos": {},
	"georadius": {}, "georadius_ro": {}, "geosearch": {}, "geosearchstore": {},
	"get": {}, "getbit": {}, "getdel": {}, "getex": {}, "getrange": {},
	"getset": {}, "hdel": {}, "hexists": {}, "hget": {}, "hgetall": {},
	"hincrby": {}, "hincrbyfloat": {}, "hkeys": {}, "hlen": {}, "hmget": {},
	"hmset": {}, "hrandfield": {}, "hscan": {}, "hset": {}, "hsetnx": {},
	"hstrlen": {}, "hvals": {}, "incr": {}, "incrby": {}, "incrbyfloat": {},
	"lindex": {}, "linsert": {}, "llen": {}, "lmove": {}, "lpop": {},
	"lpos": {}, "lpush": {}, "lpushx": {}, "lrange": {}, "lrem": {},
	"lset": {}, "ltrim": {}, "mget": {}, "mset": {}, "msetnx": {},
	"persist": {}, "pexpire": {}, "pexpireat": {}, "pexpiretime": {},
	"pfadd": {}, "pfcount": {}, "psetex": {}, "pttl": {}, "rename": {},
	"renamenx": {}, "restore": {}, "rpop": {}, "rpoplpush": {}, "rpush": {},
	"rpushx": {}, "sadd": {}, "scard": {}, "sdiff": {}, "sdiffstore": {},
	"sinter": {}, "sinterstore": {}, "sismember": {},
	"smembers": {}, "smismember": {}, "smove": {}, "spop": {}, "srandmember": {},
	"srem": {}, "sscan": {}, "strlen": {}, "sunion": {}, "sunionstore": {},
	"set": {}, "setbit": {}, "setex": {}, "setnx": {}, "setrange": {},
	"touch": {}, "ttl": {}, "type": {}, "unlink": {}, "xack": {},
	"xadd": {}, "xautoclaim": {}, "xclaim": {}, "xdel": {},
	"xlen": {}, "xpending": {}, "xrange": {},
	"xrevrange": {}, "xtrim": {}, "zadd": {}, "zcard": {}, "zcount": {},
	"zincrby": {}, "zlexcount": {}, "zmscore": {}, "zpopmax": {},
	"zpopmin": {}, "zrandmember": {}, "zrange": {}, "zrangebylex": {},
	"zrangebyscore": {}, "zrangestore": {}, "zrank": {}, "zrem": {},
	"zremrangebylex": {}, "zremrangebyrank": {}, "zremrangebyscore": {},
	"zrevrange": {}, "zrevrangebylex": {}, "zrevrangebyscore": {},
	"zrevrank": {}, "zscan": {}, "zscore": {},
}

func redisCommandKey(cmd redis.Cmder) (string, bool) {
	if cmd == nil {
		return "", false
	}
	if _, ok := redisFirstKeyCommands[strings.ToLower(cmd.Name())]; !ok {
		return "", false
	}
	args := cmd.Args()
	if len(args) < 2 {
		return "", false
	}
	key, ok := redisKeyString(args[1])
	if !ok {
		return "", false
	}
	if len(key) > redisKeyLimit {
		key = key[:redisKeyLimit] + "..."
	}
	return key, true
}

func redisKeyString(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case []byte:
		return string(v), true
	case int:
		return strconv.Itoa(v), true
	case int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprint(v), true
	default:
		return "", false
	}
}
