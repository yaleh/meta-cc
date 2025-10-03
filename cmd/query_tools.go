package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/filter"
	"github.com/yale/meta-cc/internal/locator"
	"github.com/yale/meta-cc/internal/parser"
	"github.com/yale/meta-cc/pkg/output"
)

var (
	queryToolsStatus string
	queryToolsTool   string
	queryToolsWhere  string
)

var queryToolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Query tool calls",
	Long: `Query tool calls with advanced filtering options.

Supports filtering by:
  - Tool name (--tool)
  - Status (--status: success|error)
  - General condition (--where: "field=value")

Examples:
  meta-cc query tools --status error
  meta-cc query tools --tool Bash --limit 20
  meta-cc query tools --where "status=error" --sort-by timestamp
  meta-cc query tools --tool Edit --status error --output md`,
	RunE: runQueryTools,
}

func init() {
	queryCmd.AddCommand(queryToolsCmd)

	// Tool-specific filters
	queryToolsCmd.Flags().StringVar(&queryToolsStatus, "status", "", "Filter by status (success|error)")
	queryToolsCmd.Flags().StringVar(&queryToolsTool, "tool", "", "Filter by tool name")
	queryToolsCmd.Flags().StringVar(&queryToolsWhere, "where", "", "Filter condition (key=value)")
}

func runQueryTools(cmd *cobra.Command, args []string) error {
	// Step 1: Locate and parse session
	loc := locator.NewSessionLocator()
	sessionPath, err := loc.Locate(locator.LocateOptions{
		SessionID:   sessionID,
		ProjectPath: projectPath,
	})
	if err != nil {
		return fmt.Errorf("failed to locate session: %w", err)
	}

	sessionParser := parser.NewSessionParser(sessionPath)
	entries, err := sessionParser.ParseEntries()
	if err != nil {
		return fmt.Errorf("failed to parse session: %w", err)
	}

	// Step 2: Extract tool calls
	toolCalls := parser.ExtractToolCalls(entries)

	// Step 3: Apply filters
	filtered, err := applyToolFilters(toolCalls)
	if err != nil {
		return err
	}

	// Step 4: Sort if requested
	if querySortBy != "" {
		sortToolCalls(filtered, querySortBy, queryReverse)
	}

	// Step 5: Apply offset
	if queryOffset > 0 {
		if queryOffset < len(filtered) {
			filtered = filtered[queryOffset:]
		} else {
			filtered = []parser.ToolCall{} // Offset beyond length
		}
	}

	// Step 6: Apply limit
	if queryLimit > 0 && len(filtered) > queryLimit {
		filtered = filtered[:queryLimit]
	}

	// Step 7: Format output
	var outputStr string
	var formatErr error

	switch outputFormat {
	case "json":
		outputStr, formatErr = output.FormatJSON(filtered)
	case "md", "markdown":
		outputStr, formatErr = output.FormatMarkdown(filtered)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	if formatErr != nil {
		return fmt.Errorf("failed to format output: %w", formatErr)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}

func applyToolFilters(toolCalls []parser.ToolCall) ([]parser.ToolCall, error) {
	var result []parser.ToolCall

	// Apply --where filter first (with validation)
	if queryToolsWhere != "" {
		filtered, err := filter.ApplyWhere(toolCalls, queryToolsWhere, "tool_calls")
		if err != nil {
			return nil, fmt.Errorf("invalid where condition: %w", err)
		}
		toolCalls = filtered.([]parser.ToolCall)
	}

	// Apply individual flag filters
	for _, tc := range toolCalls {
		// Apply status filter
		if queryToolsStatus != "" {
			if queryToolsStatus == "error" {
				if tc.Status != "error" && tc.Error == "" {
					continue
				}
			} else if queryToolsStatus == "success" {
				if tc.Status == "error" || tc.Error != "" {
					continue
				}
			}
		}

		// Apply tool name filter
		if queryToolsTool != "" && tc.ToolName != queryToolsTool {
			continue
		}

		// If all filters pass, include this tool call
		result = append(result, tc)
	}

	return result, nil
}

func sortToolCalls(toolCalls []parser.ToolCall, sortBy string, reverse bool) {
	sort.Slice(toolCalls, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "tool":
			less = toolCalls[i].ToolName < toolCalls[j].ToolName
		case "status":
			less = toolCalls[i].Status < toolCalls[j].Status
		case "uuid":
			less = toolCalls[i].UUID < toolCalls[j].UUID
		default:
			// Default: sort by tool name
			less = toolCalls[i].ToolName < toolCalls[j].ToolName
		}

		if reverse {
			return !less
		}
		return less
	})
}
