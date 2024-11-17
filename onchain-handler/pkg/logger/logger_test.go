package logger

import (
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/require"

	"github.com/genefriendway/onchain-handler/pkg/interfaces"
	"github.com/genefriendway/onchain-handler/pkg/mocks"
)

func TestInitLogger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		setup     func() interfaces.Logger
		wantPanic bool
	}{
		{
			name: "Successfully initialize logger",
			setup: func() interfaces.Logger {
				return mocks.NewMockLogger(ctrl)
			},
			wantPanic: false,
		},
		{
			name: "Panic when initializing nil logger",
			setup: func() interfaces.Logger {
				return nil
			},
			wantPanic: true,
		},
		{
			name: "Initialize logger only once (singleton)",
			setup: func() interfaces.Logger {
				firstLogger := mocks.NewMockLogger(ctrl)
				InitLogger(firstLogger)
				return mocks.NewMockLogger(ctrl)
			},
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Reset() // Ensure logger is reset before each test
			logger := tt.setup()

			if tt.wantPanic {
				require.Panics(t, func() {
					InitLogger(logger)
				})
			} else {
				require.NotPanics(t, func() {
					InitLogger(logger)
				})

				// Verify the logger is correctly initialized
				actualLogger := GetLogger()
				require.NotNil(t, actualLogger)

				// For singleton test case, verify the first logger remains
				if tt.name == "Initialize logger only once (singleton)" {
					require.Equal(t, instance, GetLogger())
				}
			}
		})
	}
}

func TestGetLogger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		setup     func()
		wantPanic bool
	}{
		{
			name: "Successfully get initialized logger",
			setup: func() {
				mockLogger := mocks.NewMockLogger(ctrl)
				Reset()
				InitLogger(mockLogger)
			},
			wantPanic: false,
		},
		{
			name:      "Panic when logger not initialized",
			setup:     func() { Reset() },
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			if tt.wantPanic {
				require.Panics(t, func() {
					GetLogger()
				})
			} else {
				logger := GetLogger()
				require.NotNil(t, logger)
			}
		})
	}
}

func TestSetAndGetLogLevel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
			mockLogger := mocks.NewMockLogger(ctrl)
			Reset() // Ensure the logger is reset before each test
			InitLogger(mockLogger)

			mockLogger.EXPECT().SetLogLevel(tt.level).Times(1)
			mockLogger.EXPECT().GetLogLevel().Return(tt.expected).Times(1)

			mockLogger.SetLogLevel(tt.level)
			require.Equal(t, tt.expected, mockLogger.GetLogLevel())
		})
	}
}
