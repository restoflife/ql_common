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
