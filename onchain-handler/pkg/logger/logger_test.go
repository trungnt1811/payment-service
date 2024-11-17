package logger

import (
	"testing"

	"github.com/genefriendway/onchain-handler/pkg/interfaces"
)

type MockLogger struct {
	level interfaces.Level
}

func (m *MockLogger) SetLogLevel(level interfaces.Level) {
	m.level = level
}

func (m *MockLogger) GetLogLevel() interfaces.Level {
	return m.level
}

func (m *MockLogger) Panic(message string)                                          {}
func (m *MockLogger) Panicf(format string, values ...interface{})                   {}
func (m *MockLogger) Fatal(message string)                                          {}
func (m *MockLogger) Fatalf(format string, values ...interface{})                   {}
func (m *MockLogger) Error(message string)                                          {}
func (m *MockLogger) Errorf(format string, values ...interface{})                   {}
func (m *MockLogger) Info(message string)                                           {}
func (m *MockLogger) Infof(format string, values ...interface{})                    {}
func (m *MockLogger) Debug(message string)                                          {}
func (m *MockLogger) Debugf(format string, values ...interface{})                   {}
func (m *MockLogger) Warn(message string)                                           {}
func (m *MockLogger) Warnf(format string, values ...interface{})                    {}
func (m *MockLogger) WithInterface(key string, value interface{}) interfaces.Logger { return m }
func (m *MockLogger) WithFields(fields map[string]interface{}) interfaces.Logger    { return m }
func (m *MockLogger) SetConfigModeByCode(code string)                               {}
func (m *MockLogger) SetConfig(config interface{})                                  {}
func (m *MockLogger) GetConfig() interface{}                                        { return nil }
func (m *MockLogger) SetServiceName(serviceName string)                             {}
func (m *MockLogger) GetServiceName() string                                        { return "" }

func TestInitLogger(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		logger   *MockLogger
		expected *MockLogger
	}{
		{name: "Initialize Logger", logger: &MockLogger{}, expected: &MockLogger{}},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Reset() // Ensure the logger is reset before each test
			InitLogger(tt.logger)
			if instance != tt.logger {
				t.Errorf("Expected logger instance to be %v, got %v", tt.logger, instance)
			}
		})
	}
}

func TestGetLogger(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		init     bool
		expected *MockLogger
	}{
		{name: "Get Logger After Init", init: true, expected: &MockLogger{}},
		{name: "Get Logger Without Init", init: false, expected: nil},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.init {
				InitLogger(tt.expected)
			} else {
				// Reset instance and once for testing purposes
				Reset()
			}

			defer func() {
				if r := recover(); r != nil && tt.expected != nil {
					t.Errorf("Did not expect panic, but got one")
				} else if r == nil && tt.expected == nil {
					t.Errorf("Expected panic, but did not get one")
				}
			}()

			logger := GetLogger()
			if (logger == nil && tt.expected != nil) || (logger != nil && tt.expected == nil) {
				t.Errorf("Expected logger to be %v, got %v", tt.expected, logger)
			} else if logger != nil && tt.expected != nil && logger.(*MockLogger).level != tt.expected.level {
				t.Errorf("Expected logger level to be %v, got %v", tt.expected.level, logger.(*MockLogger).level)
			}
		})
	}
}

func TestSetAndGetLogLevel(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		level    interfaces.Level
		expected interfaces.Level
	}{
		{name: "Set and Get Info Level", level: interfaces.InfoLevel, expected: interfaces.InfoLevel},
		{name: "Set and Get Debug Level", level: interfaces.DebugLevel, expected: interfaces.DebugLevel},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			Reset() // Ensure the logger is reset before each test
			InitLogger(mockLogger)

			mockLogger.SetLogLevel(tt.level)
			if mockLogger.GetLogLevel() != tt.expected {
				t.Errorf("Expected log level to be %v, got %v", tt.expected, mockLogger.GetLogLevel())
			}
		})
	}
}
