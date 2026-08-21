package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

var defaultLogger *slog.Logger

func Init(level string, env string) *slog.Logger {
	var lv slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lv = slog.LevelDebug
	case "warn":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	if strings.EqualFold(env, "production") && lv < slog.LevelInfo {
		lv = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: lv}
	var w io.Writer = os.Stdout
	defaultLogger = slog.New(slog.NewJSONHandler(w, opts))
	slog.SetDefault(defaultLogger)
	return defaultLogger
}

func L() *slog.Logger {
	if defaultLogger == nil {
		return slog.Default()
	}
	return defaultLogger
}
