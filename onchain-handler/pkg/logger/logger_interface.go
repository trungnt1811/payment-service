package logger

import (
	"fmt"
	"sync"
)

// Environment mode constants
const (
	DEVELOPMENT_ENVIRONMENT_CODE_MODE = "development"
	PRODUCTION_ENVIRONMENT_CODE_MODE  = "production"
)

// Level represents the logging level
type Level int8

const (
	// Log levels ordered by increasing severity
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

// String returns the string representation of the log level
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	case PanicLevel:
		return "PANIC"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", l)
	}
}

// Logger interface defines common logging operations
type Logger interface {
	// Level management
	SetLogLevel(level Level)
	GetLogLevel() Level

	// Basic logging methods
	Debug(message string)
	Debugf(format string, values ...interface{})
	Info(message string)
	Infof(format string, values ...interface{})
	Warn(message string)
	Warnf(format string, values ...interface{})
	Error(message string)
	Errorf(format string, values ...interface{})
	Fatal(message string)
	Fatalf(format string, values ...interface{})
	Panic(message string)
	Panicf(format string, values ...interface{})

	// Contextual logging
	WithInterface(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger

	// Configuration management
	SetConfigModeByCode(code string)
	SetConfig(config interface{})
	GetConfig() interface{}

	// Service information
	SetServiceName(serviceName string)
	GetServiceName() string
}

// LoggerFactory defines the interface for creating logger instances
type LoggerFactory interface {
	CreateLogger(config interface{}) (Logger, error)
}

var (
	instance Logger
	once     sync.Once
	mu       sync.RWMutex
)

// InitLogger initializes the logger with the specified implementation.
// It uses sync.Once to ensure the logger is initialized only once,
// making it safe for concurrent use.
func InitLogger(l Logger) {
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
func GetLogger() Logger {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		panic("Logger not initialized. Call InitLogger first.")
	}
	return instance
}

// IsLevelEnabled checks if a given log level is enabled based on the current logger level
func IsLevelEnabled(level Level) bool {
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
