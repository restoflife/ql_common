package logger

import (
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config 定义日志系统的配置结构体
type Config struct {
	Level      string `json:"level"`       // 文件日志输出等级（如 "info"、"debug"）
	Filename   string `json:"file"`        // 日志文件路径
	MaxSize    int    `json:"max_size"`    // 每个日志文件最大尺寸（MB）
	MaxBackups int    `json:"max_backups"` // 保留的旧日志文件个数
	MaxAge     int    `json:"max_age"`     // 保留旧日志的最大天数
	Console    string `json:"console"`     // 控制台输出的日志等级
	Format     string `json:"format"`      // 输出格式："json" 或 "text"
}

var (
	mu         sync.Mutex    // 互斥锁，确保并发安全
	allLoggers []*zap.Logger // 存储所有初始化过的 logger
	defaultLog *zap.Logger   // 默认 logger 实例
)

// GetAll 返回所有注册的日志实例
func GetAll() []*zap.Logger {
	return allLoggers
}

// New 初始化默认日志实例
func New(g *Config) {
	defaultLog = g.NewLogger()
	mu.Lock()
	allLoggers = append(allLoggers, defaultLog)
	mu.Unlock()
}

// NewLogger 基于当前配置创建新的 zap.Logger 实例
func (l *Config) NewLogger() *zap.Logger {
	// 根据格式创建 encoder（编码器）
	encoder := createEncoder(l.Format, false)     // 用于文件
	consoleEncoder := createEncoder("text", true) // 用于控制台

	// 构建多个输出核心（core），文件 + 控制台
	cores := make([]zapcore.Core, 0)
	cores = append(
		cores,
		zapcore.NewCore(
			encoder,
			zapcore.AddSync(&lumberjack.Logger{ // 文件输出 + 日志切割
				Filename:   l.Filename,
				MaxSize:    l.MaxSize,
				MaxBackups: l.MaxBackups,
				MaxAge:     l.MaxAge,
				LocalTime:  true,
			}),
			createLevelEnablerFunc(l.Level), // 控制输出等级
		),
		zapcore.NewCore(
			consoleEncoder,
			zapcore.Lock(os.Stderr), // 控制台输出到标准错误
			createLevelEnablerFunc(l.Console),
		),
	)

	// 创建 logger
	var logger *zap.Logger
	logger = zap.New(zapcore.NewTee(cores...)) // 多个 core 合并

	// 注册到全局 logger 列表中
	mu.Lock()
	allLoggers = append(allLoggers, logger)
	mu.Unlock()

	return logger
}

// Logger 返回默认 logger 实例，若未初始化则 panic
func Logger() *zap.Logger {
	if defaultLog == nil {
		panic("logger not initialized, call New first")
	}
	return defaultLog
}

// createLevelEnablerFunc 根据字符串日志级别创建 zap.LevelEnablerFunc
func createLevelEnablerFunc(input string) zap.LevelEnablerFunc {
	var lv = new(zapcore.Level)
	if err := lv.UnmarshalText([]byte(input)); err != nil {
		return nil
	}
	return func(lev zapcore.Level) bool {
		return lev >= *lv
	}
}

// createEncoder 创建日志编码器：支持 JSON 或 控制台格式
func createEncoder(format string, isConsole bool) zapcore.Encoder {
	var cfg zapcore.EncoderConfig
	if isConsole {
		cfg = zap.NewDevelopmentEncoderConfig() // 开发模式：更易读
	} else {
		cfg = zap.NewProductionEncoderConfig() // 生产模式：简洁实用
		cfg.CallerKey = "func"
	}
	cfg.EncodeTime = timeEncoder                        // 时间格式化函数
	cfg.EncodeLevel = zapcore.LowercaseLevelEncoder     // 等级小写输出
	cfg.EncodeDuration = zapcore.SecondsDurationEncoder // 时间间隔单位：秒
	cfg.EncodeCaller = zapcore.ShortCallerEncoder       // 短路径调用者信息

	switch format {
	case "json":
		return zapcore.NewJSONEncoder(cfg)
	default:
		return zapcore.NewConsoleEncoder(cfg) // 默认控制台格式
	}
}

// timeEncoder 时间格式化函数（RFC3339 格式）
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.RFC3339))
}

// 以下是对默认 logger 的简化调用封装（适合外部直接调用）

func Info(msg string, fields ...zapcore.Field)  { Logger().Info(msg, fields...) }
func Debug(msg string, fields ...zapcore.Field) { Logger().Debug(msg, fields...) }
func Warn(msg string, fields ...zapcore.Field)  { Logger().Warn(msg, fields...) }
func Error(msg string, fields ...zapcore.Field) { Logger().Error(msg, fields...) }
func Panic(msg string, fields ...zapcore.Field) { Logger().Panic(msg, fields...) }
func Fatal(msg string, fields ...zapcore.Field) { Logger().Fatal(msg, fields...) }

// 支持格式化输出的调用方式（使用 zap.SugaredLogger）

func Infof(format string, args ...any)  { Logger().Sugar().Infof(format, args...) }
func Debugf(format string, args ...any) { Logger().Sugar().Debugf(format, args...) }
func Warnf(format string, args ...any)  { Logger().Sugar().Warnf(format, args...) }
func Errorf(format string, args ...any) { Logger().Sugar().Errorf(format, args...) }
func Panicf(format string, args ...any) { Logger().Sugar().Panicf(format, args...) }
func Fatalf(format string, args ...any) { Logger().Sugar().Fatalf(format, args...) }

// SyncAll 刷新所有日志缓冲（写入磁盘或终端）
func SyncAll() {
	mu.Lock()
	defer mu.Unlock()
	for _, l := range allLoggers {
		if l != nil {
			_ = l.Sync() // zap.Sync 可能返回 error，但一般忽略
		}
	}
}
