package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaleh/meta-cc/internal/parser"
)

// Test QueryParams validation
func TestQueryParamsValidation(t *testing.T) {
	tests := []struct {
		name    string
		params  QueryParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid_basic_query",
			params: QueryParams{
				Resource: "tools",
				Scope:    "project",
				Filter: FilterSpec{
					ToolName: "Read",
				},
			},
			wantErr: false,
		},
		{
			name: "valid_with_aggregate",
			params: QueryParams{
				Resource: "tools",
				Scope:    "session",
				Aggregate: AggregateSpec{
					Function: "count",
					Field:    "tool_name",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid_resource",
			params: QueryParams{
				Resource: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid resource",
		},
		{
			name: "invalid_scope",
			params: QueryParams{
				Resource: "entries",
				Scope:    "invalid",
			},
			wantErr: true,
			errMsg:  "invalid scope",
		},
		{
			name: "invalid_aggregate_function",
			params: QueryParams{
				Resource: "tools",
				Aggregate: AggregateSpec{
					Function: "invalid",
				},
			},
			wantErr: true,
			errMsg:  "invalid aggregate.function",
		},
		{
			name:   "default_values",
			params: QueryParams{
				// No resource specified
			},
			wantErr: false, // Should use defaults
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQueryParams(tt.params)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test QueryParams defaults
func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name     string
		params   QueryParams
		expected QueryParams
	}{
		{
			name:   "empty_params",
			params: QueryParams{},
			expected: QueryParams{
				Resource: "entries",
				Scope:    "project",
				Output: OutputSpec{
					Format: "jsonl",
				},
			},
		},
		{
			name: "partial_params",
			params: QueryParams{
				Resource: "tools",
			},
			expected: QueryParams{
				Resource: "tools",
				Scope:    "project",
				Output: OutputSpec{
					Format: "jsonl",
				},
			},
		},
		{
			name: "full_params_no_change",
			params: QueryParams{
				Resource: "messages",
				Scope:    "session",
				Output: OutputSpec{
					Format: "tsv",
					Limit:  10,
				},
			},
			expected: QueryParams{
				Resource: "messages",
				Scope:    "session",
				Output: OutputSpec{
					Format: "tsv",
					Limit:  10,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyDefaults(tt.params)
			assert.Equal(t, tt.expected.Resource, result.Resource)
			assert.Equal(t, tt.expected.Scope, result.Scope)
			assert.Equal(t, tt.expected.Output.Format, result.Output.Format)
			if tt.expected.Output.Limit > 0 {
				assert.Equal(t, tt.expected.Output.Limit, result.Output.Limit)
			}
		})
	}
}

// Test FilterSpec isEmpty
func TestFilterSpecIsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		filter FilterSpec
		want   bool
	}{
		{
			name:   "empty_filter",
			filter: FilterSpec{},
			want:   true,
		},
		{
			name: "has_type",
			filter: FilterSpec{
				Type: "assistant",
			},
			want: false,
		},
		{
			name: "has_tool_name",
			filter: FilterSpec{
				ToolName: "Read",
			},
			want: false,
		},
		{
			name: "has_role",
			filter: FilterSpec{
				Role: "user",
			},
			want: false,
		},
		{
			name: "has_time_range",
			filter: FilterSpec{
				TimeRange: &TimeRange{
					Start: "2025-10-23T00:00:00Z",
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.IsEmpty()
			assert.Equal(t, tt.want, got)
		})
	}
}

// Test AggregateSpec isEmpty
func TestAggregateSpecIsEmpty(t *testing.T) {
	tests := []struct {
		name      string
		aggregate AggregateSpec
		want      bool
	}{
		{
			name:      "empty_aggregate",
			aggregate: AggregateSpec{},
			want:      true,
		},
		{
			name: "has_function",
			aggregate: AggregateSpec{
				Function: "count",
			},
			want: false,
		},
		{
			name: "has_function_and_field",
			aggregate: AggregateSpec{
				Function: "count",
				Field:    "tool_name",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.aggregate.IsEmpty()
			assert.Equal(t, tt.want, got)
		})
	}
}

// Integration tests for Query function

func TestQueryIntegration(t *testing.T) {
	entries := createTestEntries()

	tests := []struct {
		name    string
		params  QueryParams
		wantErr bool
		verify  func(t *testing.T, result interface{})
	}{
		{
			name: "query_all_entries",
			params: QueryParams{
				Resource: "entries",
			},
			wantErr: false,
			verify: func(t *testing.T, result interface{}) {
				entries, ok := result.([]parser.SessionEntry)
				require.True(t, ok)
				assert.Len(t, entries, 3)
			},
		},
		{
			name: "query_user_messages",
			params: QueryParams{
				Resource: "messages",
				Filter: FilterSpec{
					Role: "user",
				},
			},
			wantErr: false,
			verify: func(t *testing.T, result interface{}) {
				messages, ok := result.([]MessageView)
				require.True(t, ok)
				assert.GreaterOrEqual(t, len(messages), 1)
				for _, msg := range messages {
					assert.Equal(t, "user", msg.Role)
				}
			},
		},
		{
			name: "query_tools_with_filter",
			params: QueryParams{
				Resource: "tools",
				Filter: FilterSpec{
					ToolName: "Read",
				},
			},
			wantErr: false,
			verify: func(t *testing.T, result interface{}) {
				tools, ok := result.([]parser.ToolCall)
				require.True(t, ok)
				assert.Len(t, tools, 1)
				assert.Equal(t, "Read", tools[0].ToolName)
			},
		},
		{
			name: "query_tools_with_count",
			params: QueryParams{
				Resource: "tools",
				Aggregate: AggregateSpec{
					Function: "count",
				},
			},
			wantErr: false,
			verify: func(t *testing.T, result interface{}) {
				results, ok := result.([]map[string]interface{})
				require.True(t, ok)
				assert.Len(t, results, 1)
				assert.Contains(t, results[0], "count")
			},
		},
		{
			name: "query_tools_count_by_name",
			params: QueryParams{
				Resource: "tools",
				Aggregate: AggregateSpec{
					Function: "count",
					Field:    "tool_name",
				},
			},
			wantErr: false,
			verify: func(t *testing.T, result interface{}) {
				results, ok := result.([]map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, results)
				for _, r := range results {
					assert.Contains(t, r, "tool_name")
					assert.Contains(t, r, "count")
				}
			},
		},
		{
			name: "query_invalid_resource",
			params: QueryParams{
				Resource: "invalid",
			},
			wantErr: true,
		},
		{
			name: "query_entries_by_type",
			params: QueryParams{
				Resource: "entries",
				Filter: FilterSpec{
					Type: "user",
				},
			},
			wantErr: false,
			verify: func(t *testing.T, result interface{}) {
				entries, ok := result.([]parser.SessionEntry)
				require.True(t, ok)
				assert.GreaterOrEqual(t, len(entries), 1)
				for _, e := range entries {
					assert.Equal(t, "user", e.Type)
				}
			},
		},
		{
			name: "query_messages_by_session",
			params: QueryParams{
				Resource: "messages",
				Filter: FilterSpec{
					SessionID: "session-1",
				},
			},
			wantErr: false,
			verify: func(t *testing.T, result interface{}) {
				messages, ok := result.([]MessageView)
				require.True(t, ok)
				assert.NotEmpty(t, messages)
				for _, msg := range messages {
					assert.Equal(t, "session-1", msg.SessionID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Query(entries, tt.params)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			if tt.verify != nil {
				tt.verify(t, result)
			}
		})
	}
}

func TestQueryEmptyEntries(t *testing.T) {
	result, err := Query([]parser.SessionEntry{}, QueryParams{
		Resource: "entries",
	})

	require.NoError(t, err)
	entries, ok := result.([]parser.SessionEntry)
	require.True(t, ok)
	assert.Empty(t, entries)
}

func TestQueryValidationError(t *testing.T) {
	entries := createTestEntries()

	_, err := Query(entries, QueryParams{
		Resource: "invalid_resource",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}
