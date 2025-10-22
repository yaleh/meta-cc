package cmd

import (
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/spf13/cobra"
	mcerrors "github.com/yaleh/meta-cc/internal/errors"
	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/pkg/output"
)

var (
	queryConvStartTurn     int
	queryConvEndTurn       int
	queryConvPattern       string
	queryConvPatternTarget string
	queryConvMinDuration   int
	queryConvMaxDuration   int
)

var queryConversationCmd = &cobra.Command{
	Use:   "conversation",
	Short: "Query conversation turns (user+assistant pairs)",
	Long: `Query conversation turns with user and assistant message pairs.

Supports:
  - Turn range filtering (--start-turn, --end-turn)
  - Pattern matching (--pattern: regex on user/assistant content)
  - Pattern target (--pattern-target: user, assistant, any)
  - Duration filtering (--min-duration, --max-duration in milliseconds)
  - Timestamp sorting
  - Limit and pagination

Examples:
  meta-cc query conversation --start-turn 100 --end-turn 200
  meta-cc query conversation --min-duration 30000
  meta-cc query conversation --pattern "fix bug" --pattern-target user
  meta-cc query conversation --pattern "error" --min-duration 5000 --limit 20`,
	RunE: runQueryConversation,
}

func init() {
	queryCmd.AddCommand(queryConversationCmd)

	queryConversationCmd.Flags().IntVar(&queryConvStartTurn, "start-turn", -1, "Starting turn sequence")
	queryConversationCmd.Flags().IntVar(&queryConvEndTurn, "end-turn", -1, "Ending turn sequence")
	queryConversationCmd.Flags().StringVar(&queryConvPattern, "pattern", "", "Pattern to match (regex)")
	queryConversationCmd.Flags().StringVar(&queryConvPatternTarget, "pattern-target", "any", "Pattern target: user, assistant, any")
	queryConversationCmd.Flags().IntVar(&queryConvMinDuration, "min-duration", -1, "Minimum response duration (ms)")
	queryConversationCmd.Flags().IntVar(&queryConvMaxDuration, "max-duration", -1, "Maximum response duration (ms)")
}

// ConversationTurn represents a conversation turn with user and assistant messages
type ConversationTurn struct {
	TurnSequence     int               `json:"turn_sequence"`
	UserMessage      *UserMessage      `json:"user_message,omitempty"`
	AssistantMessage *AssistantMessage `json:"assistant_message,omitempty"`
	Duration         int               `json:"duration_ms"`
	Timestamp        string            `json:"timestamp"`
}

func runQueryConversation(cmd *cobra.Command, args []string) error {
	// Step 1: Initialize and load session using pipeline
	p := NewSessionPipeline(getGlobalOptions())
	if err := p.Load(LoadOptions{AutoDetect: true}); err != nil {
		return fmt.Errorf("failed to locate session: %w", err)
	}

	// Step 2: Build turn index and extract conversation turns
	entries := p.GetEntries()
	turnIndex := p.BuildTurnIndex()
	turns := buildConversationTurns(entries, turnIndex)

	// Step 3: Apply turn range filtering
	if queryConvStartTurn != -1 || queryConvEndTurn != -1 {
		turns = filterByTurnRange(turns, queryConvStartTurn, queryConvEndTurn)
	}

	// Step 4: Apply pattern matching
	if queryConvPattern != "" {
		var err error
		turns, err = filterByPatternCompiled(turns, queryConvPattern, queryConvPatternTarget)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}
	}

	// Step 5: Apply duration filtering
	if queryConvMinDuration != -1 || queryConvMaxDuration != -1 {
		turns = filterByDuration(turns, queryConvMinDuration, queryConvMaxDuration)
	}

	// Step 6: Apply default deterministic sorting (by turn sequence)
	sortConversationTurns(turns, "turn_sequence", false)

	// Step 6b: Apply custom sort if requested (overrides default)
	if querySortBy != "" {
		sortConversationTurns(turns, querySortBy, queryReverse)
	}

	// Step 7: Apply offset
	if queryOffset > 0 {
		if queryOffset < len(turns) {
			turns = turns[queryOffset:]
		} else {
			turns = []ConversationTurn{}
		}
	}

	// Step 8: Apply limit
	if queryLimit > 0 && len(turns) > queryLimit {
		turns = turns[:queryLimit]
	}

	// Step 9: Format output
	var outputStr string
	var formatErr error

	switch outputFormat {
	case "jsonl":
		outputStr, formatErr = output.FormatJSONL(turns)
	case "tsv":
		outputStr, formatErr = output.FormatTSV(turns)
	default:
		return fmt.Errorf("unsupported output format: %s (supported: jsonl, tsv): %w", outputFormat, mcerrors.ErrInvalidInput)
	}

	if formatErr != nil {
		return fmt.Errorf("failed to format output: %w", formatErr)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}

func buildConversationTurns(entries []parser.SessionEntry, turnIndex map[string]int) []ConversationTurn {
	userByTurn, turnTimestamps := conversationUserMessages(entries, turnIndex)
	assistantByTurn := conversationAssistantMessages(entries, turnIndex)
	turns := collectConversationTurns(userByTurn, assistantByTurn)

	var results []ConversationTurn
	for _, turn := range turns {
		user := userByTurn[turn]
		assistant := assistantByTurn[turn]
		duration := calculateTurnDuration(user, assistant)
		timestamp := firstTimestamp(user, assistant, turnTimestamps[turn])
		results = append(results, ConversationTurn{
			TurnSequence:     turn,
			UserMessage:      user,
			AssistantMessage: assistant,
			Duration:         duration,
			Timestamp:        timestamp,
		})
	}

	return results
}

func conversationUserMessages(entries []parser.SessionEntry, turnIndex map[string]int) (map[int]*UserMessage, map[int]string) {
	userByTurn := make(map[int]*UserMessage)
	turnTimestamps := make(map[int]string)

	for _, entry := range entries {
		if entry.Type != "user" || entry.Message == nil {
			continue
		}
		content := aggregateUserContent(entry.Message.Content)
		if content == "" || isSystemMessage(content) {
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
	var content string
	for _, block := range blocks {
		if block.Type == "text" {
			content += block.Text
		}
	}
	return content
}

func conversationAssistantMessages(entries []parser.SessionEntry, turnIndex map[string]int) map[int]*AssistantMessage {
	assistantByTurn := make(map[int]*AssistantMessage)

	for _, entry := range entries {
		if entry.Type != "assistant" || entry.Message == nil {
			continue
		}

		contentBlocks, textLength, toolCount := conversationAssistantBlocks(entry.Message.Content)
		tokensInput, tokensOutput := conversationTokenUsage(entry.Message.Usage)
		turn := turnIndex[entry.UUID]

		assistantByTurn[turn] = &AssistantMessage{
			TurnSequence:  turn,
			UUID:          entry.UUID,
			Timestamp:     entry.Timestamp,
			Model:         entry.Message.Model,
			ContentBlocks: contentBlocks,
			TextLength:    textLength,
			ToolUseCount:  toolCount,
			TokensInput:   tokensInput,
			TokensOutput:  tokensOutput,
			StopReason:    entry.Message.StopReason,
		}
	}

	return assistantByTurn
}

func conversationAssistantBlocks(blocks []parser.ContentBlock) ([]ContentBlock, int, int) {
	var contentBlocks []ContentBlock
	textLength := 0
	toolUseCount := 0

	for _, block := range blocks {
		switch block.Type {
		case "text":
			textLength += len(block.Text)
			contentBlocks = append(contentBlocks, ContentBlock{Type: "text", Text: block.Text})
		case "tool_use":
			toolUseCount++
			toolName := ""
			if block.ToolUse != nil {
				toolName = block.ToolUse.Name
			}
			contentBlocks = append(contentBlocks, ContentBlock{Type: "tool_use", ToolName: toolName})
		}
	}

	return contentBlocks, textLength, toolUseCount
}

func conversationTokenUsage(usage map[string]interface{}) (int, int) {
	if usage == nil {
		return 0, 0
	}

	input := 0
	if val, ok := usage["input_tokens"].(float64); ok {
		input = int(val)
	}

	outputTokens := 0
	if val, ok := usage["output_tokens"].(float64); ok {
		outputTokens = int(val)
	}

	return input, outputTokens
}

func collectConversationTurns(userByTurn map[int]*UserMessage, assistantByTurn map[int]*AssistantMessage) []int {
	turnSet := make(map[int]struct{})
	for turn := range userByTurn {
		turnSet[turn] = struct{}{}
	}
	for turn := range assistantByTurn {
		turnSet[turn] = struct{}{}
	}

	var turns []int
	for turn := range turnSet {
		turns = append(turns, turn)
	}
	return turns
}

func calculateTurnDuration(user *UserMessage, assistant *AssistantMessage) int {
	if user == nil || assistant == nil {
		return 0
	}
	return calculateDuration(user.Timestamp, assistant.Timestamp)
}

func firstTimestamp(user *UserMessage, assistant *AssistantMessage, defaultTimestamp string) string {
	if defaultTimestamp != "" {
		return defaultTimestamp
	}
	if assistant != nil {
		return assistant.Timestamp
	}
	if user != nil {
		return user.Timestamp
	}
	return ""
}

func calculateDuration(userTime, assistantTime string) int {
	userT, err1 := time.Parse(time.RFC3339, userTime)
	assistantT, err2 := time.Parse(time.RFC3339, assistantTime)

	if err1 != nil || err2 != nil {
		return 0
	}

	duration := assistantT.Sub(userT)
	return int(duration.Milliseconds())
}

func filterByTurnRange(turns []ConversationTurn, startTurn, endTurn int) []ConversationTurn {
	var filtered []ConversationTurn
	for _, turn := range turns {
		if startTurn != -1 && turn.TurnSequence < startTurn {
			continue
		}
		if endTurn != -1 && turn.TurnSequence > endTurn {
			continue
		}
		filtered = append(filtered, turn)
	}
	return filtered
}

func filterByPatternCompiled(turns []ConversationTurn, patternStr, target string) ([]ConversationTurn, error) {
	// Validate pattern first
	_, err := regexp.Compile(patternStr)
	if err != nil {
		return nil, err
	}
	return filterByPattern(turns, patternStr, target), nil
}

func filterByPattern(turns []ConversationTurn, patternStr, target string) []ConversationTurn {
	pattern, err := regexp.Compile(patternStr)
	if err != nil {
		return turns // Return unfiltered on error
	}

	var filtered []ConversationTurn

	for _, turn := range turns {
		match := false

		// Check user message
		if (target == "user" || target == "any") && turn.UserMessage != nil {
			if pattern.MatchString(turn.UserMessage.Content) {
				match = true
			}
		}

		// Check assistant message
		if (target == "assistant" || target == "any") && turn.AssistantMessage != nil {
			for _, block := range turn.AssistantMessage.ContentBlocks {
				if block.Type == "text" && pattern.MatchString(block.Text) {
					match = true
					break
				}
			}
		}

		if match {
			filtered = append(filtered, turn)
		}
	}

	return filtered
}

func filterByDuration(turns []ConversationTurn, minDuration, maxDuration int) []ConversationTurn {
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

func sortConversationTurns(turns []ConversationTurn, sortBy string, reverse bool) {
	sort.SliceStable(turns, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "turn_sequence":
			less = turns[i].TurnSequence < turns[j].TurnSequence
		case "timestamp":
			less = turns[i].Timestamp < turns[j].Timestamp
		case "duration":
			less = turns[i].Duration < turns[j].Duration
		default:
			// Default: sort by turn sequence (deterministic)
			less = turns[i].TurnSequence < turns[j].TurnSequence
		}

		if reverse {
			return !less
		}
		return less
	})
}
