package ql_common_test

import (
	"context"

	driver "github.com/redis/go-redis/v9"
	"github.com/restoflife/ql_common/db"
	"github.com/restoflife/ql_common/logger"
	"github.com/restoflife/ql_common/mongo"
	"github.com/restoflife/ql_common/redis"
	"go.uber.org/zap"
	"xorm.io/xorm"
)

// These assignments make accidental changes to the supported API fail at compile time.
var (
	_ func(context.Context, map[string]*db.XORMConfigLite, *zap.Logger, ...db.Option) error = db.BootUpXORMContext
	_ func(string) (*xorm.EngineGroup, error)                                               = db.GetEngineGroup
	_ func() error                                                                          = db.ShutdownXormE
	_ func(context.Context, map[string]*redis.Config) error                                 = redis.BootUpRedisContext
	_ func(string) (driver.UniversalClient, error)                                          = redis.GetRedis
	_ func() error                                                                          = redis.ShutdownRedisE
	_ func(context.Context, map[string]*mongo.Config, *zap.Logger) error                    = mongo.BootUpMongoContext
	_ func() error                                                                          = mongo.ShutdownMongoE
	_ func(*logger.Config) error                                                            = logger.Init
	_ func() error                                                                          = logger.CloseAll
	_ func(error, int) error                                                                = logger.WithStackSkip
)
