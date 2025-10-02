package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/filter"
	"github.com/yale/meta-cc/internal/locator"
	"github.com/yale/meta-cc/internal/parser"
	"github.com/yale/meta-cc/pkg/output"
)

var (
	extractType   string
	extractFilter string
)

// parseCmd represents the parse subcommand
var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parse Claude Code session files",
	Long: `Parse Claude Code session files and extract structured data.

Examples:
  meta-cc parse extract --type turns
  meta-cc parse extract --type tools --output md
  meta-cc parse extract --type tools --filter "status=error"`,
}

// parseExtractCmd represents the parse extract sub-subcommand
var parseExtractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract data from session",
	Long: `Extract structured data from Claude Code session files.

Supported types:
  - turns:  All conversation turns (user and assistant messages)
  - tools:  Tool calls with their results
  - errors: Failed tool calls and error messages`,
	RunE: runParseExtract,
}

func init() {
	// Add parse subcommand to root
	rootCmd.AddCommand(parseCmd)

	// Add extract sub-subcommand to parse
	parseCmd.AddCommand(parseExtractCmd)

	// extract subcommand parameters
	parseExtractCmd.Flags().StringVarP(&extractType, "type", "t", "turns", "Data type to extract: turns|tools|errors")
	parseExtractCmd.Flags().StringVarP(&extractFilter, "filter", "f", "", "Filter data (e.g., \"status=error\")")

	// --output parameter is already defined in root.go as global parameter
}

func runParseExtract(cmd *cobra.Command, args []string) error {
	// Step 1: Validate parameters
	validTypes := map[string]bool{
		"turns":  true,
		"tools":  true,
		"errors": true,
	}

	if !validTypes[extractType] {
		return fmt.Errorf("invalid type '%s': must be one of: turns, tools, errors", extractType)
	}

	// Step 2: Locate session file (using Phase 1 locator)
	loc := locator.NewSessionLocator()
	sessionPath, err := loc.Locate(locator.LocateOptions{
		SessionID:   sessionID,   // from global parameter
		ProjectPath: projectPath, // from global parameter
	})
	if err != nil {
		return fmt.Errorf("failed to locate session file: %w", err)
	}

	// Step 3: Parse session file (using Phase 2 parser)
	sessionParser := parser.NewSessionParser(sessionPath)
	entries, err := sessionParser.ParseEntries()
	if err != nil {
		return fmt.Errorf("failed to parse session file: %w", err)
	}

	// Step 4: Extract data based on type
	var data interface{}

	switch extractType {
	case "turns":
		data = entries
	case "tools":
		toolCalls := parser.ExtractToolCalls(entries)
		data = toolCalls
	case "errors":
		// Extract failed tool calls
		toolCalls := parser.ExtractToolCalls(entries)
		var errorCalls []parser.ToolCall
		for _, tc := range toolCalls {
			if tc.Status == "error" || tc.Error != "" {
				errorCalls = append(errorCalls, tc)
			}
		}
		data = errorCalls
	}

	// Step 4.5: Apply filter (if provided)
	if extractFilter != "" {
		filterObj, err := filter.ParseFilter(extractFilter)
		if err != nil {
			return fmt.Errorf("invalid filter: %w", err)
		}

		data = filter.ApplyFilter(data, filterObj)
	}

	// Step 5: Format output
	var outputStr string
	var formatErr error

	switch outputFormat {
	case "json":
		outputStr, formatErr = output.FormatJSON(data)
	case "md", "markdown":
		outputStr, formatErr = output.FormatMarkdown(data)
	case "csv":
		outputStr, formatErr = output.FormatCSV(data)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	if formatErr != nil {
		return fmt.Errorf("failed to format output: %w", formatErr)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)

	return nil
}
