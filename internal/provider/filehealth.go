package provider

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/yaleh/meta-cc/internal/parser"
)

// Session-file health probing (DIR-098).
//
// The 2026-07-30 dogfooding failure: a corrupt transcript
// (8eda8f4e-2c74-4176-ba6b-8c45e890df42.jsonl, 0 bytes — "no Claude entries")
// sat undetected in the project session directory until it hard-crashed a
// query, twice. DIR-018/DIR-094 made the *query* paths tolerant of such a file,
// but nothing let a caller ASK whether the corpus it was about to query was
// healthy: the discovery tools enumerated files without reporting their health,
// so corruption was always learned at crash time.
//
// This probe is that missing question. It turns "is this file usable?" into
// data — one struct per file, no error return — so the two discovery tools can
// report corpus health in the same call that lists the corpus.

// FileHealth is the structural health of one session (JSONL) file, computed by
// reading the file exactly once.
//
// The vocabulary intentionally mirrors the reader that consumes these files
// (internal/query/engine.readJSONLFile): a line is a record only when it is a
// well-formed JSON object, and the first unusable line is reported by its
// 1-based index among non-empty lines. A file this probe calls parseable is a
// file that reader loads without returning an error.
//
// `Reason` is the malformed marker: it is non-empty if and only if the file is
// unusable, so a caller can filter on one field instead of re-deriving the
// verdict from four others. `Error` carries the underlying I/O error text when
// the problem was at the filesystem layer rather than the record layer.
type FileHealth struct {
	// Path is the file the probe was pointed at, echoed so a health list is
	// self-describing and can be correlated with the caller's own path list.
	Path string `json:"path"`
	// Bytes is the file size from stat, counted even for a file that cannot be
	// read, because "0 bytes" is itself the diagnosis.
	Bytes int64 `json:"bytes"`
	// Entries is the number of non-empty lines. For an empty file it is 0,
	// which is the signal a loader that only checks for errors cannot see.
	Entries int `json:"entries"`
	// Parseable is true only when the file yielded at least one record, every
	// record parsed, and the file could be read — i.e. it can actually
	// contribute records to a query. An empty file is not parseable: it parses
	// without error but contributes nothing, which is exactly the shape that
	// stayed silent before this task.
	Parseable bool `json:"parseable"`
	// Empty is true when the file holds no records at all (0 bytes, or
	// whitespace only).
	Empty bool `json:"empty"`
	// ParseErrors counts non-empty lines that are not JSON objects.
	ParseErrors int `json:"parse_errors,omitempty"`
	// Error is the filesystem-level failure (stat/open/read, or a path that is
	// not a regular file), empty when the file was read to the end. An empty
	// Error is the "safe to hand to the session readers" test.
	Error string `json:"error,omitempty"`
	// Reason is non-empty iff the file is malformed, and names why.
	Reason string `json:"reason,omitempty"`
}

// Malformed reports whether the file is unusable and therefore worth naming in
// a health report.
func (h FileHealth) Malformed() bool { return h.Reason != "" }

// MalformedFile is the compact {file, reason} aggregate entry the discovery
// tools expose. It is deliberately a pair rather than the full FileHealth: an
// aggregate over a whole project corpus has to stay bounded, and a caller that
// wants per-file detail can ask inspect_session_files for it.
type MalformedFile struct {
	File   string `json:"file"`
	Reason string `json:"reason"`
}

// ProbeFileHealth reads one session file and reports its health. It never
// returns an error: an unreadable file is a health outcome, not a failure of
// the caller's request.
func ProbeFileHealth(path string) FileHealth {
	health := FileHealth{Path: path}

	info, err := os.Stat(path)
	if err != nil {
		return health.unusable(fmt.Sprintf("cannot stat file: %v", err), err)
	}
	health.Bytes = info.Size()

	if info.IsDir() {
		// Reported as an Error rather than a bare reason: `Error == ""` is the
		// "safe to hand to the session readers" test, and a directory would
		// otherwise pass it and fail deep in the reader with EISDIR.
		return health.unusable("not a session file: path is a directory",
			fmt.Errorf("%s is a directory, not a session file", path))
	}

	file, err := os.Open(path)
	if err != nil {
		return health.unusable(fmt.Sprintf("cannot open file: %v", err), err)
	}
	defer file.Close()

	// Read with the same line reader the query paths use, so the two cannot
	// disagree about where a record ends (oversized lines, image payloads).
	reader := bufio.NewReader(file)
	firstBadEntry := 0
	for {
		line, _, readErr := parser.ReadLineFiltered(reader, parser.StrategyDefault)
		trimmed := bytes.TrimSpace(bytes.TrimRight(line, "\r\n"))

		if len(trimmed) > 0 {
			health.Entries++
			if !isSessionRecord(trimmed) {
				health.ParseErrors++
				if firstBadEntry == 0 {
					firstBadEntry = health.Entries
				}
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return health.unusable(
				fmt.Sprintf("read error after %d record(s): %v", health.Entries, readErr), readErr)
		}
	}

	health.Parseable = health.Entries > 0 && health.ParseErrors == 0
	switch {
	case health.Entries == 0:
		health.Empty = true
		health.Reason = fmt.Sprintf("file is empty (%d bytes, 0 records)", health.Bytes)
	case health.ParseErrors > 0:
		health.Reason = fmt.Sprintf("invalid JSON at line %d (%d of %d records are not session objects)",
			firstBadEntry, health.ParseErrors, health.Entries)
	}
	return health
}

// ProbeFilesHealth probes each path in order, so the result lines up with the
// caller's own file list.
func ProbeFilesHealth(paths []string) []FileHealth {
	health := make([]FileHealth, 0, len(paths))
	for _, path := range paths {
		health = append(health, ProbeFileHealth(path))
	}
	return health
}

// MalformedFiles reduces probed health to the {file, reason} aggregates to
// report. The result is never nil: a healthy corpus yields an empty list, so
// "no malformed files" and "this tool does not report file health" can never
// look alike to a caller.
func MalformedFiles(health []FileHealth) []MalformedFile {
	malformed := make([]MalformedFile, 0, len(health))
	for _, h := range health {
		if !h.Malformed() {
			continue
		}
		malformed = append(malformed, MalformedFile{File: h.Path, Reason: h.Reason})
	}
	return malformed
}

// unusable marks a health record as malformed and records the cause. The
// FileHealth is returned by value, so the caller always receives a fully
// populated record rather than a partially written one.
func (h FileHealth) unusable(reason string, cause error) FileHealth {
	h.Parseable = false
	h.Reason = reason
	if cause != nil {
		h.Error = cause.Error()
	}
	return h
}

// isSessionRecord reports whether one JSONL line is a session record, i.e. a
// well-formed JSON *object*.
//
// The object check is not redundant with json.Valid: valid JSON whose top level
// is an array or scalar (the DIR-099 "wrong-shape" corruption) unmarshals into
// map[string]interface{} with an error, so a valid-but-not-an-object line is
// exactly as unusable to the query paths as a truncated one — and reporting it
// as healthy would reproduce the silent-loss bug this probe exists to end.
func isSessionRecord(line []byte) bool {
	if !json.Valid(line) {
		return false
	}
	for _, b := range line {
		switch b {
		case ' ', '\t', '\r', '\n':
			continue
		}
		return b == '{'
	}
	return false
}
