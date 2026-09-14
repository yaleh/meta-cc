package executor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/analysis"
	"github.com/yaleh/meta-cc/internal/config"
	querypkg "github.com/yaleh/meta-cc/internal/query"
)

// DIR-097 end-to-end coverage: query_session_signals(type="errors") now hands
// back a stable five-field projection instead of the raw JSONL record, while
// raw=true keeps the old shape reachable. These tests exercise the real
// handler path (the ErrorSignalJoinJQ program plus the Go projection), not the
// projection alone, so the jq/Go contract between them is covered too.

// setupErrorSignalSession writes one Claude session JSONL into a hermetic
// projects root and returns the resolved project path the session is scoped
// to. body is called with that resolved path (the records must carry it as
// their cwd) and returns the session's record lines.
func setupErrorSignalSession(t *testing.T, sessionID string, body func(projectPath string) []string) (projectPath string) {
	t.Helper()

	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	t.Setenv("HOME", t.TempDir())

	rawProjectPath := t.TempDir()
	absProject, err := filepath.Abs(rawProjectPath)
	require.NoError(t, err)
	resolvedProject, err := filepath.EvalSymlinks(absProject)
	require.NoError(t, err)

	hash := strings.NewReplacer("\\", "-", "/", "-", ":", "-").Replace(resolvedProject)
	sessionDir := filepath.Join(projectsRoot, hash)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
	require.NoError(t, os.WriteFile(sessionFile, []byte(strings.Join(body(resolvedProject), "\n")+"\n"), 0o644))

	return resolvedProject
}

// corpusShapedErrorSession reproduces this project's own corpus shape: every
// failing tool_result carries a bare-string toolUseResult ("Error: " + the
// block's content), the block carries the text in content, and the failing
// tool's name exists only on the issuing assistant record.
func corpusShapedErrorSession(sessionID string, projectPath string) []string {
	user := func(ts, extra string) string {
		return `{"type":"user","timestamp":"` + ts + `","sessionId":"` + sessionID + `","cwd":"` + projectPath + `",` + extra + `}`
	}
	assistant := func(ts, id, name string) string {
		return `{"type":"assistant","timestamp":"` + ts + `","sessionId":"` + sessionID + `","cwd":"` + projectPath +
			`","message":{"role":"assistant","content":[{"type":"tool_use","id":"` + id + `","name":"` + name + `","input":{}}]}}`
	}
	return []string{
		user("2026-01-01T00:00:00Z", `"message":{"role":"user","content":"do stuff"}`),
		assistant("2026-01-01T00:00:01Z", "toolu_str", "Bash"),
		user("2026-01-01T00:00:02Z",
			`"message":{"role":"user","content":[{"type":"tool_result","content":"Exit code 128\nfatal: Needed a single revision","is_error":true,"tool_use_id":"toolu_str"}]},`+
				`"toolUseResult":"Error: Exit code 128\nfatal: Needed a single revision"`),
		assistant("2026-01-01T00:00:03Z", "toolu_str2", "Bash"),
		user("2026-01-01T00:00:04Z",
			`"message":{"role":"user","content":[{"type":"tool_result","content":"bash: nope: command not found","is_error":true,"tool_use_id":"toolu_str2"}]},`+
				`"toolUseResult":"Error: bash: nope: command not found"`),
		// A successful call: it must never appear in type=errors output.
		assistant("2026-01-01T00:00:05Z", "toolu_ok", "Read"),
		user("2026-01-01T00:00:06Z",
			`"message":{"role":"user","content":[{"type":"tool_result","content":"file contents","is_error":false,"tool_use_id":"toolu_ok"}]},`+
				`"toolUseResult":{"type":"file","file":{"filePath":"/tmp/foo","content":"file contents"}}`),
	}
}

// objectVariantErrorSession reproduces the other toolUseResult shape: the
// block's content is empty and the failure text exists only inside the
// object-shaped toolUseResult.
func objectVariantErrorSession(sessionID string, projectPath string) []string {
	return []string{
		`{"type":"assistant","timestamp":"2026-01-01T00:00:01Z","sessionId":"` + sessionID + `","cwd":"` + projectPath +
			`","message":{"role":"assistant","content":[{"type":"tool_use","id":"toolu_obj","name":"Bash","input":{}}]}}`,
		`{"type":"user","timestamp":"2026-01-01T00:00:02Z","sessionId":"` + sessionID + `","cwd":"` + projectPath +
			`","message":{"role":"user","content":[{"type":"tool_result","content":"","is_error":true,"tool_use_id":"toolu_obj"}]},` +
			`"toolUseResult":{"stdout":"","stderr":"bash: definitely-not-a-command: command not found","interrupted":false,"isImage":false,"returnCodeInterpretation":"command not found","noOutputExpected":false}}`,
	}
}

// projectedField reads one of the five documented fields off a projected
// entry, failing the test if the field is absent.
func projectedField(t *testing.T, entry interface{}, field string) string {
	t.Helper()
	rec, ok := entry.(map[string]interface{})
	require.True(t, ok, "unexpected projected entry shape: %#v", entry)
	require.Contains(t, rec, field, "projected record is missing %q: %#v", field, rec)
	value, ok := rec[field].(string)
	require.True(t, ok, "projected field %q must be a string, got %#v", field, rec[field])
	return value
}

// TestHandleQueryToolErrors_ProjectsCorpusShapedRecords is AC1 verified
// against this project's corpus shape: every returned record exposes all five
// fields at the top level, with the tool name recovered from the issuing
// assistant record and the error text taken from the tool_result block.
func TestHandleQueryToolErrors_ProjectsCorpusShapedRecords(t *testing.T) {
	const sessionID = "error-projection-corpus"
	projectPath := setupErrorSignalSession(t, sessionID, func(p string) []string {
		return corpusShapedErrorSession(sessionID, p)
	})

	e := NewToolExecutor()
	result, err := handleQueryToolErrors(e, "project", map[string]interface{}{"working_dir": projectPath})
	require.NoError(t, err)
	require.Len(t, result.Entries, 2, "the successful Read call must not be projected as an error")

	byToolUse := map[string]map[string]string{}
	for _, entry := range result.Entries {
		rec := entry.(map[string]interface{})
		// AC1: exactly the five documented fields, all present, all strings.
		require.Len(t, rec, len(querypkg.ErrorSignalFields), "unexpected field set: %#v", rec)
		fields := map[string]string{}
		for _, field := range querypkg.ErrorSignalFields {
			fields[field] = projectedField(t, entry, field)
		}
		require.Equal(t, sessionID, fields["session_id"])
		require.NotEmpty(t, fields["timestamp"])
		byToolUse[fields["error_text"]] = fields
	}

	gitErr, ok := byToolUse["Exit code 128\nfatal: Needed a single revision"]
	require.True(t, ok, "the git failure must be projected, got: %#v", byToolUse)
	require.Equal(t, "Bash", gitErr["tool_name"])
	require.Equal(t, "bash_exit_code", gitErr["category"])
	require.Equal(t, "2026-01-01T00:00:02Z", normalizeFixtureTime(gitErr["timestamp"]))

	nf, ok := byToolUse["bash: nope: command not found"]
	require.True(t, ok, "the command_not_found failure must be projected, got: %#v", byToolUse)
	require.Equal(t, "Bash", nf["tool_name"])
	require.Equal(t, "command_not_found", nf["category"])
}

// normalizeFixtureTime tolerates both the millisecond-precision and
// second-precision RFC3339 spellings a normalizer may emit for the same
// fixture timestamp.
func normalizeFixtureTime(ts string) string {
	return strings.Replace(ts, ".000Z", "Z", 1)
}

// TestHandleQueryToolErrors_ProjectsObjectVariantToolUseResult covers the
// second toolUseResult variant end-to-end: with the block's content empty, the
// failure text can only come from the object, and the record must still expose
// the full five-field shape.
func TestHandleQueryToolErrors_ProjectsObjectVariantToolUseResult(t *testing.T) {
	const sessionID = "error-projection-object"
	projectPath := setupErrorSignalSession(t, sessionID, func(p string) []string {
		return objectVariantErrorSession(sessionID, p)
	})

	e := NewToolExecutor()
	result, err := handleQueryToolErrors(e, "project", map[string]interface{}{"working_dir": projectPath})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1)

	rec := result.Entries[0].(map[string]interface{})
	require.Len(t, rec, len(querypkg.ErrorSignalFields), "object-variant records expose the same five fields: %#v", rec)
	require.Equal(t, "bash: definitely-not-a-command: command not found", projectedField(t, result.Entries[0], "error_text"))
	require.Equal(t, "Bash", projectedField(t, result.Entries[0], "tool_name"))
	require.Equal(t, "command_not_found", projectedField(t, result.Entries[0], "category"))
	require.Equal(t, sessionID, projectedField(t, result.Entries[0], "session_id"))
	require.NotEmpty(t, projectedField(t, result.Entries[0], "timestamp"))
}

// TestHandleQueryToolErrors_CategoriesMatchAnalyzeErrors is AC2 end-to-end on
// the shared corpus-shaped fixture: analyze_errors and
// query_session_signals(type=errors) must partition the same failing calls
// into the same labels, so a consumer can move between the two tools without
// remapping categories.
func TestHandleQueryToolErrors_CategoriesMatchAnalyzeErrors(t *testing.T) {
	const sessionID = "error-projection-corpus"
	projectPath := setupErrorSignalSession(t, sessionID, func(p string) []string {
		return corpusShapedErrorSession(sessionID, p)
	})

	e := NewToolExecutor()
	result, err := handleQueryToolErrors(e, "project", map[string]interface{}{"working_dir": projectPath})
	require.NoError(t, err)

	projectedLabels := map[string]int{}
	for _, entry := range result.Entries {
		projectedLabels[projectedField(t, entry, "category")]++
	}

	raw, err := analysis.New().AnalyzeErrors(map[string]interface{}{"working_dir": projectPath})
	require.NoError(t, err)

	var analysisResult struct {
		TotalErrors int `json:"total_errors"`
		ByType      []struct {
			Label string `json:"label"`
			Count int    `json:"count"`
		} `json:"by_type"`
	}
	require.NoError(t, json.Unmarshal([]byte(raw), &analysisResult))

	analyzedLabels := map[string]int{}
	for _, g := range analysisResult.ByType {
		analyzedLabels[g.Label] = g.Count
	}

	require.Equal(t, analysisResult.TotalErrors, len(result.Entries),
		"both tools must see the same failing calls")
	require.Equal(t, analyzedLabels, projectedLabels,
		"the two tools must agree on every error's category label")
}

// TestHandleQueryToolErrors_RawFlagReturnsOriginalRecord is AC5: raw=true
// keeps the pre-DIR-097 record shape reachable, so a consumer that genuinely
// needs the untouched JSONL is not locked out by the projection.
func TestHandleQueryToolErrors_RawFlagReturnsOriginalRecord(t *testing.T) {
	const sessionID = "error-projection-corpus"
	projectPath := setupErrorSignalSession(t, sessionID, func(p string) []string {
		return corpusShapedErrorSession(sessionID, p)
	})

	e := NewToolExecutor()
	result, err := handleQueryToolErrors(e, "project", map[string]interface{}{
		"working_dir": projectPath,
		"raw":         true,
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 2, "raw mode returns the same failing calls, unprojected")

	for _, entry := range result.Entries {
		rec, ok := entry.(map[string]interface{})
		require.True(t, ok, "unexpected raw entry shape: %#v", entry)
		// The original record: type/message/toolUseResult and the camelCase
		// session identifier, with none of the five projected fields added.
		require.Equal(t, "user", rec["type"])
		require.Contains(t, rec, "message")
		require.Contains(t, rec, "toolUseResult")
		require.Equal(t, sessionID, rec["sessionId"])
		// None of the projection-only fields appear. (timestamp is the raw
		// record's own field, so it is deliberately not in this list.)
		for _, field := range []string{"session_id", "tool_name", "error_text", "category"} {
			require.NotContains(t, rec, field, "raw mode must not project: %#v", rec)
		}
	}
}

// TestHandleQueryToolErrors_LimitAppliesToProjectedRecords pins that the
// caller's limit bounds projected records rather than the raw join, so a
// record whose tool_result block is accompanied by its assistant tool_use
// cannot silently halve the result count.
func TestHandleQueryToolErrors_LimitAppliesToProjectedRecords(t *testing.T) {
	const sessionID = "error-projection-corpus"
	projectPath := setupErrorSignalSession(t, sessionID, func(p string) []string {
		return corpusShapedErrorSession(sessionID, p)
	})

	e := NewToolExecutor()
	result, err := handleQueryToolErrors(e, "project", map[string]interface{}{
		"working_dir": projectPath,
		"limit":       float64(1),
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1)

	// And with no limit, both are returned — proving the limit, not the join,
	// is what truncated.
	unlimited, err := handleQueryToolErrors(e, "project", map[string]interface{}{"working_dir": projectPath})
	require.NoError(t, err)
	require.Len(t, unlimited.Entries, 2)
}

// TestExecuteTool_QuerySessionSignals_JQFilterComposesOverProjectedRecords is
// AC4 through the real ExecuteTool path: the caller-supplied jq_filter is a
// post-filter over the projected records, so it can select on any of the five
// projected fields (notably .category, which the raw record never had).
func TestExecuteTool_QuerySessionSignals_JQFilterComposesOverProjectedRecords(t *testing.T) {
	const sessionID = "error-projection-corpus"
	projectPath := setupErrorSignalSession(t, sessionID, func(p string) []string {
		return corpusShapedErrorSession(sessionID, p)
	})

	e := NewToolExecutor()
	cfg := &config.Config{}
	base := map[string]interface{}{"type": "errors", "working_dir": projectPath}

	// extractDataArray consumes a file_ref response's temp file, so each
	// output is extracted exactly once.
	out, err := e.ExecuteTool(cfg, "query_session_signals", base)
	require.NoError(t, err)
	require.Len(t, extractDataArray(t, out), 2, "baseline returns both projected errors")

	byCategory := map[string]interface{}{
		"type": "errors", "working_dir": projectPath,
		"jq_filter": `.[] | select(.category == "bash_exit_code")`,
	}
	out, err = e.ExecuteTool(cfg, "query_session_signals", byCategory)
	require.NoError(t, err)
	data := extractDataArray(t, out)
	require.Len(t, data, 1, "jq_filter must select on the projected .category field")
	require.Equal(t, "bash_exit_code", data[0]["category"])
	require.Equal(t, "Bash", data[0]["tool_name"])
	require.Equal(t, "Exit code 128\nfatal: Needed a single revision", data[0]["error_text"])

	byTool := map[string]interface{}{
		"type": "errors", "working_dir": projectPath,
		"jq_filter": `.[] | select(.tool_name == "Bash" and (.error_text | test("command not found")))`,
	}
	out, err = e.ExecuteTool(cfg, "query_session_signals", byTool)
	require.NoError(t, err)
	require.Len(t, extractDataArray(t, out), 1, "jq_filter must compose over the projected .error_text field")

	empty := map[string]interface{}{
		"type": "errors", "working_dir": projectPath,
		"jq_filter": `.[] | select(.category == "no_such_category")`,
	}
	out, err = e.ExecuteTool(cfg, "query_session_signals", empty)
	require.NoError(t, err)
	require.Empty(t, extractDataArray(t, out))

	// The canonical histogram recipe documented in
	// docs/guides/mcp-query-tools.md and docs/guides/two-stage-query-guide.md,
	// run through the real tool so a documented example cannot rot. It
	// aggregates over the whole array (group_by needs every record at once,
	// hence no leading .[]) and then unpacks with a trailing .[]: the pipeline
	// turns each jq output LINE into one entry, so without the trailing .[]
	// the entire histogram would arrive as a single entry holding an array.
	histogram := map[string]interface{}{
		"type": "errors", "working_dir": projectPath,
		"jq_filter": `group_by(.category) | map({category: .[0].category, count: length}) | .[]`,
	}
	out, err = e.ExecuteTool(cfg, "query_session_signals", histogram)
	require.NoError(t, err)
	data = extractDataArray(t, out)
	require.Len(t, data, 2, "one bucket per distinct category: %#v", data)
	require.Equal(t, map[string]interface{}{"category": "bash_exit_code", "count": float64(1)}, data[0])
	require.Equal(t, map[string]interface{}{"category": "command_not_found", "count": float64(1)}, data[1])
}

// TestExecuteTool_QuerySessionSignals_RawFlagAcceptedBySchema proves the
// opt-in flag travels the whole way from the MCP argument surface: an unknown
// key is rejected by schema validation, so this also pins that "raw" is
// declared on query_session_signals rather than merely read off the args map.
func TestExecuteTool_QuerySessionSignals_RawFlagAcceptedBySchema(t *testing.T) {
	const sessionID = "error-projection-corpus"
	projectPath := setupErrorSignalSession(t, sessionID, func(p string) []string {
		return corpusShapedErrorSession(sessionID, p)
	})

	e := NewToolExecutor()
	cfg := &config.Config{}

	out, err := e.ExecuteTool(cfg, "query_session_signals", map[string]interface{}{
		"type": "errors", "working_dir": projectPath, "raw": true,
	})
	require.NoError(t, err)

	data := extractDataArray(t, out)
	require.Len(t, data, 2)
	require.Contains(t, data[0], "toolUseResult", "raw=true must return the unprojected record")
	require.NotContains(t, data[0], "error_text")
}
