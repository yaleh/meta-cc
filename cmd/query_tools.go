package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	internalOutput "github.com/yaleh/meta-cc/internal/output"
	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/internal/query"
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
	calls, err := query.RunToolsQuery(buildToolsQueryOptions(getGlobalOptions()))
	if err != nil {
		return handleToolsQueryError(err)
	}

	if handled, err := handleToolEstimate(cmd, calls); handled || err != nil {
		return err
	}

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

func buildToolsQueryOptions(globalOpts GlobalOptions) query.ToolsQueryOptions {
	limit := limitFlag
	if limit == 0 && queryLimit > 0 {
		limit = queryLimit
	}

	offset := offsetFlag
	if offset == 0 && queryOffset > 0 {
		offset = queryOffset
	}

	return query.ToolsQueryOptions{
		Pipeline:   toPipelineOptions(globalOpts),
		Limit:      limit,
		Offset:     offset,
		SortBy:     querySortBy,
		Reverse:    queryReverse,
		Status:     queryToolsStatus,
		Tool:       queryToolsTool,
		Where:      queryToolsWhere,
		Expression: queryToolsFilter,
	}
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

func handleToolsQueryError(err error) error {
	switch {
	case errors.Is(err, query.ErrSessionLoad):
		return internalOutput.OutputError(err, internalOutput.ErrSessionNotFound, outputFormat)
	case errors.Is(err, query.ErrFilterInvalid):
		return internalOutput.OutputError(err, internalOutput.ErrFilterError, outputFormat)
	default:
		return internalOutput.OutputError(err, internalOutput.ErrInternalError, outputFormat)
	}
}
