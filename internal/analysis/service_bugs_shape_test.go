package analysis_test

// DIR-096 regression tests: analyze_bugs gained a first-class pattern-count
// cap (max_patterns) and its per-pattern `examples` entries were promoted from
// bare error strings to structured objects.
//
// Before this fix, `limit` (examples per pattern) was the only size knob, and
// it could not bound the response at all: the pattern count itself was
// unbounded (56 patterns on this project's 5-day corpus) and each example was
// the raw, untruncated error text (largest observed: 10,148 characters). A
// default call produced 70,248 bytes -- well past the 32KB inline threshold --
// and each example was a bare string with no session/timestamp structure.
//
// The tests below pin, against the live Service (not a stub):
//   - max_patterns caps the returned pattern count (and 0 means unlimited);
//   - every example is an object carrying session_id + timestamp + error_text
//     (+ signature, + fix_text when the error->success pair was found);
//   - a default (no-args) call fits the inline threshold even though the
//     analyzer's unshaped result for the same corpus does not -- proving the
//     bound comes from this task's shaping, not from a small seeded corpus;
//   - stats_only materializes no examples at all.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/analysis"
	"github.com/yaleh/meta-cc/internal/analyzer"
	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/types"
)

// inlineThresholdBytes mirrors config's default Output.InlineThreshold
// (internal/config/config.go): a response at or above it is spilled to
// file_ref mode instead of being returned inline. Restated here (rather than
// imported) because these tests assert the *contract* the MCP layer applies,
// not the config getter.
const inlineThresholdBytes = 32768

// oversizedBugCorpus builds a session whose tool calls encode `distinct`
// distinct bug patterns, each observed `recurrences` times as a Bash
// error->success pair (the analyzer's 3-position lookahead pairs them), with
// every error carrying a multi-kilobyte message. That mirrors the live
// blowup: few knobs, one oversized example per pattern, and a pattern count
// far above any sane default.
func oversizedBugCorpus(distinct, recurrences int) []types.SessionEntry {
	var entries []types.SessionEntry
	for p := 0; p < distinct; p++ {
		for r := 0; r < recurrences; r++ {
			errID := fmt.Sprintf("tu-err-%d-%d", p, r)
			okID := fmt.Sprintf("tu-ok-%d-%d", p, r)
			ts := fmt.Sprintf("2025-10-02T10:%02d:%02d.000Z", p%60, r%60)
			errText := fmt.Sprintf("boom pattern-%d: %s", p, strings.Repeat("e", 3000))
			entries = append(entries,
				toolUseEntry("u-"+errID, ts, errID, "Bash", map[string]interface{}{"command": "boom"}),
				toolResultEntry("r-"+errID, ts, errID, errText, "error", errText),
				toolUseEntry("u-"+okID, ts, okID, "Bash", map[string]interface{}{"command": "fixed"}),
				toolResultEntry("r-"+okID, ts, okID, "fixed output", "success", ""),
			)
		}
	}
	return entries
}

// withSessionID stamps a session id on every entry, so the structured
// examples have a session to attribute each occurrence to (the JSONL helper
// writes whatever SessionID the entry carries).
func withSessionID(entries []types.SessionEntry, sessionID string) []types.SessionEntry {
	for i := range entries {
		entries[i].SessionID = sessionID
	}
	return entries
}

type bugPatternShape struct {
	ErrorSignature string `json:"error_signature"`
	FixCount       int    `json:"fix_count"`
	Recurrences    int    `json:"recurrences"`
	UnfixedErrors  int    `json:"unfixed_errors"`
	Examples       []struct {
		SessionID string `json:"session_id"`
		Timestamp string `json:"timestamp"`
		ErrorText string `json:"error_text"`
		FixText   string `json:"fix_text"`
		Signature string `json:"signature"`
	} `json:"examples"`
}

type bugResultShape struct {
	Patterns          []bugPatternShape `json:"patterns"`
	TotalPairs        int               `json:"total_pairs"`
	TotalErrors       int               `json:"total_errors"`
	TotalPatterns     int               `json:"total_patterns"`
	TruncatedPatterns int               `json:"truncated_patterns"`
}

func decodeBugResult(t *testing.T, out string) bugResultShape {
	t.Helper()
	var res bugResultShape
	require.NoError(t, json.Unmarshal([]byte(out), &res), "analyze_bugs output must decode: %s", out)
	return res
}

func TestService_AnalyzeBugs_MaxPatternsCapsPatternCount(t *testing.T) {
	const distinct = 12
	entries := withSessionID(oversizedBugCorpus(distinct, 1), "sess-cap")
	projectPath := setupProjectDirWithEntries(t, entries)

	svc := analysis.New()

	out, err := svc.AnalyzeBugs(map[string]interface{}{
		"working_dir":  projectPath,
		"max_patterns": float64(5),
	})
	require.NoError(t, err)

	res := decodeBugResult(t, out)
	assert.Len(t, res.Patterns, 5, "max_patterns=5 must return at most 5 patterns")
	assert.Equal(t, distinct, res.TotalPatterns, "the uncapped pattern count must still be reported")
	assert.Equal(t, distinct-5, res.TruncatedPatterns, "truncated_patterns must report how many were dropped")
}

func TestService_AnalyzeBugs_MaxPatternsZeroIsUnlimited(t *testing.T) {
	const distinct = 12
	entries := withSessionID(oversizedBugCorpus(distinct, 1), "sess-unlimited")
	projectPath := setupProjectDirWithEntries(t, entries)

	svc := analysis.New()

	out, err := svc.AnalyzeBugs(map[string]interface{}{
		"working_dir":  projectPath,
		"max_patterns": float64(0),
	})
	require.NoError(t, err)

	res := decodeBugResult(t, out)
	assert.Len(t, res.Patterns, distinct, "max_patterns=0 must mean unlimited")
	assert.Equal(t, 0, res.TruncatedPatterns)
}

// TestService_AnalyzeBugs_DefaultIsDeterministic pins that the cap selects the
// same patterns run-to-run: the analyzer builds its pattern slice by ranging
// over a map (randomized order) and sorts by Recurrences only, so ties left
// "which N survive the cap" to chance before this task gave the ordering a
// total key (recurrences, fix_count, signature).
func TestService_AnalyzeBugs_DefaultIsDeterministic(t *testing.T) {
	entries := withSessionID(oversizedBugCorpus(30, 1), "sess-determinism")
	projectPath := setupProjectDirWithEntries(t, entries)

	svc := analysis.New()
	args := map[string]interface{}{"working_dir": projectPath}

	first, err := svc.AnalyzeBugs(args)
	require.NoError(t, err)
	for i := 0; i < 3; i++ {
		next, err := svc.AnalyzeBugs(args)
		require.NoError(t, err)
		require.Equal(t, first, next, "a default analyze_bugs call must be reproducible (iteration %d)", i)
	}
}

func TestService_AnalyzeBugs_ExamplesAreStructuredObjects(t *testing.T) {
	ts := "2025-10-02T10:00:00.000Z"
	entries := withSessionID([]types.SessionEntry{
		toolUseEntry("u-err", ts, "tu-err", "Bash", map[string]interface{}{"command": "boom"}),
		toolResultEntry("r-err", ts, "tu-err", "boom: command not found", "error", "boom: command not found"),
		toolUseEntry("u-ok", ts, "tu-ok", "Bash", map[string]interface{}{"command": "fixed"}),
		toolResultEntry("r-ok", ts, "tu-ok", "the fix worked", "success", ""),
	}, "sess-shape")
	projectPath := setupProjectDirWithEntries(t, entries)

	svc := analysis.New()
	out, err := svc.AnalyzeBugs(map[string]interface{}{"working_dir": projectPath})
	require.NoError(t, err)

	res := decodeBugResult(t, out)
	require.Len(t, res.Patterns, 1, "one distinct error signature expected")
	require.NotEmpty(t, res.Patterns[0].Examples, "the pattern must carry its occurrence")

	ex := res.Patterns[0].Examples[0]
	assert.Equal(t, "sess-shape", ex.SessionID, "example must carry the session it was observed in")
	assert.Equal(t, ts, ex.Timestamp, "example must carry the occurrence timestamp")
	assert.Equal(t, "boom: command not found", ex.ErrorText, "example must carry the error text")
	assert.Equal(t, "the fix worked", ex.FixText, "example must carry the paired fix output")
	assert.Equal(t, res.Patterns[0].ErrorSignature, ex.Signature, "example must carry its own signature")

	// The pre-fix shape was a bare JSON string in the examples array; guard
	// against a silent regression to it.
	var raw struct {
		Patterns []struct {
			Examples []interface{} `json:"examples"`
		} `json:"patterns"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &raw))
	require.Len(t, raw.Patterns, 1)
	require.NotEmpty(t, raw.Patterns[0].Examples)
	_, isString := raw.Patterns[0].Examples[0].(string)
	assert.False(t, isString, "examples entries must be objects, not bare strings")
}

func TestService_AnalyzeBugs_DefaultFitsInlineThreshold(t *testing.T) {
	const distinct = 30
	entries := withSessionID(oversizedBugCorpus(distinct, 1), "sess-threshold")
	projectPath := setupProjectDirWithEntries(t, entries)

	svc := analysis.New()

	defaultOut, err := svc.AnalyzeBugs(map[string]interface{}{"working_dir": projectPath})
	require.NoError(t, err)
	assert.Less(t, len(defaultOut), inlineThresholdBytes,
		"a default analyze_bugs call must stay inline (<%d bytes), got %d", inlineThresholdBytes, len(defaultOut))

	// The bound comes from this task's shaping, not from a small seeded
	// corpus: the analyzer's own uncapped result for the very same corpus
	// (every pattern, every raw error string) is far past the threshold.
	raw, err := analyzer.AnalyzeBugs(entries, types.ExtractToolCalls(entries), 0)
	require.NoError(t, err)
	rawJSON, err := json.Marshal(raw)
	require.NoError(t, err)
	assert.Greater(t, len(rawJSON), inlineThresholdBytes,
		"the unshaped analyzer result must exceed the threshold, otherwise this test proves nothing (got %d)", len(rawJSON))
}

func TestService_AnalyzeBugs_StatsOnlyDoesNotMaterializeExamples(t *testing.T) {
	entries := withSessionID(oversizedBugCorpus(30, 2), "sess-stats")
	projectPath := setupProjectDirWithEntries(t, entries)

	svc := analysis.New()

	out, err := svc.AnalyzeBugs(map[string]interface{}{
		"working_dir": projectPath,
		"stats_only":  true,
	})
	require.NoError(t, err)

	assert.NotContains(t, out, "examples", "stats_only must not materialize examples")
	assert.NotContains(t, out, "fix_text", "stats_only must not materialize fix text")
	assert.NotContains(t, out, strings.Repeat("e", 300), "stats_only must not carry raw error text")
	assert.Less(t, len(out), 4096, "stats_only output must stay small regardless of example size")
}

// TestService_AnalyzeBugs_RealProjectCorpusFitsInlineThreshold is the
// corpus-level check the task asks for: the project's own live session corpus
// (the one that produced the 70,248-byte default response) must come back
// under the inline threshold with no arguments at all.
//
// The corpus lives under the *main* checkout's project hash, not this
// worktree's, so the main checkout is resolved through the shared git common
// dir. On a machine with no corpus for this repository the test skips rather
// than asserting anything vacuous.
func TestService_AnalyzeBugs_RealProjectCorpusFitsInlineThreshold(t *testing.T) {
	mainRoot := mainCheckoutRoot(t)
	svc := analysis.New()

	loc := locator.NewSessionLocator()
	files, err := loc.AllSessionsFromProject(mainRoot)
	if err != nil || len(files) == 0 {
		t.Skipf("no session corpus for %s (err=%v, files=%d)", mainRoot, err, len(files))
	}

	out, err := svc.AnalyzeBugs(map[string]interface{}{"working_dir": mainRoot})
	require.NoError(t, err)

	res := decodeBugResult(t, out)
	require.NotEmpty(t, res.Patterns, "the real corpus must yield patterns, otherwise this test proves nothing")
	assert.Less(t, len(out), inlineThresholdBytes,
		"default analyze_bugs on %s must stay inline (<%d bytes), got %d", mainRoot, inlineThresholdBytes, len(out))
	assert.LessOrEqual(t, len(res.Patterns), 20, "default max_patterns must cap the real corpus's pattern list")
	t.Logf("real corpus: %d sessions, %d patterns returned (%d total), %d bytes",
		len(files), len(res.Patterns), res.TotalPatterns, len(out))
}

// mainCheckoutRoot returns the repository's main (non-worktree) checkout path,
// resolved through the git common dir shared by every worktree.
func mainCheckoutRoot(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "--path-format=absolute", "--git-common-dir")
	cmd.Dir = mustGetwd(t)
	out, err := cmd.Output()
	require.NoError(t, err, "git rev-parse --git-common-dir must succeed")
	return filepath.Dir(strings.TrimSpace(string(out)))
}

func mustGetwd(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	return wd
}
