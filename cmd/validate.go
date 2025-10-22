package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yaleh/meta-cc/internal/validation"
)

var (
	validateAPIFile  string
	validateAPIFast  bool
	validateAPIQuiet bool
	validateAPIJSON  bool
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate CLI surface areas",
}

var validateAPICmd = &cobra.Command{
	Use:   "api",
	Short: "Validate MCP server tool definitions",
	RunE: func(cmd *cobra.Command, args []string) error {
		tools, err := validation.ParseTools(validateAPIFile)
		if err != nil {
			return fmt.Errorf("failed to parse tools: %w", err)
		}

		validator := validation.NewValidator(validateAPIFast)
		report := validator.Validate(tools)

		reporter := validation.NewReporter(validateAPIQuiet, validateAPIJSON)
		reporter.SetWriter(cmd.OutOrStdout())
		reporter.Print(report)

		if report.Failed > 0 {
			return fmt.Errorf("validation failed (%d violations)", report.Failed)
		}

		return nil
	},
}

func init() {
	validateAPICmd.Flags().StringVar(&validateAPIFile, "file", "cmd/mcp-server/tools.go", "Path to tools.go")
	validateAPICmd.Flags().BoolVar(&validateAPIFast, "fast", true, "Run fast checks only")
	validateAPICmd.Flags().BoolVar(&validateAPIQuiet, "quiet", false, "Suppress output except errors")
	validateAPICmd.Flags().BoolVar(&validateAPIJSON, "json", false, "Output validation report as JSON")

	validateCmd.AddCommand(validateAPICmd)
	rootCmd.AddCommand(validateCmd)
}
