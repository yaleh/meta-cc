package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/analyzer"
	"github.com/yale/meta-cc/internal/locator"
	"github.com/yale/meta-cc/internal/parser"
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
	// Step 1: Locate session file
	loc := locator.NewSessionLocator()
	sessionPath, err := loc.Locate(locator.LocateOptions{
		SessionID:   sessionID,
		ProjectPath: projectPath,
	})
	if err != nil {
		return fmt.Errorf("failed to locate session file: %w", err)
	}

	// Step 2: Parse session file
	sessionParser := parser.NewSessionParser(sessionPath)
	entries, err := sessionParser.ParseEntries()
	if err != nil {
		return fmt.Errorf("failed to parse session file: %w", err)
	}

	// Step 3: Detect idle periods
	result := analyzer.DetectIdlePeriods(entries, idleThresholdMin)

	// Step 4: Format and output
	var outputStr string
	var formatErr error

	switch outputFormat {
	case "json":
		outputStr, formatErr = output.FormatJSON(result)
	case "md", "markdown":
		outputStr, formatErr = formatIdlePeriodsMarkdown(result)
	default:
		return fmt.Errorf("unsupported output format: %s (analyze idle-periods supports json and md)", outputFormat)
	}

	if formatErr != nil {
		return fmt.Errorf("failed to format output: %w", formatErr)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}

func formatIdlePeriodsMarkdown(result analyzer.IdlePeriodAnalysis) (string, error) {
	if len(result.IdlePeriods) == 0 {
		return "# Idle Period Analysis\n\nNo idle periods detected.\n", nil
	}

	var md string
	md += "# Idle Period Analysis\n\n"
	md += fmt.Sprintf("Found %d idle period(s):\n\n", len(result.IdlePeriods))

	for i, period := range result.IdlePeriods {
		md += fmt.Sprintf("## Period %d\n\n", i+1)
		md += fmt.Sprintf("- **Duration**: %.1f minutes\n", period.DurationMin)
		md += fmt.Sprintf("- **Between Turns**: %d → %d\n", period.StartTurn, period.EndTurn)

		if period.ContextBefore != nil {
			md += "\n**Context Before**:\n"
			md += fmt.Sprintf("- Turn %d (%s)\n", period.ContextBefore.Turn, period.ContextBefore.Role)
			if period.ContextBefore.Tool != "" {
				md += fmt.Sprintf("- Tool: %s\n", period.ContextBefore.Tool)
			}
			if period.ContextBefore.Status != "" {
				md += fmt.Sprintf("- Status: %s\n", period.ContextBefore.Status)
			}
			if period.ContextBefore.Preview != "" {
				md += fmt.Sprintf("- Preview: %s\n", period.ContextBefore.Preview)
			}
		}

		if period.ContextAfter != nil {
			md += "\n**Context After**:\n"
			md += fmt.Sprintf("- Turn %d (%s)\n", period.ContextAfter.Turn, period.ContextAfter.Role)
			if period.ContextAfter.Tool != "" {
				md += fmt.Sprintf("- Tool: %s\n", period.ContextAfter.Tool)
			}
			if period.ContextAfter.Preview != "" {
				md += fmt.Sprintf("- Preview: %s\n", period.ContextAfter.Preview)
			}
		}

		md += "\n---\n\n"
	}

	return md, nil
}
