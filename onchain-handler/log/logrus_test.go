package log

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogrusLogger_SetConfigModeByCode(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogrusLogger(&buf, InfoLevel)

	logger.SetConfigModeByCode("test_code")
	// Since SetConfigModeByCode is currently empty, we just verify it doesn't panic
	require.NotNil(t, logger)
}

func TestLogrusLogger_SetAndGetConfig(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogrusLogger(&buf, InfoLevel)

	testConfig := map[string]string{"key": "value"}
	logger.SetConfig(testConfig)

	retrievedConfig := logger.GetConfig()
	require.Equal(t, testConfig, retrievedConfig)
}

func TestLogrusLogger_SetAndGetServiceName(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogrusLogger(&buf, InfoLevel)

	testServiceName := "test-service"
	logger.SetServiceName(testServiceName)

	retrievedServiceName := logger.GetServiceName()
	require.Equal(t, testServiceName, retrievedServiceName)
}

func TestLogrusLogger_FormatFunctions(t *testing.T) {
	tests := []struct {
		name        string
		logFunc     func(*LogrusLogger, string, ...interface{})
		level       Level
		format      string
		args        []interface{}
		expectedLog string
	}{
		{
			name:        "Debugf",
			logFunc:     (*LogrusLogger).Debugf,
			level:       DebugLevel,
			format:      "Debug %s",
			args:        []interface{}{"message"},
			expectedLog: "Debug message",
		},
		{
			name:        "Infof",
			logFunc:     (*LogrusLogger).Infof,
			level:       InfoLevel,
			format:      "Info %s",
			args:        []interface{}{"message"},
			expectedLog: "Info message",
		},
		{
			name:        "Warnf",
			logFunc:     (*LogrusLogger).Warnf,
			level:       WarnLevel,
			format:      "Warn %s",
			args:        []interface{}{"message"},
			expectedLog: "Warn message",
		},
		{
			name:        "Errorf",
			logFunc:     (*LogrusLogger).Errorf,
			level:       ErrorLevel,
			format:      "Error %s",
			args:        []interface{}{"message"},
			expectedLog: "Error message",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogrusLogger(&buf, test.level)

			test.logFunc(logger, test.format, test.args...)
			require.Contains(t, buf.String(), test.expectedLog)
		})
	}
}

func TestLogrusLogger_PanicAndFatalF(t *testing.T) {
	tests := []struct {
		name    string
		logFunc func(*LogrusLogger, string, ...interface{})
		format  string
		args    []interface{}
	}{
		{
			name:    "Panicf",
			logFunc: (*LogrusLogger).Panicf,
			format:  "Panic %s",
			args:    []interface{}{"message"},
		},
		// {
		// 	name:    "Fatalf",
		// 	logFunc: (*LogrusLogger).Fatalf,
		// 	format:  "Fatal %s",
		// 	args:    []interface{}{"message"},
		// }, => fatal is going to exit the program
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogrusLogger(&buf, PanicLevel)

			require.Panics(t, func() {
				test.logFunc(logger, test.format, test.args...)
			})
			require.Contains(t, buf.String(), fmt.Sprintf(test.format, test.args...))
		})
	}
}
