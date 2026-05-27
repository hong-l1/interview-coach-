package zapx

import "go.uber.org/zap"

func (l *Logger) Debug(msg string, args ...Field) {
	l.Logger.Debug(msg, ToZaoField(args)...)
}

func (l *Logger) Info(msg string, args ...Field) {
	l.Logger.Info(msg, ToZaoField(args)...)
}

func (l *Logger) Warn(msg string, args ...Field) {
	l.Logger.Warn(msg, ToZaoField(args)...)
}

func (l *Logger) Error(msg string, args ...Field) {
	l.Logger.Error(msg, ToZaoField(args)...)
}

func ToZaoField(args []Field) []zap.Field {
	ans := make([]zap.Field, 0, len(args))
	for k := range args {
		ans = append(ans, zap.Any(args[k].Key, args[k].Value))
	}
	return ans
}
func String(key, val string) Field {
	return Field{Key: key, Value: val}
}
func Error(err error) Field {
	return Field{Key: "error", Value: err}
}
func Int64(key string, val int64) Field {
	return Field{Key: key, Value: val}
}
func Int32(key string, val int32) Field {
	return Field{Key: key, Value: val}
}
