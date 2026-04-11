package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

func Init(env string) {
	var handler slog.Handler

	switch env {
	case "release":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	default:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})
	}

	Log = slog.New(handler)
	slog.SetDefault(Log)
}
