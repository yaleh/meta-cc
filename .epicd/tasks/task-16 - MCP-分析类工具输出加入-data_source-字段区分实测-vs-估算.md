---
id: TASK-16
title: MCP 分析类工具输出加入 data_source 字段区分实测 vs 估算
status: 'Basic: Done'
assignee: []
created_date: '2026-06-24 12:13'
updated_date: '2026-06-24 12:37'
labels:
  - 'kind:basic'
dependencies: []
ordinal: 3000
---

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
# Plan: MCP 分析类工具输出加入 data_source 字段区分实测 vs 估算

## Context
meta-cc 的六个分析类 MCP 工具（analyze_errors、quality_scan、get_work_patterns、get_timeline、analyze_bugs、get_tech_debt）混合了直接从 session trace 统计的数据与基于启发式规则推断的数据，但调用方无法区分两者的可信度。加入 `data_source` 字段（取值 `measured` | `estimated`）实现 BAIME Layer 7 溯源标准，支持系统性自我观测偏差的检测（BAIME TASK-152 中 delta_H = -1.46 已观测到该问题）。

受影响的六个结果 struct 位于 `/home/yale/work/meta-cc/internal/analyzer/`：

| Struct | 文件 | 字段来源分析 |
|---|---|---|
| `BugAnalysisResult` | bugs_analysis.go | Patterns/TotalPairs 来自 toolCalls 直接计数 → **measured** |
| `QualityScanResult` | quality_analysis.go | 四个 Dimension 均为 toolCalls 原始计数推出的比率 → **measured** |
| `TechDebtResult` | tech_debt.go | Markers/HotspotFiles 来自扫描 tool output → **measured**；OpenIssues 基于「error 后无 success」启发式 → **estimated** |
| `ErrorAnalysisResult` | errors_analysis.go | TotalErrors/ByTool/ByType 来自直接计数 → **measured** |
| `WorkPatternsResult` | work_patterns.go | ToolFrequency/HourlyActivity/PeakHour 直接计数 → **measured**；ContextSwitches 基于话题切换启发式 → **estimated** |
| `TimelineResult` | timeline.go | Events 来自 session entries 直接解析 → **measured** |

由于同一个 struct 内部存在混合（TechDebtResult.OpenIssues 是 estimated，WorkPatternsResult.ContextSwitches 是 estimated，其余字段是 measured），需要在 **顶层 struct** 加一个 `DataSource` 字段标记整体主导来源，并在注释中标注混合情况。

## Phase 1: 添加 DataSource 类型和单元测试桩

在 `/home/yale/work/meta-cc/internal/analyzer/` 目录新增一个共享文件 `data_source.go`，定义：
- `type DataSource string`
- `const DataSourceMeasured DataSource = "measured"`
- `const DataSourceEstimated DataSource = "estimated"`

同时在六个对应的 `_test.go` 文件中，为 `DataSource` 字段添加断言桩（先写 failing test，供 Phase 2 实现时通过）。

Run `make dev` 验证编译通过（test 可以 fail）。

### DoD
- [ ] `grep -q 'DataSourceMeasured' /home/yale/work/meta-cc/internal/analyzer/data_source.go`
- [ ] `make -C /home/yale/work/meta-cc dev`

## Phase 2: 在六个 Result struct 添加 DataSource 字段并赋值

编辑以下六个文件，在顶层 Result struct 添加 `DataSource DataSource \`json:"data_source"\`` 字段，并在 return 语句中赋值：

1. `bugs_analysis.go` → `DataSourceMeasured`（所有字段直接计数）
2. `quality_analysis.go` → `DataSourceMeasured`（所有比率来自计数）
3. `tech_debt.go` → `DataSourceMeasured`（Markers/HotspotFiles 直接扫描；OpenIssues 基于启发式，但整体主要来自 measured，注释标注 OpenIssues 为 estimated）
4. `errors_analysis.go` → `DataSourceMeasured`
5. `work_patterns.go` → `DataSourceMeasured`（ContextSwitches 启发式，注释标注）
6. `timeline.go` → `DataSourceMeasured`

Run `make commit` 验证测试通过。

### DoD
- [ ] `grep -rq 'DataSourceMeasured\|DataSourceEstimated' /home/yale/work/meta-cc/internal/analyzer/bugs_analysis.go`
- [ ] `grep -rq 'data_source' /home/yale/work/meta-cc/internal/analyzer/bugs_analysis.go`
- [ ] `make -C /home/yale/work/meta-cc commit`

## Phase 3: 更新 MCP 工具文档

编辑 `/home/yale/work/meta-cc/docs/guides/mcp.md`，为受影响的六个工具说明添加 `data_source` 字段的含义，并新增一节 `## Data Source Provenance (BAIME Layer 7)` 解释 `measured` vs `estimated` 的区别以及混合 struct 中的注意事项。

### DoD
- [ ] `grep -q 'data_source' /home/yale/work/meta-cc/docs/guides/mcp.md`
- [ ] `grep -q 'measured' /home/yale/work/meta-cc/docs/guides/mcp.md`
- [ ] `grep -q 'Data Source Provenance' /home/yale/work/meta-cc/docs/guides/mcp.md`

## Constraints
- 每个 Phase 代码修改 ≤200 行
- TDD：Phase 1 先写测试桩，Phase 2 实现令其通过
- 只修改分析类 Result struct，不修改 query_session_content 等查询 struct
- 不改变 MCP 工具名称或参数 schema

## Acceptance Gate
- [ ] `grep -rq 'data_source' /home/yale/work/meta-cc/internal/analyzer`
- [ ] `make -C /home/yale/work/meta-cc commit`
- [ ] `grep -q 'data_source' /home/yale/work/meta-cc/docs/guides/mcp.md`
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
cap:propose=approved

claimed: 2026-06-24T12:29:49Z

Phase 1 ✓ 2026-06-24T00:00:00Z — Created data_source.go with DataSource type and DataSourceMeasured/DataSourceEstimated constants. Added failing test stubs to all 6 test files. make dev passes.

DoD #1: PASS — grep -q 'DataSourceMeasured' data_source.go

DoD #2: PASS — make dev

Phase 2 ✓ 2026-06-24T00:00:00Z — Added DataSource field to all 6 Result structs (BugAnalysisResult, QualityScanResult, TechDebtResult, ErrorAnalysisResult, WorkPatternsResult, TimelineResult). All assigned DataSourceMeasured at return sites. Mixed-provenance structs (TechDebtResult.OpenIssues, WorkPatternsResult.ContextSwitches) annotated in comments. make commit PASS.

DoD #3: PASS — grep -rq 'DataSourceMeasured' bugs_analysis.go

DoD #4: PASS — grep -rq 'data_source' bugs_analysis.go

DoD #5: PASS — make commit

Phase 3 ✓ 2026-06-24T00:00:00Z — Updated docs/guides/mcp.md: added per-tool data_source table after Analysis Tools section, added 'Data Source Provenance (BAIME Layer 7)' section explaining measured vs estimated and documenting mixed-provenance fields.

DoD #6: PASS — grep -q 'data_source' docs/guides/mcp.md

DoD #7: PASS — grep -q 'measured' docs/guides/mcp.md

DoD #8: PASS — grep -q 'Data Source Provenance' docs/guides/mcp.md

## Execution Summary
Result: Done
Commit: 4099d49

workerLoop DoD #1: PASS — grep -rq 'data_source' /home/yale/work/meta-cc/internal/analyzer

workerLoop DoD #2: PASS — make -C /home/yale/work/meta-cc commit

workerLoop DoD #3: PASS — grep -q 'data_source' /home/yale/work/meta-cc/docs/guides/mcp.md

WARNING: agent-summary missing

Completed: 2026-06-24T12:37:31Z
<!-- SECTION:NOTES:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 grep -rq 'data_source' /home/yale/work/meta-cc/internal/analyzer
- [ ] #2 make -C /home/yale/work/meta-cc commit
- [ ] #3 grep -q 'data_source' /home/yale/work/meta-cc/docs/guides/mcp.md
<!-- DOD:END -->
