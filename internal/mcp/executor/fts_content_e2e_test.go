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
