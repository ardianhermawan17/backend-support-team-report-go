package logger

import (
	"context"
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

func New(service, env string) *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	base := slog.New(handler).With("service", service, "env", env)
	return &Logger{Logger: base}
}

func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.Logger.InfoContext(ctx, msg, args...)
}
