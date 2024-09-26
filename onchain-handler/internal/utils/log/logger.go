package log

import (
	"io"
	"time"

	"github.com/rs/zerolog"
)

var LG *ZerologLogger

// ZerologLogger implements applogger.Logger interface and provide logging service
// using zerolog.
type ZerologLogger struct {
	Instance *zerolog.Logger
}

func NewZerologLogger(output io.Writer, level zerolog.Level) *ZerologLogger {
	return newZerologLogger(output, level, false)
}

func NewZerologLoggerWithColor(output io.Writer, level zerolog.Level) *ZerologLogger {
	return newZerologLogger(output, level, true)
}

func newZerologLogger(output io.Writer, level zerolog.Level, withColor bool) *ZerologLogger {
	zlog := zerolog.New(zerolog.ConsoleWriter{
		Out:        output,
		NoColor:    withColor,
		TimeFormat: time.RFC3339,
	})

	instance := &ZerologLogger{
		Instance: &zlog,
	}

	instance.SetLogLevel(level)

	zerolog.TimestampFieldName = "t"
	zerolog.LevelFieldName = "l"
	zerolog.MessageFieldName = "m"

	return instance
}

/**
* Setting
**/

func (logger *ZerologLogger) SetLogLevel(level zerolog.Level) {
	updatedLogger := logger.Instance.Level(level)

	logger.Instance = &updatedLogger
}

func (logger *ZerologLogger) GetLogLevel() zerolog.Level {
	return logger.Instance.GetLevel()
}

/**
* Logging function
**/

func (logger *ZerologLogger) Panic(message string) {
	logger.Instance.Panic().Timestamp().Msg(message)
}

func (logger *ZerologLogger) Panicf(format string, values ...interface{}) {
	logger.Instance.Panic().Timestamp().Msgf(format, values...)
}

func (logger *ZerologLogger) Fatal(message string) {
	logger.Instance.Fatal().Timestamp().Msg(message)
}

func (logger *ZerologLogger) Fatalf(format string, values ...interface{}) {
	logger.Instance.Fatal().Timestamp().Msgf(format, values...)
}

func (logger *ZerologLogger) Error(message string) {
	logger.Instance.Error().Timestamp().Msg(message)
}

func (logger *ZerologLogger) Errorf(format string, values ...interface{}) {
	logger.Instance.Error().Timestamp().Msgf(format, values...)
}

func (logger *ZerologLogger) Info(message string) {
	logger.Instance.Info().Timestamp().Msg(message)
}

func (logger *ZerologLogger) Infof(format string, values ...interface{}) {
	logger.Instance.Info().Timestamp().Msgf(format, values...)
}

func (logger *ZerologLogger) Debug(message string) {
	logger.Instance.Debug().Timestamp().Msg(message)
}

func (logger *ZerologLogger) Debugf(format string, values ...interface{}) {
	logger.Instance.Debug().Timestamp().Msgf(format, values...)
}

func (logger *ZerologLogger) Warn(message string) {
	logger.Instance.Warn().Timestamp().Msg(message)
}

func (logger *ZerologLogger) Warnf(format string, values ...interface{}) {
	logger.Instance.Warn().Timestamp().Msgf(format, values...)
}

func (logger *ZerologLogger) WithInterface(key string, value interface{}) *ZerologLogger {
	Instance := logger.Instance.With().Interface(key, value).Logger()
	return &ZerologLogger{
		Instance: &Instance,
	}
}

func (logger *ZerologLogger) WithFields(fields interface{}) *ZerologLogger {
	Instance := logger.Instance.With().Fields(fields).Logger()
	return &ZerologLogger{
		Instance: &Instance,
	}
}
