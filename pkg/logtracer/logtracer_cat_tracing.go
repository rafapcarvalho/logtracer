package logtracer

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"
)

type CategoryLogger struct {
	logger *slog.Logger
}

type WithoutTracer struct{}

func (cl *CategoryLogger) Info(ctx context.Context, msg string, args ...any) {
	cl.execute(ctx, LevelInfo, msg, args...)
}

func (cl *CategoryLogger) Error(ctx context.Context, msg string, args ...any) {
	cl.execute(ctx, LevelError, msg, args...)
}

func (cl *CategoryLogger) Warn(ctx context.Context, msg string, args ...any) {
	cl.execute(ctx, LevelWarn, msg, args...)
}

func (cl *CategoryLogger) Debug(ctx context.Context, msg string, args ...any) {
	cl.execute(ctx, LevelDebug, msg, args...)
}

func (cl *CategoryLogger) Infof(ctx context.Context, format string, args ...any) {
	cl.execute(ctx, LevelInfo, fmt.Sprintf(format, args...))
}

func (cl *CategoryLogger) Errorf(ctx context.Context, format string, args ...any) {
	cl.execute(ctx, LevelError, fmt.Sprintf(format, args...))
}

func (cl *CategoryLogger) Warnf(ctx context.Context, format string, args ...any) {
	cl.execute(ctx, LevelWarn, fmt.Sprintf(format, args...))
}

func (cl *CategoryLogger) Debugf(ctx context.Context, format string, args ...any) {
	cl.execute(ctx, LevelDebug, fmt.Sprintf(format, args...))
}

func (cl *CategoryLogger) execute(ctx context.Context, level LogLevel, msg string, args ...any) {

	var slogLevel = getLogLevel(level)
	if !cl.logger.Enabled(ctx, slogLevel) {
		return
	}
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
