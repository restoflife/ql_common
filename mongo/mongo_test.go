package mongo

import (
	"context"
	"errors"
	"math"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

func TestObjectIDValidationBeforeWrite(t *testing.T) {
	for _, document := range []any{bson.M{"_id": "custom"}, bson.M{"_id": 1}, bson.M{"_id": nil}} {
		if _, err := objectIDDocument(document); err == nil {
			t.Fatal("accepted unsupported ID")
		}
	}
	for _, document := range []any{bson.M{"value": 1}, bson.M{"_id": bson.NewObjectID()}} {
		if _, err := objectIDDocument(document); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLoggerPreservesCredentials(t *testing.T) {
	c := &Config{URI: "mongodb://localhost:27017", Username: "user", Password: "secret", AuthSource: "admin", MaxPoolSize: 10, MinPoolSize: 2}
	for _, l := range []*zap.Logger{nil, zap.NewNop()} {
		opts, err := clientOptions(c, l)
		if err != nil {
			t.Fatal(err)
		}
		if opts.Auth == nil || opts.Auth.Username != "user" || opts.Auth.Password != "secret" || opts.Auth.AuthSource != "admin" {
			t.Fatal("credentials overwritten")
		}
		if *opts.MaxPoolSize != 10 || *opts.MinPoolSize != 2 {
			t.Fatal("pool options lost")
		}
		if l != nil && opts.Monitor == nil {
			t.Fatal("monitor lost")
		}
	}
}

func TestPagination(t *testing.T) {
	for _, tc := range []struct{ page, size, max, skip, limit int64 }{
		{2, 100, 20, 20, 20}, {0, 0, 0, 0, 10}, {2, 0, 100, 10, 10}, {-1, -1, -1, 0, 10},
		{3, 5, 3, 6, 3}, {math.MaxInt64, 100, 100, math.MaxInt64, 100},
	} {
		var opts options.FindOptions
		for _, f := range Pagination(tc.page, tc.size, tc.max).List() {
			if err := f(&opts); err != nil {
				t.Fatal(err)
			}
		}
		if *opts.Skip != tc.skip || *opts.Limit != tc.limit {
			t.Fatalf("%+v: skip=%d limit=%d", tc, *opts.Skip, *opts.Limit)
		}
	}
}

func TestInvalidConfigAndContext(t *testing.T) {
	for _, c := range []*Config{nil, {}, {URI: "invalid"}, {URI: "mongodb://localhost", MaxPoolSize: 1, MinPoolSize: 2}, {URI: "mongodb://localhost", Password: "secret"}} {
		if _, err := clientOptions(c, nil); err == nil {
			t.Fatalf("accepted %+v", c)
		}
	}
	if err := BootUpMongoContext(nil, nil, nil); err == nil {
		t.Fatal("nil context accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := FindOneContext(ctx, "missing", "db", "coll", nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := FindContext(nil, "missing", "db", "coll", nil); err == nil {
		t.Fatal("nil context accepted")
	}
	if _, err := GetClient("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if err := ShutdownMongoE(); err != nil {
		t.Fatal(err)
	}
	if err := ShutdownMongoE(); err != nil {
		t.Fatal(err)
	}
}
