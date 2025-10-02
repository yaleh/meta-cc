# **Claude Code 元认知分析系统 - 简化原型架构设计**

---

## **第一部分：整体系统架构**

### **1.1 系统架构总览**

```plantuml
@startuml
!theme plain
skinparam componentStyle rectangle

package "Claude Code 任务会话" {
  [Claude Code Process] as CC
  [.claude/settings.json] as Settings
  [CLAUDE.md] as Memory
}

package "用户交互接口" {
  [/meta-suggest] as SlashSuggest
  [/meta-analyze] as SlashAnalyze
  [/meta-apply] as SlashApply
  [@meta-coach] as Subagent
  [meta-insight MCP] as MCP
}

package "命令行分析工具 (cc-meta-cli)" {
  component "Query Engine" as QueryEngine {
    [Session Indexer] as Indexer
    [Multi-Dimension Search] as Search
    [Statistics Generator] as Stats
  }
  
  component "Analysis Engine" as AnalysisEngine {
    [Pattern Detector] as Pattern
    [Suggestion Generator] as SugGen
    [Config Diff Generator] as ConfigGen
  }
  
  database "Local Cache" as Cache {
    [SQLite Index] as SQLite
    [JSON Summaries] as JSON
  }
}

package "数据源" {
  folder "~/.claude/projects/" as Projects {
    file "[hash]/session.jsonl" as JSONL
  }
  
  [Settings] --> JSONL : reads
}

' 数据流
CC --> SlashSuggest : invokes
CC --> SlashAnalyze : invokes
CC --> SlashApply : invokes
CC --> Subagent : spawns
CC --> MCP : calls tools

SlashSuggest --> QueryEngine : queries
SlashAnalyze --> AnalysisEngine : triggers
SlashApply --> AnalysisEngine : applies
Subagent --> QueryEngine : explores
MCP --> QueryEngine : provides data

QueryEngine --> JSONL : reads
QueryEngine --> Cache : updates
AnalysisEngine --> Cache : reads
AnalysisEngine --> Settings : modifies

note right of QueryEngine
  核心命令行工具
  - 离线索引构建
  - 多维度查询
  - 导出分析报告
end note

note right of AnalysisEngine
  分析与建议生成
  - 模式识别
  - 建议生成
  - 配置修改建议
end note

@enduml
```

---

### **1.2 数据流与交互序列**

```plantuml
@startuml
!theme plain

actor Developer as Dev
participant "Claude Code\nTask Session" as CC
participant "Slash Command\n/meta-suggest" as Slash
participant "cc-meta-cli\nQuery Engine" as CLI
database "Session History\n.jsonl" as JSONL
database "Local Index\nSQLite" as DB

== 后台索引构建（自动） ==
CC -> JSONL : writes turn data
note right: 每次推理后\n自动追加到 .jsonl

[-> CLI : cron job / file watcher
activate CLI
CLI -> JSONL : read new entries
CLI -> DB : update index
note right: 多维度索引:\n- 时间序列\n- 工具使用\n- 角色分类\n- 语义向量
deactivate CLI

== 开发者主动查询 ==
Dev -> CC : types "/meta-suggest"
activate CC
CC -> Slash : invokes
activate Slash

Slash -> CLI : cc-meta-cli query suggestions \n--session current
activate CLI
CLI -> DB : query indexed data
CLI -> DB : retrieve patterns
CLI --> Slash : return JSON:\n{suggestions: [...]}
deactivate CLI

Slash -> Slash : format for display
Slash --> CC : render suggestions
deactivate Slash
CC --> Dev : displays formatted output
deactivate CC

== 应用建议 ==
Dev -> CC : types "/meta-apply sugg-123"
activate CC
CC -> Slash : invokes apply
activate Slash

Slash -> CLI : cc-meta-cli get-suggestion sugg-123
activate CLI
CLI -> DB : fetch suggestion details
CLI --> Slash : return {type: "config", diff: {...}}
deactivate CLI

Slash -> CC : applies config changes
note right: 修改 settings.json\n或创建新的 hook/command

Slash --> CC : confirms application
deactivate Slash
CC --> Dev : shows result
deactivate CC

@enduml
```

---

## **第二部分：命令行工具架构 (cc-meta-cli)**

### **2.1 命令行工具组件设计**

```plantuml
@startuml
!theme plain

package "cc-meta-cli" {
  
  component "CLI Interface" as CLI {
    [Argument Parser] as ArgParser
    [Output Formatter] as Formatter
    [Config Loader] as ConfigLoader
  }
  
  component "Query Engine" as QueryEngine {
    
    component "Indexers" {
      [Time Indexer] as TimeIdx
      [Tool Indexer] as ToolIdx
      [Role Indexer] as RoleIdx
      [Semantic Indexer] as SemanticIdx
    }
    
    component "Search Strategies" {
      [Temporal Search] as TempSearch
      [Tool Pattern Search] as ToolSearch
      [Semantic Search] as SemSearch
      [Composite Search] as CompSearch
    }
    
    component "Aggregators" {
      [Statistics Aggregator] as StatsAgg
      [Pattern Aggregator] as PatternAgg
      [Timeline Aggregator] as TimelineAgg
    }
  }
  
  component "Analysis Engine" as Analysis {
    [Pattern Recognizer] as Recognizer
    [Suggestion Generator] as Generator
    [Config Validator] as Validator
  }
  
  component "Storage Layer" as Storage {
    [SQLite Handler] as SQLiteH
    [JSONL Parser] as JSONLParser
    [Cache Manager] as CacheM
  }
  
  ' Connections
  CLI --> QueryEngine
  CLI --> Analysis
  QueryEngine --> Storage
  Analysis --> Storage
  Analysis --> QueryEngine
  
}

note right of Indexers
  构建多维度索引:
  - 时间戳索引
  - 工具调用索引
  - 角色消息索引
  - 语义向量索引
end note

note right of Analysis
  分析能力:
  - 识别重复模式
  - 检测低效操作
  - 生成优化建议
  - 验证配置安全性
end note

@enduml
```

---

### **2.2 命令行工具命令结构**

```plantuml
@startmindmap
!theme plain

* cc-meta-cli

** index
*** build
**** --session <path>
**** --project <dir>
**** --all
*** rebuild
**** --force
*** stats
**** --summary

** query
*** sessions
**** --project <dir>
**** --since <date>
**** --until <date>
*** turns
**** --session <id>
**** --role <user|assistant>
**** --tool <name>
**** --range <start:end>
*** tools
**** --frequency
**** --errors-only
**** --by-session <id>
*** patterns
**** --type <error|efficiency|prompt>
**** --threshold <float>
*** search
**** --semantic <query>
**** --top-k <n>

** analyze
*** session
**** --id <session-id>
**** --level <tactical|strategic|meta>
**** --focus <topic>
*** project
**** --path <dir>
**** --aggregate
*** compare
**** --session-a <id>
**** --session-b <id>

** suggest
*** list
**** --priority <high|medium|low>
**** --category <tactical|strategic|meta>
**** --status <pending|applied|ignored>
*** get
**** --id <suggestion-id>
*** apply
**** --id <suggestion-id>
**** --dry-run
*** ignore
**** --id <suggestion-id>

** export
*** report
**** --format <json|markdown|html>
**** --session <id>
*** config
**** --suggestions <ids>
**** --merge

** config
*** set
**** --key <path>
**** --value <val>
*** get
**** --key <path>
*** show

@endmindmap
```

---

### **2.3 核心数据结构**

```plantuml
@startuml
!theme plain

class SessionIndex {
  + session_id: string
  + project_hash: string
  + start_time: timestamp
  + end_time: timestamp
  + turn_count: int
  + total_tokens: int
  + status: enum
  --
  + get_turns(): Turn[]
  + get_stats(): Statistics
}

class Turn {
  + turn_id: string
  + session_id: string
  + sequence: int
  + role: enum
  + timestamp: timestamp
  + content_blocks: ContentBlock[]
  + tokens_used: int
  --
  + extract_text(): string
  + get_tools(): ToolCall[]
}

class ToolCall {
  + tool_id: string
  + turn_id: string
  + tool_name: string
  + tool_input: JSON
  + tool_output: JSON
  + duration_ms: int
  + success: boolean
  + error_message: string
  --
  + is_repeated(): boolean
  + get_pattern_signature(): string
}

class Pattern {
  + pattern_id: string
  + type: enum
  + frequency: int
  + sessions: string[]
  + first_seen: timestamp
  + last_seen: timestamp
  + signature: string
  + metadata: JSON
  --
  + get_examples(): Turn[]
  + calculate_impact(): float
}

class Suggestion {
  + suggestion_id: string
  + priority: enum
  + category: enum
  + title: string
  + description: string
  + rationale: string
  + action_type: enum
  + executable: JSON
  + status: enum
  + created_at: timestamp
  + applied_at: timestamp
  --
  + to_config_diff(): ConfigDiff
  + validate(): boolean
  + estimate_impact(): ImpactScore
}

class Statistics {
  + session_id: string
  + tool_usage_freq: Map<string, int>
  + avg_turn_duration: float
  + error_rate: float
  + prompt_quality_score: float
  + efficiency_score: float
  --
  + compare_with(other: Statistics): Comparison
  + generate_report(): Report
}

class SemanticVector {
  + vector_id: string
  + turn_id: string
  + embedding: float[]
  + text_preview: string
  + metadata: JSON
  --
  + similarity(other: SemanticVector): float
  + nearest_neighbors(k: int): SemanticVector[]
}

' Relationships
SessionIndex "1" *-- "many" Turn
Turn "1" *-- "many" ToolCall
Turn "1" -- "1" SemanticVector
Pattern "1" -- "many" Turn : identifies
Suggestion "1" -- "many" Pattern : based_on
SessionIndex "1" -- "1" Statistics : generates

enum TurnRole {
  USER
  ASSISTANT
  SYSTEM
}

enum PatternType {
  ERROR_REPETITION
  TOOL_MISUSE
  INEFFICIENT_WORKFLOW
  PROMPT_AMBIGUITY
  CONTEXT_LOSS
}

enum SuggestionCategory {
  TACTICAL
  STRATEGIC
  METACOGNITIVE
}

enum ActionType {
  CONFIG_MODIFICATION
  HOOK_ADDITION
  COMMAND_CREATION
  PROMPT_TEMPLATE
  WORKFLOW_CHANGE
}

@enduml
```

---

## **第三部分：用户交互接口设计**

### **3.1 Slash Commands 接口规范**

```plantuml
@startuml
!theme plain

interface ISlashCommand {
  + name: string
  + description: string
  + argument_hint: string
  + allowed_tools: string[]
  --
  + execute(args: string): CommandResult
  + validate_args(args: string): boolean
  + get_help(): string
}

class MetaSuggestCommand implements ISlashCommand {
  - cli_path: string
  --
  + execute(args: string): CommandResult
  --
  **Logic:**
  1. Call: cc-meta-cli suggest list
  2. Parse JSON output
  3. Format for display
  4. Return structured result
}

class MetaAnalyzeCommand implements ISlashCommand {
  - cli_path: string
  --
  + execute(args: string): CommandResult
  --
  **Logic:**
  1. Extract focus topic from args
  2. Get current session path
  3. Call: cc-meta-cli analyze session
     --focus $TOPIC
  4. Return formatted analysis
}

class MetaApplyCommand implements ISlashCommand {
  - cli_path: string
  - validator: ConfigValidator
  --
  + execute(args: string): CommandResult
  --
  **Logic:**
  1. Extract suggestion_id
  2. Call: cc-meta-cli suggest get --id
  3. Validate action safety
  4. Apply configuration changes
  5. Mark suggestion as applied
  6. Return confirmation
}

class MetaExportCommand implements ISlashCommand {
  - cli_path: string
  --
  + execute(args: string): CommandResult
  --
  **Logic:**
  1. Get current session
  2. Call: cc-meta-cli export report
     --format markdown
  3. Save to file
  4. Return file path
}

class CommandResult {
  + success: boolean
  + message: string
  + data: JSON
  + display_format: enum
  --
  + to_markdown(): string
  + to_json(): string
}

note right of MetaApplyCommand
  **Safety Checks:**
  - Validate JSON syntax
  - Check file permissions
  - Backup existing configs
  - Rollback on failure
end note

@enduml
```

---

### **3.2 Subagent 架构设计**

```plantuml
@startuml
!theme plain

abstract class BaseSubagent {
  # name: string
  # description: string
  # tools: string[]
  # model: string
  # system_prompt: string
  --
  + initialize(context: SessionContext): void
  + process_query(query: string): Response
  # call_cli(command: string): CLIResult
  # format_response(data: JSON): string
}

class MetaCoachSubagent extends BaseSubagent {
  - conversation_state: ConversationState
  - analysis_cache: Cache
  --
  + initialize(context: SessionContext): void
  + process_query(query: string): Response
  --
  **Capabilities:**
  - Analyze current session
  - Provide strategic guidance
  - Suggest next steps
  - Facilitate reflection
}

class SessionReviewerSubagent extends BaseSubagent {
  - review_criteria: Criteria[]
  --
  + initialize(context: SessionContext): void
  + process_query(query: string): Response
  + set_criteria(criteria: Criteria[]): void
  --
  **Capabilities:**
  - Deep dive into session history
  - Multi-turn analysis dialogue
  - Comparative analysis
  - Generate detailed reports
}

class PatternExplorerSubagent extends BaseSubagent {
  - pattern_types: PatternType[]
  --
  + initialize(context: SessionContext): void
  + process_query(query: string): Response
  + filter_patterns(filters: Filter[]): Pattern[]
  --
  **Capabilities:**
  - Identify recurring patterns
  - Explain pattern impact
  - Suggest pattern-breaking strategies
  - Track pattern evolution
}

class ConversationState {
  + current_topic: string
  + insights_shared: Insight[]
  + questions_asked: Question[]
  + user_preferences: Preferences
  --
  + add_insight(insight: Insight): void
  + track_question(q: Question): void
  + update_preferences(p: Preferences): void
}

class SessionContext {
  + session_id: string
  + project_path: string
  + transcript_path: string
  + current_turn: int
  + developer_history: History
  --
  + load_from_env(): SessionContext
  + get_recent_turns(n: int): Turn[]
}

MetaCoachSubagent --> SessionContext : uses
MetaCoachSubagent --> ConversationState : maintains
SessionReviewerSubagent --> SessionContext : uses
PatternExplorerSubagent --> SessionContext : uses

note bottom of BaseSubagent
  **CLI Integration:**
  All subagents can invoke:
  - cc-meta-cli query
  - cc-meta-cli analyze
  - cc-meta-cli suggest
  
  via call_cli() method
end note

@enduml
```

---

### **3.3 MCP Server 接口设计**

```plantuml
@startuml
!theme plain

package "meta-insight MCP Server" {
  
  interface MCPServer {
    + name: string
    + version: string
    + tools: Tool[]
    + resources: Resource[]
    --
    + handle_tool_call(tool: string, args: JSON): ToolResult
    + list_resources(): Resource[]
    + read_resource(uri: string): ResourceContent
  }
  
  class MetaInsightMCP implements MCPServer {
    - cli_wrapper: CLIWrapper
    - cache: QueryCache
    --
    + handle_tool_call(tool: string, args: JSON): ToolResult
    + list_resources(): Resource[]
    + read_resource(uri: string): ResourceContent
  }
  
  class CLIWrapper {
    - cli_path: string
    - timeout: int
    --
    + execute(command: string[], input: JSON): CLIResult
    + parse_output(output: string): JSON
    + handle_error(error: Error): ErrorResponse
  }
  
  class QueryCache {
    - ttl: int
    - max_size: int
    - storage: Map<string, CacheEntry>
    --
    + get(key: string): JSON | null
    + set(key: string, value: JSON): void
    + invalidate(pattern: string): void
  }
}

' MCP Tools Definition
class Tool_QuerySessionStats {
  + name: "query_session_stats"
  + description: "..."
  + input_schema: JSONSchema
  --
  **Input:**
  {
    "session_id": string,
    "metric": enum
  }
  --
  **Output:**
  {
    "tool_usage": Map<string, int>,
    "error_rate": float,
    "efficiency_score": float
  }
  --
  **CLI Mapping:**
  cc-meta-cli query tools
    --session $session_id
    --frequency
}

class Tool_SearchSimilarSituations {
  + name: "search_similar_situations"
  + description: "..."
  + input_schema: JSONSchema
  --
  **Input:**
  {
    "query": string,
    "top_k": int
  }
  --
  **Output:**
  {
    "results": SearchResult[],
    "total": int
  }
  --
  **CLI Mapping:**
  cc-meta-cli query search
    --semantic "$query"
    --top-k $top_k
}

class Tool_GetPatternInsights {
  + name: "get_pattern_insights"
  + description: "..."
  + input_schema: JSONSchema
  --
  **Input:**
  {
    "category": enum,
    "session_id": string?
  }
  --
  **Output:**
  {
    "patterns": Pattern[],
    "recommendations": string[]
  }
  --
  **CLI Mapping:**
  cc-meta-cli query patterns
    --type $category
    --session $session_id
}

class Tool_AnalyzeCurrentSession {
  + name: "analyze_current_session"
  + description: "..."
  + input_schema: JSONSchema
  --
  **Input:**
  {
    "level": enum,
    "focus": string?
  }
  --
  **Output:**
  {
    "insights": Insight[],
    "suggestions": Suggestion[]
  }
  --
  **CLI Mapping:**
  cc-meta-cli analyze session
    --id current
    --level $level
    --focus $focus
}

class Tool_GetSuggestions {
  + name: "get_suggestions"
  + description: "..."
  + input_schema: JSONSchema
  --
  **Input:**
  {
    "status": enum,
    "priority": enum?
  }
  --
  **Output:**
  {
    "suggestions": Suggestion[]
  }
  --
  **CLI Mapping:**
  cc-meta-cli suggest list
    --status $status
    --priority $priority
}

' MCP Resources Definition
class Resource_SessionSummary {
  + uri: "session://current/summary"
  + name: "Current Session Summary"
  + mime_type: "application/json"
  --
  **CLI Mapping:**
  cc-meta-cli query sessions
    --current
}

class Resource_ToolUsageReport {
  + uri: "report://tools/usage"
  + name: "Tool Usage Report"
  + mime_type: "text/markdown"
  --
  **CLI Mapping:**
  cc-meta-cli export report
    --format markdown
    --session current
}

class Resource_PatternDatabase {
  + uri: "patterns://all"
  + name: "Pattern Database"
  + mime_type: "application/json"
  --
  **CLI Mapping:**
  cc-meta-cli query patterns
    --all
}

MetaInsightMCP --> CLIWrapper : uses
MetaInsightMCP --> QueryCache : uses
MetaInsightMCP ..> Tool_QuerySessionStats : provides
MetaInsightMCP ..> Tool_SearchSimilarSituations : provides
MetaInsightMCP ..> Tool_GetPatternInsights : provides
MetaInsightMCP ..> Tool_AnalyzeCurrentSession : provides
MetaInsightMCP ..> Tool_GetSuggestions : provides
MetaInsightMCP ..> Resource_SessionSummary : exposes
MetaInsightMCP ..> Resource_ToolUsageReport : exposes
MetaInsightMCP ..> Resource_PatternDatabase : exposes

note right of MetaInsightMCP
  **MCP Server 配置:**
  
  # 添加到 Claude Code
  claude mcp add meta-insight \\
    npx cc-meta-mcp-server
  
  **使用示例:**
  > Use meta-insight MCP to
    search for similar debugging
    situations in my history
end note

@enduml
```

---

## **第四部分：数据处理流程**

### **4.1 索引构建流程**

```plantuml
@startuml
!theme plain
title 索引构建与更新流程

|Session History|
start
:新的 Turn 写入\n.jsonl 文件;

|cc-meta-cli|
:检测文件变化\n(inotify/polling);

:解析 JSONL 增量;
note right
  只处理新增的行
  避免重复解析
end note

fork
  :时间序列索引;
  :插入 Turn 记录到\ntimeline 表;
fork again
  :工具调用索引;
  :提取 tool_use blocks;
  :插入到 tool_calls 表;
fork again
  :角色分类索引;
  :按 role 分组;
  :更新统计计数;
fork again
  :语义向量索引;
  :提取文本内容;
  :生成 embedding;
  :存储到向量数据库;
end fork

:更新会话元数据;
note right
  - turn_count
  - last_update
  - total_tokens
end note

:触发模式检测;

if (达到分析阈值?) then (yes)
  :运行模式识别;
  :更新 patterns 表;
else (no)
  :跳过;
endif

:写入索引版本号;

stop

@enduml
```

---

### **4.2 查询执行流程**

```plantuml
@startuml
!theme plain
title 多维度查询执行流程

actor User
participant "CLI Interface" as CLI
participant "Query Planner" as Planner
database "SQLite Index" as DB
database "Vector Store" as VDB
participant "Result Aggregator" as Agg

User -> CLI : cc-meta-cli query \\n--semantic "auth error" \\n--tool Bash \\n--since 2025-06-01

activate CLI
CLI -> Planner : parse query
activate Planner

Planner -> Planner : identify query types
note right
  检测到多个查询维度:
  1. 语义查询 (semantic)
  2. 工具过滤 (tool)
  3. 时间范围 (since)
end note

Planner -> Planner : build execution plan
note right
  优化查询顺序:
  1. 先执行选择性高的过滤
     (时间、工具)
  2. 再执行计算密集的操作
     (语义搜索)
end note

' Parallel execution
par Temporal Filter
  Planner -> DB : SELECT turn_id FROM timeline \\nWHERE timestamp >= '2025-06-01'
  DB --> Planner : turn_ids_temporal
else Tool Filter
  Planner -> DB : SELECT turn_id FROM tool_calls \\nWHERE tool_name = 'Bash'
  DB --> Planner : turn_ids_tool
end

Planner -> Planner : intersect(turn_ids_temporal, turn_ids_tool)
note right: 取交集缩小范围

Planner -> VDB : semantic_search("auth error", \\n  filter_ids=intersected_ids)
activate VDB
VDB -> VDB : compute similarities
VDB --> Planner : ranked_turn_ids
deactivate VDB

Planner -> DB : fetch_turn_details(ranked_turn_ids)
activate DB
DB --> Planner : turn_records[]
deactivate DB

Planner -> Agg : aggregate(turn_records)
activate Agg
Agg -> Agg : group_by_session
Agg -> Agg : calculate_statistics
Agg -> Agg : format_output
Agg --> Planner : formatted_results
deactivate Agg

Planner --> CLI : query_results
deactivate Planner

CLI -> CLI : render(query_results)
CLI --> User : display formatted output

deactivate CLI

@enduml
```

---

### **4.3 建议生成与应用流程**

```plantuml
@startuml
!theme plain
title 建议生成与应用完整流程

|Analysis Engine|
start

:接收分析请求\n(session_id, level);

:从索引加载会话数据;
note right
  包括:
  - 工具使用统计
  - 错误模式
  - 时间分布
  - 提示质量指标
end note

:执行模式识别;

fork
  :检测工具滥用模式;
  :生成自动化建议;
fork again
  :检测重复错误;
  :生成配置优化建议;
fork again
  :分析提示质量;
  :生成提示改进模板;
fork again
  :检测上下文丢失;
  :生成工作流建议;
end fork

:聚合所有建议;

:按优先级排序;
note right
  优先级算法:
  - 高: 影响 > 阈值 && 易实施
  - 中: 影响中等 || 实施复杂
  - 低: 影响小
end note

:为每条建议生成\n可执行内容;

split
  :Config 类型建议;
  :生成 JSON diff;
  :验证语法和安全性;
split again
  :Hook 类型建议;
  :生成 shell 脚本;
  :添加安全检查;
split again
  :Prompt 类型建议;
  :生成模板文档;
  :添加使用示例;
end split

:保存到建议数据库;

:返回建议列表;

stop

|User Interface|

:用户查看建议\n(/meta-suggest);

if (是否应用?) then (yes)
  :用户调用\n/meta-apply <id>;
  
  |Application Engine|
  
  :加载建议详情;
  
  :执行安全检查;
  note right
    - 备份原配置
    - 验证语法
    - 检查权限
  end note
  
  if (检查通过?) then (yes)
    
    switch (建议类型)
    case (Config Modification)
      :读取 settings.json;
      :应用 JSON patch;
      :写回文件;
    case (Hook Addition)
      :创建 hook 脚本;
      :设置执行权限;
      :更新 settings.json;
    case (Command Creation)
      :创建 .md 文件;
      :写入模板;
      :放置到正确目录;
    case (Prompt Template)
      :追加到 CLAUDE.md;
      :或创建独立文档;
    endswitch
    
    :标记建议为已应用;
    
    :记录应用日志;
    
    :返回成功消息;
    
  else (no)
    :返回错误信息;
    :建议手动操作;
  endif
  
else (no)
  :标记为已查看;
endif

stop

@enduml
```

---

## **第五部分：部署与配置**

### **5.1 系统组件部署图**

```plantuml
@startuml
!theme plain

node "开发者机器" {
  
  frame "Claude Code 环境" {
    component [Claude Code Process] as CC
    folder ".claude/" {
      file "settings.json"
      folder "commands/" {
        file "meta-suggest.md"
        file "meta-analyze.md"
        file "meta-apply.md"
      }
      folder "agents/" {
        file "meta-coach.md"
      }
      folder "projects/" {
        folder "[hash]/" {
          file "session.jsonl"
        }
      }
    }
  }
  
  frame "cc-meta-cli 工具" {
    component [CLI Binary] as CLI
    database "index.db" {
      [timeline]
      [tool_calls]
      [patterns]
      [suggestions]
    }
    database "vector.db" {
      [embeddings]
    }
    folder "cache/" {
      file "summaries.json"
    }
  }
  
  frame "MCP Server (可选)" {
    component [meta-insight-mcp] as MCP
  }
}

CC --> CLI : invokes via Slash Commands
CC --> MCP : calls via MCP protocol
MCP --> CLI : delegates queries
CLI --> "session.jsonl" : reads
CLI --> "index.db" : reads/writes
CLI --> "vector.db" : reads/writes

note right of CLI
  **安装:**
  npm install -g cc-meta-cli
  
  或
  
  brew install cc-meta-cli
end note

note right of MCP
  **配置:**
  claude mcp add meta-insight \\
    npx @cc-meta/mcp-server
end note

@enduml
```

---

### **5.2 配置文件结构**

```plantuml
@startyaml
#highlight "config" / "indexing" / "analysis"

config:
  version: "1.0.0"
  cli_path: "/usr/local/bin/cc-meta-cli"
  
  indexing:
    enabled: true
    mode: "incremental"  # or "full"
    watch_mode: true
    watch_interval: 5  # seconds
    
    indices:
      - name: "timeline"
        enabled: true
        fields: ["timestamp", "session_id", "turn_id", "role"]
      
      - name: "tool_calls"
        enabled: true
        fields: ["tool_name", "input_hash", "exit_code", "duration"]
      
      - name: "semantic"
        enabled: true
        model: "text-embedding-3-small"
        dimension: 1536
        
  analysis:
    auto_trigger:
      enabled: true
      tactical_interval: 5  # every 5 turns
      strategic_interval: 20  # every 20 turns
    
    pattern_detection:
      min_frequency: 3
      time_window: "7d"
      categories:
        - "error_repetition"
        - "tool_misuse"
        - "inefficient_workflow"
        - "prompt_ambiguity"
    
    suggestion_generation:
      max_suggestions: 10
      priority_threshold: 0.7
      auto_apply_safe: false
      
  storage:
    database: "sqlite"
    path: "~/.cc-meta/index.db"
    vector_store: "chromadb"
    vector_path: "~/.cc-meta/vectors"
    cache_ttl: 3600  # seconds
    
  mcp_server:
    enabled: true
    port: 9001
    transport: "stdio"  # or "sse"
    
  privacy:
    anonymize_prompts: false
    exclude_patterns:
      - "*.env"
      - "*.key"
      - "secrets/**"
      
  logging:
    level: "info"
    path: "~/.cc-meta/logs/"
    
@endyaml
```

---

### **5.3 接口协议定义**

```plantuml
@startuml
!theme plain

package "CLI Output Formats" {
  
  class JSONOutput {
    + version: string = "1.0"
    + timestamp: string
    + command: string
    + result: JSON
    + metadata: Metadata
    --
    **Example:**
    {
      "version": "1.0",
      "timestamp": "2025-06-02T10:00:00Z",
      "command": "query sessions",
      "result": {
        "sessions": [...]
      },
      "metadata": {
        "execution_time_ms": 45,
        "cache_hit": false
      }
    }
  }
  
  class MarkdownOutput {
    + header: string
    + sections: Section[]
    + footer: string
    --
    **Example:**
    # Session Analysis Report
    
    ## Overview
    - Total turns: 42
    - Duration: 1h 23m
    
    ## Tool Usage
    | Tool | Count |
    |------|-------|
    | Edit | 15    |
    | Bash | 8     |
  }
  
  class TableOutput {
    + headers: string[]
    + rows: string[][]
    + format: "ascii" | "csv"
    --
    **Example:**
    ┌────────┬────────┐
    │ Tool   │ Count  │
    ├────────┼────────┤
    │ Edit   │ 15     │
    │ Bash   │ 8      │
    └────────┴────────┘
  }
}

package "MCP Protocol Messages" {
  
  class ToolCallRequest {
    + jsonrpc: "2.0"
    + id: string
    + method: "tools/call"
    + params: {
        name: string,
        arguments: JSON
      }
  }
  
  class ToolCallResponse {
    + jsonrpc: "2.0"
    + id: string
    + result: {
        content: Content[],
        isError: boolean
      }
  }
  
  class ResourceReadRequest {
    + jsonrpc: "2.0"
    + id: string
    + method: "resources/read"
    + params: {
        uri: string
      }
  }
  
  class ResourceReadResponse {
    + jsonrpc: "2.0"
    + id: string
    + result: {
        contents: ResourceContent[]
      }
  }
}

package "Internal Data Transfer" {
  
  class CLIResult {
    + success: boolean
    + exit_code: int
    + stdout: string
    + stderr: string
    + data: JSON
    --
    + parse_json(): JSON
    + is_error(): boolean
  }
  
  class QueryResult {
    + total: int
    + items: JSON[]
    + pagination: {
        page: int,
        per_page: int,
        has_more: boolean
      }
    + execution_time_ms: int
  }
  
  class SuggestionPayload {
    + id: string
    + priority: "high" | "medium" | "low"
    + category: Category
    + title: string
    + description: string
    + rationale: string
    + action: {
        type: ActionType,
        target: string,
        content: JSON,
        safety_level: int
      }
    --
    + to_config_diff(): ConfigDiff
    + to_markdown(): string
  }
}

note right of JSONOutput
  **CLI 输出格式:**
  
  默认使用 JSON (--output json)
  可通过 --output 参数切换:
  - json
  - markdown
  - table
  - csv
end note

note right of ToolCallRequest
  **MCP 工具调用:**
  
  Claude Code 通过 MCP 协议
  调用 meta-insight 服务器的工具
  
  服务器委托给 cc-meta-cli
end note

@enduml
```

---

## **第六部分：使用场景示例**

### **6.1 场景 1：快速查看建议**

```plantuml
@startuml
!theme plain

actor Developer as Dev
participant "Claude Code" as CC
participant "/meta-suggest" as Slash
participant "cc-meta-cli" as CLI

Dev -> CC : types "/meta-suggest"
activate CC

CC -> Slash : invoke command
activate Slash

Slash -> CLI : cc-meta-cli suggest list \\n--priority high,medium \\n--status pending \\n--output json
activate CLI

CLI -> CLI : query suggestions table
CLI -> CLI : filter by criteria
CLI -> CLI : sort by priority

CLI --> Slash : return JSON:\n{\n  "suggestions": [\n    {\n      "id": "s-001",\n      "priority": "high",\n      "title": "Automate lint",\n      ...\n    }\n  ]\n}
deactivate CLI

Slash -> Slash : format for display
Slash --> CC : rendered markdown

CC --> Dev : displays:\n\n📊 **元认知建议** (2 条)\n\n🔴 高优先级\n1. [s-001] 自动化 lint 检查\n   - 检测到 8 次手动运行\n   - 建议添加 PostToolUse Hook\n   - `/meta-apply s-001`

deactivate Slash
deactivate CC

note right of Dev
  开发者看到建议后
  可以选择:
  1. 应用 (/meta-apply)
  2. 查看详情 (/meta-analyze)
  3. 与 coach 讨论
  4. 忽略
end note

@enduml
```

---

### **6.2 场景 2：深度对话式分析**

```plantuml
@startuml
!theme plain

actor Developer as Dev
participant "Meta-Coach\nSubagent" as Coach
participant "cc-meta-cli" as CLI

Dev -> Coach : "I feel stuck on this\nauthentication bug"
activate Coach

Coach -> CLI : cc-meta-cli query turns \\n--session current \\n--tool Bash \\n--range -10:0
activate CLI
CLI --> Coach : recent bash commands
deactivate CLI

Coach -> CLI : cc-meta-cli query patterns \\n--type error \\n--session current
activate CLI
CLI --> Coach : error patterns found
deactivate CLI

Coach --> Dev : **我的观察:**\n\n在过去 10 轮中:\n- 你运行了 5 次测试\n- 都失败在 test_auth\n- 错误信息相同:\n  "Token validation failed"\n\n**问题:**\n你是否检查了 token 的:\n1. 过期时间?\n2. 签名算法?\n3. Secret key 配置?

Dev -> Coach : "Hmm, I only checked #1"

Coach -> CLI : cc-meta-cli query search \\n--semantic "token secret config" \\n--top-k 3
activate CLI
CLI --> Coach : similar past situations
deactivate CLI

Coach --> Dev : **好消息!**\n\n我在你的历史中找到了\n类似情况（3 个月前）:\n\n你当时在 .env 文件中\n配置了 JWT_SECRET，但\n忘记在测试环境中设置。\n\n**建议:**\n1. 检查测试配置\n2. 添加配置验证 Hook\n3. 我可以帮你设置吗?

Dev -> Coach : "Yes, please help me set it up"

Coach -> CLI : cc-meta-cli suggest get \\n--category config_validation
activate CLI
CLI --> Coach : pre-existing suggestion found
deactivate CLI

Coach --> Dev : **太好了!**\n\n我找到了一个现成的建议:\n[s-042] 配置验证 Hook\n\n这个 Hook 会在启动时\n验证所有必需的环境变量\n\n要应用吗?\n> /meta-apply s-042

deactivate Coach

note right of Coach
  Coach 结合了:
  1. 当前会话分析
  2. 历史模式检索
  3. 预生成建议库
  
  提供上下文相关的
  对话式指导
end note

@enduml
```

---

### **6.3 场景 3：通过 MCP 进行语义搜索**

```plantuml
@startuml
!theme plain

actor Developer as Dev
participant "Claude Code" as CC
participant "meta-insight\nMCP Server" as MCP
participant "cc-meta-cli" as CLI

Dev -> CC : "Use meta-insight MCP to\nsearch for past solutions\nto database migration issues"

activate CC

CC -> MCP : tool_call:\n{\n  "name": "search_similar_situations",\n  "arguments": {\n    "query": "database migration",\n    "top_k": 5\n  }\n}
activate MCP

MCP -> CLI : cc-meta-cli query search \\n--semantic "database migration" \\n--top-k 5 \\n--output json
activate CLI

CLI -> CLI : generate query embedding
CLI -> CLI : search vector database
CLI -> CLI : rank by similarity
CLI -> CLI : fetch turn details

CLI --> MCP : {\n  "results": [\n    {\n      "session_id": "...",\n      "turn_id": "...",\n      "similarity": 0.89,\n      "text_preview": "...",\n      "solution": "..."\n    }\n  ]\n}
deactivate CLI

MCP --> CC : tool_result:\n{\n  "content": [\n    {\n      "type": "text",\n      "text": "Found 3 relevant cases..."\n    }\n  ]\n}
deactivate MCP

CC -> CC : synthesize response

CC --> Dev : **我在你的历史中找到了\n3 个相关案例:**\n\n**案例 1** (2024-12-15)\n- 情况: PostgreSQL 迁移失败\n- 解决方案: 使用事务回滚\n- 相似度: 89%\n- [查看完整对话](session://...)\n\n**案例 2** (2025-01-03)\n...\n\n**建议:**\n基于这些案例，你可能需要:\n1. 检查外键约束\n2. 使用 --dry-run 预检查\n3. 分步骤执行迁移

deactivate CC

note right of Dev
  MCP 工具提供了:
  1. 语义搜索能力
  2. 跨会话检索
  3. 结构化结果
  
  Claude 可以基于这些
  数据进行推理和综合
end note

@enduml
```

---

## **第七部分：扩展性与演进路径**

### **7.1 功能模块化与扩展点**

```plantuml
@startuml
!theme plain

package "Core System" {
  [CLI Engine] as Core
  [Index Manager] as Index
  [Query Processor] as Query
  [Analysis Engine] as Analysis
}

package "Extension Points" {
  
  interface IIndexer {
    + index_turn(turn: Turn): void
    + build_index(session: Session): void
  }
  
  interface ISearchStrategy {
    + search(query: Query): Result[]
    + supports(query_type: string): boolean
  }
  
  interface IPatternDetector {
    + detect(session: Session): Pattern[]
    + get_pattern_type(): string
  }
  
  interface ISuggestionGenerator {
    + generate(patterns: Pattern[]): Suggestion[]
    + get_category(): string
  }
  
  interface IOutputFormatter {
    + format(data: JSON): string
    + get_format_type(): string
  }
}

package "Built-in Implementations" {
  [TimeIndexer] as TimeIdx
  [ToolIndexer] as ToolIdx
  [SemanticIndexer] as SemIdx
  
  [TemporalSearch] as TempSearch
  [ToolPatternSearch] as ToolSearch
  [SemanticSearch] as SemSearch
  
  [ErrorRepetitionDetector] as ErrDet
  [ToolMisuseDetector] as ToolDet
  
  [ConfigSuggestionGen] as ConfigGen
  [PromptSuggestionGen] as PromptGen
  
  [JSONFormatter] as JSON
  [MarkdownFormatter] as MD
}

package "Custom Extensions" {
  [CustomIndexer] as CustomIdx
  [CustomDetector] as CustomDet
  [CustomGenerator] as CustomGen
  
  note right of CustomIdx
    **用户可以实现:**
    
    例如：索引特定领域数据
    - Git commit 关联
    - Issue tracker 集成
    - Code metrics 索引
  end note
}

Core --> Index
Core --> Query
Core --> Analysis

Index ..> IIndexer : uses
Query ..> ISearchStrategy : uses
Analysis ..> IPatternDetector : uses
Analysis ..> ISuggestionGenerator : uses
Core ..> IOutputFormatter : uses

TimeIdx ..|> IIndexer
ToolIdx ..|> IIndexer
SemIdx ..|> IIndexer
CustomIdx ..|> IIndexer

TempSearch ..|> ISearchStrategy
ToolSearch ..|> ISearchStrategy
SemSearch ..|> ISearchStrategy

ErrDet ..|> IPatternDetector
ToolDet ..|> IPatternDetector
CustomDet ..|> IPatternDetector

ConfigGen ..|> ISuggestionGenerator
PromptGen ..|> ISuggestionGenerator
CustomGen ..|> ISuggestionGenerator

JSON ..|> IOutputFormatter
MD ..|> IOutputFormatter

@enduml
```

---

### **7.2 渐进式实施路线图**

```plantuml
@startgantt
!theme plain
projectscale weekly

title cc-meta-cli 开发路线图

-- Phase 1: 基础设施 (2-3 周) --
[设计核心数据结构] lasts 3 days
[实现 JSONL 解析器] lasts 3 days
[构建 SQLite 索引模块] lasts 5 days
[时间序列索引] lasts 3 days
[工具调用索引] lasts 3 days
[角色分类索引] lasts 2 days
[实现基础查询引擎] lasts 5 days

-- Phase 2: 查询能力 (1-2 周) --
[实现多维度查询] lasts 4 days
[时间范围查询] lasts 2 days
[工具过滤查询] lasts 2 days
[复合查询优化] lasts 3 days
[输出格式化] lasts 2 days

-- Phase 3: 分析引擎 (2-3 周) --
[模式检测框架] lasts 5 days
[错误重复检测器] lasts 3 days
[工具滥用检测器] lasts 3 days
[工作流低效检测器] lasts 3 days
[建议生成引擎] lasts 5 days
[配置建议生成] lasts 3 days
[提示模板生成] lasts 3 days

-- Phase 4: 用户接口 (1-2 周) --
[Slash Command: /meta-suggest] lasts 2 days
[Slash Command: /meta-analyze] lasts 2 days
[Slash Command: /meta-apply] lasts 3 days
[Subagent: meta-coach] lasts 4 days
[Subagent: pattern-explorer] lasts 3 days

-- Phase 5: 高级功能 (2-3 周) --
[语义索引 (向量)] lasts 5 days
[MCP Server 实现] lasts 5 days
[工具: query_session_stats] lasts 2 days
[工具: search_similar] lasts 3 days
[工具: get_patterns] lasts 2 days
[增量索引优化] lasts 3 days
[缓存机制] lasts 2 days

-- Phase 6: 优化与发布 (1 周) --
[性能测试] lasts 2 days
[文档编写] lasts 3 days
[CI/CD 配置] lasts 2 days
[发布 v1.0] lasts 1 day

@endgantt
```

---

## **总结：简化原型的核心设计原则**

### **设计哲学**

```plantuml
@startmindmap
!theme plain
* cc-meta-cli
** 简洁性
*** 单一命令行工具
*** 标准化输出格式
*** 无额外后台服务
** 可集成性
*** 与 Claude Code 无缝对接
*** 通过 Slash Commands 调用
*** 通过 MCP 扩展功能
*** 通过 Subagents 深化交互
** 渐进式
*** 核心功能优先
*** 按需添加高级特性
*** 插件化架构
** 数据驱动
*** 基于实际会话历史
*** 多维度索引
*** 统计与模式识别
** 开发者友好
*** 清晰的 CLI 接口
*** 详细的输出信息
*** 易于调试
@endmindmap
```

---

### **关键特性矩阵**

| 特性 | 基础版 | 标准版 | 高级版 |
|------|--------|--------|--------|
| **索引能力** |
| 时间序列索引 | ✅ | ✅ | ✅ |
| 工具调用索引 | ✅ | ✅ | ✅ |
| 角色分类索引 | ✅ | ✅ | ✅ |
| 语义向量索引 | ❌ | ✅ | ✅ |
| **查询能力** |
| 基础过滤查询 | ✅ | ✅ | ✅ |
| 复合查询优化 | ❌ | ✅ | ✅ |
| 语义搜索 | ❌ | ✅ | ✅ |
| 跨会话聚合 | ❌ | ❌ | ✅ |
| **分析能力** |
| 统计摘要 | ✅ | ✅ | ✅ |
| 模式检测 | 基础 | ✅ | ✅ |
| 建议生成 | 手动 | 自动 | 自动+智能 |
| **接口** |
| CLI 命令 | ✅ | ✅ | ✅ |
| Slash Commands | 2 个 | 4 个 | 6+ 个 |
| Subagents | ❌ | 1 个 | 2+ 个 |
| MCP Server | ❌ | ❌ | ✅ |
| **性能** |
| 增量索引 | ✅ | ✅ | ✅ |
| 查询缓存 | ❌ | ✅ | ✅ |
| 并行处理 | ❌ | ❌ | ✅ |

这个简化原型通过清晰的模块划分、标准化的接口设计和渐进式的实施路径，为开发者提供了一个轻量但功能完整的元认知分析系统。