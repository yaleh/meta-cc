# Phase 27 E2E 测试要求更新摘要

**日期**: 2025-10-26
**更新**: 在 Phase 27 计划中补充 E2E 测试要求

---

## 更新内容

### 1. 新增 E2E 测试框架说明

**位置**: `docs/core/plan.md` - Phase 27 开头部分（第 372 行）

添加了完整的 E2E 测试框架说明，包括：

- **3 种测试方法**：
  - 直接 stdio 测试（快速验证）
  - 自动化测试脚本（推荐）
  - MCP Inspector（交互调试）

- **测试文档**（已创建）：
  - `docs/guides/mcp-e2e-testing.md` - 完整指南（13,000 字）
  - `docs/guides/mcp-testing-quickstart.md` - 快速参考
  - `docs/analysis/mcp-e2e-testing-recommendations.md` - 方法分析

- **集成方式**：
  - 每个 Stage 包含 E2E 测试命令
  - 集成到 Makefile（`make test-e2e-mcp`）
  - CI/CD pipeline 支持

---

### 2. Stage 27.1: 删除旧接口

**新增 E2E 测试验证**（第 723 行）：

```bash
# 验证工具已删除
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
  ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | \
  jq -e '.result.tools[] | select(.name == "query")' && \
  echo "❌ FAILED: query still exists" || echo "✓ query removed"

# 验证快捷工具仍可用
./tests/e2e/mcp-e2e-simple.sh
```

**验证点**：
- ✅ query 和 query_raw 工具不再可用
- ✅ 10 个快捷查询工具正常工作

---

### 3. Stage 27.2: get_session_directory

**新增 E2E 测试验证**（第 790 行）：

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

**验证点**：
- ✅ project 范围返回目录和统计信息
- ✅ session 范围返回当前会话目录
- ✅ 错误处理正确

---

### 4. Stage 27.3: inspect_session_files

**新增 E2E 测试验证**（第 916 行）：

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

**验证点**：
- ✅ 文件元数据正确（大小、行数、类型分布）
- ✅ 时间范围准确
- ✅ 样本收集功能正常

---

### 5. Stage 27.4: execute_stage2_query

**新增 E2E 测试验证**（第 1039 行）：

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

**验证点**：
- ✅ 基础过滤功能
- ✅ 排序功能
- ✅ 转换功能
- ✅ 性能 < 100ms（3MB 数据）

---

### 6. Stage 27.5: 文档和测试完善

**更新内容**（第 1141 行）：

**E2E 测试基础设施**（已完成）：
- ✅ `tests/e2e/mcp-e2e-simple.sh` - 自动化测试脚本
- ✅ `docs/guides/mcp-e2e-testing.md` - E2E 测试完整指南
- ✅ `docs/guides/mcp-testing-quickstart.md` - 快速参考
- ✅ `docs/analysis/mcp-e2e-testing-recommendations.md` - 测试方法分析

**E2E 测试扩展**（Stage 27.5 完成）：
- 更新测试脚本，添加 Phase 27 工具测试
- 添加性能基准测试
- 验证所有测试通过

**集成到 CI/CD**：
```makefile
# Makefile 新增 target
test-e2e-mcp: build
	@bash tests/e2e/mcp-e2e-simple.sh ./meta-cc-mcp

test-all: test test-e2e-mcp
	@echo "✅ All tests passed (unit + E2E)"
```

---

### 7. 更新完成标准

**位置**: 第 1165 行

添加了完整的 E2E 测试完成标准：

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
- ✅ **E2E 测试指南完整**
- ✅ 快速参考手册可用

---

## 更新统计

| 更新类型 | 位置 | 说明 |
|---------|------|------|
| **新增框架说明** | 第 372 行 | E2E 测试框架完整说明 |
| **Stage 27.1 测试** | 第 723 行 | 验证工具删除和兼容性 |
| **Stage 27.2 测试** | 第 790 行 | get_session_directory E2E 测试 |
| **Stage 27.3 测试** | 第 916 行 | inspect_session_files E2E 测试 |
| **Stage 27.4 测试** | 第 1039 行 | execute_stage2_query 完整测试 |
| **Stage 27.5 测试** | 第 1141 行 | 测试基础设施和扩展 |
| **完成标准** | 第 1211 行 | E2E 测试完成标准 |

**总计**：7 个主要更新，覆盖所有 5 个 Stage 和完成标准

---

## 测试覆盖

### 功能测试

- ✅ 工具列表验证（tools/list）
- ✅ 工具调用验证（tools/call）
- ✅ 参数验证（scope, files, filter, sort, transform, limit）
- ✅ 错误处理（无效工具名、无效参数）
- ✅ 返回值格式（JSON schema 验证）

### 性能测试

- ✅ 执行时间 < 100ms（3MB 数据）
- ✅ 内存使用 < 10MB
- ✅ 79x 加速验证（智能文件选择）

### 集成测试

- ✅ 两阶段工作流（get_session_directory → execute_stage2_query）
- ✅ 多工具协同（inspect → execute）
- ✅ 跨会话文件查询

### 回归测试

- ✅ 10 个快捷查询工具兼容性
- ✅ 破坏性变更验证（query/query_raw 删除）
- ✅ 向后兼容性检查

---

## 测试工具

### 已创建的资源

| 资源 | 类型 | 用途 |
|------|------|------|
| `tests/e2e/mcp-e2e-simple.sh` | 测试脚本 | 自动化 E2E 测试 |
| `docs/guides/mcp-e2e-testing.md` | 文档 | 完整测试指南（13,000 字）|
| `docs/guides/mcp-testing-quickstart.md` | 文档 | 快速参考手册 |
| `docs/analysis/mcp-e2e-testing-recommendations.md` | 分析 | 方法对比和建议 |

### 推荐工作流

**开发快速验证**：
```bash
make build && ./tests/e2e/mcp-e2e-simple.sh
```

**交互式调试**：
```bash
mcp-inspector ./meta-cc-mcp
```

**CI/CD 集成**：
```bash
make test-all  # 包含单元测试和 E2E 测试
```

---

## 下一步行动

### P0（实施 Phase 27 时）

1. ✅ **验证测试框架可用**:
   ```bash
   ./tests/e2e/mcp-e2e-simple.sh ./meta-cc-mcp
   ```

2. **每个 Stage 完成后运行 E2E 测试**:
   - Stage 27.1 完成 → 运行验证脚本
   - Stage 27.2 完成 → 测试 get_session_directory
   - Stage 27.3 完成 → 测试 inspect_session_files
   - Stage 27.4 完成 → 完整工作流测试
   - Stage 27.5 完成 → 更新测试脚本

3. **集成到 Makefile**:
   ```bash
   vim Makefile
   # 添加 test-e2e-mcp target
   ```

### P1（Phase 27 完成后）

1. **扩展测试用例**:
   - 添加边界条件测试
   - 添加并发测试
   - 添加大数据量测试（100MB+）

2. **集成到 CI**:
   - GitHub Actions workflow
   - 自动化回归测试
   - 性能趋势监控

3. **文档完善**:
   - 添加故障排查章节
   - 添加性能调优指南
   - 添加常见问题 FAQ

---

## 总结

✅ **Phase 27 现已包含完整的 E2E 测试要求**

- 📄 **7 个位置更新**：框架说明 + 5 个 Stage + 完成标准
- 🧪 **完整测试覆盖**：功能、性能、集成、回归
- 📚 **丰富文档支持**：完整指南、快速参考、方法分析
- 🔧 **现成工具**：自动化脚本、测试命令、Makefile 集成

**准备就绪**：可以立即开始 Phase 27 实施，每个 Stage 都有明确的 E2E 测试验证标准。
