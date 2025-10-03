# Stage 8.11: 工作流模式数据支持（Workflow Pattern Data）

## Overview

**Objective**: 实现工作流模式检测功能，为 @meta-coach 提供数据支持，帮助识别低效的工作模式

**Code Budget**: ~100 lines (Go code)

**Time Estimate**: 1-2 hours

**Priority**: Medium（提升 @meta-coach 分析能力，但不影响核心功能）

**Status**: 📋 Planned

## Design Principles

**职责边界**：
- ✅ **meta-cc 职责**: 检测重复模式、统计频率、计算时间跨度（基于规则）
- ✅ **@meta-coach 职责**: 语义理解模式含义、判断是否低效、生成优化建议
- ❌ **meta-cc 不做**: 判断模式好坏、生成建议、语义分析

**数据流向**：
```
meta-cc analyze sequences/file-churn/idle-periods → 统计数据 → @meta-coach → 语义理解 → 优化建议
```

## Commands to Implement

### 1. analyze sequences - 工具序列检测

**命令**：
```bash
meta-cc analyze sequences --min-length <N> --min-occurrences <M> [--output json|md]
```

**功能**：
- 检测重复出现的工具调用序列
- 支持自定义序列最小长度
- 返回序列频率和出现位置

**输出示例**：
```json
{
  "sequences": [
    {
      "pattern": "Read → Edit → Bash",
      "length": 3,
      "count": 5,
      "occurrences": [
        {
          "start_turn": 10,
          "end_turn": 12,
          "tools": [
            {"turn": 10, "tool": "Read", "file": "test_auth.js"},
            {"turn": 11, "tool": "Edit", "file": "test_auth.js"},
            {"turn": 12, "tool": "Bash", "command": "npm test"}
          ]
        },
        {"start_turn": 15, "end_turn": 17, "tools": [...]},
        {"start_turn": 19, "end_turn": 21, "tools": [...]},
        {"start_turn": 24, "end_turn": 26, "tools": [...]},
        {"start_turn": 28, "end_turn": 30, "tools": [...]}
      ],
      "time_span_minutes": 23
    }
  ]
}
```

**@meta-coach 应用**：
```markdown
sequences=$(meta-cc analyze sequences --min-length 3 --min-occurrences 3 --output json)

# Claude 分析：
# - "Read → Edit → Bash" 出现 5 次
# - 语义理解：可能在反复测试同一个修改
# - 建议：创建 /test-single 命令专注单个测试
```

### 2. analyze file-churn - 文件频繁修改检测

**命令**：
```bash
meta-cc analyze file-churn --threshold <N> [--output json|md]
```

**功能**：
- 检测被频繁访问的文件
- 阈值：访问次数 ≥ N 的文件
- 返回文件访问统计和时间跨度

**输出示例**：
```json
{
  "high_churn_files": [
    {
      "file": "test_auth.js",
      "read_count": 8,
      "edit_count": 5,
      "write_count": 0,
      "total_accesses": 13,
      "time_span_minutes": 23,
      "first_access": 1735689600,
      "last_access": 1735690980
    },
    {
      "file": "utils/auth.js",
      "read_count": 4,
      "edit_count": 3,
      "write_count": 1,
      "total_accesses": 8,
      "time_span_minutes": 15
    }
  ]
}
```

**@meta-coach 应用**：
```markdown
file_churn=$(meta-cc analyze file-churn --threshold 5 --output json)

# Claude 分析：
# - test_auth.js 被读取 8 次、编辑 5 次
# - 语义理解：可能对该文件逻辑不清晰
# - 建议：使用 Grep 搜索相关函数调用，理解整体流程
```

### 3. analyze idle-periods - 时间间隔分析

**命令**：
```bash
meta-cc analyze idle-periods --threshold <duration> [--output json|md]
```

**功能**：
- 检测会话中的长时间空闲期
- 阈值：超过指定时长（如 "5 minutes"）的空闲
- 返回空闲时段的开始/结束时间

**输出示例**：
```json
{
  "idle_periods": [
    {
      "start_turn": 15,
      "end_turn": 20,
      "duration_minutes": 7.5,
      "start_timestamp": 1735689700,
      "end_timestamp": 1735690150,
      "context_before": {
        "turn": 15,
        "tool": "Bash",
        "status": "error"
      },
      "context_after": {
        "turn": 20,
        "role": "user",
        "preview": "Let me try a different approach"
      }
    }
  ]
}
```

**@meta-coach 应用**：
```markdown
idle_periods=$(meta-cc analyze idle-periods --threshold "5 minutes" --output json)

# Claude 分析：
# - 检测到 7.5 分钟的空闲期
# - 空闲前：Bash 错误
# - 空闲后：用户说"尝试不同方法"
# - 语义理解：可能在思考或查找资料
# - 建议：下次遇到卡点时可以直接问我
```

## Implementation Plan

### Step 1: 数据结构定义

**文件**: `internal/analyzer/workflow.go`

```go
// SequenceAnalysis 序列分析结果
type SequenceAnalysis struct {
    Sequences []SequencePattern `json:"sequences"`
}

type SequencePattern struct {
    Pattern       string              `json:"pattern"`
    Length        int                 `json:"length"`
    Count         int                 `json:"count"`
    Occurrences   []SequenceOccurrence `json:"occurrences"`
    TimeSpanMin   int                 `json:"time_span_minutes"`
}

type SequenceOccurrence struct {
    StartTurn int               `json:"start_turn"`
    EndTurn   int               `json:"end_turn"`
    Tools     []ToolInSequence  `json:"tools"`
}

type ToolInSequence struct {
    Turn    int    `json:"turn"`
    Tool    string `json:"tool"`
    File    string `json:"file,omitempty"`
    Command string `json:"command,omitempty"`
}

// FileChurnAnalysis 文件频繁修改分析
type FileChurnAnalysis struct {
    HighChurnFiles []FileChurnDetail `json:"high_churn_files"`
}

type FileChurnDetail struct {
    File           string `json:"file"`
    ReadCount      int    `json:"read_count"`
    EditCount      int    `json:"edit_count"`
    WriteCount     int    `json:"write_count"`
    TotalAccesses  int    `json:"total_accesses"`
    TimeSpanMin    int    `json:"time_span_minutes"`
    FirstAccess    int64  `json:"first_access"`
    LastAccess     int64  `json:"last_access"`
}

// IdlePeriodAnalysis 空闲时段分析
type IdlePeriodAnalysis struct {
    IdlePeriods []IdlePeriod `json:"idle_periods"`
}

type IdlePeriod struct {
    StartTurn       int           `json:"start_turn"`
    EndTurn         int           `json:"end_turn"`
    DurationMin     float64       `json:"duration_minutes"`
    StartTimestamp  int64         `json:"start_timestamp"`
    EndTimestamp    int64         `json:"end_timestamp"`
    ContextBefore   *TurnContext  `json:"context_before,omitempty"`
    ContextAfter    *TurnContext  `json:"context_after,omitempty"`
}

type TurnContext struct {
    Turn    int    `json:"turn"`
    Role    string `json:"role,omitempty"`
    Tool    string `json:"tool,omitempty"`
    Status  string `json:"status,omitempty"`
    Preview string `json:"preview,omitempty"`
}
```

### Step 2: 命令实现

**文件**: `cmd/analyze_sequences.go` (~40 lines)

```go
var analyzeSequencesCmd = &cobra.Command{
    Use:   "sequences",
    Short: "Detect repeated tool call sequences",
    Run:   runAnalyzeSequences,
}

func init() {
    analyzeSequencesCmd.Flags().Int("min-length", 3, "Minimum sequence length")
    analyzeSequencesCmd.Flags().Int("min-occurrences", 3, "Minimum occurrences to report")
    analyzeCmd.AddCommand(analyzeSequencesCmd)
}

func runAnalyzeSequences(cmd *cobra.Command, args []string) {
    // 1. Locate and parse session
    // 2. Extract tool sequences
    // 3. Count occurrences
    // 4. Filter by thresholds
    // 5. Format and output
}
```

**文件**: `cmd/analyze_file_churn.go` (~30 lines)

**文件**: `cmd/analyze_idle.go` (~30 lines)

### Step 3: 分析逻辑实现

**文件**: `internal/analyzer/sequences.go` (~60 lines)

```go
func DetectSequences(turns []parser.Turn, minLength, minOccurrences int) SequenceAnalysis {
    // 1. Extract tool names from turns
    // 2. Find all n-grams (n >= minLength)
    // 3. Count occurrences
    // 4. Filter by minOccurrences
    // 5. Return results
}

func extractToolSequence(turns []parser.Turn, start, length int) SequenceOccurrence {
    // Extract specific sequence occurrence
}
```

**文件**: `internal/analyzer/file_churn.go` (~40 lines)

```go
func DetectFileChurn(turns []parser.Turn, threshold int) FileChurnAnalysis {
    // 1. Group tool calls by file
    // 2. Count Read/Edit/Write operations
    // 3. Filter by threshold
    // 4. Calculate time spans
    // 5. Return results
}
```

**文件**: `internal/analyzer/idle.go` (~40 lines)

```go
func DetectIdlePeriods(turns []parser.Turn, thresholdDuration string) IdlePeriodAnalysis {
    // 1. Parse threshold duration
    // 2. Calculate gaps between turns
    // 3. Filter by threshold
    // 4. Extract context before/after
    // 5. Return results
}
```

## Testing Strategy

### Unit Tests

```go
func TestDetectSequences(t *testing.T) {
    tests := []struct {
        name            string
        minLength       int
        minOccurrences  int
        expectedCount   int
    }{
        {"length=3, occ=3", 3, 3, 1},
        {"length=2, occ=5", 2, 5, 2},
    }
    // ...
}

func TestDetectFileChurn(t *testing.T) {
    // 测试文件频繁修改检测
}

func TestDetectIdlePeriods(t *testing.T) {
    // 测试空闲时段检测
}
```

### Integration Tests

```bash
# Test sequences
./meta-cc analyze sequences --min-length 3 --min-occurrences 3 --output json | jq '.sequences | length'

# Test file churn
./meta-cc analyze file-churn --threshold 5 --output json | jq '.high_churn_files | length'

# Test idle periods
./meta-cc analyze idle-periods --threshold "5 minutes" --output json | jq '.idle_periods | length'
```

## Usage Examples

### Example 1: @meta-coach 工作流诊断

```markdown
# .claude/agents/meta-coach.md

## 工作流模式分析

当用户说"感觉效率低"或"不知道哪里有问题"时，使用以下命令获取数据：

\`\`\`bash
# 检测工具序列
sequences=$(meta-cc analyze sequences --min-length 3 --min-occurrences 3 --output json)

# 检测文件频繁修改
file_churn=$(meta-cc analyze file-churn --threshold 5 --output json)

# 检测空闲时段
idle_periods=$(meta-cc analyze idle-periods --threshold "5 minutes" --output json)

# 获取最近活动
recent=$(meta-cc query tools --last-n-turns 20 --output json)
\`\`\`

基于以上数据，我会：
1. 识别重复的工具序列 → 判断是否低效
2. 发现频繁修改的文件 → 判断是否困惑
3. 分析空闲时段 → 判断是否卡住
4. 结合最近活动 → 给出具体建议
```

### Example 2: Slash Command - 工作流健康检查

```markdown
# /meta-workflow-check
---
name: meta-workflow-check
description: 检查工作流模式，识别低效操作
---

\`\`\`bash
sequences=$(meta-cc analyze sequences --min-length 3 --min-occurrences 3 --output json)
file_churn=$(meta-cc analyze file-churn --threshold 5 --output json)
idle_periods=$(meta-cc analyze idle-periods --threshold "5 minutes" --output json)
\`\`\`

Claude，基于以上数据：
1. 识别重复的工作流模式
2. 标记可能的低效点
3. 给出优化建议
```

### Example 3: MCP Server 集成

```json
{
  "name": "get_workflow_patterns",
  "description": "检测工作流模式（工具序列、文件访问、空闲时段）",
  "inputSchema": {
    "type": "object",
    "properties": {
      "min_occurrences": {
        "type": "number",
        "default": 3
      }
    }
  }
}
```

**自然语言查询**：
```
User: "帮我分析一下我的工作流，看看有没有低效的地方"

Claude 调用:
- get_workflow_patterns(min_occurrences=3)
  → 内部调用 analyze sequences, file-churn, idle-periods

Claude 分析返回的数据：
- "你在过去 20 分钟内重复了 5 次 Read → Edit → Bash 序列"
- "test_auth.js 被读取 8 次，可能对其逻辑不太清楚"
- "有一个 7 分钟的空闲期，之前是测试失败"
- 建议：...
```

## Success Criteria

- ✅ `analyze sequences` 检测重复序列
- ✅ `analyze file-churn` 识别频繁修改文件
- ✅ `analyze idle-periods` 检测长时间空闲
- ✅ 所有单元测试通过
- ✅ @meta-coach 能成功使用新命令进行分析
- ✅ 数据输出不包含语义判断，仅统计事实

## Documentation Updates

### Files to Update
1. `.claude/agents/meta-coach.md` - 添加工作流模式分析章节
2. `.claude/commands/meta-workflow-check.md` - 新建 Slash Command
3. `README.md` - 添加工作流分析示例
4. `docs/examples-usage.md` - 添加完整场景

## Dependencies

- ✅ Phase 0-7 completed (parser infrastructure)
- ✅ Stage 8.1-8.4 completed (query framework)
- 📋 Stage 8.10 completed (上下文查询，为 idle-periods 提供上下文)

## Next Steps

After Stage 8.11:
- 📋 Phase 9: 上下文长度应对（输出管理）
- 📋 Stage 8.8-8.9: MCP Server 集成（使用 8.10-8.11 新命令）
- 📋 更新 @meta-coach 使用新的工作流分析能力
