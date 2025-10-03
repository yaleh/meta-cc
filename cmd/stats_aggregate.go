package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/filter"
	"github.com/yale/meta-cc/internal/locator"
	"github.com/yale/meta-cc/internal/parser"
	"github.com/yale/meta-cc/internal/stats"
	"github.com/yale/meta-cc/pkg/output"
)

var (
	aggregateGroupBy string
	aggregateMetrics string
	aggregateFilter  string
)

// statsCmd represents the stats parent command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Statistical analysis of session data",
	Long:  `Perform statistical analysis and aggregation on Claude Code session data.`,
}

// statsAggregateCmd represents the stats aggregate subcommand
var statsAggregateCmd = &cobra.Command{
	Use:   "aggregate",
	Short: "Aggregate data with group-by and metrics",
	Long: `Aggregate tool call data by grouping and calculating metrics.

Supported group-by fields:
  - tool:   Group by tool name
  - status: Group by execution status
  - uuid:   Group by session entry UUID

Supported metrics:
  - count:      Number of records in group
  - error_rate: Percentage of errors (0.0-1.0)

Examples:
  # Error rate by tool
  meta-cc stats aggregate --group-by tool --metrics error_rate

  # Multiple metrics
  meta-cc stats aggregate --group-by tool --metrics "count,error_rate"

  # With filtering
  meta-cc stats aggregate --group-by status --metrics count --filter "tool='Bash'"`,
	RunE: runStatsAggregate,
}

func init() {
	// Add stats command to root
	rootCmd.AddCommand(statsCmd)

	// Add aggregate subcommand to stats
	statsCmd.AddCommand(statsAggregateCmd)

	// Flags for aggregate
	statsAggregateCmd.Flags().StringVar(&aggregateGroupBy, "group-by", "tool", "Field to group by (tool|status|uuid)")
	statsAggregateCmd.Flags().StringVar(&aggregateMetrics, "metrics", "count", "Metrics to calculate (comma-separated)")
	statsAggregateCmd.Flags().StringVar(&aggregateFilter, "filter", "", "Filter expression (SQL-like)")
}

func runStatsAggregate(cmd *cobra.Command, args []string) error {
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

	// Step 3: Apply filter if provided (using Stage 10.1 filter engine)
	if aggregateFilter != "" {
		expr, err := filter.ParseExpression(aggregateFilter)
		if err != nil {
			return fmt.Errorf("invalid filter: %w", err)
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
				return fmt.Errorf("filter evaluation error: %w", err)
			}

			if match {
				filtered = append(filtered, tc)
			}
		}
		toolCalls = filtered
	}

	// Step 4: Parse metrics
	metricsList := strings.Split(aggregateMetrics, ",")
	for i, m := range metricsList {
		metricsList[i] = strings.TrimSpace(m)
	}

	// Step 5: Perform aggregation
	config := stats.AggregateConfig{
		GroupBy: aggregateGroupBy,
		Metrics: metricsList,
	}

	results, err := stats.Aggregate(toolCalls, config)
	if err != nil {
		return fmt.Errorf("aggregation failed: %w", err)
	}

	// Step 6: Format output
	var outputStr string
	switch outputFormat {
	case "json":
		outputStr, err = output.FormatJSON(results)
	case "md", "markdown":
		outputStr, err = output.FormatMarkdown(results)
	case "csv":
		outputStr, err = output.FormatCSV(results)
	default:
		outputStr, err = output.FormatJSON(results)
	}

	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}
