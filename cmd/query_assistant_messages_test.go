package cmd

import (
	"regexp"
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
)

// TestExtractAssistantMessages tests basic extraction of assistant messages
func TestExtractAssistantMessages(t *testing.T) {
	entries := []parser.SessionEntry{
		{
			Type:      "user",
			UUID:      "user-1",
			Timestamp: "2025-10-09T10:00:00Z",
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{Type: "text", Text: "Hello"},
				},
			},
		},
		{
			Type:      "assistant",
			UUID:      "asst-1",
			Timestamp: "2025-10-09T10:00:01Z",
			Message: &parser.Message{
				Role:  "assistant",
				Model: "claude-sonnet-4-5",
				Content: []parser.ContentBlock{
					{Type: "text", Text: "Hi there!"},
					{Type: "tool_use", ToolUse: &parser.ToolUse{ID: "t1", Name: "Read", Input: map[string]interface{}{}}},
				},
				Usage: map[string]interface{}{
					"input_tokens":  100.0,
					"output_tokens": 50.0,
				},
			},
		},
	}

	turnIndex := map[string]int{
		"user-1": 1,
		"asst-1": 2,
	}

	messages := extractAssistantMessages(entries, turnIndex)

	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.TurnSequence != 2 {
		t.Errorf("expected turn_sequence 2, got %d", msg.TurnSequence)
	}
	if msg.UUID != "asst-1" {
		t.Errorf("expected uuid asst-1, got %s", msg.UUID)
	}
	if msg.Model != "claude-sonnet-4-5" {
		t.Errorf("expected model claude-sonnet-4-5, got %s", msg.Model)
	}
	if msg.TextLength != 9 {
		t.Errorf("expected text_length 9, got %d", msg.TextLength)
	}
	if msg.ToolUseCount != 1 {
		t.Errorf("expected tool_use_count 1, got %d", msg.ToolUseCount)
	}
	if msg.TokensInput != 100 {
		t.Errorf("expected tokens_input 100, got %d", msg.TokensInput)
	}
	if msg.TokensOutput != 50 {
		t.Errorf("expected tokens_output 50, got %d", msg.TokensOutput)
	}
}

// TestAssistantMessagesPatternMatching tests regex filtering
func TestAssistantMessagesPatternMatching(t *testing.T) {
	messages := []AssistantMessage{
		{TurnSequence: 1, ContentBlocks: []ContentBlock{{Type: "text", Text: "This is an error message"}}},
		{TurnSequence: 2, ContentBlocks: []ContentBlock{{Type: "text", Text: "This is a success message"}}},
		{TurnSequence: 3, ContentBlocks: []ContentBlock{{Type: "text", Text: "Another error occurred"}}},
	}

	pattern := regexp.MustCompile("error")
	filtered := filterAssistantMessagesByPattern(messages, pattern)

	if len(filtered) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(filtered))
	}
	if filtered[0].TurnSequence != 1 {
		t.Errorf("expected turn 1, got %d", filtered[0].TurnSequence)
	}
	if filtered[1].TurnSequence != 3 {
		t.Errorf("expected turn 3, got %d", filtered[1].TurnSequence)
	}
}

// TestAssistantMessagesToolFiltering tests tool count filtering
func TestAssistantMessagesToolFiltering(t *testing.T) {
	messages := []AssistantMessage{
		{TurnSequence: 1, ToolUseCount: 0},
		{TurnSequence: 2, ToolUseCount: 3},
		{TurnSequence: 3, ToolUseCount: 5},
		{TurnSequence: 4, ToolUseCount: 7},
	}

	// Min tools filter
	filtered := filterAssistantMessagesByToolCount(messages, 3, -1)
	if len(filtered) != 3 {
		t.Fatalf("expected 3 messages with min_tools=3, got %d", len(filtered))
	}

	// Max tools filter
	filtered = filterAssistantMessagesByToolCount(messages, -1, 5)
	if len(filtered) != 3 {
		t.Fatalf("expected 3 messages with max_tools=5, got %d", len(filtered))
	}

	// Both filters
	filtered = filterAssistantMessagesByToolCount(messages, 3, 5)
	if len(filtered) != 2 {
		t.Fatalf("expected 2 messages with min=3 max=5, got %d", len(filtered))
	}
	if filtered[0].TurnSequence != 2 || filtered[1].TurnSequence != 3 {
		t.Errorf("expected turns 2 and 3, got %d and %d", filtered[0].TurnSequence, filtered[1].TurnSequence)
	}
}

// TestAssistantMessagesTokenFiltering tests token filtering
func TestAssistantMessagesTokenFiltering(t *testing.T) {
	messages := []AssistantMessage{
		{TurnSequence: 1, TokensOutput: 100},
		{TurnSequence: 2, TokensOutput: 500},
		{TurnSequence: 3, TokensOutput: 1000},
		{TurnSequence: 4, TokensOutput: 2000},
	}

	filtered := filterAssistantMessagesByTokens(messages, 500)
	if len(filtered) != 3 {
		t.Fatalf("expected 3 messages with min_tokens=500, got %d", len(filtered))
	}
	if filtered[0].TurnSequence != 2 {
		t.Errorf("expected first turn to be 2, got %d", filtered[0].TurnSequence)
	}
}

// TestAssistantMessagesSorting tests deterministic sorting
func TestAssistantMessagesSorting(t *testing.T) {
	messages := []AssistantMessage{
		{TurnSequence: 3, TokensOutput: 100},
		{TurnSequence: 1, TokensOutput: 500},
		{TurnSequence: 2, TokensOutput: 200},
	}

	// Default: sort by turn_sequence
	sortAssistantMessages(messages, "turn_sequence", false)
	if messages[0].TurnSequence != 1 || messages[1].TurnSequence != 2 || messages[2].TurnSequence != 3 {
		t.Errorf("expected sorted by turn_sequence")
	}

	// Sort by tokens_output
	sortAssistantMessages(messages, "tokens_output", false)
	if messages[0].TokensOutput != 100 || messages[1].TokensOutput != 200 || messages[2].TokensOutput != 500 {
		t.Errorf("expected sorted by tokens_output")
	}

	// Reverse sort
	sortAssistantMessages(messages, "tokens_output", true)
	if messages[0].TokensOutput != 500 || messages[1].TokensOutput != 200 || messages[2].TokensOutput != 100 {
		t.Errorf("expected reverse sorted by tokens_output")
	}
}
