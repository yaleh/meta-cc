package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yaleh/meta-cc/internal/query"
)

var (
	successfulPromptsMinQuality float64
)

var querySuccessfulPromptsCmd = &cobra.Command{
	Use:   "successful-prompts",
	Short: "Query successful prompt patterns",
	RunE:  runQuerySuccessfulPrompts,
}

func init() {
	querySuccessfulPromptsCmd.Flags().Float64Var(&successfulPromptsMinQuality, "min-quality-score", 0.0, "Minimum quality score (0.0-1.0)")
	queryCmd.AddCommand(querySuccessfulPromptsCmd)
}

func runQuerySuccessfulPrompts(cmd *cobra.Command, args []string) error {
	p := NewSessionPipeline(getGlobalOptions())
	if err := p.Load(LoadOptions{AutoDetect: true}); err != nil {
		return fmt.Errorf("failed to locate session: %w", err)
	}

	entries := p.GetEntries()
	result := query.BuildSuccessfulPrompts(entries, successfulPromptsMinQuality, queryLimit)

	if outputFormat == "md" {
		return outputSuccessfulPromptsMarkdown(cmd, result)
	}

	for _, prompt := range result.Prompts {
		jsonBytes, err := json.Marshal(prompt)
		if err != nil {
			return fmt.Errorf("failed to marshal prompt: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(jsonBytes))
	}
	return nil
}

func outputSuccessfulPromptsMarkdown(cmd *cobra.Command, result *query.SuccessfulPromptsResult) error {
	fmt.Fprintf(cmd.OutOrStdout(), "# Successful Prompts (%d)\n\n", len(result.Prompts))
	for i, prompt := range result.Prompts {
		fmt.Fprintf(cmd.OutOrStdout(), "## Prompt %d (Turn %d)\n\n", i+1, prompt.TurnSequence)
		fmt.Fprintf(cmd.OutOrStdout(), "**Quality Score**: %.2f\n\n", prompt.QualityScore)
		fmt.Fprintf(cmd.OutOrStdout(), "**Prompt**: %s\n\n", prompt.UserPrompt)
		fmt.Fprintf(cmd.OutOrStdout(), "**Outcome**: %+v\n\n", prompt.Outcome)
	}
	return nil
}
