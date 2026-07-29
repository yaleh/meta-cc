// DIR-078: offline documentation-contract gate. It complements
// consistency_test.go (curated prose patterns) by catching drift those checks
// miss: an example calling a removed tool, a rotten link, and a Go prerequisite
// below the go.mod baseline. Markdown is classified current/historical/migration
// and only current docs are checked; historical and migration content is
// allowlisted so it never false-positives. Diagnostics name file, line, claim,
// and accepted replacement.
package release

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// docScope labels a markdown file for the current-doc contract.
type docScope string

const (
	scopeCurrent    docScope = "current"    // makes active claims today; fully checked
	scopeHistorical docScope = "historical" // past state (proposals/plans/phases/archives/ADRs/fixtures); exempt
	scopeMigration  docScope = "migration"  // labeled old->new mapping; removed invocations allowed
)

// historicalDirPrefixes are trees whose markdown is historical by definition
// (prefix test on the slash-normalized repo-relative path; slash is significant).
var historicalDirPrefixes = []string{
	"docs/proposals/",
	"docs/plans/",
	"docs/phases/",
	"docs/archive/",
	"docs/tasks/",
	"docs/experiments/",
	"docs/analysis/",
	"docs/architecture/adr/", // Architecture Decision Records are historical by nature
	"docs/testing/fixtures/", // captured MCP output fixtures, not current guidance
	"plans/",
	"tasks/",
	"_experiments/",
}

// scopeEntry is one checked-in classification: why a file outside a historical
// directory is held to a non-current scope.
type scopeEntry struct {
	scope  docScope
	reason string
}

// docScopeAllowlist is the checked-in manifest of individually classified files
// (the status manifest for intentional historical and migration references).
// Add an entry, with a reason, to exempt a current-directory file.
var docScopeAllowlist = map[string]scopeEntry{
	"docs/examples/mcp-query-cookbook.md": {
		scope:  scopeHistorical,
		reason: "SUPERSEDED banner: v2.0 query examples retained as a historical record",
	},
	"docs/core/plan.md": {
		scope:  scopeHistorical,
		reason: "phase roadmap describing completed phases, not current usage guidance",
	},
	"docs/methodology/empirical-methodology-development.md": {
		scope:  scopeHistorical,
		reason: "historical case study; tool names are illustrative pseudocode",
	},
	"docs/architecture/cli-removal-dependency-analysis.md": {
		scope:  scopeHistorical,
		reason: "historical analysis supporting the CLI removal",
	},
	"docs/guides/mcp-v2-migration.md": {
		scope:  scopeMigration,
		reason: "labeled v1->v2 migration mapping",
	},
	"docs/guides/migration-to-unified-query.md": {
		scope:  scopeMigration,
		reason: "labeled migration mapping to the (since removed) unified query tool",
	},
	"docs/guides/unified-query-api.md": {
		scope:  scopeMigration,
		reason: "documents the removed unified query API; superseded by the consolidated tools",
	},
}

// classifyDocScope returns a path's scope and a reason (allowlist wins over prefixes).
func classifyDocScope(rel string) (docScope, string) {
	rel = filepath.ToSlash(rel)
	if e, ok := docScopeAllowlist[rel]; ok {
		return e.scope, e.reason
	}
	for _, p := range historicalDirPrefixes {
		if strings.HasPrefix(rel, p) {
			return scopeHistorical, "under historical directory prefix " + p
		}
	}
	return scopeCurrent, ""
}

// currentMarkdownFiles returns the slash-normalized repo-relative paths of every
// current markdown file under docs/ plus top-level README/CLAUDE (fails if none).
func currentMarkdownFiles(t *testing.T, root string) []string {
	t.Helper()
	var rel []string
	walkErr := filepath.Walk(filepath.Join(root, "docs"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		r, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		r = filepath.ToSlash(r)
		if s, _ := classifyDocScope(r); s == scopeCurrent {
			rel = append(rel, r)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walking docs/: %v", walkErr)
	}
	for _, top := range []string{"README.md", "CLAUDE.md"} {
		if _, statErr := os.Stat(filepath.Join(root, top)); statErr == nil {
			if s, _ := classifyDocScope(top); s == scopeCurrent {
				rel = append(rel, top)
			}
		}
	}
	if len(rel) == 0 {
		t.Fatal("no current markdown files resolved — the scope classification is broken")
	}
	return rel
}

// twoStageWorkflow is the live replacement for the removed unified `query` /
// `query_raw` tools and for the specialized v1 tools that mcp-v2-migration.md
// §Removed Tools maps onto "query + jq" (query_context, query_tools_advanced,
// query_time_series): that unified query tool was itself removed, so its live
// equivalent is the two-stage workflow.
const twoStageWorkflow = "the two-stage workflow: get_session_directory -> inspect_session_files -> execute_stage2_query"

// removedToolReplacement maps every MCP tool removed from the registry to its
// current replacement (mirrors the "Migration from Previous Tools" table in
// docs/guides/mcp-query-tools.md and §Removed Tools in mcp-v2-migration.md, the
// latter resolved transitively to live tools). Its keys define the removed-tool
// set, so a detected invocation always has an accepted replacement diagnostic.
var removedToolReplacement = map[string]string{
	"query_user_messages":      `query_session_content(role="user")`,
	"query_assistant_messages": `query_session_content(role="assistant")`,
	"query_summaries":          `query_session_content(role="assistant", contains="## Summary")`,
	"query_conversation_flow":  `query_session_content(role="all")`,
	"query_conversation":       `query_session_content(role="all")`, // doc: query_conversation_flow
	"query_tool_blocks":        `query_session_content(role="tool")`,
	"query_tool_results":       `query_session_content(role="tool")`,
	"query_tool_errors":        `query_session_signals(type="errors")`,
	"query_token_usage":        `query_session_signals(type="tokens")`,
	"query_system_errors":      `query_session_signals(type="system_errors")`,
	"query_timestamps":         `query_session_signals(type="timestamps")`,
	"query_tools":              `query_session_signals(type="tool_stats")`,
	"query_tool_stats":         `query_session_signals(type="tool_stats")`,
	"query_file_snapshots":     `query_file_activity(type="snapshots")`,
	"query_files":              `query_file_activity(type="snapshots")`, // doc: query_file_snapshots
	"query_context":            twoStageWorkflow,                        // doc: query + custom jq
	"query_tools_advanced":     twoStageWorkflow,                        // doc: query + jq filtering
	"query_time_series":        twoStageWorkflow,                        // doc: query + jq grouping
	"query_raw":                twoStageWorkflow,
}

// removedToolInvocationRe matches a removed tool used as a call: an identifier
// boundary, the name, optional whitespace, then "(". The paren excludes file
// names and prose. Built from removedToolReplacement keys (RE2 has no lookbehind,
// so the boundary is group 1 and the name group 2).
var removedToolInvocationRe = func() *regexp.Regexp {
	names := make([]string, 0, len(removedToolReplacement))
	for n := range removedToolReplacement {
		names = append(names, n)
	}
	sort.Strings(names)
	return regexp.MustCompile(`(^|[^A-Za-z0-9_])(` + strings.Join(names, "|") + `)[ \t]*\(`)
}()

// fenceState tracks whether subsequent lines are inside a fenced code block.
// A naive toggle desyncs on nested example blocks (a ```markdown block that
// shows inner ``` snippets), which let a removed-tool call slip through unseen.
// consume instead treats any fence line as an opener when outside a block, but
// only a BARE fence (no info string) as a closer — matching CommonMark, where a
// closing fence carries no info string — so a nested ```lang example stays
// inside the block and its calls remain detectable.
type fenceState struct{ in bool }

// consume reports whether line is a fence delimiter (and so not scannable
// content) while advancing the in/out-of-code state.
func (s *fenceState) consume(line string) bool {
	t := strings.TrimSpace(line)
	if !strings.HasPrefix(t, "```") && !strings.HasPrefix(t, "~~~") {
		return false
	}
	if !s.in {
		s.in = true
	} else if strings.TrimLeft(t, "`~") == "" {
		s.in = false // only a bare fence closes an open block
	}
	return true
}

// codeHit is a flagged line inside a fenced code block.
type codeHit struct {
	line int
	tool string
}

// removedToolInvocations scans fenced code blocks for removed-tool calls.
// Inline code, prose, and table rows sit outside fences and are ignored, since
// migration mappings are expressed as tables.
func removedToolInvocations(content string) []codeHit {
	var hits []codeHit
	var fs fenceState
	for i, line := range strings.Split(content, "\n") {
		if fs.consume(line) {
			continue
		}
		if !fs.in {
			continue
		}
		for _, m := range removedToolInvocationRe.FindAllStringSubmatch(line, -1) {
			hits = append(hits, codeHit{line: i + 1, tool: m[2]})
		}
	}
	return hits
}

// goVersionRe matches a documented "Go MAJOR.MINOR" prerequisite claim.
var goVersionRe = regexp.MustCompile(`Go (\d+)\.(\d+)`)

// goVersionHit is a documented Go version below the baseline.
type goVersionHit struct {
	line    int
	version string
}

// goModBaseline returns the major/minor of the go.mod go directive (min toolchain).
func goModBaseline(t *testing.T, root string) (int, int) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("reading go.mod: %v", err)
	}
	m := regexp.MustCompile(`(?m)^go (\d+)\.(\d+)`).FindStringSubmatch(string(data))
	if m == nil {
		t.Fatal("go.mod has no 'go MAJOR.MINOR' directive — cannot derive the baseline toolchain")
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	return major, minor
}

// belowBaselineGoVersions returns every documented Go version strictly below
// baseMajor.baseMinor, compared numerically (Go 1.9 < Go 1.24), not lexically.
func belowBaselineGoVersions(content string, baseMajor, baseMinor int) []goVersionHit {
	var hits []goVersionHit
	for i, line := range strings.Split(content, "\n") {
		for _, m := range goVersionRe.FindAllStringSubmatch(line, -1) {
			major, _ := strconv.Atoi(m[1])
			minor, _ := strconv.Atoi(m[2])
			if major < baseMajor || (major == baseMajor && minor < baseMinor) {
				hits = append(hits, goVersionHit{line: i + 1, version: m[1] + "." + m[2]})
			}
		}
	}
	return hits
}

// linkCheckedDocs are the current non-guide navigation/index pages whose
// relative links must all resolve. Every current docs/guides/*.md page is
// link-checked too (see currentGuideDocs), so the whole current guide surface
// — not just a curated subset — is covered.
var linkCheckedDocs = []string{
	"README.md",
	"docs/DOCUMENTATION_MAP.md",
	"docs/QUICK_ACCESS.md",
	"docs/core/principles.md",
	"docs/reference/features.md",
}

// currentGuideDocs returns the slash-normalized repo-relative paths of every
// current docs/guides/*.md page (migration/historical guides are excluded by
// classifyDocScope), so the link gate covers the entire current guide surface.
func currentGuideDocs(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, "docs", "guides"))
	if err != nil {
		t.Fatalf("reading docs/guides: %v", err)
	}
	var rel []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		r := "docs/guides/" + e.Name()
		if s, _ := classifyDocScope(r); s == scopeCurrent {
			rel = append(rel, r)
		}
	}
	return rel
}

// linkCheckedFiles is the full set of pages the link gate covers: the curated
// non-guide navigation pages plus every current guide.
func linkCheckedFiles(t *testing.T, root string) []string {
	t.Helper()
	files := append([]string(nil), linkCheckedDocs...)
	return append(files, currentGuideDocs(t, root)...)
}

// mdLinkRe matches an inline markdown link and captures its target.
var mdLinkRe = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

// brokenLink is a relative link target that does not resolve on disk.
type brokenLink struct {
	line   int
	target string
}

// brokenRelativeLinks returns the relative markdown links in content (resolved
// against baseDir) whose targets do not exist. Absolute URLs, mailto links,
// absolute paths, and same-file "#anchor" links are skipped; a "#fragment" on a
// valid path is dropped before the check. Fenced code blocks are ignored.
func brokenRelativeLinks(baseDir, content string) []brokenLink {
	var broken []brokenLink
	var fs fenceState
	for i, line := range strings.Split(content, "\n") {
		if fs.consume(line) {
			continue
		}
		if fs.in {
			continue
		}
		for _, m := range mdLinkRe.FindAllStringSubmatch(line, -1) {
			target := strings.Trim(strings.TrimSpace(m[1]), "<>")
			if sp := strings.IndexByte(target, ' '); sp >= 0 {
				target = target[:sp] // drop an optional "title"
			}
			if target == "" || strings.HasPrefix(target, "#") {
				continue
			}
			if strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") || strings.HasPrefix(target, "/") {
				continue
			}
			path := target
			if h := strings.IndexByte(path, '#'); h >= 0 {
				path = path[:h]
			}
			if path == "" {
				continue
			}
			if _, err := os.Stat(filepath.Join(baseDir, filepath.FromSlash(path))); err != nil {
				broken = append(broken, brokenLink{line: i + 1, target: m[1]})
			}
		}
	}
	return broken
}

// TestCurrentDocsDoNotInvokeRemovedTools: no current document's executable
// examples may call a tool removed from the registry (AC: a removed-tool
// invocation added to a current guide or example fails).
func TestCurrentDocsDoNotInvokeRemovedTools(t *testing.T) {
	root := repoRoot(t)
	for _, rel := range currentMarkdownFiles(t, root) {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("reading %s: %v", rel, err)
		}
		for _, h := range removedToolInvocations(string(data)) {
			t.Errorf(
				"removed-tool invocation: %s:%d calls %q, which is no longer registered.\n"+
					"  Accepted replacement: %s.\n"+
					"  Fix the example, or—if this is an intentional labeled migration mapping—add %s\n"+
					"  to docScopeAllowlist (scope migration) in internal/release/doc_contract_test.go.",
				rel, h.line, h.tool, removedToolReplacement[h.tool], rel,
			)
		}
	}
}

// TestCurrentDocsGoVersionAtLeastBaseline: every Go prerequisite in a current
// page must be at or above the go.mod baseline (AC: a documented Go
// prerequisite below the baseline is detected).
func TestCurrentDocsGoVersionAtLeastBaseline(t *testing.T) {
	root := repoRoot(t)
	baseMajor, baseMinor := goModBaseline(t, root)
	for _, rel := range currentMarkdownFiles(t, root) {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("reading %s: %v", rel, err)
		}
		for _, h := range belowBaselineGoVersions(string(data), baseMajor, baseMinor) {
			t.Errorf(
				"stale Go prerequisite: %s:%d documents Go %s, below the go.mod baseline Go %d.%d.\n"+
					"  Fix: update %s to require Go %d.%d or later (or correct go.mod if the baseline itself moved).",
				rel, h.line, h.version, baseMajor, baseMinor, rel, baseMajor, baseMinor,
			)
		}
	}
}

// TestCurrentNavigationLinksResolve: every relative link in the current
// index/navigation pages and in every current docs/guides/*.md page must
// resolve (AC: a broken relative link in a current index or guide is detected).
func TestCurrentNavigationLinksResolve(t *testing.T) {
	root := repoRoot(t)
	files := linkCheckedFiles(t, root)
	if guides := currentGuideDocs(t, root); len(guides) == 0 {
		t.Fatal("no current docs/guides/*.md pages resolved — the guide link check would silently cover nothing")
	}
	for _, rel := range files {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		data, err := os.ReadFile(abs)
		if err != nil {
			t.Errorf("link-checked page is missing: %s: %v", rel, err)
			continue
		}
		for _, b := range brokenRelativeLinks(filepath.Dir(abs), string(data)) {
			t.Errorf(
				"broken link: %s:%d links to %q, which does not resolve.\n"+
					"  Fix: correct the target path or remove the link.",
				rel, b.line, b.target,
			)
		}
	}
}

// Unit tests below use synthetic inputs to lock in false-positive protection.

func TestClassifyDocScope(t *testing.T) {
	cases := []struct {
		path string
		want docScope
	}{
		{"docs/guides/mcp.md", scopeCurrent},
		{"docs/reference/features.md", scopeCurrent},
		{"docs/methodology/role-based-documentation.md", scopeCurrent}, // R1: fixed in place, not exempted
		{"README.md", scopeCurrent},
		{"docs/proposals/proposal-x.md", scopeHistorical},
		{"docs/plans/47-49-query-ux-improvements.md", scopeHistorical},
		{"docs/archive/mcp-usage.md", scopeHistorical},
		{"docs/architecture/adr/ADR-003-mcp-server-integration.md", scopeHistorical},
		{"docs/testing/fixtures/signals-errors-basic.md", scopeHistorical}, // R2: captured fixture output
		{"plans/anything.md", scopeHistorical},
		{"tasks/DIR-001.md", scopeHistorical},
		{"docs/examples/mcp-query-cookbook.md", scopeHistorical},
		{"docs/core/plan.md", scopeHistorical},
		{"docs/guides/mcp-v2-migration.md", scopeMigration},
		{"docs/guides/unified-query-api.md", scopeMigration},
	}
	for _, c := range cases {
		if got, _ := classifyDocScope(c.path); got != c.want {
			t.Errorf("classifyDocScope(%q) = %s, want %s", c.path, got, c.want)
		}
	}
}

func TestRemovedToolInvocations(t *testing.T) {
	content := strings.Join([]string{
		"# Guide", // 1
		"",        // 2
		`Use query_session_signals(type="errors").`, // 3 current tool, prose -> no hit
		"",                               // 4
		"```javascript",                  // 5
		"query_tool_errors({limit: 10})", // 6 -> hit
		"```",                            // 7
		"",                               // 8
		"Mentions `query_token_usage` inline here.",   // 9 prose inline code -> no hit
		"| `query_tools` | `query_session_signals` |", // 10 migration table row -> no hit
		"```", // 11
		"see cmd/query_tools.go for the old source", // 12 file name, no paren -> no hit
		"```", // 13
	}, "\n")

	hits := removedToolInvocations(content)
	if len(hits) != 1 {
		t.Fatalf("expected exactly 1 removed-tool invocation, got %d: %+v", len(hits), hits)
	}
	if hits[0].tool != "query_tool_errors" {
		t.Errorf("wrong tool: got %q, want query_tool_errors", hits[0].tool)
	}
	if hits[0].line != 6 {
		t.Errorf("wrong line: got %d, want 6", hits[0].line)
	}
	if _, ok := removedToolReplacement[hits[0].tool]; !ok {
		t.Errorf("detected tool %q has no replacement mapping for diagnostics", hits[0].tool)
	}
}

// TestRemovedToolManifestCoversV1Tools locks in the DIR-078 audit fix (R1): the
// five Phase-25.4 tools removed in mcp-v2-migration.md §Removed Tools must each
// be in the manifest with a live replacement, and each must be detected as a
// fenced invocation.
func TestRemovedToolManifestCoversV1Tools(t *testing.T) {
	for _, tool := range []string{
		"query_context", "query_tools_advanced", "query_time_series", "query_conversation", "query_files",
	} {
		repl, ok := removedToolReplacement[tool]
		if !ok || repl == "" {
			t.Errorf("removed tool %q is missing from removedToolReplacement (needs a live replacement)", tool)
			continue
		}
		content := "```javascript\n" + tool + "({limit: 5})\n```\n"
		hits := removedToolInvocations(content)
		if len(hits) != 1 || hits[0].tool != tool {
			t.Errorf("removedToolInvocations did not detect a fenced %q call: %+v", tool, hits)
		}
	}
}

// TestRemovedToolInvocationNestedFence is the regression for the missed
// docs/methodology/role-based-documentation.md:1214 detection: a removed call
// inside a nested example block (a ```markdown block showing inner ``` snippets)
// must still be flagged, which a naive fence toggle desyncs and misses.
func TestRemovedToolInvocationNestedFence(t *testing.T) {
	content := strings.Join([]string{
		"```markdown",                 // 1 outer example block opens
		"Sample doc with an example:", // 2
		"```python",                   // 3 nested fence (info string) -> still inside outer block
		"query_files(threshold=5)",    // 4 -> must be hit
		"```",                         // 5 bare fence closes
		"```",                         // 6
	}, "\n")
	hits := removedToolInvocations(content)
	if len(hits) != 1 || hits[0].tool != "query_files" || hits[0].line != 4 {
		t.Fatalf("nested-fence call not detected: got %+v, want one query_files hit at line 4", hits)
	}
}

// TestCurrentGuideDocs locks in the DIR-078 audit fix (R3): the link gate covers
// every current docs/guides/*.md page, while migration guides stay excluded.
func TestCurrentGuideDocs(t *testing.T) {
	root := repoRoot(t)
	guides := currentGuideDocs(t, root)
	if len(guides) < 5 {
		t.Fatalf("expected the current guide surface to be sizeable, got %d: %v", len(guides), guides)
	}
	index := make(map[string]bool, len(guides))
	for _, g := range guides {
		index[g] = true
	}
	for _, want := range []string{"docs/guides/git-hooks.md", "docs/guides/capabilities.md", "docs/guides/mcp-query-tools.md"} {
		if !index[want] {
			t.Errorf("current guide %q missing from currentGuideDocs", want)
		}
	}
	for _, migr := range []string{"docs/guides/mcp-v2-migration.md", "docs/guides/unified-query-api.md"} {
		if index[migr] {
			t.Errorf("migration guide %q must not be link-checked as a current guide", migr)
		}
	}
}

func TestBelowBaselineGoVersions(t *testing.T) {
	content := strings.Join([]string{
		"Requires Go 1.24 or later.", // 1 == baseline -> ok
		"Install Go 1.21",            // 2 below -> hit
		"Go 1.9 is too old",          // 3 below (numeric, not lexical) -> hit
		"Go 1.25 also works",         // 4 above -> ok
	}, "\n")

	hits := belowBaselineGoVersions(content, 1, 24)
	if len(hits) != 2 {
		t.Fatalf("expected 2 below-baseline versions, got %d: %+v", len(hits), hits)
	}
	if hits[0].version != "1.21" || hits[0].line != 2 {
		t.Errorf("first hit = %+v, want version 1.21 at line 2", hits[0])
	}
	if hits[1].version != "1.9" || hits[1].line != 3 {
		t.Errorf("second hit = %+v, want version 1.9 at line 3", hits[1])
	}
}

func TestBrokenRelativeLinks(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "exists.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "page.md"), []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}

	content := strings.Join([]string{
		"[ok](exists.md)",                 // 1 resolves -> ok
		"[ok anchor](exists.md#section)",  // 2 file resolves -> ok
		"[missing](nope.md)",              // 3 -> broken
		"[dir](sub/)",                     // 4 directory resolves -> ok
		"[web](https://example.com/x.md)", // 5 absolute URL -> skipped
		"[anchor](#top)",                  // 6 same-file anchor -> skipped
		"[missing nested](sub/gone.md)",   // 7 -> broken
	}, "\n")

	broken := brokenRelativeLinks(dir, content)
	if len(broken) != 2 {
		t.Fatalf("expected 2 broken links, got %d: %+v", len(broken), broken)
	}
	if broken[0].line != 3 || broken[1].line != 7 {
		t.Errorf("broken link lines = %d,%d; want 3,7", broken[0].line, broken[1].line)
	}
}
