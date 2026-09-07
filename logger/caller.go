package logger

import (
	"go.uber.org/zap/zapcore"
)

// errorCallerCore removes caller metadata from entries below ErrorLevel.
type errorCallerCore struct {
	zapcore.Core
}

func (c errorCallerCore) With(fields []zapcore.Field) zapcore.Core {
	return errorCallerCore{Core: c.Core.With(fields)}
}

func (c errorCallerCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if !c.Enabled(entry.Level) {
		return checked
	}
	return checked.AddCore(entry, c)
}

func (c errorCallerCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	if entry.Level < zapcore.ErrorLevel {
		entry.Caller = zapcore.EntryCaller{}
	}
	return c.Core.Write(entry, fields)
}
