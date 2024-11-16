package zap

import (
	"fmt"
	"testing"

	"go.uber.org/zap/zapcore"

	"github.com/stretchr/testify/require"

	pkglogger "github.com/genefriendway/onchain-handler/pkg/logger"
)

func TestDefaultZapConfig(t *testing.T) {
	serviceName := "testService"
	config := DefaultZapConfig(serviceName)

	require.False(t, config.Development, "Expected Development to be false")
	require.Equal(t, "json", config.Encoding, "Expected Encoding to be 'json'")
	require.Equal(t, []string{"stdout"}, config.OutputPaths, "Expected OutputPaths to be ['stdout']")
	require.Equal(t, []string{"stderr"}, config.ErrorOutputPaths, "Expected ErrorOutputPaths to be ['stderr']")
	require.NotNil(t, config.InitialFields, "Expected InitialFields to be initialized")
	require.Equal(t, "timestamp", config.TimeKey, "Expected TimeKey to be 'timestamp'")
	require.Equal(t, "level", config.LevelKey, "Expected LevelKey to be 'level'")
	require.Equal(t, "message", config.MessageKey, "Expected MessageKey to be 'message'")
	require.Equal(t, "stacktrace", config.StacktraceKey, "Expected StacktraceKey to be 'stacktrace'")
	require.Equal(t, "caller", config.CallerKey, "Expected CallerKey to be 'caller'")
	require.Equal(t, "function", config.FunctionKey, "Expected FunctionKey to be 'function'")
	require.Equal(t, "2006-01-02T15:04:05.000Z0700", config.TimeFormat, "Expected TimeFormat to be '2006-01-02T15:04:05.000Z0700'")
}

func TestDevelopmentAndProductionConfigs(t *testing.T) {
	serviceName := "testService"

	devConfig := defaultZapConfigDevelop(serviceName)
	require.True(t, devConfig.Development)
	require.Equal(t, "development", devConfig.InitialFields["environment"])
	require.Equal(t, serviceName, devConfig.InitialFields["service"])

	prodConfig := defaultZapConfigProduction(serviceName)
	require.False(t, prodConfig.Development)
	require.Equal(t, "production", prodConfig.InitialFields["environment"])
	require.Equal(t, serviceName, prodConfig.InitialFields["service"])
}

func TestNewZapLogger(t *testing.T) {
	tests := []struct {
		name         string
		initialLevel pkglogger.Level
		config       *ZapLoggerConfig
		wantErr      bool
	}{
		{"default-info", pkglogger.InfoLevel, nil, false},
		{"custom-debug", pkglogger.DebugLevel, DefaultZapConfig("test"), false},
		{"custom-error", pkglogger.ErrorLevel, defaultZapConfigDevelop("test"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewZapLoggerWithConfig(tt.initialLevel, tt.config)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, logger)
			require.Equal(t, tt.initialLevel, logger.GetLogLevel())
		})
	}
}

func TestZapLogger_Operations(t *testing.T) {
	logger, err := NewZapLogger(pkglogger.InfoLevel)
	require.NoError(t, err)

	t.Run("level operations", func(t *testing.T) {
		logger.SetLogLevel(pkglogger.DebugLevel)
		require.Equal(t, pkglogger.DebugLevel, logger.GetLogLevel())
	})

	t.Run("with fields", func(t *testing.T) {
		fields := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		}
		newLogger := logger.WithFields(fields)
		require.NotNil(t, newLogger)
		require.NotEqual(t, logger, newLogger)
	})

	t.Run("service name", func(t *testing.T) {
		serviceName := "test-service"
		logger.SetServiceName(serviceName)
		require.Equal(t, serviceName, logger.GetServiceName())
	})
}

func TestConvertZapLevel(t *testing.T) {
	tests := []struct {
		level    pkglogger.Level
		expected zapcore.Level
	}{
		{pkglogger.DebugLevel, zapcore.DebugLevel},
		{pkglogger.InfoLevel, zapcore.InfoLevel},
		{pkglogger.WarnLevel, zapcore.WarnLevel},
		{pkglogger.ErrorLevel, zapcore.ErrorLevel},
		{pkglogger.FatalLevel, zapcore.FatalLevel},
		{pkglogger.PanicLevel, zapcore.PanicLevel},
		{pkglogger.Level(99), zapcore.InfoLevel}, // default case
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("level-%v", tt.level), func(t *testing.T) {
			result := convertZapLevel(tt.level)
			require.Equal(t, tt.expected, result)
		})
	}
}
