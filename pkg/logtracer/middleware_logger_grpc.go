package logtracer

import (
	"context"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"time"
)

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
		newCtx = StartSpan(newCtx, info.FullMethod)
		defer EndSpan(newCtx)

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
		newCtx = StartSpan(newCtx, info.FullMethod)
		defer EndSpan(newCtx)

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

		newCtx := StartSpan(ctx, method)
		defer EndSpan(newCtx)

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

		newCtx := StartSpan(ctx, method)
		defer EndSpan(newCtx)

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
