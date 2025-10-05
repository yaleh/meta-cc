package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/filter"
	internalOutput "github.com/yale/meta-cc/internal/output"
)

// ErrorEntry represents a single error occurrence
type ErrorEntry struct {
	UUID      string `json:"uuid"`
	Timestamp string `json:"timestamp"`
	ToolName  string `json:"tool_name"`
	Error     string `json:"error"`
	Signature string `json:"signature"`
}

var queryErrorsCmd = &cobra.Command{
	Use:   "errors",
	Short: "Query tool errors",
	Long: `Extract all tool errors from session.

Returns a simple list of errors with signatures for downstream analysis.
Use jq, awk, or LLM for pattern detection and aggregation.

Examples:
  # All errors
  meta-cc query errors

  # Last 50 errors
  meta-cc query errors | jq '.[-50:]'

  # Group by signature
  meta-cc query errors | jq 'group_by(.signature)'

  # Count patterns
  meta-cc query errors | jq 'group_by(.signature) | map({sig: .[0].signature, count: length})'`,
	RunE: runQueryErrors,
}

func init() {
	queryCmd.AddCommand(queryErrorsCmd)
}

func runQueryErrors(cmd *cobra.Command, args []string) error {
	// 1. Initialize pipeline
	p := NewSessionPipeline(getGlobalOptions())

	// 2. Load session
	if err := p.Load(LoadOptions{AutoDetect: true}); err != nil {
		return internalOutput.OutputError(err, internalOutput.ErrSessionNotFound, outputFormat)
	}

	// 3. Extract tool calls
	tools := p.ExtractToolCalls()

	// 4. Filter errors only
	var errors []ErrorEntry
	for _, tool := range tools {
		if tool.Status == "error" || tool.Error != "" {
			errors = append(errors, ErrorEntry{
				UUID:      tool.UUID,
				Timestamp: tool.Timestamp,
				ToolName:  tool.ToolName,
				Error:     tool.Error,
				Signature: generateErrorSignature(tool.ToolName, tool.Error),
			})
		}
	}

	// 5. Sort by timestamp (deterministic output - lexicographic sort of ISO 8601)
	sort.Slice(errors, func(i, j int) bool {
		return errors[i].Timestamp < errors[j].Timestamp
	})

	// 6. Apply pagination if specified
	if limitFlag > 0 || offsetFlag > 0 {
		paginationConfig := filter.PaginationConfig{
			Limit:  limitFlag,
			Offset: offsetFlag,
		}
		errors = applyErrorPagination(errors, paginationConfig)
	}

	// 7. Check for empty results
	if len(errors) == 0 {
		return internalOutput.WarnNoResults(outputFormat)
	}

	// 8. Format output
	outputStr, formatErr := internalOutput.FormatOutput(errors, outputFormat)
	if formatErr != nil {
		return internalOutput.OutputError(formatErr, internalOutput.ErrInternalError, outputFormat)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}

// generateErrorSignature creates a simple signature: {tool}:{error_prefix}
func generateErrorSignature(toolName, errorText string) string {
	// Take first 50 chars of error for signature
	prefix := errorText
	if len(errorText) > 50 {
		prefix = errorText[:50]
	}

	// Normalize whitespace
	prefix = strings.Join(strings.Fields(prefix), " ")

	return fmt.Sprintf("%s:%s", toolName, prefix)
}

// applyErrorPagination applies pagination to error entries
func applyErrorPagination(errors []ErrorEntry, config filter.PaginationConfig) []ErrorEntry {
	start := config.Offset
	if start >= len(errors) {
		return []ErrorEntry{}
	}

	end := len(errors)
	if config.Limit > 0 {
		end = start + config.Limit
		if end > len(errors) {
			end = len(errors)
		}
	}

	return errors[start:end]
}
