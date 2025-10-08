# meta-cc 设计原则与核心约束

本文档定义了 meta-cc 项目的核心设计原则、开发约束和架构决策。

---

## 一、核心约束

### 代码量控制
- 每个 Phase：代码修改量 ≤ 500 行
- 每个 Stage：代码修改量 ≤ 200 行

### 开发方法
- **测试驱动开发（TDD）**：每个 Stage 先写测试，后写实现
- **交付要求**：每个 Phase 更新 README.md，说明当前 build 使用方法
- **验证策略**：使用真实 Claude Code 会话历史进行测试

### 测试环境
- 测试 fixture：`tests/fixtures/` （包含样本和错误会话文件）
- 真实验证项目：meta-cc, NarrativeForge, claude-tmux
- 集成测试：`tests/integration/slash_commands_test.sh`

---

## 二、架构设计原则

### 1. 职责最小化原则

meta-cc 仅负责 Claude Code 会话历史知识的提取，不做分析决策。

**应该做的：**
- ✅ 提取：Turn、ToolCall、Error 等结构化数据
- ✅ 检测：基于规则的模式识别（重复错误签名、工具序列）

**不应该做的：**
- ❌ 分析：不做语义分析、不做决策（如窗口大小、聚合方式）
- ❌ 过滤：不预判用户需求，复杂过滤交给 jq/awk

### 2. Pipeline 模式

抽象通用数据处理流程，消除跨命令重复代码。

**标准流程：**
```
定位会话 → 加载 JSONL → 提取数据 → 输出格式化
```

### 3. 输出确定性

所有输出按稳定字段排序，解决 Go map 迭代随机性问题。

**排序规则：**
- `query tools` → 按 Timestamp 排序
- `query messages` → 按 turn_sequence 排序
- `query errors` → 按 Timestamp 排序

### 4. 延迟决策

将分析决策推给下游工具/LLM，meta-cc 只提供原始数据。

**职责边界：**
- ❌ meta-cc 不应实现：窗口过滤、错误聚合、模式计数
- ✅ 交给 jq/awk：`meta-cc query errors | jq '.[length-50:]'`
- ✅ 交给 Claude：Slash Commands 从 JSONL 生成语义建议

---

## 三、输出格式设计原则

### 核心原则

1. **双格式原则**：仅保留 JSONL 和 TSV
2. **格式一致性**：所有场景输出有效格式（包括错误）
3. **数据日志分离**：stdout=数据, stderr=日志
4. **Unix 可组合性**：meta-cc 提供简单检索，复杂过滤交给 jq/awk/grep
5. **客户端渲染**：移除 Markdown，由 Claude Code 自行渲染

### 格式选择

- **JSONL**（默认，`--stream`）：机器处理，jq 友好，流式性能
- **TSV**（`--output tsv`）：轻量级，awk/grep 友好，体积小

---

## 四、MCP Server 设计原则

### 架构分离：CLI vs MCP

**meta-cc CLI**（核心数据层）：
- ✅ 职责：JSONL 解析、数据提取、模式检测
- ✅ 输出：JSONL 格式（原始数据，无过滤）
- ✅ 不做：查询过滤、聚合统计、语义分析
- ✅ 可执行文件：`meta-cc`

**meta-cc-mcp**（MCP Server 层）：
- ✅ 职责：接收 MCP 请求，调用 CLI，使用 gojq 过滤/聚合
- ✅ 处理流程：
  1. 调用 `meta-cc` CLI 获取原始 JSONL
  2. 使用 `gojq` 库应用 jq 表达式过滤
  3. 可选统计聚合（stats_only/stats_first）
  4. **混合输出模式**（Phase 16）：
     - 输出 ≤ 阈值 → inline 模式（直接返回数据，无截断）
     - 输出 > 阈值 → file_ref 模式（返回临时文件元数据）
     - 阈值可配置（默认 8KB，参数或环境变量）
- ✅ 可执行文件：`meta-cc-mcp`（独立二进制）

### 查询语法：jq

**技术选择**：
- ✅ 使用 gojq 库（github.com/itchyny/gojq）
- ✅ jq 语法最适合 LLM（Claude 可直接生成表达式）
- ✅ 功能强大（过滤、投影、聚合、排序、分组）
- ✅ 纯 Go 实现，流式处理

**示例 jq_filter**：
```bash
# 过滤错误
.[] | select(.Status == "error")

# 统计工具分布
.[] | select(.Status == "error") | .ToolName | group_by(.) | map({tool: .[0], count: length})
```

### MCP 输出控制（Phase 15 + 16）

1. **单一格式**：仅支持 JSONL
2. **统计模式**：
   - `stats_only=true`：仅返回统计（如 `{"tool":"Bash","count":311}`）
   - `stats_first=true`：先统计后详情（用 `---` 分隔）
3. **阈值配置**（Phase 16.6）：
   - 参数：`inline_threshold_bytes`（默认 8192）
   - 环境变量：`META_CC_INLINE_THRESHOLD`
   - 用途：控制 inline/file_ref 模式切换点

### MCP 输出模式（Phase 16）

**混合输出策略**：根据输出大小自动选择输出模式

1. **Inline 模式**（输出 ≤ 8KB）：
   - 直接返回 JSONL 数据
   - 适合小查询结果（如 limit=5-10）
   - 单轮交互完成
   - 返回格式：
     ```json
     {
       "mode": "inline",
       "data": [
         {"Timestamp": "...", "ToolName": "Bash", "Status": "success"},
         ...
       ]
     }
     ```

2. **File Reference 模式**（输出 > 8KB）：
   - 写入临时 JSONL 文件（`/tmp/meta-cc-mcp-*`）
   - 返回文件元数据（路径、大小、行数、字段列表、摘要）
   - Claude 使用 Read/Grep/Bash 检索文件
   - 适合大查询结果（如全项目工具调用历史）
   - 返回格式：
     ```json
     {
       "mode": "file_ref",
       "file_ref": {
         "path": "/tmp/meta-cc-mcp-{session_hash}-{timestamp}-{query_type}.jsonl",
         "size_bytes": 524288,
         "line_count": 1523,
         "fields": ["Timestamp", "ToolName", "Status", "Error"],
         "summary": {
           "first_line": {...},
           "last_line": {...}
         }
       }
     }
     ```

**文件生命周期管理**：
- 临时文件位于 `/tmp/meta-cc-mcp-*`
- 文件命名：`meta-cc-mcp-{session_hash}-{timestamp}-{query_type}.jsonl`
- 按会话 hash 分组
- MCP 启动时自动清理 7 天前文件
- 可选手动清理工具：`cleanup_temp_files`

**使用场景**：
- 小查询（≤8KB）→ Claude 直接分析 inline 数据
- 大查询（>8KB）→ Claude 使用 Read/Grep/Bash 检索临时文件

**技术指标**：
- Inline 阈值：8KB（覆盖 ~80% 查询场景）
- File Reference 压缩率：>99%（仅返回元数据 ~100 bytes）
- 临时文件清理周期：7 天

**模式切换与阈值配置**（Phase 16.6 优化）：
- **inline_threshold_bytes**（默认 8KB，可配置）：决定使用 inline 还是 file_ref 模式
- **配置方式**：参数 `{"inline_threshold_bytes": 16384}` 或环境变量 `META_CC_INLINE_THRESHOLD=16384`
- **工作流程**：
  1. 查询结果 ≤ 阈值 → inline 模式 → 直接返回（无截断）
  2. 查询结果 > 阈值 → file_ref 模式 → 写入临时文件，返回元数据（~100 bytes）
- **设计哲学**：完全依赖 hybrid mode，无数据截断，信息完整性保证

### 默认查询范围与输出控制

**查询范围**：
- ✅ **默认查询范围为项目级**（所有会话）
- ✅ 工具名带 `_session` 后缀表示**仅查询当前会话**
- ✅ 保持 API 清晰：无后缀 = 项目级，`_session` = 会话级

**结果数量限制**：
- ✅ **默认无结果数量限制**（依赖混合输出模式）
- ✅ Claude 可显式传递 `limit` 参数来控制结果数量
- ✅ meta-cc-mcp 不预判用户需要多少数据

**输出控制策略**：
- ✅ 小结果（≤8KB）→ inline 模式，直接返回数据
- ✅ 大结果（>8KB）→ file_ref 模式，返回文件元数据，Claude 使用 Read/Grep/Bash 检索
- ✅ 混合输出模式确保大结果不会消耗过多 token

**设计理念**：
- meta-cc-mcp 不预判用户需要多少数据
- 让 Claude 根据对话上下文决定是否需要 limit
- 混合输出模式提供技术保障，无需人为设置默认限制

**何时显式使用 limit 参数**：
1. 用户明确要求"前 N 个结果"（如"最近 10 个错误"）
2. 只需要样本数据（如"给我看几个例子"）
3. 快速探索场景（先看少量数据，再决定是否扩展）

**工具命名约定**：

| 项目级（默认） | 会话级 | 说明 |
|--------------|--------|------|
| `query_tools` | `query_tools_session` | 工具调用查询 |
| `query_user_messages` | `query_user_messages_session` | 用户消息搜索 |
| `get_stats` | `get_session_stats` | 统计信息 |

---

## 五、职责分离与集成层次

### CLI 工具职责（meta-cc 核心）

meta-cc CLI 是系统的核心，提供清晰、简洁的对外接口：

- ✅ JSONL 解析和数据提取
- ✅ 基于规则的模式检测（错误重复、工具频率）
- ✅ 结构化数据输出（JSONL/TSV）
- ✅ 索引构建和查询优化

**接口设计原则**：
- 子命令应集中和简洁，避免碎片化
- 提供对项目/会话历史的统一访问
- 输出格式稳定，便于上层工具消费

### Claude Code 集成层次

**1. MCP Server（核心层）**

- ✅ meta-cc-mcp 作为主要接入点
- ✅ Claude 自主决策何时调用
- ✅ 支持 jq 表达式过滤和统计
- ✅ 项目级/会话级数据访问

**MCP 职责**：
- 调用 meta-cc CLI 获取原始数据
- 使用 gojq 应用 jq_filter 过滤
- 统计聚合（stats_only/stats_first）
- 输出长度控制（max_output_bytes）

**参数命名约定**：

MCP 工具参数与 CLI 参数保持一致，遵循 Unix 命名习惯：

| MCP 工具 | MCP 参数 | CLI 命令 | CLI 参数 | 一致性 |
|---------|---------|---------|---------|------|
| `query_user_messages` | `pattern` | `query user-messages` | `--pattern` | ✅ 完全一致 |
| `query_tool_sequences` | `pattern` | `analyze sequences` | `--pattern` | ✅ 完全一致 |

**设计原则**：
- 参数名遵循 Unix 标准术语（`pattern` 而非 `match`）
- MCP → CLI 参数直接映射，无需转换
- 参考：grep、sed、awk 均使用 "pattern" 而非 "match"

**2. Subagents（语义层）**

Subagents 分为两类：

**工具型 Agent**（辅助 Claude 在对话中执行复杂查询）：
- ✅ **@meta-query**：组织 CLI + Unix 管道进行复杂聚合
- ✅ 用于 Claude 需要多步 Unix 工具组合的场景
- ✅ 示例：`meta-cc query tools | jq ... | sort | uniq -c`

**业务型 Agent**（独立的元认知分析助手）：
- ✅ 各 Subagent 独立调用 meta-cc 工具
- ✅ **互不依赖或调用**（保持独立性）
- ✅ **必须说明 MCP 输出控制策略**（参考 `.claude/agents/meta-coach.md`）
- ✅ 支持多轮对话和上下文关联（在单个 Subagent 内部）

**MCP 输出控制要求**（Phase 16+ 推荐）：
- `stats_only=true`：仅统计（>99% 压缩）
- **Hybrid Mode**（默认）：自动选择 inline（≤8KB）或 file_ref（>8KB），无数据截断
- ~~`content_summary=true`~~：已弃用，使用 Hybrid Mode 代替（信息完整性更好）
- ~~`max_message_length=500`~~：已弃用（默认 0 = 无截断），使用 Hybrid Mode 代替
- `limit=10-20`：限制结果数量（仅在明确需要时使用）

**MCP 输出模式适配**（Phase 16）：
- 小查询（≤8KB）→ inline 模式，直接分析 data 字段
- 大查询（>8KB）→ file_ref 模式：
  1. 检查返回的 `mode` 字段
  2. 使用 Read 工具查看文件前 100 行（了解结构）
  3. 使用 Bash + jq/grep/awk 统计/过滤
  4. 使用 Grep 搜索特定模式

**业务型 Subagent 职责**：
- @meta-coach：综合分析、语义理解、建议生成
- @error-analyst：错误模式分析、根本原因诊断
- @workflow-tuner：工作流优化、自动化建议

**使用决策**：
- 单步 jq 可完成 → 使用 MCP
- 多步 Unix 管道 → Claude 在对话中使用 @meta-query
- 语义分析和建议 → 用户直接调用业务型 Subagent

**3. Slash Commands（快捷层）**

- ✅ 直接调用 meta-cc CLI
- ✅ 固定格式的快速报告
- ✅ 适合重复性查询

### 集成层次对比

**核心集成层**：

| 特性 | MCP Server | Slash Commands |
|------|-----------|----------------|
| 调用方式 | Claude 自主 | 用户 `/` |
| 查询能力 | jq 过滤/统计 | 固定逻辑 |
| 输出控制 | ✅ 内置（stats_only, max_output_bytes） | 固定格式 |
| 使用场景 | 对话中自然查询（80%） | 快速统计报告（20%） |

**Subagent 层**（基于 MCP）：

| 类型 | Agent | 调用方式 | 核心能力 | 输出控制 |
|------|-------|---------|----------|---------|
| 工具型 | @meta-query | Claude 自主（对话中） | Unix 管道聚合 | 手动 |
| 业务型 | @meta-coach | 用户 `@` | 综合分析、建议生成 | MCP 参数 |
| 业务型 | @error-analyst | 用户 `@` | 错误模式分析 | MCP 参数 |
| 业务型 | @workflow-tuner | 用户 `@` | 工作流优化建议 | MCP 参数 |

### 分离的好处

1. **职责清晰**：CLI（数据）→ MCP（过滤）→ Subagent（语义）
2. **性能优化**：gojq 库处理过滤，避免多次调用 CLI
3. **可测试性**：各层独立，易于测试
4. **灵活性**：MCP 覆盖常见场景（80%），@meta-query 处理复杂管道（20%）
5. **独立性**：业务型 Subagents 互不依赖，避免复杂的嵌套调用

---

## 六、索引设计原则

### 索引作为优化，而非必需

**阶段1（MVP）：无索引**
- 直接解析 JSONL 文件
- 适用于单会话分析（<1000 turns）
- 实现快速（1-2周）

**阶段2：可选索引**
- 仅在需要跨会话查询时启用
- 加速重复查询（如查找历史相似错误）
- SQLite 轻量级，零配置

**为什么索引是可选的？**
- 大多数场景只需分析当前会话
- 避免引入复杂性
- 渐进式优化路径

---

## 七、会话定位策略

### 优先级顺序

1. 命令行参数 `--session <id>`（遍历所有项目查找）
2. 命令行参数 `--project <path>`（转换为哈希，返回最新会话）
3. **自动检测（当前工作目录）**（默认方式）

### 自动检测优势

- ✅ Slash Command 执行时，Bash 工具的 cwd = 项目根目录
- ✅ 无需传递任何参数，用户体验最佳
- ✅ 与 Claude Code 实际行为完美匹配
- ✅ 支持多项目切换（通过 `--project` 参数）

---

## 八、测试策略

### 单元测试
- 每个 Stage 必须有对应的单元测试
- 测试覆盖率目标：≥ 80%
- 使用 `go test ./...` 运行所有测试

### 集成测试
- 每个 Phase 结束后运行集成测试
- 使用真实的会话文件 fixture
- 验证命令端到端流程

### 测试数据管理
- 测试 fixture 存放在 `tests/fixtures/`
- 使用真实的（脱敏的）Claude Code 会话文件
- 包含多种场景：正常会话、错误重复、工具密集使用等

---

## 九、风险和缓解措施

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Claude Code 会话文件格式变化 | 高 | 使用真实文件测试，版本化 fixture |
| 环境变量不可用 | 中 | 提供多种定位方式（参数、路径推断） |
| 测试覆盖不足 | 中 | TDD 强制要求，每个 Stage 先写测试 |
| 代码量超标 | 低 | 每个 Stage 结束检查行数，及时拆分 |
| Claude Code 集成失败 | 高 | 在测试环境充分验证 |

---

## 参考文档

- [meta-cc 项目总体实施计划](./plan.md)
- [Claude Code 元认知分析系统 - 技术方案](./proposals/meta-cognition-proposal.md)
- [Claude Code 官方文档](https://docs.claude.com/en/docs/claude-code/overview)
