package cmd

import (
	"github.com/spf13/cobra"
)

// analyzeCmd represents the analyze subcommand
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze Claude Code session patterns",
	Long: `Analyze Claude Code session data to detect patterns and insights.

Note: analyze errors has been replaced by 'query errors' in Phase 14.
Use 'meta-cc query errors' for error extraction with jq for aggregation.

Examples:
  meta-cc analyze sequences
  meta-cc analyze file-churn
  meta-cc analyze idle`,
}

func init() {
	// Add analyze subcommand to root
	rootCmd.AddCommand(analyzeCmd)

	// Note: analyze errors subcommand has been removed
	// Use 'query errors' instead
}
