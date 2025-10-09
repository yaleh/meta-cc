package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yaleh/meta-cc/internal/output"
)

var (
	sessionID    string
	projectPath  string
	outputFormat string
	sessionOnly  bool // Phase 13: Force session-only analysis (opt-out of project default)

	// Phase 9.1: Pagination and size estimation flags
	limitFlag        int
	offsetFlag       int
	estimateSizeFlag bool

	// Phase 9.2: Chunking flags
	chunkSizeFlag int
	outputDirFlag string

	// Phase 9.3: Field projection flags
	fieldsFlag         string
	ifErrorIncludeFlag string

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
		_ = cmd.Help()
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

// getGlobalOptions returns GlobalOptions from global flags
func getGlobalOptions() GlobalOptions {
	projPath := projectPath

	// Phase 13: 默认使用当前工作目录作为项目路径（principles.md §7）
	// 条件：
	//  - 未明确指定 --project
	//  - 未设置 --session 或 --session-only
	//  - 未设置环境变量 CC_SESSION_ID（用于测试和特殊场景）
	if projPath == "" && sessionID == "" && !sessionOnly && os.Getenv("CC_SESSION_ID") == "" {
		if cwd, err := os.Getwd(); err == nil {
			projPath = cwd
		}
		// 如果获取当前目录失败，保持空字符串，让 pipeline 处理错误
	}

	return GlobalOptions{
		SessionID:   sessionID,
		ProjectPath: projPath,
		SessionOnly: sessionOnly,
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Set version string with build info
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, BuildTime)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&sessionID, "session", "", "Session ID (or use $CC_SESSION_ID)")
	rootCmd.PersistentFlags().StringVar(&projectPath, "project", "", "Project path (defaults to current directory)")
	rootCmd.PersistentFlags().BoolVar(&sessionOnly, "session-only", false, "Analyze current session only (opt-out of project-level default)")
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
	_ = viper.BindPFlag("session", rootCmd.PersistentFlags().Lookup("session"))
	_ = viper.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project"))
	_ = viper.BindEnv("session", "CC_SESSION_ID")
	_ = viper.BindEnv("project", "CC_PROJECT_PATH")
}

func initConfig() {
	viper.AutomaticEnv()
}
