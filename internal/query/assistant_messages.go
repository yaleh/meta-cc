package query

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	mcerrors "github.com/yaleh/meta-cc/internal/errors"
	"github.com/yaleh/meta-cc/internal/parser"
)

type AssistantMessagesOptions struct {
	Pattern   string
	MinTools  int
	MaxTools  int
	MinTokens int
	MinLength int
	MaxLength int
	Limit     int
	Offset    int
	SortBy    string
	Reverse   bool
}

type AssistantMessage struct {
	TurnSequence  int                     `json:"turn_sequence"`
	UUID          string                  `json:"uuid"`
	Timestamp     string                  `json:"timestamp"`
	Model         string                  `json:"model"`
	ContentBlocks []AssistantContentBlock `json:"content_blocks"`
	TextLength    int                     `json:"text_length"`
	ToolUseCount  int                     `json:"tool_use_count"`
	TokensInput   int                     `json:"tokens_input"`
	TokensOutput  int                     `json:"tokens_output"`
	StopReason    string                  `json:"stop_reason,omitempty"`
}

type AssistantContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ToolName string `json:"tool_name,omitempty"`
}

type assistantMessageRaw struct {
	msg         AssistantMessage
	textContent string
}

func BuildAssistantMessages(entries []parser.SessionEntry, opts AssistantMessagesOptions) ([]AssistantMessage, error) {
	turnIndex := buildTurnIndex(entries)
	raw := extractAssistantMessages(entries, turnIndex)

	if opts.Pattern != "" {
		pattern, err := regexp.Compile(opts.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %w", mcerrors.ErrInvalidInput)
		}
		raw = filterAssistantMessagesByPattern(raw, pattern)
	}

	raw = filterAssistantMessagesByToolCount(raw, opts.MinTools, opts.MaxTools)
	raw = filterAssistantMessagesByTokens(raw, opts.MinTokens)
	raw = filterAssistantMessagesByLength(raw, opts.MinLength, opts.MaxLength)

	messages := flattenAssistantMessages(raw)
	sortAssistantMessages(messages, opts.SortBy, opts.Reverse)

	if opts.Offset > 0 {
		if opts.Offset < len(messages) {
			messages = messages[opts.Offset:]
		} else {
			messages = []AssistantMessage{}
		}
	}
	if opts.Limit > 0 && len(messages) > opts.Limit {
		messages = messages[:opts.Limit]
	}

	return messages, nil
}

func extractAssistantMessages(entries []parser.SessionEntry, turnIndex map[string]int) []assistantMessageRaw {
	var messages []assistantMessageRaw

	for _, entry := range entries {
		if entry.Type != "assistant" || entry.Message == nil {
			continue
		}

		var textLength int
		var toolUseCount int
		var blocks []AssistantContentBlock
		var textBuilder strings.Builder

		for _, block := range entry.Message.Content {
			switch block.Type {
			case "text":
				textLength += len(block.Text)
				textBuilder.WriteString(block.Text)
				blocks = append(blocks, AssistantContentBlock{Type: "text", Text: block.Text})
			case "tool_use":
				toolUseCount++
				toolName := ""
				if block.ToolUse != nil {
					toolName = block.ToolUse.Name
				}
				blocks = append(blocks, AssistantContentBlock{Type: "tool_use", ToolName: toolName})
			}
		}

		tokensInput, tokensOutput := extractTokenUsage(entry)

		message := AssistantMessage{
			TurnSequence:  turnIndex[entry.UUID],
			UUID:          entry.UUID,
			Timestamp:     entry.Timestamp,
			Model:         entry.Message.Model,
			ContentBlocks: blocks,
			TextLength:    textLength,
			ToolUseCount:  toolUseCount,
			TokensInput:   tokensInput,
			TokensOutput:  tokensOutput,
			StopReason:    entry.Message.StopReason,
		}

		messages = append(messages, assistantMessageRaw{
			msg:         message,
			textContent: textBuilder.String(),
		})
	}

	return messages
}

func extractTokenUsage(entry parser.SessionEntry) (int, int) {
	input := 0
	output := 0
	if entry.Message != nil && entry.Message.Usage != nil {
		if val, ok := entry.Message.Usage["input_tokens"].(float64); ok {
			input = int(val)
		}
		if val, ok := entry.Message.Usage["output_tokens"].(float64); ok {
			output = int(val)
		}
	}
	return input, output
}

func filterAssistantMessagesByPattern(messages []assistantMessageRaw, pattern *regexp.Regexp) []assistantMessageRaw {
	var filtered []assistantMessageRaw
	for _, msg := range messages {
		if pattern.MatchString(msg.textContent) {
			filtered = append(filtered, msg)
			continue
		}
		matched := false
		for _, block := range msg.msg.ContentBlocks {
			if block.Type == "text" && pattern.MatchString(block.Text) {
				matched = true
				break
			}
		}
		if matched {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

func filterAssistantMessagesByToolCount(messages []assistantMessageRaw, minTools, maxTools int) []assistantMessageRaw {
	if minTools == -1 && maxTools == -1 {
		return messages
	}

	var filtered []assistantMessageRaw
	for _, msg := range messages {
		count := msg.msg.ToolUseCount
		if minTools != -1 && count < minTools {
			continue
		}
		if maxTools != -1 && count > maxTools {
			continue
		}
		filtered = append(filtered, msg)
	}
	return filtered
}

func filterAssistantMessagesByTokens(messages []assistantMessageRaw, minTokens int) []assistantMessageRaw {
	if minTokens == -1 {
		return messages
	}

	var filtered []assistantMessageRaw
	for _, msg := range messages {
		if msg.msg.TokensOutput >= minTokens {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

func filterAssistantMessagesByLength(messages []assistantMessageRaw, minLength, maxLength int) []assistantMessageRaw {
	if minLength == -1 && maxLength == -1 {
		return messages
	}

	var filtered []assistantMessageRaw
	for _, msg := range messages {
		length := msg.msg.TextLength
		if minLength != -1 && length < minLength {
			continue
		}
		if maxLength != -1 && length > maxLength {
			continue
		}
		filtered = append(filtered, msg)
	}
	return filtered
}

func flattenAssistantMessages(raw []assistantMessageRaw) []AssistantMessage {
	result := make([]AssistantMessage, len(raw))
	for i, r := range raw {
		result[i] = r.msg
	}
	return result
}

func sortAssistantMessages(messages []AssistantMessage, sortBy string, reverse bool) {
	comparator := func(i, j int) bool {
		switch sortBy {
		case "timestamp":
			return messages[i].Timestamp < messages[j].Timestamp
		case "tool_use_count":
			return messages[i].ToolUseCount < messages[j].ToolUseCount
		case "text_length":
			return messages[i].TextLength < messages[j].TextLength
		default:
			return messages[i].TurnSequence < messages[j].TurnSequence
		}
	}

	sort.SliceStable(messages, func(i, j int) bool {
		less := comparator(i, j)
		if reverse {
			return !less
		}
		return less
	})
}

// -----------------------------------------------------------------------------
// Conversation helpers

type ConversationOptions struct {
	StartTurn     int
	EndTurn       int
	Pattern       string
	PatternTarget string
	MinDuration   int
	MaxDuration   int
	Limit         int
	Offset        int
	SortBy        string
	Reverse       bool
}

type ConversationTurn struct {
	TurnSequence     int               `json:"turn_sequence"`
	UserMessage      *UserMessage      `json:"user_message,omitempty"`
	AssistantMessage *AssistantMessage `json:"assistant_message,omitempty"`
	Duration         int               `json:"duration_ms"`
	Timestamp        string            `json:"timestamp"`
}

func BuildConversationTurns(entries []parser.SessionEntry, opts ConversationOptions) ([]ConversationTurn, error) {
	turnIndex := buildTurnIndex(entries)
	turns := buildConversationTurnList(entries, turnIndex)

	if opts.StartTurn != -1 || opts.EndTurn != -1 {
		turns = filterTurnsByRange(turns, opts.StartTurn, opts.EndTurn)
	}

	if opts.Pattern != "" {
		target := strings.ToLower(opts.PatternTarget)
		if target == "" {
			target = "any"
		}
		filtered, err := filterTurnsByPattern(turns, opts.Pattern, target)
		if err != nil {
			return nil, err
		}
		turns = filtered
	}

	turns = filterTurnsByDuration(turns, opts.MinDuration, opts.MaxDuration)
	sortConversationTurns(turns, opts.SortBy, opts.Reverse)

	if opts.Offset > 0 {
		if opts.Offset < len(turns) {
			turns = turns[opts.Offset:]
		} else {
			turns = []ConversationTurn{}
		}
	}
	if opts.Limit > 0 && len(turns) > opts.Limit {
		turns = turns[:opts.Limit]
	}

	return turns, nil
}

func buildConversationTurnList(entries []parser.SessionEntry, turnIndex map[string]int) []ConversationTurn {
	userByTurn, timestampByTurn := conversationUserMessages(entries, turnIndex)
	assistantByTurn := conversationAssistantMessages(entries, turnIndex)

	uniqueTurns := make(map[int]struct{})
	for turn := range userByTurn {
		uniqueTurns[turn] = struct{}{}
	}
	for turn := range assistantByTurn {
		uniqueTurns[turn] = struct{}{}
	}

	var turns []ConversationTurn
	for turn := range uniqueTurns {
		user := userByTurn[turn]
		assistant := assistantByTurn[turn]
		duration := calculateTurnDuration(user, assistant)
		timestamp := firstTimestamp(user, assistant, timestampByTurn[turn])
		turns = append(turns, ConversationTurn{
			TurnSequence:     turn,
			UserMessage:      user,
			AssistantMessage: assistant,
			Duration:         duration,
			Timestamp:        timestamp,
		})
	}

	return turns
}

func conversationUserMessages(entries []parser.SessionEntry, turnIndex map[string]int) (map[int]*UserMessage, map[int]string) {
	userByTurn := make(map[int]*UserMessage)
	turnTimestamps := make(map[int]string)

	for _, entry := range entries {
		if entry.Type != "user" || entry.Message == nil {
			continue
		}
		content := aggregateUserContent(entry.Message.Content)
		if content == "" || isSystemAssistantMessage(content) {
			continue
		}

		turn := turnIndex[entry.UUID]
		userByTurn[turn] = &UserMessage{
			TurnSequence: turn,
			UUID:         entry.UUID,
			Timestamp:    entry.Timestamp,
			Content:      content,
		}
		turnTimestamps[turn] = entry.Timestamp
	}

	return userByTurn, turnTimestamps
}

func aggregateUserContent(blocks []parser.ContentBlock) string {
	var content strings.Builder
	for _, block := range blocks {
		if block.Type == "text" {
			content.WriteString(block.Text)
		}
	}
	return content.String()
}

func conversationAssistantMessages(entries []parser.SessionEntry, turnIndex map[string]int) map[int]*AssistantMessage {
	assistantByTurn := make(map[int]*AssistantMessage)

	for _, entry := range entries {
		if entry.Type != "assistant" || entry.Message == nil {
			continue
		}

		blocks := make([]AssistantContentBlock, 0, len(entry.Message.Content))
		for _, block := range entry.Message.Content {
			if block.Type == "text" {
				blocks = append(blocks, AssistantContentBlock{Type: "text", Text: block.Text})
			}
		}

		assistantByTurn[turnIndex[entry.UUID]] = &AssistantMessage{
			TurnSequence:  turnIndex[entry.UUID],
			UUID:          entry.UUID,
			Timestamp:     entry.Timestamp,
			Model:         entry.Message.Model,
			ContentBlocks: blocks,
		}
	}

	return assistantByTurn
}

func filterTurnsByRange(turns []ConversationTurn, start, end int) []ConversationTurn {
	var filtered []ConversationTurn
	for _, turn := range turns {
		if start != -1 && turn.TurnSequence < start {
			continue
		}
		if end != -1 && turn.TurnSequence > end {
			continue
		}
		filtered = append(filtered, turn)
	}
	return filtered
}

func filterTurnsByPattern(turns []ConversationTurn, pattern, target string) ([]ConversationTurn, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", mcerrors.ErrInvalidInput)
	}

	var filtered []ConversationTurn
	for _, turn := range turns {
		match := false
		if target == "user" || target == "any" {
			if turn.UserMessage != nil && re.MatchString(turn.UserMessage.Content) {
				match = true
			}
		}
		if target == "assistant" || target == "any" {
			if turn.AssistantMessage != nil {
				for _, block := range turn.AssistantMessage.ContentBlocks {
					if block.Type == "text" && re.MatchString(block.Text) {
						match = true
						break
					}
				}
			}
		}
		if match {
			filtered = append(filtered, turn)
		}
	}
	return filtered, nil
}

func filterTurnsByDuration(turns []ConversationTurn, minDuration, maxDuration int) []ConversationTurn {
	if minDuration == -1 && maxDuration == -1 {
		return turns
	}

	var filtered []ConversationTurn
	for _, turn := range turns {
		if minDuration != -1 && turn.Duration < minDuration {
			continue
		}
		if maxDuration != -1 && turn.Duration > maxDuration {
			continue
		}
		filtered = append(filtered, turn)
	}
	return filtered
}

func calculateTurnDuration(user *UserMessage, assistant *AssistantMessage) int {
	if user == nil || assistant == nil {
		return 0
	}
	start, err1 := parseConversationTimestamp(user.Timestamp)
	end, err2 := parseConversationTimestamp(assistant.Timestamp)
	if err1 != nil || err2 != nil {
		return 0
	}
	return int(end.Sub(start).Milliseconds())
}

func firstTimestamp(user *UserMessage, assistant *AssistantMessage, fallback string) string {
	if user != nil && user.Timestamp != "" {
		return user.Timestamp
	}
	if assistant != nil && assistant.Timestamp != "" {
		return assistant.Timestamp
	}
	return fallback
}

func sortConversationTurns(turns []ConversationTurn, sortBy string, reverse bool) {
	comparator := func(i, j int) bool {
		switch sortBy {
		case "timestamp":
			return turns[i].Timestamp < turns[j].Timestamp
		case "duration":
			return turns[i].Duration < turns[j].Duration
		default:
			return turns[i].TurnSequence < turns[j].TurnSequence
		}
	}

	sort.SliceStable(turns, func(i, j int) bool {
		less := comparator(i, j)
		if reverse {
			return !less
		}
		return less
	})
}

func parseConversationTimestamp(ts string) (time.Time, error) {
	if ts == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}
	return time.Parse(time.RFC3339, ts)
}

// -----------------------------------------------------------------------------
// Utilities

func isSystemAssistantMessage(content string) bool {
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
