package executor

import (
	"testing"

	"github.com/yaleh/meta-cc/internal/conversation"
)

// TestSessionToEntrySurfacesAggregateNotInput is the DIR-071 acceptance proof
// at the MCP output layer: an opaque provider aggregate (Codex SQLite
// threads.tokens_used) is surfaced on the session entry as explicitly-labeled
// aggregate_tokens/aggregate_source fields with provenance, and NEVER as
// input_tokens — so a session-level total cannot be mistaken for input usage.
func TestSessionToEntrySurfacesAggregateNotInput(t *testing.T) {
	s := conversation.Session{
		ID:       "s1",
		Provider: conversation.ProviderCodex,
		CWD:      "/tmp",
		TokenUsage: conversation.TokenUsage{
			AggregateTokens: 4242,
			AggregateSource: conversation.AggregateSourceCodexSQLite,
		},
	}
	entry := sessionToEntry(s)

	if entry["aggregate_tokens"] != 4242 {
		t.Fatalf("aggregate_tokens not surfaced: %#v", entry)
	}
	if entry["aggregate_source"] != conversation.AggregateSourceCodexSQLite {
		t.Fatalf("aggregate_source (provenance) not surfaced: %#v", entry)
	}
	if _, ok := entry["input_tokens"]; ok {
		t.Fatalf("opaque aggregate must NOT be surfaced as input_tokens: %#v", entry)
	}
}

// TestSessionToEntryOmitsAggregateWhenAbsent guards backward compatibility:
// sessions without an aggregate (e.g. every Claude session, or a Codex session
// whose threads row had no tokens_used) gain no new fields.
func TestSessionToEntryOmitsAggregateWhenAbsent(t *testing.T) {
	s := conversation.Session{ID: "s1", Provider: conversation.ProviderClaude, CWD: "/tmp"}
	entry := sessionToEntry(s)
	if _, ok := entry["aggregate_tokens"]; ok {
		t.Fatalf("aggregate_tokens must be omitted when absent: %#v", entry)
	}
	if _, ok := entry["aggregate_source"]; ok {
		t.Fatalf("aggregate_source must be omitted when absent: %#v", entry)
	}
}
