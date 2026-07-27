package types_test

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/types"
)

func TestExtractToolCalls_PairsToolUseWithResult(t *testing.T) {
	entries := []types.SessionEntry{
		{
			Type:      "assistant",
			UUID:      "entry-1",
			Timestamp: "2026-01-01T00:00:00Z",
			Message: &types.Message{
				Role: "assistant",
				Content: []types.ContentBlock{
					{
						Type: "tool_use",
						ToolUse: &types.ToolUse{
							ID:    "tu-abc",
							Name:  "Bash",
							Input: map[string]interface{}{"command": "ls"},
						},
					},
				},
			},
		},
		{
			Type: "user",
			UUID: "entry-2",
			Message: &types.Message{
				Role: "user",
				Content: []types.ContentBlock{
					{
						Type: "tool_result",
						ToolResult: &types.ToolResult{
							ToolUseID: "tu-abc",
							Content:   "file.txt",
							Status:    "success",
						},
					},
				},
			},
		},
	}

	calls := types.ExtractToolCalls(entries)
	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}
	tc := calls[0]
	if tc.ToolName != "Bash" {
		t.Errorf("expected ToolName='Bash', got %q", tc.ToolName)
	}
	if tc.Output != "file.txt" {
		t.Errorf("expected Output='file.txt', got %q", tc.Output)
	}
	if tc.Status != "success" {
		t.Errorf("expected Status='success', got %q", tc.Status)
	}
}

func TestExtractToolCalls_NoResult(t *testing.T) {
	entries := []types.SessionEntry{
		{
			Type: "assistant",
			UUID: "e1",
			Message: &types.Message{
				Content: []types.ContentBlock{
					{
						Type: "tool_use",
						ToolUse: &types.ToolUse{
							ID:   "tu-xyz",
							Name: "Read",
						},
					},
				},
			},
		},
	}

	calls := types.ExtractToolCalls(entries)
	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}
	if calls[0].Output != "" || calls[0].Status != "" {
		t.Errorf("expected empty output/status for unmatched tool call")
	}
}

func TestExtractToolCalls_EmptyEntries(t *testing.T) {
	calls := types.ExtractToolCalls(nil)
	if len(calls) != 0 {
		t.Errorf("expected 0 tool calls for nil input")
	}
}

func TestExtractToolCalls_NormalizesStatusFromIsError_False(t *testing.T) {
	// Real Claude Code JSONL: is_error: false, no "status" field → normalized to "success"
	entries := []types.SessionEntry{
		{
			Type:      "assistant",
			UUID:      "entry-1",
			Timestamp: "2026-01-01T00:00:00Z",
			Message: &types.Message{
				Role: "assistant",
				Content: []types.ContentBlock{
					{
						Type: "tool_use",
						ToolUse: &types.ToolUse{
							ID:   "tu-001",
							Name: "Read",
							Input: map[string]interface{}{
								"file_path": "/tmp/test.go",
							},
						},
					},
				},
			},
		},
		{
			Type: "user",
			UUID: "entry-2",
			Message: &types.Message{
				Role: "user",
				Content: []types.ContentBlock{
					{
						Type: "tool_result",
						ToolResult: &types.ToolResult{
							ToolUseID: "tu-001",
							Content:   "package main\n",
							IsError:   false,
							// Status is "" — mimics real JSONL (no "status" key)
						},
					},
				},
			},
		},
	}

	calls := types.ExtractToolCalls(entries)
	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}
	tc := calls[0]
	if tc.Status != "success" {
		t.Errorf("expected Status='success' (normalized from IsError=false), got %q", tc.Status)
	}
	if tc.Error != "" {
		t.Errorf("expected Error='', got %q", tc.Error)
	}
}

func TestExtractToolCalls_NormalizesStatusFromIsError_True(t *testing.T) {
	// Real Claude Code JSONL: is_error: true, no "status" field → normalized to "error"
	entries := []types.SessionEntry{
		{
			Type:      "assistant",
			UUID:      "entry-1",
			Timestamp: "2026-01-01T00:00:00Z",
			Message: &types.Message{
				Role: "assistant",
				Content: []types.ContentBlock{
					{
						Type: "tool_use",
						ToolUse: &types.ToolUse{
							ID:   "tu-002",
							Name: "Bash",
							Input: map[string]interface{}{
								"command": "exit 1",
							},
						},
					},
				},
			},
		},
		{
			Type: "user",
			UUID: "entry-2",
			Message: &types.Message{
				Role: "user",
				Content: []types.ContentBlock{
					{
						Type: "tool_result",
						ToolResult: &types.ToolResult{
							ToolUseID: "tu-002",
							Content:   "command failed",
							IsError:   true,
							// Status is "" — mimics real JSONL (no "status" key)
						},
					},
				},
			},
		},
	}

	calls := types.ExtractToolCalls(entries)
	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}
	tc := calls[0]
	if tc.Status != "error" {
		t.Errorf("expected Status='error' (normalized from IsError=true), got %q", tc.Status)
	}
}

func TestExtractToolCalls_ExplicitStatusNotOverridden(t *testing.T) {
	// When JSONL does include a "status" field, it should NOT be overridden by normalization.
	entries := []types.SessionEntry{
		{
			Type:      "assistant",
			UUID:      "entry-1",
			Timestamp: "2026-01-01T00:00:00Z",
			Message: &types.Message{
				Role: "assistant",
				Content: []types.ContentBlock{
					{
						Type: "tool_use",
						ToolUse: &types.ToolUse{
							ID:    "tu-003",
							Name:  "Edit",
							Input: map[string]interface{}{},
						},
					},
				},
			},
		},
		{
			Type: "user",
			UUID: "entry-2",
			Message: &types.Message{
				Role: "user",
				Content: []types.ContentBlock{
					{
						Type: "tool_result",
						ToolResult: &types.ToolResult{
							ToolUseID: "tu-003",
							Content:   "edited",
							IsError:   false,
							Status:    "custom-status",
						},
					},
				},
			},
		},
	}

	calls := types.ExtractToolCalls(entries)
	if len(calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(calls))
	}
	tc := calls[0]
	if tc.Status != "custom-status" {
		t.Errorf("expected Status='custom-status' (explicit, should not be overridden), got %q", tc.Status)
	}
}
