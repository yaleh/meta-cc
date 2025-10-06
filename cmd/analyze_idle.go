package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/analyzer"
	"github.com/yale/meta-cc/pkg/output"
)

var idleThresholdMin int

// analyzeIdleCmd represents the analyze idle-periods subcommand
var analyzeIdleCmd = &cobra.Command{
	Use:   "idle-periods",
	Short: "Detect idle periods in session",
	Long: `Detect idle periods (time gaps) in Claude Code session data.

Identifies periods where there was no activity for an extended duration.
Useful for detecting points where users may have been stuck or distracted.

Examples:
  meta-cc analyze idle-periods
  meta-cc analyze idle-periods --threshold 5
  meta-cc analyze idle-periods --threshold 10 --output json`,
	RunE: runAnalyzeIdle,
}

func init() {
	analyzeCmd.AddCommand(analyzeIdleCmd)

	analyzeIdleCmd.Flags().IntVar(&idleThresholdMin, "threshold", 5, "Minimum idle duration in minutes")
}

func runAnalyzeIdle(cmd *cobra.Command, args []string) error {
	// Step 1: Initialize and load session using pipeline
	p := NewSessionPipeline(getGlobalOptions())
	if err := p.Load(LoadOptions{AutoDetect: true}); err != nil {
		return fmt.Errorf("failed to locate session file: %w", err)
	}

	// Step 2: Detect idle periods
	entries := p.GetEntries()
	result := analyzer.DetectIdlePeriods(entries, idleThresholdMin)

	// Step 4: Format and output
	var outputStr string
	var formatErr error

	switch outputFormat {
	case "jsonl":
		outputStr, formatErr = output.FormatJSONL(result)
	case "tsv":
		outputStr, formatErr = output.FormatTSV(result)
	default:
		return fmt.Errorf("unsupported output format: %s (supported: jsonl, tsv)", outputFormat)
	}

	if formatErr != nil {
		return fmt.Errorf("failed to format output: %w", formatErr)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}
