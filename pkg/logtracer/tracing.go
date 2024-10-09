package logtracer

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	tracer "go.opentelemetry.io/otel/trace"
)

func (lt *LogTracer) StartSpan(ctx context.Context, name string, opts ...tracer.SpanStartOption) (context.Context, tracer.Span) {
	if lt.tracer != nil {
		ctx, span := lt.tracer.Start(ctx, name, opts...)
		if spanContext := span.SpanContext(); spanContext.IsValid() {
			ctx = context.WithValue(ctx, "trace_id", spanContext.TraceID().String())
		}
		return ctx, span
	}
	return ctx, nil
}

func (lt *LogTracer) AddAtribute(ctx context.Context, key string, value interface{}) {
	if lt.tracer != nil {
		span := tracer.SpanFromContext(ctx)
		if span.IsRecording() {
			span.SetAttributes(attribute.String(key, fmt.Sprint(value)))
		}
	}
}
