package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var zapLog *zap.Logger

func init() {
	logLevel := os.Getenv("O2PN3_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	// setup logger
	zapLevel, err := zap.ParseAtomicLevel(logLevel)
	if err != nil {
		fmt.Println("invalid O2PN3_LOG_LEVEL value\n", err)
		os.Exit(1)
	}

	cfg := zap.Config{
		Encoding:         "console",
		Level:            zapLevel,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "message",
			LevelKey:     "level",
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			TimeKey:      "time",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	zapLog, err = cfg.Build()
	if err != nil {
		fmt.Println("fatal error starting logger\n", err)
		os.Exit(1)
	}
}

func Info(message string, fields ...zap.Field) {
	zapLog.Info(message, fields...)
}

func Debug(message string, fields ...zap.Field) {
	zapLog.Debug(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	zapLog.Error(message, fields...)
}

func Fatal(message string, fields ...zap.Field) {
	zapLog.Fatal(message, fields...)
}

func Panic(message string, fields ...zap.Field) {
	zapLog.Panic(message, fields...)
}
