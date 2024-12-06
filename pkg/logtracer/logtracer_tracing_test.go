package logtracer

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInitTracerProvider(t *testing.T) {
	cfg := Config{
		ServiceName:   "test-service",
		OTLPEndpoint:  "localhost:4318",
		EnableTracing: true,
		AdditionalResource: map[string]string{
			"env": "test",
		},
	}

	provider, err := initTracerProvider(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestShutdown(t *testing.T) {
	cfg := Config{
		ServiceName:   "test-service",
		EnableTracing: true,
		OTLPEndpoint:  "localhost:4318",
	}

	InitLogger(cfg)
	ctx := context.Background()

	err := Shutdown(ctx)
	assert.NoError(t, err)
}
