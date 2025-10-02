package parser

// ToolCall 表示一个完整的工具调用（ToolUse + ToolResult）
type ToolCall struct {
	TurnSequence int                    // 工具调用所在的 Turn 序号
	ToolName     string                 // 工具名称
	Input        map[string]interface{} // 工具输入参数
	Output       string                 // 工具输出（ToolResult.Content）
	Status       string                 // 执行状态（success/error）
	Error        string                 // 错误信息（如果有）
}

// ExtractToolCalls 从 Turn 数组中提取所有工具调用
// 流程：
//  1. 遍历所有 Turn，收集 ToolUse（按 ID 索引）
//  2. 遍历所有 Turn，查找 ToolResult，匹配 tool_use_id
//  3. 生成 ToolCall 数组
func ExtractToolCalls(turns []Turn) []ToolCall {
	// Step 1: 收集所有 ToolUse（按 ID 索引）
	toolUseMap := make(map[string]struct {
		turnSeq int
		toolUse *ToolUse
	})

	for _, turn := range turns {
		for _, block := range turn.Content {
			if block.Type == "tool_use" && block.ToolUse != nil {
				toolUseMap[block.ToolUse.ID] = struct {
					turnSeq int
					toolUse *ToolUse
				}{
					turnSeq: turn.Sequence,
					toolUse: block.ToolUse,
				}
			}
		}
	}

	// Step 2: 收集所有 ToolResult（按 tool_use_id 索引）
	toolResultMap := make(map[string]*ToolResult)

	for _, turn := range turns {
		for _, block := range turn.Content {
			if block.Type == "tool_result" && block.ToolResult != nil {
				toolResultMap[block.ToolResult.ToolUseID] = block.ToolResult
			}
		}
	}

	// Step 3: 生成 ToolCall 数组
	var toolCalls []ToolCall

	for toolUseID, tu := range toolUseMap {
		toolCall := ToolCall{
			TurnSequence: tu.turnSeq,
			ToolName:     tu.toolUse.Name,
			Input:        tu.toolUse.Input,
		}

		// 查找匹配的 ToolResult
		if result, found := toolResultMap[toolUseID]; found {
			toolCall.Output = result.Content
			toolCall.Status = result.Status
			toolCall.Error = result.Error
		}

		toolCalls = append(toolCalls, toolCall)
	}

	return toolCalls
}
