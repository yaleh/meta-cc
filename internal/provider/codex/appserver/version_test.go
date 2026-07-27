package appserver

import "testing"

// TestSupportsPagination is the DIR-032 "capability negotiation, not
// reinvention" proof: SupportsPagination reports the same stable
// thread/list cursor-pagination surface MinSupportedVersion already gates,
// exposed as its own named, unit-testable capability query.
func TestSupportsPagination(t *testing.T) {
	if MinSupportedVersion.SupportsPagination() != true {
		t.Fatalf("MinSupportedVersion itself must support pagination")
	}
	older := Version{Major: 0, Minor: 100, Patch: 0}
	if older.SupportsPagination() {
		t.Fatalf("a version below MinSupportedVersion must not report pagination support")
	}
	newer := Version{Major: 1, Minor: 0, Patch: 0}
	if !newer.SupportsPagination() {
		t.Fatalf("a version above MinSupportedVersion must report pagination support")
	}
}

// TestSupportsExperimentalTurnPagination covers both the
// capability-present and capability-absent paths (DIR-032 instruction #7):
// every real, currently-known Codex CLI version reports false (fails safely
// to the existing non-paginated thread/read(includeTurns) behavior), while
// a version at/above the placeholder floor reports true (the path a future
// confirmed method/version pair would take).
func TestSupportsExperimentalTurnPagination(t *testing.T) {
	if MinSupportedVersion.SupportsExperimentalTurnPagination() {
		t.Fatalf("no real Codex CLI version should report experimental turn pagination support yet")
	}
	future := Version{Major: 1, Minor: 0, Patch: 0}
	if future.SupportsExperimentalTurnPagination() {
		t.Fatalf("a version below the experimental floor must not report support")
	}
	if !ExperimentalTurnPaginationMinVersion.SupportsExperimentalTurnPagination() {
		t.Fatalf("the experimental floor version itself must report support (capability-present path)")
	}
}

func TestVersionLessAndString(t *testing.T) {
	a := Version{Major: 0, Minor: 145, Patch: 0}
	b := Version{Major: 0, Minor: 146, Patch: 0}
	if !a.Less(b) || b.Less(a) {
		t.Fatalf("expected a < b")
	}
	if got := (Version{Major: 1, Minor: 2, Patch: 3}).String(); got != "1.2.3" {
		t.Fatalf("String() = %q", got)
	}
	if got := (Version{Raw: "codex-cli 1.2.3"}).String(); got != "codex-cli 1.2.3" {
		t.Fatalf("String() should prefer Raw when set, got %q", got)
	}
}

func TestParseVersionBasic(t *testing.T) {
	v, err := ParseVersion("codex-cli 0.145.0")
	if err != nil {
		t.Fatalf("ParseVersion: %v", err)
	}
	if v.Major != 0 || v.Minor != 145 || v.Patch != 0 {
		t.Fatalf("unexpected parsed version: %#v", v)
	}
	if _, err := ParseVersion(""); err == nil {
		t.Fatalf("expected error for empty version string")
	}
}
