// quay-suite.sh regression tests — task gap-quay-test-output-undeclared-counts-absent.
//
// The defect these guard: quay's web test card rendered EVERY meta-cc round as `pass —/—`. The
// counts were honestly absent, not zero — quay's builtin parser only understands node:test / TAP,
// and the declared path (.quay/config.yml's loop.test_output) matches ONE line with ONE capture
// group and never sums. The fix has two halves that can each regress independently:
//
//   1. scripts/quay-suite.sh must print the already-totalled `quay-suite: pass N fail M skip K`
//      line, LAST, while passing go test's exit code and per-test failure lines through.
//   2. .quay/config.yml must run it (loop.test_command) and declare regexes (loop.test_output)
//      that actually bind to that line under quay's own matching semantics.
//
// These live in Go rather than tests/scripts/*.bats on purpose: bats is not installed on the
// developer host and `make test-bats` silently SKIPS when it is missing, so a .bats file here
// would be unverifiable exactly where `make commit` runs. This file rides `go test ./...`, the
// same sweep the acceptance gate uses.
package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// quaySuiteSummaryRE is the shape quay's loop.test_output regexes are built against — and the
// shape the task's first acceptance criterion pins, verbatim.
const quaySuiteSummaryRE = `^quay-suite: pass ([0-9]+) fail ([0-9]+) skip ([0-9]+)$`

// quayRepoRoot resolves the meta-cc checkout this test belongs to, via its own source path
// rather than the process working directory.
func quayRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller could not resolve this test's source path")
	return filepath.Dir(filepath.Dir(file))
}

// quaySuiteScript is the entrypoint under test.
func quaySuiteScript(t *testing.T) string {
	t.Helper()
	path := filepath.Join(quayRepoRoot(t), "scripts", "quay-suite.sh")
	_, err := os.Stat(path)
	require.NoError(t, err, "scripts/quay-suite.sh must exist — it is loop.test_command's value and the only producer of the summary line loop.test_output parses")
	return path
}

// runQuaySuite runs scripts/quay-suite.sh with dir as its working directory and returns its
// stdout and exit code.
//
// ⛔ The subprocess output is BUFFERED, never forwarded to this test's stdout. A test that prints
// a nested suite's output would inject extra `--- FAIL:` lines and a second summary line into the
// outer suite log — and quay's declared regex takes the FIRST summary match, not the last, so
// that would corrupt the outer round's counts rather than merely being noisy.
func runQuaySuite(t *testing.T, dir string, args ...string) (stdout string, exitCode int) {
	t.Helper()
	cmd := exec.Command("bash", append([]string{quaySuiteScript(t)}, args...)...)
	cmd.Dir = dir
	var out, errOut strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		require.ErrorAs(t, err, &exitErr,
			"quay-suite.sh failed to run at all (stderr: %s)", errOut.String())
		return out.String(), exitErr.ExitCode()
	}
	return out.String(), 0
}

// lastLine returns the final line of s, ignoring trailing newlines — the summary line must be
// LAST, so "the last line" is the assertion surface.
func lastLine(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	return lines[len(lines)-1]
}

// writeQuaySuiteFixture writes a throwaway Go module whose single test file declares exactly
// 2 passing, 3 failing and 4 skipping TOP-LEVEL test functions, plus one passing subtest nested
// inside TestPass1.
//
// The counts are deliberately distinct and non-zero so that a capture group bound to the wrong
// field is caught (2/3/4 cannot be confused with each other), and the subtest is what catches a
// subtest-blind counter: counting it would report pass 3 / tests 10 instead of pass 2 / tests 9.
func writeQuaySuiteFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	var b strings.Builder
	b.WriteString("package fixture\n\nimport \"testing\"\n\n")
	b.WriteString("func TestPass1(t *testing.T) {\n\tt.Run(\"child\", func(t *testing.T) {})\n}\n")
	b.WriteString("func TestPass2(t *testing.T) {}\n")
	for i := 1; i <= 3; i++ {
		fmt.Fprintf(&b, "func TestFail%d(t *testing.T) { t.Errorf(\"fixture failure %d\") }\n", i, i)
	}
	for i := 1; i <= 4; i++ {
		fmt.Fprintf(&b, "func TestSkip%d(t *testing.T) { t.Skip(\"fixture skip\") }\n", i)
	}

	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"),
		[]byte("module example.com/quaysuitefixture\n\ngo 1.21\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "fixture_test.go"),
		[]byte(b.String()), 0o644))
	return dir
}

// declaredTestOutput mirrors quay's readLoopTestOutput (worker-driver.ts): it reads
// .quay/config.yml's loop section and keeps only non-blank string values.
func declaredTestOutput(t *testing.T, configPath string) (testCommand string, declared map[string]string) {
	t.Helper()
	data, err := os.ReadFile(configPath)
	require.NoError(t, err, "reading %s", configPath)

	var cfg struct {
		Loop struct {
			TestCommand string            `yaml:"test_command"`
			TestOutput  map[string]string `yaml:"test_output"`
		} `yaml:"loop"`
	}
	require.NoError(t, yaml.Unmarshal(data, &cfg), "parsing %s", configPath)

	declared = map[string]string{}
	for field, re := range cfg.Loop.TestOutput {
		if strings.TrimSpace(re) != "" {
			declared[field] = re
		}
	}
	return cfg.Loop.TestCommand, declared
}

// TestQuaySuiteSummaryLineShape asserts AC1 on a real meta-cc package: exit 0, and a last line
// matching ^quay-suite: pass [0-9]+ fail [0-9]+ skip [0-9]+$.
//
// -short keeps it cheap: internal/version's ldflags test shells out to `make` and a full
// `go build`, and this test runs inside `go test ./...` — a compiler invocation nested in the
// suite (and racing the suite's own build) buys nothing that the plain package run does not
// already prove.
func TestQuaySuiteSummaryLineShape(t *testing.T) {
	stdout, exitCode := runQuaySuite(t, quayRepoRoot(t), "-short", "./internal/version/")

	require.Equal(t, 0, exitCode, "quay-suite.sh must exit 0 for a green package; tail:\n%s", tailLines(stdout, 15))

	summary := lastLine(stdout)
	require.Regexp(t, quaySuiteSummaryRE, summary,
		"the LAST line must be the summary quay's declared regexes bind to; got tail:\n%s", tailLines(stdout, 15))

	m := regexp.MustCompile(quaySuiteSummaryRE).FindStringSubmatch(summary)
	require.NotNil(t, m)
	require.NotZero(t, mustAtoi(t, m[1]),
		"a green internal/version run must report at least one passing test, not an empty count")
}

// TestQuaySuiteRedPathStaysRed asserts AC2 and the counting contract on a fixture whose outcome
// is known exactly: a failing suite must still exit non-zero, must still name its failing tests
// at column 0 (quay attributes a red round from those lines), and must count top-level test
// functions only.
func TestQuaySuiteRedPathStaysRed(t *testing.T) {
	stdout, exitCode := runQuaySuite(t, writeQuaySuiteFixture(t))

	require.NotZero(t, exitCode,
		"a fixture with failing tests must make quay-suite.sh exit non-zero — the wrapper wraps go test, it must never turn a red suite green")
	require.Contains(t, stdout, "--- FAIL: TestFail1 (",
		"per-test failure lines must survive verbatim into the suite log; they are what quay reads a red round's failing tests from")
	require.Equal(t, "quay-suite: pass 2 fail 3 skip 4", lastLine(stdout),
		"counts must be 2/3/4 — the subtest under TestPass1 must not be counted as a test of its own")
}

// TestQuaySuiteKeepsWholeModuleDefaultWithFlagsOnly pins the default package set against the
// footgun that would silently narrow it: keying "no packages given" on `$# -eq 0` means a
// flag-only invocation hands `go test` no package argument, and `go test` then falls back to the
// CURRENT DIRECTORY — testing nothing but the repo root. The fixture module keeps a subpackage
// with tests precisely so a bare-`.` run is distinguishable: it would report 0 counts, not 9.
func TestQuaySuiteKeepsWholeModuleDefaultWithFlagsOnly(t *testing.T) {
	dir := writeQuaySuiteFixture(t)
	sub := filepath.Join(dir, "sub")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "sub_test.go"),
		[]byte("package sub\n\nimport \"testing\"\n\nfunc TestSubPackageIsReached(t *testing.T) {}\n"), 0o644))

	// -count=1 only: a flag, no package. The default ./... must still be applied.
	stdout, _ := runQuaySuite(t, dir, "-count=1")

	require.Equal(t, "quay-suite: pass 3 fail 3 skip 4", lastLine(stdout),
		"a flag-only invocation must still default to ./... — the fixture's subpackage test is only reached through the recursive pattern")
}

// TestDeclaredTestOutputBindsToTheSummaryLine asserts AC3: the mechanism is only live if
// .quay/config.yml both RUNS the summary producer and declares regexes that bind to its output
// under quay's own semantics — `new RegExp(re, "m").exec(plain)` → m[1] → Number(), then
// tests = pass + fail + cancelled when `tests` itself is undeclared.
func TestDeclaredTestOutputBindsToTheSummaryLine(t *testing.T) {
	root := quayRepoRoot(t)
	testCommand, declared := declaredTestOutput(t, filepath.Join(root, ".quay", "config.yml"))

	require.Contains(t, testCommand, "scripts/quay-suite.sh",
		"loop.test_command must run the summary producer — without it loop.test_output has nothing to parse and the counts silently revert to absent")

	for _, field := range []string{"pass", "fail", "cancelled"} {
		require.Contains(t, declared, field,
			"loop.test_output.%s must be declared; undeclared fields are exactly the `—/—` this task fixes", field)
	}

	stdout, _ := runQuaySuite(t, writeQuaySuiteFixture(t))
	summary := lastLine(stdout)
	require.Regexp(t, quaySuiteSummaryRE, summary, "fixture summary line")

	counts := map[string]int{}
	for field, re := range declared {
		compiled, err := regexp.Compile(re)
		require.NoError(t, err, "loop.test_output.%s is not a valid regex", field)

		// quay reads m[1] ONLY — one capture group is the contract, not a style preference.
		// (Verified against applyDeclaredTestOutput; extra groups would be silently ignored,
		// so a two-group regex is a latent bug this assertion surfaces.)
		require.Equal(t, 1, compiled.NumSubexp(),
			"loop.test_output.%s must have exactly one capture group — quay reads m[1]", field)

		m := compiled.FindStringSubmatch(summary)
		require.NotNil(t, m,
			"loop.test_output.%s (%q) does not match the summary line %q — a declared-but-unmatched field does not fall back to the builtin parser, it stays absent",
			field, re, summary)

		n, err := strconv.Atoi(strings.TrimSpace(m[1]))
		require.NoError(t, err, "loop.test_output.%s capture %q is not a count", field, m[1])
		counts[field] = n
	}

	// quay's fallback formula: tests = pass + fail + cancelled (pre-verified-round-record.ts),
	// which is why `tests` is left undeclared and Go's skip is mapped onto `cancelled`.
	if _, declaredTests := counts["tests"]; !declaredTests {
		counts["tests"] = counts["pass"] + counts["fail"] + counts["cancelled"]
	}

	require.Equal(t, 2, counts["pass"])
	require.Equal(t, 3, counts["fail"])
	require.Equal(t, 4, counts["cancelled"], "Go's skip maps to quay's cancelled")
	require.Equal(t, 9, counts["tests"],
		"tests must equal the real test-function count (2+3+4), not a subtest-inflated 10")
}

func mustAtoi(t *testing.T, s string) int {
	t.Helper()
	n, err := strconv.Atoi(s)
	require.NoError(t, err, "not a number: %q", s)
	return n
}

// tailLines returns the last n lines of s, for failure messages that show the actual output
// without dumping an entire suite log into the test report.
func tailLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
