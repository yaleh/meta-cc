package cmd

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/spf13/cobra"
	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/pkg/output"
)

var (
	queryAssistantPattern   string
	queryAssistantMinTools  int
	queryAssistantMaxTools  int
	queryAssistantMinTokens int
	queryAssistantMinLength int
	queryAssistantMaxLength int
)

var queryAssistantMessagesCmd = &cobra.Command{
	Use:   "assistant-messages",
	Short: "Query assistant messages",
	Long: `Query assistant messages with pattern matching and filtering.

Supports:
  - Pattern matching (--pattern: regex pattern on text content)
  - Tool usage filtering (--min-tools, --max-tools)
  - Token filtering (--min-tokens-output)
  - Length filtering (--min-length, --max-length)
  - Timestamp sorting
  - Limit and pagination

Examples:
  meta-cc query assistant-messages --pattern "fix.*bug"
  meta-cc query assistant-messages --min-tools 5
  meta-cc query assistant-messages --min-tokens-output 2000
  meta-cc query assistant-messages --pattern "error" --min-tools 2 --limit 10`,
	RunE: runQueryAssistantMessages,
}

func init() {
	queryCmd.AddCommand(queryAssistantMessagesCmd)

	queryAssistantMessagesCmd.Flags().StringVar(&queryAssistantPattern, "pattern", "", "Pattern to match text content (regex)")
	queryAssistantMessagesCmd.Flags().IntVar(&queryAssistantMinTools, "min-tools", -1, "Minimum tool use count")
	queryAssistantMessagesCmd.Flags().IntVar(&queryAssistantMaxTools, "max-tools", -1, "Maximum tool use count")
	queryAssistantMessagesCmd.Flags().IntVar(&queryAssistantMinTokens, "min-tokens-output", -1, "Minimum output tokens")
	queryAssistantMessagesCmd.Flags().IntVar(&queryAssistantMinLength, "min-length", -1, "Minimum text length")
	queryAssistantMessagesCmd.Flags().IntVar(&queryAssistantMaxLength, "max-length", -1, "Maximum text length")
}

// AssistantMessage represents an assistant message with metadata
type AssistantMessage struct {
	TurnSequence  int            `json:"turn_sequence"`
	UUID          string         `json:"uuid"`
	Timestamp     string         `json:"timestamp"`
	Model         string         `json:"model"`
	ContentBlocks []ContentBlock `json:"content_blocks"`
	TextLength    int            `json:"text_length"`
	ToolUseCount  int            `json:"tool_use_count"`
	TokensInput   int            `json:"tokens_input"`
	TokensOutput  int            `json:"tokens_output"`
	StopReason    string         `json:"stop_reason,omitempty"`
}

// ContentBlock is a simplified version of parser.ContentBlock for output
type ContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ToolName string `json:"tool_name,omitempty"`
}

func runQueryAssistantMessages(cmd *cobra.Command, args []string) error {
	// Step 1: Initialize and load session using pipeline
	p := NewSessionPipeline(getGlobalOptions())
	if err := p.Load(LoadOptions{AutoDetect: true}); err != nil {
		return fmt.Errorf("failed to locate session: %w", err)
	}

	// Step 2: Build turn index and extract assistant messages
	entries := p.GetEntries()
	turnIndex := p.BuildTurnIndex()
	messages := extractAssistantMessages(entries, turnIndex)

	// Step 3: Apply pattern matching
	if queryAssistantPattern != "" {
		pattern, err := regexp.Compile(queryAssistantPattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}
		messages = filterAssistantMessagesByPattern(messages, pattern)
	}

	// Step 4: Apply tool count filtering
	if queryAssistantMinTools != -1 || queryAssistantMaxTools != -1 {
		messages = filterAssistantMessagesByToolCount(messages, queryAssistantMinTools, queryAssistantMaxTools)
	}

	// Step 5: Apply token filtering
	if queryAssistantMinTokens != -1 {
		messages = filterAssistantMessagesByTokens(messages, queryAssistantMinTokens)
	}

	// Step 6: Apply length filtering
	if queryAssistantMinLength != -1 || queryAssistantMaxLength != -1 {
		messages = filterAssistantMessagesByLength(messages, queryAssistantMinLength, queryAssistantMaxLength)
	}

	// Step 7: Apply default deterministic sorting (by turn sequence)
	sortAssistantMessages(messages, "turn_sequence", false)

	// Step 7b: Apply custom sort if requested (overrides default)
	if querySortBy != "" {
		sortAssistantMessages(messages, querySortBy, queryReverse)
	}

	// Step 8: Apply offset
	if queryOffset > 0 {
		if queryOffset < len(messages) {
			messages = messages[queryOffset:]
		} else {
			messages = []AssistantMessage{}
		}
	}

	// Step 9: Apply limit
	if queryLimit > 0 && len(messages) > queryLimit {
		messages = messages[:queryLimit]
	}

	// Step 10: Format output
	var outputStr string
	var formatErr error

	switch outputFormat {
	case "jsonl":
		outputStr, formatErr = output.FormatJSONL(messages)
	case "tsv":
		outputStr, formatErr = output.FormatTSV(messages)
	default:
		return fmt.Errorf("unsupported output format: %s (supported: jsonl, tsv)", outputFormat)
	}

	if formatErr != nil {
		return fmt.Errorf("failed to format output: %w", formatErr)
	}

	fmt.Fprintln(cmd.OutOrStdout(), outputStr)
	return nil
}

func extractAssistantMessages(entries []parser.SessionEntry, turnIndex map[string]int) []AssistantMessage {
	var messages []AssistantMessage

	for _, entry := range entries {
		// Only process assistant entries
		if entry.Type != "assistant" {
			continue
		}

		// Skip entries without Message
		if entry.Message == nil {
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
		messages = append(messages, AssistantMessage{
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
		})
	}

	return messages
}

func filterAssistantMessagesByPattern(messages []AssistantMessage, pattern *regexp.Regexp) []AssistantMessage {
	var filtered []AssistantMessage
	for _, msg := range messages {
		// Check text content in content blocks
		for _, block := range msg.ContentBlocks {
			if block.Type == "text" && pattern.MatchString(block.Text) {
				filtered = append(filtered, msg)
				break
			}
		}
	}
	return filtered
}

func filterAssistantMessagesByToolCount(messages []AssistantMessage, minTools, maxTools int) []AssistantMessage {
	var filtered []AssistantMessage
	for _, msg := range messages {
		if minTools != -1 && msg.ToolUseCount < minTools {
			continue
		}
		if maxTools != -1 && msg.ToolUseCount > maxTools {
			continue
		}
		filtered = append(filtered, msg)
	}
	return filtered
}

func filterAssistantMessagesByTokens(messages []AssistantMessage, minTokens int) []AssistantMessage {
	var filtered []AssistantMessage
	for _, msg := range messages {
		if msg.TokensOutput >= minTokens {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

func filterAssistantMessagesByLength(messages []AssistantMessage, minLength, maxLength int) []AssistantMessage {
	var filtered []AssistantMessage
	for _, msg := range messages {
		if minLength != -1 && msg.TextLength < minLength {
			continue
		}
		if maxLength != -1 && msg.TextLength > maxLength {
			continue
		}
		filtered = append(filtered, msg)
	}
	return filtered
}

func sortAssistantMessages(messages []AssistantMessage, sortBy string, reverse bool) {
	sort.SliceStable(messages, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "turn_sequence":
			less = messages[i].TurnSequence < messages[j].TurnSequence
		case "timestamp":
			less = messages[i].Timestamp < messages[j].Timestamp
		case "tokens_output":
			less = messages[i].TokensOutput < messages[j].TokensOutput
		case "text_length":
			less = messages[i].TextLength < messages[j].TextLength
		case "tool_use_count":
			less = messages[i].ToolUseCount < messages[j].ToolUseCount
		default:
			// Default: sort by turn sequence (deterministic)
			less = messages[i].TurnSequence < messages[j].TurnSequence
		}

		if reverse {
			return !less
		}
		return less
	})
}
