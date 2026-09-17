package locator

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DIR-094: the shared corpus-tolerance helper. These tests pin the contract
// every enumerating path relies on — an empty report stays invisible on the
// wire (nil slices, so `omitempty` drops the fields), and every recorded
// exclusion names both the file and the reason.

func TestSkipReport_EmptyByDefault(t *testing.T) {
	var report SkipReport
	assert.True(t, report.Empty())
	assert.Nil(t, report.Warnings(), "a clean corpus must yield nil warnings so omitempty drops the field")
	assert.Nil(t, report.Paths(), "a clean corpus must yield nil paths so omitempty drops the field")
}

func TestSkipReport_SkipNamesFileAndReason(t *testing.T) {
	var report SkipReport
	report.Skip("/corpus/bad.jsonl", errors.New("boom"))

	require.False(t, report.Empty())
	require.Len(t, report.Warnings(), 1)
	assert.Contains(t, report.Warnings()[0], "/corpus/bad.jsonl")
	assert.Contains(t, report.Warnings()[0], "boom")
	assert.Equal(t, []string{"/corpus/bad.jsonl"}, report.Paths())
}

func TestSkipReport_SkipReason(t *testing.T) {
	var report SkipReport
	report.SkipReason("/corpus/stub.jsonl", "no message entries")

	require.Len(t, report.Warnings(), 1)
	assert.Contains(t, report.Warnings()[0], "/corpus/stub.jsonl")
	assert.Contains(t, report.Warnings()[0], "no message entries")
	assert.Equal(t, []string{"/corpus/stub.jsonl"}, report.Paths())
}

func TestSkipReport_SkipNilErrorStillRecords(t *testing.T) {
	var report SkipReport
	report.Skip("/corpus/odd.jsonl", nil)
	require.Len(t, report.Warnings(), 1)
	assert.Contains(t, report.Warnings()[0], "/corpus/odd.jsonl")
}

// AdoptWarning preserves a downstream layer's own wording verbatim and, since
// that layer reports a session ID rather than a corpus path, contributes no
// entry to Paths() — the exclusion is still visible in Warnings.
func TestSkipReport_AdoptWarningKeepsTextAndAddsNoPath(t *testing.T) {
	var report SkipReport
	report.AdoptWarning("provider codex session abc: failed to load turns, skipped: nope")
	report.AdoptWarning("")

	require.Equal(t, []string{"provider codex session abc: failed to load turns, skipped: nope"}, report.Warnings())
	assert.Nil(t, report.Paths())
	assert.False(t, report.Empty())
}

// Accumulation order is the order exclusions were encountered, so a reader can
// map the warnings back onto the enumerated corpus.
func TestSkipReport_AccumulatesInOrder(t *testing.T) {
	var report SkipReport
	report.SkipReason("/corpus/a.jsonl", "first")
	report.SkipReason("/corpus/b.jsonl", "second")

	assert.Equal(t, []string{"/corpus/a.jsonl", "/corpus/b.jsonl"}, report.Paths())
	require.Len(t, report.Warnings(), 2)
	assert.Contains(t, report.Warnings()[0], "a.jsonl")
	assert.Contains(t, report.Warnings()[1], "b.jsonl")
}

// Accessors hand back copies: a caller appending to the returned slice must not
// mutate the report's own state.
func TestSkipReport_AccessorsReturnCopies(t *testing.T) {
	var report SkipReport
	report.SkipReason("/corpus/a.jsonl", "first")

	paths := report.Paths()
	paths[0] = "/mutated"
	assert.Equal(t, []string{"/corpus/a.jsonl"}, report.Paths())
}

// A nil *SkipReport is usable as a no-op sink, so call sites that do not need
// reporting (or are mid-refactor) cannot panic.
func TestSkipReport_NilReceiverIsNoOp(t *testing.T) {
	var report *SkipReport
	require.NotPanics(t, func() {
		report.Skip("/corpus/a.jsonl", errors.New("boom"))
		report.AdoptWarning("w")
		report.SkipReason("/corpus/b.jsonl", "r")
	})
	assert.True(t, report.Empty())
	assert.Nil(t, report.Warnings())
}
