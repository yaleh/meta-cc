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
- ✅ **Phase 18 已完成**（GitHub Release 准备）
- ✅ **Phase 19 已完成**（Assistant 响应查询 + 对话分析）
- ✅ **Phase 20 已完成**（插件打包与发布）
- ✅ **Phase 21 已完成**（自托管插件市场）
- ✅ **Phase 22 已完成**（统一 Meta 命令 + 多源能力发现）
- ✅ **Phase 23 已完成**（查询能力函数库化 + MCP 完全去 CLI 依赖）
- ✅ **Phase 24 已完成**（统一查询接口设计与实现 - Schema 标准化 + 统一 Query API）
- ✅ 单元测试全部通过（新增 assistant messages + conversation 测试）
- ✅ 3 个真实项目验证通过（0% 错误率）
- ✅ 11 个 Slash Commands 可用
- ✅ 3 个 Subagents 可用
- ✅ MCP Server 独立可执行文件（`meta-cc-mcp`，16 个工具，支持混合输出模式）
- ✅ MCP 输出压缩率 80%+（10.7k → ~1-2k tokens）
- ✅ 混合输出模式：自动处理大数据（≤8KB inline，>8KB file_ref，无截断）
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

card "Phase 12" as P12 #lightgreen {
  **MCP 项目级查询**
  - 项目级工具（默认）
  - 会话级工具（_session）
  - --project . 支持
  - 跨会话分析
}

card "Phase 13" as P13 #lightgreen {
  **输出格式简化**
  - JSONL/TSV 双格式
  - 格式一致性
  - 错误处理标准化
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
  - 8KB 阈值切换
  [详细文档](../guides/mcp.md)
}

card "Phase 17" as P17 #lightgreen {
  **Subagent 实现** ✅
  - @meta-coach 核心
  - @error-analyst 专用
  - @workflow-tuner 专用
  - 形式化规范
}

card "Phase 18" as P18 #lightyellow {
  **GitHub Release 准备**
  - LICENSE + 开源合规
  - CI/CD 流水线
  - Release 自动化
  - 社区文档完善
}

card "Phase 19" as P19 #lightgreen {
  **消息查询增强**
  - Assistant 响应查询
  - 对话分析
  - 完整消息链
}

card "Phase 20" as P20 #lightgreen {
  **插件打包与发布**
  - 多平台包
  - 自动安装脚本
  - 发布流程优化
}

card "Phase 21" as P21 #lightgreen {
  **自托管插件市场**
  - 市场配置
  - 一键安装
  - 版本管理
}

card "Phase 22" as P22 #lightgreen {
  **统一 Meta 命令**
  - 多源能力发现
  - 语义匹配
  - 动态加载
}

card "Phase 23" as P23 #lightgreen {
  **查询能力函数库化**
  - 提取 query* 逻辑为库
  - CLI/MCP 共用函数
  - 共享 jq/输出工具
  - 回归测试串联
}

card "Phase 24" as P24 #lightgray {
  **统一查询接口**
  - 单一 query 工具
  - 资源导向设计
  - 可组合过滤器
  - Schema 标准化
}

P0 -down-> P8
P8 -down-> P9
P9 -down-> P10
P10 -down-> P11
P11 -down-> P12
P12 -down-> P13
P13 -down-> P14
P14 -down-> P15
P15 -down-> P16
P16 -down-> P17
P17 -down-> P18
P18 -down-> P19
P19 -down-> P20
P20 -down-> P21
P21 -down-> P22
P22 -down-> P23
P23 -down-> P24

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

note right of P18
  **开源发布准备**
  社区化和自动化
end note

note right of P22
  **能力系统完成**
  统一入口 + 动态扩展
end note

@enduml
```

**Phase 优先级分类**：
- ✅ **已完成** (Phase 0-22): 完整功能实现
  - Phase 0-9: MVP + 核心查询 + 上下文管理
  - Phase 10-11: 高级查询和可组合性（部分实现）
  - Phase 12-13: MCP 项目级 + 输出简化
  - Phase 14-15: 架构重构 + MCP 增强
  - Phase 16-17: 输出模式优化 + Subagent
  - Phase 18-22: 开源发布 + 能力系统

---

## 已完成阶段总览 (Phase 0-22)

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
| 12 | MCP 项目级查询 | ✅ | 项目级工具、跨会话分析 | ~450 行 | [plans/12/](../plans/12-mcp-project-query/) |
| 13 | 输出格式简化 | ✅ | JSONL/TSV 统一、格式一致性 | ~400 行 | [plans/13/](../plans/13-output-simplification/) |
| 14 | 架构重构与 MCP 增强 | ✅ | Pipeline 模式、独立可执行文件 | ~900 行 | [plans/14/](../plans/14-architecture-refactor/) |
| 15 | MCP 输出控制与标准化 | ✅ | 输出大小控制、参数统一化 | ~350 行 | [plans/15/](../plans/15-mcp-standardization/) |
| 16 | MCP 输出模式优化 | ✅ | 混合输出模式、文件引用机制 | ~400 行 | [plans/16/](../plans/16-mcp-output-optimization/) |
| 17 | Subagent 实现 | ✅ | @meta-coach, @error-analyst, @workflow-tuner | ~1,000 行 | [Phase 17 详情](#phase-17-subagent-实现详细) |
| 18 | GitHub Release 准备 | ✅ | LICENSE, CI/CD, 自动化发布 | ~1,250 行 | [plans/18/](../plans/18-github-release-prep/) |
| 19 | 消息查询增强 | ✅ | Assistant 响应、对话分析 | ~600 行 | [plans/19/](../plans/19-message-query-enhancement/) |
| 20 | 插件打包与发布 | ✅ | 多平台包、自动安装脚本 | ~400 行 | [plans/20/](../plans/20-plugin-packaging/) |
| 21 | 自托管插件市场 | ✅ | 市场配置、一键安装 | ~200 行 | [plans/21/](../plans/21-self-hosted-marketplace/) |
| 22 | 统一 Meta 命令 | ✅ | 多源能力发现、语义匹配 | ~800 行 | [plans/22/](../plans/22-unified-meta-command/) |
| 23 | 查询能力函数库化 | ✅ | `internal/query` 库、MCP 完全去 CLI 依赖 | ~350 行 | [plans/23/](../plans/23-query-library/) |
| 24 | 统一查询接口设计与实现 | ⬜ | 单一 `query` 工具、资源导向、可组合过滤 | ~800 行 | 规划中 |

**注释**：
- **状态标识**：✅ 已完成，🟡 部分实现
- **代码量**：估算值，包含源码和测试
- Phase 7 集成到 Phase 8 的查询系统中
- Phase 10-11 核心功能已实现，部分高级特性待完善

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

---

## Phase 23: 查询能力函数库化（已完成）

**目标**：将查询逻辑抽象为可复用函数库，使 MCP 完全去除对 CLI 子进程的依赖，所有查询工具直接使用 `internal/query` 库。

**实际完成**：
- ✅ `internal/query` 库已建立，包含 12 个查询函数（RunToolsQuery, BuildAssistantMessages, BuildContextQuery 等）
- ✅ MCP 的 13 个查询工具全部迁移到使用库（query_tools, query_user_messages, query_assistant_messages, query_context, query_tool_sequences, query_file_access, query_files, query_conversation, get_session_stats, query_time_series, query_project_state, query_successful_prompts, query_tools_advanced）
- ✅ 删除所有 CLI 相关遗留代码：
  - 删除 `buildCommand()` 函数（17 行）
  - 删除 `toolCommandBuilders` 映射和 13 个 builder 函数（208 行）
  - 删除 `executeMetaCC()` 函数（72 行）
  - 删除 `scopeArgs()` 函数（9 行）
  - 删除 `ToolExecutor.metaCCPath` 字段
- ✅ 简化 ExecuteTool default 分支，移除 CLI fallback 逻辑
- ✅ 新增测试验证不调用 CLI（`executor_no_cli_test.go`，包含 3 个测试套件）
- ✅ 所有单元测试通过（`go test ./...`）

**代码变更统计**：
- 删除代码：~306 行（CLI 相关遗留代码）
- 新增代码：~190 行（测试代码）
- 净减少：~116 行

**完成标准**
- ✅ MCP 执行器不再调用 `executeMetaCC` 或 `buildCommand`
- ✅ 所有查询工具使用 `internal/query` 库
- ✅ 测试验证 MCP 不会尝试执行 CLI 二进制文件
- ✅ 所有现有测试通过（包括 `cmd/mcp-server` 测试套件）

**关键成果**：
1. **完全去除 CLI 依赖**：MCP 不再通过子进程调用 `meta-cc` 二进制文件
2. **简化架构**：所有查询逻辑统一在 `internal/query` 库中
3. **提升性能**：消除子进程创建开销和 JSONL 二次编解码
4. **提升可维护性**：减少代码重复，统一错误处理

详细计划见 [plans/23/](../plans/23-query-library/)

---

## Phase 24: 统一查询接口设计与实现（规划）

**目标**：基于实际 Claude Code JSONL schema，设计并实现统一的查询接口，将 16 个碎片化 MCP 工具简化为 1 个可组合的 `query` 工具。

**代码量**：~800 行（核心逻辑 + 测试 + 迁移支持）

### 核心设计理念

**从碎片化到统一**：
- **当前**：16 个独立工具（query_tools, query_user_messages, query_conversation...）
- **目标**：1 个统一 `query` 工具，通过参数化实现所有查询场景

**资源导向设计**：
- `entries`：原始 SessionEntry 流（完整信息）
- `messages`：消息视图（user/assistant 对话）
- `tools`：工具执行视图（tool_use + tool_result 配对）

**可组合查询管道**：
```
filter → transform → aggregate → output
```

### 关键设计决策

1. **Schema 统一为 snake_case**
   - 理由：与 Claude Code JSONL 源文件完全一致
   - 影响：需要修改 `ToolCall` 等结构体（破坏性变更）

2. **统一查询参数结构**
   - `resource`: 资源类型选择
   - `filter`: 结构化过滤条件
   - `transform`: 转换和分组逻辑
   - `aggregate`: 聚合函数
   - `output`: 输出控制

3. **向后兼容策略**
   - 保留旧工具作为 `query()` 的别名包装
   - 提供迁移指南和自动转换工具
   - 2-3 个版本的兼容期

### 阶段拆分

**Stage 24.1: Schema 标准化**
- 统一所有结构体 JSON 标签为 snake_case
- 修改 `ToolCall`, `SessionEntry`, `Message` 等结构
- 更新所有相关测试用例
- 验证与实际 JSONL 文件的完全一致性

**Stage 24.2: 统一查询接口实现**
- 在 `internal/query` 中实现 `QueryParams` 结构
- 实现资源选择器（entries/messages/tools）
- 实现过滤引擎（结构化 filter）
- 实现转换引擎（extract, group_by, join）
- 实现聚合引擎（count, sum, avg, etc.）

**Stage 24.3: MCP 工具重构**
- 在 `cmd/mcp-server` 中添加新的 `query` 工具
- 实现旧工具的别名包装（向后兼容）
- 更新工具定义和文档
- 添加查询示例和 cookbook

**Stage 24.4: 测试与验证**
- 单元测试：filter/transform/aggregate 逻辑
- 集成测试：完整查询场景
- 兼容性测试：旧工具别名功能
- 性能测试：大数据集查询

**Stage 24.5: 文档与迁移**
- 编写统一查询 API 文档
- 创建迁移指南（旧工具 → 新查询）
- 提供查询示例库
- 更新 CLAUDE.md 和 README.md

### 完成标准

- ⬜ Schema 完全统一为 snake_case
- ⬜ 统一 `query` 工具实现并通过测试
- ⬜ 所有旧工具功能可通过新接口实现
- ⬜ 向后兼容别名工作正常
- ⬜ 文档和迁移指南完整
- ⬜ 所有测试通过（包括回归测试）

### 预期收益

| 维度 | 改善 |
|-----|------|
| 工具数量 | 16 → 1 (94% 减少) |
| 参数数量 | 80+ → 20 (75% 减少) |
| 命名一致性 | 混乱 → 统一 |
| 可组合性 | 无 → 完整 |
| 学习成本 | 高 → 低 |
| 维护成本 | 高 → 低 |

### 风险管理

**风险 1: 破坏性变更**
- 缓解：提供别名包装，2-3 版本兼容期
- 缓解：提供迁移工具和详细文档

**风险 2: 性能回退**
- 缓解：查询计划优化
- 缓解：性能基准测试和对比

**风险 3: 学习曲线**
- 缓解：渐进式文档（简单到复杂）
- 缓解：丰富的示例库和 cookbook

详细设计见 `/tmp/unified_query_api_proposal.md`

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

**最后更新**：2025-10-13
**维护者**：meta-cc 开发团队
