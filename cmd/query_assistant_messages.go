package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	mcerrors "github.com/yaleh/meta-cc/internal/errors"
	"github.com/yaleh/meta-cc/internal/query"
	"github.com/yaleh/meta-cc/pkg/output"
)

var (
	queryAssistantPattern   string
	queryAssistantMinTools  int
	queryAssistantMaxTools  int
	queryAssistantMinTokens int
	queryAssistantMinLength int
	queryAssistantMaxLength int
)

type AssistantMessage = query.AssistantMessage

type AssistantMessagesOptions = query.AssistantMessagesOptions

var queryAssistantMessagesCmd = &cobra.Command{
	Use:   "assistant-messages",
	Short: "Query assistant messages",
	Long: `Query assistant messages with pattern matching and filtering.

Supports:
  - Pattern matching (--pattern: regex pattern on text content)
  - Tool usage filtering (--min-tools, --max-tools)
  - Token filtering (--min-tokens-output)
  - Length filtering (--min-length, --max-length)
  - Timestamp sorting
  - Limit and pagination

Examples:
  meta-cc query assistant-messages --pattern "fix.*bug"
  meta-cc query assistant-messages --min-tools 5
  meta-cc query assistant-messages --min-tokens-output 2000
  meta-cc query assistant-messages --pattern "error" --min-tools 2 --limit 10`,
	RunE: runQueryAssistantMessages,
}

func init() {
	queryCmd.AddCommand(queryAssistantMessagesCmd)

	queryAssistantMessagesCmd.Flags().StringVar(&queryAssistantPattern, "pattern", "", "Pattern to match text content (regex)")
	queryAssistantMessagesCmd.Flags().IntVar(&queryAssistantMinTools, "min-tools", -1, "Minimum tool use count")
	queryAssistantMessagesCmd.Flags().IntVar(&queryAssistantMaxTools, "max-tools", -1, "Maximum tool use count")
	queryAssistantMessagesCmd.Flags().IntVar(&queryAssistantMinTokens, "min-tokens-output", -1, "Minimum output tokens")
	queryAssistantMessagesCmd.Flags().IntVar(&queryAssistantMinLength, "min-length", -1, "Minimum text length")
	queryAssistantMessagesCmd.Flags().IntVar(&queryAssistantMaxLength, "max-length", -1, "Maximum text length")
}

func runQueryAssistantMessages(cmd *cobra.Command, args []string) error {
	p := NewSessionPipeline(getGlobalOptions())
	if err := p.Load(LoadOptions{AutoDetect: true}); err != nil {
		return fmt.Errorf("failed to locate session: %w", err)
	}

	entries := p.GetEntries()
	options := query.AssistantMessagesOptions{
		Pattern:   queryAssistantPattern,
		MinTools:  queryAssistantMinTools,
		MaxTools:  queryAssistantMaxTools,
		MinTokens: queryAssistantMinTokens,
		MinLength: queryAssistantMinLength,
		MaxLength: queryAssistantMaxLength,
		Limit:     queryLimit,
		Offset:    queryOffset,
		SortBy:    querySortBy,
		Reverse:   queryReverse,
	}

	messages, err := query.BuildAssistantMessages(entries, options)
	if err != nil {
		return err
	}

	var outputStr string
	var formatErr error

	switch outputFormat {
	case "jsonl":
		outputStr, formatErr = output.FormatJSONL(messages)
	case "tsv":
		outputStr, formatErr = output.FormatTSV(messages)
	default:
		return fmt.Errorf("unsupported output format: %s (supported: jsonl, tsv): %w", outputFormat, mcerrors.ErrInvalidInput)
	}

	if formatErr != nil {
		return fmt.Errorf("failed to format output: %w", formatErr)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}
