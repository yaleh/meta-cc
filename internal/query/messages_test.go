package query

import (
	"os"
	"path/filepath"
	"testing"

	pipelinepkg "github.com/yaleh/meta-cc/pkg/pipeline"
)

func createMessagesTestSession(t *testing.T, sessionID, projectHash string, content string) {
	t.Helper()

	projectsRoot := os.Getenv("META_CC_PROJECTS_ROOT")
	if projectsRoot == "" {
		t.Fatal("META_CC_PROJECTS_ROOT must be set for tests")
	}

	sessionDir := filepath.Join(projectsRoot, projectHash)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(sessionDir) })

	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write session fixture: %v", err)
	}
}

func TestRunUserMessagesQuery_PatternAndContext(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	sessionID := "test-messages-library"
	projectHash := "-home-yale-work-test-messages-library"
	fixture := `{"type":"user","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":"Fix bug in parser"}}
{"type":"assistant","timestamp":"2025-10-02T10:00:10.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"assistant","content":[{"type":"text","text":"Sure"}]}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"user","content":"Add new feature"}}
`

	createMessagesTestSession(t, sessionID, projectHash, fixture)

	opts := UserMessagesQueryOptions{
		Pipeline: pipelinepkg.GlobalOptions{SessionID: sessionID},
		Pattern:  "Fix",
		Context:  1,
	}

	msgs, err := RunUserMessagesQuery(opts)
	if err != nil {
		t.Fatalf("RunUserMessagesQuery failed: %v", err)
	}

	if len(msgs) != 1 {
		t.Fatalf("expected 1 user message, got %d", len(msgs))
	}

	if msgs[0].ContextAfter == nil || len(msgs[0].ContextAfter) == 0 {
		t.Fatalf("expected context after message")
	}

	if msgs[0].ContextAfter[0].Role != "assistant" {
		t.Fatalf("expected assistant context after, got %s", msgs[0].ContextAfter[0].Role)
	}
}

func TestRunUserMessagesQuery_LimitOffset(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	sessionID := "test-messages-library-pagination"
	projectHash := "-home-yale-work-test-messages-library-pagination"
	fixture := `{"type":"user","timestamp":"2025-10-02T10:00:00.000Z","uuid":"uuid-1","sessionId":"test","message":{"role":"user","content":"Message 1"}}
{"type":"user","timestamp":"2025-10-02T10:01:00.000Z","uuid":"uuid-2","sessionId":"test","message":{"role":"user","content":"Message 2"}}
{"type":"user","timestamp":"2025-10-02T10:02:00.000Z","uuid":"uuid-3","sessionId":"test","message":{"role":"user","content":"Message 3"}}
`

	createMessagesTestSession(t, sessionID, projectHash, fixture)

	opts := UserMessagesQueryOptions{
		Pipeline: pipelinepkg.GlobalOptions{SessionID: sessionID},
		Offset:   1,
		Limit:    1,
	}

	msgs, err := RunUserMessagesQuery(opts)
	if err != nil {
		t.Fatalf("RunUserMessagesQuery failed: %v", err)
	}

	if len(msgs) != 1 {
		t.Fatalf("expected 1 message after pagination, got %d", len(msgs))
	}

	if msgs[0].Content != "Message 2" {
		t.Fatalf("expected Message 2, got %s", msgs[0].Content)
	}
}
