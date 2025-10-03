# Stage 8.12 执行总结

## 背景与动机

**会话分析**（Turn 43，0% 错误率）：
- 项目使用 meta-cc 进行自我分析，验证了"反身性设计"的可行性
- 重复调用 `get_session_stats → analyze_errors → query_tool_sequences`，表明这是标准分析流程
- Phase 8.1-8.7 已完成，但缺少智能 prompt 建议和优化能力

**用户需求**（从会话中提取）：
1. **"应可以方便地根据上下文和项目环境建议下一步的 prompt"**
2. **"应可以结合最近的上下文和项目环境，将用户的（口语化、不严谨的）prompt 改写成更清晰、更完备的 prompt"**
3. **坚持现有设计原则**：meta-cc 不做 NLP/LLM，只提供数据检索

## 设计方案

### 职责分离架构

```
┌─────────────────────────────────────────────────────────────┐
│                    meta-cc CLI（数据层）                     │
│                          无 LLM                              │
├─────────────────────────────────────────────────────────────┤
│ 1. query user-messages --with-context N                     │
│    → 用户消息 + 前后 N 轮上下文                              │
│                                                             │
│ 2. query project-state                                      │
│    → 未完成任务、最近文件、会话质量指标                      │
│                                                             │
│ 3. query successful-prompts                                 │
│    → 历史成功 prompts + 质量评分 + 结构特征                  │
│                                                             │
│ 4. query tool-sequences --successful-only --with-metrics    │
│    → 高成功率工作流模式 + 统计指标                           │
└─────────────────────────────────────────────────────────────┘
                          ↓ 结构化 JSON 数据
┌─────────────────────────────────────────────────────────────┐
│              Claude 集成层（语义理解 + 生成）                 │
│                      Slash Commands + Subagent              │
├─────────────────────────────────────────────────────────────┤
│ /meta-suggest-next                                          │
│  → 分析数据 → 生成 3 个优先级不同的 prompt 建议               │
│                                                             │
│ /meta-refine-prompt "<user_prompt>"                        │
│  → 识别模糊点 → 补充上下文 → 输出结构化 prompt               │
│                                                             │
│ @meta-coach（增强）                                          │
│  → 主动识别模糊 prompt → 引导式提问 → 改写建议               │
└─────────────────────────────────────────────────────────────┘
```

### 核心数据结构

#### 1. 上下文窗口查询

```json
{
  "matches": [
    {
      "turn_sequence": 38,
      "content": "分析我的工作流",
      "context_before": [
        {"turn": 35, "role": "assistant", "summary": "已完成 Stage 8.7", "tool_calls": ["Edit", "Bash"]},
        {"turn": 36, "role": "user", "content": "运行测试"},
        {"turn": 37, "role": "assistant", "summary": "所有测试通过", "tool_calls": ["Bash"]}
      ],
      "context_after": [
        {"turn": 39, "role": "assistant", "summary": "调用 meta-cc 分析", "tool_calls": ["Bash"]},
        {"turn": 40, "role": "user", "content": "很好，继续"}
      ]
    }
  ]
}
```

#### 2. 项目状态

```json
{
  "session_id": "6a32f273-...",
  "recent_files": [
    {"path": "plans/8/phase.md", "operations": ["Read", "Edit"], "edit_count": 3},
    {"path": "cmd/query_tools.go", "operations": ["Write", "Edit"], "edit_count": 2}
  ],
  "incomplete_stages": [
    {"phase": 8, "stage": "8.10", "title": "上下文和关联查询", "mentioned_in_turns": [15, 28, 40]},
    {"phase": 8, "stage": "8.11", "title": "工作流模式数据支持", "mentioned_in_turns": [15, 40]}
  ],
  "last_error_free_turns": 43,
  "current_focus": "Phase 8 integration improvements"
}
```

#### 3. 成功 Prompts 模式

```json
{
  "prompts": [
    {
      "turn_sequence": 15,
      "user_prompt": "实现 Stage 8.2：query tools 命令",
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

## 实施计划

### Step 1: 扩展 `query user-messages`（~60 行）

**文件**：`cmd/query_messages.go`

**新增功能**：
- `--with-context N` 参数
- 返回每个匹配前后 N 轮的精简上下文
- 上下文包含：turn、role、summary、tool_calls

**数据结构**：
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
    Summary   string   `json:"summary"`          // 前 100 字符或工具调用摘要
    ToolCalls []string `json:"tool_calls,omitempty"`
}
```

### Step 2: 实现 `query project-state`（~80 行）

**文件**：`cmd/query_project_state.go`

**功能**：
1. 解析 `file-history-snapshot` entries
2. 提取最近修改的文件（按操作次数排序，top 10）
3. 从用户消息中提取未完成任务（正则匹配 "Stage X.Y"、"Phase X"、"TODO"）
4. 计算会话质量指标：
   - 连续无错误轮次
   - 平均完成时间
   - 工具使用效率

**输出结构**：
```go
type ProjectState struct {
    SessionID            string            `json:"session_id"`
    RecentFiles          []FileActivity    `json:"recent_files"`
    IncompleteStages     []IncompleteTask  `json:"incomplete_stages"`
    LastErrorFreeTurns   int               `json:"last_error_free_turns"`
    CurrentFocus         string            `json:"current_focus"`      // 最近讨论最多的主题
    RecentAchievements   []string          `json:"recent_achievements"`
}

type FileActivity struct {
    Path           string   `json:"path"`
    LastModifiedTurn int    `json:"last_modified_turn"`
    Operations     []string `json:"operations"`
    EditCount      int      `json:"edit_count"`
}

type IncompleteTask struct {
    Phase            int      `json:"phase"`
    Stage            string   `json:"stage"`
    Title            string   `json:"title"`
    MentionedInTurns []int    `json:"mentioned_in_turns"`
}
```

### Step 3: 实现 `query successful-prompts`（~60 行）

**文件**：`cmd/query_successful_prompts.go`

**功能**：
1. 识别用户消息 → 助手完成 → 用户确认的模式
2. 计算质量评分：
   ```
   quality_score =
     0.4 * (1 - error_rate) +        // 无错误加分
     0.3 * speed_factor +             // 快速完成加分（< 5 turns）
     0.2 * deliverable_score +        // 有明确交付物加分
     0.1 * confirmation_score         // 用户明确确认加分
   ```
3. 提取 prompt 结构特征：
   - 是否有明确目标（动词 + 对象）
   - 是否有约束条件（代码行数、时间限制）
   - 是否有验收标准
   - 是否提供了上下文

**输出结构**：
```go
type SuccessfulPrompt struct {
    TurnSequence      int               `json:"turn_sequence"`
    UserPrompt        string            `json:"user_prompt"`
    Context           PromptContext     `json:"context"`
    Outcome           PromptOutcome     `json:"outcome"`
    QualityScore      float64           `json:"quality_score"`
    PatternFeatures   PromptFeatures    `json:"pattern_features"`
}

type PromptOutcome struct {
    Status           string   `json:"status"`           // success/partial/failed
    TurnsToComplete  int      `json:"turns_to_complete"`
    ErrorCount       int      `json:"error_count"`
    Deliverables     []string `json:"deliverables"`
}

type PromptFeatures struct {
    HasClearGoal           bool `json:"has_clear_goal"`
    HasConstraints         bool `json:"has_constraints"`
    HasAcceptanceCriteria  bool `json:"has_acceptance_criteria"`
    HasContext             bool `json:"has_context"`
}
```

### Step 4: 创建 Slash Commands（~150 行配置）

#### `/meta-suggest-next`

**工作流**：
1. 调用 `query user-messages --limit 5 --with-context 2`
2. 调用 `query project-state --include-incomplete-tasks`
3. 调用 `query tool-sequences --successful-only`
4. 将数据格式化后传递给 Claude
5. Claude 分析 → 生成 3 个优先级不同的 prompt 建议

**输出格式**：
```markdown
### 建议 Prompt 1（优先级：高）
**Prompt**: "实现 Stage 8.10：上下文和关联查询功能"

**理由**：
- 项目状态显示这是下一个未完成任务（mentioned in turns 15, 28, 40）
- 最近用户意图包含"分析我的工作流"，需要上下文查询支持
- 与当前 Phase 8 目标一致

**预期工作流**: Read → Grep → Edit → Bash（基于成功模式，成功率 95%）

---

### 建议 Prompt 2（优先级：中）
...
```

#### `/meta-refine-prompt "<user_prompt>"`

**工作流**：
1. 调用 `query project-state`
2. 调用 `query successful-prompts --limit 10`
3. 识别原 prompt 的模糊点
4. 补充上下文、约束、验收标准
5. 输出结构化 prompt

**输出格式**：
```markdown
### 优化后的 Prompt

实现 Stage 8.10：上下文和关联查询功能

**目标**：
- 在 `cmd/query_context.go` 中实现 `query context` 命令
- 支持 `--error-signature`、`--window` 参数
- 输出包含错误前后 N 轮的完整上下文

**约束**：
- 代码量：~180 行（参照 plans/8/phase.md）
- 包含单元测试
- 遵循现有 query 命令结构

**验收标准**：
- 所有单元测试通过
- 能够查询 Phase 5 错误分析中的错误签名
- JSON 输出符合规范

---

### 改进说明
- **原问题**：目标模糊（"搞一下"），缺少具体实现细节
- **改进点**：
  1. 明确 Stage 编号和功能名称
  2. 列出具体交付物
  3. 设定代码行数约束
  4. 提供验收标准
```

### Step 5: 增强 `@meta-coach`（~100 行文档）

**新增章节**：

```markdown
## 新增能力：Prompt 优化指导（Phase 8.12）

### 使用场景

1. **"我不知道下一步该做什么"**
   → 调用 `query project-state` + `query user-messages`
   → 提供 3 个具体的下一步 prompt 建议

2. **"这个任务太复杂，怎么分解？"**
   → 调用 `query successful-prompts` 查找类似任务
   → 提供分解策略和子任务 prompts

3. **用户 prompt 模糊或不完整**
   → 主动识别并改写 prompt
   → 补充缺失的上下文、约束、验收标准

### 工作流示例

#### 场景 1：建议下一步

```bash
用户: @meta-coach 我有点卡住了

# Step 1: 诊断
recent_intents=$(meta-cc query user-messages --limit 10 --with-context 3 --output json)
project_state=$(meta-cc query project-state --output json)

# Step 2: 引导提问
"我注意到你最近在做 Phase 8 集成改进，遇到什么困难了吗？"

# Step 3: 建议
workflows=$(meta-cc query tool-sequences --successful-only --output json)
"根据项目状态，下一步是实现 Stage 8.10（上下文查询）。
基于成功模式（Read → Grep → Edit，成功率 95%），我建议..."
```

#### 场景 2：Prompt 改写

```bash
用户: @meta-coach "帮我优化一下代码"

# Step 1: 识别模糊点
# "优化" → 性能？可读性？结构？
# "代码" → 哪个文件？

# Step 2: 获取上下文
project_state=$(meta-cc query project-state --output json)
# 最近修改：cmd/query_tools.go

# Step 3: 改写并确认
"我理解你想优化 cmd/query_tools.go。具体是想：
1. 提升性能？
2. 改进可读性？
3. 优化架构？

基于你最近的工作，我推荐：
[输出结构化 prompt]"
```
```

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

使用当前会话（Turn 43）验证：

```bash
# 场景 1：建议下一步
/meta-suggest-next
# 预期：建议实现 Stage 8.10 或 8.11

# 场景 2：改写 prompt
/meta-refine-prompt "完成那个上下文查询"
# 预期：输出包含具体文件、参数、验收标准的完整 prompt

# 场景 3：@meta-coach 引导
@meta-coach 感觉有点复杂
# 预期：分析当前状态，提供分解建议
```

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

## 预期收益

1. **开发效率提升 30%+**
   - 减少 prompt 试错时间
   - 快速定位下一步行动

2. **Prompt 质量提升**
   - 结构化、可测试
   - 包含完整上下文和验收标准

3. **用户体验改善**
   - 从"我不知道下一步做什么"→ 明确的行动指南
   - 从口语化 prompt → 专业的技术规格

## 后续扩展（Phase 9+）

### Phase 9: Prompt 模板库

基于大量成功 prompts 分析，提取可复用模板：

```bash
meta-cc query prompt-templates --category implementation
```

### Phase 10: 自动化 Prompt 生成

结合项目状态和模板，直接生成 prompt：

```bash
meta-cc generate next-prompt --use-template implementation
```

## 架构验证

### 职责分离检查表

| 组件 | 职责 | ✅/❌ |
|------|------|-------|
| meta-cc CLI | 数据检索（上下文、项目状态、成功模式） | ✅ |
| meta-cc CLI | 不做语义判断 | ✅ |
| meta-cc CLI | 不生成 prompt | ✅ |
| meta-cc CLI | 不做 NLP/LLM | ✅ |
| Slash Commands | 格式化数据、调用 Claude | ✅ |
| Claude（在 Slash 中） | 语义理解、prompt 生成 | ✅ |
| @meta-coach | 引导式对话、多轮推理 | ✅ |

**验证通过** ✅：所有语义处理由 Claude 完成，meta-cc 仅提供结构化数据。

## 总结

Stage 8.12 通过添加 3 个新的 meta-cc 查询命令和 2 个 Slash Commands，为智能 Prompt 建议和优化提供了完整的数据基础。该设计严格遵循职责分离原则，确保 meta-cc 保持轻量、可测试，同时为 Claude 集成层提供了丰富的数据维度。

**下一步**：等待用户确认后开始实施此 Stage。
