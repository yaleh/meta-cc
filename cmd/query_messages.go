package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yaleh/meta-cc/internal/query"
	"github.com/yaleh/meta-cc/pkg/output"
)

var (
	queryMessagesPattern string
	queryMessagesContext int
)

var queryUserMessagesCmd = &cobra.Command{
	Use:   "user-messages",
	Short: "Query user messages",
	Long: `Query user messages with pattern matching.

Supports:
  - Pattern matching (--pattern: regex pattern)
  - Timestamp sorting
  - Limit and pagination

Examples:
  meta-cc query user-messages --pattern "fix.*bug"
  meta-cc query user-messages --pattern "error" --limit 10
  meta-cc query user-messages --sort-by timestamp --reverse`,
	RunE: runQueryUserMessages,
}

func init() {
	queryCmd.AddCommand(queryUserMessagesCmd)

	queryUserMessagesCmd.Flags().StringVar(&queryMessagesPattern, "pattern", "", "Pattern to match (regex)")
	queryUserMessagesCmd.Flags().IntVar(&queryMessagesContext, "with-context", 0, "Include N turns before/after each match")
}

// Re-export query package types for backwards compatibility within cmd package.
type UserMessage = query.UserMessage

type ContextEntry = query.ContextEntry

func runQueryUserMessages(cmd *cobra.Command, args []string) error {
	messages, err := query.RunUserMessagesQuery(buildUserMessagesOptions(getGlobalOptions()))
	if err != nil {
		return handleUserMessagesError(err)
	}

	var (
		outputStr string
		formatErr error
	)

	switch outputFormat {
	case "jsonl":
		outputStr, formatErr = output.FormatJSONL(messages)
	case "tsv":
		outputStr, formatErr = output.FormatTSV(messages)
	default:
		return fmt.Errorf("unsupported output format: %s (supported: jsonl, tsv)", outputFormat)
	}

	if formatErr != nil {
		return fmt.Errorf("failed to format output: %w", formatErr)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}

func buildUserMessagesOptions(globalOpts GlobalOptions) query.UserMessagesQueryOptions {
	limit := queryLimit
	if limitFlag > 0 {
		limit = limitFlag
	}

	offset := queryOffset
	if offsetFlag > 0 {
		offset = offsetFlag
	}

	return query.UserMessagesQueryOptions{
		Pipeline: toPipelineOptions(globalOpts),
		Pattern:  queryMessagesPattern,
		Context:  queryMessagesContext,
		Limit:    limit,
		Offset:   offset,
		SortBy:   querySortBy,
		Reverse:  queryReverse,
	}
}

func handleUserMessagesError(err error) error {
	switch {
	case errors.Is(err, query.ErrSessionLoad):
		return fmt.Errorf("failed to locate session: %w", err)
	case errors.Is(err, query.ErrInvalidPattern):
		return fmt.Errorf("invalid regex pattern: %w", err)
	default:
		return fmt.Errorf("query failed: %w", err)
	}
}
