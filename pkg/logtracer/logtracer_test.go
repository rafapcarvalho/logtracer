package logtracer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rafapcarvalho/logtracer/internal/handlers"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"path/filepath"
	"testing"
)

func TestInitLogger(t *testing.T) {
	cleanup := func() {
		InitLog = nil
		CfgLog = nil
		SrvcLog = nil
		TstLog = nil
		Categories = nil
		traceProvider = nil
		globalTracer = nil
	}

	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "JSON format without tracing",
			cfg: Config{
				ServiceName:   "test-service",
				LogFormat:     "json",
				EnableTracing: false,
			},
		},
		{
			name: "Text format with tracing",
			cfg: Config{
				ServiceName:   "test-service",
				LogFormat:     "text",
				EnableTracing: true,
				OTLPEndpoint:  "localhost:4318",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup()

			InitLogger(tt.cfg)

			if InitLog == nil || InitLog.logger == nil {
				t.Error("InitLog was not properly initialized")
			}
			assert.NotNil(t, InitLog)
			assert.NotNil(t, CfgLog)
			assert.NotNil(t, SrvcLog)
			assert.NotNil(t, TstLog)
			assert.NotNil(t, NoTrace)
			assert.NotNil(t, Categories)

			if tt.cfg.EnableTracing {
				assert.NotNil(t, globalTracer, "Tracer should be initialized when tracing is enabled")
			}

			ctx := context.Background()
			InitLog.Info(ctx, "Test message")
		})
	}
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	handlers.LoggerLevel = new(slog.LevelVar)
	handlers.LoggerLevel.Set(slog.LevelDebug)
	// Setup logger with a buffer for capturing output
	l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		AddSource: true,
		Level:     handlers.LoggerLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				if src, ok := a.Value.Any().(*slog.Source); ok {
					function := filepath.Base(src.Function)
					file := filepath.Base(src.File)
					a.Value = slog.StringValue(fmt.Sprintf("[%s] %s:%d", function, file, src.Line))
				}
			}
			return a
		},
	}))

	SrvcLog = newCategoryLogger(l, "test-service", "SRVC")
	SetLevel(LevelDebug)
	tests := []struct {
		name    string
		logFunc func(ctx context.Context, msg string, args ...any)
		msg     string
		want    string
	}{
		{
			name:    "Info level log",
			logFunc: SrvcLog.Info,
			msg:     "http://info test message/teste1/teste2/a",
			want:    `{"level":"INFO","msg":"http://info test message/teste1/teste2/a","component":"test-service","category":"SRVC"}`,
		},
		{
			name:    "Error level log",
			logFunc: SrvcLog.Error,
			msg:     "error test message",
			want:    `{"level":"ERROR","msg":"error test message","component":"test-service","category":"SRVC"}`,
		},
		{
			name:    "Warn level log",
			logFunc: SrvcLog.Warn,
			msg:     "warn test message",
			want:    `{"level":"WARN","msg":"warn test message","component":"test-service","category":"SRVC"}`,
		},
		{
			name:    "Debug level log",
			logFunc: SrvcLog.Debug,
			msg:     "debug test message",
			want:    `{"level":"DEBUG","msg":"debug test message","component":"test-service","category":"SRVC"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			ctx := context.Background()
			tt.logFunc(ctx, tt.msg)
			if buf.Len() == 0 {
				t.Fatal("No log output captured")
			}
			checkLogOutput(t, buf.String(), tt.want, "time", "source")
		})
	}
}

func TestLogWithAttributes(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				if src, ok := a.Value.Any().(*slog.Source); ok {
					function := filepath.Base(src.Function)
					file := filepath.Base(src.File)
					a.Value = slog.StringValue(fmt.Sprintf("[%s] %s:%d", function, file, src.Line))
				}
			}
			return a
		},
	}))

	SrvcLog = newCategoryLogger(l, "test-service", "SRVC")

	tests := []struct {
		name string
		args []any
		want string
	}{
		{
			name: "Single string attribute",
			args: []any{"key1", "value1"},
			want: `{"level":"INFO","msg":"test message","component":"test-service","category":"SRVC","key1":"value1"}`,
		},
		{
			name: "Multiple attributes",
			args: []any{"key1", "value1", "key2", 123, "key3", true},
			want: `{"level":"INFO","msg":"test message","component":"test-service","category":"SRVC","key1":"value1","key2":123,"key3":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			ctx := context.Background()
			SrvcLog.Info(ctx, "test message", tt.args...)
			checkLogOutput(t, buf.String(), tt.want, "time", "source")
		})
	}
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func checkLogOutput(t *testing.T, got, want string, skip ...string) {
	t.Helper()
	gotObj := make(map[string]any)
	wantObj := make(map[string]any)

	if err := json.Unmarshal([]byte(got), &gotObj); err != nil {
		t.Errorf("got invalid JSON: %v -> %s", err, got)
		return
	}
	if err := json.Unmarshal([]byte(want), &wantObj); err != nil {
		t.Errorf("want invalid JSON: %v -> %s", err, want)
		return
	}

	skip = append(skip, "time") // Always skip time comparison
	for k, v := range wantObj {
		if !contains(skip, k) {
			w, has := gotObj[k]
			if !has {
				t.Errorf("got missing key %s, want %v", k, v)
			}
			if fmt.Sprint(w) != fmt.Sprint(v) {
				t.Errorf("key '%s' got '%v', want '%v'", k, w, v)
			}
		}
	}
}
