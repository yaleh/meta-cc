package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yaleh/meta-cc/internal/query"
)

var (
	projectStateIncludeTasks bool
)

var queryProjectStateCmd = &cobra.Command{
	Use:   "project-state",
	Short: "Query current project state from session",
	RunE:  runQueryProjectState,
}

func init() {
	queryProjectStateCmd.Flags().BoolVar(&projectStateIncludeTasks, "include-incomplete-tasks", true, "Include incomplete tasks analysis")
	queryCmd.AddCommand(queryProjectStateCmd)
}

func runQueryProjectState(cmd *cobra.Command, args []string) error {
	p := NewSessionPipeline(getGlobalOptions())
	if err := p.Load(LoadOptions{AutoDetect: true}); err != nil {
		return fmt.Errorf("failed to locate session: %w", err)
	}

	entries := p.GetEntries()
	state := query.BuildProjectState(entries, query.ProjectStateOptions{IncludeIncomplete: projectStateIncludeTasks})

	if outputFormat == "md" {
		return outputProjectStateMarkdown(cmd, state)
	}

	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(state)
}

func outputProjectStateMarkdown(cmd *cobra.Command, state *query.ProjectState) error {
	var sb strings.Builder

	sb.WriteString("# Project State\n\n")
	sb.WriteString(fmt.Sprintf("**Session ID**: `%s`\n\n", state.SessionID))

	if len(state.RecentFiles) > 0 {
		sb.WriteString("## Recent Files\n\n")
		sb.WriteString("| Path | Last Turn | Operations | Edit Count |\n")
		sb.WriteString("|------|-----------|------------|------------|\n")
		for _, file := range state.RecentFiles {
			sb.WriteString(fmt.Sprintf("| %s | %d | %s | %d |\n", file.Path, file.LastModifiedTurn, strings.Join(file.Operations, ", "), file.EditCount))
		}
		sb.WriteString("\n")
	}

	if len(state.IncompleteStages) > 0 {
		sb.WriteString("## Incomplete Tasks\n\n")
		sb.WriteString("| Turn | Title |\n")
		sb.WriteString("|------|-------|\n")
		for _, task := range state.IncompleteStages {
			sb.WriteString(fmt.Sprintf("| %d | %s |\n", task.MentionedInTurn, task.Title))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("**Last Error-Free Turns**: %d\n\n", state.LastErrorFreeTurns))
	if state.CurrentFocus != "" {
		sb.WriteString("## Current Focus\n\n")
		sb.WriteString(state.CurrentFocus + "\n\n")
	}

	if len(state.RecentAchievements) > 0 {
		sb.WriteString("## Recent Achievements\n\n")
		for _, achievement := range state.RecentAchievements {
			sb.WriteString("- " + achievement + "\n")
		}
		sb.WriteString("\n")
	}

	fmt.Fprint(cmd.OutOrStdout(), sb.String())
	return nil
}
