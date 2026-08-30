package logger

import (
	"sync"

	"go.uber.org/zap"
)

var load = sync.OnceValue(func() *zap.SugaredLogger {
	log, _ := zap.NewDevelopment()
	defer log.Sync()
	return log.Sugar()
})

func L() *zap.SugaredLogger {
	return load()
}
