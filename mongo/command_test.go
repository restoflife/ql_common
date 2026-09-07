package mongo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestDefaultDatabaseAndCustomCommand(t *testing.T) {
	c, err := driver.Connect()
	if err != nil {
		t.Fatal(err)
	}
	defaultDatabases.Store(c, "game_db")
	if err := clientMap.Init([]string{"game"}, func(string) (*driver.Client, error) { return c, nil }, disconnect); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ShutdownMongoE() })
	coll, err := GetCollection("game", "", "records")
	if err != nil || coll.Database().Name() != "game_db" {
		t.Fatal("default database lost", err)
	}
	db, err := GetDatabase("game", "archive")
	if err != nil || db.Name() != "archive" {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := RunCommandContext(ctx, "game", "", bson.D{{Key: "ping", Value: 1}}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := ShutdownMongoE(); err != nil {
		t.Fatal(err)
	}
	if _, ok := defaultDatabases.Load(c); ok {
		t.Fatal("default database leaked")
	}
}
func TestExtendedOptions(t *testing.T) {
	opts, err := clientOptions(&Config{URI: "mongodb://localhost", MaxConnecting: 3, Timeout: 4 * time.Second, MaxPoolSize: 10, MinPoolSize: 2}, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	if *opts.MaxConnecting != 3 || *opts.Timeout != 4*time.Second || opts.Monitor == nil {
		t.Fatal("options lost")
	}
}
