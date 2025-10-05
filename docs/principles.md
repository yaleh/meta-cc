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

### 默认查询范围

- ✅ **默认查询范围为项目级**（所有会话）
- ✅ 工具名带 `_session` 后缀表示**仅查询当前会话**
- ✅ 保持 API 清晰：无后缀 = 项目级，`_session` = 会话级
- ✅ 利用 `--project .` 标志实现跨会话查询

**工具命名约定：**

| 项目级（默认） | 会话级 | 说明 |
|--------------|--------|------|
| `query_tools` | `query_tools_session` | 工具调用查询 |
| `analyze_errors` | `analyze_errors_session` | 错误分析 |
| `query_user_messages` | `query_user_messages_session` | 用户消息搜索 |
| `get_stats` | `get_session_stats` | 统计信息 |

**设计理由：**
- 元认知需要跨会话分析（发现长期模式）
- 用户期望 MCP 提供项目全局视角
- 会话级查询作为快速上下文检索的补充

---

## 五、职责分离：CLI vs Claude

### CLI 工具职责（无 LLM）

- ✅ JSONL 解析和数据提取
- ✅ 基于规则的模式检测（错误重复、工具频率）
- ✅ 结构化数据输出（JSONL/TSV）
- ✅ 索引构建和查询优化

### Claude 职责（在 Slash/Subagent/MCP 中）

- ✅ 语义理解和分析
- ✅ 建议生成和优先级判断
- ✅ 上下文关联和推理
- ✅ 与用户的对话式交互

### 分离的好处

1. **性能**：CLI 处理纯数据，速度快
2. **成本**：不为简单统计调用 LLM
3. **可测试性**：CLI 输出确定性，易于测试
4. **灵活性**：同一份数据，可被多个上层工具复用

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
