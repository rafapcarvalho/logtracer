package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/rafapcarvalho/logtracer/examples/gin-exampe1/file1"
	logger "github.com/rafapcarvalho/logtracer/pkg/logtracer"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"
)

func main() {
	cfg := logger.Config{
		ServiceName:   "example-service",
		LogFormat:     "json",
		EnableTracing: true,
		OTLPEndpoint:  "localhost:4318",
		AdditionalResource: map[string]string{
			"environment": "production",
		},
	}

	logger.InitLogger(cfg)

	r := gin.New()
	r.Use(logger.GinMiddleware(cfg.ServiceName))
	r.Use(logger.OTELMiddleware(cfg.ServiceName))

	r.GET("/example", func(c *gin.Context) {
		ctx, span := logger.StartSpan(c.Request.Context(), "example-handler")
		defer span.End()

		logger.AddAttribute(ctx, "custom.attribute", "diferente")
		logger.SrvcLog.Warn(ctx, "handling example request")
		file1.CallFile1(ctx)
		c.JSON(http.StatusOK, gin.H{"message": "hello, World!"})
	})
	r.GET("/error", func(c *gin.Context) {
		ctx, span := logger.StartSpan(c.Request.Context(), "example-handler")
		defer span.End()

		logger.AddAttribute(ctx, "custom.attribute", "hendler de teste de error")
		logger.SrvcLog.Warn(ctx, "handling error request")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "hello, World!"})
	})

	srv := &http.Server{
		Addr:    ":8091",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.SrvcLog.Error(context.Background(), "Failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.SrvcLog.Info(context.Background(), "Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.SrvcLog.Errorf(ctx, "Server forced to shutdown error: %v", err)
	}

	if err := logger.Shutdown(ctx); err != nil {
		logger.SrvcLog.Errorf(ctx, "Error shutting down tracer, error: %v", err)
	}

	logger.SrvcLog.Info(context.Background(), "Server exiting")
}