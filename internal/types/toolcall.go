package types

// ToolCall represents a complete tool invocation (ToolUse paired with its ToolResult).
// All JSON tags use snake_case to match the Claude Code JSONL schema.
type ToolCall struct {
	UUID      string                 `json:"uuid"`      // UUID of the SessionEntry containing the tool_use
	ToolName  string                 `json:"tool_name"` // name of the tool
	Input     map[string]interface{} `json:"input"`     // tool input parameters
	Output    string                 `json:"output"`    // tool output (ToolResult.Content)
	Status    string                 `json:"status"`    // execution status (success/error)
	Error     string                 `json:"error"`     // error message (if any)
	Timestamp string                 `json:"timestamp"` // ISO 8601 timestamp
}

// FileActionType maps a tool name to a file action category ("Read", "Edit", or "Write").
// Returns an empty string if the tool does not perform a file action.
func FileActionType(toolName string) string {
	switch toolName {
	case "Read":
		return "Read"
	case "Edit":
		return "Edit"
	case "Write":
		return "Write"
	case "NotebookEdit":
		return "Edit"
	default:
		return ""
	}
}

// ExtractToolCalls extracts all tool calls from a SessionEntry slice.
// It pairs each ToolUse with its corresponding ToolResult by tool_use_id.
//
// Ordering guarantee (DIR-058): the returned slice is in entry (JSONL) order —
// each ToolCall appears at the position of its tool_use block as entries are
// walked in slice order (and, within a single entry, in content-block order).
// The output is deterministic: this function never ranges over a map to build
// the result slice. If the same tool_use_id appears more than once, only the
// first occurrence is emitted. Callers that need a different order (e.g.
// turn-number order) must sort explicitly.
func ExtractToolCalls(entries []SessionEntry) []ToolCall {
	// Step 1: collect all ToolResult blocks indexed by tool_use_id
	toolResultMap := make(map[string]*ToolResult)
	for _, entry := range entries {
		if entry.Message == nil {
			continue
		}
		for _, block := range entry.Message.Content {
			if block.Type == "tool_result" && block.ToolResult != nil {
				toolResultMap[block.ToolResult.ToolUseID] = block.ToolResult
			}
		}
	}

	// Step 2: re-iterate entries in slice order and emit a ToolCall at each
	// tool_use block, looking up its paired result in toolResultMap. Iterating
	// entries (not a map) keeps the emit order deterministic across calls.
	var toolCalls []ToolCall
	emitted := make(map[string]struct{})
	for _, entry := range entries {
		if entry.Message == nil {
			continue
		}
		for _, block := range entry.Message.Content {
			if block.Type != "tool_use" || block.ToolUse == nil {
				continue
			}
			toolUseID := block.ToolUse.ID
			if _, dup := emitted[toolUseID]; dup {
				continue // one ToolCall per tool_use_id; first occurrence wins
			}
			emitted[toolUseID] = struct{}{}

			tc := ToolCall{
				UUID:      entry.UUID,
				ToolName:  block.ToolUse.Name,
				Input:     block.ToolUse.Input,
				Timestamp: entry.Timestamp,
			}
			if result, found := toolResultMap[toolUseID]; found {
				tc.Output = result.Content
				tc.Status = result.Status
				tc.Error = result.Error

				// Normalize Status from IsError when the JSONL omits a "status" field.
				// Real Claude Code tool_result blocks only carry "is_error":true/false,
				// so result.Status is "" for every successful call.
				if tc.Status == "" {
					if result.IsError {
						tc.Status = "error"
					} else {
						tc.Status = "success"
					}
				}
			}
			toolCalls = append(toolCalls, tc)
		}
	}

	return toolCalls
}
