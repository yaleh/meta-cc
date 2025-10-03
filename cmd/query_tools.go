package cmd

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/filter"
	"github.com/yale/meta-cc/internal/locator"
	internalOutput "github.com/yale/meta-cc/internal/output"
	"github.com/yale/meta-cc/internal/parser"
	"github.com/yale/meta-cc/pkg/output"
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
	// Step 1: Locate and parse session
	loc := locator.NewSessionLocator()
	sessionPath, err := loc.Locate(locator.LocateOptions{
		SessionID:   sessionID,
		ProjectPath: projectPath,
	})
	if err != nil {
		return fmt.Errorf("failed to locate session: %w", err)
	}

	sessionParser := parser.NewSessionParser(sessionPath)
	entries, err := sessionParser.ParseEntries()
	if err != nil {
		return fmt.Errorf("failed to parse session: %w", err)
	}

	// Step 2: Extract tool calls
	toolCalls := parser.ExtractToolCalls(entries)

	// Step 3: Apply filters
	filtered, err := applyToolFilters(toolCalls)
	if err != nil {
		return err
	}

	// Step 4: Sort if requested
	if querySortBy != "" {
		sortToolCalls(filtered, querySortBy, queryReverse)
	}

	// Step 5: Handle --estimate-size flag
	if estimateSizeFlag {
		estimate, err := output.EstimateToolCallsSize(filtered, outputFormat)
		if err != nil {
			return fmt.Errorf("failed to estimate size: %w", err)
		}

		// Output estimate as JSON
		estimateStr, _ := output.FormatJSON(estimate)
		fmt.Fprintln(cmd.OutOrStdout(), estimateStr)
		return nil
	}

	// Step 6: Apply pagination using new filter package
	// Use global flags (limitFlag, offsetFlag) if set, otherwise fall back to queryLimit/queryOffset
	limit := limitFlag
	offset := offsetFlag
	if limit == 0 && queryLimit > 0 {
		limit = queryLimit
	}
	if offset == 0 && queryOffset > 0 {
		offset = queryOffset
	}

	paginationConfig := filter.PaginationConfig{
		Limit:  limit,
		Offset: offset,
	}
	paginated := filter.ApplyPagination(filtered, paginationConfig)

	// Step 7: Handle chunking mode
	if chunkSizeFlag > 0 {
		// Validate output directory is specified
		if outputDirFlag == "" {
			return fmt.Errorf("--output-dir is required when using --chunk-size")
		}

		// Create chunks
		metadata, err := output.ChunkToolCalls(paginated, chunkSizeFlag, outputDirFlag, outputFormat)
		if err != nil {
			return fmt.Errorf("failed to create chunks: %w", err)
		}

		// Output chunk summary to stderr (not stdout)
		fmt.Fprintf(cmd.ErrOrStderr(), "Generated %d chunk(s)\n", len(metadata))
		for _, meta := range metadata {
			fmt.Fprintf(cmd.ErrOrStderr(), "  Chunk %d: %s (%d records, %d bytes)\n",
				meta.Index, filepath.Base(meta.File), meta.Records, meta.SizeBytes)
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "Manifest: %s\n", filepath.Join(outputDirFlag, "manifest.json"))

		return nil
	}

	// Step 8: Handle summary-first mode
	if summaryFirstFlag {
		summaryOutput, err := output.FormatSummaryFirst(paginated, topNFlag, outputFormat)
		if err != nil {
			return fmt.Errorf("failed to generate summary: %w", err)
		}

		// Print summary followed by details
		fmt.Fprintln(cmd.OutOrStdout(), summaryOutput.Summary)
		fmt.Fprintln(cmd.OutOrStdout(), summaryOutput.Details)

		// Check for empty results
		if len(paginated) == 0 {
			return internalOutput.NewExitCodeError(internalOutput.ExitNoResults, "No results found")
		}

		return nil
	}

	// Step 9: Handle streaming output if requested
	if queryStream {
		streamWriter := output.NewStreamWriter(cmd.OutOrStdout())
		for _, tool := range paginated {
			if err := streamWriter.WriteRecord(tool); err != nil {
				return fmt.Errorf("stream write error: %w", err)
			}
		}
		// Check for empty results
		if len(paginated) == 0 {
			return internalOutput.NewExitCodeError(internalOutput.ExitNoResults, "No results found")
		}
		return nil
	}

	// Step 10: Apply field projection if requested
	projectionConfig := output.ParseProjectionConfig(fieldsFlag, ifErrorIncludeFlag)

	// If projection is requested, project the fields
	if len(projectionConfig.Fields) > 0 {
		projected, err := output.ProjectToolCalls(paginated, projectionConfig)
		if err != nil {
			return fmt.Errorf("failed to project fields: %w", err)
		}

		// Format projected output
		outputStr, formatErr := output.FormatProjectedOutput(projected, outputFormat)
		if formatErr != nil {
			return fmt.Errorf("failed to format projected output: %w", formatErr)
		}

		fmt.Fprintln(cmd.OutOrStdout(), outputStr)

		// Check for empty results
		if len(paginated) == 0 {
			return internalOutput.NewExitCodeError(internalOutput.ExitNoResults, "No results found")
		}

		return nil
	}

	// Step 11: Format output (non-chunked, non-projected, non-summary mode)
	var outputStr string
	var formatErr error

	switch outputFormat {
	case "json":
		outputStr, formatErr = output.FormatJSON(paginated)
	case "md", "markdown":
		outputStr, formatErr = output.FormatMarkdown(paginated)
	case "csv":
		outputStr, formatErr = output.FormatCSV(paginated)
	case "tsv":
		outputStr = output.FormatTSV(paginated)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	if formatErr != nil {
		return fmt.Errorf("failed to format output: %w", formatErr)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)

	// Check for empty results and return appropriate exit code
	if len(paginated) == 0 {
		return internalOutput.NewExitCodeError(internalOutput.ExitNoResults, "No results found")
	}

	return nil
}

func applyToolFilters(toolCalls []parser.ToolCall) ([]parser.ToolCall, error) {
	// Apply --filter expression first (most powerful)
	if queryToolsFilter != "" {
		expr, err := filter.ParseExpression(queryToolsFilter)
		if err != nil {
			return nil, fmt.Errorf("invalid filter expression: %w", err)
		}

		var filtered []parser.ToolCall
		for _, tc := range toolCalls {
			// Convert ToolCall to map for expression evaluation
			record := map[string]interface{}{
				"tool":   tc.ToolName,
				"status": tc.Status,
				"uuid":   tc.UUID,
				"error":  tc.Error,
			}

			match, err := expr.Evaluate(record)
			if err != nil {
				return nil, fmt.Errorf("filter evaluation error: %w", err)
			}

			if match {
				filtered = append(filtered, tc)
			}
		}
		toolCalls = filtered
	}

	// Apply --where filter (simple key=value pairs)
	if queryToolsWhere != "" {
		filtered, err := filter.ApplyWhere(toolCalls, queryToolsWhere, "tool_calls")
		if err != nil {
			return nil, fmt.Errorf("invalid where condition: %w", err)
		}
		toolCalls = filtered.([]parser.ToolCall)
	}

	// Apply individual flag filters (for backwards compatibility)
	var result []parser.ToolCall
	for _, tc := range toolCalls {
		// Apply status filter
		if queryToolsStatus != "" {
			if queryToolsStatus == "error" {
				if tc.Status != "error" && tc.Error == "" {
					continue
				}
			} else if queryToolsStatus == "success" {
				if tc.Status == "error" || tc.Error != "" {
					continue
				}
			}
		}

		// Apply tool name filter
		if queryToolsTool != "" && tc.ToolName != queryToolsTool {
			continue
		}

		// If all filters pass, include this tool call
		result = append(result, tc)
	}

	return result, nil
}

func sortToolCalls(toolCalls []parser.ToolCall, sortBy string, reverse bool) {
	sort.Slice(toolCalls, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "tool":
			less = toolCalls[i].ToolName < toolCalls[j].ToolName
		case "status":
			less = toolCalls[i].Status < toolCalls[j].Status
		case "uuid":
			less = toolCalls[i].UUID < toolCalls[j].UUID
		default:
			// Default: sort by tool name
			less = toolCalls[i].ToolName < toolCalls[j].ToolName
		}

		if reverse {
			return !less
		}
		return less
	})
}
