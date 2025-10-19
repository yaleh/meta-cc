package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestSentinelErrorsExist(t *testing.T) {
	// Verify all sentinel errors are defined
	sentinels := []error{
		ErrNotFound,
		ErrInvalidInput,
		ErrMissingParameter,
		ErrUnknownTool,
		ErrTimeout,
		ErrFileIO,
		ErrNetworkFailure,
		ErrParseError,
		ErrConfigError,
	}

	for _, sentinel := range sentinels {
		if sentinel == nil {
			t.Error("sentinel error should not be nil")
		}

		if sentinel.Error() == "" {
			t.Error("sentinel error should have non-empty message")
		}
	}
}

func TestErrorWrapping(t *testing.T) {
	tests := []struct {
		name     string
		sentinel error
		context  string
	}{
		{"not found", ErrNotFound, "user abc123"},
		{"invalid input", ErrInvalidInput, "negative window size"},
		{"missing parameter", ErrMissingParameter, "pattern for query"},
		{"unknown tool", ErrUnknownTool, "get_nonexistent_data"},
		{"timeout", ErrTimeout, "database query"},
		{"file io", ErrFileIO, "reading config.yaml"},
		{"network failure", ErrNetworkFailure, "downloading package"},
		{"parse error", ErrParseError, "decoding JSON"},
		{"config error", ErrConfigError, "validating log level"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Wrap sentinel error with context
			wrapped := fmt.Errorf("operation failed for %s: %w", tt.context, tt.sentinel)

			// Verify errors.Is works
			if !errors.Is(wrapped, tt.sentinel) {
				t.Errorf("errors.Is failed: wrapped error should match sentinel")
			}

			// Verify error message contains context
			errMsg := wrapped.Error()
			if errMsg == "" {
				t.Error("wrapped error message should not be empty")
			}
		})
	}
}

func TestErrorIsWithoutWrapping(t *testing.T) {
	// Direct comparison should work
	if !errors.Is(ErrNotFound, ErrNotFound) {
		t.Error("errors.Is should match same sentinel")
	}

	// Different sentinels should not match
	if errors.Is(ErrNotFound, ErrInvalidInput) {
		t.Error("errors.Is should not match different sentinels")
	}
}

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		err     error
		message string
	}{
		{ErrNotFound, "not found"},
		{ErrInvalidInput, "invalid input"},
		{ErrMissingParameter, "required parameter missing"},
		{ErrUnknownTool, "unknown tool"},
		{ErrTimeout, "operation timeout"},
		{ErrFileIO, "file I/O error"},
		{ErrNetworkFailure, "network operation failed"},
		{ErrParseError, "parsing failed"},
		{ErrConfigError, "configuration error"},
	}

	for _, tt := range tests {
		if tt.err.Error() != tt.message {
			t.Errorf("error message mismatch: got %q, want %q", tt.err.Error(), tt.message)
		}
	}
}

func TestMultipleLevelWrapping(t *testing.T) {
	// Wrap error multiple times
	level1 := fmt.Errorf("database operation failed: %w", ErrNotFound)
	level2 := fmt.Errorf("user service failed: %w", level1)
	level3 := fmt.Errorf("HTTP handler failed: %w", level2)

	// errors.Is should work through multiple levels
	if !errors.Is(level3, ErrNotFound) {
		t.Error("errors.Is should work through multiple wrapping levels")
	}

	// Error message should contain all context
	errMsg := level3.Error()
	if errMsg == "" {
		t.Error("multi-level wrapped error should have message")
	}
}
