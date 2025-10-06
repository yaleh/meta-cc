package main

import (
	"encoding/json"
)

const (
	// DefaultInlineThresholdBytes is the size threshold for inline vs file_ref mode
	DefaultInlineThresholdBytes = 8 * 1024 // 8KB

	// Output mode constants
	OutputModeInline  = "inline"
	OutputModeFileRef = "file_ref"
)

// OutputModeConfig holds configuration for output mode selection
type OutputModeConfig struct {
	// InlineThresholdBytes is the maximum size for inline mode (default: 8KB)
	InlineThresholdBytes int
}

// DefaultOutputModeConfig returns the default configuration
func DefaultOutputModeConfig() *OutputModeConfig {
	return &OutputModeConfig{
		InlineThresholdBytes: DefaultInlineThresholdBytes,
	}
}

// calculateOutputSize measures the byte size of data when serialized to JSONL format.
//
// Parameters:
//   - data: Array of records to measure (typically []interface{} or []map[string]interface{})
//
// Returns:
//   - Total byte count including newlines
//
// The function serializes each record to JSON and counts bytes including newline separators.
// This provides accurate size measurement for determining inline vs file_ref mode.
func calculateOutputSize(data []interface{}) int {
	if len(data) == 0 {
		return 0
	}

	totalSize := 0

	for _, record := range data {
		// Serialize to JSON
		jsonBytes, err := json.Marshal(record)
		if err != nil {
			// If serialization fails, estimate conservatively
			continue
		}

		// Count JSON bytes + newline
		totalSize += len(jsonBytes) + 1 // +1 for newline
	}

	return totalSize
}

// selectOutputMode determines whether to use inline or file_ref mode based on data size.
//
// Parameters:
//   - size: Byte size of the output data (from calculateOutputSize)
//   - explicitMode: Explicit mode override ("inline", "file_ref", or "" for auto-select)
//
// Returns:
//   - "inline" for data ≤8KB or explicit inline mode
//   - "file_ref" for data >8KB or explicit file_ref mode
//
// Mode Selection Logic:
//  1. If explicitMode is "inline" or "file_ref" → use it (override)
//  2. If explicitMode is invalid/empty → auto-select based on size
//  3. Auto-select: size ≤ 8192 bytes → inline, otherwise → file_ref
//
// Override Examples:
//   - selectOutputMode(100*1024, "inline") → "inline" (force inline for large data)
//   - selectOutputMode(1*1024, "file_ref") → "file_ref" (force file_ref for small data)
func selectOutputMode(size int, explicitMode string) string {
	// Check for explicit mode override
	if explicitMode == OutputModeInline || explicitMode == OutputModeFileRef {
		return explicitMode
	}

	// Auto-select based on size threshold
	config := DefaultOutputModeConfig()

	if size <= config.InlineThresholdBytes {
		return OutputModeInline
	}

	return OutputModeFileRef
}

// selectOutputModeWithConfig is the same as selectOutputMode but allows custom configuration.
// This is useful for testing different thresholds or future configuration options.
func selectOutputModeWithConfig(size int, explicitMode string, config *OutputModeConfig) string {
	// Check for explicit mode override
	if explicitMode == OutputModeInline || explicitMode == OutputModeFileRef {
		return explicitMode
	}

	// Auto-select based on size threshold
	if size <= config.InlineThresholdBytes {
		return OutputModeInline
	}

	return OutputModeFileRef
}
