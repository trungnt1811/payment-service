package types

// Level defines the severity level for logging.
type Level string

const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
	FatalLevel Level = "fatal"
	PanicLevel Level = "panic"
)

// Logger interface defines common logging operations.
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
	WithFields(fields map[string]interface{}) Logger
}
