# MCP Server E2E Testing Guide

## 概述

本文档介绍如何在**不重启 Claude Code** 的情况下，对新构建的 `meta-cc-mcp` binary 进行端到端（E2E）测试。

## 背景

- MCP 服务器使用 **stdio 协议**（stdin 接收请求，stdout 返回响应）
- Claude Code 启动时会自动启动配置的 MCP 服务器进程
- 测试新 binary 无需重启 Claude Code，可以直接在命令行测试

---

## 方案对比

| 方案 | 复杂度 | 灵活性 | 依赖 | 推荐场景 |
|------|--------|--------|------|----------|
| 1. 直接 stdio 测试 | ⭐ | ⭐⭐ | 无 | 快速验证单个工具 |
| 2. Bash 测试脚本 | ⭐⭐ | ⭐⭐⭐ | jq | 自动化回归测试 |
| 3. MCP Inspector | ⭐⭐⭐ | ⭐⭐⭐⭐ | Node.js | 交互式调试和探索 |
| 4. Node.js SDK 客户端 | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Node.js + SDK | 复杂场景和集成测试 |

---

## 方案 1: 直接 stdio 测试（最简单）

### 原理

直接通过 stdin 向 MCP binary 发送 JSON-RPC 请求，从 stdout 读取响应。

### 使用方法

#### 1.1 列出所有工具

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./meta-cc-mcp | jq .
```

**预期输出**：
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "tools": [
      {
        "name": "query_user_messages",
        "description": "...",
        "inputSchema": {...}
      },
      ...
    ]
  }
}
```

#### 1.2 调用工具（示例：query_user_messages）

```bash
echo '{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "query_user_messages",
    "arguments": {
      "pattern": "test",
      "limit": 5
    }
  }
}' | ./meta-cc-mcp | jq .
```

#### 1.3 测试 get_session_directory（新工具）

```bash
echo '{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "get_session_directory",
    "arguments": {
      "scope": "project"
    }
  }
}' | ./meta-cc-mcp | jq .
```

#### 1.4 测试 execute_stage2_query（新工具）

```bash
# 获取最近 3 个会话文件
FILES=($(ls -t ~/.claude/projects/-home-*/*.jsonl | head -3))

# 构造文件列表 JSON
FILE_JSON=$(printf '"%s",' "${FILES[@]}" | sed 's/,$//')

echo "{
  \"jsonrpc\": \"2.0\",
  \"id\": 4,
  \"method\": \"tools/call\",
  \"params\": {
    \"name\": \"execute_stage2_query\",
    \"arguments\": {
      \"files\": [$FILE_JSON],
      \"filter\": \"select(.type == \\\"user\\\")\",
      \"sort\": \"sort_by(.timestamp)\",
      \"limit\": 10
    }
  }
}" | ./meta-cc-mcp | jq .
```

### 优点
- ✅ 零依赖（只需 jq 格式化输出）
- ✅ 快速验证
- ✅ 适合快速迭代

### 缺点
- ❌ 手动构造 JSON 容易出错
- ❌ 不适合复杂测试场景
- ❌ 无法测试交互式场景

---

## 方案 2: Bash 测试脚本（推荐用于 CI）

### 创建测试脚本

创建 `tests/e2e/mcp-e2e-test.sh`:

```bash
#!/bin/bash
# E2E test suite for meta-cc-mcp server
# Usage: ./tests/e2e/mcp-e2e-test.sh [binary_path]

set -e

BINARY="${1:-./meta-cc-mcp}"
TESTS_PASSED=0
TESTS_FAILED=0

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Helper: Send JSON-RPC request
send_request() {
    local request="$1"
    echo "$request" | "$BINARY" 2>/dev/null
}

# Helper: Test tool call
test_tool_call() {
    local test_name="$1"
    local tool_name="$2"
    local arguments="$3"
    local expected_pattern="$4"

    echo -n "Testing $test_name... "

    local request=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "id": $RANDOM,
  "method": "tools/call",
  "params": {
    "name": "$tool_name",
    "arguments": $arguments
  }
}
EOF
)

    local response=$(send_request "$request")

    if echo "$response" | jq -e '.error' >/dev/null 2>&1; then
        echo -e "${RED}FAILED${NC}"
        echo "  Error: $(echo "$response" | jq -r '.error.message')"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi

    if [ -n "$expected_pattern" ]; then
        if ! echo "$response" | grep -q "$expected_pattern"; then
            echo -e "${RED}FAILED${NC}"
            echo "  Expected pattern not found: $expected_pattern"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            return 1
        fi
    fi

    echo -e "${GREEN}PASSED${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

# Test suite
echo "=========================================="
echo "MCP E2E Test Suite"
echo "Binary: $BINARY"
echo "=========================================="
echo ""

# Test 1: tools/list
echo "=== Phase 1: Tool Discovery ==="
TOOLS_LIST=$(send_request '{"jsonrpc":"2.0","id":1,"method":"tools/list"}')
TOOL_COUNT=$(echo "$TOOLS_LIST" | jq '.result.tools | length')
echo "Available tools: $TOOL_COUNT"
echo ""

# Test 2: Legacy tools (backward compatibility)
echo "=== Phase 2: Legacy Tools ==="
test_tool_call "query_user_messages" "query_user_messages" '{"pattern":".*","limit":5}' "result"
test_tool_call "query_tools" "query_tools" '{"limit":5}' "result"
test_tool_call "query_tool_errors" "query_tool_errors" '{"limit":5}' "result"
echo ""

# Test 3: New tools (Phase 27)
echo "=== Phase 3: Phase 27 New Tools ==="
test_tool_call "get_session_directory" "get_session_directory" '{"scope":"project"}' "directory"
echo ""

# Test 4: Stage 2 query (if implemented)
if echo "$TOOLS_LIST" | jq -e '.result.tools[] | select(.name == "execute_stage2_query")' >/dev/null 2>&1; then
    echo "=== Phase 4: Stage 2 Query ==="

    # Get recent session files
    SESSION_DIR=$(echo '{"jsonrpc":"2.0","id":100,"method":"tools/call","params":{"name":"get_session_directory","arguments":{"scope":"project"}}}' | \
        "$BINARY" 2>/dev/null | jq -r '.result.content[0].text | fromjson | .directory')

    if [ -n "$SESSION_DIR" ] && [ -d "$SESSION_DIR" ]; then
        FILES=$(ls -t "$SESSION_DIR"/*.jsonl 2>/dev/null | head -3 | jq -R . | jq -s .)

        STAGE2_ARGS=$(cat <<EOF
{
  "files": $FILES,
  "filter": "select(.type == \\"user\\")",
  "sort": "sort_by(.timestamp)",
  "limit": 5
}
EOF
)
        test_tool_call "execute_stage2_query" "execute_stage2_query" "$STAGE2_ARGS" "results"
    else
        echo -e "${YELLOW}SKIPPED${NC} (no session directory found)"
    fi
    echo ""
fi

# Summary
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed: ${RED}$TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
```

### 使用方法

```bash
# 测试默认 binary（当前目录）
./tests/e2e/mcp-e2e-test.sh

# 测试指定 binary
./tests/e2e/mcp-e2e-test.sh /path/to/meta-cc-mcp

# 测试新构建的 binary
make build && ./tests/e2e/mcp-e2e-test.sh ./meta-cc-mcp
```

### 集成到 Makefile

在 `Makefile` 中添加：

```makefile
test-e2e-mcp: build
	@echo "Running MCP E2E tests..."
	@bash tests/e2e/mcp-e2e-test.sh ./meta-cc-mcp

test-all: test test-e2e-mcp
	@echo "✅ All tests passed (unit + E2E)"
```

### 优点
- ✅ 自动化测试套件
- ✅ 可集成到 CI/CD
- ✅ 易于扩展新测试用例
- ✅ 提供清晰的测试报告

### 缺点
- ❌ 需要维护测试脚本
- ❌ 复杂 JSON 构造较繁琐

---

## 方案 3: MCP Inspector（推荐用于交互式调试）

### 安装

```bash
npm install -g @modelcontextprotocol/inspector
```

### 使用方法

#### 3.1 启动 Inspector

```bash
# 方式 1: 使用当前目录 binary
mcp-inspector ./meta-cc-mcp

# 方式 2: 使用绝对路径
mcp-inspector /home/yale/work/meta-cc/meta-cc-mcp
```

#### 3.2 在浏览器中测试

1. Inspector 会启动本地 Web 服务器（通常是 `http://localhost:5173`）
2. 浏览器会自动打开交互式界面
3. 左侧面板显示所有可用工具
4. 点击工具可查看参数 schema
5. 填写参数并执行，右侧显示结果

#### 3.3 测试 Phase 27 新工具

在 Inspector UI 中：

1. **测试 get_session_directory**:
   - 选择 `get_session_directory` 工具
   - 设置 `scope: "project"`
   - 点击 Execute
   - 查看返回的目录路径和文件统计

2. **测试 inspect_session_files**:
   - 使用上一步获取的目录路径
   - 选择 `inspect_session_files` 工具
   - 填写文件路径列表
   - 查看文件元数据（类型分布、时间范围等）

3. **测试 execute_stage2_query**:
   - 选择 `execute_stage2_query` 工具
   - 填写参数：
     ```json
     {
       "files": ["/path/to/file1.jsonl", "/path/to/file2.jsonl"],
       "filter": "select(.type == \"user\")",
       "sort": "sort_by(.timestamp)",
       "limit": 10
     }
     ```
   - 执行并查看结果

### 优点
- ✅ 交互式界面，直观易用
- ✅ 自动解析工具 schema
- ✅ 实时查看请求/响应
- ✅ 支持历史记录
- ✅ 非常适合调试和探索

### 缺点
- ❌ 需要安装 Node.js 和 npm
- ❌ 不适合自动化测试

---

## 方案 4: Node.js MCP SDK 客户端（最灵活）

### 安装依赖

```bash
mkdir -p tests/mcp-client
cd tests/mcp-client
npm init -y
npm install @modelcontextprotocol/sdk
```

### 创建测试客户端

创建 `tests/mcp-client/test-client.js`:

```javascript
#!/usr/bin/env node

import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';

async function main() {
  const binaryPath = process.argv[2] || './meta-cc-mcp';

  console.log('Starting MCP client...');
  console.log(`Binary: ${binaryPath}\n`);

  // Create transport
  const transport = new StdioClientTransport({
    command: binaryPath,
    args: []
  });

  // Create client
  const client = new Client(
    {
      name: 'meta-cc-test-client',
      version: '1.0.0'
    },
    {
      capabilities: {}
    }
  );

  await client.connect(transport);

  console.log('✓ Connected to MCP server\n');

  // Test 1: List tools
  console.log('=== Test 1: List Tools ===');
  const tools = await client.listTools();
  console.log(`Found ${tools.tools.length} tools:`);
  tools.tools.forEach(tool => {
    console.log(`  - ${tool.name}`);
  });
  console.log('');

  // Test 2: Call get_session_directory
  console.log('=== Test 2: Get Session Directory ===');
  try {
    const result = await client.callTool({
      name: 'get_session_directory',
      arguments: { scope: 'project' }
    });
    console.log('Result:', JSON.stringify(result, null, 2));
  } catch (error) {
    console.error('Error:', error.message);
  }
  console.log('');

  // Test 3: Call query_user_messages
  console.log('=== Test 3: Query User Messages ===');
  try {
    const result = await client.callTool({
      name: 'query_user_messages',
      arguments: { pattern: '.*', limit: 5 }
    });
    const data = JSON.parse(result.content[0].text);
    console.log(`Found ${data.length} messages`);
  } catch (error) {
    console.error('Error:', error.message);
  }
  console.log('');

  // Test 4: Call execute_stage2_query (if implemented)
  console.log('=== Test 4: Execute Stage 2 Query ===');
  const hasStage2 = tools.tools.some(t => t.name === 'execute_stage2_query');

  if (hasStage2) {
    try {
      // First get directory
      const dirResult = await client.callTool({
        name: 'get_session_directory',
        arguments: { scope: 'project' }
      });
      const dirData = JSON.parse(dirResult.content[0].text);

      // Get recent files (mock - in real scenario use fs)
      const files = [
        dirData.directory + '/session1.jsonl',
        dirData.directory + '/session2.jsonl'
      ];

      const queryResult = await client.callTool({
        name: 'execute_stage2_query',
        arguments: {
          files: files,
          filter: 'select(.type == "user")',
          sort: 'sort_by(.timestamp)',
          limit: 5
        }
      });

      const queryData = JSON.parse(queryResult.content[0].text);
      console.log(`Query returned ${queryData.results.length} results`);
      console.log(`Execution time: ${queryData.metadata.execution_time_ms}ms`);
    } catch (error) {
      console.error('Error:', error.message);
    }
  } else {
    console.log('execute_stage2_query not implemented yet');
  }
  console.log('');

  // Cleanup
  await client.close();
  console.log('✓ Tests completed');
}

main().catch(console.error);
```

### 使用方法

```bash
# 安装依赖（一次性）
cd tests/mcp-client && npm install && cd ../..

# 运行测试
node tests/mcp-client/test-client.js

# 测试特定 binary
node tests/mcp-client/test-client.js /path/to/meta-cc-mcp
```

### 优点
- ✅ 功能最强大
- ✅ 可编写复杂测试逻辑
- ✅ 易于集成到 CI/CD
- ✅ 支持异步和错误处理
- ✅ 可复用测试代码

### 缺点
- ❌ 需要 Node.js 环境
- ❌ 代码量较大
- ❌ 需要维护 npm 依赖

---

## 实践建议

### 快速验证流程

```bash
# 1. 构建新 binary
make build

# 2. 快速验证（方案 1）
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./meta-cc-mcp | jq '.result.tools | length'

# 3. 完整 E2E 测试（方案 2）
./tests/e2e/mcp-e2e-test.sh ./meta-cc-mcp

# 4. 交互式调试（方案 3，可选）
mcp-inspector ./meta-cc-mcp
```

### 开发迭代流程

```bash
# 快速开发循环
while true; do
    # 修改代码
    vim cmd/mcp-server/handlers_stage2.go

    # 构建 + 快速验证
    make build && echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./meta-cc-mcp | jq -e '.result'

    # 如果通过，运行完整测试
    if [ $? -eq 0 ]; then
        ./tests/e2e/mcp-e2e-test.sh
    fi
done
```

### CI/CD 集成

在 `.github/workflows/test.yml` 中添加：

```yaml
- name: Run MCP E2E Tests
  run: |
    make build
    bash tests/e2e/mcp-e2e-test.sh ./meta-cc-mcp
```

---

## 常见问题

### Q1: 测试时如何模拟真实的 Claude Code 环境？

**A**: 设置环境变量：

```bash
export META_CC_PROJECTS_ROOT="$HOME/.claude/projects"
export META_CC_CWD="/home/yale/work/meta-cc"
./meta-cc-mcp
```

### Q2: 如何测试需要当前会话的工具？

**A**: 使用 `META_CC_SESSION_FILE` 环境变量：

```bash
export META_CC_SESSION_FILE="$HOME/.claude/projects/-home-yale-work-meta-cc/latest.jsonl"
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_session_stats","arguments":{}}}' | ./meta-cc-mcp
```

### Q3: 如何调试 MCP 服务器的日志？

**A**: 日志输出到 stderr，不影响 JSON-RPC 通信：

```bash
# 查看日志
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./meta-cc-mcp 2>debug.log | jq .

# 实时查看日志
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./meta-cc-mcp 2>&1 >/dev/null | grep -v "^{"
```

### Q4: 为什么我的测试卡住了？

**A**: MCP 服务器等待换行符（`\n`）。确保使用 `echo` 而不是 `echo -n`：

```bash
# ✓ 正确
echo '{"jsonrpc":"2.0",...}' | ./meta-cc-mcp

# ✗ 错误（会卡住）
echo -n '{"jsonrpc":"2.0",...}' | ./meta-cc-mcp
```

---

## 推荐工作流

### 场景 1: 快速验证单个修改

→ **使用方案 1（直接 stdio 测试）**

```bash
make build && echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"NEW_TOOL","arguments":{}}}' | ./meta-cc-mcp | jq .
```

### 场景 2: 完整的回归测试

→ **使用方案 2（Bash 测试脚本）**

```bash
make build && ./tests/e2e/mcp-e2e-test.sh
```

### 场景 3: 调试复杂问题

→ **使用方案 3（MCP Inspector）**

```bash
mcp-inspector ./meta-cc-mcp
# 在浏览器中交互式调试
```

### 场景 4: CI/CD 自动化

→ **使用方案 2 + 4（Bash + Node.js SDK）**

```bash
# Makefile
test-e2e-full: test-e2e-mcp test-e2e-node
	@echo "✅ All E2E tests passed"

test-e2e-mcp:
	@bash tests/e2e/mcp-e2e-test.sh

test-e2e-node:
	@node tests/mcp-client/test-client.js
```

---

## 下一步

1. **创建基础测试脚本**（方案 2）
   - 复制上述 `mcp-e2e-test.sh` 到 `tests/e2e/`
   - 运行验证现有工具

2. **添加 Phase 27 测试用例**
   - 测试 `get_session_directory`
   - 测试 `inspect_session_files`
   - 测试 `execute_stage2_query`

3. **集成到 CI**
   - 添加 `make test-e2e-mcp` target
   - 在 GitHub Actions 中运行

4. **（可选）安装 MCP Inspector**
   - 用于交互式调试
   - 探索新工具行为

---

## 参考资源

- [MCP Protocol Specification](https://spec.modelcontextprotocol.io/)
- [MCP SDK Documentation](https://github.com/modelcontextprotocol/sdk)
- [MCP Inspector Repository](https://github.com/modelcontextprotocol/inspector)
- [meta-cc MCP Guide](./mcp.md)
