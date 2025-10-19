package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/yaleh/meta-cc/internal/config"
)

// loggerContextKey is the context key for storing the logger
type loggerContextKey struct{}

// InitLogger initializes the global slog logger with centralized configuration
func InitLogger(cfg *config.Config) {
	// Use log level from centralized config
	// Config handles LOG_LEVEL fallback for backward compatibility
	logLevel := cfg.Log.Level

	// Create JSON handler for structured logging
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true, // Include file:line for debugging
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// NewRequestLogger creates a request-scoped logger with request_id and tool_name
func NewRequestLogger(toolName string) (*slog.Logger, string) {
	requestID := uuid.New().String()
	logger := slog.Default().With(
		"request_id", requestID,
		"tool_name", toolName,
	)
	return logger, requestID
}

// WithLogger attaches a logger to the context
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

// LoggerFromContext retrieves the logger from the context
// If no logger is found, returns the default logger
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerContextKey{}).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// classifyError classifies an error into a category for logging
func classifyError(err error) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()

	// Parse errors
	if strings.Contains(errMsg, "parse") || strings.Contains(errMsg, "unmarshal") ||
		strings.Contains(errMsg, "invalid JSON") || strings.Contains(errMsg, "decode") {
		return "parse_error"
	}

	// Validation errors
	if strings.Contains(errMsg, "validation") || strings.Contains(errMsg, "invalid") ||
		strings.Contains(errMsg, "missing") || strings.Contains(errMsg, "required") {
		return "validation_error"
	}

	// I/O errors
	if strings.Contains(errMsg, "no such file") || strings.Contains(errMsg, "permission denied") ||
		strings.Contains(errMsg, "cannot open") || strings.Contains(errMsg, "read") ||
		strings.Contains(errMsg, "write") {
		return "io_error"
	}

	// Execution errors
	if strings.Contains(errMsg, "execution") || strings.Contains(errMsg, "command") ||
		strings.Contains(errMsg, "process") {
		return "execution_error"
	}

	// Network errors
	if strings.Contains(errMsg, "network") || strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "http") {
		return "network_error"
	}

	// Default: general error
	return "general_error"
}
