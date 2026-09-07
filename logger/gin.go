package logger

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	stackPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 64<<10)
		},
	}

	notlogged = []string{"/favicon.ico", "/health", "/ready"}
)

// ConfigGin provides the corresponding package operation.
type ConfigGin struct {
	Output    io.Writer
	SkipPaths []string
}

// FormatterParams provides the corresponding package operation.
type FormatterParams struct {
	Request      *http.Request
	StatusCode   int
	Latency      time.Duration
	ClientIP     string
	Method       string
	Path         string
	ErrorMessage string
	isTerm       bool
	BodySize     int
	Keys         map[any]any
	RequestID    string
	ResponseTime int64
}

// GinLogger uses the supplied Zap logger and omits query strings.
func GinLogger(logger *zap.Logger) gin.HandlerFunc {
	return WithConfig(logger, ConfigGin{SkipPaths: notlogged})
}

func ginErrorStack(errors []*gin.Error) string {
	var stacks []string
	for _, item := range errors {
		if stacktrace := errorStack(item.Err); stacktrace != "" {
			stacks = append(stacks, stacktrace)
		}
	}
	return strings.Join(stacks, "\n")
}

func ginErrorMessage(errors []*gin.Error) string {
	var message strings.Builder
	for i, item := range errors {
		fmt.Fprintf(&message, "Error #%02d: %s\n", i+1, item.Err)
	}
	return message.String()
}

// WithWriter provides the corresponding package operation.
func WithWriter(logger *zap.Logger, out io.Writer, notlogged ...string) gin.HandlerFunc {
	return WithConfig(logger, ConfigGin{
		Output:    out,
		SkipPaths: notlogged,
	})
}

// formatBodySize provides the corresponding package operation.
func formatBodySize(bytes int) string {
	if bytes < 1024 {
		return "< 1KB"
	}
	if bytes < 1024*1024 {
		return "< 1MB"
	}
	return "> 1MB"
}

// formatLatency provides the corresponding package operation.
func formatLatency(latency time.Duration) string {
	if latency < time.Millisecond {
		return "< 1ms"
	}
	if latency < time.Second {
		return strconv.FormatInt(latency.Milliseconds(), 10) + "ms"
	}
	return latency.String()
}

// WithConfig provides the corresponding package operation.
func WithConfig(log *zap.Logger, conf ConfigGin) gin.HandlerFunc {
	if log == nil {
		log = current()
	}
	log = log.WithOptions(zap.WithCaller(false))
	// An explicit writer replaces the destination; a nil writer preserves Zap.
	if conf.Output != nil {
		log = log.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewCore(createEncoder("json", false), zapcore.Lock(zapcore.AddSync(conf.Output)), core)
		}))
	}

	skip := make(map[string]struct{}, len(conf.SkipPaths))
	for _, path := range conf.SkipPaths {
		skip[path] = struct{}{}
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		if _, ok := skip[path]; !ok {
			latency := time.Since(start)
			statusCode := c.Writer.Status()
			privateErrors := c.Errors.ByType(gin.ErrorTypePrivate)
			errorMessage := ginErrorMessage(privateErrors)
			stacktrace := ginErrorStack(privateErrors)

			// Query strings are deliberately excluded from access logs.
			fullPath := path

			if errorMessage != "" || statusCode >= http.StatusInternalServerError {
				fields := []zap.Field{
					zap.String("path", fullPath),
					zap.Int("code", statusCode),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
					zap.String("latency", formatLatency(latency)),
					zap.String("error", errorMessage),
				}
				if stacktrace != "" {
					fields = append(fields, zap.String("stacktrace", stacktrace))
				}
				log.WithOptions(zap.AddStacktrace(zapcore.FatalLevel+1)).Error("[gin]", fields...)
			} else if statusCode >= http.StatusBadRequest {
				log.Warn("[gin]",
					zap.String("path", fullPath),
					zap.Int("code", statusCode),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
					zap.String("latency", formatLatency(latency)),
				)
			} else {
				// info provides the corresponding package operation.
				log.Info("[gin]",
					zap.String("path", fullPath),
					zap.Int("code", statusCode),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
					zap.String("latency", formatLatency(latency)),
					// zap.Int("body_size", c.Writer.Size()),
				)
			}
		}
	}
}

// Recovery provides the corresponding package operation.
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = current()
	}
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Restore length before reuse; only allocate a stack buffer on panic.
				stack := stackPool.Get().([]byte)
				stack = stack[:cap(stack)]
				defer stackPool.Put(stack)
				stack = stack[:runtime.Stack(stack, false)]
				logger.WithOptions(zap.AddStacktrace(zapcore.FatalLevel+1)).Error("[panic recovery]",
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
					zap.Any("error", err),
					zap.String("stacktrace", string(stack)),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Internal server error",
				})
			}
		}()

		c.Next()
	}
}
