# Phase 8 扩展总结：上下文查询与工作流分析

## 概述

基于软件工程和架构分析，Phase 8 新增 **Stage 8.10-8.11**，扩展上下文查询和工作流模式检测能力，同时严格遵守 **meta-cc 职责边界原则**。

## 核心设计原则（已确认）

### 职责分离

**meta-cc 职责**（无 LLM/NLP）：
- ✅ 数据提取、过滤、聚合、统计
- ✅ 模式检测（基于规则，如序列重复、频率统计）
- ✅ 输出结构化数据（JSON/Markdown）
- ❌ **绝不做**: 语义判断、Prompt 生成、建议输出

**Claude 集成层职责**（Slash Commands/Subagent/MCP）：
- ✅ 语义理解（理解用户意图）
- ✅ 上下文关联（将多个 meta-cc 输出关联分析）
- ✅ 建议生成（基于数据给出可操作建议）
- ✅ 对话交互（Subagent 的多轮对话）

### 数据流向

```
用户需求（口语化）
  ↓
Slash Command/Subagent（理解意图）
  ↓
调用 meta-cc 命令（精准检索）
  ↓
meta-cc 返回结构化数据（无语义判断）
  ↓
Claude 语义分析 + 建议生成
  ↓
用户接收可操作建议
```

## 新增功能（Stage 8.10-8.11）

### Stage 8.10: 上下文和关联查询

**代码量**: ~180 lines
**时间**: 2-3 hours
**优先级**: High

**新增命令**:

1. **`query context`** - 错误上下文查询
   ```bash
   meta-cc query context --error-signature err-a1b2 --window 3
   ```
   返回：错误发生前后 N 个 turns 的完整上下文

2. **`query file-access`** - 文件操作历史
   ```bash
   meta-cc query file-access --file test_auth.js
   ```
   返回：文件的所有 Read/Edit/Write 操作统计

3. **`query tool-sequences`** - 工具序列模式
   ```bash
   meta-cc query tool-sequences --min-occurrences 3
   ```
   返回：重复出现的工具调用序列（如 Read → Edit → Bash）

4. **时间窗口查询**
   ```bash
   meta-cc query tools --since "5 minutes ago"
   meta-cc query tools --last-n-turns 10
   ```

**应用场景**:
- Slash Commands 获取错误上下文进行分析
- @meta-coach 使用文件访问历史识别困惑点
- MCP Server 提供更丰富的查询接口

### Stage 8.11: 工作流模式数据支持

**代码量**: ~100 lines
**时间**: 1-2 hours
**优先级**: Medium

**新增命令**:

1. **`analyze sequences`** - 工具序列检测
   ```bash
   meta-cc analyze sequences --min-length 3 --min-occurrences 3
   ```
   返回：重复的工具调用模式（纯统计，无语义判断）

2. **`analyze file-churn`** - 文件频繁修改检测
   ```bash
   meta-cc analyze file-churn --threshold 5
   ```
   返回：访问次数 ≥ 5 的文件统计

3. **`analyze idle-periods`** - 时间间隔分析
   ```bash
   meta-cc analyze idle-periods --threshold "5 minutes"
   ```
   返回：超过 5 分钟的空闲时段

**应用场景**:
- @meta-coach 使用序列检测识别重复模式
- Slash Commands 使用 file-churn 识别问题热点
- 工作流诊断时使用 idle-periods 分析卡点

## 新增 Slash Commands（示例）

### /meta-error-context

```markdown
---
name: meta-error-context
argument-hint: [error-pattern-id]
---

\`\`\`bash
error_id=${1:-"latest"}
context=$(meta-cc query context --error-signature "$error_id" --window 3 --output json)
\`\`\`

Claude，基于以上上下文数据：
1. 分析错误发生前的操作
2. 检查错误发生后的尝试是否有效
3. 建议具体的调试步骤
```

### /meta-workflow-check

```markdown
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

## @meta-coach 增强

### 新增章节：工作流模式分析

```markdown
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
1. 识别重复的工具序列 → 判断是否低效（语义理解）
2. 发现频繁修改的文件 → 判断是否困惑（语义理解）
3. 分析空闲时段 → 判断是否卡住（语义理解）
4. 结合最近活动 → 给出具体建议（语义生成）
```

## MCP Server 集成（Stage 8.8 更新）

新增 MCP 工具：

```json
{
  "name": "query_context",
  "description": "获取特定错误/文件的上下文",
  "inputSchema": {
    "error_signature": "string (optional)",
    "file": "string (optional)",
    "window": "number (default 3)"
  }
}
```

```json
{
  "name": "get_workflow_patterns",
  "description": "检测工作流模式（工具序列、文件访问）",
  "inputSchema": {
    "min_occurrences": "number (default 3)"
  }
}
```

**自然语言查询示例**:

```
User: "分析那个 TypeError 错误的上下文"
→ Claude 调用: query_context(error_signature="err-a1b2", window=3)
→ Claude 分析返回的上下文数据并给出建议

User: "帮我分析一下我的工作流，看看有没有低效的地方"
→ Claude 调用: get_workflow_patterns(min_occurrences=3)
→ Claude 分析返回的模式数据并给出优化建议
```

## 实施优先级

### 高优先级（立即实施）
1. **Stage 8.10**: 上下文和关联查询（2-3h）
   - 为 Slash Commands/Subagent 提供上下文检索
   - 解决"如何获取错误上下文"的痛点

2. **Stage 8.11**: 工作流模式数据支持（1-2h）
   - 为 @meta-coach 提供工作流分析数据
   - 帮助识别低效模式

### 中优先级（后续实施）
3. **Stage 8.8-8.9**: MCP Server 集成（2h）
   - 使用 8.10-8.11 新命令
   - 完善 MCP 工具集

## 文档更新

### 已更新
- ✅ `docs/plan.md` - Phase 8-11 描述更新
- ✅ `plans/8/phase.md` - 完整 Stage 8.10-8.11 规划
- ✅ `plans/8/README.md` - 代码量更新
- ✅ `plans/8/stage-8.10.md` - 详细实施计划
- ✅ `plans/8/stage-8.11.md` - 详细实施计划

### 待创建（实施时）
- 📋 `.claude/commands/meta-error-context.md`
- 📋 `.claude/commands/meta-workflow-check.md`
- 📋 更新 `.claude/agents/meta-coach.md`（工作流分析章节）

## 成功标准

### 功能验证
- ✅ `query context` 返回完整上下文
- ✅ `query file-access` 统计文件操作
- ✅ `analyze sequences` 检测重复序列
- ✅ 时间过滤器正常工作
- ✅ 工作流模式检测提供数据（无语义判断）

### 集成验证
- ✅ Slash Commands 能使用新命令
- ✅ @meta-coach 能基于新数据进行分析
- ✅ MCP Server 支持自然语言查询
- ✅ 所有输出不包含语义判断，仅统计事实

## 关键改进点总结

### 相比原建议的修正

**原建议**（不符合设计原则）：
- ❌ `meta-cc suggest next-prompt`（语义判断）
- ❌ `meta-cc refine prompt`（NLP 处理）

**修正后**（符合设计原则）：
- ✅ `meta-cc query context`（数据提取）
- ✅ `meta-cc analyze sequences`（统计分析）
- ✅ Slash Commands/Subagent 负责语义理解

### 核心价值

1. **严格职责分离**
   - meta-cc 纯数据处理，易于测试和维护
   - Claude 集成层负责语义理解，充分发挥 LLM 能力

2. **丰富数据接口**
   - 上下文查询解决"如何获取相关信息"
   - 工作流模式检测提供分析所需数据

3. **可扩展性**
   - 新命令易于集成到 Slash/Subagent/MCP
   - 数据格式标准化，便于后续增强

## 下一步行动

用户手动启动执行时：

1. 实施 Stage 8.10（2-3h）
2. 实施 Stage 8.11（1-2h）
3. 更新 Slash Commands 和 @meta-coach
4. 测试所有集成
5. 更新主文档

---

**文档版本**: 1.0
**创建时间**: 2025-10-03
**依据**: 软件工程和架构分析建议（修正版）
