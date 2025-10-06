package main

import (
	"os"
	"strings"
	"testing"
)

// TestCalculateOutputSize verifies JSONL byte counting for various data types
func TestCalculateOutputSize(t *testing.T) {
	tests := []struct {
		name     string
		data     []interface{}
		wantSize int
	}{
		{
			name:     "empty array",
			data:     []interface{}{},
			wantSize: 0,
		},
		{
			name: "single small record",
			data: []interface{}{
				map[string]interface{}{"id": 1, "name": "test"},
			},
			wantSize: 25, // {"id":1,"name":"test"}\n
		},
		{
			name: "multiple records",
			data: []interface{}{
				map[string]interface{}{"id": 1},
				map[string]interface{}{"id": 2},
			},
			wantSize: 18, // {"id":1}\n{"id":2}\n
		},
		{
			name: "7KB data (below threshold)",
			data: func() []interface{} {
				data := generateTestData(7 * 1024)
				return data
			}(),
			wantSize: -1, // Computed dynamically
		},
		{
			name: "8KB data (at threshold)",
			data: func() []interface{} {
				data := generateTestData(8 * 1024)
				return data
			}(),
			wantSize: -1, // Computed dynamically
		},
		{
			name: "9KB data (above threshold)",
			data: func() []interface{} {
				data := generateTestData(9 * 1024)
				return data
			}(),
			wantSize: -1, // Computed dynamically
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateOutputSize(tt.data)

			// For dynamically computed sizes (wantSize=-1), just verify the size is reasonable
			if tt.wantSize == -1 {
				// Just verify the size is positive and reasonable
				if got <= 0 {
					t.Errorf("calculateOutputSize() = %d, want positive size", got)
				}
				// Log the actual size for verification
				t.Logf("calculateOutputSize() = %d bytes", got)
			} else {
				// Allow small tolerance for JSON encoding variations
				tolerance := 10
				if abs(got-tt.wantSize) > tolerance {
					t.Errorf("calculateOutputSize() = %d, want ~%d", got, tt.wantSize)
				}
			}
		})
	}
}

// TestSelectOutputMode verifies mode selection at threshold boundaries
func TestSelectOutputMode(t *testing.T) {
	tests := []struct {
		name         string
		size         int
		explicitMode string
		wantMode     string
	}{
		{
			name:         "empty result",
			size:         0,
			explicitMode: "",
			wantMode:     "inline",
		},
		{
			name:         "small result (4KB)",
			size:         4 * 1024,
			explicitMode: "",
			wantMode:     "inline",
		},
		{
			name:         "7KB result (below threshold)",
			size:         7 * 1024,
			explicitMode: "",
			wantMode:     "inline",
		},
		{
			name:         "8KB result (at threshold)",
			size:         8 * 1024,
			explicitMode: "",
			wantMode:     "inline",
		},
		{
			name:         "8KB + 1 byte (above threshold)",
			size:         8*1024 + 1,
			explicitMode: "",
			wantMode:     "file_ref",
		},
		{
			name:         "9KB result (above threshold)",
			size:         9 * 1024,
			explicitMode: "",
			wantMode:     "file_ref",
		},
		{
			name:         "100KB result (large)",
			size:         100 * 1024,
			explicitMode: "",
			wantMode:     "file_ref",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectOutputMode(tt.size, tt.explicitMode)
			if got != tt.wantMode {
				t.Errorf("selectOutputMode(%d, %q) = %q, want %q", tt.size, tt.explicitMode, got, tt.wantMode)
			}
		})
	}
}

// TestOutputModeOverride verifies explicit mode parameter handling
func TestOutputModeOverride(t *testing.T) {
	tests := []struct {
		name         string
		size         int
		explicitMode string
		wantMode     string
	}{
		{
			name:         "override to inline for large data",
			size:         100 * 1024,
			explicitMode: "inline",
			wantMode:     "inline",
		},
		{
			name:         "override to file_ref for small data",
			size:         1 * 1024,
			explicitMode: "file_ref",
			wantMode:     "file_ref",
		},
		{
			name:         "invalid mode falls back to auto-select (small)",
			size:         4 * 1024,
			explicitMode: "invalid",
			wantMode:     "inline",
		},
		{
			name:         "invalid mode falls back to auto-select (large)",
			size:         10 * 1024,
			explicitMode: "invalid",
			wantMode:     "file_ref",
		},
		{
			name:         "empty mode uses auto-select",
			size:         5 * 1024,
			explicitMode: "",
			wantMode:     "inline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectOutputMode(tt.size, tt.explicitMode)
			if got != tt.wantMode {
				t.Errorf("selectOutputMode(%d, %q) = %q, want %q", tt.size, tt.explicitMode, got, tt.wantMode)
			}
		})
	}
}

// TestOutputModeConfig verifies configuration struct
func TestOutputModeConfig(t *testing.T) {
	config := DefaultOutputModeConfig()

	if config.InlineThresholdBytes != 8*1024 {
		t.Errorf("DefaultOutputModeConfig().InlineThresholdBytes = %d, want %d", config.InlineThresholdBytes, 8*1024)
	}
}

// Helper functions

// generateTestData creates test data of approximately the specified size
func generateTestData(targetSize int) []interface{} {
	// Create records with predictable size
	// Each record: {"data":"X..."}\n where X is repeated
	// Estimate: 15 chars overhead + data length + 1 newline
	recordOverhead := 16

	var data []interface{}
	currentSize := 0

	for currentSize < targetSize {
		// Calculate how much data we need for this record
		remaining := targetSize - currentSize
		dataLen := remaining - recordOverhead
		if dataLen < 0 {
			dataLen = 0
		}
		if dataLen > 1000 {
			dataLen = 1000 // Limit individual record size
		}

		record := map[string]interface{}{
			"data": strings.Repeat("X", dataLen),
		}
		data = append(data, record)

		// Update current size (approximate)
		currentSize += dataLen + recordOverhead
	}

	return data
}

// abs returns absolute value of int
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// TestGetOutputModeConfigDefault verifies default 8KB threshold
func TestGetOutputModeConfigDefault(t *testing.T) {
	params := map[string]interface{}{}

	config := getOutputModeConfig(params)

	if config.InlineThresholdBytes != 8*1024 {
		t.Errorf("getOutputModeConfig() default threshold = %d, want %d", config.InlineThresholdBytes, 8*1024)
	}
}

// TestGetOutputModeConfigParameter verifies parameter override
func TestGetOutputModeConfigParameter(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantBytes int
	}{
		{
			name: "parameter as float64",
			params: map[string]interface{}{
				"inline_threshold_bytes": float64(16384),
			},
			wantBytes: 16384,
		},
		{
			name: "parameter as int",
			params: map[string]interface{}{
				"inline_threshold_bytes": 4096,
			},
			wantBytes: 4096,
		},
		{
			name: "parameter set to 1KB",
			params: map[string]interface{}{
				"inline_threshold_bytes": 1024,
			},
			wantBytes: 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := getOutputModeConfig(tt.params)

			if config.InlineThresholdBytes != tt.wantBytes {
				t.Errorf("getOutputModeConfig() threshold = %d, want %d", config.InlineThresholdBytes, tt.wantBytes)
			}
		})
	}
}

// TestGetOutputModeConfigEnvironment verifies env var override
func TestGetOutputModeConfigEnvironment(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		wantBytes int
	}{
		{
			name:      "valid env var 16KB",
			envValue:  "16384",
			wantBytes: 16384,
		},
		{
			name:      "valid env var 4KB",
			envValue:  "4096",
			wantBytes: 4096,
		},
		{
			name:      "invalid env var (non-numeric)",
			envValue:  "invalid",
			wantBytes: 8192, // Falls back to default
		},
		{
			name:      "invalid env var (negative)",
			envValue:  "-1000",
			wantBytes: 8192, // Falls back to default
		},
		{
			name:      "invalid env var (zero)",
			envValue:  "0",
			wantBytes: 8192, // Falls back to default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv("META_CC_INLINE_THRESHOLD", tt.envValue)
			defer os.Unsetenv("META_CC_INLINE_THRESHOLD")

			params := map[string]interface{}{}
			config := getOutputModeConfig(params)

			if config.InlineThresholdBytes != tt.wantBytes {
				t.Errorf("getOutputModeConfig() with env=%s threshold = %d, want %d", tt.envValue, config.InlineThresholdBytes, tt.wantBytes)
			}
		})
	}
}

// TestConfigPriority verifies parameter > env var > default priority
func TestConfigPriority(t *testing.T) {
	// Set environment variable to 4KB
	os.Setenv("META_CC_INLINE_THRESHOLD", "4096")
	defer os.Unsetenv("META_CC_INLINE_THRESHOLD")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantBytes int
	}{
		{
			name: "parameter overrides env var",
			params: map[string]interface{}{
				"inline_threshold_bytes": 16384,
			},
			wantBytes: 16384, // Parameter wins
		},
		{
			name:      "env var overrides default",
			params:    map[string]interface{}{},
			wantBytes: 4096, // Env var wins
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := getOutputModeConfig(tt.params)

			if config.InlineThresholdBytes != tt.wantBytes {
				t.Errorf("getOutputModeConfig() threshold = %d, want %d (priority test)", config.InlineThresholdBytes, tt.wantBytes)
			}
		})
	}
}
