package ftsindex

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/conversation"
)

// TestRawSearchableText_ToolSearchItems is the DIR-072 query-surface proof
// for tool search items: a search call indexes the same "name + input" text
// a function tool call would (the query IS first-class searchable content),
// and a search output indexes its output — so hits surface through the
// existing FTS path without any tool-stats conflation (these kinds never
// enter the legacy ToolCalls projection tool stats consume).
func TestRawSearchableText_ToolSearchItems(t *testing.T) {
	call := conversation.Item{Kind: conversation.ItemKindToolSearchCall, ToolName: "repo_search", Input: []byte(`{"query":"TODO"}`)}
	if got := rawSearchableText(call); got != `repo_search {"query":"TODO"}` {
		t.Fatalf("tool search call searchable text = %q", got)
	}
	callNoInput := conversation.Item{Kind: conversation.ItemKindToolSearchCall, ToolName: "repo_search"}
	if got := rawSearchableText(callNoInput); got != "repo_search" {
		t.Fatalf("input-less tool search call searchable text = %q", got)
	}
	output := conversation.Item{Kind: conversation.ItemKindToolSearchOutput, Output: "3 matches found"}
	if got := rawSearchableText(output); got != "3 matches found" {
		t.Fatalf("tool search output searchable text = %q", got)
	}
}

// TestRawSearchableText_MetadataItemsIndexNothing proves the DIR-072
// metadata families (world_state, settings_applied) contribute NO searchable
// text: they are bounded metadata, not conversation content — and settings
// values in particular must never reach the full-text index (they can carry
// credentials; the typed item never embeds them and the index must agree).
func TestRawSearchableText_MetadataItemsIndexNothing(t *testing.T) {
	ws := conversation.Item{Kind: conversation.ItemKindWorldState, WorldState: &conversation.WorldState{CWD: "/tmp/project", ApprovalPolicy: "never"}}
	if got := rawSearchableText(ws); got != "" {
		t.Fatalf("world_state item must index no searchable text, got %q", got)
	}
	settings := conversation.Item{Kind: conversation.ItemKindSettingsApplied, Settings: &conversation.SettingsApplied{Keys: []string{"api_key", "model"}}}
	if got := rawSearchableText(settings); got != "" {
		t.Fatalf("settings_applied item must index no searchable text, got %q", got)
	}
}
