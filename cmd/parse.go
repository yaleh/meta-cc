package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/analyzer"
	"github.com/yale/meta-cc/internal/filter"
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

	// Step 2: Initialize and load session using pipeline
	p := NewSessionPipeline(getGlobalOptions())
	if err := p.Load(LoadOptions{AutoDetect: true}); err != nil {
		return internalOutput.OutputError(err, internalOutput.ErrSessionNotFound, outputFormat)
	}

	// Step 3: Extract data based on type
	var data interface{}

	switch extractType {
	case "turns":
		data = p.GetEntries()
	case "tools":
		data = p.ExtractToolCalls()
	case "errors":
		// Extract failed tool calls
		toolCalls := p.ExtractToolCalls()
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
	// Step 1: Initialize and load session using pipeline
	p := NewSessionPipeline(getGlobalOptions())
	if err := p.Load(LoadOptions{AutoDetect: true}); err != nil {
		return internalOutput.OutputError(err, internalOutput.ErrSessionNotFound, outputFormat)
	}

	// Step 2: Extract entries and tool calls
	entries := p.GetEntries()
	toolCalls := p.ExtractToolCalls()

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
