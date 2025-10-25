package main

const (
	// DefaultMaxMessageLength is the default for max_message_length parameter.
	// Set to 0 to disable truncation (use hybrid output mode for large results).
	DefaultMaxMessageLength = 0

	DefaultPreviewLength = 100
)

// TruncateMessageContent truncates the 'content' field in user messages
// to prevent context overflow from large session summaries.
//
// Use hybrid output mode instead of truncation for complete data preservation.
//
// Parameters:
//   - messages: Array of message objects (maps)
//   - maxLen: Maximum length for content field (0 or negative = no truncation)
//
// Returns:
//   - New array with truncated messages (originals unchanged)
//
// Truncated messages include:
//   - content_truncated: true
//   - original_length: original content length
func TruncateMessageContent(messages []interface{}, maxLen int) []interface{} {
	if maxLen <= 0 {
		return messages
	}

	truncated := make([]interface{}, len(messages))

	for i, msg := range messages {
		// Convert to map for manipulation
		msgMap, ok := msg.(map[string]interface{})
		if !ok {
			// Not a map, return as-is
			truncated[i] = msg
			continue
		}

		// Create copy to avoid mutating original
		newMap := make(map[string]interface{})
		for k, v := range msgMap {
			newMap[k] = v
		}

		// Truncate content field if present and exceeds maxLen
		if content, ok := newMap["content"].(string); ok {
			if len(content) > maxLen {
				newMap["content"] = content[:maxLen] + "... [TRUNCATED]"
				newMap["content_truncated"] = true
				newMap["original_length"] = len(content)
			}
		}

		truncated[i] = newMap
	}

	return truncated
}

// ApplyContentSummary returns only message metadata (no full content).
// Useful for pattern matching without needing full message text.
//
// Use hybrid output mode instead for complete data preservation.
//
// Parameters:
//   - messages: Array of message objects (maps)
//
// Returns:
//   - New array with summary objects containing:
//   - turn_sequence: message turn number
//   - timestamp: message timestamp
//   - content_preview: first 100 characters of content
//
// All other fields are removed to reduce output size.
func ApplyContentSummary(messages []interface{}) []interface{} {
	summary := make([]interface{}, len(messages))

	for i, msg := range messages {
		msgMap, ok := msg.(map[string]interface{})
		if !ok {
			// Not a map, return as-is
			summary[i] = msg
			continue
		}

		// Extract preview (first 100 chars)
		preview := ""
		if content, ok := msgMap["content"].(string); ok {
			if len(content) > DefaultPreviewLength {
				preview = content[:DefaultPreviewLength] + "..."
			} else {
				preview = content
			}
		}

		// Create summary object
		summary[i] = map[string]interface{}{
			"turn_sequence":   msgMap["turn_sequence"],
			"timestamp":       msgMap["timestamp"],
			"content_preview": preview,
		}
	}

	return summary
}
