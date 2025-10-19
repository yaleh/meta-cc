// Package errors defines sentinel errors for common error conditions across meta-cc.
//
// Sentinel errors enable consistent error handling and programmatic error checking
// using errors.Is() and errors.As(). They follow the Go 1.13+ error wrapping conventions.
//
// Usage:
//
//	// Wrapping with context:
//	if user == nil {
//	    return fmt.Errorf("failed to load user %s: %w", userID, errors.ErrNotFound)
//	}
//
//	// Checking programmatically:
//	if errors.Is(err, errors.ErrNotFound) {
//	    // Handle not found case
//	}
package errors

import "errors"

// Sentinel errors for common conditions across meta-cc.
//
// These errors are designed to be wrapped with context using fmt.Errorf with %w:
//
//	return fmt.Errorf("operation failed for %s: %w", context, errors.ErrXxx)
var (
	// ErrNotFound indicates a requested resource was not found.
	// Use for missing files, sessions, records, etc.
	ErrNotFound = errors.New("not found")

	// ErrInvalidInput indicates input validation failed.
	// Use for out-of-range values, invalid formats, constraint violations.
	ErrInvalidInput = errors.New("invalid input")

	// ErrMissingParameter indicates a required parameter was not provided.
	// Use for missing CLI arguments, missing MCP tool parameters.
	ErrMissingParameter = errors.New("required parameter missing")

	// ErrUnknownTool indicates an unsupported tool was requested.
	// Use when MCP tool name is not recognized.
	ErrUnknownTool = errors.New("unknown tool")

	// ErrTimeout indicates an operation exceeded its time limit.
	// Use for operations with time constraints.
	ErrTimeout = errors.New("operation timeout")

	// ErrFileIO indicates a file I/O operation failed.
	// Use for file read/write/create/delete failures, directory operations.
	ErrFileIO = errors.New("file I/O error")

	// ErrNetworkFailure indicates a network operation failed.
	// Use for HTTP requests, downloads, connection failures.
	ErrNetworkFailure = errors.New("network operation failed")

	// ErrParseError indicates parsing or deserialization failed.
	// Use for JSON/YAML/TOML parsing errors, invalid format errors.
	ErrParseError = errors.New("parsing failed")

	// ErrConfigError indicates a configuration error.
	// Use for invalid config values, missing required config, config validation failures.
	ErrConfigError = errors.New("configuration error")
)
