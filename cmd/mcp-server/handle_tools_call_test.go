package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestHandleToolsCall_ValidRequest tests request structure validation
func TestHandleToolsCall_ValidRequest(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "get_session_stats",
			"arguments": map[string]interface{}{
				"scope": "session",
			},
		},
	}

	var buf bytes.Buffer
	origStdout := outputWriter
	outputWriter = &buf
	defer func() { outputWriter = origStdout }()

	handleToolsCall(context.Background(), req)

	var resp JSONRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc=2.0, got %s", resp.JSONRPC)
	}

	// Response should be valid JSON-RPC (may have error if meta-cc not available)
	// ID comes back as float64 from JSON unmarshaling
	if idFloat, ok := resp.ID.(float64); !ok || idFloat != 1.0 {
		t.Errorf("expected ID=1.0, got %v (%T)", resp.ID, resp.ID)
	}
}

// TestHandleToolsCall_MissingName tests error handling for missing tool name (alternate)
func TestHandleToolsCall_MissingName(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"arguments": map[string]interface{}{},
		},
	}

	var buf bytes.Buffer
	origStdout := outputWriter
	outputWriter = &buf
	defer func() { outputWriter = origStdout }()

	handleToolsCall(context.Background(), req)

	var resp JSONRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error == nil {
		t.Error("expected error for missing tool name")
	}

	if resp.Error.Code != -32602 {
		t.Errorf("expected error code -32602, got %d", resp.Error.Code)
	}

	if !strings.Contains(resp.Error.Message, "missing tool name") {
		t.Errorf("expected error message about missing tool name, got %s", resp.Error.Message)
	}
}

// TestHandleToolsCall_NonExistentTool tests error handling for non-existent tool
func TestHandleToolsCall_NonExistentTool(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "nonexistent_tool_xyz",
			"arguments": map[string]interface{}{
				"scope": "session",
			},
		},
	}

	var buf bytes.Buffer
	origStdout := outputWriter
	outputWriter = &buf
	defer func() { outputWriter = origStdout }()

	handleToolsCall(context.Background(), req)

	var resp JSONRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error == nil {
		t.Error("expected error for invalid tool name")
	}

	// Should get an error code (either -32000 or -32603 depending on executor state)
	if resp.Error.Code >= 0 {
		t.Errorf("expected negative error code, got %d", resp.Error.Code)
	}

	// Error message should mention the tool name
	if !strings.Contains(resp.Error.Message, "nonexistent_tool_xyz") && !strings.Contains(resp.Error.Message, "unknown tool") {
		t.Errorf("expected error message to mention unknown tool, got: %s", resp.Error.Message)
	}
}

// TestHandleToolsCall_ArgumentDefaults tests that missing arguments are handled
func TestHandleToolsCall_ArgumentDefaults(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "get_session_stats",
			// No arguments map - should use defaults
		},
	}

	var buf bytes.Buffer
	origStdout := outputWriter
	outputWriter = &buf
	defer func() { outputWriter = origStdout }()

	handleToolsCall(context.Background(), req)

	var resp JSONRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Should get a response (may be error or success depending on meta-cc availability)
	if resp.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc=2.0, got %s", resp.JSONRPC)
	}

	// ID comes back as float64 from JSON unmarshaling
	if idFloat, ok := resp.ID.(float64); !ok || idFloat != 4.0 {
		t.Errorf("expected ID=4.0, got %v (%T)", resp.ID, resp.ID)
	}
}

// TestHandleToolsCall_ResponseTiming tests that execution completes quickly
func TestHandleToolsCall_ResponseTiming(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "get_session_stats",
			"arguments": map[string]interface{}{
				"scope": "session",
			},
		},
	}

	var buf bytes.Buffer
	origStdout := outputWriter
	outputWriter = &buf
	defer func() { outputWriter = origStdout }()

	start := time.Now()
	handleToolsCall(context.Background(), req)
	elapsed := time.Since(start)

	// Execution should complete reasonably fast (<5 seconds even with command execution)
	if elapsed > 5*time.Second {
		t.Errorf("execution took too long: %v", elapsed)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Should get a valid response (error or success)
	if resp.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc=2.0, got %s", resp.JSONRPC)
	}
}
