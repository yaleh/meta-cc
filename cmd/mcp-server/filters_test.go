package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestTruncateMessageContent tests message content truncation
func TestTruncateMessageContent(t *testing.T) {
	tests := []struct {
		name           string
		messages       []interface{}
		maxLen         int
		expectTruncate bool
		expectContent  string
	}{
		{
			name: "truncate long content",
			messages: []interface{}{
				map[string]interface{}{
					"turn_sequence": float64(1),
					"timestamp":     "2025-10-06T12:00:00Z",
					"content":       strings.Repeat("a", 1000),
				},
			},
			maxLen:         500,
			expectTruncate: true,
			expectContent:  strings.Repeat("a", 500) + "... [TRUNCATED]",
		},
		{
			name: "short content not truncated",
			messages: []interface{}{
				map[string]interface{}{
					"turn_sequence": float64(1),
					"timestamp":     "2025-10-06T12:00:00Z",
					"content":       "short content",
				},
			},
			maxLen:         500,
			expectTruncate: false,
			expectContent:  "short content",
		},
		{
			name: "zero maxLen returns original",
			messages: []interface{}{
				map[string]interface{}{
					"content": strings.Repeat("a", 1000),
				},
			},
			maxLen:         0,
			expectTruncate: false,
			expectContent:  strings.Repeat("a", 1000),
		},
		{
			name: "negative maxLen returns original",
			messages: []interface{}{
				map[string]interface{}{
					"content": strings.Repeat("a", 1000),
				},
			},
			maxLen:         -1,
			expectTruncate: false,
			expectContent:  strings.Repeat("a", 1000),
		},
		{
			name:           "empty messages",
			messages:       []interface{}{},
			maxLen:         500,
			expectTruncate: false,
		},
		{
			name: "message without content field",
			messages: []interface{}{
				map[string]interface{}{
					"turn_sequence": float64(1),
					"timestamp":     "2025-10-06T12:00:00Z",
				},
			},
			maxLen:         500,
			expectTruncate: false,
		},
		{
			name: "non-map message",
			messages: []interface{}{
				"string message",
			},
			maxLen:         500,
			expectTruncate: false,
		},
		{
			name: "multiple messages with mixed lengths",
			messages: []interface{}{
				map[string]interface{}{
					"turn_sequence": float64(1),
					"content":       "short",
				},
				map[string]interface{}{
					"turn_sequence": float64(2),
					"content":       strings.Repeat("b", 1000),
				},
			},
			maxLen:         500,
			expectTruncate: false, // We'll check individually
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateMessageContent(tt.messages, tt.maxLen)

			if len(result) != len(tt.messages) {
				t.Errorf("expected %d messages, got %d", len(tt.messages), len(result))
				return
			}

			// Skip detailed checks for empty or special cases
			if len(tt.messages) == 0 || tt.name == "message without content field" || tt.name == "non-map message" {
				return
			}

			// Check first message
			if len(result) > 0 {
				msgMap, ok := result[0].(map[string]interface{})
				if !ok && tt.name != "non-map message" {
					t.Error("expected first result to be a map")
					return
				}

				if ok && tt.expectContent != "" {
					content, _ := msgMap["content"].(string)
					if content != tt.expectContent {
						t.Errorf("expected content=%q, got %q", tt.expectContent, content)
					}
				}

				if ok && tt.expectTruncate {
					truncated, _ := msgMap["content_truncated"].(bool)
					if !truncated {
						t.Error("expected content_truncated=true")
					}

					origLen, _ := msgMap["original_length"].(int)
					if origLen != 1000 {
						t.Errorf("expected original_length=1000, got %d", origLen)
					}
				}

				if ok && !tt.expectTruncate && tt.name == "short content not truncated" {
					if _, exists := msgMap["content_truncated"]; exists {
						t.Error("content_truncated should not exist for non-truncated messages")
					}
				}
			}

			// Special check for multiple messages test
			if tt.name == "multiple messages with mixed lengths" {
				// First message should not be truncated
				msg1, _ := result[0].(map[string]interface{})
				if _, exists := msg1["content_truncated"]; exists {
					t.Error("first message should not be truncated")
				}

				// Second message should be truncated
				msg2, _ := result[1].(map[string]interface{})
				truncated, _ := msg2["content_truncated"].(bool)
				if !truncated {
					t.Error("second message should be truncated")
				}
			}
		})
	}
}

// TestApplyContentSummary tests content summary mode
func TestApplyContentSummary(t *testing.T) {
	tests := []struct {
		name           string
		messages       []interface{}
		expectPreview  string
		expectFields   []string
		unexpectFields []string
	}{
		{
			name: "long content creates preview",
			messages: []interface{}{
				map[string]interface{}{
					"turn_sequence": float64(42),
					"timestamp":     "2025-10-06T12:00:00Z",
					"content":       strings.Repeat("a", 200),
					"extra_field":   "should be removed",
				},
			},
			expectPreview:  strings.Repeat("a", 100) + "...",
			expectFields:   []string{"turn_sequence", "timestamp", "content_preview"},
			unexpectFields: []string{"content", "extra_field"},
		},
		{
			name: "short content no ellipsis",
			messages: []interface{}{
				map[string]interface{}{
					"turn_sequence": float64(1),
					"timestamp":     "2025-10-06T12:00:00Z",
					"content":       "short",
				},
			},
			expectPreview:  "short",
			expectFields:   []string{"turn_sequence", "timestamp", "content_preview"},
			unexpectFields: []string{"content"},
		},
		{
			name: "missing content field",
			messages: []interface{}{
				map[string]interface{}{
					"turn_sequence": float64(1),
					"timestamp":     "2025-10-06T12:00:00Z",
				},
			},
			expectPreview:  "",
			expectFields:   []string{"turn_sequence", "timestamp", "content_preview"},
			unexpectFields: []string{"content"},
		},
		{
			name:          "empty messages",
			messages:      []interface{}{},
			expectFields:  []string{},
			expectPreview: "",
		},
		{
			name: "non-map message",
			messages: []interface{}{
				"string message",
			},
			expectFields: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyContentSummary(tt.messages)

			if len(result) != len(tt.messages) {
				t.Errorf("expected %d messages, got %d", len(tt.messages), len(result))
				return
			}

			if len(result) == 0 {
				return
			}

			// Check first message
			if tt.name != "non-map message" {
				msgMap, ok := result[0].(map[string]interface{})
				if !ok {
					if tt.name != "non-map message" {
						t.Error("expected first result to be a map")
					}
					return
				}

				// Check expected fields exist
				for _, field := range tt.expectFields {
					if _, exists := msgMap[field]; !exists {
						t.Errorf("expected field %q to exist", field)
					}
				}

				// Check unexpected fields don't exist
				for _, field := range tt.unexpectFields {
					if _, exists := msgMap[field]; exists {
						t.Errorf("field %q should not exist", field)
					}
				}

				// Check preview content
				if tt.expectPreview != "" {
					preview, _ := msgMap["content_preview"].(string)
					if preview != tt.expectPreview {
						t.Errorf("expected preview=%q, got %q", tt.expectPreview, preview)
					}
				}
			}
		})
	}
}

// TestTruncateMessageContent_Immutability tests that original messages are not mutated
func TestTruncateMessageContent_Immutability(t *testing.T) {
	original := []interface{}{
		map[string]interface{}{
			"turn_sequence": float64(1),
			"content":       strings.Repeat("a", 1000),
		},
	}

	// Store original content
	originalMap := original[0].(map[string]interface{})
	originalContent := originalMap["content"].(string)

	// Truncate
	TruncateMessageContent(original, 500)

	// Check original is unchanged
	afterContent := originalMap["content"].(string)
	if afterContent != originalContent {
		t.Error("TruncateMessageContent mutated the original message")
	}

	if len(afterContent) != 1000 {
		t.Errorf("original content length changed from 1000 to %d", len(afterContent))
	}
}

// TestApplyContentSummary_Immutability tests that original messages are not mutated
func TestApplyContentSummary_Immutability(t *testing.T) {
	original := []interface{}{
		map[string]interface{}{
			"turn_sequence": float64(1),
			"content":       strings.Repeat("a", 200),
			"extra_field":   "value",
		},
	}

	// Store original
	originalMap := original[0].(map[string]interface{})
	originalContent := originalMap["content"].(string)

	// Apply summary
	ApplyContentSummary(original)

	// Check original is unchanged
	afterContent := originalMap["content"].(string)
	if afterContent != originalContent {
		t.Error("ApplyContentSummary mutated the original message")
	}

	if _, exists := originalMap["extra_field"]; !exists {
		t.Error("ApplyContentSummary removed fields from original")
	}
}

// TestTruncateJSONL tests truncation with JSONL input/output
func TestTruncateJSONL(t *testing.T) {
	jsonlInput := `{"turn_sequence":1,"timestamp":"2025-10-06T12:00:00Z","content":"` + strings.Repeat("a", 1000) + `"}
{"turn_sequence":2,"timestamp":"2025-10-06T12:01:00Z","content":"short"}`

	// Parse JSONL to messages
	lines := strings.Split(strings.TrimSpace(jsonlInput), "\n")
	var messages []interface{}
	for _, line := range lines {
		var obj interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Fatalf("failed to parse JSONL: %v", err)
		}
		messages = append(messages, obj)
	}

	// Truncate
	truncated := TruncateMessageContent(messages, 500)

	// Verify first message is truncated
	msg1 := truncated[0].(map[string]interface{})
	content1 := msg1["content"].(string)
	if len(content1) != 515 { // 500 + "... [TRUNCATED]" (15 chars)
		t.Errorf("expected truncated content length=515, got %d", len(content1))
	}

	truncatedFlag := msg1["content_truncated"].(bool)
	if !truncatedFlag {
		t.Error("expected content_truncated=true")
	}

	// Verify second message is not truncated
	msg2 := truncated[1].(map[string]interface{})
	content2 := msg2["content"].(string)
	if content2 != "short" {
		t.Errorf("expected content=short, got %s", content2)
	}

	if _, exists := msg2["content_truncated"]; exists {
		t.Error("second message should not have content_truncated field")
	}
}
