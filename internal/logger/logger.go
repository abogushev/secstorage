package logger

import (
	"log"

	"go.uber.org/zap"
)

var Log *zap.Logger

func init() {
	cfg := zap.NewProductionConfig()
	cfg.DisableStacktrace = true
	cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	logger, err := cfg.Build()
	if err != nil {
		log.Fatal("cannot init zap", err)
	}
	Log = logger
}
