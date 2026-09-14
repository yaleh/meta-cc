package locator

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestExclusionFor_HealthyFile proves the shared rule stays silent for a file
// that actually contributed entries — the tolerance must not turn into
// warning-on-everything, or the warnings become noise callers learn to ignore.
func TestExclusionFor_HealthyFile(t *testing.T) {
	if got := ExclusionFor("session.jsonl", 3, nil); got != nil {
		t.Fatalf("healthy file must not be excluded, got %v", got)
	}
}

// TestExclusionFor_ZeroEntriesIsExcluded is the DIR-094 core rule: a file that
// parses cleanly but yields nothing (empty, or a metadata-only stub) is just
// as excluded from a corpus result as an unreadable one, and must be reported
// with the shared sentinel so callers can classify it without string matching.
func TestExclusionFor_ZeroEntriesIsExcluded(t *testing.T) {
	exclusion := ExclusionFor("stub.jsonl", 0, nil)
	if exclusion == nil {
		t.Fatal("a zero-entry file must be excluded, not silently accepted")
	}
	if !errors.Is(exclusion.Err, ErrNoMessageEntries) {
		t.Fatalf("exclusion.Err = %v, want ErrNoMessageEntries", exclusion.Err)
	}
	if !strings.Contains(exclusion.Reason(), "no message entries") {
		t.Fatalf("Reason() = %q, must explain the zero-entry case", exclusion.Reason())
	}
}

// TestExclusionFor_ParseErrorIsPreserved proves a genuine read/parse failure
// is reported as itself (wrapped, not replaced), so a caller can still tell an
// unreadable file from an empty one — and so errors.Is keeps working through
// the exclusion round-trip.
func TestExclusionFor_ParseErrorIsPreserved(t *testing.T) {
	parseErr := fmt.Errorf("failed to parse line 2: %w", errors.New("unexpected end of JSON input"))
	exclusion := ExclusionFor("malformed.jsonl", 0, parseErr)
	if exclusion == nil {
		t.Fatal("a parse failure must be excluded, not silently accepted")
	}
	if !errors.Is(exclusion.Err, parseErr) {
		t.Fatalf("exclusion.Err = %v, want the original parse error", exclusion.Err)
	}
	if errors.Is(exclusion.Err, ErrNoMessageEntries) {
		t.Fatal("a parse failure must not be misclassified as a benign zero-message stub")
	}
}

// TestSessionFileExclusionWarning pins the canonical warning wording. Every
// corpus-enumerating tool funnels through this one string, so its shape is the
// contract callers and tests rely on when asserting that an excluded file was
// named rather than dropped.
func TestSessionFileExclusionWarning(t *testing.T) {
	tests := []struct {
		name      string
		exclusion SessionFileExclusion
		want      string
	}{
		{
			name:      "zero-message stub",
			exclusion: SessionFileExclusion{File: "/p/stub.jsonl", Err: ErrNoMessageEntries},
			want:      "skipped session file /p/stub.jsonl: contains no message entries (empty or metadata-only session stub)",
		},
		{
			name:      "parse failure",
			exclusion: SessionFileExclusion{File: "/p/bad.jsonl", Err: errors.New("failed to parse line 3: boom")},
			want:      "skipped session file /p/bad.jsonl: failed to parse line 3: boom",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.exclusion.Warning(); got != tt.want {
				t.Fatalf("Warning() = %q, want %q", got, tt.want)
			}
		})
	}
}
