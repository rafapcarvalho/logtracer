package logtracer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGRPCInterceptors(t *testing.T) {
	lt := &LogTracer{}

	t.Run("Server Interceptors", func(t *testing.T) {
		unary := lt.UnaryServerInterceptor()
		assert.NotNil(t, unary)

		stream := lt.StreamServerInterceptor()
		assert.NotNil(t, stream)
	})

	t.Run("Client Interceptors", func(t *testing.T) {
		unary := lt.UnaryClientInterceptor()
		assert.NotNil(t, unary)

		stream := lt.StreamClientInterceptor()
		assert.NotNil(t, stream)
	})

	t.Run("OTEL Interceptors", func(t *testing.T) {
		serverUnary, serverStream := OTELGRPCServerInterceptor()
		assert.NotNil(t, serverUnary)
		assert.NotNil(t, serverStream)

		clientUnary, clientStream := OTELGRPCClientInterceptor()
		assert.NotNil(t, clientUnary)
		assert.NotNil(t, clientStream)
	})
}
