package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yale/meta-cc/internal/output"
)

var (
	cfgFile      string
	sessionID    string
	projectPath  string
	outputFormat string

	// Phase 9.1: Pagination and size estimation flags
	limitFlag        int
	offsetFlag       int
	estimateSizeFlag bool

	// Phase 9.2: Chunking flags
	chunkSizeFlag int
	outputDirFlag string

	// Phase 9.3: Field projection flags
	fieldsFlag          string
	ifErrorIncludeFlag  string

	// Phase 9.4: Compact output format flags
	summaryFirstFlag bool
	topNFlag         int
)

// Build information (injected during build via -ldflags)
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "meta-cc",
	Short: "Meta-Cognition tool for Claude Code",
	Long: `meta-cc analyzes Claude Code session history to provide
metacognitive insights and workflow optimization.`,
	SilenceErrors: true, // We handle errors ourselves
	SilenceUsage:  true, // Don't show usage on errors
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		// Check if it's an ExitCodeError
		if exitErr, ok := err.(*output.ExitCodeError); ok {
			// ExitCodeError already has the message printed by Cobra
			// We just need to exit with the appropriate code
			os.Exit(exitErr.Code)
		}
		// For other errors, Cobra will handle them (exit 1)
		return err
	}
	return nil
}

func init() {
	cobra.OnInitialize(initConfig)

	// Set version string with build info
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, BuildTime)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&sessionID, "session", "", "Session ID (or use $CC_SESSION_ID)")
	rootCmd.PersistentFlags().StringVar(&projectPath, "project", "", "Project path")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "jsonl", "Output format: jsonl|tsv")

	// Phase 9.1: Pagination and size estimation flags
	rootCmd.PersistentFlags().IntVar(&limitFlag, "limit", 0, "Limit output to N records (0 = no limit)")
	rootCmd.PersistentFlags().IntVar(&offsetFlag, "offset", 0, "Skip first M records")
	rootCmd.PersistentFlags().BoolVar(&estimateSizeFlag, "estimate-size", false, "Estimate output size without generating full output")

	// Phase 9.2: Chunking flags
	rootCmd.PersistentFlags().IntVar(&chunkSizeFlag, "chunk-size", 0, "Split output into chunks of N records (0 = no chunking)")
	rootCmd.PersistentFlags().StringVar(&outputDirFlag, "output-dir", "", "Output directory for chunks (required with --chunk-size)")

	// Phase 9.3: Field projection flags
	rootCmd.PersistentFlags().StringVar(&fieldsFlag, "fields", "", "Output only specified fields (comma-separated, e.g., 'UUID,ToolName,Status')")
	rootCmd.PersistentFlags().StringVar(&ifErrorIncludeFlag, "if-error-include", "", "Include additional fields for error records (comma-separated)")

	// Phase 9.4: Compact output format flags
	rootCmd.PersistentFlags().BoolVar(&summaryFirstFlag, "summary-first", false, "Output summary before detailed records")
	rootCmd.PersistentFlags().IntVar(&topNFlag, "top", 0, "Show only top N detailed records (requires --summary-first, 0 = all)")

	// Bind environment variables
	viper.BindPFlag("session", rootCmd.PersistentFlags().Lookup("session"))
	viper.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project"))
	viper.BindEnv("session", "CC_SESSION_ID")
	viper.BindEnv("project", "CC_PROJECT_PATH")
}

func initConfig() {
	viper.AutomaticEnv()
}
