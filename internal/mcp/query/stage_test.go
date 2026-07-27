package query

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

const (
	stageTestUser1 = `{"type":"user","timestamp":"2025-01-15T10:00:00Z","message":{"content":"fix bug"}}`
)

func TestHandleExecuteStage2Query_TransformAllNull_WarningsInResponse(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_stage_warn.jsonl")
	if err := os.WriteFile(testFile, []byte(stageTestUser1+"\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	args := map[string]interface{}{
		"files":     []interface{}{testFile},
		"filter":    `select(.type == "user")`,
		"transform": ".nonexistent",
	}

	result, err := HandleExecuteStage2Query(context.Background(), args)
	if err != nil {
		t.Fatalf("HandleExecuteStage2Query failed: %v", err)
	}

	// Marshal to JSON and unmarshal to check warnings
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	warningsRaw, ok := parsed["warnings"]
	if !ok {
		t.Fatal("expected 'warnings' key in response")
	}

	warnings, ok := warningsRaw.([]interface{})
	if !ok {
		t.Fatalf("expected warnings to be an array, got %T", warningsRaw)
	}

	if len(warnings) == 0 {
		t.Error("expected non-empty warnings array when transform produces all-null results")
	}
}

func TestHandleExecuteStage2Query_TransformValidField_EmptyWarnings(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_stage_nowarn.jsonl")
	if err := os.WriteFile(testFile, []byte(stageTestUser1+"\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	args := map[string]interface{}{
		"files":     []interface{}{testFile},
		"filter":    `select(.type == "user")`,
		"transform": "{type: .type}",
	}

	result, err := HandleExecuteStage2Query(context.Background(), args)
	if err != nil {
		t.Fatalf("HandleExecuteStage2Query failed: %v", err)
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	warningsRaw, ok := parsed["warnings"]
	if !ok {
		t.Fatal("expected 'warnings' key in response")
	}

	warnings, ok := warningsRaw.([]interface{})
	if !ok {
		t.Fatalf("expected warnings to be an array, got %T", warningsRaw)
	}

	if len(warnings) != 0 {
		t.Errorf("expected empty warnings for valid transform, got: %v", warnings)
	}
}

// TestHandleGetSessionMetadata_SessionScope_SingleFile verifies that HandleGetSessionMetadata
// with scope=session returns metadata for exactly the single current session file.
func TestHandleGetSessionMetadata_SessionScope_SingleFile(t *testing.T) {
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to restore cwd: %v", err)
		}
	}()

	// Set up a fake project session directory
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	projectPath := t.TempDir()
	if err := os.Chdir(projectPath); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	resolvedPath, err := filepath.EvalSymlinks(projectPath)
	if err != nil {
		resolvedPath = projectPath
	}
	// Compute project hash
	h := strings.ReplaceAll(resolvedPath, "\\", "-")
	h = strings.ReplaceAll(h, "/", "-")
	h = strings.ReplaceAll(h, ":", "-")

	sessionDir := filepath.Join(projectsRoot, h)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

	// Create two session files so project scope would return 2
	f1 := filepath.Join(sessionDir, "sess1.jsonl")
	f2 := filepath.Join(sessionDir, "sess2.jsonl")
	if err := os.WriteFile(f1, []byte(`{"type":"user"}`+"\n"), 0644); err != nil {
		t.Fatalf("failed to write sess1: %v", err)
	}
	if err := os.WriteFile(f2, []byte(`{"type":"user"}`+"\n"), 0644); err != nil {
		t.Fatalf("failed to write sess2: %v", err)
	}

	args := map[string]interface{}{
		"scope": "session",
	}

	result, err := HandleGetSessionMetadata(context.Background(), args)
	if err != nil {
		t.Fatalf("HandleGetSessionMetadata failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}

	fileCount, ok := resultMap["file_count"].(int)
	if !ok {
		t.Fatalf("expected file_count to be int, got %T: %v", resultMap["file_count"], resultMap["file_count"])
	}

	if fileCount != 1 {
		t.Errorf("expected session scope to return file_count=1, got %d", fileCount)
	}
}

// ---------------------------------------------------------------------------
// DIR-024: provider-aware get_session_directory / get_session_metadata
// ---------------------------------------------------------------------------

// setupClaudeSessionDir creates a temp Claude projects root + session dir
// (without chdir'ing — tests pass working_dir explicitly instead) and returns
// (projectsRoot, sessionDir, projectPath).
func setupClaudeSessionDir(t *testing.T) (projectsRoot, sessionDir, projectPath string) {
	t.Helper()

	projectsRoot = t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	projectPath = t.TempDir()
	resolvedPath, err := filepath.EvalSymlinks(projectPath)
	if err != nil {
		resolvedPath = projectPath
	}
	hash := strings.ReplaceAll(resolvedPath, "\\", "-")
	hash = strings.ReplaceAll(hash, "/", "-")
	hash = strings.ReplaceAll(hash, ":", "-")

	sessionDir = filepath.Join(projectsRoot, hash)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}
	return projectsRoot, sessionDir, projectPath
}

// setupCodexHomeDir creates a temp Codex home with a state_5.sqlite database
// containing a single thread whose rollout_path/cwd match sessionID/cwd, and
// writes a minimal rollout file at that path. Returns the rollout file path.
func setupCodexHomeDir(t *testing.T, sessionID, cwd string) (codexHome, rolloutPath string) {
	t.Helper()

	codexHome = t.TempDir()
	t.Setenv("META_CC_CODEX_ROOT", codexHome)

	rolloutPath = filepath.Join(codexHome, sessionID+".jsonl")
	content := `{"timestamp":"2026-06-14T06:00:00Z","type":"session_meta","payload":{"id":"` + sessionID + `","cwd":"` + cwd + `"}}
{"timestamp":"2026-06-14T06:00:01Z","type":"event_msg","payload":{"type":"user_message","message":"hello from codex"}}
`
	if err := os.WriteFile(rolloutPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write rollout file: %v", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(codexHome, "state_5.sqlite"))
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE threads (
		id TEXT PRIMARY KEY,
		rollout_path TEXT,
		cwd TEXT,
		title TEXT,
		model TEXT,
		model_provider TEXT,
		tokens_used INTEGER,
		source TEXT,
		created_at INTEGER
	)`); err != nil {
		t.Fatalf("failed to create threads table: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO threads(id, rollout_path, cwd, title, model, model_provider, tokens_used, source, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sessionID, rolloutPath, cwd, "stage test", "gpt-5", "openai", 42, "cli", int64(1700000000)); err != nil {
		t.Fatalf("failed to insert thread row: %v", err)
	}

	return codexHome, rolloutPath
}

// TestHandleGetSessionDirectory_ProviderClaudeDefault_Unchanged verifies that
// omitting "provider" (or passing "claude") preserves the exact pre-DIR-024
// response shape for get_session_directory.
func TestHandleGetSessionDirectory_ProviderClaudeDefault_Unchanged(t *testing.T) {
	_, sessionDir, projectPath := setupClaudeSessionDir(t)
	if err := os.WriteFile(filepath.Join(sessionDir, "s1.jsonl"), []byte(`{"type":"user"}`+"\n"), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	for _, providerArg := range []interface{}{nil, "claude"} {
		args := map[string]interface{}{"scope": "project", "working_dir": projectPath}
		if providerArg != nil {
			args["provider"] = providerArg
		}
		result, err := HandleGetSessionDirectory(context.Background(), args)
		if err != nil {
			t.Fatalf("HandleGetSessionDirectory failed: %v", err)
		}
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map result, got %T", result)
		}
		if resultMap["directory"] != sessionDir {
			t.Errorf("expected directory=%s, got %v", sessionDir, resultMap["directory"])
		}
		if resultMap["provider"] != "claude" {
			t.Errorf("expected provider=claude, got %v", resultMap["provider"])
		}
		if resultMap["file_count"] != 1 {
			t.Errorf("expected file_count=1, got %v", resultMap["file_count"])
		}
	}
}

// TestHandleGetSessionDirectory_ProviderCodex_ReturnsOnlyRolloutFiles verifies
// that provider=codex resolves only Codex rollout files, never Claude paths.
func TestHandleGetSessionDirectory_ProviderCodex_ReturnsOnlyRolloutFiles(t *testing.T) {
	// Make Claude unavailable/irrelevant: a Claude project directory that has
	// no bearing on the Codex result should never leak into it.
	_, claudeSessionDir, projectPath := setupClaudeSessionDir(t)
	if err := os.WriteFile(filepath.Join(claudeSessionDir, "claude-session.jsonl"), []byte(`{"type":"user"}`+"\n"), 0644); err != nil {
		t.Fatalf("failed to write claude session file: %v", err)
	}

	_, rolloutPath := setupCodexHomeDir(t, "codex-sess-1", projectPath)

	result, err := HandleGetSessionDirectory(context.Background(), map[string]interface{}{
		"scope":       "project",
		"provider":    "codex",
		"working_dir": projectPath,
	})
	if err != nil {
		t.Fatalf("HandleGetSessionDirectory(provider=codex) failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	if resultMap["provider"] != "codex" {
		t.Errorf("expected provider=codex, got %v", resultMap["provider"])
	}
	files, ok := resultMap["files"].([]string)
	if !ok {
		t.Fatalf("expected files to be []string, got %T", resultMap["files"])
	}
	if len(files) != 1 || files[0] != rolloutPath {
		t.Errorf("expected files=[%s], got %v", rolloutPath, files)
	}
	for _, f := range files {
		if strings.Contains(f, claudeSessionDir) {
			t.Errorf("codex result leaked a claude file path: %s", f)
		}
	}
}

// TestHandleGetSessionDirectory_ProviderCodex_Unavailable_FailsClosed verifies
// that a missing Codex state database produces an actionable error rather
// than an empty/successful result that could be mistaken for "no data".
func TestHandleGetSessionDirectory_ProviderCodex_Unavailable_FailsClosed(t *testing.T) {
	t.Setenv("META_CC_CODEX_ROOT", t.TempDir()) // no state_5.sqlite present
	projectPath := t.TempDir()

	_, err := HandleGetSessionDirectory(context.Background(), map[string]interface{}{
		"scope":       "project",
		"provider":    "codex",
		"working_dir": projectPath,
	})
	if err == nil {
		t.Fatal("expected error for unavailable codex provider, got nil")
	}
	if !strings.Contains(err.Error(), "codex provider unavailable") {
		t.Errorf("expected actionable 'codex provider unavailable' error, got: %v", err)
	}
}

// TestHandleGetSessionDirectory_ProviderAll_PerProviderBreakdown verifies
// that provider=all returns an explicit per-provider breakdown rather than
// merging Claude and Codex files into one directory/schema.
func TestHandleGetSessionDirectory_ProviderAll_PerProviderBreakdown(t *testing.T) {
	_, claudeSessionDir, projectPath := setupClaudeSessionDir(t)
	if err := os.WriteFile(filepath.Join(claudeSessionDir, "claude-session.jsonl"), []byte(`{"type":"user"}`+"\n"), 0644); err != nil {
		t.Fatalf("failed to write claude session file: %v", err)
	}
	_, rolloutPath := setupCodexHomeDir(t, "codex-sess-1", projectPath)

	result, err := HandleGetSessionDirectory(context.Background(), map[string]interface{}{
		"scope":       "project",
		"provider":    "all",
		"working_dir": projectPath,
	})
	if err != nil {
		t.Fatalf("HandleGetSessionDirectory(provider=all) failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	if resultMap["provider"] != "all" {
		t.Errorf("expected provider=all, got %v", resultMap["provider"])
	}
	providers, ok := resultMap["providers"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected providers to be a map, got %T", resultMap["providers"])
	}

	claudeResult, ok := providers["claude"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected providers.claude to be a map, got %T", providers["claude"])
	}
	if claudeResult["directory"] != claudeSessionDir {
		t.Errorf("expected claude directory=%s, got %v", claudeSessionDir, claudeResult["directory"])
	}

	codexResult, ok := providers["codex"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected providers.codex to be a map, got %T", providers["codex"])
	}
	codexFiles, ok := codexResult["files"].([]string)
	if !ok || len(codexFiles) != 1 || codexFiles[0] != rolloutPath {
		t.Errorf("expected codex files=[%s], got %v", rolloutPath, codexResult["files"])
	}
}

// TestHandleGetSessionDirectory_InvalidProvider_FailsClosed verifies that an
// unrecognized provider name fails with an actionable error instead of
// silently defaulting to Claude.
func TestHandleGetSessionDirectory_InvalidProvider_FailsClosed(t *testing.T) {
	_, err := HandleGetSessionDirectory(context.Background(), map[string]interface{}{
		"scope":    "project",
		"provider": "bogus",
	})
	if err == nil {
		t.Fatal("expected error for invalid provider, got nil")
	}
	if !strings.Contains(err.Error(), "invalid provider") {
		t.Errorf("expected 'invalid provider' error, got: %v", err)
	}
}

// TestHandleGetSessionMetadata_ProviderClaudeDefault_Unchanged verifies that
// get_session_metadata's default (provider=claude) response shape is
// unaffected by DIR-024.
func TestHandleGetSessionMetadata_ProviderClaudeDefault_Unchanged(t *testing.T) {
	_, sessionDir, projectPath := setupClaudeSessionDir(t)
	if err := os.WriteFile(filepath.Join(sessionDir, "s1.jsonl"), []byte(`{"type":"user"}`+"\n"), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	result, err := HandleGetSessionMetadata(context.Background(), map[string]interface{}{
		"scope":       "project",
		"working_dir": projectPath,
	})
	if err != nil {
		t.Fatalf("HandleGetSessionMetadata failed: %v", err)
	}
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	if resultMap["provider"] != "claude" {
		t.Errorf("expected provider=claude, got %v", resultMap["provider"])
	}
	if resultMap["file_count"] != 1 {
		t.Errorf("expected file_count=1, got %v", resultMap["file_count"])
	}
	if _, ok := resultMap["jsonl_schema"]; !ok {
		t.Error("expected jsonl_schema field to be present")
	}
}

// TestHandleGetSessionMetadata_ProviderCodex_ReturnsCodexSchema verifies that
// provider=codex returns Codex rollout files with Codex-specific schema
// hints rather than the Claude JSONL schema.
func TestHandleGetSessionMetadata_ProviderCodex_ReturnsCodexSchema(t *testing.T) {
	projectPath := t.TempDir()
	_, rolloutPath := setupCodexHomeDir(t, "codex-sess-2", projectPath)

	result, err := HandleGetSessionMetadata(context.Background(), map[string]interface{}{
		"scope":       "project",
		"provider":    "codex",
		"working_dir": projectPath,
	})
	if err != nil {
		t.Fatalf("HandleGetSessionMetadata(provider=codex) failed: %v", err)
	}
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	if resultMap["provider"] != "codex" {
		t.Errorf("expected provider=codex, got %v", resultMap["provider"])
	}

	files, ok := resultMap["files"].([]map[string]interface{})
	if !ok || len(files) != 1 {
		t.Fatalf("expected 1 codex file entry, got %#v", resultMap["files"])
	}
	if files[0]["path"] != rolloutPath {
		t.Errorf("expected file path=%s, got %v", rolloutPath, files[0]["path"])
	}

	schema, ok := resultMap["jsonl_schema"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected jsonl_schema to be a map, got %T", resultMap["jsonl_schema"])
	}
	if _, ok := schema["legacy_response_item_fields"]; !ok {
		t.Error("expected codex-specific 'legacy_response_item_fields' schema section")
	}
	if _, ok := schema["user_message_fields"]; ok {
		t.Error("codex jsonl_schema should not reuse the claude-specific 'user_message_fields' key")
	}
}

// TestHandleGetSessionMetadata_ProviderAll_PerProviderBreakdown verifies that
// provider=all keeps Claude and Codex schemas/files separate.
func TestHandleGetSessionMetadata_ProviderAll_PerProviderBreakdown(t *testing.T) {
	_, sessionDir, projectPath := setupClaudeSessionDir(t)
	if err := os.WriteFile(filepath.Join(sessionDir, "s1.jsonl"), []byte(`{"type":"user"}`+"\n"), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}
	setupCodexHomeDir(t, "codex-sess-3", projectPath)

	result, err := HandleGetSessionMetadata(context.Background(), map[string]interface{}{
		"scope":       "project",
		"provider":    "all",
		"working_dir": projectPath,
	})
	if err != nil {
		t.Fatalf("HandleGetSessionMetadata(provider=all) failed: %v", err)
	}
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	providers, ok := resultMap["providers"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected providers to be a map, got %T", resultMap["providers"])
	}
	if _, ok := providers["claude"]; !ok {
		t.Error("expected providers.claude to be present")
	}
	if _, ok := providers["codex"]; !ok {
		t.Error("expected providers.codex to be present")
	}
}

// TestHandleGetSessionMetadata_InvalidProvider_FailsClosed verifies that an
// unrecognized provider name fails closed for get_session_metadata too.
func TestHandleGetSessionMetadata_InvalidProvider_FailsClosed(t *testing.T) {
	_, err := HandleGetSessionMetadata(context.Background(), map[string]interface{}{
		"scope":    "project",
		"provider": "bogus",
	})
	if err == nil {
		t.Fatal("expected error for invalid provider, got nil")
	}
	if !strings.Contains(err.Error(), "invalid provider") {
		t.Errorf("expected 'invalid provider' error, got: %v", err)
	}
}

// TestHandleGetSessionDirectory_ProviderAll_BothUnavailable_FailsClosed
// verifies that provider=all fails outright (rather than returning an empty
// success) when neither provider has any data.
func TestHandleGetSessionDirectory_ProviderAll_BothUnavailable_FailsClosed(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	t.Setenv("META_CC_CODEX_ROOT", t.TempDir())
	projectPath := t.TempDir()

	_, err := HandleGetSessionDirectory(context.Background(), map[string]interface{}{
		"scope":       "project",
		"provider":    "all",
		"working_dir": projectPath,
	})
	if err == nil {
		t.Fatal("expected error when neither provider has data, got nil")
	}
	if !strings.Contains(err.Error(), "no session data available for any provider") {
		t.Errorf("expected actionable 'no session data available' error, got: %v", err)
	}
}

// TestHandleGetSessionMetadata_ProviderAll_BothUnavailable_FailsClosed mirrors
// the above for get_session_metadata.
func TestHandleGetSessionMetadata_ProviderAll_BothUnavailable_FailsClosed(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())
	t.Setenv("META_CC_CODEX_ROOT", t.TempDir())
	projectPath := t.TempDir()

	_, err := HandleGetSessionMetadata(context.Background(), map[string]interface{}{
		"scope":       "project",
		"provider":    "all",
		"working_dir": projectPath,
	})
	if err == nil {
		t.Fatal("expected error when neither provider has data, got nil")
	}
	if !strings.Contains(err.Error(), "no session data available for any provider") {
		t.Errorf("expected actionable 'no session data available' error, got: %v", err)
	}
}

// TestCommonDirectory covers the shared-parent detection used by the Codex
// discovery response: same directory collapses to one path, divergent
// directories collapse to nil, and an empty list also yields nil.
func TestCommonDirectory(t *testing.T) {
	if got := commonDirectory(nil); got != nil {
		t.Errorf("commonDirectory(nil) = %v, want nil", got)
	}
	if got := commonDirectory([]string{"/a/b/1.jsonl", "/a/b/2.jsonl"}); got != "/a/b" {
		t.Errorf("commonDirectory(same dir) = %v, want /a/b", got)
	}
	if got := commonDirectory([]string{"/a/b/1.jsonl", "/a/c/2.jsonl"}); got != nil {
		t.Errorf("commonDirectory(diverging dirs) = %v, want nil", got)
	}
}

// TestHandleGetSessionDirectory_WorkingDirOverride verifies that an explicit
// working_dir argument is honored for the default (Claude) provider.
func TestHandleGetSessionDirectory_WorkingDirOverride(t *testing.T) {
	_, sessionDir, projectPath := setupClaudeSessionDir(t)
	if err := os.WriteFile(filepath.Join(sessionDir, "s1.jsonl"), []byte(`{"type":"user"}`+"\n"), 0644); err != nil {
		t.Fatalf("failed to write session file: %v", err)
	}

	// Run from an unrelated cwd; working_dir must still resolve to projectPath.
	unrelatedDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to restore cwd: %v", err)
		}
	}()
	if err := os.Chdir(unrelatedDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	result, err := HandleGetSessionDirectory(context.Background(), map[string]interface{}{
		"scope":       "project",
		"working_dir": projectPath,
	})
	if err != nil {
		t.Fatalf("HandleGetSessionDirectory failed: %v", err)
	}
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}
	if resultMap["directory"] != sessionDir {
		t.Errorf("expected directory=%s, got %v", sessionDir, resultMap["directory"])
	}
}
