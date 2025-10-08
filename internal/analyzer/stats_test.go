package analyzer

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/parser"
)

func TestCalculateStats_BasicCounts(t *testing.T) {
	// 准备测试数据
	entries := []parser.SessionEntry{
		{
			Type:      "user",
			UUID:      "uuid-1",
			Timestamp: "2025-10-02T10:00:00.000Z",
		},
		{
			Type:      "assistant",
			UUID:      "uuid-2",
			Timestamp: "2025-10-02T10:01:00.000Z",
		},
		{
			Type:      "user",
			UUID:      "uuid-3",
			Timestamp: "2025-10-02T10:02:00.000Z",
		},
	}

	toolCalls := []parser.ToolCall{
		{UUID: "uuid-t1", ToolName: "Grep", Status: "success"},
		{UUID: "uuid-t2", ToolName: "Read", Status: "success"},
		{UUID: "uuid-t3", ToolName: "Grep", Status: "error", Error: "pattern error"},
	}

	stats := CalculateStats(entries, toolCalls)

	// 验证基础计数
	if stats.TurnCount != 3 {
		t.Errorf("Expected TurnCount 3, got %d", stats.TurnCount)
	}

	if stats.UserTurnCount != 2 {
		t.Errorf("Expected UserTurnCount 2, got %d", stats.UserTurnCount)
	}

	if stats.AssistantTurnCount != 1 {
		t.Errorf("Expected AssistantTurnCount 1, got %d", stats.AssistantTurnCount)
	}

	if stats.ToolCallCount != 3 {
		t.Errorf("Expected ToolCallCount 3, got %d", stats.ToolCallCount)
	}

	if stats.ErrorCount != 1 {
		t.Errorf("Expected ErrorCount 1, got %d", stats.ErrorCount)
	}
}

func TestCalculateStats_Duration(t *testing.T) {
	entries := []parser.SessionEntry{
		{
			Type:      "user",
			UUID:      "uuid-1",
			Timestamp: "2025-10-02T10:00:00.000Z",
		},
		{
			Type:      "assistant",
			UUID:      "uuid-2",
			Timestamp: "2025-10-02T10:05:30.000Z",
		},
	}

	stats := CalculateStats(entries, []parser.ToolCall{})

	// 会话时长应为 5 分 30 秒 = 330 秒
	expectedDuration := int64(330)
	if stats.DurationSeconds != expectedDuration {
		t.Errorf("Expected DurationSeconds %d, got %d", expectedDuration, stats.DurationSeconds)
	}
}

func TestCalculateStats_ToolFrequency(t *testing.T) {
	toolCalls := []parser.ToolCall{
		{UUID: "uuid-1", ToolName: "Grep"},
		{UUID: "uuid-2", ToolName: "Read"},
		{UUID: "uuid-3", ToolName: "Grep"},
		{UUID: "uuid-4", ToolName: "Grep"},
		{UUID: "uuid-5", ToolName: "Bash"},
	}

	stats := CalculateStats([]parser.SessionEntry{}, toolCalls)

	// 验证工具频率统计
	if stats.ToolFrequency["Grep"] != 3 {
		t.Errorf("Expected Grep frequency 3, got %d", stats.ToolFrequency["Grep"])
	}

	if stats.ToolFrequency["Read"] != 1 {
		t.Errorf("Expected Read frequency 1, got %d", stats.ToolFrequency["Read"])
	}

	if stats.ToolFrequency["Bash"] != 1 {
		t.Errorf("Expected Bash frequency 1, got %d", stats.ToolFrequency["Bash"])
	}
}

func TestCalculateStats_ErrorRate(t *testing.T) {
	toolCalls := []parser.ToolCall{
		{UUID: "uuid-1", ToolName: "Grep", Status: "success"},
		{UUID: "uuid-2", ToolName: "Read", Status: "error", Error: "file not found"},
		{UUID: "uuid-3", ToolName: "Bash", Status: "error", Error: "command failed"},
		{UUID: "uuid-4", ToolName: "Grep", Status: "success"},
	}

	stats := CalculateStats([]parser.SessionEntry{}, toolCalls)

	// 错误率应为 2/4 = 50%
	expectedErrorRate := 50.0
	if stats.ErrorRate != expectedErrorRate {
		t.Errorf("Expected ErrorRate %.1f%%, got %.1f%%", expectedErrorRate, stats.ErrorRate)
	}
}

func TestCalculateStats_EmptyData(t *testing.T) {
	stats := CalculateStats([]parser.SessionEntry{}, []parser.ToolCall{})

	if stats.TurnCount != 0 {
		t.Errorf("Expected TurnCount 0 for empty data, got %d", stats.TurnCount)
	}

	if stats.ToolCallCount != 0 {
		t.Errorf("Expected ToolCallCount 0 for empty data, got %d", stats.ToolCallCount)
	}

	if stats.ErrorRate != 0.0 {
		t.Errorf("Expected ErrorRate 0.0 for empty data, got %.1f", stats.ErrorRate)
	}
}

func TestCalculateStats_SingleEntry(t *testing.T) {
	entries := []parser.SessionEntry{
		{
			Type:      "user",
			UUID:      "uuid-1",
			Timestamp: "2025-10-02T10:00:00.000Z",
		},
	}

	stats := CalculateStats(entries, []parser.ToolCall{})

	// 单个条目，时长应为 0
	if stats.DurationSeconds != 0 {
		t.Errorf("Expected DurationSeconds 0 for single entry, got %d", stats.DurationSeconds)
	}
}

func TestCalculateStats_TopTools(t *testing.T) {
	toolCalls := []parser.ToolCall{
		{UUID: "uuid-1", ToolName: "Grep"},
		{UUID: "uuid-2", ToolName: "Read"},
		{UUID: "uuid-3", ToolName: "Grep"},
		{UUID: "uuid-4", ToolName: "Bash"},
		{UUID: "uuid-5", ToolName: "Grep"},
		{UUID: "uuid-6", ToolName: "Write"},
	}

	stats := CalculateStats([]parser.SessionEntry{}, toolCalls)

	// 验证 TopTools（前 3 名）
	if len(stats.TopTools) < 3 {
		t.Fatalf("Expected at least 3 TopTools, got %d", len(stats.TopTools))
	}

	// 验证第一名
	if stats.TopTools[0].Name != "Grep" || stats.TopTools[0].Count != 3 {
		t.Errorf("Expected top tool 'Grep' with count 3, got '%s' with count %d",
			stats.TopTools[0].Name, stats.TopTools[0].Count)
	}
}
