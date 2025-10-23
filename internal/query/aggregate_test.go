package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaleh/meta-cc/internal/parser"
)

func TestApplyAggregate(t *testing.T) {
	entries := createTestEntries()
	tools := extractToolExecutions(entries)

	tests := []struct {
		name      string
		resources interface{}
		aggregate AggregateSpec
		wantCount int
	}{
		{
			name:      "empty_aggregate_no_change",
			resources: tools,
			aggregate: AggregateSpec{},
			wantCount: 1, // No aggregation, return as is
		},
		{
			name:      "count_all",
			resources: tools,
			aggregate: AggregateSpec{Function: "count"},
			wantCount: 1, // Single result with count
		},
		{
			name:      "count_by_tool_name",
			resources: tools,
			aggregate: AggregateSpec{Function: "count", Field: "tool_name"},
			wantCount: 1, // 1 unique tool name (Read)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyAggregate(tt.resources, tt.aggregate)
			require.NotNil(t, result)

			// Result should be []map[string]interface{} for aggregates
			if tt.aggregate.IsEmpty() {
				// No aggregation, return original type
				assert.IsType(t, tools, result)
			} else {
				// Aggregation returns map slice
				resultMaps, ok := result.([]map[string]interface{})
				require.True(t, ok, "Aggregated result should be []map[string]interface{}")
				assert.Equal(t, tt.wantCount, len(resultMaps))
			}
		})
	}
}

func TestAggregateCount(t *testing.T) {
	tools := []parser.ToolCall{
		{ToolName: "Read", Status: "success"},
		{ToolName: "Read", Status: "error"},
		{ToolName: "Edit", Status: "success"},
	}

	tests := []struct {
		name      string
		field     string
		wantCount int
		wantKeys  []string
	}{
		{
			name:      "count_all",
			field:     "",
			wantCount: 1,
			wantKeys:  []string{"count"},
		},
		{
			name:      "count_by_tool_name",
			field:     "tool_name",
			wantCount: 2, // Read: 2, Edit: 1
			wantKeys:  []string{"tool_name", "count"},
		},
		{
			name:      "count_by_status",
			field:     "status",
			wantCount: 2, // success: 2, error: 1
			wantKeys:  []string{"status", "count"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to []interface{}
			var items []interface{}
			for _, tool := range tools {
				items = append(items, tool)
			}

			result := aggregateCount(items, tt.field)

			require.Len(t, result, tt.wantCount)

			// Verify keys in result
			for _, item := range result {
				for _, key := range tt.wantKeys {
					assert.Contains(t, item, key)
				}
			}

			// Verify total count
			if tt.field == "" {
				assert.Equal(t, 3, result[0]["count"])
			}
		})
	}
}

func TestAggregateGroup(t *testing.T) {
	tools := []parser.ToolCall{
		{ToolName: "Read", Status: "success"},
		{ToolName: "Read", Status: "error"},
		{ToolName: "Edit", Status: "success"},
	}

	// Convert to []interface{}
	var items []interface{}
	for _, tool := range tools {
		items = append(items, tool)
	}

	result := aggregateGroup(items, "tool_name")

	require.Len(t, result, 2) // 2 groups: Read, Edit

	// Verify group structure
	for _, item := range result {
		assert.Contains(t, item, "tool_name")
		assert.Contains(t, item, "count")
	}
}

func TestExtractFieldValue(t *testing.T) {
	tests := []struct {
		name     string
		resource interface{}
		field    string
		want     string
	}{
		{
			name:     "tool_name_from_tool",
			resource: parser.ToolCall{ToolName: "Read"},
			field:    "tool_name",
			want:     "Read",
		},
		{
			name:     "status_from_tool",
			resource: parser.ToolCall{Status: "success"},
			field:    "status",
			want:     "success",
		},
		{
			name:     "role_from_message",
			resource: MessageView{Role: "user"},
			field:    "role",
			want:     "user",
		},
		{
			name:     "type_from_entry",
			resource: parser.SessionEntry{Type: "assistant"},
			field:    "type",
			want:     "assistant",
		},
		{
			name:     "unknown_field",
			resource: parser.ToolCall{ToolName: "Read"},
			field:    "unknown",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFieldValue(tt.resource, tt.field)
			assert.Equal(t, tt.want, got)
		})
	}
}
