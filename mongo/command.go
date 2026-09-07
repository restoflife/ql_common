package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"strings"
)

// GetDatabase returns the configured default database when dbName is empty.
// The returned handle is shared and must not be disconnected by a request.
func GetDatabase(name, dbName string) (*mongo.Database, error) {
	client, err := GetClient(name)
	if err != nil {
		return nil, err
	}
	if dbName == "" {
		value, ok := defaultDatabases.Load(client)
		if ok {
			dbName = value.(string)
		}
	}
	if strings.TrimSpace(dbName) == "" || strings.ContainsRune(dbName, '\x00') {
		return nil, ErrDatabaseRequired
	}
	return client.Database(dbName), nil
}

// RunCommandContext executes an ordered BSON command (use bson.D) on the selected database.
// Decode the returned result and check its error. Prefer typed CRUD helpers for normal queries.
func RunCommandContext(ctx context.Context, name, dbName string, command any, opts ...options.Lister[options.RunCmdOptions]) (*mongo.SingleResult, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	database, err := GetDatabase(name, dbName)
	if err != nil {
		return nil, err
	}
	return database.RunCommand(ctx, command, opts...), nil
}

// Deprecated: Use RunCommandContext.
func RunCommand(name, dbName string, command any, opts ...options.Lister[options.RunCmdOptions]) (*mongo.SingleResult, error) {
	return RunCommandContext(context.Background(), name, dbName, command, opts...)
}

// InsertOneResultContext supports arbitrary _id types (including strings).
// InsertOneContext remains compatible with existing callers that require ObjectID.
func InsertOneResultContext(ctx context.Context, name, dbName, collName string, document any, opts ...options.Lister[options.InsertOneOptions]) (*mongo.InsertOneResult, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}
	return collection.InsertOne(ctx, document, opts...)
}
