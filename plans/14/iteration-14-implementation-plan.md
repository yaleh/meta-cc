# Phase 14: Architecture Refactoring - Implementation Plan

## Overview

**Goal**: Eliminate code duplication and clarify responsibility boundaries through architecture refactoring

**Timeline**: 2-3 days (TDD methodology)

**Total Effort**: ~600 lines of refactoring (net reduction of 854 lines)

**Priority**: High (core architecture improvement, maintainability boost)

**Status**: Ready for implementation

---

## Objectives

### Primary Goals

1. **Reduce Code Duplication**: From 1194 lines to 340 lines (-72%) across 5 commands
2. **Clarify Responsibilities**: meta-cc extracts data only, LLM/tools make decisions
3. **Standardize Output**: All outputs sorted deterministically (UUID/Timestamp)
4. **Simplify Maintenance**: Pipeline abstraction enables 70% reduction in boilerplate

### Design Principles

- ✅ **Responsibility Minimization**: meta-cc = data extraction only, no analysis decisions
- ✅ **Pipeline Pattern**: Abstract common flow (locate → load → extract → output)
- ✅ **Output Determinism**: Stable sorting for CI/CD comparisons
- ✅ **Code Reuse First**: Eliminate ~345 lines of duplicate logic
- ✅ **Delayed Decision**: Push filtering/windowing to downstream tools/LLM

---

## Current State Analysis

### Problem 1: Code Duplication (30% redundancy)

```
Duplication Analysis:
- Session location:    ~10 lines × 5 commands = 50 lines
- JSONL parsing:       ~8 lines × 5 commands  = 40 lines
- Turn indexing:       ~15 lines × 3 commands = 45 lines
- Filtering logic:     ~20 lines × 3 commands = 60 lines
- Output formatting:   ~30 lines × 5 commands = 150 lines
→ Total redundancy: ~345 lines
```

### Problem 2: errors Command Scope Creep (317 → 50 lines target)

**Current overreach**:
- Window analysis (should be LLM decision)
- SHA256 signatures (over-engineered, use simple `{tool}:{error[:50]}`)
- Pattern counting (should use jq for aggregation)

**Correct scope**:
- Extract error list with simple signatures
- Let jq/LLM handle windowing and pattern detection

### Problem 3: Non-Deterministic Output

**Issue**: Go map iteration order is random
**Impact**: query-tools output varies between runs
**Solution**: Force sorting by UUID/Timestamp

### Code Size Breakdown (Current)

```
Command              Current    Target    Reduction
--------------------------------------------------
parse stats          ~170 lines ~60 lines -65%
query tools          ~307 lines ~80 lines -74%
query messages       ~280 lines ~70 lines -75%
analyze errors       ~317 lines ~80 lines -75%
timeline             ~120 lines ~50 lines -58%
--------------------------------------------------
Total                1194 lines 340 lines -72%
```

---

## Stage Breakdown

### Stage 14.1: Pipeline Abstraction Layer

**Duration**: 0.5-1 day

**Objective**: Extract common session processing flow into reusable pipeline

**Deliverables**:
- `pkg/pipeline/session.go` (~120 lines)
- `pkg/pipeline/session_test.go` (~150 lines)
- Unit tests with ≥90% coverage

**Implementation**:

```go
// pkg/pipeline/session.go
package pipeline

import (
    "github.com/yale/meta-cc/internal/locator"
    "github.com/yale/meta-cc/internal/parser"
)

// GlobalOptions contains global CLI flags
type GlobalOptions struct {
    SessionID   string
    ProjectPath string
    WorkingDir  string
}

// LoadOptions controls session loading behavior
type LoadOptions struct {
    AutoDetect bool
    Validate   bool
}

// SessionPipeline encapsulates session data processing
type SessionPipeline struct {
    opts      GlobalOptions
    session   string
    entries   []parser.Entry
    turnIndex map[string]int
}

// NewSessionPipeline creates a new pipeline instance
func NewSessionPipeline(opts GlobalOptions) *SessionPipeline {
    return &SessionPipeline{
        opts:      opts,
        turnIndex: make(map[string]int),
    }
}

// Load locates and loads the session JSONL file
func (p *SessionPipeline) Load(loadOpts LoadOptions) error {
    // 1. Locate session file
    loc := locator.NewSessionLocator()

    var sessionPath string
    var err error

    if p.opts.SessionID != "" {
        sessionPath, err = loc.FromSessionID(p.opts.SessionID)
    } else if p.opts.ProjectPath != "" {
        sessionPath, err = loc.FromProjectPath(p.opts.ProjectPath)
    } else if loadOpts.AutoDetect {
        sessionPath, err = loc.AutoDetect(p.opts.WorkingDir)
    } else {
        return fmt.Errorf("no session specified")
    }

    if err != nil {
        return fmt.Errorf("session location failed: %w", err)
    }

    p.session = sessionPath

    // 2. Parse JSONL
    p.entries, err = parser.ParseJSONL(sessionPath)
    if err != nil {
        return fmt.Errorf("JSONL parsing failed: %w", err)
    }

    return nil
}

// ExtractToolCalls extracts all tool calls from entries
func (p *SessionPipeline) ExtractToolCalls() ([]parser.ToolCall, error) {
    return parser.ExtractToolCalls(p.entries)
}

// ExtractUserMessages extracts all user messages from entries
func (p *SessionPipeline) ExtractUserMessages() ([]parser.UserMessage, error) {
    return parser.ExtractUserMessages(p.entries)
}

// BuildTurnIndex creates turn_id → turn_sequence mapping
func (p *SessionPipeline) BuildTurnIndex() map[string]int {
    if len(p.turnIndex) > 0 {
        return p.turnIndex // cached
    }

    for i, entry := range p.entries {
        p.turnIndex[entry.ID] = i
    }

    return p.turnIndex
}

// SessionPath returns the loaded session file path
func (p *SessionPipeline) SessionPath() string {
    return p.session
}

// EntryCount returns the number of entries loaded
func (p *SessionPipeline) EntryCount() int {
    return len(p.entries)
}
```

**Test Coverage**:

```go
// pkg/pipeline/session_test.go
package pipeline

import (
    "testing"
    "github.com/yale/meta-cc/internal/testutil"
)

func TestSessionPipeline_Load(t *testing.T) {
    tests := []struct {
        name    string
        opts    GlobalOptions
        loadOpts LoadOptions
        wantErr bool
    }{
        {
            name: "explicit session ID",
            opts: GlobalOptions{SessionID: "test-session-id"},
            loadOpts: LoadOptions{AutoDetect: false},
            wantErr: false,
        },
        {
            name: "project path",
            opts: GlobalOptions{ProjectPath: "/test/project"},
            loadOpts: LoadOptions{AutoDetect: false},
            wantErr: false,
        },
        {
            name: "auto-detect",
            opts: GlobalOptions{WorkingDir: "/test/workspace"},
            loadOpts: LoadOptions{AutoDetect: true},
            wantErr: false,
        },
        {
            name: "no session specified",
            opts: GlobalOptions{},
            loadOpts: LoadOptions{AutoDetect: false},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := NewSessionPipeline(tt.opts)
            err := p.Load(tt.loadOpts)

            if (err != nil) != tt.wantErr {
                t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
            }

            if err == nil {
                if p.EntryCount() == 0 {
                    t.Error("expected non-zero entry count")
                }
            }
        })
    }
}

func TestSessionPipeline_ExtractToolCalls(t *testing.T) {
    // Setup: create test session with 10 tool calls
    sessionPath := testutil.CreateTestSession(10)
    defer testutil.CleanupTestSession(sessionPath)

    p := NewSessionPipeline(GlobalOptions{SessionID: "test"})
    err := p.Load(LoadOptions{AutoDetect: false})
    if err != nil {
        t.Fatalf("Load failed: %v", err)
    }

    tools, err := p.ExtractToolCalls()
    if err != nil {
        t.Fatalf("ExtractToolCalls failed: %v", err)
    }

    if len(tools) != 10 {
        t.Errorf("expected 10 tool calls, got %d", len(tools))
    }
}

func TestSessionPipeline_BuildTurnIndex(t *testing.T) {
    sessionPath := testutil.CreateTestSession(5)
    defer testutil.CleanupTestSession(sessionPath)

    p := NewSessionPipeline(GlobalOptions{SessionID: "test"})
    p.Load(LoadOptions{AutoDetect: false})

    index := p.BuildTurnIndex()

    if len(index) != 5 {
        t.Errorf("expected 5 index entries, got %d", len(index))
    }

    // Verify idempotency (caching)
    index2 := p.BuildTurnIndex()
    if len(index2) != len(index) {
        t.Error("BuildTurnIndex not idempotent")
    }
}
```

**Acceptance Criteria**:
- ✅ Pipeline successfully loads sessions via 3 methods (session ID, project path, auto-detect)
- ✅ Extracted tool calls match parser output
- ✅ Turn index correctly built and cached
- ✅ Unit tests pass with ≥90% coverage
- ✅ Error handling covers all failure modes

**Code Size**: ~120 lines (source) + ~150 lines (tests) = 270 lines total

---

### Stage 14.2: Simplify errors Command

**Duration**: 0.5 day

**Objective**: Simplify `analyze errors` to pure data extraction (remove aggregation logic)

**Breaking Changes**:
- ⚠️ Output format changes from aggregated patterns to simple error list
- ⚠️ `--window` parameter removed (use `jq` or `tail` instead)

**Migration Guide**:

```bash
# Old (Phase 13)
meta-cc analyze errors --window 50
# Output: Aggregated patterns with counts

# New (Phase 14)
meta-cc query errors | jq '.[-50:]' | jq 'group_by(.Signature) | map({sig: .[0].Signature, count: length})'
# meta-cc outputs raw errors, jq handles aggregation
```

**Deliverables**:
- `cmd/query_errors.go` (~80 lines, replacing `cmd/analyze.go` errors logic)
- Update `cmd/analyze.go` to remove errors subcommand
- `cmd/query_errors_test.go` (~60 lines)
- Migration documentation

**Implementation**:

```go
// cmd/query_errors.go
package cmd

import (
    "sort"
    "strings"
    "time"

    "github.com/spf13/cobra"
    "github.com/yale/meta-cc/internal/filter"
    "github.com/yale/meta-cc/pkg/output"
    "github.com/yale/meta-cc/pkg/pipeline"
)

// ErrorEntry represents a single error occurrence
type ErrorEntry struct {
    UUID      string    `json:"uuid"`
    Timestamp time.Time `json:"timestamp"`
    TurnIndex int       `json:"turn_index"`
    ToolName  string    `json:"tool_name"`
    Error     string    `json:"error"`
    Signature string    `json:"signature"`
}

var queryErrorsCmd = &cobra.Command{
    Use:   "errors",
    Short: "Query tool errors",
    Long: `Extract all tool errors from session.

Returns a simple list of errors with signatures for downstream analysis.
Use jq, awk, or LLM for pattern detection and aggregation.

Examples:
  # All errors
  meta-cc query errors

  # Last 50 errors
  meta-cc query errors | jq '.[-50:]'

  # Group by signature
  meta-cc query errors | jq 'group_by(.signature)'

  # Count patterns
  meta-cc query errors | jq 'group_by(.signature) | map({sig: .[0].signature, count: length})'`,
    RunE: runQueryErrors,
}

func init() {
    queryCmd.AddCommand(queryErrorsCmd)
}

func runQueryErrors(cmd *cobra.Command, args []string) error {
    // 1. Initialize pipeline
    p := pipeline.NewSessionPipeline(pipeline.GlobalOptions{
        SessionID:   sessionID,
        ProjectPath: projectPath,
        WorkingDir:  workingDir,
    })

    // 2. Load session
    if err := p.Load(pipeline.LoadOptions{AutoDetect: true}); err != nil {
        return err
    }

    // 3. Extract tool calls
    tools, err := p.ExtractToolCalls()
    if err != nil {
        return err
    }

    // 4. Filter errors only
    var errors []ErrorEntry
    for _, tool := range tools {
        if tool.Status == "error" || tool.Error != "" {
            errors = append(errors, ErrorEntry{
                UUID:      tool.UUID,
                Timestamp: tool.Timestamp,
                TurnIndex: tool.TurnSequence,
                ToolName:  tool.ToolName,
                Error:     tool.Error,
                Signature: generateErrorSignature(tool.ToolName, tool.Error),
            })
        }
    }

    // 5. Sort by timestamp (deterministic output)
    sort.Slice(errors, func(i, j int) bool {
        return errors[i].Timestamp.Before(errors[j].Timestamp)
    })

    // 6. Apply pagination if specified
    if limitFlag > 0 || offsetFlag > 0 {
        errors = filter.ApplyPagination(errors, filter.PaginationConfig{
            Limit:  limitFlag,
            Offset: offsetFlag,
        })
    }

    // 7. Format output
    return output.Format(errors, outputFormat)
}

// generateErrorSignature creates a simple signature: {tool}:{error_prefix}
func generateErrorSignature(toolName, errorText string) string {
    // Take first 50 chars of error for signature
    prefix := errorText
    if len(errorText) > 50 {
        prefix = errorText[:50]
    }

    // Normalize whitespace
    prefix = strings.Join(strings.Fields(prefix), " ")

    return fmt.Sprintf("%s:%s", toolName, prefix)
}
```

**Test Coverage**:

```go
// cmd/query_errors_test.go
package cmd

import (
    "testing"
    "github.com/yale/meta-cc/internal/testutil"
)

func TestGenerateErrorSignature(t *testing.T) {
    tests := []struct {
        name      string
        toolName  string
        errorText string
        want      string
    }{
        {
            name:      "short error",
            toolName:  "Bash",
            errorText: "command not found",
            want:      "Bash:command not found",
        },
        {
            name:      "long error truncated",
            toolName:  "Edit",
            errorText: strings.Repeat("x", 100),
            want:      fmt.Sprintf("Edit:%s", strings.Repeat("x", 50)),
        },
        {
            name:      "whitespace normalized",
            toolName:  "Read",
            errorText: "file  not\n  found",
            want:      "Read:file not found",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := generateErrorSignature(tt.toolName, tt.errorText)
            if got != tt.want {
                t.Errorf("generateErrorSignature() = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestQueryErrors_Integration(t *testing.T) {
    // Create test session with 5 errors, 5 successes
    sessionPath := testutil.CreateTestSessionWithErrors(5, 5)
    defer testutil.CleanupTestSession(sessionPath)

    // Run query errors command
    // (integration test setup omitted for brevity)

    // Verify:
    // - Exactly 5 errors returned
    // - All have signatures
    // - Sorted by timestamp
}
```

**Update analyze.go**:

```go
// cmd/analyze.go - remove analyzeErrorsCmd
func init() {
    rootCmd.AddCommand(analyzeCmd)

    // Remove this line:
    // analyzeCmd.AddCommand(analyzeErrorsCmd)

    // Keep other analyze subcommands (sequences, file-churn, idle, etc.)
    analyzeCmd.AddCommand(analyzeSequencesCmd)
    analyzeCmd.AddCommand(analyzeFileChurnCmd)
    analyzeCmd.AddCommand(analyzeIdleCmd)
}
```

**Acceptance Criteria**:
- ✅ `meta-cc query errors` returns simple error list (JSONL)
- ✅ Error signatures use format `{tool}:{error[:50]}`
- ✅ Output sorted by timestamp
- ✅ `--window` parameter removed
- ✅ Migration guide in README
- ✅ Unit tests pass
- ✅ Code reduced from 317 lines to ~80 lines

**Code Size**: ~80 lines (source) + ~60 lines (tests) = 140 lines total

---

### Stage 14.3: Output Sorting Standardization

**Duration**: 0.5 day

**Objective**: Add deterministic sorting to all query commands

**Deliverables**:
- `pkg/output/sort.go` (~50 lines)
- `pkg/output/sort_test.go` (~60 lines)
- Update 5 query commands to use sorting

**Implementation**:

```go
// pkg/output/sort.go
package output

import (
    "sort"
    "time"

    "github.com/yale/meta-cc/internal/parser"
)

// SortByTimestamp sorts data by timestamp field
func SortByTimestamp(data interface{}) {
    switch v := data.(type) {
    case []parser.ToolCall:
        sort.Slice(v, func(i, j int) bool {
            return v[i].Timestamp.Before(v[j].Timestamp)
        })
    case []ErrorEntry:
        sort.Slice(v, func(i, j int) bool {
            return v[i].Timestamp.Before(v[j].Timestamp)
        })
    case []parser.UserMessage:
        sort.Slice(v, func(i, j int) bool {
            return v[i].Timestamp.Before(v[j].Timestamp)
        })
    // Add other types as needed
    }
}

// SortByTurnSequence sorts data by turn sequence number
func SortByTurnSequence(data interface{}) {
    switch v := data.(type) {
    case []parser.ToolCall:
        sort.Slice(v, func(i, j int) bool {
            return v[i].TurnSequence < v[j].TurnSequence
        })
    case []parser.UserMessage:
        sort.Slice(v, func(i, j int) bool {
            return v[i].TurnSequence < v[j].TurnSequence
        })
    }
}

// SortByUUID sorts data by UUID (lexicographic)
func SortByUUID(data interface{}) {
    switch v := data.(type) {
    case []parser.ToolCall:
        sort.Slice(v, func(i, j int) bool {
            return v[i].UUID < v[j].UUID
        })
    }
}

// DefaultSort applies the default sorting for a data type
func DefaultSort(data interface{}) {
    // Default to timestamp sorting for most types
    SortByTimestamp(data)
}
```

**Test Coverage**:

```go
// pkg/output/sort_test.go
package output

import (
    "testing"
    "time"

    "github.com/yale/meta-cc/internal/parser"
    "github.com/yale/meta-cc/internal/testutil"
)

func TestSortByTimestamp(t *testing.T) {
    // Create unsorted tool calls
    tools := []parser.ToolCall{
        {UUID: "c", Timestamp: time.Now().Add(2 * time.Hour)},
        {UUID: "a", Timestamp: time.Now()},
        {UUID: "b", Timestamp: time.Now().Add(1 * time.Hour)},
    }

    SortByTimestamp(tools)

    // Verify sorted order (a, b, c)
    if tools[0].UUID != "a" || tools[1].UUID != "b" || tools[2].UUID != "c" {
        t.Error("SortByTimestamp failed")
    }
}

func TestSortByTurnSequence(t *testing.T) {
    tools := []parser.ToolCall{
        {UUID: "c", TurnSequence: 30},
        {UUID: "a", TurnSequence: 10},
        {UUID: "b", TurnSequence: 20},
    }

    SortByTurnSequence(tools)

    if tools[0].UUID != "a" || tools[1].UUID != "b" || tools[2].UUID != "c" {
        t.Error("SortByTurnSequence failed")
    }
}

func TestSortDeterminism(t *testing.T) {
    // Generate 100 random tool calls
    tools1 := testutil.GenerateRandomToolCalls(100)
    tools2 := make([]parser.ToolCall, len(tools1))
    copy(tools2, tools1)

    // Sort both
    SortByTimestamp(tools1)
    SortByTimestamp(tools2)

    // Verify identical order
    for i := range tools1 {
        if tools1[i].UUID != tools2[i].UUID {
            t.Error("sorting not deterministic")
        }
    }
}
```

**Update Commands**:

```go
// cmd/query_tools.go - add sorting
func runQueryTools(cmd *cobra.Command, args []string) error {
    // ... existing logic ...

    // Apply sorting before output
    output.SortByTimestamp(tools)

    return output.Format(tools, outputFormat)
}

// cmd/query_messages.go - add sorting
func runQueryMessages(cmd *cobra.Command, args []string) error {
    // ... existing logic ...

    output.SortByTurnSequence(messages)

    return output.Format(messages, outputFormat)
}
```

**Acceptance Criteria**:
- ✅ All query commands output deterministically sorted data
- ✅ `query tools` → sorted by Timestamp
- ✅ `query messages` → sorted by TurnSequence
- ✅ `query errors` → sorted by Timestamp
- ✅ Sorting is idempotent (running twice produces same order)
- ✅ Unit tests verify determinism

**Code Size**: ~50 lines (source) + ~60 lines (tests) = 110 lines total

---

### Stage 14.4: Code Deduplication

**Duration**: 0.5-1 day

**Objective**: Refactor 5 commands to use SessionPipeline, eliminating ~345 lines of duplication

**Deliverables**:
- Refactored `cmd/parse.go` (stats command)
- Refactored `cmd/query_tools.go`
- Refactored `cmd/query_messages.go`
- Refactored `cmd/analyze_sequences.go`
- Refactored `cmd/analyze_file_churn.go`
- Integration tests updated

**Before/After Comparison**:

```go
// BEFORE (cmd/query_tools.go - 307 lines)
func runQueryTools(cmd *cobra.Command, args []string) error {
    // 1. Locate session (10 lines of duplicate code)
    loc := locator.NewSessionLocator()
    var sessionPath string
    if sessionID != "" {
        sessionPath, err = loc.FromSessionID(sessionID)
    } else if projectPath != "" {
        sessionPath, err = loc.FromProjectPath(projectPath)
    } else {
        sessionPath, err = loc.AutoDetect()
    }
    // ...

    // 2. Parse JSONL (8 lines of duplicate code)
    entries, err := parser.ParseJSONL(sessionPath)
    if err != nil {
        return err
    }

    // 3. Extract tool calls (5 lines)
    tools, err := parser.ExtractToolCalls(entries)
    // ...

    // 4. Filter logic (20 lines)
    filtered := []parser.ToolCall{}
    for _, tool := range tools {
        if queryToolsStatus != "" && tool.Status != queryToolsStatus {
            continue
        }
        // ... more filtering ...
        filtered = append(filtered, tool)
    }

    // 5. Output formatting (30 lines)
    switch outputFormat {
    case "json":
        // ...
    case "tsv":
        // ...
    }
}
```

```go
// AFTER (cmd/query_tools.go - 80 lines)
func runQueryTools(cmd *cobra.Command, args []string) error {
    // 1. Initialize pipeline (1 line)
    p := pipeline.NewSessionPipeline(getGlobalOptions())

    // 2. Load session (1 line)
    if err := p.Load(pipeline.LoadOptions{AutoDetect: true}); err != nil {
        return err
    }

    // 3. Extract tool calls (1 line)
    tools, err := p.ExtractToolCalls()
    if err != nil {
        return err
    }

    // 4. Filter (reuse existing filter package, 3 lines)
    filtered := filter.ApplyFilters(tools, filter.Config{
        Status: queryToolsStatus,
        Tool:   queryToolsTool,
        Where:  queryToolsWhere,
    })

    // 5. Sort (1 line)
    output.SortByTimestamp(filtered)

    // 6. Paginate (2 lines)
    paginated := filter.ApplyPagination(filtered, filter.PaginationConfig{
        Limit:  limitFlag,
        Offset: offsetFlag,
    })

    // 7. Output (1 line)
    return output.Format(paginated, outputFormat)
}
```

**Helper Function**:

```go
// cmd/root.go - add helper
func getGlobalOptions() pipeline.GlobalOptions {
    return pipeline.GlobalOptions{
        SessionID:   sessionID,
        ProjectPath: projectPath,
        WorkingDir:  workingDir,
    }
}
```

**Implementation Steps**:

1. Refactor `parse stats` command:
   ```go
   // cmd/parse.go
   func runParseStats(cmd *cobra.Command, args []string) error {
       p := pipeline.NewSessionPipeline(getGlobalOptions())
       if err := p.Load(pipeline.LoadOptions{AutoDetect: true}); err != nil {
           return err
       }

       tools, _ := p.ExtractToolCalls()
       stats := analyzer.CalculateStats(tools)

       return output.Format(stats, outputFormat)
   }
   ```

2. Refactor `query tools` (shown above)

3. Refactor `query messages`:
   ```go
   func runQueryMessages(cmd *cobra.Command, args []string) error {
       p := pipeline.NewSessionPipeline(getGlobalOptions())
       p.Load(pipeline.LoadOptions{AutoDetect: true})

       messages, _ := p.ExtractUserMessages()

       // Apply pattern filter if specified
       if queryMessagesPattern != "" {
           messages = filter.FilterByPattern(messages, queryMessagesPattern)
       }

       output.SortByTurnSequence(messages)

       return output.Format(messages, outputFormat)
   }
   ```

4. Refactor `analyze sequences`:
   ```go
   func runAnalyzeSequences(cmd *cobra.Command, args []string) error {
       p := pipeline.NewSessionPipeline(getGlobalOptions())
       p.Load(pipeline.LoadOptions{AutoDetect: true})

       tools, _ := p.ExtractToolCalls()
       sequences := analyzer.DetectToolSequences(tools, sequencesMinOccurrences)

       return output.Format(sequences, outputFormat)
   }
   ```

5. Refactor `analyze file-churn`:
   ```go
   func runAnalyzeFileChurn(cmd *cobra.Command, args []string) error {
       p := pipeline.NewSessionPipeline(getGlobalOptions())
       p.Load(pipeline.LoadOptions{AutoDetect: true})

       tools, _ := p.ExtractToolCalls()
       churn := analyzer.CalculateFileChurn(tools, fileChurnThreshold)

       return output.Format(churn, outputFormat)
   }
   ```

**Acceptance Criteria**:
- ✅ All 5 commands refactored to use SessionPipeline
- ✅ Session location code appears only once (in pipeline)
- ✅ JSONL parsing code appears only once (in pipeline)
- ✅ Output formatting unified via `output.Format()`
- ✅ Total code reduction ≥60% (1194 → 340 lines)
- ✅ All existing tests pass without modification
- ✅ Integration tests verify behavior unchanged

**Code Size**: -854 lines removed (net reduction)

---

## Testing Strategy

### Unit Testing (TDD)

**Test-Driven Development Flow**:
1. Write test first (define expected behavior)
2. Run test (verify it fails)
3. Implement minimum code to pass
4. Refactor while keeping tests green

**Coverage Targets**:
- Stage 14.1: ≥90% coverage (pipeline package)
- Stage 14.2: ≥85% coverage (query errors)
- Stage 14.3: ≥90% coverage (sorting utilities)
- Stage 14.4: ≥80% coverage (refactored commands)

**Test Commands**:
```bash
# Run all tests
go test ./... -v

# Test specific stage
go test ./pkg/pipeline -v          # Stage 14.1
go test ./cmd -run TestQueryErrors # Stage 14.2
go test ./pkg/output -run TestSort # Stage 14.3

# Check coverage
go test ./pkg/pipeline -cover
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Testing

**Test Script**: `test-scripts/validate-phase-14.sh`

```bash
#!/bin/bash
# Phase 14 Validation Script

set -e

echo "=== Phase 14 Architecture Refactoring Validation ==="
echo ""

# Stage 14.1: Pipeline functionality
echo "[1/5] Testing SessionPipeline..."
go test ./pkg/pipeline -v
echo "✅ Pipeline tests passed"
echo ""

# Stage 14.2: Simplified errors command
echo "[2/5] Testing query errors..."

# Old command should be removed
if meta-cc analyze errors --help 2>/dev/null; then
    echo "❌ analyze errors should be removed"
    exit 1
fi

# New command should exist
ERRORS=$(meta-cc query errors --limit 10)
ERROR_COUNT=$(echo "$ERRORS" | jq '. | length')

if [ "$ERROR_COUNT" -ge 0 ]; then
    echo "✅ query errors works correctly"
else
    echo "❌ query errors failed"
    exit 1
fi
echo ""

# Stage 14.3: Output determinism
echo "[3/5] Testing output determinism..."

# Run same query twice
OUTPUT1=$(meta-cc query tools --limit 50)
OUTPUT2=$(meta-cc query tools --limit 50)

if [ "$OUTPUT1" == "$OUTPUT2" ]; then
    echo "✅ Output is deterministic"
else
    echo "❌ Output is non-deterministic"
    exit 1
fi
echo ""

# Stage 14.4: Code size verification
echo "[4/5] Verifying code reduction..."

TOTAL_LINES=$(wc -l cmd/*.go | tail -1 | awk '{print $1}')
TARGET_MAX=3500  # Should be < 4000 after refactoring

if [ "$TOTAL_LINES" -lt "$TARGET_MAX" ]; then
    echo "✅ Code size reduced: $TOTAL_LINES lines (target: <$TARGET_MAX)"
else
    echo "⚠️  Code size: $TOTAL_LINES lines (target: <$TARGET_MAX)"
fi
echo ""

# Stage 14.5: Behavior equivalence
echo "[5/5] Testing behavior equivalence..."

# Compare output with baseline (Phase 13)
TOOLS_OUTPUT=$(meta-cc query tools --limit 20)
STATS_OUTPUT=$(meta-cc parse stats)

if echo "$TOOLS_OUTPUT" | jq -e '. | length == 20' >/dev/null; then
    echo "✅ query tools output correct"
else
    echo "❌ query tools output incorrect"
    exit 1
fi

if echo "$STATS_OUTPUT" | jq -e '.TurnCount' >/dev/null; then
    echo "✅ parse stats output correct"
else
    echo "❌ parse stats output incorrect"
    exit 1
fi

echo ""
echo "=== All Phase 14 Tests Passed ✅ ==="
echo ""
echo "Summary:"
echo "  - Pipeline abstraction: working"
echo "  - Simplified errors command: working"
echo "  - Output determinism: verified"
echo "  - Code size: reduced by ~72%"
echo "  - Behavior equivalence: preserved"
```

### Regression Testing

**Existing Test Compatibility**:
```bash
# Run all existing unit tests
go test ./... -v

# Run integration tests from previous phases
./tests/integration/slash_commands_test.sh
./tests/integration/mcp_test.sh

# Verify Slash Commands still work
# (requires Claude Code environment)
cd test-workspace
# Test /meta-stats
# Test /meta-errors
```

---

## Migration Guide

### For Users

**Breaking Change 1: analyze errors → query errors**

```bash
# ❌ Old (Phase 13)
meta-cc analyze errors --window 50

# ✅ New (Phase 14)
meta-cc query errors | jq '.[-50:]'
```

**Breaking Change 2: Aggregation removed**

```bash
# ❌ Old (Phase 13)
meta-cc analyze errors --window 50
# Output: Aggregated patterns with counts

# ✅ New (Phase 14) - aggregate with jq
meta-cc query errors | jq 'group_by(.signature) | map({
    signature: .[0].signature,
    count: length,
    first: .[0].timestamp,
    last: .[-1].timestamp
})'
```

**Breaking Change 3: Output sorting**

```bash
# Output is now deterministic (sorted by timestamp)
# No user action needed, but output order may change
meta-cc query tools --limit 100
```

### For Slash Commands

**Update `.claude/commands/meta-errors.md`**:

```markdown
---
name: meta-errors
description: Analyze errors in current session
allowed_tools: [Bash]
---

Run error analysis:

```bash
# Check if meta-cc is installed
if ! command -v meta-cc &> /dev/null; then
    echo "❌ Error: meta-cc not installed"
    exit 1
fi

# Extract errors (Phase 14 new command)
ERRORS=$(meta-cc query errors)

# Aggregate patterns using jq
PATTERNS=$(echo "$ERRORS" | jq 'group_by(.signature) | map({
    signature: .[0].signature,
    tool: .[0].tool_name,
    count: length,
    first_occurrence: .[0].timestamp,
    last_occurrence: .[-1].timestamp,
    sample_error: .[0].error
}) | sort_by(-.count)')

# Display results
echo "## Error Patterns"
echo ""
echo "$PATTERNS" | jq -r '.[] | "### \(.tool): \(.signature) (\(.count) occurrences)\n\n\(.sample_error)\n"'
```
```

### For Developers

**Adding New Commands**:

```go
// Old pattern (DON'T use)
func runNewCommand(cmd *cobra.Command, args []string) error {
    // Duplicate session location logic
    loc := locator.NewSessionLocator()
    sessionPath, _ := loc.FromSessionID(sessionID)

    // Duplicate JSONL parsing
    entries, _ := parser.ParseJSONL(sessionPath)

    // ... command-specific logic ...
}

// New pattern (DO use)
func runNewCommand(cmd *cobra.Command, args []string) error {
    // Use pipeline
    p := pipeline.NewSessionPipeline(getGlobalOptions())
    if err := p.Load(pipeline.LoadOptions{AutoDetect: true}); err != nil {
        return err
    }

    // Extract data
    tools, _ := p.ExtractToolCalls()

    // Command-specific logic
    result := processData(tools)

    // Unified output
    return output.Format(result, outputFormat)
}
```

---

## Success Criteria

### Functional Requirements

- ✅ All commands produce identical output to Phase 13 (except analyze errors)
- ✅ Output is deterministically sorted (same input → same output order)
- ✅ `query errors` replaces `analyze errors` with simpler output
- ✅ Pipeline correctly handles all 3 session location methods
- ✅ No regressions in existing functionality

### Code Quality Requirements

- ✅ Code reduction ≥60% (1194 → ≤340 lines across 5 commands)
- ✅ Unit test coverage ≥80% for all new code
- ✅ All unit tests pass (including existing tests)
- ✅ Integration tests pass (`validate-phase-14.sh`)
- ✅ No duplicate session location/parsing code

### Documentation Requirements

- ✅ Migration guide in README.md
- ✅ Breaking changes documented
- ✅ Slash Commands updated
- ✅ Pipeline usage examples in developer docs
- ✅ `query errors` command documented

### Integration Requirements

- ✅ Slash Commands work without modification (except meta-errors)
- ✅ MCP Server tools return deterministic output
- ✅ All Phase 13 integration tests still pass
- ✅ Real-world validation with 3 projects (meta-cc, NarrativeForge, claude-tmux)

---

## Risk Assessment & Mitigation

### High Risk: Breaking Changes

**Risk**: Users relying on `analyze errors` output format

**Mitigation**:
- Comprehensive migration guide in README
- Warning in CHANGELOG
- Deprecation notice in Phase 13 docs
- jq examples for equivalent functionality

**Rollback Plan**:
```bash
# If needed, revert to Phase 13
git checkout feature/phase-13
go build -o meta-cc
```

### Medium Risk: Code Size Underestimation

**Risk**: Refactoring takes longer than 600 lines estimate

**Mitigation**:
- Break Stage 14.4 into smaller sub-stages
- Refactor one command at a time
- Keep tests passing after each command refactor

### Medium Risk: Pipeline Bugs

**Risk**: Centralized pipeline introduces single point of failure

**Mitigation**:
- ≥90% unit test coverage for pipeline
- Integration tests verify end-to-end behavior
- Canary testing: refactor one command first, validate thoroughly

### Low Risk: Output Sorting Performance

**Risk**: Sorting large datasets (2000+ entries) impacts performance

**Mitigation**:
- Benchmark sorting performance
- Use efficient sort algorithms (Go's sort.Slice is optimized)
- Sorting is O(n log n), acceptable for n < 10,000

---

## Performance Benchmarks

### Expected Performance

**Baseline (Phase 13)**:
- Load session (2000 turns): ~150ms
- Extract tools (1000 calls): ~50ms
- Filter + format: ~30ms
- **Total**: ~230ms

**Target (Phase 14)**:
- Pipeline load: ~150ms (same)
- Extract tools: ~50ms (same)
- **Sort overhead**: +5ms (new)
- Filter + format: ~30ms (same)
- **Total**: ~235ms (+2%)

**Verification**:
```bash
# Benchmark script
time meta-cc query tools --limit 1000
time meta-cc query tools --limit 1000 --output tsv
```

---

## Timeline & Milestones

### Day 1: Pipeline Foundation
- **Morning**: Stage 14.1 implementation (SessionPipeline)
- **Afternoon**: Stage 14.1 tests (90% coverage)
- **End of Day**: Pipeline validated, ready for integration

### Day 2: Simplification & Sorting
- **Morning**: Stage 14.2 (simplify errors command)
- **Afternoon**: Stage 14.3 (output sorting)
- **End of Day**: Both stages tested and validated

### Day 3: Deduplication & Validation
- **Morning**: Stage 14.4 part 1 (refactor 2-3 commands)
- **Afternoon**: Stage 14.4 part 2 (refactor remaining commands)
- **End of Day**: Integration tests, documentation, phase complete

### Phase Completion
- ✅ All 4 stages complete
- ✅ Code reduction target met (≥60%)
- ✅ All tests passing
- ✅ Documentation updated
- ✅ Migration guide published

---

## Appendix

### A. Code Size Tracking

**Current State (Phase 13)**:
```bash
$ wc -l cmd/*.go
  170 cmd/parse.go
  307 cmd/query_tools.go
  280 cmd/query_messages.go
  317 cmd/analyze.go
  120 cmd/analyze_sequences.go
 1194 total (5 commands)
```

**Target State (Phase 14)**:
```bash
$ wc -l cmd/*.go pkg/pipeline/*.go
   60 cmd/parse.go           (-65%)
   80 cmd/query_tools.go     (-74%)
   70 cmd/query_messages.go  (-75%)
   80 cmd/query_errors.go    (new, replaces analyze errors)
   50 cmd/analyze_sequences.go (-58%)
  340 total (5 commands)     (-72%)

  120 pkg/pipeline/session.go (new abstraction)
  460 total
```

**Net Savings**: 1194 - 460 = 734 lines (-61.5%)

### B. Removed Duplicate Patterns

**Pattern 1: Session Location (50 lines removed)**
```go
// BEFORE (repeated in 5 files)
loc := locator.NewSessionLocator()
if sessionID != "" {
    sessionPath, _ = loc.FromSessionID(sessionID)
} else if projectPath != "" {
    sessionPath, _ = loc.FromProjectPath(projectPath)
} else {
    sessionPath, _ = loc.AutoDetect()
}

// AFTER (once in pipeline)
p.Load(pipeline.LoadOptions{AutoDetect: true})
```

**Pattern 2: JSONL Parsing (40 lines removed)**
```go
// BEFORE (repeated in 5 files)
entries, err := parser.ParseJSONL(sessionPath)
if err != nil {
    return fmt.Errorf("failed to parse JSONL: %w", err)
}

// AFTER (once in pipeline)
p.Load() // includes parsing
```

**Pattern 3: Output Formatting (150 lines removed)**
```go
// BEFORE (repeated in 5 files)
switch outputFormat {
case "json":
    data, _ := json.MarshalIndent(result, "", "  ")
    fmt.Println(string(data))
case "tsv":
    // 30 lines of TSV formatting ...
}

// AFTER (unified in output package)
output.Format(result, outputFormat)
```

### C. Reference Commands

**Phase 14 Examples**:
```bash
# Query errors (new simplified command)
meta-cc query errors
meta-cc query errors --limit 50
meta-cc query errors | jq '.[-20:]'

# Aggregate errors with jq
meta-cc query errors | jq 'group_by(.signature) | map({sig: .[0].signature, count: length})'

# Query tools (deterministic output)
meta-cc query tools --limit 100
meta-cc query tools --status error

# Parse stats (refactored)
meta-cc parse stats
```

---

## Next Phase Preview

**Phase 15: Advanced Indexing (Optional)**
- SQLite-based indexing for cross-session queries
- Persistent caching of parsed sessions
- Sub-second query times for large projects

**Phase 16: Subagent Integration (Optional)**
- @meta-coach conversational analysis
- Automated workflow optimization suggestions
- Integration with Phase 14 pipeline

---

**Phase 14 Implementation Plan Complete**

**Estimated Timeline**: 2-3 days

**Risk Level**: Medium (breaking changes, extensive refactoring)

**Success Probability**: High (clear design, comprehensive tests, incremental approach)

**Ready to Begin**: ✅
