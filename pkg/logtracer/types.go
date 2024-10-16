package logtracer

import (
	"go.opentelemetry.io/otel/propagation"
	provider "go.opentelemetry.io/otel/sdk/trace"
	tracer "go.opentelemetry.io/otel/trace"
	"log/slog"
)

type LogTracer struct {
	logger         *slog.Logger
	tracerProvider *provider.TracerProvider
	tracer         tracer.Tracer
	propagator     propagation.TextMapPropagator
	InitLog        *CategoryLogger
	CfgLog         *CategoryLogger
	SrvcLog        *CategoryLogger
	TstLog         *CategoryLogger
	NoTrace        WithoutTracer
	Categories     map[string]*CategoryLogger
}

type SpanOption func(*SpanOptions)

type SpanOptions struct {
	Attributes map[string]string
	ID         string
}

func WithAttribute(key, value string) SpanOption {
	return func(o *SpanOptions) {
		if o.Attributes == nil {
			o.Attributes = make(map[string]string)
		}
		o.Attributes[key] = value
	}
}

func WithId(id string) SpanOption {
	return func(o *SpanOptions) {
		o.ID = id
	}
}

type Config struct {
	CustomID           string
	ServiceName        string
	LogFormat          string
	EnableTracing      bool
	OTLPEndpoint       string
	AdditionalResource map[string]string
}
