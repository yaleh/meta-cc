package locator

import "fmt"

// SkipReport is the one implementation of the corpus-tolerance contract every
// enumerating path in this repo must honor:
//
//  1. one unusable session file must never erase the results derived from the
//     other files in the corpus (DIR-018 / DIR-030), and
//  2. an excluded file must never be dropped SILENTLY — the exclusion is
//     recorded here so callers can carry it into their response metadata.
//
// DIR-018 established both halves for internal/analysis/service.go's loadData.
// DIR-094 found the rest of the corpus-enumerating surface (the Claude
// provider's ListSessions behind query_sessions, get_timeline, and the other
// analysis tools) implementing only half the contract — either failing the
// whole batch on one bad file, or skipping it with no trace. Routing all of
// them through this single type is what makes the tolerance *unified* rather
// than re-derived at each call site.
//
// It lives in locator because that package owns corpus enumeration
// (AllSessionsFromProject) and is already a dependency of both the Claude
// provider and the analysis service, so the shared definition adds no new
// layering edge.
//
// A SkipReport is not safe for concurrent use; each enumeration pass owns one.
type SkipReport struct {
	paths    []string
	warnings []string
}

// Skip records that path was excluded from the corpus because of err. The
// recorded warning names the file, so a caller surfacing it tells the reader
// exactly which data is missing.
func (r *SkipReport) Skip(path string, err error) {
	if err == nil {
		r.SkipReason(path, "unknown reason")
		return
	}
	r.SkipReason(path, err.Error())
}

// SkipReason records an exclusion whose reason is already a string, for the
// cases where the underlying error's own text would be redundant or
// misleading (e.g. the Claude provider's "no message entries in <path>"
// sentinel, which already carries the path this method prefixes).
func (r *SkipReport) SkipReason(path, reason string) {
	if r == nil {
		return
	}
	r.paths = append(r.paths, path)
	r.warnings = append(r.warnings, fmt.Sprintf("skipped session file %s: %s", path, reason))
}

// AdoptWarning records an exclusion that a downstream layer already rendered
// as a human-readable warning, preserving that text verbatim. It is used when
// the originating layer reports a session ID rather than a corpus file path
// (providerrecords.Build's per-session messages), so no path is recorded and
// the caller's own wording is not rewritten here.
func (r *SkipReport) AdoptWarning(warning string) {
	if r == nil || warning == "" {
		return
	}
	r.warnings = append(r.warnings, warning)
}

// Warnings returns the human-readable exclusions, or nil when nothing was
// skipped — so callers that embed them in an `omitempty` field leave the wire
// format unchanged for a clean corpus (the DIR-018 contract).
func (r *SkipReport) Warnings() []string {
	if r == nil || len(r.warnings) == 0 {
		return nil
	}
	return append([]string(nil), r.warnings...)
}

// Paths returns the corpus file paths that were excluded, or nil. It is the
// machine-readable counterpart to Warnings, for the `skipped_files` response
// metadata DIR-094 adds. Exclusions adopted via AdoptWarning carry no path and
// therefore do not appear here — they remain visible in Warnings.
func (r *SkipReport) Paths() []string {
	if r == nil || len(r.paths) == 0 {
		return nil
	}
	return append([]string(nil), r.paths...)
}

// Empty reports whether nothing was excluded.
func (r *SkipReport) Empty() bool {
	return r == nil || (len(r.paths) == 0 && len(r.warnings) == 0)
}
