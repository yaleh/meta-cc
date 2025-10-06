package main

import (
	"strings"
	"testing"
)

func TestDeprecatedTools_AnalyzeErrors(t *testing.T) {
	executor := NewToolExecutor()

	args := map[string]interface{}{
		"scope": "project",
	}

	_, err := executor.ExecuteTool("analyze_errors", args)

	// Should return deprecation error
	if err == nil {
		t.Error("expected deprecation error")
	}

	if !strings.Contains(err.Error(), "DEPRECATED") {
		t.Error("expected DEPRECATED in error message")
	}

	if !strings.Contains(err.Error(), "query_tools") {
		t.Error("expected migration hint to query_tools")
	}
}

func TestDeprecatedTools_AggregateStats(t *testing.T) {
	executor := NewToolExecutor()

	args := map[string]interface{}{
		"group_by": "tool",
	}

	_, err := executor.ExecuteTool("aggregate_stats", args)

	// Should return deprecation error
	if err == nil {
		t.Error("expected deprecation error")
	}

	if !strings.Contains(err.Error(), "DEPRECATED") {
		t.Error("expected DEPRECATED in error message")
	}

	if !strings.Contains(err.Error(), "query_tools") {
		t.Error("expected migration hint to query_tools")
	}
}

func TestToolsList_NoAggregateStats(t *testing.T) {
	tools := getToolDefinitions()

	for _, tool := range tools {
		if tool.Name == "aggregate_stats" {
			t.Error("aggregate_stats should be removed from tools list")
		}
	}
}

func TestToolsList_AnalyzeErrorsMarked(t *testing.T) {
	tools := getToolDefinitions()

	found := false
	for _, tool := range tools {
		if tool.Name == "analyze_errors" {
			found = true
			if !strings.Contains(tool.Description, "DEPRECATED") {
				t.Error("analyze_errors should be marked as DEPRECATED")
			}
		}
	}

	if !found {
		t.Error("analyze_errors tool not found")
	}
}
