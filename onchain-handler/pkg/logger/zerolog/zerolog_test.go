package zerolog

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/genefriendway/onchain-handler/pkg/interfaces"
)

func TestNewZerologLogger(t *testing.T) {
	tests := []struct {
		name     string
		level    interfaces.Level
		expected interfaces.Level
		useColor bool
	}{
		{"Info level without colors", interfaces.InfoLevel, interfaces.InfoLevel, false},
		{"Debug level with colors", interfaces.DebugLevel, interfaces.DebugLevel, true},
		{"Error level without colors", interfaces.ErrorLevel, interfaces.ErrorLevel, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewZerologLogger(&buf, test.level, test.useColor)

			require.NotNil(t, logger)
			require.Equal(t, test.expected, logger.GetLogLevel())
		})
	}
}

func TestSetAndGetConfig(t *testing.T) {
	var buf bytes.Buffer
	logger := NewZerologLogger(&buf, interfaces.InfoLevel, false)

	config := map[string]string{"key": "value"}
	logger.SetConfig(config)

	require.Equal(t, config, logger.GetConfig())
}

func TestSetAndGetServiceName(t *testing.T) {
	var buf bytes.Buffer
	logger := NewZerologLogger(&buf, interfaces.InfoLevel, false)

	serviceName := "my-service"
	logger.SetServiceName(serviceName)

	require.Equal(t, serviceName, logger.GetServiceName())
}

func TestSetConfigModeByCode(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		expectedLevel interfaces.Level
	}{
		{"Development mode", interfaces.DEVELOPMENT_ENVIRONMENT_CODE_MODE, interfaces.DebugLevel},
		{"Production mode", interfaces.PRODUCTION_ENVIRONMENT_CODE_MODE, interfaces.InfoLevel},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewZerologLogger(&buf, interfaces.InfoLevel, false)

			logger.SetConfigModeByCode(test.code)
			require.Equal(t, test.expectedLevel, logger.GetLogLevel())
		})
	}
}

func TestLoggingMethods(t *testing.T) {
	tests := []struct {
		logFunc  func(logger *ZerologLogger, msg string)
		message  string
		expected string
	}{
		{(*ZerologLogger).Debug, "debug message", "debug message"},
		{(*ZerologLogger).Info, "info message", "info message"},
		{(*ZerologLogger).Warn, "warn message", "warn message"},
		{(*ZerologLogger).Error, "error message", "error message"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewZerologLogger(&buf, interfaces.DebugLevel, false)

			test.logFunc(logger, test.message)
			require.Contains(t, buf.String(), test.expected)
		})
	}
}

func TestLoggingMethodsf(t *testing.T) {
	tests := []struct {
		logFunc  func(logger *ZerologLogger, format string, v ...interface{})
		format   string
		args     []interface{}
		expected string
	}{
		{(*ZerologLogger).Debugf, "debug %s", []interface{}{"message"}, "debug message"},
		{(*ZerologLogger).Infof, "info %s", []interface{}{"message"}, "info message"},
		{(*ZerologLogger).Warnf, "warn %s", []interface{}{"message"}, "warn message"},
		{(*ZerologLogger).Errorf, "error %s", []interface{}{"message"}, "error message"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewZerologLogger(&buf, interfaces.DebugLevel, false)

			test.logFunc(logger, test.format, test.args...)
			require.Contains(t, buf.String(), test.expected)
		})
	}
}

func TestWithInterface(t *testing.T) {
	tests := []struct {
		key      string
		value    interface{}
		expected string
	}{
		{"key1", "value1", `"key1":"value1"`},
		{"key2", 123, `"key2":123`},
		{"key3", true, `"key3":true`},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewZerologLogger(&buf, interfaces.DebugLevel, false)

			newLogger := logger.WithInterface(test.key, test.value)
			newLogger.Info("test message")
			require.Contains(t, buf.String(), test.expected)
		})
	}
}

func TestWithFields(t *testing.T) {
	tests := []struct {
		fields   map[string]interface{}
		expected []string
	}{
		{
			fields: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
				"key3": true,
			},
			expected: []string{`"key1":"value1"`, `"key2":123`, `"key3":true`},
		},
	}

	for _, test := range tests {
		t.Run("WithFields", func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewZerologLogger(&buf, interfaces.DebugLevel, false)

			newLogger := logger.WithFields(test.fields)
			newLogger.Info("test message")
			for _, expected := range test.expected {
				require.Contains(t, buf.String(), expected)
			}
		})
	}
}
