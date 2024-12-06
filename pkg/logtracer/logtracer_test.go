package logtracer

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInitLogger(t *testing.T) {
	cleanup := func() {
		InitLog = nil
		CfgLog = nil
		SrvcLog = nil
		TstLog = nil
		Categories = nil
		traceProvider = nil
		globalTracer = nil
	}

	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "JSON format without tracing",
			cfg: Config{
				ServiceName:   "test-service",
				LogFormat:     "json",
				EnableTracing: false,
			},
		},
		{
			name: "Text format with tracing",
			cfg: Config{
				ServiceName:   "test-service",
				LogFormat:     "text",
				EnableTracing: true,
				OTLPEndpoint:  "localhost:4318",
			},
		},
	}

	/*	defer func() {
			if r := recover(); r != nil {
				t.Errorf("Test panicked: %v", r)
			}
		}()
	*/
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup()

			InitLogger(tt.cfg)

			if InitLog == nil || InitLog.logger == nil {
				t.Error("InitLog was not properly initialized")
			}
			assert.NotNil(t, InitLog)
			assert.NotNil(t, CfgLog)
			assert.NotNil(t, SrvcLog)
			assert.NotNil(t, TstLog)
			assert.NotNil(t, NoTrace)
			assert.NotNil(t, Categories)

			if tt.cfg.EnableTracing {
				assert.NotNil(t, globalTracer, "Tracer should be initialized when tracing is enabled")
			}

			ctx := context.Background()
			InitLog.Info(ctx, "Test message")
		})
	}
}
