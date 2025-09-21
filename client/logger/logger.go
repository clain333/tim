package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var (
	zapLogger *zap.Logger
)

func InitLogger(logFile string) error {
	fileWriter, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(fileWriter),
		zapcore.ErrorLevel,
	)

	zapLogger = zap.New(core, zap.AddCaller())
	return nil
}

func Error(msg string, fields ...zap.Field) {
	if zapLogger != nil {
		zapLogger.Error(msg, fields...)
	}
}

func Sync() {
	if zapLogger != nil {
		_ = zapLogger.Sync()
	}
}
