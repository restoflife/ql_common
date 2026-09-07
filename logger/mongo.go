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
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// MongoLogger provides the corresponding package operation.
type MongoLogger struct {
	logger *zap.Logger
	plain  *zap.Logger
	show   atomic.Bool
}

// NewMongoLogger provides the corresponding package operation.
func NewMongoLogger(zapLogger *zap.Logger) *MongoLogger {
	if zapLogger == nil {
		zapLogger = zap.NewNop()
	}
	result := &MongoLogger{logger: zapLogger, plain: zapLogger.WithOptions(zap.WithCaller(false))}
	result.show.Store(true)
	return result
}

// CommandMonitor provides the corresponding package operation.
func (m *MongoLogger) CommandMonitor() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started:   m.commandStarted,
		Succeeded: m.commandSucceeded,
		Failed:    m.commandFailed,
	}
}

// commandStarted provides the corresponding package operation.
func (m *MongoLogger) commandStarted(ctx context.Context, evt *event.CommandStartedEvent) {
	if !m.show.Load() {
		return
	}

	if m.plain.Core().Enabled(zapcore.DebugLevel) {
		m.plain.Debug("MongoDB Command Started",
			zap.String("database", evt.DatabaseName),
			zap.String("command", evt.CommandName),
			zap.String("connection_id", evt.ConnectionID),
		)
	}
}

// commandSucceeded provides the corresponding package operation.
func (m *MongoLogger) commandSucceeded(ctx context.Context, evt *event.CommandSucceededEvent) {
	if !m.show.Load() || evt.CommandName == "endSessions" {
		return
	}

	duration := evt.Duration * time.Nanosecond
	m.plain.Info("MongoDB Command Executed",
		zap.String("command", evt.CommandName),
		zap.String("database", evt.DatabaseName),
		zap.String("latency", duration.String()),
		zap.String("connection_id", evt.ConnectionID),
	)
}

// commandFailed provides the corresponding package operation.
func (m *MongoLogger) commandFailed(ctx context.Context, evt *event.CommandFailedEvent) {
	if !m.show.Load() {
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

// GetClientOptions provides the corresponding package operation.
func (m *MongoLogger) GetClientOptions(uri string) *options.ClientOptions {
	zapSink := &ZapMongoSink{Logger: m.logger, Plain: m.plain}

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

// ShowMongo provides the corresponding package operation.
func (m *MongoLogger) ShowMongo(b ...bool) {
	if len(b) > 0 {
		m.show.Store(b[0])
	}
}

// IsShowMongo provides the corresponding package operation.
func (m *MongoLogger) IsShowMongo() bool {
	return m.show.Load()
}

// ZapMongoSink provides the corresponding package operation.
type ZapMongoSink struct {
	Logger *zap.Logger
	Plain  *zap.Logger
}

func (z *ZapMongoSink) Info(level int, msg string, keysAndValues ...any) {
	fields := make([]zap.Field, 0)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			fields = append(fields, zap.Any(fmt.Sprintf("%v", keysAndValues[i]), keysAndValues[i+1]))
		}
	}
	logger := z.Plain
	if logger == nil {
		logger = z.Logger.WithOptions(zap.WithCaller(false))
	}
	logger.Info(msg, fields...)
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
