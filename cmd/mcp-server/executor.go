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

type toolPipelineConfig struct {
	jqFilter         string
	statsOnly        bool
	statsFirst       bool
	outputFormat     string
	maxMessageLength int
	contentSummary   bool
}

func newToolPipelineConfig(args map[string]interface{}) toolPipelineConfig {
	return toolPipelineConfig{
		jqFilter:         getStringParam(args, "jq_filter", ".[]"),
		statsOnly:        getBoolParam(args, "stats_only", false),
		statsFirst:       getBoolParam(args, "stats_first", false),
		outputFormat:     getStringParam(args, "output_format", "jsonl"),
		maxMessageLength: getIntParam(args, "max_message_length", 0),
		contentSummary:   getBoolParam(args, "content_summary", false),
	}
}

func (c toolPipelineConfig) requiresMessageFilters() bool {
	return c.maxMessageLength > 0 || c.contentSummary
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

func determineScope(toolName string, args map[string]interface{}) string {
	defaultScope := "project"
	if toolName == "get_session_stats" {
		defaultScope = "session"
	}
	return getStringParam(args, "scope", defaultScope)
}

func recordToolSuccess(toolName, scope string, start time.Time) {
	elapsed := time.Since(start)
	RecordToolCall(toolName, scope, "success")
	RecordToolExecutionDuration(toolName, scope, elapsed)
}

func recordToolFailure(toolName, scope string, start time.Time, errorType string) {
	elapsed := time.Since(start)
	RecordToolCall(toolName, scope, "error")
	RecordToolExecutionDuration(toolName, scope, elapsed)
	RecordError(toolName, errorType, GetErrorSeverity(errorType))
}

func (e *ToolExecutor) executeSpecialTool(cfg *config.Config, toolName, scope string, args map[string]interface{}, start time.Time) (string, bool, error) {
	switch toolName {
	case "cleanup_temp_files":
		output, err := executeCleanupTool(args)
		if err != nil {
			errorType := classifyError(err)
			recordToolFailure(toolName, scope, start, errorType)
			return "", true, err
		}
		recordToolSuccess(toolName, scope, start)
		return output, true, nil

	case "list_capabilities":
		output, err := executeListCapabilitiesTool(cfg, args)
		if err != nil {
			errorType := classifyError(err)
			recordToolFailure(toolName, scope, start, errorType)
			return "", true, err
		}
		recordToolSuccess(toolName, scope, start)
		return output, true, nil

	case "get_capability":
		output, err := executeGetCapabilityTool(cfg, args)
		if err != nil {
			errorType := classifyError(err)
			recordToolFailure(toolName, scope, start, errorType)
			return "", true, err
		}
		recordToolSuccess(toolName, scope, start)
		return output, true, nil

	default:
		return "", false, nil
	}
}

// ExecuteTool executes a meta-cc command and applies jq filtering
func (e *ToolExecutor) ExecuteTool(cfg *config.Config, toolName string, args map[string]interface{}) (string, error) {
	scope := determineScope(toolName, args)
	start := time.Now()

	if output, handled, err := e.executeSpecialTool(cfg, toolName, scope, args, start); handled {
		return output, err
	}

	config := newToolPipelineConfig(args)
	cmdArgs := e.buildCommand(toolName, args, scope, config.outputFormat)
	if cmdArgs == nil {
		recordToolFailure(toolName, scope, start, "validation_error")
		return "", fmt.Errorf("unknown tool %s in executor: %w", toolName, mcerrors.ErrUnknownTool)
	}

	rawOutput, err := e.executeMetaCC(cmdArgs)
	if err != nil {
		errorType := classifyError(err)
		slog.Error("meta-cc execution failed",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", errorType,
		)
		recordToolFailure(toolName, scope, start, errorType)
		return "", err
	}

	filtered, err := ApplyJQFilter(rawOutput, config.jqFilter)
	if err != nil {
		slog.Error("jq filter application failed",
			"tool_name", toolName,
			"jq_filter", config.jqFilter,
			"error", err.Error(),
			"error_type", "execution_error",
		)
		return "", fmt.Errorf("jq filter error for tool %s: %w", toolName, err)
	}

	parsedData, err := e.parseJSONL(filtered)
	if err != nil {
		slog.Error("JSONL parsing failed",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", "parse_error",
		)
		return "", fmt.Errorf("JSONL parse error for tool %s: %w", toolName, mcerrors.ErrParseError)
	}

	if toolName == "query_user_messages" && config.requiresMessageFilters() {
		parsedData = e.applyMessageFiltersToData(parsedData, config.maxMessageLength, config.contentSummary)
	}

	output, err := e.buildResponse(cfg, parsedData, args, toolName, config)
	if err != nil {
		return "", err
	}

	slog.Debug("tool execution pipeline completed successfully",
		"tool_name", toolName,
		"output_length", len(output),
	)

	recordToolSuccess(toolName, scope, start)
	return output, nil
}

func (e *ToolExecutor) buildResponse(cfg *config.Config, parsedData []interface{}, args map[string]interface{}, toolName string, pipeline toolPipelineConfig) (string, error) {
	if pipeline.statsOnly {
		return e.buildStatsOnlyResponse(parsedData, toolName)
	}

	if pipeline.statsFirst {
		return e.buildStatsFirstResponse(cfg, parsedData, args, toolName)
	}

	return e.buildStandardResponse(cfg, parsedData, args, toolName)
}

func (e *ToolExecutor) buildStatsOnlyResponse(parsedData []interface{}, toolName string) (string, error) {
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
}

func (e *ToolExecutor) buildStatsFirstResponse(cfg *config.Config, parsedData []interface{}, args map[string]interface{}, toolName string) (string, error) {
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
	response, err := adaptResponse(cfg, parsedData, args, toolName)
	if err != nil {
		slog.Error("response adaptation failed (stats_first)",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", "execution_error",
		)
		return "", err
	}

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

func (e *ToolExecutor) buildStandardResponse(cfg *config.Config, parsedData []interface{}, args map[string]interface{}, toolName string) (string, error) {
	response, err := adaptResponse(cfg, parsedData, args, toolName)
	if err != nil {
		slog.Error("response adaptation failed",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", "execution_error",
		)
		return "", fmt.Errorf("response adaptation error for tool %s: %w", toolName, err)
	}

	output, err := serializeResponse(response)
	if err != nil {
		slog.Error("response serialization failed",
			"tool_name", toolName,
			"error", err.Error(),
			"error_type", "parse_error",
		)
		return "", err
	}

	return output, nil
}

func (e *ToolExecutor) buildCommand(toolName string, args map[string]interface{}, scope string, outputFormat string) []string {
	builder, ok := toolCommandBuilders[toolName]
	if !ok {
		return nil
	}

	cmdArgs := make([]string, 0, 8)
	cmdArgs = append(cmdArgs, scopeArgs(scope)...)
	cmdArgs = append(cmdArgs, builder(args)...)

	if len(cmdArgs) == 0 {
		return nil
	}

	cmdArgs = append(cmdArgs, "--output", outputFormat)
	return cmdArgs
}

type commandBuilder func(args map[string]interface{}) []string

var toolCommandBuilders = map[string]commandBuilder{
	"query_tools":              buildQueryToolsCommand,
	"query_user_messages":      buildQueryUserMessagesCommand,
	"get_session_stats":        buildGetSessionStatsCommand,
	"query_context":            buildQueryContextCommand,
	"query_tool_sequences":     buildQueryToolSequencesCommand,
	"query_file_access":        buildQueryFileAccessCommand,
	"query_project_state":      buildQueryProjectStateCommand,
	"query_successful_prompts": buildQuerySuccessfulPromptsCommand,
	"query_tools_advanced":     buildQueryToolsAdvancedCommand,
	"query_time_series":        buildQueryTimeSeriesCommand,
	"query_assistant_messages": buildQueryAssistantMessagesCommand,
	"query_conversation":       buildQueryConversationCommand,
	"query_files":              buildQueryFilesCommand,
}

func scopeArgs(scope string) []string {
	switch scope {
	case "project":
		return []string{"--project", "."}
	case "session":
		return []string{"--session-only"}
	default:
		return nil
	}
}

func buildQueryToolsCommand(args map[string]interface{}) []string {
	cmd := []string{"query", "tools"}
	if tool := getStringParam(args, "tool", ""); tool != "" {
		cmd = append(cmd, "--tool", tool)
	}
	if status := getStringParam(args, "status", ""); status != "" {
		cmd = append(cmd, "--status", status)
	}
	if limit := getIntParam(args, "limit", 0); limit > 0 {
		cmd = append(cmd, "--limit", strconv.Itoa(limit))
	}
	return cmd
}

func buildQueryUserMessagesCommand(args map[string]interface{}) []string {
	cmd := []string{"query", "user-messages"}
	if pattern := getStringParam(args, "pattern", ""); pattern != "" {
		cmd = append(cmd, "--pattern", pattern)
	}
	if limit := getIntParam(args, "limit", 0); limit > 0 {
		cmd = append(cmd, "--limit", strconv.Itoa(limit))
	}
	return cmd
}

func buildGetSessionStatsCommand(args map[string]interface{}) []string {
	return []string{"parse", "stats"}
}

func buildQueryContextCommand(args map[string]interface{}) []string {
	cmd := []string{"query", "context"}
	if errorSig := getStringParam(args, "error_signature", ""); errorSig != "" {
		cmd = append(cmd, "--error-signature", errorSig)
	}
	if window := getIntParam(args, "window", 0); window > 0 {
		cmd = append(cmd, "--window", strconv.Itoa(window))
	}
	return cmd
}

func buildQueryToolSequencesCommand(args map[string]interface{}) []string {
	cmd := []string{"analyze", "sequences"}
	if pattern := getStringParam(args, "pattern", ""); pattern != "" {
		cmd = append(cmd, "--pattern", pattern)
	}
	if minOccur := getIntParam(args, "min_occurrences", 0); minOccur > 0 {
		cmd = append(cmd, "--min-occurrences", strconv.Itoa(minOccur))
	}
	if includeBuiltin := getBoolParam(args, "include_builtin_tools", false); includeBuiltin {
		cmd = append(cmd, "--include-builtin-tools")
	}
	return cmd
}

func buildQueryFileAccessCommand(args map[string]interface{}) []string {
	cmd := []string{"query", "file-access"}
	if file := getStringParam(args, "file", ""); file != "" {
		cmd = append(cmd, "--file", file)
	}
	return cmd
}

func buildQueryProjectStateCommand(args map[string]interface{}) []string {
	return []string{"query", "project-state"}
}

func buildQuerySuccessfulPromptsCommand(args map[string]interface{}) []string {
	cmd := []string{"query", "successful-prompts"}
	if limit := getIntParam(args, "limit", 0); limit > 0 {
		cmd = append(cmd, "--limit", strconv.Itoa(limit))
	}
	if minQuality := getFloatParam(args, "min_quality_score", 0); minQuality > 0 {
		cmd = append(cmd, "--min-quality-score", fmt.Sprintf("%.2f", minQuality))
	}
	return cmd
}

func buildQueryToolsAdvancedCommand(args map[string]interface{}) []string {
	cmd := []string{"query", "tools"}
	if where := getStringParam(args, "where", ""); where != "" {
		cmd = append(cmd, "--where", where)
	}
	if limit := getIntParam(args, "limit", 0); limit > 0 {
		cmd = append(cmd, "--limit", strconv.Itoa(limit))
	}
	return cmd
}

func buildQueryTimeSeriesCommand(args map[string]interface{}) []string {
	cmd := []string{"stats", "timeseries"}
	if interval := getStringParam(args, "interval", ""); interval != "" {
		cmd = append(cmd, "--interval", interval)
	}
	if metric := getStringParam(args, "metric", ""); metric != "" {
		cmd = append(cmd, "--metric", metric)
	}
	if where := getStringParam(args, "where", ""); where != "" {
		cmd = append(cmd, "--where", where)
	}
	return cmd
}

func buildQueryAssistantMessagesCommand(args map[string]interface{}) []string {
	cmd := []string{"query", "assistant-messages"}
	if pattern := getStringParam(args, "pattern", ""); pattern != "" {
		cmd = append(cmd, "--pattern", pattern)
	}
	if minTools := getIntParam(args, "min_tools", 0); minTools > 0 {
		cmd = append(cmd, "--min-tools", strconv.Itoa(minTools))
	}
	if maxTools := getIntParam(args, "max_tools", 0); maxTools > 0 {
		cmd = append(cmd, "--max-tools", strconv.Itoa(maxTools))
	}
	if minTokens := getIntParam(args, "min_tokens_output", 0); minTokens > 0 {
		cmd = append(cmd, "--min-tokens-output", strconv.Itoa(minTokens))
	}
	if minLength := getIntParam(args, "min_length", 0); minLength > 0 {
		cmd = append(cmd, "--min-length", strconv.Itoa(minLength))
	}
	if maxLength := getIntParam(args, "max_length", 0); maxLength > 0 {
		cmd = append(cmd, "--max-length", strconv.Itoa(maxLength))
	}
	if limit := getIntParam(args, "limit", 0); limit > 0 {
		cmd = append(cmd, "--limit", strconv.Itoa(limit))
	}
	return cmd
}

func buildQueryConversationCommand(args map[string]interface{}) []string {
	cmd := []string{"query", "conversation"}
	if startTurn := getIntParam(args, "start_turn", 0); startTurn > 0 {
		cmd = append(cmd, "--start-turn", strconv.Itoa(startTurn))
	}
	if endTurn := getIntParam(args, "end_turn", 0); endTurn > 0 {
		cmd = append(cmd, "--end-turn", strconv.Itoa(endTurn))
	}
	if pattern := getStringParam(args, "pattern", ""); pattern != "" {
		cmd = append(cmd, "--pattern", pattern)
	}
	if target := getStringParam(args, "pattern_target", ""); target != "" {
		cmd = append(cmd, "--pattern-target", target)
	}
	if minDuration := getIntParam(args, "min_duration", 0); minDuration > 0 {
		cmd = append(cmd, "--min-duration", strconv.Itoa(minDuration))
	}
	if maxDuration := getIntParam(args, "max_duration", 0); maxDuration > 0 {
		cmd = append(cmd, "--max-duration", strconv.Itoa(maxDuration))
	}
	if limit := getIntParam(args, "limit", 0); limit > 0 {
		cmd = append(cmd, "--limit", strconv.Itoa(limit))
	}
	return cmd
}

func buildQueryFilesCommand(args map[string]interface{}) []string {
	cmd := []string{"analyze", "file-churn"}
	if threshold := getIntParam(args, "threshold", 0); threshold > 0 && threshold != 5 {
		cmd = append(cmd, "--threshold", strconv.Itoa(threshold))
	}
	return cmd
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
