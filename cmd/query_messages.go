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
	queryMessagesMatch string
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
}

// UserMessage represents a user message with metadata
type UserMessage struct {
	UUID      string `json:"uuid"`
	Timestamp string `json:"timestamp"`
	Content   string `json:"content"`
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

	// Step 2: Extract user messages
	userMessages := extractUserMessages(entries)

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

	// Step 5: Apply offset
	if queryOffset > 0 {
		if queryOffset < len(userMessages) {
			userMessages = userMessages[queryOffset:]
		} else {
			userMessages = []UserMessage{}
		}
	}

	// Step 6: Apply limit
	if queryLimit > 0 && len(userMessages) > queryLimit {
		userMessages = userMessages[:queryLimit]
	}

	// Step 7: Format output
	var outputStr string
	var formatErr error

	switch outputFormat {
	case "json":
		outputStr, formatErr = output.FormatJSON(userMessages)
	case "md", "markdown":
		outputStr, formatErr = output.FormatMarkdown(userMessages)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	if formatErr != nil {
		return fmt.Errorf("failed to format output: %w", formatErr)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}

func extractUserMessages(entries []parser.SessionEntry) []UserMessage {
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
			messages = append(messages, UserMessage{
				UUID:      entry.UUID,
				Timestamp: entry.Timestamp,
				Content:   content,
			})
		}
	}

	return messages
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
