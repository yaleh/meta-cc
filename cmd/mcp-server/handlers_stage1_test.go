package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleGetSessionDirectory_ProjectScope tests project scope with existing files
func TestHandleGetSessionDirectory_ProjectScope(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(originalWd))
	}()

	// Create temp directory as mock Claude projects root
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	// Create temp directory as project path
	projectPath := t.TempDir()

	// Change to project directory
	require.NoError(t, os.Chdir(projectPath))

	// Resolve symlinks for consistent hashing
	resolvedPath, err2 := filepath.EvalSymlinks(projectPath)
	if err2 != nil {
		resolvedPath = projectPath
	}
	projectHash := pathToHash(resolvedPath)

	// Create session directory
	sessionDir := filepath.Join(projectsRoot, projectHash)
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	// Create test session files with different timestamps
	file1 := filepath.Join(sessionDir, "session1.jsonl")
	file2 := filepath.Join(sessionDir, "session2.jsonl")

	require.NoError(t, os.WriteFile(file1, []byte(`{"type":"user","timestamp":"2025-01-15T10:30:00Z"}`+"\n"), 0644))
	time.Sleep(10 * time.Millisecond) // Ensure different mtimes
	require.NoError(t, os.WriteFile(file2, []byte(`{"type":"assistant","timestamp":"2025-10-26T14:20:00Z"}`+"\n"), 0644))

	// Execute
	args := map[string]interface{}{
		"scope": "project",
	}

	result, err := handleGetSessionDirectory(context.Background(), args)
	require.NoError(t, err)

	// Verify result structure
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "result should be a map")

	assert.Equal(t, sessionDir, resultMap["directory"])
	assert.Equal(t, "project", resultMap["scope"])
	assert.Equal(t, 2, resultMap["file_count"])

	// Verify total_size_bytes is reasonable
	totalSize, ok := resultMap["total_size_bytes"].(int64)
	require.True(t, ok)
	assert.Greater(t, totalSize, int64(0))

	// Verify time range
	oldest, ok := resultMap["oldest_file"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, oldest)

	newest, ok := resultMap["newest_file"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, newest)

	// Verify oldest < newest
	oldestTime, err := time.Parse(time.RFC3339, oldest)
	require.NoError(t, err)
	newestTime, err := time.Parse(time.RFC3339, newest)
	require.NoError(t, err)
	assert.True(t, oldestTime.Before(newestTime) || oldestTime.Equal(newestTime))
}

// TestHandleGetSessionDirectory_SessionScope tests session scope
func TestHandleGetSessionDirectory_SessionScope(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(originalWd))
	}()

	// Create temp directory as mock Claude projects root
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	// Create temp directory as project path
	projectPath := t.TempDir()

	// Change to project directory
	require.NoError(t, os.Chdir(projectPath))

	// Resolve symlinks for consistent hashing
	resolvedPath, err2 := filepath.EvalSymlinks(projectPath)
	if err2 != nil {
		resolvedPath = projectPath
	}
	projectHash := pathToHash(resolvedPath)

	// Create session directory
	sessionDir := filepath.Join(projectsRoot, projectHash)
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	// Create test session file
	sessionFile := filepath.Join(sessionDir, "current-session.jsonl")
	testData := `{"type":"user","timestamp":"2025-10-26T14:20:00Z"}` + "\n"
	require.NoError(t, os.WriteFile(sessionFile, []byte(testData), 0644))

	// Execute
	args := map[string]interface{}{
		"scope": "session",
	}

	result, err := handleGetSessionDirectory(context.Background(), args)
	require.NoError(t, err)

	// Verify result structure
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, sessionDir, resultMap["directory"])
	assert.Equal(t, "session", resultMap["scope"])
	assert.Equal(t, 1, resultMap["file_count"])

	totalSize, ok := resultMap["total_size_bytes"].(int64)
	require.True(t, ok)
	assert.Equal(t, int64(len(testData)), totalSize)
}

// TestHandleGetSessionDirectory_EmptyDirectory tests empty directory
func TestHandleGetSessionDirectory_EmptyDirectory(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(originalWd))
	}()

	// Create temp directory as mock Claude projects root
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	// Create temp directory as project path
	projectPath := t.TempDir()

	// Change to project directory
	require.NoError(t, os.Chdir(projectPath))

	// Resolve symlinks for consistent hashing
	resolvedPath, err2 := filepath.EvalSymlinks(projectPath)
	if err2 != nil {
		resolvedPath = projectPath
	}
	projectHash := pathToHash(resolvedPath)

	// Create empty session directory
	sessionDir := filepath.Join(projectsRoot, projectHash)
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	// Create at least one .jsonl file so the locator doesn't fail
	// (empty directory means no sessions found, which is tested in DirectoryNotFound)
	dummyFile := filepath.Join(sessionDir, "dummy.jsonl")
	require.NoError(t, os.WriteFile(dummyFile, []byte{}, 0644))

	// Execute
	args := map[string]interface{}{
		"scope": "project",
	}

	result, err := handleGetSessionDirectory(context.Background(), args)
	require.NoError(t, err)

	// Verify result for directory with empty file
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, sessionDir, resultMap["directory"])
	assert.Equal(t, "project", resultMap["scope"])
	assert.Equal(t, 1, resultMap["file_count"])
	assert.Equal(t, int64(0), resultMap["total_size_bytes"])
	assert.NotEmpty(t, resultMap["oldest_file"])
	assert.NotEmpty(t, resultMap["newest_file"])
}

// TestHandleGetSessionDirectory_InvalidScope tests invalid scope parameter
func TestHandleGetSessionDirectory_InvalidScope(t *testing.T) {
	args := map[string]interface{}{
		"scope": "invalid",
	}

	_, err := handleGetSessionDirectory(context.Background(), args)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid scope")
}

// TestHandleGetSessionDirectory_MissingScope tests missing scope parameter
func TestHandleGetSessionDirectory_MissingScope(t *testing.T) {
	args := map[string]interface{}{}

	_, err := handleGetSessionDirectory(context.Background(), args)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scope parameter is required")
}

// TestHandleGetSessionDirectory_DirectoryNotFound tests non-existent directory
func TestHandleGetSessionDirectory_DirectoryNotFound(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(originalWd))
	}()

	// Create temp directory as mock Claude projects root (but no sessions)
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))

	// Create temp directory as project path
	projectPath := t.TempDir()

	// Change to project directory
	require.NoError(t, os.Chdir(projectPath))

	// Don't create any session files

	args := map[string]interface{}{
		"scope": "project",
	}

	_, err2 := handleGetSessionDirectory(context.Background(), args)
	require.Error(t, err2)
	assert.Contains(t, err2.Error(), "no sessions found for project")
}

// TestGetSessionDirectory_JSON tests JSON serialization
func TestGetSessionDirectory_JSON(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(originalWd))
	}()

	// Create temp directory as mock Claude projects root
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	// Create temp directory as project path
	projectPath := t.TempDir()

	// Change to project directory
	require.NoError(t, os.Chdir(projectPath))

	// Resolve symlinks for consistent hashing
	resolvedPath, err2 := filepath.EvalSymlinks(projectPath)
	if err2 != nil {
		resolvedPath = projectPath
	}
	projectHash := pathToHash(resolvedPath)

	// Create session directory
	sessionDir := filepath.Join(projectsRoot, projectHash)
	require.NoError(t, os.MkdirAll(sessionDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "test.jsonl"), []byte(`{"test":"data"}`+"\n"), 0644))

	args := map[string]interface{}{
		"scope": "project",
	}

	result, err := handleGetSessionDirectory(context.Background(), args)
	require.NoError(t, err)

	// Verify JSON serialization
	jsonBytes, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(jsonBytes, &decoded))

	assert.Equal(t, sessionDir, decoded["directory"])
	assert.Equal(t, "project", decoded["scope"])
	assert.Equal(t, float64(1), decoded["file_count"]) // JSON unmarshals numbers as float64
}

// TestHandleGetSessionDirectory_SubagentFileCount tests that the response includes
// a subagent_file_count field.
func TestHandleGetSessionDirectory_SubagentFileCount(t *testing.T) {
	// Save original working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(originalWd))
	}()

	// Create temp directory as mock Claude projects root
	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	// Create temp directory as project path
	projectPath := t.TempDir()
	require.NoError(t, os.Chdir(projectPath))

	resolvedPath, err2 := filepath.EvalSymlinks(projectPath)
	if err2 != nil {
		resolvedPath = projectPath
	}
	projectHash := pathToHash(resolvedPath)

	sessionDir := filepath.Join(projectsRoot, projectHash)
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	// Create a top-level session file (UUID = "session1")
	sessionUUID := "session1"
	sessionFile := filepath.Join(sessionDir, sessionUUID+".jsonl")
	require.NoError(t, os.WriteFile(sessionFile, []byte(`{"type":"user"}`+"\n"), 0644))

	// Create subagent files under <uuid>/subagents/
	subDir := filepath.Join(sessionDir, sessionUUID, "subagents")
	require.NoError(t, os.MkdirAll(subDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "sub1.jsonl"), []byte(`{"type":"user"}`+"\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "sub2.jsonl"), []byte(`{"type":"user"}`+"\n"), 0644))

	args := map[string]interface{}{
		"scope": "project",
	}

	result, err := handleGetSessionDirectory(context.Background(), args)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "result should be a map")

	// Verify subagent_file_count is present and correct
	subCount, ok := resultMap["subagent_file_count"]
	require.True(t, ok, "result should contain subagent_file_count field")
	assert.Equal(t, 2, subCount, "expected 2 subagent files")

	// file_count should only count top-level files
	assert.Equal(t, 1, resultMap["file_count"], "file_count should be 1 (top-level only)")
}

// TestHandleGetSessionDirectory_SubagentFileCount_WithSubagents creates dirs with
// subagents and confirms subagent_file_count > 0.
func TestHandleGetSessionDirectory_SubagentFileCount_WithSubagents(t *testing.T) {
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(originalWd))
	}()

	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	projectPath := t.TempDir()
	require.NoError(t, os.Chdir(projectPath))

	resolvedPath, err2 := filepath.EvalSymlinks(projectPath)
	if err2 != nil {
		resolvedPath = projectPath
	}
	projectHash := pathToHash(resolvedPath)

	sessionDir := filepath.Join(projectsRoot, projectHash)
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	// Top-level session
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "main.jsonl"),
		[]byte(`{"type":"user"}`+"\n"), 0644))

	// Subagent files
	subDir := filepath.Join(sessionDir, "main", "subagents")
	require.NoError(t, os.MkdirAll(subDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "agent.jsonl"),
		[]byte(`{"type":"user"}`+"\n"), 0644))

	result, err := handleGetSessionDirectory(context.Background(), map[string]interface{}{"scope": "project"})
	require.NoError(t, err)

	resultMap := result.(map[string]interface{})
	subCount, ok := resultMap["subagent_file_count"]
	require.True(t, ok, "subagent_file_count field must be present")

	subCountInt, ok := subCount.(int)
	require.True(t, ok, "subagent_file_count must be an int")
	assert.Greater(t, subCountInt, 0, "expected subagent_file_count > 0")
}

// TestHandleGetSessionDirectory_SubagentFileCount_NoSubagents confirms subagent_file_count==0
// when no subagent directories exist.
func TestHandleGetSessionDirectory_SubagentFileCount_NoSubagents(t *testing.T) {
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(originalWd))
	}()

	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)

	projectPath := t.TempDir()
	require.NoError(t, os.Chdir(projectPath))

	resolvedPath, err2 := filepath.EvalSymlinks(projectPath)
	if err2 != nil {
		resolvedPath = projectPath
	}
	projectHash := pathToHash(resolvedPath)

	sessionDir := filepath.Join(projectsRoot, projectHash)
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	// Only top-level session files, no subagent subdirs
	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "session.jsonl"),
		[]byte(`{"type":"user"}`+"\n"), 0644))

	result, err := handleGetSessionDirectory(context.Background(), map[string]interface{}{"scope": "project"})
	require.NoError(t, err)

	resultMap := result.(map[string]interface{})
	subCount, ok := resultMap["subagent_file_count"]
	require.True(t, ok, "subagent_file_count field must be present")
	assert.Equal(t, 0, subCount, "expected subagent_file_count == 0 when no subagent dirs exist")
}

// TestCountLines_LargeImageLine_NoError verifies that countLines handles lines
// larger than 10MB (e.g. base64 image data) without error and returns correct count.
func TestCountLines_LargeImageLine_NoError(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "large.jsonl")

	// Build a line with a 5MB payload
	imageData := strings.Repeat("A", 5*1024*1024)
	line1 := `{"type":"user","data":"` + imageData + `"}`
	line2 := `{"type":"assistant","message":"normal"}`
	line3 := `{"type":"user","message":"another"}`

	content := line1 + "\n" + line2 + "\n" + line3 + "\n"
	require.NoError(t, os.WriteFile(file, []byte(content), 0644))

	count, err := countLines(file)
	require.NoError(t, err, "countLines should not error on large line")
	assert.Equal(t, 3, count, "should count 3 lines")
}

// TestCountLines_EmptyFile verifies that countLines returns 0 for an empty file.
func TestCountLines_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "empty.jsonl")
	require.NoError(t, os.WriteFile(file, []byte(""), 0644))

	count, err := countLines(file)
	require.NoError(t, err, "countLines should not error on empty file")
	assert.Equal(t, 0, count, "empty file should have 0 lines")
}

// TestCountLines_NoTrailingNewline verifies that countLines correctly counts
// a file where the last line has no trailing newline.
func TestCountLines_NoTrailingNewline(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "notrail.jsonl")

	// 3 lines, last line has no trailing newline
	content := `{"type":"user"}` + "\n" + `{"type":"assistant"}` + "\n" + `{"type":"user"}`
	require.NoError(t, os.WriteFile(file, []byte(content), 0644))

	count, err := countLines(file)
	require.NoError(t, err, "countLines should not error on file without trailing newline")
	assert.Equal(t, 3, count, "should count 3 lines even without trailing newline")
}
