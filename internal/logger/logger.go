// Package logger configures structured logging to file and stderr.
package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// New returns a logger writing to stderr and to <logDir>/genetica.log.
// The returned closer must be called on shutdown.
func New(logDir, level string) (*slog.Logger, io.Closer, error) {
	f, err := os.OpenFile(filepath.Join(logDir, "genetica.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, nil, err
	}
	h := slog.NewTextHandler(io.MultiWriter(os.Stderr, f), &slog.HandlerOptions{Level: parseLevel(level)})
	return slog.New(h), f, nil
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
