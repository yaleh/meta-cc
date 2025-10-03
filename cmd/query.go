package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/filter"
	"github.com/yale/meta-cc/internal/parser"
)

var (
	// Common query parameters
	queryLimit   int
	querySortBy  string
	queryReverse bool
	queryOffset  int
	queryStream  bool // Enable JSONL streaming output

	// Time filter parameters
	querySince      string
	queryLastNTurns int
	queryFromTs     int64
	queryToTs       int64
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query Claude Code session data",
	Long: `Query and retrieve specific data from Claude Code sessions.

The query command provides specialized subcommands for different data types:
  - tools:         Query tool calls with detailed filtering
  - user-messages: Query user messages with pattern matching

Examples:
  meta-cc query tools --status error --limit 20
  meta-cc query user-messages --match "fix.*bug"
  meta-cc query tools --tool Bash --sort-by timestamp`,
}

func init() {
	rootCmd.AddCommand(queryCmd)

	// Common query parameters as persistent flags
	queryCmd.PersistentFlags().IntVarP(&queryLimit, "limit", "l", 0, "Limit number of results (0 = no limit)")
	queryCmd.PersistentFlags().StringVarP(&querySortBy, "sort-by", "s", "", "Sort by field (timestamp, tool, status)")
	queryCmd.PersistentFlags().BoolVarP(&queryReverse, "reverse", "r", false, "Reverse sort order")
	queryCmd.PersistentFlags().IntVar(&queryOffset, "offset", 0, "Skip first N results (for pagination)")
	queryCmd.PersistentFlags().BoolVar(&queryStream, "stream", false, "Output as JSON Lines (JSONL) for streaming")

	// Time filter parameters as persistent flags
	queryCmd.PersistentFlags().StringVar(&querySince, "since", "", "Filter entries since duration ago (e.g., '5 minutes ago', '1 hour ago')")
	queryCmd.PersistentFlags().IntVar(&queryLastNTurns, "last-n-turns", 0, "Query last N turns only")
	queryCmd.PersistentFlags().Int64Var(&queryFromTs, "from", 0, "Start timestamp (Unix)")
	queryCmd.PersistentFlags().Int64Var(&queryToTs, "to", 0, "End timestamp (Unix)")
}

// applyTimeFilter applies time-based filtering to session entries
func applyTimeFilter(entries []parser.SessionEntry) ([]parser.SessionEntry, error) {
	timeFilter := filter.TimeFilter{
		Since:      querySince,
		LastNTurns: queryLastNTurns,
		FromTs:     queryFromTs,
		ToTs:       queryToTs,
	}

	return timeFilter.Apply(entries)
}
