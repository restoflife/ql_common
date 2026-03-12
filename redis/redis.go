package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/restoflife/ql_common/logger"
	"go.uber.org/zap"
)

// redisMgr 保存所有命名的 Redis 客户端实例
var redisMgr = map[string]redis.UniversalClient{}

// Nil Redis 空值错误
var Nil = redis.Nil

// MustBootUpRedis 初始化并连接多个 Redis 实例，根据配置支持单机、哨兵、集群模式
func MustBootUpRedis(configs map[string]*Config) error {
	for name, c := range configs {
		var client redis.UniversalClient

		switch c.Mode {
		case SENTINEL:
			// 使用 Redis Sentinel 模式
			client = redis.NewFailoverClient(&redis.FailoverOptions{
				MasterName:    c.MasterName,
				SentinelAddrs: c.Slaves,
				Password:      c.Password,
				DB:            c.DB,
				PoolSize:      c.PoolSize,
				MinIdleConns:  c.MinIdle,
			})
		case CLUSTER:
			// 使用 Redis Cluster 模式
			client = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:        c.Slaves,
				Password:     c.Password,
				PoolSize:     c.PoolSize,
				MinIdleConns: c.MinIdle,
			})
		default:
			// 默认使用 Standalone 模式
			client = redis.NewClient(&redis.Options{
				Addr:         c.Addr,
				Password:     c.Password,
				DB:           c.DB,
				PoolSize:     c.PoolSize,
				MinIdleConns: c.MinIdle,
			})
		}

		// 测试 Redis 连接是否可用
		if err := client.Ping(context.Background()).Err(); err != nil {
			return fmt.Errorf("redis [%s] 连接失败: %w", name, err)
		}

		// 检查是否重复加载 Redis 实例
		if _, ok := redisMgr[name]; ok {
			return fmt.Errorf("redis实例 [%s] 重复加载", name)
		}

		// 将 Redis 实例加入管理器
		redisMgr[name] = client
		logger.Info("Redis连接成功", zap.String("name", name), zap.String("mode", c.Mode))
	}

	// 启动 Redis 健康检查协程，定期 Ping 所有实例
	go func() {
		ticker := time.NewTicker(5 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			for name, cli := range redisMgr {
				if err := cli.Ping(context.Background()).Err(); err != nil {
					logger.Error("Redis连接检测失败", zap.String("name", name), zap.Error(err))
				}
			}
		}
	}()

	return nil
}

// GetRedis 根据名称获取 Redis 实例
func GetRedis(name string) (redis.UniversalClient, error) {
	client, ok := redisMgr[name]
	if !ok {
		return nil, fmt.Errorf("redis实例 [%s] 不存在", name)
	}
	return client, nil
}

// ShutdownRedis 优雅关闭所有 Redis 客户端连接
func ShutdownRedis() {
	for name, client := range redisMgr {
		if err := client.Close(); err != nil {
			fmt.Printf("Redis关闭失败 [%s]: %v\n", name, err)
		}
	}
}

// ==================== String 操作 ====================

// Set 设置字符串值
func Set(name, key string, value any, expiration time.Duration) error {
	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	return client.Set(context.Background(), key, value, expiration).Err()
}

// SetNX 仅当键不存在时设置值
func SetNX(name, key string, value any, expiration time.Duration) (bool, error) {
	client, err := GetRedis(name)
	if err != nil {
		return false, err
	}
	return client.SetNX(context.Background(), key, value, expiration).Result()
}

// SetXX 仅当键存在时设置值
func SetXX(name, key string, value any, expiration time.Duration) (bool, error) {
	client, err := GetRedis(name)
	if err != nil {
		return false, err
	}
	return client.SetXX(context.Background(), key, value, expiration).Result()
}

// Get 获取字符串值
func Get(name, key string) (string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.Get(context.Background(), key).Result()
}

// GetDel 获取并删除字符串值
func GetDel(name, key string) (string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.GetDel(context.Background(), key).Result()
}

// GetEx 获取值并设置过期时间
func GetEx(name, key string, expiration time.Duration) (string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.GetEx(context.Background(), key, expiration).Result()
}

// Incr 递增整数
func Incr(name, key string) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.Incr(context.Background(), key).Result()
}

// IncrBy 递增指定值
func IncrBy(name, key string, value int64) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.IncrBy(context.Background(), key, value).Result()
}

// IncrByFloat 递增浮点数
func IncrByFloat(name, key string, value float64) (float64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.IncrByFloat(context.Background(), key, value).Result()
}

// Decr 递减整数
func Decr(name, key string) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.Decr(context.Background(), key).Result()
}

// DecrBy 递减指定值
func DecrBy(name, key string, value int64) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.DecrBy(context.Background(), key, value).Result()
}

// Append 追加字符串
func Append(name, key, value string) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.Append(context.Background(), key, value).Result()
}

// MGet 批量获取值
func MGet(name string, keys ...string) ([]any, error) {
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.MGet(context.Background(), keys...).Result()
}

// MSet 批量设置值
func MSet(name string, values ...any) error {
	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	return client.MSet(context.Background(), values...).Err()
}

// GetRange 获取子字符串
func GetRange(name, key string, start, end int64) (string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.GetRange(context.Background(), key, start, end).Result()
}

// SetRange 替换子字符串
func SetRange(name, key string, offset int64, value string) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.SetRange(context.Background(), key, offset, value).Result()
}

// StrLen 获取字符串长度
func StrLen(name, key string) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.StrLen(context.Background(), key).Result()
}

// ==================== Hash 操作 ====================

// HSet 设置哈希字段值
func HSet(name, key, field string, value any) error {
	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	return client.HSet(context.Background(), key, field, value).Err()
}

// HSetNX 仅当字段不存在时设置
func HSetNX(name, key, field string, value any) (bool, error) {
	client, err := GetRedis(name)
	if err != nil {
		return false, err
	}
	return client.HSetNX(context.Background(), key, field, value).Result()
}

// HGet 获取哈希字段值
func HGet(name, key, field string) (string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.HGet(context.Background(), key, field).Result()
}

// HGetAll 获取所有哈希字段和值
func HGetAll(name, key string) (map[string]string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.HGetAll(context.Background(), key).Result()
}

// HMSet 批量设置哈希字段值
func HMSet(name, key string, values any) error {
	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	return client.HMSet(context.Background(), key, values).Err()
}

// HMGet 批量获取哈希字段值
func HMGet(name, key string, fields ...string) ([]any, error) {
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.HMGet(context.Background(), key, fields...).Result()
}

// HDel 删除哈希字段
func HDel(name, key string, fields ...string) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.HDel(context.Background(), key, fields...).Result()
}

// HExists 检查哈希字段是否存在
func HExists(name, key, field string) (bool, error) {
	client, err := GetRedis(name)
	if err != nil {
		return false, err
	}
	return client.HExists(context.Background(), key, field).Result()
}

// HIncrBy 哈希字段值递增
func HIncrBy(name, key, field string, incr int64) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.HIncrBy(context.Background(), key, field, incr).Result()
}

// HIncrByFloat 哈希字段值浮点数递增
func HIncrByFloat(name, key, field string, incr float64) (float64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.HIncrByFloat(context.Background(), key, field, incr).Result()
}

// HKeys 获取所有哈希字段
func HKeys(name, key string) ([]string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.HKeys(context.Background(), key).Result()
}

// HLen 获取哈希字段数量
func HLen(name, key string) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.HLen(context.Background(), key).Result()
}

// HVals 获取所有哈希值
func HVals(name, key string) ([]string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.HVals(context.Background(), key).Result()
}

// ==================== List 操作 ====================

// LPush 将元素推入列表左侧
func LPush(name, key string, values ...any) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.LPush(context.Background(), key, values...).Result()
}

// RPush 将元素推入列表右侧
func RPush(name, key string, values ...any) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.RPush(context.Background(), key, values...).Result()
}

// LPop 弹出列表左侧元素
func LPop(name, key string) (string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.LPop(context.Background(), key).Result()
}

// RPop 弹出列表右侧元素
func RPop(name, key string) (string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.RPop(context.Background(), key).Result()
}

// LRange 获取列表范围内元素
func LRange(name, key string, start, stop int64) ([]string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.LRange(context.Background(), key, start, stop).Result()
}

// LLen 获取列表长度
func LLen(name, key string) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.LLen(context.Background(), key).Result()
}

// LIndex 获取列表指定索引的元素
func LIndex(name, key string, index int64) (string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.LIndex(context.Background(), key, index).Result()
}

// LInsert 在列表指定位置插入元素
func LInsert(name, key, op string, pivot, value any) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.LInsert(context.Background(), key, op, pivot, value).Result()
}

// LRem 从列表中移除元素
func LRem(name, key string, count int64, value any) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.LRem(context.Background(), key, count, value).Result()
}

// LTrim 裁剪列表
func LTrim(name, key string, start, stop int64) error {
	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	return client.LTrim(context.Background(), key, start, stop).Err()
}

// LSet 设置列表指定索引的值
func LSet(name, key string, index int64, value any) error {
	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	return client.LSet(context.Background(), key, index, value).Err()
}

// ==================== Set 操作 ====================

// SAdd 向集合添加成员
func SAdd(name, key string, members ...any) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.SAdd(context.Background(), key, members...).Result()
}

// SRem 从集合移除成员
func SRem(name, key string, members ...any) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.SRem(context.Background(), key, members...).Result()
}

// SMembers 获取集合所有成员
func SMembers(name, key string) ([]string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.SMembers(context.Background(), key).Result()
}

// SIsMember 检查成员是否在集合中
func SIsMember(name, key string, member any) (bool, error) {
	client, err := GetRedis(name)
	if err != nil {
		return false, err
	}
	return client.SIsMember(context.Background(), key, member).Result()
}

// SCard 获取集合基数
func SCard(name, key string) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.SCard(context.Background(), key).Result()
}

// SPop 随机弹出集合成员
func SPop(name, key string) (string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.SPop(context.Background(), key).Result()
}

// SUnion 返回多个集合的并集
func SUnion(name string, keys ...string) ([]string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.SUnion(context.Background(), keys...).Result()
}

// SInter 返回多个集合的交集
func SInter(name string, keys ...string) ([]string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.SInter(context.Background(), keys...).Result()
}

// SDiff 返回多个集合的差集
func SDiff(name string, keys ...string) ([]string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.SDiff(context.Background(), keys...).Result()
}

// ==================== Sorted Set 操作 ====================

// ZAdd 向有序集合添加成员
func ZAdd(name, key string, members ...redis.Z) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.ZAdd(context.Background(), key, members...).Result()
}

// ZRem 从有序集合移除成员
func ZRem(name, key string, members ...any) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.ZRem(context.Background(), key, members...).Result()
}

// ZRange 获取有序集合指定范围的成员
func ZRange(name, key string, start, stop int64) ([]string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.ZRange(context.Background(), key, start, stop).Result()
}

// ZRangeWithScores 获取有序集合指定范围的成员及分数
func ZRangeWithScores(name, key string, start, stop int64) ([]redis.Z, error) {
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.ZRangeWithScores(context.Background(), key, start, stop).Result()
}

// ZRank 获取成员在有序集合中的排名
func ZRank(name, key, member string) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.ZRank(context.Background(), key, member).Result()
}

// ZScore 获取成员的分数
func ZScore(name, key, member string) (float64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.ZScore(context.Background(), key, member).Result()
}

// ZIncrBy 递增成员的分数
func ZIncrBy(name, key, member string, increment float64) (float64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.ZIncrBy(context.Background(), key, increment, member).Result()
}

// ZCard 获取有序集合基数
func ZCard(name, key string) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.ZCard(context.Background(), key).Result()
}

// ZCount 获取指定分数范围内的成员数量
func ZCount(name, key, min, max string) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.ZCount(context.Background(), key, min, max).Result()
}

// ZRemRangeByRank 按排名范围移除成员
func ZRemRangeByRank(name, key string, start, stop int64) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.ZRemRangeByRank(context.Background(), key, start, stop).Result()
}

// ZRemRangeByScore 按分数范围移除成员
func ZRemRangeByScore(name, key, min, max string) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.ZRemRangeByScore(context.Background(), key, min, max).Result()
}

// ==================== Key 操作 ====================

// Exists 检查键是否存在
func Exists(name string, keys ...string) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.Exists(context.Background(), keys...).Result()
}

// Del 删除键
func Del(name string, keys ...string) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.Del(context.Background(), keys...).Result()
}

// Expire 设置键的过期时间
func Expire(name, key string, expiration time.Duration) (bool, error) {
	client, err := GetRedis(name)
	if err != nil {
		return false, err
	}
	return client.Expire(context.Background(), key, expiration).Result()
}

// ExpireAt 设置键的过期时间点
func ExpireAt(name, key string, tm time.Time) (bool, error) {
	client, err := GetRedis(name)
	if err != nil {
		return false, err
	}
	return client.ExpireAt(context.Background(), key, tm).Result()
}

// TTL 获取键的剩余过期时间
func TTL(name, key string) (time.Duration, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.TTL(context.Background(), key).Result()
}

// Persist 移除键的过期时间
func Persist(name, key string) (bool, error) {
	client, err := GetRedis(name)
	if err != nil {
		return false, err
	}
	return client.Persist(context.Background(), key).Result()
}

// Rename 重命名键
func Rename(name, key, newkey string) error {
	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	return client.Rename(context.Background(), key, newkey).Err()
}

// RenameNX 仅当新键不存在时重命名
func RenameNX(name, key, newkey string) (bool, error) {
	client, err := GetRedis(name)
	if err != nil {
		return false, err
	}
	return client.RenameNX(context.Background(), key, newkey).Result()
}

// Type 获取键的类型
func Type(name, key string) (string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.Type(context.Background(), key).Result()
}

// Keys 查找匹配的键
func Keys(name, pattern string) ([]string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.Keys(context.Background(), pattern).Result()
}

// Scan 迭代键
func Scan(name string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return nil, 0, err
	}
	return client.Scan(context.Background(), cursor, match, count).Result()
}

// ==================== 批量操作 ====================

// Pipeline 执行批量命令
func Pipeline(name string, fn func(redis.Pipeliner) error) error {
	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	_, err = client.Pipelined(context.Background(), fn)
	return err
}

// TxPipeline 执行事务批量命令
func TxPipeline(name string, fn func(redis.Pipeliner) error) error {
	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	_, err = client.TxPipelined(context.Background(), fn)
	return err
}

// ==================== 脚本操作 ====================

// Eval 执行 Lua 脚本
func Eval(name string, script string, keys []string, args ...any) (any, error) {
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.Eval(context.Background(), script, keys, args...).Result()
}

// EvalSha 执行 Lua 脚本（使用 SHA）
func EvalSha(name, sha1 string, keys []string, args ...any) (any, error) {
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.EvalSha(context.Background(), sha1, keys, args...).Result()
}

// ScriptLoad 加载 Lua 脚本
func ScriptLoad(name, script string) (string, error) {
	client, err := GetRedis(name)
	if err != nil {
		return "", err
	}
	return client.ScriptLoad(context.Background(), script).Result()
}

// ScriptExists 检查脚本是否已加载
func ScriptExists(name string, hashes ...string) ([]bool, error) {
	client, err := GetRedis(name)
	if err != nil {
		return nil, err
	}
	return client.ScriptExists(context.Background(), hashes...).Result()
}

// ==================== HyperLogLog 操作 ====================

// PFAdd 添加元素到 HyperLogLog
func PFAdd(name, key string, elements ...any) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.PFAdd(context.Background(), key, elements...).Result()
}

// PFCount 获取 HyperLogLog 的基数估计
func PFCount(name string, keys ...string) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.PFCount(context.Background(), keys...).Result()
}

// PFMerge 合并多个 HyperLogLog
func PFMerge(name, dest string, keys ...string) error {
	client, err := GetRedis(name)
	if err != nil {
		return err
	}
	return client.PFMerge(context.Background(), dest, keys...).Err()
}

// ==================== Bitmap 操作 ====================

// SetBit 设置位
func SetBit(name, key string, offset int64, value int) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.SetBit(context.Background(), key, offset, value).Result()
}

// GetBit 获取位
func GetBit(name, key string, offset int64) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.GetBit(context.Background(), key, offset).Result()
}

// BitCount 获取位图中设置为 1 的数量
func BitCount(name, key string) (int64, error) {
	client, err := GetRedis(name)
	if err != nil {
		return 0, err
	}
	return client.BitCount(context.Background(), key, nil).Result()
}
