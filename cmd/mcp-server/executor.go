package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

const (
	DefaultMaxOutputBytes = 51200 // 50KB
)

type ToolExecutor struct {
	metaCCPath string
}

func NewToolExecutor() *ToolExecutor {
	// Find meta-cc executable
	metaCCPath, err := exec.LookPath("meta-cc")
	if err != nil {
		// Assume meta-cc is in the same directory or current directory
		metaCCPath = "./meta-cc"
	}

	return &ToolExecutor{
		metaCCPath: metaCCPath,
	}
}

// ExecuteTool executes a meta-cc command and applies jq filtering
func (e *ToolExecutor) ExecuteTool(toolName string, args map[string]interface{}) (string, error) {
	// Handle deprecated tools
	if toolName == "analyze_errors" {
		return "", fmt.Errorf("[DEPRECATED] analyze_errors is deprecated. Use query_tools with status='error' filter instead. " +
			"Example: {\"name\":\"query_tools\",\"arguments\":{\"status\":\"error\",\"jq_filter\":\".[] | select(.Status == \\\"error\\\")\",\"stats_only\":true}}")
	}

	if toolName == "aggregate_stats" {
		return "", fmt.Errorf("[DEPRECATED] aggregate_stats is deprecated. Use query_tools with jq_filter and stats_only=true instead. " +
			"Example: {\"name\":\"query_tools\",\"arguments\":{\"jq_filter\":\"group_by(.ToolName) | map({tool: .[0].ToolName, count: length})\",\"stats_only\":true}}")
	}

	// Extract common parameters
	jqFilter := getStringParam(args, "jq_filter", ".[]")
	statsOnly := getBoolParam(args, "stats_only", false)
	statsFirst := getBoolParam(args, "stats_first", false)
	maxOutputBytes := getIntParam(args, "max_output_bytes", DefaultMaxOutputBytes)
	scope := getStringParam(args, "scope", "project")
	outputFormat := getStringParam(args, "output_format", "jsonl")

	// Build meta-cc command
	cmdArgs := e.buildCommand(toolName, args, scope, outputFormat)
	if cmdArgs == nil {
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}

	// Execute meta-cc
	rawOutput, err := e.executeMetaCC(cmdArgs)
	if err != nil {
		return "", err
	}

	// Apply jq filter
	filtered, err := ApplyJQFilter(rawOutput, jqFilter)
	if err != nil {
		return "", fmt.Errorf("jq filter error: %w", err)
	}

	// Generate stats if requested
	var output string
	if statsOnly {
		output, err = GenerateStats(filtered)
		if err != nil {
			return "", err
		}
	} else if statsFirst {
		stats, _ := GenerateStats(filtered)
		output = stats + "\n---\n" + filtered
	} else {
		output = filtered
	}

	// Apply output length limit
	if len(output) > maxOutputBytes {
		output = output[:maxOutputBytes]
		output += fmt.Sprintf("\n[OUTPUT TRUNCATED: exceeded %d bytes limit]", maxOutputBytes)
	}

	return output, nil
}

func (e *ToolExecutor) buildCommand(toolName string, args map[string]interface{}, scope string, outputFormat string) []string {
	cmdArgs := []string{}

	// Add project flag for project-level queries
	if scope == "project" {
		cmdArgs = append(cmdArgs, "--project", ".")
	}

	// Map tool name to meta-cc command
	switch toolName {
	case "query_tools":
		cmdArgs = append(cmdArgs, "query", "tools")
		if tool := getStringParam(args, "tool", ""); tool != "" {
			cmdArgs = append(cmdArgs, "--tool", tool)
		}
		if status := getStringParam(args, "status", ""); status != "" {
			cmdArgs = append(cmdArgs, "--status", status)
		}
		if limit := getIntParam(args, "limit", 0); limit > 0 {
			cmdArgs = append(cmdArgs, "--limit", strconv.Itoa(limit))
		}

	case "extract_tools":
		cmdArgs = append(cmdArgs, "extract", "tools")
		if limit := getIntParam(args, "limit", 0); limit > 0 {
			cmdArgs = append(cmdArgs, "--limit", strconv.Itoa(limit))
		}

	case "query_user_messages":
		cmdArgs = append(cmdArgs, "query", "user-messages")
		if pattern := getStringParam(args, "pattern", ""); pattern != "" {
			cmdArgs = append(cmdArgs, "--pattern", pattern)
		}
		if limit := getIntParam(args, "limit", 0); limit > 0 {
			cmdArgs = append(cmdArgs, "--limit", strconv.Itoa(limit))
		}

	case "get_session_stats":
		cmdArgs = append(cmdArgs, "parse", "stats")

	case "analyze_errors":
		cmdArgs = append(cmdArgs, "analyze", "errors")

	case "query_context":
		cmdArgs = append(cmdArgs, "query", "context")
		if errorSig := getStringParam(args, "error_signature", ""); errorSig != "" {
			cmdArgs = append(cmdArgs, "--error-signature", errorSig)
		}
		if window := getIntParam(args, "window", 0); window > 0 {
			cmdArgs = append(cmdArgs, "--window", strconv.Itoa(window))
		}

	case "query_tool_sequences":
		cmdArgs = append(cmdArgs, "analyze", "sequences")
		if pattern := getStringParam(args, "pattern", ""); pattern != "" {
			cmdArgs = append(cmdArgs, "--pattern", pattern)
		}
		if minOccur := getIntParam(args, "min_occurrences", 0); minOccur > 0 {
			cmdArgs = append(cmdArgs, "--min-occurrences", strconv.Itoa(minOccur))
		}

	case "query_file_access":
		cmdArgs = append(cmdArgs, "query", "file-access")
		if file := getStringParam(args, "file", ""); file != "" {
			cmdArgs = append(cmdArgs, "--file", file)
		}

	case "query_project_state":
		cmdArgs = append(cmdArgs, "query", "project-state")

	case "query_successful_prompts":
		cmdArgs = append(cmdArgs, "query", "successful-prompts")
		if limit := getIntParam(args, "limit", 0); limit > 0 {
			cmdArgs = append(cmdArgs, "--limit", strconv.Itoa(limit))
		}
		if minQuality := getFloatParam(args, "min_quality_score", 0); minQuality > 0 {
			cmdArgs = append(cmdArgs, "--min-quality", fmt.Sprintf("%.2f", minQuality))
		}

	case "query_tools_advanced":
		cmdArgs = append(cmdArgs, "query", "tools")
		if where := getStringParam(args, "where", ""); where != "" {
			cmdArgs = append(cmdArgs, "--where", where)
		}
		if limit := getIntParam(args, "limit", 0); limit > 0 {
			cmdArgs = append(cmdArgs, "--limit", strconv.Itoa(limit))
		}

	case "query_time_series":
		cmdArgs = append(cmdArgs, "stats", "timeseries")
		if interval := getStringParam(args, "interval", ""); interval != "" {
			cmdArgs = append(cmdArgs, "--interval", interval)
		}
		if metric := getStringParam(args, "metric", ""); metric != "" {
			cmdArgs = append(cmdArgs, "--metric", metric)
		}
		if where := getStringParam(args, "where", ""); where != "" {
			cmdArgs = append(cmdArgs, "--where", where)
		}

	case "query_files":
		cmdArgs = append(cmdArgs, "analyze", "file-churn")
		if sortBy := getStringParam(args, "sort_by", ""); sortBy != "" {
			cmdArgs = append(cmdArgs, "--sort-by", sortBy)
		}
		if top := getIntParam(args, "top", 0); top > 0 {
			cmdArgs = append(cmdArgs, "--top", strconv.Itoa(top))
		}
		if where := getStringParam(args, "where", ""); where != "" {
			cmdArgs = append(cmdArgs, "--where", where)
		}

	default:
		return nil
	}

	// Always output JSONL (unless specified otherwise)
	cmdArgs = append(cmdArgs, "--output", outputFormat)

	return cmdArgs
}

func (e *ToolExecutor) executeMetaCC(cmdArgs []string) (string, error) {
	cmd := exec.Command(e.metaCCPath, cmdArgs...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set current directory for meta-cc
	cmd.Dir, _ = os.Getwd()

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("meta-cc error: %s", stderr.String())
	}

	return stdout.String(), nil
}

// Helper functions
func getStringParam(args map[string]interface{}, key, defaultVal string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return defaultVal
}

func getBoolParam(args map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	return defaultVal
}

func getIntParam(args map[string]interface{}, key string, defaultVal int) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	if v, ok := args[key].(int); ok {
		return v
	}
	return defaultVal
}

func getFloatParam(args map[string]interface{}, key string, defaultVal float64) float64 {
	if v, ok := args[key].(float64); ok {
		return v
	}
	return defaultVal
}
