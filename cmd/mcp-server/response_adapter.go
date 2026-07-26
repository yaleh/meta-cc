package main

import (
	"github.com/yaleh/meta-cc/internal/config"
	filterpkg "github.com/yaleh/meta-cc/internal/filter"
	responsepkg "github.com/yaleh/meta-cc/internal/mcp/response"
)

// adaptResponse adapts CLI output to hybrid mode format (inline or file_ref).
func adaptResponse(cfg *config.Config, data []interface{}, params map[string]interface{}, toolName string, pagination *filterpkg.PaginationMetadata) (interface{}, error) {
	return responsepkg.AdaptResponse(cfg, data, params, toolName, pagination)
}

// buildInlineResponse constructs inline mode response
func buildInlineResponse(data []interface{}, pagination *filterpkg.PaginationMetadata) map[string]interface{} {
	return responsepkg.BuildInlineResponse(data, pagination)
}

// buildFileRefResponse constructs file reference mode response
func buildFileRefResponse(filePath string, data []interface{}, pagination *filterpkg.PaginationMetadata) (map[string]interface{}, error) {
	return responsepkg.BuildFileRefResponse(filePath, data, pagination)
}

// serializeResponse converts response to JSON string
func serializeResponse(response interface{}) (string, error) {
	return responsepkg.SerializeResponse(response)
}

// getSessionHash returns current session hash for temp file naming
func getSessionHash(cfg *config.Config) string {
	return responsepkg.GetSessionHash(cfg)
}
