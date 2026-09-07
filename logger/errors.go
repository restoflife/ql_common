package logger

import "errors"

var (
	ErrInvalidFileLevel    = errors.New("invalid file log level")
	ErrInvalidConsoleLevel = errors.New("invalid console log level")
	ErrInvalidFormat       = errors.New("invalid log format")
	ErrInvalidCallerSkip   = errors.New("caller skip must be non-negative")
	ErrInvalidRotation     = errors.New("log rotation values must be non-negative")
)
