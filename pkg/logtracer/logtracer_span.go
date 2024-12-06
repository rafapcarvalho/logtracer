package logtracer

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	tracer "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type customIDKey struct{}

func StartSpan(ctx context.Context, name string, opts ...SpanOption) context.Context {
	options := &SpanOptions{}
	for _, opt := range opts {
		opt(options)
	}

	/*if options.ID != "" {
		ctx = context.WithValue(ctx, customIDKey{}, options.ID)
	}
	*/
	if customID != "" {
		ctx = context.WithValue(ctx, customID.String(), options.ID)
	}

	var span tracer.Span
	if globalTracer != nil {
		var attrs []attribute.KeyValue
		for k, v := range options.Attributes {
			attrs = append(attrs, attribute.String(k, v))
		}
		/*		if ctx.Value(customID.String()).(string) != "" {
				attrs = append(attrs, attribute.String("custom.id", ctx.Value(customID.String()).(string)))
			}*/
		if options.ID != "" {
			attrs = append(attrs, attribute.String("custom.id", options.ID))
		}
		ctx, span = globalTracer.Start(ctx, name, tracer.WithAttributes(attrs...))
	} else {
		noopTrace := noop.NewTracerProvider()
		noopNewTracer := noopTrace.Tracer("")
		ctx, span = noopNewTracer.Start(ctx, name)
	}

	return context.WithValue(ctx, spanKey{}, span)
}

type spanKey struct{}

func AddAttribute(ctx context.Context, key string, value interface{}) {
	if span, ok := ctx.Value(spanKey{}).(tracer.Span); ok && span.IsRecording() {
		span.SetAttributes(attribute.String(key, fmt.Sprint(value)))
	}
}

func EndSpan(ctx context.Context) {
	if span, ok := ctx.Value(spanKey{}).(tracer.Span); ok {
		span.End()
	}
}

func GetCustomID(ctx context.Context) string {
	id, _ := ctx.Value(customID.String()).(string)
	return id
}

func getOrCreateTraceID(ctx context.Context) string {
	customID := GetCustomID(ctx)
	if customID != "" {
		return customID
	}

	spanCtx := tracer.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}
	return ""
}

func recordLogSpan(ctx context.Context, level LogLevel, msg string, args ...any) {
	span := tracer.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("execute.level", level.String()),
		attribute.String("execute.message", msg),
	}

	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if ok {
				attrs = append(attrs, attribute.String(key, fmt.Sprint(args[i+1])))
			}
		}
	}

	span.AddEvent("log", tracer.WithAttributes(attrs...))

	if level == LevelError {
		span.SetStatus(codes.Error, "execution error")
	}
}
