package analysis_test

// DIR-096: analyze_bugs gained a first-class pattern-count cap (max_patterns)
// and grew structured examples. Before it, the pattern count was unbounded and
// defaulted to every distinct error signature, so even limit:3 (which bounds
// only the per-pattern example list) produced a 92KB response on a 5-day
// corpus and spilled to file_ref mode. These tests pin the four acceptance
// criteria through the real Service boundary — the same path an MCP caller
// takes — rather than through the analyzer primitive.
//
// Parameters are passed as float64 throughout because that is what the MCP
// JSON-RPC layer actually delivers (encoding/json decodes every number into
// float64); see intArg.

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/analysis"
	"github.com/yaleh/meta-cc/internal/mcp/response"
	"github.com/yaleh/meta-cc/internal/types"
)

// bugsCorpus is a synthetic session whose shape mirrors this project's real
// corpus: many distinct error signatures, each recurring a handful of times,
// each followed by its own successful call so the pattern has a paired fix.
type bugsCorpus struct {
	sessionID string
	entries   []types.SessionEntry
}

// buildBugsCorpus returns a corpus with `signatures` distinct error signatures.
// Signature i recurs recurrences(i) times. Error texts are padded to
// errTextLen bytes, which is what actually drives response size.
func buildBugsCorpus(t *testing.T, sessionID string, signatures int, recurrences func(i int) int, errTextLen int) bugsCorpus {
	t.Helper()
	var entries []types.SessionEntry
	seq := 0
	nextTS := func() string {
		seq++
		return fmt.Sprintf("2026-01-01T%02d:%02d:%02d.000Z", seq/3600, (seq/60)%60, seq%60)
	}

	for p := 0; p < signatures; p++ {
		errText := fmt.Sprintf("signature-%03d: %s", p, strings.Repeat("e", errTextLen))
		for r := 0; r < recurrences(p); r++ {
			useID := fmt.Sprintf("use-%d-%d", p, r)
			entries = append(entries,
				toolUseEntry(fmt.Sprintf("u-%d-%d", p, r), nextTS(), useID, "Bash", nil),
				toolResultEntry(fmt.Sprintf("r-%d-%d", p, r), nextTS(), useID, "", "error", errText),
			)
			// The paired fix: a same-tool success within three positions.
			fixID := fmt.Sprintf("fix-%d-%d", p, r)
			entries = append(entries,
				toolUseEntry(fmt.Sprintf("fu-%d-%d", p, r), nextTS(), fixID, "Bash", nil),
				toolResultEntry(fmt.Sprintf("fr-%d-%d", p, r), nextTS(), fixID, "fixed", "success", ""),
			)
		}
	}

	// The corpus helpers do not set SessionID; stamp it so the example's
	// session_id assertion has a real value to resolve through the entry index.
	for i := range entries {
		entries[i].SessionID = sessionID
	}
	return bugsCorpus{sessionID: sessionID, entries: entries}
}

// uniform returns a recurrence function where every signature recurs n times.
func uniform(n int) func(int) int { return func(int) int { return n } }

type bugsResponse struct {
	TotalPatterns int `json:"total_patterns"`
	TotalErrors   int `json:"total_errors"`
	Patterns      []struct {
		ErrorSignature string `json:"error_signature"`
		Recurrences    int    `json:"recurrences"`
		Examples       []struct {
			SessionID string `json:"session_id"`
			Timestamp string `json:"timestamp"`
			ErrorText string `json:"error_text"`
			FixText   string `json:"fix_text"`
			Signature string `json:"signature"`
		} `json:"examples"`
	} `json:"patterns"`
}

func analyzeBugsJSON(t *testing.T, projectPath string, extra map[string]interface{}) (string, bugsResponse) {
	t.Helper()
	args := map[string]interface{}{"working_dir": projectPath}
	for k, v := range extra {
		args[k] = v
	}
	out, err := analysis.New().AnalyzeBugs(args)
	require.NoError(t, err, "AnalyzeBugs(%v)", args)
	var parsed bugsResponse
	require.NoError(t, json.Unmarshal([]byte(out), &parsed), "output must be valid JSON")
	return out, parsed
}

// TestAnalyzeBugs_ServiceMaxPatternsCapsReturnedPatterns pins AC #1 at the
// Service boundary: max_patterns=5 returns at most 5 patterns.
func TestAnalyzeBugs_ServiceMaxPatternsCapsReturnedPatterns(t *testing.T) {
	corpus := buildBugsCorpus(t, "sess-max", 12, uniform(2), 40)
	projectPath := setupProjectDirWithEntries(t, corpus.entries)

	out, parsed := analyzeBugsJSON(t, projectPath, map[string]interface{}{"max_patterns": float64(5)})
	assert.LessOrEqual(t, len(parsed.Patterns), 5, "AC #1: max_patterns=5 must return at most 5 patterns")
	assert.Equal(t, 5, len(parsed.Patterns), "corpus has 12 signatures, so the cap is binding")
	assert.Equal(t, 12, parsed.TotalPatterns, "the pre-cap pattern count must stay visible")
	assert.Equal(t, 24, parsed.TotalErrors, "totals describe the corpus, not the capped slice")
	assert.NotContains(t, out, `"patterns":[]`, "a capped result must not look empty")

	// Explicit 0 means unlimited, per the parameter contract.
	_, unlimited := analyzeBugsJSON(t, projectPath, map[string]interface{}{"max_patterns": float64(0)})
	assert.Equal(t, 12, len(unlimited.Patterns), "max_patterns=0 means unlimited")
}

// TestAnalyzeBugs_ServiceDefaultCallIsCapped verifies the no-args default is
// bounded on both axes: at most analyzer.DefaultMaxPatterns patterns, and at
// most the default per-pattern example count. The example cap is what keeps the
// response bounded as a frequently-recurring signature accumulates examples
// over the corpus's lifetime.
func TestAnalyzeBugs_ServiceDefaultCallIsCapped(t *testing.T) {
	// 30 signatures, each recurring enough to exceed the example cap.
	corpus := buildBugsCorpus(t, "sess-default", 30, uniform(6), 40)
	projectPath := setupProjectDirWithEntries(t, corpus.entries)

	_, parsed := analyzeBugsJSON(t, projectPath, nil)
	require.NotEmpty(t, parsed.Patterns)
	assert.LessOrEqual(t, len(parsed.Patterns), 20, "default must cap the pattern count")
	assert.Equal(t, 30, parsed.TotalPatterns)

	for _, p := range parsed.Patterns {
		assert.LessOrEqual(t, len(p.Examples), 3,
			"default must cap examples per pattern (pattern %s carried %d)",
			p.ErrorSignature, len(p.Examples))
	}

	// Passing limit:0 must still mean "every example", so the cap stays an
	// opt-out rather than a silent loss of data.
	_, explicit := analyzeBugsJSON(t, projectPath, map[string]interface{}{"limit": float64(0)})
	require.NotEmpty(t, explicit.Patterns)
	assert.Equal(t, 6, len(explicit.Patterns[0].Examples),
		"limit:0 must remain unlimited examples per pattern")
}

// TestAnalyzeBugs_ServiceExamplesAreStructured pins AC #2 end-to-end: every
// example is addressable by session_id + timestamp + error_text, with
// signature always present and fix_text present exactly when a fix was paired.
func TestAnalyzeBugs_ServiceExamplesAreStructured(t *testing.T) {
	corpus := buildBugsCorpus(t, "sess-structured", 3, uniform(2), 60)
	projectPath := setupProjectDirWithEntries(t, corpus.entries)

	_, parsed := analyzeBugsJSON(t, projectPath, nil)
	require.NotEmpty(t, parsed.Patterns)

	seen, withFix := 0, 0
	for _, p := range parsed.Patterns {
		for _, e := range p.Examples {
			seen++
			assert.Equal(t, "sess-structured", e.SessionID, "AC #2: example must carry session_id")
			assert.NotEmpty(t, e.Timestamp, "AC #2: example must carry timestamp")
			assert.NotEmpty(t, e.ErrorText, "AC #2: example must carry error_text")
			assert.NotEmpty(t, e.Signature, "AC #2: example must carry signature")
			assert.Equal(t, p.ErrorSignature, e.Signature,
				"each example's signature must match its pattern's")
			if e.FixText != "" {
				withFix++
			}
		}
	}
	assert.NotZero(t, seen, "the corpus must produce examples")
	assert.NotZero(t, withFix, "every error in this corpus is followed by a fix, so fix_text must appear")
}

// TestAnalyzeBugs_DefaultCallFitsInlineThreshold pins AC #3.
//
// The first case reproduces the finding's shape — 75 distinct patterns, the
// count that produced a 92KB response — and shows the unbounded defaults
// (max_patterns:0, limit:0, i.e. the pre-DIR-096 behaviour) exceed the inline
// threshold while the default call does not. The second case isolates the
// other unbounded axis: a few signatures recurring many times.
func TestAnalyzeBugs_DefaultCallFitsInlineThreshold(t *testing.T) {
	threshold := response.DefaultInlineThresholdBytes

	t.Run("many distinct patterns", func(t *testing.T) {
		corpus := buildBugsCorpus(t, "sess-many", 75, uniform(1), 300)
		projectPath := setupProjectDirWithEntries(t, corpus.entries)

		unbounded, err := analysis.New().AnalyzeBugs(map[string]interface{}{
			"working_dir":  projectPath,
			"max_patterns": float64(0),
			"limit":        float64(0),
		})
		require.NoError(t, err)
		require.Greater(t, len(unbounded), threshold,
			"precondition: the unbounded shape must exceed the inline threshold, otherwise this test proves nothing")

		bounded, parsed := analyzeBugsJSON(t, projectPath, nil)
		require.Equal(t, 75, parsed.TotalPatterns)
		require.LessOrEqual(t, len(parsed.Patterns), 20)
		assert.Less(t, len(bounded), threshold,
			"AC #3: a default analyze_bugs call must fit the inline threshold (%d bytes), got %d",
			threshold, len(bounded))
	})

	t.Run("few signatures recurring many times", func(t *testing.T) {
		corpus := buildBugsCorpus(t, "sess-recur", 4, uniform(200), 200)
		projectPath := setupProjectDirWithEntries(t, corpus.entries)

		out, parsed := analyzeBugsJSON(t, projectPath, nil)
		require.Equal(t, 4, parsed.TotalPatterns)
		for _, p := range parsed.Patterns {
			require.Len(t, p.Examples, 3,
				"800 error occurrences must not expand into 800 examples")
		}
		assert.Less(t, len(out), threshold,
			"a default call must stay inline even when signatures recur hundreds of times: got %d", len(out))
	})
}

// TestAnalyzeBugs_StatsOnlyDoesNotMaterializeExamples pins AC #4: stats_only
// must not build examples at all, so its size is independent of how large the
// underlying error texts are.
func TestAnalyzeBugs_StatsOnlyDoesNotMaterializeExamples(t *testing.T) {
	// Each corpus needs its own projects root (setupProjectDirWithEntries
	// re-points META_CC_PROJECTS_ROOT), so the two sizes are captured from
	// sequential subtests and compared afterwards.
	var shortLen, longLen, fullLen int

	t.Run("short error texts", func(t *testing.T) {
		corpus := buildBugsCorpus(t, "sess-short", 6, uniform(3), 20)
		projectPath := setupProjectDirWithEntries(t, corpus.entries)

		out, err := analysis.New().AnalyzeBugs(map[string]interface{}{
			"working_dir": projectPath, "stats_only": true,
		})
		require.NoError(t, err)

		assert.NotContains(t, out, "examples", "stats_only must not emit an examples field")
		assert.NotContains(t, out, "signature-000", "stats_only must not emit example error text")
		shortLen = len(out)
	})

	t.Run("long error texts", func(t *testing.T) {
		corpus := buildBugsCorpus(t, "sess-long", 6, uniform(3), 4000)
		projectPath := setupProjectDirWithEntries(t, corpus.entries)

		out, err := analysis.New().AnalyzeBugs(map[string]interface{}{
			"working_dir": projectPath, "stats_only": true,
		})
		require.NoError(t, err)

		assert.NotContains(t, out, "examples", "stats_only must not emit an examples field")
		assert.NotContains(t, out, "signature-000", "stats_only must not emit example error text")
		longLen = len(out)

		// For contrast, the full path on this same corpus does carry the
		// examples. Checked here rather than after the subtests because the
		// corpus is only reachable while this subtest's env is in scope.
		fullOut, err := analysis.New().AnalyzeBugs(map[string]interface{}{"working_dir": projectPath})
		require.NoError(t, err)
		assert.Contains(t, fullOut, "examples")
		fullLen = len(fullOut)
	})

	// A 200x larger error text must not change the stats_only payload size:
	// the stats path never materializes the text in the first place.
	assert.InDelta(t, shortLen, longLen, 64,
		"AC #4: stats_only output must not grow with error text length (%d vs %d bytes)",
		shortLen, longLen)
	assert.Greater(t, fullLen, longLen,
		"the full path must carry strictly more than the stats_only path")
}
