package query

import (
	"sort"
	"strings"

	"github.com/yaleh/meta-cc/internal/parser"
)

type ProjectState struct {
	SessionID          string           `json:"session_id"`
	RecentFiles        []FileActivity   `json:"recent_files"`
	IncompleteStages   []IncompleteTask `json:"incomplete_stages,omitempty"`
	LastErrorFreeTurns int              `json:"last_error_free_turns"`
	CurrentFocus       string           `json:"current_focus"`
	RecentAchievements []string         `json:"recent_achievements"`
}

type FileActivity struct {
	Path             string   `json:"path"`
	LastModifiedTurn int      `json:"last_modified_turn"`
	Operations       []string `json:"operations"`
	EditCount        int      `json:"edit_count"`
}

type IncompleteTask struct {
	Phase           int    `json:"phase,omitempty"`
	Stage           string `json:"stage,omitempty"`
	Title           string `json:"title"`
	MentionedInTurn int    `json:"mentioned_in_turn"`
}

type ProjectStateOptions struct {
	IncludeIncomplete bool
}

func BuildProjectState(entries []parser.SessionEntry, opts ProjectStateOptions) *ProjectState {
	sessionID := ""
	if len(entries) > 0 {
		sessionID = entries[0].SessionID
	}

	turnIndex := buildTurnIndex(entries)
	recentFiles := extractRecentFiles(entries, turnIndex)

	var incomplete []IncompleteTask
	if opts.IncludeIncomplete {
		incomplete = extractIncompleteTasks(entries, turnIndex)
	}

	errorFree := calculateErrorFreeTurns(entries, turnIndex)
	focus := determineCurrentFocus(entries, turnIndex)
	achievements := extractRecentAchievements(entries, turnIndex)

	return &ProjectState{
		SessionID:          sessionID,
		RecentFiles:        recentFiles,
		IncompleteStages:   incomplete,
		LastErrorFreeTurns: errorFree,
		CurrentFocus:       focus,
		RecentAchievements: achievements,
	}
}

func extractRecentFiles(entries []parser.SessionEntry, turnIndex map[string]int) []FileActivity {
	fileMap := make(map[string]*FileActivity)

	for _, entry := range entries {
		if entry.Type != "file-history-snapshot" {
			continue
		}

		for _, msgEntry := range entries {
			if msgEntry.Message == nil {
				continue
			}
			for _, block := range msgEntry.Message.Content {
				if block.Type != "tool_use" || block.ToolUse == nil {
					continue
				}
				toolName := block.ToolUse.Name
				if toolName != "Read" && toolName != "Edit" && toolName != "Write" && toolName != "NotebookEdit" {
					continue
				}
				filePath, ok := block.ToolUse.Input["file_path"].(string)
				if !ok || filePath == "" {
					continue
				}
				if _, exists := fileMap[filePath]; !exists {
					fileMap[filePath] = &FileActivity{Path: filePath}
				}
				activity := fileMap[filePath]
				turn := turnIndex[msgEntry.UUID]
				if turn > activity.LastModifiedTurn {
					activity.LastModifiedTurn = turn
				}
				if !containsString(activity.Operations, toolName) {
					activity.Operations = append(activity.Operations, toolName)
				}
				if toolName == "Edit" || toolName == "Write" || toolName == "NotebookEdit" {
					activity.EditCount++
				}
			}
		}
	}

	var result []FileActivity
	for _, activity := range fileMap {
		result = append(result, *activity)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].LastModifiedTurn == result[j].LastModifiedTurn {
			return result[i].Path < result[j].Path
		}
		return result[i].LastModifiedTurn > result[j].LastModifiedTurn
	})

	if len(result) > 10 {
		result = result[:10]
	}

	return result
}

func extractIncompleteTasks(entries []parser.SessionEntry, turnIndex map[string]int) []IncompleteTask {
	var tasks []IncompleteTask

	for _, entry := range entries {
		if entry.Message == nil || entry.Message.Role != "assistant" {
			continue
		}
		for _, block := range entry.Message.Content {
			if block.Type != "text" || block.Text == "" {
				continue
			}
			textLower := strings.ToLower(block.Text)
			if strings.Contains(textLower, "incomplete") || strings.Contains(textLower, "todo") {
				tasks = append(tasks, IncompleteTask{
					Title:           block.Text,
					MentionedInTurn: turnIndex[entry.UUID],
				})
			}
		}
	}

	return tasks
}

func determineCurrentFocus(entries []parser.SessionEntry, turnIndex map[string]int) string {
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if entry.Message == nil || entry.Message.Role != "assistant" {
			continue
		}
		for _, block := range entry.Message.Content {
			if block.Type == "text" && strings.TrimSpace(block.Text) != "" {
				return block.Text
			}
		}
	}
	return ""
}

func calculateErrorFreeTurns(entries []parser.SessionEntry, turnIndex map[string]int) int {
	count := 0
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if entry.Message == nil || !entry.IsMessage() {
			continue
		}
		hasError := false
		for _, block := range entry.Message.Content {
			if block.Type == "tool_result" && block.ToolResult != nil && block.ToolResult.IsError {
				hasError = true
				break
			}
		}
		if hasError {
			break
		}
		count++
	}
	return count
}

func extractRecentAchievements(entries []parser.SessionEntry, turnIndex map[string]int) []string {
	var achievements []string
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if entry.Message == nil || entry.Message.Role != "assistant" {
			continue
		}
		for _, block := range entry.Message.Content {
			if block.Type != "text" || block.Text == "" {
				continue
			}
			textLower := strings.ToLower(block.Text)
			if strings.Contains(textLower, "completed") || strings.Contains(textLower, "implemented") {
				achievements = append(achievements, block.Text)
			}
		}
		if len(achievements) >= 5 {
			break
		}
	}
	return achievements
}

func containsString(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}
