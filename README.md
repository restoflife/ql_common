# ql_common

Go 后端基础组件库：XORM/MySQL、MongoDB、Redis 和 Zap/Gin 日志。
要求 Go 1.25.0 或更高版本。没有应用入口，不会自动读取配置或启动服务。

## 模块

| 包 | 作用 |
| --- | --- |
| `db` | 命名主从引擎组、Context 会话、事务回调、可选表结构同步 |
| `mongo` | 命名客户端、CRUD、分页、聚合、索引，MongoDB Driver v2 |
| `redis` | 命名单机、Sentinel、Cluster 客户端与常用命令 |
| `logger` | Zap 日志、文件轮转、Gin 请求/恢复中间件、数据库日志适配 |

## 接入示例

以下代码展示 Redis 接入；地址需替换成自己的配置。库不读取环境变量。

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/restoflife/ql_common/logger"
    cache "github.com/restoflife/ql_common/redis"
)

func main() {
    if err := logger.Init(&logger.Config{
        Level: "info", Console: "info", Filename: "logs/app.log",
        MaxSize: 10, MaxBackups: 5, MaxAge: 7, Format: "json",
    }); err != nil {
        log.Fatal(err)
    }
    defer logger.CloseAll()

    startup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    err := cache.BootUpRedisContext(startup, map[string]*cache.Config{
        "default": {Addr: "127.0.0.1:6379", PoolSize: 10, MinIdle: 2},
    })
    cancel()
    if err != nil {
        logger.Errorf("Redis 初始化失败: %v", err)
        return
    }
    defer cache.ShutdownRedis()

    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    if err := cache.SetContext(ctx, "default", "example", "value", time.Minute); err != nil {
        logger.Errorf("写入失败: %v", err)
    }
}
```

服务端收到退出信号时：先停止接收请求、等待在途请求结束，再关闭存储连接，
最后关闭日志。示例中的 defer 不替代 HTTP 服务的 graceful shutdown。

### SQL

```go
err := db.BootUpXORMContext(ctx, map[string]*db.XORMConfigLite{
    "default": {
        Driver: "mysql",
        Dsn: "user:password@tcp(127.0.0.1:3306)/app?parseTime=true",
        MaxOpen: 20, MaxIdle: 5, MaxLife: 3600,
    },
}, logger.Logger())
// 必须先检查 err。
err = db.Transaction(ctx, "default", func(s *xorm.Session) error {
    _, err := s.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", 10, 1)
    return err
})
```

- `MaxLife` 单位为秒。池配置为零时保留驱动默认值。
- 事务回调返回错误时回滚；发生 panic 时回滚并继续传播 panic；提交错误返回给调用方。
- `NewSessionContext` 返回的会话必须用 `defer db.Close(session)` 关闭。
- 只自动注册 MySQL 驱动；其他 XORM 支持的驱动需应用自行导入。
- 只有 `Synchronization=true` 且提供 `SetSyncFunc` 时才调用同步函数。
  DDL/同步回调的外部副作用不属于初始化回滚范围。回调不得递归调用本包初始化或关闭。

### MongoDB

```go
err := mongo.BootUpMongoContext(ctx, map[string]*mongo.Config{
    "default": {
        URI: "mongodb://127.0.0.1:27017",
        Username: "app", Password: "password", AuthSource: "admin",
        MaxPoolSize: 20,
    },
}, logger.Logger())
// 必须先检查 err。
cursor, err := mongo.FindContext(ctx, "default", "app", "users", bson.M{})
if err != nil { return err }
defer cursor.Close(context.Background())
for cursor.Next(ctx) {
    var user bson.M
    if err := cursor.Decode(&user); err != nil { return err }
}
return cursor.Err()
```

- 上述片段放在返回 error 的业务函数内；`ctx` 由应用设置 deadline，
  并覆盖整个游标读取过程。大结果集不要无界加载到内存。
- 显式用户名密码不会被日志选项覆盖；CA 文件开启 TLS 时最低版本为 TLS 1.2。
- `dbName` 为空时使用 `Config.Database`，显式传入时覆盖默认数据库。
- `FindOneContext` 会直接返回查询错误，包括驱动的 `mongo.ErrNoDocuments`。
- `DistinctContext` 返回强类型 `*mongo.DistinctResult`，需要调用其 `Decode`。
- `InsertOne`/`InsertOneContext` 只支持 ObjectID 主键，不支持的显式 _id 在写入前拒绝；自定义主键使用 `GetCollection`
  获得原生 Collection 并调用驱动。批量写入失败可能已部分成功，业务需设计幂等性。
- `Pagination(page, pageSize, maxSize)`：页码最小 1，默认 pageSize=10，
  maxSize<=0 时使用 100；先修正页大小再计算偏移，并防止整数溢出。

### Redis

- 单机：`Mode=""` 或 `"standalone"`，设置 `Addr`。
- Sentinel：`Mode="sentinel"`，设置 `MasterName` 和 `Slaves`（哨兵地址列表）。
- Cluster：`Mode="cluster"`，设置 `Slaves`（集群节点），`DB` 必须为 0。
- 未知模式、负池参数、空必要地址会返回错误，不再默默退回单机模式。
- 命令均有同名 `Context` 后缀版本，如 `GetContext`、`ScanContext`、
  `PipelineContext`、`EvalContext`。Context 不能为 nil。
- 批处理回调中也要向驱动命令传递同一个 ctx。
- 用 `errors.Is(err, redis.Nil)` 判断不存在的键；不要把缺失当作连接故障。
- 优先使用 SCAN，避免在大库执行 KEYS。管道不是自动重试/业务幂等保证。
- 需要未封装的能力时，使用 `GetRedis` 返回的原生 `UniversalClient`。

### 日志

- 推荐 `logger.Init(config) error` 和 `config.Build() (*zap.Logger, error)`；
  历史 `New`、`NewLogger` 遇到非法配置会在初始化时 panic。
- 配置按值复制，不修改调用方对象；空文件名禁用文件输出，未配置任何目标时输出到控制台。
- 默认 logger 未初始化时，普通包级日志调用安全地不输出；`Logger()` 返回 nil，
  `MustLogger()` 仍会 panic。Fatal/Panic 仍具有终止/抛出语义，不应用于可恢复故障。
- `SyncAll` 包括默认日志；`CloseAll` 同时关闭轮转文件并清空注册表。
  关闭前停止日志生产，关闭后不要继续使用已取得的 logger 指针。
- Gin 注册顺序建议：`r.Use(logger.GinLogger(log), logger.Recovery(log))`，
  这样访问日志能观察到恢复中间件设置的 500 状态。
- `GinLogger` 使用传入 Zap 的目标；`WithWriter` 的显式 writer 替换目标并输出 JSON，
  仍遵守传入 logger 的等级设置。
- 默认不记录查询字符串、请求头、Mongo 命令/回复正文或 Mongo URI。SQL 日志会绑定参数，调用方必须避免记录敏感值。
  SQL 文本中的字面量、业务错误、panic 内容和应用自行添加的字段仍可能含敏感数据；
  应使用参数化 SQL，并在业务层脱敏。

## 生命周期和错误契约

- 一个包内的初始化/关闭操作串行化，查询注册表并发安全。
- 同一批配置全部成功后才对外可见；失败会清理本次已创建的连接，不影响此前成功的实例。
- 名称不能为空，重复名称返回可用 `errors.Is(err, 包名.ErrDuplicate)` 判断的错误；
  未注册名称返回 `包名.ErrNotFound`。
- `ShutdownXormE`、`ShutdownRedisE`、`ShutdownMongoE` 汇总关闭错误；
  原无返回值 Shutdown 函数保留，并记录关闭失败。
- 重复关闭安全；关闭后名称解除注册，可以重新初始化。
- 关闭不是在途请求排空器：从注册表取得的客户端/会话仍由调用方协调使用期限。
- 初始化期间不要修改传入的配置 map、结构体或切片。
- 不自动启动永久 Ping 协程。应用自行调度健康检查，用有 deadline 的 ctx 调用
  Redis Client.Ping、Mongo Client.Ping 或 XORM Session.PingContext。
- 兼容启动函数 `MustBootUp*` 返回 error、不 panic，使用 10 秒启动超时；
  显式 `BootUp*Context` 的总时限由调用方决定，清理使用独立关闭上下文。

## 公开 API 兼容策略

- 已发布的导出函数、类型、配置字段和错误变量不直接改名、删除或改变既有含义。
- 新能力优先增加新 API；旧 API 保留并标记弃用，至少跨一个主版本后才允许删除。
- 修复安全问题、数据错误或资源泄漏时可以调整错误行为，但必须在升级说明中列出。
- 每次发布前运行单元测试、竞态测试和集成测试；调用方升级后仍需编译并验证自身项目。
- 数据库名称、集合名称、业务错误码、项目目录和配置文件路径由调用方传入，公共库不保存项目默认值。
## 升级注意

原有公开函数保留，但以下行为刻意修正：

1. `MaxLife` 从错误的毫秒解释修正为配置注释约定的秒，请检查已有配置。
2. 初始化从“部分注册”改为批次原子注册；自动后台 Ping 被移除。
3. 默认日志纳入刷新/关闭，空文件名不再生成隐式临时日志文件。
4. Gin 不再记录 query，Mongo 不再记录正文；SQL 日志会绑定参数以便排查问题。
5. `FindOne` 和 `Distinct` 现在直接返回操作错误，不再只把错误留在结果对象内。
6. 分页修正非法页码/页大小；nil 配置、未知 Redis 模式等尽早返回错误。

旧 Redis 方法不接受请求 ctx，保留原有 Background 行为；新业务请使用 Context 版本。
旧 Mongo 方法的超时仅覆盖该次调用，不覆盖调用方后续游标遍历；游标调用必须自行传入 ctx。

## 验证

无需数据库的测试：

```sh
go test ./...
go vet ./...
go test -race ./...
go test -cover ./...
```

单元测试覆盖注册表并发/失败清理、事务提交/回滚/panic、配置验证、日志安全、
堆栈复用，以及所有 Context 操作入口的参数/取消/缺失实例处理。
这些测试不等于驱动网络行为或真实服务器兼容性验证。

真实数据库集成测试需显式设置以下变量，且只能使用可丢弃的测试服务：

- `QL_TEST_MYSQL_DSN`：测试库 DSN，需有建表/删表权限。
- `QL_TEST_REDIS_ADDR`：测试 Redis 地址。
- `QL_TEST_MONGO_URI`、`QL_TEST_MONGO_USER`、`QL_TEST_MONGO_PASSWORD`：
  测试 Mongo 连接与可选认证，认证库为 admin。

```sh
go test -race -tags=integration -count=1 -timeout=3m ./...
```

未设置变量时对应集成测试跳过；测试创建专属表/键/集合并清理，不应连接生产环境。
CI 配置包含 Linux/Windows 单元竞态测试，以及隔离容器中的数据库集成测试。
Sentinel/Cluster 故障切换、TLS 互通、长时间负载和业务数据迁移不在现有测试覆盖范围。

## 发布前

本仓库修改不会自动发布版本。发布前应运行完整 CI、复核迁移项，并由维护者确认开源许可证、
版本号和兼容性策略；当前未擅自添加许可证或创建发布标签。

### 统一连接管理与扩展命令

先使用 BootUpXORMContext / BootUpRedisContext / BootUpMongoContext 注册命名实例，
再调用对应包的 Context 方法；请求传入自身 Context，不要重复创建或关闭客户端。

- db.GetEngineGroup(name)：获取注册的共享读写连接组。XORMConfigLite.MasterPool / SlavePools 可分别设置连接数及 time.Duration 生命周期；不设置时沿用旧参数。
- redis.Config 支持 TLSConfig、MaxActiveConns、MaxIdleConns、ConnMaxIdleTime。
- redis.Z 是驱动 Z 类型的别名，可直接 redis.ZAddContext(ctx, "game", "rank", redis.Z{Score: 100, Member: "123"})。
- redis.ZRevRangeContext / ZRevRangeWithScoresContext：按分数倒序查询。
- redis.DoContext(ctx, "game", "COMMAND", arg1, arg2)：执行未封装命令，参数独立传入，返回值为 any，必须检查 error。
- mongo.Config 支持 MaxConnecting、Timeout。CRUD 方法的 dbName 为空时使用 Config.Database，显式库名则覆盖默认值。
- mongo.InsertOneResultContext：支持字符串等任意 _id 类型，返回驱动 InsertOneResult；原 InsertOneContext 保持 ObjectID 返回值兼容。
- mongo.RunCommandContext(ctx, "game", "", bson.D{{Key: "ping", Value: 1}})：执行有序 BSON 命令。先检查调用错误，再检查结果 Decode 的错误。

常用 String/Hash/List/Set/ZSet、Pipeline、Lua，以及 MongoDB CRUD/聚合/索引已有 Context 方法。
未覆盖的高级功能也可通过 GetRedis / GetClient / GetCollection 获取同一注册实例的驱动句柄。
ShutdownXormE / ShutdownRedisE / ShutdownMongoE 在进程退出且请求排空后统一调用；Mongo 游标与 MySQL Session 仍由调用方释放。