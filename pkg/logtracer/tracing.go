package logtracer

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	tracer "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func StartSpan(ctx context.Context, name string, opts ...tracer.SpanStartOption) (context.Context, tracer.Span) {
	if globalTracer != nil {
		return globalTracer.Start(ctx, name, opts...)
	}
	noopTrace := noop.NewTracerProvider()
	noopNewTracer := noopTrace.Tracer("")
	return noopNewTracer.Start(ctx, name, opts...)
}

func AddAttribute(ctx context.Context, key string, value interface{}) {
	span := tracer.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attribute.String(key, fmt.Sprint(value)))
	}
}
