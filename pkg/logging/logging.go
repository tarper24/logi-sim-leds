package logging

import (
	"log/slog"
	"os"
)

// Setup configures the default slog logger.
// If debug is true, log level is Debug; otherwise Warn.
// If logPath is non-empty, output goes to that file as JSON; otherwise stderr as text.
func Setup(debug bool, logPath string) error {
	level := slog.LevelWarn
	if debug {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}

	if logPath != "" {
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		slog.SetDefault(slog.New(slog.NewJSONHandler(f, opts)))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, opts)))
	}

	return nil
}
