package logtracer

import (
	"github.com/rafapcarvalho/logtracer/internal/handlers"
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

func getLogLevel(level LogLevel) slog.Level {
	var newLevel slog.Level
	switch level {
	case LevelError:
		newLevel = slog.LevelError
	case LevelWarn:
		newLevel = slog.LevelWarn
	case LevelDebug:
		newLevel = slog.LevelDebug
	default:
		newLevel = slog.LevelInfo
	}
	return newLevel
}
