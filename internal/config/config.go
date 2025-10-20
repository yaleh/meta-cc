// Package config provides centralized configuration management for meta-cc.
//
// Configuration follows 12-Factor App principles using environment variables
// with fail-fast validation at startup.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration.
type Config struct {
	// Log holds logging-related configuration
	Log LogConfig

	// Output holds output-related configuration
	Output OutputConfig

	// Capability holds capability source configuration
	Capability CapabilityConfig

	// Session holds session information from Claude Code
	Session SessionConfig
}

// LogConfig holds logging-related configuration.
type LogConfig struct {
	// Level controls the minimum log level for output.
	// Valid values: DEBUG, INFO, WARN, ERROR
	// Default: INFO
	Level slog.Level

	// Format specifies the log output format.
	// Valid values: text (human-readable), json (structured)
	// Default: text
	Format string

	// Enabled controls whether logging is enabled.
	// Valid values: true, false
	// Default: true
	Enabled bool
}

// OutputConfig holds output-related configuration.
type OutputConfig struct {
	// Mode controls MCP output mode.
	// Valid values: auto, inline, file_ref
	// Default: auto
	Mode string

	// InlineThreshold sets the byte threshold for switching to file_ref mode.
	// Default: 8192 (8KB)
	InlineThreshold int
}

// CapabilityConfig holds capability source configuration.
type CapabilityConfig struct {
	// Sources specifies colon-separated paths to capability directories.
	// Default: "" (uses built-in capabilities)
	Sources string
}

// SessionConfig holds session information from Claude Code.
// Note: Session information is no longer loaded from environment variables.
// This structure is retained for future use and backward compatibility.
type SessionConfig struct {
	// SessionID is the unique session identifier from Claude Code.
	// No longer populated from environment variables.
	SessionID string

	// ProjectHash is the project path hash from Claude Code.
	// No longer populated from environment variables.
	ProjectHash string
}

// Load loads configuration from environment variables with validation.
//
// Returns an error if configuration is invalid. The application should
// fail-fast on configuration errors to prevent runtime issues.
func Load() (*Config, error) {
	cfg := &Config{
		Log:        loadLogConfig(),
		Output:     loadOutputConfig(),
		Capability: loadCapabilityConfig(),
		Session:    loadSessionConfig(),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration and returns an error if invalid.
func (c *Config) Validate() error {
	// Validate log format
	if c.Log.Format != "text" && c.Log.Format != "json" {
		return fmt.Errorf("invalid META_CC_LOG_FORMAT: %s (must be 'text' or 'json')",
			c.Log.Format)
	}

	// Validate output mode
	validModes := []string{"auto", "inline", "file_ref"}
	if !contains(validModes, c.Output.Mode) {
		return fmt.Errorf("invalid META_CC_OUTPUT_MODE: %s (must be one of: %v)",
			c.Output.Mode, validModes)
	}

	// Validate inline threshold (must be positive)
	if c.Output.InlineThreshold <= 0 {
		return fmt.Errorf("META_CC_INLINE_THRESHOLD must be positive, got: %d",
			c.Output.InlineThreshold)
	}

	return nil
}

// loadLogConfig loads logging configuration from environment.
func loadLogConfig() LogConfig {
	cfg := LogConfig{
		Level:   slog.LevelInfo, // Default
		Format:  getEnvOrDefault("META_CC_LOG_FORMAT", "text"),
		Enabled: getEnvBool("META_CC_LOGGING_ENABLED", true),
	}

	// Parse log level with fallback to old LOG_LEVEL for backward compatibility
	levelStr := os.Getenv("META_CC_LOG_LEVEL")
	if levelStr == "" {
		// Fallback to LOG_LEVEL (deprecated) for MCP server compatibility
		levelStr = os.Getenv("LOG_LEVEL")
	}
	if levelStr != "" {
		cfg.Level = parseLogLevel(levelStr)
	}

	return cfg
}

// loadOutputConfig loads output configuration from environment.
func loadOutputConfig() OutputConfig {
	return OutputConfig{
		Mode:            getEnvOrDefault("META_CC_OUTPUT_MODE", "auto"),
		InlineThreshold: getEnvInt("META_CC_INLINE_THRESHOLD", 8192),
	}
}

// loadCapabilityConfig loads capability configuration from environment.
func loadCapabilityConfig() CapabilityConfig {
	return CapabilityConfig{
		Sources: os.Getenv("META_CC_CAPABILITY_SOURCES"),
	}
}

// loadSessionConfig loads session configuration from environment.
// Note: Environment variables are no longer used. This returns an empty config.
func loadSessionConfig() SessionConfig {
	return SessionConfig{
		SessionID:   "",
		ProjectHash: "",
	}
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		valueLower := strings.ToLower(value)
		return valueLower == "true" || valueLower == "1" || valueLower == "yes"
	}
	return defaultValue
}

func parseLogLevel(levelStr string) slog.Level {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// LevelString returns the string representation of the log level.
func (c LogConfig) LevelString() string {
	switch c.Level {
	case slog.LevelDebug:
		return "DEBUG"
	case slog.LevelInfo:
		return "INFO"
	case slog.LevelWarn:
		return "WARN"
	case slog.LevelError:
		return "ERROR"
	default:
		return "INFO"
	}
}

// SourcesSlice returns the capability sources as a slice of paths.
// Returns empty slice if no sources are configured.
func (c CapabilityConfig) SourcesSlice() []string {
	if c.Sources == "" {
		return nil
	}
	return strings.Split(c.Sources, ":")
}
