# Phase 11: Unix Composability (Unix 工具可组合性)

## 概述

**目标**: 优化输出格式和 CLI 设计，完善 Unix 管道支持，使 meta-cc 能够与 jq、grep、awk 等标准 Unix 工具无缝组合

**代码量**: ~200 行 (Go 源代码 + Markdown 文档)

**依赖**: Phase 0-10 (完整的 meta-cc 工具链 + 查询和统计命令)

**交付物**:
- JSONL 流式输出（`--stream` 模式）
- 标准化退出码（0=success, 1=error, 2=no results）
- stderr/stdout 分离（日志 vs 数据）
- `docs/cookbook.md`：常见分析模式
- `docs/cli-composability.md`：与 jq/grep/awk 组合示例

---

## Phase 目标

优化 meta-cc 的 CLI 设计，使其符合 Unix 哲学和最佳实践：

### 核心需求

1. **流式输出**：支持 JSONL 格式，每行一个 JSON 对象，便于流式处理
2. **标准化退出码**：遵循 Unix 惯例，便于脚本编写和错误处理
3. **I/O 分离**：日志输出到 stderr，数据输出到 stdout，支持管道过滤
4. **文档和示例**：提供实用的组合使用模式和最佳实践

### 设计原则

Phase 11 遵循 Unix 哲学：

- ✅ **Do one thing well**: 每个命令专注于单一功能
- ✅ **Text streams**: 使用文本流作为通用接口
- ✅ **Composability**: 工具可以通过管道组合使用
- ✅ **Consistent interface**: 统一的命令行接口和行为

**Unix 管道示例**:
```bash
# meta-cc: 提取错误的 Bash 工具调用
meta-cc query tools --stream | jq -c 'select(.ToolName == "Bash" and .Status == "error")'

# 组合 jq 和 awk 统计
meta-cc query tools --stream | jq '.ToolName' | sort | uniq -c | awk '{print $2 ": " $1}'

# 错误模式分析
meta-cc query tools --where "status='error'" --stream | \
  jq -r '.Error' | \
  grep -oP '(permission|not found|timeout)' | \
  sort | uniq -c
```

---

## 成功标准

**功能验收**:
- ✅ 所有 Stage 单元测试通过（TDD）
- ✅ `--stream` 标志在所有 query 和 stats 命令中可用
- ✅ 输出的 JSONL 格式有效（每行可被 jq 解析）
- ✅ 退出码符合 Unix 惯例（0/1/2）
- ✅ 所有日志输出到 stderr，数据输出到 stdout

**集成验收**:
- ✅ 管道工作流验证（与 jq, grep, awk, sed 等组合）
- ✅ 脚本场景验证（基于退出码的条件逻辑）
- ✅ 无回归（所有现有测试通过）

**代码质量**:
- ✅ 实际代码量: ~200 行（目标 180-220 行）
- ✅ 每个 Stage ≤ 60 行（Go 源代码）
- ✅ 测试覆盖率: ≥ 80%
- ✅ 无新增外部依赖

**文档质量**:
- ✅ Cookbook: 10+ 实用分析模式
- ✅ Composability Guide: 5+ 工具集成示例
- ✅ 所有示例可执行并验证通过

---

## Stage 11.1: JSONL 流式输出

### 目标

实现 `--stream` 标志，支持 JSONL（JSON Lines）流式输出格式，便于 Unix 管道处理。

### 背景

**当前行为**:
```bash
# 默认输出：完整 JSON 数组
meta-cc query tools --output json
[
  {"uuid": "1", "tool": "Bash", ...},
  {"uuid": "2", "tool": "Edit", ...}
]
```

**问题**:
- JSON 数组需要完整解析后才能处理
- 不适合大数据集的流式处理
- 管道工具（如 jq）需要额外参数（`jq -c '.[]'`）才能逐行处理

**期望行为**:
```bash
# 流式输出：每行一个 JSON 对象（JSONL）
meta-cc query tools --stream
{"uuid":"1","tool":"Bash",...}
{"uuid":"2","tool":"Edit",...}

# 直接用于管道
meta-cc query tools --stream | jq 'select(.Status == "error")'
```

### 实现步骤

#### 1. 定义流式输出接口

**文件**: `internal/output/stream.go` (新建, ~30 行)

```go
package output

import (
	"encoding/json"
	"io"
)

// StreamWriter writes data as JSON Lines (JSONL) format
type StreamWriter struct {
	writer io.Writer
}

// NewStreamWriter creates a new JSONL stream writer
func NewStreamWriter(w io.Writer) *StreamWriter {
	return &StreamWriter{writer: w}
}

// WriteRecord writes a single record as a JSON line
func (sw *StreamWriter) WriteRecord(record interface{}) error {
	// Marshal to compact JSON
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	// Write JSON line
	_, err = sw.writer.Write(data)
	if err != nil {
		return err
	}

	// Write newline
	_, err = sw.writer.Write([]byte("\n"))
	return err
}

// WriteRecords writes multiple records as JSON lines
func (sw *StreamWriter) WriteRecords(records interface{}) error {
	// Use reflection to iterate over slice/array
	// For each record, call WriteRecord
	// Implementation details...
	return nil
}
```

#### 2. 为 query 命令添加 --stream 标志

**文件**: `cmd/query.go` (修改, ~10 行)

```go
var (
	// ... existing flags ...
	queryStream bool  // NEW: Enable streaming output
)

func init() {
	// ... existing flags ...

	// Add --stream flag to query command
	queryCmd.PersistentFlags().BoolVar(&queryStream, "stream", false, "Output as JSON Lines (JSONL) for streaming")
}
```

#### 3. 更新 query tools 命令以支持流式输出

**文件**: `cmd/query_tools.go` (修改, ~15 行)

```go
func runQueryTools(cmd *cobra.Command, args []string) error {
	// ... existing logic (parsing, filtering) ...

	// Determine output format
	if queryStream {
		// Use streaming output
		streamWriter := output.NewStreamWriter(os.Stdout)
		for _, tool := range tools {
			if err := streamWriter.WriteRecord(tool); err != nil {
				return fmt.Errorf("stream write error: %v", err)
			}
		}
		return nil
	}

	// ... existing output logic (JSON array, Markdown, TSV) ...
}
```

#### 4. 扩展到其他命令

**文件**: `cmd/stats_aggregate.go`, `cmd/stats_timeseries.go`, `cmd/stats_files.go` (修改, 各 ~5 行)

```go
// 每个 stats 子命令都支持 --stream 标志
// 使用与 query tools 相同的模式
```

### TDD 步骤

**测试文件**: `internal/output/stream_test.go` (新建, ~60 行)

```go
package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestStreamWriter_WriteRecord(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)

	record := map[string]interface{}{
		"tool":   "Bash",
		"status": "success",
	}

	err := sw.WriteRecord(record)
	if err != nil {
		t.Fatalf("WriteRecord failed: %v", err)
	}

	// Verify output is valid JSON followed by newline
	output := buf.String()
	if !strings.HasSuffix(output, "\n") {
		t.Error("Output should end with newline")
	}

	// Verify JSON is valid
	var decoded map[string]interface{}
	line := strings.TrimSpace(output)
	if err := json.Unmarshal([]byte(line), &decoded); err != nil {
		t.Errorf("Invalid JSON output: %v", err)
	}

	// Verify content
	if decoded["tool"] != "Bash" {
		t.Errorf("Expected tool='Bash', got '%s'", decoded["tool"])
	}
}

func TestStreamWriter_WriteMultipleRecords(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)

	records := []map[string]interface{}{
		{"id": 1, "tool": "Bash"},
		{"id": 2, "tool": "Edit"},
		{"id": 3, "tool": "Read"},
	}

	for _, record := range records {
		if err := sw.WriteRecord(record); err != nil {
			t.Fatalf("WriteRecord failed: %v", err)
		}
	}

	// Verify output is JSONL (3 lines)
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		var decoded map[string]interface{}
		if err := json.Unmarshal([]byte(line), &decoded); err != nil {
			t.Errorf("Line %d invalid JSON: %v", i+1, err)
		}
	}
}

func TestStreamWriter_EmptyData(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)

	// No records written
	output := buf.String()
	if output != "" {
		t.Error("Expected empty output for no records")
	}
}

func TestStreamWriter_ComplexNestedData(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)

	record := map[string]interface{}{
		"tool": "Bash",
		"input": map[string]interface{}{
			"command": "ls -la",
			"timeout": 5000,
		},
		"tags": []string{"filesystem", "list"},
	}

	err := sw.WriteRecord(record)
	if err != nil {
		t.Fatalf("WriteRecord failed: %v", err)
	}

	// Verify nested structure is preserved
	var decoded map[string]interface{}
	line := strings.TrimSpace(buf.String())
	if err := json.Unmarshal([]byte(line), &decoded); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Verify nested input map
	input, ok := decoded["input"].(map[string]interface{})
	if !ok {
		t.Fatal("Input field should be a map")
	}

	if input["command"] != "ls -la" {
		t.Errorf("Expected command='ls -la', got '%v'", input["command"])
	}
}
```

**集成测试**: `tests/integration/streaming_test.sh` (新建, ~30 行)

```bash
#!/bin/bash
# Test JSONL streaming output

set -e

echo "=== JSONL Streaming Test ==="

# Test 1: Verify --stream produces valid JSONL
echo "[1/3] Testing JSONL format..."
STREAM_OUTPUT=$(meta-cc query tools --stream --limit 5 2>/dev/null)
LINE_COUNT=$(echo "$STREAM_OUTPUT" | wc -l)

if [ "$LINE_COUNT" -ne 5 ]; then
    echo "✗ Expected 5 lines, got $LINE_COUNT"
    exit 1
fi

# Verify each line is valid JSON
echo "$STREAM_OUTPUT" | while IFS= read -r line; do
    if ! echo "$line" | jq empty 2>/dev/null; then
        echo "✗ Invalid JSON line: $line"
        exit 1
    fi
done

echo "✓ JSONL format valid"

# Test 2: Verify streaming works with jq
echo "[2/3] Testing jq integration..."
FILTERED=$(meta-cc query tools --stream --limit 100 2>/dev/null | jq -c 'select(.Status == "error")' | wc -l)
echo "✓ Found $FILTERED error records via jq pipeline"

# Test 3: Verify streaming works with grep
echo "[3/3] Testing grep integration..."
BASH_COUNT=$(meta-cc query tools --stream --limit 100 2>/dev/null | grep -c '"ToolName":"Bash"' || true)
echo "✓ Found $BASH_COUNT Bash tool calls via grep pipeline"

echo ""
echo "=== All Streaming Tests Passed ✅ ==="
```

### 交付物

**新建文件**:
- `internal/output/stream.go` (~30 行)
- `internal/output/stream_test.go` (~60 行)
- `tests/integration/streaming_test.sh` (~30 行)

**修改文件**:
- `cmd/query.go` (~10 行)
- `cmd/query_tools.go` (~15 行)
- `cmd/stats_aggregate.go` (~5 行)
- `cmd/stats_timeseries.go` (~5 行)
- `cmd/stats_files.go` (~5 行)

**代码量**: ~50 行（新增 Go 源代码，不含测试）

### 验收标准

- ✅ `meta-cc query tools --stream` 输出有效的 JSONL
- ✅ 每行是独立的 JSON 对象（可被 jq 解析）
- ✅ 管道工作流：`meta-cc query tools --stream | jq 'select(.Status == "error")'`
- ✅ 所有单元测试通过
- ✅ 集成测试通过（jq, grep, awk 组合）
- ✅ 无性能回退（流式输出应该更快）

---

## Stage 11.2: 退出码标准化

### 目标

标准化所有命令的退出码，遵循 Unix 惯例，便于脚本编写和错误处理。

### 背景

**Unix 退出码惯例**:
- `0`: 成功（Success）
- `1`: 一般错误（General error）
- `2`: 无结果/未找到（No results / Not found）
- `3+`: 特定错误码（可选）

**当前问题**:
- meta-cc 可能没有统一的退出码策略
- 难以在脚本中判断命令是否成功

**期望行为**:
```bash
# 成功查询（有结果）
meta-cc query tools --where "tool='Bash'"
echo $?  # 0

# 成功查询（无结果）
meta-cc query tools --where "tool='NonExistent'"
echo $?  # 2

# 错误（语法错误、文件不存在等）
meta-cc query tools --where "invalid syntax"
echo $?  # 1
```

### 实现步骤

#### 1. 定义退出码常量

**文件**: `internal/output/exitcode.go` (新建, ~20 行)

```go
package output

import "os"

// Standard Unix exit codes
const (
	ExitSuccess   = 0  // Command succeeded
	ExitError     = 1  // General error (parsing, I/O, etc.)
	ExitNoResults = 2  // Command succeeded but no results found
)

// Exit with standard code
func Exit(code int) {
	os.Exit(code)
}

// ExitWithMessage prints error message to stderr and exits
func ExitWithMessage(code int, message string) {
	if code != ExitSuccess {
		fmt.Fprintln(os.Stderr, message)
	}
	os.Exit(code)
}

// DetermineExitCode determines the appropriate exit code based on results
func DetermineExitCode(hasResults bool, err error) int {
	if err != nil {
		return ExitError
	}
	if !hasResults {
		return ExitNoResults
	}
	return ExitSuccess
}
```

#### 2. 更新命令以使用标准退出码

**文件**: `cmd/query_tools.go` (修改, ~10 行)

```go
func runQueryTools(cmd *cobra.Command, args []string) error {
	// ... existing logic ...

	// Filter tools
	tools, err := applyFilters(tools)
	if err != nil {
		// Error during filtering: exit 1
		return err  // Cobra will exit with 1
	}

	// Output results
	if err := outputTools(tools); err != nil {
		return err  // Exit 1
	}

	// Determine exit code based on results
	hasResults := len(tools) > 0
	exitCode := output.DetermineExitCode(hasResults, nil)

	// Exit with appropriate code
	// Note: We can't call os.Exit() directly in RunE function
	// Instead, we return a special error type that signals the exit code
	if exitCode == output.ExitNoResults {
		// Special handling: print to stderr, exit 2
		fmt.Fprintln(os.Stderr, "No results found")
		os.Exit(output.ExitNoResults)
	}

	return nil  // Exit 0
}
```

**Note**: Cobra 的 `RunE` 函数返回 `error` 时默认退出码为 1。为了支持退出码 2，我们需要一个自定义错误类型：

**文件**: `internal/output/exitcode.go` (扩展, ~15 行)

```go
// ExitCodeError is a special error type that carries an exit code
type ExitCodeError struct {
	Code    int
	Message string
}

func (e *ExitCodeError) Error() string {
	return e.Message
}

// NewExitCodeError creates a new exit code error
func NewExitCodeError(code int, message string) *ExitCodeError {
	return &ExitCodeError{
		Code:    code,
		Message: message,
	}
}
```

**文件**: `cmd/root.go` (修改, ~10 行)

```go
// In Execute() function, handle ExitCodeError
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Check if it's an ExitCodeError
		if exitErr, ok := err.(*output.ExitCodeError); ok {
			fmt.Fprintln(os.Stderr, exitErr.Message)
			os.Exit(exitErr.Code)
		}

		// Default error handling (exit 1)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(output.ExitError)
	}
}
```

#### 3. 应用到所有命令

**文件**: `cmd/stats_aggregate.go`, `cmd/stats_timeseries.go`, `cmd/stats_files.go` (修改, 各 ~5 行)

```go
// 每个命令在无结果时返回 ExitNoResults (2)
// 每个命令在错误时返回 ExitError (1)
```

### TDD 步骤

**测试文件**: `internal/output/exitcode_test.go` (新建, ~40 行)

```go
package output

import "testing"

func TestDetermineExitCode(t *testing.T) {
	tests := []struct {
		name       string
		hasResults bool
		err        error
		expected   int
	}{
		{
			name:       "success with results",
			hasResults: true,
			err:        nil,
			expected:   ExitSuccess,
		},
		{
			name:       "success without results",
			hasResults: false,
			err:        nil,
			expected:   ExitNoResults,
		},
		{
			name:       "error",
			hasResults: false,
			err:        fmt.Errorf("some error"),
			expected:   ExitError,
		},
		{
			name:       "error with results (still error)",
			hasResults: true,
			err:        fmt.Errorf("some error"),
			expected:   ExitError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineExitCode(tt.hasResults, tt.err)
			if result != tt.expected {
				t.Errorf("Expected exit code %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestExitCodeError(t *testing.T) {
	err := NewExitCodeError(ExitNoResults, "No results found")

	if err.Code != ExitNoResults {
		t.Errorf("Expected code %d, got %d", ExitNoResults, err.Code)
	}

	if err.Error() != "No results found" {
		t.Errorf("Expected message 'No results found', got '%s'", err.Error())
	}
}
```

**集成测试**: `tests/integration/exitcode_test.sh` (新建, ~40 行)

```bash
#!/bin/bash
# Test exit code behavior

set +e  # Don't exit on command failure

echo "=== Exit Code Test ==="

# Test 1: Success with results (exit 0)
echo "[1/3] Testing success with results..."
meta-cc query tools --limit 5 >/dev/null 2>&1
EXIT_CODE=$?

if [ $EXIT_CODE -ne 0 ]; then
    echo "✗ Expected exit 0, got $EXIT_CODE"
    exit 1
fi
echo "✓ Exit 0 for success with results"

# Test 2: Success without results (exit 2)
echo "[2/3] Testing success without results..."
meta-cc query tools --where "tool='NonExistentTool'" >/dev/null 2>&1
EXIT_CODE=$?

if [ $EXIT_CODE -ne 2 ]; then
    echo "✗ Expected exit 2, got $EXIT_CODE"
    exit 1
fi
echo "✓ Exit 2 for no results"

# Test 3: Error (exit 1)
echo "[3/3] Testing error..."
meta-cc query tools --where "invalid AND syntax AND" >/dev/null 2>&1
EXIT_CODE=$?

if [ $EXIT_CODE -ne 1 ]; then
    echo "✗ Expected exit 1, got $EXIT_CODE"
    exit 1
fi
echo "✓ Exit 1 for error"

echo ""
echo "=== All Exit Code Tests Passed ✅ ==="
```

### 交付物

**新建文件**:
- `internal/output/exitcode.go` (~35 行)
- `internal/output/exitcode_test.go` (~40 行)
- `tests/integration/exitcode_test.sh` (~40 行)

**修改文件**:
- `cmd/root.go` (~10 行)
- `cmd/query_tools.go` (~10 行)
- 其他查询和统计命令 (各 ~5 行)

**代码量**: ~30 行（新增 Go 源代码，不含测试）

### 验收标准

- ✅ 有结果时退出码为 0
- ✅ 无结果时退出码为 2
- ✅ 错误时退出码为 1
- ✅ 脚本可以基于退出码进行条件判断
- ✅ 所有单元测试通过
- ✅ 集成测试通过

---

## Stage 11.3: stderr/stdout 分离

### 目标

确保所有日志输出到 stderr，数据输出到 stdout，支持干净的管道过滤。

### 背景

**Unix 惯例**:
- `stdout` (文件描述符 1): 命令的主要输出（数据）
- `stderr` (文件描述符 2): 诊断信息（日志、警告、错误）

**当前问题**:
- 可能混合输出日志和数据到 stdout
- 管道处理时需要额外过滤日志行

**期望行为**:
```bash
# 只输出数据到 stdout
meta-cc query tools --output json > data.json

# 只输出日志到 stderr
meta-cc query tools --output json 2> debug.log

# 分离输出
meta-cc query tools --output json > data.json 2> debug.log

# 管道处理（忽略日志）
meta-cc query tools --stream 2>/dev/null | jq '.ToolName'
```

### 实现步骤

#### 1. 审计所有输出点

**审计清单** (使用 grep 查找):
```bash
# 查找所有 fmt.Println, fmt.Printf, log.Println 等
grep -rn "fmt.Print" cmd/ internal/
grep -rn "log.Print" cmd/ internal/
grep -rn "os.Stdout" cmd/ internal/
grep -rn "os.Stderr" cmd/ internal/
```

#### 2. 定义日志和数据输出策略

**文件**: `internal/output/writer.go` (新建, ~30 行)

```go
package output

import (
	"io"
	"os"
)

// Output destinations
var (
	// DataWriter is the writer for command output data (default: stdout)
	DataWriter io.Writer = os.Stdout

	// LogWriter is the writer for diagnostic messages (default: stderr)
	LogWriter io.Writer = os.Stderr
)

// WriteData writes data to stdout
func WriteData(format string, args ...interface{}) {
	fmt.Fprintf(DataWriter, format, args...)
}

// WriteLog writes log message to stderr
func WriteLog(format string, args ...interface{}) {
	fmt.Fprintf(LogWriter, format, args...)
}

// WriteLogLine writes log message with newline to stderr
func WriteLogLine(format string, args ...interface{}) {
	fmt.Fprintf(LogWriter, format+"\n", args...)
}

// SetWriters sets custom writers (useful for testing)
func SetWriters(dataWriter, logWriter io.Writer) {
	DataWriter = dataWriter
	LogWriter = logWriter
}
```

#### 3. 更新所有命令以使用正确的输出目标

**文件**: `cmd/query_tools.go` (修改, ~10 行)

```go
func runQueryTools(cmd *cobra.Command, args []string) error {
	// Log messages go to stderr
	if verbose {
		output.WriteLogLine("Parsing session file...")
	}

	// ... existing logic ...

	// Data output goes to stdout (via formatter)
	if err := outputTools(tools); err != nil {
		return err
	}

	return nil
}
```

**审计所有命令**:
- `cmd/parse.go`
- `cmd/analyze.go`
- `cmd/query*.go`
- `cmd/stats*.go`

确保：
- 进度信息 → stderr
- 调试信息 → stderr
- 警告/错误 → stderr
- 数据输出 → stdout

#### 4. 更新格式化器

**文件**: `cmd/output.go` (修改, ~10 行)

```go
// outputTools outputs tools in the specified format
func outputTools(tools []parser.ToolCall) error {
	// Determine output format
	format := getOutputFormat()

	// All formatters write to stdout via output.DataWriter
	switch format {
	case "json":
		return formatJSON(tools, output.DataWriter)
	case "markdown":
		return formatMarkdown(tools, output.DataWriter)
	case "tsv":
		return formatTSV(tools, output.DataWriter)
	default:
		return fmt.Errorf("unknown output format: %s", format)
	}
}
```

### TDD 步骤

**测试文件**: `internal/output/writer_test.go` (新建, ~40 行)

```go
package output

import (
	"bytes"
	"testing"
)

func TestWriteData(t *testing.T) {
	var dataBuf, logBuf bytes.Buffer
	SetWriters(&dataBuf, &logBuf)

	WriteData("data output")

	if dataBuf.String() != "data output" {
		t.Errorf("Expected 'data output', got '%s'", dataBuf.String())
	}

	if logBuf.String() != "" {
		t.Error("LogWriter should be empty")
	}
}

func TestWriteLog(t *testing.T) {
	var dataBuf, logBuf bytes.Buffer
	SetWriters(&dataBuf, &logBuf)

	WriteLog("log message")

	if logBuf.String() != "log message" {
		t.Errorf("Expected 'log message', got '%s'", logBuf.String())
	}

	if dataBuf.String() != "" {
		t.Error("DataWriter should be empty")
	}
}

func TestWriteLogLine(t *testing.T) {
	var dataBuf, logBuf bytes.Buffer
	SetWriters(&dataBuf, &logBuf)

	WriteLogLine("log line")

	if logBuf.String() != "log line\n" {
		t.Errorf("Expected 'log line\\n', got '%s'", logBuf.String())
	}
}

func TestSetWriters(t *testing.T) {
	var customData, customLog bytes.Buffer
	SetWriters(&customData, &customLog)

	if DataWriter != &customData {
		t.Error("DataWriter not set correctly")
	}

	if LogWriter != &customLog {
		t.Error("LogWriter not set correctly")
	}
}
```

**集成测试**: `tests/integration/stdio_test.sh` (新建, ~40 行)

```bash
#!/bin/bash
# Test stderr/stdout separation

set -e

echo "=== stderr/stdout Separation Test ==="

# Test 1: Data only on stdout
echo "[1/3] Testing data on stdout..."
DATA=$(meta-cc query tools --limit 5 --output json 2>/dev/null)
if [ -z "$DATA" ]; then
    echo "✗ No data on stdout"
    exit 1
fi
echo "✓ Data on stdout"

# Test 2: Logs only on stderr
echo "[2/3] Testing logs on stderr..."
# Assuming --verbose flag exists for logging
LOGS=$(meta-cc query tools --limit 5 --verbose >/dev/null 2>&1)
# If logs are on stderr, this should work
echo "✓ Logs on stderr (manual verification needed)"

# Test 3: Pipeline works without log interference
echo "[3/3] Testing pipeline without log interference..."
COUNT=$(meta-cc query tools --stream --limit 10 2>/dev/null | jq -s '. | length')
if [ "$COUNT" -ne 10 ]; then
    echo "✗ Expected 10 records, got $COUNT (logs may be interfering)"
    exit 1
fi
echo "✓ Pipeline works cleanly"

echo ""
echo "=== All stdio Tests Passed ✅ ==="
```

### 交付物

**新建文件**:
- `internal/output/writer.go` (~30 行)
- `internal/output/writer_test.go` (~40 行)
- `tests/integration/stdio_test.sh` (~40 行)

**修改文件**:
- 所有 `cmd/*.go` 文件：将日志输出改为 stderr (审计后确定具体行数，估计 ~20 行总计)

**代码量**: ~40 行（新增 Go 源代码，不含测试和审计修改）

### 验收标准

- ✅ 数据输出仅在 stdout
- ✅ 日志输出仅在 stderr
- ✅ 管道工作流不受日志干扰：`meta-cc ... 2>/dev/null | jq`
- ✅ 可以分别重定向数据和日志
- ✅ 所有单元测试通过
- ✅ 集成测试通过

---

## Stage 11.4: 文档 - Cookbook 和组合使用指南

### 目标

创建实用的分析模式文档和工具组合示例，帮助用户充分利用 meta-cc 的 Unix 可组合性。

### 交付物

#### 1. Cookbook: 常见分析模式

**文件**: `docs/cookbook.md` (新建, ~300 行)

```markdown
# meta-cc Cookbook: Common Analysis Patterns

This cookbook provides practical examples for common analysis tasks using meta-cc.

## Table of Contents

1. [Error Analysis](#error-analysis)
2. [Performance Profiling](#performance-profiling)
3. [Tool Usage Statistics](#tool-usage-statistics)
4. [File Modification Tracking](#file-modification-tracking)
5. [Time-Based Analysis](#time-based-analysis)
6. [Advanced Filtering](#advanced-filtering)
7. [Data Export and Reporting](#data-export-and-reporting)
8. [Debugging Workflows](#debugging-workflows)
9. [CI/CD Integration](#cicd-integration)
10. [Custom Metrics](#custom-metrics)

---

## Error Analysis

### 1.1 Find all errors in the current session

```bash
# Simple error query
meta-cc query tools --where "status='error'" --output json

# Stream for pipeline processing
meta-cc query tools --stream | jq 'select(.Status == "error")'
```

### 1.2 Group errors by tool

```bash
# Using stats aggregate
meta-cc stats aggregate --group-by tool --metrics "count,error_rate" | \
  jq '.[] | select(.metrics.error_rate > 0)'

# Using jq for custom grouping
meta-cc query tools --stream | \
  jq -s 'group_by(.ToolName) |
         map({tool: .[0].ToolName,
              error_count: map(select(.Status == "error")) | length})'
```

### 1.3 Extract error messages and patterns

```bash
# Extract unique error patterns
meta-cc query tools --where "status='error'" --stream | \
  jq -r '.Error' | \
  grep -oP '(permission denied|not found|timeout|failed to)' | \
  sort | uniq -c | sort -rn

# Find permission errors
meta-cc query tools --stream | \
  jq -c 'select(.Status == "error" and (.Error | contains("permission")))' | \
  jq -r '[.ToolName, .Error] | @tsv'
```

### 1.4 Time-based error analysis

```bash
# Errors in the last hour
meta-cc query tools --where "status='error'" --since "1 hour ago" --stream

# Error rate trend over time
meta-cc stats time-series --metric error-rate --interval hour | \
  jq -r '.[] | [.timestamp, .value] | @tsv' | \
  gnuplot -e "set terminal dumb; plot '-' using 2 with lines"
```

---

## Performance Profiling

### 2.1 Find slow operations

```bash
# Tools that took longer than 5 seconds
meta-cc query tools --where "duration > 5000" --stream | \
  jq -r '[.ToolName, .Duration] | @tsv' | \
  sort -k2 -rn

# Top 10 slowest operations
meta-cc query tools --sort-by duration --reverse --limit 10 --output json
```

### 2.2 Average duration by tool

```bash
meta-cc stats aggregate --group-by tool --metrics "count,avg_duration,p90,p95" | \
  jq -r '.[] | [.group_value, .metrics.avg_duration, .metrics.p95] | @tsv' | \
  column -t
```

### 2.3 Performance timeline

```bash
# Tool usage frequency over time
meta-cc stats time-series --metric tool-calls --interval hour | \
  jq -r '.[] | [.timestamp, .value] | @csv'

# Duration trends
meta-cc stats time-series --metric avg-duration --interval day
```

---

## Tool Usage Statistics

### 3.1 Tool usage distribution

```bash
# Count by tool
meta-cc stats aggregate --group-by tool --metrics count | \
  jq -r '.[] | [.group_value, .metrics.count] | @tsv' | \
  awk '{print $2 ": " $1}' | \
  sort -rn

# Pie chart data
meta-cc stats aggregate --group-by tool --metrics count | \
  jq -r '.[] | "\(.group_value): \(.metrics.count)"'
```

### 3.2 Most frequently used tools

```bash
# Top 5 tools
meta-cc stats aggregate --group-by tool --metrics count | \
  jq 'sort_by(.metrics.count) | reverse | .[0:5]'
```

### 3.3 Tool success rates

```bash
meta-cc stats aggregate --group-by tool --metrics "count,error_rate" | \
  jq -r '.[] | [.group_value,
                .metrics.count,
                (.metrics.error_rate * 100 | tostring + "%")] | @tsv' | \
  column -t -s $'\t' -N "Tool,Count,Error Rate"
```

---

## File Modification Tracking

### 4.1 Most edited files

```bash
meta-cc stats files --sort-by edit-count --top 20 | \
  jq -r '.[] | [.file_path, .edit_count] | @tsv' | \
  column -t
```

### 4.2 Files with high error rates

```bash
meta-cc stats files --sort-by error-rate | \
  jq '.[] | select(.error_rate > 0.1)' | \
  jq -r '[.file_path, .error_count, (.error_rate * 100 | tostring + "%")] | @tsv'
```

### 4.3 File operation summary

```bash
meta-cc stats files --top 10 | \
  jq -r '.[] | [.file_path,
                .read_count,
                .edit_count,
                .write_count,
                .error_count] | @tsv' | \
  column -t -s $'\t' -N "File,Reads,Edits,Writes,Errors"
```

---

## Time-Based Analysis

### 5.1 Activity by hour of day

```bash
meta-cc stats time-series --metric tool-calls --interval hour | \
  jq -r '.[] | [(.timestamp | strftime("%H:00")), .value] | @tsv' | \
  awk '{hours[$1] += $2} END {for (h in hours) print h, hours[h]}' | \
  sort
```

### 5.2 Session duration

```bash
# Get first and last timestamps
FIRST=$(meta-cc query tools --limit 1 --sort-by timestamp | jq -r '.[0].Timestamp')
LAST=$(meta-cc query tools --limit 1 --sort-by timestamp --reverse | jq -r '.[0].Timestamp')

# Calculate duration (requires date command)
START_SEC=$(date -d "$FIRST" +%s)
END_SEC=$(date -d "$LAST" +%s)
DURATION=$((END_SEC - START_SEC))

echo "Session duration: $((DURATION / 3600)) hours $((DURATION % 3600 / 60)) minutes"
```

### 5.3 Identify productive hours

```bash
meta-cc stats time-series --metric tool-calls --interval hour | \
  jq -r '.[] | [(.timestamp | strftime("%Y-%m-%d %H:00")), .value] | @tsv' | \
  awk '{if ($2 > 20) print $1 " - High activity: " $2 " tool calls"}'
```

---

## Advanced Filtering

### 6.1 Complex boolean queries

```bash
# Bash errors OR long-running operations
meta-cc query tools --where "(tool='Bash' AND status='error') OR duration>5000" --stream

# Successful file operations
meta-cc query tools --where "tool IN ('Read','Edit','Write') AND status='success'"
```

### 6.2 Pattern matching

```bash
# Tools matching pattern
meta-cc query tools --where "tool LIKE 'meta%'" --stream

# Error messages matching regex
meta-cc query tools --where "status='error' AND error REGEXP 'permission.*denied'"
```

### 6.3 Range queries

```bash
# Operations in a time range
meta-cc query tools --where "timestamp BETWEEN '2025-10-01' AND '2025-10-03'"

# Moderate duration operations (not too fast, not too slow)
meta-cc query tools --where "duration BETWEEN 1000 AND 5000"
```

---

## Data Export and Reporting

### 7.1 Export to CSV

```bash
# Tool statistics to CSV
meta-cc stats aggregate --group-by tool --metrics "count,error_rate,avg_duration" | \
  jq -r '["Tool","Count","Error Rate","Avg Duration"],
         (.[] | [.group_value,
                 .metrics.count,
                 .metrics.error_rate,
                 .metrics.avg_duration]) | @csv'
```

### 7.2 Generate HTML report

```bash
# Create simple HTML table
cat <<EOF > report.html
<html><body>
<h1>meta-cc Session Report</h1>
<table border="1">
<tr><th>Tool</th><th>Count</th><th>Error Rate</th></tr>
EOF

meta-cc stats aggregate --group-by tool --metrics "count,error_rate" | \
  jq -r '.[] | "<tr><td>\(.group_value)</td><td>\(.metrics.count)</td><td>\(.metrics.error_rate)</td></tr>"' >> report.html

echo "</table></body></html>" >> report.html
```

### 7.3 JSON summary for dashboards

```bash
# Create dashboard data
cat <<EOF > dashboard.json
{
  "session_stats": $(meta-cc analyze stats),
  "tool_distribution": $(meta-cc stats aggregate --group-by tool --metrics count),
  "error_summary": $(meta-cc analyze errors),
  "file_hotspots": $(meta-cc stats files --top 10)
}
EOF
```

---

## Debugging Workflows

### 8.1 Trace tool call sequences

```bash
# Get tool call order
meta-cc query tools --sort-by timestamp | \
  jq -r '.[] | [.Timestamp, .ToolName, .Status] | @tsv'

# Identify error sequences (errors followed by retries)
meta-cc query sequences --pattern "error,success" --window 3
```

### 8.2 Find repeated errors

```bash
# Group errors by error message
meta-cc query tools --where "status='error'" --stream | \
  jq -r '.Error' | \
  sort | uniq -c | sort -rn | head -10
```

### 8.3 Context around errors

```bash
# Get 2 tools before and after each error
meta-cc query context --error-signature "permission denied" --window 2
```

---

## CI/CD Integration

### 9.1 Fail build on high error rate

```bash
#!/bin/bash
# ci-check-errors.sh

ERROR_RATE=$(meta-cc stats aggregate --group-by status --metrics count | \
  jq -r '.[] | select(.group_value == "error") | .metrics.count' || echo 0)

TOTAL=$(meta-cc analyze stats | jq -r '.total_tools')

if [ "$TOTAL" -gt 0 ]; then
  RATE=$(awk "BEGIN {print $ERROR_RATE / $TOTAL}")
  if (( $(awk "BEGIN {print ($RATE > 0.1)}") )); then
    echo "Error rate too high: $(awk "BEGIN {print $RATE * 100}")%"
    exit 1
  fi
fi

echo "Error rate acceptable: $(awk "BEGIN {print $RATE * 100}")%"
```

### 9.2 Generate performance report

```bash
#!/bin/bash
# ci-performance-report.sh

echo "=== Performance Report ==="
echo ""

echo "Slowest Operations:"
meta-cc query tools --sort-by duration --reverse --limit 5 | \
  jq -r '.[] | "  - \(.ToolName): \(.Duration)ms"'

echo ""
echo "Average Duration by Tool:"
meta-cc stats aggregate --group-by tool --metrics avg_duration | \
  jq -r '.[] | "  - \(.group_value): \(.metrics.avg_duration | round)ms"'
```

---

## Custom Metrics

### 10.1 Calculate custom ratios

```bash
# Edit/Read ratio (how often we edit vs read)
EDITS=$(meta-cc query tools --where "tool='Edit'" --stream | wc -l)
READS=$(meta-cc query tools --where "tool='Read'" --stream | wc -l)
RATIO=$(awk "BEGIN {print $EDITS / ($READS + 1)}")
echo "Edit/Read ratio: $RATIO"
```

### 10.2 Identify anti-patterns

```bash
# Find repeated read-edit-read patterns (inefficient)
meta-cc query sequences --pattern "Read,Edit,Read" | \
  jq '. | length' | \
  awk '{if ($1 > 10) print "Warning: " $1 " inefficient read-edit-read patterns detected"}'
```

### 10.3 Tool diversity score

```bash
# Calculate how many different tools are used
UNIQUE_TOOLS=$(meta-cc stats aggregate --group-by tool --metrics count | jq '. | length')
TOTAL_CALLS=$(meta-cc analyze stats | jq -r '.total_tools')
DIVERSITY=$(awk "BEGIN {print $UNIQUE_TOOLS / sqrt($TOTAL_CALLS)}")
echo "Tool diversity score: $DIVERSITY"
```

---

## Tips and Best Practices

### Use `--stream` for large datasets

Streaming output is more efficient for large result sets:
```bash
# Good: Stream and filter incrementally
meta-cc query tools --stream | jq 'select(.Status == "error")' | head -100

# Avoid: Load all data into memory first
meta-cc query tools --limit 100000 | jq 'select(.Status == "error")'
```

### Combine with standard Unix tools

meta-cc integrates well with:
- **jq**: JSON processing and filtering
- **grep**: Text pattern matching
- **awk**: Text processing and calculations
- **sort, uniq**: Data aggregation
- **column**: Table formatting
- **gnuplot**: Data visualization

### Redirect logs when scripting

Always redirect stderr to avoid log interference:
```bash
# In scripts
DATA=$(meta-cc query tools 2>/dev/null)

# In pipelines
meta-cc query tools --stream 2>/dev/null | jq '.ToolName'
```

### Use exit codes in conditionals

```bash
if meta-cc query tools --where "status='error'"; then
  echo "Errors found!"
  # Handle errors...
else
  EXIT_CODE=$?
  if [ $EXIT_CODE -eq 2 ]; then
    echo "No errors (good!)"
  else
    echo "Query failed"
  fi
fi
```

---

## See Also

- [CLI Composability Guide](./cli-composability.md) - Integration with jq, grep, awk
- [meta-cc README](../README.md) - Full command reference
- [Examples and Usage](./examples-usage.md) - Getting started guide
```

#### 2. CLI Composability Guide: 工具集成示例

**文件**: `docs/cli-composability.md` (新建, ~200 行)

```markdown
# CLI Composability: Integrating meta-cc with Unix Tools

This guide demonstrates how to integrate meta-cc with standard Unix tools for powerful data analysis workflows.

## Table of Contents

1. [jq Integration](#jq-integration)
2. [grep Integration](#grep-integration)
3. [awk Integration](#awk-integration)
4. [sed Integration](#sed-integration)
5. [Combining Multiple Tools](#combining-multiple-tools)
6. [Advanced Patterns](#advanced-patterns)

---

## jq Integration

jq is a lightweight JSON processor, perfect for filtering and transforming meta-cc output.

### Basic Filtering

```bash
# Select errors only
meta-cc query tools --stream | jq 'select(.Status == "error")'

# Select specific tool
meta-cc query tools --stream | jq 'select(.ToolName == "Bash")'

# Multiple conditions
meta-cc query tools --stream | \
  jq 'select(.ToolName == "Bash" and .Status == "error")'
```

### Field Extraction

```bash
# Extract tool names only
meta-cc query tools --stream | jq -r '.ToolName'

# Extract multiple fields as TSV
meta-cc query tools --stream | \
  jq -r '[.ToolName, .Status, .Duration] | @tsv'

# Create custom objects
meta-cc query tools --stream | \
  jq '{tool: .ToolName, failed: (.Status == "error"), duration_sec: (.Duration / 1000)}'
```

### Aggregation

```bash
# Count by tool (using jq)
meta-cc query tools --stream | \
  jq -s 'group_by(.ToolName) |
         map({tool: .[0].ToolName, count: length}) |
         sort_by(.count) | reverse'

# Average duration by tool
meta-cc query tools --stream | \
  jq -s 'group_by(.ToolName) |
         map({tool: .[0].ToolName,
              avg_duration: (map(.Duration) | add / length)})'
```

### Conditional Processing

```bash
# Different output based on status
meta-cc query tools --stream | \
  jq 'if .Status == "error" then
        {error: .ToolName, message: .Error}
      else
        {success: .ToolName}
      end'

# Add severity field
meta-cc query tools --stream | \
  jq '. + if .Duration > 5000 then
             {severity: "high"}
           elif .Duration > 2000 then
             {severity: "medium"}
           else
             {severity: "low"}
           end'
```

---

## grep Integration

grep is excellent for pattern matching and filtering text output.

### Pattern Matching

```bash
# Find permission errors
meta-cc query tools --where "status='error'" --stream | \
  jq -r '.Error' | \
  grep -i "permission"

# Case-insensitive search
meta-cc query tools --stream | \
  jq -r '.ToolName + ": " + .Error' | \
  grep -i "timeout\|failed\|denied"

# Invert match (exclude pattern)
meta-cc query tools --stream | \
  jq -r '.ToolName' | \
  grep -v "Bash"  # Exclude Bash tools
```

### Counting and Statistics

```bash
# Count occurrences
meta-cc query tools --stream | \
  jq -r '.Error' | \
  grep -c "permission denied"

# Count unique error patterns
meta-cc query tools --where "status='error'" --stream | \
  jq -r '.Error' | \
  grep -oP '(permission|timeout|not found|failed)' | \
  sort | uniq -c
```

### Context Lines

```bash
# Show 2 lines before and after match
meta-cc query tools --output json | \
  jq -r '.[] | .Error' | \
  grep -B 2 -A 2 "permission denied"
```

---

## awk Integration

awk is powerful for text processing, calculations, and formatting.

### Field Processing

```bash
# Extract and format fields
meta-cc query tools --stream | \
  jq -r '[.ToolName, .Status, .Duration] | @tsv' | \
  awk '{print "Tool:", $1, "Status:", $2, "Duration:", $3 "ms"}'

# Custom column formatting
meta-cc stats aggregate --group-by tool | \
  jq -r '.[] | [.group_value, .metrics.count, .metrics.error_rate] | @tsv' | \
  awk '{printf "%-15s Count: %5d Error Rate: %.2f%%\n", $1, $2, $3 * 100}'
```

### Calculations

```bash
# Sum durations
meta-cc query tools --stream | \
  jq -r '.Duration' | \
  awk '{sum += $1} END {print "Total duration:", sum "ms"}'

# Average, min, max
meta-cc query tools --stream | \
  jq -r '.Duration' | \
  awk '{sum+=$1; if($1>max) max=$1; if(min==""|$1<min) min=$1; count++}
       END {print "Avg:", sum/count, "Min:", min, "Max:", max}'
```

### Conditional Processing

```bash
# Flag slow operations
meta-cc query tools --stream | \
  jq -r '[.ToolName, .Duration] | @tsv' | \
  awk '{if ($2 > 5000) print "SLOW:", $1, $2 "ms"; else print "OK:", $1, $2 "ms"}'

# Count by category
meta-cc query tools --stream | \
  jq -r '.Duration' | \
  awk '{
    if ($1 < 1000) fast++;
    else if ($1 < 5000) medium++;
    else slow++;
  }
  END {print "Fast:", fast, "Medium:", medium, "Slow:", slow}'
```

### Grouping and Aggregation

```bash
# Group and sum by tool
meta-cc query tools --stream | \
  jq -r '[.ToolName, .Duration] | @tsv' | \
  awk '{duration[$1] += $2; count[$1]++}
       END {for (tool in duration)
              print tool, "Total:", duration[tool] "ms",
                   "Avg:", duration[tool]/count[tool] "ms"}'
```

---

## sed Integration

sed is useful for stream editing and text transformation.

### Text Replacement

```bash
# Normalize tool names
meta-cc query tools --stream | \
  jq -r '.ToolName' | \
  sed 's/Bash/Shell/g; s/Edit/Modify/g'

# Clean up error messages
meta-cc query tools --where "status='error'" --stream | \
  jq -r '.Error' | \
  sed 's|/home/[^/]*/|~/|g'  # Replace home paths with ~
```

### Filtering Lines

```bash
# Delete empty lines
meta-cc query tools --stream | \
  jq -r '.Error // empty' | \
  sed '/^$/d'

# Keep only lines matching pattern
meta-cc query tools --stream | \
  jq -r '.ToolName + ": " + .Status' | \
  sed -n '/error/p'  # Print only lines with "error"
```

---

## Combining Multiple Tools

Real-world scenarios often combine multiple tools for complex analysis.

### Example 1: Error Pattern Analysis

```bash
# Find top error patterns with counts
meta-cc query tools --where "status='error'" --stream | \
  jq -r '.Error' | \                    # Extract error messages
  grep -oP '(permission|timeout|not found|failed to \w+)' | \ # Extract patterns
  sed 's/failed to \w*/failed to .../g' | \  # Normalize
  sort | \                               # Sort
  uniq -c | \                            # Count unique
  sort -rn | \                           # Sort by count desc
  head -10 | \                           # Top 10
  awk '{printf "%3d: %s\n", $1, substr($0, index($0, $2))}'  # Format
```

### Example 2: Performance Report

```bash
# Generate performance summary
{
  echo "=== Performance Report ==="
  echo ""
  echo "Tool Usage Statistics:"
  meta-cc stats aggregate --group-by tool --metrics "count,avg_duration" | \
    jq -r '.[] | [.group_value, .metrics.count, .metrics.avg_duration] | @tsv' | \
    awk '{printf "  %-15s Count: %5d Avg: %6.0fms\n", $1, $2, $3}' | \
    sort -k3 -rn

  echo ""
  echo "Slowest Operations:"
  meta-cc query tools --sort-by duration --reverse --limit 5 --stream | \
    jq -r '[.ToolName, .Duration, .Status] | @tsv' | \
    awk '{printf "  %s: %dms (%s)\n", $1, $2, $3}'
} > performance_report.txt
```

### Example 3: File Modification Heatmap

```bash
# Identify most problematic files
meta-cc stats files --sort-by error-count | \
  jq -r '.[] | select(.error_count > 0) |
             [.file_path, .edit_count, .error_count, .error_rate] | @tsv' | \
  awk '{printf "%s\t%d edits\t%d errors\t%.1f%%\n", $1, $2, $3, $4*100}' | \
  column -t -s $'\t'
```

---

## Advanced Patterns

### Streaming Pipeline with Feedback

```bash
# Process stream with progress indicator
meta-cc query tools --stream 2>/dev/null | \
  jq -c 'select(.Status == "error")' | \
  tee >(wc -l >&2) | \  # Count to stderr
  jq -r '.Error' | \
  sort | uniq -c
```

### Parallel Processing

```bash
# Process large dataset in parallel (requires GNU parallel)
meta-cc query tools --stream | \
  parallel --pipe -N 1000 \
    'jq -s "group_by(.ToolName) | map({tool: .[0].ToolName, count: length})"' | \
  jq -s 'flatten | group_by(.tool) |
         map({tool: .[0].tool, count: (map(.count) | add)})'
```

### Creating Reusable Filters

```bash
# Save common jq filter
cat > /tmp/filter-errors.jq <<'EOF'
select(.Status == "error") |
{
  tool: .ToolName,
  error: .Error,
  duration: .Duration,
  severity: (if .Duration > 5000 then "high" else "low" end)
}
EOF

# Use saved filter
meta-cc query tools --stream | jq -f /tmp/filter-errors.jq
```

---

## Performance Tips

### 1. Use `--stream` for Large Datasets

Streaming avoids loading everything into memory:
```bash
# Good: Process incrementally
meta-cc query tools --stream | jq 'select(.Status == "error")' | head -100

# Avoid: Load all, then filter
meta-cc query tools --output json | jq '.[] | select(.Status == "error")' | head -100
```

### 2. Filter Early

Apply filters in meta-cc before piping:
```bash
# Good: Filter in query
meta-cc query tools --where "status='error'" --stream | jq -r '.Error'

# Less efficient: Filter after
meta-cc query tools --stream | jq 'select(.Status == "error") | .Error'
```

### 3. Use Compact Output

Use `-c` flag in jq for compact JSON:
```bash
# Compact output (faster, less space)
meta-cc query tools --stream | jq -c 'select(.Status == "error")'
```

### 4. Redirect stderr

Suppress logs when scripting:
```bash
meta-cc query tools --stream 2>/dev/null | jq '.ToolName'
```

---

## Troubleshooting

### jq: parse error

**Problem**: jq fails to parse output

**Solution**: Ensure `--stream` or `--output json` is used
```bash
# Wrong: Default output might not be JSON
meta-cc query tools | jq '.ToolName'

# Correct:
meta-cc query tools --output json | jq '.[] | .ToolName'
meta-cc query tools --stream | jq -r '.ToolName'
```

### Logs interfering with pipeline

**Problem**: stderr logs appear in pipeline output

**Solution**: Redirect stderr to /dev/null
```bash
meta-cc query tools --stream 2>/dev/null | jq '.ToolName'
```

### Exit code confusion

**Problem**: Script doesn't handle "no results" correctly

**Solution**: Check for exit code 2
```bash
meta-cc query tools --where "tool='NonExistent'" 2>/dev/null
EXIT_CODE=$?

case $EXIT_CODE in
  0) echo "Found results" ;;
  2) echo "No results" ;;
  *) echo "Error" ;;
esac
```

---

## See Also

- [Cookbook](./cookbook.md) - Common analysis patterns
- [meta-cc README](../README.md) - Full command reference
- [jq Manual](https://stedolan.github.io/jq/manual/) - jq documentation
```

### 验收标准

- ✅ `docs/cookbook.md` 包含 10+ 实用分析模式
- ✅ `docs/cli-composability.md` 包含 jq/grep/awk/sed 集成示例
- ✅ 所有文档示例可执行并验证通过
- ✅ README.md 更新，包含 Unix 可组合性章节
- ✅ 文档清晰、实用、易懂

---

## 集成测试：Unix 可组合性端到端验证

### 测试脚本

**文件**: `tests/integration/unix_composability_test.sh` (新建, ~100 行)

```bash
#!/bin/bash
# Phase 11 Unix Composability Integration Test

set -e

echo "=== Phase 11 Unix Composability Test ==="
echo ""

# Step 1: Test JSONL streaming
echo "[1/4] Testing JSONL streaming..."

STREAM_OUTPUT=$(meta-cc query tools --stream --limit 5 2>/dev/null)
LINE_COUNT=$(echo "$STREAM_OUTPUT" | wc -l)

if [ "$LINE_COUNT" -ne 5 ]; then
    echo "  ✗ Expected 5 lines, got $LINE_COUNT"
    exit 1
fi

# Verify each line is valid JSON
echo "$STREAM_OUTPUT" | while IFS= read -r line; do
    if ! echo "$line" | jq empty 2>/dev/null; then
        echo "  ✗ Invalid JSON line: $line"
        exit 1
    fi
done

echo "  ✓ JSONL streaming works"

# Step 2: Test exit codes
echo "[2/4] Testing exit codes..."

# Test: Success with results (exit 0)
meta-cc query tools --limit 5 >/dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "  ✗ Expected exit 0 for success"
    exit 1
fi

# Test: Success without results (exit 2)
meta-cc query tools --where "tool='NonExistentTool'" >/dev/null 2>&1
if [ $? -ne 2 ]; then
    echo "  ✗ Expected exit 2 for no results"
    exit 1
fi

echo "  ✓ Exit codes work correctly"

# Step 3: Test stderr/stdout separation
echo "[3/4] Testing stderr/stdout separation..."

# Data should be on stdout
DATA=$(meta-cc query tools --limit 5 --output json 2>/dev/null)
if [ -z "$DATA" ]; then
    echo "  ✗ No data on stdout"
    exit 1
fi

# Verify data is valid JSON
if ! echo "$DATA" | jq empty 2>/dev/null; then
    echo "  ✗ Invalid JSON on stdout"
    exit 1
fi

echo "  ✓ stderr/stdout separation works"

# Step 4: Test pipeline workflows
echo "[4/4] Testing pipeline workflows..."

# jq integration
JQ_COUNT=$(meta-cc query tools --stream --limit 10 2>/dev/null | jq -s '. | length')
if [ "$JQ_COUNT" -ne 10 ]; then
    echo "  ✗ jq pipeline failed: expected 10, got $JQ_COUNT"
    exit 1
fi

# grep integration
GREP_COUNT=$(meta-cc query tools --stream --limit 100 2>/dev/null | grep -c '"ToolName"' || true)
if [ "$GREP_COUNT" -eq 0 ]; then
    echo "  ✗ grep pipeline failed: no matches"
    exit 1
fi

# awk integration
AWK_OUTPUT=$(meta-cc query tools --stream --limit 5 2>/dev/null | \
              jq -r '.ToolName' | \
              awk '{print "Tool: " $1}' | \
              wc -l)
if [ "$AWK_OUTPUT" -ne 5 ]; then
    echo "  ✗ awk pipeline failed"
    exit 1
fi

echo "  ✓ Pipeline workflows work (jq, grep, awk)"

echo ""
echo "=== All Phase 11 Tests Passed ✅ ==="
echo ""
echo "Summary:"
echo "  - JSONL streaming: working"
echo "  - Exit codes: 0 (success), 2 (no results)"
echo "  - I/O separation: stdout=data, stderr=logs"
echo "  - Pipelines: jq, grep, awk integration verified"
```

---

## 文档更新

### README.md 新增章节

**新增内容** (~150 行):

````markdown
## Unix Composability (Phase 11)

Phase 11 optimizes meta-cc for seamless integration with Unix pipelines and standard tools.

### JSONL Streaming Output

Stream data for efficient pipeline processing:

```bash
# Basic streaming
meta-cc query tools --stream

# Pipeline with jq
meta-cc query tools --stream | jq 'select(.Status == "error")'

# Pipeline with grep
meta-cc query tools --stream | jq -r '.Error' | grep -i "permission"

# Pipeline with awk
meta-cc query tools --stream | \
  jq -r '[.ToolName, .Duration] | @tsv' | \
  awk '{sum+=$2} END {print "Total:", sum "ms"}'
```

**JSONL Format**:
```
{"uuid":"1","tool":"Bash","status":"success",...}
{"uuid":"2","tool":"Edit","status":"success",...}
{"uuid":"3","tool":"Read","status":"error",...}
```

### Standard Exit Codes

meta-cc follows Unix conventions for exit codes:

| Exit Code | Meaning | Example |
|-----------|---------|---------|
| 0 | Success (with results) | `meta-cc query tools --limit 10` |
| 1 | Error (parsing, I/O, etc.) | `meta-cc query tools --where "invalid syntax"` |
| 2 | Success (no results) | `meta-cc query tools --where "tool='NonExistent'"` |

**Usage in scripts**:
```bash
if meta-cc query tools --where "status='error'"; then
  echo "Errors found!"
  # Handle errors...
else
  EXIT_CODE=$?
  if [ $EXIT_CODE -eq 2 ]; then
    echo "No errors (good!)"
  else
    echo "Query failed"
    exit 1
  fi
fi
```

### stderr/stdout Separation

meta-cc separates logs and data for clean pipeline processing:

- **stdout**: Command output data (JSON, Markdown, TSV)
- **stderr**: Diagnostic messages (logs, warnings, errors)

```bash
# Redirect data only
meta-cc query tools --output json > data.json

# Redirect logs only
meta-cc query tools --output json 2> debug.log

# Separate both
meta-cc query tools --output json > data.json 2> debug.log

# Suppress logs in pipelines
meta-cc query tools --stream 2>/dev/null | jq '.ToolName'
```

### Common Pipeline Patterns

**Error Analysis**:
```bash
# Find top error patterns
meta-cc query tools --where "status='error'" --stream | \
  jq -r '.Error' | \
  grep -oP '(permission|timeout|not found)' | \
  sort | uniq -c | sort -rn
```

**Performance Profiling**:
```bash
# Average duration by tool
meta-cc stats aggregate --group-by tool --metrics avg_duration | \
  jq -r '.[] | [.group_value, .metrics.avg_duration] | @tsv' | \
  column -t
```

**Tool Usage Statistics**:
```bash
# Tool distribution
meta-cc query tools --stream | \
  jq -r '.ToolName' | \
  sort | uniq -c | sort -rn | \
  awk '{print $2 ": " $1}'
```

**File Modification Tracking**:
```bash
# Most edited files with error rates
meta-cc stats files --sort-by edit-count --top 10 | \
  jq -r '.[] | [.file_path, .edit_count, (.error_rate * 100 | tostring + "%")] | @tsv' | \
  column -t
```

### See Also

- [Cookbook](docs/cookbook.md) - 10+ practical analysis patterns
- [CLI Composability Guide](docs/cli-composability.md) - Integration with jq, grep, awk
- [Examples and Usage](docs/examples-usage.md) - Getting started guide
````

---

## Phase 11 验收清单

### 功能验收

- [ ] **Stage 11.1: JSONL 流式输出**
  - [ ] `--stream` 标志在所有 query 和 stats 命令中可用
  - [ ] 输出为有效的 JSONL 格式（每行一个 JSON 对象）
  - [ ] 与 jq 管道集成工作正常
  - [ ] 单元测试通过

- [ ] **Stage 11.2: 退出码标准化**
  - [ ] 退出码 0：成功（有结果）
  - [ ] 退出码 1：错误（解析、I/O 等）
  - [ ] 退出码 2：成功（无结果）
  - [ ] 脚本可以基于退出码进行条件判断
  - [ ] 单元测试通过

- [ ] **Stage 11.3: stderr/stdout 分离**
  - [ ] 所有数据输出到 stdout
  - [ ] 所有日志输出到 stderr
  - [ ] 管道不受日志干扰
  - [ ] 可以分别重定向数据和日志
  - [ ] 单元测试通过

- [ ] **Stage 11.4: 文档**
  - [ ] `docs/cookbook.md` 完成（10+ 模式）
  - [ ] `docs/cli-composability.md` 完成（5+ 工具示例）
  - [ ] 所有示例可执行并验证
  - [ ] README.md 更新

### 集成验收

- [ ] **端到端测试**
  - [ ] 集成测试脚本通过（`unix_composability_test.sh`）
  - [ ] jq 管道工作流验证
  - [ ] grep 管道工作流验证
  - [ ] awk 管道工作流验证
  - [ ] 真实项目验证（meta-cc, NarrativeForge, claude-tmux）

- [ ] **无回归验证**
  - [ ] 所有现有单元测试通过
  - [ ] 所有现有集成测试通过
  - [ ] 现有命令行为不变

### 代码质量

- [ ] **代码量验收**
  - [ ] Stage 11.1: ~50 行（流式输出）
  - [ ] Stage 11.2: ~30 行（退出码）
  - [ ] Stage 11.3: ~40 行（I/O 分离）
  - [ ] Stage 11.4: ~80 行（文档）
  - [ ] 总计: ~200 行（目标 180-220 行）

- [ ] **测试覆盖率**
  - [ ] 单元测试覆盖率 ≥ 80%
  - [ ] 所有 TDD 测试通过
  - [ ] 边界条件测试完整

### 文档验收

- [ ] **Cookbook 质量**
  - [ ] 10+ 实用分析模式
  - [ ] 代码示例可执行
  - [ ] 覆盖常见场景
  - [ ] 包含最佳实践建议

- [ ] **Composability Guide 质量**
  - [ ] jq 集成示例（5+ 个）
  - [ ] grep/awk/sed 示例
  - [ ] 组合工具示例
  - [ ] 性能优化建议

- [ ] **README.md 更新**
  - [ ] 添加"Unix Composability"章节
  - [ ] 包含流式输出示例
  - [ ] 包含退出码说明
  - [ ] 包含管道工作流示例

---

## 项目结构（Phase 11 完成后）

```
meta-cc/
├── cmd/
│   ├── query.go                      # 更新：--stream 标志
│   ├── query_tools.go                # 更新：流式输出支持
│   ├── stats_aggregate.go            # 更新：流式输出支持
│   ├── stats_timeseries.go           # 更新：流式输出支持
│   ├── stats_files.go                # 更新：流式输出支持
│   └── root.go                       # 更新：ExitCodeError 处理
├── internal/
│   └── output/
│       ├── stream.go                 # 新增：JSONL 流式输出
│       ├── stream_test.go            # 新增：流式输出测试
│       ├── exitcode.go               # 新增：退出码定义和处理
│       ├── exitcode_test.go          # 新增：退出码测试
│       ├── writer.go                 # 新增：I/O 分离工具
│       └── writer_test.go            # 新增：I/O 分离测试
├── tests/
│   └── integration/
│       ├── streaming_test.sh         # 新增：流式输出集成测试
│       ├── exitcode_test.sh          # 新增：退出码集成测试
│       ├── stdio_test.sh             # 新增：I/O 分离集成测试
│       └── unix_composability_test.sh # 新增：综合集成测试
├── docs/
│   ├── cookbook.md                   # 新增：常见分析模式
│   └── cli-composability.md          # 新增：工具组合指南
├── plans/
│   └── 11/
│       ├── plan.md                   # 本文档
│       └── README.md                 # 快速参考
└── README.md                          # 更新：Unix 可组合性文档
```

---

## 依赖关系

**Phase 11 依赖**:
- Phase 0-10（完整的 meta-cc 工具链 + 查询和统计命令）
- Phase 9（输出格式化基础）

**Phase 11 提供**:
- JSONL 流式输出能力
- 标准化退出码
- 干净的 I/O 分离
- Unix 管道集成最佳实践

**后续 Phase 可选扩展**:
- Phase 12（查询语言增强）：SQL-like 语法优化
- Phase 13（索引功能）：SQLite 索引、跨会话查询

---

## 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 流式输出破坏现有格式 | 高 | 仅当 `--stream` 标志启用时使用；充分测试现有格式 |
| 退出码与 Cobra 框架冲突 | 中 | 使用自定义 `ExitCodeError` 类型；在 `Execute()` 中处理 |
| I/O 分离审计遗漏输出点 | 中 | 使用 grep 全面审计；代码审查 |
| 文档示例过时 | 低 | 集成测试验证所有示例；版本化文档 |
| 管道性能问题 | 低 | 性能测试验证；流式输出应该更快 |

---

## 实施优先级

**必须完成**（Phase 11 核心功能）:
1. Stage 11.1（JSONL 流式输出）- 提供管道友好的输出
2. Stage 11.2（退出码标准化）- 支持脚本编写
3. Stage 11.3（stderr/stdout 分离）- 提供干净的管道体验
4. 集成测试和基本文档更新

**推荐完成**（提升用户体验）:
5. Stage 11.4（Cookbook 和 Composability Guide）- 提供实用示例

**可选完成**（进一步优化）:
6. 性能优化（如果流式输出性能不达标）
7. 更多管道集成示例（与其他工具如 sed, cut, paste 等）

---

## Phase 11 总结

Phase 11 优化 meta-cc 的 CLI 设计，使其成为真正的 Unix 一等公民：

### 核心成果

1. **JSONL 流式输出**（Stage 11.1）
   - 每行一个 JSON 对象
   - 支持增量处理
   - 与 jq、grep、awk 无缝集成

2. **标准化退出码**（Stage 11.2）
   - 0 = 成功（有结果）
   - 1 = 错误
   - 2 = 成功（无结果）
   - 支持脚本条件判断

3. **stderr/stdout 分离**（Stage 11.3）
   - stdout：数据输出
   - stderr：日志输出
   - 干净的管道体验

4. **实用文档**（Stage 11.4）
   - Cookbook：10+ 分析模式
   - Composability Guide：工具集成示例
   - 最佳实践和优化建议

### 集成价值

- **Shell 脚本**: 可靠的退出码，易于错误处理
- **数据管道**: JSONL 流式输出，高效的增量处理
- **Unix 工具**: 与 jq、grep、awk、sed 等工具完美组合
- **自动化**: 清晰的 I/O 分离，便于日志记录和数据提取

### 用户价值

- ✅ 管道友好：与标准 Unix 工具无缝组合
- ✅ 脚本友好：标准化退出码，可靠的错误处理
- ✅ 高效处理：流式输出，支持大数据集
- ✅ 最佳实践：实用的文档和示例

**Phase 11 完成后，meta-cc 成为一个遵循 Unix 哲学的优秀 CLI 工具，能够与标准工具链完美集成。**

---

## 参考文档

- [Unix 哲学](https://en.wikipedia.org/wiki/Unix_philosophy)
- [JSON Lines 格式](https://jsonlines.org/)
- [jq Manual](https://stedolan.github.io/jq/manual/)
- [meta-cc 技术方案](../../docs/proposals/meta-cognition-proposal.md)
- [meta-cc 总体实施计划](../../docs/plan.md)
- [Phase 10 实施计划](../10/plan.md)（高级查询基础）

---

**Phase 11 实施准备就绪。开始 TDD 开发流程。**
