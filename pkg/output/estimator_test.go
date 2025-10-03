package output

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/yale/meta-cc/internal/parser"
)

// generateTestToolCalls creates test ToolCall data
func generateTestToolCalls(count int) []parser.ToolCall {
	calls := make([]parser.ToolCall, count)

	for i := 0; i < count; i++ {
		calls[i] = parser.ToolCall{
			UUID:     "uuid-" + string(rune('A'+(i%26))),
			ToolName: "TestTool",
			Status:   "success",
			Input:    map[string]interface{}{"test": "value"},
			Output:   "Test output content",
		}
	}

	return calls
}

func TestEstimateToolCallsSize(t *testing.T) {
	tools := generateTestToolCalls(100)

	tests := []struct {
		name          string
		format        string
		expectedMinKB float64
		expectedMaxKB float64
	}{
		{
			name:          "JSON format",
			format:        "json",
			expectedMinKB: 5.0,
			expectedMaxKB: 50.0,
		},
		{
			name:          "Markdown format",
			format:        "md",
			expectedMinKB: 10.0,
			expectedMaxKB: 50.0,
		},
		{
			name:          "CSV format",
			format:        "csv",
			expectedMinKB: 5.0,
			expectedMaxKB: 30.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			estimate, err := EstimateToolCallsSize(tools, tt.format)
			if err != nil {
				t.Fatalf("EstimateToolCallsSize failed: %v", err)
			}

			if estimate.EstimatedKB < tt.expectedMinKB || estimate.EstimatedKB > tt.expectedMaxKB {
				t.Errorf("expected size between %.1f-%.1f KB, got %.1f KB",
					tt.expectedMinKB, tt.expectedMaxKB, estimate.EstimatedKB)
			}

			if estimate.RecordCount != 100 {
				t.Errorf("expected record count 100, got %d", estimate.RecordCount)
			}

			if estimate.Format != tt.format {
				t.Errorf("expected format %s, got %s", tt.format, estimate.Format)
			}
		})
	}
}

func TestEstimateSizeAccuracy(t *testing.T) {
	// Test accuracy requirement: ≥95%
	tools := generateTestToolCalls(100)

	estimate, err := EstimateToolCallsSize(tools, "json")
	if err != nil {
		t.Fatalf("EstimateToolCallsSize failed: %v", err)
	}

	// Get actual size
	actual, err := json.Marshal(tools)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	actualSize := len(actual)

	// Calculate accuracy
	diff := int(math.Abs(float64(estimate.EstimatedBytes - actualSize)))
	accuracy := 1.0 - float64(diff)/float64(actualSize)

	t.Logf("Estimated: %d bytes, Actual: %d bytes, Accuracy: %.2f%%",
		estimate.EstimatedBytes, actualSize, accuracy*100)

	if accuracy < 0.95 {
		t.Errorf("estimate accuracy %.2f%% is below 95%% threshold", accuracy*100)
	}
}

func TestEstimateStatsSize(t *testing.T) {
	formats := []string{"json", "md", "csv"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			estimate := EstimateStatsSize(format)

			// Stats should be small (< 5 KB)
			if estimate.EstimatedKB > 5.0 {
				t.Errorf("stats size too large: %.1f KB", estimate.EstimatedKB)
			}

			if estimate.RecordCount != 1 {
				t.Errorf("expected record count 1, got %d", estimate.RecordCount)
			}

			if estimate.Format != format {
				t.Errorf("expected format %s, got %s", format, estimate.Format)
			}
		})
	}
}

func TestEstimateSizeEmptyData(t *testing.T) {
	tools := []parser.ToolCall{}

	estimate, err := EstimateToolCallsSize(tools, "json")
	if err != nil {
		t.Fatalf("EstimateToolCallsSize failed: %v", err)
	}

	// Empty array should be "[]" = 2 bytes
	if estimate.EstimatedBytes < 2 || estimate.EstimatedBytes > 10 {
		t.Errorf("expected ~2 bytes for empty array, got %d", estimate.EstimatedBytes)
	}

	if estimate.RecordCount != 0 {
		t.Errorf("expected record count 0, got %d", estimate.RecordCount)
	}
}

func TestEstimateSizeUnsupportedFormat(t *testing.T) {
	tools := generateTestToolCalls(10)

	// Unsupported format should still return estimate (fallback)
	estimate, err := EstimateToolCallsSize(tools, "xml")
	if err != nil {
		t.Fatalf("EstimateToolCallsSize failed: %v", err)
	}

	// Should have some estimate
	if estimate.EstimatedBytes <= 0 {
		t.Error("expected positive estimate for unsupported format")
	}
}
