package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	// OutputModeLegacy returns raw data array (backward compatibility)
	OutputModeLegacy = "legacy"
)

// adaptResponse adapts CLI output to hybrid mode format (inline or file_ref).
// No data truncation occurs - all data is preserved via inline or file_ref mode.
//
// Parameters:
//   - data: Parsed JSONL records from CLI
//   - params: MCP tool parameters (including output_mode, inline_threshold_bytes)
//   - toolName: Name of the MCP tool being executed
//
// Returns:
//   - Inline mode: {"mode": "inline", "data": [...]}
//   - File ref mode: {"mode": "file_ref", "file_ref": {...}}
//   - Legacy mode: [...] (raw array)
func adaptResponse(data []interface{}, params map[string]interface{}, toolName string) (interface{}, error) {
	// Check for legacy mode (backward compatibility)
	if outputMode := getStringParam(params, "output_mode", ""); outputMode == OutputModeLegacy {
		return data, nil
	}

	// Determine output mode (no truncation - rely on hybrid mode)
	size := calculateOutputSize(data)
	config := getOutputModeConfig(params)
	mode := selectOutputModeWithConfig(size, getStringParam(params, "output_mode", ""), config)

	// Build response based on mode
	switch mode {
	case OutputModeInline:
		return buildInlineResponse(data), nil

	case OutputModeFileRef:
		// Create temp file
		sessionHash := getSessionHash()
		filePath := createTempFilePath(sessionHash, toolName)

		// Write data to temp file
		if err := writeJSONLFile(filePath, data); err != nil {
			return nil, fmt.Errorf("failed to write temp file: %w", err)
		}

		// Build file reference response
		return buildFileRefResponse(filePath, data)

	default:
		return nil, fmt.Errorf("unknown output mode: %s", mode)
	}
}

// buildInlineResponse constructs inline mode response
func buildInlineResponse(data []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"mode": OutputModeInline,
		"data": data,
	}
}

// buildFileRefResponse constructs file reference mode response
func buildFileRefResponse(filePath string, data []interface{}) (map[string]interface{}, error) {
	// Generate file reference metadata
	fileRef, err := generateFileReference(filePath, data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate file reference: %w", err)
	}

	// Convert FileReference struct to map
	fileRefMap := map[string]interface{}{
		"path":       fileRef.Path,
		"size_bytes": fileRef.SizeBytes,
		"line_count": fileRef.LineCount,
		"fields":     fileRef.Fields,
		"summary":    fileRef.Summary,
	}

	// Build response
	response := map[string]interface{}{
		"mode":     OutputModeFileRef,
		"file_ref": fileRefMap,
	}

	return response, nil
}

// serializeResponse converts response to JSON string
func serializeResponse(response interface{}) (string, error) {
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to serialize response: %w", err)
	}
	return string(jsonBytes), nil
}

// getSessionHash returns current session hash for temp file naming
// Falls back to "unknown" if session info is not available
func getSessionHash() string {
	// Try to get session hash from environment
	if sessionID := os.Getenv("CC_SESSION_ID"); sessionID != "" {
		// Use first 8 chars of session ID as hash
		if len(sessionID) > 8 {
			return sessionID[:8]
		}
		return sessionID
	}

	// Try to get project hash
	if projectHash := os.Getenv("CC_PROJECT_HASH"); projectHash != "" {
		if len(projectHash) > 8 {
			return projectHash[:8]
		}
		return projectHash
	}

	// Fallback to "unknown"
	return "unknown"
}
