package assistant

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/types"
)

func assistantFixture() []types.SessionEntry {
	return []types.SessionEntry{
		{Type: "user", UUID: "u1", Timestamp: "2026-01-01T10:00:00Z", Message: &types.Message{Role: "user", Content: []types.ContentBlock{{Type: "text", Text: "alpha request"}}}},
		{Type: "assistant", UUID: "a1", ParentUUID: "u1", Timestamp: "2026-01-01T10:00:02Z", Message: &types.Message{Role: "assistant", Model: "model-a", StopReason: "end_turn", Usage: map[string]interface{}{"input_tokens": float64(10), "output_tokens": float64(20)}, Content: []types.ContentBlock{{Type: "text", Text: "short"}, {Type: "tool_use", ToolUse: &types.ToolUse{Name: "Read"}}}}},
		{Type: "user", UUID: "u2", ParentUUID: "a1", Timestamp: "2026-01-01T10:01:00Z", Message: &types.Message{Role: "user", Content: []types.ContentBlock{{Type: "text", Text: "beta request"}}}},
		{Type: "assistant", UUID: "a2", ParentUUID: "u2", Timestamp: "2026-01-01T10:01:05Z", Message: &types.Message{Role: "assistant", Model: "model-b", Usage: map[string]interface{}{"input_tokens": "bad", "output_tokens": float64(50)}, Content: []types.ContentBlock{{Type: "text", Text: "a much longer answer"}, {Type: "tool_use"}, {Type: "tool_use", ToolUse: &types.ToolUse{Name: "Bash"}}}}},
		{Type: "user", UUID: "system", ParentUUID: "a2", Timestamp: "2026-01-01T10:02:00Z", Message: &types.Message{Role: "user", Content: []types.ContentBlock{{Type: "text", Text: "<command-message>hidden"}}}},
		{Type: "assistant", UUID: "nil-message"},
		{Type: "progress", UUID: "ignored", Message: &types.Message{}},
	}
}

func defaultMessageOptions() AssistantMessagesOptions {
	return AssistantMessagesOptions{MinTools: -1, MaxTools: -1, MinTokens: -1, MinLength: -1, MaxLength: -1}
}

func defaultConversationOptions() ConversationOptions {
	return ConversationOptions{StartTurn: -1, EndTurn: -1, MinDuration: -1, MaxDuration: -1}
}

func TestBuildAssistantMessagesExtractsMetadata(t *testing.T) {
	got, err := BuildAssistantMessages(assistantFixture(), defaultMessageOptions())
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "a1", got[0].UUID)
	assert.Equal(t, 5, got[0].TextLength)
	assert.Equal(t, 1, got[0].ToolUseCount)
	assert.Equal(t, 10, got[0].TokensInput)
	assert.Equal(t, 20, got[0].TokensOutput)
	assert.Equal(t, "end_turn", got[0].StopReason)
	assert.Equal(t, "Read", got[0].ContentBlocks[1].ToolName)
	assert.Equal(t, 0, got[1].TokensInput)
	assert.Equal(t, "", got[1].ContentBlocks[1].ToolName)
}

func TestBuildAssistantMessagesFilters(t *testing.T) {
	tests := []struct {
		name string
		edit func(*AssistantMessagesOptions)
		want []string
	}{
		{"pattern", func(o *AssistantMessagesOptions) { o.Pattern = "longer" }, []string{"a2"}},
		{"pattern miss", func(o *AssistantMessagesOptions) { o.Pattern = "missing" }, nil},
		{"minimum tools", func(o *AssistantMessagesOptions) { o.MinTools = 2 }, []string{"a2"}},
		{"maximum tools", func(o *AssistantMessagesOptions) { o.MaxTools = 1 }, []string{"a1"}},
		{"token minimum", func(o *AssistantMessagesOptions) { o.MinTokens = 30 }, []string{"a2"}},
		{"length minimum", func(o *AssistantMessagesOptions) { o.MinLength = 10 }, []string{"a2"}},
		{"length maximum", func(o *AssistantMessagesOptions) { o.MaxLength = 10 }, []string{"a1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultMessageOptions()
			tt.edit(&opts)
			got, err := BuildAssistantMessages(assistantFixture(), opts)
			require.NoError(t, err)
			assert.Equal(t, tt.want, messageUUIDs(got))
		})
	}
}

func messageUUIDs(messages []AssistantMessage) []string {
	var ids []string
	for _, message := range messages {
		ids = append(ids, message.UUID)
	}
	return ids
}

func TestBuildAssistantMessagesSortAndPage(t *testing.T) {
	tests := []struct {
		name    string
		sortBy  string
		reverse bool
		want    []string
	}{
		{"timestamp", "timestamp", false, []string{"a1", "a2"}},
		{"tools reverse", "tool_use_count", true, []string{"a2", "a1"}},
		{"length reverse", "text_length", true, []string{"a2", "a1"}},
		{"sequence reverse", "", true, []string{"a2", "a1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultMessageOptions()
			opts.SortBy, opts.Reverse = tt.sortBy, tt.reverse
			got, err := BuildAssistantMessages(assistantFixture(), opts)
			require.NoError(t, err)
			assert.Equal(t, tt.want, messageUUIDs(got))
		})
	}

	opts := defaultMessageOptions()
	opts.Offset, opts.Limit = 1, 1
	got, err := BuildAssistantMessages(assistantFixture(), opts)
	require.NoError(t, err)
	assert.Equal(t, []string{"a2"}, messageUUIDs(got))
	opts.Offset = 10
	got, err = BuildAssistantMessages(assistantFixture(), opts)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestBuildConversationTurnsFiltersAndDuration(t *testing.T) {
	got, err := BuildConversationTurns(assistantFixture(), defaultConversationOptions())
	require.NoError(t, err)
	require.Len(t, got, 4)
	assert.Zero(t, got[0].Duration)

	tests := []struct {
		name string
		edit func(*ConversationOptions)
		want int
	}{
		{"range", func(o *ConversationOptions) { o.StartTurn, o.EndTurn = 2, 2 }, 1},
		{"user pattern", func(o *ConversationOptions) { o.Pattern, o.PatternTarget = "alpha", "user" }, 1},
		{"assistant pattern", func(o *ConversationOptions) { o.Pattern, o.PatternTarget = "longer", "assistant" }, 1},
		{"default pattern target", func(o *ConversationOptions) { o.Pattern = "request" }, 2},
		{"minimum duration", func(o *ConversationOptions) { o.MinDuration = 1 }, 0},
		{"maximum duration", func(o *ConversationOptions) { o.MaxDuration = 0 }, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultConversationOptions()
			tt.edit(&opts)
			turns, err := BuildConversationTurns(assistantFixture(), opts)
			require.NoError(t, err)
			assert.Len(t, turns, tt.want)
		})
	}
}

func TestConversationHelpersHandlePartialAndInvalidTurns(t *testing.T) {
	assert.Zero(t, calculateTurnDuration(nil, &AssistantMessage{}))
	assert.Zero(t, calculateTurnDuration(&types.UserMessage{Timestamp: "bad"}, &AssistantMessage{Timestamp: "also-bad"}))
	assert.Equal(t, "user", firstTimestamp(&types.UserMessage{Timestamp: "user"}, &AssistantMessage{Timestamp: "assistant"}, "fallback"))
	assert.Equal(t, "assistant", firstTimestamp(nil, &AssistantMessage{Timestamp: "assistant"}, "fallback"))
	assert.Equal(t, "fallback", firstTimestamp(nil, nil, "fallback"))

	for _, prefix := range []string{"<command-message>", "<command-name>", "<command-args>", "<local-command", "Caveat:", "# meta-"} {
		assert.True(t, isSystemAssistantMessage("  "+prefix+" hidden"))
	}
	assert.False(t, isSystemAssistantMessage(""))
	assert.False(t, isSystemAssistantMessage("normal request"))
}
