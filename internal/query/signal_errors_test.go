package query

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/analyzer"
)

// DIR-097: query_session_signals(type="errors") must expose the same five
// fields on every record, whatever irregular shape the underlying source
// record has. These tests pin the projection against the two toolUseResult
// variants and the error-text sources found in this project's corpus.

// pairedError builds one entry of the shape ErrorSignalJoinJQ emits for a
// failing tool_result, plus the __tool_use index entry that carries the tool
// name found on the issuing assistant record.
func pairedError(toolUseID, toolName string, record, block map[string]interface{}) []interface{} {
	return []interface{}{
		map[string]interface{}{joinToolUseKey: map[string]interface{}{"id": toolUseID, "name": toolName}},
		map[string]interface{}{joinRecordKey: record, joinBlockKey: block},
	}
}

// realCorpusErrorRecord is a verbatim record from this project's own Claude
// Code session corpus (session 96b3f998, 2026-09-14): the failing
// `git rev-parse` tool_result whose toolUseResult is a bare string. Only cwd
// and sessionId are rewritten so the fixture is stable across checkouts.
func realCorpusErrorRecord() (record, block map[string]interface{}) {
	return map[string]interface{}{
			"parentUuid":    "c0dc6dff-1399-48b0-b0f2-fc8bf3deb6f8",
			"isSidechain":   false,
			"type":          "user",
			"message":       map[string]interface{}{"role": "user"},
			"uuid":          "9f853870-0992-4ac8-8845-d55763e1db05",
			"timestamp":     "2026-09-14T17:54:09.847Z",
			"toolUseResult": "Error: Exit code 128\nfatal: Needed a single revision",
			"sessionId":     "96b3f998-f618-4705-aa96-0f6e8c2c3494",
			"version":       "2.1.261",
		}, map[string]interface{}{
			"type":        "tool_result",
			"content":     "Exit code 128\nfatal: Needed a single revision",
			"is_error":    true,
			"tool_use_id": "call_00_xS8pYGo8SWC2v3Tzl4Jp4487",
		}
}

// objectVariantErrorRecord is the other toolUseResult shape: the Bash harness
// writes an object (stdout/stderr/returnCodeInterpretation) and leaves the
// block's content empty, so the failure text exists ONLY inside the object.
func objectVariantErrorRecord() (record, block map[string]interface{}) {
	return map[string]interface{}{
			"type":      "user",
			"uuid":      "object-variant-u1",
			"timestamp": "2026-09-14T18:00:00.000Z",
			"sessionId": "object-variant-session",
			"toolUseResult": map[string]interface{}{
				"stdout":                   "",
				"stderr":                   "bash: definitely-not-a-command: command not found",
				"interrupted":              false,
				"isImage":                  false,
				"returnCodeInterpretation": "command not found",
				"noOutputExpected":         false,
			},
			"message": map[string]interface{}{"role": "user"},
		}, map[string]interface{}{
			"type":        "tool_result",
			"content":     "",
			"is_error":    true,
			"tool_use_id": "object-variant-toolu",
		}
}

func TestProjectErrorSignals_ExposesFiveFieldsForBothToolUseResultVariants(t *testing.T) {
	cases := []struct {
		name          string
		fixture       func() (map[string]interface{}, map[string]interface{})
		toolUseID     string
		toolName      string
		wantErrorText string
		wantCategory  string
		wantSessionID string
		wantTimestamp string
	}{
		{
			name:          "string variant (this project's corpus)",
			fixture:       realCorpusErrorRecord,
			toolUseID:     "call_00_xS8pYGo8SWC2v3Tzl4Jp4487",
			toolName:      "Bash",
			wantErrorText: "Exit code 128\nfatal: Needed a single revision",
			wantCategory:  "bash_exit_code",
			wantSessionID: "96b3f998-f618-4705-aa96-0f6e8c2c3494",
			wantTimestamp: "2026-09-14T17:54:09.847Z",
		},
		{
			name:          "object variant (stderr-only failure text)",
			fixture:       objectVariantErrorRecord,
			toolUseID:     "object-variant-toolu",
			toolName:      "Bash",
			wantErrorText: "bash: definitely-not-a-command: command not found",
			wantCategory:  "command_not_found",
			wantSessionID: "object-variant-session",
			wantTimestamp: "2026-09-14T18:00:00.000Z",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			record, block := tc.fixture()

			projected := ProjectErrorSignals(pairedError(tc.toolUseID, tc.toolName, record, block))
			require.Len(t, projected, 1, "the __tool_use index entry must not become an output record")

			rec, ok := projected[0].(map[string]interface{})
			require.True(t, ok, "unexpected projected shape: %#v", projected[0])

			// AC1: all five fields present on the record itself — not merely
			// reachable through a nested path.
			for _, field := range ErrorSignalFields {
				require.Contains(t, rec, field, "projected record is missing %q: %#v", field, rec)
			}
			require.Len(t, rec, len(ErrorSignalFields),
				"the projection is a fixed contract, not the raw record with fields added: %#v", rec)

			require.Equal(t, tc.wantTimestamp, rec["timestamp"])
			require.Equal(t, tc.wantSessionID, rec["session_id"])
			require.Equal(t, tc.toolName, rec["tool_name"])
			require.Equal(t, tc.wantErrorText, rec["error_text"])
			require.Equal(t, tc.wantCategory, rec["category"])
		})
	}
}

// TestProjectErrorSignals_CategoryAgreesWithAnalyzer is AC2: the projected
// category is produced by the very function analyze_errors groups by, so the
// two tools cannot drift into different vocabularies.
func TestProjectErrorSignals_CategoryAgreesWithAnalyzer(t *testing.T) {
	cases := []struct {
		toolName, errorText, wantCategory string
	}{
		{"Bash", "Exit code 128\nfatal: Needed a single revision", "bash_exit_code"},
		{"Bash", "bash: nope: command not found", "command_not_found"},
		{"Bash", "Exit code 143\nCommand timed out after 120s", "command_timeout"},
		{"Bash", "cat: /tmp/nope: No such file or directory", "file_not_found"},
		// A real corpus error text that ClassifyErrorType deliberately leaves
		// uncategorized ("File does not exist" matches none of the rules).
		// The projection inherits that verdict rather than inventing a label,
		// which is the point of sharing the classifier.
		{"Read", "File does not exist. Note: your current working directory is /tmp", "uncategorized"},
		{"Bash", "unknown gate: ready-pool-check", "uncategorized"},
		{"Bash", "", "tool_error_no_message"},
	}

	for _, tc := range cases {
		t.Run(tc.wantCategory, func(t *testing.T) {
			projected := ProjectErrorSignals(pairedError("toolu_x", tc.toolName,
				map[string]interface{}{"timestamp": "t", "sessionId": "s"},
				map[string]interface{}{"tool_use_id": "toolu_x", "content": tc.errorText}))
			require.Len(t, projected, 1)

			rec := projected[0].(map[string]interface{})
			require.Equal(t, tc.wantCategory, rec["category"])
			// The label is exactly what analyze_errors would assign to the
			// same (tool, text) pair it parses out of the same record.
			require.Equal(t, analyzer.ClassifyErrorType(tc.toolName, tc.errorText), rec["category"])
		})
	}
}

// TestProjectErrorSignals_ErrorTextPrecedence pins the source order that keeps
// the projection's text identical to what the parser hands analyze_errors: an
// explicit block-level "error" wins, then the block's "content", and only then
// the record-level "toolUseResult".
func TestProjectErrorSignals_ErrorTextPrecedence(t *testing.T) {
	base := func() (map[string]interface{}, map[string]interface{}) {
		return map[string]interface{}{
			"timestamp":     "2026-01-01T00:00:00Z",
			"session_id":    "sess",
			"toolUseResult": "Error: from toolUseResult",
		}, map[string]interface{}{"tool_use_id": "toolu_p", "is_error": true}
	}

	t.Run("explicit error field wins", func(t *testing.T) {
		record, block := base()
		block["error"] = "from error field"
		block["content"] = "from content"
		require.Equal(t, "from error field", ErrorText(record, block))
	})

	t.Run("content wins over toolUseResult", func(t *testing.T) {
		record, block := base()
		block["content"] = "from content"
		require.Equal(t, "from content", ErrorText(record, block))
	})

	t.Run("toolUseResult is the last resort", func(t *testing.T) {
		record, block := base()
		require.Equal(t, "Error: from toolUseResult", ErrorText(record, block))
	})

	t.Run("array-shaped content is joined", func(t *testing.T) {
		record, block := base()
		block["content"] = []interface{}{
			map[string]interface{}{"type": "text", "text": "first line"},
			map[string]interface{}{"type": "text", "text": "second line"},
		}
		require.Equal(t, "first line\nsecond line", ErrorText(record, block))
	})

	t.Run("object toolUseResult with no text fields yields empty", func(t *testing.T) {
		// An unrecognized object must not be serialized wholesale: Edit
		// results carry whole file bodies, which would make the projection's
		// size unbounded.
		require.Equal(t, "", toolUseResultText(map[string]interface{}{
			"type": "file",
			"file": map[string]interface{}{"content": "a very large file body"},
		}))
		require.Equal(t, "", toolUseResultText(map[string]interface{}{"success": true}))
		require.Equal(t, "", toolUseResultText(42))
	})
}

// TestProjectErrorSignal_FallsBackToBlockLevelToolName covers a record that
// carries the tool name on the tool_result block itself (the shape
// cmd/mcp-server/test.jsonl uses), where no assistant record is available to
// join against.
func TestProjectErrorSignal_FallsBackToBlockLevelToolName(t *testing.T) {
	projected := ProjectErrorSignal(
		map[string]interface{}{"timestamp": "t", "sessionId": "s"},
		map[string]interface{}{"tool_use_id": "missing-from-index", "name": "Bash", "content": "boom"},
		map[string]string{},
	)
	require.Equal(t, "Bash", projected["tool_name"])
	require.Equal(t, "uncategorized", projected["category"])
}

// TestProjectErrorSignals_OneRecordPerFailingBlock keeps a record with several
// failing tool_result blocks from collapsing into a single output record, and
// keeps its successful blocks out of the output entirely.
func TestProjectErrorSignals_OneRecordPerFailingBlock(t *testing.T) {
	record := map[string]interface{}{"timestamp": "t", "sessionId": "s"}

	entries := []interface{}{
		map[string]interface{}{joinToolUseKey: map[string]interface{}{"id": "ok", "name": "Read"}},
		map[string]interface{}{joinToolUseKey: map[string]interface{}{"id": "e1", "name": "Bash"}},
		map[string]interface{}{joinToolUseKey: map[string]interface{}{"id": "e2", "name": "Edit"}},
		// Only the failing blocks are paired in; the successful block never
		// reaches the projection because ErrorSignalJoinJQ filters on is_error.
		map[string]interface{}{joinRecordKey: record, joinBlockKey: map[string]interface{}{"tool_use_id": "e1", "content": "first failure"}},
		map[string]interface{}{joinRecordKey: record, joinBlockKey: map[string]interface{}{"tool_use_id": "e2", "content": "second failure"}},
	}

	projected := ProjectErrorSignals(entries)
	require.Len(t, projected, 2)

	require.Equal(t, "first failure", projected[0].(map[string]interface{})["error_text"])
	require.Equal(t, "Bash", projected[0].(map[string]interface{})["tool_name"])
	require.Equal(t, "second failure", projected[1].(map[string]interface{})["error_text"])
	require.Equal(t, "Edit", projected[1].(map[string]interface{})["tool_name"])
}

// TestProjectErrorSignals_ToleratesMalformedEntries keeps a defensive posture:
// anything that is not the paired shape is skipped rather than panicking or
// emitting a half-built record.
func TestProjectErrorSignals_ToleratesMalformedEntries(t *testing.T) {
	projected := ProjectErrorSignals([]interface{}{
		"not a map",
		map[string]interface{}{joinRecordKey: "not a map either"},
		map[string]interface{}{"unrelated": true},
	})
	require.Empty(t, projected)
}

// TestSessionIDOf_ToleratesBothSpellings pins the raw (sessionId) and
// normalized (session_id) session identifier spellings.
func TestSessionIDOf_ToleratesBothSpellings(t *testing.T) {
	require.Equal(t, "raw", SessionIDOf(map[string]interface{}{"sessionId": "raw"}))
	require.Equal(t, "norm", SessionIDOf(map[string]interface{}{"session_id": "norm"}))
	require.Equal(t, "norm", SessionIDOf(map[string]interface{}{"session_id": "norm", "sessionId": "raw"}))
	require.Equal(t, "", SessionIDOf(map[string]interface{}{}))
}
