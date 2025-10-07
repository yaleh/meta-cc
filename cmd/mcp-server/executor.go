package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
	// Handle cleanup tool separately (no meta-cc command needed)
	if toolName == "cleanup_temp_files" {
		return executeCleanupTool(args)
	}

	// Extract common parameters
	jqFilter := getStringParam(args, "jq_filter", ".[]")
	statsOnly := getBoolParam(args, "stats_only", false)
	statsFirst := getBoolParam(args, "stats_first", false)
	scope := getStringParam(args, "scope", "project")
	outputFormat := getStringParam(args, "output_format", "jsonl")

	// Extract message truncation parameters (for query_user_messages)
	// Default to 0 (no truncation) - rely on hybrid mode for large results
	maxMessageLength := getIntParam(args, "max_message_length", 0)
	contentSummary := getBoolParam(args, "content_summary", false)

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

	// Parse JSONL to interface array for hybrid mode adaptation
	parsedData, err := e.parseJSONL(filtered)
	if err != nil {
		return "", fmt.Errorf("JSONL parse error: %w", err)
	}

	// Apply message filters for query_user_messages (only if explicitly requested)
	// By default, rely on hybrid mode (no truncation)
	if toolName == "query_user_messages" && (maxMessageLength > 0 || contentSummary) {
		parsedData = e.applyMessageFiltersToData(parsedData, maxMessageLength, contentSummary)
	}

	// Handle stats_only and stats_first modes
	if statsOnly {
		// Convert back to JSONL for stats generation
		jsonlData, err := e.dataToJSONL(parsedData)
		if err != nil {
			return "", err
		}
		output, err := GenerateStats(jsonlData)
		if err != nil {
			return "", err
		}
		return output, nil
	} else if statsFirst {
		// Convert back to JSONL for stats generation
		jsonlData, err := e.dataToJSONL(parsedData)
		if err != nil {
			return "", err
		}
		stats, _ := GenerateStats(jsonlData)

		// Adapt response for data portion
		response, err := adaptResponse(parsedData, args, toolName)
		if err != nil {
			return "", err
		}

		// Serialize and prepend stats
		serialized, err := serializeResponse(response)
		if err != nil {
			return "", err
		}

		return stats + "\n---\n" + serialized, nil
	}

	// Adapt response to hybrid mode (inline or file_ref)
	response, err := adaptResponse(parsedData, args, toolName)
	if err != nil {
		return "", fmt.Errorf("response adaptation error: %w", err)
	}

	// Serialize response (no truncation - rely on hybrid mode)
	output, err := serializeResponse(response)
	if err != nil {
		return "", err
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
		// New parameter: include built-in tools (default: false for cleaner patterns)
		if includeBuiltin := getBoolParam(args, "include_builtin_tools", false); includeBuiltin {
			cmdArgs = append(cmdArgs, "--include-builtin-tools")
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
			cmdArgs = append(cmdArgs, "--min-quality-score", fmt.Sprintf("%.2f", minQuality))
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

	case "cleanup_temp_files":
		// Handle cleanup tool directly (no meta-cc command)
		return nil

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
		// Check if this is an exit error with specific exit code
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()

			// Exit code 2 means "no results found" - this is not an error
			// Return stdout content (may be "[]" or empty string)
			if exitCode == 2 {
				return stdout.String(), nil
			}
		}

		// For exit code 1 or other errors, return error message
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg == "" {
			// If stderr is empty, include command details for debugging
			return "", fmt.Errorf("meta-cc error: command failed with exit code (stderr empty)\nCommand: %s %s", e.metaCCPath, strings.Join(cmdArgs, " "))
		}
		return "", fmt.Errorf("meta-cc error: %s\nCommand: %s %s", stderrMsg, e.metaCCPath, strings.Join(cmdArgs, " "))
	}

	return stdout.String(), nil
}

// parseJSONL parses JSONL string into array of interfaces
func (e *ToolExecutor) parseJSONL(jsonlData string) ([]interface{}, error) {
	jsonlData = strings.TrimSpace(jsonlData)

	// Handle special cases: empty input or "[]" (exit code 2 scenario)
	if jsonlData == "" || jsonlData == "[]" {
		return []interface{}{}, nil
	}

	lines := strings.Split(jsonlData, "\n")
	var data []interface{}

	for _, line := range lines {
		if line == "" {
			continue
		}

		var obj interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		data = append(data, obj)
	}

	return data, nil
}

// dataToJSONL converts array of interfaces to JSONL string
func (e *ToolExecutor) dataToJSONL(data []interface{}) (string, error) {
	var output strings.Builder
	for _, record := range data {
		jsonBytes, err := json.Marshal(record)
		if err != nil {
			return "", err
		}
		output.Write(jsonBytes)
		output.WriteString("\n")
	}
	return output.String(), nil
}

// applyMessageFiltersToData applies content truncation or summary mode to user messages (data array)
func (e *ToolExecutor) applyMessageFiltersToData(messages []interface{}, maxMessageLength int, contentSummary bool) []interface{} {
	if contentSummary {
		return ApplyContentSummary(messages)
	}
	return TruncateMessageContent(messages, maxMessageLength)
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
