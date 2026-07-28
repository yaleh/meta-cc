package codex

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/mcp/filters"
	"github.com/yaleh/meta-cc/internal/provider/codex/appserver"
	providerrecords "github.com/yaleh/meta-cc/internal/provider/records"
)

// userItemFixture builds an app-server "userMessage" ThreadItem carrying text.
func userItemFixture(t *testing.T, id, text string) appserver.ThreadItem {
	t.Helper()
	raw, err := json.Marshal(map[string]interface{}{
		"type": "userMessage",
		"id":   id,
		"content": []map[string]string{
			{"type": "text", "text": text},
		},
	})
	if err != nil {
		t.Fatalf("marshal userMessage fixture: %v", err)
	}
	var item appserver.ThreadItem
	if err := json.Unmarshal(raw, &item); err != nil {
		t.Fatalf("unmarshal userMessage fixture: %v", err)
	}
	return item
}

// agentItemFixture builds an app-server "agentMessage" ThreadItem carrying text.
func agentItemFixture(t *testing.T, id, text string) appserver.ThreadItem {
	t.Helper()
	raw, err := json.Marshal(map[string]interface{}{
		"type": "agentMessage",
		"id":   id,
		"text": text,
	})
	if err != nil {
		t.Fatalf("marshal agentMessage fixture: %v", err)
	}
	var item appserver.ThreadItem
	if err := json.Unmarshal(raw, &item); err != nil {
		t.Fatalf("unmarshal agentMessage fixture: %v", err)
	}
	return item
}

// turnFixture builds an app-server Turn with a user+agent message pair,
// mirroring the shape of a real thread/read(includeTurns) response.
func turnFixture(t *testing.T, id string, startedAt int64, userText, agentText string) appserver.Turn {
	t.Helper()
	ts := startedAt
	return appserver.Turn{
		ID:        id,
		Status:    "completed",
		StartedAt: &ts,
		Items: []appserver.ThreadItem{
			userItemFixture(t, id+"-u", userText),
			agentItemFixture(t, id+"-a", agentText),
		},
	}
}

// TestAppServerBackendContextWindowContract is the DIR-036 "both Codex
// backends obey the same public context-window contract" proof: a 5-turn
// app-server-backed thread (not a rollout file) matching the middle turn
// must produce the same bounded/ordered/correctly-flagged window as the
// rollout-backed equivalent (see
// internal/mcp/executor/codex_context_turns_e2e_test.go's rollout fixture
// tests), using the exact provider/session abstraction
// (Provider.GetSession/LoadTurns + providerrecords.Normalize) DIR-036 routes
// context loading through — not a rollout-file rescan.
func TestAppServerBackendContextWindowContract(t *testing.T) {
	sessionID := "appserver-context-sess"
	thread := appserver.Thread{
		ID:            sessionID,
		SessionID:     sessionID,
		CWD:           "/tmp/project",
		ModelProvider: "openai",
		Turns: []appserver.Turn{
			turnFixture(t, "turn-1", 1700000000, "turn one baseline", "ack one"),
			turnFixture(t, "turn-2", 1700000010, "turn two baseline", "ack two"),
			turnFixture(t, "turn-3", 1700000020, "measure 吞吐 rate please", "ack three"),
			turnFixture(t, "turn-4", 1700000030, "turn four baseline", "ack four"),
			turnFixture(t, "turn-5", 1700000040, "turn five baseline", "ack five"),
		},
	}

	src := &fakeThreadSource{thread: thread}
	backend := &appServerBackend{connect: connectFake(src, &noopCloser{}, nil)}
	p := newProvider(locator.NewCodexLocator(), ModeAppServer, backend)

	ctx := context.Background()

	loadNormalized := func(sid string) ([]interface{}, error) {
		session, err := p.GetSession(ctx, sid)
		if err != nil {
			return nil, err
		}
		turns, err := p.LoadTurns(ctx, sid)
		if err != nil {
			return nil, err
		}
		recs := providerrecords.Normalize(session, turns)
		out := make([]interface{}, len(recs))
		for i, r := range recs {
			out[i] = r
		}
		return out, nil
	}

	records, err := loadNormalized(sessionID)
	if err != nil {
		t.Fatalf("failed to load normalized records via app-server backend: %v", err)
	}
	if p.Backend() != "app_server" {
		t.Fatalf("expected app_server backend to have actually answered, got %q", p.Backend())
	}

	// Simulate the query's match step: role=user, pattern="吞吐".
	var matched []interface{}
	for _, r := range records {
		obj := r.(map[string]interface{})
		if obj["type"] != "user" {
			continue
		}
		msg, ok := obj["message"].(map[string]interface{})
		if !ok {
			continue
		}
		content, _ := msg["content"].(string)
		if strings.Contains(content, "吞吐") {
			matched = append(matched, obj)
		}
	}
	if len(matched) != 1 {
		t.Fatalf("expected exactly 1 matched record before context expansion, got %d: %#v", len(matched), matched)
	}

	result, warnings, err := filters.ExpandContextTurnsCanonical(matched, 1, loadNormalized, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if len(result) == 0 {
		t.Fatal("context_turns must not erase the app-server-backed match (DIR-036)")
	}

	matchCount := 0
	for _, entry := range result {
		obj := entry.(map[string]interface{})
		sid := obj["session_id"]
		if sid != sessionID {
			t.Fatalf("result crossed session boundary: %#v", obj)
		}
		if ctx, ok := obj["context"].(bool); ok && !ctx {
			matchCount++
		}
	}
	if matchCount != 1 {
		t.Fatalf("expected exactly 1 matched (context:false) record, got %d in %#v", matchCount, result)
	}
}
