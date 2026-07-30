package response

import (
	"fmt"
	"time"
)

// DIR-080 / ADR-009: optional issue/task-ready evidence bundle. This is a
// WORKFLOW feature, deliberately separate from the base response envelope:
// it packages bounded, redacted excerpts plus stable provenance, the query
// criteria that produced them, and completeness/fallback diagnostics so a
// finding can be attached to a task without exposing full transcripts.

const (
	// EvidenceBundleVersion versions the bundle contract.
	EvidenceBundleVersion = 1
	// maxEvidenceExcerpts bounds how many records a bundle carries.
	maxEvidenceExcerpts = 5
	// maxCriteriaValueLen bounds each query-criteria string value.
	maxCriteriaValueLen = 200
)

// requiredDiagnosticsKeys is EXACTLY the DIR-079 QueryDiagnostics vocabulary
// emitted by internal/mcp/query/stage.go (backend … skip_warnings). Evidence
// bundles reuse these keys rather than inventing a second vocabulary, per
// the DIR-080 dependency contract.
var requiredDiagnosticsKeys = []string{
	"backend",
	"provider",
	"provider_effective",
	"files_considered",
	"files_loaded",
	"files_skipped",
	"records_scanned",
	"matches_returned",
	"truncated",
	"degraded",
	"skip_warnings",
}

// BuildEvidenceBundle assembles a bounded evidence bundle:
//   - criteria must carry "tool" (the query that produced the evidence);
//     string values are length-bounded;
//   - excerpts are the first maxEvidenceExcerpts records, optionally
//     projected to paths (validated by ProjectRecords), always redacted;
//   - completeness reuses the DIR-079 diagnostics vocabulary verbatim and
//     records missing keys in completeness_notes instead of hiding them;
//   - provenance reports observed providers and record/excerpt counts.
func BuildEvidenceBundle(criteria map[string]interface{}, records []interface{}, diagnostics map[string]interface{}, paths []string) (map[string]interface{}, error) {
	if criteria == nil {
		criteria = map[string]interface{}{}
	}
	if tool, _ := criteria["tool"].(string); tool == "" {
		return nil, fmt.Errorf("evidence bundle criteria must include \"tool\" (the query tool that produced the evidence)")
	}

	boundedCriteria := make(map[string]interface{}, len(criteria))
	for k, v := range criteria {
		if s, ok := v.(string); ok && len(s) > maxCriteriaValueLen {
			boundedCriteria[k] = s[:maxCriteriaValueLen] + fmt.Sprintf("...[truncated: %d chars]", len(s))
			continue
		}
		boundedCriteria[k] = v
	}

	source := records
	if len(paths) > 0 {
		projected, err := ProjectRecords(records, paths)
		if err != nil {
			return nil, fmt.Errorf("evidence bundle projection failed: %w", err)
		}
		source = projected
	}
	if len(source) > maxEvidenceExcerpts {
		source = source[:maxEvidenceExcerpts]
	}
	excerpts := make([]interface{}, 0, len(source))
	fieldsRedacted := 0
	for _, record := range source {
		sanitized, n := redactAndTruncate(record, 0, maxSampleStringLen)
		excerpts = append(excerpts, sanitized)
		fieldsRedacted += n
	}

	if diagnostics == nil {
		diagnostics = map[string]interface{}{}
	}
	completeness := make(map[string]interface{}, len(requiredDiagnosticsKeys)+1)
	notes := []string{}
	for _, key := range requiredDiagnosticsKeys {
		if v, ok := diagnostics[key]; ok {
			completeness[key] = v
			continue
		}
		completeness[key] = diagnosticsZeroValue(key)
		notes = append(notes, fmt.Sprintf("diagnostics key %q was not supplied by the query; defaulted", key))
	}
	completeness["completeness_notes"] = notes

	return map[string]interface{}{
		"bundle_version": EvidenceBundleVersion,
		"generated_at":   time.Now().UTC().Format(time.RFC3339),
		"query_criteria": boundedCriteria,
		"provenance": map[string]interface{}{
			"providers":     evidenceProviders(records),
			"total_records": len(records),
			"excerpt_count": len(excerpts),
		},
		"completeness": completeness,
		"excerpts":     excerpts,
		"redaction": map[string]interface{}{
			"applied":         true,
			"fields_redacted": fieldsRedacted,
		},
	}, nil
}

func diagnosticsZeroValue(key string) interface{} {
	switch key {
	case "skip_warnings":
		return []string{}
	case "truncated", "degraded":
		return false
	case "backend", "provider", "provider_effective":
		return ""
	default:
		return 0
	}
}

// evidenceProviders reports the distinct provider values observed in the
// result records (from the derived shape's bounded value enumeration).
func evidenceProviders(records []interface{}) []string {
	shape := DeriveResultShape(records)
	if shape == nil || shape.Root == nil || shape.Root.Properties == nil {
		return []string{}
	}
	provider := shape.Root.Properties["provider"]
	if provider == nil || len(provider.Values) == 0 {
		return []string{}
	}
	return provider.Values
}
