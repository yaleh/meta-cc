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

var errorClassificationRules = []struct {
	errorType string
	patterns  []string
}{
	{errorType: "parse_error", patterns: []string{"parse", "unmarshal", "invalid JSON", "decode"}},
	{errorType: "validation_error", patterns: []string{"validation", "invalid", "missing", "required"}},
	{errorType: "io_error", patterns: []string{"no such file", "permission denied", "cannot open", "read", "write"}},
	{errorType: "execution_error", patterns: []string{"execution", "command", "process"}},
	{errorType: "network_error", patterns: []string{"network", "connection", "timeout", "http"}},
}

// classifyError classifies an error into a category for logging
func classifyError(err error) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()
	for _, rule := range errorClassificationRules {
		if containsAny(errMsg, rule.patterns) {
			return rule.errorType
		}
	}

	return "general_error"
}

func containsAny(haystack string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(haystack, needle) {
			return true
		}
	}
	return false
}
