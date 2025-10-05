package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yale/meta-cc/internal/testutil"
)

func TestSessionPipeline_LoadProjectLevel(t *testing.T) {
	// Test that SessionPipeline loads all sessions when ProjectPath is set
	homeDir, _ := os.UserHomeDir()
	projectPath := "/home/yale/work/testproject"
	projectHash := "-home-yale-work-testproject"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	os.MkdirAll(sessionDir, 0755)
	defer os.RemoveAll(sessionDir)

	// Create 3 session files with different content
	session1 := filepath.Join(sessionDir, "session1.jsonl")
	session2 := filepath.Join(sessionDir, "session2.jsonl")
	session3 := filepath.Join(sessionDir, "session3.jsonl")

	// Session 1: 2 user messages
	session1Content := `{"type":"user","uuid":"uuid1","timestamp":"2024-01-01T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"message from session 1-1"}]}}
{"type":"user","uuid":"uuid2","timestamp":"2024-01-01T10:01:00Z","message":{"role":"user","content":[{"type":"text","text":"message from session 1-2"}]}}
`
	// Session 2: 1 user message
	session2Content := `{"type":"user","uuid":"uuid3","timestamp":"2024-01-02T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"message from session 2-1"}]}}
`
	// Session 3: 2 user messages
	session3Content := `{"type":"user","uuid":"uuid4","timestamp":"2024-01-03T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"message from session 3-1"}]}}
{"type":"user","uuid":"uuid5","timestamp":"2024-01-03T10:01:00Z","message":{"role":"user","content":[{"type":"text","text":"message from session 3-2"}]}}
`

	os.WriteFile(session1, []byte(session1Content), 0644)
	os.WriteFile(session2, []byte(session2Content), 0644)
	os.WriteFile(session3, []byte(session3Content), 0644)

	// Make session3 the newest
	os.Chtimes(session1, testutil.TimeFromUnix(1000), testutil.TimeFromUnix(1000))
	os.Chtimes(session2, testutil.TimeFromUnix(2000), testutil.TimeFromUnix(2000))
	os.Chtimes(session3, testutil.TimeFromUnix(3000), testutil.TimeFromUnix(3000))

	// Test with ProjectPath set (should load ALL sessions)
	opts := GlobalOptions{
		ProjectPath: projectPath,
		SessionOnly: false,
	}

	pipeline := NewSessionPipeline(opts)
	err := pipeline.Load(LoadOptions{AutoDetect: true})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	entries := pipeline.GetEntries()

	// Should have entries from ALL 3 sessions (5 total user messages)
	if len(entries) != 5 {
		t.Errorf("Expected 5 entries (from all 3 sessions), got %d", len(entries))
	}

	// Verify we have entries from all sessions by checking UUIDs
	uuidMap := make(map[string]bool)
	for _, entry := range entries {
		uuidMap[entry.UUID] = true
	}

	expectedUUIDs := []string{"uuid1", "uuid2", "uuid3", "uuid4", "uuid5"}
	for _, uuid := range expectedUUIDs {
		if !uuidMap[uuid] {
			t.Errorf("Expected to find entry with UUID %s", uuid)
		}
	}
}

func TestSessionPipeline_LoadSessionOnly(t *testing.T) {
	// Test that SessionPipeline loads only latest session when SessionOnly is true
	homeDir, _ := os.UserHomeDir()
	projectPath := "/home/yale/work/testproject2"
	projectHash := "-home-yale-work-testproject2"

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	os.MkdirAll(sessionDir, 0755)
	defer os.RemoveAll(sessionDir)

	// Create 2 session files
	session1 := filepath.Join(sessionDir, "session1.jsonl")
	session2 := filepath.Join(sessionDir, "session2.jsonl")

	session1Content := `{"type":"user","uuid":"uuid1","timestamp":"2024-01-01T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"old session"}]}}
`
	session2Content := `{"type":"user","uuid":"uuid2","timestamp":"2024-01-02T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"new session"}]}}
`

	os.WriteFile(session1, []byte(session1Content), 0644)
	os.WriteFile(session2, []byte(session2Content), 0644)
	os.Chtimes(session1, testutil.TimeFromUnix(1000), testutil.TimeFromUnix(1000))
	os.Chtimes(session2, testutil.TimeFromUnix(2000), testutil.TimeFromUnix(2000))

	// Test with SessionOnly=true (should load ONLY latest session)
	opts := GlobalOptions{
		ProjectPath: projectPath,
		SessionOnly: true, // Explicitly request session-only mode
	}

	pipeline := NewSessionPipeline(opts)
	err := pipeline.Load(LoadOptions{AutoDetect: true})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	entries := pipeline.GetEntries()

	// Should have entries from ONLY the latest session (1 entry)
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry (latest session only), got %d", len(entries))
	}

	// Verify it's from the newest session
	if entries[0].UUID != "uuid2" {
		t.Errorf("Expected UUID from newest session (uuid2), got %s", entries[0].UUID)
	}
}
