package query

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/testutil"
)

// DIR-098: malformed-file health in the two discovery tools.
//
// The 2026-07-30 dogfooding failure was not that a corrupt transcript existed —
// it was that nothing could tell you one did. get_session_directory and
// inspect_session_files enumerated files without reporting their health, so a
// 0-byte transcript sat in the project directory until a query crashed on it.
// These tests pin the contract that closes that gap, using the SAME shared
// fixture (internal/testutil) the cross-tool tolerance gate uses, so the shapes
// the tolerance gate installs and the shapes these tools report cannot drift
// apart.

// writeCorpusFile writes one session file into a corpus directory.
func writeCorpusFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

// installSharedCorruptShapes writes every shape from DIR-099's fixture — the
// 0-byte 8eda8f4e transcript included — into a session directory, and returns
// the paths in fixture order.
func installSharedCorruptShapes(t *testing.T, dir string) []string {
	t.Helper()
	paths := make([]string, 0, len(testutil.CorruptFiles()))
	for _, corrupt := range testutil.CorruptFiles() {
		paths = append(paths, writeCorpusFile(t, dir, corrupt.Name+".jsonl",
			string(testutil.CorpusFile(t, corrupt.Fixture))))
	}
	return paths
}

// decodeResponse unmarshals a get_session_directory response.
func decodeResponse(t *testing.T, result interface{}) map[string]interface{} {
	t.Helper()
	encoded, err := json.Marshal(result)
	require.NoError(t, err)
	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	return decoded
}

// malformedFilesOf returns the {file, reason} entries a response reports, and
// requires the key to be present: a caller must be able to distinguish "none
// found" from "not reported".
func malformedFilesOf(t *testing.T, response map[string]interface{}) []map[string]interface{} {
	t.Helper()
	raw, present := response["malformed_files"]
	require.True(t, present, "the response must always carry malformed_files; got keys %v", keysOf(response))
	require.NotNil(t, raw, "malformed_files must be an empty list, never null")
	entries, ok := raw.([]interface{})
	require.True(t, ok, "malformed_files must be a list; got %T", raw)

	decoded := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		object, ok := entry.(map[string]interface{})
		require.True(t, ok, "each malformed_files entry must be an object; got %v", entry)
		decoded = append(decoded, object)
	}
	return decoded
}

func keysOf(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// ── get_session_directory ────────────────────────────────────────────────────

// AC: on the machine's project corpus — the 8eda8f4e-style empty file this
// task was filed from — get_session_directory NAMES the file as malformed WITH
// a reason, in the same call that lists the corpus. No query tool is invoked
// anywhere in this test: the finding is discoverable before anything crashes.
func TestHandleGetSessionDirectory_NamesTheEmptyFileAsMalformed(t *testing.T) {
	_, sessionDir, projectPath := setupClaudeSessionDir(t)
	writeCorpusFile(t, sessionDir, "healthy.jsonl", `{"type":"user"}`+"\n")
	corruptPaths := installSharedCorruptShapes(t, sessionDir)

	result, err := HandleGetSessionDirectory(context.Background(), map[string]interface{}{
		"scope": "project", "working_dir": projectPath,
	})
	require.NoError(t, err)

	malformed := malformedFilesOf(t, decodeResponse(t, result))

	named := make(map[string]string, len(malformed))
	for _, entry := range malformed {
		file, ok := entry["file"].(string)
		require.True(t, ok, "a malformed entry must name a file; got %v", entry)
		reason, ok := entry["reason"].(string)
		require.True(t, ok, "a malformed entry must carry a reason; got %v", entry)
		require.NotEmpty(t, reason, "%s must be named WITH a reason, not just flagged", file)
		named[file] = reason
	}

	// The exact file from the finding, by the ID it was filed under.
	emptyFile := filepath.Join(sessionDir, "8eda8f4e-2c74-4176-ba6b-8c45e890df42.jsonl")
	require.Contains(t, named, emptyFile,
		"the 8eda8f4e-style empty transcript must be named as malformed; got %v", keysOfStrings(named))
	assert.Contains(t, named[emptyFile], "empty",
		"the reason must say the file is empty, got %q", named[emptyFile])

	// Every corrupt shape, not just the empty one: truncation and wrong-shape
	// are the same class of surprise.
	for _, path := range corruptPaths {
		assert.Contains(t, named, path, "every unusable file must be named; got %v", keysOfStrings(named))
	}
	// And the healthy file must NOT be flagged: a probe that cries wolf is
	// worse than none, because the field stops being read.
	assert.NotContains(t, named, filepath.Join(sessionDir, "healthy.jsonl"),
		"a healthy transcript must not be reported as malformed")
}

func keysOfStrings(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// AC: a healthy corpus produces an EMPTY malformed_files list, and otherwise
// unchanged output. "Otherwise unchanged" is pinned mechanically as a key set:
// the pre-DIR-098 keys are still there, and malformed_files is the only
// addition, so this task cannot quietly reshape the response.
func TestHandleGetSessionDirectory_HealthyCorpusIsUnchanged(t *testing.T) {
	_, sessionDir, projectPath := setupClaudeSessionDir(t)
	writeCorpusFile(t, sessionDir, "s1.jsonl", `{"type":"user"}`+"\n")

	result, err := HandleGetSessionDirectory(context.Background(), map[string]interface{}{
		"scope": "project", "working_dir": projectPath,
	})
	require.NoError(t, err)
	response := decodeResponse(t, result)

	assert.Empty(t, malformedFilesOf(t, response), "a healthy corpus must report no malformed files")

	// The pre-DIR-098 response, verbatim: every key it carried still carries
	// the same value.
	assert.Equal(t, sessionDir, response["directory"])
	assert.Equal(t, "claude", response["provider"])
	assert.Equal(t, "project", response["scope"])
	assert.Equal(t, float64(1), response["file_count"])
	assert.Equal(t, float64(0), response["subagent_file_count"])
	assert.Contains(t, response, "total_size_bytes")
	assert.Contains(t, response, "oldest_file")
	assert.Contains(t, response, "newest_file")

	assert.Equal(t,
		[]string{"directory", "file_count", "malformed_files", "newest_file", "oldest_file",
			"provider", "scope", "subagent_file_count", "total_size_bytes"},
		keysOf(response),
		"malformed_files must be the ONLY addition to the response shape")
}

// ── inspect_session_files ────────────────────────────────────────────────────

// AC: inspect_session_files reports a parse error as structured data for the
// bad file instead of failing the call — and still inspects its siblings at
// full length.
func TestHandleInspectSessionFiles_ReportsBadFilesWithoutFailingTheCall(t *testing.T) {
	dir := t.TempDir()
	good := writeCorpusFile(t, dir, "good.jsonl",
		`{"type":"user","timestamp":"2026-07-30T10:00:00Z"}`+"\n"+
			`{"type":"assistant","timestamp":"2026-07-30T10:00:01Z"}`+"\n")
	empty := writeCorpusFile(t, dir, "8eda8f4e-2c74-4176-ba6b-8c45e890df42.jsonl", "")
	truncated := writeCorpusFile(t, dir, "truncated.jsonl", `{"type":"user"`+"\n")
	missing := filepath.Join(dir, "missing.jsonl")

	result, err := HandleInspectSessionFiles(context.Background(), map[string]interface{}{
		"files": []interface{}{good, empty, truncated, missing},
	})
	require.NoError(t, err, "a bad file must be reported, not raised")

	response := decodeResponse(t, result)
	malformed := malformedFilesOf(t, response)
	named := make(map[string]string, len(malformed))
	for _, entry := range malformed {
		file, _ := entry["file"].(string)
		reason, _ := entry["reason"].(string)
		require.NotEmpty(t, reason, "%s must be named WITH a reason", file)
		named[file] = reason
	}

	require.Contains(t, named, empty, "the 8eda8f4e-style empty file must be named as malformed")
	assert.Contains(t, named[empty], "empty")
	require.Contains(t, named, truncated, "the truncated file must be named as malformed")
	assert.Contains(t, named[truncated], "invalid JSON", "the reason must name the failure, got %q", named[truncated])
	require.Contains(t, named, missing, "an unreadable path must be named as malformed")
	assert.Contains(t, named[missing], "cannot stat file")
	assert.NotContains(t, named, good, "the healthy file must not be reported as malformed")

	// The usable file is still inspected at full length, and the summary
	// reconciles with the health channel rather than silently shrinking.
	assert.Equal(t, float64(4), response["summary"].(map[string]interface{})["total_files"],
		"total_files counts the paths the caller asked about, so it always reconciles with malformed_files")
	// Every READABLE path is still described in files — including the corrupt
	// ones, whose metadata is what makes an empty file visible as "0 records"
	// rather than absent. Only a path that could not be read at all is
	// reported solely through the health channel.
	described := make(map[string]float64)
	files := response["files"].([]interface{})
	for _, raw := range files {
		entry := raw.(map[string]interface{})
		described[entry["path"].(string)] = entry["line_count"].(float64)
	}
	assert.Equal(t, float64(2), described[good], "the usable file must be inspected at full length")
	assert.Contains(t, described, empty, "a readable but empty file must still be described")
	assert.Contains(t, described, truncated)
	assert.NotContains(t, described, missing, "an unreadable path has no metadata to describe")

	// Per-file health is structured, including the parse-error count for the
	// file whose records did not parse.
	health := fileHealthOf(t, response)
	require.Contains(t, health, truncated)
	assert.Greater(t, health[truncated]["parse_errors"].(float64), float64(0),
		"a truncated line must be reported as a parse error, not just as an unusable file")
	assert.Equal(t, false, health[truncated]["parseable"])
	assert.Equal(t, true, health[good]["parseable"])
	assert.Equal(t, float64(2), health[good]["entries"])
	// The unreadable path carries the underlying error text, so the caller can
	// tell "the file is gone" from "the file is corrupt".
	assert.NotEmpty(t, health[missing]["error"])
}

func fileHealthOf(t *testing.T, response map[string]interface{}) map[string]map[string]interface{} {
	t.Helper()
	raw, present := response["file_health"]
	require.True(t, present, "the response must carry file_health; got keys %v", keysOf(response))
	entries, ok := raw.([]interface{})
	require.True(t, ok, "file_health must be a list; got %T", raw)

	byPath := make(map[string]map[string]interface{}, len(entries))
	for _, entry := range entries {
		object, ok := entry.(map[string]interface{})
		require.True(t, ok, "each file_health entry must be an object; got %v", entry)
		path, ok := object["path"].(string)
		require.True(t, ok, "each file_health entry must name its path; got %v", entry)
		byPath[path] = object
	}
	return byPath
}

// AC: a healthy corpus produces an empty malformed_files list and otherwise
// unchanged output — the pre-existing per-file shape is exactly what the
// inspector produced before, with the health channel added alongside it.
func TestHandleInspectSessionFiles_HealthyCorpusIsUnchanged(t *testing.T) {
	dir := t.TempDir()
	good := writeCorpusFile(t, dir, "s1.jsonl",
		`{"type":"user","timestamp":"2026-07-30T10:00:00Z"}`+"\n"+
			`{"type":"assistant","timestamp":"2026-07-30T10:00:01Z"}`+"\n")

	result, err := HandleInspectSessionFiles(context.Background(), map[string]interface{}{
		"files": []interface{}{good},
	})
	require.NoError(t, err)

	response := decodeResponse(t, result)
	assert.Empty(t, malformedFilesOf(t, response))

	assert.Equal(t,
		[]string{"file_health", "files", "malformed_files", "summary"},
		keysOf(response),
		"malformed_files and file_health must be the ONLY additions to the response shape")

	summary := response["summary"].(map[string]interface{})
	assert.Equal(t, float64(1), summary["total_files"])
	assert.Equal(t, float64(2), summary["total_records"])

	files := response["files"].([]interface{})
	require.Len(t, files, 1)
	entry := files[0].(map[string]interface{})
	assert.Equal(t, good, entry["path"])
	assert.Equal(t, float64(2), entry["line_count"])
	assert.Equal(t, map[string]interface{}{"user": float64(1), "assistant": float64(1)}, entry["record_types"])

	health := fileHealthOf(t, response)
	require.Contains(t, health, good)
	assert.Equal(t, true, health[good]["parseable"])
	assert.Equal(t, false, health[good]["empty"])
	// A healthy file omits the failure fields entirely, so the per-file channel
	// stays small enough to report for a whole corpus in one response.
	assert.NotContains(t, health[good], "parse_errors")
	assert.NotContains(t, health[good], "reason")
	assert.NotContains(t, health[good], "error")
}

// The discovery tools must answer "is this corpus safe to query?" for a whole
// corpus in one call — the point of reporting health in the LISTING tool rather
// than leaving it to a query. A sample file path plus a corrupt sibling is
// exactly the two-stage workflow's first step.
func TestHandleGetSessionDirectory_HealthIsVisibleWithoutAnyQuery(t *testing.T) {
	_, sessionDir, projectPath := setupClaudeSessionDir(t)
	installSharedCorruptShapes(t, sessionDir)

	result, err := HandleGetSessionDirectory(context.Background(), map[string]interface{}{
		"scope": "project", "working_dir": projectPath,
	})
	require.NoError(t, err)

	encoded, err := json.Marshal(result)
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(encoded), "8eda8f4e-2c74-4176-ba6b-8c45e890df42"),
		"the corrupt transcript must be named in the directory listing itself: %s", encoded)
}
