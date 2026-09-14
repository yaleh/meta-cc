package provider

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/yaleh/meta-cc/internal/locator"
	"github.com/yaleh/meta-cc/internal/types"
)

// FileKind selects the entry rule a health probe applies to a session file.
// The corpus is heterogeneous — a Claude project directory holds
// types.SessionEntry JSONL, a Codex rollout holds a different line schema
// entirely — so "did this file contribute anything?" cannot be answered by one
// hard-coded rule.
type FileKind int

const (
	// KindClaudeSession is a Claude Code session log: JSONL lines that decode
	// into types.SessionEntry, where user/assistant entries are the messages a
	// query can actually see. This is the rule claude.Provider.ListSessions
	// applies (parseClaudeEntries), so the probe and the listing reach the same
	// verdict about the same file.
	KindClaudeSession FileKind = iota
	// KindRaw is any other provider's raw session file. Its line schema is not
	// this package's business, so only readability is judged: a readable file
	// with content counts as healthy, and the probe never claims a Codex
	// rollout is "empty of messages" merely because it is not Claude-shaped.
	KindRaw
)

// FileHealth is the per-file health probe result. It is the raw material for
// both Stage 1 discovery tools (get_session_directory and
// inspect_session_files), which is the point: before this, each enumerated
// files without saying whether they were usable, so a corrupt file only
// announced itself later as a crash inside a downstream query.
//
// The JSON shape is the caller-visible contract:
//
//	{"file": "...", "bytes": 0, "entries": 0, "parseable": true, "empty": true}
//	{"file": "...", "bytes": 147, "entries": 0, "parseable": false, "empty": false,
//	 "error": "1 of 1 non-empty lines are not valid JSON (truncated or corrupt file)"}
type FileHealth struct {
	File      string `json:"file"`
	Bytes     int64  `json:"bytes"`
	Entries   int    `json:"entries"`
	Parseable bool   `json:"parseable"`
	Empty     bool   `json:"empty"`
	Error     string `json:"error,omitempty"`

	// nonBlank, ioErr and defect are the inputs to the shared malformed
	// verdict (MalformedFiles) and to Reason. They stay unexported because they
	// are probe bookkeeping, not part of the reported shape — a caller that
	// wants the explanation reads MalformedFile.Reason.
	nonBlank int
	// ioErr is "this file could not be read", which is the cause the DIR-094
	// shared rule (locator.ExclusionFor) decides on.
	ioErr error
	// defect is "this file was read, but some of its lines are not JSON".
	// It is deliberately NOT an ioErr: the Claude listing parser tolerates
	// malformed lines, so a file that still yields message entries is not
	// excluded from the corpus and must not be reported as if it were. The
	// defect is still worth naming — that is the difference between a
	// discoverable problem and a silent one.
	defect error
}

// MalformedFile names one enumerated session file that contributed nothing
// usable, and why. This is the aggregate both discovery tools publish as
// "malformed_files" — empty for a healthy corpus, never nil.
type MalformedFile struct {
	File   string `json:"file"`
	Reason string `json:"reason"`
}

// ProbeFileHealth reads one session file and reports what a caller can expect
// from it. It never returns an error: a file that cannot be read is a health
// *result* ("parseable": false plus the cause), not a failure of the probe.
// That inversion is the point of this task — parse errors become data.
func ProbeFileHealth(file string, kind FileKind) FileHealth {
	return probeFile(file, kind, false)
}

// probeFile is ProbeFileHealth with an optional early exit. stopAtFirstEntry
// is used by MalformedFiles, which only needs the verdict: once a file has
// yielded one entry it can never be malformed, so the rest of it (possibly
// megabytes) does not need reading. Files that ARE malformed yield no entries
// and are therefore always scanned in full, so the reported reason stays exact
// on exactly the files where a reason is reported.
func probeFile(file string, kind FileKind, stopAtFirstEntry bool) FileHealth {
	h := FileHealth{File: file, Parseable: true}

	info, err := os.Stat(file)
	if err != nil {
		return h.fail(fmt.Errorf("cannot stat file: %w", err))
	}
	h.Bytes = info.Size()

	f, err := os.Open(file)
	if err != nil {
		return h.fail(fmt.Errorf("cannot open file: %w", err))
	}
	defer f.Close()

	badLines := 0
	reader := bufio.NewReader(f)
	for {
		line, readErr := reader.ReadBytes('\n')
		if trimmed := bytes.TrimSpace(line); len(trimmed) > 0 {
			h.nonBlank++
			switch kind {
			case KindClaudeSession:
				if !isJSONObject(trimmed) {
					badLines++
				} else {
					var entry types.SessionEntry
					if jsonErr := json.Unmarshal(trimmed, &entry); jsonErr == nil && entry.IsMessage() {
						h.Entries++
					}
				}
			default:
				h.Entries++
			}
			if stopAtFirstEntry && h.Entries > 0 {
				return h
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return h.fail(fmt.Errorf("error reading file: %w", readErr))
		}
	}

	// "Empty" is literal — zero bytes, or nothing but blank lines. A
	// metadata-only stub is NOT empty (it has content) even though it yields
	// zero entries; both are malformed, but they are not the same defect and
	// the reasons say so.
	h.Empty = h.Bytes == 0 || h.nonBlank == 0

	if badLines > 0 {
		h.defect = fmt.Errorf("%d of %d non-empty lines are not valid JSON (truncated or corrupt file)", badLines, h.nonBlank)
		h.Parseable = false
		h.Error = h.defect.Error()
	}
	return h
}

// Readable reports whether the file could be read at all, regardless of whether
// its contents are well-formed. It is the gate for "can I scan this file for
// descriptive detail?" — a partially corrupt file is still worth inspecting,
// while an unreadable one has nothing to scan. Parseable is the stricter claim
// (readable AND every line valid JSON); the two are separate because a file can
// be readable, defective, and still contribute entries — in which case it is
// reported as defective but must NOT be excluded from the corpus.
func (h FileHealth) Readable() bool {
	return h.ioErr == nil
}

// fail records a probe-level problem and marks the file unparseable.
func (h FileHealth) fail(err error) FileHealth {
	h.Parseable = false
	h.ioErr = err
	h.Error = err.Error()
	return h
}

// isJSONObject reports whether a line is a JSON object, the only shape a
// session-log line can take. This is deliberately distinct from "decodes into
// types.SessionEntry": a line can be perfectly valid JSON and still not be a
// message (mode changes, summaries, permission records), and calling those
// "not valid JSON" would misname the defect.
func isJSONObject(line []byte) bool {
	var raw map[string]json.RawMessage
	return json.Unmarshal(line, &raw) == nil
}

// MalformedFiles probes every path and returns the unhealthy subset, in input
// order. The verdict comes from locator.ExclusionFor — the DIR-094 rule that
// already decides this for query_sessions and the analysis loaders — so every
// corpus-enumerating tool keeps reaching the same conclusion about the same
// file instead of each inventing its own.
func MalformedFiles(files []string, kind FileKind) []MalformedFile {
	health := make([]FileHealth, 0, len(files))
	for _, file := range files {
		health = append(health, probeFile(file, kind, true))
	}
	return MalformedFilesFromHealth(health)
}

// MalformedFilesFromHealth applies the shared exclusion rule to already-probed
// health, for callers that also report the per-file health
// (inspect_session_files does both).
func MalformedFilesFromHealth(health []FileHealth) []MalformedFile {
	out := make([]MalformedFile, 0)
	for _, h := range health {
		exclusion := locator.ExclusionFor(h.File, h.Entries, h.ioErr)
		if exclusion == nil {
			continue
		}
		out = append(out, MalformedFile{File: h.File, Reason: reasonFor(h, exclusion)})
	}
	return out
}

// reasonFor prefers the specific over the generic. locator.ExclusionFor can
// only distinguish "unreadable" from "contributed nothing", so a truncated
// file would otherwise be explained with the empty/metadata-only-stub clause —
// true but not what is wrong with it. When the probe saw the actual defect, it
// names it.
func reasonFor(h FileHealth, exclusion *locator.SessionFileExclusion) string {
	if h.ioErr == nil && h.defect != nil {
		return h.defect.Error()
	}
	return exclusion.Reason()
}
