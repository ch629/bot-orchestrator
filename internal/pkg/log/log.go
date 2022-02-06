package log

import (
	"fmt"

	"go.uber.org/zap"
)

var log *zap.Logger

func init() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("failed to setup logger: %s", err))
	}
	log = logger
	zap.ReplaceGlobals(logger)
}

func Info(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	log.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	log.Fatal(msg, fields...)
}

func With(fields ...zap.Field) *zap.Logger {
	return log.With(fields...)
}
