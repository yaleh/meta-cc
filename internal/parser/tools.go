package parser

// ToolCall represents a complete tool invocation (ToolUse + ToolResult)
// All JSON tags use snake_case to match Claude Code JSONL schema
type ToolCall struct {
	UUID      string                 `json:"uuid"`      // UUID of the SessionEntry containing the tool_use
	ToolName  string                 `json:"tool_name"` // Name of the tool
	Input     map[string]interface{} `json:"input"`     // Tool input parameters
	Output    string                 `json:"output"`    // Tool output (ToolResult.Content)
	Status    string                 `json:"status"`    // Execution status (success/error)
	Error     string                 `json:"error"`     // Error message (if any)
	Timestamp string                 `json:"timestamp"` // Timestamp of the tool call (ISO 8601 format)
}

// ExtractToolCalls extracts all tool calls from SessionEntry array
// Process:
//  1. Iterate all SessionEntry, collect ToolUse (indexed by ID)
//  2. Iterate all SessionEntry, find ToolResult, match by tool_use_id
//  3. Generate ToolCall array
func ExtractToolCalls(entries []SessionEntry) []ToolCall {
	// Step 1: Collect all ToolUse (indexed by ID)
	toolUseMap := make(map[string]struct {
		uuid      string
		toolUse   *ToolUse
		timestamp string
	})

	for _, entry := range entries {
		// Skip entries without Message
		if entry.Message == nil {
			continue
		}

		for _, block := range entry.Message.Content {
			if block.Type == "tool_use" && block.ToolUse != nil {
				toolUseMap[block.ToolUse.ID] = struct {
					uuid      string
					toolUse   *ToolUse
					timestamp string
				}{
					uuid:      entry.UUID,
					toolUse:   block.ToolUse,
					timestamp: entry.Timestamp,
				}
			}
		}
	}

	// Step 2: Collect all ToolResult (indexed by tool_use_id)
	toolResultMap := make(map[string]*ToolResult)

	for _, entry := range entries {
		// Skip entries without Message
		if entry.Message == nil {
			continue
		}

		for _, block := range entry.Message.Content {
			if block.Type == "tool_result" && block.ToolResult != nil {
				toolResultMap[block.ToolResult.ToolUseID] = block.ToolResult
			}
		}
	}

	// Step 3: Generate ToolCall array
	var toolCalls []ToolCall

	for toolUseID, tu := range toolUseMap {
		toolCall := ToolCall{
			UUID:      tu.uuid,
			ToolName:  tu.toolUse.Name,
			Input:     tu.toolUse.Input,
			Timestamp: tu.timestamp,
		}

		// Find matching ToolResult
		if result, found := toolResultMap[toolUseID]; found {
			toolCall.Output = result.Content
			toolCall.Status = result.Status
			toolCall.Error = result.Error
		}

		toolCalls = append(toolCalls, toolCall)
	}

	return toolCalls
}
