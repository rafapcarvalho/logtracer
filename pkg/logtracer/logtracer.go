package logtracer

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"os"
	"strings"
)

var LoggerLevel = new(slog.LevelVar)

func New(cfg Config) (*LogTracer, error) {
	var handler slog.Handler
	logFormat := strings.ToLower(cfg.LogFormat)
	if logFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: LoggerLevel,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: LoggerLevel,
		})
	}

	logger := slog.New(handler)

	var tp *sdktrace.TracerProvider
	var tracer trace.Tracer
	var err error

	if cfg.EnableTracing {
		tp, err = initTracerProvider(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize tracer provider: %w", err)
		}
		tracer = tp.Tracer(cfg.ServiceName)
	}

	propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})

	lt := &LogTracer{
		logger:         logger,
		tracerProvider: tp,
		tracer:         tracer,
		propagator:     propagator,
		Categories:     make(map[string]*CategoryLogger),
	}

	lt.InitLog = lt.newCategoryLogger(cfg.ServiceName, "INIT")
	lt.CfgLog = lt.newCategoryLogger(cfg.ServiceName, "CFG")
	lt.SrvcLog = lt.newCategoryLogger(cfg.ServiceName, "SRV")
	lt.TstLog = lt.newCategoryLogger(cfg.ServiceName, "TST")

	return lt, nil
}

func (lt *LogTracer) newCategoryLogger(serviceName, category string) *CategoryLogger {
	return &CategoryLogger{
		logger: lt.logger.With("component", serviceName, "category", category),
		tracer: lt.tracer,
	}
}

func initTracerProvider(cfg Config) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	resourceAttrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(cfg.ServiceName),
	}
	for k, v := range cfg.AdditionalResource {
		resourceAttrs = append(resourceAttrs, attribute.String(k, v))
	}

	rsrc := resource.NewWithAttributes(semconv.SchemaURL, resourceAttrs...)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(rsrc),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{},
		),
	)

	return tp, nil
}

func (lt *LogTracer) Shutdown(ctx context.Context) error {
	if lt.tracerProvider != nil {
		return lt.tracerProvider.Shutdown(ctx)
	}
	return nil
}

func (lt *LogTracer) SetLogLevel(level LogLevel) {
	var slogLevel slog.Level
	switch level {
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelError:
		slogLevel = slog.LevelError
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelDebug:
		slogLevel = slog.LevelDebug
	default:
		slogLevel = slog.LevelInfo
	}
	LoggerLevel.Set(slogLevel)
}
