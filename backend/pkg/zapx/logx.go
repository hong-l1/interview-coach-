package zapx

import "go.uber.org/zap"

type Logger struct {
	zap.Logger
}

func NewLogger(logger *zap.Logger) *Logger {
	return &Logger{*logger}
}

type Log interface {
	Debug(msg string, args ...Field)
	Info(msg string, args ...Field)
	Warn(msg string, args ...Field)
	Error(msg string, args ...Field)
}
type Field struct {
	Key   string
	Value any
}
