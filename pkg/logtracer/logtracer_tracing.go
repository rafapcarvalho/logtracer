package logtracer

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"time"
)

func initTracerProvider(cfg Config) (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	exporter, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
		otlptracehttp.WithInsecure(),
	)

	if err != nil {
		SrvcLog.Error(ctx, "Failed to create OTLP exporter",
			"error", err,
			"endpoint", cfg.OTLPEndpoint,
		)
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

	otel.SetErrorHandler(&errorLogger{})
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagator)
	return tp, nil
}

type errorLogger struct{}

func (e errorLogger) Handle(err error) {
	SrvcLog.Error(context.Background(), "Trace export failed",
		"error", err,
	)
}

func Shutdown(ctx context.Context) error {
	if traceProvider != nil {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return traceProvider.Shutdown(ctx)
	}
	return nil
}
