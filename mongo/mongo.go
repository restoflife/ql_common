package mongo

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/restoflife/ql_common/internal/registry"
	"github.com/restoflife/ql_common/logger"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

var clientMap registry.Registry[*mongo.Client]
var defaultDatabases sync.Map // client pointer -> configured business database

// Deprecated: Use BootUpMongoContext.
func MustBootUpMongo(configs map[string]*Config, mongoLog *zap.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return BootUpMongoContext(ctx, configs, mongoLog)
}

// BootUpMongoContext publishes a complete batch or cleans up all new clients.
func BootUpMongoContext(ctx context.Context, configs map[string]*Config, mongoLog *zap.Logger) error {
	if ctx == nil {
		return ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	names := make([]string, 0, len(configs))
	opts := make(map[string]*options.ClientOptions, len(configs))
	for name, cfg := range configs {
		option, err := clientOptions(cfg, mongoLog)
		if err != nil {
			return configError(name, err)
		}
		opts[name] = option
		names = append(names, name)
	}
	return clientMap.Init(names, func(name string) (*mongo.Client, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		client, err := mongo.Connect(opts[name])
		if err != nil {
			return nil, err
		}
		if err := client.Ping(ctx, nil); err != nil {
			return nil, errors.Join(err, disconnect(client))
		}
		defaultDatabases.Store(client, configs[name].Database)
		return client, nil
	}, disconnect)

}

func clientOptions(cfg *Config, mongoLog *zap.Logger) (*options.ClientOptions, error) {
	if cfg == nil || cfg.URI == "" {
		return nil, ErrURIRequired
	}
	if cfg.MaxPoolSize > 0 && cfg.MinPoolSize > cfg.MaxPoolSize {
		return nil, ErrPoolSize
	}
	opts := options.Client().ApplyURI(cfg.URI)
	if mongoLog != nil {
		opts = logger.NewMongoLogger(mongoLog).GetClientOptions(cfg.URI)
	}
	// Apply explicit credentials last, preserving logging and URI options.
	if cfg.Username != "" {
		opts.SetAuth(options.Credential{Username: cfg.Username, Password: cfg.Password, AuthSource: cfg.AuthSource})
	} else if cfg.Password != "" {
		return nil, ErrUsernameRequired
	}
	if cfg.CACertFile != "" {
		tlsConfig, err := getTLSConfigFromCA(cfg.CACertFile)
		if err != nil {
			return nil, err
		}
		opts.SetTLSConfig(tlsConfig)
	}
	if cfg.MaxPoolSize > 0 {
		opts.SetMaxPoolSize(cfg.MaxPoolSize)
	}
	if cfg.MinPoolSize > 0 {
		opts.SetMinPoolSize(cfg.MinPoolSize)
	}
	if cfg.Timeout < 0 {
		return nil, ErrNegativeTimeout
	}
	if cfg.Timeout > 0 {
		opts.SetTimeout(cfg.Timeout)
	}
	if cfg.MaxConnecting > 0 {
		opts.SetMaxConnecting(cfg.MaxConnecting)
	}
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	return opts, nil
}

func disconnect(client *mongo.Client) error {
	defaultDatabases.Delete(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return client.Disconnect(ctx)
}

func getTLSConfigFromCA(caFile string) (*tls.Config, error) {
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	caPool := x509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM(caCert); !ok {
		return nil, ErrInvalidCACertificate
	}
	return &tls.Config{
		RootCAs:    caPool,
		MinVersion: tls.VersionTLS12,
	}, nil
}

// InsertOne provides the corresponding package operation.
//
// mongo provides the corresponding package operation.
// Deprecated: Use InsertOneContext.
func InsertOne(name, dbName, collName string, document any) (bson.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return InsertOneContext(ctx, name, dbName, collName, document)
}

// InsertMany provides the corresponding package operation.
// mongo provides the corresponding package operation.
// Deprecated: Use InsertManyContext.
func InsertMany(name, dbName, collName string, documents []any) ([]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return InsertManyContext(ctx, name, dbName, collName, documents)
}

// FindOne provides the corresponding package operation.
// objectID provides the corresponding package operation.
// mongo provides the corresponding package operation.
// Deprecated: Use FindOneContext.
func FindOne(name, dbName, collName string, filter any, opts ...options.Lister[options.FindOneOptions]) (*mongo.SingleResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return FindOneContext(ctx, name, dbName, collName, filter, opts...)
}

// Find provides the corresponding package operation.
// Deprecated: Use FindContext.
func Find(name, dbName, collName string, filter any, opts ...options.Lister[options.FindOptions]) (*mongo.Cursor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return FindContext(ctx, name, dbName, collName, filter, opts...)
}

// UpdateOne provides the corresponding package operation.
// Deprecated: Use UpdateOneContext.
func UpdateOne(name, dbName, collName string, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return UpdateOneContext(ctx, name, dbName, collName, filter, update, opts...)
}

// UpdateMany provides the corresponding package operation.
// Deprecated: Use UpdateManyContext.
func UpdateMany(name, dbName, collName string, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return UpdateManyContext(ctx, name, dbName, collName, filter, update, opts...)
}

// UpdateByID provides the corresponding package operation.
// Deprecated: Use UpdateByIDContext.
func UpdateByID(name, dbName, collName string, id bson.ObjectID, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return UpdateByIDContext(ctx, name, dbName, collName, id, update, opts...)
}

// DeleteOne provides the corresponding package operation.
// Deprecated: Use DeleteOneContext.
func DeleteOne(name, dbName, collName string, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return DeleteOneContext(ctx, name, dbName, collName, filter, opts...)
}

// DeleteMany provides the corresponding package operation.
// Deprecated: Use DeleteManyContext.
func DeleteMany(name, dbName, collName string, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return DeleteManyContext(ctx, name, dbName, collName, filter, opts...)
}

// CountDocuments provides the corresponding package operation.
// Deprecated: Use CountDocumentsContext.
func CountDocuments(name, dbName, collName string, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return CountDocumentsContext(ctx, name, dbName, collName, filter, opts...)
}

// EstimatedDocumentCount provides the corresponding package operation.
// Deprecated: Use EstimatedDocumentCountContext.
func EstimatedDocumentCount(name, dbName, collName string, opts ...options.Lister[options.EstimatedDocumentCountOptions]) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return EstimatedDocumentCountContext(ctx, name, dbName, collName, opts...)
}

// Pagination provides the corresponding package operation.
func Pagination(page, pageSize, maxSize int64) options.Lister[options.FindOptions] {
	if page < 1 {
		page = 1
	}
	if maxSize <= 0 {
		maxSize = 100
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > maxSize {
		pageSize = maxSize
	}
	skip := int64(math.MaxInt64)
	if page-1 <= math.MaxInt64/pageSize {
		skip = (page - 1) * pageSize
	}
	return options.Find().SetSkip(skip).SetLimit(pageSize)
}

// Aggregate provides the corresponding package operation.
// Deprecated: Use AggregateContext.
func Aggregate(name, dbName, collName string, pipeline any, opts ...options.Lister[options.AggregateOptions]) (*mongo.Cursor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return AggregateContext(ctx, name, dbName, collName, pipeline, opts...)
}

// Distinct provides the corresponding package operation.
// Deprecated: Use DistinctContext.
func Distinct(name, dbName, collName string, fieldName string, filter any, opts ...options.Lister[options.DistinctOptions]) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return DistinctContext(ctx, name, dbName, collName, fieldName, filter, opts...)
}

// DropIndex provides the corresponding package operation.
// Deprecated: Use DropIndexContext.
func DropIndex(name, dbName, collName string, indexName string, opts ...options.Lister[options.DropIndexesOptions]) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return DropIndexContext(ctx, name, dbName, collName, indexName, opts...)
}

// CreateIndex provides the corresponding package operation.
// Deprecated: Use CreateIndexContext.
func CreateIndex(name, dbName, collName string, indexes []mongo.IndexModel, opts ...options.Lister[options.CreateIndexesOptions]) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return CreateIndexContext(ctx, name, dbName, collName, indexes, opts...)
}

// GetClient provides the corresponding package operation.
func GetClient(name string) (*mongo.Client, error) { return clientMap.Get(name) }

// GetCollection provides the corresponding package operation.
func GetCollection(name, dbName, collName string) (*mongo.Collection, error) {
	database, err := GetDatabase(name, dbName)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(collName) == "" || strings.ContainsRune(collName, '\x00') {
		return nil, ErrCollectionRequired
	}
	return database.Collection(collName), nil
}

// ShutdownMongo provides the corresponding package operation.
// Deprecated: Use ShutdownMongoE.
func ShutdownMongo() {
	if err := ShutdownMongoE(); err != nil {
		logger.Error("Mongo shutdown failed", zap.Error(err))
	}
}

// ShutdownMongoE unregisters and closes clients. Drain in-flight users first.
func ShutdownMongoE() error { return clientMap.Close(disconnect) }
