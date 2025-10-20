package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	mcerrors "github.com/yaleh/meta-cc/internal/errors"
	"github.com/yaleh/meta-cc/internal/query"
	"github.com/yaleh/meta-cc/pkg/output"
)

var (
	sequencesMinOccurrences int
	sequencesPattern        string
	includeBuiltinTools     bool
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

	analyzeSequencesCmd.Flags().IntVar(&sequencesMinOccurrences, "min-occurrences", 3, "Minimum occurrences to report")
	analyzeSequencesCmd.Flags().StringVar(&sequencesPattern, "pattern", "", "Specific pattern to match (e.g. \"Read → Edit\" or \"Read -> Edit\")")
	analyzeSequencesCmd.Flags().BoolVar(&includeBuiltinTools, "include-builtin-tools", false, "Include built-in tools (Bash, Read, Edit, etc.) in sequence analysis. Default: false (exclude for cleaner workflow patterns, 35x faster)")
}

func runAnalyzeSequences(cmd *cobra.Command, args []string) error {
	// Step 1: Initialize and load session using pipeline
	p := NewSessionPipeline(getGlobalOptions())
	if err := p.Load(LoadOptions{AutoDetect: true}); err != nil {
		return fmt.Errorf("failed to locate session file: %w", err)
	}

	entries := p.GetEntries()

	// Step 2: Use query package for all sequence detection
	// When pattern="" (empty), it auto-detects all repeated sequences
	// The includeBuiltinTools flag controls filtering (default: false = exclude built-in tools)
	queryResult, err := query.BuildToolSequenceQuery(entries, sequencesMinOccurrences, sequencesPattern, includeBuiltinTools)
	if err != nil {
		return fmt.Errorf("failed to query tool sequences: %w", err)
	}
	sequences := queryResult.Sequences

	// Step 3: Format and output
	var outputStr string
	var formatErr error

	switch outputFormat {
	case "jsonl":
		outputStr, formatErr = output.FormatJSONL(sequences)
	case "tsv":
		outputStr, formatErr = output.FormatTSV(sequences)
	default:
		return fmt.Errorf("unsupported output format: %s (supported: jsonl, tsv): %w", outputFormat, mcerrors.ErrInvalidInput)
	}

	if formatErr != nil {
		return fmt.Errorf("failed to format output: %w", formatErr)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}
