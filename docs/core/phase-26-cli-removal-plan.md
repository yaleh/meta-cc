# Phase 26: CLI 代码清理（MCP 独立化）

## 概述

**目标**：移除过时的 CLI 相关代码，实现 MCP-only 架构，简化项目结构和维护负担。

**动机**：
- MCP server 已实现为独立可执行文件（`meta-cc-mcp`），提供完整功能
- CLI 命令（`meta-cc`）已过时，功能重复且增加维护成本
- 简化架构可提升代码质量和可维护性
- 减少构建复杂度和部署包大小

**状态**：📋 计划阶段（待确认）

---

## 当前状态分析

### MCP Server 依赖（必须保留）

通过依赖分析，MCP server 仅依赖以下 internal 包：
```
internal/config     - 配置管理
internal/errors     - 错误处理
internal/locator    - 会话文件定位
internal/query      - 查询引擎（核心）
```

### CLI 相关文件（待移除）

**cmd/ 目录下的 CLI 命令文件**（~50 个文件，~15,000 行代码）：
```
cmd/root.go                          - CLI 主命令
cmd/parse.go                         - parse 命令
cmd/analyze*.go                      - analyze 命令族
cmd/query*.go                        - query 命令族（CLI 版本）
cmd/stats*.go                        - stats 命令族
cmd/validate.go                      - validate 命令
cmd/pipeline.go                      - CLI pipeline 模式
cmd/*_test.go                        - CLI 测试文件
cmd/*_integration_test.go            - CLI 集成测试
```

**Makefile 目标**：
```
build-cli            - CLI 构建目标
BINARY_NAME          - meta-cc 二进制名称
相关的交叉编译配置
```

**文档**：
```
docs/reference/cli.md                - CLI 参考文档
docs/tutorials/examples.md           - CLI 使用示例（部分）
CLAUDE.md                            - CLI 相关说明（部分）
README.md                            - CLI 安装说明（部分）
```

### 共享代码（保留并审查）

**internal/ 包**（需要逐一审查依赖）：
```
internal/parser      - JSONL 解析器（MCP 依赖）
internal/analyzer    - 分析器（可能仅 CLI 使用）
internal/stats       - 统计计算（可能仅 CLI 使用）
internal/filter      - 过滤器（可能仅 CLI 使用）
internal/output      - 输出格式化（可能仅 CLI 使用）
internal/validation  - 验证逻辑（可能仅 CLI 使用）
```

**需要确认的问题**：
1. `internal/analyzer`, `internal/stats` 等包是否被 MCP server 间接依赖？
2. 是否有测试文件共享工具函数需要迁移？

---

## 实施计划

### Stage 26.1: 依赖分析与验证

**目标**：明确 MCP server 的完整依赖树，确定可安全删除的代码范围

**任务**：
1. **完整依赖树分析**
   ```bash
   # 分析 MCP server 的所有依赖（直接和传递）
   cd cmd/mcp-server
   go list -f '{{.Deps}}' . | tr ' ' '\n' | grep meta-cc/internal
   ```

2. **识别孤立包**（仅被 CLI 使用）
   - 使用 `go mod why` 确认每个 internal 包的使用者
   - 标记仅被 CLI 使用的包

3. **测试覆盖率分析**
   - 识别共享测试工具（`testutil` 等）
   - 确保 MCP server 测试不依赖 CLI 测试

**交付物**：
- `docs/architecture/cli-removal-dependency-analysis.md` - 依赖分析报告
- 待删除文件清单（按类别分组）
- 待保留文件清单（带保留原因）

**验证标准**：
- ✅ 所有 internal 包的依赖关系已明确
- ✅ 已识别所有孤立包（零依赖或仅 CLI 依赖）
- ✅ MCP server 的最小依赖集已确定

**估算代码量**：~50 行（分析脚本 + 文档）

---

### Stage 26.2: 移除 CLI 命令文件

**目标**：删除 `cmd/` 目录下的所有 CLI 相关文件，保留 `cmd/mcp-server/`

**任务**：
1. **删除 CLI 命令文件**
   ```bash
   # 删除 CLI 主命令和子命令
   rm cmd/root.go cmd/parse.go cmd/analyze*.go
   rm cmd/query*.go cmd/stats*.go cmd/validate.go
   rm cmd/pipeline.go cmd/fixtures_helpers_test.go
   ```

2. **删除 CLI 测试文件**
   ```bash
   rm cmd/*_test.go cmd/*_integration_test.go
   # 保留 cmd/mcp-server/ 下的所有测试
   ```

3. **删除 CLI 入口点**
   - 删除 `main.go`（如果存在）或确认唯一入口点是 `cmd/mcp-server/main.go`

**交付物**：
- 清理后的 `cmd/` 目录（仅包含 `cmd/mcp-server/`）

**验证标准**：
- ✅ `cmd/` 目录仅包含 `mcp-server/` 子目录
- ✅ 所有 CLI 相关的 `.go` 文件已删除
- ✅ `make build-mcp` 仍然成功构建
- ✅ MCP server 测试全部通过

**估算代码量**：-15,000 行（删除）

---

### Stage 26.3: 清理孤立的 internal 包

**目标**：删除仅被 CLI 使用的 internal 包

**任务**：
1. **删除孤立包**（基于 Stage 26.1 的分析）
   - 候选包：`internal/analyzer`, `internal/stats`, `internal/filter`, `internal/output`, `internal/validation`
   - 仅在确认零依赖后删除

2. **迁移共享工具**
   - 如果 `internal/testutil` 被 MCP 测试使用，保留
   - 如果仅被 CLI 测试使用，删除

3. **更新 import 路径**
   - 确保没有孤立的 import 语句

**交付物**：
- 清理后的 `internal/` 目录（仅保留 MCP 依赖的包）

**验证标准**：
- ✅ `go mod tidy` 成功
- ✅ `make test-all` 全部通过（MCP 测试）
- ✅ 无未使用的 internal 包

**估算代码量**：-5,000 行（估算，取决于孤立包数量）

---

### Stage 26.4: 更新构建脚本

**目标**：简化 Makefile，移除 CLI 相关构建目标

**任务**：
1. **删除 CLI 构建目标**
   ```makefile
   # 删除
   build-cli:
   BINARY_NAME := meta-cc

   # 保留
   build-mcp:
   MCP_BINARY_NAME := meta-cc-mcp
   ```

2. **简化主构建目标**
   ```makefile
   # 修改前
   build: build-cli build-mcp

   # 修改后
   build: build-mcp
   ```

3. **更新交叉编译**
   - 仅构建 `meta-cc-mcp` 二进制文件
   - 移除 `meta-cc` 二进制的平台构建

4. **更新 bundle-release**
   - 仅打包 `meta-cc-mcp` 二进制
   - 移除 CLI 二进制引用

**交付物**：
- 简化的 `Makefile`

**验证标准**：
- ✅ `make build` 仅构建 `meta-cc-mcp`
- ✅ `make cross-compile` 成功（仅 MCP server）
- ✅ `make bundle-release` 成功（仅包含 MCP server）
- ✅ 所有构建目标正常工作

**估算代码量**：-50 行（Makefile）

---

### Stage 26.5: 更新文档

**目标**：移除 CLI 相关文档，更新架构说明，反映 MCP-only 架构

**任务**：
1. **删除 CLI 参考文档**
   ```bash
   rm docs/reference/cli.md
   ```

2. **更新核心文档**
   - `README.md`: 移除 CLI 安装说明，强调 MCP-only
   - `CLAUDE.md`: 移除 CLI 使用说明
   - `docs/core/plan.md`: 添加 Phase 26 完成记录
   - `docs/core/principles.md`: 更新架构描述（如有 CLI 引用）

3. **更新集成指南**
   - `docs/guides/integration.md`: 移除 CLI 集成方式
   - `docs/guides/mcp.md`: 强调为唯一集成方式

4. **更新教程**
   - `docs/tutorials/examples.md`: 移除 CLI 示例，仅保留 MCP 示例

5. **更新架构文档**
   - `docs/architecture/proposals/meta-cognition-proposal.md`: 更新架构图（移除 CLI 层）

**交付物**：
- 更新后的文档集（反映 MCP-only 架构）
- 迁移指南（可选）：`docs/guides/migration-from-cli.md`

**验证标准**：
- ✅ 所有文档链接有效（无 404）
- ✅ 无 CLI 相关的过时说明
- ✅ README 和 CLAUDE.md 正确反映 MCP-only 架构
- ✅ 快速开始指南可用（MCP 安装和使用）

**估算代码量**：~500 行（文档更新 + 新迁移指南）

---

### Stage 26.6: CI/CD 和发布更新

**目标**：更新 CI/CD 流程，移除 CLI 构建和测试

**任务**：
1. **更新 GitHub Actions**
   - `.github/workflows/ci.yml`: 移除 CLI 测试步骤
   - 仅保留 MCP server 构建和测试

2. **更新发布流程**
   - `scripts/release.sh`: 移除 CLI 二进制打包
   - 仅发布 `meta-cc-mcp` 二进制

3. **更新插件元数据**
   - `.claude-plugin/marketplace.json`: 确认无 CLI 引用

**交付物**：
- 简化的 CI/CD 配置

**验证标准**：
- ✅ CI 流程成功（仅 MCP 测试）
- ✅ 发布脚本成功（仅 MCP 二进制）
- ✅ 插件元数据正确

**估算代码量**：~50 行（配置更新）

---

## 完成标准

### 功能验证
- ✅ `make build` 成功构建 `meta-cc-mcp` 二进制
- ✅ `make test-all` 所有测试通过（MCP server 测试）
- ✅ `make lint` 无错误
- ✅ `make cross-compile` 成功构建所有平台的 MCP 二进制
- ✅ MCP server 在 Claude Code 中正常工作（16 个工具可用）

### 代码质量
- ✅ `go mod tidy` 无变更（依赖已清理）
- ✅ 测试覆盖率 ≥80%（MCP server）
- ✅ 无未使用的 internal 包
- ✅ 无孤立的 import 语句

### 架构验证
- ✅ `cmd/` 目录仅包含 `mcp-server/` 子目录
- ✅ `internal/` 目录仅包含 MCP 依赖的包
- ✅ 项目根目录无 CLI 相关文件（如 `main.go`）

### 文档完整性
- ✅ 所有文档链接有效
- ✅ README.md 正确描述 MCP-only 架构
- ✅ 快速开始指南可用（MCP 安装）
- ✅ 无 CLI 相关的过时说明

### 发布验证
- ✅ CI/CD 流程成功
- ✅ 发布包仅包含 MCP 二进制
- ✅ 插件安装和更新正常

---

## 影响评估

### 正面影响
1. **代码库简化**：减少 ~20,000 行代码（CLI 命令 + 孤立包 + 测试）
2. **维护负担降低**：单一架构（MCP-only），减少重复功能维护
3. **构建时间缩短**：仅构建 MCP server，减少 CI/CD 时间
4. **部署简化**：单一二进制文件（`meta-cc-mcp`），无需选择
5. **文档一致性**：单一集成方式，降低用户困惑

### 潜在风险
1. **用户迁移**：
   - **风险**：现有 CLI 用户需要迁移到 MCP
   - **缓解**：提供迁移指南，MCP 功能完全覆盖 CLI
   - **评估**：低风险（CLI 使用量极低，MCP 为主要使用方式）

2. **测试覆盖率下降**：
   - **风险**：删除 CLI 测试可能降低整体覆盖率
   - **缓解**：确保 MCP 测试覆盖所有核心功能
   - **评估**：低风险（MCP 测试已覆盖核心查询功能）

3. **依赖分析错误**：
   - **风险**：误删 MCP 依赖的包导致构建失败
   - **缓解**：Stage 26.1 严格依赖分析，逐步验证
   - **评估**：低风险（依赖树清晰，`go list` 验证）

### 回滚计划
如果发现重大问题，可以通过以下方式回滚：
1. 恢复 `cmd/` 目录的 CLI 文件（从 git history）
2. 恢复 `internal/` 孤立包
3. 恢复 Makefile 的 CLI 构建目标
4. 恢复文档

**回滚成本**：1-2 小时（git revert + 测试）

---

## 时间估算

| Stage | 描述 | 预估时间 |
|-------|------|---------|
| 26.1 | 依赖分析与验证 | 2 小时 |
| 26.2 | 移除 CLI 命令文件 | 1 小时 |
| 26.3 | 清理孤立包 | 2 小时 |
| 26.4 | 更新构建脚本 | 1 小时 |
| 26.5 | 更新文档 | 3 小时 |
| 26.6 | CI/CD 和发布更新 | 1 小时 |
| **总计** | | **10 小时** |

**代码变更量**：
- 删除：~20,000 行（CLI 代码 + 孤立包）
- 新增：~500 行（文档 + 配置）
- 净减少：~19,500 行

---

## 后续优化（未来 Phase）

完成 Phase 26 后，可考虑的进一步优化：

1. **internal 包重组**（Phase 27？）
   - 将 `internal/query` 拆分为更小的子包
   - 改进包边界和职责划分

2. **MCP server 性能优化**（Phase 28？）
   - 会话缓存优化
   - 查询执行并行化

3. **插件生态扩展**（Phase 29？）
   - 更多 capabilities
   - 第三方插件支持

---

## 参考资料

### 内部文档
- [设计原则](./principles.md) - 架构决策依据
- [实施计划](./plan.md) - 整体 Phase 规划
- [MCP 指南](../guides/mcp.md) - MCP server 完整参考

### 相关 Phase
- [Phase 14: 架构重构](../../plans/14-architecture-refactor/) - MCP 独立化起点
- [Phase 23-25: 查询接口重构](../archive/phase-23-25-query-refactoring.md) - 查询层简化

### TODO
- [TODO.md](../../TODO.md) - 原始需求描述

---

**最后更新**：2025-10-25
**维护者**：meta-cc 开发团队
**状态**：📋 计划阶段，等待确认
