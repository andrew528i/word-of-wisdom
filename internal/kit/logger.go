package kit

import (
	"sync"

	"go.uber.org/zap"
)

var (
	once   sync.Once
	Logger *zap.SugaredLogger
)

func init() {
	once.Do(func() {
		l, err := zap.NewProduction()
		if err != nil {
			panic(err)
		}

		Logger = l.Sugar()
	})
}
