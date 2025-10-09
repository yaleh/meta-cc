package cmd

import (
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/parser"
)

// TestBuildConversationTurns tests basic conversation turn pairing
func TestBuildConversationTurns(t *testing.T) {
	// Create mock entries with user+assistant pairs
	entries := []parser.SessionEntry{
		{
			UUID:      "user-1",
			Type:      "user",
			Timestamp: "2024-01-01T10:00:00Z",
			Message: &parser.Message{
				Content: []parser.ContentBlock{
					{Type: "text", Text: "Hello world"},
				},
			},
		},
		{
			UUID:      "assistant-1",
			Type:      "assistant",
			Timestamp: "2024-01-01T10:00:05Z",
			Message: &parser.Message{
				Model: "claude-3",
				Content: []parser.ContentBlock{
					{Type: "text", Text: "Hello! How can I help?"},
				},
				Usage: map[string]interface{}{
					"input_tokens":  float64(10),
					"output_tokens": float64(20),
				},
				StopReason: "end_turn",
			},
		},
		{
			UUID:      "user-2",
			Type:      "user",
			Timestamp: "2024-01-01T10:01:00Z",
			Message: &parser.Message{
				Content: []parser.ContentBlock{
					{Type: "text", Text: "Show me files"},
				},
			},
		},
		{
			UUID:      "assistant-2",
			Type:      "assistant",
			Timestamp: "2024-01-01T10:01:03Z",
			Message: &parser.Message{
				Model: "claude-3",
				Content: []parser.ContentBlock{
					{Type: "text", Text: "Here are the files:"},
				},
				Usage: map[string]interface{}{
					"input_tokens":  float64(15),
					"output_tokens": float64(25),
				},
			},
		},
	}

	turnIndex := map[string]int{
		"user-1":      1,
		"assistant-1": 1,
		"user-2":      2,
		"assistant-2": 2,
	}

	turns := buildConversationTurns(entries, turnIndex)

	if len(turns) != 2 {
		t.Fatalf("expected 2 turns, got %d", len(turns))
	}

	// Verify turn 1
	if turns[0].TurnSequence != 1 {
		t.Errorf("turn 1: expected sequence 1, got %d", turns[0].TurnSequence)
	}
	if turns[0].UserMessage == nil {
		t.Fatal("turn 1: user message is nil")
	}
	if turns[0].UserMessage.Content != "Hello world" {
		t.Errorf("turn 1: expected user content 'Hello world', got '%s'", turns[0].UserMessage.Content)
	}
	if turns[0].AssistantMessage == nil {
		t.Fatal("turn 1: assistant message is nil")
	}
	if turns[0].AssistantMessage.TextLength != 22 {
		t.Errorf("turn 1: expected text length 22, got %d", turns[0].AssistantMessage.TextLength)
	}
	if turns[0].Duration <= 0 {
		t.Errorf("turn 1: expected positive duration, got %d", turns[0].Duration)
	}

	// Verify turn 2
	if turns[1].TurnSequence != 2 {
		t.Errorf("turn 2: expected sequence 2, got %d", turns[1].TurnSequence)
	}
	if turns[1].UserMessage == nil {
		t.Fatal("turn 2: user message is nil")
	}
	if turns[1].AssistantMessage == nil {
		t.Fatal("turn 2: assistant message is nil")
	}
}

// TestConversationTurnDuration tests duration calculation
func TestConversationTurnDuration(t *testing.T) {
	tests := []struct {
		name          string
		userTime      string
		assistantTime string
		expectedMs    int
	}{
		{
			name:          "5 second difference",
			userTime:      "2024-01-01T10:00:00Z",
			assistantTime: "2024-01-01T10:00:05Z",
			expectedMs:    5000,
		},
		{
			name:          "1 minute difference",
			userTime:      "2024-01-01T10:00:00Z",
			assistantTime: "2024-01-01T10:01:00Z",
			expectedMs:    60000,
		},
		{
			name:          "sub-second difference",
			userTime:      "2024-01-01T10:00:00.100Z",
			assistantTime: "2024-01-01T10:00:00.500Z",
			expectedMs:    400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := calculateDuration(tt.userTime, tt.assistantTime)
			if duration != tt.expectedMs {
				t.Errorf("expected %d ms, got %d ms", tt.expectedMs, duration)
			}
		})
	}
}

// TestConversationTurnFiltering tests turn range filtering
func TestConversationTurnFiltering(t *testing.T) {
	turns := []ConversationTurn{
		{TurnSequence: 1},
		{TurnSequence: 2},
		{TurnSequence: 3},
		{TurnSequence: 4},
		{TurnSequence: 5},
	}

	// Test start-turn filter
	filtered := filterByTurnRange(turns, 3, -1)
	if len(filtered) != 3 {
		t.Errorf("start-turn=3: expected 3 turns, got %d", len(filtered))
	}
	if filtered[0].TurnSequence != 3 {
		t.Errorf("start-turn=3: expected first turn 3, got %d", filtered[0].TurnSequence)
	}

	// Test end-turn filter
	filtered = filterByTurnRange(turns, -1, 3)
	if len(filtered) != 3 {
		t.Errorf("end-turn=3: expected 3 turns, got %d", len(filtered))
	}
	if filtered[len(filtered)-1].TurnSequence != 3 {
		t.Errorf("end-turn=3: expected last turn 3, got %d", filtered[len(filtered)-1].TurnSequence)
	}

	// Test both filters
	filtered = filterByTurnRange(turns, 2, 4)
	if len(filtered) != 3 {
		t.Errorf("start-turn=2, end-turn=4: expected 3 turns, got %d", len(filtered))
	}
	if filtered[0].TurnSequence != 2 || filtered[len(filtered)-1].TurnSequence != 4 {
		t.Errorf("start-turn=2, end-turn=4: expected turns 2-4")
	}
}

// TestConversationPatternMatching tests pattern matching on user/assistant content
func TestConversationPatternMatching(t *testing.T) {
	turns := []ConversationTurn{
		{
			TurnSequence: 1,
			UserMessage: &UserMessage{
				Content: "fix the bug",
			},
			AssistantMessage: &AssistantMessage{
				ContentBlocks: []ContentBlock{
					{Type: "text", Text: "I'll help you debug the issue"},
				},
			},
		},
		{
			TurnSequence: 2,
			UserMessage: &UserMessage{
				Content: "show me the files",
			},
			AssistantMessage: &AssistantMessage{
				ContentBlocks: []ContentBlock{
					{Type: "text", Text: "Here are the files"},
				},
			},
		},
		{
			TurnSequence: 3,
			UserMessage: &UserMessage{
				Content: "check for errors",
			},
			AssistantMessage: &AssistantMessage{
				ContentBlocks: []ContentBlock{
					{Type: "text", Text: "No errors found"},
				},
			},
		},
	}

	// Test pattern on user content
	filtered := filterByPattern(turns, `bug`, "user")
	if len(filtered) != 1 {
		t.Errorf("pattern='bug' target=user: expected 1 turn, got %d", len(filtered))
	}
	if len(filtered) > 0 && filtered[0].TurnSequence != 1 {
		t.Errorf("pattern='bug' target=user: expected turn 1, got %d", filtered[0].TurnSequence)
	}

	// Test pattern on assistant content
	filtered = filterByPattern(turns, `files`, "assistant")
	if len(filtered) != 1 {
		t.Errorf("pattern='files' target=assistant: expected 1 turn, got %d", len(filtered))
	}
	if len(filtered) > 0 && filtered[0].TurnSequence != 2 {
		t.Errorf("pattern='files' target=assistant: expected turn 2, got %d", filtered[0].TurnSequence)
	}

	// Test pattern on any content
	filtered = filterByPattern(turns, `error`, "any")
	if len(filtered) != 1 {
		t.Errorf("pattern='error' target=any: expected 1 turn, got %d", len(filtered))
	}
	if len(filtered) > 0 && filtered[0].TurnSequence != 3 {
		t.Errorf("pattern='error' target=any: expected turn 3, got %d", filtered[0].TurnSequence)
	}
}

// TestConversationIncompleteTurns tests handling of incomplete turns
func TestConversationIncompleteTurns(t *testing.T) {
	entries := []parser.SessionEntry{
		{
			UUID:      "user-1",
			Type:      "user",
			Timestamp: "2024-01-01T10:00:00Z",
			Message: &parser.Message{
				Content: []parser.ContentBlock{
					{Type: "text", Text: "Hello"},
				},
			},
		},
		// No assistant response for turn 1
		{
			UUID:      "user-2",
			Type:      "user",
			Timestamp: "2024-01-01T10:01:00Z",
			Message: &parser.Message{
				Content: []parser.ContentBlock{
					{Type: "text", Text: "Are you there?"},
				},
			},
		},
		{
			UUID:      "assistant-2",
			Type:      "assistant",
			Timestamp: "2024-01-01T10:01:05Z",
			Message: &parser.Message{
				Model: "claude-3",
				Content: []parser.ContentBlock{
					{Type: "text", Text: "Yes, here now"},
				},
			},
		},
	}

	turnIndex := map[string]int{
		"user-1":      1,
		"user-2":      2,
		"assistant-2": 2,
	}

	turns := buildConversationTurns(entries, turnIndex)

	// Should have 2 turns
	if len(turns) != 2 {
		t.Fatalf("expected 2 turns (including incomplete), got %d", len(turns))
	}

	// Turn 1 should have user but no assistant
	if turns[0].UserMessage == nil {
		t.Error("turn 1: expected user message")
	}
	if turns[0].AssistantMessage != nil {
		t.Error("turn 1: expected no assistant message")
	}
	if turns[0].Duration != 0 {
		t.Errorf("turn 1: expected zero duration, got %d", turns[0].Duration)
	}

	// Turn 2 should have both
	if turns[1].UserMessage == nil {
		t.Error("turn 2: expected user message")
	}
	if turns[1].AssistantMessage == nil {
		t.Error("turn 2: expected assistant message")
	}
	if turns[1].Duration <= 0 {
		t.Errorf("turn 2: expected positive duration, got %d", turns[1].Duration)
	}
}

// TestConversationSorting tests deterministic ordering
func TestConversationSorting(t *testing.T) {
	turns := []ConversationTurn{
		{TurnSequence: 3, Duration: 5000},
		{TurnSequence: 1, Duration: 3000},
		{TurnSequence: 2, Duration: 10000},
	}

	// Sort by turn sequence
	sortConversationTurns(turns, "turn_sequence", false)
	if turns[0].TurnSequence != 1 || turns[1].TurnSequence != 2 || turns[2].TurnSequence != 3 {
		t.Error("sort by turn_sequence failed")
	}

	// Sort by duration
	sortConversationTurns(turns, "duration", false)
	if turns[0].Duration != 3000 || turns[1].Duration != 5000 || turns[2].Duration != 10000 {
		t.Error("sort by duration failed")
	}

	// Sort by duration (reverse)
	sortConversationTurns(turns, "duration", true)
	if turns[0].Duration != 10000 || turns[1].Duration != 5000 || turns[2].Duration != 3000 {
		t.Error("sort by duration (reverse) failed")
	}
}

// TestConversationDurationFiltering tests filtering by response time
func TestConversationDurationFiltering(t *testing.T) {
	turns := []ConversationTurn{
		{TurnSequence: 1, Duration: 1000},
		{TurnSequence: 2, Duration: 5000},
		{TurnSequence: 3, Duration: 10000},
		{TurnSequence: 4, Duration: 15000},
	}

	// Test min duration filter
	filtered := filterByDuration(turns, 5000, -1)
	if len(filtered) != 3 {
		t.Errorf("min-duration=5000: expected 3 turns, got %d", len(filtered))
	}

	// Test max duration filter
	filtered = filterByDuration(turns, -1, 10000)
	if len(filtered) != 3 {
		t.Errorf("max-duration=10000: expected 3 turns, got %d", len(filtered))
	}

	// Test both filters
	filtered = filterByDuration(turns, 5000, 10000)
	if len(filtered) != 2 {
		t.Errorf("min-duration=5000, max-duration=10000: expected 2 turns, got %d", len(filtered))
	}
}

// Helper function to parse timestamp
func parseTimestamp(ts string) time.Time {
	t, _ := time.Parse(time.RFC3339, ts)
	return t
}
