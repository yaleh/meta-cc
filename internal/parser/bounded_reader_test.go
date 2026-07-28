package parser

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestReadLineBounded_NormalLine(t *testing.T) {
	input := "hello world\n"
	r := bufio.NewReader(strings.NewReader(input))
	got, err := ReadLineBounded(r, 1024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != input {
		t.Errorf("expected %q, got %q", input, got)
	}
}

func TestReadLineBounded_MultipleLines(t *testing.T) {
	input := "first\nsecond\nthird"
	r := bufio.NewReader(strings.NewReader(input))

	got1, err1 := ReadLineBounded(r, 1024)
	if err1 != nil {
		t.Fatalf("unexpected error on line 1: %v", err1)
	}
	if string(got1) != "first\n" {
		t.Errorf("expected %q, got %q", "first\n", got1)
	}

	got2, err2 := ReadLineBounded(r, 1024)
	if err2 != nil {
		t.Fatalf("unexpected error on line 2: %v", err2)
	}
	if string(got2) != "second\n" {
		t.Errorf("expected %q, got %q", "second\n", got2)
	}

	got3, err3 := ReadLineBounded(r, 1024)
	if err3 != io.EOF {
		t.Fatalf("expected io.EOF for final unterminated line, got: %v", err3)
	}
	if string(got3) != "third" {
		t.Errorf("expected %q, got %q", "third", got3)
	}
}

func TestReadLineBounded_EmptyInput(t *testing.T) {
	r := bufio.NewReader(strings.NewReader(""))
	got, err := ReadLineBounded(r, 1024)
	if err != io.EOF {
		t.Fatalf("expected io.EOF for empty input, got: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no data, got: %q", got)
	}
}

func TestReadLineBounded_EmptyLine(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("\n"))
	got, err := ReadLineBounded(r, 1024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "\n" {
		t.Errorf("expected a bare newline, got: %q", got)
	}
}

// TestReadLineBounded_SpansInternalBufferRefills verifies that a line longer
// than bufio.Reader's own internal buffer size (which forces ReadSlice to
// return bufio.ErrBufferFull one or more times) is still assembled correctly
// as long as it stays under maxBytes.
func TestReadLineBounded_SpansInternalBufferRefills(t *testing.T) {
	long := strings.Repeat("x", 5000) // > default small buffer size below
	input := long + "\nnext\n"

	r := bufio.NewReaderSize(strings.NewReader(input), 64) // tiny internal buffer forces refills
	got, err := ReadLineBounded(r, 1<<20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != long+"\n" {
		t.Errorf("expected long line reassembled, got len=%d", len(got))
	}

	got2, err2 := ReadLineBounded(r, 1<<20)
	if err2 != nil {
		t.Fatalf("unexpected error reading second line: %v", err2)
	}
	if string(got2) != "next\n" {
		t.Errorf("expected %q, got %q", "next\n", got2)
	}
}

// TestReadLineBounded_ExceedsMaxBytes verifies the cap: a line at/above
// maxBytes is rejected with ErrLineTooLong rather than silently accepted
// unbounded — the same failure mode bufio.Scanner surfaces via ErrTooLong
// when a token exceeds its configured max buffer size.
func TestReadLineBounded_ExceedsMaxBytes(t *testing.T) {
	oversized := strings.Repeat("a", 200) // no trailing newline: keeps growing
	r := bufio.NewReaderSize(strings.NewReader(oversized), 16)

	got, err := ReadLineBounded(r, 100)
	if !errors.Is(err, ErrLineTooLong) {
		t.Fatalf("expected ErrLineTooLong, got: %v", err)
	}
	if len(got) != 100 {
		t.Errorf("expected truncated result capped at maxBytes=100, got len=%d", len(got))
	}
}

// TestReadLineBounded_ExceedsMaxBytesWithNewlineBeyondCap verifies the cap
// fires even when the oversized line does eventually contain a newline —
// the newline arriving after the cap must not save it.
func TestReadLineBounded_ExceedsMaxBytesWithNewlineBeyondCap(t *testing.T) {
	oversized := strings.Repeat("b", 300) + "\n" + "next\n"
	r := bufio.NewReaderSize(strings.NewReader(oversized), 32)

	_, err := ReadLineBounded(r, 100)
	if !errors.Is(err, ErrLineTooLong) {
		t.Fatalf("expected ErrLineTooLong, got: %v", err)
	}
}

// TestReadLineBounded_LargeButWithinCap exercises a line sized similarly to
// how client.go uses this helper (a large but legitimate multi-MB frame)
// well within a generous cap, proving large-but-valid content is not
// rejected.
func TestReadLineBounded_LargeButWithinCap(t *testing.T) {
	const maxBytes = 64 * 1024 * 1024
	payload := bytes.Repeat([]byte("z"), 5*1024*1024) // 5MB, well under 64MB cap
	input := append(append([]byte{}, payload...), '\n')

	r := bufio.NewReader(bytes.NewReader(input))
	got, err := ReadLineBounded(r, maxBytes)
	if err != nil {
		t.Fatalf("unexpected error for large-but-within-cap line: %v", err)
	}
	if len(got) != len(input) {
		t.Errorf("expected full line of len=%d, got len=%d", len(input), len(got))
	}
}
