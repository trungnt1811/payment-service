package log

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

type LogrusLogger struct {
	logger      *logrus.Entry
	level       Level
	serviceName string
	config      interface{}
}

func NewLogrusLogger(output io.Writer, level Level) *LogrusLogger {
	if output == nil {
		output = os.Stdout
	}

	l := logrus.New()
	l.SetOutput(output)
	l.SetLevel(convertLogrusLevel(level))

	return &LogrusLogger{
		logger: l.WithFields(logrus.Fields{}), // Initialize as an Entry with no fields
		level:  level,
	}
}

func (l *LogrusLogger) SetLogLevel(level Level) {
	l.level = level
	l.logger.Logger.SetLevel(convertLogrusLevel(level))
}

func (l *LogrusLogger) GetLogLevel() Level {
	return l.level
}

func (l *LogrusLogger) Panic(message string) {
	l.logger.Panic(message)
}

func (l *LogrusLogger) Panicf(format string, values ...interface{}) {
	l.logger.Panicf(format, values...)
}

func (l *LogrusLogger) Fatal(message string) {
	l.logger.Fatal(message)
}

func (l *LogrusLogger) Fatalf(format string, values ...interface{}) {
	l.logger.Fatalf(format, values...)
}

func (l *LogrusLogger) Error(message string) {
	l.logger.Error(message)
}

func (l *LogrusLogger) Errorf(format string, values ...interface{}) {
	l.logger.Errorf(format, values...)
}

func (l *LogrusLogger) Info(message string) {
	l.logger.Info(message)
}

func (l *LogrusLogger) Infof(format string, values ...interface{}) {
	l.logger.Infof(format, values...)
}

func (l *LogrusLogger) Debug(message string) {
	l.logger.Debug(message)
}

func (l *LogrusLogger) Debugf(format string, values ...interface{}) {
	l.logger.Debugf(format, values...)
}

func (l *LogrusLogger) Warn(message string) {
	l.logger.Warn(message)
}

func (l *LogrusLogger) Warnf(format string, values ...interface{}) {
	l.logger.Warnf(format, values...)
}

func (l *LogrusLogger) WithInterface(key string, value interface{}) Logger {
	return &LogrusLogger{
		logger: l.logger.WithField(key, value),
		level:  l.level,
	}
}

func (l *LogrusLogger) WithFields(fields map[string]interface{}) Logger {
	return &LogrusLogger{
		logger: l.logger.WithFields(fields),
		level:  l.level,
	}
}

func (l *LogrusLogger) SetConfigModeByCode(code string) {
	// Implement logic to set configuration mode by code
}

func (l *LogrusLogger) SetConfig(config interface{}) {
	l.config = config
}

func (l *LogrusLogger) GetConfig() interface{} {
	return l.config
}

func (l *LogrusLogger) SetServiceName(serviceName string) {
	l.serviceName = serviceName
}

func (l *LogrusLogger) GetServiceName() string {
	return l.serviceName
}

func convertLogrusLevel(level Level) logrus.Level {
	switch level {
	case DebugLevel:
		return logrus.DebugLevel
	case InfoLevel:
		return logrus.InfoLevel
	case WarnLevel:
		return logrus.WarnLevel
	case ErrorLevel:
		return logrus.ErrorLevel
	case FatalLevel:
		return logrus.FatalLevel
	case PanicLevel:
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}
