package logger

import (
	"sync"

	"github.com/genefriendway/onchain-handler/pkg/interfaces"
)

var (
	instance interfaces.Logger
	once     sync.Once
	mu       sync.RWMutex
)

// InitLogger initializes the logger with the specified implementation.
// It uses sync.Once to ensure the logger is initialized only once,
// making it safe for concurrent use.
func InitLogger(l interfaces.Logger) {
	if l == nil {
		panic("cannot initialize logger with nil value")
	}

	once.Do(func() {
		mu.Lock()
		defer mu.Unlock()
		instance = l
	})
}

// GetLogger returns the initialized logger instance.
// If the logger hasn't been initialized, it panics with a descriptive message.
func GetLogger() interfaces.Logger {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		panic("Logger not initialized. Call InitLogger first.")
	}
	return instance
}

// IsLevelEnabled checks if a given log level is enabled based on the current logger level
func IsLevelEnabled(level interfaces.Level) bool {
	if instance == nil {
		return false
	}

	return level >= instance.GetLogLevel()
}

// Reset clears the logger instance for testing purposes
// Should only be used in tests
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	instance = nil
	once = sync.Once{}
}
