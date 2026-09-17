package executor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/config"
)

func executeContentEnvelope(t *testing.T, e *ToolExecutor, args map[string]interface{}, disabled bool) map[string]interface{} {
	t.Helper()
	if disabled {
		t.Setenv("META_CC_DISABLE_FTS_INDEX", "1")
	} else {
		t.Setenv("META_CC_DISABLE_FTS_INDEX", "0")
	}
	out, err := e.ExecuteTool(&config.Config{}, "query_session_content", args)
	require.NoError(t, err)
	var envelope map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(out), &envelope))
	if envelope["mode"] == "file_ref" {
		ref := envelope["file_ref"].(map[string]interface{})
		raw, err := os.ReadFile(ref["path"].(string))
		require.NoError(t, err)
		defer os.Remove(ref["path"].(string))
		var data []interface{}
		for _, line := range splitNonEmptyLines(string(raw)) {
			var item interface{}
			require.NoError(t, json.Unmarshal([]byte(line), &item))
			data = append(data, item)
		}
		envelope["data"] = data
	}
	delete(envelope, "warnings")
	delete(envelope, "file_ref")
	delete(envelope, "mode")
	return envelope
}

func splitNonEmptyLines(raw string) []string {
	var lines []string
	start := 0
	for i := 0; i <= len(raw); i++ {
		if i != len(raw) && raw[i] != '\n' {
			continue
		}
		if i > start {
			lines = append(lines, raw[start:i])
		}
		start = i + 1
	}
	return lines
}

func TestQuerySessionContent_UnicodeCaseFoldingParity(t *testing.T) {
	// Skipped in -short mode (DIR-083). This is a dual-scan parity test: every
	// query is executed twice -- once with the FTS index disabled and once
	// enabled -- and the two full envelopes are deep-compared. That shape is
	// the expensive one, and its cost tracks whatever the codex fixture has to
	// do per query: measured 2026-07-30 at 4.3s on this test alone, back when
	// the fixture still drove a real `codex app-server` child per query (that
	// cost was removed by cf6efaf pinning META_CC_CODEX_BACKEND=files, leaving
	// ~0.1s today). The guard is kept because the skip is what bounds the cost
	// on the hot path if that pin is ever weakened or removed.
	//
	// The parity guarantee itself is NOT weakened by this skip, only relocated
	// off the hot path: `make push` and CI run `go test` WITHOUT -short (see
	// the Makefile's `test-all` target, reached via `check-push-ready`), so
	// "the FTS/Unicode fast path agrees with the canonical scan" is still
	// asserted exactly where it matters. Plain
	// `go test ./internal/mcp/executor/` (non-short) also still runs it.
	if testing.Short() {
		t.Skip("skipping DIR-083 FTS/Unicode dual-scan parity assertion in -short mode; it still runs on the non-short `go test` / `make push` path")
	}

	project := setupCodexRolloutFixtureProject(t, "rollout-context-turns-sample.jsonl")
	rollout := filepath.Join(os.Getenv("META_CC_CODEX_ROOT"), "rollout.jsonl")
	raw, err := os.ReadFile(rollout)
	require.NoError(t, err)
	raw = []byte(strings.Replace(string(raw), "baseline", "KELVIN ÉCOLE ÅNGSTRÖM", 1))
	require.NoError(t, os.WriteFile(rollout, raw, 0o644))

	e := NewToolExecutor()
	for _, query := range []string{"kelvin", "Kelvin", "école", "École", "ångström", "Ångström"} {
		args := map[string]interface{}{"role": "user", "provider": "codex", "contains": query, "working_dir": project}
		direct := executeContentEnvelope(t, e, args, true)
		indexed := executeContentEnvelope(t, e, args, false)
		if !reflect.DeepEqual(direct, indexed) {
			t.Fatalf("Unicode query %q differs from canonical scan\ndirect:  %#v\nindexed: %#v", query, direct, indexed)
		}
		data, ok := direct["data"].([]interface{})
		if !ok || len(data) == 0 {
			t.Fatalf("canonical Unicode query %q returned no records: %#v", query, direct)
		}
	}
}

func TestQuerySessionContent_FTSParityWithCanonicalScan(t *testing.T) {
	// Skipped in -short mode (DIR-083), for the same reason as the Unicode
	// case-folding parity test above: this runs 7 sub-cases, each executing the
	// query twice (FTS index disabled vs enabled) and deep-comparing the two
	// envelopes, so it is the other half of the dual-scan parity pair -- the
	// two were measured 2026-07-30 at 5.3s + 4.3s of serial wall time on every
	// `-short` run. That specific cost was already removed by cf6efaf (the
	// codex fixture now pins META_CC_CODEX_BACKEND=files instead of spawning an
	// app-server child per query), leaving ~0.1s today; the skip is kept as the
	// guard that bounds this cost on the hot path if that pin is weakened.
	//
	// The parity guarantee itself is NOT weakened by this skip, only relocated
	// off the hot path: `make push` and CI run `go test` WITHOUT -short (see
	// the Makefile's `test-all` target, reached via `check-push-ready`), so
	// "the FTS fast path agrees with the canonical scan" is still asserted
	// exactly where it matters. Plain `go test ./internal/mcp/executor/`
	// (non-short) also still runs all 7 sub-cases.
	if testing.Short() {
		t.Skip("skipping DIR-083 FTS dual-scan parity assertion in -short mode; it still runs on the non-short `go test` / `make push` path")
	}

	project := setupCodexRolloutFixtureProject(t, "rollout-context-turns-sample.jsonl")
	e := NewToolExecutor()
	cases := []struct {
		name string
		args map[string]interface{}
	}{
		{"provider role pattern", map[string]interface{}{"role": "user", "provider": "codex", "pattern": "baseline", "working_dir": project}},
		{"literal contains", map[string]interface{}{"role": "assistant", "provider": "codex", "contains": "ack three", "working_dir": project}},
		{"time bounds", map[string]interface{}{"role": "user", "provider": "codex", "contains": "baseline", "since": "2026-07-20T09:00:10Z", "working_dir": project}},
		{"canonical context", map[string]interface{}{"role": "user", "provider": "codex", "pattern": "measure", "context_turns": float64(1), "working_dir": project}},
		{"grouping", map[string]interface{}{"role": "user", "provider": "codex", "contains": "baseline", "group_by_session": true, "working_dir": project}},
		{"pagination", map[string]interface{}{"role": "user", "provider": "codex", "contains": "baseline", "offset": float64(1), "page_size": float64(2), "working_dir": project}},
		{"jq composition", map[string]interface{}{"role": "user", "provider": "codex", "contains": "baseline", "jq_filter": `.[] | select(.turn_id == "turn-4")`, "working_dir": project}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			direct := executeContentEnvelope(t, e, tc.args, true)
			indexed := executeContentEnvelope(t, e, tc.args, false)
			if !reflect.DeepEqual(direct, indexed) {
				t.Fatalf("indexed response differs from canonical scan\ndirect:  %#v\nindexed: %#v", direct, indexed)
			}
		})
	}
}
