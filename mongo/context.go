package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// InsertOneContext is InsertOne with caller-owned cancellation and deadline.
func InsertOneContext(ctx context.Context, name, dbName, collName string, document any) (bson.ObjectID, error) {
	if ctx == nil {
		err := ErrNilContext
		return bson.NilObjectID, err
	}
	if err := ctx.Err(); err != nil {
		return bson.NilObjectID, err
	}

	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return bson.NilObjectID, err
	}

	raw, err := objectIDDocument(document)
	if err != nil {
		return bson.NilObjectID, err
	}
	result, err := collection.InsertOne(ctx, raw)
	if err != nil {
		return bson.NilObjectID, operationError("insert one", err)
	}

	objId, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		return bson.NilObjectID, ErrUnexpectedInsertID
	}

	return objId, nil
}

// Reject unsupported IDs before issuing a write, rather than reporting failure
// after a successful insert. Use the native collection for custom ID types.
func objectIDDocument(document any) (bson.Raw, error) {
	raw, err := bson.Marshal(document)
	if err != nil {
		return nil, err
	}
	id := bson.Raw(raw).Lookup("_id")
	if id.Type != 0 && id.Type != bson.TypeObjectID {
		return nil, ErrObjectIDRequired
	}
	return bson.Raw(raw), nil
}

// InsertManyContext is InsertMany with caller-owned cancellation and deadline.
func InsertManyContext(ctx context.Context, name, dbName, collName string, documents []any) ([]any, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	result, err := collection.InsertMany(ctx, documents)
	if err != nil {
		if result != nil {
			return result.InsertedIDs, operationError("insert many partially", err)
		}
		return nil, operationError("insert many", err)
	}

	return result.InsertedIDs, nil
}

// FindOneContext is FindOne with caller-owned cancellation and deadline.
func FindOneContext(ctx context.Context, name, dbName, collName string, filter any, opts ...options.Lister[options.FindOneOptions]) (*mongo.SingleResult, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	result := collection.FindOne(ctx, filter, opts...)
	return result, result.Err()
}

// FindContext is Find with caller-owned cancellation and deadline.
// Keep ctx alive while iterating and close the returned cursor.
func FindContext(ctx context.Context, name, dbName, collName string, filter any, opts ...options.Lister[options.FindOptions]) (*mongo.Cursor, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	cursor, err := collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, operationError("find", err)
	}

	return cursor, nil
}

// UpdateOneContext is UpdateOne with caller-owned cancellation and deadline.
func UpdateOneContext(ctx context.Context, name, dbName, collName string, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	result, err := collection.UpdateOne(ctx, filter, update, opts...)
	if err != nil {
		return nil, operationError("update one", err)
	}

	return result, nil
}

// UpdateManyContext is UpdateMany with caller-owned cancellation and deadline.
func UpdateManyContext(ctx context.Context, name, dbName, collName string, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (*mongo.UpdateResult, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	result, err := collection.UpdateMany(ctx, filter, update, opts...)
	if err != nil {
		return nil, operationError("update many", err)
	}

	return result, nil
}

// UpdateByIDContext is UpdateByID with caller-owned cancellation and deadline.
func UpdateByIDContext(ctx context.Context, name, dbName, collName string, id bson.ObjectID, update any, opts ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update, opts...)
	if err != nil {
		return nil, operationError("update by ID", err)
	}

	return result, nil
}

// DeleteOneContext is DeleteOne with caller-owned cancellation and deadline.
func DeleteOneContext(ctx context.Context, name, dbName, collName string, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	result, err := collection.DeleteOne(ctx, filter, opts...)
	if err != nil {
		return nil, operationError("delete one", err)
	}

	return result, nil
}

// DeleteManyContext is DeleteMany with caller-owned cancellation and deadline.
func DeleteManyContext(ctx context.Context, name, dbName, collName string, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	result, err := collection.DeleteMany(ctx, filter, opts...)
	if err != nil {
		return nil, operationError("delete many", err)
	}

	return result, nil
}

// CountDocumentsContext is CountDocuments with caller-owned cancellation and deadline.
func CountDocumentsContext(ctx context.Context, name, dbName, collName string, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return 0, err
	}

	count, err := collection.CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, operationError("count documents", err)
	}

	return count, nil
}

// EstimatedDocumentCountContext is EstimatedDocumentCount with caller-owned cancellation and deadline.
func EstimatedDocumentCountContext(ctx context.Context, name, dbName, collName string, opts ...options.Lister[options.EstimatedDocumentCountOptions]) (int64, error) {
	if ctx == nil {
		err := ErrNilContext
		return 0, err
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return 0, err
	}

	count, err := collection.EstimatedDocumentCount(ctx, opts...)
	if err != nil {
		return 0, operationError("estimate document count", err)
	}

	return count, nil
}

// AggregateContext is Aggregate with caller-owned cancellation and deadline.
// Keep ctx alive while iterating and close the returned cursor.
func AggregateContext(ctx context.Context, name, dbName, collName string, pipeline any, opts ...options.Lister[options.AggregateOptions]) (*mongo.Cursor, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	cursor, err := collection.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return nil, operationError("aggregate", err)
	}

	return cursor, nil
}

// DistinctContext is Distinct with caller-owned cancellation and deadline.
func DistinctContext(ctx context.Context, name, dbName, collName string, fieldName string, filter any, opts ...options.Lister[options.DistinctOptions]) (*mongo.DistinctResult, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	result := collection.Distinct(ctx, fieldName, filter, opts...)

	return result, result.Err()
}

// DropIndexContext is DropIndex with caller-owned cancellation and deadline.
func DropIndexContext(ctx context.Context, name, dbName, collName string, indexName string, opts ...options.Lister[options.DropIndexesOptions]) error {
	if ctx == nil {
		err := ErrNilContext
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return err
	}

	err = collection.Indexes().DropOne(ctx, indexName, opts...)
	if err != nil {
		return operationError("drop index", err)
	}

	return nil
}

// CreateIndexContext is CreateIndex with caller-owned cancellation and deadline.
func CreateIndexContext(ctx context.Context, name, dbName, collName string, indexes []mongo.IndexModel, opts ...options.Lister[options.CreateIndexesOptions]) ([]string, error) {
	if ctx == nil {
		err := ErrNilContext
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	collection, err := GetCollection(name, dbName, collName)
	if err != nil {
		return nil, err
	}

	indexNames, err := collection.Indexes().CreateMany(ctx, indexes, opts...)
	if err != nil {
		return nil, operationError("create index", err)
	}

	return indexNames, nil
}
