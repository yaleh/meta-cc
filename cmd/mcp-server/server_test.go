package main

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func TestHandleInitialize(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  map[string]interface{}{},
	}

	var buf bytes.Buffer
	origStdout := outputWriter
	outputWriter = &buf
	defer func() { outputWriter = origStdout }()

	handleInitialize(context.Background(), req)

	var resp JSONRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc=2.0, got %s", resp.JSONRPC)
	}

	// ID is float64 when unmarshaled from JSON
	if id, ok := resp.ID.(float64); !ok || int(id) != 1 {
		t.Errorf("expected id=1, got %v (type %T)", resp.ID, resp.ID)
	}

	if resp.Error != nil {
		t.Errorf("expected no error, got %v", resp.Error)
	}

	// Check that result contains protocolVersion
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("expected result to be a map")
	}

	if _, hasVersion := result["protocolVersion"]; !hasVersion {
		t.Error("expected protocolVersion in result")
	}
}

func TestHandleToolsList(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
		Params:  map[string]interface{}{},
	}

	var buf bytes.Buffer
	origStdout := outputWriter
	outputWriter = &buf
	defer func() { outputWriter = origStdout }()

	handleToolsList(context.Background(), req)

	var resp JSONRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc=2.0, got %s", resp.JSONRPC)
	}

	if resp.Error != nil {
		t.Errorf("expected no error, got %v", resp.Error)
	}

	// Check that result contains tools array
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("expected result to be a map")
	}

	toolsInterface, ok := result["tools"]
	if !ok {
		t.Fatal("expected tools field in result")
	}

	// tools is []interface{} when unmarshaled from JSON
	toolsSlice, ok := toolsInterface.([]interface{})
	if !ok {
		t.Fatalf("expected tools to be a slice, got %T", toolsInterface)
	}

	// TASK-7: Consolidated 10 old query_* tools into 3 new tools
	// Total: 15 tools
	// - 3 consolidated query tools (query_session_content, query_session_signals, query_file_activity)
	// - 1 utility tool (cleanup_temp_files)
	// - 4 two-stage query tools (get_session_directory, inspect_session_files, execute_stage2_query, get_session_metadata)
	// - 6 analysis tools (analyze_errors, quality_scan, get_work_patterns, get_timeline, analyze_bugs, get_tech_debt)
	// - 1 doc session signals tool (query_edit_sequences)
	if len(toolsSlice) != 15 {
		t.Errorf("expected 15 tools after TASK-7, got %d", len(toolsSlice))
	}
}

func TestHandleRequest_UnknownMethod(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "unknown/method",
		Params:  map[string]interface{}{},
	}

	var buf bytes.Buffer
	origStdout := outputWriter
	outputWriter = &buf
	defer func() { outputWriter = origStdout }()

	handleRequest(req)

	var resp JSONRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error == nil {
		t.Error("expected error for unknown method")
	}

	if resp.Error.Code != -32601 {
		t.Errorf("expected error code -32601, got %d", resp.Error.Code)
	}
}

func TestWriteResponse(t *testing.T) {
	var buf bytes.Buffer
	origStdout := outputWriter
	outputWriter = &buf
	defer func() { outputWriter = origStdout }()

	result := map[string]interface{}{
		"test": "value",
	}

	writeResponse(123, result)

	var resp JSONRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc=2.0, got %s", resp.JSONRPC)
	}

	// ID is float64 when unmarshaled from JSON
	if id, ok := resp.ID.(float64); !ok || int(id) != 123 {
		t.Errorf("expected id=123, got %v (type %T)", resp.ID, resp.ID)
	}

	if resp.Error != nil {
		t.Errorf("expected no error, got %v", resp.Error)
	}

	resultMap, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("expected result to be a map")
	}

	if resultMap["test"] != "value" {
		t.Errorf("expected test=value, got %v", resultMap["test"])
	}
}

func TestWriteError(t *testing.T) {
	var buf bytes.Buffer
	origStdout := outputWriter
	outputWriter = &buf
	defer func() { outputWriter = origStdout }()

	writeError(456, -32600, "Invalid Request")

	var resp JSONRPCResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc=2.0, got %s", resp.JSONRPC)
	}

	// ID is float64 when unmarshaled from JSON
	if id, ok := resp.ID.(float64); !ok || int(id) != 456 {
		t.Errorf("expected id=456, got %v (type %T)", resp.ID, resp.ID)
	}

	if resp.Error == nil {
		t.Fatal("expected error to be present")
	}

	if resp.Error.Code != -32600 {
		t.Errorf("expected error code -32600, got %d", resp.Error.Code)
	}

	if resp.Error.Message != "Invalid Request" {
		t.Errorf("expected message='Invalid Request', got %s", resp.Error.Message)
	}
}

func TestHandleToolsCall_MissingToolName(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			// Missing "name" field
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
}

func TestHandleRequest_AllMethods(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		expectErr bool
	}{
		{
			name:      "initialize",
			method:    "initialize",
			expectErr: false,
		},
		{
			name:      "tools/list",
			method:    "tools/list",
			expectErr: false,
		},
		{
			name:      "unknown method",
			method:    "unknown/method",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  tt.method,
				Params:  map[string]interface{}{},
			}

			var buf bytes.Buffer
			origStdout := outputWriter
			outputWriter = &buf
			defer func() { outputWriter = origStdout }()

			handleRequest(req)

			var resp JSONRPCResponse
			if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			if tt.expectErr {
				if resp.Error == nil {
					t.Error("expected error")
				}
			} else {
				if resp.Error != nil {
					t.Errorf("expected no error, got %v", resp.Error)
				}
			}
		})
	}
}

func TestQuerySessionContentHasOutputWarning(t *testing.T) {
	tools := getToolDefinitions()

	var querySessionContentTool *Tool
	for _, tool := range tools {
		if tool.Name == "query_session_content" {
			querySessionContentTool = &tool
			break
		}
	}

	if querySessionContentTool == nil {
		t.Fatal("query_session_content tool not found")
	}

	// Verify max_message_length parameter exists
	props := querySessionContentTool.InputSchema.Properties
	if _, hasParam := props["max_message_length"]; !hasParam {
		t.Error("query_session_content should have max_message_length parameter")
	}
}
