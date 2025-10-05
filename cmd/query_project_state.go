package cmd

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/locator"
	"github.com/yale/meta-cc/internal/parser"
)

var (
	projectStateIncludeTasks bool
)

// queryProjectStateCmd represents the project-state query command
var queryProjectStateCmd = &cobra.Command{
	Use:   "project-state",
	Short: "Query current project state from session",
	Long: `Query current project state including recent files, incomplete tasks, and session quality.

This command analyzes the session to extract:
- Recent files modified (from file-history-snapshot)
- Incomplete stages/tasks mentioned
- Error-free turn count
- Current focus areas

Example:
  meta-cc query project-state --output json
  meta-cc query project-state --include-incomplete-tasks --output md`,
	RunE: runQueryProjectState,
}

func init() {
	queryProjectStateCmd.Flags().BoolVar(&projectStateIncludeTasks, "include-incomplete-tasks", true, "Include incomplete tasks analysis")

	queryCmd.AddCommand(queryProjectStateCmd)
}

func runQueryProjectState(cmd *cobra.Command, args []string) error {
	// Locate session file
	loc := locator.NewSessionLocator()
	sessionPath, err := loc.Locate(locator.LocateOptions{
		SessionID:   sessionID,
		ProjectPath: projectPath, // from global parameter
		SessionOnly: sessionOnly, // Phase 13: opt-out of project default

	})
	if err != nil {
		return fmt.Errorf("failed to locate session: %w", err)
	}

	// Read and parse session file
	sessionParser := parser.NewSessionParser(sessionPath)
	entries, err := sessionParser.ParseEntries()
	if err != nil {
		return fmt.Errorf("failed to read session file: %w", err)
	}

	// Build project state
	state, err := buildProjectState(entries, projectStateIncludeTasks)
	if err != nil {
		return fmt.Errorf("failed to build project state: %w", err)
	}

	// Output result
	if outputFormat == "md" {
		return outputProjectStateMarkdown(cmd, state)
	}

	// JSON output (default)
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(state)
}

// ProjectState represents the current project state
type ProjectState struct {
	SessionID          string           `json:"session_id"`
	RecentFiles        []FileActivity   `json:"recent_files"`
	IncompleteStages   []IncompleteTask `json:"incomplete_stages,omitempty"`
	LastErrorFreeTurns int              `json:"last_error_free_turns"`
	CurrentFocus       string           `json:"current_focus"`
	RecentAchievements []string         `json:"recent_achievements"`
}

// FileActivity represents a file's activity
type FileActivity struct {
	Path             string   `json:"path"`
	LastModifiedTurn int      `json:"last_modified_turn"`
	Operations       []string `json:"operations"`
	EditCount        int      `json:"edit_count"`
}

// IncompleteTask represents an incomplete stage or task
type IncompleteTask struct {
	Phase           int    `json:"phase,omitempty"`
	Stage           string `json:"stage,omitempty"`
	Title           string `json:"title"`
	MentionedInTurn int    `json:"mentioned_in_turn"`
}

func buildProjectState(entries []parser.SessionEntry, includeTasks bool) (*ProjectState, error) {
	// Extract session ID
	sessionID := ""
	if len(entries) > 0 {
		sessionID = entries[0].SessionID
	}

	// Build turn index
	turnIndex := buildTurnIndex(entries)

	// Extract recent files from file-history-snapshot
	recentFiles := extractRecentFiles(entries, turnIndex)

	// Extract incomplete tasks if requested
	var incompleteTasks []IncompleteTask
	if includeTasks {
		incompleteTasks = extractIncompleteTasks(entries, turnIndex)
	}

	// Calculate error-free turns
	errorFreeTurns := calculateErrorFreeTurns(entries, turnIndex)

	// Determine current focus
	currentFocus := determineCurrentFocus(entries, turnIndex)

	// Extract recent achievements
	achievements := extractRecentAchievements(entries, turnIndex)

	return &ProjectState{
		SessionID:          sessionID,
		RecentFiles:        recentFiles,
		IncompleteStages:   incompleteTasks,
		LastErrorFreeTurns: errorFreeTurns,
		CurrentFocus:       currentFocus,
		RecentAchievements: achievements,
	}, nil
}

// extractRecentFiles extracts recently modified files from file-history-snapshot entries
func extractRecentFiles(entries []parser.SessionEntry, turnIndex map[string]int) []FileActivity {
	fileMap := make(map[string]*FileActivity)

	for _, entry := range entries {
		// Look for file-history-snapshot entries
		if entry.Type != "file-history-snapshot" {
			continue
		}

		// Parse file operations from the entry
		// Note: file-history-snapshot entries contain file paths in a specific format
		// We'll need to extract this from the entry data
		// For now, use a simple heuristic based on tool calls
		for _, msgEntry := range entries {
			if msgEntry.Message == nil {
				continue
			}

			for _, block := range msgEntry.Message.Content {
				if block.Type == "tool_use" && block.ToolUse != nil {
					toolName := block.ToolUse.Name
					if toolName == "Read" || toolName == "Edit" || toolName == "Write" {
						// Extract file_path from input
						if filePath, ok := block.ToolUse.Input["file_path"].(string); ok {
							if _, exists := fileMap[filePath]; !exists {
								fileMap[filePath] = &FileActivity{
									Path:       filePath,
									Operations: []string{},
									EditCount:  0,
								}
							}

							activity := fileMap[filePath]
							msgTurn := turnIndex[msgEntry.UUID]

							// Update last modified turn
							if msgTurn > activity.LastModifiedTurn {
								activity.LastModifiedTurn = msgTurn
							}

							// Track operation
							if !contains(activity.Operations, toolName) {
								activity.Operations = append(activity.Operations, toolName)
							}

							// Count edits
							if toolName == "Edit" || toolName == "Write" {
								activity.EditCount++
							}
						}
					}
				}
			}
		}
	}

	// Convert map to slice and sort by last modified turn
	var result []FileActivity
	for _, activity := range fileMap {
		result = append(result, *activity)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].LastModifiedTurn > result[j].LastModifiedTurn
	})

	// Limit to top 10
	if len(result) > 10 {
		result = result[:10]
	}

	return result
}

// extractIncompleteTasks extracts incomplete tasks from user messages
func extractIncompleteTasks(entries []parser.SessionEntry, turnIndex map[string]int) []IncompleteTask {
	var tasks []IncompleteTask

	// Patterns to match task mentions
	stagePattern := regexp.MustCompile(`(?i)Stage\s+(\d+)\.(\d+)`)
	phasePattern := regexp.MustCompile(`(?i)Phase\s+(\d+)`)
	todoPattern := regexp.MustCompile(`(?i)(?:TODO|待办|未完成|implement|添加|实现):\s*(.+)`)

	for _, entry := range entries {
		if entry.Type != "user" || entry.Message == nil {
			continue
		}

		turn := turnIndex[entry.UUID]

		// Extract text content
		var content string
		for _, block := range entry.Message.Content {
			if block.Type == "text" {
				content += block.Text
			}
		}

		// Match stage patterns
		if matches := stagePattern.FindStringSubmatch(content); len(matches) >= 3 {
			tasks = append(tasks, IncompleteTask{
				Phase:           atoi(matches[1]),
				Stage:           fmt.Sprintf("%s.%s", matches[1], matches[2]),
				Title:           extractTaskTitle(content),
				MentionedInTurn: turn,
			})
		}

		// Match phase patterns
		if matches := phasePattern.FindStringSubmatch(content); len(matches) >= 2 {
			tasks = append(tasks, IncompleteTask{
				Phase:           atoi(matches[1]),
				Title:           extractTaskTitle(content),
				MentionedInTurn: turn,
			})
		}

		// Match TODO patterns
		if matches := todoPattern.FindStringSubmatch(content); len(matches) >= 2 {
			tasks = append(tasks, IncompleteTask{
				Title:           strings.TrimSpace(matches[1]),
				MentionedInTurn: turn,
			})
		}
	}

	// Remove duplicates and sort by turn (most recent first)
	tasks = deduplicateTasks(tasks)
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].MentionedInTurn > tasks[j].MentionedInTurn
	})

	// Limit to top 10
	if len(tasks) > 10 {
		tasks = tasks[:10]
	}

	return tasks
}

// calculateErrorFreeTurns calculates consecutive error-free turns from the end
func calculateErrorFreeTurns(entries []parser.SessionEntry, turnIndex map[string]int) int {
	// Find all error turns
	errorTurns := make(map[int]bool)

	for _, entry := range entries {
		if entry.Message == nil {
			continue
		}

		for _, block := range entry.Message.Content {
			if block.Type == "tool_result" && block.ToolResult != nil {
				if block.ToolResult.Status == "error" || block.ToolResult.Error != "" {
					turn := turnIndex[entry.UUID]
					errorTurns[turn] = true
				}
			}
		}
	}

	// Count from the last turn backwards
	maxTurn := len(turnIndex)
	errorFreeTurns := 0

	for turn := maxTurn; turn >= 1; turn-- {
		if errorTurns[turn] {
			break
		}
		errorFreeTurns++
	}

	return errorFreeTurns
}

// determineCurrentFocus determines the current focus area
func determineCurrentFocus(entries []parser.SessionEntry, turnIndex map[string]int) string {
	// Analyze last 5 user messages to determine focus
	userMessages := []string{}

	for i := len(entries) - 1; i >= 0 && len(userMessages) < 5; i-- {
		entry := entries[i]
		if entry.Type != "user" || entry.Message == nil {
			continue
		}

		var content string
		for _, block := range entry.Message.Content {
			if block.Type == "text" {
				content += block.Text
			}
		}

		if content != "" {
			userMessages = append(userMessages, content)
		}
	}

	// Look for common themes
	allText := strings.Join(userMessages, " ")
	allText = strings.ToLower(allText)

	themes := map[string][]string{
		"Testing":        {"test", "测试", "unit test", "e2e"},
		"Implementation": {"implement", "实现", "添加", "add", "create"},
		"Bug fixing":     {"fix", "修复", "bug", "error", "问题"},
		"Documentation":  {"doc", "文档", "readme", "comment"},
		"Refactoring":    {"refactor", "重构", "optimize", "优化"},
		"Phase":          {"phase", "stage", "阶段"},
	}

	for theme, keywords := range themes {
		for _, keyword := range keywords {
			if strings.Contains(allText, keyword) {
				return theme
			}
		}
	}

	return "General development"
}

// extractRecentAchievements extracts recent achievements from assistant messages
func extractRecentAchievements(entries []parser.SessionEntry, turnIndex map[string]int) []string {
	var achievements []string

	// Look for completion patterns in last 10 assistant messages
	achievementPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:completed|完成|finished).*?Stage\s+\d+\.\d+`),
		regexp.MustCompile(`(?i)all tests (?:pass|passed|passing|通过)`),
		regexp.MustCompile(`(?i)successfully (?:implemented|created|added)`),
		regexp.MustCompile(`(?i)(?:实现|添加|完成)了.*`),
	}

	assistantMsgs := 0
	for i := len(entries) - 1; i >= 0 && assistantMsgs < 10; i-- {
		entry := entries[i]
		if entry.Type != "assistant" || entry.Message == nil {
			continue
		}

		assistantMsgs++

		var content string
		for _, block := range entry.Message.Content {
			if block.Type == "text" {
				content += block.Text
			}
		}

		for _, pattern := range achievementPatterns {
			if match := pattern.FindString(content); match != "" {
				// Clean up and add
				achievement := strings.TrimSpace(match)
				if len(achievement) > 100 {
					achievement = achievement[:100] + "..."
				}
				achievements = append(achievements, achievement)
				break // One achievement per message
			}
		}
	}

	// Limit to top 5
	if len(achievements) > 5 {
		achievements = achievements[:5]
	}

	return achievements
}

// Helper functions
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func atoi(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

func extractTaskTitle(content string) string {
	// Extract first line or first 100 chars as title
	lines := strings.Split(content, "\n")
	title := lines[0]
	if len(title) > 100 {
		title = title[:100] + "..."
	}
	return strings.TrimSpace(title)
}

func deduplicateTasks(tasks []IncompleteTask) []IncompleteTask {
	seen := make(map[string]bool)
	var result []IncompleteTask

	for _, task := range tasks {
		key := fmt.Sprintf("%d-%s-%s", task.Phase, task.Stage, task.Title)
		if !seen[key] {
			seen[key] = true
			result = append(result, task)
		}
	}

	return result
}

func outputProjectStateMarkdown(cmd *cobra.Command, state *ProjectState) error {
	var sb strings.Builder

	sb.WriteString("# Project State\n\n")
	sb.WriteString(fmt.Sprintf("**Session ID**: %s\n\n", state.SessionID))

	// Recent files
	sb.WriteString("## Recent Files\n\n")
	if len(state.RecentFiles) == 0 {
		sb.WriteString("No recent file activity.\n\n")
	} else {
		sb.WriteString("| File | Last Modified Turn | Operations | Edit Count |\n")
		sb.WriteString("|------|-------------------|------------|------------|\n")
		for _, file := range state.RecentFiles {
			sb.WriteString(fmt.Sprintf("| %s | %d | %s | %d |\n",
				file.Path, file.LastModifiedTurn, strings.Join(file.Operations, ", "), file.EditCount))
		}
		sb.WriteString("\n")
	}

	// Incomplete stages
	if len(state.IncompleteStages) > 0 {
		sb.WriteString("## Incomplete Stages/Tasks\n\n")
		sb.WriteString("| Phase | Stage | Title | Mentioned in Turn |\n")
		sb.WriteString("|-------|-------|-------|-------------------|\n")
		for _, task := range state.IncompleteStages {
			phase := ""
			if task.Phase > 0 {
				phase = fmt.Sprintf("%d", task.Phase)
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %d |\n",
				phase, task.Stage, task.Title, task.MentionedInTurn))
		}
		sb.WriteString("\n")
	}

	// Session quality
	sb.WriteString("## Session Quality\n\n")
	sb.WriteString(fmt.Sprintf("- **Error-Free Turns**: %d\n", state.LastErrorFreeTurns))
	sb.WriteString(fmt.Sprintf("- **Current Focus**: %s\n\n", state.CurrentFocus))

	// Recent achievements
	if len(state.RecentAchievements) > 0 {
		sb.WriteString("## Recent Achievements\n\n")
		for _, achievement := range state.RecentAchievements {
			sb.WriteString(fmt.Sprintf("- %s\n", achievement))
		}
	}

	fmt.Fprint(cmd.OutOrStdout(), sb.String())
	return nil
}
