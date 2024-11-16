package zerolog

import (
	"io"
	"os"

	"github.com/rs/zerolog"

	"github.com/genefriendway/onchain-handler/pkg/logger"
)

type ZerologLogger struct {
	logger      *zerolog.Logger
	level       logger.Level
	serviceName string
	config      interface{}
}

func NewZerologLogger(output io.Writer, level logger.Level, useColors bool) *ZerologLogger {
	if output == nil {
		output = os.Stdout
	}

	if useColors {
		output = zerolog.ConsoleWriter{Out: output, TimeFormat: "2006-01-02 15:04:05"}
	}

	zl := zerolog.New(output).With().Timestamp().Logger()
	logger := &ZerologLogger{
		logger: &zl,
		level:  level,
	}
	logger.SetLogLevel(level)
	return logger
}

func (l *ZerologLogger) SetLogLevel(level logger.Level) {
	l.level = level
	zerolog.SetGlobalLevel(convertLevel(level))
}

func (l *ZerologLogger) GetLogLevel() logger.Level {
	return l.level
}

func (l *ZerologLogger) Panic(message string) {
	l.logger.Panic().Msg(message)
}

func (l *ZerologLogger) Panicf(format string, values ...interface{}) {
	l.logger.Panic().Msgf(format, values...)
}

func (l *ZerologLogger) Fatal(message string) {
	l.logger.Fatal().Msg(message)
}

func (l *ZerologLogger) Fatalf(format string, values ...interface{}) {
	l.logger.Fatal().Msgf(format, values...)
}

func (l *ZerologLogger) Error(message string) {
	l.logger.Error().Msg(message)
}

func (l *ZerologLogger) Errorf(format string, values ...interface{}) {
	l.logger.Error().Msgf(format, values...)
}

func (l *ZerologLogger) Info(message string) {
	l.logger.Info().Msg(message)
}

func (l *ZerologLogger) Infof(format string, values ...interface{}) {
	l.logger.Info().Msgf(format, values...)
}

func (l *ZerologLogger) Debug(message string) {
	l.logger.Debug().Msg(message)
}

func (l *ZerologLogger) Debugf(format string, values ...interface{}) {
	l.logger.Debug().Msgf(format, values...)
}

func (l *ZerologLogger) Warn(message string) {
	l.logger.Warn().Msg(message)
}

func (l *ZerologLogger) Warnf(format string, values ...interface{}) {
	l.logger.Warn().Msgf(format, values...)
}

func (l *ZerologLogger) WithInterface(key string, value interface{}) logger.Logger {
	newLogger := l.logger.With().Interface(key, value).Logger()
	return &ZerologLogger{logger: &newLogger, level: l.level, serviceName: l.serviceName, config: l.config}
}

func (l *ZerologLogger) WithFields(fields map[string]interface{}) logger.Logger {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	newLogger := ctx.Logger()
	return &ZerologLogger{logger: &newLogger, level: l.level, serviceName: l.serviceName, config: l.config}
}

func (l *ZerologLogger) SetConfigModeByCode(code string) {
	// Implement the logic to set the configuration mode by code
	// This is a placeholder implementation
	if code == logger.DEVELOPMENT_ENVIRONMENT_CODE_MODE {
		l.SetLogLevel(logger.DebugLevel)
	} else if code == logger.PRODUCTION_ENVIRONMENT_CODE_MODE {
		l.SetLogLevel(logger.InfoLevel)
	}
}

func (l *ZerologLogger) SetConfig(config interface{}) {
	l.config = config
}

func (l *ZerologLogger) GetConfig() interface{} {
	return l.config
}

func (l *ZerologLogger) SetServiceName(serviceName string) {
	l.serviceName = serviceName
}

func (l *ZerologLogger) GetServiceName() string {
	return l.serviceName
}

func convertLevel(level logger.Level) zerolog.Level {
	switch level {
	case logger.DebugLevel:
		return zerolog.DebugLevel
	case logger.InfoLevel:
		return zerolog.InfoLevel
	case logger.WarnLevel:
		return zerolog.WarnLevel
	case logger.ErrorLevel:
		return zerolog.ErrorLevel
	case logger.FatalLevel:
		return zerolog.FatalLevel
	case logger.PanicLevel:
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}
