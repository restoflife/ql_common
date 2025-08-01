package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func Transaction(ctx context.Context, name string, fn func(redis.Pipeliner) error) error {
	client, err := GetRedis(name)
	if err != nil {
		return err
	}

	// 使用 TxPipelined 执行事务块（自动 MULTI/EXEC）
	_, err = client.TxPipelined(ctx, fn)
	return err
}
