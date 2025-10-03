# Stage 8.12: Prompt 建议与优化功能（数据层）

## 概述

**目标**：为智能 Prompt 建议和优化提供数据检索能力

**职责边界**：
- ✅ meta-cc：提供结构化数据检索（上下文、项目状态、成功模式）
- ✅ Claude 集成层：语义理解、prompt 生成、建议排序
- ❌ meta-cc 绝不实现 `suggest-next-prompt` 或 `refine-prompt` 命令

**代码预算**：~200 行（Go 代码）

**时间估算**：2-3 小时

## 动机

根据当前会话分析（Turn 43，使用 meta-cc 自我分析），识别出以下需求：

### 应用场景
1. **建议下一步 Prompt**：
   - 基于用户最近意图、项目未完成任务、成功工作流模式
   - 提供 3 个优先级不同的具体 prompt 建议

2. **改写不严谨 Prompt**：
   - 将口语化 prompt 改写为结构化格式
   - 补充项目上下文、约束条件、验收标准
   - 参考历史成功 prompts 的模式

### 数据需求（meta-cc 提供）
- 最近用户消息 + 上下文窗口（前后 N 轮）
- 项目状态（未完成任务、最近修改文件）
- 成功工作流模式（高成功率的工具序列）
- 历史成功 prompts（带结果反馈）

## 新增 meta-cc 命令

### 1. 增强的上下文查询

```bash
# 查询用户消息，带完整上下文
meta-cc query user-messages \
  --match "实现|添加|修复|优化|分析" \
  --limit 10 \
  --with-context 3 \
  --output json

# 输出示例
{
  "matches": [
    {
      "turn_sequence": 38,
      "timestamp": "2025-10-03T10:25:00Z",
      "content": "分析我的工作流",
      "context_before": [
        {
          "turn": 35,
          "role": "assistant",
          "summary": "已完成 Stage 8.7",
          "tool_calls": ["Edit", "Bash"]
        },
        {
          "turn": 36,
          "role": "user",
          "content": "运行测试"
        },
        {
          "turn": 37,
          "role": "assistant",
          "summary": "所有测试通过",
          "tool_calls": ["Bash"]
        }
      ],
      "context_after": [
        {
          "turn": 39,
          "role": "assistant",
          "summary": "调用 meta-cc 分析",
          "tool_calls": ["Bash"]
        },
        {
          "turn": 40,
          "role": "user",
          "content": "很好，继续"
        }
      ]
    }
  ],
  "total_matches": 5
}
```

**实现要点**：
- 扩展 `query user-messages` 添加 `--with-context N` 参数
- 为每个匹配返回前后 N 轮的上下文
- 上下文包含 role、summary、tool_calls（精简数据）

---

### 2. 项目状态查询

```bash
# 查询当前项目状态
meta-cc query project-state \
  --source file-history \
  --include-incomplete-tasks \
  --output json

# 输出示例
{
  "session_id": "6a32f273-...",
  "recent_files": [
    {
      "path": "plans/8/phase.md",
      "last_modified_turn": 42,
      "operations": ["Read", "Edit"],
      "edit_count": 3
    },
    {
      "path": "cmd/query_tools.go",
      "last_modified_turn": 35,
      "operations": ["Write", "Edit"],
      "edit_count": 2
    }
  ],
  "incomplete_stages": [
    {
      "phase": 8,
      "stage": "8.10",
      "title": "上下文和关联查询",
      "mentioned_in_turns": [15, 28, 40]
    },
    {
      "phase": 8,
      "stage": "8.11",
      "title": "工作流模式数据支持",
      "mentioned_in_turns": [15, 40]
    }
  ],
  "last_error_free_turns": 43,
  "current_focus": "Phase 8 integration improvements",
  "recent_achievements": [
    "Stage 8.7 完成",
    "所有测试通过"
  ]
}
```

**实现要点**：
- 分析 `file-history-snapshot` entries 提取最近修改的文件
- 从用户消息中提取未完成任务（正则匹配 "Stage X.Y"、"Phase X"）
- 计算无错误轮次数（连续成功的 turns）
- 识别当前关注点（最近讨论最多的主题）

---

### 3. 成功工作流模式查询

```bash
# 查询成功的工具序列模式
meta-cc query tool-sequences \
  --min-occurrences 3 \
  --successful-only \
  --with-metrics \
  --output json

# 输出示例
{
  "sequences": [
    {
      "pattern": "Read -> Grep -> Edit",
      "occurrences": 15,
      "success_rate": 0.95,
      "avg_duration_minutes": 2.5,
      "example_turns": [12, 25, 38],
      "context": "代码修改工作流"
    },
    {
      "pattern": "Bash -> Read -> Edit -> Bash",
      "occurrences": 8,
      "success_rate": 1.0,
      "avg_duration_minutes": 5.2,
      "example_turns": [8, 18, 30],
      "context": "测试驱动开发循环"
    }
  ],
  "total_sequences": 2
}
```

**实现要点**：
- 基于 Stage 8.11 的工具序列检测
- 添加成功率过滤（`--successful-only`）
- 计算平均持续时间、成功率指标
- 提供示例 turns 用于上下文查询

---

### 4. 成功 Prompts 模式查询

```bash
# 查询历史成功的 prompts
meta-cc query successful-prompts \
  --limit 10 \
  --min-quality-score 0.8 \
  --output json

# 输出示例
{
  "prompts": [
    {
      "turn_sequence": 15,
      "user_prompt": "实现 Stage 8.2：query tools 命令",
      "context": {
        "phase": "Phase 8",
        "task_type": "implementation"
      },
      "outcome": {
        "status": "success",
        "turns_to_complete": 5,
        "error_count": 0,
        "deliverables": ["cmd/query_tools.go", "tests"]
      },
      "quality_score": 0.95,
      "pattern_features": {
        "has_clear_goal": true,
        "has_constraints": true,
        "has_acceptance_criteria": true,
        "has_context": true
      }
    }
  ]
}
```

**实现要点**：
- 识别用户消息 → 助手完成 → 用户确认的模式
- 质量评分基于：无错误、快速完成、有明确交付物
- 提取 prompt 的结构特征（目标、约束、验收标准）

---

## Claude 集成层实现

### Slash Command: `/meta-suggest-next`

**文件**：`.claude/commands/meta-suggest-next.md`

```markdown
---
name: meta-suggest-next
description: 根据会话上下文和项目状态建议下一步 prompt
allowed_tools: [Bash]
---

# meta-suggest-next：智能 Prompt 建议

基于当前会话上下文和项目状态，建议下一步最佳 prompt。

```bash
#!/bin/bash

# Step 1: 获取最近用户意图
recent_intents=$(meta-cc query user-messages --match "." --limit 5 --with-context 2 --output json)

# Step 2: 获取项目状态
project_state=$(meta-cc query project-state --include-incomplete-tasks --output json)

# Step 3: 获取成功工作流模式
workflows=$(meta-cc query tool-sequences --min-occurrences 3 --successful-only --output json)

# 将数据传递给 Claude 进行语义分析
cat <<EOF
## 上下文数据

### 最近用户意图
\`\`\`json
$recent_intents
\`\`\`

### 项目状态
\`\`\`json
$project_state
\`\`\`

### 成功工作流模式
\`\`\`json
$workflows
\`\`\`

## 任务

基于以上数据，分析并建议下一步最佳 prompt。考虑：

1. **未完成任务优先级**：从 incomplete_stages 中选择最紧急/重要的
2. **用户意图连续性**：根据 recent_intents 推断下一步自然延续
3. **成功模式复用**：建议使用已验证的高成功率工作流

输出格式（Markdown）：

### 建议 Prompt 1（优先级：高）
**Prompt**: [具体的、可执行的 prompt]

**理由**：
- [基于数据的分析，引用具体数据点]

**预期工作流**: [建议使用的工具序列，基于 workflows 数据]

### 建议 Prompt 2（优先级：中）
...

### 建议 Prompt 3（优先级：低/可选）
...
EOF
\```
```

---

### Slash Command: `/meta-refine-prompt`

**文件**：`.claude/commands/meta-refine-prompt.md`

```markdown
---
name: meta-refine-prompt
description: 将口语化 prompt 改写为清晰、完备的 prompt
allowed_tools: [Bash]
argument-hint: "<user_prompt>"
---

# meta-refine-prompt：Prompt 优化器

将用户的口语化、不严谨的 prompt 改写为更清晰、更完备的 prompt。

```bash
#!/bin/bash

USER_PROMPT="${1:-请输入要优化的 prompt}"

# Step 1: 获取项目上下文
project_state=$(meta-cc query project-state --output json)

# Step 2: 获取成功 prompts 参考
successful_prompts=$(meta-cc query successful-prompts --limit 10 --output json)

# 传递给 Claude 进行改写
cat <<EOF
## 原始 Prompt
> $USER_PROMPT

## 项目上下文
\`\`\`json
$project_state
\`\`\`

## 成功 Prompt 参考
\`\`\`json
$successful_prompts
\`\`\`

## 任务

基于以上数据，将原始 prompt 改写为结构化、完备的 prompt。

### 改写要求

1. **明确目标**：
   - 使用具体动词（实现、添加、修复、重构）
   - 指定对象和范围（具体文件、功能、模块）

2. **提供上下文**：
   - 关联当前项目状态（基于 project_state）
   - 说明为什么要做这个任务

3. **设定边界**：
   - 明确约束（代码行数、时间限制、依赖范围）
   - 列出交付物（文件、测试、文档）

4. **验收标准**：
   - 可测试的完成标志
   - 质量指标（测试覆盖率、性能要求）

### 输出格式

#### 优化后的 Prompt

[完整的、结构化的 prompt，包含上述 4 个要素]

#### 改进说明

- **原问题**：[原 prompt 的具体不足]
- **改进点**：[具体改进措施，引用成功 prompts 的模式]
- **预期效果**：[改写后的优势]
EOF
\```
```

---

### Subagent 增强: `@meta-coach`

在现有 `.claude/agents/meta-coach.md` 中添加：

```markdown
## 新增能力：Prompt 优化指导（Phase 8.12）

### 使用场景

当用户表达以下困惑时主动提供 prompt 优化：

1. **"我不知道下一步该做什么"**
   → 调用 `query project-state` 和 `query user-messages`
   → 提供 3 个具体的下一步 prompt 建议

2. **"这个任务太复杂，怎么分解？"**
   → 调用 `query successful-prompts` 查找类似任务
   → 提供分解策略和子任务 prompts

3. **用户 prompt 模糊或不完整**
   → 调用 `query project-state` 获取上下文
   → 主动改写 prompt，补充缺失信息

### 工作流

#### 场景 1：建议下一步

```bash
# 用户输入："我有点卡住了"

# Step 1: 诊断
recent_intents=$(meta-cc query user-messages --limit 10 --with-context 3 --output json)
project_state=$(meta-cc query project-state --output json)

# Step 2: 引导提问
"我注意到你最近在做 [从 project_state 提取]，遇到什么困难了吗？"

# Step 3: 建议
workflows=$(meta-cc query tool-sequences --successful-only --output json)
"根据项目状态，下一步是 [从 incomplete_stages 提取]。
基于成功模式 [引用 workflows]，我建议这样开始..."
```

#### 场景 2：Prompt 改写

```bash
# 用户输入："帮我优化一下代码"（模糊）

# Step 1: 识别模糊点
# "优化" → 性能？可读性？结构？
# "代码" → 哪个文件？哪个模块？

# Step 2: 获取上下文
project_state=$(meta-cc query project-state --output json)
# 当前在做：Phase 8
# 最近修改：cmd/query_tools.go

# Step 3: 改写并确认
"我理解你想优化 cmd/query_tools.go 的代码。
具体是想：
1. 提升性能（减少查询时间）？
2. 改进代码可读性（重构函数）？
3. 还是优化架构（拆分模块）？

基于你最近的工作，我推荐：
[提供结构化 prompt]"
```

### 数据来源总结

| 功能 | 使用的 meta-cc 命令 | 数据用途 |
|------|---------------------|---------|
| 建议下一步 | `query user-messages --with-context`<br>`query project-state`<br>`query tool-sequences` | 理解意图<br>识别未完成任务<br>参考成功模式 |
| 改写 Prompt | `query project-state`<br>`query successful-prompts` | 补充上下文<br>参考结构模式 |
| 诊断卡点 | `query user-messages --with-context`<br>`query tool-sequences` | 识别重复模式<br>发现低效操作 |
```

---

## 实施步骤

### Step 1: 扩展 `query user-messages`（~60 行）

**文件**：`cmd/query_messages.go`

**新增参数**：
- `--with-context N`：返回每个匹配前后 N 轮的上下文

**实现**：
```go
type UserMessageMatch struct {
    TurnSequence   int              `json:"turn_sequence"`
    Timestamp      string           `json:"timestamp"`
    Content        string           `json:"content"`
    ContextBefore  []ContextEntry   `json:"context_before,omitempty"`
    ContextAfter   []ContextEntry   `json:"context_after,omitempty"`
}

type ContextEntry struct {
    Turn      int      `json:"turn"`
    Role      string   `json:"role"`
    Summary   string   `json:"summary"`
    ToolCalls []string `json:"tool_calls,omitempty"`
}
```

---

### Step 2: 实现 `query project-state`（~80 行）

**文件**：`cmd/query_project_state.go`

**功能**：
- 解析 `file-history-snapshot` entries
- 提取最近修改的文件（按操作次数排序）
- 从用户消息中提取未完成任务（正则匹配）
- 计算会话质量指标（无错误轮次、平均完成时间）

**输出结构**：
```go
type ProjectState struct {
    SessionID            string            `json:"session_id"`
    RecentFiles          []FileActivity    `json:"recent_files"`
    IncompleteStages     []IncompleteTask  `json:"incomplete_stages"`
    LastErrorFreeTurns   int               `json:"last_error_free_turns"`
    CurrentFocus         string            `json:"current_focus"`
    RecentAchievements   []string          `json:"recent_achievements"`
}
```

---

### Step 3: 实现 `query successful-prompts`（~60 行）

**文件**：`cmd/query_successful_prompts.go`

**功能**：
- 识别用户消息 → 助手完成 → 用户确认的模式
- 计算质量评分（基于完成速度、错误率、交付物）
- 提取 prompt 结构特征

**质量评分算法**：
```go
quality_score = (
    0.4 * (1 - error_rate) +
    0.3 * speed_factor +      // 快速完成加分
    0.2 * deliverable_score + // 有明确交付物加分
    0.1 * confirmation_score  // 用户明确确认加分
)
```

---

### Step 4: 创建 Slash Commands（~150 行配置）

**文件**：
- `.claude/commands/meta-suggest-next.md`
- `.claude/commands/meta-refine-prompt.md`

**特点**：
- 调用多个 meta-cc 命令
- 数据格式化后传递给 Claude
- 明确的任务指令（输出格式、分析要求）

---

### Step 5: 增强 `@meta-coach`（~100 行文档）

**文件**：`.claude/agents/meta-coach.md`

**新增章节**：
- Prompt 优化能力说明
- 3 个场景的工作流示例
- 数据来源总结表

---

## 测试计划

### 单元测试

```bash
# 测试上下文查询
go test ./cmd -run TestQueryUserMessagesWithContext

# 测试项目状态提取
go test ./cmd -run TestQueryProjectState

# 测试成功 prompts 识别
go test ./cmd -run TestQuerySuccessfulPrompts
```

### 集成测试

```bash
# 测试 Slash Commands
/meta-suggest-next
# 验证：返回 3 个具体 prompt 建议

/meta-refine-prompt "帮我搞一下那个查询功能"
# 验证：改写为结构化 prompt

# 测试 @meta-coach
@meta-coach 我不知道下一步做什么
# 验证：提供数据驱动的建议
```

### 真实场景验证

使用当前会话（Turn 43）测试：

```bash
# 场景 1：建议下一步
/meta-suggest-next
# 预期建议：实现 Stage 8.10 或 8.11

# 场景 2：改写 prompt
/meta-refine-prompt "完成那个上下文查询"
# 预期输出：包含具体文件、参数、验收标准的完整 prompt

# 场景 3：@meta-coach 引导
@meta-coach 感觉有点复杂
# 预期：分析当前状态，提供分解建议
```

---

## 成功标准

### 功能完整性
- ✅ 3 个新 meta-cc 命令正确输出 JSON
- ✅ 2 个 Slash Commands 可用
- ✅ @meta-coach 增强文档完整

### 数据质量
- ✅ 上下文窗口准确（前后 N 轮）
- ✅ 项目状态提取准确（未完成任务、最近文件）
- ✅ 成功 prompts 识别准确（质量评分合理）

### 用户体验
- ✅ Prompt 建议具体、可执行
- ✅ Prompt 改写包含上下文和约束
- ✅ @meta-coach 引导自然、有帮助

---

## 后续扩展

### Phase 9: Prompt 模板库

基于大量成功 prompts 分析，提取可复用模板：

```bash
# 查询模板
meta-cc query prompt-templates --category implementation

# 输出
{
  "templates": [
    {
      "name": "实现新功能",
      "pattern": "实现 [Stage X.Y]: [功能名称]\n\n目标：...\n约束：...\n验收标准：...",
      "success_rate": 0.92,
      "avg_completion_turns": 5
    }
  ]
}
```

### Phase 10: 自动化 Prompt 生成

结合项目状态和模板，直接生成 prompt：

```bash
# 自动生成下一步 prompt
meta-cc generate next-prompt --use-template implementation

# 输出完整的、可直接使用的 prompt
```

---

## 总结

Stage 8.12 通过添加 3 个新的 meta-cc 查询命令，为智能 Prompt 建议和优化提供了数据基础。遵循职责分离原则：

- **meta-cc**：纯数据检索（上下文、项目状态、成功模式）
- **Claude（Slash/Subagent）**：语义理解、prompt 生成、建议排序

这种架构确保了 meta-cc 保持轻量、可测试，同时为 Claude 集成层提供了丰富的数据维度。

**下一步**：实施此 Stage，然后在真实会话中验证效果。
