package response

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dir079Diagnostics returns a fully-populated diagnostics map using exactly
// the DIR-079 QueryDiagnostics vocabulary keys emitted by
// internal/mcp/query/stage.go (no second vocabulary is invented).
func dir079Diagnostics() map[string]interface{} {
	return map[string]interface{}{
		"backend":            "jsonl",
		"provider":           "claude",
		"provider_effective": "claude",
		"files_considered":   3,
		"files_loaded":       3,
		"files_skipped":      0,
		"records_scanned":    128,
		"matches_returned":   2,
		"truncated":          false,
		"degraded":           false,
		"skip_warnings":      []string{},
	}
}

func TestBuildEvidenceBundle_CompleteContract(t *testing.T) {
	bundle, err := BuildEvidenceBundle(
		map[string]interface{}{"tool": "query_session_content", "role": "assistant"},
		claudeFixtureRecords(),
		dir079Diagnostics(),
		[]string{".timestamp", ".provider"},
	)
	require.NoError(t, err)

	assert.Equal(t, EvidenceBundleVersion, bundle["bundle_version"])
	assert.NotEmpty(t, bundle["generated_at"])

	criteria := bundle["query_criteria"].(map[string]interface{})
	assert.Equal(t, "query_session_content", criteria["tool"])

	completeness := bundle["completeness"].(map[string]interface{})
	for _, key := range requiredDiagnosticsKeys {
		assert.Contains(t, completeness, key, "completeness must reuse the DIR-079 vocabulary: %s", key)
	}
	assert.Equal(t, 3, completeness["files_considered"])
	assert.Empty(t, completeness["completeness_notes"], "no notes when every key is present")

	excerpts := bundle["excerpts"].([]interface{})
	assert.Len(t, excerpts, 2)
	excerpt := excerpts[0].(map[string]interface{})
	assert.Contains(t, excerpt, ".timestamp")
	assert.NotContains(t, excerpt, "message", "excerpts must be projected, not full transcripts")

	provenance := bundle["provenance"].(map[string]interface{})
	assert.Equal(t, []string{"claude"}, provenance["providers"])
	assert.Equal(t, 2, provenance["total_records"])
	assert.Equal(t, 2, provenance["excerpt_count"])
}

func TestBuildEvidenceBundle_MissingDiagnosticsNoted(t *testing.T) {
	partial := map[string]interface{}{"provider": "codex", "degraded": true}
	bundle, err := BuildEvidenceBundle(
		map[string]interface{}{"tool": "execute_stage2_query"},
		codexFixtureRecords(),
		partial,
		nil,
	)
	require.NoError(t, err)

	completeness := bundle["completeness"].(map[string]interface{})
	for _, key := range requiredDiagnosticsKeys {
		assert.Contains(t, completeness, key, "vocabulary keys must always be present")
	}
	notes, _ := completeness["completeness_notes"].([]string)
	assert.NotEmpty(t, notes, "missing diagnostics must be surfaced, not papered over")
	assert.Contains(t, strings.Join(notes, ","), "files_considered")
	assert.Equal(t, true, completeness["degraded"])
}

func TestBuildEvidenceBundle_ExcerptsRedacted(t *testing.T) {
	bundle, err := BuildEvidenceBundle(
		map[string]interface{}{"tool": "query_session_signals"},
		sensitiveFixtureRecords(),
		dir079Diagnostics(),
		nil,
	)
	require.NoError(t, err)

	serialized, err := json.Marshal(bundle)
	require.NoError(t, err)
	assert.NotContains(t, string(serialized), "sk-live-1234567890")
	assert.Contains(t, string(serialized), redactedMarker)
}

func TestBuildEvidenceBundle_ExcerptCap(t *testing.T) {
	records := make([]interface{}, 10)
	for i := range records {
		records[i] = map[string]interface{}{"id": float64(i)}
	}
	bundle, err := BuildEvidenceBundle(map[string]interface{}{"tool": "x"}, records, dir079Diagnostics(), nil)
	require.NoError(t, err)
	assert.Len(t, bundle["excerpts"].([]interface{}), maxEvidenceExcerpts)
}

func TestBuildEvidenceBundle_RequiresToolCriteria(t *testing.T) {
	_, err := BuildEvidenceBundle(map[string]interface{}{"role": "user"}, nil, dir079Diagnostics(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tool")
}

func TestBuildEvidenceBundle_ProjectionErrorsPropagate(t *testing.T) {
	_, err := BuildEvidenceBundle(
		map[string]interface{}{"tool": "x"},
		claudeFixtureRecords(),
		dir079Diagnostics(),
		[]string{".not_a_path"},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown projection path")
}

func TestBuildEvidenceBundle_MixedProviderProvenance(t *testing.T) {
	bundle, err := BuildEvidenceBundle(
		map[string]interface{}{"tool": "x"},
		heterogeneousFixtureRecords(),
		dir079Diagnostics(),
		nil,
	)
	require.NoError(t, err)
	provenance := bundle["provenance"].(map[string]interface{})
	assert.Equal(t, []string{"claude", "codex"}, provenance["providers"])
}

func TestBuildEvidenceBundle_CriteriaStringsBounded(t *testing.T) {
	bundle, err := BuildEvidenceBundle(
		map[string]interface{}{"tool": "x", "pattern": strings.Repeat("p", 5000)},
		nil,
		dir079Diagnostics(),
		nil,
	)
	require.NoError(t, err)
	criteria := bundle["query_criteria"].(map[string]interface{})
	pattern, _ := criteria["pattern"].(string)
	assert.LessOrEqual(t, len(pattern), maxCriteriaValueLen+64)
}
