package response

import (
	"encoding/json"

	"github.com/yaleh/meta-cc/internal/config"
)

const (
	// DefaultInlineThresholdBytes is the size threshold for inline vs file_ref mode.
	// 32KB (~8K tokens) covers most MCP query results inline while staying within
	// ~5% of the 200K token context window.
	DefaultInlineThresholdBytes = 32 * 1024 // 32KB

	// Output mode constants
	OutputModeAuto    = "auto"
	OutputModeInline  = "inline"
	OutputModeFileRef = "file_ref"
)

// IsValidOutputMode returns true if mode is a recognised output_mode value.
func IsValidOutputMode(mode string) bool {
	switch mode {
	case OutputModeAuto, OutputModeInline, OutputModeFileRef:
		return true
	}
	return false
}

// OutputModeConfig holds configuration for output mode selection
type OutputModeConfig struct {
	// InlineThresholdBytes is the maximum size for inline mode (default: 32KB)
	InlineThresholdBytes int
}

// DefaultOutputModeConfig returns the default configuration
func DefaultOutputModeConfig() *OutputModeConfig {
	return &OutputModeConfig{
		InlineThresholdBytes: DefaultInlineThresholdBytes,
	}
}

// CalculateOutputSize measures the byte size of data when serialized to JSONL format.
func CalculateOutputSize(data []interface{}) int {
	if len(data) == 0 {
		return 0
	}

	totalSize := 0
	for _, record := range data {
		jsonBytes, err := json.Marshal(record)
		if err != nil {
			continue
		}
		totalSize += len(jsonBytes) + 1 // +1 for newline
	}

	return totalSize
}

// SelectOutputMode determines whether to use inline or file_ref mode based on data size.
func SelectOutputMode(size int, explicitMode string) string {
	if explicitMode == OutputModeInline || explicitMode == OutputModeFileRef {
		return explicitMode
	}

	cfg := DefaultOutputModeConfig()
	if size <= cfg.InlineThresholdBytes {
		return OutputModeInline
	}
	return OutputModeFileRef
}

// SelectOutputModeWithConfig is the same as SelectOutputMode but allows custom configuration.
func SelectOutputModeWithConfig(size int, explicitMode string, cfg *OutputModeConfig) string {
	if explicitMode == OutputModeInline || explicitMode == OutputModeFileRef {
		return explicitMode
	}
	if size <= cfg.InlineThresholdBytes {
		return OutputModeInline
	}
	return OutputModeFileRef
}

// GetOutputModeConfig returns output mode configuration from centralized config and parameters.
// Priority: parameter > centralized config (from environment) > default (32768 bytes)
func GetOutputModeConfig(globalCfg *config.Config, params map[string]interface{}) *OutputModeConfig {
	cfg := DefaultOutputModeConfig()

	if thresholdParam, ok := params["inline_threshold_bytes"]; ok {
		if threshold, ok := thresholdParam.(float64); ok {
			cfg.InlineThresholdBytes = int(threshold)
			return cfg
		}
		if threshold, ok := thresholdParam.(int); ok {
			cfg.InlineThresholdBytes = threshold
			return cfg
		}
	}

	cfg.InlineThresholdBytes = globalCfg.Output.InlineThreshold
	return cfg
}
