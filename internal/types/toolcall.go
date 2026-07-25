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
func ExtractToolCalls(entries []SessionEntry) []ToolCall {
	// Step 1: collect all ToolUse blocks indexed by ID
	type toolUseRecord struct {
		uuid      string
		toolUse   *ToolUse
		timestamp string
	}
	toolUseMap := make(map[string]toolUseRecord)

	for _, entry := range entries {
		if entry.Message == nil {
			continue
		}
		for _, block := range entry.Message.Content {
			if block.Type == "tool_use" && block.ToolUse != nil {
				toolUseMap[block.ToolUse.ID] = toolUseRecord{
					uuid:      entry.UUID,
					toolUse:   block.ToolUse,
					timestamp: entry.Timestamp,
				}
			}
		}
	}

	// Step 2: collect all ToolResult blocks indexed by tool_use_id
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

	// Step 3: pair ToolUse with ToolResult
	var toolCalls []ToolCall
	for toolUseID, tu := range toolUseMap {
		tc := ToolCall{
			UUID:      tu.uuid,
			ToolName:  tu.toolUse.Name,
			Input:     tu.toolUse.Input,
			Timestamp: tu.timestamp,
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

	return toolCalls
}
