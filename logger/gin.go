package logger

import (
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 使用 sync.Pool 优化堆栈内存的重复申请和释放
var (
	stackPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 64<<10) // 64KB 缓冲区
		},
	}

	// 默认不记录日志的路径
	notlogged = []string{"/favicon.ico", "/health", "/ready"}
)

// ConfigGin 配置 Gin 日志中间件的结构体
type ConfigGin struct {
	Output    io.Writer // 日志输出目标（如文件、stdout）
	SkipPaths []string  // 指定不记录日志的请求路径
}

// FormatterParams 是日志格式化器使用的参数结构体
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
	ResponseTime int64 // 响应时间（毫秒）
}

// GinLogger 创建一个默认的日志中间件（使用 gin.DefaultWriter）
func GinLogger(logger *zap.Logger) gin.HandlerFunc {
	return WithWriter(logger, gin.DefaultWriter, notlogged...)
}

// WithWriter 使用指定 Writer 创建日志中间件
func WithWriter(logger *zap.Logger, out io.Writer, notlogged ...string) gin.HandlerFunc {
	return WithConfig(logger, ConfigGin{
		Output:    out,
		SkipPaths: notlogged,
	})
}

// formatBodySize 格式化响应体大小
func formatBodySize(bytes int) string {
	if bytes < 1024 {
		return "< 1KB"
	}
	if bytes < 1024*1024 {
		return "< 1MB"
	}
	return "> 1MB"
}

// formatLatency 格式化延迟时间
func formatLatency(latency time.Duration) string {
	if latency < time.Millisecond {
		return "< 1ms"
	}
	if latency < time.Second {
		return strconv.FormatInt(latency.Milliseconds(), 10) + "ms"
	}
	return latency.String()
}

// WithConfig 使用指定配置构建 Gin 日志中间件
func WithConfig(log *zap.Logger, conf ConfigGin) gin.HandlerFunc {
	// 设置日志输出目标
	out := conf.Output
	if out == nil {
		out = gin.DefaultWriter
	}

	// 构建跳过路径的 map
	skip := make(map[string]struct{}, len(conf.SkipPaths))
	for _, path := range conf.SkipPaths {
		skip[path] = struct{}{}
	}

	// 返回 Gin 中间件处理函数
	return func(c *gin.Context) {
		start := time.Now()           // 请求开始时间
		path := c.Request.URL.Path    // 请求路径
		raw := c.Request.URL.RawQuery // 请求查询参数

		c.Next() // 继续处理请求（执行后续中间件及业务逻辑）

		// 判断是否需要跳过日志
		if _, ok := skip[path]; !ok {
			latency := time.Since(start)
			statusCode := c.Writer.Status()
			errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

			// 优化：只在需要时拼接查询参数
			fullPath := path
			if raw != "" {
				fullPath = path + "?" + raw
			}

			// 根据状态码选择日志级别
			if errorMessage != "" || statusCode >= http.StatusInternalServerError {
				// 错误日志
				log.Error("[gin]",
					zap.String("path", fullPath),
					zap.Int("code", statusCode),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
					zap.String("user-agent", c.Request.UserAgent()),
					zap.String("latency", formatLatency(latency)),
					zap.String("error", errorMessage),
				)
			} else if statusCode >= http.StatusBadRequest {
				// 警告日志（4xx 错误）
				log.Warn("[gin]",
					zap.String("path", fullPath),
					zap.Int("code", statusCode),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
					zap.String("latency", formatLatency(latency)),
				)
			} else {
				// info 日志（正常请求）
				log.Info("[gin]",
					zap.String("path", fullPath),
					zap.Int("code", statusCode),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
					zap.String("user-agent", c.Request.UserAgent()),
					zap.String("latency", formatLatency(latency)),
					// zap.Int("body_size", c.Writer.Size()),
				)
			}
		}
	}
}

// Recovery 是一个 panic 恢复中间件，避免应用崩溃，并记录堆栈信息
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从池中获取堆栈缓冲
		stack := stackPool.Get().([]byte)
		defer stackPool.Put(stack[:0]) // 使用完后重置并放回池中

		defer func() {
			var rawReq []byte
			// 尝试将请求 dump 出来，便于日志排查（仅在非生产环境或开启调试时）
			if c.Request != nil && os.Getenv("GIN_DEBUG") == "true" {
				rawReq, _ = httputil.DumpRequest(c.Request, false)
			}

			// 捕获 panic 并记录日志
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack = stack[:runtime.Stack(stack, false)]
				logger.Error("[panic recovery]",
					zap.String("path", c.Request.RequestURI),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
					zap.Any("error", err),
					zap.ByteString("request", rawReq),
					zap.String("stack", string(stack)),
				)

				// 返回友好的错误响应
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Internal server error",
				})
			}
		}()

		c.Next() // 继续执行后续逻辑
	}
}
