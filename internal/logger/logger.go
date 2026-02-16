package logger

import (
	"Proteus/internal/config"
	"Proteus/internal/logger/slog"
	"io"
	"log"
	"os"

	"github.com/pressly/goose/v3"
)

// Logger defines the interface for structured logging with different severity levels.
type Logger interface {
	// LogFatal logs a fatal message with an error and optional key-value arguments.
	LogFatal(msg string, err error, args ...any)
	// LogError logs an error message with an error and optional key-value arguments.
	LogError(string, error, ...any)
	// LogInfo logs an informational message with optional key-value arguments.
	LogInfo(msg string, args ...any)
	// Debug logs a debug message with optional key-value arguments.
	Debug(msg string, args ...any)
}

// NewLogger creates a new Logger instance based on the provided configuration.
// Returns the logger and an *os.File if a file is used for logging.
func NewLogger(config config.Logger) (Logger, *os.File) {
	goose.SetLogger(log.New(io.Discard, "", 0))
	return slog.NewLogger(config)
}
