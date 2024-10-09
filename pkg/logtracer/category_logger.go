package logtracer

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"strings"
)

func (cl *CategoryLogger) Info(ctx context.Context, msg string, args ...any) {
	cl.log(ctx, LevelInfo, msg, args...)
}

func (cl *CategoryLogger) Error(ctx context.Context, msg string, args ...any) {
	cl.log(ctx, LevelError, msg, args...)
}

func (cl *CategoryLogger) Warn(ctx context.Context, msg string, args ...any) {
	cl.log(ctx, LevelWarn, msg, args...)
}

func (cl *CategoryLogger) Debug(ctx context.Context, msg string, args ...any) {
	cl.log(ctx, LevelDebug, msg, args...)
}

func (cl *CategoryLogger) Gin(ctx context.Context, msg string, args ...any) {
	cl.log(ctx, LevelInfo, msg, args...)
}

func (cl *CategoryLogger) Grpc(ctx context.Context, msg string, args ...any) {
	cl.log(ctx, LevelInfo, msg, args...)
}

func (cl *CategoryLogger) log(ctx context.Context, level LogLevel, msg string, args ...any) {
	id := getOrCreateTraceID(ctx)

	var slogLevel = getLogLevel(level)

	cl.logger.Log(ctx, slogLevel, msg, append([]any{"id", id, "severity", level.String()}, args...)...)

	if cl.tracer != nil {
		recordLogSpan(ctx, level, msg, args)
	}
}

func (cl *CategoryLogger) logf(ctx context.Context, level LogLevel, format string, args ...any) {
	id := getOrCreateTraceID(ctx)

	format = strings.ReplaceAll(format, "/", "∕")
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			args[i] = strings.ReplaceAll(str, "/", "∕")
		}
	}

	var slogLevel = getLogLevel(level)

	cl.logger.Log(ctx, slogLevel, fmt.Sprintf(format, args...), append([]any{"id", id})...)

	if cl.tracer != nil {
		recordLogSpan(ctx, level, format, args)
	}
}

func getOrCreateTraceID(ctx context.Context) string {
	if id, ok := ctx.Value("trace_id").(string); ok {
		return id
	}
	return uuid.New().String()
}

func recordLogSpan(ctx context.Context, level LogLevel, msg string, args ...any) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		attrs := []attribute.KeyValue{
			attribute.String("log.severity", level.String()),
			attribute.String("log.message", fmt.Sprintf(msg, args...)),
		}
		switch level {
		case LevelError:
			span.SetStatus(codes.Error, msg)
			span.RecordError(fmt.Errorf(msg))
		case LevelWarn:
			span.AddEvent("warning", trace.WithAttributes(attrs...))
		default:
			span.AddEvent("log", trace.WithAttributes(attrs...))
		}
	}
}

func getLogLevel(level LogLevel) slog.Level {
	var newLevel slog.Level
	switch level {
	case LevelError:
		newLevel = slog.LevelError
	case LevelWarn:
		newLevel = slog.LevelWarn
	case LevelDebug:
		newLevel = slog.LevelDebug
	default:
		newLevel = slog.LevelInfo
	}
	return newLevel
}
