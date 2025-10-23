# Phase 24 实施计划：统一查询接口设计与实现

## 项目信息

- **Phase**: 24
- **目标**: 基于实际 Claude Code JSONL schema，设计并实现统一的查询接口，将 16 个碎片化 MCP 工具简化为 1 个可组合的查询工具
- **预计工期**: 2-3 周
- **代码量预算**: ≤800 行（5 个 Stage × 160 行/Stage）
- **测试覆盖目标**: ≥80%

---

## 一、总体目标与设计理念

### 核心目标

从当前的 **16 个独立 MCP 工具** → **1 个统一 `query` 工具**

### 设计理念

1. **资源导向设计**：查询"什么资源"（entries/messages/tools），而非"怎么做"（query_tools/query_sequences）
2. **可组合查询管道**：filter → transform → aggregate → output
3. **Schema 统一为 snake_case**：与 JSONL 源文件保持一致
4. **向后兼容迁移**：保留旧工具作为别名，2-3 版本兼容期

### 关键设计决策

**资源类型层次**：
```
entries (原始 SessionEntry 流)
  ↓ transform
messages (user/assistant 消息视图)
  ↓ transform
tools (tool_use + tool_result 配对视图)
```

**查询管道**：
```
SessionEntry[]
  → filter(过滤条件)
  → transform(转换/分组/提取)
  → aggregate(聚合函数)
  → Result
```

---

## 二、Stage 划分与依赖关系

### Stage 依赖图

```
Stage 24.1 (Schema标准化)
    ↓
Stage 24.2 (统一查询接口实现) ← 关键路径
    ↓
Stage 24.3 (MCP工具重构)
    ↓
Stage 24.4 (测试与验证)
    ↓
Stage 24.5 (文档与迁移)
```

**并行机会**：
- Stage 24.4 的单元测试可以在 24.2 完成后立即开始
- Stage 24.5 的文档编写可以与 24.3 并行

**关键路径**：Stage 24.1 → 24.2 → 24.3 → 24.4

---

## 三、Stage 24.1：Schema 标准化

### 目标

统一所有数据结构为 snake_case，与 JSONL 源文件保持一致。

### 当前问题

**混乱的命名风格**：
- `internal/parser/tools.go`: `ToolCall` 使用 PascalCase（`ToolName`, `UUID`, `Timestamp`）
- `internal/query/types.go`: 查询结果混用 snake_case 和 PascalCase
- JSONL 源文件: 全部 snake_case（`session_id`, `parent_uuid`, `git_branch`）

### 详细步骤

#### Step 1.1: 定义标准 Schema（2小时）

**文件**: `internal/query/unified_schema.go`（新建）

```go
package query

// SessionEntry - 标准化 JSONL 条目（snake_case）
type SessionEntry struct {
    Type       string   `json:"type"`
    UUID       string   `json:"uuid"`
    Timestamp  string   `json:"timestamp"`
    SessionID  string   `json:"session_id"`   // 统一为 snake_case
    ParentUUID string   `json:"parent_uuid"`
    CWD        string   `json:"cwd"`
    GitBranch  string   `json:"git_branch"`   // 统一为 snake_case
    Version    string   `json:"version"`
    Message    *Message `json:"message,omitempty"`
}

// Message - 消息内容
type Message struct {
    Role       string         `json:"role"`
    Content    []ContentBlock `json:"content"`
    ID         string         `json:"id,omitempty"`
    Model      string         `json:"model,omitempty"`
    StopReason string         `json:"stop_reason,omitempty"`
    Usage      *TokenUsage    `json:"usage,omitempty"`
}

// ToolExecution - 工具执行结果（统一 snake_case）
type ToolExecution struct {
    ToolUseID     string                 `json:"tool_use_id"`
    SessionID     string                 `json:"session_id"`
    Timestamp     string                 `json:"timestamp"`
    ToolName      string                 `json:"tool_name"`
    Input         map[string]interface{} `json:"input"`
    Output        string                 `json:"output"`
    Status        string                 `json:"status"`
    Error         string                 `json:"error,omitempty"`
    AssistantUUID string                 `json:"assistant_uuid"`
    UserUUID      string                 `json:"user_uuid"`
}

// MessageView - 消息视图（扁平化）
type MessageView struct {
    UUID          string         `json:"uuid"`
    SessionID     string         `json:"session_id"`
    ParentUUID    string         `json:"parent_uuid"`
    Timestamp     string         `json:"timestamp"`
    Role          string         `json:"role"`
    Content       string         `json:"content,omitempty"`
    ContentBlocks []ContentBlock `json:"content_blocks"`
}
```

**测试用例**（`unified_schema_test.go`）：
```go
func TestSchemaJSONSerialization(t *testing.T) {
    // 验证 snake_case JSON 序列化
    exec := ToolExecution{
        ToolUseID:  "test-123",
        SessionID:  "session-456",
        ToolName:   "Read",
        Timestamp:  "2025-10-23T00:00:00Z",
    }

    data, err := json.Marshal(exec)
    require.NoError(t, err)

    // 验证字段名为 snake_case
    assert.Contains(t, string(data), "tool_use_id")
    assert.Contains(t, string(data), "session_id")
    assert.NotContains(t, string(data), "ToolUseID")
}
```

**验收标准**：
- ✅ 所有字段使用 snake_case JSON tag
- ✅ 与 JSONL 源文件 schema 100% 匹配
- ✅ 测试覆盖率 ≥80%

**代码量估算**: 150 行代码 + 50 行测试

---

#### Step 1.2: 迁移现有代码（3小时）

**影响文件**：
1. `internal/parser/tools.go` - 更新 ToolCall 为 snake_case
2. `internal/query/*.go` - 更新所有查询函数返回类型
3. `cmd/mcp-server/executor.go` - 更新输出格式

**迁移策略**：
```go
// 方案 A: 类型别名（向后兼容）
type ToolCall = ToolExecution  // 别名保留旧名称

// 方案 B: 转换函数
func ConvertToolCallToExecution(old ToolCall) ToolExecution {
    return ToolExecution{
        ToolUseID: old.UUID,
        ToolName:  old.ToolName,
        // ...
    }
}
```

**验收标准**：
- ✅ 所有 `make test` 通过
- ✅ MCP 工具输出为 snake_case
- ✅ 旧工具仍然可用（通过别名）

**代码量估算**: 100 行代码修改

---

## 四、Stage 24.2：统一查询接口实现

### 目标

实现核心 `Query()` 函数和过滤/转换/聚合引擎。

### 详细步骤

#### Step 2.1: 查询参数设计（2小时）

**文件**: `internal/query/unified_query.go`（新建）

```go
package query

// QueryParams - 统一查询参数
type QueryParams struct {
    // Tier 1: Resource Selection
    Resource string // "entries" | "messages" | "tools"

    // Tier 2: Scope
    Scope string // "session" | "project"

    // Tier 3: Filtering
    Filter FilterSpec

    // Tier 4: Transformation
    Transform TransformSpec

    // Tier 5: Aggregation
    Aggregate AggregateSpec

    // Tier 6: Output Control
    Output OutputSpec

    // Advanced: jq filter
    JQFilter string
}

// FilterSpec - 结构化过滤条件
type FilterSpec struct {
    // Entry-level
    Type       string
    SessionID  string
    UUID       string
    GitBranch  string
    TimeRange  *TimeRange

    // Message-level
    Role         string
    ContentType  string
    ContentMatch string

    // Tool-level
    ToolName   string
    ToolStatus string
    HasError   bool
}

// TransformSpec - 转换/分组/提取
type TransformSpec struct {
    Extract []string // JSONPath 表达式
    GroupBy string   // 分组字段
    Join    *JoinSpec
}

// AggregateSpec - 聚合函数
type AggregateSpec struct {
    Function string // "count" | "sum" | "avg" | "min" | "max" | "group"
    Field    string // 聚合字段
}

// OutputSpec - 输出控制
type OutputSpec struct {
    Format    string // "jsonl" | "tsv" | "summary"
    Limit     int
    SortBy    string
    SortOrder string // "asc" | "desc"
}
```

**测试用例**：
```go
func TestQueryParamsValidation(t *testing.T) {
    tests := []struct {
        name    string
        params  QueryParams
        wantErr bool
    }{
        {
            name: "valid_basic_query",
            params: QueryParams{
                Resource: "tools",
                Scope:    "project",
                Filter: FilterSpec{
                    ToolName: "Read",
                },
            },
            wantErr: false,
        },
        {
            name: "invalid_resource",
            params: QueryParams{
                Resource: "invalid",
            },
            wantErr: true,
        },
    }
    // ...
}
```

**验收标准**：
- ✅ 参数结构完整定义
- ✅ 参数验证逻辑实现
- ✅ 测试覆盖率 ≥80%

**代码量估算**: 120 行代码 + 80 行测试

---

#### Step 2.2: 资源选择器实现（3小时）

**文件**: 继续 `internal/query/unified_query.go`

```go
// Query - 统一查询入口
func Query(entries []SessionEntry, params QueryParams) ([]interface{}, error) {
    // 1. Validate params
    if err := validateParams(params); err != nil {
        return nil, err
    }

    // 2. Select resource view
    resources, err := selectResource(entries, params.Resource)
    if err != nil {
        return nil, err
    }

    // 3. Apply filters
    filtered := applyFilter(resources, params.Filter)

    // 4. Apply transformations
    transformed := applyTransform(filtered, params.Transform)

    // 5. Apply aggregations
    aggregated := applyAggregate(transformed, params.Aggregate)

    // 6. Format output
    return formatOutput(aggregated, params.Output), nil
}

// selectResource 选择资源视图
func selectResource(entries []SessionEntry, resource string) ([]interface{}, error) {
    switch resource {
    case "entries":
        return entriesToInterface(entries), nil
    case "messages":
        return extractMessages(entries), nil
    case "tools":
        return extractToolExecutions(entries), nil
    default:
        return nil, fmt.Errorf("unknown resource type: %s", resource)
    }
}

// extractMessages 提取消息视图
func extractMessages(entries []SessionEntry) []interface{} {
    var messages []interface{}
    for _, entry := range entries {
        if entry.Message == nil {
            continue
        }
        msg := MessageView{
            UUID:          entry.UUID,
            SessionID:     entry.SessionID,
            ParentUUID:    entry.ParentUUID,
            Timestamp:     entry.Timestamp,
            Role:          entry.Message.Role,
            ContentBlocks: entry.Message.Content,
        }
        messages = append(messages, msg)
    }
    return messages
}

// extractToolExecutions 提取工具执行视图
func extractToolExecutions(entries []SessionEntry) []interface{} {
    // 复用现有的 ToolCall 提取逻辑，转换为 ToolExecution
    toolCalls := ExtractToolCalls(entries)
    var executions []interface{}
    for _, tc := range toolCalls {
        exec := ToolExecution{
            ToolUseID:  tc.UUID, // 需要实际的 tool_use_id
            ToolName:   tc.ToolName,
            Input:      tc.Input,
            Output:     tc.Output,
            Status:     tc.Status,
            Error:      tc.Error,
            Timestamp:  tc.Timestamp,
        }
        executions = append(executions, exec)
    }
    return executions
}
```

**测试用例**：
```go
func TestSelectResource(t *testing.T) {
    fixture := loadTestFixture(t, "sample_session.jsonl")

    tests := []struct {
        name         string
        resource     string
        expectType   string
        expectCount  int
    }{
        {
            name:        "select_entries",
            resource:    "entries",
            expectType:  "SessionEntry",
            expectCount: 10,
        },
        {
            name:        "select_messages",
            resource:    "messages",
            expectType:  "MessageView",
            expectCount: 6,
        },
        {
            name:        "select_tools",
            resource:    "tools",
            expectType:  "ToolExecution",
            expectCount: 4,
        },
    }
    // ...
}
```

**验收标准**：
- ✅ 3 种资源类型正确提取
- ✅ 与现有提取逻辑输出一致
- ✅ 测试覆盖率 ≥80%

**代码量估算**: 150 行代码 + 100 行测试

---

#### Step 2.3: 过滤引擎实现（3小时）

**文件**: 继续 `internal/query/unified_query.go`

```go
// applyFilter 应用结构化过滤器
func applyFilter(resources []interface{}, filter FilterSpec) []interface{} {
    if isEmptyFilter(filter) {
        return resources
    }

    var result []interface{}
    for _, resource := range resources {
        if matchesFilter(resource, filter) {
            result = append(result, resource)
        }
    }
    return result
}

// matchesFilter 检查资源是否匹配过滤条件
func matchesFilter(resource interface{}, filter FilterSpec) bool {
    // Entry-level filters
    if filter.Type != "" {
        if entry, ok := resource.(SessionEntry); ok {
            if entry.Type != filter.Type {
                return false
            }
        }
    }

    if filter.SessionID != "" {
        if hasSessionID, ok := resource.(interface{ GetSessionID() string }); ok {
            if hasSessionID.GetSessionID() != filter.SessionID {
                return false
            }
        }
    }

    // Tool-level filters
    if filter.ToolName != "" {
        if exec, ok := resource.(ToolExecution); ok {
            if exec.ToolName != filter.ToolName {
                return false
            }
        }
    }

    if filter.ToolStatus != "" {
        if exec, ok := resource.(ToolExecution); ok {
            if exec.Status != filter.ToolStatus {
                return false
            }
        }
    }

    // Message-level filters
    if filter.Role != "" {
        if msg, ok := resource.(MessageView); ok {
            if msg.Role != filter.Role {
                return false
            }
        }
    }

    return true
}
```

**测试用例**：
```go
func TestApplyFilter(t *testing.T) {
    resources := []interface{}{
        ToolExecution{ToolName: "Read", Status: "success"},
        ToolExecution{ToolName: "Read", Status: "error"},
        ToolExecution{ToolName: "Edit", Status: "success"},
    }

    tests := []struct {
        name   string
        filter FilterSpec
        want   int
    }{
        {
            name:   "filter_by_tool_name",
            filter: FilterSpec{ToolName: "Read"},
            want:   2,
        },
        {
            name:   "filter_by_status",
            filter: FilterSpec{ToolStatus: "error"},
            want:   1,
        },
        {
            name: "filter_by_tool_and_status",
            filter: FilterSpec{
                ToolName:   "Read",
                ToolStatus: "error",
            },
            want: 1,
        },
    }
    // ...
}
```

**验收标准**：
- ✅ 支持所有 FilterSpec 字段
- ✅ AND 逻辑正确组合
- ✅ 测试覆盖率 ≥80%

**代码量估算**: 100 行代码 + 80 行测试

---

#### Step 2.4: 聚合引擎实现（2小时）

**文件**: 继续 `internal/query/unified_query.go`

```go
// applyAggregate 应用聚合函数
func applyAggregate(resources []interface{}, agg AggregateSpec) []interface{} {
    if agg.Function == "" {
        return resources
    }

    switch agg.Function {
    case "count":
        return aggregateCount(resources, agg.Field)
    case "group":
        return aggregateGroup(resources, agg.Field)
    default:
        return resources
    }
}

// aggregateCount 计数聚合
func aggregateCount(resources []interface{}, field string) []interface{} {
    if field == "" {
        // 简单计数
        return []interface{}{
            map[string]interface{}{"count": len(resources)},
        }
    }

    // 按字段分组计数
    counts := make(map[string]int)
    for _, resource := range resources {
        value := extractField(resource, field)
        counts[value]++
    }

    var result []interface{}
    for value, count := range counts {
        result = append(result, map[string]interface{}{
            field:   value,
            "count": count,
        })
    }
    return result
}

// extractField 从资源提取字段值
func extractField(resource interface{}, field string) string {
    switch field {
    case "tool_name":
        if exec, ok := resource.(ToolExecution); ok {
            return exec.ToolName
        }
    case "status":
        if exec, ok := resource.(ToolExecution); ok {
            return exec.Status
        }
    case "role":
        if msg, ok := resource.(MessageView); ok {
            return msg.Role
        }
    }
    return ""
}
```

**测试用例**：
```go
func TestApplyAggregate(t *testing.T) {
    resources := []interface{}{
        ToolExecution{ToolName: "Read"},
        ToolExecution{ToolName: "Read"},
        ToolExecution{ToolName: "Edit"},
    }

    tests := []struct {
        name string
        agg  AggregateSpec
        want int
    }{
        {
            name: "count_total",
            agg:  AggregateSpec{Function: "count"},
            want: 1, // 单个结果：{"count": 3}
        },
        {
            name: "count_by_tool",
            agg:  AggregateSpec{Function: "count", Field: "tool_name"},
            want: 2, // Read: 2, Edit: 1
        },
    }
    // ...
}
```

**验收标准**：
- ✅ count 聚合正确实现
- ✅ group 聚合正确实现
- ✅ 测试覆盖率 ≥80%

**代码量估算**: 80 行代码 + 60 行测试

---

## 五、Stage 24.3：MCP 工具重构

### 目标

使用统一查询接口重构 16 个 MCP 工具。

### 详细步骤

#### Step 3.1: MCP 适配器实现（3小时）

**文件**: `cmd/mcp-server/unified_adapter.go`（新建）

```go
package main

import (
    "github.com/yaleh/meta-cc/internal/query"
)

// convertMCPParamsToQueryParams 转换 MCP 工具参数为统一查询参数
func convertMCPParamsToQueryParams(toolName string, args map[string]interface{}) query.QueryParams {
    params := query.QueryParams{
        Scope: getStringParam(args, "scope", "project"),
        Output: query.OutputSpec{
            Format: getStringParam(args, "output_format", "jsonl"),
            Limit:  getIntParam(args, "limit", 0),
        },
        JQFilter: getStringParam(args, "jq_filter", ".[]"),
    }

    switch toolName {
    case "query_tools":
        params.Resource = "tools"
        params.Filter = query.FilterSpec{
            ToolName:   getStringParam(args, "tool", ""),
            ToolStatus: getStringParam(args, "status", ""),
        }

    case "query_user_messages":
        params.Resource = "messages"
        params.Filter = query.FilterSpec{
            Role:         "user",
            ContentMatch: getStringParam(args, "pattern", ""),
        }

    case "query_assistant_messages":
        params.Resource = "messages"
        params.Filter = query.FilterSpec{
            Role: "assistant",
        }

    case "query_files":
        params.Resource = "tools"
        params.Filter = query.FilterSpec{
            ToolName: "Read|Edit|Write",
        }
        params.Aggregate = query.AggregateSpec{
            Function: "count",
            Field:    "file_path",
        }
    }

    return params
}

// executeUnifiedQuery 执行统一查询
func executeUnifiedQuery(cfg *config.Config, toolName string, args map[string]interface{}) (string, error) {
    // 1. 转换参数
    params := convertMCPParamsToQueryParams(toolName, args)

    // 2. 加载会话数据
    entries, err := loadSessionEntries(cfg, params.Scope)
    if err != nil {
        return "", err
    }

    // 3. 执行查询
    results, err := query.Query(entries, params)
    if err != nil {
        return "", err
    }

    // 4. 应用 jq 过滤器
    if params.JQFilter != "" && params.JQFilter != ".[]" {
        results, err = applyJQFilter(results, params.JQFilter)
        if err != nil {
            return "", err
        }
    }

    // 5. 格式化输出
    return formatOutput(results, params.Output)
}
```

**测试用例**：
```go
func TestConvertMCPParams(t *testing.T) {
    tests := []struct {
        name     string
        toolName string
        args     map[string]interface{}
        want     query.QueryParams
    }{
        {
            name:     "query_tools_with_filter",
            toolName: "query_tools",
            args: map[string]interface{}{
                "tool":   "Read",
                "status": "error",
            },
            want: query.QueryParams{
                Resource: "tools",
                Filter: query.FilterSpec{
                    ToolName:   "Read",
                    ToolStatus: "error",
                },
            },
        },
    }
    // ...
}
```

**验收标准**：
- ✅ 所有 16 个工具正确转换
- ✅ 参数映射无遗漏
- ✅ 测试覆盖率 ≥80%

**代码量估算**: 150 行代码 + 100 行测试

---

#### Step 3.2: 逐步迁移工具（4小时）

**策略**：一次迁移 2-3 个工具，验证后继续

**迁移顺序**（从简单到复杂）：
1. `get_session_stats` - 最简单，只需 count
2. `query_tools` - 基础过滤
3. `query_user_messages` - 模式匹配
4. `query_assistant_messages` - 复杂过滤
5. `query_conversation` - 关联查询
6. `query_files` - 聚合查询
7. 其他工具...

**迁移模板**：
```go
// 旧实现（保留作为备份）
func executeQueryToolsOld(cfg *config.Config, args map[string]interface{}) (string, error) {
    // ... 现有实现
}

// 新实现（使用统一接口）
func executeQueryTools(cfg *config.Config, args map[string]interface{}) (string, error) {
    return executeUnifiedQuery(cfg, "query_tools", args)
}
```

**验收标准**：
- ✅ 每个工具输出与旧版本一致
- ✅ 所有单元测试通过
- ✅ 集成测试通过

**代码量估算**: 200 行代码修改

---

## 六、Stage 24.4：测试与验证

### 目标

全面测试统一查询接口，确保功能正确性和性能。

### 详细步骤

#### Step 4.1: 单元测试（3小时）

**测试文件**: `internal/query/unified_query_test.go`

**测试维度**：
1. **参数验证测试**：无效参数拒绝
2. **资源选择测试**：entries/messages/tools 正确提取
3. **过滤逻辑测试**：各种过滤条件组合
4. **聚合功能测试**：count/group/sum 正确计算
5. **输出格式测试**：jsonl/tsv 格式正确

**测试用例示例**：
```go
func TestUnifiedQueryE2E(t *testing.T) {
    // 加载测试 fixture
    entries := loadTestFixture(t, "sample_session.jsonl")

    tests := []struct {
        name    string
        params  query.QueryParams
        want    int
        wantErr bool
    }{
        {
            name: "query_failed_reads",
            params: query.QueryParams{
                Resource: "tools",
                Filter: query.FilterSpec{
                    ToolName:   "Read",
                    ToolStatus: "error",
                },
            },
            want: 3, // 期望 3 个失败的 Read
        },
        {
            name: "count_by_tool",
            params: query.QueryParams{
                Resource: "tools",
                Aggregate: query.AggregateSpec{
                    Function: "count",
                    Field:    "tool_name",
                },
            },
            want: 5, // 期望 5 种工具
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            results, err := query.Query(entries, tt.params)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Len(t, results, tt.want)
        })
    }
}
```

**验收标准**：
- ✅ 测试覆盖率 ≥80%
- ✅ 所有边界条件测试
- ✅ 错误处理测试

**代码量估算**: 150 行测试代码

---

#### Step 4.2: 集成测试（2小时）

**测试文件**: `cmd/mcp-server/unified_integration_test.go`

**测试场景**：
1. **端到端测试**：MCP 请求 → 统一查询 → 格式化输出
2. **性能测试**：大数据集查询性能（10k+ entries）
3. **向后兼容测试**：旧工具调用结果一致性

**测试用例示例**：
```go
func TestMCPToolBackwardCompatibility(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    tools := []string{
        "query_tools",
        "query_user_messages",
        "query_assistant_messages",
    }

    for _, toolName := range tools {
        t.Run(toolName, func(t *testing.T) {
            // 执行新实现
            newResult, err := executeUnifiedQuery(cfg, toolName, testArgs)
            require.NoError(t, err)

            // 执行旧实现（如果保留）
            // oldResult, err := executeToolOld(cfg, toolName, testArgs)
            // require.NoError(t, err)

            // 验证结果结构一致
            var newData []map[string]interface{}
            err = json.Unmarshal([]byte(newResult), &newData)
            require.NoError(t, err)

            // 验证必需字段存在
            for _, item := range newData {
                assert.Contains(t, item, "timestamp")
            }
        })
    }
}
```

**验收标准**：
- ✅ 所有 MCP 工具集成测试通过
- ✅ 性能无回退（与旧实现相比）
- ✅ 真实项目验证通过

**代码量估算**: 100 行测试代码

---

#### Step 4.3: 性能基准测试（2小时）

**测试文件**: `internal/query/unified_query_bench_test.go`

```go
func BenchmarkUnifiedQueryTools(b *testing.B) {
    entries := loadLargeTestFixture(b, 10000)
    params := query.QueryParams{
        Resource: "tools",
        Filter: query.FilterSpec{
            ToolStatus: "error",
        },
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := query.Query(entries, params)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkUnifiedQueryAggregate(b *testing.B) {
    entries := loadLargeTestFixture(b, 10000)
    params := query.QueryParams{
        Resource: "tools",
        Aggregate: query.AggregateSpec{
            Function: "count",
            Field:    "tool_name",
        },
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := query.Query(entries, params)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

**性能目标**：
- ✅ 10k entries 查询 < 100ms
- ✅ 聚合查询 < 200ms
- ✅ 内存使用 < 50MB

**验收标准**：
- ✅ 基准测试通过
- ✅ 无性能回退

**代码量估算**: 50 行测试代码

---

## 七、Stage 24.5：文档与迁移

### 目标

编写完整文档，提供迁移指南和示例。

### 详细步骤

#### Step 5.1: API 文档（3小时）

**文件**: `docs/guides/unified-query-api.md`（新建）

**文档结构**：
```markdown
# 统一查询 API 指南

## 概述

统一查询接口将 16 个 MCP 工具简化为 1 个可组合的 `query` 工具。

## 快速开始

### 基础查询

查询所有失败的 Read 工具调用：
```javascript
query({
  resource: "tools",
  filter: {
    tool_name: "Read",
    tool_status: "error"
  }
})
```

### 聚合查询

统计每个工具的调用次数：
```javascript
query({
  resource: "tools",
  aggregate: {
    function: "count",
    field: "tool_name"
  }
})
```

## API 参考

### QueryParams

| 参数 | 类型 | 说明 | 默认值 |
|-----|------|-----|--------|
| resource | string | 资源类型 | "entries" |
| scope | string | 查询范围 | "project" |
| filter | FilterSpec | 过滤条件 | {} |
| aggregate | AggregateSpec | 聚合函数 | {} |
| output | OutputSpec | 输出控制 | {} |

### FilterSpec

| 字段 | 类型 | 说明 |
|-----|------|-----|
| tool_name | string | 工具名称 |
| tool_status | string | 执行状态 |
| role | string | 消息角色 |
| session_id | string | 会话 ID |

### 示例集合

#### 示例 1: 时间范围查询
...

#### 示例 2: 复杂组合查询
...
```

**验收标准**：
- ✅ 所有参数完整文档
- ✅ 10+ 示例覆盖常见场景
- ✅ 清晰的结构和导航

**代码量估算**: 0 行代码（纯文档）

---

#### Step 5.2: 迁移指南（2小时）

**文件**: `docs/guides/migration-to-unified-query.md`（新建）

**文档结构**：
```markdown
# 迁移到统一查询 API

## 迁移路径

### 当前工具 → 统一查询映射

| 当前工具 | 统一查询等价 |
|---------|------------|
| `query_tools` | `query({resource: "tools", filter: {...}})` |
| `query_user_messages` | `query({resource: "messages", filter: {role: "user"}})` |
| `query_files` | `query({resource: "tools", aggregate: {function: "count"}})` |

### 逐步迁移

#### 阶段 1: 学习新接口（第 1 周）
- 阅读 API 文档
- 尝试基础查询
- 对比旧工具输出

#### 阶段 2: 并行使用（第 2-4 周）
- 新查询使用 `query`
- 旧脚本保留旧工具
- 逐步迁移脚本

#### 阶段 3: 完全迁移（第 5-8 周）
- 所有查询使用 `query`
- 移除旧工具依赖

### 常见迁移问题

#### Q: 旧工具何时移除？
A: v3.0.0 版本将移除旧工具，预计 6 个月后。

#### Q: 如何处理复杂查询？
A: 使用 `jq_filter` 高级出口。

### 示例迁移

#### 迁移前
```javascript
query_tools({tool: "Read", status: "error"})
```

#### 迁移后
```javascript
query({
  resource: "tools",
  filter: {tool_name: "Read", tool_status: "error"}
})
```
```

**验收标准**：
- ✅ 清晰的迁移路径
- ✅ 所有工具迁移示例
- ✅ 常见问题解答

**代码量估算**: 0 行代码（纯文档）

---

#### Step 5.3: 更新现有文档（1小时）

**影响文件**：
1. `docs/guides/mcp.md` - 更新 MCP 工具列表
2. `README.md` - 更新功能描述
3. `CHANGELOG.md` - 添加 v2.0.0 变更记录

**CHANGELOG 示例**：
```markdown
## [2.0.0] - 2025-11-XX

### Added
- 统一查询 API：1 个 `query` 工具替代 16 个专用工具
- 资源导向查询：entries/messages/tools 三层视图
- 可组合查询管道：filter → transform → aggregate

### Changed
- MCP 工具输出统一为 snake_case
- 所有查询结果保持 JSONL schema 一致性

### Deprecated
- 16 个旧 MCP 工具标记为 deprecated（v3.0.0 移除）

### Migration
- 参见 [迁移指南](docs/guides/migration-to-unified-query.md)
```

**验收标准**：
- ✅ 所有文档更新完成
- ✅ CHANGELOG 准确记录

**代码量估算**: 0 行代码（纯文档）

---

## 八、风险分析与缓解措施

### 风险 1: 破坏性变更

**影响**: 现有用户的查询脚本失效

**缓解措施**：
1. ✅ **向后兼容层**：保留旧工具作为别名（`query_tools` → `query`）
2. ✅ **长兼容期**：2-3 个大版本（6-12 个月）
3. ✅ **迁移工具**：提供自动转换脚本
4. ✅ **详细文档**：迁移指南 + 示例库

**实施方案**：
```go
// 旧工具作为别名
func executeQueryTools(cfg *config.Config, args map[string]interface{}) (string, error) {
    // 转换为统一查询
    return executeUnifiedQuery(cfg, "query_tools", args)
}

// 添加 deprecation 警告（v2.1.0）
func executeQueryToolsWithWarning(cfg *config.Config, args map[string]interface{}) (string, error) {
    logDeprecationWarning("query_tools", "query")
    return executeQueryTools(cfg, args)
}
```

---

### 风险 2: 性能回退

**影响**: 统一接口可能不如专用工具优化

**缓解措施**：
1. ✅ **基准测试**：所有查询场景性能对比
2. ✅ **查询优化**：为常见模式优化代码路径
3. ✅ **缓存机制**：复用资源提取结果

**性能目标**：
- ≤10% 性能回退（可接受）
- 大多数场景性能持平或改善

---

### 风险 3: 学习曲线

**影响**: 用户需要学习新的查询语法

**缓解措施**：
1. ✅ **渐进式教程**：从简单到复杂
2. ✅ **交互式示例**：Cookbook 覆盖常见场景
3. ✅ **错误提示改进**：友好的参数验证错误

**学习路径设计**：
- Level 1: 基础过滤（学习成本 5 分钟）
- Level 2: 聚合查询（学习成本 10 分钟）
- Level 3: 复杂组合（学习成本 20 分钟）

---

### 风险 4: Schema 不一致

**影响**: 新旧输出格式不一致导致混乱

**缓解措施**：
1. ✅ **Schema 验证测试**：确保所有字段 snake_case
2. ✅ **输出对比测试**：新旧工具输出结构一致性
3. ✅ **文档明确说明**：Schema 变更清单

**Schema 变更检查清单**：
```go
func TestSchemaConsistency(t *testing.T) {
    // 验证所有输出字段为 snake_case
    result := executeQuery(...)
    for key := range result {
        assert.True(t, isSnakeCase(key), "field %s not snake_case", key)
    }
}
```

---

## 九、代码量汇总

### 按 Stage 统计

| Stage | 代码（行） | 测试（行） | 文档（行） | 总计 |
|-------|----------|----------|----------|------|
| 24.1 Schema标准化 | 150 | 50 | 0 | 200 |
| 24.2 统一接口实现 | 450 | 320 | 0 | 770 |
| 24.3 MCP工具重构 | 350 | 100 | 0 | 450 |
| 24.4 测试与验证 | 0 | 300 | 0 | 300 |
| 24.5 文档与迁移 | 0 | 0 | ~500 | 500 |
| **总计** | **950** | **770** | **500** | **2220** |

**说明**：
- 代码总量 **950 行**（超出预算 150 行，但在合理范围内）
- 测试代码 **770 行**（测试/代码比 81%，符合 ≥80% 目标）
- 文档 **500 行**（不计入代码预算）

**调整方案**：
- Stage 24.2 可拆分为 2 个子 Stage（2.1-2.2, 2.3-2.4）
- Stage 24.3 可拆分为 2 个子 Stage（3.1, 3.2）
- 这样可以保持每个 Stage ≤200 行代码的约束

---

## 十、时间估算

### 按 Stage 估算

| Stage | 估算时间 | 关键路径 |
|-------|---------|---------|
| 24.1 Schema标准化 | 5 小时 | ✅ |
| 24.2 统一接口实现 | 10 小时 | ✅ |
| 24.3 MCP工具重构 | 7 小时 | ✅ |
| 24.4 测试与验证 | 7 小时 | ✅ |
| 24.5 文档与迁移 | 6 小时 | ⬜ |
| **总计** | **35 小时** | - |

**工期估算**：
- **理想情况**（全职）：5 个工作日
- **实际情况**（兼职）：2-3 周
- **缓冲时间**：+20%（应对意外）

---

## 十一、验收标准总结

### 功能验收

- ✅ 统一 `query` 工具实现完成
- ✅ 3 种资源类型（entries/messages/tools）正确提取
- ✅ 过滤/聚合引擎功能完整
- ✅ 16 个 MCP 工具成功迁移
- ✅ 所有旧工具输出一致性验证通过

### 质量验收

- ✅ 单元测试覆盖率 ≥80%
- ✅ 所有 `make test` 通过
- ✅ 所有 `make lint` 通过
- ✅ 性能无显著回退（≤10%）
- ✅ 真实项目验证（meta-cc, NarrativeForge, claude-tmux）

### 文档验收

- ✅ API 参考文档完整
- ✅ 迁移指南清晰
- ✅ 10+ 示例覆盖常见场景
- ✅ CHANGELOG 准确记录

### 发布验收

- ✅ 向后兼容层实现（旧工具别名）
- ✅ Deprecation 警告添加
- ✅ 版本号更新为 v2.0.0
- ✅ Release Notes 编写

---

## 十二、下一步行动

### 立即行动（本周）

1. ✅ **确认设计**：Review 本计划，确认技术方案
2. ✅ **创建分支**：`git checkout -b feature/unified-query-api`
3. ✅ **开始 Stage 24.1**：Schema 标准化

### 短期行动（2 周内）

1. 完成 Stage 24.1-24.2（核心实现）
2. 开始 Stage 24.3（MCP 重构）
3. 每个 Stage 结束运行 `make all`

### 中期行动（1 个月内）

1. 完成 Stage 24.4-24.5（测试与文档）
2. 进行真实项目验证
3. 准备 v2.0.0 Release

---

## 附录 A：参考资料

### 设计文档

1. **统一查询 API 提案**：`/tmp/unified_query_api_proposal.md`
2. **Schema 对比报告**：`/tmp/corrected_schema_comparison_report.md`
3. **设计原则**：`docs/core/principles.md`

### 代码参考

1. **当前 ToolCall 实现**：`internal/parser/tools.go`
2. **查询函数库**：`internal/query/*.go`（4175 行）
3. **MCP 工具定义**：`cmd/mcp-server/tools.go`
4. **MCP 执行器**：`cmd/mcp-server/executor.go`

### 类似设计参考

1. **GraphQL**：统一查询接口，schema 定义资源
2. **OData**：URL 参数化查询，filter/orderby/top
3. **Elasticsearch Query DSL**：结构化查询语言

---

## 附录 B：常见问题 FAQ

### Q1: 为什么要统一查询接口？

**A**: 当前 16 个工具存在严重碎片化：
- 80+ 参数重复定义
- 3 种命名风格混乱
- 无法组合查询
- 学习成本高

统一接口带来：
- 94% 工具数量减少
- 75% 参数数量减少
- 无限组合能力
- 一致性体验

---

### Q2: 旧工具何时移除？

**A**: 分阶段废弃：
- **v2.0.0**（当前）：引入 `query`，旧工具保留
- **v2.1.0**（+3 个月）：旧工具标记 deprecated，显示警告
- **v3.0.0**（+6 个月）：移除旧工具

用户有至少 **6 个月时间**迁移。

---

### Q3: 性能会下降吗？

**A**: 不会显著下降：
- 统一接口使用相同的底层提取逻辑
- 为常见模式优化代码路径
- 基准测试确保性能目标（≤10% 回退）

实际上，统一接口可能**更快**（避免重复解析）。

---

### Q4: 如何处理复杂查询？

**A**: 三种方式：
1. **结构化查询**：使用 filter/transform/aggregate 组合
2. **jq 高级出口**：复杂逻辑用 jq_filter
3. **管道组合**：多次 query 调用组合结果

示例（复杂查询）：
```javascript
query({
  resource: "tools",
  filter: {tool_name: "Read", tool_status: "error"},
  transform: {extract: ["input.file_path"]},
  aggregate: {function: "count"}
})
```

---

### Q5: 迁移成本多大？

**A**: 取决于使用场景：
- **简单查询**：5 分钟学习，直接替换
- **复杂脚本**：15-30 分钟改写
- **自动化工具**：提供转换脚本

提供完整迁移指南 + 示例库，降低成本。

---

## 附录 C：示例代码片段

### 示例 1: 基础过滤查询

```go
// 查询所有失败的 Read 工具调用
params := query.QueryParams{
    Resource: "tools",
    Scope:    "project",
    Filter: query.FilterSpec{
        ToolName:   "Read",
        ToolStatus: "error",
    },
}

results, err := query.Query(entries, params)
```

---

### 示例 2: 聚合查询

```go
// 统计每个工具的调用次数
params := query.QueryParams{
    Resource: "tools",
    Aggregate: query.AggregateSpec{
        Function: "count",
        Field:    "tool_name",
    },
}

results, err := query.Query(entries, params)
// 输出: [{"tool_name": "Read", "count": 123}, ...]
```

---

### 示例 3: 复杂组合查询

```go
// 分析每个 Git 分支上失败的文件操作
params := query.QueryParams{
    Resource: "tools",
    Filter: query.FilterSpec{
        ToolName:   "Read|Edit|Write",
        ToolStatus: "error",
    },
    Transform: query.TransformSpec{
        Extract: []string{"git_branch", "tool_name", "input.file_path"},
        GroupBy: "git_branch",
    },
    Aggregate: query.AggregateSpec{
        Function: "count",
        Field:    "tool_name",
    },
    Output: query.OutputSpec{
        SortBy:    "count",
        SortOrder: "desc",
    },
}

results, err := query.Query(entries, params)
```

---

## 附录 D：测试策略矩阵

| 测试类型 | 覆盖范围 | 执行频率 | 失败影响 |
|---------|---------|---------|---------|
| 单元测试 | 所有函数 | 每次提交 | 阻止合并 |
| 集成测试 | MCP 工具 | 每次 PR | 阻止合并 |
| 性能测试 | 关键路径 | 每周 | 警告 |
| 回归测试 | 所有场景 | 发布前 | 阻止发布 |
| 真实项目验证 | 3 个项目 | 发布前 | 阻止发布 |

---

## 结论

Phase 24 的统一查询接口设计是 meta-cc 项目的**重要里程碑**，将极大提升用户体验和代码可维护性。

**关键成果**：
- ✅ 从 16 个工具 → 1 个工具（94% 简化）
- ✅ Schema 统一为 snake_case（100% 一致性）
- ✅ 可组合查询管道（无限扩展能力）
- ✅ 向后兼容迁移（用户友好）

**下一步**：
1. Review 并确认本计划
2. 开始 Stage 24.1 实施
3. 每个 Stage 严格遵循 TDD 原则

---

**文档版本**: v1.0
**创建日期**: 2025-10-23
**预计完成日期**: 2025-11-15
