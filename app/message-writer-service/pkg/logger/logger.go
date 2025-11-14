package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
)

var (
	log  *zap.Logger
	once sync.Once
)

// Get initializes (once) and returns a global zap.Logger instance
func Get() *zap.Logger {
	once.Do(func() {
		var err error
		// If in production, use production config
		if env := os.Getenv("ENV"); env == "production" {
			// Use production config to display logs in json format for tracing tools
			log, err = zap.NewProduction()
		} else {
			// Otherwise, use development config
			log, err = zap.NewDevelopment()
		}
		if err != nil {
			panic("Failed to initialize zap logger: " + err.Error())
		}
	})
	return log
}
