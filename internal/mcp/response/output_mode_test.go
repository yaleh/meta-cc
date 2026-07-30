package response

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yaleh/meta-cc/internal/config"
)

func TestOutputModeValidationAndDefaults(t *testing.T) {
	for _, mode := range []string{OutputModeAuto, OutputModeInline, OutputModeFileRef} {
		assert.True(t, IsValidOutputMode(mode))
	}
	assert.False(t, IsValidOutputMode("unknown"))
	assert.Equal(t, DefaultInlineThresholdBytes, DefaultOutputModeConfig().InlineThresholdBytes)
}

func TestCalculateOutputSizeSkipsUnmarshalableRecords(t *testing.T) {
	assert.Zero(t, CalculateOutputSize(nil))
	assert.Equal(t, len("{\"id\":1}\n"), CalculateOutputSize([]interface{}{map[string]interface{}{"id": 1}, func() {}}))
}

func TestSelectOutputMode(t *testing.T) {
	assert.Equal(t, OutputModeInline, SelectOutputMode(999999, OutputModeInline))
	assert.Equal(t, OutputModeFileRef, SelectOutputMode(1, OutputModeFileRef))
	assert.Equal(t, OutputModeInline, SelectOutputMode(DefaultInlineThresholdBytes, OutputModeAuto))
	assert.Equal(t, OutputModeFileRef, SelectOutputMode(DefaultInlineThresholdBytes+1, ""))

	cfg := &OutputModeConfig{InlineThresholdBytes: 10}
	assert.Equal(t, OutputModeInline, SelectOutputModeWithConfig(10, "", cfg))
	assert.Equal(t, OutputModeFileRef, SelectOutputModeWithConfig(11, OutputModeAuto, cfg))
	assert.Equal(t, OutputModeInline, SelectOutputModeWithConfig(99, OutputModeInline, cfg))
	assert.Equal(t, OutputModeFileRef, SelectOutputModeWithConfig(1, OutputModeFileRef, cfg))
}

func TestGetOutputModeConfigPrecedence(t *testing.T) {
	global := &config.Config{Output: config.OutputConfig{InlineThreshold: 100}}
	assert.Equal(t, 12, GetOutputModeConfig(global, map[string]interface{}{"inline_threshold_bytes": float64(12)}).InlineThresholdBytes)
	assert.Equal(t, 13, GetOutputModeConfig(global, map[string]interface{}{"inline_threshold_bytes": 13}).InlineThresholdBytes)
	assert.Equal(t, 100, GetOutputModeConfig(global, map[string]interface{}{"inline_threshold_bytes": "bad"}).InlineThresholdBytes)
	assert.Equal(t, 100, GetOutputModeConfig(global, nil).InlineThresholdBytes)
}
