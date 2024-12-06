package logtracer

import (
	"context"
	"testing"
)

func TestWithoutTracer(t *testing.T) {
	InitLogger(Config{ServiceName: "test-service", LogFormat: "json"})
	ctx := context.Background()

	tests := []struct {
		name    string
		logFunc func(context.Context, string, ...any)
	}{
		{"Info", NoTrace.Info},
		{"Error", NoTrace.Error},
		{"Warn", NoTrace.Warn},
		{"Debug", NoTrace.Debug},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.logFunc(ctx, "test message")
		})
	}

	tests2 := []struct {
		name    string
		logFunc func(context.Context, string, ...any)
	}{
		{"Infof", NoTrace.Infof},
		{"Errorf", NoTrace.Errorf},
		{"Warnf", NoTrace.Warnf},
		{"Debugf", NoTrace.Debugf},
	}

	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			tt.logFunc(ctx, "test message %s", "formatted")
		})
	}
}
