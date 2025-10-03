package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/filter"
	"github.com/yale/meta-cc/internal/locator"
	"github.com/yale/meta-cc/internal/parser"
	"github.com/yale/meta-cc/internal/stats"
	"github.com/yale/meta-cc/pkg/output"
)

var (
	filesSortBy string
	filesTop    int
	filesFilter string
)

// statsFilesCmd represents the stats files subcommand
var statsFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "Analyze file-level statistics",
	Long: `Show statistics for files accessed during the session.

Tracks read, edit, write operations and error counts per file.

Supported sort fields:
  - total_ops:   Total number of operations (default)
  - edit_count:  Number of edit operations
  - read_count:  Number of read operations
  - write_count: Number of write operations
  - error_count: Number of errors
  - error_rate:  Percentage of operations that failed (0.0-1.0)

Examples:
  # Most edited files
  meta-cc stats files --sort-by edit_count --top 20

  # Files with most errors
  meta-cc stats files --sort-by error_count --top 10

  # Files with highest error rate
  meta-cc stats files --sort-by error_rate --filter "error_count>0"`,
	RunE: runStatsFiles,
}

func init() {
	// Add files subcommand to stats
	statsCmd.AddCommand(statsFilesCmd)

	// Flags for files
	statsFilesCmd.Flags().StringVar(&filesSortBy, "sort-by", "total_ops", "Sort by field (total_ops|edit_count|read_count|write_count|error_count|error_rate)")
	statsFilesCmd.Flags().IntVar(&filesTop, "top", 0, "Show only top N files (0 = all)")
	statsFilesCmd.Flags().StringVar(&filesFilter, "filter", "", "Filter expression (SQL-like)")
}

func runStatsFiles(cmd *cobra.Command, args []string) error {
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

	// Step 3: Analyze file statistics
	fileStats := stats.AnalyzeFileStats(toolCalls)

	// Step 4: Apply filter if provided
	if filesFilter != "" {
		expr, err := filter.ParseExpression(filesFilter)
		if err != nil {
			return fmt.Errorf("invalid filter: %w", err)
		}

		var filtered []stats.FileStats
		for _, fs := range fileStats {
			// Convert FileStats to map for expression evaluation
			record := map[string]interface{}{
				"file_path":   fs.FilePath,
				"read_count":  fs.ReadCount,
				"edit_count":  fs.EditCount,
				"write_count": fs.WriteCount,
				"error_count": fs.ErrorCount,
				"total_ops":   fs.TotalOps,
				"error_rate":  fs.ErrorRate,
			}

			match, err := expr.Evaluate(record)
			if err != nil {
				return fmt.Errorf("filter evaluation error: %w", err)
			}

			if match {
				filtered = append(filtered, fs)
			}
		}
		fileStats = filtered
	}

	// Step 5: Apply sorting
	stats.SortFileStats(fileStats, filesSortBy)

	// Step 6: Apply top N limit
	if filesTop > 0 && filesTop < len(fileStats) {
		fileStats = fileStats[:filesTop]
	}

	// Step 7: Format output
	var outputStr string
	switch outputFormat {
	case "json":
		outputStr, err = output.FormatJSON(fileStats)
	case "md", "markdown":
		outputStr, err = output.FormatMarkdown(fileStats)
	case "csv":
		outputStr, err = output.FormatCSV(fileStats)
	default:
		outputStr, err = output.FormatJSON(fileStats)
	}

	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}
