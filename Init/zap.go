package Init

import (
	"awesomeProject4/pkg/zapx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger() zapx.Log {
	jack := &lumberjack.Logger{
		Filename:   "./logs/log.log",
		MaxSize:    50,
		MaxBackups: 3,
		MaxAge:     7,
	}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig()), zapcore.AddSync(jack), zapcore.DebugLevel)
	l := zap.New(core, zap.AddCaller())
	return zapx.NewLogger(l)
}
