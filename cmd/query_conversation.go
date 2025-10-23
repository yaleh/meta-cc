package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	mcerrors "github.com/yaleh/meta-cc/internal/errors"
	"github.com/yaleh/meta-cc/internal/query"
	"github.com/yaleh/meta-cc/pkg/output"
)

var (
	queryConvStartTurn     int
	queryConvEndTurn       int
	queryConvPattern       string
	queryConvPatternTarget string
	queryConvMinDuration   int
	queryConvMaxDuration   int
)

type ConversationTurn = query.ConversationTurn

type ConversationOptions = query.ConversationOptions

var queryConversationCmd = &cobra.Command{
	Use:   "conversation",
	Short: "Query conversation turns (user+assistant pairs)",
	RunE:  runQueryConversation,
}

func init() {
	queryCmd.AddCommand(queryConversationCmd)

	queryConversationCmd.Flags().IntVar(&queryConvStartTurn, "start-turn", -1, "Starting turn sequence")
	queryConversationCmd.Flags().IntVar(&queryConvEndTurn, "end-turn", -1, "Ending turn sequence")
	queryConversationCmd.Flags().StringVar(&queryConvPattern, "pattern", "", "Pattern to match (regex)")
	queryConversationCmd.Flags().StringVar(&queryConvPatternTarget, "pattern-target", "any", "Pattern target: user, assistant, any")
	queryConversationCmd.Flags().IntVar(&queryConvMinDuration, "min-duration", -1, "Minimum response duration (ms)")
	queryConversationCmd.Flags().IntVar(&queryConvMaxDuration, "max-duration", -1, "Maximum response duration (ms)")
}

func runQueryConversation(cmd *cobra.Command, args []string) error {
	p := NewSessionPipeline(getGlobalOptions())
	if err := p.Load(LoadOptions{AutoDetect: true}); err != nil {
		return fmt.Errorf("failed to locate session: %w", err)
	}

	entries := p.GetEntries()
	options := query.ConversationOptions{
		StartTurn:     queryConvStartTurn,
		EndTurn:       queryConvEndTurn,
		Pattern:       queryConvPattern,
		PatternTarget: queryConvPatternTarget,
		MinDuration:   queryConvMinDuration,
		MaxDuration:   queryConvMaxDuration,
		Limit:         queryLimit,
		Offset:        queryOffset,
		SortBy:        querySortBy,
		Reverse:       queryReverse,
	}

	turns, err := query.BuildConversationTurns(entries, options)
	if err != nil {
		return err
	}

	var outputStr string
	var formatErr error

	switch outputFormat {
	case "jsonl":
		outputStr, formatErr = output.FormatJSONL(turns)
	case "tsv":
		outputStr, formatErr = output.FormatTSV(turns)
	default:
		return fmt.Errorf("unsupported output format: %s (supported: jsonl, tsv): %w", outputFormat, mcerrors.ErrInvalidInput)
	}

	if formatErr != nil {
		return fmt.Errorf("failed to format output: %w", formatErr)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}
