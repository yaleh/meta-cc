package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/locator"
	"github.com/yale/meta-cc/internal/parser"
	"github.com/yale/meta-cc/internal/query"
)

var (
	sequencesMinOccur int
	sequencesPattern  string
)

// querySequencesCmd represents the tool-sequences query command
var querySequencesCmd = &cobra.Command{
	Use:   "tool-sequences",
	Short: "Query repeated tool call sequences",
	Long: `Query repeated tool call sequences in Claude Code session history.

This command detects repeated patterns of tool calls, such as:
- Read → Edit → Bash (test cycle)
- Grep → Read → Read (exploration pattern)

Results include occurrence count, turn numbers, and time span.

Example:
  meta-cc query tool-sequences --min-occurrences 3
  meta-cc query tool-sequences --pattern "Read -> Edit -> Bash"
  meta-cc query tool-sequences --min-occurrences 2 --output md`,
	RunE: runQuerySequences,
}

func init() {
	querySequencesCmd.Flags().IntVar(&sequencesMinOccur, "min-occurrences", 3, "Minimum occurrences to report")
	querySequencesCmd.Flags().StringVar(&sequencesPattern, "pattern", "", "Specific sequence pattern to match (e.g., 'Read -> Edit -> Bash')")

	queryCmd.AddCommand(querySequencesCmd)
}

func runQuerySequences(cmd *cobra.Command, args []string) error {
	// Locate session file
	loc := locator.NewSessionLocator()
	sessionPath, err := loc.Locate(locator.LocateOptions{
		SessionID:   sessionID,
		ProjectPath: projectPath,
	})
	if err != nil {
		return fmt.Errorf("failed to locate session: %w", err)
	}

	// Read and parse session file
	sessionParser := parser.NewSessionParser(sessionPath)
	entries, err := sessionParser.ParseEntries()
	if err != nil {
		return fmt.Errorf("failed to read session file: %w", err)
	}

	// Apply time filter
	entries, err = applyTimeFilter(entries)
	if err != nil {
		return fmt.Errorf("failed to apply time filter: %w", err)
	}

	// Build tool sequence query
	result, err := query.BuildToolSequenceQuery(entries, sequencesMinOccur, sequencesPattern)
	if err != nil {
		return fmt.Errorf("failed to build tool sequence query: %w", err)
	}

	// Output result
	if outputFormat == "md" {
		return outputSequencesMarkdown(result)
	}

	// JSON output (default)
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func outputSequencesMarkdown(result *query.ToolSequenceQuery) error {
	var sb strings.Builder

	sb.WriteString("# Tool Sequence Patterns\n\n")
	sb.WriteString(fmt.Sprintf("**Patterns Found**: %d\n\n", len(result.Sequences)))

	for i, seq := range result.Sequences {
		sb.WriteString(fmt.Sprintf("## Pattern %d: %s\n\n", i+1, seq.Pattern))
		sb.WriteString(fmt.Sprintf("**Occurrences**: %d\n\n", seq.Count))
		sb.WriteString(fmt.Sprintf("**Time Span**: %d minutes\n\n", seq.TimeSpanMin))

		if len(seq.Occurrences) > 0 {
			sb.WriteString("### Occurrence Locations\n\n")
			sb.WriteString("| # | Start Turn | End Turn |\n")
			sb.WriteString("|---|------------|----------|\n")
			for j, occ := range seq.Occurrences {
				sb.WriteString(fmt.Sprintf("| %d | %d | %d |\n",
					j+1, occ.StartTurn, occ.EndTurn))
			}
			sb.WriteString("\n")
		}
	}

	fmt.Print(sb.String())
	return nil
}
