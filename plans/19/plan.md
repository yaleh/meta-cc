# Phase 19: Message Query Enhancement - TDD Implementation Plan

## Phase Overview

**Objective**: Implement assistant response querying and complete conversation querying capabilities with backward compatibility for existing `query_user_messages` interface.

**Code Volume**: ~600 lines | **Priority**: Medium | **Status**: Planning

**Dependencies**:
- Phase 0-18 (Complete meta-cc CLI + MCP Server + hybrid output mode)
- Phase 16 (Hybrid output mode for large queries)

**Deliverables**:
- Content serialization support for Message/ContentBlock types
- New CLI command: `meta-cc query assistant-messages`
- New CLI command: `meta-cc query conversation`
- Two new MCP tools: `query_assistant_messages`, `query_conversation` (14→16 tools)
- Updated documentation and examples

---

## Phase Objectives

### Core Problems

**Problem 1: Content Serialization Gap**
- Current: `Message.Content` marked `json:"-"` prevents serialization
- Impact: Cannot query assistant messages or export conversation data
- Root cause: ContentBlock structure contains ToolUse/ToolResult with complex types

**Problem 2: Limited Message Query Capabilities**
- Current: Only `query_user_messages` available
- Missing: Assistant response analysis (length, tool usage, token metrics)
- Missing: Conversation turn analysis (user+assistant pairing, response time)

**Problem 3: No Conversation Context**
- Current: User and assistant messages queried separately
- Need: Paired analysis for interaction patterns, iteration efficiency

### Solution Architecture

```
Phase 19 Implementation Strategy:

1. Serialization Layer (Stage 19.1)
   - Custom MarshalJSON for Message/ContentBlock
   - Preserve backward compatibility with existing parser

2. Assistant Query (Stage 19.2)
   - CLI: meta-cc query assistant-messages
   - MCP: query_assistant_messages
   - Data: AssistantMessage struct with metrics

3. Conversation Query (Stage 19.3)
   - CLI: meta-cc query conversation
   - MCP: query_conversation
   - Data: ConversationTurn (pairs user+assistant)

4. MCP Integration (Stage 19.4)
   - Add tools to MCP server (tools.go, executor.go)
   - Hybrid output mode support (inline/file_ref)

5. Documentation (Stage 19.5)
   - Update CLAUDE.md, examples-usage.md
   - MCP tool reference
```

### Design Principles

1. **Backward Compatibility**: Preserve existing `query_user_messages` interface
2. **Three-Tool Strategy**: Separate concerns (user, assistant, conversation)
3. **Delayed Decision**: Provide complete data, analysis to Claude/jq
4. **Hybrid Output**: Large queries automatically use file_ref mode
5. **Performance Optimization**: Dedicated tools avoid loading unnecessary data

---

## Success Criteria

**Functional Acceptance**:
- ✅ All stage unit tests pass (TDD methodology)
- ✅ Content serialization works for Message/ContentBlock types
- ✅ `query assistant-messages` CLI command functional
- ✅ `query conversation` CLI command functional
- ✅ Two new MCP tools integrated and tested
- ✅ Hybrid output mode supports all new tools
- ✅ Backward compatibility maintained for `query_user_messages`

**Integration Acceptance**:
- ✅ MCP server returns valid responses for new tools
- ✅ Large query results (>8KB) automatically use file_ref mode
- ✅ jq filtering works with new output formats
- ✅ CLI and MCP outputs are consistent

**Code Quality**:
- ✅ Total code: ~600 lines (within Phase 19 budget)
  - Stage 19.1: ~80 lines (serialization)
  - Stage 19.2: ~150 lines (assistant query)
  - Stage 19.3: ~200 lines (conversation query)
  - Stage 19.4: ~100 lines (MCP integration)
  - Stage 19.5: ~70 lines (documentation)
- ✅ Each stage ≤ 200 lines
- ✅ Test coverage: ≥ 80%
- ✅ `make all` passes after each stage

---

## Stage 19.1: Content Serialization Support

### Objective

Add custom JSON marshaling for `Message` and `ContentBlock` types to enable serialization while maintaining backward compatibility with existing parser logic.

### Acceptance Criteria

- [ ] `Message.MarshalJSON()` correctly serializes Content field
- [ ] `ContentBlock.MarshalJSON()` handles text/tool_use/tool_result types
- [ ] Existing parser logic unaffected (unmarshaling still works)
- [ ] Round-trip test: unmarshal → marshal → unmarshal yields same data
- [ ] Edge cases handled: empty content, nil fields, mixed content blocks
- [ ] Unit tests achieve ≥80% coverage

### TDD Approach

**Test File**: `internal/parser/types_serialization_test.go` (~40 lines)

```go
// Test functions:
// - TestMessageMarshalJSON - Basic message serialization
// - TestMessageMarshalJSONEmpty - Empty content handling
// - TestContentBlockMarshalJSON - All block types (text, tool_use, tool_result)
// - TestContentBlockMarshalJSONNil - Nil field handling
// - TestSerializationRoundTrip - Unmarshal → Marshal → Unmarshal
```

**Test Strategy**:
1. Create test fixtures with various content block combinations
2. Marshal to JSON, verify structure
3. Unmarshal back, verify data integrity
4. Test edge cases (nil, empty, mixed types)

**Implementation File**: `internal/parser/types.go` (~40 lines addition)

```go
// Add custom marshaling methods:
// - (m *Message) MarshalJSON() ([]byte, error)
// - (cb *ContentBlock) MarshalJSON() ([]byte, error)
// - (cb *ContentBlock) UnmarshalJSON(data []byte) error (enhance existing)

// Example implementation structure:
func (m *Message) MarshalJSON() ([]byte, error) {
    type Alias Message
    return json.Marshal(&struct {
        Content []ContentBlock `json:"content"`
        *Alias
    }{
        Content: m.Content,
        Alias:   (*Alias)(m),
    })
}

func (cb *ContentBlock) MarshalJSON() ([]byte, error) {
    // Handle type-specific serialization
    switch cb.Type {
    case "text":
        return json.Marshal(struct{Type, Text string}{cb.Type, cb.Text})
    case "tool_use":
        return json.Marshal(struct{Type string; ToolUse *ToolUse}{cb.Type, cb.ToolUse})
    case "tool_result":
        return json.Marshal(struct{Type string; ToolResult *ToolResult}{cb.Type, cb.ToolResult})
    }
}
```

### File Changes

**Modified Files**:
- `internal/parser/types.go` (+40 lines)
  - Add `MarshalJSON` for `Message`
  - Add `MarshalJSON` for `ContentBlock`
  - Enhance `UnmarshalJSON` for `ContentBlock` (if needed)

**New Files**:
- `internal/parser/types_serialization_test.go` (+40 lines)

### Test Commands

```bash
# Run Stage 19.1 tests
go test -v ./internal/parser -run TestMessage.*JSON
go test -v ./internal/parser -run TestContentBlock.*JSON
go test -v ./internal/parser -run TestSerialization.*

# Run full test suite
make test

# Verify no regressions in existing parser tests
go test -v ./internal/parser
```

### Testing Protocol

**After Implementation**:
1. Run `make all` to verify lint, test, build
2. If any errors occur, fix immediately
3. **HALT if errors persist after 2 fix attempts**

### Dependencies

None (foundation stage)

### Estimated Time

1 hour (40 lines implementation + 40 lines tests)

---

## Stage 19.2: Assistant Message Query

### Objective

Implement assistant message querying with support for content analysis, tool usage filtering, and token metrics.

### Acceptance Criteria

- [ ] CLI command `meta-cc query assistant-messages` functional
- [ ] Pattern matching works (regex on text content)
- [ ] Filtering by tool count (`--min-tools`, `--max-tools`)
- [ ] Filtering by token usage (`--min-tokens-output`)
- [ ] AssistantMessage struct includes all required fields
- [ ] JSONL and TSV output formats supported
- [ ] Deterministic sorting by turn_sequence
- [ ] Unit tests achieve ≥80% coverage
- [ ] Integration test with real session data

### TDD Approach

**Test File**: `cmd/query_assistant_messages_test.go` (~60 lines)

```go
// Test functions:
// - TestExtractAssistantMessages - Basic extraction
// - TestAssistantMessagesPatternMatching - Regex filtering
// - TestAssistantMessagesToolFiltering - Tool count filters
// - TestAssistantMessagesTokenFiltering - Token usage filters
// - TestAssistantMessagesSorting - Deterministic ordering
// - TestAssistantMessagesOutputFormats - JSONL/TSV output
```

**Test Strategy**:
1. Create mock session entries with assistant messages
2. Test extraction logic with various content blocks
3. Verify filtering by pattern, tool count, token usage
4. Test output format consistency
5. Integration test with fixture data

**Implementation File**: `cmd/query_assistant_messages.go` (~90 lines)

```go
// Core structures:
type AssistantMessage struct {
    TurnSequence  int            `json:"turn_sequence"`
    UUID          string         `json:"uuid"`
    Timestamp     string         `json:"timestamp"`
    ContentBlocks []ContentBlock `json:"content_blocks"`
    TextLength    int            `json:"text_length"`
    ToolUseCount  int            `json:"tool_use_count"`
    TokensInput   int            `json:"tokens_input"`
    TokensOutput  int            `json:"tokens_output"`
    StopReason    string         `json:"stop_reason,omitempty"`
}

// ContentBlock mirrors parser.ContentBlock but simplified
type ContentBlock struct {
    Type     string   `json:"type"`
    Text     string   `json:"text,omitempty"`
    ToolName string   `json:"tool_name,omitempty"`
}

// Core functions:
// - extractAssistantMessages(entries, turnIndex) []AssistantMessage
// - filterByPattern(messages, pattern) []AssistantMessage
// - filterByToolCount(messages, min, max) []AssistantMessage
// - filterByTokens(messages, minTokens) []AssistantMessage
// - sortAssistantMessages(messages, sortBy, reverse)
```

**CLI Flags**:
```bash
--pattern          # Regex pattern to match text content
--min-tools        # Minimum tool use count
--max-tools        # Maximum tool use count
--min-tokens-output # Minimum output tokens
--sort-by          # Sort field (turn_sequence, timestamp, tokens_output)
--reverse          # Reverse sort order
--limit            # Result limit
--offset           # Result offset
```

**Usage Examples**:
```bash
# Query all assistant messages
meta-cc query assistant-messages

# Find responses with specific pattern
meta-cc query assistant-messages --pattern "fix.*bug"

# Find tool-heavy responses
meta-cc query assistant-messages --min-tools 5

# Find large responses
meta-cc query assistant-messages --min-tokens-output 2000

# Combined filtering
meta-cc query assistant-messages --pattern "error" --min-tools 2 --limit 10
```

### File Changes

**New Files**:
- `cmd/query_assistant_messages.go` (+90 lines)
- `cmd/query_assistant_messages_test.go` (+60 lines)

**Modified Files**:
- `cmd/query.go` (+5 lines) - Register subcommand

### Test Commands

```bash
# Run Stage 19.2 tests
go test -v ./cmd -run TestExtractAssistantMessages
go test -v ./cmd -run TestAssistantMessages.*

# Integration test with fixture
meta-cc query assistant-messages --session tests/fixtures/sample-session.jsonl

# Test output formats
meta-cc query assistant-messages --output jsonl | jq
meta-cc query assistant-messages --output tsv | head

# Run full test suite
make test
```

### Testing Protocol

**After Implementation**:
1. Run `make all` to verify lint, test, build
2. Test with real session data (meta-cc project)
3. Verify output format consistency (JSONL/TSV)
4. **HALT if errors persist after 2 fix attempts**

### Dependencies

- Stage 19.1 (serialization support)

### Estimated Time

1.5 hours (90 lines implementation + 60 lines tests)

---

## Stage 19.3: Conversation Query

### Objective

Implement conversation turn querying that pairs user messages with assistant responses, including response time calculation.

### Acceptance Criteria

- [ ] CLI command `meta-cc query conversation` functional
- [ ] ConversationTurn correctly pairs user+assistant messages
- [ ] Duration calculation works (timestamp difference in ms)
- [ ] Turn range filtering works (`--start-turn`, `--end-turn`)
- [ ] Pattern matching on user or assistant content
- [ ] Handles incomplete turns (user without assistant, vice versa)
- [ ] JSONL and TSV output formats supported
- [ ] Deterministic sorting by turn_sequence
- [ ] Unit tests achieve ≥80% coverage
- [ ] Integration test with real session data

### TDD Approach

**Test File**: `cmd/query_conversation_test.go` (~80 lines)

```go
// Test functions:
// - TestBuildConversationTurns - Basic turn pairing
// - TestConversationTurnDuration - Duration calculation
// - TestConversationTurnFiltering - Turn range filters
// - TestConversationPatternMatching - Pattern on user/assistant
// - TestConversationIncompleteTurns - Handle missing parts
// - TestConversationSorting - Deterministic ordering
// - TestConversationOutputFormats - JSONL/TSV output
```

**Test Strategy**:
1. Create mock entries with user+assistant pairs
2. Test turn pairing logic with sequential entries
3. Verify duration calculation from timestamps
4. Test filtering by turn range and pattern
5. Handle edge cases (missing user/assistant, same-turn ordering)
6. Integration test with fixture data

**Implementation File**: `cmd/query_conversation.go` (~120 lines)

```go
// Core structures:
type ConversationTurn struct {
    TurnSequence     int               `json:"turn_sequence"`
    UserMessage      *UserMessage      `json:"user_message,omitempty"`
    AssistantMessage *AssistantMessage `json:"assistant_message,omitempty"`
    Duration         int               `json:"duration_ms"`      // Time from user to assistant
    Timestamp        string            `json:"timestamp"`        // Turn start time
}

// Reuse UserMessage from query_messages.go
// Reuse AssistantMessage from query_assistant_messages.go

// Core functions:
// - buildConversationTurns(entries, turnIndex) []ConversationTurn
// - calculateDuration(userTime, assistantTime) int
// - filterByTurnRange(turns, start, end) []ConversationTurn
// - filterByPattern(turns, pattern, target) []ConversationTurn
// - sortConversationTurns(turns, sortBy, reverse)
```

**CLI Flags**:
```bash
--start-turn       # Starting turn sequence
--end-turn         # Ending turn sequence
--pattern          # Regex pattern (applies to user or assistant)
--pattern-target   # Filter target: "user", "assistant", "any" (default: "any")
--min-duration     # Minimum response duration (ms)
--max-duration     # Maximum response duration (ms)
--sort-by          # Sort field (turn_sequence, duration, timestamp)
--reverse          # Reverse sort order
--limit            # Result limit
--offset           # Result offset
```

**Usage Examples**:
```bash
# Query all conversation turns
meta-cc query conversation

# Query specific turn range
meta-cc query conversation --start-turn 100 --end-turn 200

# Find slow responses
meta-cc query conversation --min-duration 30000

# Pattern matching on user input
meta-cc query conversation --pattern "fix.*bug" --pattern-target user

# Combined filtering
meta-cc query conversation --start-turn 50 --pattern "error" --min-duration 5000 --limit 20
```

### File Changes

**New Files**:
- `cmd/query_conversation.go` (+120 lines)
- `cmd/query_conversation_test.go` (+80 lines)

**Modified Files**:
- `cmd/query.go` (+5 lines) - Register subcommand

### Test Commands

```bash
# Run Stage 19.3 tests
go test -v ./cmd -run TestBuildConversationTurns
go test -v ./cmd -run TestConversation.*

# Integration test with fixture
meta-cc query conversation --session tests/fixtures/sample-session.jsonl

# Test turn pairing
meta-cc query conversation --start-turn 1 --end-turn 10 --output jsonl | jq '.[] | {turn: .turn_sequence, duration: .duration_ms}'

# Test pattern filtering
meta-cc query conversation --pattern "error" --pattern-target user

# Run full test suite
make test
```

### Testing Protocol

**After Implementation**:
1. Run `make all` to verify lint, test, build
2. Test with real session data (verify turn pairing)
3. Validate duration calculations (check timestamps)
4. Test edge cases (first turn, incomplete turns)
5. **HALT if errors persist after 2 fix attempts**

### Dependencies

- Stage 19.1 (serialization support)
- Stage 19.2 (AssistantMessage struct)

### Estimated Time

2 hours (120 lines implementation + 80 lines tests)

---

## Stage 19.4: MCP Tool Integration

### Objective

Integrate the two new query tools into the MCP server with hybrid output mode support.

### Acceptance Criteria

- [ ] `query_assistant_messages` tool registered in MCP server
- [ ] `query_conversation` tool registered in MCP server
- [ ] Tool descriptions accurate and complete
- [ ] Parameter validation works for all tool parameters
- [ ] Hybrid output mode supported (inline ≤8KB, file_ref >8KB)
- [ ] Integration tests pass for both tools
- [ ] Tool count increased from 14 to 16
- [ ] Backward compatibility maintained for existing tools

### TDD Approach

**Test File**: `cmd/mcp-server/tools_test.go` (additions ~40 lines)

```go
// Add test functions:
// - TestToolsListContainsNewTools - Verify tool registration
// - TestQueryAssistantMessagesToolDefinition - Tool schema validation
// - TestQueryConversationToolDefinition - Tool schema validation
```

**Test File**: `cmd/mcp-server/executor_test.go` (additions ~30 lines)

```go
// Add test functions:
// - TestExecuteQueryAssistantMessages - Basic execution
// - TestExecuteQueryConversation - Basic execution
// - TestNewToolsHybridOutput - Verify hybrid mode support
```

**Test File**: `cmd/mcp-server/integration_test.go` (additions ~30 lines)

```go
// Add test functions:
// - TestIntegrationQueryAssistantMessages - End-to-end test
// - TestIntegrationQueryConversation - End-to-end test
```

**Implementation Files**:

1. `cmd/mcp-server/tools.go` (+40 lines)

```go
// Add to listTools() function:
{
    Name: "query_assistant_messages",
    Description: "Query assistant messages with pattern matching, tool usage, and token filtering. Returns assistant responses with content blocks and metrics.",
    InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "pattern": {"type": "string", "description": "Regex pattern to match text content"},
            "min_tools": {"type": "integer", "description": "Minimum tool use count"},
            "max_tools": {"type": "integer", "description": "Maximum tool use count"},
            "min_tokens_output": {"type": "integer", "description": "Minimum output tokens"},
            "limit": {"type": "integer", "description": "Maximum results"},
            "scope": {"type": "string", "enum": ["session", "project"], "default": "project"},
            "stats_only": {"type": "boolean", "default": false},
            "stats_first": {"type": "boolean", "default": false},
            "jq_filter": {"type": "string", "description": "jq expression for filtering"},
            "output_format": {"type": "string", "enum": ["jsonl", "tsv"], "default": "jsonl"},
            "inline_threshold_bytes": {"type": "integer", "description": "Threshold for inline vs file_ref mode"}
        }
    }`),
},
{
    Name: "query_conversation",
    Description: "Query conversation turns (user+assistant pairs) with turn range, pattern matching, and duration filtering. Returns paired messages with response time.",
    InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "start_turn": {"type": "integer", "description": "Starting turn sequence"},
            "end_turn": {"type": "integer", "description": "Ending turn sequence"},
            "pattern": {"type": "string", "description": "Regex pattern (user or assistant)"},
            "pattern_target": {"type": "string", "enum": ["user", "assistant", "any"], "default": "any"},
            "min_duration": {"type": "integer", "description": "Minimum response duration (ms)"},
            "max_duration": {"type": "integer", "description": "Maximum response duration (ms)"},
            "limit": {"type": "integer", "description": "Maximum results"},
            "scope": {"type": "string", "enum": ["session", "project"], "default": "project"},
            "stats_only": {"type": "boolean", "default": false},
            "stats_first": {"type": "boolean", "default": false},
            "jq_filter": {"type": "string", "description": "jq expression for filtering"},
            "output_format": {"type": "string", "enum": ["jsonl", "tsv"], "default": "jsonl"},
            "inline_threshold_bytes": {"type": "integer", "description": "Threshold for inline vs file_ref mode"}
        }
    }`),
},
```

2. `cmd/mcp-server/executor.go` (+60 lines)

```go
// Add cases to executeTool() switch statement:

case "query_assistant_messages":
    args := []string{"query", "assistant-messages"}
    if pattern, ok := params["pattern"].(string); ok && pattern != "" {
        args = append(args, "--pattern", pattern)
    }
    if minTools, ok := params["min_tools"].(float64); ok {
        args = append(args, "--min-tools", fmt.Sprintf("%d", int(minTools)))
    }
    if maxTools, ok := params["max_tools"].(float64); ok {
        args = append(args, "--max-tools", fmt.Sprintf("%d", int(maxTools)))
    }
    if minTokens, ok := params["min_tokens_output"].(float64); ok {
        args = append(args, "--min-tokens-output", fmt.Sprintf("%d", int(minTokens)))
    }
    if limit, ok := params["limit"].(float64); ok && limit > 0 {
        args = append(args, "--limit", fmt.Sprintf("%d", int(limit)))
    }
    // ... additional parameter handling
    return executeCommand("query_assistant_messages", args, params)

case "query_conversation":
    args := []string{"query", "conversation"}
    if startTurn, ok := params["start_turn"].(float64); ok {
        args = append(args, "--start-turn", fmt.Sprintf("%d", int(startTurn)))
    }
    if endTurn, ok := params["end_turn"].(float64); ok {
        args = append(args, "--end-turn", fmt.Sprintf("%d", int(endTurn)))
    }
    if pattern, ok := params["pattern"].(string); ok && pattern != "" {
        args = append(args, "--pattern", pattern)
    }
    if target, ok := params["pattern_target"].(string); ok && target != "" {
        args = append(args, "--pattern-target", target)
    }
    if minDuration, ok := params["min_duration"].(float64); ok {
        args = append(args, "--min-duration", fmt.Sprintf("%d", int(minDuration)))
    }
    if maxDuration, ok := params["max_duration"].(float64); ok {
        args = append(args, "--max-duration", fmt.Sprintf("%d", int(maxDuration)))
    }
    if limit, ok := params["limit"].(float64); ok && limit > 0 {
        args = append(args, "--limit", fmt.Sprintf("%d", int(limit)))
    }
    // ... additional parameter handling
    return executeCommand("query_conversation", args, params)
```

### File Changes

**Modified Files**:
- `cmd/mcp-server/tools.go` (+40 lines)
- `cmd/mcp-server/executor.go` (+60 lines)
- `cmd/mcp-server/tools_test.go` (+40 lines)
- `cmd/mcp-server/executor_test.go` (+30 lines)
- `cmd/mcp-server/integration_test.go` (+30 lines)

### Test Commands

```bash
# Run MCP server tests
go test -v ./cmd/mcp-server -run TestToolsListContainsNewTools
go test -v ./cmd/mcp-server -run TestQueryAssistantMessages.*
go test -v ./cmd/mcp-server -run TestQueryConversation.*

# Integration tests
go test -v ./cmd/mcp-server -run TestIntegration.*

# Manual MCP testing (if MCP client available)
# Test query_assistant_messages
echo '{"method":"tools/call","params":{"name":"query_assistant_messages","arguments":{"pattern":"error","limit":5}}}' | meta-cc-mcp

# Test query_conversation
echo '{"method":"tools/call","params":{"name":"query_conversation","arguments":{"start_turn":1,"end_turn":10}}}' | meta-cc-mcp

# Run full test suite
make test
```

### Testing Protocol

**After Implementation**:
1. Run `make all` to verify lint, test, build
2. Test MCP tool registration (verify 16 tools listed)
3. Test parameter validation for both tools
4. Test hybrid output mode with large queries
5. **HALT if errors persist after 2 fix attempts**

### Dependencies

- Stage 19.2 (assistant query implementation)
- Stage 19.3 (conversation query implementation)
- Phase 16 (hybrid output mode)

### Estimated Time

1 hour (60 lines implementation + 40 lines tests)

---

## Stage 19.5: Documentation Updates

### Objective

Update all relevant documentation to reflect the new query capabilities and tool availability.

### Acceptance Criteria

- [ ] CLAUDE.md updated with new MCP tools (14→16)
- [ ] docs/mcp-output-modes.md updated with tool descriptions
- [ ] docs/examples-usage.md updated with usage examples
- [ ] docs/principles.md Section 7 matches implementation
- [ ] CHANGELOG.md updated with Phase 19 changes
- [ ] All documentation accurate and consistent

### Documentation Changes

**1. CLAUDE.md** (+20 lines)

```markdown
### Query Tools (16 available)

**Basic Queries**:
- `get_session_stats` - Session statistics and metrics
- `query_tools` - Filter tool calls by name, status
- `query_user_messages` - Search user messages with regex
- `query_assistant_messages` - Query assistant responses (NEW)
  - Parameters: `pattern`, `min_tools`, `max_tools`, `min_tokens_output`, `limit`
  - Use case: Analyze response length, tool usage patterns, token metrics
- `query_conversation` - Query conversation turns (NEW)
  - Parameters: `start_turn`, `end_turn`, `pattern`, `min_duration`, `limit`
  - Use case: Interaction patterns, response time analysis
- `query_files` - File operation statistics

**Advanced Queries**:
- ... (existing tools)
```

**2. docs/mcp-output-modes.md** (+15 lines)

Add to "Query Tools" section:

```markdown
#### query_assistant_messages

Query assistant responses with pattern matching and metrics:
- Pattern matching on text content (regex)
- Tool usage filtering (min/max tool count)
- Token filtering (minimum output tokens)
- Returns: AssistantMessage with content_blocks, tool_use_count, tokens

Example:
```
query_assistant_messages(pattern="fix.*bug", min_tools=2, limit=10)
```

#### query_conversation

Query conversation turns (user+assistant pairs):
- Turn range filtering (start_turn, end_turn)
- Pattern matching on user or assistant content
- Response time filtering (min/max duration)
- Returns: ConversationTurn with user_message, assistant_message, duration_ms

Example:
```
query_conversation(start_turn=100, end_turn=200, min_duration=5000)
```
```

**3. docs/examples-usage.md** (+20 lines)

Add new section "Querying Assistant Messages and Conversations":

```markdown
## Querying Assistant Messages and Conversations

### Assistant Message Analysis

```bash
# Find tool-heavy responses
meta-cc query assistant-messages --min-tools 5

# Find large responses
meta-cc query assistant-messages --min-tokens-output 2000

# Pattern matching
meta-cc query assistant-messages --pattern "error.*occurred" --output jsonl | jq
```

### Conversation Analysis

```bash
# Query turn range
meta-cc query conversation --start-turn 50 --end-turn 100

# Find slow responses
meta-cc query conversation --min-duration 30000

# Pattern matching on user input
meta-cc query conversation --pattern "fix bug" --pattern-target user

# Analyze response times
meta-cc query conversation --output jsonl | jq '.[] | {turn: .turn_sequence, duration: .duration_ms, user: .user_message.content[:50]}'
```

### MCP Tool Usage

```python
# Using query_assistant_messages
result = mcp_client.call_tool("query_assistant_messages", {
    "pattern": "error",
    "min_tools": 2,
    "limit": 10
})

# Using query_conversation
result = mcp_client.call_tool("query_conversation", {
    "start_turn": 1,
    "end_turn": 50,
    "min_duration": 5000
})
```
```

**4. docs/principles.md** (+10 lines)

Verify Section 7 matches implementation (should already be accurate based on design decisions).

**5. CHANGELOG.md** (+5 lines)

```markdown
## [Unreleased]

### Added (Phase 19)
- Content serialization support for Message and ContentBlock types
- CLI command: `meta-cc query assistant-messages` for assistant response analysis
- CLI command: `meta-cc query conversation` for conversation turn analysis
- MCP tools: `query_assistant_messages` and `query_conversation` (14→16 tools)
- Support for filtering by tool usage, token metrics, and response time
```

### File Changes

**Modified Files**:
- `CLAUDE.md` (+20 lines)
- `docs/mcp-output-modes.md` (+15 lines)
- `docs/examples-usage.md` (+20 lines)
- `docs/principles.md` (+10 lines verification)
- `CHANGELOG.md` (+5 lines)

### Verification Commands

```bash
# Check documentation consistency
grep -r "query_assistant_messages" docs/
grep -r "query_conversation" docs/
grep -r "14 tools" docs/ # Should be updated to "16 tools"

# Verify code references match docs
git diff docs/ CLAUDE.md CHANGELOG.md

# Check for broken links (if tooling available)
# markdown-link-check docs/*.md
```

### Testing Protocol

**After Documentation**:
1. Review all updated documentation for accuracy
2. Verify tool counts are consistent (16 tools)
3. Test examples in examples-usage.md
4. Check cross-references between documents
5. Verify CHANGELOG.md completeness

### Dependencies

- Stage 19.4 (MCP integration complete)

### Estimated Time

30 minutes (70 lines documentation updates)

---

## Phase Integration Strategy

### Build Verification

After completing all stages, verify the complete Phase 19 implementation:

```bash
# 1. Full build
make all

# 2. Unit tests
go test -v ./...

# 3. Integration tests
make test-integration

# 4. CLI functional tests
meta-cc query assistant-messages --session tests/fixtures/sample-session.jsonl
meta-cc query conversation --session tests/fixtures/sample-session.jsonl

# 5. MCP server tests
go test -v ./cmd/mcp-server

# 6. Test with real session data
meta-cc query assistant-messages --pattern "error" --min-tools 2 --limit 10
meta-cc query conversation --start-turn 100 --end-turn 200
```

### Backward Compatibility Verification

Ensure existing functionality remains intact:

```bash
# Test existing query_user_messages
meta-cc query user-messages --pattern "test"

# Test MCP server with existing tools
# (use MCP client if available)

# Verify output formats
meta-cc query user-messages --output jsonl | jq
meta-cc query assistant-messages --output tsv | head
meta-cc query conversation --output jsonl | jq
```

### Performance Benchmarks

Measure performance impact:

```bash
# Benchmark assistant query
time meta-cc query assistant-messages --limit 1000

# Benchmark conversation query
time meta-cc query conversation --limit 1000

# Compare with user messages query
time meta-cc query user-messages --limit 1000

# Verify hybrid output mode triggers correctly
# (Large queries should generate file_ref responses)
```

### Rollout Checklist

Before marking Phase 19 complete:

- [ ] All 5 stages completed and tested
- [ ] `make all` passes without errors
- [ ] Test coverage ≥80% (verify with `make test-coverage`)
- [ ] Documentation updated and accurate
- [ ] CLI commands functional and tested
- [ ] MCP tools integrated and tested
- [ ] Hybrid output mode works for new tools
- [ ] Backward compatibility verified
- [ ] Performance acceptable (no significant regression)
- [ ] CHANGELOG.md updated
- [ ] Git commit includes Phase 19 changes

---

## File Change Inventory

### Summary by Stage

| Stage | New Files | Modified Files | Total Lines |
|-------|-----------|----------------|-------------|
| 19.1  | 1         | 1              | ~80         |
| 19.2  | 2         | 1              | ~150        |
| 19.3  | 2         | 1              | ~200        |
| 19.4  | 0         | 5              | ~100        |
| 19.5  | 0         | 5              | ~70         |
| **Total** | **5** | **13** | **~600** |

### Detailed File Changes

**New Files (5)**:
1. `internal/parser/types_serialization_test.go` (40 lines)
2. `cmd/query_assistant_messages.go` (90 lines)
3. `cmd/query_assistant_messages_test.go` (60 lines)
4. `cmd/query_conversation.go` (120 lines)
5. `cmd/query_conversation_test.go` (80 lines)

**Modified Files (13)**:
1. `internal/parser/types.go` (+40 lines) - Serialization methods
2. `cmd/query.go` (+10 lines) - Register subcommands
3. `cmd/mcp-server/tools.go` (+40 lines) - Tool registration
4. `cmd/mcp-server/executor.go` (+60 lines) - Tool execution
5. `cmd/mcp-server/tools_test.go` (+40 lines) - Tool tests
6. `cmd/mcp-server/executor_test.go` (+30 lines) - Executor tests
7. `cmd/mcp-server/integration_test.go` (+30 lines) - Integration tests
8. `CLAUDE.md` (+20 lines) - Tool documentation
9. `docs/mcp-output-modes.md` (+15 lines) - Tool descriptions
10. `docs/examples-usage.md` (+20 lines) - Usage examples
11. `docs/principles.md` (+10 lines) - Verification
12. `CHANGELOG.md` (+5 lines) - Phase 19 entry

---

## Risk Assessment and Mitigation

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Serialization breaks existing parser | Low | High | Thorough testing of existing parser tests, backward compatibility verification |
| Content block structure complexity | Medium | Medium | Start with simple test cases, iterate on edge cases |
| Performance regression with large conversations | Medium | Medium | Benchmark before/after, optimize if needed |
| MCP tool parameter validation issues | Low | Medium | Comprehensive unit tests for parameter handling |
| Hybrid output mode interaction bugs | Low | Medium | Reuse Phase 16 patterns, test with large queries |
| Documentation inconsistency | Medium | Low | Cross-reference all docs, verify examples work |

### Contingency Plans

**If Stage 19.1 serialization fails**:
- Consider alternative approach: create separate DTO structs instead of custom marshaling
- Maintain backward compatibility by keeping existing types unchanged

**If Stage 19.2/19.3 exceed code limits**:
- Extract common logic into shared helper functions
- Simplify filtering logic (delegate more to jq)
- Split large functions into smaller units

**If performance regresses significantly**:
- Profile with `go test -bench` to identify bottlenecks
- Add caching for repeated queries
- Optimize turn index building (if bottleneck)

**If testing fails repeatedly**:
- HALT development per testing protocol
- Document blockers and failure state
- Seek code review or architectural consultation

---

## Testing Strategy

### Unit Testing

**Coverage Requirements**:
- Each stage: ≥80% coverage
- Critical paths: 100% coverage (serialization, turn pairing)
- Edge cases: Comprehensive test cases

**Test Organization**:
```
internal/parser/
  types_serialization_test.go  - Serialization tests

cmd/
  query_assistant_messages_test.go  - Assistant query tests
  query_conversation_test.go        - Conversation query tests

cmd/mcp-server/
  tools_test.go         - Tool registration tests
  executor_test.go      - Execution tests
  integration_test.go   - End-to-end tests
```

### Integration Testing

**Fixture-Based Tests**:
- Use `tests/fixtures/sample-session.jsonl`
- Test with real session data structure
- Verify output format consistency

**End-to-End Workflows**:
```bash
# Test CLI workflow
meta-cc query assistant-messages --pattern "error" > output.jsonl
jq '.[] | select(.tool_use_count > 2)' output.jsonl

# Test MCP workflow
# (via MCP client if available)
```

### Regression Testing

**Verify No Breaking Changes**:
```bash
# Existing query_user_messages should work unchanged
meta-cc query user-messages --pattern "test"

# Existing MCP tools should work unchanged
# (test all 14 existing tools)

# Output formats should remain consistent
meta-cc query user-messages --output tsv
```

### Performance Testing

**Benchmarks**:
```bash
# Measure query performance
go test -bench=. ./cmd/...

# Measure with large datasets
time meta-cc query conversation --limit 10000

# Compare with baseline (pre-Phase 19)
```

---

## Post-Phase Verification

### Functional Verification

After completing Phase 19, verify:

1. **Serialization Works**:
   ```bash
   # Test round-trip serialization
   meta-cc query assistant-messages --output jsonl | jq '.[] | .content_blocks'
   ```

2. **Assistant Query Works**:
   ```bash
   # Test pattern matching
   meta-cc query assistant-messages --pattern "error"

   # Test tool filtering
   meta-cc query assistant-messages --min-tools 5

   # Test token filtering
   meta-cc query assistant-messages --min-tokens-output 2000
   ```

3. **Conversation Query Works**:
   ```bash
   # Test turn range
   meta-cc query conversation --start-turn 1 --end-turn 10

   # Test duration filtering
   meta-cc query conversation --min-duration 10000

   # Test pattern matching
   meta-cc query conversation --pattern "bug" --pattern-target user
   ```

4. **MCP Tools Work**:
   ```bash
   # Test tool listing (should show 16 tools)
   echo '{"method":"tools/list"}' | meta-cc-mcp | jq '.result | length'

   # Test query_assistant_messages
   # (via MCP client)

   # Test query_conversation
   # (via MCP client)
   ```

5. **Hybrid Output Mode Works**:
   ```bash
   # Large queries should trigger file_ref mode
   # (test with >8KB expected output)
   meta-cc query conversation --limit 1000
   ```

### Documentation Verification

1. **Check Tool Count Consistency**:
   ```bash
   grep -r "14 tools" docs/ CLAUDE.md
   # Should find nothing (all updated to "16 tools")

   grep -r "16 tools" docs/ CLAUDE.md
   # Should find multiple references
   ```

2. **Verify Examples Work**:
   ```bash
   # Run examples from docs/examples-usage.md
   # Verify they produce expected output
   ```

3. **Check Cross-References**:
   ```bash
   # Verify all documentation cross-references are valid
   # Check links between CLAUDE.md, docs/plan.md, docs/principles.md
   ```

### Integration Verification

1. **Backward Compatibility**:
   - Existing `query_user_messages` works unchanged
   - All existing MCP tools work unchanged
   - Output formats remain consistent

2. **New Functionality**:
   - New CLI commands work as documented
   - New MCP tools return expected data
   - Hybrid output mode triggers appropriately

3. **Performance**:
   - No significant regression (<10% slowdown acceptable)
   - Large queries complete in reasonable time
   - Memory usage within acceptable bounds

---

## Success Metrics

### Quantitative Metrics

- **Code Quality**:
  - Test coverage ≥ 80%
  - Zero linting errors (`make lint`)
  - Zero test failures (`make test`)

- **Functionality**:
  - 2 new CLI commands functional
  - 2 new MCP tools functional
  - 16 total MCP tools available

- **Performance**:
  - Query execution time <500ms (typical session)
  - Large queries (>1000 results) complete <5s
  - Memory usage <100MB (typical query)

### Qualitative Metrics

- **Usability**:
  - Clear documentation for all new features
  - Intuitive CLI interface
  - Helpful error messages

- **Maintainability**:
  - Clean code structure
  - Comprehensive tests
  - Clear separation of concerns

- **Reliability**:
  - Handles edge cases gracefully
  - Provides meaningful errors
  - Backward compatible

---

## Timeline Estimate

| Stage | Description | Estimated Time |
|-------|-------------|----------------|
| 19.1  | Serialization support | 1 hour |
| 19.2  | Assistant query | 1.5 hours |
| 19.3  | Conversation query | 2 hours |
| 19.4  | MCP integration | 1 hour |
| 19.5  | Documentation | 0.5 hours |
| **Total** | **All stages** | **6 hours** |

**Contingency**: +2 hours for unexpected issues (total: 8 hours)

---

## Conclusion

Phase 19 implements comprehensive message query capabilities while maintaining backward compatibility and leveraging the hybrid output mode from Phase 16. The three-tool strategy (user, assistant, conversation) provides flexible analysis options while keeping implementation manageable within the 600-line Phase budget.

Key success factors:
- TDD methodology ensures high quality
- Staged approach minimizes risk
- Hybrid output mode handles large queries
- Backward compatibility preserved
- Clear documentation guides usage

Upon completion, meta-cc will provide complete visibility into Claude Code sessions, enabling powerful workflow analysis and optimization insights.
