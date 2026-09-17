package executor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/analyzer"
	"github.com/yaleh/meta-cc/internal/config"
)

// error_projection_test.go covers DIR-097: query_session_signals(type="errors")
// must project every record to the stable {timestamp, session_id, tool_name,
// error_text, category} shape regardless of the underlying toolUseResult
// string/object variance and of whether the tool_result block's `content` is
// a string or an array of text blocks.

// rawErrorRecord builds a raw user-role error record the way the Claude JSONL
// path delivers it: an `is_error: true` tool_result block inside
// message.content, plus the record-level toolUseResult whose type varies.
func rawErrorRecord(timestamp, sessionID string, block map[string]interface{}, toolUseResult interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type":          "user",
		"timestamp":     timestamp,
		"sessionId":     sessionID,
		"toolUseResult": toolUseResult,
		"message": map[string]interface{}{
			"role":    "user",
			"content": []interface{}{block},
		},
	}
}

func errorBlock(toolUseID string, content interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type":        "tool_result",
		"tool_use_id": toolUseID,
		"is_error":    true,
		"content":     content,
	}
}

// TestProjectErrorRecord_StringVariantAndObjectVariant is the DIR-097 core
// assertion: the two toolUseResult variants project to the same five-field
// shape, with the error text taken from the nested message.content[].content
// (not from toolUseResult).
func TestProjectErrorRecord_StringVariantAndObjectVariant(t *testing.T) {
	const sid = "sess-1"
	toolNames := map[string]string{toolUseKey(sid, "toolu_1"): "Bash"}

	cases := []struct {
		name          string
		toolUseResult interface{}
	}{
		{"string toolUseResult", "Error: Exit code 1\nboom"},
		{"object toolUseResult", map[string]interface{}{
			"stdout": "", "stderr": "Exit code 1\nboom", "interrupted": false,
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := projectErrorRecord(
				rawErrorRecord("2026-01-01T00:00:02Z", sid, errorBlock("toolu_1", "Exit code 1\nboom"), tc.toolUseResult),
				toolNames,
			)
			require.Equal(t, map[string]interface{}{
				"timestamp":  "2026-01-01T00:00:02Z",
				"session_id": sid,
				"tool_name":  "Bash",
				"error_text": "Exit code 1\nboom",
				"category":   "bash_exit_code",
			}, got)
		})
	}
}

// TestProjectErrorRecord_ArrayContentJoinsTextBlocks pins the array-content
// variant: message.content[].content as an array of text blocks must yield the
// same newline-joined text internal/types.ToolResult.UnmarshalJSON produces,
// so error_text (and therefore category) matches what analyze_errors sees.
func TestProjectErrorRecord_ArrayContentJoinsTextBlocks(t *testing.T) {
	content := []interface{}{
		map[string]interface{}{"type": "text", "text": "File not found: /tmp/x"},
		map[string]interface{}{"type": "text", "text": ""},
		map[string]interface{}{"type": "text", "text": "hint: check the path"},
	}
	got := projectErrorRecord(
		rawErrorRecord("2026-01-01T00:00:03Z", "sess-2", errorBlock("toolu_2", content), map[string]interface{}{"type": "text"}),
		map[string]string{toolUseKey("sess-2", "toolu_2"): "Read"},
	)
	require.Equal(t, "File not found: /tmp/x\nhint: check the path", got["error_text"])
	require.Equal(t, "file_not_found", got["category"])
	require.Equal(t, "Read", got["tool_name"])
}

// TestProjectErrorRecord_FallsBackToToolUseResult covers the defensive path:
// when no tool_result block yields text, the record-level toolUseResult (either
// variant) is the last resort rather than emitting an empty error_text and a
// misleading tool_error_no_message label.
func TestProjectErrorRecord_FallsBackToToolUseResult(t *testing.T) {
	cases := []struct {
		name          string
		toolUseResult interface{}
		wantText      string
		wantCategory  string
	}{
		{"object variant uses stderr", map[string]interface{}{"stdout": "out", "stderr": "command not found: npm"}, "command not found: npm", "command_not_found"},
		{"object variant falls through to stdout", map[string]interface{}{"stderr": "", "stdout": "permission denied"}, "permission denied", "permission_denied"},
		{"string variant", "Error: connection refused", "Error: connection refused", "connection_error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := projectErrorRecord(
				rawErrorRecord("2026-01-01T00:00:04Z", "sess-3", errorBlock("toolu_3", ""), tc.toolUseResult),
				map[string]string{toolUseKey("sess-3", "toolu_3"): "Bash"},
			)
			require.Equal(t, tc.wantText, got["error_text"])
			require.Equal(t, tc.wantCategory, got["category"])
		})
	}
}

// TestProjectErrorRecord_AlwaysExposesAllFiveFields is the AC1 contract in its
// purest form: whatever the input record lacks, the projected record carries
// exactly the five documented keys — never a missing key a consumer's jq would
// have to guard for.
func TestProjectErrorRecord_AlwaysExposesAllFiveFields(t *testing.T) {
	cases := map[string]map[string]interface{}{
		"empty record":     {},
		"no matching tool": rawErrorRecord("2026-01-01T00:00:05Z", "sess-4", errorBlock("unknown-id", "boom"), nil),
		"session_id only": map[string]interface{}{
			"type": "user", "session_id": "sess-5",
			"message": map[string]interface{}{
				"content": []interface{}{errorBlock("toolu_5", "boom")},
			},
		},
	}
	for name, rec := range cases {
		t.Run(name, func(t *testing.T) {
			got := projectErrorRecord(rec, nil)
			require.Len(t, got, len(projectedErrorFields))
			for _, field := range projectedErrorFields {
				require.Contains(t, got, field, "projected record must always expose %q", field)
			}
		})
	}

	// sessionId/session_id are the same identity under two names (raw Claude
	// JSONL uses sessionId; the normalized record carries both) — either one
	// populates the projected session_id.
	got := projectErrorRecord(cases["session_id only"], nil)
	require.Equal(t, "sess-5", got["session_id"])
}

// TestProjectErrorRecord_NonErrorBlocksIgnored guards the block selection: only
// `is_error: true` tool_result blocks feed error_text, so a record carrying a
// successful result alongside a failed one does not project the success text.
func TestProjectErrorRecord_NonErrorBlocksIgnored(t *testing.T) {
	rec := map[string]interface{}{
		"type": "user", "timestamp": "2026-01-01T00:00:06Z", "sessionId": "sess-6",
		"message": map[string]interface{}{"content": []interface{}{
			map[string]interface{}{"type": "tool_result", "tool_use_id": "toolu_ok", "is_error": false, "content": "all good"},
			map[string]interface{}{"type": "tool_result", "tool_use_id": "toolu_err", "is_error": true, "content": "Exit code 2"},
		}},
	}
	got := projectErrorRecord(rec, map[string]string{
		toolUseKey("sess-6", "toolu_ok"):  "Read",
		toolUseKey("sess-6", "toolu_err"): "Bash",
	})
	require.Equal(t, "Exit code 2", got["error_text"])
	require.Equal(t, "Bash", got["tool_name"], "the resolved name must come from the failing block, not the succeeding one")
}

// TestCollectToolNames covers the cross-record join index: assistant tool_use
// blocks are keyed by (session, tool_use_id) so a project-scope stream spanning
// several sessions cannot resolve one session's id to another session's tool.
func TestCollectToolNames(t *testing.T) {
	entries := []interface{}{
		map[string]interface{}{
			"type": "assistant", "sessionId": "sess-a",
			"message": map[string]interface{}{"content": []interface{}{
				map[string]interface{}{"type": "tool_use", "id": "toolu_x", "name": "Bash"},
			}},
		},
		map[string]interface{}{
			"type": "assistant", "sessionId": "sess-b",
			"message": map[string]interface{}{"content": []interface{}{
				map[string]interface{}{"type": "tool_use", "id": "toolu_x", "name": "Read"},
			}},
		},
		map[string]interface{}{"type": "user", "sessionId": "sess-a"},
	}
	names := collectToolNames(entries)
	require.Equal(t, "Bash", names[toolUseKey("sess-a", "toolu_x")])
	require.Equal(t, "Read", names[toolUseKey("sess-b", "toolu_x")])
}

// ─── end-to-end fixtures ─────────────────────────────────────────────────────

// setupErrorProjectionFixture wires a Claude projects root holding three
// sessions whose error records cover the shapes DIR-097 was filed about:
// a string toolUseResult, an object toolUseResult, and an array-valued
// tool_result content. It returns the resolved project path the sessions are
// scoped to (used as working_dir by the queries under test).
func setupErrorProjectionFixture(t *testing.T) (projectPath string) {
	t.Helper()

	projectsRoot := t.TempDir()
	t.Setenv("META_CC_PROJECTS_ROOT", projectsRoot)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CODEX_HOME", filepath.Join(t.TempDir(), "codex-home"))

	absProject, err := filepath.Abs(t.TempDir())
	require.NoError(t, err)
	resolvedProject, err := filepath.EvalSymlinks(absProject)
	require.NoError(t, err)

	hash := strings.NewReplacer("\\", "-", "/", "-", ":", "-").Replace(resolvedProject)
	sessionDir := filepath.Join(projectsRoot, hash)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))

	// Each session file is one failed tool call: a user prompt, the assistant
	// tool_use that names the tool, then the user-role tool_result that carries
	// the error. Only the last line varies across the three fixtures — that is
	// the point: the projection must yield one shape for all of them.
	writeSession := func(sessionID, toolUseID, toolName, ts, resultTS, resultRecord string) {
		lines := []string{
			`{"type":"user","timestamp":"` + ts + `","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `","message":{"role":"user","content":"go"}}`,
			`{"type":"assistant","timestamp":"` + ts + `","sessionId":"` + sessionID + `","cwd":"` + resolvedProject +
				`","message":{"role":"assistant","content":[{"type":"tool_use","id":"` + toolUseID + `","name":"` + toolName + `","input":{}}]}}`,
			`{"type":"user","timestamp":"` + resultTS + `","sessionId":"` + sessionID + `","cwd":"` + resolvedProject + `",` +
				`"toolUseResult":` + toolUseResultFor(sessionID) + `,"message":{"role":"user","content":[` + resultRecord + `]}}`,
		}
		sessionFile := filepath.Join(sessionDir, sessionID+".jsonl")
		require.NoError(t, os.WriteFile(sessionFile, []byte(strings.Join(lines, "\n")+"\n"), 0o644))
	}

	// Session 1: string toolUseResult + string content (the overwhelmingly
	// common Claude Code shape; see this project's own corpus).
	writeSession("sess-raw-string", "toolu_bash", "Bash", "2026-01-01T00:00:01Z", "2026-01-01T00:00:02Z",
		`{"type":"tool_result","tool_use_id":"toolu_bash","is_error":true,"content":"Exit code 1\nNo module named pre_commit"}`)

	// Session 2: object toolUseResult (Bash-shaped: stdout/stderr/interrupted)
	// alongside the same nested string content.
	writeSession("sess-raw-object", "toolu_read", "Read", "2026-01-01T00:01:01Z", "2026-01-01T00:01:02Z",
		`{"type":"tool_result","tool_use_id":"toolu_read","is_error":true,"content":"File not found: /tmp/missing"}`)

	// Session 3: array-valued tool_result content (text blocks), whose error
	// text is not a plain string at all.
	writeSession("sess-raw-array", "toolu_npm", "Bash", "2026-01-01T00:02:01Z", "2026-01-01T00:02:02Z",
		`{"type":"tool_result","tool_use_id":"toolu_npm","is_error":true,"content":[{"type":"text","text":"command not found: npm"},{"type":"text","text":""}]}`)

	return resolvedProject
}

// toolUseResultFor returns the record-level toolUseResult for a fixture
// session — a string for sess-raw-string (matching real Claude Code records,
// which prefix the message with "Error: "), an object for the other two.
func toolUseResultFor(sessionID string) string {
	switch sessionID {
	case "sess-raw-string":
		return `"Error: Exit code 1\nNo module named pre_commit"`
	case "sess-raw-object":
		return `{"stdout":"","stderr":"File not found: /tmp/missing","interrupted":false,"isImage":false}`
	default:
		return `{"stdout":"","stderr":"command not found: npm","interrupted":false,"isImage":false}`
	}
}

// projectedEntry asserts the entry is a projected record and returns it.
func projectedEntry(t *testing.T, entry interface{}) map[string]interface{} {
	t.Helper()
	rec, ok := entry.(map[string]interface{})
	require.True(t, ok, "expected a projected map entry, got %#v", entry)
	require.Len(t, rec, len(projectedErrorFields), "projected record must expose exactly the five documented fields: %#v", rec)
	for _, field := range projectedErrorFields {
		require.Contains(t, rec, field, "projected record is missing %q: %#v", field, rec)
	}
	return rec
}

// TestHandleQueryToolErrors_ProjectsStableShape is AC1/AC2 end-to-end over the
// three-record fixture: every returned record exposes all five fields with the
// right values, and each category is the label analyze_errors would assign to
// the same (tool_name, error_text) pair — asserted against ClassifyErrorType
// itself, so the two tools cannot drift apart unnoticed.
func TestHandleQueryToolErrors_ProjectsStableShape(t *testing.T) {
	projectPath := setupErrorProjectionFixture(t)
	e := NewToolExecutor()

	result, err := handleQueryToolErrors(e, "project", map[string]interface{}{
		"working_dir": projectPath,
		"provider":    "claude",
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 3, "all three fixture error records must be returned")

	bySession := map[string]map[string]interface{}{}
	for _, entry := range result.Entries {
		rec := projectedEntry(t, entry)
		bySession[rec["session_id"].(string)] = rec
	}

	// Session 1 — string toolUseResult, string content.
	require.Equal(t, map[string]interface{}{
		"timestamp":  "2026-01-01T00:00:02Z",
		"session_id": "sess-raw-string",
		"tool_name":  "Bash",
		"error_text": "Exit code 1\nNo module named pre_commit",
		"category":   "bash_exit_code",
	}, bySession["sess-raw-string"])

	// Session 2 — object toolUseResult; error_text still comes from the
	// nested message.content[].content, not from the object's stderr.
	require.Equal(t, map[string]interface{}{
		"timestamp":  "2026-01-01T00:01:02Z",
		"session_id": "sess-raw-object",
		"tool_name":  "Read",
		"error_text": "File not found: /tmp/missing",
		"category":   "file_not_found",
	}, bySession["sess-raw-object"])

	// Session 3 — array-valued tool_result content, joined from its text blocks.
	require.Equal(t, map[string]interface{}{
		"timestamp":  "2026-01-01T00:02:02Z",
		"session_id": "sess-raw-array",
		"tool_name":  "Bash",
		"error_text": "command not found: npm",
		"category":   "command_not_found",
	}, bySession["sess-raw-array"])

	// AC2: the labels are analyze_errors' labels, produced by the identical
	// classifier over the identical inputs.
	for _, rec := range bySession {
		require.Equal(t,
			analyzer.ClassifyErrorType(rec["tool_name"].(string), rec["error_text"].(string)),
			rec["category"],
			"category must be the analyze_errors label for the same (tool_name, error_text)")
	}
}

// TestHandleQueryToolErrors_RawFlagReturnsOriginalRecord is AC5: raw=true is
// the opt-out that preserves the pre-DIR-097 pass-through, so a consumer that
// needs a field the projection drops can still get it.
func TestHandleQueryToolErrors_RawFlagReturnsOriginalRecord(t *testing.T) {
	projectPath := setupErrorProjectionFixture(t)
	e := NewToolExecutor()

	result, err := handleQueryToolErrors(e, "project", map[string]interface{}{
		"working_dir": projectPath,
		"provider":    "claude",
		"raw":         true,
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 3)

	for _, entry := range result.Entries {
		rec, ok := entry.(map[string]interface{})
		require.True(t, ok)
		require.NotContains(t, rec, "error_text", "raw=true must not project")
		require.NotContains(t, rec, "category", "raw=true must not project")
		require.Contains(t, rec, "toolUseResult", "raw=true must keep the original record shape")
		require.Contains(t, rec, "message")
	}
}

// TestExecuteTool_QuerySessionSignals_Errors_JQFilterComposesOverProjection is
// AC4 through the real tool path: jq_filter is applied by the pipeline as a
// post-filter over the handler's result, so it must compose with the projected
// fields (not the raw record's) — including for a field that does not exist on
// the raw record at all, which is what makes the projection useful.
func TestExecuteTool_QuerySessionSignals_Errors_JQFilterComposesOverProjection(t *testing.T) {
	projectPath := setupErrorProjectionFixture(t)
	e := NewToolExecutor()
	cfg := &config.Config{}

	// jq_filter runs against the whole entries array (see
	// pipeline.applyJQPostFilter), so the documented idiom is `.[] | ...`.
	cases := []struct {
		name    string
		filter  string
		wantLen int
	}{
		{"by category", `.[] | select(.category == "command_not_found")`, 1},
		{"by tool_name", `.[] | select(.tool_name == "Bash")`, 2},
		{"project the fields", `.[] | {session_id, error_text, category}`, 3},
		{"select(false) still empties", `.[] | select(false)`, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := e.ExecuteTool(cfg, "query_session_signals", map[string]interface{}{
				"type": "errors", "provider": "claude", "working_dir": projectPath, "jq_filter": tc.filter,
			})
			require.NoError(t, err)
			require.Len(t, extractDataArray(t, out), tc.wantLen)
		})
	}

	// The projected field names are what the caller's jq sees — a filter over
	// the raw JSONL path (message.content[]) now matches nothing, because the
	// projection replaced those records.
	out, err := e.ExecuteTool(cfg, "query_session_signals", map[string]interface{}{
		"type": "errors", "provider": "claude", "working_dir": projectPath,
		"jq_filter": `.[] | select(.message.content != null)`,
	})
	require.NoError(t, err)
	require.Len(t, extractDataArray(t, out), 0)
}

