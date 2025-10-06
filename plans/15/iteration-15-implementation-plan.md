# Phase 15: MCP Output Control & Tools Standardization - Implementation Plan

## Overview

**Goal**: Implement MCP output size control and standardize tool parameters to prevent context overflow

**Timeline**: 2-3 days (TDD methodology)

**Total Effort**: ~350 lines (Output control ~150 + Parameter standardization ~200)

**Priority**: High (resolves MCP context overflow issues, completes Phase 14 MCP enhancements)

**Status**: Ready for implementation

---

## Objectives

### Primary Goals

1. **Message Content Truncation**: Prevent user messages with large content (e.g., session summaries ~10.7k tokens) from overwhelming MCP responses
2. **Standardize Tool Parameters**: Ensure all MCP tools support unified parameters (jq_filter, stats_only, stats_first, max_output_bytes, max_message_length)
3. **Simplify Tool Descriptions**: Reduce all tool descriptions to ≤100 characters for better MCP client readability
4. **Remove Redundant Tools**: Eliminate aggregate_stats and analyze_errors (deprecated in Phase 14)

### Design Principles

- ✅ **Early Truncation**: Apply max_message_length before jq filtering to reduce processing overhead
- ✅ **Content Summary Mode**: Option to return only message metadata (turn/timestamp/preview) without full content
- ✅ **Parameter Consistency**: All tools expose the same control parameters
- ✅ **Clear Warning Messages**: Tool descriptions warn about potential large outputs
- ✅ **Backward Compatibility**: Default parameters maintain existing behavior

---

## Problem Statement

### Root Cause Analysis

**Problem**: MCP query `query_user_messages` returns ~10.7k tokens (exceeds Claude context budget)

**Why it happens**:
1. User messages can contain session summaries (thousands of lines of conversation history)
2. `jq_filter ".[]"` returns full objects including massive `content` fields
3. `max_output_bytes` only truncates at the end (too late to prevent processing overhead)
4. MCP returns entire 10.7k tokens, filling up available context

**Example scenario**:
```bash
# User message contains session summary (~8k lines)
meta-cc query user-messages --match "session.*summary"

# jq_filter ".[]" returns full object:
{
  "uuid": "abc123...",
  "timestamp": "2025-10-06T12:00:00Z",
  "turn_sequence": 42,
  "content": "<8000 lines of conversation history...>"  # ← Problem!
}

# Result: ~10.7k tokens sent to Claude
```

### Solution Architecture

**Two-layer output control**:

1. **Message-level truncation** (`max_message_length`):
   - Truncate `content` field before jq filtering
   - Default: 500 characters per message
   - Applied in `TruncateMessageContent()` function

2. **Content summary mode** (`content_summary`):
   - Return only metadata: `{turn, timestamp, content_preview}`
   - No full content included
   - Ideal for pattern matching without content analysis

**Performance impact**:
```
Before: 10.7k tokens
After (max_message_length=500): ~1-2k tokens
Compression: 81-91%
```

---

## Stage Breakdown

### Stage 15.1: MCP Output Size Control

**Duration**: 1 day

**Objective**: Implement message content truncation to prevent context overflow

**Code Size**: ~150 lines (source ~80 + tests ~70)

#### Deliverables

1. **`cmd/mcp-server/filters.go`** (~80 lines)
   - `TruncateMessageContent(messages []interface{}, maxLen int) []interface{}`
   - `ApplyContentSummary(messages []interface{}) []interface{}`
   - JSON unmarshaling/marshaling helpers

2. **`cmd/mcp-server/executor.go`** (~50 lines update)
   - Extract `max_message_length` parameter (default: 500)
   - Extract `content_summary` parameter (default: false)
   - Call truncation functions after jq filter for user messages

3. **`cmd/mcp-server/executor_test.go`** (~70 lines)
   - Test message truncation with various lengths
   - Test content summary mode
   - Test edge cases (empty content, nil messages)

4. **Update tool descriptions** (all tools)
   - Add output size warning to descriptions

#### Implementation

**filters.go**:
```go
// cmd/mcp-server/filters.go
package main

import (
	"encoding/json"
	"strings"
)

const (
	DefaultMaxMessageLength = 500
)

// TruncateMessageContent truncates the 'content' field in user messages
// to prevent context overflow from large session summaries.
func TruncateMessageContent(messages []interface{}, maxLen int) []interface{} {
	if maxLen <= 0 {
		return messages
	}

	truncated := make([]interface{}, len(messages))

	for i, msg := range messages {
		// Convert to map for manipulation
		msgMap, ok := msg.(map[string]interface{})
		if !ok {
			truncated[i] = msg
			continue
		}

		// Create copy to avoid mutating original
		newMap := make(map[string]interface{})
		for k, v := range msgMap {
			newMap[k] = v
		}

		// Truncate content field if present
		if content, ok := newMap["content"].(string); ok {
			if len(content) > maxLen {
				newMap["content"] = content[:maxLen] + "... [TRUNCATED]"
				newMap["content_truncated"] = true
				newMap["original_length"] = len(content)
			}
		}

		truncated[i] = newMap
	}

	return truncated
}

// ApplyContentSummary returns only message metadata (no full content).
// Useful for pattern matching without needing full message text.
func ApplyContentSummary(messages []interface{}) []interface{} {
	summary := make([]interface{}, len(messages))

	for i, msg := range messages {
		msgMap, ok := msg.(map[string]interface{})
		if !ok {
			summary[i] = msg
			continue
		}

		// Extract preview (first 100 chars)
		preview := ""
		if content, ok := msgMap["content"].(string); ok {
			if len(content) > 100 {
				preview = content[:100] + "..."
			} else {
				preview = content
			}
		}

		// Create summary object
		summary[i] = map[string]interface{}{
			"turn_sequence":   msgMap["turn_sequence"],
			"timestamp":       msgMap["timestamp"],
			"content_preview": preview,
		}
	}

	return summary
}

// ConvertToMessages converts JSON string to message array
func ConvertToMessages(jsonStr string) ([]interface{}, error) {
	var messages []interface{}

	// Try array first
	if err := json.Unmarshal([]byte(jsonStr), &messages); err == nil {
		return messages, nil
	}

	// Try single object
	var singleMsg interface{}
	if err := json.Unmarshal([]byte(jsonStr), &singleMsg); err == nil {
		return []interface{}{singleMsg}, nil
	}

	// Try JSONL (newline-delimited)
	lines := strings.Split(strings.TrimSpace(jsonStr), "\n")
	messages = make([]interface{}, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}
		var msg interface{}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// ConvertFromMessages converts message array back to JSON string
func ConvertFromMessages(messages []interface{}, format string) (string, error) {
	if format == "jsonl" {
		// JSONL format: one object per line
		lines := make([]string, len(messages))
		for i, msg := range messages {
			data, err := json.Marshal(msg)
			if err != nil {
				return "", err
			}
			lines[i] = string(data)
		}
		return strings.Join(lines, "\n"), nil
	}

	// JSON array format
	data, err := json.Marshal(messages)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
```

**executor.go updates**:
```go
// cmd/mcp-server/executor.go (update ExecuteTool)

func (e *ToolExecutor) ExecuteTool(toolName string, args map[string]interface{}) (string, error) {
	// ... existing deprecation checks ...

	// Extract common parameters
	jqFilter := getStringParam(args, "jq_filter", ".[]")
	statsOnly := getBoolParam(args, "stats_only", false)
	statsFirst := getBoolParam(args, "stats_first", false)
	maxOutputBytes := getIntParam(args, "max_output_bytes", DefaultMaxOutputBytes)
	maxMessageLength := getIntParam(args, "max_message_length", DefaultMaxMessageLength)
	contentSummary := getBoolParam(args, "content_summary", false)
	scope := getStringParam(args, "scope", "project")
	outputFormat := getStringParam(args, "output_format", "jsonl")

	// ... existing command building ...

	// Execute meta-cc
	rawOutput, err := e.executeMetaCC(cmdArgs)
	if err != nil {
		return "", err
	}

	// Apply message-specific filters for user message queries
	if toolName == "query_user_messages" {
		rawOutput, err = e.applyMessageFilters(rawOutput, maxMessageLength, contentSummary, outputFormat)
		if err != nil {
			return "", fmt.Errorf("message filter error: %w", err)
		}
	}

	// Apply jq filter (after message truncation)
	filtered, err := ApplyJQFilter(rawOutput, jqFilter)
	if err != nil {
		return "", fmt.Errorf("jq filter error: %w", err)
	}

	// ... existing stats generation ...

	// Apply output length limit
	if len(output) > maxOutputBytes {
		output = output[:maxOutputBytes]
		output += fmt.Sprintf("\n[OUTPUT TRUNCATED: exceeded %d bytes limit]", maxOutputBytes)
	}

	return output, nil
}

// applyMessageFilters applies content truncation or summary mode
func (e *ToolExecutor) applyMessageFilters(rawOutput string, maxLen int, summary bool, format string) (string, error) {
	// Convert to message array
	messages, err := ConvertToMessages(rawOutput)
	if err != nil {
		return rawOutput, nil // Fallback to original on parse error
	}

	// Apply appropriate filter
	var filtered []interface{}
	if summary {
		filtered = ApplyContentSummary(messages)
	} else if maxLen > 0 {
		filtered = TruncateMessageContent(messages, maxLen)
	} else {
		filtered = messages
	}

	// Convert back to string
	return ConvertFromMessages(filtered, format)
}
```

**Test Coverage** (`executor_test.go`):
```go
// cmd/mcp-server/executor_test.go

func TestTruncateMessageContent(t *testing.T) {
	tests := []struct {
		name    string
		messages []interface{}
		maxLen   int
		wantTruncated bool
	}{
		{
			name: "short content not truncated",
			messages: []interface{}{
				map[string]interface{}{
					"content": "short message",
					"turn_sequence": 1,
				},
			},
			maxLen: 100,
			wantTruncated: false,
		},
		{
			name: "long content truncated",
			messages: []interface{}{
				map[string]interface{}{
					"content": strings.Repeat("x", 1000),
					"turn_sequence": 1,
				},
			},
			maxLen: 100,
			wantTruncated: true,
		},
		{
			name: "max_len zero disables truncation",
			messages: []interface{}{
				map[string]interface{}{
					"content": strings.Repeat("x", 1000),
					"turn_sequence": 1,
				},
			},
			maxLen: 0,
			wantTruncated: false,
		},
		{
			name: "missing content field handled",
			messages: []interface{}{
				map[string]interface{}{
					"turn_sequence": 1,
				},
			},
			maxLen: 100,
			wantTruncated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateMessageContent(tt.messages, tt.maxLen)

			if len(result) != len(tt.messages) {
				t.Fatalf("expected %d messages, got %d", len(tt.messages), len(result))
			}

			resultMap := result[0].(map[string]interface{})
			if tt.wantTruncated {
				if truncated, ok := resultMap["content_truncated"].(bool); !ok || !truncated {
					t.Error("expected content_truncated=true")
				}
				if origLen, ok := resultMap["original_length"].(int); !ok || origLen != 1000 {
					t.Errorf("expected original_length=1000, got %v", origLen)
				}
				content := resultMap["content"].(string)
				if len(content) > tt.maxLen+20 { // +20 for "... [TRUNCATED]"
					t.Errorf("content not truncated properly: %d chars", len(content))
				}
			} else {
				if _, ok := resultMap["content_truncated"]; ok {
					t.Error("expected no truncation flag")
				}
			}
		})
	}
}

func TestApplyContentSummary(t *testing.T) {
	messages := []interface{}{
		map[string]interface{}{
			"turn_sequence": 1,
			"timestamp": "2025-10-06T12:00:00Z",
			"content": strings.Repeat("x", 500),
			"other_field": "should be removed",
		},
	}

	result := ApplyContentSummary(messages)

	if len(result) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result))
	}

	resultMap := result[0].(map[string]interface{})

	// Should only have these fields
	if _, ok := resultMap["turn_sequence"]; !ok {
		t.Error("missing turn_sequence")
	}
	if _, ok := resultMap["timestamp"]; !ok {
		t.Error("missing timestamp")
	}
	if preview, ok := resultMap["content_preview"].(string); !ok {
		t.Error("missing content_preview")
	} else if len(preview) > 105 { // 100 + "..."
		t.Errorf("preview too long: %d chars", len(preview))
	}

	// Should NOT have full content
	if _, ok := resultMap["content"]; ok {
		t.Error("full content should be removed in summary mode")
	}
	if _, ok := resultMap["other_field"]; ok {
		t.Error("other fields should be removed in summary mode")
	}
}

func TestConvertMessagesRoundTrip(t *testing.T) {
	original := []interface{}{
		map[string]interface{}{
			"turn_sequence": 1,
			"content": "test message",
		},
	}

	// Convert to JSONL string
	jsonlStr, err := ConvertFromMessages(original, "jsonl")
	if err != nil {
		t.Fatalf("ConvertFromMessages failed: %v", err)
	}

	// Convert back to messages
	messages, err := ConvertToMessages(jsonlStr)
	if err != nil {
		t.Fatalf("ConvertToMessages failed: %v", err)
	}

	// Verify round-trip
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	msgMap := messages[0].(map[string]interface{})
	if content := msgMap["content"].(string); content != "test message" {
		t.Errorf("content mismatch: %s", content)
	}
}
```

#### Acceptance Criteria

- ✅ `TruncateMessageContent()` correctly truncates messages longer than `max_message_length`
- ✅ Truncated messages include `content_truncated` and `original_length` fields
- ✅ `ApplyContentSummary()` returns only metadata (turn, timestamp, preview)
- ✅ Content summary preview is ≤103 characters (100 + "...")
- ✅ `max_message_length=0` disables truncation
- ✅ `content_summary=true` overrides `max_message_length`
- ✅ Unit tests pass with ≥90% coverage
- ✅ Integration test: `query_user_messages` output reduced from ~10.7k to ~1-2k tokens

#### Verification Commands

```bash
# Unit tests
go test ./cmd/mcp-server -run TestTruncate -v
go test ./cmd/mcp-server -run TestApplyContentSummary -v
go test ./cmd/mcp-server -run TestConvertMessages -v

# Coverage check
go test ./cmd/mcp-server -coverprofile=coverage.out
go tool cover -func=coverage.out | grep filters.go

# Integration test (requires meta-cc-mcp executable)
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"query_user_messages","arguments":{"pattern":".*","max_message_length":500}}}' | ./meta-cc-mcp

# Verify output size
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"query_user_messages","arguments":{"pattern":".*","content_summary":true}}}' | ./meta-cc-mcp | wc -c
# Expected: <5000 bytes (vs ~30000 without truncation)
```

---

### Stage 15.2: Standardize MCP Tool Parameters

**Duration**: 0.5 day

**Objective**: Ensure all MCP tools expose unified control parameters

**Code Size**: ~100 lines (tool definitions update)

#### Deliverables

1. **`cmd/mcp-server/tools.go`** (update existing ~200 lines)
   - Add `max_message_length` to StandardToolParameters()
   - Add `content_summary` to StandardToolParameters()
   - Update all tool definitions to include new parameters

2. **`cmd/mcp-server/tools_test.go`** (update)
   - Test that all tools have standard parameters
   - Verify parameter types are correct

#### Implementation

**tools.go updates**:
```go
// cmd/mcp-server/tools.go

func StandardToolParameters() map[string]Property {
	return map[string]Property{
		"scope": {
			Type:        "string",
			Description: "Query scope: 'project' (default) or 'session'",
		},
		"jq_filter": {
			Type:        "string",
			Description: "jq expression for filtering (default: '.[]')",
		},
		"stats_only": {
			Type:        "boolean",
			Description: "Return only statistics (default: false)",
		},
		"stats_first": {
			Type:        "boolean",
			Description: "Return stats first, then details (default: false)",
		},
		"max_output_bytes": {
			Type:        "number",
			Description: "Max output size in bytes (default: 51200)",
		},
		"max_message_length": {  // NEW
			Type:        "number",
			Description: "Max message content length in chars (default: 500, 0=unlimited)",
		},
		"content_summary": {  // NEW
			Type:        "boolean",
			Description: "Return only message metadata (default: false)",
		},
		"output_format": {
			Type:        "string",
			Description: "Output format: jsonl or tsv (default: jsonl)",
		},
	}
}

// Update tool descriptions to include output warnings
func getToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name: "query_user_messages",
			Description: "Search user messages. WARNING: Can return large output; use max_message_length or content_summary.",
			InputSchema: InputSchema{
				Type:     "object",
				Required: []string{"pattern"},
				Properties: MergeParameters(map[string]Property{
					"pattern": {
						Type:        "string",
						Description: "Regex pattern to match (required)",
					},
					"limit": {
						Type:        "number",
						Description: "Max results (default: 10)",
					},
				}),
			},
		},
		{
			Name: "query_tools",
			Description: "Query tool calls with filters. Default scope: project.",
			InputSchema: InputSchema{
				Type:     "object",
				Required: []string{},
				Properties: MergeParameters(map[string]Property{
					"tool": {
						Type:        "string",
						Description: "Filter by tool name",
					},
					"status": {
						Type:        "string",
						Description: "Filter by status (error/success)",
					},
					"limit": {
						Type:        "number",
						Description: "Max results (default: 20)",
					},
				}),
			},
		},
		{
			Name: "get_session_stats",
			Description: "Get session statistics. Default scope: session.",
			InputSchema: InputSchema{
				Type:       "object",
				Required:   []string{},
				Properties: StandardToolParameters(), // All standard params, no specific ones
			},
		},
		// ... other tools with similar updates ...
	}
}
```

**tools_test.go updates**:
```go
// cmd/mcp-server/tools_test.go

func TestAllToolsHaveStandardParameters(t *testing.T) {
	tools := getToolDefinitions()

	requiredParams := []string{
		"jq_filter", "stats_only", "stats_first",
		"max_output_bytes", "max_message_length", "content_summary",
	}

	for _, tool := range tools {
		// Skip deprecated tools
		if tool.Name == "analyze_errors" || tool.Name == "aggregate_stats" {
			continue
		}

		for _, param := range requiredParams {
			if _, ok := tool.InputSchema.Properties[param]; !ok {
				t.Errorf("tool %s missing parameter: %s", tool.Name, param)
			}
		}
	}
}

func TestStandardToolParameters_IncludesNewParams(t *testing.T) {
	params := StandardToolParameters()

	// Verify max_message_length exists
	if maxMsgLen, ok := params["max_message_length"]; !ok {
		t.Error("missing max_message_length parameter")
	} else {
		if maxMsgLen.Type != "number" {
			t.Errorf("max_message_length should be number, got %s", maxMsgLen.Type)
		}
		if !strings.Contains(maxMsgLen.Description, "500") {
			t.Error("max_message_length description should mention default 500")
		}
	}

	// Verify content_summary exists
	if contentSum, ok := params["content_summary"]; !ok {
		t.Error("missing content_summary parameter")
	} else {
		if contentSum.Type != "boolean" {
			t.Errorf("content_summary should be boolean, got %s", contentSum.Type)
		}
	}
}
```

#### Acceptance Criteria

- ✅ All 15 MCP tools include `max_message_length` parameter
- ✅ All 15 MCP tools include `content_summary` parameter
- ✅ `query_user_messages` description includes output size warning
- ✅ Unit tests verify all tools have standard parameters
- ✅ Parameter types correctly validated (number vs boolean vs string)

#### Verification Commands

```bash
# Run parameter tests
go test ./cmd/mcp-server -run TestAllToolsHaveStandard -v
go test ./cmd/mcp-server -run TestStandardToolParameters -v

# Verify tool list includes parameters
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./meta-cc-mcp | jq '.result.tools[] | select(.name=="query_user_messages") | .inputSchema.properties | keys'
# Expected: [..., "max_message_length", "content_summary", ...]

# Build and test
make all
```

---

### Stage 15.3: Simplify MCP Tool Descriptions

**Duration**: 0.5 day

**Objective**: Reduce all tool descriptions to ≤100 characters for clarity

**Code Size**: ~50 lines (tool description updates)

#### Deliverables

1. **`cmd/mcp-server/tools.go`** (update descriptions)
   - Shorten all descriptions to ≤100 chars
   - Follow format: `<action> <object> <scope>`
   - Move detailed usage to docs

2. **`cmd/mcp-server/tools_test.go`** (new tests)
   - Verify all descriptions ≤100 characters
   - Ensure no empty descriptions

#### Implementation

**Simplified descriptions**:
```go
// cmd/mcp-server/tools.go (excerpt)

func getToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name: "query_tools",
			Description: "Query tool calls. Scope: project. Use filters: tool, status, limit.", // 73 chars
			// ...
		},
		{
			Name: "extract_tools",
			Description: "Extract tool call history. Scope: project. Limit: 100 max.", // 67 chars
			// ...
		},
		{
			Name: "query_user_messages",
			Description: "Search user messages. Scope: project. WARNING: Use max_message_length.", // 79 chars
			// ...
		},
		{
			Name: "get_session_stats",
			Description: "Get session statistics. Scope: session.", // 44 chars
			// ...
		},
		{
			Name: "query_context",
			Description: "Query error context. Required: error_signature. Window: 3.", // 64 chars
			// ...
		},
		{
			Name: "query_tool_sequences",
			Description: "Query workflow patterns. Scope: project. Min occurrences: 3.", // 67 chars
			// ...
		},
		{
			Name: "query_file_access",
			Description: "Query file operation history. Required: file path.", // 56 chars
			// ...
		},
		{
			Name: "query_project_state",
			Description: "Query project state evolution. Scope: project.", // 51 chars
			// ...
		},
		{
			Name: "query_successful_prompts",
			Description: "Query successful prompt patterns. Quality: 0.8 min. Limit: 10.", // 69 chars
			// ...
		},
		{
			Name: "query_tools_advanced",
			Description: "Query tools with SQL filters. Required: where clause.", // 59 chars
			// ...
		},
		{
			Name: "query_time_series",
			Description: "Analyze metrics over time. Interval: hour/day/week.", // 57 chars
			// ...
		},
		{
			Name: "query_files",
			Description: "File-level operation stats. Top: 20. Sort by: total_ops.", // 63 chars
			// ...
		},
		// Deprecated tools (will be removed in Stage 15.4)
		{
			Name: "analyze_errors",
			Description: "[DEPRECATED] Use query_tools with status='error'.", // 56 chars
			// ...
		},
		{
			Name: "aggregate_stats",
			Description: "[DEPRECATED] Use query_tools with jq_filter and stats_only.", // 66 chars
			// ...
		},
	}
}
```

**Test validation**:
```go
// cmd/mcp-server/tools_test.go

func TestToolDescriptionLength(t *testing.T) {
	tools := getToolDefinitions()

	for _, tool := range tools {
		descLen := len(tool.Description)

		if descLen == 0 {
			t.Errorf("tool %s has empty description", tool.Name)
		}

		if descLen > 100 {
			t.Errorf("tool %s description too long: %d chars (max 100)\nDescription: %s",
				tool.Name, descLen, tool.Description)
		}

		// Ensure description starts with capital letter
		if !strings.Contains("ABCDEFGHIJKLMNOPQRSTUVWXYZ[", string(tool.Description[0])) {
			t.Errorf("tool %s description should start with capital letter: %s",
				tool.Name, tool.Description)
		}
	}
}

func TestDeprecatedToolsMarked(t *testing.T) {
	tools := getToolDefinitions()

	deprecatedTools := []string{"analyze_errors", "aggregate_stats"}

	for _, tool := range tools {
		for _, depName := range deprecatedTools {
			if tool.Name == depName {
				if !strings.Contains(tool.Description, "[DEPRECATED]") {
					t.Errorf("tool %s should be marked [DEPRECATED] in description", tool.Name)
				}
			}
		}
	}
}
```

#### Acceptance Criteria

- ✅ All tool descriptions ≤100 characters
- ✅ Descriptions follow format: `<action> <object> <scope/constraints>`
- ✅ Deprecated tools marked with `[DEPRECATED]` prefix
- ✅ No empty descriptions
- ✅ All descriptions start with capital letter
- ✅ Unit test validates description length constraints

#### Verification Commands

```bash
# Test description length
go test ./cmd/mcp-server -run TestToolDescriptionLength -v

# Test deprecated tool marking
go test ./cmd/mcp-server -run TestDeprecatedToolsMarked -v

# Verify tool list shows short descriptions
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./meta-cc-mcp | jq '.result.tools[] | {name, desc: .description | length}'
# All should be ≤100
```

---

### Stage 15.4: MCP Tool Documentation

**Duration**: 0.5 day

**Objective**: Create comprehensive MCP tools reference documentation

**Code Size**: ~200 lines documentation

#### Deliverables

1. **`docs/mcp-tools-reference.md`** (~200 lines)
   - Complete parameter reference for all tools
   - Usage scenarios for each tool
   - Examples with jq filtering
   - Output control strategies
   - Migration guide from deprecated tools

2. **Update `docs/plan.md`** (Phase 15 completion notes)
   - Document Phase 15 deliverables
   - Update MCP section with output control info

#### Documentation Structure

**docs/mcp-tools-reference.md**:
```markdown
# MCP Tools Reference

## Overview

This document provides comprehensive reference for all meta-cc MCP tools.

**Quick Links**:
- [Standard Parameters](#standard-parameters)
- [Output Control](#output-control-strategies)
- [Tool Catalog](#tool-catalog)
- [Migration Guide](#migration-guide)

---

## Standard Parameters

All MCP tools support these standard parameters:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `scope` | string | "project" | Query scope: "project" or "session" |
| `jq_filter` | string | ".[]" | jq expression for filtering results |
| `stats_only` | boolean | false | Return only statistics (no details) |
| `stats_first` | boolean | false | Return stats before details |
| `max_output_bytes` | number | 51200 | Max output size in bytes (50KB) |
| `max_message_length` | number | 500 | Max message content length (0=unlimited) |
| `content_summary` | boolean | false | Return only message metadata |
| `output_format` | string | "jsonl" | Output format: "jsonl" or "tsv" |

---

## Output Control Strategies

### Problem: Large User Messages

**Scenario**: `query_user_messages` returns ~10.7k tokens due to session summaries in message content.

**Solution 1: Truncate content** (recommended for most cases)
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": "implement.*feature",
    "max_message_length": 500
  }
}
```
Result: ~1-2k tokens (81-91% reduction)

**Solution 2: Content summary** (for metadata-only queries)
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": ".*",
    "content_summary": true
  }
}
```
Result: Only `{turn_sequence, timestamp, content_preview}` returned

**Solution 3: Combine jq filter + stats_only**
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": "error|fail",
    "jq_filter": "group_by(.turn_sequence) | length",
    "stats_only": true
  }
}
```
Result: Single number (count of messages matching pattern)

### When to Use Each Strategy

| Strategy | Use Case | Output Size | Content Access |
|----------|----------|-------------|----------------|
| `max_message_length` | Need message content but want to limit size | 1-5k tokens | Truncated (500 chars) |
| `content_summary` | Only need metadata (timestamps, turn numbers) | <1k tokens | Preview only (100 chars) |
| `stats_only` + `jq_filter` | Aggregation queries (counts, averages) | <100 tokens | No content |
| Default (no limits) | Small projects or specific queries | Variable | Full content |

---

## Tool Catalog

### query_tools

**Description**: Query tool calls with filters. Scope: project.

**Specific Parameters**:
- `tool` (string): Filter by tool name (e.g., "Bash", "Edit")
- `status` (string): Filter by status ("error" or "success")
- `limit` (number): Max results (default: 20)

**Usage Scenarios**:

1. **Find all Bash errors**:
```json
{
  "name": "query_tools",
  "arguments": {
    "tool": "Bash",
    "status": "error",
    "limit": 50
  }
}
```

2. **Count tool usage by type**:
```json
{
  "name": "query_tools",
  "arguments": {
    "jq_filter": "group_by(.ToolName) | map({tool: .[0].ToolName, count: length})",
    "stats_only": true
  }
}
```

3. **Get recent 10 tool calls**:
```json
{
  "name": "query_tools",
  "arguments": {
    "limit": 10,
    "jq_filter": ".[] | {tool: .ToolName, status: .Status, time: .Timestamp}"
  }
}
```

---

### query_user_messages

**Description**: Search user messages. WARNING: Can return large output.

**Specific Parameters**:
- `pattern` (string, required): Regex pattern to match
- `limit` (number): Max results (default: 10)

**⚠️ Output Control**: This tool can return very large outputs due to session summaries in messages. **Always use** `max_message_length` or `content_summary`.

**Usage Scenarios**:

1. **Search for feature requests** (with truncation):
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": "implement|add|create",
    "max_message_length": 300,
    "limit": 20
  }
}
```

2. **Get message timeline** (metadata only):
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": ".*",
    "content_summary": true
  }
}
```

3. **Count error-related messages**:
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": "error|fail|bug",
    "jq_filter": "length",
    "stats_only": true
  }
}
```

---

### get_session_stats

**Description**: Get session statistics. Scope: session.

**Specific Parameters**: None (uses only standard parameters)

**Usage Scenarios**:

1. **Get full session stats**:
```json
{
  "name": "get_session_stats",
  "arguments": {}
}
```

2. **Get only turn count**:
```json
{
  "name": "get_session_stats",
  "arguments": {
    "jq_filter": ".TurnCount"
  }
}
```

---

### extract_tools

**Description**: Extract tool call history. Scope: project. Limit: 100 max.

**Specific Parameters**:
- `limit` (number): Max tools to extract (default: 100)

**Usage Scenarios**:

1. **Get last 50 tools**:
```json
{
  "name": "extract_tools",
  "arguments": {
    "limit": 50
  }
}
```

2. **Extract and count by status**:
```json
{
  "name": "extract_tools",
  "arguments": {
    "limit": 100,
    "jq_filter": "group_by(.Status) | map({status: .[0].Status, count: length})"
  }
}
```

---

### query_context

**Description**: Query error context. Required: error_signature.

**Specific Parameters**:
- `error_signature` (string, required): Error pattern ID
- `window` (number): Context window size (default: 3)

**Usage Scenarios**:

1. **Get context for specific error**:
```json
{
  "name": "query_context",
  "arguments": {
    "error_signature": "Bash:command not found",
    "window": 5
  }
}
```

---

## Migration Guide

### From analyze_errors → query_tools

**Old** (Phase 14, deprecated):
```json
{
  "name": "analyze_errors",
  "arguments": {}
}
```

**New** (Phase 15+):
```json
{
  "name": "query_tools",
  "arguments": {
    "status": "error",
    "jq_filter": "group_by(.ToolName) | map({tool: .[0].ToolName, count: length})",
    "stats_only": true
  }
}
```

### From aggregate_stats → query_tools + jq

**Old** (Phase 14, deprecated):
```json
{
  "name": "aggregate_stats",
  "arguments": {
    "group_by": "tool"
  }
}
```

**New** (Phase 15+):
```json
{
  "name": "query_tools",
  "arguments": {
    "jq_filter": "group_by(.ToolName) | map({tool: .[0].ToolName, count: length, errors: [.[] | select(.Status == \"error\")] | length})",
    "stats_only": true
  }
}
```

---

## Performance Benchmarks

### Output Size Comparison

| Tool | Default Output | With max_message_length=500 | With content_summary |
|------|----------------|------------------------------|----------------------|
| `query_user_messages` | ~10.7k tokens | ~1.5k tokens (86% reduction) | ~800 tokens (93% reduction) |
| `query_tools` | ~3k tokens | N/A (not applicable) | N/A |
| `extract_tools` | ~5k tokens | N/A | N/A |

### Query Performance

- jq filtering: +5-10ms overhead
- Message truncation: +2-5ms overhead
- Content summary: +1-3ms overhead

**Total overhead**: <15ms for typical queries (<200 results)

---

## Best Practices

1. **Always limit output for user message queries**:
   ```json
   {"max_message_length": 500}  // or content_summary: true
   ```

2. **Use stats_only for aggregations**:
   ```json
   {"stats_only": true, "jq_filter": "group_by(.X) | map({...})"}
   ```

3. **Prefer jq filtering over post-processing**:
   - Do: `{"jq_filter": ".[0:10]"}`
   - Don't: Retrieve 1000 results and take first 10 in code

4. **Use scope wisely**:
   - `scope: "session"` for current session only (faster)
   - `scope: "project"` for cross-session analysis (slower)

5. **Combine parameters for optimal output**:
   ```json
   {
     "limit": 50,              // Limit DB query
     "jq_filter": ".[] | select(.Status == \"error\")",  // Filter in memory
     "max_message_length": 500,  // Truncate large fields
     "stats_first": true         // Show stats before details
   }
   ```

---

## Troubleshooting

### "OUTPUT TRUNCATED" message

**Cause**: Output exceeded `max_output_bytes` (default 50KB)

**Solutions**:
1. Increase limit: `{"max_output_bytes": 102400}`
2. Use jq to select fewer fields: `{"jq_filter": ".[] | {uuid, status}"}`
3. Enable stats_only: `{"stats_only": true}`

### Large token usage in Claude

**Cause**: Message content includes session summaries

**Solutions**:
1. Use `max_message_length`: `{"max_message_length": 300}`
2. Use `content_summary`: `{"content_summary": true}`
3. Apply jq filter: `{"jq_filter": ".[] | {turn, timestamp}"}`

### Tool returns empty array

**Cause**: jq filter removed all results or no matches found

**Debug**:
1. Remove jq filter: `{"jq_filter": ".[]"}`
2. Check stats: `{"stats_only": true}`
3. Verify pattern/filters are correct

---

## Reference

- [MCP Protocol Specification](https://spec.modelcontextprotocol.io)
- [jq Manual](https://jqlang.github.io/jq/manual/)
- [meta-cc CLI Documentation](../README.md)
```

#### Acceptance Criteria

- ✅ Complete parameter reference table created
- ✅ All 15 tools documented with usage scenarios
- ✅ Migration guide from deprecated tools provided
- ✅ Output control strategies clearly explained
- ✅ Best practices section included
- ✅ Performance benchmarks documented
- ✅ Troubleshooting section covers common issues

#### Verification Commands

```bash
# Verify documentation exists
ls -la docs/mcp-tools-reference.md

# Check documentation length
wc -l docs/mcp-tools-reference.md
# Expected: ~400-500 lines

# Verify all tools documented
grep "^### " docs/mcp-tools-reference.md | wc -l
# Expected: ≥12 (one per tool)

# Check for completeness
grep -c "Usage Scenarios" docs/mcp-tools-reference.md
# Expected: ≥12
```

---

## Testing Strategy

### Unit Testing (TDD)

**Test-Driven Development Flow**:
1. Write test first (define expected behavior)
2. Run test (verify it fails)
3. Implement minimum code to pass
4. Refactor while keeping tests green

**Coverage Targets**:
- Stage 15.1: ≥90% coverage (filters.go)
- Stage 15.2: ≥85% coverage (tools.go updates)
- Stage 15.3: ≥90% coverage (description validation)
- Overall: ≥85% coverage for cmd/mcp-server

**Test Commands**:
```bash
# Run all MCP server tests
go test ./cmd/mcp-server -v

# Test specific stages
go test ./cmd/mcp-server -run TestTruncate -v          # Stage 15.1
go test ./cmd/mcp-server -run TestAllToolsHave -v      # Stage 15.2
go test ./cmd/mcp-server -run TestToolDescription -v   # Stage 15.3

# Check coverage
go test ./cmd/mcp-server -cover
go test ./cmd/mcp-server -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Testing

**Test Script**: `test-scripts/validate-phase-15.sh`

```bash
#!/bin/bash
# Phase 15 Validation Script

set -e

echo "=== Phase 15 MCP Output Control Validation ==="
echo ""

# Stage 15.1: Message truncation
echo "[1/4] Testing message content truncation..."

# Test max_message_length parameter
RESULT=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"query_user_messages","arguments":{"pattern":".*","limit":5,"max_message_length":100}}}' | ./meta-cc-mcp)

# Verify truncation applied
if echo "$RESULT" | jq -e '.result.content[0].text' | grep -q "TRUNCATED"; then
    echo "✅ Message truncation works"
else
    echo "⚠️  No truncation marker found (might be OK if messages are short)"
fi
echo ""

# Stage 15.1: Content summary
echo "[2/4] Testing content summary mode..."

RESULT=$(echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"query_user_messages","arguments":{"pattern":".*","limit":5,"content_summary":true}}}' | ./meta-cc-mcp)

# Verify only preview returned
if echo "$RESULT" | jq -e '.result.content[0].text' | grep -q "content_preview"; then
    echo "✅ Content summary mode works"
else
    echo "❌ Content summary mode failed"
    exit 1
fi
echo ""

# Stage 15.2: Standard parameters
echo "[3/4] Testing standard parameters presence..."

TOOLS=$(echo '{"jsonrpc":"2.0","id":3,"method":"tools/list"}' | ./meta-cc-mcp)

# Check that query_user_messages has max_message_length
if echo "$TOOLS" | jq -e '.result.tools[] | select(.name=="query_user_messages") | .inputSchema.properties.max_message_length' >/dev/null; then
    echo "✅ max_message_length parameter exists"
else
    echo "❌ max_message_length parameter missing"
    exit 1
fi

# Check that query_user_messages has content_summary
if echo "$TOOLS" | jq -e '.result.tools[] | select(.name=="query_user_messages") | .inputSchema.properties.content_summary' >/dev/null; then
    echo "✅ content_summary parameter exists"
else
    echo "❌ content_summary parameter missing"
    exit 1
fi
echo ""

# Stage 15.3: Description length
echo "[4/4] Testing tool description length..."

MAX_LEN=$(echo "$TOOLS" | jq -r '.result.tools[].description | length' | sort -n | tail -1)

if [ "$MAX_LEN" -le 100 ]; then
    echo "✅ All descriptions ≤100 characters (max: $MAX_LEN)"
else
    echo "❌ Description too long: $MAX_LEN characters"
    exit 1
fi
echo ""

echo "=== All Phase 15 Tests Passed ✅ ==="
echo ""
echo "Summary:"
echo "  - Message truncation: working"
echo "  - Content summary: working"
echo "  - Standard parameters: present"
echo "  - Description length: compliant"
```

**Run integration tests**:
```bash
# Make script executable
chmod +x test-scripts/validate-phase-15.sh

# Run tests (requires meta-cc-mcp binary)
make build
./test-scripts/validate-phase-15.sh
```

### Regression Testing

**Ensure Phase 14 functionality still works**:
```bash
# Run all existing unit tests
go test ./... -v

# Run Phase 14 integration tests
./test-scripts/validate-phase-14.sh

# Verify jq filtering still works
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"query_tools","arguments":{"jq_filter":".[] | select(.Status == \"error\")","limit":10}}}' | ./meta-cc-mcp

# Verify stats_only still works
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"query_tools","arguments":{"stats_only":true}}}' | ./meta-cc-mcp
```

---

## Success Criteria

### Functional Requirements

- ✅ Message content truncation implemented (`max_message_length`)
- ✅ Content summary mode implemented (`content_summary`)
- ✅ All tools expose standard parameters (including new ones)
- ✅ All tool descriptions ≤100 characters
- ✅ Deprecated tools marked with `[DEPRECATED]` prefix
- ✅ Output compression ≥80% for large user messages (10.7k → 1-2k tokens)

### Code Quality Requirements

- ✅ Unit test coverage ≥85% for cmd/mcp-server
- ✅ All unit tests pass (including existing tests)
- ✅ Integration tests pass (`validate-phase-15.sh`)
- ✅ No regressions in Phase 14 functionality
- ✅ Code follows existing patterns in cmd/mcp-server

### Documentation Requirements

- ✅ Complete MCP tools reference created (`docs/mcp-tools-reference.md`)
- ✅ Migration guide from deprecated tools provided
- ✅ Output control strategies documented
- ✅ Best practices section included
- ✅ Performance benchmarks documented
- ✅ Troubleshooting section created

### Integration Requirements

- ✅ MCP server correctly applies message truncation
- ✅ MCP server correctly applies content summary
- ✅ All tools callable via MCP protocol
- ✅ Tool parameters validated by MCP client
- ✅ Real-world validation: `query_user_messages` output reduced to <5k bytes

---

## Risk Assessment & Mitigation

### High Risk: Breaking Existing MCP Clients

**Risk**: Existing Claude Code sessions expect current output format

**Mitigation**:
- Default parameters maintain backward compatibility
- `max_message_length=0` disables truncation (unlimited)
- `content_summary=false` by default (full content returned)
- Deprecation warnings for tools to be removed

**Rollback Plan**:
```bash
# If issues arise, revert to Phase 14
git checkout feature/phase-14
make build
```

### Medium Risk: Truncation Too Aggressive

**Risk**: 500-character default may cut off important context

**Mitigation**:
- Make `max_message_length` configurable (default: 500, 0=unlimited)
- Add `original_length` field to show how much was truncated
- Document in tool description: "use max_message_length or content_summary"
- Provide examples in documentation

### Medium Risk: Documentation Outdated

**Risk**: MCP tools reference becomes stale as features are added

**Mitigation**:
- Link documentation to code (generate from tool definitions)
- Add CI check: `go test` verifies tool count matches docs
- Include "Last Updated" timestamp in documentation
- Create GitHub issue template for doc updates

### Low Risk: Performance Degradation

**Risk**: Truncation/filtering adds overhead to queries

**Mitigation**:
- Benchmark truncation performance (expected: <5ms overhead)
- Apply truncation only when necessary (not on all tools)
- Use efficient string operations (slicing vs regex)

**Performance baseline**:
```bash
# Benchmark truncation
go test ./cmd/mcp-server -bench=BenchmarkTruncate -benchmem
```

---

## Timeline & Milestones

### Day 1: Output Control Implementation

- **Morning**: Stage 15.1 implementation (filters.go)
- **Afternoon**: Stage 15.1 tests (executor_test.go)
- **End of Day**: Message truncation validated, coverage ≥90%

### Day 2: Standardization & Documentation

- **Morning**: Stage 15.2 (standardize parameters)
- **Early Afternoon**: Stage 15.3 (simplify descriptions)
- **Late Afternoon**: Stage 15.4 part 1 (start documentation)
- **End of Day**: Parameters standardized, descriptions simplified

### Day 3: Documentation & Validation

- **Morning**: Stage 15.4 part 2 (complete documentation)
- **Afternoon**: Integration testing, regression testing
- **End of Day**: Phase complete, documentation published

### Phase Completion Checklist

- ✅ All 4 stages complete
- ✅ Output compression target met (≥80%)
- ✅ All tests passing (unit + integration)
- ✅ Documentation complete and reviewed
- ✅ Migration guide published
- ✅ Deprecated tools removed (if applicable)

---

## Performance Benchmarks

### Expected Performance

**Output Size Reduction**:
```
Baseline (Phase 14):
- query_user_messages: ~10.7k tokens
- query_tools: ~3k tokens
- extract_tools: ~5k tokens

Target (Phase 15):
- query_user_messages (max_message_length=500): ~1.5k tokens (-86%)
- query_user_messages (content_summary=true): ~800 tokens (-93%)
- Other tools: unchanged (~3-5k tokens)
```

**Processing Overhead**:
```
Baseline (Phase 14): ~150ms average query time

Target (Phase 15):
- Message truncation: +5ms
- Content summary: +3ms
- Total: ~155-160ms (+3-7%)
```

**Memory Usage**:
```
Baseline: ~10MB per query

Target:
- Truncation: -40% memory (fewer strings in memory)
- Content summary: -60% memory (metadata only)
```

**Verification**:
```bash
# Benchmark output size
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"query_user_messages","arguments":{"pattern":".*","limit":10}}}' | ./meta-cc-mcp | wc -c
# Before: ~30000 bytes
# After (max_message_length=500): ~5000 bytes

# Benchmark query time
time echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"query_user_messages","arguments":{"pattern":".*","limit":50,"max_message_length":500}}}' | ./meta-cc-mcp >/dev/null
# Expected: <200ms
```

---

## Appendix

### A. Tool Parameter Matrix

| Tool | Specific Params | Standard Params | max_message_length | content_summary |
|------|----------------|-----------------|-------------------|-----------------|
| query_tools | tool, status, limit | ✅ All | ✅ | ✅ |
| extract_tools | limit | ✅ All | ✅ | ✅ |
| query_user_messages | pattern, limit | ✅ All | ✅ | ✅ |
| get_session_stats | (none) | ✅ All | ✅ | ✅ |
| query_context | error_signature, window | ✅ All | ✅ | ✅ |
| query_tool_sequences | pattern, min_occurrences | ✅ All | ✅ | ✅ |
| query_file_access | file | ✅ All | ✅ | ✅ |
| query_project_state | (none) | ✅ All | ✅ | ✅ |
| query_successful_prompts | limit, min_quality_score | ✅ All | ✅ | ✅ |
| query_tools_advanced | where, limit | ✅ All | ✅ | ✅ |
| query_time_series | interval, metric, where | ✅ All | ✅ | ✅ |
| query_files | sort_by, top, where | ✅ All | ✅ | ✅ |

### B. Output Size Examples

**Example 1: Default query (no truncation)**
```json
// Request
{"name": "query_user_messages", "arguments": {"pattern": ".*", "limit": 5}}

// Response (10.7k tokens)
[
  {
    "turn_sequence": 42,
    "timestamp": "2025-10-06T12:00:00Z",
    "content": "<8000 lines of session summary...>"
  },
  ...
]
```

**Example 2: With max_message_length=500**
```json
// Request
{"name": "query_user_messages", "arguments": {"pattern": ".*", "limit": 5, "max_message_length": 500}}

// Response (1.5k tokens)
[
  {
    "turn_sequence": 42,
    "timestamp": "2025-10-06T12:00:00Z",
    "content": "<first 500 chars of session summary...>... [TRUNCATED]",
    "content_truncated": true,
    "original_length": 25000
  },
  ...
]
```

**Example 3: With content_summary=true**
```json
// Request
{"name": "query_user_messages", "arguments": {"pattern": ".*", "limit": 5, "content_summary": true}}

// Response (800 tokens)
[
  {
    "turn_sequence": 42,
    "timestamp": "2025-10-06T12:00:00Z",
    "content_preview": "This is a session summary containing..."
  },
  ...
]
```

### C. Migration Examples

**Before (Phase 14)**:
```bash
# Query user messages (returns 10.7k tokens)
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"query_user_messages","arguments":{"pattern":".*"}}}' | ./meta-cc-mcp
```

**After (Phase 15)**:
```bash
# Option 1: Truncate content
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"query_user_messages","arguments":{"pattern":".*","max_message_length":500}}}' | ./meta-cc-mcp
# Returns 1.5k tokens

# Option 2: Summary mode
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"query_user_messages","arguments":{"pattern":".*","content_summary":true}}}' | ./meta-cc-mcp
# Returns 800 tokens

# Option 3: Combine with jq
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"query_user_messages","arguments":{"pattern":".*","jq_filter":".[0:10] | .[] | {turn, timestamp}","stats_only":false}}}' | ./meta-cc-mcp
# Returns <500 tokens
```

---

## Next Phase Preview

**Phase 16: Subagent Integration (Optional)**
- `@meta-coach` conversational analysis
- Automated workflow optimization suggestions
- Integration with Phase 14/15 MCP enhancements

**Phase 17: Advanced Analytics (Optional)**
- Cross-session pattern detection
- Temporal analysis (workflow evolution over time)
- Predictive error detection

---

**Phase 15 Implementation Plan Complete**

**Estimated Timeline**: 2-3 days

**Risk Level**: Low-Medium (new features, minimal breaking changes)

**Success Probability**: High (clear design, comprehensive tests, incremental approach)

**Ready to Begin**: ✅
