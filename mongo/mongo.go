package mongo

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"qingliu/logger"
)

var (
	clientMap = make(map[string]*mongo.Client)
	mu        sync.RWMutex
)

// MustBootUpMongo 初始化多个 Mongo 客户端
func MustBootUpMongo(configs map[string]*Config) error {
	for name, cfg := range configs {
		err := func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			clientOpts := options.Client().ApplyURI(cfg.URI)

			// 用户名/密码鉴权（如果未配置在 URI 中）
			if cfg.Username != "" && cfg.Password != "" {
				cred := options.Credential{
					Username:   cfg.Username,
					Password:   cfg.Password,
					AuthSource: cfg.AuthSource,
				}
				clientOpts.SetAuth(cred)
			}

			// TLS/CAFile
			if cfg.CACertFile != "" {
				tlsConfig, err := getTLSConfigFromCA(cfg.CACertFile)
				if err != nil {
					return fmt.Errorf("加载 CA 文件失败: %w", err)
				}
				clientOpts.SetTLSConfig(tlsConfig)
			}

			if cfg.MaxPoolSize > 0 {
				clientOpts.SetMaxPoolSize(cfg.MaxPoolSize)
			}
			if cfg.MinPoolSize > 0 {
				clientOpts.SetMinPoolSize(cfg.MinPoolSize)
			}

			client, err := mongo.Connect(ctx, clientOpts)
			if err != nil {
				return fmt.Errorf("mongo [%s] 连接失败: %w", name, err)
			}

			if err = client.Ping(ctx, nil); err != nil {
				return fmt.Errorf("mongo [%s] ping 失败: %w", name, err)
			}

			// 列出所有数据库
			// dbs, err := client.ListDatabaseNames(ctx, bson.M{})
			// if err != nil {
			// 	logger.Error("列出数据库失败", zap.String("name", name), zap.Error(err))
			// } else {
			// 	logger.Info("Mongo数据库列表", zap.String("name", name), zap.Strings("databases", dbs))
			// }

			mu.Lock()
			defer mu.Unlock()
			if _, ok := clientMap[name]; ok {
				return fmt.Errorf("mongo [%s] 已存在", name)
			}
			clientMap[name] = client

			logger.Info("Mongo连接成功", zap.String("name", name), zap.String("uri", cfg.URI))
			return nil
		}()
		if err != nil {
			return err
		}
	}

	// 启动定时 Ping 健康检查
	go func() {
		ticker := time.NewTicker(5 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			mu.RLock()
			for name, cli := range clientMap {
				if err := cli.Ping(context.Background(), nil); err != nil {
					logger.Error("Mongo健康检查失败", zap.String("name", name), zap.Error(err))
				}
			}
			mu.RUnlock()
		}
	}()

	return nil
}

func getTLSConfigFromCA(caFile string) (*tls.Config, error) {
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	caPool := x509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM(caCert); !ok {
		return nil, errors.New("无法解析 CA 证书")
	}
	return &tls.Config{
		RootCAs: caPool,
	}, nil
}

// GetClient 获取 Mongo 客户端
func GetClient(name string) (*mongo.Client, error) {
	mu.RLock()
	defer mu.RUnlock()

	client, ok := clientMap[name]
	if !ok {
		return nil, fmt.Errorf("mongo实例 [%s] 不存在", name)
	}
	return client, nil
}

// GetCollection 获取具体集合
func GetCollection(name, dbName, collName string) (*mongo.Collection, error) {
	client, err := GetClient(name)
	if err != nil {
		return nil, err
	}
	return client.Database(dbName).Collection(collName), nil
}

// ShutdownMongo 关闭所有 Mongo 实例连接
func ShutdownMongo() {
	mu.Lock()
	defer mu.Unlock()

	for name, client := range clientMap {
		if err := client.Disconnect(context.Background()); err != nil {
			logger.Error("Mongo 关闭失败", zap.String("name", name), zap.Error(err))
		}
	}
}
