package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	mcerrors "github.com/yaleh/meta-cc/internal/errors"
	internalOutput "github.com/yaleh/meta-cc/internal/output"
	"github.com/yaleh/meta-cc/internal/query"
	"github.com/yaleh/meta-cc/pkg/output"
)

var (
	timeSeriesMetric   string
	timeSeriesInterval string
	timeSeriesFilter   string
)

// statsTimeSeriesCmd represents the stats time-series subcommand
var statsTimeSeriesCmd = &cobra.Command{
	Use:   "time-series",
	Short: "Analyze metrics over time",
	Long: `Generate time series data for specified metrics.

Supported metrics:
  - tool-calls: Count of tool calls per time bucket
  - error-rate: Percentage of errors per time bucket (0.0-1.0)

Supported intervals:
  - hour: Bucket by hour
  - day:  Bucket by day
  - week: Bucket by week (ISO week, Monday start)

Examples:
  # Tool calls per hour
  meta-cc stats time-series --metric tool-calls --interval hour

  # Error rate per day
  meta-cc stats time-series --metric error-rate --interval day

  # Tool calls per week with filtering
  meta-cc stats time-series --metric tool-calls --interval week --filter "tool='Bash'"`,
	RunE: runStatsTimeSeries,
}

func init() {
	// Add time-series subcommand to stats
	statsCmd.AddCommand(statsTimeSeriesCmd)

	// Flags for time-series
	statsTimeSeriesCmd.Flags().StringVar(&timeSeriesMetric, "metric", "tool-calls", "Metric to analyze (tool-calls|error-rate)")
	statsTimeSeriesCmd.Flags().StringVar(&timeSeriesInterval, "interval", "hour", "Time interval (hour|day|week)")
	statsTimeSeriesCmd.Flags().StringVar(&timeSeriesFilter, "filter", "", "Filter expression (SQL-like)")
}

func runStatsTimeSeries(cmd *cobra.Command, args []string) error {
	// Step 1: Initialize and load session using pipeline
	p := NewSessionPipeline(getGlobalOptions())
	if err := p.Load(LoadOptions{AutoDetect: true}); err != nil {
		return fmt.Errorf("failed to locate session: %w", err)
	}

	// Step 2: Extract tool calls
	toolCalls := p.ExtractToolCalls()

	// Step 3: Perform time series analysis via shared query helper
	points, err := query.AnalyzeTimeSeries(toolCalls, timeSeriesMetric, timeSeriesInterval, timeSeriesFilter)
	if err != nil {
		return fmt.Errorf("time series analysis failed: %w", err)
	}

	// Step 5: Format output
	var outputStr string
	switch outputFormat {
	case "jsonl":
		outputStr, err = output.FormatJSONL(points)
	case "tsv":
		outputStr, err = output.FormatTSV(points)
	default:
		return fmt.Errorf("unsupported output format: %s (supported: jsonl, tsv): %w", outputFormat, mcerrors.ErrInvalidInput)
	}

	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)

	// Check for empty results and return appropriate exit code
	if len(points) == 0 {
		return internalOutput.NewExitCodeError(internalOutput.ExitNoResults, "No results found")
	}

	return nil
}
