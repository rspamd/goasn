package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
)

func SetupLogger(debug bool) error {
	config := zap.NewProductionConfig()
	if debug {
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}
	var err error
	Logger, err = config.Build()
	return err
}
