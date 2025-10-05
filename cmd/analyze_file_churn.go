package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/analyzer"
	"github.com/yale/meta-cc/internal/locator"
	"github.com/yale/meta-cc/internal/parser"
	"github.com/yale/meta-cc/pkg/output"
)

var fileChurnThreshold int

// analyzeFileChurnCmd represents the analyze file-churn subcommand
var analyzeFileChurnCmd = &cobra.Command{
	Use:   "file-churn",
	Short: "Detect files with frequent access",
	Long: `Detect files that are frequently accessed in Claude Code session data.

Identifies files that are read, edited, or written multiple times.
Useful for detecting files that may need clarification or refactoring.

Examples:
  meta-cc analyze file-churn
  meta-cc analyze file-churn --threshold 5
  meta-cc analyze file-churn --output json`,
	RunE: runAnalyzeFileChurn,
}

func init() {
	analyzeCmd.AddCommand(analyzeFileChurnCmd)

	analyzeFileChurnCmd.Flags().IntVar(&fileChurnThreshold, "threshold", 5, "Minimum access count to report")
}

func runAnalyzeFileChurn(cmd *cobra.Command, args []string) error {
	// Step 1: Locate session file
	loc := locator.NewSessionLocator()
	sessionPath, err := loc.Locate(locator.LocateOptions{
		SessionID:   sessionID,
		ProjectPath: projectPath, // from global parameter
		SessionOnly: sessionOnly,  // Phase 13: opt-out of project default

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

	// Step 3: Detect file churn
	result := analyzer.DetectFileChurn(entries, fileChurnThreshold)

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

func formatFileChurnMarkdown(result analyzer.FileChurnAnalysis) (string, error) {
	if len(result.HighChurnFiles) == 0 {
		return "# File Churn Analysis\n\nNo high churn files detected.\n", nil
	}

	var md string
	md += "# File Churn Analysis\n\n"
	md += fmt.Sprintf("Found %d high churn file(s):\n\n", len(result.HighChurnFiles))

	for i, file := range result.HighChurnFiles {
		md += fmt.Sprintf("## %d. %s\n\n", i+1, file.File)
		md += fmt.Sprintf("- **Total Accesses**: %d\n", file.TotalAccesses)
		md += fmt.Sprintf("- **Read**: %d times\n", file.ReadCount)
		md += fmt.Sprintf("- **Edit**: %d times\n", file.EditCount)
		md += fmt.Sprintf("- **Write**: %d times\n", file.WriteCount)
		md += fmt.Sprintf("- **Time Span**: %d minutes\n", file.TimeSpanMin)
		md += "\n---\n\n"
	}

	return md, nil
}
