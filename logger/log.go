package logger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config provides the corresponding package operation.
type Config struct {
	Level      string `json:"level"`
	Filename   string `json:"file"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
	Console    string `json:"console"`
	Format     string `json:"format"`
	CallerSkip int    `json:"caller_skip"`
}

const (
	DefaultLevel   = "info"
	DefaultMaxSize = 10 // MB
	DefaultFormat  = "json"
)

var (
	lifecycle         sync.Mutex
	mu                sync.Mutex
	allLoggers        []*zap.Logger
	defaultLog        atomic.Pointer[zap.Logger]
	defaultCallerSkip atomic.Int32
	logClosers        []io.Closer
	nopLog            = zap.NewNop()
)

// GetAll provides the corresponding package operation.
func GetAll() []*zap.Logger {
	mu.Lock()
	defer mu.Unlock()
	result := make([]*zap.Logger, len(allLoggers))
	copy(result, allLoggers)
	return result
}

// New provides the corresponding package operation.
// New initializes the default logger, panicking on invalid configuration.
// Prefer Init when configuration comes from users or files.
// Deprecated: Use Init.
func New(g *Config) {
	if err := Init(g); err != nil {
		panic(err)
	}
}

// Init validates a copy of the config and safely publishes the default logger.
func Init(g *Config) error {
	lifecycle.Lock()
	defer lifecycle.Unlock()
	l, err := g.Build()
	if err != nil {
		return err
	}
	defaultLog.Store(l)
	var callerSkip int
	if g != nil {
		callerSkip = g.CallerSkip
	}
	defaultCallerSkip.Store(int32(callerSkip))
	return nil
}

func normalized(g *Config) (Config, error) {
	var c Config
	if g != nil {
		c = *g
	}
	if c.Level == "" {
		c.Level = DefaultLevel
	}
	if c.MaxSize == 0 {
		c.MaxSize = DefaultMaxSize
	}
	if c.Format == "" {
		c.Format = DefaultFormat
	}
	if c.Filename == "" && c.Console == "" {
		c.Console = c.Level
	}
	if createLevelEnablerFunc(c.Level) == nil {
		return c, fmt.Errorf("%w: %q", ErrInvalidFileLevel, c.Level)
	}
	if c.Console != "" && createLevelEnablerFunc(c.Console) == nil {
		return c, fmt.Errorf("%w: %q", ErrInvalidConsoleLevel, c.Console)
	}
	if c.Format != "json" && c.Format != "text" {
		return c, fmt.Errorf("%w: %q", ErrInvalidFormat, c.Format)
	}
	if c.CallerSkip < 0 {
		return c, ErrInvalidCallerSkip
	}
	if c.MaxSize < 0 || c.MaxBackups < 0 || c.MaxAge < 0 {
		return c, ErrInvalidRotation
	}
	return c, nil
}

// Build constructs and registers a logger. An empty filename disables file output.
func (g *Config) Build() (*zap.Logger, error) {
	c, err := normalized(g)
	if err != nil {
		return nil, err
	}
	mu.Lock()
	defer mu.Unlock()
	l := c.newLogger()
	allLoggers = append(allLoggers, l)
	return l, nil
}

// newLogger builds a logger from a normalized configuration.
func (l *Config) newLogger() *zap.Logger {
	encoder := createEncoder(l.Format, false)
	consoleEncoder := createEncoder("text", true)

	cores := make([]zapcore.Core, 0, 2)

	if l.Filename != "" {
		writer := &lumberjack.Logger{Filename: l.Filename, MaxSize: l.MaxSize, MaxBackups: l.MaxBackups, MaxAge: l.MaxAge, LocalTime: true}
		logClosers = append(logClosers, writer)
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(writer), createLevelEnablerFunc(l.Level)))
	}

	if consoleLevel := createLevelEnablerFunc(l.Console); consoleLevel != nil {
		cores = append(cores,
			zapcore.NewCore(
				consoleEncoder,
				zapcore.Lock(os.Stderr),
				consoleLevel,
			),
		)
	}

	core := errorCallerCore{Core: zapcore.NewTee(cores...)}
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
}

// NewLogger provides the corresponding package operation.
// Deprecated: Use Config.Build.
func (l *Config) NewLogger() *zap.Logger {
	result, err := l.Build()
	if err != nil {
		panic(err)
	}
	return result
}

// Logger provides the corresponding package operation.
func Logger() *zap.Logger {
	return defaultLog.Load()
}

// MustLogger provides the corresponding package operation.
func MustLogger() *zap.Logger {
	if l := defaultLog.Load(); l != nil {
		return l
	}
	panic("logger not initialized, call Init or New first")
}

func current() *zap.Logger {
	if l := defaultLog.Load(); l != nil {
		return l
	}
	return nopLog
}

// createLevelEnablerFunc provides the corresponding package operation.
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

// createEncoder provides the corresponding package operation.
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

// timeEncoder provides the corresponding package operation.
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.RFC3339))
}

func callerLogger() *zap.Logger {
	skip := int(defaultCallerSkip.Load())
	if skip == 0 {
		return current()
	}
	return current().WithOptions(zap.AddCallerSkip(skip))
}

func Info(msg string, fields ...zapcore.Field)  { callerLogger().Info(msg, fields...) }
func Debug(msg string, fields ...zapcore.Field) { callerLogger().Debug(msg, fields...) }
func Warn(msg string, fields ...zapcore.Field)  { callerLogger().Warn(msg, fields...) }
func Error(msg string, fields ...zapcore.Field) { callerLogger().Error(msg, fields...) }
func Panic(msg string, fields ...zapcore.Field) { callerLogger().Panic(msg, fields...) }
func Fatal(msg string, fields ...zapcore.Field) { callerLogger().Fatal(msg, fields...) }

func Infof(format string, args ...any)  { callerLogger().Sugar().Infof(format, args...) }
func Debugf(format string, args ...any) { callerLogger().Sugar().Debugf(format, args...) }
func Warnf(format string, args ...any)  { callerLogger().Sugar().Warnf(format, args...) }
func Errorf(format string, args ...any) { callerLogger().Sugar().Errorf(format, args...) }
func Panicf(format string, args ...any) { callerLogger().Sugar().Panicf(format, args...) }
func Fatalf(format string, args ...any) { callerLogger().Sugar().Fatalf(format, args...) }

// SyncAll provides the corresponding package operation.
func SyncAll() {
	mu.Lock()
	defer mu.Unlock()
	for _, l := range allLoggers {
		if l != nil {
			_ = l.Sync()
		}
	}
}

// CloseAll flushes registered loggers and closes rotated files. Drain log producers
// first; do not reuse returned logger pointers after shutdown.
func CloseAll() error {
	lifecycle.Lock()
	defer lifecycle.Unlock()
	mu.Lock()
	defer mu.Unlock()
	defaultLog.Store(nil)
	defaultCallerSkip.Store(0)
	var result error
	for _, l := range allLoggers {
		result = errors.Join(result, l.Sync())
	}
	for _, closer := range logClosers {
		result = errors.Join(result, closer.Close())
	}
	allLoggers = nil
	logClosers = nil
	return result
}
