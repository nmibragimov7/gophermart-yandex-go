package logger

import (
	"log"

	"go.uber.org/zap"
)

func Init() *zap.SugaredLogger {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Printf("failed to build development logger: %s", err.Error())
	}

	return logger.Sugar()
}
