package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	sessionID    string
	projectPath  string
	outputFormat string
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
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Set version string with build info
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, BuildTime)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&sessionID, "session", "", "Session ID (or use $CC_SESSION_ID)")
	rootCmd.PersistentFlags().StringVar(&projectPath, "project", "", "Project path")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "json", "Output format: json|md|csv")

	// Bind environment variables
	viper.BindPFlag("session", rootCmd.PersistentFlags().Lookup("session"))
	viper.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project"))
	viper.BindEnv("session", "CC_SESSION_ID")
	viper.BindEnv("project", "CC_PROJECT_PATH")
}

func initConfig() {
	viper.AutomaticEnv()
}
