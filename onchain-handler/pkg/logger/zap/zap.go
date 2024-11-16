package zap

import (
	"fmt"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/genefriendway/onchain-handler/pkg/logger"
)

// DefaultZapConfig returns default configuration for Zap logger
func DefaultZapConfig(serviceName string) *ZapLoggerConfig {
	return &ZapLoggerConfig{
		Development:      false,
		Encoding:         "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields:    make(map[string]interface{}),
		TimeKey:          "timestamp",
		LevelKey:         "level",
		MessageKey:       "message",
		StacktraceKey:    "stacktrace",
		CallerKey:        "caller",
		FunctionKey:      "function",
		TimeFormat:       "2006-01-02T15:04:05.000Z0700",
	}
}

func defaultZapConfigDevelop(serviceName string) *ZapLoggerConfig {
	return &ZapLoggerConfig{
		Development:      true,
		Encoding:         "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields: map[string]interface{}{
			"environment": "development",
			"service":     serviceName,
		},
		TimeKey:       "timestamp",
		LevelKey:      "level",
		MessageKey:    "message",
		StacktraceKey: "stacktrace",
		CallerKey:     "caller",
		FunctionKey:   "function",
		TimeFormat:    "2006-01-02T15:04:05.000Z0700",
	}
}

func defaultZapConfigProduction(serviceName string) *ZapLoggerConfig {
	return &ZapLoggerConfig{
		Development:      false,
		Encoding:         "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields: map[string]interface{}{
			"environment": "production",
			"service":     serviceName,
		},
		TimeKey:       "timestamp",
		LevelKey:      "level",
		MessageKey:    "message",
		StacktraceKey: "stacktrace",
		CallerKey:     "caller",
		FunctionKey:   "function",
		TimeFormat:    "2006-01-02T15:04:05.000Z0700",
	}
}

func convertZapLevel(level logger.Level) zapcore.Level {
	switch level {
	case logger.DebugLevel:
		return zapcore.DebugLevel
	case logger.InfoLevel:
		return zapcore.InfoLevel
	case logger.WarnLevel:
		return zapcore.WarnLevel
	case logger.ErrorLevel:
		return zapcore.ErrorLevel
	case logger.FatalLevel:
		return zapcore.FatalLevel
	case logger.PanicLevel:
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

// ZapLogger implements the Logger interface using Uber's zap logging library
type ZapLogger struct {
	logger      *zap.SugaredLogger
	atomicLevel zap.AtomicLevel // Added to store zap's atomic level
	level       atomic.Value    // stores Level
	config      atomic.Value    // stores *ZapLoggerConfig
	serviceName atomic.Value    // stores string
}

// ZapLoggerConfig holds configuration for the ZapLogger
type ZapLoggerConfig struct {
	Development      bool
	Encoding         string // "json" or "console"
	OutputPaths      []string
	ErrorOutputPaths []string
	InitialFields    map[string]interface{}
	TimeKey          string
	LevelKey         string
	MessageKey       string
	StacktraceKey    string
	CallerKey        string
	FunctionKey      string
	TimeFormat       string
}

// ZapLoggerFactory implements LoggerFactory for creating ZapLogger instances
type ZapLoggerFactory struct{}

func (f *ZapLoggerFactory) CreateLogger(config interface{}) (logger.Logger, error) {
	if config == nil {
		return NewZapLogger(logger.InfoLevel)
	}

	zapConfig, ok := config.(*ZapLoggerConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected *ZapLoggerConfig, got %T", config)
	}

	return NewZapLoggerWithConfig(logger.InfoLevel, zapConfig)
}

// NewZapLogger creates a new ZapLogger with default configuration
func NewZapLogger(level logger.Level) (logger.Logger, error) {
	return NewZapLoggerWithConfig(level, nil)
}

// NewZapLoggerWithConfig creates a new ZapLogger with the specified configuration
func NewZapLoggerWithConfig(level logger.Level, config *ZapLoggerConfig) (*ZapLogger, error) {
	if config == nil {
		config = DefaultZapConfig("")
	}

	encoderConfig := createEncoderConfig(config)

	// Create atomic level
	atomicLevel := zap.NewAtomicLevelAt(convertZapLevel(level))

	zapConfig := zap.Config{
		Level:            atomicLevel,
		Development:      config.Development,
		Encoding:         config.Encoding,
		EncoderConfig:    encoderConfig,
		OutputPaths:      config.OutputPaths,
		ErrorOutputPaths: config.ErrorOutputPaths,
		InitialFields:    config.InitialFields,
	}

	baseLogger, err := zapConfig.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	logger := &ZapLogger{
		logger:      baseLogger.Sugar(),
		atomicLevel: atomicLevel,
	}
	logger.setLevel(level)
	logger.setConfig(config)

	return logger, nil
}

func createEncoderConfig(config *ZapLoggerConfig) zapcore.EncoderConfig {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        config.TimeKey,
		LevelKey:       config.LevelKey,
		NameKey:        "logger",
		CallerKey:      config.CallerKey,
		FunctionKey:    config.FunctionKey,
		MessageKey:     config.MessageKey,
		StacktraceKey:  config.StacktraceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if config.Development {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	}

	return encoderConfig
}

// Thread-safe atomic operations
func (l *ZapLogger) setLevel(level logger.Level) {
	l.level.Store(level)
	// Update the underlying zap logger's level
	if l.atomicLevel != (zap.AtomicLevel{}) {
		l.atomicLevel.SetLevel(convertZapLevel(level))
	}
}

func (l *ZapLogger) getLevel() logger.Level {
	return l.level.Load().(logger.Level)
}

func (l *ZapLogger) setConfig(config *ZapLoggerConfig) {
	l.config.Store(config)
}

func (l *ZapLogger) IsLevelEnabled(level logger.Level) bool {
	// Add nil check for atomicLevel
	if l.atomicLevel == (zap.AtomicLevel{}) {
		// If atomicLevel is not set, fall back to comparing with stored level
		return level >= l.getLevel()
	}
	return l.atomicLevel.Enabled(convertZapLevel(level))
}

func (l *ZapLogger) getConfig() *ZapLoggerConfig {
	return l.config.Load().(*ZapLoggerConfig)
}

// Logger interface implementation
func (l *ZapLogger) SetLogLevel(level logger.Level) {
	l.setLevel(level)
}

func (l *ZapLogger) GetLogLevel() logger.Level {
	return l.getLevel()
}

func (l *ZapLogger) Debug(message string) {
	if l.IsLevelEnabled(logger.DebugLevel) {
		l.logger.Debug(message)
	}
}

func (l *ZapLogger) Debugf(format string, values ...interface{}) {
	if l.IsLevelEnabled(logger.DebugLevel) {
		l.logger.Debugf(format, values...)
	}
}

func (l *ZapLogger) Info(message string) {
	if l.IsLevelEnabled(logger.DebugLevel) {
		l.logger.Info(message)
	}
}

func (l *ZapLogger) Infof(format string, values ...interface{}) {
	if l.IsLevelEnabled(logger.DebugLevel) {
		l.logger.Infof(format, values...)
	}
}

func (l *ZapLogger) Warn(message string) {
	if l.IsLevelEnabled(logger.WarnLevel) {
		l.logger.Warn(message)
	}
}

func (l *ZapLogger) Warnf(format string, values ...interface{}) {
	if l.IsLevelEnabled(logger.WarnLevel) {
		l.logger.Warnf(format, values...)
	}
}

func (l *ZapLogger) Error(message string) {
	if l.IsLevelEnabled(logger.ErrorLevel) {
		l.logger.Error(message)
	}
}

func (l *ZapLogger) Errorf(format string, values ...interface{}) {
	if l.IsLevelEnabled(logger.ErrorLevel) {
		l.logger.Errorf(format, values...)
	}
}

func (l *ZapLogger) Fatal(message string) {
	if l.IsLevelEnabled(logger.FatalLevel) {
		l.logger.Fatal(message)
	}
}

func (l *ZapLogger) Fatalf(format string, values ...interface{}) {
	if l.IsLevelEnabled(logger.FatalLevel) {
		l.logger.Fatalf(format, values...)
	}
}

func (l *ZapLogger) Panic(message string) {
	if l.IsLevelEnabled(logger.PanicLevel) {
		l.logger.Panic(message)
	}
}

func (l *ZapLogger) Panicf(format string, values ...interface{}) {
	if l.IsLevelEnabled(logger.PanicLevel) {
		l.logger.Panicf(format, values...)
	}
}

// WithInterface creates a new logger with the added field while preserving atomicLevel
func (l *ZapLogger) WithInterface(key string, value interface{}) logger.Logger {
	return &ZapLogger{
		logger:      l.logger.With(key, value),
		atomicLevel: l.atomicLevel,
		level:       l.level,
		config:      l.config,
		serviceName: l.serviceName,
	}
}

// WithFields creates a new logger with the added fields while preserving atomicLevel
func (l *ZapLogger) WithFields(fields map[string]interface{}) logger.Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &ZapLogger{
		logger:      l.logger.With(args...),
		atomicLevel: l.atomicLevel,
		level:       l.level,
		config:      l.config,
		serviceName: l.serviceName,
	}
}

func (l *ZapLogger) SetConfigModeByCode(code string) {
	var config *ZapLoggerConfig
	serviceName := l.GetServiceName()

	switch code {
	case logger.DEVELOPMENT_ENVIRONMENT_CODE_MODE:
		config = defaultZapConfigDevelop(serviceName)
	case logger.PRODUCTION_ENVIRONMENT_CODE_MODE:
		config = defaultZapConfigProduction(serviceName)
	default:
		config = DefaultZapConfig(serviceName)
	}

	l.setConfig(config)
}

func (l *ZapLogger) SetConfig(config interface{}) {
	if zapConfig, ok := config.(*ZapLoggerConfig); ok {
		l.setConfig(zapConfig)
	}
}

func (l *ZapLogger) GetConfig() interface{} {
	return l.getConfig()
}

func (l *ZapLogger) SetServiceName(serviceName string) {
	l.serviceName.Store(serviceName)
}

func (l *ZapLogger) GetServiceName() string {
	if name := l.serviceName.Load(); name != nil {
		return name.(string)
	}
	return ""
}

// Cleanup safely closes the logger and releases resources
func (l *ZapLogger) Cleanup() error {
	if l.logger != nil {
		return l.logger.Sync()
	}
	return nil
}
