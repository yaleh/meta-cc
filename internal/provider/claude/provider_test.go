package claude

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/locator"
)

func TestProviderID(t *testing.T) {
	p := NewProvider(locator.NewSessionLocator(), ".")
	if got := p.ID(); got != conversation.ProviderClaude {
		t.Fatalf("ID() = %s", got)
	}
}

func TestIsAvailable(t *testing.T) {
	root := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", root)
	if !NewProvider(locator.NewSessionLocator(), ".").IsAvailable(context.Background()) {
		t.Fatalf("expected available")
	}

	t.Setenv("META_CC_PROJECTS_ROOT", filepath.Join(root, "missing"))
	if NewProvider(locator.NewSessionLocator(), ".").IsAvailable(context.Background()) {
		t.Fatalf("expected unavailable")
	}
}

func TestListSessionsAndLoadTurns(t *testing.T) {
	root := t.TempDir()
	project := t.TempDir()
	resolvedProject, err := filepath.EvalSymlinks(project)
	if err != nil {
		t.Fatal(err)
	}
	projectDir := filepath.Join(root, strings.NewReplacer("\\", "-", "/", "-", ":", "-").Replace(resolvedProject))
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "tests", "fixtures", "sample-session.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	sessionFile := filepath.Join(projectDir, "sample.jsonl")
	if err := os.WriteFile(sessionFile, data, 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("META_CC_PROJECTS_ROOT", root)
	p := NewProvider(locator.NewSessionLocator(), project)
	sessions, err := p.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 1 || sessions[0].Provider != conversation.ProviderClaude {
		t.Fatalf("unexpected sessions: %#v", sessions)
	}

	turns, err := p.LoadTurns(context.Background(), sessions[0].ID)
	if err != nil {
		t.Fatalf("LoadTurns: %v", err)
	}
	if len(turns) != 2 && len(turns) != 1 {
		t.Fatalf("unexpected turns: %#v", turns)
	}
	if len(turns) > 0 && len(turns[0].ToolCalls) > 0 && turns[0].ToolCalls[0].Name != "Grep" {
		t.Fatalf("unexpected tool call: %#v", turns[0].ToolCalls[0])
	}
}

// TestGetSessionEnforcesWorkingDirBoundary is the DIR-032 cwd-boundary
// regression test for a real gap findSessionFile had: it called
// locator.FromSessionID (a global, unscoped search across every
// project-hash directory) and returned whatever it found with no
// comparison against p.workingDir — the same class of cross-project leak
// DIR-030's adversarial audit found and fixed on the
// provider_query.go/ExecuteQueryForSession path, just not yet exercised on
// this constructor-level GetSession path. This seeds two distinct projects
// with their own sessions and proves a Provider scoped to project A cannot
// resolve project B's session_id via GetSession, even though
// FromSessionID alone would happily find it.
func TestGetSessionEnforcesWorkingDirBoundary(t *testing.T) {
	root := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", root)

	data, err := os.ReadFile(filepath.Join("..", "..", "..", "tests", "fixtures", "sample-session.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	const fixtureSessionID = "6a32f273-191a-49c8-a5fc-a5dcba08531a"

	// seedProject writes a copy of the fixture under its own project-hash
	// directory with a distinct session ID substituted in, so the two
	// projects' sessions are genuinely different sessions (not two copies
	// of the same one) — the realistic shape of the cross-project leak
	// this test guards against.
	seedProject := func(newSessionID string) (resolvedProject string) {
		project := t.TempDir()
		resolved, err := filepath.EvalSymlinks(project)
		if err != nil {
			t.Fatal(err)
		}
		projectDir := filepath.Join(root, strings.NewReplacer("\\", "-", "/", "-", ":", "-").Replace(resolved))
		if err := os.MkdirAll(projectDir, 0o755); err != nil {
			t.Fatal(err)
		}
		content := strings.ReplaceAll(string(data), fixtureSessionID, newSessionID)
		if err := os.WriteFile(filepath.Join(projectDir, newSessionID+".jsonl"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return resolved
	}

	projectA := seedProject("session-in-project-a")
	projectB := seedProject("session-in-project-b")

	// A Provider scoped to project A must NOT be able to resolve project
	// B's session via GetSession — before the DIR-032 fix,
	// locator.FromSessionID's unscoped global search would have found it
	// anyway.
	pA := NewProvider(locator.NewSessionLocator(), projectA)
	if _, err := pA.GetSession(context.Background(), "session-in-project-b"); err == nil {
		t.Fatalf("expected GetSession to reject a session_id outside the configured working_dir boundary, got no error")
	}

	// Sanity: the same lookup DOES succeed when correctly scoped to project B.
	pB := NewProvider(locator.NewSessionLocator(), projectB)
	if _, err := pB.GetSession(context.Background(), "session-in-project-b"); err != nil {
		t.Fatalf("expected GetSession to succeed within the correct project boundary, got %v", err)
	}
	// And project A's own session remains reachable from project A.
	if _, err := pA.GetSession(context.Background(), "session-in-project-a"); err != nil {
		t.Fatalf("expected GetSession to succeed for project A's own session, got %v", err)
	}
}

// TestLoadTurnsCorrelatesToolResultFromFollowingUserEntry is a DIR-046
// regression test for a join-direction bug in buildTurns/joinToolCalls: a
// tool_use call's result lives in the NEXT user-typed JSONL entry (the one
// whose parentUuid equals the assistant's uuid), never in the user entry
// that precedes/triggers that assistant turn. Before the fix,
// joinToolCalls read tool_result blocks off turnPair.user (the trigger,
// which for a multi-round-trip session carries the PREVIOUS tool call's
// results, not this assistant's), so results[thisToolUse.ID] was never
// found and every ToolCall.Output/IsError silently defaulted to ""/false —
// which in turn made Normalize (records.go) drop the tool_result record
// entirely (`if call.Output == "" && !call.IsError { continue }`),
// starving handleQueryToolErrors/handleQueryTools' status filter under
// provider="all" of virtually all Claude-sourced error/success signal
// (see internal/mcp/executor's DIR-046 fixture test for the end-to-end
// symptom). This test seeds two full round trips in one session — a
// successful Read and a failed Bash — and asserts both ToolCalls carry
// their OWN (not a neighboring call's) Output/IsError.
func TestLoadTurnsCorrelatesToolResultFromFollowingUserEntry(t *testing.T) {
	root := t.TempDir()
	project := t.TempDir()
	resolvedProject, err := filepath.EvalSymlinks(project)
	if err != nil {
		t.Fatal(err)
	}
	projectDir := filepath.Join(root, strings.NewReplacer("\\", "-", "/", "-", ":", "-").Replace(resolvedProject))
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}

	const sessionID = "multi-round-trip-session"
	lines := []string{
		`{"type":"user","uuid":"u1","parentUuid":null,"timestamp":"2026-01-01T00:00:00Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"user","content":[{"type":"text","text":"read a file, then run a failing command"}]}}`,
		`{"type":"assistant","uuid":"a1","parentUuid":"u1","timestamp":"2026-01-01T00:00:01Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"call-read","name":"Read","input":{"file_path":"/tmp/foo"}}]}}`,
		`{"type":"user","uuid":"u2","parentUuid":"a1","timestamp":"2026-01-01T00:00:02Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"call-read","is_error":false,"content":"file contents"}]}}`,
		`{"type":"assistant","uuid":"a2","parentUuid":"u2","timestamp":"2026-01-01T00:00:03Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"call-bash","name":"Bash","input":{"command":"exit 1"}}]}}`,
		`{"type":"user","uuid":"u3","parentUuid":"a2","timestamp":"2026-01-01T00:00:04Z","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"call-bash","is_error":true,"content":"command failed"}]}}`,
	}
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("META_CC_PROJECTS_ROOT", root)
	p := NewProvider(locator.NewSessionLocator(), resolvedProject)
	turns, err := p.LoadTurns(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("LoadTurns: %v", err)
	}

	var allCalls []conversation.ToolCall
	for _, turn := range turns {
		allCalls = append(allCalls, turn.ToolCalls...)
	}

	callByID := make(map[string]conversation.ToolCall)
	for _, call := range allCalls {
		callByID[call.ID] = call
	}

	readCall, ok := callByID["call-read"]
	if !ok {
		t.Fatalf("expected a ToolCall for call-read, got calls: %#v", allCalls)
	}
	if readCall.IsError {
		t.Fatalf("call-read must be correlated with its OWN successful result (is_error=false), got IsError=true")
	}
	if readCall.Output != "file contents" {
		t.Fatalf("call-read.Output = %q, want %q (its own tool_result, not call-bash's or empty)", readCall.Output, "file contents")
	}

	bashCall, ok := callByID["call-bash"]
	if !ok {
		t.Fatalf("expected a ToolCall for call-bash, got calls: %#v", allCalls)
	}
	if !bashCall.IsError {
		t.Fatalf("call-bash must be correlated with its OWN failed result (is_error=true), got IsError=false")
	}
	if bashCall.Output != "command failed" {
		t.Fatalf("call-bash.Output = %q, want %q (its own tool_result, not call-read's or empty)", bashCall.Output, "command failed")
	}
}
