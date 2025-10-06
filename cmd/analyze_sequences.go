package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/analyzer"
	"github.com/yale/meta-cc/pkg/output"
)

var (
	sequencesMinLength      int
	sequencesMinOccurrences int
)

// analyzeSequencesCmd represents the analyze sequences subcommand
var analyzeSequencesCmd = &cobra.Command{
	Use:   "sequences",
	Short: "Detect repeated tool call sequences",
	Long: `Detect repeated tool call sequences in Claude Code session data.

Identifies patterns where the same sequence of tools is called multiple times.
Useful for detecting potentially inefficient workflows.

Examples:
  meta-cc analyze sequences
  meta-cc analyze sequences --min-length 3 --min-occurrences 3
  meta-cc analyze sequences --output json`,
	RunE: runAnalyzeSequences,
}

func init() {
	analyzeCmd.AddCommand(analyzeSequencesCmd)

	analyzeSequencesCmd.Flags().IntVar(&sequencesMinLength, "min-length", 3, "Minimum sequence length")
	analyzeSequencesCmd.Flags().IntVar(&sequencesMinOccurrences, "min-occurrences", 3, "Minimum occurrences to report")
}

func runAnalyzeSequences(cmd *cobra.Command, args []string) error {
	// Step 1: Initialize and load session using pipeline
	p := NewSessionPipeline(getGlobalOptions())
	if err := p.Load(LoadOptions{AutoDetect: true}); err != nil {
		return fmt.Errorf("failed to locate session file: %w", err)
	}

	// Step 2: Detect sequences
	entries := p.GetEntries()
	result := analyzer.DetectToolSequences(entries, sequencesMinLength, sequencesMinOccurrences)

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

func formatSequencesMarkdown(result analyzer.SequenceAnalysis) (string, error) {
	if len(result.Sequences) == 0 {
		return "# Tool Sequence Analysis\n\nNo repeated sequences detected.\n", nil
	}

	var md string
	md += "# Tool Sequence Analysis\n\n"
	md += fmt.Sprintf("Found %d repeated sequence(s):\n\n", len(result.Sequences))

	for i, seq := range result.Sequences {
		md += fmt.Sprintf("## Sequence %d: %s\n\n", i+1, seq.Pattern)
		md += fmt.Sprintf("- **Length**: %d tools\n", seq.Length)
		md += fmt.Sprintf("- **Occurrences**: %d times\n", seq.Count)
		md += fmt.Sprintf("- **Time Span**: %d minutes\n", seq.TimeSpanMin)
		md += "\n"

		md += "**Occurrences**:\n\n"
		limit := seq.Count
		if limit > 5 {
			limit = 5
		}
		for j := 0; j < limit; j++ {
			occ := seq.Occurrences[j]
			md += fmt.Sprintf("- Turn %d → %d\n", occ.StartTurn, occ.EndTurn)
		}
		if seq.Count > 5 {
			md += fmt.Sprintf("- ... and %d more\n", seq.Count-5)
		}
		md += "\n---\n\n"
	}

	return md, nil
}
