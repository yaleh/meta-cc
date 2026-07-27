package ftsindex

import (
	"strconv"
	"strings"

	"github.com/yaleh/meta-cc/internal/conversation"
)

// rawSearchableText extracts the pre-privacy-cap text to index for an Item,
// per its Kind. Unrecognized kinds (ItemKindUnknown, and anything future
// providers add) deliberately index nothing from Item.Raw: Raw is
// provenance for hydration, not vetted for secrets, so it is excluded from
// the searchable text by default (Contract: "secrets are not intentionally
// indexed").
func rawSearchableText(item conversation.Item) string {
	switch item.Kind {
	case conversation.ItemKindUserMessage, conversation.ItemKindAgentMessage, conversation.ItemKindReasoning:
		return item.Text
	case conversation.ItemKindToolCall:
		if len(item.Input) > 0 {
			return item.ToolName + " " + string(item.Input)
		}
		return item.ToolName
	case conversation.ItemKindToolResult:
		return item.Output
	case conversation.ItemKindCommandExecution:
		parts := make([]string, 0, 2)
		if item.Command != "" {
			parts = append(parts, item.Command)
		}
		if item.Output != "" {
			parts = append(parts, item.Output)
		}
		return strings.Join(parts, "\n")
	case conversation.ItemKindFileChange:
		return strings.Join(item.Paths, " ")
	case conversation.ItemKindWebSearch:
		return item.Query
	case conversation.ItemKindPlanUpdate:
		return strings.Join(item.PlanSteps, "\n")
	default:
		return ""
	}
}

// itemBody returns the privacy-capped searchable text for item, plus
// whether it was truncated. This is the single choke point both the
// indexer and its tests go through, so "an oversized tool output is never
// fully indexed" is guaranteed by construction rather than by convention.
func itemBody(item conversation.Item, limit int) (string, bool) {
	return truncateBody(rawSearchableText(item), limit)
}

// itemID returns a stable identity for an Item within a Turn: the
// provider-native ID when present, otherwise a synthetic
// "<turn_id>#<index>" fallback so every indexed row still has a unique,
// reconstructable (turn_id, item_id) key for hydration even when the
// source event carried no native identifier (see conversation.Item.ID's
// doc comment: "empty when the source has no stable identity").
func itemID(turnID string, idx int, item conversation.Item) string {
	if item.ID != "" {
		return item.ID
	}
	return turnID + "#" + strconv.Itoa(idx)
}
