package config

import (
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	// Clear environment
	clearTestEnv(t)

	// Set valid configuration
	os.Setenv("META_CC_LOG_LEVEL", "DEBUG")
	os.Setenv("META_CC_LOG_FORMAT", "json")
	os.Setenv("META_CC_OUTPUT_MODE", "inline")
	os.Setenv("META_CC_INLINE_THRESHOLD", "4096")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Log.Level != slog.LevelDebug {
		t.Errorf("expected DEBUG level, got %v", cfg.Log.Level)
	}

	if cfg.Log.Format != "json" {
		t.Errorf("expected json format, got %s", cfg.Log.Format)
	}

	if cfg.Output.Mode != "inline" {
		t.Errorf("expected inline mode, got %s", cfg.Output.Mode)
	}

	if cfg.Output.InlineThreshold != 4096 {
		t.Errorf("expected 4096 threshold, got %d", cfg.Output.InlineThreshold)
	}
}

func TestLoadDefaults(t *testing.T) {
	// Clear all config env vars
	clearTestEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("should load with defaults: %v", err)
	}

	// Verify defaults
	if cfg.Log.Level != slog.LevelInfo {
		t.Errorf("default log level should be INFO, got %v", cfg.Log.Level)
	}

	if cfg.Log.Format != "text" {
		t.Errorf("default log format should be text, got %s", cfg.Log.Format)
	}

	if !cfg.Log.Enabled {
		t.Error("logging should be enabled by default")
	}

	if cfg.Output.Mode != "auto" {
		t.Errorf("default output mode should be auto, got %s", cfg.Output.Mode)
	}

	if cfg.Output.InlineThreshold != 8192 {
		t.Errorf("default inline threshold should be 8192, got %d", cfg.Output.InlineThreshold)
	}
}

func TestValidateInvalidLogFormat(t *testing.T) {
	clearTestEnv(t)
	os.Setenv("META_CC_LOG_FORMAT", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error for invalid log format")
	}

	if !strings.Contains(err.Error(), "invalid META_CC_LOG_FORMAT") {
		t.Errorf("error should mention invalid log format: %v", err)
	}
}

func TestValidateInvalidOutputMode(t *testing.T) {
	clearTestEnv(t)
	os.Setenv("META_CC_OUTPUT_MODE", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error for invalid output mode")
	}

	if !strings.Contains(err.Error(), "invalid META_CC_OUTPUT_MODE") {
		t.Errorf("error should mention invalid output mode: %v", err)
	}
}

func TestValidateInvalidInlineThreshold(t *testing.T) {
	clearTestEnv(t)
	os.Setenv("META_CC_INLINE_THRESHOLD", "-100")

	_, err := Load()
	if err == nil {
		t.Fatal("expected validation error for invalid inline threshold")
	}

	if !strings.Contains(err.Error(), "META_CC_INLINE_THRESHOLD must be positive") {
		t.Errorf("error should mention positive threshold: %v", err)
	}
}

func TestLogLevelParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"DEBUG", slog.LevelDebug},
		{"debug", slog.LevelDebug},
		{"INFO", slog.LevelInfo},
		{"info", slog.LevelInfo},
		{"WARN", slog.LevelWarn},
		{"WARNING", slog.LevelWarn},
		{"ERROR", slog.LevelError},
		{"error", slog.LevelError},
		{"invalid", slog.LevelInfo}, // Falls back to INFO
		{"", slog.LevelInfo},        // Default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level := parseLogLevel(tt.input)
			if level != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, level, tt.expected)
			}
		})
	}
}

func TestLogLevelFallback(t *testing.T) {
	// Test backward compatibility: LOG_LEVEL fallback
	clearTestEnv(t)
	os.Setenv("LOG_LEVEL", "DEBUG")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Log.Level != slog.LevelDebug {
		t.Errorf("should use LOG_LEVEL fallback, got %v", cfg.Log.Level)
	}
}

func TestLogLevelPriority(t *testing.T) {
	// META_CC_LOG_LEVEL takes priority over LOG_LEVEL
	clearTestEnv(t)
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("META_CC_LOG_LEVEL", "ERROR")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Log.Level != slog.LevelError {
		t.Errorf("META_CC_LOG_LEVEL should take priority, got %v", cfg.Log.Level)
	}
}

func TestLoggingEnabled(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"TRUE", true},
		{"1", true},
		{"yes", true},
		{"YES", true},
		{"false", false},
		{"FALSE", false},
		{"0", false},
		{"no", false},
		{"", true}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			clearTestEnv(t)
			if tt.input != "" {
				os.Setenv("META_CC_LOGGING_ENABLED", tt.input)
			}

			cfg, err := Load()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.Log.Enabled != tt.expected {
				t.Errorf("META_CC_LOGGING_ENABLED=%q: expected %v, got %v",
					tt.input, tt.expected, cfg.Log.Enabled)
			}
		})
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    slog.Level
		expected string
	}{
		{slog.LevelDebug, "DEBUG"},
		{slog.LevelInfo, "INFO"},
		{slog.LevelWarn, "WARN"},
		{slog.LevelError, "ERROR"},
	}

	for _, tt := range tests {
		cfg := LogConfig{Level: tt.level}
		if got := cfg.LevelString(); got != tt.expected {
			t.Errorf("LevelString() = %q, want %q", got, tt.expected)
		}
	}
}

func TestCapabilitySourcesSlice(t *testing.T) {
	tests := []struct {
		sources  string
		expected []string
	}{
		{"", nil},
		{"/path/one", []string{"/path/one"}},
		{"/path/one:/path/two", []string{"/path/one", "/path/two"}},
		{"/path/one:/path/two:/path/three", []string{"/path/one", "/path/two", "/path/three"}},
	}

	for _, tt := range tests {
		cfg := CapabilityConfig{Sources: tt.sources}
		got := cfg.SourcesSlice()

		if len(got) != len(tt.expected) {
			t.Errorf("SourcesSlice(%q) length = %d, want %d", tt.sources, len(got), len(tt.expected))
			continue
		}

		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("SourcesSlice(%q)[%d] = %q, want %q", tt.sources, i, got[i], tt.expected[i])
			}
		}
	}
}

func TestSessionConfig(t *testing.T) {
	clearTestEnv(t)
	os.Setenv("CC_SESSION_ID", "test-session-123")
	os.Setenv("CC_PROJECT_HASH", "-home-user-project")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Session.SessionID != "test-session-123" {
		t.Errorf("expected session ID test-session-123, got %s", cfg.Session.SessionID)
	}

	if cfg.Session.ProjectHash != "-home-user-project" {
		t.Errorf("expected project hash -home-user-project, got %s", cfg.Session.ProjectHash)
	}
}

// Test helpers

func clearTestEnv(t *testing.T) {
	t.Helper()
	envVars := []string{
		"META_CC_LOG_LEVEL",
		"META_CC_LOG_FORMAT",
		"META_CC_LOGGING_ENABLED",
		"META_CC_OUTPUT_MODE",
		"META_CC_INLINE_THRESHOLD",
		"META_CC_CAPABILITY_SOURCES",
		"LOG_LEVEL", // Deprecated fallback
		"CC_SESSION_ID",
		"CC_PROJECT_HASH",
	}

	for _, v := range envVars {
		os.Unsetenv(v)
	}
}
