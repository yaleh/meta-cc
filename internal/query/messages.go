package query

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/yaleh/meta-cc/internal/parser"
	pipelinepkg "github.com/yaleh/meta-cc/pkg/pipeline"
)

// ErrInvalidPattern indicates the provided regex pattern could not be compiled.
var ErrInvalidPattern = errors.New("query: invalid regex pattern")

// RunUserMessagesQuery extracts user messages from the session pipeline, applies
// pattern filtering, context expansion, sorting, and pagination according to options.
func RunUserMessagesQuery(opts UserMessagesQueryOptions) ([]UserMessage, error) {
	pipe := pipelinepkg.NewSessionPipeline(opts.Pipeline)
	if err := pipe.Load(pipelinepkg.LoadOptions{AutoDetect: true}); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSessionLoad, err)
	}

	entries := pipe.Entries()
	turnIndex := pipe.BuildTurnIndex()

	messages := extractUserMessages(entries, turnIndex)

	if opts.Pattern != "" {
		pattern, err := regexp.Compile(opts.Pattern)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidPattern, err)
		}

		var filtered []UserMessage
		for _, msg := range messages {
			if pattern.MatchString(msg.Content) {
				filtered = append(filtered, msg)
			}
		}
		messages = filtered
	}

	// Default deterministic sort by turn sequence before optional overrides
	sortUserMessages(messages, "turn_sequence", false)

	if opts.SortBy != "" {
		sortUserMessages(messages, opts.SortBy, opts.Reverse)
	}

	if opts.Context > 0 {
		messages = addContextToMessages(messages, entries, turnIndex, opts.Context)
	}

	messages = applyUserMessagePagination(messages, opts.Limit, opts.Offset)

	return messages, nil
}

func extractUserMessages(entries []parser.SessionEntry, turnIndex map[string]int) []UserMessage {
	var messages []UserMessage

	for _, entry := range entries {
		if entry.Type != "user" || entry.Message == nil {
			continue
		}

		var contentBuilder strings.Builder
		for _, block := range entry.Message.Content {
			if block.Type == "text" {
				contentBuilder.WriteString(block.Text)
			}
		}

		content := contentBuilder.String()
		if content == "" || isSystemMessage(content) {
			continue
		}

		turn := turnIndex[entry.UUID]

		messages = append(messages, UserMessage{
			TurnSequence: turn,
			UUID:         entry.UUID,
			Timestamp:    entry.Timestamp,
			Content:      content,
		})
	}

	return messages
}

func addContextToMessages(messages []UserMessage, entries []parser.SessionEntry, turnIndex map[string]int, window int) []UserMessage {
	if window <= 0 {
		return messages
	}

	entryByTurn := map[int]parser.SessionEntry{}
	for _, entry := range entries {
		if !entry.IsMessage() {
			continue
		}
		turn := turnIndex[entry.UUID]
		entryByTurn[turn] = entry
	}

	for i := range messages {
		turn := messages[i].TurnSequence

		messages[i].ContextBefore = collectContextEntries(entryByTurn, turn-window, turn-1)
		messages[i].ContextAfter = collectContextEntries(entryByTurn, turn+1, turn+window)
	}

	return messages
}

func collectContextEntries(entryByTurn map[int]parser.SessionEntry, start, end int) []ContextEntry {
	var context []ContextEntry

	for turn := start; turn <= end; turn++ {
		entry, ok := entryByTurn[turn]
		if !ok || entry.Message == nil {
			continue
		}

		summary := summarizeContent(entry.Message.Content)
		toolCalls := collectToolCalls(entry.Message.Content)

		context = append(context, ContextEntry{
			Turn:      turn,
			Role:      entry.Message.Role,
			Summary:   summary,
			ToolCalls: toolCalls,
		})
	}

	return context
}

func summarizeContent(blocks []parser.ContentBlock) string {
	var builder strings.Builder

	for _, block := range blocks {
		if block.Type == "text" {
			builder.WriteString(block.Text)
		}
	}

	summary := builder.String()
	if len(summary) > 120 {
		return summary[:120] + "..."
	}
	return summary
}

func collectToolCalls(blocks []parser.ContentBlock) []string {
	var tools []string
	for _, block := range blocks {
		if block.Type == "tool_use" && block.ToolUse != nil {
			tools = append(tools, block.ToolUse.Name)
		}
	}
	return tools
}

func sortUserMessages(messages []UserMessage, sortBy string, reverse bool) {
	sort.SliceStable(messages, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "turn_sequence":
			less = messages[i].TurnSequence < messages[j].TurnSequence
		case "timestamp":
			less = messages[i].Timestamp < messages[j].Timestamp
		case "uuid":
			less = messages[i].UUID < messages[j].UUID
		default:
			less = messages[i].TurnSequence < messages[j].TurnSequence
		}

		if reverse {
			return !less
		}
		return less
	})
}

func applyUserMessagePagination(messages []UserMessage, limit, offset int) []UserMessage {
	start := offset
	if start > len(messages) {
		return []UserMessage{}
	}

	end := len(messages)
	if limit > 0 && start+limit < end {
		end = start + limit
	}

	return messages[start:end]
}

func isSystemMessage(content string) bool {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return false
	}

	systemPrefixes := []string{
		"<command-message>",
		"<command-name>",
		"<command-args>",
		"<local-command",
		"Caveat:",
		"# meta-",
	}

	for _, prefix := range systemPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}

	return false
}
