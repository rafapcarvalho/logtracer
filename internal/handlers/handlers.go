package handlers

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

var LoggerLevel = new(slog.LevelVar)

func StdoutJSON() slog.Handler {
	return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       LoggerLevel,
		ReplaceAttr: replace,
	})
}

func StdoutTXT() slog.Handler {
	return slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       LoggerLevel,
		ReplaceAttr: replace,
	})
}

func replace(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		if src, ok := a.Value.Any().(*slog.Source); ok {
			function := filepath.Base(src.Function) // Pega apenas o nome da função, sem o pacote
			file := filepath.Base(src.File)
			formattedSource := fmt.Sprintf("[%s] %s:%d", function, file, src.Line)
			return slog.Attr{
				Key:   "source",
				Value: slog.StringValue(formattedSource),
			}
		}
	}
	return a
}
