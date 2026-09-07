package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync/atomic"
	"xorm.io/builder"
	"xorm.io/xorm/log"
)

// XormLogger is a concurrency-safe XORM logger. SQL arguments are bound for logging.
type XormLogger struct {
	logger *zap.Logger
	plain  *zap.Logger
	show   atomic.Bool
	level  atomic.Int32
}

func NewXormLogger(l *zap.Logger) *XormLogger {
	if l == nil {
		l = zap.NewNop()
	}
	result := &XormLogger{logger: l, plain: l.WithOptions(zap.WithCaller(false))}
	result.show.Store(true)
	result.level.Store(int32(log.LOG_DEBUG))
	return result
}
func (o *XormLogger) BeforeSQL(ctx log.LogContext) {}
func (o *XormLogger) AfterSQL(ctx log.LogContext) {
	if !o.IsShowSQL() {
		return
	}
	level := zapcore.InfoLevel
	threshold := log.LOG_INFO
	if ctx.Err != nil {
		level = zapcore.ErrorLevel
		threshold = log.LOG_ERR
	}
	if o.Level() > threshold {
		return
	}
	sql, convertErr := builder.ConvertToBoundSQL(ctx.SQL, ctx.Args)
	if convertErr != nil {
		sql = ctx.SQL
	}
	logger := o.plain
	if ctx.Err != nil {
		logger = o.logger
	}
	if entry := logger.Check(level, SQL); entry != nil {
		entry.Write(zap.String("sql", sql), zap.String("latency", ctx.ExecuteTime.String()), zap.Error(ctx.Err), zap.Error(convertErr))
	}
}
func (o *XormLogger) Debugf(format string, v ...interface{}) {
	if o.Level() <= log.LOG_DEBUG {
		o.plain.Debug(fmt.Sprintf(format, v...))
	}
}
func (o *XormLogger) Infof(format string, v ...interface{}) {
	if o.Level() <= log.LOG_INFO {
		o.plain.Info(fmt.Sprintf(format, v...))
	}
}
func (o *XormLogger) Warnf(format string, v ...interface{}) {
	if o.Level() <= log.LOG_WARNING {
		o.plain.Warn(fmt.Sprintf(format, v...))
	}
}
func (o *XormLogger) Errorf(format string, v ...interface{}) {
	if o.Level() <= log.LOG_ERR {
		o.logger.Error(fmt.Sprintf(format, v...))
	}
}
func (o *XormLogger) Level() log.LogLevel         { return log.LogLevel(o.level.Load()) }
func (o *XormLogger) SetLevel(level log.LogLevel) { o.level.Store(int32(level)) }
func (o *XormLogger) ShowSQL(show ...bool) {
	if len(show) > 0 {
		o.show.Store(show[0])
	}
}
func (o *XormLogger) IsShowSQL() bool { return o.show.Load() }
