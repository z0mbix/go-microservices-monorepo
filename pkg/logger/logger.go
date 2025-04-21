package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// New returns a new slog.Logger instance with the specified log level
func New(level string) (*slog.Logger, error) {
	level = strings.ToLower(level)
	var logLevel slog.Level

	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		return nil, fmt.Errorf("invalid log level: %s", level)
	}

	logHandler := slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{Level: logLevel},
	)
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	return logger, nil
}
