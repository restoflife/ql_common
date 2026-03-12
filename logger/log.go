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

// 默认配置值
const (
	DefaultLevel   = "info"
	DefaultMaxSize = 10 // MB
	DefaultFormat  = "json"
)

var (
	mu         sync.Mutex    // 互斥锁，确保并发安全
	allLoggers []*zap.Logger // 存储所有初始化过的 logger
	defaultLog *zap.Logger   // 默认 logger 实例
)

// GetAll 返回所有注册的日志实例
func GetAll() []*zap.Logger {
	mu.Lock()
	defer mu.Unlock()
	result := make([]*zap.Logger, len(allLoggers))
	copy(result, allLoggers)
	return result
}

// New 初始化默认日志实例
func New(g *Config) {
	// 设置默认值
	if g.Level == "" {
		g.Level = DefaultLevel
	}
	if g.MaxSize == 0 {
		g.MaxSize = DefaultMaxSize
	}
	if g.Format == "" {
		g.Format = DefaultFormat
	}

	defaultLog = g.newLogger()
}

// newLogger 基于当前配置创建新的 zap.Logger 实例（内部使用，不添加到全局列表）
func (l *Config) newLogger() *zap.Logger {
	// 根据格式创建 encoder（编码器）
	encoder := createEncoder(l.Format, false)     // 用于文件
	consoleEncoder := createEncoder("text", true) // 用于控制台

	// 构建多个输出核心（core），文件 + 控制台
	cores := make([]zapcore.Core, 0, 2)

	// 始终添加文件 core
	cores = append(cores,
		zapcore.NewCore(
			encoder,
			zapcore.AddSync(&lumberjack.Logger{
				Filename:   l.Filename,
				MaxSize:    l.MaxSize,
				MaxBackups: l.MaxBackups,
				MaxAge:     l.MaxAge,
				LocalTime:  true,
			}),
			createLevelEnablerFunc(l.Level),
		),
	)

	// 只有当控制台级别有效时才添加控制台 core
	if consoleLevel := createLevelEnablerFunc(l.Console); consoleLevel != nil {
		cores = append(cores,
			zapcore.NewCore(
				consoleEncoder,
				zapcore.Lock(os.Stderr),
				consoleLevel,
			),
		)
	}

	// 创建 logger
	return zap.New(zapcore.NewTee(cores...))
}

// NewLogger 基于当前配置创建新的 zap.Logger 实例（公开版本，会添加到全局列表）
func (l *Config) NewLogger() *zap.Logger {
	logger := l.newLogger()

	// 注册到全局 logger 列表中
	mu.Lock()
	allLoggers = append(allLoggers, logger)
	mu.Unlock()

	return logger
}

// Logger 返回默认 logger 实例，若未初始化则返回 nil
func Logger() *zap.Logger {
	return defaultLog
}

// MustLogger 返回默认 logger 实例，若未初始化则 panic
func MustLogger() *zap.Logger {
	if defaultLog == nil {
		panic("logger not initialized, call New first")
	}
	return defaultLog
}

// createLevelEnablerFunc 根据字符串日志级别创建 zap.LevelEnablerFunc
func createLevelEnablerFunc(input string) zap.LevelEnablerFunc {
	if input == "" {
		return nil
	}
	var lv zapcore.Level
	if err := lv.UnmarshalText([]byte(input)); err != nil {
		return nil
	}
	return func(lev zapcore.Level) bool {
		return lev >= lv
	}
}

// createEncoder 创建日志编码器：支持 JSON 或控制台格式
func createEncoder(format string, isConsole bool) zapcore.Encoder {
	var cfg zapcore.EncoderConfig
	if isConsole {
		cfg = zap.NewDevelopmentEncoderConfig()
	} else {
		cfg = zap.NewProductionEncoderConfig()
	}
	cfg.EncodeTime = timeEncoder
	cfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	cfg.EncodeDuration = zapcore.SecondsDurationEncoder
	cfg.EncodeCaller = zapcore.ShortCallerEncoder

	switch format {
	case "json":
		return zapcore.NewJSONEncoder(cfg)
	default:
		return zapcore.NewConsoleEncoder(cfg)
	}
}

// timeEncoder 时间格式化函数（RFC3339 格式）
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.RFC3339))
}

// 以下是对默认 logger 的简化调用封装

func Info(msg string, fields ...zapcore.Field)  { defaultLog.Info(msg, fields...) }
func Debug(msg string, fields ...zapcore.Field) { defaultLog.Debug(msg, fields...) }
func Warn(msg string, fields ...zapcore.Field)  { defaultLog.Warn(msg, fields...) }
func Error(msg string, fields ...zapcore.Field) { defaultLog.Error(msg, fields...) }
func Panic(msg string, fields ...zapcore.Field) { defaultLog.Panic(msg, fields...) }
func Fatal(msg string, fields ...zapcore.Field) { defaultLog.Fatal(msg, fields...) }

// 支持格式化输出的调用方式

func Infof(format string, args ...any)  { defaultLog.Sugar().Infof(format, args...) }
func Debugf(format string, args ...any) { defaultLog.Sugar().Debugf(format, args...) }
func Warnf(format string, args ...any)  { defaultLog.Sugar().Warnf(format, args...) }
func Errorf(format string, args ...any) { defaultLog.Sugar().Errorf(format, args...) }
func Panicf(format string, args ...any) { defaultLog.Sugar().Panicf(format, args...) }
func Fatalf(format string, args ...any) { defaultLog.Sugar().Fatalf(format, args...) }

// SyncAll 刷新所有日志缓冲
func SyncAll() {
	mu.Lock()
	defer mu.Unlock()
	for _, l := range allLoggers {
		if l != nil {
			_ = l.Sync()
		}
	}
}
