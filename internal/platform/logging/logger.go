package logging

import (
	"log/slog"
	"os"
)

// New returns a stdlib slog.Logger writing structured text logs to stdout.
func New() *slog.Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(handler)
}
