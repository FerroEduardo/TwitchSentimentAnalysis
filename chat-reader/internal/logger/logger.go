package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger() *zap.SugaredLogger {
	production := os.Getenv("PRODUCTION")
	var config zap.Config
	if production != "" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logger := zap.Must(config.Build()).Sugar()

	return logger
}
