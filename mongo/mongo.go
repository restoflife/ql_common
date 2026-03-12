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

	"github.com/restoflife/ql_common/logger"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

var (
	clientMap = make(map[string]*mongo.Client)
	mu        sync.RWMutex
)

// MustBootUpMongo 初始化多个 Mongo 客户端
func MustBootUpMongo(configs map[string]*Config, mongoLog *zap.Logger) error {
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

			if mongoLog != nil {
				mongoLogger := logger.NewMongoLogger(mongoLog)
				mongoLogger.ShowMongo(true)
				// 获取配置了日志的客户端选项
				clientOpts = mongoLogger.GetClientOptions(cfg.URI)
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

			client, err := mongo.Connect(clientOpts)
			if err != nil {
				return fmt.Errorf("mongo [%s] 连接失败：%w", name, err)
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

// InsertOne 插入单个文档
//  mongo.InsertOne("配置名称", "库名", "集合名", "文档")
func InsertOne(name, dbName, collName string, document any) (bson.ObjectID, error) {
	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return bson.NilObjectID, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, document)
	if err != nil {
		return bson.NilObjectID, fmt.Errorf("InsertOne fail：%w", err)
	}

	objId, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		return bson.NilObjectID, errors.New("无法获取插入的 ID")
	}

	return objId, nil
}

// InsertMany 批量插入多个文档
// mongo.InsertMany("配置名称", "库名", "集合名", "文档")
func InsertMany(name, dbName, collName string, documents []any) ([]any, error) {
	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertMany(ctx, documents)
	if err != nil {
		return nil, fmt.Errorf("批量插入失败：%w", err)
	}

	return result.InsertedIDs, nil
}

// FindOne 查询单个文档
// objectID, err := bson.ObjectIDFromHex("_id主键")
// mongo.FindOne("配置名称", "库名", "集合名", bson.M{"key": "value"})
func FindOne(name, dbName, collName string, filter any, opts ...options.Lister[options.FindOneOptions]) (*mongo.SingleResult, error) {
	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return collection.FindOne(ctx, filter, opts...), nil
}

// Find 查询多个文档
func Find(name, dbName, collName string, filter any, opts ...options.Lister[options.FindOptions]) (*mongo.Cursor, error) {
	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, fmt.Errorf("查询失败：%w", err)
	}

	return cursor, nil
}

// UpdateOne 更新单个文档
func UpdateOne(name, dbName, collName string, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		return nil, fmt.Errorf("更新失败：%w", err)
	}

	return result, nil
}

// UpdateMany 批量更新多个文档
func UpdateMany(name, dbName, collName string, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error) {
	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := collection.UpdateMany(ctx, filter, update, opts...)
	if err != nil {
		return nil, fmt.Errorf("批量更新失败：%w", err)
	}

	return result, nil
}

// UpdateByID 根据 ID 更新文档
func UpdateByID(name, dbName, collName string, id bson.ObjectID, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update, opts...)
	if err != nil {
		return nil, fmt.Errorf("根据 ID 更新失败：%w", err)
	}

	return result, nil
}

// DeleteOne 删除单个文档
func DeleteOne(name, dbName, collName string, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, filter, opts...)
	if err != nil {
		return nil, fmt.Errorf("删除失败：%w", err)
	}

	return result, nil
}

// DeleteMany 批量删除多个文档
func DeleteMany(name, dbName, collName string, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := collection.DeleteMany(ctx, filter, opts...)
	if err != nil {
		return nil, fmt.Errorf("批量删除失败：%w", err)
	}

	return result, nil
}

// CountDocuments 统计文档数量
func CountDocuments(name, dbName, collName string, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := collection.CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, fmt.Errorf("统计数量失败：%w", err)
	}

	return count, nil
}

// EstimatedDocumentCount 获取集合中文档的估计总数
func EstimatedDocumentCount(name, dbName, collName string, opts ...options.Lister[options.EstimatedDocumentCountOptions]) (int64, error) {
	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := collection.EstimatedDocumentCount(ctx, opts...)
	if err != nil {
		return 0, fmt.Errorf("获取估计数量失败：%w", err)
	}

	return count, nil
}

// Pagination 分页查询配置
func Pagination(page, pageSize, maxSize int64) options.Lister[options.FindOptions] {
	skip := (page - 1) * pageSize
	if skip < 0 {
		skip = 0
	}

	if pageSize <= 0 {
		pageSize = 10
	}
	// maxSize 用于限制单次查询的最大文档数，防止 pageSize 过大导致性能问题或内存溢出
	if pageSize > maxSize {
		pageSize = maxSize
	}

	return options.Find().SetSkip(skip).SetLimit(pageSize)
}

// Aggregate 聚合查询
func Aggregate(name, dbName, collName string, pipeline any, opts ...options.Lister[options.AggregateOptions]) (*mongo.Cursor, error) {
	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := collection.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return nil, fmt.Errorf("聚合查询失败：%w", err)
	}

	return cursor, nil
}

// Distinct 查询不同值
func Distinct(name, dbName, collName string, fieldName string, filter any, opts ...options.Lister[options.DistinctOptions]) (interface{}, error) {
	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result := collection.Distinct(ctx, fieldName, filter, opts...)

	return result, nil
}

// DropIndex 删除索引
func DropIndex(name, dbName, collName string, indexName string, opts ...options.Lister[options.DropIndexesOptions]) error {
	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = collection.Indexes().DropOne(ctx, indexName, opts...)
	if err != nil {
		return fmt.Errorf("删除索引失败：%w", err)
	}

	return nil
}

// CreateIndex 创建索引
func CreateIndex(name, dbName, collName string, indexes []mongo.IndexModel, opts ...options.Lister[options.CreateIndexesOptions]) ([]string, error) {
	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexNames, err := collection.Indexes().CreateMany(ctx, indexes, opts...)
	if err != nil {
		return nil, fmt.Errorf("创建索引失败：%w", err)
	}

	return indexNames, nil
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
