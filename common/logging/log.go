package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"sync"
)

var logger *zap.Logger
var logInitOnce sync.Once

func cvtZapLevel(level int) zapcore.Level {
	switch level {
	case 0:
		return zap.DebugLevel

	case 1:
		return zap.InfoLevel

	case 2:
		return zap.WarnLevel

	case 3:
		return zap.ErrorLevel

	default:
		return zap.InfoLevel
	}
}

func LogInit(filename string, maxsize, maxBackups, maxAge, level int) {
	logInitOnce.Do(func() {
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   filename,
			MaxSize:    maxsize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
		})

		zapCfg := zap.NewDevelopmentEncoderConfig()
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(zapCfg),
			w,
			cvtZapLevel(level),
		)
		logger = zap.New(core)
	})
}

func Info(msg string, field ...zap.Field) {
	logger.Info(msg, field...)
}

func Debug(msg string, field ...zap.Field) {
	logger.Debug(msg, field...)
}

func Warn(msg string, field ...zap.Field) {
	logger.Warn(msg, field...)
}

func Error(msg string, field ...zap.Field) {
	logger.Error(msg, field...)
}
