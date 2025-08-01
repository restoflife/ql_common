package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"xorm.io/builder"
	"xorm.io/xorm/log"
)

// XormLogger 实现了 xorm 的 log.Logger 接口，使用 zap 作为底层日志库
type XormLogger struct {
	logger *zap.Logger   // zap 的 logger 实例
	off    bool          // 是否关闭日志
	show   bool          // 是否显示 SQL
	level  log.LogLevel  // xorm 的日志级别
	logLvl zapcore.Level // zap 的日志级别
}

// NewXormLogger 创建一个新的 XormLogger 实例
func NewXormLogger(zapLogger *zap.Logger) *XormLogger {
	return &XormLogger{
		logger: zapLogger,
		show:   true,
	}
}

// BeforeSQL 在 SQL 执行前调用（可用于埋点，当前未使用）
func (o *XormLogger) BeforeSQL(ctx log.LogContext) {
	// 可用于记录执行前时间或打印 SQL 参数
}

// AfterSQL 在 SQL 执行后调用，记录 SQL、耗时和错误信息
func (o *XormLogger) AfterSQL(ctx log.LogContext) {
	sql, _ := builder.ConvertToBoundSQL(ctx.SQL, ctx.Args)
	o.logLvl = zapcore.InfoLevel
	if ctx.Err != nil {
		o.logLvl = zapcore.ErrorLevel
	}
	if o.logger.Core().Enabled(o.logLvl) {
		o.logger.Check(o.logLvl, SQL).Write(
			zap.String("sql", sql),
			zap.String("latency", ctx.ExecuteTime.String()),
			zap.Error(ctx.Err),
		)
	}
}

// Debugf 打印 debug 级别日志
func (o *XormLogger) Debugf(format string, v ...interface{}) {
	o.logger.Debug(fmt.Sprintf(format, v...))
}

// Infof 打印 info 级别日志
func (o *XormLogger) Infof(format string, v ...interface{}) {
	o.logger.Info(fmt.Sprintf(format, v...))
}

// Warnf 打印 warning 级别日志
func (o *XormLogger) Warnf(format string, v ...interface{}) {
	o.logger.Warn(fmt.Sprintf(format, v...))
}

// Errorf 打印 error 级别日志
func (o *XormLogger) Errorf(format string, v ...interface{}) {
	o.logger.Error(fmt.Sprintf(format, v...))
}

// Level 返回当前日志级别（xorm 使用）
func (o *XormLogger) Level() log.LogLevel {
	if o.off {
		return log.LOG_OFF
	}

	for _, l := range []zapcore.Level{
		zapcore.DebugLevel,
		zapcore.InfoLevel,
		zapcore.WarnLevel,
		zapcore.ErrorLevel,
	} {
		if o.logger.Core().Enabled(l) {
			switch l {
			case zapcore.DebugLevel:
				return log.LOG_DEBUG
			case zapcore.InfoLevel:
				return log.LOG_INFO
			case zapcore.WarnLevel:
				return log.LOG_WARNING
			case zapcore.ErrorLevel:
				return log.LOG_ERR
			}
		}
	}
	return log.LOG_UNKNOWN
}

// SetLevel 设置 xorm 的日志级别（不影响 zap 内部）
func (o *XormLogger) SetLevel(l log.LogLevel) {
	o.level = l
}

// ShowSQL 设置是否打印 SQL 日志
func (o *XormLogger) ShowSQL(b ...bool) {
	if len(b) > 0 {
		o.show = b[0]
	}
}

// IsShowSQL 返回当前是否打印 SQL 日志的状态
func (o *XormLogger) IsShowSQL() bool {
	return o.show
}
