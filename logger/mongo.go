/*
 * @Author:   admin
 * @IDE:      GoLand
 * @Date:     2025/10/17 15:38
 * @FilePath: qingliu//mongo.go
 */

package logger

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// MongoLogger 实现了 MongoDB 的自定义日志记录
type MongoLogger struct {
	logger *zap.Logger
	show   bool // 是否显示 MongoDB 操作
}

// NewMongoLogger 创建一个新的 MongoLogger 实例
func NewMongoLogger(zapLogger *zap.Logger) *MongoLogger {
	return &MongoLogger{
		logger: zapLogger,
		show:   true,
	}
}

// CommandMonitor 返回 MongoDB 命令监视器
func (m *MongoLogger) CommandMonitor() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started:   m.commandStarted,
		Succeeded: m.commandSucceeded,
		Failed:    m.commandFailed,
	}
}

// commandStarted 命令开始执行时调用
func (m *MongoLogger) commandStarted(ctx context.Context, evt *event.CommandStartedEvent) {
	if !m.show {
		return
	}

	// 记录命令开始，可以用于性能监控
	if m.logger.Core().Enabled(zapcore.DebugLevel) {
		m.logger.Debug("MongoDB Command Started",
			zap.String("database", evt.DatabaseName),
			zap.String("command", evt.Command.String()),
			zap.String("connection_id", evt.ConnectionID),
		)
	}
}

// commandSucceeded 命令成功执行时调用
func (m *MongoLogger) commandSucceeded(ctx context.Context, evt *event.CommandSucceededEvent) {
	if !m.show {
		return
	}

	duration := evt.Duration * time.Nanosecond
	replyDoc, err := bson.MarshalExtJSON(evt.Reply, false, false)
	if err != nil {
		m.logger.Error("解析 MongoDB 回复失败", zap.Error(err))
		return
	}
	m.logger.Info("MongoDB Command Executed",
		zap.String("command", evt.CommandName),
		zap.String("database", evt.DatabaseName),
		zap.String("latency", duration.String()),
		zap.String("connection_id", evt.ConnectionID),
		// zap.Any("reply", m.sanitizeReply(evt.Reply)),
		zap.ByteString("reply", replyDoc),
	)
}

// commandFailed 命令执行失败时调用
func (m *MongoLogger) commandFailed(ctx context.Context, evt *event.CommandFailedEvent) {
	if !m.show {
		return
	}

	duration := evt.Duration * time.Nanosecond

	m.logger.Error("MongoDB Command Failed",
		zap.String("command", evt.CommandName),
		zap.String("database", evt.DatabaseName),
		zap.String("latency", duration.String()),
		zap.String("connection_id", evt.ConnectionID),
		zap.Error(evt.Failure),
	)
}

// sanitizeReply 清理回复中的敏感信息
func (m *MongoLogger) sanitizeReply(reply bson.Raw) any {
	var doc bson.M
	if err := bson.Unmarshal(reply, &doc); err != nil {
		return "<unable to parse reply>"
	}

	// 移除敏感字段
	delete(doc, "password")
	delete(doc, "token")
	delete(doc, "secret")

	// 限制大字段的输出
	if data, ok := doc["data"]; ok {
		if byteData, ok := data.([]byte); ok && len(byteData) > 100 {
			doc["data"] = fmt.Sprintf("<binary data %d bytes>", len(byteData))
		}
	}

	return doc
}

// GetClientOptions 返回配置了日志记录的 MongoDB 客户端选项
func (m *MongoLogger) GetClientOptions(uri string) *options.ClientOptions {
	// 创建 zap 适配器用于 MongoDB 驱动内部日志
	zapSink := &ZapMongoSink{Logger: m.logger}

	return options.Client().
		ApplyURI(uri).
		SetMonitor(m.CommandMonitor()).
		SetLoggerOptions(
			options.Logger().
				SetComponentLevel(options.LogComponentCommand, options.LogLevelInfo).
				SetComponentLevel(options.LogComponentConnection, options.LogLevelInfo).
				SetComponentLevel(options.LogComponentTopology, options.LogLevelInfo).
				SetSink(zapSink),
		)
}

// ShowMongo 设置是否显示 MongoDB 操作日志
func (m *MongoLogger) ShowMongo(b ...bool) {
	if len(b) > 0 {
		m.show = b[0]
	}
}

// IsShowMongo 返回当前是否显示 MongoDB 操作日志的状态
func (m *MongoLogger) IsShowMongo() bool {
	return m.show
}

// ZapMongoSink 适配器，将 MongoDB 驱动日志转发到 zap
type ZapMongoSink struct {
	Logger *zap.Logger
}

func (z *ZapMongoSink) Info(level int, msg string, keysAndValues ...any) {
	fields := make([]zap.Field, 0)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			fields = append(fields, zap.Any(fmt.Sprintf("%v", keysAndValues[i]), keysAndValues[i+1]))
		}
	}
	z.Logger.Info(msg, fields...)
}

func (z *ZapMongoSink) Error(err error, msg string, keysAndValues ...any) {
	fields := make([]zap.Field, 0)
	fields = append(fields, zap.Error(err))
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			fields = append(fields, zap.Any(fmt.Sprintf("%v", keysAndValues[i]), keysAndValues[i+1]))
		}
	}
	z.Logger.Error(msg, fields...)
}
