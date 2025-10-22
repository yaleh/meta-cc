# Phase 23: 查询能力函数库化（Query Pipeline Library Extraction）

## 概述

**目标**：抽取 `cmd/query_*` 及 MCP 执行路径中的查询逻辑，构建可复用的 `internal/query` 函数库，供 CLI 与 MCP 共享，以消除子进程调用与重复 JSONL 解析。

**预估代码量**：~550 行（实现 ~320，测试 ~230），单文件/单 Stage 均控制在 200 行以内。

**主要依赖**：
- `pkg/pipeline/session.go` 已提供会话定位 + 解析能力，可直接在库层复用。
- `internal/filter`、`pkg/output` 提供稳定的过滤/输出工具。
- `cmd/mcp-server/executor.go` 当前通过子进程调用 CLI，改造后需转调库函数。

**核心交付物**：
- 新建 `internal/query` 包，暴露 `RunToolsQuery`, `RunUserMessagesQuery`, `RunSessionStats` 等高层 API。
- CLI 查询命令改用库函数，行为完全保持。
- MCP 执行器改为库调用，复用共享的 jq/输出路径。
- 行为回归测试（CLI + MCP）覆盖主要查询路径。
- 更新开发文档，指向新的函数入口。

---

## 当前痛点

1. **跨进程复制开销**：`cmd/mcp-server/executor.go:90-155` 启动 `meta-cc` 子进程，重复解析 JSONL。
2. **逻辑分散**：过滤、排序、分页逻辑散落在 `cmd/query_tools.go:52-193`、`cmd/query_messages.go:1-112` 等文件，难以复用。
3. **测试成本高**：MCP 行为依赖 CLI 测试，用例跨层、可控性差。
4. **文档难维护**：CLI 和 MCP 各自描述查询流程，导致差异风险。

---

## Phase 目标

1. **统一查询调用**：构建共享函数库供 CLI/MCP 调用。
2. **保持现有语义**：所有现有命令和工具的标志位、输出格式保持不变。
3. **提高可测性**：新增针对库函数的单元/集成测试，并串联 CLI/MCP 端到端用例。
4. **为 Phase 24 铺路**：在新接口中引入统一的 `QueryOptions` 草案（Scope/Pagination/JQ 等），便于下阶段参数简化。

---

## 交付范围

| 模块 | 范围 | 说明 |
|------|------|------|
| `internal/query` | ✅ 新增 | 封装工具、消息、统计查询公共逻辑 |
| CLI (`cmd/query_*`) | ✅ 调整 | 改为调用 `internal/query`，保留 flag 解析 |
| MCP (`cmd/mcp-server`) | ✅ 调整 | 执行器调用库函数；保留混合输出、临时文件处理 |
| 文档 | ✅ 更新 | `docs/core/plan.md`、开发指南、测试说明 |
| Tests | ✅ 新增 | Library 单测、CLI/MCP 回归测试 |
| CLI Flags | 🔄 兼容 | 保持现有输入，打印兼容警告（若需要） |
| Phase 24 参数重构 | ❌ 不含 | 仅铺垫接口，不改动参数形态 |

---

## 成功标准

- ✅ CLI 与 MCP 查询路径均调用同一库函数。
- ✅ 主要命令（tools, user-messages, session stats）输出经回归测试验证，与当前版本二进制比对一致。
- ✅ 新增库函数测试覆盖率 ≥ 85%。
- ✅ 文档明确指出复用方式与后续 Phase 24 的衔接点。
- ✅ MCP 不再依赖 `exec.Command("meta-cc")`，`cmd/mcp-server/executor.go` 中子进程调用被移除。

---

## 验收清单

1. 代码层面
   - [ ] `internal/query` 包已创建，含入口函数与内部 `QueryOptions` 草案。
   - [ ] CLI 与 MCP 调用改造完成，旧代码删除或标记遗留。
   - [ ] 关键查询路径单元测试、回归测试通过。

2. 文档层面
   - [ ] `docs/core/plan.md` Phase 23 状态更新。
   - [ ] 开发者指南中新增库调用示例。
   - [ ] Release note 草稿说明内部改造（若需要）。

3. 质量衡量
   - [ ] `make test`、`make test-all`、Phase 23 新增测试脚本均通过。
   - [ ] MCP 端集成测试（`cmd/mcp-server` 下现有测试）仍然通过。

---

## 里程碑

| 里程碑 | 说明 | 预计时间 |
|--------|------|----------|
| M23.1 | `internal/query` 包骨架 + CLI 接入 | 1.5d |
| M23.2 | MCP 执行器切换 + 辅助组件抽离 | 1.5d |
| M23.3 | 回归测试 + 文档更新 | 1d |

---

## 风险与应对

| 风险 | 等级 | 缓解措施 |
|------|------|----------|
| CLI/MCP 输出差异 | 中 | 使用黄金文件 / snapshot 测试对比改造前后输出 |
| 共享函数耦合过高 | 中 | 通过 `QueryOptions` 分层（输入、过滤、输出），并保持函数粒度清晰 |
| 迁移造成性能回退 | 低 | 保留 `SessionPipeline` 缓存机制，避免重复解析 |
| Phase 24 调整冲突 | 低 | 在接口中预留扩展字段，保持兼容 |

---

## 下一步

- 根据本计划编写 `iteration-23-implementation-plan.md`（详述 Stage 23.1-23.3）。
- 评估所需的黄金文件/fixture 更新或新增。
- 与 Phase 24 负责人确认 `QueryOptions` 草案字段，以减少反复改动。
