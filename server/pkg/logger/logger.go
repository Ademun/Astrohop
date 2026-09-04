package logger

import (
	"go.uber.org/zap"
)

func GetLogger() *zap.SugaredLogger {
	log, _ := zap.NewDevelopment()
	defer func() {
		if err := log.Sync(); err != nil {
			panic(err)
		}
	}()
	return log.Sugar()
}
