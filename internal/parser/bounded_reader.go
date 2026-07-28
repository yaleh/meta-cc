package parser

import (
	"bufio"
	"errors"
)

// ErrLineTooLong is returned by ReadLineBounded when a single line grows to
// or beyond the caller-supplied maxBytes cap before a newline is found. It
// mirrors the role of bufio.Scanner's ErrTooLong for callers that used to
// configure scanner.Buffer(initial, max) and relied on Scan() failing (not
// silently truncating or growing without bound) once a token crossed that
// max size.
var ErrLineTooLong = errors.New("parser: line exceeds maxBytes limit")

// ReadLineBounded reads one line (including the trailing '\n', if present)
// from r, capping the total accumulated size of that line at maxBytes.
//
// It is the line-length-safe helper for non-JSONL, line-oriented read loops
// that used to be hand-rolled as a raw `bufio.NewReader(r)` +
// `reader.ReadBytes('\n')` loop with no cap — the sibling to
// ReadLineFiltered, which instead is specialized for Claude/Codex session
// JSONL content (base64 image-data stripping). Neither tool fits every
// caller: ReadLineFiltered's image/data-field filtering is wrong for
// arbitrary non-JSONL text (e.g. scanning source files for TODO markers, or
// framing a live JSON-RPC subprocess stream), while an unbounded
// ReadBytes('\n') loop lets a malformed or hostile input grow memory without
// limit. ReadLineBounded intentionally carries no framing/content opinions
// beyond the length cap, to fill that gap.
//
// Return semantics mirror bufio.Reader.ReadBytes: on success the returned
// line includes the trailing '\n'. If the underlying reader reaches EOF with
// a trailing partial line (no '\n'), that data is returned along with
// io.EOF, the same way ReadBytes and bufio.Scanner's final (unterminated)
// token both surface a caller's last, newline-less chunk of data. If the
// line grows to or beyond maxBytes before a '\n' is found, ReadLineBounded
// stops accumulating and returns ErrLineTooLong; the bytes returned in that
// case are only the first maxBytes read and must not be treated as a
// complete line (this matches bufio.Scanner discarding an over-long token
// entirely rather than returning a truncated one).
func ReadLineBounded(r *bufio.Reader, maxBytes int) ([]byte, error) {
	var line []byte
	for {
		chunk, err := r.ReadSlice('\n')
		line = append(line, chunk...)
		if len(line) > maxBytes {
			return line[:maxBytes], ErrLineTooLong
		}
		if err == nil {
			return line, nil
		}
		if errors.Is(err, bufio.ErrBufferFull) {
			// No newline yet within the current internal buffer; keep
			// accumulating and read more.
			continue
		}
		// A genuine terminal condition: io.EOF (possibly with a trailing,
		// newline-less partial line already appended above) or an
		// underlying read error.
		return line, err
	}
}
