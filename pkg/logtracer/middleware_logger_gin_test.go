package logtracer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGinMiddleware(t *testing.T) {
	middleware := GinMiddleware("test-service")
	assert.NotNil(t, middleware)

	otelMiddleware := OTELMiddleware("test-service")
	assert.NotNil(t, otelMiddleware)
}
