package logtracer

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"time"
)

func GinMiddleware(lt *LogTracer, serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()
		end := time.Now()
		latency := end.Sub(start)

		if query != "" {
			path = fmt.Sprintf("%s?%s", path, query)
		}

		lt.Categories["GIN"].Info(c.Request.Context(),
			"HTTP request",
			"status", c.Writer.Status(),
			"method", c.Request.Method,
			"path", path,
			"ip", c.ClientIP(),
			"latency", latency,
			"user-agent", c.Request.UserAgent(),
		)
	}
}

func OTELMiddleware(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}

func extractB3Headers(ctx context.Context) map[string]string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	headers := make(map[string]string)
	b3Headers := []string{"x-b3-traceid", "x-b3-spanid", "x-b3-parentspanid", "x-b3-sampled", "x-b3-flags"}

	for _, header := range b3Headers {
		if values := md.Get(header); len(values) > 0 {
			headers[header] = values[0]
		}
	}

	return headers
}

func (lt *LogTracer) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()

		b3Headers := extractB3Headers(ctx)
		newCtx := lt.propagator.Extract(ctx, propagation.MapCarrier(b3Headers))
		newCtx, span := lt.StartSpan(newCtx, info.FullMethod)
		defer span.End()

		lt.AddAtribute(newCtx, "grpc.Method", info.FullMethod)

		resp, err := handler(newCtx, req)

		duration := time.Since(startTime)
		statusCode := "OK"
		if err != nil {
			statusCode = "ERROR"
		}

		lt.Categories["GRPC"].Info(newCtx,
			"gRPC request",
			"method", info.FullMethod,
			"duration", duration,
			"status", statusCode,
		)
		return resp, err
	}
}

func (lt *LogTracer) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		startTime := time.Now()

		b3Headers := extractB3Headers(ss.Context())
		newCtx := lt.propagator.Extract(ss.Context(), propagation.MapCarrier(b3Headers))
		newCtx, span := lt.StartSpan(newCtx, info.FullMethod)
		defer span.End()

		lt.AddAtribute(newCtx, "grpc.method", info.FullMethod)

		wrapped := &wrappedServerStream{ServerStream: ss, ctx: newCtx}
		err := handler(srv, wrapped)

		duration := time.Since(startTime)
		statusCode := "OK"
		if err != nil {
			statusCode = "ERROR"
		}

		lt.SrvcLog.Info(newCtx,
			"gRPC stream",
			"method", info.FullMethod,
			"duration", duration,
			"status", statusCode,
		)

		return err
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
