package logtracer

import (
	"go.opentelemetry.io/otel/propagation"
	provider "go.opentelemetry.io/otel/sdk/trace"
	tracer "go.opentelemetry.io/otel/trace"
	"log/slog"
)

type LogLevel int

const (
	LevelInfo LogLevel = iota
	LevelError
	LevelWarn
	LevelDebug
)

func (l LogLevel) String() string {
	return [...]string{"Info", "Error", "Warn", "Debug"}[l]
}

type LogTracer struct {
	logger         *slog.Logger
	tracerProvider *provider.TracerProvider
	tracer         tracer.Tracer
	propagator     propagation.TextMapPropagator
	InitLog        *CategoryLogger
	CfgLog         *CategoryLogger
	SrvcLog        *CategoryLogger
	TstLog         *CategoryLogger
	Categories     map[string]*CategoryLogger
}

type Config struct {
	ServiceName        string
	LogFormat          string
	EnableTracing      bool
	OTLPEndpoint       string
	AdditionalResource map[string]string
}
