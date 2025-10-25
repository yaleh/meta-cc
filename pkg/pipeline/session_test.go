package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test helper: Creates a minimal valid session file in Claude directory structure
func createTestSession(t *testing.T, sessionID, projectPath string) string {
	t.Helper()

	projectsRoot := os.Getenv("META_CC_PROJECTS_ROOT")
	if projectsRoot == "" {
		var err error
		projectsRoot, err = os.MkdirTemp("", "meta-cc-projects-*")
		if err != nil {
			t.Fatalf("failed to create temp projects dir: %v", err)
		}
		// Ensure cleanup when helper creates the directory
		t.Cleanup(func() { os.RemoveAll(projectsRoot) })
	}
	// Generate project hash using same logic as pathToHash in locator
	// First resolve symlinks for consistent hashing on macOS (/var -> /private/var)
	resolvedPath, err := filepath.EvalSymlinks(projectPath)
	if err != nil {
		// If resolution fails (e.g., path doesn't exist), use original path
		resolvedPath = projectPath
	}
	// Replace both / and \ with -, then replace : (Windows drive letters)
	projectHash := strings.ReplaceAll(resolvedPath, "\\", "-")
	projectHash = strings.ReplaceAll(projectHash, "/", "-")
	projectHash = strings.ReplaceAll(projectHash, ":", "-")
	sessionDir := filepath.Join(projectsRoot, projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	// Convert path to forward slashes for JSON (Windows paths have backslashes which break JSON parsing)
	jsonSafePath := filepath.ToSlash(projectPath)
	content := `{"type":"user","timestamp":"2025-10-02T06:07:13.673Z","message":{"role":"user","content":"test"},"uuid":"uuid-1","parentUuid":null,"sessionId":"` + sessionID + `","cwd":"` + jsonSafePath + `"}
{"type":"assistant","timestamp":"2025-10-02T06:08:57.769Z","message":{"id":"msg_01","type":"message","role":"assistant","model":"claude-sonnet-4","content":[{"type":"tool_use","id":"toolu_01","name":"Bash","input":{"command":"ls"}}],"stop_reason":"tool_use"},"uuid":"uuid-2","parentUuid":"uuid-1","sessionId":"` + sessionID + `","cwd":"` + jsonSafePath + `"}
{"type":"user","timestamp":"2025-10-02T06:09:10.123Z","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_01","content":"file1.txt\nfile2.txt"}]},"uuid":"uuid-3","parentUuid":"uuid-2","sessionId":"` + sessionID + `"}
`
	if err := os.WriteFile(sessionFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(sessionDir) })

	return sessionFile
}

func TestNewSessionPipeline(t *testing.T) {
	opts := GlobalOptions{
		SessionID:   "test-session",
		ProjectPath: "/test/path",
		SessionOnly: false,
	}

	p := NewSessionPipeline(opts)

	if p == nil {
		t.Fatal("NewSessionPipeline returned nil")
	}

	if p.opts.SessionID != "test-session" {
		t.Errorf("Expected SessionID 'test-session', got %q", p.opts.SessionID)
	}

	if p.turnIndex == nil {
		t.Error("turnIndex map should be initialized")
	}
}

func TestSessionPipeline_Load_ExplicitSessionID(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	sessionID := "test-explicit-session"
	projectPath := "/test/project"
	sessionFile := createTestSession(t, sessionID, projectPath)

	p := NewSessionPipeline(GlobalOptions{
		SessionID: sessionID,
	})

	err := p.Load(LoadOptions{AutoDetect: false})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if p.session != sessionFile {
		t.Errorf("Expected session path %q, got %q", sessionFile, p.session)
	}

	if p.EntryCount() == 0 {
		t.Error("Expected non-zero entry count")
	}
}

func TestSessionPipeline_Load_ProjectPath(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	sessionID := "test-project-session"
	// Use temp dir for cross-platform compatibility
	tempDir, err := os.MkdirTemp("", "testproject2")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	projectPath := tempDir
	createTestSession(t, sessionID, projectPath)

	p := NewSessionPipeline(GlobalOptions{
		ProjectPath: projectPath,
	})

	err = p.Load(LoadOptions{AutoDetect: false})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if p.EntryCount() == 0 {
		t.Error("Expected non-zero entry count")
	}
}

func TestSessionPipeline_Load_AutoDetect(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	// Create a session in a test directory and set it as ProjectPath
	// (AutoDetect relies on the locator's default behavior which uses project path)
	// Use temp dir for cross-platform compatibility
	tempDir, err := os.MkdirTemp("", "testautodetect")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testProjectPath := tempDir
	sessionID := "test-auto-session"
	createTestSession(t, sessionID, testProjectPath)

	p := NewSessionPipeline(GlobalOptions{
		ProjectPath: testProjectPath,
	})

	err = p.Load(LoadOptions{AutoDetect: true})
	if err != nil {
		t.Fatalf("Load with AutoDetect failed: %v", err)
	}

	if p.EntryCount() == 0 {
		t.Error("Expected non-zero entry count")
	}
}

func TestSessionPipeline_Load_NoSessionSpecified(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	p := NewSessionPipeline(GlobalOptions{})

	err := p.Load(LoadOptions{AutoDetect: false})
	if err == nil {
		t.Error("Expected error when no session specified and AutoDetect is false")
	}
}

func TestSessionPipeline_Load_SessionNotFound(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	p := NewSessionPipeline(GlobalOptions{
		SessionID: "nonexistent-session-id",
	})

	err := p.Load(LoadOptions{AutoDetect: false})
	if err == nil {
		t.Error("Expected error for nonexistent session")
	}
}

func TestSessionPipeline_Load_InvalidJSONL(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	// Create a session file with invalid JSONL
	sessionID := "invalid-jsonl-session"
	projectHash := "-test-invalid"
	projectsRoot := os.Getenv("META_CC_PROJECTS_ROOT")
	sessionDir := filepath.Join(projectsRoot, projectHash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}
	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte("invalid json\n{broken"), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}
	defer os.RemoveAll(sessionDir)

	p := NewSessionPipeline(GlobalOptions{
		SessionID: sessionID,
	})

	err := p.Load(LoadOptions{AutoDetect: false})
	if err == nil {
		t.Error("Expected error for invalid JSONL")
	}
}

func TestSessionPipeline_ExtractToolCalls(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	sessionID := "test-tools-session"
	projectPath := "/test/tools"
	createTestSession(t, sessionID, projectPath)

	p := NewSessionPipeline(GlobalOptions{
		SessionID: sessionID,
	})

	err := p.Load(LoadOptions{AutoDetect: false})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	tools := p.ExtractToolCalls()

	if len(tools) == 0 {
		t.Error("Expected at least one tool call")
	}

	// Verify tool call structure
	if len(tools) > 0 {
		if tools[0].ToolName != "Bash" {
			t.Errorf("Expected tool name 'Bash', got %q", tools[0].ToolName)
		}
	}
}

func TestSessionPipeline_ExtractToolCalls_BeforeLoad(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	p := NewSessionPipeline(GlobalOptions{})

	tools := p.ExtractToolCalls()

	if len(tools) != 0 {
		t.Error("Expected zero tools before Load")
	}
}

func TestSessionPipeline_BuildTurnIndex(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	sessionID := "test-index-session"
	projectPath := "/test/index"
	createTestSession(t, sessionID, projectPath)

	p := NewSessionPipeline(GlobalOptions{
		SessionID: sessionID,
	})

	err := p.Load(LoadOptions{AutoDetect: false})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	index := p.BuildTurnIndex()

	if len(index) == 0 {
		t.Error("Expected non-zero turn index")
	}

	// Verify index mapping: UUID -> turn sequence
	if _, exists := index["uuid-1"]; !exists {
		t.Error("Expected UUID 'uuid-1' in turn index")
	}

	if _, exists := index["uuid-2"]; !exists {
		t.Error("Expected UUID 'uuid-2' in turn index")
	}

	// Verify turn sequence is correct
	if index["uuid-1"] != 0 {
		t.Errorf("Expected turn index 0 for uuid-1, got %d", index["uuid-1"])
	}

	if index["uuid-2"] != 1 {
		t.Errorf("Expected turn index 1 for uuid-2, got %d", index["uuid-2"])
	}
}

func TestSessionPipeline_BuildTurnIndex_Idempotent(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	sessionID := "test-idempotent-session"
	projectPath := "/test/idempotent"
	createTestSession(t, sessionID, projectPath)

	p := NewSessionPipeline(GlobalOptions{
		SessionID: sessionID,
	})

	if err := p.Load(LoadOptions{AutoDetect: false}); err != nil {
		t.Fatalf("failed to load session: %v", err)
	}

	index1 := p.BuildTurnIndex()
	index2 := p.BuildTurnIndex()

	if len(index1) != len(index2) {
		t.Error("BuildTurnIndex not idempotent: different lengths")
	}

	// Verify same keys
	for uuid := range index1 {
		if index1[uuid] != index2[uuid] {
			t.Errorf("BuildTurnIndex not idempotent: UUID %q has different values", uuid)
		}
	}
}

func TestSessionPipeline_SessionPath(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	sessionID := "test-path-session"
	projectPath := "/test/path"
	sessionFile := createTestSession(t, sessionID, projectPath)

	p := NewSessionPipeline(GlobalOptions{
		SessionID: sessionID,
	})

	if p.SessionPath() != "" {
		t.Error("SessionPath should be empty before Load")
	}

	if err := p.Load(LoadOptions{AutoDetect: false}); err != nil {
		t.Fatalf("failed to load session: %v", err)
	}

	if p.SessionPath() != sessionFile {
		t.Errorf("Expected SessionPath %q, got %q", sessionFile, p.SessionPath())
	}
}

func TestSessionPipeline_EntryCount(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	sessionID := "test-count-session"
	projectPath := "/test/count"
	createTestSession(t, sessionID, projectPath)

	p := NewSessionPipeline(GlobalOptions{
		SessionID: sessionID,
	})

	if p.EntryCount() != 0 {
		t.Error("EntryCount should be 0 before Load")
	}

	if err := p.Load(LoadOptions{AutoDetect: false}); err != nil {
		t.Fatalf("failed to load session: %v", err)
	}

	count := p.EntryCount()
	if count != 3 {
		t.Errorf("Expected entry count 3, got %d", count)
	}
}

func TestSessionPipeline_Load_WithSessionOnly(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	// This tests the SessionOnly flag behavior
	sessionID := "test-session-only"
	projectPath := "/test/session-only"
	createTestSession(t, sessionID, projectPath)

	p := NewSessionPipeline(GlobalOptions{
		SessionID:   sessionID,
		SessionOnly: true, // Should work same as SessionOnly: false for explicit SessionID
	})

	err := p.Load(LoadOptions{AutoDetect: false})
	if err != nil {
		t.Fatalf("Load with SessionOnly failed: %v", err)
	}

	if p.EntryCount() == 0 {
		t.Error("Expected non-zero entry count")
	}
}

func TestSessionPipeline_Load_WithValidation(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	sessionID := "test-validation-session"
	projectPath := "/test/validation"
	createTestSession(t, sessionID, projectPath)

	p := NewSessionPipeline(GlobalOptions{
		SessionID: sessionID,
	})

	err := p.Load(LoadOptions{
		AutoDetect: false,
		Validate:   true,
	})
	if err != nil {
		t.Fatalf("Load with validation failed: %v", err)
	}

	if p.EntryCount() == 0 {
		t.Error("Expected non-zero entry count")
	}
}

func TestSessionPipeline_Entries(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	sessionID := "test-entries-session"
	projectPath := "/test/entries"
	createTestSession(t, sessionID, projectPath)

	p := NewSessionPipeline(GlobalOptions{
		SessionID: sessionID,
	})

	// Before Load, should return empty slice
	entries := p.Entries()
	if len(entries) != 0 {
		t.Error("Expected zero entries before Load")
	}

	if err := p.Load(LoadOptions{AutoDetect: false}); err != nil {
		t.Fatalf("failed to load session: %v", err)
	}

	// After Load, should return actual entries
	entries = p.Entries()
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Verify entry structure
	if len(entries) > 0 {
		if entries[0].UUID != "uuid-1" {
			t.Errorf("Expected first entry UUID 'uuid-1', got %q", entries[0].UUID)
		}
	}
}
