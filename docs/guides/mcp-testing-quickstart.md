# MCP Testing Quick Start

快速参考：如何在不重启 Claude Code 的情况下测试新构建的 `meta-cc-mcp` binary。

## TL;DR - 最快验证方法

```bash
# 1. 构建
make build

# 2. 测试（选择一种）
./tests/e2e/mcp-e2e-simple.sh ./meta-cc-mcp              # 推荐：自动化测试
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
  ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | jq .           # 快速：单个命令
```

---

## 方法 1: 自动化测试脚本（推荐）

### 使用现有脚本

```bash
# 运行完整测试
./tests/e2e/mcp-e2e-simple.sh

# 测试特定 binary
./tests/e2e/mcp-e2e-simple.sh /path/to/meta-cc-mcp

# 开发循环
make build && ./tests/e2e/mcp-e2e-simple.sh
```

### 输出示例

```
==========================================
MCP Simple E2E Test
==========================================
Binary: ./meta-cc-mcp
==========================================

✓ PASSED - Found 15 tools
✓ PASSED - Tool executed successfully
⚠ get_session_directory not found (not implemented yet)
```

---

## 方法 2: 命令行直接测试

### 列出所有工具

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
  ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | jq .
```

### 调用特定工具

```bash
# query_tool_errors
echo '{
  "jsonrpc":"2.0",
  "id":2,
  "method":"tools/call",
  "params":{
    "name":"query_tool_errors",
    "arguments":{"limit":5}
  }
}' | ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | jq .

# query_tools
echo '{
  "jsonrpc":"2.0",
  "id":3,
  "method":"tools/call",
  "params":{
    "name":"query_tools",
    "arguments":{"limit":10}
  }
}' | ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | jq .

# query_user_messages
echo '{
  "jsonrpc":"2.0",
  "id":4,
  "method":"tools/call",
  "params":{
    "name":"query_user_messages",
    "arguments":{"pattern":".*","limit":5}
  }
}' | ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | jq .
```

### 测试 Phase 27 新工具（实现后）

```bash
# get_session_directory
echo '{
  "jsonrpc":"2.0",
  "id":5,
  "method":"tools/call",
  "params":{
    "name":"get_session_directory",
    "arguments":{"scope":"project"}
  }
}' | ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | jq .

# inspect_session_files
echo '{
  "jsonrpc":"2.0",
  "id":6,
  "method":"tools/call",
  "params":{
    "name":"inspect_session_files",
    "arguments":{
      "files":["/path/to/file1.jsonl","/path/to/file2.jsonl"],
      "include_samples":false
    }
  }
}' | ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | jq .

# execute_stage2_query
echo '{
  "jsonrpc":"2.0",
  "id":7,
  "method":"tools/call",
  "params":{
    "name":"execute_stage2_query",
    "arguments":{
      "files":["/path/to/file1.jsonl"],
      "filter":"select(.type == \"user\")",
      "sort":"sort_by(.timestamp)",
      "limit":10
    }
  }
}' | ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | jq .
```

---

## 方法 3: MCP Inspector（交互式）

### 安装

```bash
npm install -g @modelcontextprotocol/inspector
```

### 使用

```bash
# 启动 Inspector
mcp-inspector ./meta-cc-mcp

# 浏览器会自动打开 http://localhost:5173
# 在 UI 中：
# 1. 左侧面板显示所有工具
# 2. 点击工具查看参数 schema
# 3. 填写参数并执行
# 4. 右侧查看结果
```

### 优点
- ✅ 可视化界面
- ✅ 自动解析 schema
- ✅ 实时查看请求/响应
- ✅ 最适合探索和调试

---

## 常见测试场景

### 场景 1: 快速验证 binary 可运行

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
  ./meta-cc-mcp 2>&1 | grep -q '"result"' && echo "✓ OK" || echo "✗ FAILED"
```

### 场景 2: 验证工具数量

```bash
TOOL_COUNT=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
  ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | jq '.result.tools | length')
echo "Found $TOOL_COUNT tools"
```

### 场景 3: 检查特定工具是否存在

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
  ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | \
  jq -e '.result.tools[] | select(.name == "get_session_directory")' >/dev/null && \
  echo "✓ get_session_directory exists" || \
  echo "⚠ get_session_directory not found"
```

### 场景 4: 测试工具错误处理

```bash
# 无效的工具名
echo '{
  "jsonrpc":"2.0",
  "id":99,
  "method":"tools/call",
  "params":{
    "name":"nonexistent_tool",
    "arguments":{}
  }
}' | ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | jq '.error'
```

---

## 调试技巧

### 查看完整日志

```bash
# 保存日志到文件
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
  ./meta-cc-mcp 2>debug.log | jq .

# 查看日志
cat debug.log
```

### 过滤日志

```bash
# 只看 ERROR 级别日志
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
  ./meta-cc-mcp 2>&1 >/dev/null | grep ERROR
```

### 测试超时处理

```bash
# 3 秒超时
timeout 3 bash -c 'echo "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/list\"}" | ./meta-cc-mcp'
```

---

## 开发工作流

### 快速迭代循环

```bash
# 方式 1: 单行命令
while true; do
  make build && \
  echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
    ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | jq -e '.result' && \
  echo "✓ Build OK" || echo "✗ Build failed"
  sleep 2
done

# 方式 2: 使用测试脚本
while true; do
  make build && ./tests/e2e/mcp-e2e-simple.sh
  sleep 2
done
```

### 完整验证流程

```bash
# 1. 格式化代码
make fmt

# 2. 运行单元测试
make test

# 3. 构建
make build

# 4. E2E 测试
./tests/e2e/mcp-e2e-simple.sh

# 5. （可选）使用 Inspector 交互测试
mcp-inspector ./meta-cc-mcp
```

---

## 集成到 Makefile

在 `Makefile` 中添加：

```makefile
# E2E 测试
test-e2e-mcp: build
	@echo "Running MCP E2E tests..."
	@bash tests/e2e/mcp-e2e-simple.sh ./meta-cc-mcp

# 完整测试（单元 + E2E）
test-all: test test-e2e-mcp
	@echo "✅ All tests passed"

# 开发快速验证
dev-test: build
	@echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
	  ./meta-cc-mcp 2>&1 | grep '"jsonrpc"' | jq -e '.result' >/dev/null && \
	  echo "✅ MCP server OK" || echo "❌ MCP server failed"
```

使用：

```bash
make dev-test        # 快速验证
make test-e2e-mcp   # 完整 E2E 测试
make test-all       # 所有测试
```

---

## 常见问题

### Q: 为什么每次测试都要重启服务器？

**A**: 因为 MCP 服务器设计为长期运行的进程。每次 stdin 输入后，服务器会处理请求并返回响应，然后继续等待下一个请求。但在测试脚本中，我们发送单个请求后立即关闭 stdin，导致服务器退出。

这对于快速测试是可接受的。如果需要测试持久会话，使用 MCP Inspector。

### Q: 为什么需要 `grep '"jsonrpc"'`？

**A**: 因为 MCP 服务器会输出日志到 stderr（但由于某些原因可能混入 stdout）。`grep '"jsonrpc"'` 过滤出实际的 JSON-RPC 响应。

### Q: 如何模拟 Claude Code 的环境？

**A**: 设置环境变量：

```bash
export META_CC_PROJECTS_ROOT="$HOME/.claude/projects"
export META_CC_CWD="$PWD"
./meta-cc-mcp
```

### Q: 测试卡住了怎么办？

**A**: 使用 `timeout` 命令：

```bash
timeout 5 bash -c 'echo "{...}" | ./meta-cc-mcp'
```

---

## 下一步

1. **立即验证**:
   ```bash
   ./tests/e2e/mcp-e2e-simple.sh
   ```

2. **添加到 CI**:
   ```yaml
   # .github/workflows/test.yml
   - name: Run MCP E2E Tests
     run: make test-e2e-mcp
   ```

3. **安装 MCP Inspector**（可选）:
   ```bash
   npm install -g @modelcontextprotocol/inspector
   mcp-inspector ./meta-cc-mcp
   ```

---

## 参考资源

- [完整 E2E 测试指南](./mcp-e2e-testing.md)
- [MCP 协议规范](https://spec.modelcontextprotocol.io/)
- [MCP Inspector](https://github.com/modelcontextprotocol/inspector)
- [meta-cc MCP 工具参考](./mcp.md)
