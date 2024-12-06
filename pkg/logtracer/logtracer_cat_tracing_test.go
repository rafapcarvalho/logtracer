package logtracer

import (
	"context"
	"log/slog"
	"testing"
)

func TestCategoryLogger(t *testing.T) {
	InitLogger(Config{ServiceName: "test-service", LogFormat: "json"})
	ctx := context.Background()
	logger := newCategoryLogger(slog.Default(), "test-service", "TEST")

	tests := []struct {
		name    string
		logFunc func(context.Context, string, ...any)
	}{
		{"Info", logger.Info},
		{"Error", logger.Error},
		{"Warn", logger.Warn},
		{"Debug", logger.Debug},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.logFunc(ctx, "test message")
		})
	}
}
