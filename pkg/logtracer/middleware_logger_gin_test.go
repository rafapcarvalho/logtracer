package logtracer

import (
	"bytes"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"path/filepath"
	"testing"
	"time"
)

func TestGinMiddleware(t *testing.T) {
	middleware := GinMiddleware("test-service")
	assert.NotNil(t, middleware)

	otelMiddleware := OTELMiddleware("test-service")
	assert.NotNil(t, otelMiddleware)
}

func TestLogHTTPRequest(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				if src, ok := a.Value.Any().(*slog.Source); ok {
					function := filepath.Base(src.Function)
					file := filepath.Base(src.File)
					a.Value = slog.StringValue(fmt.Sprintf("[%s] %s:%d", function, file, src.Line))
				}
			}
			return a
		},
	}))

	ginLog = newCategoryLogger(l, "test-service", "GIN")

	tests := []struct {
		name      string
		status    int
		method    string
		path      string
		ip        string
		latency   time.Duration
		userAgent string
		want      string
	}{
		{
			name:      "Successful request",
			status:    200,
			method:    "GET",
			path:      "/api/v1/test",
			ip:        "127.0.0.1",
			latency:   100 * time.Millisecond,
			userAgent: "test-agent",
			want:      `{"level":"INFO","msg":"HTTP request","component":"test-service","category":"GIN","status":200,"method":"GET","path":"/api/v1/test","ip":"127.0.0.1","latency":"100ms","user-agent":"test-agent"}`,
		},
		{
			name:      "Client error request with microseconds",
			status:    404,
			method:    "POST",
			path:      "/api/v1/missing",
			ip:        "192.168.1.1",
			latency:   500 * time.Microsecond,
			userAgent: "postman",
			want:      `{"level":"ERROR","msg":"HTTP request","component":"test-service","category":"GIN","status":404,"method":"POST","path":"/api/v1/missing","ip":"192.168.1.1","latency":"500µs","user-agent":"postman"}`,
		},
		{
			name:      "Server error request with seconds",
			status:    500,
			method:    "PUT",
			path:      "/api/v1/error",
			ip:        "10.0.0.1",
			latency:   1500 * time.Millisecond,
			userAgent: "curl",
			want:      `{"level":"ERROR","msg":"HTTP request","component":"test-service","category":"GIN","status":500,"method":"PUT","path":"/api/v1/error","ip":"10.0.0.1","latency":"1.50s","user-agent":"curl"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			logHTTPRequest(context.Background(), tt.status, tt.method, tt.path, tt.ip, tt.latency, tt.userAgent)
			if buf.Len() == 0 {
				t.Fatal("No log output captured")
			}
			checkLogOutput(t, buf.String(), tt.want, "time", "source")
		})
	}
}
