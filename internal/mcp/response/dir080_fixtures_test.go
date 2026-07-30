package response

// DIR-080 / ADR-009: representative result fixtures for self-describing
// file_ref metadata tests (shape description, bounded samples, generated jq
// recipes, projection). Shapes mirror what the consolidated query tools
// actually emit so recipes validate against realistic records.

// claudeFixtureRecords returns Claude-style query_session_content records:
// nested message objects, array content blocks, optional usage, and provider
// provenance.
func claudeFixtureRecords() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"type":      "assistant",
			"timestamp": "2026-07-01T10:00:00Z",
			"sessionId": "sess-a",
			"provider":  "claude",
			"message": map[string]interface{}{
				"role": "assistant",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": "hello world"},
				},
				"usage": map[string]interface{}{"input_tokens": 12.0, "output_tokens": 7.0},
			},
		},
		map[string]interface{}{
			"type":      "assistant",
			"timestamp": "2026-07-01T10:01:00Z",
			"sessionId": "sess-a",
			"provider":  "claude",
			// usage intentionally absent on this record: exercises Optional.
			"message": map[string]interface{}{
				"role": "assistant",
				"content": []interface{}{
					map[string]interface{}{"type": "tool_use", "name": "Read"},
				},
			},
		},
	}
}

// codexFixtureRecords returns Codex-style normalized records: a nested
// payload object distinct from Claude's message shape, plus provider
// provenance.
func codexFixtureRecords() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"type":      "response_item",
			"timestamp": "2026-07-02T09:00:00Z",
			"provider":  "codex",
			"payload": map[string]interface{}{
				"type": "message",
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "input_text", "text": "fix the build"},
				},
			},
		},
		map[string]interface{}{
			"type":      "response_item",
			"timestamp": "2026-07-02T09:01:00Z",
			"provider":  "codex",
			"payload": map[string]interface{}{
				"type":     "function_call",
				"name":     "shell",
				"is_error": true,
			},
		},
	}
}

// groupedFixtureRecords returns the shape querypkg.GroupBySession emits:
// session-level aggregates wrapping the flat turns array.
func groupedFixtureRecords() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"session_id":  "sess-a",
			"match_count": 2.0,
			"first_match": "2026-07-01T10:00:00Z",
			"last_match":  "2026-07-01T10:01:00Z",
			"turns": []interface{}{
				map[string]interface{}{"type": "assistant", "timestamp": "2026-07-01T10:00:00Z"},
				map[string]interface{}{"type": "assistant", "timestamp": "2026-07-01T10:01:00Z"},
			},
		},
		map[string]interface{}{
			"session_id":  "sess-b",
			"match_count": 1.0,
			"first_match": "2026-07-03T08:00:00Z",
			"last_match":  "2026-07-03T08:00:00Z",
			"turns": []interface{}{
				map[string]interface{}{"type": "user", "timestamp": "2026-07-03T08:00:00Z"},
			},
		},
	}
}

// heterogeneousFixtureRecords returns records where the same nested path
// (.message.content) is a string on some records and an array on others —
// the exact shape that breaks naive jq contracts.
func heterogeneousFixtureRecords() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"type":     "user",
			"provider": "claude",
			"message": map[string]interface{}{
				"role":    "user",
				"content": "plain string content",
			},
		},
		map[string]interface{}{
			"type":     "assistant",
			"provider": "claude",
			"message": map[string]interface{}{
				"role": "assistant",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": "array content"},
				},
			},
		},
		map[string]interface{}{
			"type":     "user",
			"provider": "codex",
			"message":  nil, // nullable message
		},
	}
}

// sensitiveFixtureRecords contains secret-looking fields and oversized
// strings for redaction / truncation tests.
func sensitiveFixtureRecords() []interface{} {
	long := make([]byte, 1000)
	for i := range long {
		long[i] = 'x'
	}
	return []interface{}{
		map[string]interface{}{
			"type":          "assistant",
			"api_key":       "sk-live-1234567890",
			"Authorization": "Bearer eyJhbGciOi.secret-payload",
			"session_token": "tok_abcdef",
			"nested": map[string]interface{}{
				"client_secret": "shhh",
				"password":      "hunter2",
			},
			"text":      string(long),
			"safe_note": "short",
		},
	}
}
