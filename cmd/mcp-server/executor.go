package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/yaleh/meta-cc/internal/config"
	mcerrors "github.com/yaleh/meta-cc/internal/errors"
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
func (e *ToolExecutor) ExecuteTool(cfg *config.Config, toolName string, args map[string]interface{}) (string, error) {
	// Get scope for metrics
	scope := getStringParam(args, "scope", "project")

	// Start timing for tool execution metrics
	start := time.Now()

	// Handle cleanup tool separately (no meta-cc command needed)
	if toolName == "cleanup_temp_files" {
		output, err := executeCleanupTool(args)
		elapsed := time.Since(start)

		if err != nil {
			RecordToolCall(toolName, scope, "error")
			RecordToolExecutionDuration(toolName, scope, elapsed)
			RecordError(toolName, classifyError(err), GetErrorSeverity(classifyError(err)))
			return "", err
		}

		RecordToolCall(toolName, scope, "success")
		RecordToolExecutionDuration(toolName, scope, elapsed)
		return output, nil
	}

	// Handle list_capabilities tool separately (no meta-cc command needed)
	if toolName == "list_capabilities" {
		output, err := executeListCapabilitiesTool(cfg, args)
		elapsed := time.Since(start)

		if err != nil {
			RecordToolCall(toolName, scope, "error")
			RecordToolExecutionDuration(toolName, scope, elapsed)
			RecordError(toolName, classifyError(err), GetErrorSeverity(classifyError(err)))
			return "", err
		}

		RecordToolCall(toolName, scope, "success")
		RecordToolExecutionDuration(toolName, scope, elapsed)
		return output, nil
	}

	// Handle get_capability tool separately (no meta-cc command needed)
	if toolName == "get_capability" {
		output, err := executeGetCapabilityTool(cfg, args)
		elapsed := time.Since(start)

		if err != nil {
			RecordToolCall(toolName, scope, "error")
			RecordToolExecutionDuration(toolName, scope, elapsed)
			RecordError(toolName, classifyError(err), GetErrorSeverity(classifyError(err)))
			return "", err
		}

		RecordToolCall(toolName, scope, "success")
		RecordToolExecutionDuration(toolName, scope, elapsed)
		return output, nil
	}

	// Extract common parameters
	jqFilter := getStringParam(args, "jq_filter", ".[]")
	statsOnly := getBoolParam(args, "stats_only", false)
	statsFirst := getBoolParam(args, "stats_first", false)
	outputFormat := getStringParam(args, "output_format", "jsonl")

	// Extract message truncation parameters (for query_user_messages)
	// Default to 0 (no truncation) - rely on hybrid mode for large results
	maxMessageLength := getIntParam(args, "max_message_length", 0)
	contentSummary := getBoolParam(args, "content_summary", false)

	// Build meta-cc command
	cmdArgs := e.buildCommand(toolName, args, scope, outputFormat)
	if cmdArgs == nil {
		elapsed := time.Since(start)
		RecordToolCall(toolName, scope, "error")
		RecordToolExecutionDuration(toolName, scope, elapsed)
		RecordError(toolName, "validation_error", "error")
		return "", fmt.Errorf("unknown tool %s in executor: %w", toolName, mcerrors.ErrUnknownTool)
	}

	// Execute meta-cc
	rawOutput, err := e.executeMetaCC(cmdArgs)
	if err != nil {
		elapsed := time.Since(start)
		errorType := classifyError(err)
		slog.Error("meta-cc execution failed",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", errorType,
		)
		RecordToolCall(toolName, scope, "error")
		RecordToolExecutionDuration(toolName, scope, elapsed)
		RecordError(toolName, errorType, GetErrorSeverity(errorType))
		return "", err
	}

	// Apply jq filter
	filtered, err := ApplyJQFilter(rawOutput, jqFilter)
	if err != nil {
		slog.Error("jq filter application failed",
			"tool_name", toolName,
			"jq_filter", jqFilter,
			"error", err.Error(),
			"error_type", "execution_error",
		)
		// Preserve detailed error message from ApplyJQFilter
		return "", fmt.Errorf("jq filter error for tool %s: %w", toolName, err)
	}

	// Parse JSONL to interface array for hybrid mode adaptation
	parsedData, err := e.parseJSONL(filtered)
	if err != nil {
		slog.Error("JSONL parsing failed",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", "parse_error",
		)
		return "", fmt.Errorf("JSONL parse error for tool %s: %w", toolName, mcerrors.ErrParseError)
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
			slog.Error("dataToJSONL conversion failed (stats_only)",
				"tool_name", toolName,
				"error", err.Error(),
				"error_type", "parse_error",
			)
			return "", err
		}
		output, err := GenerateStats(jsonlData)
		if err != nil {
			slog.Error("stats generation failed",
				"tool_name", toolName,
				"error", err.Error(),
				"error_type", "execution_error",
			)
			return "", err
		}
		return output, nil
	} else if statsFirst {
		// Convert back to JSONL for stats generation
		jsonlData, err := e.dataToJSONL(parsedData)
		if err != nil {
			slog.Error("dataToJSONL conversion failed (stats_first)",
				"tool_name", toolName,
				"error", err.Error(),
				"error_type", "parse_error",
			)
			return "", err
		}
		stats, _ := GenerateStats(jsonlData)

		// Adapt response for data portion
		response, err := adaptResponse(cfg, parsedData, args, toolName)
		if err != nil {
			slog.Error("response adaptation failed (stats_first)",
				"tool_name", toolName,
				"error", err.Error(),
				"error_type", "execution_error",
			)
			return "", err
		}

		// Serialize and prepend stats
		serialized, err := serializeResponse(response)
		if err != nil {
			slog.Error("response serialization failed (stats_first)",
				"tool_name", toolName,
				"error", err.Error(),
				"error_type", "parse_error",
			)
			return "", err
		}

		return stats + "\n---\n" + serialized, nil
	}

	response, err := adaptResponse(cfg, parsedData, args, toolName)
	if err != nil {
		slog.Error("response adaptation failed",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", "execution_error",
		)
		return "", fmt.Errorf("response adaptation error for tool %s: %w", toolName, err)
	}

	// Serialize response (no truncation - rely on hybrid mode)
	output, err := serializeResponse(response)
	if err != nil {
		slog.Error("response serialization failed",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", "parse_error",
		)
		return "", err
	}

	slog.Debug("tool execution pipeline completed successfully",
		"tool_name", toolName,
		"output_length", len(output),
	)

	// Record successful tool execution metrics
	elapsed := time.Since(start)
	RecordToolCall(toolName, scope, "success")
	RecordToolExecutionDuration(toolName, scope, elapsed)

	return output, nil
}

func (e *ToolExecutor) buildCommand(toolName string, args map[string]interface{}, scope string, outputFormat string) []string {
	cmdArgs := []string{}

	// Add scope flags based on scope parameter
	if scope == "project" {
		// Project-level: explicitly set --project . to load all sessions
		cmdArgs = append(cmdArgs, "--project", ".")
	} else if scope == "session" {
		// Session-level: use --session-only flag to load only current session
		cmdArgs = append(cmdArgs, "--session-only")
	}
	// If scope is neither (shouldn't happen with default), CLI will use project-level default

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

	case "query_assistant_messages":
		cmdArgs = append(cmdArgs, "query", "assistant-messages")
		if pattern := getStringParam(args, "pattern", ""); pattern != "" {
			cmdArgs = append(cmdArgs, "--pattern", pattern)
		}
		if minTools := getIntParam(args, "min_tools", 0); minTools > 0 {
			cmdArgs = append(cmdArgs, "--min-tools", strconv.Itoa(minTools))
		}
		if maxTools := getIntParam(args, "max_tools", 0); maxTools > 0 {
			cmdArgs = append(cmdArgs, "--max-tools", strconv.Itoa(maxTools))
		}
		if minTokens := getIntParam(args, "min_tokens_output", 0); minTokens > 0 {
			cmdArgs = append(cmdArgs, "--min-tokens-output", strconv.Itoa(minTokens))
		}
		if minLength := getIntParam(args, "min_length", 0); minLength > 0 {
			cmdArgs = append(cmdArgs, "--min-length", strconv.Itoa(minLength))
		}
		if maxLength := getIntParam(args, "max_length", 0); maxLength > 0 {
			cmdArgs = append(cmdArgs, "--max-length", strconv.Itoa(maxLength))
		}
		if limit := getIntParam(args, "limit", 0); limit > 0 {
			cmdArgs = append(cmdArgs, "--limit", strconv.Itoa(limit))
		}

	case "query_conversation":
		cmdArgs = append(cmdArgs, "query", "conversation")
		if startTurn := getIntParam(args, "start_turn", 0); startTurn > 0 {
			cmdArgs = append(cmdArgs, "--start-turn", strconv.Itoa(startTurn))
		}
		if endTurn := getIntParam(args, "end_turn", 0); endTurn > 0 {
			cmdArgs = append(cmdArgs, "--end-turn", strconv.Itoa(endTurn))
		}
		if pattern := getStringParam(args, "pattern", ""); pattern != "" {
			cmdArgs = append(cmdArgs, "--pattern", pattern)
		}
		if target := getStringParam(args, "pattern_target", ""); target != "" {
			cmdArgs = append(cmdArgs, "--pattern-target", target)
		}
		if minDuration := getIntParam(args, "min_duration", 0); minDuration > 0 {
			cmdArgs = append(cmdArgs, "--min-duration", strconv.Itoa(minDuration))
		}
		if maxDuration := getIntParam(args, "max_duration", 0); maxDuration > 0 {
			cmdArgs = append(cmdArgs, "--max-duration", strconv.Itoa(maxDuration))
		}
		if limit := getIntParam(args, "limit", 0); limit > 0 {
			cmdArgs = append(cmdArgs, "--limit", strconv.Itoa(limit))
		}

	case "query_files":
		cmdArgs = append(cmdArgs, "analyze", "file-churn")
		// Only pass --threshold (data extraction parameter)
		// Filtering (where), sorting (sort_by), and limiting (top) should be done via jq_filter
		if threshold := getIntParam(args, "threshold", 0); threshold > 0 && threshold != 5 {
			cmdArgs = append(cmdArgs, "--threshold", strconv.Itoa(threshold))
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
	dir, err := os.Getwd()
	if err != nil {
		slog.Error("failed to get working directory",
			"error", err.Error(),
			"error_type", "io_error",
		)
		return "", fmt.Errorf("failed to get working directory: %w", mcerrors.ErrFileIO)
	}
	cmd.Dir = dir

	slog.Debug("executing meta-cc command",
		"command", e.metaCCPath,
		"args", strings.Join(cmdArgs, " "),
		"working_dir", dir,
	)

	if err := cmd.Run(); err != nil {
		// Check if this is an exit error with specific exit code
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()

			// Exit code 2 means "no results found" - this is not an error
			// Return stdout content (may be "[]" or empty string)
			if exitCode == 2 {
				slog.Debug("meta-cc returned no results",
					"exit_code", exitCode,
					"command", strings.Join(cmdArgs, " "),
				)
				return stdout.String(), nil
			}

			// For other exit codes, log as error
			stderrMsg := strings.TrimSpace(stderr.String())
			slog.Error("meta-cc command failed",
				"exit_code", exitCode,
				"stderr", stderrMsg,
				"command", strings.Join(cmdArgs, " "),
				"error_type", "execution_error",
			)
		} else {
			slog.Error("meta-cc command execution error",
				"error", err.Error(),
				"command", strings.Join(cmdArgs, " "),
				"error_type", classifyError(err),
			)
		}

		// For exit code 1 or other errors, return error message
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg == "" {
			// If stderr is empty, include command details for debugging
			return "", fmt.Errorf("meta-cc command '%s %s' failed with exit code (stderr empty): %w",
				e.metaCCPath, strings.Join(cmdArgs, " "), mcerrors.ErrFileIO)
		}
		return "", fmt.Errorf("meta-cc command '%s %s' failed: %s: %w",
			e.metaCCPath, strings.Join(cmdArgs, " "), stderrMsg, mcerrors.ErrFileIO)
	}

	slog.Debug("meta-cc command completed",
		"output_length", len(stdout.String()),
	)

	return stdout.String(), nil
}

// parseJSONL parses JSONL string into array of interfaces
func (e *ToolExecutor) parseJSONL(jsonlData string) ([]interface{}, error) {
	jsonlData = strings.TrimSpace(jsonlData)

	// Handle special cases: empty input or "[]" (exit code 2 scenario)
	if jsonlData == "" || jsonlData == "[]" {
		slog.Debug("parseJSONL: empty input or no results",
			"input", jsonlData,
		)
		return []interface{}{}, nil
	}

	lines := strings.Split(jsonlData, "\n")
	var data []interface{}

	for i, line := range lines {
		if line == "" {
			continue
		}

		var obj interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			slog.Error("failed to parse JSONL line",
				"line_number", i+1,
				"line_content", line,
				"error", err.Error(),
				"error_type", "parse_error",
			)
			return nil, fmt.Errorf("invalid JSON on line %d: %w", i+1, mcerrors.ErrParseError)
		}
		data = append(data, obj)
	}

	slog.Debug("parseJSONL completed",
		"record_count", len(data),
	)

	return data, nil
}

// dataToJSONL converts array of interfaces to JSONL string
func (e *ToolExecutor) dataToJSONL(data []interface{}) (string, error) {
	var output strings.Builder
	for i, record := range data {
		jsonBytes, err := json.Marshal(record)
		if err != nil {
			slog.Error("failed to marshal record to JSON",
				"record_index", i,
				"error", err.Error(),
				"error_type", "parse_error",
			)
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
