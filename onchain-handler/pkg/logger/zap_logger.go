package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/genefriendway/onchain-handler/pkg/logger/types"
)

type zapLogger struct {
	sugaredLogger *zap.SugaredLogger
	currentLevel  zap.AtomicLevel
}

var (
	instance *zapLogger
	once     sync.Once
)

// mapLogLevel maps interfaces.Level to zapcore.Level.
func mapLogLevel(level types.Level) zapcore.Level {
	switch level {
	case types.DebugLevel:
		return zapcore.DebugLevel
	case types.InfoLevel:
		return zapcore.InfoLevel
	case types.WarnLevel:
		return zapcore.WarnLevel
	case types.ErrorLevel:
		return zapcore.ErrorLevel
	case types.FatalLevel:
		return zapcore.FatalLevel
	case types.PanicLevel:
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

// customLevelEncoder replaces "level" with "severity".
func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	severityMapping := map[zapcore.Level]string{
		zapcore.DebugLevel: "DEBUG",
		zapcore.InfoLevel:  "INFO",
		zapcore.WarnLevel:  "WARNING",
		zapcore.ErrorLevel: "ERROR",
		zapcore.PanicLevel: "CRITICAL",
		zapcore.FatalLevel: "ALERT",
	}
	enc.AppendString(severityMapping[level])
}

// newZapLogger initializes a new zap-based logger instance.
func newZapLogger(level types.Level) *zapLogger {
	atomicLevel := zap.NewAtomicLevelAt(mapLogLevel(level))

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:      "timestamp",
		LevelKey:     "severity",
		CallerKey:    "caller",
		MessageKey:   "message",
		EncodeLevel:  customLevelEncoder,
		EncodeTime:   zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05Z07:00"),
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.Lock(zapcore.AddSync(os.Stdout)),
		atomicLevel,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &zapLogger{
		sugaredLogger: logger.Sugar(),
		currentLevel:  atomicLevel,
	}
}

// GetLogger returns the singleton logger instance.
func GetLogger() types.Logger {
	once.Do(func() {
		instance = newZapLogger(types.InfoLevel)
	})
	return instance
}

// SetLogger replaces the singleton logger instance.
func SetLogger(customLogger types.Logger) {
	instance = customLogger.(*zapLogger)
}

// SetLogLevel sets the log level dynamically.
func SetLogLevel(level types.Level) {
	if instance != nil {
		instance.currentLevel.SetLevel(mapLogLevel(level))
	}
}

// Logger Interface Implementation
func (z *zapLogger) SetLogLevel(level types.Level) {
	SetLogLevel(level)
}

func (z *zapLogger) GetLogLevel() types.Level {
	return types.Level(z.currentLevel.String())
}

// Individual Log Methods
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
func (z *zapLogger) WithFields(fields map[string]interface{}) types.Logger {
	return &zapLogger{
		sugaredLogger: z.sugaredLogger.With(fields),
		currentLevel:  z.currentLevel,
	}
}
