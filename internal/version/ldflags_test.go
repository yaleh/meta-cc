package version_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestLDFLAGSCommitAndBuildTimeAreWired is a regression test for DIR-049:
// the Makefile injects commit/build-time info via `-X` linker flags when
// building ./cmd/mcp-server (see Makefile's LDFLAGS/LDFLAGS_VALUE, used by
// the `build`, `install`, and `cross-compile` targets). Those flags used to
// target a nonexistent package (github.com/yaleh/meta-cc/cmd.Commit /
// .BuildTime), which Go's linker silently no-ops on when an -X target is
// unresolvable -- no build error, no warning, just a binary that never got
// the values.
//
// This test does NOT hardcode a duplicate copy of the -X target paths --
// that would pass even if the Makefile itself were wrong, since it would
// never actually exercise the Makefile's own LDFLAGS computation. Instead
// it asks the real Makefile for its computed LDFLAGS_VALUE (via the
// `print-ldflags-value` target added alongside this test), feeds that
// verbatim to `go build -ldflags <value>`, runs the resulting binary, and
// asserts its "MCP server starting" startup log reports a commit that
// matches the real `git rev-parse --short HEAD` and is never empty/
// "unknown". If the Makefile's -X target path is ever reverted to the
// nonexistent github.com/yaleh/meta-cc/cmd package, this test fails,
// because the linker will then silently fail to resolve the symbol and
// internal/version.Commit will keep its "unknown" default.
func TestLDFLAGSCommitAndBuildTimeAreWired(t *testing.T) {
	repoRoot := findRepoRoot(t)

	// Per `go help test`: "Tests that open files within the package's
	// module ... only match future runs in which the files ... are
	// unchanged" -- i.e. go's test result cache DOES invalidate on changes
	// to files the test binary itself reads directly (tracked via os.Open/
	// os.ReadFile), but NOT on changes to files only read by a subprocess
	// this test launches (like `make` or `go build` below). Without this
	// direct read, editing the Makefile alone (no .go file touched) could
	// leave a stale cached PASS in place after the Makefile regresses --
	// silently defeating this entire regression test. This read was
	// verified empirically to fix that: `go clean -testcache` once, then
	// toggling the Makefile's LDFLAGS_VALUE back to the broken
	// github.com/yaleh/meta-cc/cmd.* path with no Go file touched still
	// forces a real re-run (not "(cached)") and fails.
	makefilePath := filepath.Join(repoRoot, "Makefile")
	makefileBytes, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("reading %s: %v", makefilePath, err)
	}
	if !strings.Contains(string(makefileBytes), "print-ldflags-value:") {
		t.Fatalf("Makefile no longer defines the print-ldflags-value target this test depends on")
	}

	ldflagsValue := makePrintLDFLAGSValue(t, repoRoot)
	if ldflagsValue == "" {
		t.Fatalf("`make print-ldflags-value` returned an empty string")
	}

	wantCommit := gitShortHEAD(t, repoRoot)

	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "meta-cc-mcp-ldflags-test")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	buildCmd := exec.Command("go", "build", "-ldflags", ldflagsValue, "-o", binPath, "./cmd/mcp-server")
	buildCmd.Dir = repoRoot
	buildCmd.Env = os.Environ()
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("go build with Makefile-derived LDFLAGS failed: %v\nldflags: %s\n%s", err, ldflagsValue, out)
	}

	// Run the built binary with immediate EOF on stdin: cmd/mcp-server's
	// main loop reads JSON-RPC requests line-by-line from stdin via a
	// bufio.Scanner and exits cleanly once it observes EOF, after logging
	// its "MCP server starting" entry (JSON, to stdout) at startup.
	runCmd := exec.Command(binPath)
	runCmd.Stdin = strings.NewReader("")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	runCmd.Stdout = &stdout
	runCmd.Stderr = &stderr
	if err := runCmd.Run(); err != nil {
		t.Fatalf("running built binary failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}

	gotCommit, gotBuildTime, found := parseStartupLog(stdout.String())
	if !found {
		t.Fatalf("did not find \"MCP server starting\" log line in binary stdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String())
	}

	if gotCommit == "" || gotCommit == "unknown" {
		t.Errorf("startup log commit = %q, must not be empty/\"unknown\" for a build produced by the real Makefile LDFLAGS -- the -X target did not resolve to a real symbol", gotCommit)
	}
	if gotCommit != wantCommit {
		t.Errorf("startup log commit = %q, want %q (git rev-parse --short HEAD) -- the Makefile-derived LDFLAGS did not embed the expected commit", gotCommit, wantCommit)
	}
	if gotBuildTime == "" || gotBuildTime == "unknown" {
		t.Errorf("startup log build_time = %q, must not be empty/\"unknown\" for a build produced by the real Makefile LDFLAGS -- the -X target did not resolve to a real symbol", gotBuildTime)
	}
}

// makePrintLDFLAGSValue shells out to `make print-ldflags-value` in
// repoRoot to obtain the exact -X assignments the real build/install/
// cross-compile targets use (see Makefile's LDFLAGS_VALUE).
func makePrintLDFLAGSValue(t *testing.T, repoRoot string) string {
	t.Helper()
	cmd := exec.Command("make", "-s", "print-ldflags-value")
	cmd.Dir = repoRoot
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("`make print-ldflags-value` failed: %v\n%s", err, out)
	}
	return strings.TrimSpace(string(out))
}

// gitShortHEAD returns `git rev-parse --short HEAD` for repoRoot, mirroring
// exactly how the Makefile's COMMIT variable is computed.
func gitShortHEAD(t *testing.T, repoRoot string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse --short HEAD failed: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// parseStartupLog scans newline-delimited JSON log output for the
// "MCP server starting" entry and extracts its commit/build_time fields.
func parseStartupLog(output string) (commit, buildTime string, found bool) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if msg, _ := entry["msg"].(string); msg == "MCP server starting" {
			c, _ := entry["commit"].(string)
			b, _ := entry["build_time"].(string)
			return c, b, true
		}
	}
	return "", "", false
}

// findRepoRoot walks up from the current working directory to locate the
// module root (identified by go.mod), so `make`/`go build` invocations run
// against the real repository regardless of where `go test` is invoked
// from.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repo root (go.mod) starting from %s", dir)
		}
		dir = parent
	}
}
