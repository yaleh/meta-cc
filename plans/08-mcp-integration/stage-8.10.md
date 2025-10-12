# Stage 8.10: 上下文和关联查询（Context & Relation Queries）

## Overview

**Objective**: 实现上下文查询和关联查询功能，为 Slash Commands 和 @meta-coach 提供精准的上下文检索能力

**Code Budget**: ~180 lines (Go code)

**Time Estimate**: 2-3 hours

**Priority**: High（为 Claude 集成层提供上下文数据支持）

**Status**: 📋 Planned

## Design Principles

**职责边界**：
- ✅ **meta-cc 职责**: 数据提取、关联匹配、时间窗口过滤
- ✅ **Claude 集成层职责**: 语义理解、上下文分析、建议生成
- ❌ **meta-cc 不做**: 语义判断、Prompt 生成、建议输出

**数据流向**：
```
meta-cc query context → 结构化上下文数据 → Slash Command/Subagent → Claude 语义分析 → 用户建议
```

## Commands to Implement

### 1. query context - 错误上下文查询

**命令**：
```bash
meta-cc query context --error-signature <id> --window <N> [--output json|md]
```

**功能**：
- 查找指定错误模式的所有出现位置
- 返回每次出现前后 N 个 turns 的上下文
- 包含用户消息、工具调用、文件操作

**输出示例**：
```json
{
  "error_signature": "err-a1b2",
  "occurrences": [
    {
      "turn": 15,
      "context_before": [
        {"turn": 12, "role": "user", "preview": "Fix the auth bug"},
        {"turn": 13, "role": "assistant", "tools": ["Read test_auth.js"]},
        {"turn": 14, "role": "assistant", "tools": ["Edit test_auth.js"]}
      ],
      "error_turn": {
        "turn": 15,
        "tool": "Bash",
        "command": "npm test",
        "error": "TypeError: Cannot read property 'id' of undefined"
      },
      "context_after": [
        {"turn": 16, "role": "assistant", "tools": ["Edit test_auth.js"]},
        {"turn": 17, "role": "assistant", "tools": ["Bash: npm test"]}
      ]
    }
  ]
}
```

**Slash Command 应用**：
```markdown
# /meta-error-context [error-id]
context=$(meta-cc query context --error-signature "$error_id" --window 3 --output json)

Claude，基于以上上下文数据：
1. 分析错误发生前做了什么操作
2. 错误发生后的尝试是否有效
3. 建议下一步的调试方向
```

### 2. query file-access - 文件操作历史

**命令**：
```bash
meta-cc query file-access --file <path> [--group-by action] [--output json|md]
```

**功能**：
- 查询指定文件的所有操作历史（Read/Edit/Write）
- 支持按操作类型分组统计
- 返回时间序列和频率统计

**输出示例**：
```json
{
  "file": "test_auth.js",
  "total_accesses": 13,
  "operations": {
    "Read": 8,
    "Edit": 5,
    "Write": 0
  },
  "timeline": [
    {"turn": 10, "action": "Read", "timestamp": 1735689600},
    {"turn": 12, "action": "Edit", "timestamp": 1735689650},
    {"turn": 15, "action": "Read", "timestamp": 1735689700}
  ],
  "time_span_minutes": 23
}
```

**@meta-coach 应用**：
```markdown
file_history=$(meta-cc query file-access --file "$FILE" --output json)

# Claude 分析：
# - 文件被读取 8 次但只编辑 5 次，可能在反复查看
# - 建议：使用 Grep 搜索相关函数，一次性理解逻辑
```

### 3. query tool-sequences - 工具序列模式

**命令**：
```bash
meta-cc query tool-sequences --min-occurrences <N> [--pattern <seq>] [--output json|md]
```

**功能**：
- 检测重复出现的工具调用序列（如 Read → Edit → Bash）
- 返回序列出现次数和具体位置
- 可指定特定模式进行精确匹配

**输出示例**：
```json
{
  "sequences": [
    {
      "pattern": "Read → Edit → Bash",
      "count": 5,
      "occurrences": [
        {"start_turn": 10, "end_turn": 12},
        {"start_turn": 15, "end_turn": 17},
        {"start_turn": 19, "end_turn": 21},
        {"start_turn": 24, "end_turn": 26},
        {"start_turn": 28, "end_turn": 30}
      ],
      "time_span_minutes": 23
    },
    {
      "pattern": "Grep → Read → Read",
      "count": 4,
      "occurrences": [...]
    }
  ]
}
```

**@meta-coach 应用**：
```markdown
sequences=$(meta-cc query tool-sequences --min-occurrences 3 --output json)

# Claude 分析：
# - Read → Edit → Bash 出现 5 次，可能在反复测试同一修改
# - 建议：创建 /test-single 命令专注单个测试
```

### 4. 时间窗口查询扩展

**新增参数**（适用于所有 query 命令）：
```bash
--since <duration>       # 例：--since "5 minutes ago", --since "1 hour ago"
--last-n-turns <N>       # 查询最近 N 个 turns
--from <timestamp>       # 起始时间戳
--to <timestamp>         # 结束时间戳
```

**示例**：
```bash
# 查询最近 5 分钟的工具调用
meta-cc query tools --since "5 minutes ago"

# 查询最近 10 个 turns
meta-cc query tools --last-n-turns 10

# 查询特定时间范围
meta-cc query tools --from 1735689600 --to 1735693200
```

## Implementation Plan

### Step 1: 数据结构定义

**文件**: `internal/query/context.go`

```go
// ContextQuery 上下文查询结果
type ContextQuery struct {
    ErrorSignature string              `json:"error_signature,omitempty"`
    Occurrences    []ContextOccurrence `json:"occurrences"`
}

type ContextOccurrence struct {
    Turn          int          `json:"turn"`
    ContextBefore []TurnPreview `json:"context_before"`
    ErrorTurn     ErrorDetail  `json:"error_turn"`
    ContextAfter  []TurnPreview `json:"context_after"`
}

type TurnPreview struct {
    Turn      int      `json:"turn"`
    Role      string   `json:"role"`
    Preview   string   `json:"preview,omitempty"`
    Tools     []string `json:"tools,omitempty"`
    Timestamp int64    `json:"timestamp"`
}

// FileAccessQuery 文件访问查询结果
type FileAccessQuery struct {
    File          string                 `json:"file"`
    TotalAccesses int                    `json:"total_accesses"`
    Operations    map[string]int         `json:"operations"`
    Timeline      []FileAccessEvent      `json:"timeline"`
    TimeSpanMin   int                    `json:"time_span_minutes"`
}

type FileAccessEvent struct {
    Turn      int    `json:"turn"`
    Action    string `json:"action"` // Read/Edit/Write
    Timestamp int64  `json:"timestamp"`
}

// ToolSequenceQuery 工具序列查询结果
type ToolSequenceQuery struct {
    Sequences []SequencePattern `json:"sequences"`
}

type SequencePattern struct {
    Pattern       string              `json:"pattern"`
    Count         int                 `json:"count"`
    Occurrences   []SequenceOccurrence `json:"occurrences"`
    TimeSpanMin   int                 `json:"time_span_minutes"`
}

type SequenceOccurrence struct {
    StartTurn int `json:"start_turn"`
    EndTurn   int `json:"end_turn"`
}
```

### Step 2: 命令实现

**文件**: `cmd/query_context.go` (~80 lines)

```go
var queryContextCmd = &cobra.Command{
    Use:   "context",
    Short: "Query context around specific events (errors, files, etc.)",
    Run:   runQueryContext,
}

func init() {
    queryContextCmd.Flags().String("error-signature", "", "Error pattern ID to query")
    queryContextCmd.Flags().Int("window", 3, "Context window size (turns before/after)")
    queryCmd.AddCommand(queryContextCmd)
}

func runQueryContext(cmd *cobra.Command, args []string) {
    // 1. Locate and parse session
    // 2. Extract error occurrences
    // 3. Build context windows
    // 4. Format and output
}
```

**文件**: `cmd/query_file_access.go` (~50 lines)

```go
var queryFileAccessCmd = &cobra.Command{
    Use:   "file-access",
    Short: "Query file access history",
    Run:   runQueryFileAccess,
}

func init() {
    queryFileAccessCmd.Flags().String("file", "", "File path to query (required)")
    queryFileAccessCmd.Flags().Bool("group-by", false, "Group by action type")
    queryFileAccessCmd.MarkFlagRequired("file")
    queryCmd.AddCommand(queryFileAccessCmd)
}
```

**文件**: `cmd/query_sequences.go` (~50 lines)

```go
var querySequencesCmd = &cobra.Command{
    Use:   "tool-sequences",
    Short: "Query repeated tool call sequences",
    Run:   runQuerySequences,
}

func init() {
    querySequencesCmd.Flags().Int("min-occurrences", 3, "Minimum occurrences to report")
    querySequencesCmd.Flags().String("pattern", "", "Specific sequence pattern to match")
    queryCmd.AddCommand(querySequencesCmd)
}
```

### Step 3: 时间过滤器

**文件**: `internal/filter/time.go` (~40 lines)

```go
type TimeFilter struct {
    Since       string // "5 minutes ago"
    LastNTurns  int
    FromTs      int64
    ToTs        int64
}

func (f *TimeFilter) Apply(turns []parser.Turn) []parser.Turn {
    // 实现时间窗口过滤逻辑
}

func parseDuration(s string) (time.Duration, error) {
    // 解析 "5 minutes ago", "1 hour ago" 等
}
```

## Testing Strategy

### Unit Tests

**文件**: `internal/query/context_test.go`

```go
func TestQueryContext(t *testing.T) {
    tests := []struct {
        name          string
        errorSig      string
        window        int
        expectedCount int
    }{
        {"single occurrence", "err-a1b2", 3, 1},
        {"multiple occurrences", "err-a1b2", 2, 5},
        {"window=0", "err-a1b2", 0, 1},
    }
    // ...
}

func TestQueryFileAccess(t *testing.T) {
    // 测试文件访问查询
}

func TestQuerySequences(t *testing.T) {
    // 测试序列检测
}

func TestTimeFilter(t *testing.T) {
    // 测试时间过滤
}
```

### Integration Tests

```bash
# Test context query
./meta-cc query context --error-signature err-a1b2 --window 3 --output json | jq '.occurrences | length'

# Test file access
./meta-cc query file-access --file test_auth.js --output json | jq '.total_accesses'

# Test sequences
./meta-cc query tool-sequences --min-occurrences 3 --output json | jq '.sequences | length'

# Test time filters
./meta-cc query tools --since "10 minutes ago" --output json | jq '. | length'
./meta-cc query tools --last-n-turns 5 --output json | jq '. | length'
```

## Usage Examples

### Example 1: 错误上下文分析（Slash Command）

```markdown
# /meta-error-context [error-id]
---
name: meta-error-context
argument-hint: [error-pattern-id]
---

\`\`\`bash
error_id=${1:-"latest"}

# 获取错误上下文
context=$(meta-cc query context --error-signature "$error_id" --window 3 --output json)

# 获取相关文件历史
files=$(echo "$context" | jq -r '.occurrences[0].error_turn.file')
if [ -n "$files" ]; then
  file_history=$(meta-cc query file-access --file "$files" --output json)
fi
\`\`\`

Claude，基于以上数据：
1. 分析错误发生的上下文
2. 检查相关文件的修改历史
3. 建议具体的调试步骤
```

### Example 2: 工作流模式分析（@meta-coach）

```markdown
# .claude/agents/meta-coach.md

## 工作流模式分析

\`\`\`bash
# 获取工具序列模式
sequences=$(meta-cc query tool-sequences --min-occurrences 3 --output json)

# 获取文件访问频率（top 5）
# （需要先实现 stats files 或手动聚合）

# 获取最近活动
recent=$(meta-cc query tools --last-n-turns 10 --output json)
\`\`\`

基于以上数据，我会：
1. 识别重复的工具序列（可能是低效模式）
2. 发现频繁修改的文件（可能是问题热点）
3. 分析最近活动趋势
```

### Example 3: MCP Server 集成

```json
{
  "name": "query_context",
  "description": "获取特定错误/文件/关键词的上下文",
  "inputSchema": {
    "type": "object",
    "properties": {
      "error_signature": {
        "type": "string",
        "description": "错误模式 ID（可选）"
      },
      "window": {
        "type": "number",
        "default": 3,
        "description": "上下文窗口大小"
      }
    }
  }
}
```

**自然语言查询**：
```
User: "分析那个 TypeError 错误的上下文"

Claude 调用: query_context(error_signature="err-a1b2", window=3)

Claude 分析返回的上下文数据并给出建议
```

## Success Criteria

- ✅ `query context` 返回完整的错误上下文
- ✅ `query file-access` 统计文件操作历史
- ✅ `query tool-sequences` 检测重复序列
- ✅ 时间过滤器（--since, --last-n-turns）正常工作
- ✅ 所有单元测试通过
- ✅ 集成测试验证通过
- ✅ Slash Commands 和 @meta-coach 能成功使用新命令

## Documentation Updates

### Files to Update
1. `README.md` - 添加上下文查询示例
2. `docs/examples-usage.md` - 添加完整使用场景
3. `.claude/commands/meta-error-context.md` - 新建 Slash Command
4. `.claude/agents/meta-coach.md` - 更新工作流分析部分

## Dependencies

- ✅ Phase 0-7 completed (parser and basic query infrastructure)
- ✅ Stage 8.1-8.4 completed (query framework)
- 📋 `internal/parser` 支持提取文件路径（可能需要增强）

## Next Steps

After Stage 8.10:
- 📋 Stage 8.11: 工作流模式数据支持
- 📋 Stage 8.8-8.9: MCP Server 集成
- 📋 Phase 9: 上下文长度应对
