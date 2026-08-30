package logger

import (
	"sync"

	"go.uber.org/zap"
)

var load = sync.OnceValue(func() *zap.SugaredLogger {
	log, _ := zap.NewDevelopment()
	defer func() {
		if err := log.Sync(); err != nil {
			panic(err)
		}
	}()
	return log.Sugar()
})

func L() *zap.SugaredLogger {
	return load()
}
