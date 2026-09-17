package claude

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/conversation"
	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/testutil"
)

// seedProjectDir creates the project-hash transcript directory that mirrors
// how Claude Code stores session JSONL under META_CC_PROJECTS_ROOT, returning
// the resolved project path (for NewProvider) and the on-disk hash directory
// (for writing session files into).
func seedProjectDir(t *testing.T, root string) (resolvedProject, projectDir string) {
	t.Helper()
	project := t.TempDir()
	resolved, err := filepath.EvalSymlinks(project)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, strings.NewReplacer("\\", "-", "/", "-", ":", "-").Replace(resolved))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return resolved, dir
}

// stubSessionJSONL returns the on-disk shape of a session Claude Code created
// but never exchanged a turn in: only mode/permission-mode/system metadata
// entries, no user/assistant messages. These are valid JSON but carry no
// queryable content, and (before the fix) poisoned every project-wide query.
func stubSessionJSONL(sessionID string) []byte {
	lines := []string{
		`{"type":"mode","sessionId":"` + sessionID + `","mode":"default"}`,
		`{"type":"permission-mode","sessionId":"` + sessionID + `","permissionMode":"default"}`,
		`{"type":"system","sessionId":"` + sessionID + `","subtype":"init"}`,
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

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

// TestListSessionsSkipsZeroMessageStub is the regression test for the
// project-wide-query poisoning bug: a "stub" session file — one Claude Code
// writes at session start containing only mode/permission-mode/system
// metadata entries and no user/assistant messages — made sessionFromFile
// return "no Claude entries", and ListSessions propagated that error
// immediately, failing the ENTIRE project listing (and thus query_sessions /
// query_session_content under provider=all) even though every other session
// file was healthy. This seeds one valid session alongside one stub and
// asserts ListSessions skips the stub and returns the valid session with no
// error. This extends the DIR-030 "one bad session must not erase the rest"
// guarantee — previously honored only at the LoadTurns stage in records.Build
// — down to the ListSessions stage, which failed before that tolerance was
// ever reached.
func TestListSessionsSkipsZeroMessageStub(t *testing.T) {
	root := t.TempDir()
	resolvedProject, projectDir := seedProjectDir(t, root)

	data, err := os.ReadFile(filepath.Join("..", "..", "..", "tests", "fixtures", "sample-session.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "valid.jsonl"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "stub.jsonl"), stubSessionJSONL("stub-session"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("META_CC_PROJECTS_ROOT", root)
	p := NewProvider(locator.NewSessionLocator(), resolvedProject)
	sessions, err := p.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("ListSessions must skip a zero-message stub, not fail the whole listing: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected exactly 1 session (valid returned, stub skipped), got %d: %#v", len(sessions), sessions)
	}
}

// TestListSessionsReportsEveryExcludedFile is the DIR-094 regression test for
// the half abc8135 left open: ListSessions had stopped hard-failing on a
// zero-message stub, but skipped it with a bare `continue`, so a file dropped
// from the listing was invisible to the caller — the opposite failure mode
// from the original whole-batch error, and just as much a silent loss of data.
// It also covers the case abc8135 did NOT reach: an unreadable/malformed file
// still aborted the entire listing.
//
// The checked-in tests/fixtures/malformed-corpus documents all three
// non-contributing shapes (empty, metadata-only stub, truncated JSON) and is
// seeded via the same testutil helper every other per-path test uses, so
// ListSessions, query_sessions, get_timeline and the five analysis tools are
// all held to identical inputs.
func TestListSessionsReportsEveryExcludedFile(t *testing.T) {
	root := t.TempDir()
	resolvedProject, projectDir := seedProjectDir(t, root)
	excluded := testutil.SeedMalformedCorpus(t, projectDir, resolvedProject)

	t.Setenv("META_CC_PROJECTS_ROOT", root)
	p := NewProvider(locator.NewSessionLocator(), resolvedProject)

	sessions, err := p.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("a single bad session file must not fail the whole listing: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected exactly the 1 healthy session, got %d: %#v", len(sessions), sessions)
	}
	if sessions[0].CWD != resolvedProject {
		t.Fatalf("healthy session CWD = %q, want %q", sessions[0].CWD, resolvedProject)
	}

	warnings := p.Warnings()
	if len(warnings) != len(excluded) {
		t.Fatalf("expected one warning per excluded file (%d), got %d: %v", len(excluded), len(warnings), warnings)
	}
	for _, name := range excluded {
		if !testutil.WarningsNameFile(warnings, name) {
			t.Errorf("no warning names the excluded file %q; got %v", name, warnings)
		}
	}
}

// TestSessionFromFileDistinguishesEmptyFromError guards the sentinel contract
// sessionFromFile relies on: a readable file with zero message entries returns
// errNoMessageEntries (a benign, well-understood state), while a genuine I/O
// failure returns a DIFFERENT error. DIR-094 kept this distinction because
// sessionFromFile backs targeted lookups (GetSession/findSessionFile), where
// there is no rest-of-the-batch to protect and a caller deserves the real
// error; ListSessions, which IS a corpus enumeration, now skips and reports
// BOTH classes instead of failing fast — see
// TestListSessionsReportsEveryExcludedFile. A real mid-read I/O error is not
// portably simulable in a unit test, so the "real error stays distinct" half
// is asserted here via a nonexistent file (an os.Open *PathError), which is
// the same error class any genuine read failure surfaces as.
func TestSessionFromFileDistinguishesEmptyFromError(t *testing.T) {
	root := t.TempDir()
	resolvedProject, projectDir := seedProjectDir(t, root)

	stubPath := filepath.Join(projectDir, "stub.jsonl")
	if err := os.WriteFile(stubPath, stubSessionJSONL("stub-session"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("META_CC_PROJECTS_ROOT", root)
	p := NewProvider(locator.NewSessionLocator(), resolvedProject)

	// Benign case: readable file, zero message entries → sentinel.
	if _, err := p.sessionFromFile(stubPath); !errors.Is(err, errNoMessageEntries) {
		t.Fatalf("empty stub: expected errNoMessageEntries, got %v", err)
	}

	// Real-error case: nonexistent file → NOT the sentinel, so ListSessions
	// keeps failing hard on genuine problems.
	if _, err := p.sessionFromFile(filepath.Join(projectDir, "does-not-exist.jsonl")); errors.Is(err, errNoMessageEntries) {
		t.Fatalf("missing file: expected a real I/O error, got errNoMessageEntries")
	}
}

// sessionJSONL returns the on-disk shape of a transcript carrying one
// user/assistant exchange. For a subagent transcript the entry's sessionId is
// the PARENT session's uuid and the subagent's identity lives only in the
// filename — the layout the two real subagent transcripts on this host
// (agent-a16670a1c4453c992.jsonl, agent-ad81c407898530c85.jsonl) actually have.
func sessionJSONL(sessionID, cwd string) []byte {
	lines := []string{
		`{"type":"user","sessionId":"` + sessionID + `","cwd":"` + cwd + `","timestamp":"2026-09-17T08:00:00Z","uuid":"u1","message":{"role":"user","content":"do the thing"}}`,
		`{"type":"assistant","sessionId":"` + sessionID + `","cwd":"` + cwd + `","timestamp":"2026-09-17T08:00:01Z","uuid":"a1","parentUuid":"u1","message":{"role":"assistant","model":"claude-sonnet-5","content":[{"type":"text","text":"done"}],"usage":{"input_tokens":5,"output_tokens":7}}}`,
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

// TestListSessionsIncludesSubagents is the AC for
// gap-claude-listsessions-misses-subagents: a project's listing must enumerate
// the subagent transcripts sitting beside its sessions, address each by the
// agent id a caller can hand back to another tool, and still ignore the
// non-session siblings (tool-results/) that live at the same depth.
//
// Both halves are asserted here because either alone is useless: a listing that
// omitted the subagent leaves it undiscoverable, and a listing that included it
// under its PARENT's sessionId (which is what the transcript's own entries
// carry) would print an id that no session_id-taking tool can resolve.
func TestListSessionsIncludesSubagents(t *testing.T) {
	root := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", root)

	resolvedProject, projectDir := seedProjectDir(t, root)

	const (
		parentSessionID = "7a5d362c-056b-4553-88c8-47cdaabb02be"
		agentID         = "a16670a1c4453c992"
	)

	parentFile := filepath.Join(projectDir, parentSessionID+".jsonl")
	if err := os.WriteFile(parentFile, sessionJSONL(parentSessionID, resolvedProject), 0o644); err != nil {
		t.Fatal(err)
	}

	subagentsDir := filepath.Join(projectDir, parentSessionID, locator.SubagentDirName)
	if err := os.MkdirAll(subagentsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	subagentFile := filepath.Join(subagentsDir, "agent-"+agentID+".jsonl")
	if err := os.WriteFile(subagentFile, sessionJSONL(parentSessionID, resolvedProject), 0o644); err != nil {
		t.Fatal(err)
	}

	// A same-depth sibling that is NOT a session transcript.
	toolResultsDir := filepath.Join(projectDir, parentSessionID, "tool-results")
	if err := os.MkdirAll(toolResultsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(toolResultsDir, "result.jsonl"),
		sessionJSONL("tool-result-not-a-session", resolvedProject), 0o644); err != nil {
		t.Fatal(err)
	}

	p := NewProvider(locator.NewSessionLocator(), resolvedProject)

	sessions, err := p.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 2 {
		ids := make([]string, 0, len(sessions))
		for _, s := range sessions {
			ids = append(ids, s.ID)
		}
		t.Fatalf("expected the parent session + the subagent transcript, got %d: %v", len(sessions), ids)
	}

	byID := make(map[string]conversation.Session, len(sessions))
	for _, s := range sessions {
		byID[s.ID] = s
	}
	if _, ok := byID[parentSessionID]; !ok {
		t.Errorf("parent session %s missing from listing: %v", parentSessionID, byID)
	}
	subagent, ok := byID[agentID]
	if !ok {
		t.Fatalf("subagent %s missing from listing (ids present: %v)", agentID, byID)
	}
	if subagent.ID == parentSessionID {
		t.Errorf("subagent listed under its parent's sessionId %s — the id it prints must be resolvable", parentSessionID)
	}

	// The id the listing just printed must be usable: GetSession and LoadTurns
	// are the exact-lookup entry points every session_id-taking tool shares.
	got, err := p.GetSession(context.Background(), agentID)
	if err != nil {
		t.Fatalf("GetSession(%s) after listing it: %v", agentID, err)
	}
	if got.ID != agentID {
		t.Errorf("GetSession(%s).ID = %s", agentID, got.ID)
	}
	if path, err := FilePath(got); err != nil || path != subagentFile {
		t.Errorf("GetSession(%s) resolved to %q (err %v), want %s", agentID, path, err, subagentFile)
	}
	turns, err := p.LoadTurns(context.Background(), agentID)
	if err != nil {
		t.Fatalf("LoadTurns(%s) after listing it: %v", agentID, err)
	}
	if len(turns) == 0 {
		t.Errorf("LoadTurns(%s) returned no turns", agentID)
	}

	// The parent must still resolve to its own transcript, not the subagent's.
	parent, err := p.GetSession(context.Background(), parentSessionID)
	if err != nil {
		t.Fatalf("GetSession(%s): %v", parentSessionID, err)
	}
	if path, _ := FilePath(parent); path != parentFile {
		t.Errorf("GetSession(%s) resolved to %q, want %s", parentSessionID, path, parentFile)
	}
}

// TestSessionFromEntriesSetsIsSubagent covers the Claude-side producer of
// Session.IsSubagent and Session.ParentThreadID — the half that was never
// written. The field itself and its emission both already existed
// (internal/conversation/types.go, internal/mcp/executor/query_sessions_handler.go),
// but nothing on this path ever set it, so `is_subagent` was unreachable on
// the Claude path no matter what the corpus contained.
//
// Claude Code stores a subagent transcript at
//
//	<projectsRoot>/<projectHash>/<parentSessionID>/subagents/agent-<agentID>.jsonl
//
// so the transcript's own path is what identifies it as a subagent AND names
// the session that spawned it — no spawn metadata has to be parsed out of the
// JSONL. The fixtures below reproduce exactly that layout.
//
// Listing such a transcript and attributing it are two different jobs (see
// TestListSessionsIncludesSubagents for the listing half): a record can be
// enumerated under the right id and still leave its consumer unable to tell it
// came from a subagent at all.
func TestSessionFromEntriesSetsIsSubagent(t *testing.T) {
	root := t.TempDir()
	resolvedProject, projectDir := seedProjectDir(t, root)

	const parentSessionID = "7a5d362c-056b-4553-88c8-47cdaabb02be"

	topLevel := filepath.Join(projectDir, parentSessionID+".jsonl")
	if err := os.WriteFile(topLevel, sessionJSONL(parentSessionID, resolvedProject), 0o644); err != nil {
		t.Fatal(err)
	}

	subagentDir := filepath.Join(projectDir, parentSessionID, locator.SubagentDirName)
	if err := os.MkdirAll(subagentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// A real subagent transcript carries its PARENT's sessionId (verified
	// against the live corpus on this host), which is why ParentThreadID is
	// derived from the path rather than from the entries.
	subagentFile := filepath.Join(subagentDir, "agent-a16670a1c4453c992.jsonl")
	if err := os.WriteFile(subagentFile, sessionJSONL(parentSessionID, resolvedProject), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("transcript under a session uuid's subagents dir is a subagent", func(t *testing.T) {
		session, err := SessionFromFile(subagentFile)
		if err != nil {
			t.Fatalf("SessionFromFile(%s): %v", subagentFile, err)
		}
		if !session.IsSubagent {
			t.Errorf("IsSubagent = false for a transcript under <uuid>/subagents/, want true")
		}
		if session.ParentThreadID != parentSessionID {
			t.Errorf("ParentThreadID = %q, want the fixture's parent session uuid %q",
				session.ParentThreadID, parentSessionID)
		}
		if session.Lineage != conversation.LineageStatusChild {
			t.Errorf("Lineage = %q, want %q (the path reliably reports the parent edge)",
				session.Lineage, conversation.LineageStatusChild)
		}
	})

	t.Run("top-level transcript is not a subagent and has no parent", func(t *testing.T) {
		session, err := SessionFromFile(topLevel)
		if err != nil {
			t.Fatalf("SessionFromFile(%s): %v", topLevel, err)
		}
		if session.IsSubagent {
			t.Errorf("IsSubagent = true for a top-level transcript, want false")
		}
		if session.ParentThreadID != "" {
			t.Errorf("ParentThreadID = %q for a top-level transcript, want empty", session.ParentThreadID)
		}
	})

	t.Run("a subagents dir with a non-uuid parent segment names no parent", func(t *testing.T) {
		// IsSubagent deliberately follows the SAME rule the rest of the
		// codebase uses to recognise a subagent transcript
		// (locator.IsSubagentTranscript: filed directly inside a subagents/
		// directory), so a record can never disagree with the id sessionIDFor
		// gives it. ParentThreadID is the stricter half: it is claimed only
		// when the path actually names a session uuid, so a path that
		// witnesses no parent leaves the link empty rather than fabricating
		// one.
		strayDir := filepath.Join(projectDir, "subagents")
		if err := os.MkdirAll(strayDir, 0o755); err != nil {
			t.Fatal(err)
		}
		strayFile := filepath.Join(strayDir, "agent-a16670a1c4453c992.jsonl")
		if err := os.WriteFile(strayFile, sessionJSONL(parentSessionID, resolvedProject), 0o644); err != nil {
			t.Fatal(err)
		}

		session, err := SessionFromFile(strayFile)
		if err != nil {
			t.Fatalf("SessionFromFile(%s): %v", strayFile, err)
		}
		if !session.IsSubagent {
			t.Errorf("IsSubagent = false for a transcript filed directly inside subagents/, want true — the same rule sessionIDFor addresses it by")
		}
		if session.ParentThreadID != "" {
			t.Errorf("ParentThreadID = %q, want empty: the parent segment is not a session uuid", session.ParentThreadID)
		}
	})
}
