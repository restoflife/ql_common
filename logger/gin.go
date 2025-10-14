package logger

import (
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mattn/go-isatty"
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
	notlogged = []string{"/favicon.ico"}
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

// WithConfig 使用指定配置构建 Gin 日志中间件
func WithConfig(log *zap.Logger, conf ConfigGin) gin.HandlerFunc {
	// 设置日志输出目标
	out := conf.Output
	if out == nil {
		out = gin.DefaultWriter
	}

	// 判断输出是否为终端（决定是否彩色输出等）
	isTerm := true
	if w, ok := out.(*os.File); !ok || os.Getenv("TERM") == "dumb" ||
		(!isatty.IsTerminal(w.Fd()) && !isatty.IsCygwinTerminal(w.Fd())) {
		isTerm = false
	}

	// 构建跳过路径的 map
	skip := make(map[string]struct{})
	for _, path := range conf.SkipPaths {
		skip[path] = struct{}{}
	}

	// 返回 Gin 中间件处理函数
	return func(c *gin.Context) {
		start := time.Now()           // 请求开始时间
		path := c.Request.URL.Path    // 请求路径
		raw := c.Request.URL.RawQuery // 请求查询参数
		c.Next()                      // 继续处理请求（执行后续中间件及业务逻辑）

		// 判断是否需要跳过日志
		if _, ok := skip[path]; !ok {
			param := FormatterParams{
				Request:      c.Request,
				isTerm:       isTerm,
				Keys:         c.Keys,
				Latency:      time.Since(start),
				ClientIP:     c.ClientIP(),
				Method:       c.Request.Method,
				StatusCode:   c.Writer.Status(),
				ErrorMessage: c.Errors.ByType(gin.ErrorTypePrivate).String(),
				BodySize:     c.Writer.Size(),
			}

			if raw != "" {
				path = path + "?" + raw
			}
			param.Path = path

			// 没有错误时输出 info 日志
			if len(param.ErrorMessage) == 0 {
				log.Info("[gin]",
					zap.String("path", path),
					zap.Int("code", param.StatusCode),
					zap.String("method", param.Method),
					zap.String("user-agent", c.Request.UserAgent()),
					zap.String("latency", param.Latency.String()),
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
			// 尝试将请求 dump 出来，便于日志排查
			if c.Request != nil {
				rawReq, _ = httputil.DumpRequest(c.Request, true)
			}

			// 捕获 panic 并记录日志
			if err := recover(); err != nil {
				stack = stack[:runtime.Stack(stack, false)]
				logger.Error("[recovery]",
					zap.String("path", c.Request.RequestURI),
					zap.Any("error", err),
					zap.ByteString("request", rawReq),
					zap.String("stack", string(stack)),
				)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		c.Next() // 继续执行后续逻辑
	}
}
