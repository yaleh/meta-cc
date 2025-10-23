# Phase 24: 统一查询接口设计与实现

## 快速概览

**目标**: 将 16 个碎片化 MCP 工具简化为 1 个可组合的统一查询工具

**核心设计**:
```
16 个独立工具 → 1 个 query 工具
资源导向: entries → messages → tools
可组合管道: filter → transform → aggregate → output
Schema 统一: 100% snake_case
```

---

## 文档索引

- **[完整实施计划](iteration-24-implementation-plan.md)** - 详细的 Stage 划分和实施步骤
- **[设计参考](../../tmp/unified_query_api_proposal.md)** - 统一查询 API 设计提案
- **[Schema 分析](../../tmp/corrected_schema_comparison_report.md)** - 当前 schema 对比报告

---

## 关键指标

| 维度 | 当前 | 目标 | 改善 |
|-----|------|------|------|
| MCP 工具数量 | 16 个 | 1 个 | **94% 减少** |
| 参数总数 | ~80 个 | ~20 个 | **75% 减少** |
| 命名风格 | 混乱 | 统一 | **100% 一致** |
| 可组合性 | 无 | 完整 | **质的飞跃** |
| 学习曲线 | 高 | 低 | **显著改善** |

---

## Stage 划分

### Stage 24.1: Schema 标准化
- **目标**: 统一所有数据结构为 snake_case
- **工期**: 5 小时
- **代码量**: 150 行 + 50 行测试

### Stage 24.2: 统一查询接口实现
- **目标**: 实现核心 Query() 函数和过滤/聚合引擎
- **工期**: 10 小时
- **代码量**: 450 行 + 320 行测试

### Stage 24.3: MCP 工具重构
- **目标**: 使用统一接口重构 16 个 MCP 工具
- **工期**: 7 小时
- **代码量**: 350 行 + 100 行测试

### Stage 24.4: 测试与验证
- **目标**: 全面测试和性能验证
- **工期**: 7 小时
- **代码量**: 0 行 + 300 行测试

### Stage 24.5: 文档与迁移
- **目标**: 编写文档和迁移指南
- **工期**: 6 小时
- **代码量**: 500 行文档

---

## 快速开始

### 当前问题

**查询失败的 Read 工具调用**（旧方式）：
```javascript
query_tools({
  tool: "Read",
  status: "error"
})
```

**统一接口**（新方式）：
```javascript
query({
  resource: "tools",
  filter: {
    tool_name: "Read",
    tool_status: "error"
  }
})
```

### 核心设计

**资源类型**:
- `entries`: 原始 SessionEntry 流
- `messages`: user/assistant 消息视图
- `tools`: tool_use + tool_result 配对视图

**查询管道**:
```
SessionEntry[]
  → filter(过滤条件)
  → transform(转换/分组/提取)
  → aggregate(聚合函数)
  → Result
```

---

## 风险与缓解

### 风险 1: 破坏性变更
- **缓解**: 保留旧工具作为别名，2-3 版本兼容期
- **状态**: ✅ 已规划

### 风险 2: 性能回退
- **缓解**: 基准测试，查询优化，缓存机制
- **目标**: ≤10% 性能回退

### 风险 3: 学习曲线
- **缓解**: 渐进式教程，交互式示例，友好错误提示
- **目标**: 5-20 分钟学习时间

### 风险 4: Schema 不一致
- **缓解**: Schema 验证测试，输出对比测试
- **目标**: 100% snake_case 一致性

---

## 验收标准

### 功能验收
- ✅ 统一 query 工具实现完成
- ✅ 3 种资源类型正确提取
- ✅ 过滤/聚合引擎功能完整
- ✅ 16 个 MCP 工具成功迁移

### 质量验收
- ✅ 测试覆盖率 ≥80%
- ✅ 所有 make test 通过
- ✅ 所有 make lint 通过
- ✅ 性能无显著回退

### 文档验收
- ✅ API 参考文档完整
- ✅ 迁移指南清晰
- ✅ 10+ 示例覆盖常见场景

---

## 时间线

| 阶段 | 时间 | 状态 |
|-----|------|------|
| Phase 规划 | Week 0 | ✅ 完成 |
| Stage 24.1-24.2 | Week 1 | ⬜ 待开始 |
| Stage 24.3 | Week 2 | ⬜ 待开始 |
| Stage 24.4-24.5 | Week 3 | ⬜ 待开始 |
| 真实项目验证 | Week 3 | ⬜ 待开始 |
| v2.0.0 发布 | Week 4 | ⬜ 待开始 |

**预计完成日期**: 2025-11-15

---

## 代码量汇总

| Stage | 代码 | 测试 | 文档 | 总计 |
|-------|------|------|------|------|
| 24.1 | 150 | 50 | 0 | 200 |
| 24.2 | 450 | 320 | 0 | 770 |
| 24.3 | 350 | 100 | 0 | 450 |
| 24.4 | 0 | 300 | 0 | 300 |
| 24.5 | 0 | 0 | ~500 | 500 |
| **总计** | **950** | **770** | **500** | **2220** |

**测试覆盖率**: 81% (770 / 950)

---

## 示例查询

### 示例 1: 基础过滤

**需求**: 查询所有失败的 Read 工具调用

```javascript
query({
  resource: "tools",
  filter: {
    tool_name: "Read",
    tool_status: "error"
  }
})
```

---

### 示例 2: 聚合查询

**需求**: 统计每个工具的调用次数

```javascript
query({
  resource: "tools",
  aggregate: {
    function: "count",
    field: "tool_name"
  }
})
```

**输出**:
```jsonl
{"tool_name": "Read", "count": 1523}
{"tool_name": "Edit", "count": 892}
{"tool_name": "Bash", "count": 445}
```

---

### 示例 3: 时间范围查询

**需求**: 查询最近 24 小时的失败工具调用

```javascript
query({
  resource: "tools",
  filter: {
    tool_status: "error",
    time_range: {
      start: "2025-10-22T00:00:00Z"
    }
  }
})
```

---

### 示例 4: 复杂组合查询

**需求**: 分析每个 Git 分支上失败的文件操作

```javascript
query({
  resource: "tools",
  filter: {
    tool_name: "Read|Edit|Write",
    tool_status: "error"
  },
  transform: {
    extract: ["git_branch", "tool_name", "input.file_path"],
    group_by: "git_branch"
  },
  aggregate: {
    function: "count",
    field: "tool_name"
  },
  output: {
    sort_by: "count",
    sort_order: "desc"
  }
})
```

**输出**:
```jsonl
{"git_branch": "feature/phase-14", "tool_name": "Read", "count": 234}
{"git_branch": "feature/phase-14", "tool_name": "Edit", "count": 123}
{"git_branch": "main", "tool_name": "Read", "count": 45}
```

---

## 迁移路径

### 旧工具 → 统一查询映射

| 旧工具 | 统一查询等价 |
|--------|-------------|
| `query_tools` | `query({resource: "tools", filter: {...}})` |
| `query_user_messages` | `query({resource: "messages", filter: {role: "user"}})` |
| `query_assistant_messages` | `query({resource: "messages", filter: {role: "assistant"}})` |
| `query_conversation` | `query({resource: "messages", transform: {group_by: "parent_uuid"}})` |
| `query_files` | `query({resource: "tools", aggregate: {function: "count"}})` |

### 向后兼容策略

**v2.0.0** (当前):
- 引入 `query` 统一接口
- 保留所有旧工具（别名包装）
- 无 breaking changes

**v2.1.0** (+3 个月):
- 旧工具标记为 deprecated
- 显示迁移提示

**v3.0.0** (+6 个月):
- 移除旧工具
- 仅保留 `query` 统一接口

---

## 参考资料

### 设计文档
- [统一查询 API 提案](../../tmp/unified_query_api_proposal.md)
- [Schema 对比报告](../../tmp/corrected_schema_comparison_report.md)
- [设计原则](../../docs/core/principles.md)

### 代码参考
- [当前 ToolCall 实现](../../internal/parser/tools.go)
- [查询函数库](../../internal/query/)
- [MCP 工具定义](../../cmd/mcp-server/tools.go)

### 类似设计
- GraphQL: 统一查询接口
- OData: URL 参数化查询
- Elasticsearch: 结构化查询 DSL

---

## 下一步行动

1. **Review 本计划** - 确认技术方案和时间线
2. **创建分支** - `git checkout -b feature/unified-query-api`
3. **开始 Stage 24.1** - Schema 标准化
4. **每个 Stage 运行** - `make all` 验证质量

---

## 联系方式

- **负责人**: Yale (yaleh@github)
- **项目**: meta-cc
- **Phase**: 24
- **计划文档**: `/home/yale/work/meta-cc/plans/iteration_24/`

---

**最后更新**: 2025-10-23
