package logtracer

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"
)

func (WithoutTracer) Info(ctx context.Context, msg string, args ...any) {
	noTrace.executeNoTrace(ctx, LevelInfo, msg, args...)
}

func (WithoutTracer) Error(ctx context.Context, msg string, args ...any) {
	noTrace.executeNoTrace(ctx, LevelError, msg, args...)
}

func (WithoutTracer) Warn(ctx context.Context, msg string, args ...any) {
	noTrace.executeNoTrace(ctx, LevelWarn, msg, args...)
}

func (WithoutTracer) Debug(ctx context.Context, msg string, args ...any) {
	noTrace.executeNoTrace(ctx, LevelDebug, msg, args...)
}

func (WithoutTracer) Infof(ctx context.Context, format string, args ...any) {
	noTrace.executeNoTrace(ctx, LevelInfo, fmt.Sprintf(format, args...))
}

func (WithoutTracer) Errorf(ctx context.Context, format string, args ...any) {
	noTrace.executeNoTrace(ctx, LevelError, fmt.Sprintf(format, args...))
}

func (WithoutTracer) Warnf(ctx context.Context, format string, args ...any) {
	noTrace.executeNoTrace(ctx, LevelWarn, fmt.Sprintf(format, args...))
}

func (WithoutTracer) Debugf(ctx context.Context, format string, args ...any) {
	noTrace.executeNoTrace(ctx, LevelDebug, fmt.Sprintf(format, args...))
}

func (cl *CategoryLogger) executeNoTrace(ctx context.Context, level LogLevel, msg string, args ...any) {

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
}
