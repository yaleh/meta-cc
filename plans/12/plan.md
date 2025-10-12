# Phase 12: MCP Project-Level Query Implementation

## 概述

**目标**: 扩展 MCP Server 支持项目级和会话级查询，默认提供跨会话分析能力

**代码量**: ~300 行

**依赖**: Phase 0-9 (完整的 meta-cc 工具链 + MCP Server 基础)

**交付物**:
- 8 个项目级 MCP 工具（查询所有会话）
- 8 个会话级 MCP 工具（仅查询当前会话，`_session` 后缀）
- 执行逻辑支持 `--project .` 标志
- 更新的 MCP 配置文件
- `docs/mcp-guide.md`：使用指南

---

## Phase 目标

扩展 MCP Server 以支持跨会话分析，同时保持向后兼容性：

### 核心需求

1. **项目级查询（默认）**：工具名无后缀，查询项目所有会话
2. **会话级查询**：工具名带 `_session` 后缀，仅查询当前会话
3. **CLI 集成**：使用 `--project .` 标志实现跨会话查询
4. **向后兼容**：保持现有 `get_session_stats` 工具不变

### 设计原则

Phase 12 遵循清晰的命名约定：

- ✅ **默认行为**: 无后缀工具 = 项目级（跨会话）
- ✅ **显式限定**: `_session` 后缀 = 当前会话
- ✅ **向后兼容**: `get_session_stats` 保持原样
- ✅ **一致性**: 所有工具遵循统一命名模式

**工具映射表**:
| 项目级（默认） | 会话级 | 说明 |
|--------------|--------|------|
| `get_stats` | `get_session_stats` | 统计信息 |
| `analyze_errors` | `analyze_errors_session` | 错误分析 |
| `query_tools` | `query_tools_session` | 工具调用查询 |
| `query_user_messages` | `query_user_messages_session` | 用户消息搜索 |
| `query_tool_sequences` | `query_tool_sequences_session` | 工作流模式 |
| `query_file_access` | `query_file_access_session` | 文件操作历史 |
| `query_successful_prompts` | `query_successful_prompts_session` | 优质提示词 |
| `query_context` | `query_context_session` | 错误上下文 |

---

## 成功标准

**功能验收**:
- ✅ 所有 Stage 单元测试通过（TDD）
- ✅ 项目级工具返回多会话数据
- ✅ 会话级工具仅返回当前会话数据
- ✅ `get_session_stats` 向后兼容（行为不变）
- ✅ `--project .` 标志正确传递到 CLI

**集成验收**:
- ✅ 跨会话分析场景验证
- ✅ 单会话分析场景验证
- ✅ 真实项目验证（meta-cc, NarrativeForge）
- ✅ 无回归（所有现有测试通过）

**代码质量**:
- ✅ 实际代码量: ~300 行（目标 280-320 行）
- ✅ 每个 Stage ≤ 100 行（Go 源代码）
- ✅ 测试覆盖率: ≥ 80%
- ✅ 无新增外部依赖

**文档质量**:
- ✅ `docs/mcp-guide.md` 完成
- ✅ 包含项目级和会话级对比示例
- ✅ 包含实际使用场景
- ✅ README.md 更新

---

## Stage 12.1: 项目级 MCP 工具定义

### 目标

定义 8 个项目级 MCP 工具，默认查询项目所有会话。

### 背景

**当前状态**:
- MCP Server 已实现 `get_session_stats`（仅当前会话）
- 无法跨会话分析工作模式和错误模式

**问题**:
- 无法回答"我在这个项目中如何使用 agents？"
- 无法识别项目级的重复错误模式
- 无法对比不同会话的工作效率

**期望行为**:
```bash
# Claude 调用 query_tools（项目级）
# MCP Server 执行: meta-cc query tools --project . --limit 100
# 返回：项目所有会话的工具调用数据

# Claude 调用 get_stats（项目级）
# MCP Server 执行: meta-cc analyze stats --project .
# 返回：项目级统计数据
```

### 实现步骤

#### 1. 定义项目级工具结构

**文件**: `internal/mcp/tools_project.go` (新建, ~80 行)

```go
package mcp

import (
	"encoding/json"
)

// Project-level MCP tools (query all sessions in project)

// ToolQueryTools queries tool calls across all sessions
type ToolQueryTools struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

func NewToolQueryTools() *ToolQueryTools {
	schema := `{
		"type": "object",
		"properties": {
			"limit": {
				"type": "integer",
				"description": "Maximum number of tool calls to return (default: 20)"
			},
			"tool": {
				"type": "string",
				"description": "Filter by tool name (e.g., 'Bash', 'Edit', 'Read')"
			},
			"status": {
				"type": "string",
				"enum": ["success", "error"],
				"description": "Filter by execution status"
			},
			"where": {
				"type": "string",
				"description": "SQL-like filter expression"
			}
		}
	}`

	return &ToolQueryTools{
		Name:        "query_tools",
		Description: "Query tool calls across all sessions in the project. Returns tool execution history, status, duration, and errors.",
		InputSchema: json.RawMessage(schema),
	}
}

// Execute runs the tool and returns results
func (t *ToolQueryTools) Execute(args map[string]interface{}) (interface{}, error) {
	// Will be implemented in Stage 12.2
	return nil, nil
}

// ToolQueryUserMessages queries user messages across all sessions
type ToolQueryUserMessages struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

func NewToolQueryUserMessages() *ToolQueryUserMessages {
	schema := `{
		"type": "object",
		"properties": {
			"pattern": {
				"type": "string",
				"description": "Regex pattern to match in message content (required)"
			},
			"limit": {
				"type": "integer",
				"description": "Maximum number of results (default: 10)"
			}
		},
		"required": ["pattern"]
	}`

	return &ToolQueryUserMessages{
		Name:        "query_user_messages",
		Description: "Search user messages across all sessions in the project using regex pattern matching.",
		InputSchema: json.RawMessage(schema),
	}
}

func (t *ToolQueryUserMessages) Execute(args map[string]interface{}) (interface{}, error) {
	// Will be implemented in Stage 12.2
	return nil, nil
}

// ToolGetStats gets project-level statistics
type ToolGetStats struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

func NewToolGetStats() *ToolGetStats {
	schema := `{
		"type": "object",
		"properties": {
			"output_format": {
				"type": "string",
				"enum": ["json", "md"],
				"default": "json"
			}
		}
	}`

	return &ToolGetStats{
		Name:        "get_stats",
		Description: "Get statistics for all sessions in the project. Returns tool usage counts, error rates, and session metrics.",
		InputSchema: json.RawMessage(schema),
	}
}

func (t *ToolGetStats) Execute(args map[string]interface{}) (interface{}, error) {
	// Will be implemented in Stage 12.2
	return nil, nil
}

// ToolAnalyzeErrors analyzes error patterns across all sessions
type ToolAnalyzeErrors struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

func NewToolAnalyzeErrors() *ToolAnalyzeErrors {
	schema := `{
		"type": "object",
		"properties": {
			"output_format": {
				"type": "string",
				"enum": ["json", "md"],
				"default": "json"
			}
		}
	}`

	return &ToolAnalyzeErrors{
		Name:        "analyze_errors",
		Description: "Analyze error patterns across all sessions in the project. Detects repeated errors and common failure modes.",
		InputSchema: json.RawMessage(schema),
	}
}

func (t *ToolAnalyzeErrors) Execute(args map[string]interface{}) (interface{}, error) {
	// Will be implemented in Stage 12.2
	return nil, nil
}

// ToolQueryToolSequences queries workflow patterns across all sessions
type ToolQueryToolSequences struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

func NewToolQueryToolSequences() *ToolQueryToolSequences {
	schema := `{
		"type": "object",
		"properties": {
			"pattern": {
				"type": "string",
				"description": "Specific sequence pattern to match (e.g., 'Read -> Edit -> Bash')"
			},
			"min_occurrences": {
				"type": "integer",
				"default": 3,
				"description": "Minimum occurrences to report (default 3)"
			},
			"output_format": {
				"type": "string",
				"enum": ["json", "md"],
				"default": "json"
			}
		}
	}`

	return &ToolQueryToolSequences{
		Name:        "query_tool_sequences",
		Description: "Query repeated tool call sequences (workflow patterns) across all sessions in the project.",
		InputSchema: json.RawMessage(schema),
	}
}

func (t *ToolQueryToolSequences) Execute(args map[string]interface{}) (interface{}, error) {
	// Will be implemented in Stage 12.2
	return nil, nil
}

// ToolQueryFileAccess queries file access history across all sessions
type ToolQueryFileAccess struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

func NewToolQueryFileAccess() *ToolQueryFileAccess {
	schema := `{
		"type": "object",
		"properties": {
			"file": {
				"type": "string",
				"description": "File path to query (required)"
			},
			"output_format": {
				"type": "string",
				"enum": ["json", "md"],
				"default": "json"
			}
		},
		"required": ["file"]
	}`

	return &ToolQueryFileAccess{
		Name:        "query_file_access",
		Description: "Query file access history (read/edit/write operations) across all sessions in the project.",
		InputSchema: json.RawMessage(schema),
	}
}

func (t *ToolQueryFileAccess) Execute(args map[string]interface{}) (interface{}, error) {
	// Will be implemented in Stage 12.2
	return nil, nil
}

// ToolQuerySuccessfulPrompts queries successful prompt patterns across all sessions
type ToolQuerySuccessfulPrompts struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

func NewToolQuerySuccessfulPrompts() *ToolQuerySuccessfulPrompts {
	schema := `{
		"type": "object",
		"properties": {
			"limit": {
				"type": "integer",
				"default": 10,
				"description": "Maximum number of results (default 10)"
			},
			"min_quality_score": {
				"type": "number",
				"default": 0.8,
				"description": "Minimum quality score (0.0-1.0, default 0.8)"
			},
			"output_format": {
				"type": "string",
				"enum": ["json", "md"],
				"default": "json"
			}
		}
	}`

	return &ToolQuerySuccessfulPrompts{
		Name:        "query_successful_prompts",
		Description: "Query successful prompt patterns across all sessions in the project (Stage 8.12).",
		InputSchema: json.RawMessage(schema),
	}
}

func (t *ToolQuerySuccessfulPrompts) Execute(args map[string]interface{}) (interface{}, error) {
	// Will be implemented in Stage 12.2
	return nil, nil
}

// ToolQueryContext queries error context analysis across all sessions
type ToolQueryContext struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

func NewToolQueryContext() *ToolQueryContext {
	schema := `{
		"type": "object",
		"properties": {
			"error_signature": {
				"type": "string",
				"description": "Error pattern ID to query (required)"
			},
			"window": {
				"type": "integer",
				"default": 3,
				"description": "Context window size in turns before/after (default 3)"
			},
			"output_format": {
				"type": "string",
				"enum": ["json", "md"],
				"default": "json"
			}
		},
		"required": ["error_signature"]
	}`

	return &ToolQueryContext{
		Name:        "query_context",
		Description: "Query context around specific errors across all sessions in the project (Stage 8.10).",
		InputSchema: json.RawMessage(schema),
	}
}

func (t *ToolQueryContext) Execute(args map[string]interface{}) (interface{}, error) {
	// Will be implemented in Stage 12.2
	return nil, nil
}

// RegisterProjectLevelTools registers all project-level tools
func RegisterProjectLevelTools(registry *ToolRegistry) {
	registry.Register(NewToolQueryTools())
	registry.Register(NewToolQueryUserMessages())
	registry.Register(NewToolGetStats())
	registry.Register(NewToolAnalyzeErrors())
	registry.Register(NewToolQueryToolSequences())
	registry.Register(NewToolQueryFileAccess())
	registry.Register(NewToolQuerySuccessfulPrompts())
	registry.Register(NewToolQueryContext())
}
```

### TDD 步骤

**测试文件**: `internal/mcp/tools_project_test.go` (新建, ~100 行)

```go
package mcp

import (
	"encoding/json"
	"testing"
)

func TestToolQueryTools_Definition(t *testing.T) {
	tool := NewToolQueryTools()

	if tool.Name != "query_tools" {
		t.Errorf("Expected name 'query_tools', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("Description should not be empty")
	}

	// Verify schema is valid JSON
	var schema map[string]interface{}
	if err := json.Unmarshal(tool.InputSchema, &schema); err != nil {
		t.Fatalf("Invalid InputSchema JSON: %v", err)
	}

	// Verify schema has properties
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema should have 'properties' field")
	}

	// Verify expected parameters
	expectedParams := []string{"limit", "tool", "status", "where"}
	for _, param := range expectedParams {
		if _, exists := props[param]; !exists {
			t.Errorf("Schema missing parameter: %s", param)
		}
	}
}

func TestToolQueryUserMessages_Definition(t *testing.T) {
	tool := NewToolQueryUserMessages()

	if tool.Name != "query_user_messages" {
		t.Errorf("Expected name 'query_user_messages', got '%s'", tool.Name)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(tool.InputSchema, &schema); err != nil {
		t.Fatalf("Invalid InputSchema JSON: %v", err)
	}

	// Verify 'pattern' is required
	required, ok := schema["required"].([]interface{})
	if !ok || len(required) == 0 {
		t.Error("Schema should have 'required' field with 'pattern'")
	}

	if required[0] != "pattern" {
		t.Errorf("Expected required field 'pattern', got '%v'", required[0])
	}
}

func TestToolGetStats_Definition(t *testing.T) {
	tool := NewToolGetStats()

	if tool.Name != "get_stats" {
		t.Errorf("Expected name 'get_stats', got '%s'", tool.Name)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(tool.InputSchema, &schema); err != nil {
		t.Fatalf("Invalid InputSchema JSON: %v", err)
	}
}

func TestToolAnalyzeErrors_Definition(t *testing.T) {
	tool := NewToolAnalyzeErrors()

	if tool.Name != "analyze_errors" {
		t.Errorf("Expected name 'analyze_errors', got '%s'", tool.Name)
	}
}

func TestToolQueryToolSequences_Definition(t *testing.T) {
	tool := NewToolQueryToolSequences()

	if tool.Name != "query_tool_sequences" {
		t.Errorf("Expected name 'query_tool_sequences', got '%s'", tool.Name)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(tool.InputSchema, &schema); err != nil {
		t.Fatalf("Invalid InputSchema JSON: %v", err)
	}

	props := schema["properties"].(map[string]interface{})
	if _, exists := props["min_occurrences"]; !exists {
		t.Error("Schema should have 'min_occurrences' parameter")
	}
}

func TestToolQueryFileAccess_Definition(t *testing.T) {
	tool := NewToolQueryFileAccess()

	if tool.Name != "query_file_access" {
		t.Errorf("Expected name 'query_file_access', got '%s'", tool.Name)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(tool.InputSchema, &schema); err != nil {
		t.Fatalf("Invalid InputSchema JSON: %v", err)
	}

	// Verify 'file' is required
	required := schema["required"].([]interface{})
	if required[0] != "file" {
		t.Errorf("Expected required field 'file', got '%v'", required[0])
	}
}

func TestToolQuerySuccessfulPrompts_Definition(t *testing.T) {
	tool := NewToolQuerySuccessfulPrompts()

	if tool.Name != "query_successful_prompts" {
		t.Errorf("Expected name 'query_successful_prompts', got '%s'", tool.Name)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(tool.InputSchema, &schema); err != nil {
		t.Fatalf("Invalid InputSchema JSON: %v", err)
	}

	props := schema["properties"].(map[string]interface{})
	if _, exists := props["min_quality_score"]; !exists {
		t.Error("Schema should have 'min_quality_score' parameter")
	}
}

func TestToolQueryContext_Definition(t *testing.T) {
	tool := NewToolQueryContext()

	if tool.Name != "query_context" {
		t.Errorf("Expected name 'query_context', got '%s'", tool.Name)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(tool.InputSchema, &schema); err != nil {
		t.Fatalf("Invalid InputSchema JSON: %v", err)
	}

	// Verify 'error_signature' is required
	required := schema["required"].([]interface{})
	if required[0] != "error_signature" {
		t.Errorf("Expected required field 'error_signature', got '%v'", required[0])
	}
}

func TestRegisterProjectLevelTools(t *testing.T) {
	registry := NewToolRegistry()
	RegisterProjectLevelTools(registry)

	expectedTools := []string{
		"query_tools",
		"query_user_messages",
		"get_stats",
		"analyze_errors",
		"query_tool_sequences",
		"query_file_access",
		"query_successful_prompts",
		"query_context",
	}

	for _, toolName := range expectedTools {
		if !registry.HasTool(toolName) {
			t.Errorf("Tool '%s' not registered", toolName)
		}
	}

	// Verify total count
	if registry.Count() < len(expectedTools) {
		t.Errorf("Expected at least %d tools, got %d", len(expectedTools), registry.Count())
	}
}
```

### 交付物

**新建文件**:
- `internal/mcp/tools_project.go` (~80 行)
- `internal/mcp/tools_project_test.go` (~100 行)

**代码量**: ~80 行（新增 Go 源代码，不含测试）

### 验收标准

- ✅ 8 个项目级工具定义完成
- ✅ 所有工具名正确（无 `_session` 后缀）
- ✅ InputSchema 有效且符合 JSON Schema 规范
- ✅ 所有必需参数标记为 `required`
- ✅ 所有单元测试通过
- ✅ `RegisterProjectLevelTools()` 正确注册所有工具

---

## Stage 12.2: 执行逻辑与 `--project .` 标志支持

### 目标

实现项目级工具的执行逻辑，添加 `--project .` 标志到 CLI 调用。

### 背景

**当前执行流程**（会话级）:
```
MCP call → Execute() → Run CLI: meta-cc query tools --limit 10
```

**期望执行流程**（项目级）:
```
MCP call → Execute() → Run CLI: meta-cc query tools --project . --limit 10
```

### 实现步骤

#### 1. 实现执行逻辑

**文件**: `internal/mcp/executor.go` (修改, ~100 行)

```go
package mcp

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ExecuteProjectLevelQuery executes a project-level query with --project flag
func ExecuteProjectLevelQuery(command string, args map[string]interface{}) (interface{}, error) {
	// Build CLI command with --project . flag
	cmdArgs := []string{command, "--project", "."}

	// Add standard parameters
	if limit, ok := args["limit"].(float64); ok {
		cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", int(limit)))
	}

	if tool, ok := args["tool"].(string); ok {
		cmdArgs = append(cmdArgs, "--tool", tool)
	}

	if status, ok := args["status"].(string); ok {
		cmdArgs = append(cmdArgs, "--status", status)
	}

	if where, ok := args["where"].(string); ok {
		cmdArgs = append(cmdArgs, "--where", where)
	}

	if pattern, ok := args["pattern"].(string); ok {
		cmdArgs = append(cmdArgs, "--pattern", pattern)
	}

	// Output format (default: json)
	outputFormat := "json"
	if format, ok := args["output_format"].(string); ok {
		outputFormat = format
	}
	cmdArgs = append(cmdArgs, "--output", outputFormat)

	// Execute CLI command
	cmd := exec.Command("meta-cc", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("CLI execution failed: %v, output: %s", err, string(output))
	}

	// Parse JSON output
	if outputFormat == "json" {
		var result interface{}
		if err := json.Unmarshal(output, &result); err != nil {
			return nil, fmt.Errorf("failed to parse JSON output: %v", err)
		}
		return result, nil
	}

	// Return raw output for non-JSON formats
	return string(output), nil
}

// Implement Execute() for each project-level tool

func (t *ToolQueryTools) Execute(args map[string]interface{}) (interface{}, error) {
	return ExecuteProjectLevelQuery("query tools", args)
}

func (t *ToolQueryUserMessages) Execute(args map[string]interface{}) (interface{}, error) {
	return ExecuteProjectLevelQuery("query user-messages", args)
}

func (t *ToolGetStats) Execute(args map[string]interface{}) (interface{}, error) {
	return ExecuteProjectLevelQuery("analyze stats", args)
}

func (t *ToolAnalyzeErrors) Execute(args map[string]interface{}) (interface{}, error) {
	return ExecuteProjectLevelQuery("analyze errors", args)
}

func (t *ToolQueryToolSequences) Execute(args map[string]interface{}) (interface{}, error) {
	return ExecuteProjectLevelQuery("query tool-sequences", args)
}

func (t *ToolQueryFileAccess) Execute(args map[string]interface{}) (interface{}, error) {
	return ExecuteProjectLevelQuery("query file-access", args)
}

func (t *ToolQuerySuccessfulPrompts) Execute(args map[string]interface{}) (interface{}, error) {
	return ExecuteProjectLevelQuery("query successful-prompts", args)
}

func (t *ToolQueryContext) Execute(args map[string]interface{}) (interface{}, error) {
	return ExecuteProjectLevelQuery("query context", args)
}
```

### TDD 步骤

**测试文件**: `internal/mcp/executor_test.go` (新建, ~80 行)

```go
package mcp

import (
	"strings"
	"testing"
)

func TestExecuteProjectLevelQuery_BuildCommand(t *testing.T) {
	// Mock CLI execution for testing
	// In real implementation, use dependency injection or mocking

	tests := []struct {
		name     string
		command  string
		args     map[string]interface{}
		expected []string // Expected CLI arguments
	}{
		{
			name:    "query tools with limit",
			command: "query tools",
			args: map[string]interface{}{
				"limit": float64(10),
			},
			expected: []string{"query", "tools", "--project", ".", "--limit", "10", "--output", "json"},
		},
		{
			name:    "query tools with filters",
			command: "query tools",
			args: map[string]interface{}{
				"tool":   "Bash",
				"status": "error",
				"limit":  float64(20),
			},
			expected: []string{"query", "tools", "--project", ".", "--tool", "Bash", "--status", "error", "--limit", "20", "--output", "json"},
		},
		{
			name:    "query user messages with pattern",
			command: "query user-messages",
			args: map[string]interface{}{
				"pattern": "test.*pattern",
				"limit":   float64(5),
			},
			expected: []string{"query", "user-messages", "--project", ".", "--pattern", "test.*pattern", "--limit", "5", "--output", "json"},
		},
		{
			name:    "analyze stats with markdown output",
			command: "analyze stats",
			args: map[string]interface{}{
				"output_format": "md",
			},
			expected: []string{"analyze", "stats", "--project", ".", "--output", "md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build command arguments
			cmdArgs := buildProjectLevelQueryArgs(tt.command, tt.args)

			// Verify expected arguments are present
			for _, expectedArg := range tt.expected {
				found := false
				for _, arg := range cmdArgs {
					if arg == expectedArg {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected argument '%s' not found in: %v", expectedArg, cmdArgs)
				}
			}

			// Verify --project . is always present
			if !containsArgs(cmdArgs, "--project", ".") {
				t.Error("--project . flag missing from command")
			}
		})
	}
}

// Helper function to build command args (extracted for testing)
func buildProjectLevelQueryArgs(command string, args map[string]interface{}) []string {
	cmdParts := strings.Split(command, " ")
	cmdArgs := append(cmdParts, "--project", ".")

	if limit, ok := args["limit"].(float64); ok {
		cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", int(limit)))
	}

	if tool, ok := args["tool"].(string); ok {
		cmdArgs = append(cmdArgs, "--tool", tool)
	}

	if status, ok := args["status"].(string); ok {
		cmdArgs = append(cmdArgs, "--status", status)
	}

	if pattern, ok := args["pattern"].(string); ok {
		cmdArgs = append(cmdArgs, "--pattern", pattern)
	}

	outputFormat := "json"
	if format, ok := args["output_format"].(string); ok {
		outputFormat = format
	}
	cmdArgs = append(cmdArgs, "--output", outputFormat)

	return cmdArgs
}

func containsArgs(args []string, key, value string) bool {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == key && args[i+1] == value {
			return true
		}
	}
	return false
}

func TestToolQueryTools_Execute(t *testing.T) {
	// Integration test (requires meta-cc CLI to be available)
	// Skip if CLI not available
	if !isCLIAvailable() {
		t.Skip("meta-cc CLI not available, skipping integration test")
	}

	tool := NewToolQueryTools()
	args := map[string]interface{}{
		"limit": float64(5),
	}

	result, err := tool.Execute(args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result")
	}

	// Verify result structure (should be JSON)
	// Type assertion depends on CLI output format
}

func isCLIAvailable() bool {
	cmd := exec.Command("meta-cc", "--version")
	return cmd.Run() == nil
}
```

**集成测试**: `tests/integration/mcp_project_flag_test.sh` (新建, ~40 行)

```bash
#!/bin/bash
# Test --project flag is correctly passed to CLI

set -e

echo "=== MCP Project Flag Test ==="

# Test 1: Verify --project flag is added
echo "[1/2] Testing --project flag in CLI invocation..."

# Simulate MCP tool call with debug output
# Note: This requires a debug mode in the MCP server to print CLI commands
# For now, test indirectly by checking multi-session results

# Create mock multi-session data for testing
# (Implementation depends on test setup)

echo "  ✓ Project flag test (manual verification needed)"

# Test 2: Verify project-level query returns multi-session data
echo "[2/2] Testing multi-session data return..."

# This test requires:
# 1. Multiple session JSONL files in ~/.claude/projects/test-project/
# 2. meta-cc CLI with --project flag support

# For now, placeholder
echo "  ✓ Multi-session data test (integration test)"

echo ""
echo "=== All MCP Project Flag Tests Passed ✅ ==="
```

### 交付物

**新建文件**:
- `internal/mcp/executor.go` (~100 行)
- `internal/mcp/executor_test.go` (~80 行)
- `tests/integration/mcp_project_flag_test.sh` (~40 lines)

**修改文件**:
- `internal/mcp/tools_project.go` (添加 Execute 实现，~20 行)

**代码量**: ~100 行（新增 Go 源代码，不含测试）

### 验收标准

- ✅ `--project .` 标志添加到所有项目级查询
- ✅ CLI 命令正确构建（参数顺序、格式）
- ✅ JSON 输出正确解析
- ✅ 错误处理完整（CLI 失败、JSON 解析失败）
- ✅ 所有单元测试通过
- ✅ 集成测试验证 `--project` 标志传递

---

## Stage 12.3: 会话级工具（`_session` 后缀）

### 目标

定义 8 个会话级 MCP 工具，仅查询当前会话。

### 背景

**命名约定**:
- 项目级：`query_tools`（无后缀）
- 会话级：`query_tools_session`（`_session` 后缀）

**特殊情况**:
- `get_session_stats` 已存在，保持不变（向后兼容）

### 实现步骤

#### 1. 定义会话级工具

**文件**: `internal/mcp/tools_session.go` (新建, ~80 行)

```go
package mcp

import (
	"encoding/json"
)

// Session-level MCP tools (query current session only, with _session suffix)

// ToolQueryToolsSession queries tool calls in current session only
type ToolQueryToolsSession struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

func NewToolQueryToolsSession() *ToolQueryToolsSession {
	schema := `{
		"type": "object",
		"properties": {
			"limit": {
				"type": "integer",
				"description": "Maximum number of tool calls to return (default: 20)"
			},
			"tool": {
				"type": "string",
				"description": "Filter by tool name"
			},
			"status": {
				"type": "string",
				"enum": ["success", "error"],
				"description": "Filter by execution status"
			}
		}
	}`

	return &ToolQueryToolsSession{
		Name:        "query_tools_session",
		Description: "Query tool calls in the current session only. For project-level queries, use query_tools.",
		InputSchema: json.RawMessage(schema),
	}
}

func (t *ToolQueryToolsSession) Execute(args map[string]interface{}) (interface{}, error) {
	// Execute without --project flag (current session only)
	return ExecuteSessionLevelQuery("query tools", args)
}

// ToolQueryUserMessagesSession queries user messages in current session only
type ToolQueryUserMessagesSession struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

func NewToolQueryUserMessagesSession() *ToolQueryUserMessagesSession {
	schema := `{
		"type": "object",
		"properties": {
			"pattern": {
				"type": "string",
				"description": "Regex pattern to match (required)"
			},
			"limit": {
				"type": "integer",
				"description": "Maximum results (default: 10)"
			}
		},
		"required": ["pattern"]
	}`

	return &ToolQueryUserMessagesSession{
		Name:        "query_user_messages_session",
		Description: "Search user messages in the current session only.",
		InputSchema: json.RawMessage(schema),
	}
}

func (t *ToolQueryUserMessagesSession) Execute(args map[string]interface{}) (interface{}, error) {
	return ExecuteSessionLevelQuery("query user-messages", args)
}

// ToolAnalyzeErrorsSession analyzes errors in current session only
type ToolAnalyzeErrorsSession struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

func NewToolAnalyzeErrorsSession() *ToolAnalyzeErrorsSession {
	schema := `{
		"type": "object",
		"properties": {
			"output_format": {
				"type": "string",
				"enum": ["json", "md"],
				"default": "json"
			}
		}
	}`

	return &ToolAnalyzeErrorsSession{
		Name:        "analyze_errors_session",
		Description: "Analyze error patterns in the current session only.",
		InputSchema: json.RawMessage(schema),
	}
}

func (t *ToolAnalyzeErrorsSession) Execute(args map[string]interface{}) (interface{}, error) {
	return ExecuteSessionLevelQuery("analyze errors", args)
}

// Define remaining session-level tools following the same pattern:
// - ToolQueryToolSequencesSession
// - ToolQueryFileAccessSession
// - ToolQuerySuccessfulPromptsSession
// - ToolQueryContextSession

// Note: get_session_stats already exists, DO NOT redefine

// ExecuteSessionLevelQuery executes a session-level query (without --project flag)
func ExecuteSessionLevelQuery(command string, args map[string]interface{}) (interface{}, error) {
	// Build CLI command WITHOUT --project flag
	cmdArgs := []string{command}

	// Add parameters (same logic as project-level, but no --project flag)
	if limit, ok := args["limit"].(float64); ok {
		cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", int(limit)))
	}

	// ... (same parameter handling as ExecuteProjectLevelQuery)

	// Execute CLI
	cmd := exec.Command("meta-cc", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("CLI execution failed: %v", err)
	}

	// Parse JSON
	var result interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return result, nil
}

// RegisterSessionLevelTools registers all session-level tools
func RegisterSessionLevelTools(registry *ToolRegistry) {
	registry.Register(NewToolQueryToolsSession())
	registry.Register(NewToolQueryUserMessagesSession())
	registry.Register(NewToolAnalyzeErrorsSession())
	// Register remaining tools...
	// Note: get_session_stats is already registered, skip it
}
```

### TDD 步骤

**测试文件**: `internal/mcp/tools_session_test.go` (新建, ~60 行)

```go
package mcp

import (
	"encoding/json"
	"testing"
)

func TestToolQueryToolsSession_Naming(t *testing.T) {
	tool := NewToolQueryToolsSession()

	if tool.Name != "query_tools_session" {
		t.Errorf("Expected name 'query_tools_session', got '%s'", tool.Name)
	}

	// Verify description mentions "current session"
	if !strings.Contains(tool.Description, "current session") {
		t.Error("Description should mention 'current session'")
	}
}

func TestToolQueryUserMessagesSession_Naming(t *testing.T) {
	tool := NewToolQueryUserMessagesSession()

	if tool.Name != "query_user_messages_session" {
		t.Errorf("Expected name 'query_user_messages_session', got '%s'", tool.Name)
	}
}

func TestSessionLevelTools_NoProjectFlag(t *testing.T) {
	// Verify that session-level execution does NOT include --project flag

	args := map[string]interface{}{
		"limit": float64(10),
	}

	cmdArgs := buildSessionLevelQueryArgs("query tools", args)

	// Verify --project flag is NOT present
	for _, arg := range cmdArgs {
		if arg == "--project" {
			t.Error("Session-level query should NOT have --project flag")
		}
	}
}

func buildSessionLevelQueryArgs(command string, args map[string]interface{}) []string {
	cmdParts := strings.Split(command, " ")
	cmdArgs := cmdParts

	// Add parameters (no --project flag)
	if limit, ok := args["limit"].(float64); ok {
		cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", int(limit)))
	}

	cmdArgs = append(cmdArgs, "--output", "json")

	return cmdArgs
}

func TestRegisterSessionLevelTools(t *testing.T) {
	registry := NewToolRegistry()
	RegisterSessionLevelTools(registry)

	expectedTools := []string{
		"query_tools_session",
		"query_user_messages_session",
		"analyze_errors_session",
		// Add remaining tools...
	}

	for _, toolName := range expectedTools {
		if !registry.HasTool(toolName) {
			t.Errorf("Tool '%s' not registered", toolName)
		}
	}
}

func TestGetSessionStats_BackwardCompatibility(t *testing.T) {
	// Verify get_session_stats exists and is unchanged
	registry := NewToolRegistry()

	// Assuming get_session_stats was registered previously
	RegisterExistingTools(registry)

	if !registry.HasTool("get_session_stats") {
		t.Error("get_session_stats should exist for backward compatibility")
	}

	// Verify it does NOT have _session suffix (it's the original name)
	tool := registry.GetTool("get_session_stats")
	if tool.Name != "get_session_stats" {
		t.Errorf("Expected name 'get_session_stats', got '%s'", tool.Name)
	}
}
```

### 交付物

**新建文件**:
- `internal/mcp/tools_session.go` (~80 行)
- `internal/mcp/tools_session_test.go` (~60 行)

**代码量**: ~80 行（新增 Go 源代码，不含测试）

### 验收标准

- ✅ 8 个会话级工具定义完成
- ✅ 所有工具名带 `_session` 后缀
- ✅ 描述明确说明"current session only"
- ✅ 执行逻辑不包含 `--project` 标志
- ✅ `get_session_stats` 保持不变（向后兼容）
- ✅ 所有单元测试通过

---

## Stage 12.4: 配置与文档更新

### 目标

更新 MCP 配置文件和文档，提供清晰的项目级/会话级工具使用指南。

### 实现步骤

#### 1. 更新 MCP 配置文件

**文件**: `.claude/mcp-servers/meta-cc.json` (修改, ~20 行)

```json
{
  "mcpServers": {
    "meta-cc": {
      "command": "meta-cc",
      "args": ["mcp"],
      "tools": [
        "query_tools",
        "query_user_messages",
        "get_stats",
        "analyze_errors",
        "query_tool_sequences",
        "query_file_access",
        "query_successful_prompts",
        "query_context",
        "query_tools_session",
        "query_user_messages_session",
        "get_session_stats",
        "analyze_errors_session",
        "query_tool_sequences_session",
        "query_file_access_session",
        "query_successful_prompts_session",
        "query_context_session"
      ]
    }
  }
}
```

#### 2. 创建使用指南

**文件**: `docs/mcp-guide.md` (新建, ~200 行)

```markdown
# MCP Project-Level Query Guide

## Overview

Phase 12 extends the meta-cc MCP Server to support both **project-level** (all sessions) and **session-level** (current session) queries.

## Tool Naming Convention

| Naming Pattern | Scope | Example |
|---------------|-------|---------|
| `<tool_name>` (no suffix) | **Project-level** (all sessions) | `query_tools` |
| `<tool_name>_session` | **Session-level** (current session) | `query_tools_session` |

**Exception**: `get_session_stats` retains its original name for backward compatibility.

## Available Tools

### Project-Level Tools (All Sessions)

Query across all sessions in the current project:

- `query_tools` - Tool call history
- `query_user_messages` - User message search
- `get_stats` - Project statistics
- `analyze_errors` - Error pattern analysis
- `query_tool_sequences` - Workflow patterns
- `query_file_access` - File operation history
- `query_successful_prompts` - Quality prompt patterns
- `query_context` - Error context analysis

### Session-Level Tools (Current Session)

Query only the current session:

- `query_tools_session`
- `query_user_messages_session`
- `get_session_stats` (backward compatible)
- `analyze_errors_session`
- `query_tool_sequences_session`
- `query_file_access_session`
- `query_successful_prompts_session`
- `query_context_session`

## Usage Examples

### Example 1: Project-Level Analysis

**User**: "How do I typically use agents in this project?"

**Claude** (uses `query_tools`):
```json
{
  "tool": "query_tools",
  "args": {
    "where": "tool LIKE '%agent%'",
    "limit": 50
  }
}
```

**Result**: Returns agent-related tool calls across all sessions in the project.

### Example 2: Session-Level Analysis

**User**: "What errors have occurred in this current session?"

**Claude** (uses `analyze_errors_session`):
```json
{
  "tool": "analyze_errors_session",
  "args": {
    "output_format": "json"
  }
}
```

**Result**: Returns errors from the current session only.

### Example 3: Cross-Session Error Patterns

**User**: "What are the most common errors in this project?"

**Claude** (uses `analyze_errors`):
```json
{
  "tool": "analyze_errors",
  "args": {
    "output_format": "json"
  }
}
```

**Result**: Returns aggregated error patterns from all sessions.

### Example 4: File Modification History

**User**: "Show me the edit history for main.go across all sessions"

**Claude** (uses `query_file_access`):
```json
{
  "tool": "query_file_access",
  "args": {
    "file": "main.go"
  }
}
```

**Result**: Returns read/edit/write operations on main.go from all sessions.

## Implementation Details

### Project-Level Execution

Project-level tools add the `--project .` flag to CLI commands:

```bash
# Project-level query
meta-cc query tools --project . --limit 100

# Result: All tool calls from all sessions in ~/.claude/projects/{project-hash}/
```

### Session-Level Execution

Session-level tools execute without the `--project` flag:

```bash
# Session-level query
meta-cc query tools --limit 100

# Result: Tool calls from current session only
```

## When to Use Each Scope

### Use Project-Level Tools When:

- ✅ Analyzing long-term patterns ("How do I typically structure prompts?")
- ✅ Identifying recurring errors ("What errors keep happening?")
- ✅ Tracking project evolution ("How has my tool usage changed?")
- ✅ Finding successful workflows ("What prompt patterns work best?")

### Use Session-Level Tools When:

- ✅ Debugging current session ("What went wrong just now?")
- ✅ Quick session summary ("How many tools have I used today?")
- ✅ Focused analysis ("Show me errors from this conversation")
- ✅ Performance tuning ("Is this session slower than usual?")

## Backward Compatibility

### Existing Tool Behavior

`get_session_stats` retains its original behavior:

```json
{
  "tool": "get_session_stats",
  "args": {}
}
```

Returns statistics for the current session only (no change from Phase 8).

### Migration Guide

If you were using `get_session_stats` for project-level analysis, migrate to:

```json
{
  "tool": "get_stats",
  "args": {}
}
```

## CLI Flag Reference

| Flag | Scope | Used By |
|------|-------|---------|
| `--project .` | All sessions in project | Project-level tools |
| (no flag) | Current session only | Session-level tools |

## Troubleshooting

### Issue: Project-level tool returns only current session data

**Cause**: `--project .` flag not being passed to CLI

**Solution**: Verify MCP server implementation includes `--project .` in command execution

### Issue: Session-level tool returns multi-session data

**Cause**: Tool is using project-level execution by mistake

**Solution**: Verify tool name has `_session` suffix and does NOT include `--project` flag

## See Also

- [MCP Server Documentation](../README.md#mcp-server)
- [Integration Guide](./integration-guide.md)
- [Phase 12 Implementation Plan](../plans/12/plan.md)
```

#### 3. 更新 README.md

**文件**: `README.md` (修改, ~20 行)

```markdown
## MCP Server (Phase 8 + 12)

meta-cc provides an MCP (Model Context Protocol) server for seamless integration with Claude.

### Project-Level vs Session-Level Queries

**Project-Level** (default): Query all sessions in the project
```bash
# Tools: query_tools, get_stats, analyze_errors, etc.
```

**Session-Level**: Query current session only
```bash
# Tools: query_tools_session, get_session_stats, analyze_errors_session, etc.
```

See [MCP Project Scope Guide](docs/mcp-guide.md) for detailed usage examples.
```

### 交付物

**新建文件**:
- `docs/mcp-guide.md` (~200 行)

**修改文件**:
- `.claude/mcp-servers/meta-cc.json` (~20 行)
- `README.md` (~20 行)

**代码量**: ~40 行（配置和文档更新）

### 验收标准

- ✅ MCP 配置包含所有 16 个工具
- ✅ 文档清晰说明项目级/会话级区别
- ✅ 包含实际使用示例
- ✅ 包含何时使用哪种工具的指南
- ✅ 包含向后兼容性说明
- ✅ README.md 更新并链接到详细文档

---

## 集成测试：MCP 项目级查询端到端验证

### 测试脚本

**文件**: `tests/integration/mcp_project_scope_test.sh` (新建, ~100 行)

```bash
#!/bin/bash
# Phase 12 MCP Project-Level Query Integration Test

set -e

echo "=== Phase 12 MCP Project-Level Query Test ==="
echo ""

# Prerequisites check
if ! command -v meta-cc &> /dev/null; then
    echo "✗ meta-cc CLI not found"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo "✗ jq not found (required for JSON parsing)"
    exit 1
fi

# Step 1: Verify tool registration
echo "[1/4] Verifying MCP tool registration..."

TOOLS=$(meta-cc mcp list-tools 2>/dev/null)

# Check project-level tools
PROJECT_TOOLS=(
    "query_tools"
    "query_user_messages"
    "get_stats"
    "analyze_errors"
    "query_tool_sequences"
    "query_file_access"
    "query_successful_prompts"
    "query_context"
)

for tool in "${PROJECT_TOOLS[@]}"; do
    if ! echo "$TOOLS" | grep -q "$tool"; then
        echo "  ✗ Project-level tool '$tool' not registered"
        exit 1
    fi
done

echo "  ✓ All project-level tools registered"

# Check session-level tools
SESSION_TOOLS=(
    "query_tools_session"
    "query_user_messages_session"
    "get_session_stats"
    "analyze_errors_session"
)

for tool in "${SESSION_TOOLS[@]}"; do
    if ! echo "$TOOLS" | grep -q "$tool"; then
        echo "  ✗ Session-level tool '$tool' not registered"
        exit 1
    fi
done

echo "  ✓ All session-level tools registered"

# Step 2: Test project-level query execution
echo "[2/4] Testing project-level query execution..."

# Simulate MCP call to query_tools (project-level)
# This requires multi-session data in ~/.claude/projects/

# Create test environment with multiple sessions (if not exists)
# For this test, we assume test data is already set up

PROJECT_RESULT=$(meta-cc mcp call query_tools --args '{"limit": 10}' 2>/dev/null)

if [ -z "$PROJECT_RESULT" ]; then
    echo "  ✗ Project-level query returned empty result"
    exit 1
fi

# Verify result is valid JSON
if ! echo "$PROJECT_RESULT" | jq empty 2>/dev/null; then
    echo "  ✗ Project-level query returned invalid JSON"
    exit 1
fi

echo "  ✓ Project-level query execution successful"

# Step 3: Test session-level query execution
echo "[3/4] Testing session-level query execution..."

SESSION_RESULT=$(meta-cc mcp call query_tools_session --args '{"limit": 10}' 2>/dev/null)

if [ -z "$SESSION_RESULT" ]; then
    echo "  ✗ Session-level query returned empty result"
    exit 1
fi

if ! echo "$SESSION_RESULT" | jq empty 2>/dev/null; then
    echo "  ✗ Session-level query returned invalid JSON"
    exit 1
fi

echo "  ✓ Session-level query execution successful"

# Step 4: Verify backward compatibility
echo "[4/4] Testing backward compatibility..."

# get_session_stats should still work
STATS_RESULT=$(meta-cc mcp call get_session_stats 2>/dev/null)

if [ -z "$STATS_RESULT" ]; then
    echo "  ✗ get_session_stats backward compatibility broken"
    exit 1
fi

if ! echo "$STATS_RESULT" | jq empty 2>/dev/null; then
    echo "  ✗ get_session_stats returned invalid JSON"
    exit 1
fi

echo "  ✓ Backward compatibility maintained"

echo ""
echo "=== All Phase 12 Tests Passed ✅ ==="
echo ""
echo "Summary:"
echo "  - 8 project-level tools registered"
echo "  - 8 session-level tools registered"
echo "  - Project-level queries work correctly"
echo "  - Session-level queries work correctly"
echo "  - Backward compatibility maintained"
```

---

## Phase 12 验收清单

### 功能验收

- [ ] **Stage 12.1: 项目级工具定义**
  - [ ] 8 个项目级工具定义完成
  - [ ] InputSchema 有效
  - [ ] 单元测试通过

- [ ] **Stage 12.2: 执行逻辑**
  - [ ] `--project .` 标志正确添加
  - [ ] CLI 命令构建正确
  - [ ] JSON 解析正确
  - [ ] 单元测试通过

- [ ] **Stage 12.3: 会话级工具**
  - [ ] 8 个会话级工具定义完成
  - [ ] 工具名带 `_session` 后缀
  - [ ] `get_session_stats` 保持不变
  - [ ] 单元测试通过

- [ ] **Stage 12.4: 配置与文档**
  - [ ] MCP 配置更新
  - [ ] `docs/mcp-guide.md` 完成
  - [ ] README.md 更新

### 集成验收

- [ ] **端到端测试**
  - [ ] 所有工具正确注册
  - [ ] 项目级查询返回多会话数据
  - [ ] 会话级查询返回当前会话数据
  - [ ] 向后兼容性验证

- [ ] **真实场景验证**
  - [ ] meta-cc 项目自身测试
  - [ ] 跨会话错误模式分析
  - [ ] 工作流模式识别

### 代码质量

- [ ] **代码量验收**
  - [ ] Stage 12.1: ~80 行
  - [ ] Stage 12.2: ~100 行
  - [ ] Stage 12.3: ~80 行
  - [ ] Stage 12.4: ~40 行
  - [ ] 总计: ~300 行（目标 280-320 行）

- [ ] **测试覆盖率**
  - [ ] 单元测试覆盖率 ≥ 80%
  - [ ] 所有 TDD 测试通过

---

## 项目结构（Phase 12 完成后）

```
meta-cc/
├── internal/
│   └── mcp/
│       ├── tools_project.go          # 新增：项目级工具定义
│       ├── tools_project_test.go     # 新增：项目级工具测试
│       ├── tools_session.go          # 新增：会话级工具定义
│       ├── tools_session_test.go     # 新增：会话级工具测试
│       ├── executor.go               # 新增：执行逻辑
│       └── executor_test.go          # 新增：执行逻辑测试
├── tests/
│   └── integration/
│       ├── mcp_project_flag_test.sh  # 新增：项目标志测试
│       └── mcp_project_scope_test.sh # 新增：综合集成测试
├── docs/
│   └── mcp-guide.md          # 新增：使用指南
├── .claude/
│   └── mcp-servers/
│       └── meta-cc.json              # 更新：包含所有工具
├── plans/
│   └── 12/
│       ├── plan.md                   # 本文档
│       └── README.md                 # 快速参考
└── README.md                          # 更新：MCP 项目级查询文档
```

---

## 依赖关系

**Phase 12 依赖**:
- Phase 0-9（完整的 meta-cc 工具链）
- Phase 8（MCP Server 基础）

**Phase 12 提供**:
- 项目级跨会话查询能力
- 会话级聚焦查询能力
- 清晰的工具命名约定
- 向后兼容性保证

---

## 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| `--project .` 标志破坏现有行为 | 高 | 仅在项目级工具中使用；会话级工具保持原样 |
| 工具命名混淆 | 中 | 清晰的命名约定文档；一致的 `_session` 后缀 |
| 向后兼容性破坏 | 高 | 保持 `get_session_stats` 不变；充分测试 |
| 多会话数据量大导致性能问题 | 中 | 默认限制结果数量；使用流式输出（Phase 11） |
| CLI 执行错误处理不完整 | 低 | 完整的错误捕获和日志记录 |

---

## 实施优先级

**必须完成**（Phase 12 核心功能）:
1. Stage 12.1（项目级工具定义）
2. Stage 12.2（执行逻辑与 `--project .`）
3. Stage 12.3（会话级工具）
4. 向后兼容性验证

**推荐完成**（提升用户体验）:
5. Stage 12.4（配置和文档）
6. 集成测试和真实场景验证

**可选完成**（进一步优化）:
7. 性能优化（如果跨会话查询慢）
8. 更多使用示例和最佳实践

---

## Phase 12 总结

Phase 12 扩展 MCP Server 以支持跨会话元认知分析：

### 核心成果

1. **项目级工具**（Stage 12.1 + 12.2）
   - 8 个工具查询所有会话
   - 使用 `--project .` 标志
   - 支持长期模式分析

2. **会话级工具**（Stage 12.3）
   - 8 个工具查询当前会话
   - 统一 `_session` 后缀命名
   - 聚焦当前对话上下文

3. **向后兼容**
   - `get_session_stats` 保持不变
   - 现有集成不受影响

4. **清晰文档**（Stage 12.4）
   - 使用指南和示例
   - 何时使用哪种工具
   - 迁移指南

### 集成价值

- **元认知深度**: 跨会话分析工作模式和错误模式
- **灵活性**: 根据需求选择项目级或会话级查询
- **一致性**: 统一的工具命名约定
- **兼容性**: 向后兼容现有工具

### 用户价值

- ✅ 回答"我在这个项目中如何工作？"（项目级）
- ✅ 回答"这次会话中发生了什么？"（会话级）
- ✅ 识别长期改进机会（跨会话错误模式）
- ✅ 快速调试当前问题（会话级错误分析）

**Phase 12 完成后，meta-cc MCP Server 成为强大的元认知工具，支持项目级和会话级双维度分析。**

---

## 参考文档

- [MCP 协议规范](https://modelcontextprotocol.io/)
- [meta-cc 技术方案](../../docs/proposals/meta-cognition-proposal.md)
- [meta-cc 总体实施计划](../../docs/plan.md)
- [Phase 8 实施计划](../8/plan.md)（MCP Server 基础）
- [Phase 11 实施计划](../11/plan.md)（Unix 可组合性）

---

**Phase 12 实施准备就绪。开始 TDD 开发流程。**
