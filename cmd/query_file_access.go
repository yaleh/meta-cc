package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/query"
)

var (
	fileAccessFile string
)

// queryFileAccessCmd represents the file-access query command
var queryFileAccessCmd = &cobra.Command{
	Use:   "file-access",
	Short: "Query file access history",
	Long: `Query file access history in Claude Code session.

This command tracks all operations (Read/Edit/Write) on a specific file, including:
- Total number of accesses
- Operations breakdown
- Timeline of accesses
- Time span in minutes

Example:
  meta-cc query file-access --file test.js
  meta-cc query file-access --file /path/to/config.yaml --output md`,
	RunE: runQueryFileAccess,
}

func init() {
	queryFileAccessCmd.Flags().StringVar(&fileAccessFile, "file", "", "File path to query (required)")
	queryFileAccessCmd.MarkFlagRequired("file")

	queryCmd.AddCommand(queryFileAccessCmd)
}

func runQueryFileAccess(cmd *cobra.Command, args []string) error {
	// Initialize and load session using pipeline
	p := NewSessionPipeline(getGlobalOptions())
	if err := p.Load(LoadOptions{AutoDetect: true}); err != nil {
		return fmt.Errorf("failed to locate session: %w", err)
	}

	// Apply time filter
	entries := p.GetEntries()
	entries, err := applyTimeFilter(entries)
	if err != nil {
		return fmt.Errorf("failed to apply time filter: %w", err)
	}

	// Build file access query
	result, err := query.BuildFileAccessQuery(entries, fileAccessFile)
	if err != nil {
		return fmt.Errorf("failed to build file access query: %w", err)
	}

	// Output result
	if outputFormat == "md" {
		return outputFileAccessMarkdown(cmd, result)
	}

	// JSON output (default)
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func outputFileAccessMarkdown(cmd *cobra.Command, result *query.FileAccessQuery) error {
	var sb strings.Builder

	sb.WriteString("# File Access History\n\n")
	sb.WriteString(fmt.Sprintf("**File**: `%s`\n\n", result.File))
	sb.WriteString(fmt.Sprintf("**Total Accesses**: %d\n\n", result.TotalAccesses))
	sb.WriteString(fmt.Sprintf("**Time Span**: %d minutes\n\n", result.TimeSpanMin))

	// Operations breakdown
	if len(result.Operations) > 0 {
		sb.WriteString("## Operations Breakdown\n\n")
		sb.WriteString("| Operation | Count |\n")
		sb.WriteString("|-----------|-------|\n")
		for op, count := range result.Operations {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", op, count))
		}
		sb.WriteString("\n")
	}

	// Timeline
	if len(result.Timeline) > 0 {
		sb.WriteString("## Access Timeline\n\n")
		sb.WriteString("| Turn | Action | Timestamp |\n")
		sb.WriteString("|------|--------|----------|\n")
		for _, event := range result.Timeline {
			sb.WriteString(fmt.Sprintf("| %d | %s | %d |\n",
				event.Turn, event.Action, event.Timestamp))
		}
	}

	fmt.Fprint(cmd.OutOrStdout(), sb.String())
	return nil
}
