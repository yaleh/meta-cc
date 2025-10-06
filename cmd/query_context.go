package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/query"
)

var (
	contextErrorSig string
	contextWindow   int
)

// queryContextCmd represents the context query command
var queryContextCmd = &cobra.Command{
	Use:   "context",
	Short: "Query context around specific events (errors, files, etc.)",
	Long: `Query context around specific events in Claude Code session history.

This command finds error occurrences and returns surrounding context including:
- User messages and assistant responses
- Tool calls before and after the error
- Timestamps and turn numbers

Example:
  meta-cc query context --error-signature abc123 --window 3
  meta-cc query context --error-signature abc123 --window 5 --output md`,
	RunE: runQueryContext,
}

func init() {
	queryContextCmd.Flags().StringVar(&contextErrorSig, "error-signature", "", "Error pattern ID to query (required)")
	queryContextCmd.Flags().IntVar(&contextWindow, "window", 3, "Context window size (turns before/after)")
	_ = queryContextCmd.MarkFlagRequired("error-signature")

	queryCmd.AddCommand(queryContextCmd)
}

func runQueryContext(cmd *cobra.Command, args []string) error {
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

	// Build context query
	result, err := query.BuildContextQuery(entries, contextErrorSig, contextWindow)
	if err != nil {
		return fmt.Errorf("failed to build context query: %w", err)
	}

	// Output result
	if outputFormat == "md" {
		return outputContextMarkdown(cmd, result)
	}

	// JSON output (default)
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func outputContextMarkdown(cmd *cobra.Command, result *query.ContextQuery) error {
	var sb strings.Builder

	sb.WriteString("# Context Query Result\n\n")
	sb.WriteString(fmt.Sprintf("**Error Signature**: `%s`\n\n", result.ErrorSignature))
	sb.WriteString(fmt.Sprintf("**Occurrences Found**: %d\n\n", len(result.Occurrences)))

	for i, occ := range result.Occurrences {
		sb.WriteString(fmt.Sprintf("## Occurrence %d (Turn %d)\n\n", i+1, occ.Turn))

		// Context before
		if len(occ.ContextBefore) > 0 {
			sb.WriteString("### Context Before\n\n")
			sb.WriteString("| Turn | Role | Preview | Tools |\n")
			sb.WriteString("|------|------|---------|-------|\n")
			for _, ctx := range occ.ContextBefore {
				toolsStr := formatToolsList(ctx.Tools)
				sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
					ctx.Turn, ctx.Role, ctx.Preview, toolsStr))
			}
			sb.WriteString("\n")
		}

		// Error turn
		sb.WriteString("### Error Turn\n\n")
		sb.WriteString("| Field | Value |\n")
		sb.WriteString("|-------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Turn | %d |\n", occ.ErrorTurn.Turn))
		sb.WriteString(fmt.Sprintf("| Tool | %s |\n", occ.ErrorTurn.Tool))
		if occ.ErrorTurn.Command != "" {
			sb.WriteString(fmt.Sprintf("| Command | %s |\n", occ.ErrorTurn.Command))
		}
		if occ.ErrorTurn.File != "" {
			sb.WriteString(fmt.Sprintf("| File | %s |\n", occ.ErrorTurn.File))
		}
		sb.WriteString(fmt.Sprintf("| Error | %s |\n\n", occ.ErrorTurn.Error))

		// Context after
		if len(occ.ContextAfter) > 0 {
			sb.WriteString("### Context After\n\n")
			sb.WriteString("| Turn | Role | Preview | Tools |\n")
			sb.WriteString("|------|------|---------|-------|\n")
			for _, ctx := range occ.ContextAfter {
				toolsStr := formatToolsList(ctx.Tools)
				sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
					ctx.Turn, ctx.Role, ctx.Preview, toolsStr))
			}
			sb.WriteString("\n")
		}
	}

	fmt.Fprint(cmd.OutOrStdout(), sb.String())
	return nil
}

func formatToolsList(tools []string) string {
	if len(tools) == 0 {
		return ""
	}
	result := ""
	for i, tool := range tools {
		if i > 0 {
			result += ", "
		}
		result += tool
	}
	return result
}
