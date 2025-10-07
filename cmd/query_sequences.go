package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/parser"
	"github.com/yale/meta-cc/internal/query"
	"github.com/yale/meta-cc/internal/types"
)

var (
	sequencesMinOccur        int
	querySequencesPattern    string
	sequencesSuccessOnly     bool
	sequencesWithMetrics     bool
	queryIncludeBuiltinTools bool
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
	querySequencesCmd.Flags().StringVar(&querySequencesPattern, "pattern", "", "Specific sequence pattern to match (e.g., 'Read -> Edit -> Bash')")
	querySequencesCmd.Flags().BoolVar(&sequencesSuccessOnly, "successful-only", false, "Only show sequences with no errors")
	querySequencesCmd.Flags().BoolVar(&sequencesWithMetrics, "with-metrics", false, "Include success rate and duration metrics")
	querySequencesCmd.Flags().BoolVar(&queryIncludeBuiltinTools, "include-builtin-tools", false, "Include built-in tools (Bash, Read, Edit, etc.) in sequence analysis. Default: false (exclude for cleaner workflow patterns)")

	queryCmd.AddCommand(querySequencesCmd)
}

func runQuerySequences(cmd *cobra.Command, args []string) error {
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

	// Build tool sequence query
	result, err := query.BuildToolSequenceQuery(entries, sequencesMinOccur, querySequencesPattern, queryIncludeBuiltinTools)
	if err != nil {
		return fmt.Errorf("failed to build tool sequence query: %w", err)
	}

	// Apply filters and enhancements
	if sequencesSuccessOnly {
		result = filterSuccessfulSequences(result, entries)
	}

	if sequencesWithMetrics {
		result = addSequenceMetrics(result, entries)
	}

	// Output result
	if outputFormat == "md" {
		return outputSequencesMarkdown(cmd, result)
	}

	// JSON output (default)
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func outputSequencesMarkdown(cmd *cobra.Command, result *query.ToolSequenceQuery) error {
	var sb strings.Builder

	sb.WriteString("# Tool Sequence Patterns\n\n")
	sb.WriteString(fmt.Sprintf("**Patterns Found**: %d\n\n", len(result.Sequences)))

	for i, seq := range result.Sequences {
		sb.WriteString(fmt.Sprintf("## Pattern %d: %s\n\n", i+1, seq.Pattern))
		sb.WriteString(fmt.Sprintf("**Occurrences**: %d\n\n", seq.Count))
		sb.WriteString(fmt.Sprintf("**Time Span**: %d minutes\n\n", seq.TimeSpanMin))

		// Add metrics if available
		if seq.SuccessRate >= 0 {
			sb.WriteString(fmt.Sprintf("**Success Rate**: %.1f%%\n\n", seq.SuccessRate*100))
		}
		if seq.AvgDurationMin > 0 {
			sb.WriteString(fmt.Sprintf("**Avg Duration**: %.1f minutes\n\n", seq.AvgDurationMin))
		}
		if seq.Context != "" {
			sb.WriteString(fmt.Sprintf("**Context**: %s\n\n", seq.Context))
		}

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

	fmt.Fprint(cmd.OutOrStdout(), sb.String())
	return nil
}

// filterSuccessfulSequences filters sequences to only include those with no errors
func filterSuccessfulSequences(result *query.ToolSequenceQuery, entries []parser.SessionEntry) *query.ToolSequenceQuery {
	// Build error turn map
	errorTurns := make(map[int]bool)
	turnIndex := buildTurnIndexFromEntries(entries)

	for _, entry := range entries {
		if entry.Message == nil {
			continue
		}

		for _, block := range entry.Message.Content {
			if block.Type == "tool_result" && block.ToolResult != nil {
				if block.ToolResult.Status == "error" || block.ToolResult.Error != "" {
					turn := turnIndex[entry.UUID]
					errorTurns[turn] = true
				}
			}
		}
	}

	// Filter sequences
	var filtered []types.SequencePattern
	for _, seq := range result.Sequences {
		hasError := false
		for _, occ := range seq.Occurrences {
			for turn := occ.StartTurn; turn <= occ.EndTurn; turn++ {
				if errorTurns[turn] {
					hasError = true
					break
				}
			}
			if hasError {
				break
			}
		}

		if !hasError {
			filtered = append(filtered, seq)
		}
	}

	return &query.ToolSequenceQuery{
		Sequences: filtered,
	}
}

// addSequenceMetrics adds success rate and duration metrics to sequences
func addSequenceMetrics(result *query.ToolSequenceQuery, entries []parser.SessionEntry) *query.ToolSequenceQuery {
	// Build error turn map and timestamp map
	errorTurns := make(map[int]bool)
	turnTimestamps := make(map[int]int64)
	turnIndex := buildTurnIndexFromEntries(entries)

	for _, entry := range entries {
		turn := turnIndex[entry.UUID]

		// Parse timestamp
		var ts int64
		_, _ = fmt.Sscanf(entry.Timestamp, "%d", &ts)
		turnTimestamps[turn] = ts

		if entry.Message == nil {
			continue
		}

		for _, block := range entry.Message.Content {
			if block.Type == "tool_result" && block.ToolResult != nil {
				if block.ToolResult.Status == "error" || block.ToolResult.Error != "" {
					errorTurns[turn] = true
				}
			}
		}
	}

	// Calculate metrics for each sequence
	for i := range result.Sequences {
		seq := &result.Sequences[i]

		successCount := 0
		totalDuration := 0.0

		for _, occ := range seq.Occurrences {
			// Check if successful
			hasError := false
			for turn := occ.StartTurn; turn <= occ.EndTurn; turn++ {
				if errorTurns[turn] {
					hasError = true
					break
				}
			}

			if !hasError {
				successCount++
			}

			// Calculate duration
			startTs := turnTimestamps[occ.StartTurn]
			endTs := turnTimestamps[occ.EndTurn]
			if startTs > 0 && endTs > 0 {
				duration := float64(endTs-startTs) / 60.0 // convert to minutes
				totalDuration += duration
			}
		}

		// Calculate success rate
		if len(seq.Occurrences) > 0 {
			seq.SuccessRate = float64(successCount) / float64(len(seq.Occurrences))
		}

		// Calculate average duration
		if len(seq.Occurrences) > 0 {
			seq.AvgDurationMin = totalDuration / float64(len(seq.Occurrences))
		}

		// Determine context based on pattern
		seq.Context = determineSequenceContext(seq.Pattern)
	}

	return result
}

// buildTurnIndexFromEntries builds a turn index from entries
func buildTurnIndexFromEntries(entries []parser.SessionEntry) map[string]int {
	turnIndex := make(map[string]int)
	turn := 0

	for _, entry := range entries {
		if entry.IsMessage() {
			turn++
			turnIndex[entry.UUID] = turn
		}
	}

	return turnIndex
}

// determineSequenceContext determines the likely context/purpose of a sequence
func determineSequenceContext(pattern string) string {
	lower := strings.ToLower(pattern)

	contexts := map[string][]string{
		"代码修改工作流":  {"read", "grep", "edit"},
		"测试驱动开发循环": {"bash", "read", "edit", "bash"},
		"文件创建和验证":  {"write", "read"},
		"探索性代码阅读":  {"grep", "read", "read"},
		"调试和错误修复":  {"bash", "read", "edit"},
	}

	for context, keywords := range contexts {
		matchCount := 0
		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				matchCount++
			}
		}
		// If most keywords match, return this context
		if matchCount >= len(keywords)-1 {
			return context
		}
	}

	return ""
}
