package provider

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const healthyClaudeLines = `{"type":"user","timestamp":"2026-01-01T00:00:00.000Z","message":{"role":"user","content":[{"type":"text","text":"hello"}]}}
{"type":"assistant","timestamp":"2026-01-01T00:00:01.000Z","message":{"role":"assistant","content":[{"type":"text","text":"hi"}]}}
`

func writeFixture(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("cannot write fixture %s: %v", name, err)
	}
	return path
}

// TestProbeFileHealth_ClassifiesEverySessionFileShape pins the four shapes a
// Claude session file can take, because the whole point of the probe is that
// "zero entries" and "unreadable" and "truncated" are different reports, not
// one indistinguishable silence.
func TestProbeFileHealth_ClassifiesEverySessionFileShape(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		wantEntries   int
		wantParseable bool
		wantEmpty     bool
		wantMalformed bool
	}{
		{
			name:          "healthy exchange",
			content:       healthyClaudeLines,
			wantEntries:   2,
			wantParseable: true,
			wantMalformed: false,
		},
		{
			// The 8eda8f4e shape.
			name:          "empty file",
			content:       "",
			wantEntries:   0,
			wantParseable: true,
			wantEmpty:     true,
			wantMalformed: true,
		},
		{
			name:          "blank lines only",
			content:       "\n\n   \n",
			wantEntries:   0,
			wantParseable: true,
			wantEmpty:     true,
			wantMalformed: true,
		},
		{
			// Metadata-only session-start stub: valid JSON, no messages.
			name:          "metadata-only stub",
			content:       `{"type":"mode","mode":"default"}` + "\n" + `{"type":"permission-mode","permissionMode":"default"}` + "\n",
			wantEntries:   0,
			wantParseable: true,
			wantEmpty:     false,
			wantMalformed: true,
		},
		{
			name:          "truncated mid-JSON",
			content:       `{"type":"user","message":{"role":"user","content":[{"type":"text","text":"interrupted`,
			wantEntries:   0,
			wantParseable: false,
			wantEmpty:     false,
			wantMalformed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeFixture(t, "s.jsonl", tt.content)
			h := ProbeFileHealth(path, KindClaudeSession)

			if h.File != path {
				t.Errorf("File = %q, want %q", h.File, path)
			}
			if h.Bytes != int64(len(tt.content)) {
				t.Errorf("Bytes = %d, want %d", h.Bytes, len(tt.content))
			}
			if h.Entries != tt.wantEntries {
				t.Errorf("Entries = %d, want %d", h.Entries, tt.wantEntries)
			}
			if h.Parseable != tt.wantParseable {
				t.Errorf("Parseable = %v, want %v", h.Parseable, tt.wantParseable)
			}
			if h.Empty != tt.wantEmpty {
				t.Errorf("Empty = %v, want %v", h.Empty, tt.wantEmpty)
			}

			malformed := MalformedFiles([]string{path}, KindClaudeSession)
			if got := len(malformed) == 1; got != tt.wantMalformed {
				t.Fatalf("malformed = %v, want %v (health %#v)", got, tt.wantMalformed, h)
			}
			if tt.wantMalformed && malformed[0].Reason == "" {
				t.Error("a malformed file must carry a non-empty reason")
			}
			// Error names something that went WRONG while reading. An empty or
			// metadata-only file had nothing go wrong — it simply contributes
			// nothing, and the reason (not Error) is what says so.
			if !tt.wantParseable && h.Error == "" {
				t.Error("parseable=false must be explained by a non-empty Error")
			}
			if tt.wantParseable && h.Error != "" {
				t.Errorf("a fully parseable file must carry no Error, got %q", h.Error)
			}
		})
	}
}

// TestProbeFileHealth_UnreadableFileIsDataNotAnError is the inversion this task
// is about: a file that cannot be opened is a health *result*, so an
// enumeration that hits one can report it instead of aborting.
func TestProbeFileHealth_UnreadableFileIsDataNotAnError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.jsonl")
	h := ProbeFileHealth(path, KindClaudeSession)

	if h.Parseable {
		t.Error("a missing file must not be reported as parseable")
	}
	if h.Readable() {
		t.Error("a missing file must not be reported as readable")
	}
	if h.Error == "" {
		t.Error("a missing file must carry an error string naming the cause")
	}

	malformed := MalformedFiles([]string{path}, KindClaudeSession)
	if len(malformed) != 1 || malformed[0].File != path {
		t.Fatalf("expected the missing file named as malformed, got %#v", malformed)
	}
	if malformed[0].Reason == "" {
		t.Error("expected a reason for the missing file")
	}
}

// TestProbeFileHealth_RawKindDoesNotApplyTheClaudeMessageRule guards the other
// provider's corpus: a Codex rollout is not Claude-shaped, so judging it by
// "contains user/assistant entries" would report every healthy rollout as
// malformed.
func TestProbeFileHealth_RawKindDoesNotApplyTheClaudeMessageRule(t *testing.T) {
	rollout := `{"timestamp":"2026-01-01T00:00:00.000Z","type":"session_meta","payload":{"id":"abc"}}` + "\n" +
		`{"timestamp":"2026-01-01T00:00:01.000Z","type":"response_item","payload":{"type":"message"}}` + "\n"
	path := writeFixture(t, "rollout.jsonl", rollout)

	claudeVerdict := MalformedFiles([]string{path}, KindClaudeSession)
	if len(claudeVerdict) != 1 {
		t.Fatalf("precondition: a Claude-shaped probe should see no message entries here, got %#v", claudeVerdict)
	}

	rawVerdict := MalformedFiles([]string{path}, KindRaw)
	if len(rawVerdict) != 0 {
		t.Errorf("a readable Codex rollout with content is not malformed, got %#v", rawVerdict)
	}
}

// TestProbeFileHealth_DefectiveButContributingIsNamedYetNotExcluded pins the
// two-judgment split: a file with a corrupt line that still yields messages is
// reported as defective (so the problem is visible) but is NOT excluded (so the
// corpus does not lose a session that the listing parser accepts).
func TestProbeFileHealth_DefectiveButContributingIsNamedYetNotExcluded(t *testing.T) {
	content := `{"type":"user","timestamp":"2026-01-01T00:00:00.000Z","message":{"role":"user","content":[]}}` + "\n" +
		`this line was truncated mid-write` + "\n" +
		`{"type":"assistant","timestamp":"2026-01-01T00:00:01.000Z","message":{"role":"assistant","content":[]}}` + "\n"
	path := writeFixture(t, "partly-corrupt.jsonl", content)

	h := ProbeFileHealth(path, KindClaudeSession)
	if h.Entries != 2 {
		t.Fatalf("Entries = %d, want 2", h.Entries)
	}
	if h.Parseable {
		t.Error("a file with an unparseable line must not be reported as fully parseable")
	}
	if !strings.Contains(h.Error, "1 of 3") {
		t.Errorf("Error should name how many lines are unparseable, got %q", h.Error)
	}
	if malformed := MalformedFiles([]string{path}, KindClaudeSession); len(malformed) != 0 {
		t.Errorf("a file that contributed entries must not be excluded from the corpus, got %#v", malformed)
	}
}

// TestMalformedFiles_HealthyCorpusIsAnEmptyNonNilList keeps the reported shape
// stable: [] rather than null, so a caller can always take its length.
func TestMalformedFiles_HealthyCorpusIsAnEmptyNonNilList(t *testing.T) {
	healthy := writeFixture(t, "healthy.jsonl", healthyClaudeLines)

	malformed := MalformedFiles([]string{healthy}, KindClaudeSession)
	if malformed == nil {
		t.Fatal("MalformedFiles must never return nil — a healthy corpus reports []")
	}
	if len(malformed) != 0 {
		t.Errorf("healthy corpus produced %#v", malformed)
	}
}
