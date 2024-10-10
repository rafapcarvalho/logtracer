package logtracer

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
		newCtx, span := StartSpan(newCtx, info.FullMethod)
		defer span.End()

		AddAttribute(newCtx, "grpc.Method", info.FullMethod)

		resp, err := handler(newCtx, req)

		duration := time.Since(startTime)
		statusCode := "OK"
		if err != nil {
			statusCode = "ERROR"
		}

		grpcLog.Info(newCtx,
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
		newCtx, span := StartSpan(newCtx, info.FullMethod)
		defer span.End()

		AddAttribute(newCtx, "grpc.method", info.FullMethod)

		wrapped := &wrappedServerStream{ServerStream: ss, ctx: newCtx}
		err := handler(srv, wrapped)

		duration := time.Since(startTime)
		statusCode := "OK"
		if err != nil {
			statusCode = "ERROR"
		}

		grpcLog.Info(newCtx,
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

func (lt *LogTracer) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		startTime := time.Now()

		newCtx, span := StartSpan(ctx, method)
		defer span.End()

		AddAttribute(newCtx, "grpc.method", method)

		err := invoker(newCtx, method, req, reply, cc, opts...)

		duration := time.Since(startTime)
		statusCode := "OK"
		if err != nil {
			statusCode = "ERROR"
		}

		grpcLog.Info(newCtx,
			"gRPC client request",
			"method", method,
			"duration", duration,
			"status", statusCode,
		)

		return err
	}
}

// StreamClientInterceptor returns a gRPC stream client interceptor for logging and tracing
func (lt *LogTracer) StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		startTime := time.Now()

		newCtx, span := StartSpan(ctx, method)
		defer span.End()

		AddAttribute(newCtx, "grpc.method", method)

		clientStream, err := streamer(newCtx, desc, cc, method, opts...)

		duration := time.Since(startTime)
		statusCode := "OK"
		if err != nil {
			statusCode = "ERROR"
		}

		grpcLog.Info(newCtx,
			"gRPC client stream",
			"method", method,
			"duration", duration,
			"status", statusCode,
		)

		return clientStream, err
	}
}

// OTELGRPCServerInterceptor returns OpenTelemetry gRPC server interceptors
func OTELGRPCServerInterceptor() (grpc.UnaryServerInterceptor, grpc.StreamServerInterceptor) {
	return otelgrpc.UnaryServerInterceptor(), otelgrpc.StreamServerInterceptor()
}

// OTELGRPCClientInterceptor returns OpenTelemetry gRPC client interceptors
func OTELGRPCClientInterceptor() (grpc.UnaryClientInterceptor, grpc.StreamClientInterceptor) {
	return otelgrpc.UnaryClientInterceptor(), otelgrpc.StreamClientInterceptor()
}
