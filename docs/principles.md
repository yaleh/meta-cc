# meta-cc 设计原则与核心约束

本文档定义了 meta-cc 项目的核心设计原则、开发约束和架构决策。

> **注意**：详细的架构决策已迁移至 [Architecture Decision Records (ADRs)](adr/README.md)。本文档保留核心原则和快速参考。

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

> **详细决策**：参见 [ADR-001: 两层架构设计](adr/ADR-001-two-layer-architecture.md)

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

**查询结果排序：**
- `query tools` → 按 Timestamp 排序
- `query messages` → 按 turn_sequence 排序
- `query errors` → 按 Timestamp 排序

**代码内部确定性：**
- ✅ 对 map 键进行排序后再迭代（避免测试间歇性失败）
- ✅ 固定随机种子（如需随机数生成）
- ✅ 避免依赖时间戳顺序（使用固定时间或 mock）

**示例：**
```go
// Bad (非确定性)
for key, value := range params {
    args = append(args, fmt.Sprintf("--%s=%s", key, value))
}

// Good (确定性)
keys := make([]string, 0, len(params))
for k := range params { keys = append(keys, k) }
sort.Strings(keys)
for _, k := range keys {
    args = append(args, fmt.Sprintf("--%s=%s", k, params[k]))
}
```

### 4. 延迟决策

将分析决策推给下游工具/LLM，meta-cc 只提供原始数据。

**职责边界：**
- ❌ meta-cc 不应实现：窗口过滤、错误聚合、模式计数
- ✅ 交给 jq/awk：`meta-cc query errors | jq '.[length-50:]'`
- ✅ 交给 Claude：Slash Commands 从 JSONL 生成语义建议

---

## 三、跨平台兼容性原则

### 路径处理规范

**ALWAYS：**
- ✅ 使用 `os.TempDir()` 替代硬编码路径（`/tmp/`, `C:\Temp\`）
- ✅ 使用 `filepath.Join()` 构建跨平台路径
- ✅ 使用 `filepath.ToSlash()` 将路径转换为 JSON 安全格式
- ✅ 在测试中使用临时目录而非硬编码路径

**NEVER：**
- ❌ 硬编码 Unix 路径（`/home/`, `/var/`, `/tmp/`）
- ❌ 硬编码 Windows 路径（`C:\Users\`, `D:\Data\`）
- ❌ 在 JSON 字符串中直接使用 Windows 反斜杠路径（`\U` 是非法转义）

**示例：**
```go
// Bad
testPath := "/tmp/test-data.json"
jsonData := fmt.Sprintf(`{"path": "%s"}`, "C:\Users\test")

// Good
testPath := filepath.Join(os.TempDir(), "test-data.json")
jsonPath := filepath.ToSlash(testPath)  // C:/Users/test
jsonData := fmt.Sprintf(`{"path": "%s"}`, jsonPath)
```

### Windows 特殊处理

**文件锁定：**
- ✅ Windows 要求文件在 `os.Rename()` 前必须关闭
- ✅ 避免使用 `defer file.Close()`，改为显式关闭后 rename

**示例：**
```go
// Bad (Windows 失败)
file, _ := os.Create(tmpPath)
defer file.Close()
file.Write(data)
os.Rename(tmpPath, finalPath)  // Error: file locked

// Good (跨平台兼容)
file, _ := os.Create(tmpPath)
file.Write(data)
if err := file.Close(); err != nil { return err }
os.Rename(tmpPath, finalPath)  // Success
```

### CI/CD 跨平台测试

**测试矩阵：**
- ✅ 在 GitHub Actions 中测试 Linux、macOS、Windows
- ✅ 使用 `testing.Short()` 跳过平台特定测试
- ✅ 使用 `shell: bash` 指令确保 Windows 运行 bash 脚本

**示例：**
```go
func TestBashFeature(t *testing.T) {
    if testing.Short() || runtime.GOOS == "windows" {
        t.Skip("skipping bash-dependent test on Windows")
    }
    // Bash-specific test
}
```

---

## 四、错误处理与代码质量原则

### 强制错误检查

**ALWAYS：**
- ✅ 检查所有 `os.*` 函数的错误返回值（Chdir, Rename, Close, Remove 等）
- ✅ 检查所有 `flag.FlagSet.Set()` 调用的错误
- ✅ 检查 deferred 函数的错误（使用闭包捕获）

**示例：**
```go
// Bad
defer os.Chdir(originalDir)
flagSet.Set("output", "json")

// Good
defer func() {
    if err := os.Chdir(originalDir); err != nil {
        t.Errorf("failed to restore directory: %v", err)
    }
}()
if err := flagSet.Set("output", "json"); err != nil {
    return fmt.Errorf("failed to set flag: %w", err)
}
```

### Linting 纪律

**开发流程：**
- ✅ 每个 Stage 完成后运行 `make lint`（不仅仅在 Phase 结束时）
- ✅ 立即修复 linting 问题（避免技术债务累积）
- ✅ 使用 golangci-lint 进行全面静态分析

**CI/CD 验证：**
- ✅ GitHub Actions 在所有平台运行 linting
- ✅ Linting 失败应阻止 PR 合并

---

## 五、输出格式设计原则

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

## 六、MCP Server 设计原则

> **详细决策**：
> - [ADR-003: MCP Server 集成策略](adr/ADR-003-mcp-server-integration.md)
> - [ADR-004: 混合输出模式设计](adr/ADR-004-hybrid-output-mode.md)
> - [ADR-005: 作用域参数标准化](adr/ADR-005-scope-parameter-standardization.md)

### 架构分离：CLI vs MCP

**meta-cc CLI**（核心数据层）：
- ✅ 职责：JSONL 解析、数据提取、模式检测
- ✅ 输出：JSONL 格式（原始数据，无过滤）
- ✅ 不做：查询过滤、聚合统计、语义分析

**meta-cc-mcp**（MCP Server 层）：
- ✅ 职责：接收 MCP 请求，调用 CLI，使用 gojq 过滤/聚合
- ✅ 处理流程：CLI 获取数据 → gojq 过滤 → 统计聚合 → 混合输出模式
- ✅ 混合输出：输出 ≤ 阈值(8KB) → inline；输出 > 阈值 → file_ref

### 查询范围与输出控制（快速参考）

**查询范围**：
- ✅ 默认：`scope: "project"` （所有会话）
- ✅ 显式覆盖：`scope: "session"` （仅当前会话）

**结果数量限制**：
- ✅ 默认：无限制（依赖混合输出模式）
- ✅ 显式限制：`limit: N` （仅在明确需要时使用）

**混合输出模式**：
- ✅ 自动选择：≤8KB → inline；>8KB → file_ref
- ✅ 阈值配置：`inline_threshold_bytes` 或 `META_CC_INLINE_THRESHOLD`

---

## 七、职责分离与集成层次

> **详细决策**：参见 [ADR-003: MCP Server 集成策略](adr/ADR-003-mcp-server-integration.md)

### 集成层次概览

**1. MCP Server（核心层，80% 使用场景）**
- meta-cc-mcp 作为主要接入点
- Claude 自主决策何时调用
- 支持 jq 表达式过滤和统计

**2. Subagents（语义层，5% 使用场景）**
- @meta-coach：综合分析、语义理解、建议生成
- @error-analyst：错误模式分析、根本原因诊断
- @workflow-tuner：工作流优化、自动化建议

**3. Slash Commands（快捷层，15% 使用场景）**
- 固定格式的快速报告
- 适合重复性查询

---

## 八、测试策略

### 单元测试
- 每个 Stage 必须有对应的单元测试
- 测试覆盖率目标：≥ 80%
- 使用 `go test ./...` 运行所有测试

### 集成测试

**测试分类：**
- ✅ **单元测试**：无外部依赖，所有环境运行
- ✅ **集成测试**：需要 git 历史、Claude 会话目录等，使用 `testing.Short()` 标记

**CI/CD 策略：**
- ✅ CI 运行：`go test -short ./...`（跳过集成测试）
- ✅ 本地开发：`go test ./...`（运行所有测试）
- ✅ 集成测试必须提供清晰的 skip 消息

**示例：**
```go
func TestParseExtractCommand_RealSession(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test - requires Claude session directory")
    }
    // Integration test requiring ~/.claude/projects/
}
```

**Makefile 集成：**
```makefile
test:
	go test -short ./...

test-all:
	go test ./...
```

### 测试数据管理
- 测试 fixture 存放在 `tests/fixtures/`
- 使用真实的（脱敏的）Claude Code 会话文件
- 包含多种场景：正常会话、错误重复、工具密集使用等

---

## 九、会话定位策略

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

## 十、打包与发布原则（Phase 20）

> **详细决策**：参见 [ADR-002: 插件目录结构重构](adr/ADR-002-plugin-directory-structure.md)

### 插件结构（快速参考）

**开发**：编辑 `.claude/commands/` 和 `.claude/agents/` （Git 跟踪）
**发布**：运行 `make sync-plugin-files` 同步到 `commands/` 和 `agents/` （Git 忽略）

**关键命令**：
```bash
# 本地开发（无需构建）
vi .claude/commands/meta-stats.md

# 发布前同步
make sync-plugin-files

# 创建发布包
make bundle-release VERSION=v1.0.0
```

### 版本管理

- **语义化版本**：`MAJOR.MINOR.PATCH`（遵循 SemVer）
- **Git Tags**：`v1.0.0` 触发 Release 工作流
- **CHANGELOG**：每个 Release 包含变更摘要

---

## 十一、消息查询接口设计原则（Phase 19）

### 接口分层策略

**1. 专用接口**（细粒度，性能优化）：
- `query_user_messages`：仅查询用户消息，性能最优
- `query_assistant_messages`：仅查询 assistant 响应，支持响应长度过滤

**2. 统一接口**（粗粒度，关联分析）：
- `query_conversation`：查询完整对话（user + assistant），自动关联，包含响应时间

### 设计原则

1. **向后兼容**：保留现有接口，不破坏已有工具
2. **职责清晰**：3 个工具各司其职
3. **延迟决策**：提供完整对话数据，分析交给 Claude/jq
4. **性能优化**：专用接口避免加载无关数据

---

## 十二、风险和缓解措施

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Claude Code 会话文件格式变化 | 高 | 使用真实文件测试，版本化 fixture |
| 跨平台兼容性问题 | 高 | CI 矩阵测试（Linux/macOS/Windows），遵循跨平台原则 |
| 环境变量不可用 | 中 | 提供多种定位方式（参数、路径推断） |
| 测试覆盖不足 | 中 | TDD 强制要求，每个 Stage 先写测试 |
| 代码量超标 | 低 | 每个 Stage 结束检查行数，及时拆分 |
| Claude Code 集成失败 | 高 | 在测试环境充分验证 |

---

## 参考文档

### 架构决策记录（ADRs）
- [ADR 索引](adr/README.md) - 所有架构决策记录
- [ADR-001: 两层架构设计](adr/ADR-001-two-layer-architecture.md)
- [ADR-002: 插件目录结构重构](adr/ADR-002-plugin-directory-structure.md)
- [ADR-003: MCP Server 集成策略](adr/ADR-003-mcp-server-integration.md)
- [ADR-004: 混合输出模式设计](adr/ADR-004-hybrid-output-mode.md)
- [ADR-005: 作用域参数标准化](adr/ADR-005-scope-parameter-standardization.md)

### 项目文档
- [meta-cc 项目总体实施计划](./plan.md)
- [Claude Code 元认知分析系统 - 技术方案](./proposals/meta-cognition-proposal.md)
- [Claude Code 官方文档](https://docs.claude.com/en/docs/claude-code/overview)
