package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/testutil"
)

// setupTestProject creates a temporary project directory and returns the path and hash
// This ensures cross-platform compatibility (Windows, macOS, Linux)
func setupTestProject(t *testing.T, prefix string) (projectPath string, projectHash string, cleanup func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	projectPath = tempDir
	projectHash = filepath.ToSlash(projectPath)
	projectHash = strings.ReplaceAll(projectHash, "/", "-")

	cleanup = func() {
		os.RemoveAll(tempDir)
	}

	return projectPath, projectHash, cleanup
}

func TestSessionPipeline_LoadProjectLevel(t *testing.T) {
	// Test that SessionPipeline loads all sessions when ProjectPath is set
	homeDir, _ := os.UserHomeDir()
	projectPath, projectHash, cleanupProject := setupTestProject(t, "testproject")
	defer cleanupProject()

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
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

	if err := os.WriteFile(session1, []byte(session1Content), 0644); err != nil {
		t.Fatalf("failed to write session1: %v", err)
	}
	if err := os.WriteFile(session2, []byte(session2Content), 0644); err != nil {
		t.Fatalf("failed to write session2: %v", err)
	}
	if err := os.WriteFile(session3, []byte(session3Content), 0644); err != nil {
		t.Fatalf("failed to write session3: %v", err)
	}

	// Make session3 the newest
	if err := os.Chtimes(session1, testutil.TimeFromUnix(1000), testutil.TimeFromUnix(1000)); err != nil {
		t.Fatalf("failed to set session1 times: %v", err)
	}
	if err := os.Chtimes(session2, testutil.TimeFromUnix(2000), testutil.TimeFromUnix(2000)); err != nil {
		t.Fatalf("failed to set session2 times: %v", err)
	}
	if err := os.Chtimes(session3, testutil.TimeFromUnix(3000), testutil.TimeFromUnix(3000)); err != nil {
		t.Fatalf("failed to set session3 times: %v", err)
	}

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
	projectPath, projectHash, cleanupProject := setupTestProject(t, "testproject2")
	defer cleanupProject()

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	// Create 2 session files
	session1 := filepath.Join(sessionDir, "session1.jsonl")
	session2 := filepath.Join(sessionDir, "session2.jsonl")

	session1Content := `{"type":"user","uuid":"uuid1","timestamp":"2024-01-01T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"old session"}]}}
`
	session2Content := `{"type":"user","uuid":"uuid2","timestamp":"2024-01-02T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"new session"}]}}
`

	if err := os.WriteFile(session1, []byte(session1Content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.WriteFile(session2, []byte(session2Content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.Chtimes(session1, testutil.TimeFromUnix(1000), testutil.TimeFromUnix(1000)); err != nil {
		t.Fatalf("failed to set times: %v", err)
	}
	if err := os.Chtimes(session2, testutil.TimeFromUnix(2000), testutil.TimeFromUnix(2000)); err != nil {
		t.Fatalf("failed to set times: %v", err)
	}

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

// TestNewSessionPipeline verifies pipeline constructor
func TestNewSessionPipeline(t *testing.T) {
	tests := []struct {
		name string
		opts GlobalOptions
	}{
		{
			name: "default options",
			opts: GlobalOptions{},
		},
		{
			name: "with session ID",
			opts: GlobalOptions{
				SessionID: "test-session-123",
			},
		},
		{
			name: "with project path",
			opts: GlobalOptions{
				ProjectPath: "/test/project",
			},
		},
		{
			name: "with session only flag",
			opts: GlobalOptions{
				ProjectPath: "/test/project",
				SessionOnly: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := NewSessionPipeline(tt.opts)
			if pipeline == nil {
				t.Fatal("NewSessionPipeline returned nil")
			}
			if pipeline.opts != tt.opts {
				t.Errorf("opts not set correctly, got %+v, want %+v", pipeline.opts, tt.opts)
			}
			if pipeline.turnIndex == nil {
				t.Error("turnIndex map not initialized")
			}
			if len(pipeline.entries) != 0 {
				t.Errorf("entries should be empty initially, got %d", len(pipeline.entries))
			}
		})
	}
}

// TestSessionPipeline_SessionPath verifies SessionPath accessor
func TestSessionPipeline_SessionPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	projectPath, projectHash, cleanupProject := setupTestProject(t, "testproject-sessionpath")
	defer cleanupProject()

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	sessionFile := filepath.Join(sessionDir, "session1.jsonl")
	sessionContent := `{"type":"user","uuid":"uuid1","timestamp":"2024-01-01T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"test"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.Chtimes(sessionFile, testutil.TimeFromUnix(1000), testutil.TimeFromUnix(1000)); err != nil {
		t.Fatalf("failed to set times: %v", err)
	}

	opts := GlobalOptions{
		ProjectPath: projectPath,
		SessionOnly: true,
	}

	pipeline := NewSessionPipeline(opts)
	err := pipeline.Load(LoadOptions{AutoDetect: true})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	sessionPath := pipeline.SessionPath()
	if sessionPath == "" {
		t.Error("SessionPath returned empty string")
	}
	if !strings.Contains(sessionPath, "session1.jsonl") {
		t.Errorf("SessionPath should contain session filename, got: %s", sessionPath)
	}
}

// TestSessionPipeline_EntryCount verifies EntryCount accessor
func TestSessionPipeline_EntryCount(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	projectPath, projectHash, cleanupProject := setupTestProject(t, "testproject-entrycount")
	defer cleanupProject()

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	sessionFile := filepath.Join(sessionDir, "session1.jsonl")
	// Create 3 entries
	sessionContent := `{"type":"user","uuid":"uuid1","timestamp":"2024-01-01T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"msg1"}]}}
{"type":"user","uuid":"uuid2","timestamp":"2024-01-01T10:01:00Z","message":{"role":"user","content":[{"type":"text","text":"msg2"}]}}
{"type":"user","uuid":"uuid3","timestamp":"2024-01-01T10:02:00Z","message":{"role":"user","content":[{"type":"text","text":"msg3"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.Chtimes(sessionFile, testutil.TimeFromUnix(1000), testutil.TimeFromUnix(1000)); err != nil {
		t.Fatalf("failed to set times: %v", err)
	}

	pipeline := NewSessionPipeline(GlobalOptions{
		ProjectPath: projectPath,
		SessionOnly: true,
	})

	// Before loading
	if pipeline.EntryCount() != 0 {
		t.Errorf("EntryCount before Load should be 0, got %d", pipeline.EntryCount())
	}

	err := pipeline.Load(LoadOptions{AutoDetect: true})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// After loading
	if pipeline.EntryCount() != 3 {
		t.Errorf("EntryCount should be 3, got %d", pipeline.EntryCount())
	}
}

// TestSessionPipeline_GetEntries verifies GetEntries accessor
func TestSessionPipeline_GetEntries(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	projectPath, projectHash, cleanupProject := setupTestProject(t, "testproject-getentries")
	defer cleanupProject()

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	sessionFile := filepath.Join(sessionDir, "session1.jsonl")
	sessionContent := `{"type":"user","uuid":"uuid1","timestamp":"2024-01-01T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"test message"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.Chtimes(sessionFile, testutil.TimeFromUnix(1000), testutil.TimeFromUnix(1000)); err != nil {
		t.Fatalf("failed to set times: %v", err)
	}

	pipeline := NewSessionPipeline(GlobalOptions{
		ProjectPath: projectPath,
		SessionOnly: true,
	})
	err := pipeline.Load(LoadOptions{AutoDetect: true})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	entries := pipeline.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].UUID != "uuid1" {
		t.Errorf("Expected UUID uuid1, got %s", entries[0].UUID)
	}
	if entries[0].Type != "user" {
		t.Errorf("Expected type user, got %s", entries[0].Type)
	}
}

// TestSessionPipeline_BuildTurnIndex verifies BuildTurnIndex method
func TestSessionPipeline_BuildTurnIndex(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	projectPath, projectHash, cleanupProject := setupTestProject(t, "testproject-turnindex")
	defer cleanupProject()

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	sessionFile := filepath.Join(sessionDir, "session1.jsonl")
	sessionContent := `{"type":"user","uuid":"uuid1","timestamp":"2024-01-01T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"msg1"}]}}
{"type":"user","uuid":"uuid2","timestamp":"2024-01-01T10:01:00Z","message":{"role":"user","content":[{"type":"text","text":"msg2"}]}}
{"type":"user","uuid":"uuid3","timestamp":"2024-01-01T10:02:00Z","message":{"role":"user","content":[{"type":"text","text":"msg3"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.Chtimes(sessionFile, testutil.TimeFromUnix(1000), testutil.TimeFromUnix(1000)); err != nil {
		t.Fatalf("failed to set times: %v", err)
	}

	pipeline := NewSessionPipeline(GlobalOptions{
		ProjectPath: projectPath,
		SessionOnly: true,
	})
	err := pipeline.Load(LoadOptions{AutoDetect: true})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// First call should build the index
	index := pipeline.BuildTurnIndex()
	if len(index) != 3 {
		t.Errorf("Expected index with 3 entries, got %d", len(index))
	}

	// Verify UUID -> turn_sequence mapping
	if index["uuid1"] != 0 {
		t.Errorf("Expected uuid1 at index 0, got %d", index["uuid1"])
	}
	if index["uuid2"] != 1 {
		t.Errorf("Expected uuid2 at index 1, got %d", index["uuid2"])
	}
	if index["uuid3"] != 2 {
		t.Errorf("Expected uuid3 at index 2, got %d", index["uuid3"])
	}

	// Second call should return cached index
	index2 := pipeline.BuildTurnIndex()
	if len(index2) != len(index) {
		t.Error("BuildTurnIndex should return cached result on second call")
	}
}

// TestSessionPipeline_ExtractToolCalls verifies ExtractToolCalls method
func TestSessionPipeline_ExtractToolCalls(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	projectPath, projectHash, cleanupProject := setupTestProject(t, "testproject-toolcalls")
	defer cleanupProject()

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	sessionFile := filepath.Join(sessionDir, "session1.jsonl")
	// Create session with tool_use and tool_result
	sessionContent := `{"type":"assistant","uuid":"uuid1","timestamp":"2024-01-01T10:00:00Z","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool1","name":"Bash","input":{"command":"ls"}}]}}
{"type":"user","uuid":"uuid2","timestamp":"2024-01-01T10:00:01Z","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool1","content":"file1.txt\nfile2.txt","status":"success"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.Chtimes(sessionFile, testutil.TimeFromUnix(1000), testutil.TimeFromUnix(1000)); err != nil {
		t.Fatalf("failed to set times: %v", err)
	}

	pipeline := NewSessionPipeline(GlobalOptions{
		ProjectPath: projectPath,
		SessionOnly: true,
	})
	err := pipeline.Load(LoadOptions{AutoDetect: true})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	toolCalls := pipeline.ExtractToolCalls()
	if len(toolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(toolCalls))
	}

	tc := toolCalls[0]
	if tc.ToolName != "Bash" {
		t.Errorf("Expected tool name Bash, got %s", tc.ToolName)
	}
	if tc.Status != "success" {
		t.Errorf("Expected status success, got %s", tc.Status)
	}
	if !strings.Contains(tc.Output, "file1.txt") {
		t.Errorf("Expected output to contain file1.txt, got %s", tc.Output)
	}
}

// TestSessionPipeline_LoadError verifies error handling in Load
func TestSessionPipeline_LoadError(t *testing.T) {
	// Test with non-existent session ID
	pipeline := NewSessionPipeline(GlobalOptions{
		SessionID: "non-existent-session-id-12345",
	})

	err := pipeline.Load(LoadOptions{AutoDetect: false})
	if err == nil {
		t.Error("Expected error for non-existent session, got nil")
	}
}

// TestSessionPipeline_LoadWithSessionID verifies loading specific session by ID
func TestSessionPipeline_LoadWithSessionID(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	sessionID := "test-session-specific-id"

	// Create session file in a project directory (FromSessionID searches ~/.claude/projects/*/)
	projectHash := "test-project-hash"
	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	sessionContent := `{"type":"user","uuid":"uuid1","timestamp":"2024-01-01T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"specific session"}]}}
`
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	pipeline := NewSessionPipeline(GlobalOptions{
		SessionID: sessionID,
	})
	err := pipeline.Load(LoadOptions{AutoDetect: false})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	entries := pipeline.GetEntries()
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].UUID != "uuid1" {
		t.Errorf("Expected UUID uuid1, got %s", entries[0].UUID)
	}
}

// TestSessionPipeline_ProjectLevelSessionPath verifies session path in project mode
func TestSessionPipeline_ProjectLevelSessionPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	projectPath, projectHash, cleanupProject := setupTestProject(t, "testproject-projpath")
	defer cleanupProject()

	sessionDir := filepath.Join(homeDir, ".claude", "projects", projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	session1 := filepath.Join(sessionDir, "session1.jsonl")
	session2 := filepath.Join(sessionDir, "session2.jsonl")

	session1Content := `{"type":"user","uuid":"uuid1","timestamp":"2024-01-01T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"s1"}]}}
`
	session2Content := `{"type":"user","uuid":"uuid2","timestamp":"2024-01-02T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"s2"}]}}
`

	if err := os.WriteFile(session1, []byte(session1Content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.WriteFile(session2, []byte(session2Content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.Chtimes(session1, testutil.TimeFromUnix(1000), testutil.TimeFromUnix(1000)); err != nil {
		t.Fatalf("failed to set times: %v", err)
	}
	if err := os.Chtimes(session2, testutil.TimeFromUnix(2000), testutil.TimeFromUnix(2000)); err != nil {
		t.Fatalf("failed to set times: %v", err)
	}

	pipeline := NewSessionPipeline(GlobalOptions{
		ProjectPath: projectPath,
		SessionOnly: false, // Project-level mode
	})
	err := pipeline.Load(LoadOptions{AutoDetect: true})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	sessionPath := pipeline.SessionPath()
	// In project mode, session path should indicate project context
	if !strings.Contains(sessionPath, "project:") {
		t.Errorf("SessionPath in project mode should contain 'project:', got: %s", sessionPath)
	}
	if !strings.Contains(sessionPath, "sessions") {
		t.Errorf("SessionPath in project mode should indicate session count, got: %s", sessionPath)
	}
}
