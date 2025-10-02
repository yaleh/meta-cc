# Claude Code 元认知分析系统 - 技术方案

## 一、系统概述

### 1.1 核心定位

基于 `~/.claude/projects/` 会话历史的命令行分析工具，通过多维度索引、智能查询和模式识别，为 Claude Code 提供元认知能力。

### 1.2 核心组件

```
cc-meta CLI 工具
  ├─ 索引引擎：会话历史快速检索
  ├─ 查询引擎：多维度数据查询
  └─ 分析引擎：模式识别和建议生成

Claude Code 集成
  ├─ Slash Commands：快速操作
  ├─ Subagent：对话式分析
  └─ MCP Server：程序化访问
```

---

## 二、核心工具：cc-meta CLI

### 2.1 设计原则

**职责边界**
- CLI 工具：纯数据处理，无 LLM 调用
- 输出：结构化、高密度的分析结果
- 语义理解：由调用方（Slash Command/Subagent/MCP）中的 Claude 完成

**会话定位**
```bash
# 方式1: 通过环境变量（Claude Code 自动设置）
export CC_SESSION_ID="5b57148c-89dc-4eb5-bc37-8122e194d90d"
export CC_PROJECT_HASH="-home-yale-work-myproject"
cc-meta query turns

# 方式2: 通过命令行参数
cc-meta query turns --session 5b57148c-89dc-4eb5-bc37-8122e194d90d

# 方式3: 通过项目路径自动推断
cc-meta query turns --project /home/yale/work/myproject
# → 自动查找 ~/.claude/projects/-home-yale-work-myproject/
```

**会话文件结构**
```
~/.claude/projects/
  └─ -home-yale-work-myproject/     # 项目哈希目录
      ├─ 5b57148c-...-8122e194d90d.jsonl  # 会话1
      ├─ f1547628-...-9935-79c27ca6cc7e.jsonl  # 会话2
      └─ ...
```

### 2.2 命令结构

```bash
cc-meta - Claude Code Meta-Cognition Tool

全局选项:
  --session <id>          会话ID（或读取 $CC_SESSION_ID）
  --project <path>        项目路径（自动转换为哈希目录）
  --output <json|md|csv>  输出格式（默认：json）

COMMANDS:
  parse       解析会话文件（核心功能）
    dump      导出完整 JSONL 为结构化格式
    extract   提取特定数据（turns/tools/errors）
    stats     会话统计信息

  query       数据查询（需先建索引，可选）
    sessions  列出项目下所有会话
    turns     查询轮次
    tools     工具使用统计

  analyze     模式分析（基于规则，无 LLM）
    errors    错误模式检测
    tools     工具使用模式
    timeline  时间线分析
```

### 2.3 核心命令示例

**阶段1: 无索引，纯解析**
```bash
# 导出当前会话的所有 turns（供 Claude 分析）
cc-meta parse extract --type turns --format json

# 提取所有工具调用
cc-meta parse extract --type tools --filter "status=error"

# 生成会话统计摘要
cc-meta parse stats --metrics "tools,errors,duration"
```

**阶段2: 有索引，高级查询**
```bash
# 查询最近的 Bash 工具使用
cc-meta query tools --name Bash --limit 10

# 分析错误重复模式
cc-meta analyze errors --window 20 --threshold 3

# 生成时间线视图
cc-meta analyze timeline --group-by tool --format md
```

---

## 三、数据架构

### 3.1 核心数据流（两阶段）

**阶段1: 直接解析（MVP，无索引）**
```
JSONL 文件
    ↓
cc-meta parse extract
    ↓
结构化 JSON 输出
    ↓
Slash Command/Subagent 调用 Claude
    ↓
语义分析 + 建议生成
```

**阶段2: 索引增强（可选优化）**
```
JSONL 文件
    ↓
cc-meta index build
    ↓
SQLite 索引
    ↓
cc-meta query/analyze（基于规则）
    ↓
高密度分析结果
    ↓
Claude 语义理解
```

### 3.2 输出格式规范

**`cc-meta parse extract --type turns`**
```json
{
  "session_id": "5b57148c-89dc-4eb5-bc37-8122e194d90d",
  "project_hash": "-home-yale-work-myproject",
  "turn_count": 42,
  "turns": [
    {
      "sequence": 0,
      "role": "user",
      "timestamp": 1735689600,
      "content_preview": "帮我修复这个认证 bug",
      "has_attachments": false
    },
    {
      "sequence": 1,
      "role": "assistant",
      "timestamp": 1735689605,
      "tools_used": ["Read", "Grep"],
      "tool_calls": [
        {
          "tool": "Grep",
          "pattern": "auth.*error",
          "status": "success"
        }
      ]
    }
  ]
}
```

**`cc-meta parse extract --type tools --filter "status=error"`**
```json
{
  "total_tools": 87,
  "error_tools": 12,
  "tools": [
    {
      "turn_sequence": 15,
      "tool_name": "Bash",
      "command": "npm test",
      "status": "error",
      "exit_code": 1,
      "error_output": "FAIL test_auth.js\n  TypeError: Cannot read...",
      "timestamp": 1735689700
    }
  ]
}
```

**`cc-meta analyze errors --window 20`**
```json
{
  "analysis_type": "error_repetition",
  "window_size": 20,
  "patterns": [
    {
      "pattern_id": "err-001",
      "type": "identical_error",
      "occurrences": 5,
      "first_turn": 12,
      "last_turn": 28,
      "signature": "TypeError: Cannot read property 'id'",
      "tool": "Bash",
      "command_pattern": "npm test",
      "context": {
        "turns": [12, 15, 19, 24, 28],
        "time_span_minutes": 23
      }
    }
  ],
  "summary": {
    "total_errors": 12,
    "unique_errors": 3,
    "repeated_errors": 2
  }
}
```

### 3.3 索引结构（可选，阶段2）

**SQLite 数据库 (~/.cc-meta/index.db)**

```sql
-- 最小化索引表（仅加速查询）
CREATE TABLE sessions (
  session_id TEXT PRIMARY KEY,
  project_hash TEXT,
  first_turn_time INTEGER,
  last_turn_time INTEGER,
  turn_count INTEGER,
  tool_call_count INTEGER,
  error_count INTEGER
);

CREATE TABLE tool_calls (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  session_id TEXT,
  turn_sequence INTEGER,
  tool_name TEXT,
  status TEXT,
  timestamp INTEGER,
  error_hash TEXT  -- 用于快速匹配重复错误
);

CREATE INDEX idx_tool_session ON tool_calls(session_id, tool_name);
CREATE INDEX idx_tool_errors ON tool_calls(status, error_hash);
```

---

## 四、Claude Code 集成

### 4.1 环境变量传递机制

**Slash Command / Subagent 执行时的环境**
```bash
# Claude Code 应设置以下环境变量（需确认实现）
export CC_SESSION_ID="5b57148c-89dc-4eb5-bc37-8122e194d90d"
export CC_PROJECT_PATH="/home/user/work/myproject"
export CC_PROJECT_HASH="-home-user-work-myproject"

# Slash command 脚本中直接使用
cc-meta parse extract --type tools
# → 自动从 $CC_SESSION_ID 和 $CC_PROJECT_HASH 定位文件
```

**备选方案（如果 Claude Code 不提供环境变量）**
```bash
# 在 slash command 中手动传递
cc-meta parse extract \
  --project "$(pwd)" \
  --session-hint "latest"  # 使用最新会话
```

### 4.2 Slash Commands

**`/meta-stats`** - 会话统计（阶段1 MVP）

```markdown
# .claude/commands/meta-stats.md
---
name: meta-stats
description: 显示当前会话的统计信息
allowed_tools: [Bash]
---

运行以下命令获取会话统计：

```bash
cc-meta parse stats \
  --metrics tools,errors,duration \
  --output md
```

将结果格式化后显示给用户。
```

**`/meta-errors`** - 错误模式分析（阶段1 MVP）

```markdown
# .claude/commands/meta-errors.md
---
name: meta-errors
description: 分析当前会话中的错误模式
allowed_tools: [Bash]
---

执行错误分析：

```bash
error_data=$(cc-meta parse extract --type tools --filter "status=error" --output json)
pattern_data=$(cc-meta analyze errors --window 20 --output json)
```

基于以上数据（$error_data 和 $pattern_data），分析：
1. 是否存在重复错误？
2. 错误是否集中在某个工具/命令？
3. 给出优化建议（如添加 hook、修改工作流）
```

**`/meta-timeline`** - 会话时间线（阶段2）

```markdown
# .claude/commands/meta-timeline.md
---
name: meta-timeline
description: 生成当前会话的工具使用时间线
allowed_tools: [Bash]
---

```bash
cc-meta analyze timeline \
  --group-by tool \
  --format md
```

以可视化方式显示工具调用序列。
```

### 4.3 Subagent: @meta-coach（阶段2）

**agent 配置文件**
```markdown
# .claude/agents/meta-coach.md
---
name: meta-coach
description: 元认知教练，帮助开发者反思工作流程
model: claude-sonnet-4
allowed_tools: [Bash]
---

# 系统提示

你是开发者的元认知教练。通过分析会话历史，帮助开发者：
1. 识别重复性低效操作
2. 发现问题解决模式
3. 优化工作流程

## 可用的分析工具

使用 `cc-meta` CLI 获取结构化数据：

```bash
# 提取工具调用数据
cc-meta parse extract --type tools --output json

# 分析错误模式
cc-meta analyze errors --window 30 --output json

# 查询历史会话（如果已建索引）
cc-meta query sessions --project $CC_PROJECT_PATH --limit 10
```

## 对话风格

- 引导式提问，而非直接给答案
- 基于数据，而非猜测
- 帮助用户建立元认知意识

## 示例交互

User: 我感觉在重复做某件事...

Coach: 让我看看你最近的操作模式...
[调用 cc-meta parse extract --type tools]
发现你在过去 15 轮中运行了 6 次 `npm test`，每次都失败在同一个测试。
你觉得为什么会一直重复运行而不是专注修复？
```

### 4.4 MCP Server（阶段3，可选）

**服务器实现**
```typescript
// meta-insight-mcp/index.ts
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { exec } from "child_process";
import { promisify } from "util";

const execAsync = promisify(exec);

const server = new Server({
  name: "meta-insight",
  version: "1.0.0"
}, {
  capabilities: {
    tools: {}
  }
});

// 工具：提取会话数据
server.setRequestHandler("tools/call", async (request) => {
  if (request.params.name === "extract_session_data") {
    const { type, filter } = request.params.arguments;
    const cmd = `cc-meta parse extract --type ${type} ${filter ? `--filter "${filter}"` : ""} --output json`;
    const { stdout } = await execAsync(cmd);

    return {
      content: [{ type: "text", text: stdout }]
    };
  }
});

// 启动服务器
const transport = new StdioServerTransport();
await server.connect(transport);
```

**Claude Code 配置**
```bash
# 添加 MCP server
claude mcp add meta-insight npx -y meta-insight-mcp
```

**使用示例**
```
User: Use meta-insight MCP to check if I had similar
      errors in past sessions

Claude: [调用 MCP tool: extract_session_data]
        [分析返回的 JSON 数据]
        [给出语义化的回答]
```

---

## 五、核心功能实现

### 5.1 JSONL 解析（阶段1 核心）

**会话文件定位**
```python
import os
from pathlib import Path

def locate_session_file(session_id=None, project_path=None):
    """根据 session_id 或 project_path 定位会话文件"""

    # 方式1: 从环境变量获取
    if session_id is None:
        session_id = os.getenv("CC_SESSION_ID")
    if project_path is None:
        project_path = os.getenv("CC_PROJECT_PATH")

    # 转换项目路径为哈希目录名
    if project_path:
        project_hash = project_path.replace("/", "-")
        project_dir = Path.home() / ".claude" / "projects" / project_hash
    else:
        # 如果没有项目路径，需要遍历查找
        project_dir = find_project_dir_by_session(session_id)

    # 定位会话文件
    session_file = project_dir / f"{session_id}.jsonl"
    if not session_file.exists():
        raise FileNotFoundError(f"Session file not found: {session_file}")

    return session_file
```

**Turn 提取**
```python
import json
from typing import Iterator, Dict

def parse_turns(session_file: Path) -> Iterator[Dict]:
    """逐行解析 JSONL 文件，生成 turn 数据"""
    with open(session_file, 'r') as f:
        for line_no, line in enumerate(f):
            try:
                event = json.loads(line.strip())

                # 提取关键信息
                turn = {
                    "sequence": line_no,
                    "role": event.get("role"),
                    "timestamp": event.get("timestamp"),
                    "content_preview": extract_preview(event.get("content", [])),
                    "tools_used": extract_tools(event.get("content", []))
                }

                yield turn
            except json.JSONDecodeError:
                # 跳过损坏的行
                continue

def extract_tools(content_blocks):
    """从 content blocks 中提取工具调用"""
    tools = []
    for block in content_blocks:
        if block.get("type") == "tool_use":
            tools.append({
                "tool": block.get("name"),
                "input": block.get("input"),
                "id": block.get("id")
            })
        elif block.get("type") == "tool_result":
            # 查找对应的 tool_use，添加结果
            tool_id = block.get("tool_use_id")
            # ... 处理 tool_result
    return tools
```

### 5.2 基于规则的模式检测（阶段1）

**错误重复检测**
```python
from collections import defaultdict
from hashlib import sha256

def detect_error_repetition(turns, window=20):
    """检测重复错误模式（无 LLM）"""
    error_groups = defaultdict(list)

    # 只看最近 window 个 turns
    recent_turns = turns[-window:] if len(turns) > window else turns

    for turn in recent_turns:
        for tool in turn.get("tools_used", []):
            if tool.get("status") == "error":
                # 生成错误签名（基于工具名 + 错误输出前100字符）
                error_sig = sha256(
                    f"{tool['tool']}:{tool.get('error_output', '')[:100]}"
                    .encode()
                ).hexdigest()[:16]

                error_groups[error_sig].append({
                    "turn_sequence": turn["sequence"],
                    "tool": tool["tool"],
                    "error": tool.get("error_output")
                })

    # 输出重复 >= 3 次的错误
    patterns = []
    for sig, occurrences in error_groups.items():
        if len(occurrences) >= 3:
            patterns.append({
                "pattern_id": f"err-{sig}",
                "type": "identical_error",
                "occurrences": len(occurrences),
                "signature": occurrences[0]["error"][:200],
                "context": {
                    "turns": [o["turn_sequence"] for o in occurrences]
                }
            })

    return patterns
```

**工具使用模式**
```python
def analyze_tool_patterns(turns):
    """分析工具使用频率和顺序模式"""
    tool_freq = defaultdict(int)
    tool_sequences = []

    for turn in turns:
        tools = [t["tool"] for t in turn.get("tools_used", [])]
        for tool in tools:
            tool_freq[tool] += 1

        if len(tools) > 1:
            # 记录工具调用序列
            tool_sequences.append(" -> ".join(tools))

    # 检测高频序列
    seq_freq = defaultdict(int)
    for seq in tool_sequences:
        seq_freq[seq] += 1

    return {
        "tool_frequency": dict(tool_freq),
        "common_sequences": {
            seq: count
            for seq, count in seq_freq.items()
            if count >= 3
        }
    }
```

### 5.3 索引构建（阶段2，可选）

**增量索引**
```python
import sqlite3

def build_index(session_file: Path, db_path: Path):
    """构建会话索引（仅用于加速查询）"""
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    # 创建表
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS tool_calls (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            session_id TEXT,
            turn_sequence INTEGER,
            tool_name TEXT,
            status TEXT,
            error_hash TEXT,
            timestamp INTEGER
        )
    """)

    session_id = session_file.stem

    # 解析并索引
    for turn in parse_turns(session_file):
        for tool in turn.get("tools_used", []):
            cursor.execute("""
                INSERT INTO tool_calls
                (session_id, turn_sequence, tool_name, status, error_hash, timestamp)
                VALUES (?, ?, ?, ?, ?, ?)
            """, (
                session_id,
                turn["sequence"],
                tool["tool"],
                tool.get("status", "success"),
                hash_error(tool.get("error_output")),
                turn["timestamp"]
            ))

    conn.commit()
    conn.close()
```

---

## 六、实施计划

### 6.1 阶段1：核心解析（1-2周）

**目标：无需索引，直接解析 JSONL**

- [ ] CLI 框架搭建
  - 参数解析（--session, --project, --output）
  - 环境变量读取（CC_SESSION_ID, CC_PROJECT_PATH）
  - 会话文件定位逻辑

- [ ] JSONL 解析器
  - `parse_turns()`: 提取 turn 数据
  - `extract_tools()`: 提取工具调用和结果
  - `extract_errors()`: 识别错误工具调用

- [ ] 核心命令实现
  - `cc-meta parse extract --type turns/tools/errors`
  - `cc-meta parse stats --metrics tools,errors,duration`
  - `cc-meta analyze errors --window N`

- [ ] 输出格式化
  - JSON 输出（默认）
  - Markdown 表格输出
  - CSV 输出（可选）

- [ ] 集成测试
  - Slash Command: `/meta-stats`
  - Slash Command: `/meta-errors`

**交付物：**
- 可运行的 `cc-meta` CLI 工具
- 2 个可用的 Slash Commands
- 基础文档

### 6.2 阶段2：索引优化（1周，可选）

**目标：加速重复查询**

- [ ] SQLite 索引构建
  - `cc-meta index build`: 全量索引
  - `cc-meta index update`: 增量索引
  - 索引状态管理

- [ ] 高级查询命令
  - `cc-meta query sessions --project <path> --limit N`
  - `cc-meta query tools --name <tool> --since <date>`

- [ ] Slash Command: `/meta-timeline`

**交付物：**
- 可选的索引功能
- 更快的查询性能（跨会话）

### 6.3 阶段3：语义理解（1-2周，可选）

**目标：由 Claude 进行语义分析**

- [ ] Subagent: @meta-coach
  - agent 配置文件
  - 系统提示优化
  - 对话式分析逻辑

- [ ] MCP Server（可选）
  - MCP 协议实现
  - 工具定义（extract_session_data, analyze_patterns）
  - Claude Code 集成测试

**交付物：**
- @meta-coach subagent
- （可选）MCP server

---

## 七、关键设计决策

### 7.1 职责分离：CLI vs Claude

**CLI 工具职责（无 LLM）**
- ✅ JSONL 解析和数据提取
- ✅ 基于规则的模式检测（错误重复、工具频率）
- ✅ 结构化数据输出（JSON/Markdown）
- ✅ 索引构建和查询优化

**Claude 职责（在 Slash/Subagent/MCP 中）**
- ✅ 语义理解和分析
- ✅ 建议生成和优先级判断
- ✅ 上下文关联和推理
- ✅ 与用户的对话式交互

**为什么这样分离？**
1. **性能**：CLI 处理纯数据，速度快
2. **成本**：不为简单统计调用 LLM
3. **可测试性**：CLI 输出确定性，易于测试
4. **灵活性**：同一份数据，可被多个上层工具复用

### 7.2 会话定位策略

**优先级顺序**
1. 环境变量 `CC_SESSION_ID` + `CC_PROJECT_HASH`（最优）
2. 命令行参数 `--session <id>`
3. 项目路径推断 `--project <path>` → 转换为哈希目录
4. 自动查找最新会话（fallback）

**为什么需要多种方式？**
- Claude Code 可能不提供环境变量（需要实测确认）
- 不同集成方式（Slash/Subagent/MCP）可能有不同的上下文传递机制
- 终端手动使用时需要便捷方式

### 7.3 索引作为优化，而非必需

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

## 八、完整工作流示例

### 8.1 场景：用户重复遇到测试错误

**步骤1：用户触发分析**
```bash
# 在 Claude Code 中输入
/meta-errors
```

**步骤2：Slash Command 执行 CLI**
```bash
# .claude/commands/meta-errors.md 中的脚本
error_data=$(cc-meta parse extract --type tools --filter "status=error" --output json)
pattern_data=$(cc-meta analyze errors --window 20 --output json)
```

**步骤3：CLI 返回结构化数据**
```json
{
  "analysis_type": "error_repetition",
  "patterns": [
    {
      "pattern_id": "err-a1b2",
      "type": "identical_error",
      "occurrences": 5,
      "signature": "FAIL test_auth.js - TypeError: Cannot read property 'id' of undefined",
      "context": {
        "turns": [12, 15, 19, 24, 28],
        "time_span_minutes": 23,
        "tool": "Bash",
        "command": "npm test"
      }
    }
  ]
}
```

**步骤4：Claude 分析并给出建议**
```
检测到重复错误模式：

在过去 23 分钟内，你运行了 5 次 `npm test`，每次都失败在同一个测试：
- test_auth.js: TypeError: Cannot read property 'id' of undefined

建议：
1. 【立即】专注修复这一个测试，而不是重复运行整个测试套件
2. 【可选】添加一个 Hook，在运行测试前检查特定测试是否已修复
3. 【长期】考虑创建 /test-focus 命令，只运行失败的测试

是否需要我帮你创建这些优化？
```

### 8.2 场景：查询历史相似问题（需索引）

**步骤1：用户询问**
```
我记得之前也遇到过这个认证错误...
```

**步骤2：@meta-coach 分析**
```bash
# Subagent 调用 CLI
cc-meta query tools \
  --filter "tool=Bash AND status=error AND command LIKE '%auth%'" \
  --limit 5 \
  --output json
```

**步骤3：Claude 语义关联**
```
找到 3 次历史上的相似情况：

1. 2024-12-03：JWT token 验证失败
   → 解决方案：检查了 .env 中的 SECRET_KEY 配置

2. 2025-01-15：OAuth redirect 错误
   → 解决方案：修正了回调 URL

3. 2025-02-01：Session 过期问题（当前）
   → 状态：未解决

看起来你每次的认证问题都和配置有关。这次要不要先检查配置？
```

---

## 九、技术栈选择

### 9.1 推荐方案

**语言：Python**
- ✅ 快速开发，丰富的 JSON/CLI 库
- ✅ 易于集成 SQLite（内置 sqlite3）
- ✅ 未来可扩展嵌入 LLM（如 Anthropic SDK）
- ❌ 性能略低于 Rust（但对本场景足够）

**CLI 框架：Click**
- ✅ 简洁的命令/参数定义
- ✅ 自动生成帮助文档
- ✅ 广泛使用，成熟稳定

**数据库：SQLite（阶段2）**
- ✅ 零配置，单文件
- ✅ Python 内置支持
- ✅ 足够的查询能力

**替代方案：TypeScript + Bun**
- ✅ 与 Claude Code 技术栈一致
- ✅ 性能更好（Bun 的 SQLite 绑定）
- ❌ 生态略小于 Python

### 9.2 项目结构

```
cc-meta/
├── pyproject.toml          # 项目配置
├── src/
│   ├── cc_meta/
│   │   ├── __init__.py
│   │   ├── cli.py          # Click 命令定义
│   │   ├── parser.py       # JSONL 解析
│   │   ├── analyzer.py     # 模式检测
│   │   ├── indexer.py      # 索引构建（可选）
│   │   └── locator.py      # 会话文件定位
├── tests/
│   ├── test_parser.py
│   ├── test_analyzer.py
│   └── fixtures/           # 测试用 JSONL 样本
└── docs/
    └── integration.md      # 集成文档
```

---

## 十、总结

### 核心价值

1. **职责清晰**：CLI 做数据处理，Claude 做语义理解
2. **渐进式**：MVP（1-2周）→ 索引优化 → 语义增强
3. **低耦合**：通过环境变量/参数传递会话 ID，适配多种集成方式
4. **实用性**：基于真实会话数据，输出高密度结构化信息

### 与原提案的改进

**相比提案1**
- ✅ 明确了会话文件定位机制（环境变量/参数）
- ✅ 强调 CLI 无 LLM，语义分析由 Claude 完成
- ✅ 提供了完整的 Python 代码实现示例

**相比提案2**
- ✅ 简化了架构，去除冗余组件
- ✅ 索引改为可选，降低 MVP 复杂度
- ✅ 聚焦可操作性，而非理论设计

### 下一步行动

**验证阶段（1-2天）**
1. 确认 Claude Code 是否提供 `CC_SESSION_ID` 环境变量
2. 解析真实 JSONL 文件，确认数据结构
3. 验证 Slash Command 中调用外部 CLI 的方式

**MVP 开发（1-2周）**
1. 搭建 Python CLI 项目骨架
2. 实现核心解析和分析功能
3. 创建 2 个 Slash Commands 进行集成测试
4. 编写使用文档
