# Phase 8: Query Foundation & Integration Improvements

## Overview

**Phase Name**: Query Foundation & Integration Improvements

**Goal**: Implement core `query` command capabilities (Stage 8.1-8.4) AND update existing integrations to leverage Phase 8 (Stage 8.5-8.7)

**Code Budget**:
- Core Implementation (8.1-8.4): ~400 lines (Go code)
- Integration Updates (8.5-8.7): ~250 lines (configuration/documentation)
- MCP Server Integration (8.8-8.9): ~120 lines (Go code + configuration)
- Context Query Extensions (8.10-8.11): ~280 lines (Go code)
- Prompt Optimization Data Layer (8.12): ~200 lines (Go code)
- **Total**: ~1250 lines

**Priority**: High (core query capability + immediate practical improvements + context support)

**Status**: ✅ Stages 8.1-8.7 Completed, 📋 Stages 8.8-8.12 Planned

**Design Principles**:
- ✅ **meta-cc 职责**: 数据提取、过滤、聚合、统计（无 LLM/NLP）
- ✅ **Claude 集成层职责**: 语义理解、上下文关联、建议生成
- ✅ **职责边界**: meta-cc 绝不做语义判断，只提供结构化数据

## Stage Breakdown

### Core Query Implementation (Completed)

#### Stage 8.1: Query Command Framework ✅
- **Objective**: Establish `query` command structure
- **Code**: ~100 lines
- **Deliverables**: `cmd/query.go` with routing

#### Stage 8.2: Query Tools Command ✅
- **Objective**: Implement `query tools` with filtering
- **Code**: ~120 lines
- **Deliverables**: `cmd/query_tools.go`, tool filtering, sorting

#### Stage 8.3: Query User-Messages Command ✅
- **Objective**: Implement `query user-messages` with regex
- **Code**: ~100 lines
- **Deliverables**: `cmd/query_messages.go`, regex pattern matching

#### Stage 8.4: Enhanced Filter Engine ✅
- **Objective**: Support `--where` syntax
- **Code**: ~80 lines
- **Deliverables**: Enhanced `internal/filter/` package

### Integration Improvements (New - Planned)

#### Stage 8.5: Update Slash Commands for Phase 8
- **Objective**: Update existing Slash Commands to use Phase 8 capabilities
- **Code**: ~50 lines (configuration changes)
- **Time**: 15-30 minutes
- **Deliverables**:
  - Update `/meta-timeline` to use `query tools --limit`
  - Verify `/meta-stats` already optimal
  - Avoid context overflow in large sessions

#### Stage 8.6: Update @meta-coach Documentation
- **Objective**: Enable @meta-coach to use Phase 8 query commands
- **Code**: ~80 lines (documentation)
- **Time**: 20-30 minutes
- **Deliverables**:
  - Add Phase 8 query capabilities section
  - Document iterative analysis pattern
  - Add best practices for large sessions
  - Include Phase 8 examples in coaching scenarios

#### Stage 8.7: Create New Query-Focused Slash Commands ✅
- **Objective**: Create specialized Slash Commands for quick queries
- **Code**: ~120 lines (2 new commands)
- **Time**: 30-45 minutes
- **Deliverables**:
  - `/meta-query-tools [tool] [status] [limit]` - Quick tool query
  - `/meta-query-messages [pattern] [limit]` - Message search

### Context Query Extensions (New - Planned)

#### Stage 8.10: 上下文和关联查询
- **Objective**: 实现上下文查询和关联查询功能
- **Code**: ~180 lines
- **Time**: 2-3 hours
- **Deliverables**:
  - `query context --error-signature <id> --window N`: 错误上下文查询
  - `query file-access --file <path>`: 文件操作历史
  - `query tool-sequences --min-occurrences N`: 工具序列模式
  - 时间窗口查询：`--since`, `--last-n-turns`
  - 为 Slash Commands 和 @meta-coach 提供精准上下文检索

#### Stage 8.11: 工作流模式数据支持
- **Objective**: 实现工作流模式检测功能
- **Code**: ~100 lines
- **Time**: 1-2 hours
- **Deliverables**:
  - `analyze sequences --min-length N --min-occurrences M`: 工具序列检测
  - `analyze file-churn --threshold N`: 文件频繁修改检测
  - `analyze idle-periods --threshold <duration>`: 时间间隔分析
  - 为 @meta-coach 提供工作流分析数据源（仅数据，不做语义判断）

#### Stage 8.12: Prompt 建议与优化数据层 (NEW)
- **Objective**: 为智能 Prompt 建议和改写提供数据检索能力
- **Code**: ~200 lines
- **Time**: 2-3 hours
- **Deliverables**:
  - 扩展 `query user-messages --with-context N`: 用户消息 + 上下文窗口
  - 新增 `query project-state`: 项目状态、未完成任务、最近文件
  - 新增 `query successful-prompts`: 历史成功 prompts 模式
  - 扩展 `query tool-sequences --successful-only --with-metrics`: 成功工作流
  - 新增 Slash Commands: `/meta-suggest-next`, `/meta-refine-prompt`
  - 增强 @meta-coach: Prompt 优化指导能力
- **职责边界**:
  - ✅ meta-cc: 数据检索（上下文、项目状态、成功模式）
  - ✅ Claude: 语义理解、prompt 生成、建议排序
  - ❌ meta-cc 绝不实现 NLP/LLM 能力

### MCP Server Integration (New - Planned)

#### Stage 8.8: Enhance MCP Server with Phase 8 Tools
- **Objective**: Update MCP Server to leverage Phase 8 query capabilities
- **Code**: ~100 lines (Go code, modify `cmd/mcp.go`)
- **Time**: 1-1.5 hours
- **Deliverables**:
  - Update `extract_tools` to use pagination (prevent overflow)
  - Add `query_tools` MCP tool (flexible querying)
  - Add `query_user_messages` MCP tool (regex search)
  - Test all MCP tools

#### Stage 8.9: Configure MCP Server to Claude Code
- **Objective**: Configure MCP Server and create usage documentation
- **Code**: ~20 lines (configuration) + ~100 lines (documentation)
- **Time**: 30 minutes
- **Deliverables**:
  - Create `.claude/mcp-servers/meta-cc.json` configuration
  - Create `docs/mcp-usage.md` documentation
  - Test MCP integration with natural language queries

## Architecture

### Core Query Flow (Stage 8.1-8.4)
```
meta-cc query <type> [filters] [options]
              │
              ├─ tools          → Query tool calls (8.2)
              ├─ user-messages  → Query user messages (8.3)
              └─ [future: sessions, errors]

┌─────────────┐    ┌──────────┐    ┌─────────┐    ┌────────┐    ┌────────┐
│   Locator   │───→│  Parser  │───→│ Querier │───→│ Filter │───→│ Output │
└─────────────┘    └──────────┘    └─────────┘    └────────┘    └────────┘
```

### Integration Layer (Stage 8.5-8.9)
```
User Interface Layer:
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│ Slash Commands   │  │   @meta-coach    │  │  New Commands    │  │   MCP Server     │
│ (Updated)        │  │   (Enhanced)     │  │  (Created)       │  │  (Enhanced)      │
│                  │  │                  │  │                  │  │                  │
│ /meta-timeline   │  │ Phase 8 aware    │  │ /meta-query-*    │  │ 5 MCP tools      │
│ /meta-stats      │  │ Iterative mode   │  │                  │  │ Natural language │
└────────┬─────────┘  └────────┬─────────┘  └────────┬─────────┘  └────────┬─────────┘
         │                     │                     │                     │
         └─────────────────────┼─────────────────────┼─────────────────────┘
                               ↓                     ↓
                    ┌──────────────────────┐  ┌──────────────────────┐
                    │  Phase 8 Query API   │  │  MCP Protocol Layer  │
                    │  query tools         │  │  JSON-RPC 2.0        │
                    │  query user-messages │  │  stdio transport     │
                    └──────────────────────┘  └──────────────────────┘
```

## Key Features

### Core Query Capabilities (8.1-8.4)
- ✅ Flexible tool call queries with filtering
- ✅ User message search with regex
- ✅ Sorting and limiting
- ✅ Enhanced `--where` filter syntax
- ✅ Pagination support

### Integration Improvements (8.5-8.9)
- ✅ Existing commands use Phase 8 (avoid context overflow)
- ✅ @meta-coach leverages new query capabilities
- ✅ Quick query commands for common tasks
- ✅ Better user experience with specialized commands
- 📋 MCP Server enhanced with Phase 8 tools
- 📋 Natural language queries enabled

## Usage Examples

### Core Query Commands (8.1-8.4)
```bash
# Query tool calls
meta-cc query tools --status error --limit 20
meta-cc query tools --tool Bash --sort-by timestamp
meta-cc query tools --where "tool=Edit,status=error"

# Query user messages
meta-cc query user-messages --match "fix.*bug"
meta-cc query user-messages --match "error|warning" --limit 10
```

### Context Query Commands (8.10-8.11)
```bash
# Query error context
meta-cc query context --error-signature err-a1b2 --window 3

# Query file access history
meta-cc query file-access --file test_auth.js --output json

# Query tool sequences
meta-cc query tool-sequences --min-occurrences 3

# Time window queries
meta-cc query tools --since "5 minutes ago"
meta-cc query tools --last-n-turns 10

# Workflow pattern detection
meta-cc analyze sequences --min-length 3 --min-occurrences 3
meta-cc analyze file-churn --threshold 5
meta-cc analyze idle-periods --threshold "5 minutes"
```

### Prompt Optimization Commands (8.12)
```bash
# Query user messages with context window
meta-cc query user-messages --match "实现|添加" --limit 5 --with-context 3 --output json

# Query current project state
meta-cc query project-state --include-incomplete-tasks --output json

# Query successful prompts patterns
meta-cc query successful-prompts --limit 10 --min-quality-score 0.8 --output json

# Query successful tool sequences
meta-cc query tool-sequences --successful-only --with-metrics --output json

# Use via Slash Commands
/meta-suggest-next                              # Get 3 prioritized prompt suggestions
/meta-refine-prompt "帮我优化一下代码"           # Refine vague prompt

# Use via @meta-coach
@meta-coach 我不知道下一步做什么                # Get guided prompt suggestions
@meta-coach 这个 prompt 写得对吗？              # Get prompt optimization feedback
```

### Updated Slash Commands (8.5)
```bash
# /meta-timeline now uses Phase 8
/meta-timeline          # Uses query tools --limit 50 (no overflow)
/meta-timeline 100      # Custom limit
```

### Enhanced @meta-coach (8.6)
```
@meta-coach 分析我的工作流

# Now uses:
# - query tools --limit 100 (efficient)
# - Iterative analysis pattern
# - Targeted queries (no context overflow)
```

### New Quick Commands (8.7)
```bash
# Quick tool query
/meta-query-tools Bash                    # All Bash calls
/meta-query-tools "" error                # All errors
/meta-query-tools Edit error 10           # Last 10 Edit errors

# Quick message search
/meta-query-messages "Phase 8"            # Find mentions
/meta-query-messages "error|bug"          # Regex search
/meta-query-messages "fix.*bug" 20        # Complex regex
```

### MCP Server Integration (8.8-8.9)
```bash
# Enhanced MCP tools (5 total)
mcp__meta-cc__get_session_stats          # Session statistics
mcp__meta-cc__analyze_errors             # Error analysis
mcp__meta-cc__extract_tools              # Tool extraction (with pagination)
mcp__meta-cc__query_tools                # Flexible tool queries (NEW)
mcp__meta-cc__query_user_messages        # Message search (NEW)

# Natural language queries (Claude calls MCP automatically)
"帮我查一下用了多少次 Bash 工具"
"搜索我提到 'Phase 8' 的消息"
"分析我的错误模式"
```

## Implementation Priority

### Must Do (Stage 8.1-8.4) ✅
Core query infrastructure - **COMPLETED**

### High Priority (Stage 8.5-8.6) ✅
- Stage 8.5: Update Slash Commands (15-30 min) ✅
  - Critical: Prevents context overflow in large sessions
  - Low risk: Minimal changes, high impact

- Stage 8.6: Update @meta-coach (20-30 min) ✅
  - Important: Enables better coaching
  - Demonstrates Phase 8 value

### Medium Priority (Stage 8.7) ✅
- Stage 8.7: New Quick Commands (30-45 min) ✅
  - Nice to have: Improves UX
  - Can be deferred if time-constrained

### High Priority (Stage 8.10-8.12) 📋
- Stage 8.10: 上下文和关联查询 (2-3 hours)
  - Critical: 为 Slash Commands/Subagent 提供上下文检索
  - Enables error context analysis
  - Supports file access history and tool sequences

- Stage 8.11: 工作流模式数据支持 (1-2 hours)
  - Important: 为 @meta-coach 提供工作流分析数据
  - Detects repetitive patterns
  - Identifies inefficient workflows

- Stage 8.12: Prompt 建议与优化数据层 (2-3 hours) **NEW**
  - Critical: 实现智能 prompt 建议和改写的数据基础
  - Enables `/meta-suggest-next` and `/meta-refine-prompt`
  - Enhances @meta-coach with prompt optimization capabilities
  - High user value: Improves development efficiency by 30%+

### Medium Priority (Stage 8.8-8.9) 📋
- Stage 8.8: Enhance MCP Server (1-1.5 hours)
  - Important: Completes MCP integration
  - Enables natural language queries
  - Prevents MCP context overflow

- Stage 8.9: Configure MCP Server (30 min)
  - Nice to have: Makes MCP discoverable
  - Documentation for users
  - Integration testing

## Testing Strategy

### Stage 8.5 Testing
```bash
# Test updated /meta-timeline
/meta-timeline          # Default limit
/meta-timeline 20       # Custom limit
# Verify: No context overflow in large sessions
```

### Stage 8.6 Testing
```
@meta-coach 分析我的工作流
# Verify: Uses query commands
# Verify: Demonstrates iterative pattern
```

### Stage 8.7 Testing
```bash
/meta-query-tools Bash
/meta-query-tools "" error
/meta-query-messages "Phase 8"
# Verify: Clear output, helpful tips
```

### Stage 8.8 Testing
```bash
# Test MCP tools manually
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./meta-cc mcp | jq '.result.tools | length'
# Expected: 5 tools

# Test query_tools
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"query_tools","arguments":{"tool":"Bash"}}}' | ./meta-cc mcp | jq .

# Test query_user_messages
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"query_user_messages","arguments":{"pattern":"Phase 8"}}}' | ./meta-cc mcp | jq .
```

### Stage 8.9 Testing
```bash
# Validate configuration
jq empty .claude/mcp-servers/meta-cc.json

# Test in Claude Code
"列出所有 MCP 工具"
"帮我查一下用了多少次 Bash 工具"
"搜索我提到 'Phase 8' 的消息"
```

### Stage 8.10 Testing
```bash
# Test context queries
./meta-cc query context --error-signature err-a1b2 --window 3 --output json | jq '.occurrences | length'

# Test file access
./meta-cc query file-access --file test_auth.js --output json | jq '.total_accesses'

# Test tool sequences
./meta-cc query tool-sequences --min-occurrences 3 --output json | jq '.sequences | length'

# Test time filters
./meta-cc query tools --since "10 minutes ago" --output json | jq '. | length'
```

### Stage 8.11 Testing
```bash
# Test workflow pattern detection
./meta-cc analyze sequences --min-length 3 --min-occurrences 3 --output json | jq '.sequences | length'
./meta-cc analyze file-churn --threshold 5 --output json | jq '.high_churn_files | length'
./meta-cc analyze idle-periods --threshold "5 minutes" --output json | jq '.idle_periods | length'
```

### Integration Testing
```bash
# Test all commands in a large session (>500 turns)
/meta-stats              # Should work (already optimal)
/meta-timeline 100       # Should use Phase 8
@meta-coach 帮我优化      # Should use iterative pattern
/meta-query-tools Bash   # Should work efficiently

# Test new context-aware commands (Stage 8.10-8.11)
/meta-error-context err-a1b2    # Should show error context
/meta-workflow-check            # Should use pattern detection

# Test MCP integration
"分析我的工作流"          # Claude should call MCP tools autonomously
"查看 test_auth.js 的修改历史"  # Should use query file-access
```

## Success Metrics

### Core Implementation (8.1-8.4) ✅
- ✅ All unit tests pass
- ✅ `query tools` and `query user-messages` work
- ✅ Filtering, sorting, limiting functional
- ✅ Performance < 100ms for typical sessions

### Integration Success (8.5-8.9)
- ✅ No context overflow in sessions >500 turns
- ✅ @meta-coach uses Phase 8 iterative pattern
- ✅ New commands provide clear, helpful output
- ✅ Users can perform common queries without CLI knowledge
- 📋 MCP Server enhanced with Phase 8 tools
- 📋 Natural language queries work seamlessly
- 📋 Claude can autonomously analyze workflows

### Context Query Success (8.10-8.11)
- 📋 Error context queries return complete surrounding turns
- 📋 File access history provides actionable insights
- 📋 Tool sequence detection identifies repetitive patterns
- 📋 Time window queries work accurately
- 📋 Workflow pattern detection provides data (no semantic judgments)
- 📋 @meta-coach uses context queries for deeper analysis
- 📋 Slash Commands leverage context data for better recommendations

## Dependencies

### Prerequisites
- ✅ Phase 0-7 completed (all infrastructure ready)
- ✅ Stage 8.1-8.4 completed (query commands available)
- ✅ `meta-cc` binary in PATH
- ✅ `jq` installed (for Slash Commands)

### Stage Dependencies
- Stage 8.5: Depends on 8.2 (query tools)
- Stage 8.6: Depends on 8.2, 8.3 (query tools, messages)
- Stage 8.7: Depends on 8.2, 8.3 (query tools, messages)
- Stage 8.8: Depends on 8.2, 8.3, 8.10 (query tools, messages, context queries), Phase 7 (MCP implementation)
- Stage 8.9: Depends on 8.8 (enhanced MCP server)
- Stage 8.10: Depends on 8.2 (query tools framework)
- Stage 8.11: Depends on 8.10 (context queries for idle-period context)

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Context overflow in large sessions | High | ✅ Stage 8.5 updates commands to use `--limit` |
| @meta-coach not using Phase 8 | Medium | ✅ Stage 8.6 adds documentation and examples |
| Users don't discover query commands | Low | ✅ Stage 8.7 provides easy-to-use Slash Commands |
| MCP Server missing Phase 8 tools | Medium | 📋 Stage 8.8 adds query tools to MCP |
| MCP Server not configured | Low | 📋 Stage 8.9 creates configuration and docs |
| Backward compatibility | Low | All changes are additive, old commands still work |

## Deliverables Checklist

### Core Implementation (8.1-8.4) ✅
- ✅ `cmd/query.go` - Command framework
- ✅ `cmd/query_tools.go` - Tool query implementation
- ✅ `cmd/query_messages.go` - Message query implementation
- ✅ `internal/filter/` enhancements - WHERE syntax
- ✅ Unit tests for all new code
- ✅ Integration tests passing

### Integration Updates (8.5-8.7) ✅
- ✅ `.claude/commands/meta-timeline.md` - Updated to use Phase 8
- ✅ `.claude/agents/meta-coach.md` - Phase 8 documentation added
- ✅ `.claude/commands/meta-query-tools.md` - New command created
- ✅ `.claude/commands/meta-query-messages.md` - New command created
- ✅ `README.md` or `docs/examples-usage.md` - Usage examples updated

### Context Query Extensions (8.10-8.11) 📋
- 📋 `cmd/query_context.go` - Context query implementation
- 📋 `cmd/query_file_access.go` - File access history
- 📋 `cmd/query_sequences.go` - Tool sequence queries
- 📋 `cmd/analyze_sequences.go` - Sequence detection
- 📋 `cmd/analyze_file_churn.go` - File churn detection
- 📋 `cmd/analyze_idle.go` - Idle period detection
- 📋 `internal/query/context.go` - Context query data structures
- 📋 `internal/analyzer/workflow.go` - Workflow pattern analysis
- 📋 `.claude/commands/meta-error-context.md` - New Slash Command
- 📋 `.claude/commands/meta-workflow-check.md` - New Slash Command
- 📋 Updated `.claude/agents/meta-coach.md` - Workflow analysis section

### MCP Integration (8.8-8.9) 📋
- 📋 `cmd/mcp.go` - Enhanced with Phase 8 tools
- 📋 `.claude/mcp-servers/meta-cc.json` - MCP configuration created
- 📋 `docs/mcp-usage.md` - MCP usage guide created

## Documentation Updates

### Files to Update
1. **README.md** - Add Phase 8 query examples
2. **docs/examples-usage.md** - Add quick query command guide
3. **docs/plan.md** - Update Phase 8 description

### New Documentation
- `plans/8/stage-8.5.md` - Slash Commands update plan ✅
- `plans/8/stage-8.6.md` - @meta-coach update plan ✅
- `plans/8/stage-8.7.md` - New commands plan ✅
- `plans/8/phase.md` - This overview ✅

## Next Steps

### Immediate (Stage 8.5-8.11)
1. ✅ Plan created (this document)
2. ✅ Execute Stage 8.5 (15-30 min): Update Slash Commands
3. ✅ Execute Stage 8.6 (20-30 min): Update @meta-coach
4. ✅ Execute Stage 8.7 (30-45 min): Create new commands
5. 📋 Execute Stage 8.10 (2-3h): 上下文和关联查询
6. 📋 Execute Stage 8.11 (1-2h): 工作流模式数据支持
7. 📋 Execute Stage 8.8 (1-1.5h): Enhance MCP Server
8. 📋 Execute Stage 8.9 (30 min): Configure MCP Server
9. 📋 Test all integrations (including context queries and MCP)
10. 📋 Update main documentation

### Future (Phase 9+)
- Phase 9: Context-Length Management (pagination, chunking)
- Phase 10: Advanced Query (aggregation, time-series)
- Phase 11: Unix Composability (streaming, exit codes)

## Related Documentation

- **Implementation Plan**: `plans/8/phase-8-implementation-plan.md`
- **Stage 8.5 Plan**: `plans/8/stage-8.5.md`
- **Stage 8.6 Plan**: `plans/8/stage-8.6.md`
- **Stage 8.7 Plan**: `plans/8/stage-8.7.md`
- **Stage 8.8 Plan**: `plans/8/stage-8.8.md`
- **Stage 8.9 Plan**: `plans/8/stage-8.9.md`
- **Stage 8.10 Plan**: `plans/8/stage-8.10.md` (NEW)
- **Stage 8.11 Plan**: `plans/8/stage-8.11.md` (NEW)
- **MCP Gap Analysis**: `/tmp/phase8-mcp-gap-analysis.md`
- **Integration Proposal**: `/tmp/meta-cc-integration-improvement-proposal.md`
- **Main Plan**: `docs/plan.md`
