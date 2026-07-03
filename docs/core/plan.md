# meta-cc 项目总体实施计划

## 项目概述

基于 [技术方案](../architecture/proposals/meta-cognition-proposal.md) 的分阶段实施计划。

**核心约束与设计原则**：详见 [设计原则文档](./principles.md)

**架构决策**：详见 [ADR 索引](../architecture/adr/README.md)

**项目状态**：
- ✅ **Phase 0-9 已完成**（核心查询 + 上下文管理）
- ✅ **Phase 14 已完成**（架构重构 + MCP 独立可执行文件）
- ✅ **Phase 15 已完成**（MCP 输出控制 + 工具标准化）
- ✅ **Phase 16 已完成**（混合输出模式 + 无截断 + 可配置阈值 + 集成测试）
- ✅ **Phase 17 已完成**（Subagent 形式化实现）
- ✅ **Phase 18-22 已完成**（开源发布与生态建设：GitHub Release + 插件分发 + 统一 /meta 命令 + 消息查询完整化）
- ✅ **Phase 23-25 已完成并归档**（查询接口重构 v2.0：jq-based API + 零学习成本）
- ✅ **Phase 26 已完成**（CLI 代码清理 + MCP-only 架构 + 文档更新）
- ✅ **Phase 29-30 已完成**（Query Reliability Improvements）
- ✅ **Phase 38 已完成**（Standard Plugin Packaging — Option M implemented）
- ✅ **Phase 41-46 已完成**（Project Split）
- ✅ **Phase 50-51 已完成**（Query Noise Reduction + session_count fix）
- ✅ **Phase 56-57 已完成**（Core Type Decoupling）
- ✅ **Phase 58-59 已完成**（Architecture Hygiene Phase 1）
- ✅ **Phase 60-64 已完成**（Architecture Hygiene Phase 2：TimeRange 规范化、pkg/ 移除）
- ✅ **Phase 66-70 已完成**（JSONL Streaming Reader with early image truncation）
- ✅ **Phase 71-75 已完成**（MCP Server Final Split into internal/mcp/*）
- ✅ **Phase 76-77 已完成**（Cleanup MCP Server Duplicates）
- ✅ **Phase 78-80 已完成**（Query Architecture Cleanup：jq/ 删除、子包拆分）
- 🟡 **Phase 81-86 部分完成**（Phases 81-84 done；Phase 85 ExecuteSpecialTool 重构、Phase 86 parser→types 迁移 待完成）
- ✅ **Phase 87-94 已完成**（Multi-Provider Conversation：Codex + Claude provider）
- ✅ 单元测试全部通过（新增 assistant messages + conversation 测试）
- ✅ 3 个真实项目验证通过（0% 错误率）
- ✅ 11 个 Slash Commands 可用
- ✅ 3 个 Subagents 可用
- ✅ MCP Server 独立可执行文件（`meta-cc-mcp`，16 个工具，支持混合输出模式）
- ✅ MCP 输出压缩率 80%+（10.7k → ~1-2k tokens）
- ✅ 混合输出模式：自动处理大数据（≤32KB inline，>32KB file_ref，无截断）
- ✅ 开源基础设施完成：LICENSE, CI/CD, 发布自动化
- ✅ 消息查询完整：user messages + assistant messages + conversation turns
- ✅ 插件打包：多平台包（5 平台）+ 自动安装脚本

---

## Phase 划分总览

```plantuml
@startuml
!theme plain

card "Phase 0-7" as P0 #lightgreen {
  **✅ MVP 已完成**
  - 项目初始化
  - 会话定位
  - JSONL 解析
  - 数据提取
  - 统计分析
  - 错误分析
  - Slash Commands
  - MCP Server
}

card "Phase 8" as P8 #lightblue {
  **查询命令基础**
  - query 命令框架
  - query tools
  - query user-messages
  - 基础过滤器
}

card "Phase 9" as P9 #lightblue {
  **上下文长度应对**
  - 分页支持
  - 分片输出
  - 字段投影
  - 紧凑格式(TSV)
}

card "Phase 10" as P10 #lightyellow {
  **高级查询能力**
  - 高级过滤器
  - 聚合统计
  - 时间序列
  - 文件级统计
}

card "Phase 11" as P11 #lightyellow {
  **Unix 可组合性**
  - 流式输出
  - 退出码标准化
  - stderr/stdout分离
  - Cookbook 文档
}

card "Phase 12-13" as P1213 #lightgreen {
  **MCP 集成与优化**
  - 项目级查询工具
  - 统一输出格式（JSONL/TSV）
  - 跨会话分析能力
}

card "Phase 14" as P14 #yellow {
  **架构重构与职责清晰化**
  - Pipeline 模式抽象
  - errors 命令简化
  - 输出排序标准化
  - 代码重复消除
}

card "Phase 15" as P15 #lightgreen {
  **MCP 输出控制与标准化**
  - 输出大小控制
  - 消息内容截断
  - 工具参数统一
  - 工具描述优化
}

card "Phase 16" as P16 #lightgreen {
  **MCP 输出模式优化** ✅
  - 混合输出模式
  - 文件引用机制
  - 临时文件管理
  - 32KB 阈值切换
  [详细文档](../guides/mcp.md)
}

card "Phase 17" as P17 #lightgreen {
  **Subagent 实现** ✅
  - @meta-coach 核心
  - @error-analyst 专用
  - @workflow-tuner 专用
  - 形式化规范
}

card "Phase 18-22" as P1822 #lightgreen {
  **开源发布与生态建设**
  - GitHub Release & CI/CD
  - 插件打包与分发
  - 自托管市场
  - 统一 /meta 命令系统
  - 消息查询完整化
}

note as P2325 #lightgrey
  **Phase 23-25: 查询接口重构 (v2.0)**
  已完成并归档至 docs/archive/
  - jq-based 三层 API
  - 零学习成本查询
  - 完整迁移指南
end note

card "Phase 26" as P26 #lightgreen {
  **CLI 代码清理** ✅
  - 移除 CLI 命令文件
  - 清理孤立 internal 包
  - MCP-only 架构
  - 更新文档反映新架构
}

P0 -down-> P8
P8 -down-> P9
P9 -down-> P10
P10 -down-> P11
P11 -down-> P1213
P1213 -down-> P14
P14 -down-> P15
P15 -down-> P16
P16 -down-> P17
P17 -down-> P1822
P1822 -down-> P2325
P2325 -down-> P26

note right of P0
  **业务闭环完成**
  可在 Claude Code 中使用
end note

note right of P9
  **核心查询能力完成**
  应对大会话场景
end note

note right of P17
  **完整架构实现**
  数据层 + MCP + Subagent
end note

note right of P1822
  **开源生态完成**
  社区化 + 能力系统
end note

note right of P26
  **架构简化**
  MCP-only 架构
  减少 ~20k 行代码
end note

@enduml
```

**Phase 优先级分类**：
- ✅ **已完成** (Phase 0-27): 完整功能实现
  - Phase 0-9: MVP + 核心查询 + 上下文管理
  - Phase 10-11: 高级查询和可组合性（部分实现）
  - Phase 12-13: MCP 集成与优化（合并）
  - Phase 14-15: 架构重构 + MCP 增强
  - Phase 16-17: 输出模式优化 + Subagent
  - Phase 18-22: 开源发布与生态建设（合并）
  - Phase 23-25: 查询接口重构 v2.0（已完成并归档）
  - Phase 26: CLI 代码清理（MCP 独立化）
  - Phase 27: 两阶段查询架构 (v2.1.0)

---

## 已完成阶段总览 (Phase 0-27)

详细文档见 `plans/` 目录。下表提供快速参考：

| Phase | 名称 | 状态 | 关键交付物 | 代码量 | 详细文档 |
|-------|------|------|-----------|--------|----------|
| 0 | 项目初始化 | ✅ | Go 模块、CLI 框架、测试环境 | ~150 行 | [plans/0/](../plans/00-bootstrap/) |
| 1 | 会话文件定位 | ✅ | 自动检测、--project 标志、环境变量 | ~180 行 | [plans/1/](../plans/01-session-locator/) |
| 2 | JSONL 解析器 | ✅ | 会话文件解析、数据结构定义 | ~200 行 | [plans/2/](../plans/02-jsonl-parser/) |
| 3 | 数据提取命令 | ✅ | `parse extract` 命令、工具调用提取 | ~200 行 | [plans/3/](../plans/03-data-extraction/) |
| 4 | 统计分析命令 | ✅ | `parse stats` 命令、基础统计 | ~150 行 | [plans/4/](../plans/04-stats-analysis/) |
| 5 | 错误模式分析 | ✅ | `analyze errors` 命令、错误聚合 | ~200 行 | [plans/5/](../plans/05-error-patterns/) |
| 6 | Slash Commands 集成 | ✅ | `/meta-stats`, `/meta-errors` 命令 | ~100 行 | [plans/6/](../plans/06-slash-commands/) |
| 7 | MCP Server 实现 | ✅ | 原生 MCP 服务器、初始工具集 | ~250 行 | 集成到 Phase 8 |
| 8 | 查询命令基础 | ✅ | `query` 命令框架、工具/消息查询 | ~1,250 行 | [plans/8/](../plans/08-mcp-integration/) |
| 9 | 上下文长度管理 | ✅ | 分页、字段投影、TSV 格式 | ~806 行 | [plans/9/](../plans/09-context-management/) |
| 10 | 高级查询能力 | 🟡 | 高级过滤器、时间序列（部分实现） | ~200-400 行 | [plans/10/](../plans/10-advanced-query/) |
| 11 | Unix 可组合性 | 🟡 | 流式输出、标准化退出码（部分实现） | ~300 行 | [plans/11/](../plans/11-unix-composability/) |
| 12-13 | MCP 集成与优化 | ✅ | 项目级查询、统一输出格式、跨会话分析 | ~850 行 | [plans/12/](../plans/12-mcp-project-query/), [plans/13/](../plans/13-output-simplification/) |
| 14 | 架构重构与 MCP 增强 | ✅ | Pipeline 模式、独立可执行文件 | ~900 行 | [plans/14/](../plans/14-architecture-refactor/) |
| 15 | MCP 输出控制与标准化 | ✅ | 输出大小控制、参数统一化 | ~350 行 | [plans/15/](../plans/15-mcp-standardization/) |
| 16 | MCP 输出模式优化 | ✅ | 混合输出模式、文件引用机制 | ~400 行 | [plans/16/](../plans/16-mcp-output-optimization/) |
| 17 | Subagent 实现 | ✅ | @meta-coach, @error-analyst, @workflow-tuner | ~1,000 行 | [Phase 17 详情](#phase-17-subagent-实现详细) |
| 18-22 | 开源发布与生态建设 | ✅ | GitHub Release、插件分发、统一/meta、消息查询完整化 | ~3,250 行 | [plans/18-22/](../plans/18-github-release-prep/) (里程碑汇总) |
| 23-25 | 查询接口重构 (v2.0) | ✅ | jq-based 三层 API、零学习成本、已归档 | ~5,650 行 | [归档文档](../archive/phase-23-25-query-refactoring.md) |
| 26 | CLI 代码清理（MCP 独立化） | ✅ | 移除 CLI 代码、MCP-only 架构、简化构建 | -19,500 行 | [详细计划](./phase-26-cli-removal-plan.md) |
| 27 | 两阶段查询架构 | ✅ | 删除 query/query_raw，新增元数据+Stage 2 查询工具 | ~550 行 (净增) | [Phase 27 详情](#phase-27-两阶段查询架构详细) |
| 28 | Prompt 优化学习系统 | ✅ | Capability 驱动的 prompt 优化、保存和重用机制 | ~450 行 | [Phase 28 详情](#phase-28-prompt-优化学习系统详细) |
| 29-30 | Query Reliability Improvements | ✅ | 严格参数验证、buffer 限制扩展、schema 修复 | — | [plans/29-30/](../plans/29-30/PLAN.md) |
| 38 | Standard Plugin Packaging (Option M) | ✅ | marketplace.json 对齐官方标准 | — | [plans/38-standard-plugin-packaging.md](../plans/38-standard-plugin-packaging.md) |
| 41-46 | Project Split | ✅ | 拆分 meta-cc 项目；MCP 工具迁移至 internal | — | [plans/41-46-project-split.md](../plans/41-46-project-split.md) |
| 50-51 | Query Noise Reduction | ✅ | exclude_system_messages 参数；session_count bug 修复 | — | [plans/50-51](../plans/50-51-query-noise-reduction.md) |
| 56-57 | Core Type Decoupling | ✅ | types 包规范化；parser 类型别名 | — | [plans/56-57](../plans/56-57-core-type-decoupling.md) |
| 58-59 | Architecture Hygiene Phase 1 | ✅ | 五个结构性问题修复 | — | [plans/58-59](../plans/58-59-architecture-hygiene.md) |
| 60-64 | Architecture Hygiene Phase 2 | ✅ | TimeRange 规范化；pkg/ 移除；cmd/mcp-server 瘦身 | — | [plans/60-64](../plans/60-64-architecture-hygiene-phase2.md) |
| 66-70 | JSONL Streaming Reader | ✅ | bufio.Reader.ReadBytes；early image truncation | — | [plans/66-70](../plans/66-70-jsonl-streaming-reader.md) |
| 71-75 | MCP Server Final Split | ✅ | internal/mcp/{executor,query,filters,...} 包拆分 | — | [plans/71-75](../plans/71-75-mcp-server-final-split.md) |
| 76-77 | Cleanup MCP Server Duplicates | ✅ | 删除 cmd/mcp-server 中的 JQRunner/QueryExecutor 重复 | — | [plans/76-77](../plans/76-77-cleanup-mcp-server-duplicates.md) |
| 78-80 | Query Architecture Cleanup | ✅ | internal/query/jq/ 删除；turnindex/sequences/assistant 子包 | — | [plans/78-80](../plans/78-80-query-architecture-cleanup.md) |
| 81-84 | Architecture Cleanup Part 1 | ✅ | ParseTimestamp 导出；workflow.go 瘦身；interfaces.go types 迁移 | — | [plans/81-86](../plans/81-86-arch-cleanup.md) |
| 85-86 | Architecture Cleanup Part 2 | 🟡 | ExecuteSpecialTool 重构 (85)；parser→types 完整迁移 (86) | — | [plans/81-86](../plans/81-86-arch-cleanup.md) |
| 87-94 | Multi-Provider Conversation | ✅ | internal/provider/{claude,codex}；query_conversation_flow 工具 | — | [plans/87-94](../plans/87-94-multi-provider-conversation.md) |

**注释**：
- **状态标识**：✅ 已完成，🟡 部分实现，📋 计划中
- **代码量**：估算值，包含源码和测试；负数表示删除，净增表示删除后新增
- Phase 7 集成到 Phase 8 的查询系统中
- Phase 10-11 核心功能已实现，部分高级特性待完善
- Phase 26 为架构简化 Phase，将移除过时的 CLI 代码
- Phase 27 重构查询架构，将查询规划责任转移到 Claude Code

---

## Phase 17: Subagent 实现（详细）

**目标**：实现语义分析层 Subagents，提供端到端的元认知分析能力，**完成三层架构**

**代码量**：~1000 行（配置 + 文档，包含 @meta-query）

### 架构层次

```
┌─────────────────────────────────────────┐
│         Subagent Layer (Phase 17)       │  ← 语义理解 + 多轮对话
│   @meta-coach, @error-analyst, etc.     │
├─────────────────────────────────────────┤
│         MCP Server (Phase 14-16)        │  ← 数据查询 + 过滤
│   query_tools, query_user_messages, etc│
├─────────────────────────────────────────┤
│         meta-cc CLI (Phase 0-13)        │  ← 数据提取 + 解析
│   parse, analyze, query commands        │
└─────────────────────────────────────────┘
```

### Subagent 职责划分

**@meta-coach** (通用元认知教练)：
- 工作流分析和优化建议
- 多维度综合评估（效率、质量、模式）
- 端到端会话分析
- 自动调用 MCP 工具获取数据

**@error-analyst** (错误分析专家)：
- 深度错误模式分析
- 根因分析和解决方案
- 预防性建议

**@workflow-tuner** (工作流优化专家)：
- 工具使用模式优化
- 交互效率提升
- 最佳实践推荐

### 实现策略

1. **使用 `.claude/agents/` 目录**（Claude Code 官方机制）
2. **Subagent 定义格式**：
   ```markdown
   ---
   name: meta-coach
   description: Metacognition coach for Claude Code workflows
   dependencies: meta-cc-mcp
   ---

   # Instructions
   You are a metacognition coach...

   ## MCP Tools Available
   - query_tools
   - query_user_messages
   ...
   ```

3. **MCP 依赖声明**：确保 Subagent 知道可用的 MCP 工具

### 开发阶段

#### Stage 17.1: @meta-coach 核心实现
- 创建 `.claude/agents/meta-coach.md`
- 实现核心分析逻辑（工作流、效率、模式）
- 集成 MCP 工具调用
- 测试端到端会话分析

#### Stage 17.2: @error-analyst 专用实现
- 创建 `.claude/agents/error-analyst.md`
- 实现错误模式分析逻辑
- 根因分析和解决方案生成
- 测试错误分析场景

#### Stage 17.3: @workflow-tuner 专用实现
- 创建 `.claude/agents/workflow-tuner.md`
- 实现工具使用优化逻辑
- 交互模式分析
- 测试工作流优化场景

#### Stage 17.4: 形式化文档
- 编写 Subagent 开发指南
- 创建 Subagent 使用示例
- 更新 CLAUDE.md 和 README.md
- 测试所有 Subagent

### 完成标准
- ✅ 3 个 Subagent 实现完成
- ✅ 可通过 `@meta-coach`, `@error-analyst`, `@workflow-tuner` 调用
- ✅ Subagent 可正确调用 MCP 工具
- ✅ 端到端测试通过
- ✅ 文档完整

详细计划见 `plans/17/`（如存在）

**Phase 23-25 归档说明**：查询接口重构 v2.0 已完成并归档至 `docs/archive/phase-23-25-query-refactoring.md`，包含完整的 jq-based 三层 API 设计和实现细节。

---

## Phase 27: 两阶段查询架构（详细）

**目标**：重构查询架构，将查询规划责任转移到 Claude Code，提供轻量级元数据工具和通用查询执行器

**代码量**：~550 行净增（删除 ~200 行 query/query_raw，新增 ~750 行）

**背景**：Phase 23-25 实现的通用 query/query_raw 接口存在语义不清晰问题（流式 vs 排序 vs 最近），且将查询规划职责放在 MCP server 导致灵活性受限。Phase 27 采用两阶段模式，让 Claude Code 自主规划查询策略。

### 架构转变

```
┌─────────────────────────────────────────────────────┐
│  旧架构 (Phase 23-25)                                │
│  Claude Code → query/query_raw (复杂查询逻辑)       │
│                  ↓                                   │
│               全量扫描 + jq 过滤                     │
├─────────────────────────────────────────────────────┤
│  新架构 (Phase 27)                                   │
│  Claude Code → Stage 1: 元数据查询 (轻量)            │
│              → 自主决策文件范围                       │
│              → Stage 2: 执行查询 (精准)               │
└─────────────────────────────────────────────────────┘
```

**核心优势**：
- ✅ 性能提升 79x（智能文件选择，3MB vs 453MB）
- ✅ 查询规划灵活（Claude Code 自主决策）
- ✅ 语义清晰（分阶段职责明确）
- ✅ 代码简化（删除模糊的通用接口）

### E2E 测试框架

**Phase 27 引入完整的 E2E 测试基础设施**，支持在不重启 Claude Code 的情况下测试 MCP server：

**测试方法**：
1. **直接 stdio 测试**（快速验证）
   ```bash
   echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
     ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | jq .
   ```

2. **自动化测试脚本**（推荐）
   ```bash
   ./tests/e2e/mcp-e2e-simple.sh ./meta-cc-mcp
   ```

3. **MCP Inspector**（交互调试）
   ```bash
   npm install -g @modelcontextprotocol/inspector
   mcp-inspector ./meta-cc-mcp
   ```

**测试文档**（已创建）：
- `docs/guides/mcp-e2e-testing.md` - 完整测试指南（13,000 字）
- `docs/guides/mcp-testing-quickstart.md` - 快速参考手册
- `docs/analysis/mcp-e2e-testing-recommendations.md` - 方法对比分析

**集成方式**：
- 每个 Stage 的验收标准包含 E2E 测试命令
- 集成到 Makefile（`make test-e2e-mcp`）
- 可集成到 CI/CD pipeline

### 删除的接口

**移除 2 个通用查询工具**（语义不清晰）：
- ❌ `query` - 删除（过滤/排序/切片顺序不明确）
- ❌ `query_raw` - 删除（与 query 功能重复）

**保留 10 个快捷查询工具**（高频场景优化）：
- ✅ `query_user_messages` - 用户消息查询
- ✅ `query_tools` - 工具调用查询
- ✅ `query_tool_errors` - 工具错误查询
- ✅ `query_token_usage` - Token 使用统计
- ✅ `query_conversation_flow` - 对话流查询
- ✅ `query_system_errors` - 系统错误查询
- ✅ `query_file_snapshots` - 文件快照查询
- ✅ `query_timestamps` - 时间戳查询
- ✅ `query_summaries` - 摘要查询
- ✅ `query_tool_blocks` - 工具块查询

### 新增 MCP 工具

#### Tool 1: `get_session_directory`

**功能**：返回 Claude Code 会话历史记录目录路径

**参数**：
```json
{
  "scope": {
    "type": "string",
    "enum": ["session", "project"],
    "default": "project",
    "description": "查询范围：'session' 返回当前会话文件所在目录，'project' 返回项目所有会话目录"
  }
}
```

**返回值**：
```json
{
  "directory": "/home/user/.claude/projects/-home-user-work-meta-cc",
  "scope": "project",
  "file_count": 660,
  "total_size_bytes": 474873856
}
```

**Description**（工具描述）：
```
Returns the directory path containing Claude Code session JSONL files.

Scope:
- "session": Returns directory of the most recently modified session file
- "project": Returns directory containing all session files for current project

Output Schema:
{
  "directory": string,        // Absolute path to session directory
  "scope": "session|project",
  "file_count": number,        // Total JSONL files in directory
  "total_size_bytes": number   // Total size of all JSONL files
}

Use Cases:
- Stage 1 of two-stage query: Get directory path
- Manual exploration of session data
- External tool integration (jq, grep, etc.)
```

#### Tool 2: `inspect_session_files`

**功能**：分析 JSONL 文件，返回文件级元数据（记录数、类型分布、时间范围等）

**参数**：
```json
{
  "files": {
    "type": "array",
    "items": {"type": "string"},
    "description": "要分析的 JSONL 文件路径列表（绝对路径）",
    "required": true
  },
  "include_samples": {
    "type": "boolean",
    "default": false,
    "description": "是否包含每个文件的前 3 条记录样本"
  }
}
```

**返回值**：
```json
{
  "files": [
    {
      "path": "/path/to/session-001.jsonl",
      "size_bytes": 1592690,
      "line_count": 265,
      "record_types": {
        "user": 45,
        "assistant": 42,
        "file-history-snapshot": 178
      },
      "time_range": {
        "earliest": "2025-10-26T07:00:00.000Z",
        "latest": "2025-10-26T09:28:30.542Z"
      },
      "mtime": "2025-10-26T09:28:30.000Z",
      "samples": [...]  // 可选，前 3 条记录
    }
  ],
  "summary": {
    "total_files": 3,
    "total_records": 570,
    "total_size_bytes": 3145728,
    "time_range": {
      "earliest": "2025-10-26T02:25:00.000Z",
      "latest": "2025-10-26T10:18:00.000Z"
    }
  }
}
```

**Description**（工具描述）：
```
Analyzes JSONL session files and returns file-level metadata.

Parameters:
- files: Array of absolute file paths to analyze
- include_samples: Whether to include first 3 records from each file (default: false)

Output Schema:
{
  "files": [
    {
      "path": string,           // Absolute file path
      "size_bytes": number,     // File size
      "line_count": number,     // Total lines (including empty)
      "record_types": {         // Record type distribution
        "user": number,
        "assistant": number,
        "file-history-snapshot": number
      },
      "time_range": {           // Timestamp range in file
        "earliest": string,     // ISO8601 timestamp
        "latest": string        // ISO8601 timestamp
      },
      "mtime": string,          // File modification time (ISO8601)
      "samples": [...]          // Optional: first 3 records
    }
  ],
  "summary": {                  // Aggregated statistics
    "total_files": number,
    "total_records": number,
    "total_size_bytes": number,
    "time_range": {
      "earliest": string,
      "latest": string
    }
  }
}

JSONL Record Schema (Claude Code session format):
{
  "type": "user|assistant|file-history-snapshot",
  "uuid": string,               // Unique identifier
  "timestamp": string,          // ISO8601 timestamp
  "sessionId": string,          // Session UUID
  "message": {                  // Present for user/assistant types
    "role": string,
    "content": string | array   // Text or array of content blocks
  },
  // Additional fields vary by type
}

Use Cases:
- Query planning: Decide which files to scan based on time_range
- Performance optimization: Avoid scanning old files for recent queries
- Data exploration: Understand session structure before querying
```

#### Tool 3: `execute_stage2_query`

**功能**：在指定文件上执行结构化查询（过滤 → 排序 → 转换 → 限制）

**参数**：
```json
{
  "files": {
    "type": "array",
    "items": {"type": "string"},
    "description": "要查询的 JSONL 文件路径列表",
    "required": true
  },
  "filter": {
    "type": "string",
    "description": "jq 过滤表达式（例如：'select(.type == \"user\")'）",
    "required": true
  },
  "sort": {
    "type": "string",
    "description": "jq 排序表达式（例如：'sort_by(.timestamp)'），为空则不排序",
    "default": ""
  },
  "transform": {
    "type": "string",
    "description": "jq 转换表达式（例如：'\"\\(.timestamp[:19]) | \\(.message.content[:150])\"'），为空则返回原始 JSON",
    "default": ""
  },
  "limit": {
    "type": "integer",
    "description": "返回结果数量限制（0 表示无限制）",
    "default": 0
  }
}
```

**返回值**：
```json
{
  "results": [
    {
      "formatted": "2025-10-26T10:17:57 | 现在，参考上面的方案...",
      "raw": { "type": "user", "timestamp": "...", ... }
    }
  ],
  "metadata": {
    "files_scanned": 3,
    "records_matched": 27,
    "records_sorted": 27,
    "records_returned": 10,
    "execution_time_ms": 54.42
  }
}
```

**Description**（工具描述）：
```
Executes a structured query on specified JSONL files using jq expressions.

Execution Order:
1. Load and Filter: Stream through files, apply filter expression to each record
2. Sort: Sort all filtered records (if sort expression provided)
3. Limit: Take first/last N records (if limit > 0)
4. Transform: Apply transform expression to each result record

Parameters:
- files: Array of absolute file paths (from get_session_directory or user selection)
- filter: jq filter expression (required)
  Example: 'select(.type == "user" and (.message.content | type == "string"))'
- sort: jq sort expression (optional, empty = no sort)
  Example: 'sort_by(.timestamp)'
- transform: jq transform for output formatting (optional, empty = raw JSON)
  Example: '"\(.timestamp[:19]) | \(.message.content[:150])"'
- limit: Maximum results to return (0 = all)

Output Schema:
{
  "results": [
    {
      "formatted": string,  // Result of transform (or JSON string if no transform)
      "raw": object         // Original JSON record
    }
  ],
  "metadata": {
    "files_scanned": number,
    "records_matched": number,   // After filter
    "records_sorted": number,    // After sort
    "records_returned": number,  // After limit
    "execution_time_ms": number
  }
}

Performance:
- Streaming: Files processed one-by-one, memory-efficient
- Early stopping: If limit reached during filtering, remaining files skipped
- Typical: 55ms for 3 files (3MB, 570 lines, filter → sort → limit 10)

jq Expression Compatibility:
- Uses gojq library (99% compatible with jq 1.6)
- Supports: select, map, sort_by, group_by, has, test, etc.
- Note: Some advanced functions may not be supported (e.g., @base64d)

Error Handling:
- Invalid jq expression: Returns error with line/column info
- Timeout: 30s limit per query
- Invalid JSON: Skips malformed lines (does not fail entire query)

Example Queries:
1. Recent user messages:
   filter: 'select(.type == "user" and (.message.content | type == "string"))'
   sort: 'sort_by(.timestamp)'
   limit: 10

2. Tool errors with timestamps:
   filter: 'select(.type == "user" and .message.content[].is_error == true)'
   transform: '"\(.timestamp) | \(.message.content[].content)"'

3. Token usage statistics:
   filter: 'select(.type == "assistant" and .message.usage)'
   transform: '{timestamp: .timestamp, tokens: .message.usage.output_tokens}'
```

### 实现策略

#### Stage 27.1: 删除旧接口（破坏性变更）

**删除文件**：
- `cmd/mcp-server/handlers_query.go` 中的 `handleQuery` 和 `handleQueryRaw`
- `cmd/mcp-server/executor.go` 中对应的工具分发逻辑

**更新测试**：
- 删除 `handlers_query_test.go` 中相关测试
- 确保 10 个快捷查询工具测试仍通过

**验收标准**：
- ✅ `query` 和 `query_raw` 工具不再可用
- ✅ 10 个快捷查询工具正常工作
- ✅ 所有测试通过（删除相关测试后）

**E2E 测试验证**：
```bash
# 验证工具已删除
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
  ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | \
  jq -e '.result.tools[] | select(.name == "query")' && \
  echo "❌ FAILED: query still exists" || echo "✓ query removed"

# 验证快捷工具仍可用
./tests/e2e/mcp-e2e-simple.sh
```

#### Stage 27.2: 实现 `get_session_directory`

**新增文件**：
- `cmd/mcp-server/handlers_stage1.go` - Stage 1 工具实现
- `cmd/mcp-server/handlers_stage1_test.go` - 单元测试

**实现逻辑**：
```go
func (e *ToolExecutor) handleGetSessionDirectory(cfg *config.Config, scope string, args map[string]interface{}) (string, error) {
    scopeParam := getStringParam(args, "scope", "project")

    loc := locator.NewSessionLocator()
    cwd, _ := os.Getwd()

    var directory string
    var fileCount int
    var totalSize int64

    if scopeParam == "session" {
        // 获取最近会话文件所在目录
        sessionFile, err := loc.FromProjectPath(cwd)
        if err != nil {
            return "", err
        }
        directory = filepath.Dir(sessionFile)
        fileCount = 1
        totalSize, _ = getFileSize(sessionFile)
    } else {
        // 获取项目所有会话目录
        sessionFiles, err := loc.AllSessionsFromProject(cwd)
        if err != nil {
            return "", err
        }
        directory = filepath.Dir(sessionFiles[0])
        fileCount = len(sessionFiles)
        totalSize = getTotalSize(sessionFiles)
    }

    result := map[string]interface{}{
        "directory": directory,
        "scope": scopeParam,
        "file_count": fileCount,
        "total_size_bytes": totalSize,
    }

    return serializeJSON(result), nil
}
```

**测试场景**（单元测试）：
- ✅ Session 范围查询
- ✅ Project 范围查询
- ✅ 无会话文件时错误处理
- ✅ 返回 JSON 格式正确

**E2E 测试验证**：
```bash
# 测试 project 范围
echo '{
  "jsonrpc":"2.0",
  "id":1,
  "method":"tools/call",
  "params":{
    "name":"get_session_directory",
    "arguments":{"scope":"project"}
  }
}' | ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | \
  jq -e '.result.content[0].text | fromjson | .directory' && \
  echo "✓ get_session_directory (project) works"

# 测试 session 范围
echo '{
  "jsonrpc":"2.0",
  "id":2,
  "method":"tools/call",
  "params":{
    "name":"get_session_directory",
    "arguments":{"scope":"session"}
  }
}' | ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | \
  jq -e '.result.content[0].text | fromjson | .file_count' && \
  echo "✓ get_session_directory (session) works"

# 运行自动化测试
./tests/e2e/mcp-e2e-simple.sh
```

#### Stage 27.3: 实现 `inspect_session_files`

**新增文件**：
- `internal/query/file_inspector.go` - 文件元数据分析核心
- `internal/query/file_inspector_test.go` - 单元测试

**实现逻辑**：
```go
type FileMetadata struct {
    Path        string            `json:"path"`
    SizeBytes   int64             `json:"size_bytes"`
    LineCount   int               `json:"line_count"`
    RecordTypes map[string]int    `json:"record_types"`
    TimeRange   TimeRange         `json:"time_range"`
    MTime       string            `json:"mtime"`
    Samples     []interface{}     `json:"samples,omitempty"`
}

func InspectFiles(files []string, includeSamples bool) ([]FileMetadata, error) {
    var results []FileMetadata

    for _, filepath := range files {
        metadata := FileMetadata{
            Path: filepath,
            RecordTypes: make(map[string]int),
        }

        // 获取文件信息
        fileInfo, _ := os.Stat(filepath)
        metadata.SizeBytes = fileInfo.Size()
        metadata.MTime = fileInfo.ModTime().Format(time.RFC3339)

        // 解析 JSONL
        file, _ := os.Open(filepath)
        scanner := bufio.NewScanner(file)

        var earliest, latest time.Time
        var samples []interface{}

        lineCount := 0
        for scanner.Scan() {
            lineCount++
            line := scanner.Text()
            if line == "" {
                continue
            }

            var entry map[string]interface{}
            json.Unmarshal([]byte(line), &entry)

            // 统计类型
            if entryType, ok := entry["type"].(string); ok {
                metadata.RecordTypes[entryType]++
            }

            // 时间范围
            if timestamp, ok := entry["timestamp"].(string); ok {
                t, _ := time.Parse(time.RFC3339, timestamp)
                if earliest.IsZero() || t.Before(earliest) {
                    earliest = t
                }
                if latest.IsZero() || t.After(latest) {
                    latest = t
                }
            }

            // 样本收集
            if includeSamples && len(samples) < 3 {
                samples = append(samples, entry)
            }
        }

        metadata.LineCount = lineCount
        metadata.TimeRange.Earliest = earliest.Format(time.RFC3339)
        metadata.TimeRange.Latest = latest.Format(time.RFC3339)
        if includeSamples {
            metadata.Samples = samples
        }

        results = append(results, metadata)
        file.Close()
    }

    return results, nil
}
```

**测试场景**（单元测试）：
- ✅ 单文件分析
- ✅ 多文件分析
- ✅ 包含样本 vs 不包含样本
- ✅ 空文件处理
- ✅ 无效 JSON 处理

**E2E 测试验证**：
```bash
# 获取会话目录
SESSION_DIR=$(echo '{
  "jsonrpc":"2.0",
  "id":10,
  "method":"tools/call",
  "params":{
    "name":"get_session_directory",
    "arguments":{"scope":"project"}
  }
}' | ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | \
  jq -r '.result.content[0].text | fromjson | .directory')

# 获取最近 3 个文件
FILES=$(ls -t "$SESSION_DIR"/*.jsonl 2>/dev/null | head -3 | jq -R . | jq -s .)

# 测试 inspect_session_files（不含样本）
echo "{
  \"jsonrpc\":\"2.0\",
  \"id\":11,
  \"method\":\"tools/call\",
  \"params\":{
    \"name\":\"inspect_session_files\",
    \"arguments\":{
      \"files\":$FILES,
      \"include_samples\":false
    }
  }
}" | ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | \
  jq -e '.result.content[0].text | fromjson | .files[] | .record_types' && \
  echo "✓ inspect_session_files works"

# 测试包含样本
echo "{
  \"jsonrpc\":\"2.0\",
  \"id\":12,
  \"method\":\"tools/call\",
  \"params\":{
    \"name\":\"inspect_session_files\",
    \"arguments\":{
      \"files\":$FILES,
      \"include_samples\":true
    }
  }
}" | ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | \
  jq -e '.result.content[0].text | fromjson | .files[] | .samples' && \
  echo "✓ inspect_session_files (with samples) works"
```

#### Stage 27.4: 实现 `execute_stage2_query`

**基于可行性验证**（已完成）：
- 核心代码已在 `test_stage2_query.go` 中验证
- 移植到 `internal/query/stage2_executor.go`
- 集成到 MCP 工具处理器

**实现逻辑**（已验证）：
```go
func ExecuteStage2Query(ctx context.Context, params Stage2QueryParams) ([]Stage2QueryResult, error) {
    // 1. 过滤阶段：流式读取文件 + jq 过滤
    filteredRecords := []interface{}{}
    filterCode, _ := compileJQ(params.Filter)

    for _, file := range params.Files {
        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            var entry interface{}
            json.Unmarshal(scanner.Bytes(), &entry)

            // 应用 jq 过滤
            if match := filterCode.Run(entry); match {
                filteredRecords = append(filteredRecords, entry)
            }
        }
    }

    // 2. 排序阶段（可选）
    if params.Sort != "" {
        sortCode, _ := compileJQ(params.Sort)
        sortedRecords = sortCode.Run(filteredRecords)
    } else {
        sortedRecords = filteredRecords
    }

    // 3. 限制阶段
    if params.Limit > 0 && len(sortedRecords) > params.Limit {
        sortedRecords = sortedRecords[len(sortedRecords)-params.Limit:]
    }

    // 4. 转换阶段（可选）
    results := []Stage2QueryResult{}
    if params.Transform != "" {
        transformCode, _ := compileJQ(params.Transform)
        for _, record := range sortedRecords {
            formatted := transformCode.Run(record)
            results = append(results, Stage2QueryResult{
                Formatted: formatted,
                Raw: record,
            })
        }
    } else {
        for _, record := range sortedRecords {
            results = append(results, Stage2QueryResult{
                Formatted: jsonSerialize(record),
                Raw: record,
            })
        }
    }

    return results, nil
}
```

**测试场景**（单元测试）：
- ✅ 基础过滤（已验证）
- ✅ 过滤 + 排序（已验证）
- ✅ 过滤 + 排序 + 限制（已验证）
- ✅ 过滤 + 排序 + 限制 + 转换（已验证）
- ✅ 无效 jq 表达式错误处理
- ✅ 超时处理（30s）
- ✅ 上下文取消处理

**E2E 测试验证**：
```bash
# 获取会话文件列表
SESSION_DIR=$(echo '{
  "jsonrpc":"2.0",
  "id":20,
  "method":"tools/call",
  "params":{
    "name":"get_session_directory",
    "arguments":{"scope":"project"}
  }
}' | ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | \
  jq -r '.result.content[0].text | fromjson | .directory')

FILES=$(ls -t "$SESSION_DIR"/*.jsonl 2>/dev/null | head -3 | jq -R . | jq -s .)

# 测试 1: 基础过滤
echo "{
  \"jsonrpc\":\"2.0\",
  \"id\":21,
  \"method\":\"tools/call\",
  \"params\":{
    \"name\":\"execute_stage2_query\",
    \"arguments\":{
      \"files\":$FILES,
      \"filter\":\"select(.type == \\\"user\\\")\",
      \"limit\":5
    }
  }
}" | ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | \
  jq -e '.result.content[0].text | fromjson | .results' && \
  echo "✓ execute_stage2_query (basic) works"

# 测试 2: 过滤 + 排序 + 限制
echo "{
  \"jsonrpc\":\"2.0\",
  \"id\":22,
  \"method\":\"tools/call\",
  \"params\":{
    \"name\":\"execute_stage2_query\",
    \"arguments\":{
      \"files\":$FILES,
      \"filter\":\"select(.type == \\\"user\\\")\",
      \"sort\":\"sort_by(.timestamp)\",
      \"limit\":10
    }
  }
}" | ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | \
  jq -e '.result.content[0].text | fromjson | .metadata.execution_time_ms' && \
  echo "✓ execute_stage2_query (with sort) works"

# 测试 3: 完整工作流（过滤 + 排序 + 限制 + 转换）
echo "{
  \"jsonrpc\":\"2.0\",
  \"id\":23,
  \"method\":\"tools/call\",
  \"params\":{
    \"name\":\"execute_stage2_query\",
    \"arguments\":{
      \"files\":$FILES,
      \"filter\":\"select(.type == \\\"user\\\")\",
      \"sort\":\"sort_by(.timestamp)\",
      \"transform\":\"\\\"\\\\(.timestamp[:19]) | \\\\(.message.content[:100])\\\"\",
      \"limit\":5
    }
  }
}" | ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | \
  jq -e '.result.content[0].text | fromjson | .results[] | .formatted' && \
  echo "✓ execute_stage2_query (full pipeline) works"

# 性能验证：< 100ms for 3MB data
EXEC_TIME=$(echo "{
  \"jsonrpc\":\"2.0\",
  \"id\":24,
  \"method\":\"tools/call\",
  \"params\":{
    \"name\":\"execute_stage2_query\",
    \"arguments\":{
      \"files\":$FILES,
      \"filter\":\"select(.type == \\\"user\\\")\",
      \"sort\":\"sort_by(.timestamp)\",
      \"limit\":10
    }
  }
}" | ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | \
  jq -r '.result.content[0].text | fromjson | .metadata.execution_time_ms')

if [ "$EXEC_TIME" -lt 100 ]; then
  echo "✓ Performance: ${EXEC_TIME}ms < 100ms"
else
  echo "⚠ Performance: ${EXEC_TIME}ms >= 100ms (may need optimization)"
fi
```

#### Stage 27.5: 文档和测试完善

**更新文档**：
- `docs/guides/mcp.md` - 新增两阶段查询指南
- `docs/guides/two-stage-query-guide.md` - 完整使用教程
- `docs/examples/two-stage-query-examples.md` - 查询示例库
- `CLAUDE.md` - 快速参考

**E2E 测试基础设施**（已完成）：
- ✅ `tests/e2e/mcp-e2e-simple.sh` - 自动化测试脚本
- ✅ `docs/guides/mcp-e2e-testing.md` - E2E 测试完整指南
- ✅ `docs/guides/mcp-testing-quickstart.md` - 快速参考
- ✅ `docs/analysis/mcp-e2e-testing-recommendations.md` - 测试方法分析

**E2E 测试扩展**（Stage 27.5 完成）：
```bash
# 更新测试脚本，添加 Phase 27 工具测试
vim tests/e2e/mcp-e2e-simple.sh

# 添加以下测试：
# - get_session_directory 验证
# - inspect_session_files 验证
# - execute_stage2_query 完整工作流
# - 性能基准测试（< 100ms）

# 验证所有测试通过
./tests/e2e/mcp-e2e-simple.sh
```

**集成到 CI/CD**：
```makefile
# Makefile 新增 target
test-e2e-mcp: build
	@bash tests/e2e/mcp-e2e-simple.sh ./meta-cc-mcp

test-all: test test-e2e-mcp
	@echo "✅ All tests passed (unit + E2E)"
```

**迁移指南**（破坏性变更）：
```markdown
# 从 query/query_raw 迁移到两阶段查询

## 旧方式（已弃用）
query({
  resource: "tools",
  jq_filter: 'select(.type == "user")',
  limit: 10
})

## 新方式（推荐）
// Stage 1: 获取目录并选择文件
dir = get_session_directory(scope="project")
files = list_most_recent_files(dir.directory, limit=3)

// Stage 2: 执行查询
results = execute_stage2_query(
  files=files,
  filter='select(.type == "user")',
  sort='sort_by(.timestamp)',
  limit=10
)
```

### 完成标准

**代码实现**：
- ✅ 删除 `query` 和 `query_raw` 工具
- ✅ 3 个新 MCP 工具实现并测试通过
- ✅ 10 个快捷查询工具保持兼容
- ✅ 所有单元测试通过（覆盖率 ≥ 80%）
- ✅ MCP 工具描述包含完整 schema 说明

**性能验证**：
- ✅ Stage 2 执行时间 < 100ms（3MB 数据）
- ✅ 智能查询加速 79x（3MB vs 453MB）
- ✅ 内存使用 < 10MB（单次查询）

**E2E 测试**：
- ✅ 自动化测试脚本可运行（`./tests/e2e/mcp-e2e-simple.sh`）
- ✅ 所有 Phase 27 工具通过 E2E 验证
- ✅ 性能基准测试通过
- ✅ 错误处理测试通过
- ✅ 集成到 Makefile（`make test-e2e-mcp`）

**文档完整性**：
- ✅ API 参考文档完整
- ✅ 迁移指南清晰
- ✅ 查询示例库丰富（10+ 示例）
- ✅ E2E 测试指南完整
- ✅ 快速参考手册可用

### 风险和缓解

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| 破坏性变更影响用户 | 高 | 中 | 提供清晰迁移指南，保留快捷查询 |
| 性能不达预期 | 低 | 中 | 已验证（55ms），可缓存文件元数据 |
| jq 表达式兼容性 | 中 | 中 | 文档化支持子集，提供示例库 |
| 学习曲线陡峭 | 高 | 中 | 丰富示例，Claude Code 辅助生成 |

### 预期收益

**性能**：
- 智能查询：79x 加速（3MB vs 453MB）
- Stage 2 执行：55ms（验证值）

**代码质量**：
- 代码量净增：+550 行（删除 200，新增 750）
- 语义清晰：分阶段职责明确
- 可维护性：删除模糊接口

**用户体验**：
- 灵活性提升：Claude Code 自主规划
- 可观测性：详细元数据和执行统计
- 学习曲线：需要示例支持

详细可行性分析见 [`docs/analysis/stage2-go-implementation-feasibility.md`](../analysis/stage2-go-implementation-feasibility.md)

---

## Phase 28: Prompt 优化学习系统（详细）

**目标**：实现纯 Capability 驱动的 Prompt 学习系统，通过保存和重用优化后的 prompts 实现渐进式智能化

**代码量**：~450 行（Markdown capabilities + 文档）

**背景**：用户使用 `/meta Refine prompt: XXX` 优化 prompts 后，需要手动记录和重用。Phase 28 实现自动化的 prompt 保存、搜索和重用机制，通过项目级历史积累实现越用越智能。

### 核心设计原则

**零侵入性**：
- ✅ 无需新 MCP 工具（纯 capability 实现）
- ✅ 无需修改 `/meta` 命令（完全兼容）
- ✅ 利用现有 capability 加载机制（子目录差异化）
- ✅ 零 Go 代码修改

**渐进式智能化**：
- ✅ 首次使用：正常优化流程（无历史）
- ✅ 再次使用：自动推荐历史版本（有匹配）
- ✅ 持续改进：使用频率追踪和效果评分
- ✅ 跨项目重用：统一数据结构（`.meta-cc/`）

**用户体验**：
- ✅ 自动初始化（静默创建目录）
- ✅ 可选保存（用户确认）
- ✅ 智能推荐（相似度匹配）
- ✅ CLI 友好（grep/jq 可检索）

### 架构设计

#### 数据目录结构

```
<project-root>/.meta-cc/
├── prompts/
│   ├── library/                    # 优化后的 prompts（扁平存储）
│   │   ├── release-full-ci-monitoring-001.md
│   │   ├── debug-error-analysis-001.md
│   │   └── refactor-extract-logic-001.md
│   └── metadata/                   # 使用统计（可选）
│       └── usage.jsonl
└── config.json                     # 项目级配置（可选）
```

**文件命名约定**：`{category}-{short-description}-{id}.md`

**文件格式**（YAML frontmatter + Markdown）：
```markdown
---
id: release-full-ci-monitoring-001
title: Full Release with CI Monitoring
category: release
keywords: [发布, release, 新版本, ci, 监控]
created: 2025-10-27T09:00:00Z
updated: 2025-10-27T09:10:00Z
usage_count: 2
effectiveness: 1.0
variables: [VERSION]
status: active
---

## Original Prompts
- 提交和发布新版本
- 发布新版本

## Optimized Prompt
使用预发布自动化工作流...
```

#### Capability 架构

**利用现有差异化加载机制**：

```
capabilities/
├── commands/                       # 公开（list_capabilities 可见）
│   ├── meta-prompt.md             # 主入口（扩展版）
│   └── ...
└── prompts/                        # 内部（list_capabilities 不可见）
    ├── meta-prompt-search.md      # 搜索历史 prompts
    ├── meta-prompt-save.md        # 保存优化后的 prompt
    ├── meta-prompt-list.md        # 列出 prompts
    └── meta-prompt-utils.md       # 通用工具函数
```

**关键发现**：现有 MCP capability 加载机制原生支持差异化：
- `list_capabilities()` 只扫描顶层 `*.md` 文件（不递归子目录）
- `get_capability("prompts/xxx")` 可加载子目录文件
- 零配置实现内部 capabilities

### 用户工作流

#### 场景 1：首次使用（自动初始化）

```
用户: /meta Refine prompt: 提交和发布新版本

Claude (meta-prompt):
[自动创建 .meta-cc/prompts/ 目录，用户无感知]

未找到历史版本，生成优化 prompts...
[显示 3 个优化选项]

用户: 使用 Option 1

Claude: 任务完成。是否保存以便重用？(Y/n)

用户: Y

Claude: ✓ 已保存到 .meta-cc/prompts/library/release-full-ci-monitoring-001.md
```

#### 场景 2：再次使用（智能推荐）

```
用户: /meta Refine prompt: 发布新版本

Claude (meta-prompt):
找到 1 个相关历史 prompt：

| 标题 | 相似度 | 使用次数 | 效果 |
|------|--------|---------|------|
| Full Release with CI Monitoring | 95% | 1 | 1.0 |

你可以：
1. 使用这个历史版本
2. 重新生成新版本

用户: 1

Claude: 检测到变量 {{VERSION}}，请提供值：

用户: v2.1.0

Claude: [执行优化 prompt]
✓ 使用次数已更新: 1 → 2
```

### 实现策略

#### Phase 1: 基础设施（MVP）

**目标**：实现保存和基本重用功能

**交付物**：
- 扩展 `commands/meta-prompt.md`（集成自动初始化和保存）
- 创建 `prompts/meta-prompt-save.md`（保存逻辑）
- 定义文件格式规范
- 更新文档（CLAUDE.md FAQ）

**工作量**：4-6 小时

**验收标准**：
- ✅ 自动创建 `.meta-cc/prompts/library/` 目录
- ✅ 用户可保存优化后的 prompt
- ✅ 生成符合规范的 .md 文件
- ✅ 文件包含完整 frontmatter 和内容

#### Phase 2: 搜索和重用

**目标**：实现历史搜索和智能推荐

**交付物**：
- 创建 `prompts/meta-prompt-search.md`（搜索匹配）
- 在 `meta-prompt` 中集成历史查询
- 实现相似度匹配算法（关键词重叠）
- 实现使用追踪（更新 usage_count）

**工作量**：4-6 小时

**验收标准**：
- ✅ 再次使用时自动搜索历史
- ✅ 显示匹配的历史 prompts
- ✅ 支持选择历史版本或生成新版本
- ✅ 使用后自动更新 usage_count

#### Phase 3: 管理和列表（可选）

**目标**：提供 prompt 管理能力

**交付物**：
- 创建 `prompts/meta-prompt-list.md`（列表和过滤）
- 支持按分类、使用频率排序
- 支持查看详细信息

**工作量**：2-3 小时

**验收标准**：
- ✅ 可列出所有保存的 prompts
- ✅ 支持过滤和排序
- ✅ 可查看详细信息

### 技术亮点

**差异化加载机制**：
- 利用现有 MCP capability 加载的原生特性
- 子目录文件不被 `list_capabilities` 列出
- `/meta` 用户界面保持简洁
- 内部 capabilities 通过 `get_capability("prompts/xxx")` 调用

**CLI 友好设计**：
- YAML frontmatter 便于快速提取元数据
- 纯文本格式支持 grep/awk/jq 检索
- 扁平目录便于 ls/find 浏览
- 跨项目统一结构（`.meta-cc/`）

**相似度匹配**（简单版）：
- 关键词 Jaccard 相似度
- 历史原始 prompts 匹配
- 使用频率加权排序

### 完成标准

**代码实现**：
- ✅ 扩展 `meta-prompt` capability（自动初始化 + 历史查询）
- ✅ 3-4 个子 capabilities 实现
- ✅ 文件格式规范定义
- ✅ 使用追踪机制

**用户体验**：
- ✅ 首次使用静默初始化
- ✅ 再次使用智能推荐
- ✅ 保存确认流程
- ✅ 变量替换支持

**文档完整性**：
- ✅ 用户指南（CLAUDE.md FAQ 更新）
- ✅ 文件格式规范
- ✅ CLI 工具使用示例
- ✅ 跨项目迁移指南

### 预期收益

**用户价值**：
- 🎯 快速重用优化的 prompts（减少 80% 优化时间）
- 📊 基于使用频率的智能推荐
- 🔄 持续改进机制（效果反馈）
- 💾 项目级知识积累

**技术价值**：
- ✅ 零依赖（无需新 MCP 工具）
- ✅ 零侵入（无需修改 /meta）
- ✅ 跨项目兼容（统一数据结构）
- ✅ CLI 友好（标准 Unix 工具可用）

**未来扩展**：
- Phase 28.4: 添加索引文件（性能优化）
- Phase 28.5: 添加全局级存储（跨项目共享）
- Phase 28.6: 添加效果反馈和智能推荐
- Phase 28.7: 社区 prompt 库（公开共享）

详细设计文档见 `plans/28/`

---

## 未来规划和扩展方向

### 短期优化 (1-2 个月)

**性能和可用性**：
- 优化大型会话文件的解析性能
- 改进 MCP 工具响应时间
- 增强错误信息的可读性
- 添加更多查询示例和模板

**文档和社区**：
- 完善用户指南和教程
- 创建视频演示
- 建立社区贡献指南
- 收集用户反馈和用例

### 中期发展 (3-6 个月)

**高级查询能力 (Phase 10-11 完善)**：
- 实现完整的时间序列分析
- 添加更复杂的聚合统计
- 增强 Unix 可组合性
- 提供查询 Cookbook

**智能分析**：
- 自动识别异常模式
- 预测性分析和建议
- 个性化工作流推荐
- 团队协作分析

**集成扩展**：
- 支持更多 IDE 和编辑器
- 导出分析报告（PDF、HTML）
- 集成第三方工具（Jira、GitHub Issues）
- API 服务化

### 长期愿景 (6-12 个月)

**AI 辅助优化**：
- 基于历史数据的智能建议
- 自动化工作流优化
- 学习用户偏好和模式
- 主动式问题预防

**企业级特性**：
- 多项目和团队分析
- 权限和访问控制
- 审计和合规性报告
- 云端部署选项

**生态系统建设**：
- 插件市场和扩展机制
- 自定义 Subagent 开发
- 社区贡献的能力库
- 培训和认证计划

---

## 风险和挑战

### 技术风险

| 风险 | 影响 | 缓解措施 | 状态 |
|------|------|----------|------|
| JSONL 格式变化 | 高 | 版本检测、向后兼容性测试 | ✅ 已实施 |
| 大型会话性能 | 中 | 流式处理、增量解析、混合输出模式 | ✅ 已解决 |
| MCP 协议变化 | 中 | 遵循官方标准、定期更新 | 🔄 持续监控 |
| 跨平台兼容性 | 低 | CI/CD 多平台测试 | ✅ 已实施 |

### 产品风险

| 风险 | 影响 | 缓解措施 | 状态 |
|------|------|----------|------|
| 用户采用率低 | 高 | 完善文档、降低使用门槛、社区推广 | 🔄 进行中 |
| 功能需求偏差 | 中 | 早期用户反馈、迭代开发 | 🔄 进行中 |
| 维护负担重 | 中 | 自动化测试、CI/CD、社区贡献 | ✅ 已实施 |

### 社区风险

| 风险 | 影响 | 缓解措施 | 状态 |
|------|------|----------|------|
| 贡献者不足 | 中 | 降低贡献门槛、指导文档、激励机制 | 📋 计划中 |
| 问题响应慢 | 中 | 建立维护团队、自动化问题分类 | 📋 计划中 |

---

## 参考资料

### 内部文档
- [设计原则](./principles.md) - 核心约束和架构决策
- [技术方案](../architecture/proposals/meta-cognition-proposal.md) - 整体架构设计
- [MCP 输出模式文档](../archive/mcp-output-modes.md) - 混合输出模式详解
- [集成指南](../guides/integration.md) - 选择 MCP/Slash/Subagent
- [能力开发指南](../guides/capabilities.md) - 能力系统开发
- [ADR 索引](../architecture/adr/README.md) - 架构决策记录

### 外部资源
- [Claude Code 官方文档](https://docs.claude.com/en/docs/claude-code/overview)
- [MCP 协议规范](https://modelcontextprotocol.io)
- [Go 项目布局标准](https://github.com/golang-standards/project-layout)

### 开发工具
- [cobra](https://github.com/spf13/cobra) - CLI 框架
- [viper](https://github.com/spf13/viper) - 配置管理
- [golangci-lint](https://golangci-lint.run/) - 代码质量检查

---

**最后更新**：2025-10-25
**维护者**：meta-cc 开发团队
