package logtracer

import (
	"github.com/rafapcarvalho/logtracer/internal/handlers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"strings"
)

var (
	ginLog     *CategoryLogger
	grpcLog    *CategoryLogger
	noTrace    *CategoryLogger
	InitLog    *CategoryLogger
	CfgLog     *CategoryLogger
	SrvcLog    *CategoryLogger
	TstLog     *CategoryLogger
	Categories map[string]*CategoryLogger
	NoTrace    WithoutTracer
	customID   CustomID

	globalTracer  trace.Tracer
	traceProvider *sdktrace.TracerProvider
	propagator    propagation.TextMapPropagator
	// shutdownOnce  sync.Once
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
			logger.Error("Failed to initialize tracer provider", "error", err)
		}
	}

	if cfg.CustomID != "" {
		customID = CustomID(cfg.CustomID)
	}

	InitLog = newCategoryLogger(logger, cfg.ServiceName, "INIT")
	CfgLog = newCategoryLogger(logger, cfg.ServiceName, "CFG")
	SrvcLog = newCategoryLogger(logger, cfg.ServiceName, "SRVC")
	TstLog = newCategoryLogger(logger, cfg.ServiceName, "TEST")
	ginLog = newCategoryLogger(logger, cfg.ServiceName, "GIN")
	grpcLog = newCategoryLogger(logger, cfg.ServiceName, "GRPC")
	noTrace = newCategoryLogger(logger, cfg.ServiceName, "WithoutTrace")
	NoTrace = WithoutTracer{}

	Categories = make(map[string]*CategoryLogger)
}

func newCategoryLogger(logger *slog.Logger, serviceName, category string) *CategoryLogger {
	return &CategoryLogger{
		logger: logger.With("component", serviceName, "category", category),
	}
}
