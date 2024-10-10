package logtracer

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"runtime"
	"strings"
	"time"
)

type CategoryLogger struct {
	logger *slog.Logger
}

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

func (cl *CategoryLogger) Infof(ctx context.Context, msg string, args ...any) {
	cl.logf(ctx, LevelInfo, msg, args...)
}

func (cl *CategoryLogger) Errorf(ctx context.Context, msg string, args ...any) {
	cl.logf(ctx, LevelError, msg, args...)
}

func (cl *CategoryLogger) Warnf(ctx context.Context, msg string, args ...any) {
	cl.logf(ctx, LevelWarn, msg, args...)
}

func (cl *CategoryLogger) Debugf(ctx context.Context, msg string, args ...any) {
	cl.logf(ctx, LevelDebug, msg, args...)
}

func (cl *CategoryLogger) log(ctx context.Context, level LogLevel, msg string, args ...any) {
	var slogLevel = getLogLevel(level)
	id := getOrCreateTraceID(ctx)

	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), slogLevel, msg, pcs[0])
	if id != "" {
		r.Add("id", id)
	}
	r.Add(args...)
	_ = cl.logger.Handler().Handle(ctx, r)

	if globalTracer != nil {
		recordLogSpan(ctx, level, msg, args...)
	}
}

func (cl *CategoryLogger) logf(ctx context.Context, level LogLevel, format string, args ...any) {
	var slogLevel = getLogLevel(level)
	if !cl.logger.Enabled(ctx, slogLevel) {
		return
	}

	id := getOrCreateTraceID(ctx)

	format = strings.ReplaceAll(format, "/", "∕")
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			args[i] = strings.ReplaceAll(str, "/", "∕")
		}
	}

	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), slogLevel, fmt.Sprintf(format, args...), pcs[0])
	if id != "" {
		r.Add("id", id)
	}
	_ = cl.logger.Handler().Handle(ctx, r)

	if globalTracer != nil {
		recordLogSpan(ctx, level, format, args...)
	}
}

func getOrCreateTraceID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}
	return ""
}

func recordLogSpan(ctx context.Context, level LogLevel, msg string, args ...any) {
	ctx, span := StartSpan(ctx, "log")
	defer span.End()

	span.SetAttributes(
		attribute.String("log.level", level.String()),
		attribute.String("log.message", msg),
	)

	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if ok {
				span.SetAttributes(attribute.String(key, fmt.Sprint(args[i+1])))
			}
		}
	}

	if level == LevelError {
		span.SetStatus(codes.Error, msg)
		span.RecordError(fmt.Errorf(msg))
	} /*
		case LevelWarn:
			span.AddEvent("warning", trace.WithAttributes(attrs...))
		default:
			span.AddEvent("log", trace.WithAttributes(attrs...))
		}


		span := trace.SpanFromContext(ctx)
		if span.IsRecording() {
			msg := fmt.Sprintf(msg, args...)
			attrs := []attribute.KeyValue{
				attribute.String("log.severity", level.String()),
				attribute.String("log.message", msg),
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
		}*/
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
