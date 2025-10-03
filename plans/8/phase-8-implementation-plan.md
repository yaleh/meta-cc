# Phase 8: Query Foundation - Implementation Plan

## Phase Overview

**Phase Name**: Phase 8: 查询命令基础 (Query Foundation)

**Goal**: 实现 `meta-cc query` 命令组的核心查询能力，提供灵活的数据检索和过滤功能

**Code Budget**: ~400 lines (100 lines per Stage)

**Priority**: 高（核心检索能力）

**Status**: 待实施

## Context and Dependencies

### Completed Infrastructure (Phase 0-7)
- ✅ Session file location (`internal/locator/`)
- ✅ JSONL parsing (`internal/parser/`)
- ✅ Tool call extraction (`parser.ExtractToolCalls()`)
- ✅ Basic filtering (`internal/filter/`)
- ✅ Output formatting (`pkg/output/`)
- ✅ 66 passing unit tests

### Current Capabilities
- `meta-cc parse extract --type tools --filter "status=error"`
- `meta-cc parse stats`
- `meta-cc analyze errors`
- Filtering support: `key=value,key2=value2` format

### Phase 8 Additions
This phase introduces a **dedicated query command group** with:
1. Specialized subcommands for different data types
2. Enhanced filtering syntax (--where, --status, --tool)
3. Sorting and limiting capabilities
4. More intuitive CLI interface for data retrieval

## Design Principles

1. **Separation of Concerns**: `parse` for parsing, `query` for retrieval
2. **Unix Philosophy**: Composable, pipe-friendly output
3. **User-Friendly**: Clear, intuitive command structure
4. **Performance**: Efficient filtering at the data level
5. **Extensibility**: Easy to add new query types and filters

## Architecture

```
meta-cc query <type> [filters] [options]
              │
              ├─ tools          → Query tool calls
              ├─ user-messages  → Query user messages
              └─ [future: sessions, errors, files]

Query Flow:
┌─────────────┐    ┌──────────┐    ┌─────────┐    ┌────────┐    ┌────────┐
│   Locator   │───→│  Parser  │───→│ Querier │───→│ Filter │───→│ Output │
└─────────────┘    └──────────┘    └─────────┘    └────────┘    └────────┘
  (Phase 1)         (Phase 2)       (Phase 8)      (Phase 3)     (Phase 3)
```

## Stage Breakdown

### Stage 8.1: Query Command Framework and Routing

**Objective**: Establish `query` command structure with subcommand routing

**Code Estimate**: ~100 lines
- `cmd/query.go`: ~80 lines (main query command + init)
- `cmd/query_test.go`: ~20 lines (basic tests)

#### Test Scenarios (TDD)

**Test File**: `cmd/query_test.go`

```go
func TestQueryCommandExists(t *testing.T) {
    // Test: query command is registered
    // Expected: command exists and shows help
}

func TestQuerySubcommandRouting(t *testing.T) {
    // Test: query tools routes to correct handler
    // Test: query user-messages routes to correct handler
}

func TestQueryCommandHelp(t *testing.T) {
    // Test: meta-cc query --help shows subcommands
    // Expected: tools, user-messages listed
}

func TestQueryInvalidSubcommand(t *testing.T) {
    // Test: meta-cc query invalid-type
    // Expected: error with helpful message
}
```

#### Implementation Details

**File**: `cmd/query.go`

```go
package cmd

import (
    "github.com/spf13/cobra"
)

var (
    // Common query parameters
    queryLimit  int
    querySortBy string
    queryReverse bool
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
    Use:   "query",
    Short: "Query Claude Code session data",
    Long: `Query and retrieve specific data from Claude Code sessions.

The query command provides specialized subcommands for different data types:
  - tools:         Query tool calls with detailed filtering
  - user-messages: Query user messages with pattern matching

Examples:
  meta-cc query tools --status error --limit 20
  meta-cc query user-messages --match "fix.*bug"
  meta-cc query tools --tool Bash --sort-by timestamp`,
}

func init() {
    rootCmd.AddCommand(queryCmd)

    // Common query parameters
    queryCmd.PersistentFlags().IntVarP(&queryLimit, "limit", "l", 0, "Limit number of results (0 = no limit)")
    queryCmd.PersistentFlags().StringVarP(&querySortBy, "sort-by", "s", "", "Sort by field (timestamp, tool, status)")
    queryCmd.PersistentFlags().BoolVarP(&queryReverse, "reverse", "r", false, "Reverse sort order")
}
```

#### Acceptance Criteria
- ✅ `meta-cc query --help` displays subcommands
- ✅ `meta-cc query unknown` shows helpful error
- ✅ All tests pass (`go test ./cmd -run TestQuery`)
- ✅ Command structure documented in code comments

#### Dependencies
- None (builds on existing root command)

---

### Stage 8.2: Query Tools Command

**Objective**: Implement `query tools` with advanced filtering for tool calls

**Code Estimate**: ~120 lines
- `cmd/query_tools.go`: ~80 lines
- `cmd/query_tools_test.go`: ~40 lines

#### Test Scenarios (TDD)

**Test File**: `cmd/query_tools_test.go`

```go
func TestQueryToolsBasic(t *testing.T) {
    // Test: query tools with no filters
    // Expected: returns all tool calls as JSON
}

func TestQueryToolsFilterByStatus(t *testing.T) {
    // Test: query tools --status error
    // Expected: returns only failed tool calls
}

func TestQueryToolsFilterByTool(t *testing.T) {
    // Test: query tools --tool Bash
    // Expected: returns only Bash tool calls
}

func TestQueryToolsLimit(t *testing.T) {
    // Test: query tools --limit 10
    // Expected: returns max 10 results
}

func TestQueryToolsSortBy(t *testing.T) {
    // Test: query tools --sort-by timestamp --reverse
    // Expected: results sorted by timestamp descending
}

func TestQueryToolsCombinedFilters(t *testing.T) {
    // Test: query tools --status error --tool Edit --limit 5
    // Expected: returns up to 5 Edit errors
}
```

#### Implementation Details

**File**: `cmd/query_tools.go`

```go
package cmd

import (
    "fmt"
    "sort"

    "github.com/spf13/cobra"
    "github.com/yale/meta-cc/internal/locator"
    "github.com/yale/meta-cc/internal/parser"
    "github.com/yale/meta-cc/pkg/output"
)

var (
    queryToolsStatus string
    queryToolsTool   string
    queryToolsWhere  string
)

var queryToolsCmd = &cobra.Command{
    Use:   "tools",
    Short: "Query tool calls",
    Long: `Query tool calls with advanced filtering options.

Supports filtering by:
  - Tool name (--tool)
  - Status (--status: success|error)
  - General condition (--where: "field=value")

Examples:
  meta-cc query tools --status error
  meta-cc query tools --tool Bash --limit 20
  meta-cc query tools --where "status=error" --sort-by timestamp
  meta-cc query tools --tool Edit --status error --output md`,
    RunE: runQueryTools,
}

func init() {
    queryCmd.AddCommand(queryToolsCmd)

    // Tool-specific filters
    queryToolsCmd.Flags().StringVar(&queryToolsStatus, "status", "", "Filter by status (success|error)")
    queryToolsCmd.Flags().StringVar(&queryToolsTool, "tool", "", "Filter by tool name")
    queryToolsCmd.Flags().StringVar(&queryToolsWhere, "where", "", "Filter condition (key=value)")
}

func runQueryTools(cmd *cobra.Command, args []string) error {
    // Step 1: Locate and parse session
    loc := locator.NewSessionLocator()
    sessionPath, err := loc.Locate(locator.LocateOptions{
        SessionID:   sessionID,
        ProjectPath: projectPath,
    })
    if err != nil {
        return fmt.Errorf("failed to locate session: %w", err)
    }

    sessionParser := parser.NewSessionParser(sessionPath)
    entries, err := sessionParser.ParseEntries()
    if err != nil {
        return fmt.Errorf("failed to parse session: %w", err)
    }

    // Step 2: Extract tool calls
    toolCalls := parser.ExtractToolCalls(entries)

    // Step 3: Apply filters
    filtered := applyToolFilters(toolCalls)

    // Step 4: Sort if requested
    if querySortBy != "" {
        sortToolCalls(filtered, querySortBy, queryReverse)
    }

    // Step 5: Apply limit
    if queryLimit > 0 && len(filtered) > queryLimit {
        filtered = filtered[:queryLimit]
    }

    // Step 6: Format output
    outputStr, err := output.FormatJSON(filtered)
    if err != nil {
        return fmt.Errorf("failed to format output: %w", err)
    }

    fmt.Fprintln(cmd.OutOrStdout(), outputStr)
    return nil
}

func applyToolFilters(toolCalls []parser.ToolCall) []parser.ToolCall {
    var result []parser.ToolCall

    for _, tc := range toolCalls {
        // Apply status filter
        if queryToolsStatus != "" {
            if queryToolsStatus == "error" && tc.Status != "error" && tc.Error == "" {
                continue
            }
            if queryToolsStatus == "success" && (tc.Status == "error" || tc.Error != "") {
                continue
            }
        }

        // Apply tool name filter
        if queryToolsTool != "" && tc.ToolName != queryToolsTool {
            continue
        }

        // Apply where filter (basic implementation)
        if queryToolsWhere != "" {
            // Delegate to existing filter package
            // TODO: Enhance in Stage 8.4
        }

        result = append(result, tc)
    }

    return result
}

func sortToolCalls(toolCalls []parser.ToolCall, sortBy string, reverse bool) {
    sort.Slice(toolCalls, func(i, j int) bool {
        var less bool
        switch sortBy {
        case "timestamp":
            less = toolCalls[i].Timestamp < toolCalls[j].Timestamp
        case "tool":
            less = toolCalls[i].ToolName < toolCalls[j].ToolName
        case "status":
            less = toolCalls[i].Status < toolCalls[j].Status
        default:
            less = i < j // preserve original order
        }

        if reverse {
            return !less
        }
        return less
    })
}
```

#### Acceptance Criteria
- ✅ `meta-cc query tools` returns all tool calls
- ✅ `meta-cc query tools --status error` filters errors
- ✅ `meta-cc query tools --tool Bash` filters by tool name
- ✅ `meta-cc query tools --limit 10` limits results
- ✅ `meta-cc query tools --sort-by timestamp` sorts correctly
- ✅ All tests pass
- ✅ Works with real session files

#### Dependencies
- Stage 8.1 (query command framework)
- Existing: `parser.ExtractToolCalls()`, `output.FormatJSON()`

---

### Stage 8.3: Query User-Messages Command

**Objective**: Implement `query user-messages` with pattern matching

**Code Estimate**: ~100 lines
- `cmd/query_messages.go`: ~70 lines
- `cmd/query_messages_test.go`: ~30 lines

#### Test Scenarios (TDD)

**Test File**: `cmd/query_messages_test.go`

```go
func TestQueryUserMessagesBasic(t *testing.T) {
    // Test: query user-messages with no filters
    // Expected: returns all user messages
}

func TestQueryUserMessagesMatch(t *testing.T) {
    // Test: query user-messages --match "fix.*bug"
    // Expected: returns messages matching regex
}

func TestQueryUserMessagesLimit(t *testing.T) {
    // Test: query user-messages --limit 5
    // Expected: returns max 5 messages
}

func TestQueryUserMessagesSortByTimestamp(t *testing.T) {
    // Test: query user-messages --sort-by timestamp --reverse
    // Expected: messages sorted newest first
}

func TestQueryUserMessagesWithContext(t *testing.T) {
    // Test: query user-messages --match "error" --with-context
    // Expected: includes context (previous/next turns)
}
```

#### Implementation Details

**File**: `cmd/query_messages.go`

```go
package cmd

import (
    "fmt"
    "regexp"
    "sort"

    "github.com/spf13/cobra"
    "github.com/yale/meta-cc/internal/locator"
    "github.com/yale/meta-cc/internal/parser"
    "github.com/yale/meta-cc/pkg/output"
)

var (
    queryMessagesMatch      string
    queryMessagesWithContext bool
)

var queryUserMessagesCmd = &cobra.Command{
    Use:   "user-messages",
    Short: "Query user messages",
    Long: `Query user messages with pattern matching.

Supports:
  - Pattern matching (--match: regex pattern)
  - Timestamp sorting
  - Limit and pagination

Examples:
  meta-cc query user-messages --match "fix.*bug"
  meta-cc query user-messages --match "error" --limit 10
  meta-cc query user-messages --sort-by timestamp --reverse`,
    RunE: runQueryUserMessages,
}

func init() {
    queryCmd.AddCommand(queryUserMessagesCmd)

    queryUserMessagesCmd.Flags().StringVar(&queryMessagesMatch, "match", "", "Match pattern (regex)")
    queryUserMessagesCmd.Flags().BoolVar(&queryMessagesWithContext, "with-context", false, "Include surrounding context")
}

// UserMessage represents a user message with metadata
type UserMessage struct {
    UUID      string `json:"uuid"`
    Timestamp string `json:"timestamp"`
    Content   string `json:"content"`
    Context   string `json:"context,omitempty"` // Optional: previous/next messages
}

func runQueryUserMessages(cmd *cobra.Command, args []string) error {
    // Step 1: Locate and parse session
    loc := locator.NewSessionLocator()
    sessionPath, err := loc.Locate(locator.LocateOptions{
        SessionID:   sessionID,
        ProjectPath: projectPath,
    })
    if err != nil {
        return fmt.Errorf("failed to locate session: %w", err)
    }

    sessionParser := parser.NewSessionParser(sessionPath)
    entries, err := sessionParser.ParseEntries()
    if err != nil {
        return fmt.Errorf("failed to parse session: %w", err)
    }

    // Step 2: Extract user messages
    userMessages := extractUserMessages(entries)

    // Step 3: Apply pattern matching
    if queryMessagesMatch != "" {
        pattern, err := regexp.Compile(queryMessagesMatch)
        if err != nil {
            return fmt.Errorf("invalid regex pattern: %w", err)
        }

        var filtered []UserMessage
        for _, msg := range userMessages {
            if pattern.MatchString(msg.Content) {
                filtered = append(filtered, msg)
            }
        }
        userMessages = filtered
    }

    // Step 4: Sort if requested
    if querySortBy == "timestamp" {
        sortUserMessages(userMessages, queryReverse)
    }

    // Step 5: Apply limit
    if queryLimit > 0 && len(userMessages) > queryLimit {
        userMessages = userMessages[:queryLimit]
    }

    // Step 6: Format output
    outputStr, err := output.FormatJSON(userMessages)
    if err != nil {
        return fmt.Errorf("failed to format output: %w", err)
    }

    fmt.Fprintln(cmd.OutOrStdout(), outputStr)
    return nil
}

func extractUserMessages(entries []parser.SessionEntry) []UserMessage {
    var messages []UserMessage

    for _, entry := range entries {
        if entry.Type != "user" || entry.Message == nil {
            continue
        }

        // Extract text content from message
        var contentParts []string
        for _, block := range entry.Message.Content {
            if block.Type == "text" && block.Text != "" {
                contentParts = append(contentParts, block.Text)
            }
        }

        if len(contentParts) > 0 {
            messages = append(messages, UserMessage{
                UUID:      entry.UUID,
                Timestamp: entry.Timestamp,
                Content:   strings.Join(contentParts, "\n"),
            })
        }
    }

    return messages
}

func sortUserMessages(messages []UserMessage, reverse bool) {
    sort.Slice(messages, func(i, j int) bool {
        less := messages[i].Timestamp < messages[j].Timestamp
        if reverse {
            return !less
        }
        return less
    })
}
```

#### Acceptance Criteria
- ✅ `meta-cc query user-messages` returns all user messages
- ✅ `meta-cc query user-messages --match "fix"` filters by pattern
- ✅ Pattern matching uses regex (supports `.*`, `^`, `$`)
- ✅ `--limit` and `--sort-by` work correctly
- ✅ All tests pass
- ✅ Works with real session files

#### Dependencies
- Stage 8.1 (query command framework)
- Existing: `parser.SessionEntry`, `output.FormatJSON()`

---

### Stage 8.4: Basic Filter Engine Enhancement

**Objective**: Enhance filter engine to support `--where` syntax for complex conditions

**Code Estimate**: ~80 lines
- `internal/filter/filter.go`: ~50 lines (additions)
- `internal/filter/filter_test.go`: ~30 lines (new tests)

#### Test Scenarios (TDD)

**Test File**: `internal/filter/filter_test.go` (additions)

```go
func TestParseWhereCondition(t *testing.T) {
    // Test: "status=error"
    // Expected: single condition parsed

    // Test: "tool=Bash,status=error"
    // Expected: multiple conditions (AND logic)
}

func TestApplyWhereToToolCalls(t *testing.T) {
    // Test: Apply "status=error" to tool calls
    // Expected: only error calls returned

    // Test: Apply "tool=Edit,status=error"
    // Expected: only Edit errors returned
}

func TestWhereConditionInvalidSyntax(t *testing.T) {
    // Test: "invalid syntax"
    // Expected: error with helpful message
}

func TestWhereConditionFieldValidation(t *testing.T) {
    // Test: "invalid_field=value"
    // Expected: warning or error (configurable)
}
```

#### Implementation Details

**File**: `internal/filter/filter.go` (enhancements)

```go
// ParseWhereCondition parses a WHERE-style filter string
// Syntax: "field=value,field2=value2" (comma-separated AND conditions)
func ParseWhereCondition(where string) (*Filter, error) {
    // Reuse existing ParseFilter logic
    return ParseFilter(where)
}

// ValidateFilterField checks if a field name is valid for the data type
func ValidateFilterField(field string, dataType string) error {
    validFields := map[string][]string{
        "tool_calls": {"status", "tool", "uuid", "timestamp"},
        "entries":    {"type", "uuid", "role"},
    }

    allowed, ok := validFields[dataType]
    if !ok {
        return fmt.Errorf("unknown data type: %s", dataType)
    }

    for _, valid := range allowed {
        if field == valid {
            return nil
        }
    }

    return fmt.Errorf("invalid field '%s' for %s (valid: %v)", field, dataType, allowed)
}

// ApplyWhere is an alias for ApplyFilter with validation
func ApplyWhere(data interface{}, where string, dataType string) (interface{}, error) {
    filter, err := ParseWhereCondition(where)
    if err != nil {
        return nil, err
    }

    // Validate filter fields
    for _, cond := range filter.Conditions {
        if err := ValidateFilterField(cond.Field, dataType); err != nil {
            return nil, err
        }
    }

    return ApplyFilter(data, filter), nil
}
```

**Integration in `cmd/query_tools.go`**:

```go
// In applyToolFilters():
if queryToolsWhere != "" {
    filtered, err := filter.ApplyWhere(filtered, queryToolsWhere, "tool_calls")
    if err != nil {
        return nil, fmt.Errorf("invalid where condition: %w", err)
    }
    return filtered.([]parser.ToolCall), nil
}
```

#### Acceptance Criteria
- ✅ `--where "status=error"` filters correctly
- ✅ `--where "tool=Bash,status=error"` applies AND logic
- ✅ Invalid field names return helpful errors
- ✅ Filter validation prevents typos
- ✅ All tests pass
- ✅ Backward compatible with existing filter code

#### Dependencies
- Stage 8.2 (query tools command)
- Existing: `internal/filter/` package

---

## Integration Testing

### End-to-End Test Suite

**File**: `tests/integration/query_commands_test.sh`

```bash
#!/bin/bash
# Integration tests for Phase 8 query commands

set -e

PROJECT_ROOT="/home/yale/work/meta-cc"
META_CC="$PROJECT_ROOT/meta-cc"
TEST_SESSION="$PROJECT_ROOT/tests/fixtures/sample-session.jsonl"

# Test 1: Query tools basic
echo "Test 1: Query tools (no filters)"
$META_CC query tools --output json | jq '.[] | .tool_name' > /tmp/test1.out
test -s /tmp/test1.out || (echo "FAIL: No output"; exit 1)
echo "PASS"

# Test 2: Query tools with status filter
echo "Test 2: Query tools --status error"
$META_CC query tools --status error --output json | jq length > /tmp/test2.out
grep -E '^[0-9]+$' /tmp/test2.out || (echo "FAIL: Invalid output"; exit 1)
echo "PASS"

# Test 3: Query tools with limit
echo "Test 3: Query tools --limit 5"
COUNT=$($META_CC query tools --limit 5 --output json | jq length)
test "$COUNT" -le 5 || (echo "FAIL: Limit not applied"; exit 1)
echo "PASS"

# Test 4: Query user messages
echo "Test 4: Query user-messages"
$META_CC query user-messages --output json | jq '.[] | .content' > /tmp/test4.out
test -s /tmp/test4.out || (echo "FAIL: No messages"; exit 1)
echo "PASS"

# Test 5: Query user messages with pattern
echo "Test 5: Query user-messages --match 'test'"
$META_CC query user-messages --match "test" --output json | jq length > /tmp/test5.out
grep -E '^[0-9]+$' /tmp/test5.out || (echo "FAIL: Invalid output"; exit 1)
echo "PASS"

echo ""
echo "All integration tests passed!"
```

---

## Phase Completion Checklist

### Code Quality
- [ ] All unit tests pass (`go test ./...`)
- [ ] Integration tests pass (`./tests/integration/query_commands_test.sh`)
- [ ] Code coverage ≥ 80% for new code
- [ ] No linter warnings (`golangci-lint run`)

### Functionality
- [ ] `meta-cc query tools` works correctly
- [ ] `meta-cc query user-messages` works correctly
- [ ] Filtering (`--status`, `--tool`, `--where`) works
- [ ] Sorting (`--sort-by`, `--reverse`) works
- [ ] Limiting (`--limit`) works
- [ ] All output formats supported (json, md)

### Documentation
- [ ] README.md updated with query command examples
- [ ] Command help text is clear and accurate
- [ ] Integration tests documented
- [ ] Code comments explain complex logic

### Real-World Validation
- [ ] Test with meta-cc project session
- [ ] Test with NarrativeForge project session
- [ ] Test with large session (>1000 turns)
- [ ] Verify performance (query < 200ms for typical session)

---

## README.md Update

**Section to Add**: "Query Commands"

```markdown
## Query Commands

The `query` command provides flexible data retrieval from Claude Code sessions.

### Query Tool Calls

```bash
# Get all tool calls
meta-cc query tools

# Filter by status
meta-cc query tools --status error

# Filter by tool name
meta-cc query tools --tool Bash

# Combine filters
meta-cc query tools --tool Edit --status error --limit 10

# Sort and limit
meta-cc query tools --sort-by timestamp --reverse --limit 20

# Complex filtering
meta-cc query tools --where "tool=Bash,status=error"
```

### Query User Messages

```bash
# Get all user messages
meta-cc query user-messages

# Pattern matching (regex)
meta-cc query user-messages --match "fix.*bug"
meta-cc query user-messages --match "error|warning"

# Sort and limit
meta-cc query user-messages --sort-by timestamp --reverse --limit 5
```

### Common Options

- `--limit N`: Limit results to N items
- `--sort-by FIELD`: Sort by field (timestamp, tool, status)
- `--reverse`: Reverse sort order
- `--output FORMAT`: Output format (json, md)
```

---

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Query performance on large sessions | Medium | Implement streaming/pagination in Phase 9 |
| Filter syntax confusion | Low | Clear documentation and error messages |
| Code budget overrun | Medium | Each stage ≤100 lines, strict scope control |
| Regex pattern complexity | Low | Validate patterns, provide examples |
| Backward compatibility | Low | Keep existing `parse extract` functional |

---

## Success Metrics

- ✅ `meta-cc query tools --status error` < 100ms for typical session
- ✅ 100% of integration tests pass
- ✅ User can filter, sort, and limit results intuitively
- ✅ Query commands work across all 3 verified projects (meta-cc, NarrativeForge, claude-tmux)
- ✅ Code is maintainable and well-tested

---

## Next Phase Preview

**Phase 9: Context-Length Management** will address:
- Pagination (`--offset`, `--page`)
- Chunked output (`--chunk-size`, `--output-dir`)
- Field projection (`--fields "timestamp,tool,status"`)
- Compact formats (TSV, optimized CSV)
- Streaming output for large datasets

This builds on Phase 8's query foundation to handle large sessions efficiently.
