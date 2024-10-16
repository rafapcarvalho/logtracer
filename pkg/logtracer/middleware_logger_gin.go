package logtracer

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"time"
)

func GinMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetLevel(LevelInfo)
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()
		end := time.Now()
		latency := end.Sub(start)

		if query != "" {
			path = fmt.Sprintf("%s?%s", path, query)
		}

		logHTTPRequest(
			c.Request.Context(),
			c.Writer.Status(),
			c.Request.Method,
			path,
			c.ClientIP(),
			latency,
			c.Request.UserAgent(),
		)
	}
}

func logHTTPRequest(ctx context.Context, status int, method, path, ip string, latency time.Duration, userAgent string) {
	if status >= 400 {
		ginLog.Error(ctx,
			"HTTP request",
			"status", status,
			"method", method,
			"path", path,
			"ip", ip,
			"latency", latency,
			"user-agent", userAgent,
		)
	} else {
		ginLog.Info(ctx,
			"HTTP request",
			"status", status,
			"method", method,
			"path", path,
			"ip", ip,
			"latency", latency,
			"user-agent", userAgent,
		)
	}
}

func OTELMiddleware(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}
