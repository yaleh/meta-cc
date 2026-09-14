package locator

import (
	"errors"
	"fmt"
)

// ErrNoMessageEntries marks a session file that is readable and well-formed
// but yields no user/assistant message entries — either literally empty, or
// the "stub" Claude Code writes at session start (mode/permission-mode/system
// metadata only) before any turn is exchanged. It is a benign, expected state
// with nothing queryable to report, but it is NOT the same thing as "this file
// does not exist": a corpus enumeration that drops such a file is discarding
// data and must say so (see SessionFileExclusion).
//
// DIR-094: this sentinel lives here, not in a single provider, because the
// rule it encodes is a property of the session corpus rather than of any one
// reader of it. Every corpus-enumerating path — claude.Provider.ListSessions
// (query_sessions), analysis.Service.loadData (get_timeline and the five
// analysis tools) — must reach the same verdict on the same file; before this
// it reached two opposite ones, hard-failing a whole listing on the one hand
// and silently `continue`-ing on the other.
var ErrNoMessageEntries = errors.New("no message entries")

// emptyFileReason is the human-readable clause attached to
// ErrNoMessageEntries when a file parses cleanly but contributes nothing.
// Kept distinct from the bare sentinel text so a warning reads as an
// explanation rather than a bare "no message entries".
const emptyFileReason = "contains no message entries (empty or metadata-only session stub)"

// SessionFileExclusion records that one enumerated session file contributed
// nothing to a corpus-wide result, and why.
//
// The contract for every corpus-enumerating caller is: exclude the file from
// the result *and* name it in a warning. Neither half alone is acceptable —
// returning a whole-batch error because of one bad file erases every other
// session in the corpus (the 2026-07-30 `query_sessions` failure this task
// fixes), and skipping the file without a trace hides data loss from the
// caller (the DIR-018 contract).
type SessionFileExclusion struct {
	// File is the on-disk path of the excluded session file, as enumerated.
	File string
	// Err is the underlying cause: ErrNoMessageEntries for a zero-message
	// stub, or the parse/I-O error for a file that could not be read.
	// It is wrapped by callers that need errors.Is to keep working across
	// the API boundary (see claude.Provider.sessionFromFile).
	Err error
}

// Reason returns the human-readable clause explaining the exclusion, with no
// file path and no "skipped" prefix, so callers can compose their own phrasing.
func (e SessionFileExclusion) Reason() string {
	if errors.Is(e.Err, ErrNoMessageEntries) {
		return emptyFileReason
	}
	return e.Err.Error()
}

// Warning renders the exclusion in the canonical DIR-018 form
// "skipped session file <path>: <reason>". This one function is the shared
// vocabulary that makes the tolerance behavior uniform across tools instead of
// each path inventing (or omitting) its own wording.
func (e SessionFileExclusion) Warning() string {
	return fmt.Sprintf("skipped session file %s: %s", e.File, e.Reason())
}

// ExclusionFor is the shared "did this enumerated session file contribute
// anything?" rule. It takes the outcome of whichever parse step the calling
// path uses — which deliberately differ (claude's listing parser tolerates
// malformed lines, the analysis parser reports the offending line number) —
// and returns:
//
//   - nil when the file contributed at least one message entry;
//   - a *SessionFileExclusion wrapping parseErr when the file could not be
//     read or parsed;
//   - a *SessionFileExclusion wrapping ErrNoMessageEntries when the file
//     parsed cleanly but yielded nothing (entryCount == 0).
//
// Callers must not treat a non-nil return as fatal: the whole point is that a
// single file's problem never fails the batch. Exclude, warn, continue.
func ExclusionFor(file string, entryCount int, parseErr error) *SessionFileExclusion {
	switch {
	case parseErr != nil:
		return &SessionFileExclusion{File: file, Err: parseErr}
	case entryCount == 0:
		return &SessionFileExclusion{File: file, Err: ErrNoMessageEntries}
	default:
		return nil
	}
}
