package analyzer

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
)

func TestDetectErrorPatterns_NoErrors(t *testing.T) {
	// 无错误的会话应返回空模式列表
	entries := []parser.SessionEntry{
		{UUID: "uuid-1", Timestamp: "2025-10-02T10:00:00.000Z"},
		{UUID: "uuid-2", Timestamp: "2025-10-02T10:01:00.000Z"},
	}
	toolCalls := []parser.ToolCall{
		{UUID: "uuid-1", ToolName: "Grep", Status: "success"},
		{UUID: "uuid-2", ToolName: "Read", Status: "success"},
	}

	patterns := DetectErrorPatterns(entries, toolCalls)

	if len(patterns) != 0 {
		t.Errorf("Expected 0 patterns for no errors, got %d", len(patterns))
	}
}

func TestDetectErrorPatterns_SingleError(t *testing.T) {
	// 单个错误（出现 1 次）不应形成模式
	entries := []parser.SessionEntry{
		{UUID: "uuid-1", Timestamp: "2025-10-02T10:00:00.000Z"},
	}
	toolCalls := []parser.ToolCall{
		{UUID: "uuid-1", ToolName: "Bash", Status: "error", Error: "command not found"},
	}

	patterns := DetectErrorPatterns(entries, toolCalls)

	if len(patterns) != 0 {
		t.Errorf("Expected 0 patterns for single error, got %d", len(patterns))
	}
}

func TestDetectErrorPatterns_RepeatedError(t *testing.T) {
	// 重复 3 次的错误应形成模式
	entries := []parser.SessionEntry{
		{UUID: "uuid-1", Timestamp: "2025-10-02T10:00:00.000Z"},
		{UUID: "uuid-2", Timestamp: "2025-10-02T10:01:00.000Z"},
		{UUID: "uuid-3", Timestamp: "2025-10-02T10:02:00.000Z"},
		{UUID: "uuid-4", Timestamp: "2025-10-02T10:03:00.000Z"},
		{UUID: "uuid-5", Timestamp: "2025-10-02T10:05:00.000Z"},
	}
	toolCalls := []parser.ToolCall{
		{UUID: "uuid-1", ToolName: "Bash", Status: "error", Error: "command not found: xyz"},
		{UUID: "uuid-2", ToolName: "Grep", Status: "success"},
		{UUID: "uuid-3", ToolName: "Bash", Status: "error", Error: "command not found: xyz"},
		{UUID: "uuid-4", ToolName: "Read", Status: "success"},
		{UUID: "uuid-5", ToolName: "Bash", Status: "error", Error: "command not found: xyz"},
	}

	patterns := DetectErrorPatterns(entries, toolCalls)

	if len(patterns) != 1 {
		t.Fatalf("Expected 1 pattern, got %d", len(patterns))
	}

	pattern := patterns[0]

	if pattern.Type != "repeated_error" {
		t.Errorf("Expected type 'repeated_error', got '%s'", pattern.Type)
	}

	if pattern.Occurrences != 3 {
		t.Errorf("Expected 3 occurrences, got %d", pattern.Occurrences)
	}

	if pattern.ToolName != "Bash" {
		t.Errorf("Expected tool name 'Bash', got '%s'", pattern.ToolName)
	}

	if !contains(pattern.Context.TurnUUIDs, "uuid-1") {
		t.Error("Expected uuid-1 in turn UUIDs")
	}

	if !contains(pattern.Context.TurnUUIDs, "uuid-3") {
		t.Error("Expected uuid-3 in turn UUIDs")
	}

	if !contains(pattern.Context.TurnUUIDs, "uuid-5") {
		t.Error("Expected uuid-5 in turn UUIDs")
	}
}

func TestDetectErrorPatterns_MultiplePatterns(t *testing.T) {
	// 测试多个不同的错误模式
	entries := []parser.SessionEntry{
		{UUID: "uuid-1", Timestamp: "2025-10-02T10:00:00.000Z"},
		{UUID: "uuid-2", Timestamp: "2025-10-02T10:01:00.000Z"},
		{UUID: "uuid-3", Timestamp: "2025-10-02T10:02:00.000Z"},
		{UUID: "uuid-4", Timestamp: "2025-10-02T10:03:00.000Z"},
		{UUID: "uuid-5", Timestamp: "2025-10-02T10:04:00.000Z"},
		{UUID: "uuid-6", Timestamp: "2025-10-02T10:05:00.000Z"},
		{UUID: "uuid-7", Timestamp: "2025-10-02T10:06:00.000Z"},
		{UUID: "uuid-8", Timestamp: "2025-10-02T10:07:00.000Z"},
	}
	toolCalls := []parser.ToolCall{
		// Pattern 1: Bash command not found (3 次)
		{UUID: "uuid-1", ToolName: "Bash", Status: "error", Error: "command not found: xyz"},
		{UUID: "uuid-2", ToolName: "Bash", Status: "error", Error: "command not found: xyz"},
		{UUID: "uuid-3", ToolName: "Bash", Status: "error", Error: "command not found: xyz"},

		// Pattern 2: Read file not found (4 次)
		{UUID: "uuid-4", ToolName: "Read", Status: "error", Error: "file not found: /tmp/test.txt"},
		{UUID: "uuid-5", ToolName: "Read", Status: "error", Error: "file not found: /tmp/test.txt"},
		{UUID: "uuid-6", ToolName: "Read", Status: "error", Error: "file not found: /tmp/test.txt"},
		{UUID: "uuid-7", ToolName: "Read", Status: "error", Error: "file not found: /tmp/test.txt"},

		// Non-pattern: Single error
		{UUID: "uuid-8", ToolName: "Write", Status: "error", Error: "permission denied"},
	}

	patterns := DetectErrorPatterns(entries, toolCalls)

	if len(patterns) != 2 {
		t.Fatalf("Expected 2 patterns, got %d", len(patterns))
	}

	// 验证模式按出现次数降序排列
	if patterns[0].Occurrences < patterns[1].Occurrences {
		t.Error("Expected patterns to be sorted by occurrences (descending)")
	}
}

func TestDetectErrorPatterns_TimeSpanCalculation(t *testing.T) {
	// 测试时间跨度计算
	entries := []parser.SessionEntry{
		{UUID: "uuid-1", Timestamp: "2025-10-02T10:00:00.000Z"},
		{UUID: "uuid-2", Timestamp: "2025-10-02T10:05:30.000Z"},
		{UUID: "uuid-3", Timestamp: "2025-10-02T10:08:00.000Z"},
	}
	toolCalls := []parser.ToolCall{
		{UUID: "uuid-1", ToolName: "Bash", Status: "error", Error: "error"},
		{UUID: "uuid-2", ToolName: "Bash", Status: "error", Error: "error"},
		{UUID: "uuid-3", ToolName: "Bash", Status: "error", Error: "error"},
	}

	patterns := DetectErrorPatterns(entries, toolCalls)

	if len(patterns) != 1 {
		t.Fatalf("Expected 1 pattern, got %d", len(patterns))
	}

	// 时间跨度应为 8 分钟 = 480 秒
	expectedSpan := 480
	if patterns[0].TimeSpanSeconds != expectedSpan {
		t.Errorf("Expected time span %d seconds, got %d", expectedSpan, patterns[0].TimeSpanSeconds)
	}
}

func TestDetectErrorPatterns_TurnIndices(t *testing.T) {
	// 测试 Turn 索引计算
	entries := []parser.SessionEntry{
		{UUID: "uuid-1", Timestamp: "2025-10-02T10:00:00.000Z"},
		{UUID: "uuid-2", Timestamp: "2025-10-02T10:01:00.000Z"},
		{UUID: "uuid-3", Timestamp: "2025-10-02T10:02:00.000Z"},
		{UUID: "uuid-4", Timestamp: "2025-10-02T10:03:00.000Z"},
		{UUID: "uuid-5", Timestamp: "2025-10-02T10:04:00.000Z"},
	}
	toolCalls := []parser.ToolCall{
		{UUID: "uuid-1", ToolName: "Bash", Status: "error", Error: "error"},
		{UUID: "uuid-3", ToolName: "Bash", Status: "error", Error: "error"},
		{UUID: "uuid-5", ToolName: "Bash", Status: "error", Error: "error"},
	}

	patterns := DetectErrorPatterns(entries, toolCalls)

	if len(patterns) != 1 {
		t.Fatalf("Expected 1 pattern, got %d", len(patterns))
	}

	// 验证索引: uuid-1 -> index 0, uuid-3 -> index 2, uuid-5 -> index 4
	expectedIndices := []int{0, 2, 4}
	if len(patterns[0].Context.TurnIndices) != len(expectedIndices) {
		t.Fatalf("Expected %d indices, got %d", len(expectedIndices), len(patterns[0].Context.TurnIndices))
	}

	for i, expected := range expectedIndices {
		if patterns[0].Context.TurnIndices[i] != expected {
			t.Errorf("Expected index %d at position %d, got %d", expected, i, patterns[0].Context.TurnIndices[i])
		}
	}
}

func TestDetectErrorPatterns_ExampleOutput(t *testing.T) {
	// 演示完整的错误模式检测
	entries := []parser.SessionEntry{
		{UUID: "uuid-1", Timestamp: "2025-10-02T10:00:00.000Z"},
		{UUID: "uuid-2", Timestamp: "2025-10-02T10:02:30.000Z"},
		{UUID: "uuid-3", Timestamp: "2025-10-02T10:05:00.000Z"},
		{UUID: "uuid-4", Timestamp: "2025-10-02T10:07:00.000Z"},
		{UUID: "uuid-5", Timestamp: "2025-10-02T10:10:00.000Z"},
	}
	toolCalls := []parser.ToolCall{
		{UUID: "uuid-1", ToolName: "Bash", Status: "error", Error: "command not found: make"},
		{UUID: "uuid-2", ToolName: "Bash", Status: "error", Error: "command not found: make"},
		{UUID: "uuid-3", ToolName: "Read", Status: "error", Error: "file not found: /tmp/config.yml"},
		{UUID: "uuid-4", ToolName: "Bash", Status: "error", Error: "command not found: make"},
		{UUID: "uuid-5", ToolName: "Read", Status: "error", Error: "file not found: /tmp/config.yml"},
	}

	patterns := DetectErrorPatterns(entries, toolCalls)

	// 验证检测到的模式
	if len(patterns) != 1 {
		t.Fatalf("Expected 1 pattern (only Bash has 3+ occurrences), got %d", len(patterns))
	}

	// 验证第一个模式（应该是 Bash，3次出现）
	pattern := patterns[0]
	if pattern.ToolName != "Bash" {
		t.Errorf("Expected Bash pattern, got %s", pattern.ToolName)
	}
	if pattern.Occurrences != 3 {
		t.Errorf("Expected 3 occurrences, got %d", pattern.Occurrences)
	}
	if pattern.ErrorText != "command not found: make" {
		t.Errorf("Expected 'command not found: make', got '%s'", pattern.ErrorText)
	}

	// 验证时间信息
	if pattern.FirstSeen != "2025-10-02T10:00:00.000Z" {
		t.Errorf("Expected first seen at 10:00:00, got %s", pattern.FirstSeen)
	}
	if pattern.LastSeen != "2025-10-02T10:07:00.000Z" {
		t.Errorf("Expected last seen at 10:07:00, got %s", pattern.LastSeen)
	}
	expectedSpan := 420 // 7 minutes = 420 seconds
	if pattern.TimeSpanSeconds != expectedSpan {
		t.Errorf("Expected time span %d seconds, got %d", expectedSpan, pattern.TimeSpanSeconds)
	}

	// 验证上下文
	if len(pattern.Context.TurnUUIDs) != 3 {
		t.Errorf("Expected 3 turn UUIDs, got %d", len(pattern.Context.TurnUUIDs))
	}
	if len(pattern.Context.TurnIndices) != 3 {
		t.Errorf("Expected 3 turn indices, got %d", len(pattern.Context.TurnIndices))
	}

	// 输出示例（仅用于演示，测试时不会显示除非加 -v）
	t.Logf("Detected Pattern:")
	t.Logf("  Tool: %s", pattern.ToolName)
	t.Logf("  Error: %s", pattern.ErrorText)
	t.Logf("  Occurrences: %d", pattern.Occurrences)
	t.Logf("  First Seen: %s", pattern.FirstSeen)
	t.Logf("  Last Seen: %s", pattern.LastSeen)
	t.Logf("  Time Span: %d seconds (%.1f minutes)", pattern.TimeSpanSeconds, float64(pattern.TimeSpanSeconds)/60)
	t.Logf("  Turn UUIDs: %v", pattern.Context.TurnUUIDs)
	t.Logf("  Turn Indices: %v", pattern.Context.TurnIndices)
	t.Logf("  Signature: %s", pattern.Signature)
}

// 辅助函数
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
