### 第一部分：核心引擎 - `cch-query` 命令行工具

这是整个系统的基础，一个用于索引、查询和分析 `~/.claude/projects/` 目录下会话历史的独立命令行工具。

#### 1\. 功能定位

  - **索引 (Indexing)**：为历史会话创建快速检索的索引，特别是为语义搜索创建向量索引。
  - **查询 (Querying)**：多维度检索会话历史。
  - **分析 (Analyzing)**：基于查询结果生成高级洞察和建议。

#### 2\. 接口定义 (CLI Commands)

```text
cch-query - Claude Code History Query Tool

USAGE:
    cch-query [SUBCOMMAND] [OPTIONS]

SUBCOMMANDS:
    index       创建或更新会话历史的索引。
    search      按关键词、工具使用、时间等维度搜索会话。
    stats       生成关于工具使用、提示模式等的统计数据。
    analyze     运行预定义的分析脚本，生成可操作的建议。
    list        列出项目或会话。

GLOBAL OPTIONS:
    --project-path <path>   指定项目路径 (默认为当前目录)。
    --output <format>       输出格式 (text, json, yaml)。
```

**子命令详解:**

  - `cch-query index`
      - `--full`: 强制进行全量重新索引。
  - `cch-query search "<query>"`
      - `--session <id|current>`: 限定在特定会话或当前会话。
      - `--since <duration>`: 限定时间范围 (e.g., `1h`, `3d`)。
      - `--tool <tool_name>`: 按使用的工具过滤 (e.g., `Bash`, `Edit`)。
      - `--role <user|assistant>`: 按角色过滤。
      - `--failed`: 只显示包含错误的轮次。
  - `cch-query stats`
      - `--metric <tool_usage|prompt_quality|error_rate>`: 指定统计指标。
  - `cch-query analyze`
      - `--level <tactical|strategic>`: 指定分析的层次。
      - `--focus <tool_usage|error_patterns>`: 指定分析的焦点。

-----

### 第二部分：系统架构与数据流 (PlantUML)

#### 1\. 组件架构图

此图展示了系统的核心组件及其关系。用户通过 Claude Code 接口与系统交互，这些接口在后台调用 `cch-query` 工具来处理存储在文件系统中的会话历史。

@startuml
\!theme plain

package "Claude Code Environment" {
actor User
artifact "Task Session UI" as UI

```
package "Integration Tools" {
    component "Slash Commands\n(/meta-analyze, /meta-suggest)" as Slash
    component "Subagent\n(@meta-coach)" as Subagent
    component "MCP Server\n(metacog-mcp)" as MCP
}
```

}

package "Local File System" {
database "Session History\\n(\~/.claude/projects/\*.jsonl)" as History
database "Search & Vector Index\\n(\~/.cch-cache/)" as Index
}

package "Meta-Cognition Backend (CLI)" {
component "cch-query CLI" as CLI
}

User -right-\> UI
UI -down-\> Slash
UI -down-\> Subagent
UI -down-\> MCP

Slash -down-\> CLI
Subagent -down-\> CLI
MCP -down-\> CLI

CLI -\> History : Reads
CLI -\> Index : Reads/Writes

@enduml

#### 2\. 工作流时序图 (示例：用户使用 `/meta-suggest`)

此图展示了一个具体的工作流程：用户请求建议，系统如何在后台协同工作以提供洞察。

@startuml
\!theme plain

actor User
participant "Task Session" as Session
participant "/meta-suggest\\n(Slash Command)" as Slash
participant "cch-query CLI" as CLI
database "Session History" as History

autonumber

User -\> Session: 输入 `/meta-suggest`
Session -\> Slash: 执行命令脚本
Slash -\> CLI: 调用 `cch-query analyze --level tactical`
CLI -\> History: 读取当前和历史会话数据
History --\> CLI: 返回原始会话事件 (JSONL)
CLI -\> CLI: 内部处理：\\n1. 识别模式 (e.g., 重复错误)\\n2. 生成建议
CLI --\> Slash: 返回结构化建议 (JSON)
Slash -\> Session: 格式化建议为 Markdown
Session --\> User: 显示可操作的建议列表

@enduml

-----

### 第三部分：Claude Code 集成工具设计

这些工具是用户与 `cch-query` 引擎交互的前端。

#### 1\. Slash Commands (快速、非对话式)

  - **`/meta-analyze <topic>`**: 即时分析。
      - **描述**: 对当前会话的某个方面进行快速分析。
      - **参数**: `tool-usage`, `error-patterns`, `prompt-quality`
      - **后台调用**: `cch-query stats --metric <topic> --session current` 或 `cch-query search --failed --session current`
  - **`/meta-suggest`**: 获取建议。
      - **描述**: 从后台分析中获取可操作的建议。
      - **后台调用**: `cch-query analyze --level tactical`
  - **`/meta-apply <suggestion_id>`**: 应用建议。
      - **描述**: 执行某个建议，如修改配置或创建新工具。
      - **后台调用**: 先用 `cch-query analyze --show <suggestion_id>` 获取建议详情，然后执行相应的文件操作 (e.g., `Edit`, `Write`)。

#### 2\. Subagent (深度、对话式)

  - **`@meta-coach`**: 元认知教练。
      - **定位**: 用于解决复杂问题、复盘或当开发者感到迷茫时。
      - **工作流程**:
        1.  **启动**: 询问用户当前遇到的困难和分析目标。
        2.  **数据收集**: 在对话中，根据需要调用 `cch-query` 的不同命令。
              - *“让我们看看你最近常用的工具有哪些？”* → `cch-query stats --metric tool_usage --since 1h`
              - *“我注意到你提到了一个 bug，让我们找找历史上类似的问题。”* → `cch-query search "authentication bug" --similar`
        3.  **引导与建议**: 基于数据，以对话方式引导用户思考，并共同制定下一步计划。

#### 3\. MCP (结构化、程序化)

  - **`metacog-mcp`**: 为 Claude 提供结构化查询历史的能力。
      - **定位**: 让 Claude 在自主规划（Plan Mode）或复杂推理中，能像使用 API 一样查询会话历史。
      - **提供的工具 (Tool Definitions)**:
        ```json
        {
          "tools": [
            {
              "name": "search_history",
              "description": "在会话历史中搜索相关信息、错误或代码片段。",
              "inputSchema": {
                "type": "object",
                "properties": {
                  "text_query": { "type": "string", "description": "关键词或自然语言描述。" },
                  "tool_filter": { "type": "string", "description": "仅搜索特定工具的使用，如 'Bash'。" },
                  "time_limit": { "type": "string", "description": "时间范围，如 'last_24_hours'。" }
                }
              }
            },
            {
              "name": "get_usage_statistics",
              "description": "获取关于工具使用频率或错误率的统计数据。",
              "inputSchema": {
                "type": "object",
                "properties": {
                  "metric": { "type": "string", "enum": ["tool_usage", "error_rate"] }
                }
              }
            }
          ]
        }
        ```
      - **后台实现**: MCP 服务器接收到工具调用请求后，将其转换为对应的 `cch-query` 命令并执行。

-----

### 第四部分：核心数据结构

定义标准化的数据结构是确保系统各部分协同工作的关键。

#### 1\. 会话轮次 (Turn) - `cch-query` 的内部表示

```json
{
  "turn_id": "sha1-of-content",
  "session_id": "uuid-...",
  "project_hash": "-Users-dev-project",
  "timestamp": "2025-10-01T12:00:00Z",
  "role": "assistant",
  "content_text": "Running `npm test`...",
  "tools_used": [
    {
      "name": "Bash",
      "input": { "command": "npm test" },
      "output_preview": "... 1 test failed",
      "status": "error"
    }
  ]
}
```

#### 2\. 分析建议 (AnalysisSuggestion) - `cch-query analyze` 的输出格式

```json
{
  "suggestions": [
    {
      "id": "sugg-2a8f",
      "priority": "High",
      "category": "Tactical",
      "title": "优化重复的失败测试",
      "description": "检测到你在过去1小时内运行了5次 `npm test`，并且每次都因 `test_user_creation` 失败。建议你专注于修复这一个测试，而不是运行整个测试套件。",
      "evidence": [
        "tool_usage(Bash, `npm test`) count: 5 in last 1h",
        "tool_output_grep(`test_user_creation failed`): 5 matches"
      ],
      "recommended_action": {
        "type": "create_slash_command",
        "details": {
          "name": "/test-focus",
          "description": "运行单个失败的测试",
          "command": "npm test -- test_user_creation"
        }
      }
    }
  ]
}
```

这个简化的原型设计将复杂性封装在 `cch-query` 工具中，同时通过 Claude Code 丰富的扩展接口提供了灵活、强大且易于实现的用户交互层。
