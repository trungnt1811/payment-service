package logger

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/genefriendway/onchain-handler/pkg/interfaces"
)

type zapLogger struct {
	sugaredLogger *zap.SugaredLogger
	currentLevel  zap.AtomicLevel
}

var (
	instance *zapLogger
	mu       sync.Mutex
)

// newZapLogger initializes a new zap-based logger instance.
// newZapLogger initializes a new zap-based logger instance.
func newZapLogger(level interfaces.Level) *zapLogger {
	// Map interfaces.Level to zapcore.Level
	levelMapping := map[interfaces.Level]zapcore.Level{
		interfaces.DebugLevel: zapcore.DebugLevel,
		interfaces.InfoLevel:  zapcore.InfoLevel,
		interfaces.WarnLevel:  zapcore.WarnLevel,
		interfaces.ErrorLevel: zapcore.ErrorLevel,
		interfaces.FatalLevel: zapcore.FatalLevel,
		interfaces.PanicLevel: zapcore.PanicLevel,
	}

	// Get the zap level or default to InfoLevel
	zapLevel, exists := levelMapping[level]
	if !exists {
		zapLevel = zapcore.InfoLevel
	}

	// Configure the logger
	atomicLevel := zap.NewAtomicLevelAt(zapLevel)
	config := zap.NewProductionConfig()
	config.Level = atomicLevel

	logger, err := config.Build(zap.AddCaller(), zap.AddCallerSkip(1)) // Include caller with a skip for wrapper functions
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err)) // Provide context in the panic message
	}

	return &zapLogger{
		sugaredLogger: logger.Sugar(),
		currentLevel:  atomicLevel,
	}
}

// GetLogger returns the singleton logger instance.
func GetLogger() interfaces.Logger {
	mu.Lock()
	defer mu.Unlock()

	if instance == nil {
		// Initialize with default level if not already set
		instance = newZapLogger(interfaces.InfoLevel)
	}
	return instance
}

// SetLogger replaces the singleton logger instance.
func SetLogger(customLogger interfaces.Logger) {
	mu.Lock()
	defer mu.Unlock()
	instance = customLogger.(*zapLogger)
}

// SetLogLevel sets the log level of the logger dynamically.
func SetLogLevel(level interfaces.Level) {
	mu.Lock()
	defer mu.Unlock()

	if instance == nil {
		instance = newZapLogger(level)
	} else {
		// Update the log level dynamically
		var zapLevel zapcore.Level
		switch level {
		case interfaces.DebugLevel:
			zapLevel = zapcore.DebugLevel
		case interfaces.InfoLevel:
			zapLevel = zapcore.InfoLevel
		case interfaces.WarnLevel:
			zapLevel = zapcore.WarnLevel
		case interfaces.ErrorLevel:
			zapLevel = zapcore.ErrorLevel
		case interfaces.FatalLevel:
			zapLevel = zapcore.FatalLevel
		case interfaces.PanicLevel:
			zapLevel = zapcore.PanicLevel
		default:
			zapLevel = zapcore.InfoLevel
		}
		instance.currentLevel.SetLevel(zapLevel)
	}
}

// Logger Interface Implementation

func (z *zapLogger) SetLogLevel(level interfaces.Level) {
	SetLogLevel(level)
}

func (z *zapLogger) GetLogLevel() interfaces.Level {
	switch z.currentLevel.Level() {
	case zapcore.DebugLevel:
		return interfaces.DebugLevel
	case zapcore.InfoLevel:
		return interfaces.InfoLevel
	case zapcore.WarnLevel:
		return interfaces.WarnLevel
	case zapcore.ErrorLevel:
		return interfaces.ErrorLevel
	case zapcore.FatalLevel:
		return interfaces.FatalLevel
	case zapcore.PanicLevel:
		return interfaces.PanicLevel
	default:
		return interfaces.InfoLevel
	}
}

func (z *zapLogger) Debug(message string) { z.sugaredLogger.Debug(message) }
func (z *zapLogger) Debugf(format string, values ...interface{}) {
	z.sugaredLogger.Debugf(format, values...)
}
func (z *zapLogger) Info(message string) { z.sugaredLogger.Info(message) }
func (z *zapLogger) Infof(format string, values ...interface{}) {
	z.sugaredLogger.Infof(format, values...)
}
func (z *zapLogger) Warn(message string) { z.sugaredLogger.Warn(message) }
func (z *zapLogger) Warnf(format string, values ...interface{}) {
	z.sugaredLogger.Warnf(format, values...)
}
func (z *zapLogger) Error(message string) { z.sugaredLogger.Error(message) }
func (z *zapLogger) Errorf(format string, values ...interface{}) {
	z.sugaredLogger.Errorf(format, values...)
}
func (z *zapLogger) Fatal(message string) { z.sugaredLogger.Fatal(message) }
func (z *zapLogger) Fatalf(format string, values ...interface{}) {
	z.sugaredLogger.Fatalf(format, values...)
}
func (z *zapLogger) Panic(message string) { z.sugaredLogger.Panic(message) }
func (z *zapLogger) Panicf(format string, values ...interface{}) {
	z.sugaredLogger.Panicf(format, values...)
}

// WithFields creates a new logger instance with additional fields.
func (z *zapLogger) WithFields(fields map[string]interface{}) interfaces.Logger {
	// Attach fields to the logger and return a new logger instance
	newLogger := z.sugaredLogger.With(fields)
	return &zapLogger{
		sugaredLogger: newLogger,
		currentLevel:  z.currentLevel,
	}
}
