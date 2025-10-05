package cmd

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/locator"
	"github.com/yale/meta-cc/internal/parser"
	"github.com/yale/meta-cc/pkg/output"
)

var (
	queryMessagesMatch   string
	queryMessagesContext int
)

var queryUserMessagesCmd = &cobra.Command{
	Use:   "user-messages",
	Short: "Query user messages",
	Long: `Query user messages with pattern matching.

Supports:
  - Pattern matching (--match: regex pattern)
  - Timestamp sorting
  - Limit and pagination

Examples:
  meta-cc query user-messages --match "fix.*bug"
  meta-cc query user-messages --match "error" --limit 10
  meta-cc query user-messages --sort-by timestamp --reverse`,
	RunE: runQueryUserMessages,
}

func init() {
	queryCmd.AddCommand(queryUserMessagesCmd)

	queryUserMessagesCmd.Flags().StringVar(&queryMessagesMatch, "match", "", "Match pattern (regex)")
	queryUserMessagesCmd.Flags().IntVar(&queryMessagesContext, "with-context", 0, "Include N turns before/after each match")
}

// UserMessage represents a user message with metadata
type UserMessage struct {
	TurnSequence  int            `json:"turn_sequence"`
	UUID          string         `json:"uuid"`
	Timestamp     string         `json:"timestamp"`
	Content       string         `json:"content"`
	ContextBefore []ContextEntry `json:"context_before,omitempty"`
	ContextAfter  []ContextEntry `json:"context_after,omitempty"`
}

// ContextEntry represents a turn's context (before or after)
type ContextEntry struct {
	Turn      int      `json:"turn"`
	Role      string   `json:"role"`
	Summary   string   `json:"summary"`
	ToolCalls []string `json:"tool_calls,omitempty"`
}

func runQueryUserMessages(cmd *cobra.Command, args []string) error {
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

	// Step 2: Build turn index and extract user messages
	turnIndex := buildTurnIndex(entries)
	userMessages := extractUserMessages(entries, turnIndex)

	// Step 3: Apply pattern matching
	if queryMessagesMatch != "" {
		pattern, err := regexp.Compile(queryMessagesMatch)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}

		var filtered []UserMessage
		for _, msg := range userMessages {
			if pattern.MatchString(msg.Content) {
				filtered = append(filtered, msg)
			}
		}
		userMessages = filtered
	}

	// Step 4: Sort if requested
	if querySortBy != "" {
		sortUserMessages(userMessages, querySortBy, queryReverse)
	}

	// Step 5: Add context if requested
	if queryMessagesContext > 0 {
		userMessages = addContextToMessages(userMessages, entries, turnIndex, queryMessagesContext)
	}

	// Step 6: Apply offset
	if queryOffset > 0 {
		if queryOffset < len(userMessages) {
			userMessages = userMessages[queryOffset:]
		} else {
			userMessages = []UserMessage{}
		}
	}

	// Step 7: Apply limit
	if queryLimit > 0 && len(userMessages) > queryLimit {
		userMessages = userMessages[:queryLimit]
	}

	// Step 8: Format output
	var outputStr string
	var formatErr error

	switch outputFormat {
	case "jsonl":
		outputStr, formatErr = output.FormatJSONL(userMessages)
	case "tsv":
		outputStr, formatErr = output.FormatTSV(userMessages)
	default:
		return fmt.Errorf("unsupported output format: %s (supported: jsonl, tsv)", outputFormat)
	}

	if formatErr != nil {
		return fmt.Errorf("failed to format output: %w", formatErr)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}

func extractUserMessages(entries []parser.SessionEntry, turnIndex map[string]int) []UserMessage {
	var messages []UserMessage

	for _, entry := range entries {
		// Only process user entries
		if entry.Type != "user" {
			continue
		}

		// Skip entries without Message
		if entry.Message == nil {
			continue
		}

		// Extract text content from content blocks
		var content string
		for _, block := range entry.Message.Content {
			if block.Type == "text" {
				content += block.Text
			}
		}

		// Only include if there's actual content
		if content != "" {
			turn := turnIndex[entry.UUID]
			messages = append(messages, UserMessage{
				TurnSequence: turn,
				UUID:         entry.UUID,
				Timestamp:    entry.Timestamp,
				Content:      content,
			})
		}
	}

	return messages
}

// buildTurnIndex builds a map from UUID to turn number
func buildTurnIndex(entries []parser.SessionEntry) map[string]int {
	turnIndex := make(map[string]int)
	turn := 0

	for _, entry := range entries {
		if entry.IsMessage() {
			turn++
			turnIndex[entry.UUID] = turn
		}
	}

	return turnIndex
}

// addContextToMessages adds context before and after each message
func addContextToMessages(messages []UserMessage, entries []parser.SessionEntry, turnIndex map[string]int, contextWindow int) []UserMessage {
	// Build reverse index: turn -> entry
	entryByTurn := make(map[int]parser.SessionEntry)
	for _, entry := range entries {
		if entry.IsMessage() {
			turn := turnIndex[entry.UUID]
			entryByTurn[turn] = entry
		}
	}

	// Add context to each message
	for i := range messages {
		msg := &messages[i]

		// Context before
		for j := 1; j <= contextWindow; j++ {
			prevTurn := msg.TurnSequence - j
			if entry, ok := entryByTurn[prevTurn]; ok {
				contextEntry := buildContextEntry(entry, prevTurn, turnIndex)
				msg.ContextBefore = append([]ContextEntry{contextEntry}, msg.ContextBefore...)
			}
		}

		// Context after
		for j := 1; j <= contextWindow; j++ {
			nextTurn := msg.TurnSequence + j
			if entry, ok := entryByTurn[nextTurn]; ok {
				contextEntry := buildContextEntry(entry, nextTurn, turnIndex)
				msg.ContextAfter = append(msg.ContextAfter, contextEntry)
			}
		}
	}

	return messages
}

// buildContextEntry builds a context entry from a session entry
func buildContextEntry(entry parser.SessionEntry, turn int, turnIndex map[string]int) ContextEntry {
	contextEntry := ContextEntry{
		Turn: turn,
		Role: entry.Type,
	}

	if entry.Message != nil {
		// Extract summary (first 100 chars)
		var content string
		for _, block := range entry.Message.Content {
			if block.Type == "text" {
				content += block.Text
			}
		}
		if len(content) > 100 {
			contextEntry.Summary = content[:100] + "..."
		} else {
			contextEntry.Summary = content
		}

		// Extract tool calls
		for _, block := range entry.Message.Content {
			if block.Type == "tool_use" && block.ToolUse != nil {
				contextEntry.ToolCalls = append(contextEntry.ToolCalls, block.ToolUse.Name)
			}
		}
	}

	return contextEntry
}

func sortUserMessages(messages []UserMessage, sortBy string, reverse bool) {
	sort.Slice(messages, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "timestamp":
			less = messages[i].Timestamp < messages[j].Timestamp
		case "uuid":
			less = messages[i].UUID < messages[j].UUID
		default:
			// Default: sort by timestamp
			less = messages[i].Timestamp < messages[j].Timestamp
		}

		if reverse {
			return !less
		}
		return less
	})
}
