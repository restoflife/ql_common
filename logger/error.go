package logger

import (
	"runtime"
	"strconv"
	"strings"
)

type tracedError struct {
	error
	stacktrace string
}

func (e tracedError) Unwrap() error      { return e.error }
func (e tracedError) StackTrace() string { return e.stacktrace }

type stackError interface {
	error
	StackTrace() string
}

// WithStack provides the corresponding package operation.
func WithStack(err error) error { return WithStackSkip(err, 0) }

// WithStackSkip provides the corresponding package operation.
func WithStackSkip(err error, skip int) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(stackError); ok {
		return err
	}
	if skip < 0 {
		skip = 0
	}
	return tracedError{error: err, stacktrace: captureStack(3+skip, 12)}
}

func errorStack(err error) string {
	for err != nil {
		if traced, ok := err.(stackError); ok {
			return traced.StackTrace()
		}
		type unwrapper interface{ Unwrap() error }
		wrapped, ok := err.(unwrapper)
		if !ok {
			break
		}
		err = wrapped.Unwrap()
	}
	return ""
}

func captureStack(skip, limit int) string {
	pcs := make([]uintptr, limit+8)
	n := runtime.Callers(skip, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	var stack strings.Builder
	count := 0
	for count < limit {
		frame, more := frames.Next()
		if strings.HasPrefix(frame.Function, "github.com/gin-gonic/gin.") {
			break
		}
		if frame.Function != "" {
			stack.WriteString(frame.Function)
			stack.WriteString("\n\t")
			stack.WriteString(frame.File)
			stack.WriteByte(':')
			stack.WriteString(strconv.Itoa(frame.Line))
			stack.WriteByte('\n')
			count++
		}
		if !more {
			break
		}
	}
	return strings.TrimSuffix(stack.String(), "\n")
}
