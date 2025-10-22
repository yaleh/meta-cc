package cmd

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"github.com/yaleh/meta-cc/internal/filter"
	internalOutput "github.com/yaleh/meta-cc/internal/output"
	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/pkg/output"
)

var (
	queryToolsStatus string
	queryToolsTool   string
	queryToolsWhere  string
	queryToolsFilter string
)

var queryToolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Query tool calls",
	Long: `Query tool calls with advanced filtering options.

Supports filtering by:
  - Tool name (--tool)
  - Status (--status: success|error)
  - General condition (--where: "field=value")
  - Advanced expressions (--filter: SQL-like syntax)

Examples:
  meta-cc query tools --status error
  meta-cc query tools --tool Bash --limit 20
  meta-cc query tools --where "status=error" --sort-by timestamp
  meta-cc query tools --filter "tool='Bash' AND status='error'"
  meta-cc query tools --filter "tool IN ('Bash', 'Edit') OR duration>1000"`,
	RunE: runQueryTools,
}

func init() {
	queryCmd.AddCommand(queryToolsCmd)

	// Tool-specific filters
	queryToolsCmd.Flags().StringVar(&queryToolsStatus, "status", "", "Filter by status (success|error)")
	queryToolsCmd.Flags().StringVar(&queryToolsTool, "tool", "", "Filter by tool name")
	queryToolsCmd.Flags().StringVar(&queryToolsWhere, "where", "", "Filter condition (key=value)")
	queryToolsCmd.Flags().StringVar(&queryToolsFilter, "filter", "", "Advanced filter expression (SQL-like)")
}

func runQueryTools(cmd *cobra.Command, args []string) error {
	calls, err := loadToolCalls(getGlobalOptions())
	if err != nil {
		return err
	}

	calls, err = filterToolCalls(calls)
	if err != nil {
		return err
	}
	sortToolCallList(calls)

	if handled, err := handleToolEstimate(cmd, calls); handled || err != nil {
		return err
	}

	calls = paginateToolCalls(calls)

	if handled, err := handleToolChunking(cmd, calls); handled || err != nil {
		return err
	}

	if len(calls) == 0 {
		return internalOutput.WarnNoResults(outputFormat)
	}

	if handled, err := handleToolSummaryFirst(cmd, calls); handled || err != nil {
		return err
	}

	if handled, err := handleToolStreaming(cmd, calls); handled || err != nil {
		return err
	}

	if handled, err := handleToolProjection(cmd, calls); handled || err != nil {
		return err
	}

	return writeToolCalls(cmd, calls)
}

func loadToolCalls(opts GlobalOptions) ([]parser.ToolCall, error) {
	pipeline := NewSessionPipeline(opts)
	if err := pipeline.Load(LoadOptions{AutoDetect: true}); err != nil {
		return nil, internalOutput.OutputError(err, internalOutput.ErrSessionNotFound, outputFormat)
	}
	return pipeline.ExtractToolCalls(), nil
}

func filterToolCalls(calls []parser.ToolCall) ([]parser.ToolCall, error) {
	filtered, err := applyToolFilters(calls)
	if err != nil {
		return nil, internalOutput.OutputError(err, internalOutput.ErrFilterError, outputFormat)
	}
	return filtered, nil
}

func sortToolCallList(calls []parser.ToolCall) {
	output.SortByTimestamp(calls)
	if querySortBy != "" {
		sortToolCalls(calls, querySortBy, queryReverse)
	}
}

func handleToolEstimate(cmd *cobra.Command, calls []parser.ToolCall) (bool, error) {
	if !estimateSizeFlag {
		return false, nil
	}

	estimate, err := output.EstimateToolCallsSize(calls, outputFormat)
	if err != nil {
		return true, internalOutput.OutputError(err, internalOutput.ErrInternalError, outputFormat)
	}

	estimateStr, _ := output.FormatJSONL(estimate)
	fmt.Fprintln(cmd.OutOrStdout(), estimateStr)
	return true, nil
}

func paginateToolCalls(calls []parser.ToolCall) []parser.ToolCall {
	limit := limitFlag
	offset := offsetFlag
	if limit == 0 && queryLimit > 0 {
		limit = queryLimit
	}
	if offset == 0 && queryOffset > 0 {
		offset = queryOffset
	}

	config := filter.PaginationConfig{Limit: limit, Offset: offset}
	return filter.ApplyPagination(calls, config)
}

func handleToolChunking(cmd *cobra.Command, calls []parser.ToolCall) (bool, error) {
	if chunkSizeFlag <= 0 {
		return false, nil
	}

	if outputDirFlag == "" {
		err := fmt.Errorf("--output-dir is required when using --chunk-size")
		return true, internalOutput.OutputError(err, internalOutput.ErrInvalidArgument, outputFormat)
	}

	metadata, err := output.ChunkToolCalls(calls, chunkSizeFlag, outputDirFlag, outputFormat)
	if err != nil {
		return true, internalOutput.OutputError(err, internalOutput.ErrInternalError, outputFormat)
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "Generated %d chunk(s)\n", len(metadata))
	for _, meta := range metadata {
		fmt.Fprintf(cmd.ErrOrStderr(), "  Chunk %d: %s (%d records, %d bytes)\n",
			meta.Index, filepath.Base(meta.File), meta.Records, meta.SizeBytes)
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Manifest: %s\n", filepath.Join(outputDirFlag, "manifest.json"))
	return true, nil
}

func handleToolSummaryFirst(cmd *cobra.Command, calls []parser.ToolCall) (bool, error) {
	if !summaryFirstFlag {
		return false, nil
	}

	summaryOutput, err := output.FormatSummaryFirst(calls, topNFlag, outputFormat)
	if err != nil {
		return true, internalOutput.OutputError(err, internalOutput.ErrInternalError, outputFormat)
	}

	fmt.Fprintln(cmd.OutOrStdout(), summaryOutput.Summary)
	fmt.Fprintln(cmd.OutOrStdout(), summaryOutput.Details)
	return true, nil
}

func handleToolStreaming(cmd *cobra.Command, calls []parser.ToolCall) (bool, error) {
	if !queryStream {
		return false, nil
	}

	streamWriter := output.NewStreamWriter(cmd.OutOrStdout())
	for _, tool := range calls {
		if err := streamWriter.WriteRecord(tool); err != nil {
			return true, internalOutput.OutputError(err, internalOutput.ErrInternalError, outputFormat)
		}
	}
	return true, nil
}

func handleToolProjection(cmd *cobra.Command, calls []parser.ToolCall) (bool, error) {
	projectionConfig := output.ParseProjectionConfig(fieldsFlag, ifErrorIncludeFlag)
	if len(projectionConfig.Fields) == 0 {
		return false, nil
	}

	projected, err := output.ProjectToolCalls(calls, projectionConfig)
	if err != nil {
		return true, internalOutput.OutputError(err, internalOutput.ErrInternalError, outputFormat)
	}

	outputStr, formatErr := output.FormatProjectedOutput(projected, outputFormat)
	if formatErr != nil {
		return true, internalOutput.OutputError(formatErr, internalOutput.ErrInternalError, outputFormat)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return true, nil
}

func writeToolCalls(cmd *cobra.Command, calls []parser.ToolCall) error {
	outputStr, err := internalOutput.FormatOutput(calls, outputFormat)
	if err != nil {
		return internalOutput.OutputError(err, internalOutput.ErrInternalError, outputFormat)
	}
	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}

func applyToolFilters(toolCalls []parser.ToolCall) ([]parser.ToolCall, error) {
	filtered, err := applyExpressionFilter(toolCalls)
	if err != nil {
		return nil, err
	}

	filtered, err = applyWhereFilter(filtered)
	if err != nil {
		return nil, err
	}

	return applyFlagFilters(filtered), nil
}

func applyExpressionFilter(toolCalls []parser.ToolCall) ([]parser.ToolCall, error) {
	if queryToolsFilter == "" {
		return toolCalls, nil
	}

	expr, err := filter.ParseExpression(queryToolsFilter)
	if err != nil {
		return nil, fmt.Errorf("invalid filter expression: %w", err)
	}

	var filtered []parser.ToolCall
	for _, tc := range toolCalls {
		record := map[string]interface{}{
			"tool":   tc.ToolName,
			"status": tc.Status,
			"uuid":   tc.UUID,
			"error":  tc.Error,
		}

		match, evalErr := expr.Evaluate(record)
		if evalErr != nil {
			return nil, fmt.Errorf("filter evaluation error: %w", evalErr)
		}

		if match {
			filtered = append(filtered, tc)
		}
	}

	return filtered, nil
}

func applyWhereFilter(toolCalls []parser.ToolCall) ([]parser.ToolCall, error) {
	if queryToolsWhere == "" {
		return toolCalls, nil
	}

	filtered, err := filter.ApplyWhere(toolCalls, queryToolsWhere, "tool_calls")
	if err != nil {
		return nil, fmt.Errorf("invalid where condition: %w", err)
	}

	return filtered.([]parser.ToolCall), nil
}

func applyFlagFilters(toolCalls []parser.ToolCall) []parser.ToolCall {
	var result []parser.ToolCall

	for _, tc := range toolCalls {
		if !matchesStatus(tc) {
			continue
		}
		if queryToolsTool != "" && tc.ToolName != queryToolsTool {
			continue
		}
		result = append(result, tc)
	}

	return result
}

func matchesStatus(tc parser.ToolCall) bool {
	if queryToolsStatus == "" {
		return true
	}

	switch queryToolsStatus {
	case "error":
		return tc.Status == "error" || tc.Error != ""
	case "success":
		return tc.Status != "error" && tc.Error == ""
	default:
		return true
	}
}

func sortToolCalls(toolCalls []parser.ToolCall, sortBy string, reverse bool) {
	// Use stable sort to preserve relative order for equal values
	sort.SliceStable(toolCalls, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "timestamp":
			less = toolCalls[i].Timestamp < toolCalls[j].Timestamp
		case "tool":
			less = toolCalls[i].ToolName < toolCalls[j].ToolName
		case "status":
			less = toolCalls[i].Status < toolCalls[j].Status
		case "uuid":
			less = toolCalls[i].UUID < toolCalls[j].UUID
		default:
			// Default: sort by timestamp (deterministic)
			less = toolCalls[i].Timestamp < toolCalls[j].Timestamp
		}

		if reverse {
			return !less
		}
		return less
	})
}
