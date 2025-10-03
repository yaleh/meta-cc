package filter

import (
	"testing"

	"github.com/yale/meta-cc/internal/parser"
)

// generateTestToolCalls creates test ToolCall data
func generateTestToolCalls(count int) []parser.ToolCall {
	calls := make([]parser.ToolCall, count)

	for i := 0; i < count; i++ {
		calls[i] = parser.ToolCall{
			UUID:     string(rune('A' + (i % 26))),
			ToolName: "TestTool",
			Status:   "success",
		}
	}

	return calls
}

func TestApplyPagination(t *testing.T) {
	tools := generateTestToolCalls(100)

	tests := []struct {
		name     string
		config   PaginationConfig
		expected int
	}{
		{
			name:     "no pagination",
			config:   PaginationConfig{Limit: 0, Offset: 0},
			expected: 100,
		},
		{
			name:     "limit 50",
			config:   PaginationConfig{Limit: 50, Offset: 0},
			expected: 50,
		},
		{
			name:     "offset 50, limit 30",
			config:   PaginationConfig{Limit: 30, Offset: 50},
			expected: 30,
		},
		{
			name:     "offset beyond end",
			config:   PaginationConfig{Limit: 10, Offset: 120},
			expected: 0,
		},
		{
			name:     "limit exceeds remaining",
			config:   PaginationConfig{Limit: 100, Offset: 90},
			expected: 10,
		},
		{
			name:     "negative limit (treat as no limit)",
			config:   PaginationConfig{Limit: -1, Offset: 0},
			expected: 100,
		},
		{
			name:     "negative offset (treat as 0)",
			config:   PaginationConfig{Limit: 10, Offset: -5},
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyPagination(tools, tt.config)
			if len(result) != tt.expected {
				t.Errorf("expected %d records, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestCalculateMetadata(t *testing.T) {
	tests := []struct {
		name         string
		totalRecords int
		config       PaginationConfig
		expected     PaginationMetadata
	}{
		{
			name:         "first page",
			totalRecords: 100,
			config:       PaginationConfig{Limit: 50, Offset: 0},
			expected: PaginationMetadata{
				TotalRecords:    100,
				ReturnedRecords: 50,
				Offset:          0,
				Limit:           50,
				HasMore:         true,
			},
		},
		{
			name:         "last page (partial)",
			totalRecords: 100,
			config:       PaginationConfig{Limit: 50, Offset: 90},
			expected: PaginationMetadata{
				TotalRecords:    100,
				ReturnedRecords: 10,
				Offset:          90,
				Limit:           50,
				HasMore:         false,
			},
		},
		{
			name:         "no limit",
			totalRecords: 100,
			config:       PaginationConfig{Limit: 0, Offset: 0},
			expected: PaginationMetadata{
				TotalRecords:    100,
				ReturnedRecords: 100,
				Offset:          0,
				Limit:           0,
				HasMore:         false,
			},
		},
		{
			name:         "offset beyond end",
			totalRecords: 100,
			config:       PaginationConfig{Limit: 10, Offset: 150},
			expected: PaginationMetadata{
				TotalRecords:    100,
				ReturnedRecords: 0,
				Offset:          150,
				Limit:           10,
				HasMore:         false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := CalculateMetadata(tt.totalRecords, tt.config)

			if metadata.TotalRecords != tt.expected.TotalRecords {
				t.Errorf("TotalRecords: expected %d, got %d", tt.expected.TotalRecords, metadata.TotalRecords)
			}
			if metadata.ReturnedRecords != tt.expected.ReturnedRecords {
				t.Errorf("ReturnedRecords: expected %d, got %d", tt.expected.ReturnedRecords, metadata.ReturnedRecords)
			}
			if metadata.Offset != tt.expected.Offset {
				t.Errorf("Offset: expected %d, got %d", tt.expected.Offset, metadata.Offset)
			}
			if metadata.Limit != tt.expected.Limit {
				t.Errorf("Limit: expected %d, got %d", tt.expected.Limit, metadata.Limit)
			}
			if metadata.HasMore != tt.expected.HasMore {
				t.Errorf("HasMore: expected %v, got %v", tt.expected.HasMore, metadata.HasMore)
			}
		})
	}
}

func TestPaginationEdgeCases(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		tools := []parser.ToolCall{}
		config := PaginationConfig{Limit: 10, Offset: 0}
		result := ApplyPagination(tools, config)

		if len(result) != 0 {
			t.Errorf("expected empty result, got %d records", len(result))
		}
	})

	t.Run("exact page boundary", func(t *testing.T) {
		tools := generateTestToolCalls(100)
		config := PaginationConfig{Limit: 50, Offset: 50}
		result := ApplyPagination(tools, config)

		if len(result) != 50 {
			t.Errorf("expected 50 records, got %d", len(result))
		}
	})
}
