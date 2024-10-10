package logtracer

import (
	"context"
	"fmt"
	"github.com/rafapcarvalho/logtracer/internal/handlers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"strings"
	"sync"
	"time"
)

var (
	ginLog     *CategoryLogger
	grpcLog    *CategoryLogger
	InitLog    *CategoryLogger
	CfgLog     *CategoryLogger
	SrvcLog    *CategoryLogger
	TstLog     *CategoryLogger
	Categories map[string]*CategoryLogger

	globalTracer  trace.Tracer
	traceProvider *sdktrace.TracerProvider
	propagator    propagation.TextMapPropagator
	shutdownOnce  sync.Once
)

func InitLogger(cfg Config) {
	var logger *slog.Logger
	logFormat := strings.ToLower(cfg.LogFormat)
	if logFormat == "json" {
		logger = slog.New(handlers.StdoutJSON())
	} else {
		logger = slog.New(handlers.StdoutTXT())
	}

	if cfg.EnableTracing {
		var err error
		traceProvider, err = initTracerProvider(cfg)
		if err == nil {
			globalTracer = traceProvider.Tracer(cfg.ServiceName)
			otel.SetTracerProvider(traceProvider)

			propagator = propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
			otel.SetTextMapPropagator(propagator)
		} else {
			logger.Error("Failed to initialize tracer provider: %w", err)
		}
	}

	InitLog = newCategoryLogger(logger, cfg.ServiceName, "INIT")
	CfgLog = newCategoryLogger(logger, cfg.ServiceName, "CFG")
	SrvcLog = newCategoryLogger(logger, cfg.ServiceName, "SRVC")
	TstLog = newCategoryLogger(logger, cfg.ServiceName, "TEST")
	ginLog = newCategoryLogger(logger, cfg.ServiceName, "GIN")
	grpcLog = newCategoryLogger(logger, cfg.ServiceName, "GRPC")

	Categories = make(map[string]*CategoryLogger)
}

func newCategoryLogger(logger *slog.Logger, serviceName, category string) *CategoryLogger {
	return &CategoryLogger{
		logger: logger.With("component", serviceName, "category", category),
	}
}

func initTracerProvider(cfg Config) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	exporter, err := otlptracehttp.New(
		ctx,
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

	res := resource.NewWithAttributes(semconv.SchemaURL, resourceAttrs...)
	resource.WithProcessRuntimeDescription()
	resource.WithTelemetrySDK()

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagator)
	return tp, nil
}

func Shutdown(ctx context.Context) error {
	if traceProvider != nil {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return traceProvider.Shutdown(ctx)
	}
	return nil
}

func SetLevel(level LogLevel) {
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
	handlers.LoggerLevel.Set(slogLevel)
}
