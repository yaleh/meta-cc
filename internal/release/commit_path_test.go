// DIR-089: commit-path invariants for the DIR-078 documentation gate.
//
// `make commit` covers internal/release twice over: check-docs (a member of
// check-essential) and the full `test` run. The two invocations share a single
// Go test-result cache entry only while check-docs carries the same -short flag
// as test — otherwise the package's tests execute on every commit, cold and
// warm. Nothing in the Go suite notices either failure mode, because both are
// properties of the Makefile rather than of the package. These assertions are
// therefore structural: they read the Makefile, walk the prerequisite graph
// from `commit`, and fail on the edit that reintroduces the double execution or
// that walks the documentation gate off the commit path.
package release

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// makeTarget is one `name: prereq prereq` line plus the recipe that follows it.
type makeTarget struct {
	name  string
	deps  []string
	lines []string // recipe lines, leading tab stripped
}

// recipe returns the target's recipe as one newline-joined string.
func (m *makeTarget) recipe() string {
	return strings.Join(m.lines, "\n")
}

// parseMakefile reads the Makefile into a target table. Only bare `name: deps`
// lines count; variable assignments (`GOTEST := $(GOCMD) test`), `:=`/`?=`/`=`
// forms, and conditional lines are skipped. A tab-indented line is a recipe
// line and attaches to the most recently declared target, exactly as make
// reads it.
func parseMakefile(t *testing.T, root string) map[string]*makeTarget {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "Makefile"))
	if err != nil {
		t.Fatalf("reading Makefile: %v", err)
	}

	targets := make(map[string]*makeTarget)
	var current *makeTarget
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "\t") {
			if current != nil {
				current.lines = append(current.lines, strings.TrimPrefix(line, "\t"))
			}
			continue
		}
		name, rest, ok := splitTargetLine(line)
		if !ok {
			current = nil
			continue
		}
		current = &makeTarget{name: name}
		for _, dep := range strings.Fields(stripComment(rest)) {
			// Skip order-only prerequisites and anything computed by make.
			if dep == "|" || strings.ContainsAny(dep, "$()") {
				continue
			}
			current.deps = append(current.deps, dep)
		}
		targets[name] = current
	}
	return targets
}

// splitTargetLine splits a non-recipe line into a target name and its
// prerequisite list. It reports false for comments, blank lines, variable
// assignments, and lines whose colon belongs to a `:=`/`::=` assignment.
func splitTargetLine(line string) (name, rest string, ok bool) {
	if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, " ") {
		return "", "", false
	}
	i := strings.Index(line, ":")
	if i <= 0 {
		return "", "", false
	}
	name, rest = line[:i], line[i+1:]
	// A target name is a bare word; `FOO := bar` and `FOO = bar` are not.
	if strings.ContainsAny(name, "= \t") || strings.HasPrefix(rest, "=") {
		return "", "", false
	}
	return name, rest, true
}

// stripComment drops a trailing `# ...` comment from a prerequisite list.
func stripComment(s string) string {
	if i := strings.Index(s, "#"); i >= 0 {
		return s[:i]
	}
	return s
}

// prereqClosure returns the transitive prerequisite set of target, including
// target itself, in deterministic order.
func prereqClosure(targets map[string]*makeTarget, target string) []string {
	seen := map[string]bool{target: true}
	queue := []string{target}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		t, ok := targets[cur]
		if !ok {
			continue
		}
		for _, dep := range t.deps {
			if !seen[dep] {
				seen[dep] = true
				queue = append(queue, dep)
			}
		}
	}
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// TestCommitPathRunsDocumentationContract is the negative test for the DIR-078
// gate's presence on the commit path: delete `check-docs` from the commit
// chain (or from check-essential, which commit depends on) and the
// documentation contract silently stops gating commits. This assertion fails
// on that edit instead.
func TestCommitPathRunsDocumentationContract(t *testing.T) {
	root := repoRoot(t)
	targets := parseMakefile(t, root)

	commit, ok := targets["commit"]
	if !ok {
		t.Fatal("Makefile has no `commit` target — the pre-commit gate this test guards is gone")
	}
	closure := prereqClosure(targets, "commit")

	var reached bool
	for _, name := range closure {
		if name == "check-docs" {
			reached = true
			break
		}
	}
	if !reached {
		t.Fatalf(
			"`make commit` no longer reaches the DIR-078 documentation gate.\n"+
				"  commit prerequisites: %v\n"+
				"  transitive closure:   %v\n"+
				"  Fix: keep check-docs reachable from commit — either directly, or via\n"+
				"  check-essential (`check-essential: %s`).",
			commit.deps, closure, strings.Join(targets["check-essential"].deps, " "),
		)
	}

	gate, ok := targets["check-docs"]
	if !ok {
		t.Fatal("Makefile declares check-docs in the commit closure but defines no such target")
	}
	if !strings.Contains(gate.recipe(), "internal/release/...") {
		t.Errorf(
			"check-docs no longer runs the internal/release contract tests.\n"+
				"  recipe: %s\n"+
				"  Fix: the DIR-078 contract lives in internal/release; keep the gate pointed at it.",
			gate.recipe(),
		)
	}
}

// TestDocumentationGateSharesTestCacheEntry pins the DIR-089 dedupe: check-docs
// and `test` both cover internal/release, and Go's test-result cache keys on
// the test binary's flag set. Drop -short from check-docs and the two
// invocations address different cache entries, so the package's tests run
// twice on every `make commit` — warm as well as cold.
func TestDocumentationGateSharesTestCacheEntry(t *testing.T) {
	root := repoRoot(t)
	targets := parseMakefile(t, root)

	gate, ok := targets["check-docs"]
	if !ok {
		t.Fatal("Makefile has no check-docs target")
	}
	suite, ok := targets["test"]
	if !ok {
		t.Fatal("Makefile has no test target")
	}

	if !strings.Contains(suite.recipe(), "-short") {
		t.Fatalf("`test` no longer runs in short mode; the commit-path cache argument assumes it does.\n  recipe: %s", suite.recipe())
	}
	if !strings.Contains(gate.recipe(), "-short") {
		t.Errorf(
			"check-docs runs internal/release without -short, so its Go test-cache entry\n"+
				"cannot be reused by `test` (-short) and the package executes twice per commit.\n"+
				"  check-docs recipe: %s\n"+
				"  Fix: restore -short on the check-docs $(GOTEST) line (DIR-089).",
			gate.recipe(),
		)
	}
}

// TestReleasePackageHasNoShortGatedTests holds up the other half of the DIR-089
// argument: `make commit` runs this gate in short mode, which is only sound
// while no test here is skipped when short mode is on. Add such a guard and the
// documentation contract would stop being enforced on the commit path while
// still passing — so this assertion fails first, and points at the fix
// (drop -short from check-docs and re-establish cache reuse some other way).
func TestReleasePackageHasNoShortGatedTests(t *testing.T) {
	root := repoRoot(t)
	dir := filepath.Join(root, "internal", "release")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading %s: %v", dir, err)
	}

	// The needle is assembled at run time so this file's own source does not
	// match the scan it performs.
	needle := "testing" + "." + "Short("

	scanned := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		scanned++
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("reading %s: %v", e.Name(), err)
		}
		for i, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, needle) {
				t.Errorf(
					"internal/release/%s:%d gates a test on short mode (%s).\n"+
						"  `make commit` runs this package via check-docs with -short, so the\n"+
						"  gated test would be skipped on the commit path — the DIR-078 gate would\n"+
						"  pass while checking less. Fix: remove the short-mode guard, or drop\n"+
						"  -short from check-docs and re-solve DIR-089's cache reuse.",
					e.Name(), i+1, strings.TrimSpace(line),
				)
			}
		}
	}
	if scanned == 0 {
		t.Fatal("no Go files scanned in internal/release — this guard would silently cover nothing")
	}
}
