package logtracer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLogLevel(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LevelInfo, "Info"},
		{LevelError, "Error"},
		{LevelWarn, "Warn"},
		{LevelDebug, "Debug"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}
