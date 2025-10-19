# Test Fixture Examples

**Version**: 2.0
**Source**: Bootstrap-002 Test Strategy Development
**Last Updated**: 2025-10-18

This document provides examples of test fixtures, test helpers, and test data management for Go testing.

---

## Overview

**Test Fixtures**: Reusable test data and setup code that can be shared across multiple tests.

**Benefits**:
- Reduce duplication
- Improve maintainability
- Standardize test data
- Speed up test writing

---

## Example 1: Simple Test Helper Functions

### Pattern 5: Test Helper Pattern

```go
package parser

import (
    "os"
    "path/filepath"
    "testing"
)

// Test helper: Create test input
func createTestInput(t *testing.T, content string) *Input {
    t.Helper()  // Mark as helper for better error reporting

    return &Input{
        Content:   content,
        Timestamp: "2025-10-18T10:00:00Z",
        Type:      "tool_use",
    }
}

// Test helper: Create test file
func createTestFile(t *testing.T, name, content string) string {
    t.Helper()

    tmpDir := t.TempDir()
    filePath := filepath.Join(tmpDir, name)

    if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
        t.Fatalf("failed to create test file: %v", err)
    }

    return filePath
}

// Test helper: Load fixture
func loadFixture(t *testing.T, name string) []byte {
    t.Helper()

    data, err := os.ReadFile(filepath.Join("testdata", name))
    if err != nil {
        t.Fatalf("failed to load fixture %s: %v", name, err)
    }

    return data
}

// Usage in tests
func TestParseInput(t *testing.T) {
    input := createTestInput(t, "test content")
    result, err := ParseInput(input)

    if err != nil {
        t.Fatalf("ParseInput() error = %v", err)
    }

    if result.Type != "tool_use" {
        t.Errorf("Type = %v, want tool_use", result.Type)
    }
}
```

**Benefits**:
- No duplication of test setup
- `t.Helper()` makes errors point to test code, not helper
- Consistent test data across tests

---

## Example 2: Fixture Files in testdata/

### Directory Structure

```
internal/parser/
├── parser.go
├── parser_test.go
└── testdata/
    ├── valid_session.jsonl
    ├── invalid_session.jsonl
    ├── empty_session.jsonl
    ├── large_session.jsonl
    └── README.md
```

### Fixture Files

**testdata/valid_session.jsonl**:
```jsonl
{"type":"tool_use","tool":"Read","file":"/test/file.go","timestamp":"2025-10-18T10:00:00Z"}
{"type":"tool_use","tool":"Edit","file":"/test/file.go","timestamp":"2025-10-18T10:01:00Z","status":"success"}
{"type":"tool_use","tool":"Bash","command":"go test","timestamp":"2025-10-18T10:02:00Z","status":"success"}
```

**testdata/invalid_session.jsonl**:
```jsonl
{"type":"tool_use","tool":"Read","file":"/test/file.go","timestamp":"2025-10-18T10:00:00Z"}
invalid json line here
{"type":"tool_use","tool":"Edit","file":"/test/file.go","timestamp":"2025-10-18T10:01:00Z"}
```

### Using Fixtures in Tests

```go
func TestParseSessionFile(t *testing.T) {
    tests := []struct {
        name        string
        fixture     string
        wantErr     bool
        expectedLen int
    }{
        {
            name:        "valid session",
            fixture:     "valid_session.jsonl",
            wantErr:     false,
            expectedLen: 3,
        },
        {
            name:        "invalid session",
            fixture:     "invalid_session.jsonl",
            wantErr:     true,
            expectedLen: 0,
        },
        {
            name:        "empty session",
            fixture:     "empty_session.jsonl",
            wantErr:     false,
            expectedLen: 0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            data := loadFixture(t, tt.fixture)

            events, err := ParseSessionData(data)

            if (err != nil) != tt.wantErr {
                t.Errorf("ParseSessionData() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !tt.wantErr && len(events) != tt.expectedLen {
                t.Errorf("got %d events, want %d", len(events), tt.expectedLen)
            }
        })
    }
}
```

---

## Example 3: Builder Pattern for Test Data

### Test Data Builder

```go
package query

import "testing"

// Builder for complex test data
type TestQueryBuilder struct {
    query *Query
}

func NewTestQuery() *TestQueryBuilder {
    return &TestQueryBuilder{
        query: &Query{
            Type:    "tools",
            Filters: []Filter{},
            Options: Options{
                Limit:  0,
                Format: "jsonl",
            },
        },
    }
}

func (b *TestQueryBuilder) WithType(queryType string) *TestQueryBuilder {
    b.query.Type = queryType
    return b
}

func (b *TestQueryBuilder) WithFilter(field, op, value string) *TestQueryBuilder {
    b.query.Filters = append(b.query.Filters, Filter{
        Field:    field,
        Operator: op,
        Value:    value,
    })
    return b
}

func (b *TestQueryBuilder) WithLimit(limit int) *TestQueryBuilder {
    b.query.Options.Limit = limit
    return b
}

func (b *TestQueryBuilder) WithFormat(format string) *TestQueryBuilder {
    b.query.Options.Format = format
    return b
}

func (b *TestQueryBuilder) Build() *Query {
    return b.query
}

// Usage in tests
func TestExecuteQuery(t *testing.T) {
    // Simple query
    query1 := NewTestQuery().
        WithType("tools").
        Build()

    // Complex query
    query2 := NewTestQuery().
        WithType("messages").
        WithFilter("status", "=", "error").
        WithFilter("timestamp", ">=", "2025-10-01").
        WithLimit(10).
        WithFormat("tsv").
        Build()

    result, err := ExecuteQuery(query2)
    // ... assertions
}
```

**Benefits**:
- Fluent API for test data construction
- Easy to create variations
- Self-documenting test setup

---

## Example 4: Golden File Testing

### Pattern: Golden File Output Validation

```go
package formatter

import (
    "flag"
    "os"
    "path/filepath"
    "testing"
)

var update = flag.Bool("update", false, "update golden files")

func TestFormatOutput(t *testing.T) {
    tests := []struct {
        name  string
        input []Event
    }{
        {
            name: "simple_output",
            input: []Event{
                {Type: "Read", File: "file.go"},
                {Type: "Edit", File: "file.go"},
            },
        },
        {
            name: "complex_output",
            input: []Event{
                {Type: "Read", File: "file1.go"},
                {Type: "Edit", File: "file1.go"},
                {Type: "Bash", Command: "go test"},
                {Type: "Read", File: "file2.go"},
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Format output
            output := FormatOutput(tt.input)

            // Golden file path
            goldenPath := filepath.Join("testdata", tt.name+".golden")

            // Update golden file if flag set
            if *update {
                if err := os.WriteFile(goldenPath, []byte(output), 0644); err != nil {
                    t.Fatalf("failed to update golden file: %v", err)
                }
                t.Logf("updated golden file: %s", goldenPath)
                return
            }

            // Load expected output
            expected, err := os.ReadFile(goldenPath)
            if err != nil {
                t.Fatalf("failed to read golden file: %v", err)
            }

            // Compare
            if output != string(expected) {
                t.Errorf("output mismatch:\n=== GOT ===\n%s\n=== WANT ===\n%s", output, expected)
            }
        })
    }
}
```

**Usage**:
```bash
# Run tests normally (compares against golden files)
go test ./...

# Update golden files
go test ./... -update

# Review changes
git diff testdata/
```

**Benefits**:
- Easy to maintain expected outputs
- Visual diff of changes
- Great for complex string outputs

---

## Example 5: Table-Driven Fixtures

### Shared Test Data for Multiple Tests

```go
package analyzer

import "testing"

// Shared test fixtures
var testEvents = []struct {
    name   string
    events []Event
}{
    {
        name: "tdd_pattern",
        events: []Event{
            {Type: "Write", File: "file_test.go"},
            {Type: "Bash", Command: "go test"},
            {Type: "Edit", File: "file.go"},
            {Type: "Bash", Command: "go test"},
        },
    },
    {
        name: "refactor_pattern",
        events: []Event{
            {Type: "Read", File: "old.go"},
            {Type: "Write", File: "new.go"},
            {Type: "Edit", File: "new.go"},
            {Type: "Bash", Command: "go test"},
        },
    },
}

// Test 1 uses fixtures
func TestDetectPatterns(t *testing.T) {
    for _, fixture := range testEvents {
        t.Run(fixture.name, func(t *testing.T) {
            patterns := DetectPatterns(fixture.events)

            if len(patterns) == 0 {
                t.Error("no patterns detected")
            }
        })
    }
}

// Test 2 uses same fixtures
func TestAnalyzeWorkflow(t *testing.T) {
    for _, fixture := range testEvents {
        t.Run(fixture.name, func(t *testing.T) {
            workflow := AnalyzeWorkflow(fixture.events)

            if workflow.Type == "" {
                t.Error("workflow type not detected")
            }
        })
    }
}
```

**Benefits**:
- Fixtures shared across multiple test functions
- Consistent test data
- Easy to add new fixtures for all tests

---

## Example 6: Mock Data Generators

### Random Test Data Generation

```go
package parser

import (
    "fmt"
    "math/rand"
    "testing"
    "time"
)

// Generate random test events
func generateTestEvents(t *testing.T, count int) []Event {
    t.Helper()

    rand.Seed(time.Now().UnixNano())

    tools := []string{"Read", "Edit", "Write", "Bash", "Grep"}
    statuses := []string{"success", "error"}

    events := make([]Event, count)
    for i := 0; i < count; i++ {
        events[i] = Event{
            Type:      "tool_use",
            Tool:      tools[rand.Intn(len(tools))],
            File:      fmt.Sprintf("/test/file%d.go", rand.Intn(10)),
            Status:    statuses[rand.Intn(len(statuses))],
            Timestamp: time.Now().Add(time.Duration(i) * time.Second).Format(time.RFC3339),
        }
    }

    return events
}

// Usage in tests
func TestParseEvents_LargeDataset(t *testing.T) {
    events := generateTestEvents(t, 1000)

    parsed, err := ParseEvents(events)

    if err != nil {
        t.Fatalf("ParseEvents() error = %v", err)
    }

    if len(parsed) != 1000 {
        t.Errorf("got %d events, want 1000", len(parsed))
    }
}

func TestAnalyzeEvents_Performance(t *testing.T) {
    events := generateTestEvents(t, 10000)

    start := time.Now()
    AnalyzeEvents(events)
    duration := time.Since(start)

    if duration > 1*time.Second {
        t.Errorf("analysis took %v, want <1s", duration)
    }
}
```

**When to use**:
- Performance testing
- Stress testing
- Property-based testing
- Large dataset testing

---

## Example 7: Cleanup and Teardown

### Proper Resource Cleanup

```go
func TestWithTempDirectory(t *testing.T) {
    // Using t.TempDir() (preferred)
    tmpDir := t.TempDir()  // Automatically cleaned up

    // Create test files
    testFile := filepath.Join(tmpDir, "test.txt")
    os.WriteFile(testFile, []byte("test"), 0644)

    // Test code...
    // No manual cleanup needed
}

func TestWithCleanup(t *testing.T) {
    // Using t.Cleanup() for custom cleanup
    oldValue := globalVar
    globalVar = "test"

    t.Cleanup(func() {
        globalVar = oldValue
    })

    // Test code...
    // globalVar will be restored automatically
}

func TestWithDefer(t *testing.T) {
    // Using defer (also works)
    oldValue := globalVar
    defer func() { globalVar = oldValue }()

    globalVar = "test"

    // Test code...
}

func TestMultipleCleanups(t *testing.T) {
    // Multiple cleanups execute in LIFO order
    t.Cleanup(func() {
        fmt.Println("cleanup 1")
    })

    t.Cleanup(func() {
        fmt.Println("cleanup 2")
    })

    // Test code...

    // Output:
    // cleanup 2
    // cleanup 1
}
```

---

## Example 8: Integration Test Fixtures

### Complete Test Environment Setup

```go
package integration

import (
    "os"
    "path/filepath"
    "testing"
)

// Setup complete test environment
func setupTestEnvironment(t *testing.T) *TestEnv {
    t.Helper()

    tmpDir := t.TempDir()

    // Create directory structure
    dirs := []string{
        ".claude/logs",
        ".claude/tools",
        "src",
        "tests",
    }

    for _, dir := range dirs {
        path := filepath.Join(tmpDir, dir)
        if err := os.MkdirAll(path, 0755); err != nil {
            t.Fatalf("failed to create dir %s: %v", dir, err)
        }
    }

    // Create test files
    sessionFile := filepath.Join(tmpDir, ".claude/logs/session.jsonl")
    testSessionData := `{"type":"tool_use","tool":"Read","file":"test.go"}
{"type":"tool_use","tool":"Edit","file":"test.go"}
{"type":"tool_use","tool":"Bash","command":"go test"}`

    if err := os.WriteFile(sessionFile, []byte(testSessionData), 0644); err != nil {
        t.Fatalf("failed to create session file: %v", err)
    }

    // Create config
    configFile := filepath.Join(tmpDir, ".claude/config.json")
    configData := `{"project":"test","version":"1.0.0"}`

    if err := os.WriteFile(configFile, []byte(configData), 0644); err != nil {
        t.Fatalf("failed to create config: %v", err)
    }

    return &TestEnv{
        RootDir:     tmpDir,
        SessionFile: sessionFile,
        ConfigFile:  configFile,
    }
}

type TestEnv struct {
    RootDir     string
    SessionFile string
    ConfigFile  string
}

// Usage in integration tests
func TestIntegration_FullWorkflow(t *testing.T) {
    env := setupTestEnvironment(t)

    // Run full workflow
    result, err := RunWorkflow(env.RootDir)

    if err != nil {
        t.Fatalf("RunWorkflow() error = %v", err)
    }

    if result.EventsProcessed != 3 {
        t.Errorf("EventsProcessed = %d, want 3", result.EventsProcessed)
    }
}
```

---

## Best Practices for Fixtures

### 1. Use testdata/ Directory

```
package/
├── code.go
├── code_test.go
└── testdata/
    ├── fixture1.json
    ├── fixture2.json
    └── README.md  # Document fixtures
```

### 2. Name Fixtures Descriptively

```
❌ data1.json, data2.json
✅ valid_session.jsonl, invalid_session.jsonl, empty_session.jsonl
```

### 3. Keep Fixtures Small

```go
// Bad: 1000-line fixture
data := loadFixture(t, "large_fixture.json")

// Good: Minimal fixture
data := loadFixture(t, "minimal_valid.json")
```

### 4. Document Fixtures

**testdata/README.md**:
```markdown
# Test Fixtures

## valid_session.jsonl
Complete valid session with 3 tool uses (Read, Edit, Bash).

## invalid_session.jsonl
Session with malformed JSON on line 2 (for error testing).

## empty_session.jsonl
Empty file (for edge case testing).
```

### 5. Use Helpers for Variations

```go
func createTestEvent(t *testing.T, options ...func(*Event)) *Event {
    t.Helper()

    event := &Event{
        Type: "tool_use",
        Tool: "Read",
        Status: "success",
    }

    for _, opt := range options {
        opt(event)
    }

    return event
}

// Option functions
func WithTool(tool string) func(*Event) {
    return func(e *Event) { e.Tool = tool }
}

func WithStatus(status string) func(*Event) {
    return func(e *Event) { e.Status = status }
}

// Usage
event1 := createTestEvent(t)  // Default
event2 := createTestEvent(t, WithTool("Edit"))
event3 := createTestEvent(t, WithTool("Bash"), WithStatus("error"))
```

---

## Fixture Efficiency Comparison

| Approach | Time to Create Test | Maintainability | Flexibility |
|----------|---------------------|-----------------|-------------|
| **Inline data** | Fast (2-3 min) | Low (duplicated) | High |
| **Helper functions** | Medium (5 min) | High (reusable) | Very High |
| **Fixture files** | Slow (10 min) | Very High (centralized) | Medium |
| **Builder pattern** | Medium (8 min) | High (composable) | Very High |
| **Golden files** | Fast (2 min) | Very High (visual diff) | Low |

**Recommendation**: Use fixture files for complex data, helpers for variations, inline for simple cases.

---

**Source**: Bootstrap-002 Test Strategy Development
**Framework**: BAIME (Bootstrapped AI Methodology Engineering)
**Status**: Production-ready, validated through 4 iterations
