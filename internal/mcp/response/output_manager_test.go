package response

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetSessionCacheState() {
	sessionCacheDir = ""
	sessionCacheInitErr = nil
	sessionCacheOnce = sync.Once{}
}

func TestSessionCacheLifecycle(t *testing.T) {
	resetSessionCacheState()
	t.Cleanup(resetSessionCacheState)
	t.Setenv("CLAUDE_CODE_SESSION_ID", "coverage-test")
	t.Setenv("TMPDIR", t.TempDir())

	dir, err := GetSessionCacheDir()
	require.NoError(t, err)
	assert.Contains(t, dir, "claude-session-coverage-test")
	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
	second, err := GetSessionCacheDir()
	require.NoError(t, err)
	assert.Equal(t, dir, second)
	require.NoError(t, CleanupSessionCache())
	_, err = os.Stat(filepath.Dir(dir))
	assert.True(t, os.IsNotExist(err))
}

func TestCleanupSessionCacheWithoutInitialization(t *testing.T) {
	resetSessionCacheState()
	t.Cleanup(resetSessionCacheState)
	require.NoError(t, CleanupSessionCache())
}

func TestWriteJSONLFileWritesAtomically(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "records.jsonl")
	require.NoError(t, WriteJSONLFile(path, []interface{}{map[string]interface{}{"id": 1}, "text"}))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "{\"id\":1}\n\"text\"\n", string(data))
	_, err = os.Stat(path + ".tmp")
	assert.True(t, os.IsNotExist(err))
}

func TestWriteJSONLFileReportsCreateAndEncodeErrors(t *testing.T) {
	parentFile := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(parentFile, []byte("x"), 0o600))
	err := WriteJSONLFile(filepath.Join(parentFile, "child.jsonl"), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create directory")

	path := filepath.Join(t.TempDir(), "bad.jsonl")
	err = WriteJSONLFile(path, []interface{}{func() {}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to encode record")
	_, statErr := os.Stat(path + ".tmp")
	assert.True(t, os.IsNotExist(statErr))
}

func TestCleanupOldFilesAndToolResult(t *testing.T) {
	oldPath := CreateTempFilePath("cleanup-old", "test")
	newPath := CreateTempFilePath("cleanup-new", "test")
	require.NoError(t, os.WriteFile(oldPath, []byte("old"), 0o600))
	require.NoError(t, os.WriteFile(newPath, []byte("new"), 0o600))
	t.Cleanup(func() { _ = os.Remove(oldPath); _ = os.Remove(newPath) })
	oldTime := time.Now().Add(-48 * time.Hour)
	require.NoError(t, os.Chtimes(oldPath, oldTime, oldTime))

	removed, freed, err := CleanupOldFiles(1)
	require.NoError(t, err)
	assert.Contains(t, removed, oldPath)
	assert.GreaterOrEqual(t, freed, int64(3))
	_, err = os.Stat(newPath)
	require.NoError(t, err)

	result, err := ExecuteCleanupTool(map[string]interface{}{"max_age_days": float64(-1)})
	require.NoError(t, err)
	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(result), &decoded))
	assert.Contains(t, decoded, "removed_count")
}

func TestGetIntParam(t *testing.T) {
	assert.Equal(t, 2, getIntParam(map[string]interface{}{"v": float64(2)}, "v", 7))
	assert.Equal(t, 3, getIntParam(map[string]interface{}{"v": 3}, "v", 7))
	assert.Equal(t, 7, getIntParam(map[string]interface{}{"v": strings.NewReader("bad")}, "v", 7))
}
