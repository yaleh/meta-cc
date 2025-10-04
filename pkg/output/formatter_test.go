package output

import (
	"strings"
	"testing"

	"github.com/yale/meta-cc/internal/parser"
)

func TestFormatJSON_SessionEntries(t *testing.T) {
	entries := []parser.SessionEntry{
		{
			Type:      "user",
			UUID:      "uuid-1",
			Timestamp: "2025-10-02T06:07:13.673Z",
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{Type: "text", Text: "Hello"},
				},
			},
		},
	}

	output, err := FormatJSON(entries)
	if err != nil {
		t.Fatalf("FormatJSON failed: %v", err)
	}

	if !strings.Contains(output, `"type": "user"`) && !strings.Contains(output, `"type":"user"`) {
		t.Errorf("Expected JSON with user type, got: %s", output)
	}

	if !strings.Contains(output, `"uuid": "uuid-1"`) && !strings.Contains(output, `"uuid":"uuid-1"`) {
		t.Errorf("Expected JSON with uuid, got: %s", output)
	}
}

func TestFormatJSON_ToolCalls(t *testing.T) {
	toolCalls := []parser.ToolCall{
		{
			UUID:     "uuid-1",
			ToolName: "Grep",
			Input: map[string]interface{}{
				"pattern": "error",
			},
			Output: "match found",
			Status: "success",
		},
	}

	output, err := FormatJSON(toolCalls)
	if err != nil {
		t.Fatalf("FormatJSON failed: %v", err)
	}

	if !strings.Contains(output, `"ToolName": "Grep"`) && !strings.Contains(output, `"ToolName":"Grep"`) {
		t.Errorf("Expected JSON with Grep tool, got: %s", output)
	}
}

func TestFormatMarkdown_ToolCalls(t *testing.T) {
	toolCalls := []parser.ToolCall{
		{
			UUID:     "uuid-1",
			ToolName: "Grep",
			Input: map[string]interface{}{
				"pattern": "error",
			},
			Output: "match found",
			Status: "success",
		},
		{
			UUID:     "uuid-2",
			ToolName: "Read",
			Input: map[string]interface{}{
				"file": "test.txt",
			},
			Output: "",
			Status: "error",
			Error:  "file not found",
		},
	}

	output, err := FormatMarkdown(toolCalls)
	if err != nil {
		t.Fatalf("FormatMarkdown failed: %v", err)
	}

	// Verify Markdown table structure
	if !strings.Contains(output, "| Tool | Input | Output | Status |") {
		t.Errorf("Expected Markdown table header, got: %s", output)
	}

	if !strings.Contains(output, "| Grep |") {
		t.Errorf("Expected Grep row, got: %s", output)
	}

	if !strings.Contains(output, "| Read |") {
		t.Errorf("Expected Read row, got: %s", output)
	}

	if !strings.Contains(output, "error") {
		t.Errorf("Expected error status in table, got: %s", output)
	}
}

func TestFormatMarkdown_SessionEntries(t *testing.T) {
	entries := []parser.SessionEntry{
		{
			Type:      "user",
			UUID:      "uuid-1",
			Timestamp: "2025-10-02T06:07:13.673Z",
			Message: &parser.Message{
				Role: "user",
				Content: []parser.ContentBlock{
					{Type: "text", Text: "Hello"},
				},
			},
		},
	}

	output, err := FormatMarkdown(entries)
	if err != nil {
		t.Fatalf("FormatMarkdown failed: %v", err)
	}

	// Verify Markdown structure
	if !strings.Contains(output, "## Turn") {
		t.Errorf("Expected turn header, got: %s", output)
	}

	if !strings.Contains(output, "**Role**: user") {
		t.Errorf("Expected role field, got: %s", output)
	}
}

func TestFormatMarkdown_UserMessages(t *testing.T) {
	// Import UserMessage type from cmd package (defined in cmd/query_messages.go:44)
	type UserMessage struct {
		TurnSequence int    `json:"turn_sequence"`
		UUID         string `json:"uuid"`
		Timestamp    string `json:"timestamp"`
		Content      string `json:"content"`
	}

	userMessages := []UserMessage{
		{
			TurnSequence: 1,
			UUID:         "uuid-1",
			Timestamp:    "2025-10-04T03:41:42.444Z",
			Content:      "List all the user input history of this project.",
		},
		{
			TurnSequence: 37,
			UUID:         "uuid-2",
			Timestamp:    "2025-10-04T03:43:35.552Z",
			Content:      "Find out why it failed.",
		},
	}

	output, err := FormatMarkdown(userMessages)
	if err != nil {
		t.Fatalf("FormatMarkdown failed: %v", err)
	}

	// Verify Markdown table structure
	if !strings.Contains(output, "| Turn | UUID | Timestamp | Content |") {
		t.Errorf("Expected Markdown table header, got: %s", output)
	}

	// Verify first message row
	if !strings.Contains(output, "| 1 |") {
		t.Errorf("Expected turn 1 row, got: %s", output)
	}

	if !strings.Contains(output, "uuid-1") {
		t.Errorf("Expected uuid-1 in output, got: %s", output)
	}

	if !strings.Contains(output, "List all the user input history") {
		t.Errorf("Expected first message content, got: %s", output)
	}

	// Verify second message row
	if !strings.Contains(output, "| 37 |") {
		t.Errorf("Expected turn 37 row, got: %s", output)
	}

	if !strings.Contains(output, "Find out why it failed") {
		t.Errorf("Expected second message content, got: %s", output)
	}
}

func TestFormatCSV_ToolCalls(t *testing.T) {
	toolCalls := []parser.ToolCall{
		{
			UUID:     "uuid-1",
			ToolName: "Grep",
			Input: map[string]interface{}{
				"pattern": "error",
			},
			Output: "match found",
			Status: "success",
		},
	}

	output, err := FormatCSV(toolCalls)
	if err != nil {
		t.Fatalf("FormatCSV failed: %v", err)
	}

	// Verify CSV structure
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 lines (header + data), got %d", len(lines))
	}

	// Verify header
	header := lines[0]
	if !strings.Contains(header, "UUID,Tool,Input,Output,Status") {
		t.Errorf("Expected CSV header, got: %s", header)
	}

	// Verify data row
	dataRow := lines[1]
	if !strings.Contains(dataRow, "uuid-1") || !strings.Contains(dataRow, "Grep") {
		t.Errorf("Expected data row with UUID and tool name, got: %s", dataRow)
	}
}
