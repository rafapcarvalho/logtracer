package logtracer

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSpanOperations(t *testing.T) {
	cfg := Config{
		ServiceName:   "test-service",
		LogFormat:     "json",
		EnableTracing: true,
		OTLPEndpoint:  "localhost:4318",
	}
	InitLogger(cfg)

	ctx := context.Background()
	ctx = StartSpan(ctx, "test-span", WithAttribute("key", "value"))

	AddAttribute(ctx, "new-key", "new-value")
	assert.NotPanics(t, func() {
		EndSpan(ctx)
	})
}
