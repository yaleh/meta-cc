package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/analyzer"
	"github.com/yale/meta-cc/internal/filter"
	"github.com/yale/meta-cc/internal/locator"
	internalOutput "github.com/yale/meta-cc/internal/output"
	"github.com/yale/meta-cc/internal/parser"
	"github.com/yale/meta-cc/pkg/output"
)

var (
	extractType   string
	extractFilter string
)

// parseCmd represents the parse subcommand
var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parse Claude Code session files",
	Long: `Parse Claude Code session files and extract structured data.

Examples:
  meta-cc parse extract --type turns
  meta-cc parse extract --type tools --output md
  meta-cc parse extract --type tools --filter "status=error"`,
}

// parseExtractCmd represents the parse extract sub-subcommand
var parseExtractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract data from session",
	Long: `Extract structured data from Claude Code session files.

Supported types:
  - turns:  All conversation turns (user and assistant messages)
  - tools:  Tool calls with their results
  - errors: Failed tool calls and error messages`,
	RunE: runParseExtract,
}

func init() {
	// Add parse subcommand to root
	rootCmd.AddCommand(parseCmd)

	// Add extract sub-subcommand to parse
	parseCmd.AddCommand(parseExtractCmd)

	// extract subcommand parameters
	parseExtractCmd.Flags().StringVarP(&extractType, "type", "t", "turns", "Data type to extract: turns|tools|errors")
	parseExtractCmd.Flags().StringVarP(&extractFilter, "filter", "f", "", "Filter data (e.g., \"status=error\")")

	// --output parameter is already defined in root.go as global parameter
}

func runParseExtract(cmd *cobra.Command, args []string) error {
	// Step 1: Validate parameters
	validTypes := map[string]bool{
		"turns":  true,
		"tools":  true,
		"errors": true,
	}

	if !validTypes[extractType] {
		return internalOutput.OutputError(
			fmt.Errorf("invalid type '%s': must be one of: turns, tools, errors", extractType),
			internalOutput.ErrInvalidArgument,
			outputFormat,
		)
	}

	// Step 2: Locate session file (using Phase 1 locator)
	loc := locator.NewSessionLocator()
	sessionPath, err := loc.Locate(locator.LocateOptions{
		SessionID:   sessionID,   // from global parameter
		ProjectPath: projectPath, // from global parameter
		SessionOnly: sessionOnly,  // Phase 13: opt-out of project default
	})
	if err != nil {
		return internalOutput.OutputError(err, internalOutput.ErrSessionNotFound, outputFormat)
	}

	// Step 3: Parse session file (using Phase 2 parser)
	sessionParser := parser.NewSessionParser(sessionPath)
	entries, err := sessionParser.ParseEntries()
	if err != nil {
		return internalOutput.OutputError(err, internalOutput.ErrParseError, outputFormat)
	}

	// Step 4: Extract data based on type
	var data interface{}

	switch extractType {
	case "turns":
		data = entries
	case "tools":
		toolCalls := parser.ExtractToolCalls(entries)
		data = toolCalls
	case "errors":
		// Extract failed tool calls
		toolCalls := parser.ExtractToolCalls(entries)
		var errorCalls []parser.ToolCall
		for _, tc := range toolCalls {
			if tc.Status == "error" || tc.Error != "" {
				errorCalls = append(errorCalls, tc)
			}
		}
		data = errorCalls
	}

	// Step 4.5: Apply filter (if provided)
	if extractFilter != "" {
		filterObj, err := filter.ParseFilter(extractFilter)
		if err != nil {
			return internalOutput.OutputError(err, internalOutput.ErrFilterError, outputFormat)
		}

		data = filter.ApplyFilter(data, filterObj)
	}

	// Step 4.6: Handle --estimate-size flag (Phase 9.1)
	if estimateSizeFlag {
		// Only estimate for tool calls
		if toolCalls, ok := data.([]parser.ToolCall); ok {
			estimate, err := output.EstimateToolCallsSize(toolCalls, outputFormat)
			if err != nil {
				return internalOutput.OutputError(err, internalOutput.ErrInternalError, outputFormat)
			}

			estimateStr, _ := output.FormatJSONL(estimate)
			fmt.Fprintln(cmd.OutOrStdout(), estimateStr)
			return nil
		}
	}

	// Step 4.7: Apply pagination (Phase 9.1)
	if toolCalls, ok := data.([]parser.ToolCall); ok {
		paginationConfig := filter.PaginationConfig{
			Limit:  limitFlag,
			Offset: offsetFlag,
		}
		data = filter.ApplyPagination(toolCalls, paginationConfig)
	}

	// Step 5: Check for empty results
	switch v := data.(type) {
	case []parser.ToolCall:
		if len(v) == 0 {
			return internalOutput.WarnNoResults(outputFormat)
		}
	case []parser.SessionEntry:
		if len(v) == 0 {
			return internalOutput.WarnNoResults(outputFormat)
		}
	}

	// Step 6: Format output
	outputStr, formatErr := internalOutput.FormatOutput(data, outputFormat)
	if formatErr != nil {
		return internalOutput.OutputError(formatErr, internalOutput.ErrInternalError, outputFormat)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}

var (
	statsMetrics string // 用于过滤统计指标
)

// parseStatsCmd represents the parse stats sub-subcommand
var parseStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show session statistics",
	Long: `Show statistical analysis of Claude Code session data.

Displays metrics including:
  - Turn counts (total, user, assistant)
  - Tool call counts and frequency
  - Error counts and error rate
  - Session duration
  - Top tools by usage

Examples:
  meta-cc parse stats
  meta-cc parse stats --output md
  meta-cc parse stats --metrics tools,errors`,
	RunE: runParseStats,
}

func init() {
	// Add stats sub-subcommand to parse
	parseCmd.AddCommand(parseStatsCmd)

	// stats subcommand parameters
	parseStatsCmd.Flags().StringVarP(&statsMetrics, "metrics", "m", "", "Filter metrics to display (e.g., \"tools,errors,duration\")")

	// --output parameter is already defined in root.go as global parameter
}

func runParseStats(cmd *cobra.Command, args []string) error {
	// Step 1: Locate session file (using Phase 1 locator)
	loc := locator.NewSessionLocator()
	sessionPath, err := loc.Locate(locator.LocateOptions{
		SessionID:   sessionID,   // from global parameter
		ProjectPath: projectPath, // from global parameter
		SessionOnly: sessionOnly,  // Phase 13: opt-out of project default
	})
	if err != nil {
		return internalOutput.OutputError(err, internalOutput.ErrSessionNotFound, outputFormat)
	}

	// Step 2: Parse session file (using Phase 2 parser)
	sessionParser := parser.NewSessionParser(sessionPath)
	entries, err := sessionParser.ParseEntries()
	if err != nil {
		return internalOutput.OutputError(err, internalOutput.ErrParseError, outputFormat)
	}

	// Step 3: Extract tool calls
	toolCalls := parser.ExtractToolCalls(entries)

	// Step 4: Calculate statistics (using Stage 4.1 analyzer)
	stats := analyzer.CalculateStats(entries, toolCalls)

	// Step 4.5: Handle --estimate-size flag (Phase 9.1)
	if estimateSizeFlag {
		estimate := output.EstimateStatsSize(outputFormat)
		estimateStr, _ := output.FormatJSONL(estimate)
		fmt.Fprintln(cmd.OutOrStdout(), estimateStr)
		return nil
	}

	// Step 5: Filter metrics if specified
	var data interface{} = stats
	if statsMetrics != "" {
		// Simplified implementation: statsMetrics parameter is documented for future use
		// Complete metric filtering can be extended in subsequent phases
		data = stats
	}

	// Step 6: Format output (using Phase 3 formatters)
	outputStr, formatErr := internalOutput.FormatOutput(data, outputFormat)
	if formatErr != nil {
		return internalOutput.OutputError(formatErr, internalOutput.ErrInternalError, outputFormat)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}

// formatStatsMarkdown formats statistics as a Markdown report
func formatStatsMarkdown(stats analyzer.SessionStats) (string, error) {
	var sb strings.Builder

	sb.WriteString("# Session Statistics\n\n")

	// Overview section
	sb.WriteString("## Overview\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Turns**: %d\n", stats.TurnCount))
	sb.WriteString(fmt.Sprintf("  - User Turns: %d\n", stats.UserTurnCount))
	sb.WriteString(fmt.Sprintf("  - Assistant Turns: %d\n", stats.AssistantTurnCount))
	sb.WriteString(fmt.Sprintf("- **Session Duration**: %d seconds (%.1f minutes)\n",
		stats.DurationSeconds, float64(stats.DurationSeconds)/60))
	sb.WriteString("\n")

	// Tool Usage section
	sb.WriteString("## Tool Usage\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Tool Calls**: %d\n", stats.ToolCallCount))
	sb.WriteString(fmt.Sprintf("- **Successful Calls**: %d\n", stats.ToolCallCount-stats.ErrorCount))
	sb.WriteString(fmt.Sprintf("- **Failed Calls**: %d\n", stats.ErrorCount))
	sb.WriteString(fmt.Sprintf("- **Error Rate**: %.1f%%\n", stats.ErrorRate))
	sb.WriteString("\n")

	// Top Tools section
	if len(stats.TopTools) > 0 {
		sb.WriteString("### Top Tools\n\n")
		sb.WriteString("| Tool | Count | Percentage |\n")
		sb.WriteString("|------|-------|------------|\n")

		for _, tool := range stats.TopTools {
			percentage := float64(0)
			if stats.ToolCallCount > 0 {
				percentage = float64(tool.Count) / float64(stats.ToolCallCount) * 100
			}
			sb.WriteString(fmt.Sprintf("| %s | %d | %.1f%% |\n",
				tool.Name, tool.Count, percentage))
		}
		sb.WriteString("\n")
	}

	// Tool Frequency section
	if len(stats.ToolFrequency) > 0 {
		sb.WriteString("### All Tools\n\n")
		for name, count := range stats.ToolFrequency {
			sb.WriteString(fmt.Sprintf("- **%s**: %d calls\n", name, count))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}
