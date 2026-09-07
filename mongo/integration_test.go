//go:build integration

package mongo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
	"os"
	"testing"
	"time"
)

func TestIntegrationAuthCRUDAndCursor(t *testing.T) {
	uri := os.Getenv("QL_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("QL_TEST_MONGO_URI not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	configs := map[string]*Config{"integration": {URI: uri, Username: os.Getenv("QL_TEST_MONGO_USER"), Password: os.Getenv("QL_TEST_MONGO_PASSWORD"), AuthSource: "admin"}}
	if err := BootUpMongoContext(ctx, configs, zap.NewNop()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := ShutdownMongoE(); err != nil {
			t.Error(err)
		}
	})
	collectionName := "ql_common_test_" + bson.NewObjectID().Hex()
	collection, err := GetCollection("integration", "ql_common_test", collectionName)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := collection.Drop(cleanup); err != nil {
			t.Error(err)
		}
	}()
	ids, err := InsertManyContext(ctx, "integration", "ql_common_test", collectionName, []any{bson.M{"n": 1}, bson.M{"n": 2}, bson.M{"n": 3}})
	if err != nil || len(ids) != 3 {
		t.Fatalf("ids=%v err=%v", ids, err)
	}
	cursor, err := FindContext(ctx, "integration", "ql_common_test", collectionName, bson.M{}, options.Find().SetBatchSize(1))
	if err != nil {
		t.Fatal(err)
	}
	defer cursor.Close(context.Background())
	count := 0
	for cursor.Next(ctx) {
		count++
	}
	if cursor.Err() != nil || count != 3 {
		t.Fatalf("count=%d err=%v", count, cursor.Err())
	}
	result, err := FindOneContext(ctx, "integration", "ql_common_test", collectionName, bson.M{"n": 1})
	if err != nil {
		t.Fatal(err)
	}
	var doc bson.M
	if err := result.Decode(&doc); err != nil {
		t.Fatal(err)
	}
	distinct, err := DistinctContext(ctx, "integration", "ql_common_test", collectionName, "n", bson.M{})
	if err != nil {
		t.Fatal(err)
	}
	var values []int
	if err := distinct.Decode(&values); err != nil || len(values) != 3 {
		t.Fatalf("values=%v err=%v", values, err)
	}
	if err := BootUpMongoContext(ctx, configs, zap.NewNop()); !errors.Is(err, ErrDuplicate) {
		t.Fatal(err)
	}
}
