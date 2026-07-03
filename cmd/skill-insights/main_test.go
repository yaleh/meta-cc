package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSplitRoots(t *testing.T) {
	got := splitRoots(" /tmp/a.jsonl, ,/tmp/b ")
	want := []string{"/tmp/a.jsonl", "/tmp/b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
}

func TestSplitRootsEmpty(t *testing.T) {
	if got := splitRoots(" , "); len(got) != 0 {
		t.Fatalf("expected empty roots, got %#v", got)
	}
}

func TestRunWritesJSONToStdout(t *testing.T) {
	dir := t.TempDir()
	sessionPath := filepath.Join(dir, "session.jsonl")
	if err := os.WriteFile(sessionPath, []byte(`{"type":"user","timestamp":"2026-07-03T00:00:00Z","message":{"role":"user","content":"hello"}}`+"\n"), 0644); err != nil {
		t.Fatalf("write session: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := run([]string{"-roots", sessionPath, "-format", "json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"files_scanned": 1`) {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestRunWritesMarkdownToFile(t *testing.T) {
	dir := t.TempDir()
	sessionPath := filepath.Join(dir, "session.jsonl")
	outPath := filepath.Join(dir, "reports", "skill-insights.md")
	if err := os.WriteFile(sessionPath, []byte(`{"type":"user","timestamp":"2026-07-03T00:00:00Z","message":{"role":"user","content":"hello"}}`+"\n"), 0644); err != nil {
		t.Fatalf("write session: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := run([]string{"-roots", sessionPath, "-format", "markdown", "-out", outPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", code, stderr.String())
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(string(data), "# Skill Insights MVP") {
		t.Fatalf("unexpected markdown: %s", string(data))
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout when writing file, got %s", stdout.String())
	}
}

func TestRunRejectsUnsupportedFormat(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-roots", "/tmp/none", "-format", "xml"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unsupported format") {
		t.Fatalf("expected unsupported format error, got %s", stderr.String())
	}
}

func TestRunRejectsBadFlags(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-unknown"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Fatalf("expected flag error, got %s", stderr.String())
	}
}

func TestRunHelpExitsSuccessfully(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-h"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "Usage of skill-insights") {
		t.Fatalf("expected help output, got %s", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected help on stderr only, got stdout=%s", stdout.String())
	}
}
