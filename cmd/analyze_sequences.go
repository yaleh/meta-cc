package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/analyzer"
	"github.com/yale/meta-cc/internal/query"
	"github.com/yale/meta-cc/pkg/output"
)

var (
	sequencesMinLength      int
	sequencesMinOccurrences int
	sequencesPattern        string
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
  meta-cc analyze sequences --pattern "Read → Edit → Bash" --min-occurrences 5
  meta-cc analyze sequences --output json`,
	RunE: runAnalyzeSequences,
}

func init() {
	analyzeCmd.AddCommand(analyzeSequencesCmd)

	analyzeSequencesCmd.Flags().IntVar(&sequencesMinLength, "min-length", 3, "Minimum sequence length (ignored when --pattern is specified)")
	analyzeSequencesCmd.Flags().IntVar(&sequencesMinOccurrences, "min-occurrences", 3, "Minimum occurrences to report")
	analyzeSequencesCmd.Flags().StringVar(&sequencesPattern, "pattern", "", "Specific pattern to match (e.g. \"Read → Edit\" or \"Read -> Edit\")")
}

func runAnalyzeSequences(cmd *cobra.Command, args []string) error {
	// Step 1: Initialize and load session using pipeline
	p := NewSessionPipeline(getGlobalOptions())
	if err := p.Load(LoadOptions{AutoDetect: true}); err != nil {
		return fmt.Errorf("failed to locate session file: %w", err)
	}

	entries := p.GetEntries()

	// Step 2: Detect sequences based on whether pattern is specified
	var sequences interface{}

	if sequencesPattern != "" {
		// Use query package for specific pattern matching
		queryResult, err := query.BuildToolSequenceQuery(entries, sequencesMinOccurrences, sequencesPattern)
		if err != nil {
			return fmt.Errorf("failed to query tool sequences: %w", err)
		}
		sequences = queryResult.Sequences
	} else {
		// Use analyzer package for general sequence detection
		result := analyzer.DetectToolSequences(entries, sequencesMinLength, sequencesMinOccurrences)
		sequences = result.Sequences
	}

	// Step 3: Format and output
	var outputStr string
	var formatErr error

	switch outputFormat {
	case "jsonl":
		outputStr, formatErr = output.FormatJSONL(sequences)
	case "tsv":
		outputStr, formatErr = output.FormatTSV(sequences)
	default:
		return fmt.Errorf("unsupported output format: %s (supported: jsonl, tsv)", outputFormat)
	}

	if formatErr != nil {
		return fmt.Errorf("failed to format output: %w", formatErr)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}
