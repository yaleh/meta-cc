package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaleh/meta-cc/internal/parser"
)

func TestApplyFilter(t *testing.T) {
	entries := createTestEntries()

	tests := []struct {
		name      string
		resources interface{}
		filter    FilterSpec
		wantCount int
	}{
		{
			name:      "empty_filter_no_change",
			resources: entries,
			filter:    FilterSpec{},
			wantCount: 3, // All entries
		},
		{
			name:      "filter_by_type_user",
			resources: entries,
			filter: FilterSpec{
				Type: "user",
			},
			wantCount: 2, // 2 user entries
		},
		{
			name:      "filter_by_type_assistant",
			resources: entries,
			filter: FilterSpec{
				Type: "assistant",
			},
			wantCount: 1, // 1 assistant entry
		},
		{
			name:      "filter_by_session_id",
			resources: entries,
			filter: FilterSpec{
				SessionID: "session-1",
			},
			wantCount: 3, // All belong to session-1
		},
		{
			name:      "filter_by_git_branch",
			resources: entries,
			filter: FilterSpec{
				GitBranch: "main",
			},
			wantCount: 3, // All on main branch
		},
		{
			name:      "filter_by_uuid",
			resources: entries,
			filter: FilterSpec{
				UUID: "user-1",
			},
			wantCount: 1, // Only one entry with UUID "user-1"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyFilter(tt.resources, tt.filter)

			// Type assertion for entries
			if resultEntries, ok := result.([]parser.SessionEntry); ok {
				assert.Equal(t, tt.wantCount, len(resultEntries))
			} else {
				t.Fatalf("Result is not []parser.SessionEntry")
			}
		})
	}
}

func TestApplyFilterMessages(t *testing.T) {
	entries := createTestEntries()
	messages := extractMessages(entries)

	tests := []struct {
		name      string
		filter    FilterSpec
		wantCount int
	}{
		{
			name:      "filter_by_role_user",
			filter:    FilterSpec{Role: "user"},
			wantCount: 2, // 2 user messages
		},
		{
			name:      "filter_by_role_assistant",
			filter:    FilterSpec{Role: "assistant"},
			wantCount: 1, // 1 assistant message
		},
		{
			name:      "filter_by_session_id",
			filter:    FilterSpec{SessionID: "session-1"},
			wantCount: 3, // All messages
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyFilter(messages, tt.filter)

			// Type assertion for messages
			if resultMessages, ok := result.([]MessageView); ok {
				assert.Equal(t, tt.wantCount, len(resultMessages))
			} else {
				t.Fatalf("Result is not []MessageView")
			}
		})
	}
}

func TestApplyFilterTools(t *testing.T) {
	entries := createTestEntries()
	tools := extractToolExecutions(entries)

	tests := []struct {
		name      string
		filter    FilterSpec
		wantCount int
	}{
		{
			name:      "filter_by_tool_name",
			filter:    FilterSpec{ToolName: "Read"},
			wantCount: 1, // 1 Read tool
		},
		{
			name:      "filter_by_status_success",
			filter:    FilterSpec{ToolStatus: "success"},
			wantCount: 1, // 1 successful tool
		},
		{
			name: "filter_by_has_error_false",
			filter: FilterSpec{
				HasError: boolPtr(false),
			},
			wantCount: 1, // 1 tool without error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyFilter(tools, tt.filter)

			// Type assertion for tools
			if resultTools, ok := result.([]parser.ToolCall); ok {
				assert.Equal(t, tt.wantCount, len(resultTools))
			} else {
				t.Fatalf("Result is not []parser.ToolCall")
			}
		})
	}
}

func TestMatchesFilterEntry(t *testing.T) {
	entry := parser.SessionEntry{
		Type:       "user",
		UUID:       "test-uuid",
		SessionID:  "session-1",
		GitBranch:  "main",
		Timestamp:  "2025-10-23T00:00:00Z",
		ParentUUID: "parent-1",
	}

	tests := []struct {
		name   string
		filter FilterSpec
		want   bool
	}{
		{
			name:   "empty_filter_matches",
			filter: FilterSpec{},
			want:   true,
		},
		{
			name:   "type_matches",
			filter: FilterSpec{Type: "user"},
			want:   true,
		},
		{
			name:   "type_not_matches",
			filter: FilterSpec{Type: "assistant"},
			want:   false,
		},
		{
			name:   "uuid_matches",
			filter: FilterSpec{UUID: "test-uuid"},
			want:   true,
		},
		{
			name:   "uuid_not_matches",
			filter: FilterSpec{UUID: "other-uuid"},
			want:   false,
		},
		{
			name:   "session_id_matches",
			filter: FilterSpec{SessionID: "session-1"},
			want:   true,
		},
		{
			name:   "git_branch_matches",
			filter: FilterSpec{GitBranch: "main"},
			want:   true,
		},
		{
			name:   "parent_uuid_matches",
			filter: FilterSpec{ParentUUID: "parent-1"},
			want:   true,
		},
		{
			name: "multiple_conditions_all_match",
			filter: FilterSpec{
				Type:      "user",
				SessionID: "session-1",
				GitBranch: "main",
			},
			want: true,
		},
		{
			name: "multiple_conditions_one_fails",
			filter: FilterSpec{
				Type:      "user",
				SessionID: "session-2", // Wrong session
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesFilter(entry, tt.filter)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestMatchesFilterMessage(t *testing.T) {
	msg := MessageView{
		UUID:       "msg-1",
		SessionID:  "session-1",
		Role:       "user",
		Content:    "Read the file please",
		GitBranch:  "main",
		ParentUUID: "parent-1",
	}

	tests := []struct {
		name   string
		filter FilterSpec
		want   bool
	}{
		{
			name:   "role_matches",
			filter: FilterSpec{Role: "user"},
			want:   true,
		},
		{
			name:   "role_not_matches",
			filter: FilterSpec{Role: "assistant"},
			want:   false,
		},
		{
			name:   "session_id_matches",
			filter: FilterSpec{SessionID: "session-1"},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesFilter(msg, tt.filter)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestMatchesFilterTool(t *testing.T) {
	tool := parser.ToolCall{
		UUID:     "tool-uuid",
		ToolName: "Read",
		Status:   "success",
		Error:    "",
	}

	tests := []struct {
		name   string
		filter FilterSpec
		want   bool
	}{
		{
			name:   "tool_name_matches",
			filter: FilterSpec{ToolName: "Read"},
			want:   true,
		},
		{
			name:   "tool_name_not_matches",
			filter: FilterSpec{ToolName: "Edit"},
			want:   false,
		},
		{
			name:   "status_matches",
			filter: FilterSpec{ToolStatus: "success"},
			want:   true,
		},
		{
			name:   "status_not_matches",
			filter: FilterSpec{ToolStatus: "error"},
			want:   false,
		},
		{
			name: "has_error_false_matches",
			filter: FilterSpec{
				HasError: boolPtr(false),
			},
			want: true,
		},
		{
			name: "has_error_true_not_matches",
			filter: FilterSpec{
				HasError: boolPtr(true),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesFilter(tool, tt.filter)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestApplyFilterTimeRange(t *testing.T) {
	entries := []parser.SessionEntry{
		{
			Type:      "user",
			UUID:      "early",
			Timestamp: "2025-10-20T00:00:00Z",
		},
		{
			Type:      "user",
			UUID:      "middle",
			Timestamp: "2025-10-23T00:00:00Z",
		},
		{
			Type:      "user",
			UUID:      "late",
			Timestamp: "2025-10-25T00:00:00Z",
		},
	}

	tests := []struct {
		name      string
		filter    FilterSpec
		wantCount int
		wantUUIDs []string
	}{
		{
			name: "filter_after_start",
			filter: FilterSpec{
				TimeRange: &TimeRange{
					Start: "2025-10-22T00:00:00Z",
				},
			},
			wantCount: 2,
			wantUUIDs: []string{"middle", "late"},
		},
		{
			name: "filter_before_end",
			filter: FilterSpec{
				TimeRange: &TimeRange{
					End: "2025-10-24T00:00:00Z",
				},
			},
			wantCount: 2,
			wantUUIDs: []string{"early", "middle"},
		},
		{
			name: "filter_range",
			filter: FilterSpec{
				TimeRange: &TimeRange{
					Start: "2025-10-22T00:00:00Z",
					End:   "2025-10-24T00:00:00Z",
				},
			},
			wantCount: 1,
			wantUUIDs: []string{"middle"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyFilter(entries, tt.filter)
			resultEntries, ok := result.([]parser.SessionEntry)
			require.True(t, ok)
			assert.Equal(t, tt.wantCount, len(resultEntries))

			// Verify UUIDs
			var gotUUIDs []string
			for _, e := range resultEntries {
				gotUUIDs = append(gotUUIDs, e.UUID)
			}
			assert.ElementsMatch(t, tt.wantUUIDs, gotUUIDs)
		})
	}
}

// Helper function
func boolPtr(b bool) *bool {
	return &b
}
