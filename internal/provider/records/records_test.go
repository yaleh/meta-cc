package records

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/yaleh/meta-cc/internal/conversation"
)

func TestNormalizeCodexTurnRecords(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	session := conversation.Session{
		ID:       "codex-session",
		Provider: conversation.ProviderCodex,
		CWD:      "/tmp/project",
		Model:    "gpt-5",
		TokenUsage: conversation.TokenUsage{
			InputTokens: 999,
		},
	}
	turns := []conversation.Turn{{
		ID:            "turn-1",
		UserText:      "hello",
		AssistantText: "ack",
		TokenUsage: conversation.TokenUsage{
			InputTokens:  10,
			OutputTokens: 3,
			CacheTokens:  2,
		},
		ToolCalls: []conversation.ToolCall{{
			ID:        "call-1",
			Name:      "apply_patch",
			Input:     json.RawMessage(`{"input":"patch"}`),
			Output:    "failed",
			IsError:   true,
			Timestamp: now,
		}},
		Timestamp: now,
	}}

	got := Normalize(session, turns)
	if len(got) != 3 {
		t.Fatalf("expected user, assistant, and tool result records, got %#v", got)
	}
	assistant, ok := got[1]["message"].(map[string]interface{})
	if !ok {
		t.Fatalf("assistant message missing: %#v", got[1])
	}
	usage, ok := assistant["usage"].(map[string]interface{})
	if !ok {
		t.Fatalf("assistant usage missing: %#v", assistant)
	}
	if usage["input_tokens"] != 10 || usage["output_tokens"] != 3 || usage["cache_tokens"] != 2 {
		t.Fatalf("unexpected usage: %#v", usage)
	}
	resultContent := got[2]["message"].(map[string]interface{})["content"].([]interface{})
	result := resultContent[0].(map[string]interface{})
	if result["status"] != "error" || result["error"] != "failed" {
		t.Fatalf("tool error result not normalized: %#v", result)
	}
}

func TestNormalizeCodexSessionTokenUsageDoesNotCreateUsageRecord(t *testing.T) {
	session := conversation.Session{
		ID:       "codex-session",
		Provider: conversation.ProviderCodex,
		CWD:      "/tmp/project",
		TokenUsage: conversation.TokenUsage{
			InputTokens: 999,
		},
	}
	turns := []conversation.Turn{{
		ID:            "turn-1",
		AssistantText: "ack",
		Timestamp:     time.Unix(1700000000, 0).UTC(),
	}}

	got := Normalize(session, turns)
	message := got[0]["message"].(map[string]interface{})
	if _, ok := message["usage"]; ok {
		t.Fatalf("codex sqlite tokens_used should not become per-turn usage: %#v", message)
	}
}

func TestNormalizeToolInputNullBecomesEmptyMap(t *testing.T) {
	session := conversation.Session{
		ID:       "codex-session",
		Provider: conversation.ProviderCodex,
		CWD:      "/tmp/project",
	}
	turns := []conversation.Turn{{
		ID: "turn-1",
		ToolCalls: []conversation.ToolCall{{
			ID:    "call-null",
			Name:  "empty_input",
			Input: json.RawMessage(`null`),
		}},
		Timestamp: time.Unix(1700000000, 0).UTC(),
	}}

	got := Normalize(session, turns)
	message := got[0]["message"].(map[string]interface{})
	content := message["content"].([]interface{})
	toolUse := content[0].(map[string]interface{})
	input, ok := toolUse["input"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected non-nil input map, got %#v", toolUse["input"])
	}
	if len(input) != 0 {
		t.Fatalf("expected empty input map, got %#v", input)
	}
}

// TestFilterSessionsForScope_SessionScopeDoesNotCrossProjects is a
// regression test for a bug where scope=="session" ignored projectPath
// entirely, sorted ALL Codex sessions across every project by CreatedAt, and
// returned only the single globally most-recent one. That meant a
// session-scope lookup for project A could silently return project B's
// session data whenever B's session was created more recently, even though
// the caller explicitly scoped the query to project A.
func TestFilterSessionsForScope_SessionScopeDoesNotCrossProjects(t *testing.T) {
	sessions := []conversation.Session{
		// Project B's session is the most recent globally, so a naive
		// "sort everything, take the top one" implementation (ignoring
		// projectPath) would incorrectly select it for project A's lookup.
		{ID: "project-b-newer", Provider: conversation.ProviderCodex, CWD: "/project-b", CreatedAt: time.Unix(2, 0)},
		{ID: "project-a-older", Provider: conversation.ProviderCodex, CWD: "/project-a", CreatedAt: time.Unix(1, 0)},
	}

	filtered := FilterSessionsForScope(sessions, "session", "/project-a", conversation.ProviderCodex)

	if len(filtered) != 1 || filtered[0].ID != "project-a-older" {
		t.Fatalf("session scope leaked across projects: got %#v, want only project-a-older", filtered)
	}

	// Symmetric check: project B's own session-scope lookup must still
	// resolve to its own (only) session.
	filteredB := FilterSessionsForScope(sessions, "session", "/project-b", conversation.ProviderCodex)
	if len(filteredB) != 1 || filteredB[0].ID != "project-b-newer" {
		t.Fatalf("unexpected session-scope result for project B: got %#v", filteredB)
	}
}

// TestFilterSessionsForScope_SessionScopeWithoutProjectPathKeepsMostRecent
// documents the deliberately-preserved fallback: when projectPath (or a
// session's CWD) is unset, the project filter is a no-op and scope=="session"
// still falls back to "most recent session overall". This is the path used
// when no working_dir/project scoping information is available at all.
func TestFilterSessionsForScope_SessionScopeWithoutProjectPathKeepsMostRecent(t *testing.T) {
	sessions := []conversation.Session{
		{ID: "older", Provider: conversation.ProviderCodex, CWD: "/project-a", CreatedAt: time.Unix(1, 0)},
		{ID: "newer", Provider: conversation.ProviderCodex, CWD: "/project-b", CreatedAt: time.Unix(2, 0)},
	}

	filtered := FilterSessionsForScope(sessions, "session", "", conversation.ProviderCodex)

	if len(filtered) != 1 || filtered[0].ID != "newer" {
		t.Fatalf("expected most-recent session when projectPath is unset, got %#v", filtered)
	}
}

func TestFilterSessionsForScopeDoesNotMutateInput(t *testing.T) {
	sessions := []conversation.Session{
		{ID: "keep", Provider: conversation.ProviderCodex, CWD: "/project", CreatedAt: time.Unix(2, 0)},
		{ID: "drop", Provider: conversation.ProviderCodex, CWD: "/other", CreatedAt: time.Unix(1, 0)},
	}
	originalSecond := sessions[1]

	filtered := FilterSessionsForScope(sessions, "project", "/project", conversation.ProviderCodex)

	if len(filtered) != 1 || filtered[0].ID != "keep" {
		t.Fatalf("unexpected filtered sessions: %#v", filtered)
	}
	if sessions[1].ID != originalSecond.ID || sessions[1].CWD != originalSecond.CWD || !sessions[1].CreatedAt.Equal(originalSecond.CreatedAt) {
		t.Fatalf("FilterSessionsForScope mutated input slice: got %#v, want %#v", sessions[1], originalSecond)
	}
}
