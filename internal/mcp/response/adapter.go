package response

import (
	"encoding/json"
	"fmt"

	"github.com/yaleh/meta-cc/internal/config"
	mcerrors "github.com/yaleh/meta-cc/internal/errors"
	filterpkg "github.com/yaleh/meta-cc/internal/filter"
)

// AdaptResponse adapts data to hybrid mode format (inline or file_ref).
// paginationMeta carries pagination metadata for the response envelope (may be nil for backward compatibility).
func AdaptResponse(cfg *config.Config, data []interface{}, params map[string]interface{}, toolName string, pagination *filterpkg.PaginationMetadata) (interface{}, error) {
	size := CalculateOutputSize(data)
	modeCfg := GetOutputModeConfig(cfg, params)
	mode := SelectOutputModeWithConfig(size, getStringParam(params, "output_mode", ""), modeCfg)

	switch mode {
	case OutputModeInline:
		return BuildInlineResponse(data, pagination), nil

	case OutputModeFileRef:
		sessionHash := GetSessionHash(cfg)
		filePath := CreateTempFilePath(sessionHash, toolName)

		if err := WriteJSONLFile(filePath, data); err != nil {
			return nil, fmt.Errorf("failed to write temp file %s: %w", filePath, mcerrors.ErrFileIO)
		}

		return BuildFileRefResponse(filePath, data, pagination)

	default:
		return nil, fmt.Errorf("unknown output mode '%s' in AdaptResponse: %w", mode, mcerrors.ErrInvalidInput)
	}
}

// BuildInlineResponse constructs inline mode response.
// A nil data slice is normalized to an empty slice so JSON serialization
// produces "data":[] instead of "data":null (fixes query_summaries null return).
// pagination may be nil for backward compatibility with callers that don't use pagination.
func BuildInlineResponse(data []interface{}, pagination *filterpkg.PaginationMetadata) map[string]interface{} {
	if data == nil {
		data = []interface{}{}
	}
	resp := map[string]interface{}{
		"mode": OutputModeInline,
		"data": data,
	}
	if pagination != nil {
		resp["pagination"] = pagination
	}
	return resp
}

// BuildFileRefResponse constructs file reference mode response.
// pagination may be nil for backward compatibility with callers that don't use pagination.
func BuildFileRefResponse(filePath string, data []interface{}, pagination *filterpkg.PaginationMetadata) (map[string]interface{}, error) {
	fileRef, err := GenerateFileReference(filePath, data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate file reference for %s: %w", filePath, mcerrors.ErrFileIO)
	}

	fileRefMap := map[string]interface{}{
		"path":       fileRef.Path,
		"size_bytes": fileRef.SizeBytes,
		"line_count": fileRef.LineCount,
		"fields":     fileRef.Fields,
		"summary":    fileRef.Summary,
	}

	resp := map[string]interface{}{
		"mode":     OutputModeFileRef,
		"file_ref": fileRefMap,
	}
	if pagination != nil {
		resp["pagination"] = pagination
	}
	return resp, nil
}

// SerializeResponse converts response to JSON string
func SerializeResponse(response interface{}) (string, error) {
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to serialize response to JSON: %w", mcerrors.ErrParseError)
	}
	return string(jsonBytes), nil
}

// GetSessionHash returns current session hash for temp file naming
func GetSessionHash(cfg *config.Config) string {
	sessionID := cfg.Session.SessionID
	if sessionID != "" {
		if len(sessionID) > 8 {
			return sessionID[:8]
		}
		return sessionID
	}

	projectHash := cfg.Session.ProjectHash
	if projectHash != "" {
		if len(projectHash) > 8 {
			return projectHash[:8]
		}
		return projectHash
	}

	return "unknown"
}

// getStringParam is a local helper
func getStringParam(args map[string]interface{}, key, defaultVal string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return defaultVal
}
