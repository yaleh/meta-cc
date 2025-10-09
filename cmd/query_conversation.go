package cmd

import (
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/spf13/cobra"
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
		return fmt.Errorf("unsupported output format: %s (supported: jsonl, tsv)", outputFormat)
	}

	if formatErr != nil {
		return fmt.Errorf("failed to format output: %w", formatErr)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}

func buildConversationTurns(entries []parser.SessionEntry, turnIndex map[string]int) []ConversationTurn {
	// Build maps for user and assistant messages by turn
	userByTurn := make(map[int]*UserMessage)
	assistantByTurn := make(map[int]*AssistantMessage)
	turnTimestamps := make(map[int]string)

	// Extract user messages
	for _, entry := range entries {
		if entry.Type != "user" || entry.Message == nil {
			continue
		}

		// Extract text content
		var content string
		for _, block := range entry.Message.Content {
			if block.Type == "text" {
				content += block.Text
			}
		}

		// Skip system messages
		if content != "" && !isSystemMessage(content) {
			turn := turnIndex[entry.UUID]
			userByTurn[turn] = &UserMessage{
				TurnSequence: turn,
				UUID:         entry.UUID,
				Timestamp:    entry.Timestamp,
				Content:      content,
			}
			turnTimestamps[turn] = entry.Timestamp
		}
	}

	// Extract assistant messages
	for _, entry := range entries {
		if entry.Type != "assistant" || entry.Message == nil {
			continue
		}

		// Calculate metrics
		var textLength int
		var toolUseCount int
		var contentBlocks []ContentBlock

		for _, block := range entry.Message.Content {
			switch block.Type {
			case "text":
				textLength += len(block.Text)
				contentBlocks = append(contentBlocks, ContentBlock{
					Type: "text",
					Text: block.Text,
				})
			case "tool_use":
				toolUseCount++
				toolName := ""
				if block.ToolUse != nil {
					toolName = block.ToolUse.Name
				}
				contentBlocks = append(contentBlocks, ContentBlock{
					Type:     "tool_use",
					ToolName: toolName,
				})
			}
		}

		// Extract token counts
		tokensInput := 0
		tokensOutput := 0
		if entry.Message.Usage != nil {
			if val, ok := entry.Message.Usage["input_tokens"].(float64); ok {
				tokensInput = int(val)
			}
			if val, ok := entry.Message.Usage["output_tokens"].(float64); ok {
				tokensOutput = int(val)
			}
		}

		turn := turnIndex[entry.UUID]
		assistantByTurn[turn] = &AssistantMessage{
			TurnSequence:  turn,
			UUID:          entry.UUID,
			Timestamp:     entry.Timestamp,
			Model:         entry.Message.Model,
			ContentBlocks: contentBlocks,
			TextLength:    textLength,
			ToolUseCount:  toolUseCount,
			TokensInput:   tokensInput,
			TokensOutput:  tokensOutput,
			StopReason:    entry.Message.StopReason,
		}
	}

	// Build conversation turns
	turnSet := make(map[int]bool)
	for turn := range userByTurn {
		turnSet[turn] = true
	}
	for turn := range assistantByTurn {
		turnSet[turn] = true
	}

	var turns []ConversationTurn
	for turn := range turnSet {
		user := userByTurn[turn]
		assistant := assistantByTurn[turn]

		// Calculate duration if both messages exist
		duration := 0
		if user != nil && assistant != nil {
			duration = calculateDuration(user.Timestamp, assistant.Timestamp)
		}

		// Use user timestamp if available, else assistant timestamp
		timestamp := turnTimestamps[turn]
		if timestamp == "" && assistant != nil {
			timestamp = assistant.Timestamp
		}

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
