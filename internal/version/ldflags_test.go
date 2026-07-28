package version_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// TestLDFLAGSCommitAndBuildTimeAreWired is a regression test for DIR-049:
// the Makefile injects commit/build-time info via `-X` linker flags when
// building ./cmd/mcp-server (see Makefile's LDFLAGS/LDFLAGS_VALUE, used by
// the `build`, `install`, and `cross-compile` targets). Those flags used to
// target a nonexistent package (the repo-root cmd package's Version / Commit
// / BuildTime vars), which Go's linker silently no-ops on when an -X target
// is unresolvable -- no build error, no warning, just a binary that never
// got the values.
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
// nonexistent repo-root cmd package, this test fails, because the linker
// will then silently fail to resolve the symbol and internal/version.Commit
// will keep its "unknown" default.
//
// Its companion TestNoHandDuplicatedDeadPathLDFLAGS below guards the other
// half of the same defect class (DIR-052): hand-copied duplicates of the
// broken -X string that lived OUTSIDE the Makefile (release.yml,
// validate-artifacts.sh) and kept shipping dead LDFLAGS even after the
// Makefile itself was fixed.
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
	// toggling the Makefile's LDFLAGS_VALUE back to the broken repo-root
	// cmd-package path with no Go file touched still forces a real re-run
	// (not "(cached)") and fails.
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

// TestNoHandDuplicatedDeadPathLDFLAGS is the DIR-052 drift guard: DIR-049
// fixed the Makefile's -X targets to point at the real internal/version
// symbols, but hand-copied duplicates of the OLD broken string survived in
// .github/workflows/release.yml and tests/validation/validate-artifacts.sh,
// so the release pipeline kept silently shipping binaries with no commit/
// build-time (Go's linker no-ops on unresolvable -X targets -- no error,
// no warning). This test walks the repository and fails if any non-Makefile
// file hardcodes an -X target in the nonexistent repo-root cmd package, so
// copy-paste drift is caught by `go test` (and thus `make commit`) instead
// of only being noticed at release time. New build steps must derive their
// LDFLAGS from `make print-ldflags-value`.
func TestNoHandDuplicatedDeadPathLDFLAGS(t *testing.T) {
	repoRoot := findRepoRoot(t)

	// Assembled from pieces so this test file does not itself contain the
	// literal string it searches for (it would otherwise flag itself).
	deadPath := regexp.MustCompile(
		regexp.QuoteMeta("github.com/yaleh/meta-cc/cmd.") + `(Version|Commit|BuildTime)`)

	// Historical records that intentionally quote the broken string as part
	// of the bug's narrative (task write-ups, plans, past experiment logs),
	// plus VCS/build internals. None of these are consumed by `go build`;
	// the guard targets live build wiring (workflows, scripts, tooling).
	skipDirs := map[string]bool{
		".git":         true,
		"tasks":        true,
		"plans":        true,
		"_experiments": true,
		"build":        true,
		"dist":         true,
		"bin":          true,
		".archguard":   true,
	}

	var hits []string
	walkErr := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if rel != "." && skipDirs[rel] {
				return filepath.SkipDir
			}
			return nil
		}
		// The Makefile is the single source of truth for the -X targets (and
		// TestLDFLAGSCommitAndBuildTimeAreWired already fails if those targets
		// regress), so it is exempt from the duplication check.
		if rel == "Makefile" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for i, line := range strings.Split(string(data), "\n") {
			if deadPath.MatchString(line) {
				hits = append(hits, fmt.Sprintf("%s:%d: %s", rel, i+1, strings.TrimSpace(line)))
			}
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walking repo root %s: %v", repoRoot, walkErr)
	}

	if len(hits) > 0 {
		t.Errorf(
			"found %d hand-duplicated dead-path LDFLAGS string(s) outside the Makefile;\n"+
				"derive LDFLAGS from `make print-ldflags-value` instead of hand-writing -X targets\n"+
				"(see Makefile's LDFLAGS_VALUE / print-ldflags-value, tasks/DIR-049.md and DIR-052.md):\n  %s",
			len(hits), strings.Join(hits, "\n  "))
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
