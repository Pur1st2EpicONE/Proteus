package slog

import (
	"Proteus/internal/config"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// Logger wraps a slog.Logger and implements the Logger interface.
type Logger struct {
	logger *slog.Logger
}

// NewLogger creates a new Logger instance based on the provided configuration.
// It returns the logger and the file used for logging (or stdout if a file is not used).
func NewLogger(config config.Logger) (*Logger, *os.File) {
	var logDest *os.File
	if config.LogDir == "" {
		logDest = os.Stdout
	} else {
		logDest = openFile(config.LogDir)
		if logDest == nil {
			logDest = os.Stdout
		}
	}
	var level slog.Level
	if config.Debug {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(logDest, &slog.HandlerOptions{Level: level})
	logger := &Logger{logger: slog.New(handler)}
	slog.SetDefault(logger.logger)
	return logger, logDest
}

// openFile ensures the log directory exists and opens the log file for appending.
// Returns nil if the file cannot be created, in which case stdout will be used.
func openFile(logDir string) *os.File {
	if err := os.MkdirAll(logDir, 0777); err != nil {
		fmt.Fprintf(os.Stderr, "logger — failed to create log directory switching to stdout: %v\n", err)
		return nil
	}
	logPath := filepath.Join(logDir, "app.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger — failed to create log file switching to stdout: %v\n", err)
		return nil
	}
	return logFile
}

// LogFatal logs a fatal message with an error and exits the program.
func (l *Logger) LogFatal(msg string, err error, args ...any) {
	if err != nil {
		args = append(args, "err", err.Error())
	}
	slog.Error(msg, args...)
	os.Exit(1)
}

// LogError logs an error message with an optional error.
func (l *Logger) LogError(msg string, err error, args ...any) {
	if err != nil {
		args = append(args, "err", err.Error())
	}
	slog.Error(msg, args...)
}

// LogInfo logs an informational message.
func (l *Logger) LogInfo(msg string, args ...any) {
	slog.Info(msg, args...)
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}
