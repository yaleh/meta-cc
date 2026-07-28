package claude

import (
	"encoding/json"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/types"
)

type turnPair struct {
	user      *types.SessionEntry
	assistant *types.SessionEntry
}

func buildTurns(entries []types.SessionEntry) []turnPair {
	var pairs []turnPair
	assistantByParent := make(map[string]*types.SessionEntry)
	for i := range entries {
		if entries[i].Type == "assistant" && entries[i].ParentUUID != "" {
			assistantByParent[entries[i].ParentUUID] = &entries[i]
		}
	}

	for i := range entries {
		if entries[i].Type != "user" {
			continue
		}
		pairs = append(pairs, turnPair{
			user:      &entries[i],
			assistant: assistantByParent[entries[i].UUID],
		})
	}
	return pairs
}

// userByParentUUID indexes every user-typed entry by its parentUuid so that,
// given an assistant entry, the tool_result-carrying entry correlated with
// it can be found directly. In Claude's transcript chain, a tool_use call
// made by assistant entry A is answered by the NEXT user-typed entry whose
// parentUuid equals A's uuid — NOT by the user entry that triggered A
// (that entry's parentUuid points at the PREVIOUS assistant turn, and its
// content — if it carries tool_result blocks at all — answers that
// previous turn's calls, not A's). joinToolCalls uses this map (rather
// than turnPair.user) to correlate each tool_use with its own result; see
// the DIR-046 regression test in provider_test.go for the concrete
// off-by-one this fixes.
func userByParentUUID(entries []types.SessionEntry) map[string]*types.SessionEntry {
	m := make(map[string]*types.SessionEntry)
	for i := range entries {
		if entries[i].Type == "user" && entries[i].ParentUUID != "" {
			m[entries[i].ParentUUID] = &entries[i]
		}
	}
	return m
}

func joinToolCalls(pair turnPair, resultsByAssistantUUID map[string]*types.SessionEntry) []conversation.ToolCall {
	if pair.assistant == nil || pair.assistant.Message == nil {
		return nil
	}

	results := make(map[string]*types.ToolResult)
	if resultEntry := resultsByAssistantUUID[pair.assistant.UUID]; resultEntry != nil && resultEntry.Message != nil {
		for _, block := range resultEntry.Message.Content {
			if block.Type == "tool_result" && block.ToolResult != nil {
				results[block.ToolResult.ToolUseID] = block.ToolResult
			}
		}
	}

	var calls []conversation.ToolCall
	for _, block := range pair.assistant.Message.Content {
		if block.Type != "tool_use" || block.ToolUse == nil {
			continue
		}
		input, _ := json.Marshal(block.ToolUse.Input)
		ts, _ := time.Parse(time.RFC3339, pair.assistant.Timestamp)
		call := conversation.ToolCall{
			ID:        block.ToolUse.ID,
			Name:      block.ToolUse.Name,
			Input:     input,
			Timestamp: ts.UTC(),
		}
		if result := results[block.ToolUse.ID]; result != nil {
			call.Output = result.Content
			call.IsError = result.IsError || result.Status == "error" || result.Error != ""
		}
		calls = append(calls, call)
	}
	return calls
}

// itemsFromPair is the thin Claude-side adapter into the DIR-028 canonical
// Item stream: rather than reworking the established
// buildTurns/joinToolCalls parser pipeline, it maps the already-computed
// flattened fields (userText, assistantText, calls) into ordered Items.
// Claude's transcript format has no distinct commentary/final signal for
// assistant text (unlike Codex's Harmony-style channel field), so Phase is
// left unspecified rather than guessed — see conversation.AgentPhase.
func itemsFromPair(turnID, userText, assistantText string, calls []conversation.ToolCall, ts time.Time) []conversation.Item {
	var items []conversation.Item
	if userText != "" {
		items = append(items, conversation.Item{
			ID:        turnID + "-user",
			Kind:      conversation.ItemKindUserMessage,
			Role:      "user",
			Text:      userText,
			Timestamp: ts,
		})
	}
	if assistantText != "" {
		items = append(items, conversation.Item{
			ID:        turnID + "-assistant",
			Kind:      conversation.ItemKindAgentMessage,
			Role:      "assistant",
			Text:      assistantText,
			Timestamp: ts,
		})
	}
	for _, call := range calls {
		items = append(items, conversation.Item{
			ID:         call.ID,
			Kind:       conversation.ItemKindToolCall,
			ToolCallID: call.ID,
			ToolName:   call.Name,
			Input:      call.Input,
			Timestamp:  call.Timestamp,
		})
		if call.Output != "" || call.IsError {
			items = append(items, conversation.Item{
				Kind:       conversation.ItemKindToolResult,
				ToolCallID: call.ID,
				Output:     call.Output,
				IsError:    call.IsError,
				Timestamp:  call.Timestamp,
			})
		}
	}
	return items
}

func entryText(entry *types.SessionEntry) string {
	if entry == nil || entry.Message == nil {
		return ""
	}
	var text string
	for _, block := range entry.Message.Content {
		if block.Type == "text" && block.Text != "" {
			if text != "" {
				text += "\n"
			}
			text += block.Text
		}
	}
	return text
}
