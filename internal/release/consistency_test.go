// Package release contains offline consistency checks that treat
// internal/version/release.json as the single source of truth for the
// current meta-cc release version, and internal/mcp/tools.GetToolDefinitions()
// as the single source of truth for the current MCP tool count.
//
// These tests fail (with an actionable message) whenever a shipped manifest,
// server banner, or current-facing doc drifts from either source. They run
// as part of `go test ./...` — no network access is required, and no files
// are rewritten. Historical documents (docs/proposals, docs/plans,
// docs/phases, docs/archive, CHANGELOG.md, plans/, tasks/) intentionally
// describe past states and are NOT checked here.
package release

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/yaleh/meta-cc/internal/mcp/tools"
	"github.com/yaleh/meta-cc/internal/version"
)

// repoRoot resolves the repository root relative to this package's directory
// (internal/release), so the test works regardless of the caller's cwd.
func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	root, err := filepath.Abs(filepath.Join(wd, "..", ".."))
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("could not locate repo root (no go.mod) starting from %s: %v", wd, err)
	}
	return root
}

type marketplaceFile struct {
	Plugins []struct {
		Version     string `json:"version"`
		Description string `json:"description"`
	} `json:"plugins"`
}

type pluginFile struct {
	Version     string `json:"version"`
	Description string `json:"description"`
}

func readJSON(t *testing.T, path string, v interface{}) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("parsing %s as JSON: %v", path, err)
	}
}

// TestPluginManifestVersionsMatchReleaseVersion asserts that every shipped
// Claude/Codex plugin manifest agrees with internal/version/release.json —
// the canonical release version (AC: "Every shipped Claude/Codex/server
// version surface agrees with that source").
func TestPluginManifestVersionsMatchReleaseVersion(t *testing.T) {
	root := repoRoot(t)
	want := version.Version

	var rootMarketplace marketplaceFile
	readJSON(t, filepath.Join(root, ".claude-plugin/marketplace.json"), &rootMarketplace)

	var srcMarketplace marketplaceFile
	readJSON(t, filepath.Join(root, "plugin-src/.claude-plugin/marketplace.json"), &srcMarketplace)

	var claudePlugin pluginFile
	readJSON(t, filepath.Join(root, "plugin-src/.claude-plugin/plugin.json"), &claudePlugin)

	var codexPlugin pluginFile
	readJSON(t, filepath.Join(root, "plugin-src/.codex-plugin/plugin.json"), &codexPlugin)

	cases := []struct {
		surface string
		got     string
	}{
		{".claude-plugin/marketplace.json (plugins[0].version)", rootMarketplace.Plugins[0].Version},
		{"plugin-src/.claude-plugin/marketplace.json (plugins[0].version)", srcMarketplace.Plugins[0].Version},
		{"plugin-src/.claude-plugin/plugin.json (version)", claudePlugin.Version},
		{"plugin-src/.codex-plugin/plugin.json (version)", codexPlugin.Version},
	}

	for _, c := range cases {
		if c.got != want {
			t.Errorf(
				"version drift: %s = %q, but internal/version/release.json = %q.\n"+
					"Fix: run `./scripts/release/bump-plugin-version.sh --version %s --non-interactive` "+
					"to resync every manifest, or hand-edit %s to %q if release.json itself should change.",
				c.surface, c.got, want, want, c.surface, want,
			)
		}
	}
}

// docToolCountCheck names a CURRENT (non-historical) documentation or
// manifest file, plus the regexes that extract tool-count claims from it.
// Only files that make an active capability claim to users/agents today
// belong here. Historical proposals/plans are deliberately excluded per the
// task's design constraint against flagging historical documents.
type docToolCountCheck struct {
	relPath  string
	patterns []*regexp.Regexp
}

var currentToolCountDocs = []docToolCountCheck{
	{
		relPath: "README.md",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(\d+)\s+MCP [Tt]ools`),
			regexp.MustCompile(`\((\d+) tools\)`),
		},
	},
	{
		relPath: "CLAUDE.md",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`\((\d+) tools\)`),
			regexp.MustCompile(`[Pp]rovides (\d+) tools`),
		},
	},
	{
		relPath: "docs/DOCUMENTATION_MAP.md",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`\((\d+) tools\)`),
		},
	},
	{
		relPath: "docs/guides/integration.md",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`exposes (\d+) tools`),
		},
	},
	{
		relPath: "docs/guides/mcp.md",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`registers \*\*(\d+) tools\*\*`),
		},
	},
	{
		relPath: "docs/reference/features.md",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`exposes (\d+) MCP tools`),
		},
	},
	{
		relPath: ".claude-plugin/marketplace.json",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(\d+) MCP tools`),
		},
	},
	{
		relPath: "plugin-src/.claude-plugin/marketplace.json",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(\d+) MCP tools`),
		},
	},
	{
		relPath: "plugin-src/.claude-plugin/plugin.json",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(\d+) MCP tools`),
		},
	},
}

// TestAdvertisedToolCountMatchesRegisteredTools asserts that every
// tool-count claim in current, user-facing documentation and manifests
// equals len(tools.GetToolDefinitions()) — the real, live tools/list count
// (AC: "The current advertised tool count equals the real tools/list
// count"). The count is never hardcoded here as an independent constant;
// it is always read from the actual tool registry.
func TestAdvertisedToolCountMatchesRegisteredTools(t *testing.T) {
	root := repoRoot(t)
	actual := len(tools.GetToolDefinitions())
	if actual == 0 {
		t.Fatal("tools.GetToolDefinitions() returned 0 tools — registry looks broken")
	}

	totalMatches := 0
	for _, dc := range currentToolCountDocs {
		path := filepath.Join(root, dc.relPath)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", dc.relPath, err)
		}
		content := string(data)

		for _, re := range dc.patterns {
			matches := re.FindAllStringSubmatch(content, -1)
			for _, m := range matches {
				totalMatches++
				n, convErr := strconv.Atoi(m[1])
				if convErr != nil {
					t.Fatalf("%s: matched %q but could not parse a tool count from it: %v", dc.relPath, m[0], convErr)
				}
				if n != actual {
					t.Errorf(
						"tool-count drift: %s claims %d tools (matched %q), but "+
							"tools.GetToolDefinitions() currently registers %d tools.\n"+
							"Fix: update %s to say %d — or if you intentionally added/removed a "+
							"tool, update every surface listed in currentToolCountDocs "+
							"(internal/release/consistency_test.go) to %d.",
						dc.relPath, n, m[0], actual, dc.relPath, actual, actual,
					)
				}
			}
		}
	}

	if totalMatches == 0 {
		t.Fatal("no tool-count mentions matched any pattern in currentToolCountDocs — " +
			"the docs may have been reworded, or the regexes are stale. Update " +
			"internal/release/consistency_test.go so this check keeps covering real claims.")
	}
}

// liveToolNames returns the set of tool names registered by
// tools.GetToolDefinitions() — the live tools/list source of truth.
func liveToolNames(t *testing.T) map[string]bool {
	t.Helper()
	defs := tools.GetToolDefinitions()
	if len(defs) == 0 {
		t.Fatal("tools.GetToolDefinitions() returned 0 tools — registry looks broken")
	}
	names := make(map[string]bool, len(defs))
	for _, d := range defs {
		names[d.Name] = true
	}
	return names
}

// TestFeatureCatalogMatchesLiveToolRegistry asserts that the "## MCP Tools"
// section of docs/reference/features.md catalogs exactly the registered tool
// set: every tool bullet (`- `name`: ...`) names a live tool (no phantoms),
// every live tool appears (no omissions — e.g. query_sessions), and no tool
// is listed twice, so the per-category bullet lists partition the registry
// and their counts must sum to the live total (DIR-074). Parameter values
// inside descriptions are deliberately ignored — only line-leading tool
// bullets are matched.
func TestFeatureCatalogMatchesLiveToolRegistry(t *testing.T) {
	root := repoRoot(t)
	live := liveToolNames(t)

	data, err := os.ReadFile(filepath.Join(root, "docs/reference/features.md"))
	if err != nil {
		t.Fatalf("reading docs/reference/features.md: %v", err)
	}
	section := regexp.MustCompile(`(?ms)^## MCP Tools\n(.*?)^## `).FindStringSubmatch(string(data))
	if section == nil {
		t.Fatal("docs/reference/features.md has no '## MCP Tools' section — " +
			"the feature catalog may have been reworded. Restore the section or " +
			"update internal/release/consistency_test.go to keep covering it.")
	}

	counts := make(map[string]int)
	for _, m := range regexp.MustCompile(`(?m)^- `+"`"+`([a-z][a-z0-9_]*)`+"`"+`:`).FindAllStringSubmatch(section[1], -1) {
		counts[m[1]]++
	}
	if len(counts) == 0 {
		t.Fatal("no '- `tool_name`:' bullets found in the '## MCP Tools' section of " +
			"docs/reference/features.md — the catalog may have been reworded.")
	}

	for name, n := range counts {
		if !live[name] {
			t.Errorf("feature-catalog drift: docs/reference/features.md lists %q, but "+
				"tools.GetToolDefinitions() does not register it. Remove the stale entry "+
				"or register the tool.", name)
		}
		if n > 1 {
			t.Errorf("feature-catalog drift: docs/reference/features.md lists %q %d times; "+
				"each tool must appear in exactly one category so the category counts "+
				"partition the registry and sum to %d.", name, n, len(live))
		}
	}
	for name := range live {
		if counts[name] == 0 {
			t.Errorf("feature-catalog drift: tools.GetToolDefinitions() registers %q, but "+
				"docs/reference/features.md does not catalog it. Add it to the category it "+
				"belongs to; the category lists must sum to the live tool count (%d).",
				name, len(live))
		}
	}
}

// TestReadmeCategoryBreakdownSumsToLiveCount asserts that every README
// category breakdown of the form "N session discovery + N consolidated query
// + N two-stage + N analysis + N cleanup" sums to the live registered tool
// count, so the overview arithmetic cannot silently drift (DIR-074).
func TestReadmeCategoryBreakdownSumsToLiveCount(t *testing.T) {
	root := repoRoot(t)
	actual := len(tools.GetToolDefinitions())

	data, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatalf("reading README.md: %v", err)
	}
	re := regexp.MustCompile(`(\d+) session discovery \+ (\d+) consolidated query \+ (\d+) two-stage(?: \([^)]*\))? \+ (\d+) analysis \+ (\d+) cleanup`)
	matches := re.FindAllStringSubmatch(string(data), -1)
	if len(matches) == 0 {
		t.Fatal("README.md has no category breakdown of the form " +
			"'N session discovery + N consolidated query + N two-stage + N analysis + N cleanup'. " +
			"Restore the breakdown or update internal/release/consistency_test.go to keep covering it.")
	}
	for _, m := range matches {
		sum := 0
		for _, g := range m[1:] {
			n, convErr := strconv.Atoi(g)
			if convErr != nil {
				t.Fatalf("README.md: matched %q but could not parse %q as a count: %v", m[0], g, convErr)
			}
			sum += n
		}
		if sum != actual {
			t.Errorf("README category breakdown drift: %q sums to %d, but "+
				"tools.GetToolDefinitions() currently registers %d tools. Fix the counts "+
				"in README.md so the breakdown sums to %d.", m[0], sum, actual, actual)
		}
	}
}

// TestReadmeGoRequirementMatchesGoMod asserts that the Go prerequisite in
// README.md states the same major.minor toolchain as the go directive in
// go.mod, so the documented requirement cannot silently drift (DIR-074).
func TestReadmeGoRequirementMatchesGoMod(t *testing.T) {
	root := repoRoot(t)

	gomod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("reading go.mod: %v", err)
	}
	wantMatch := regexp.MustCompile(`(?m)^go (\d+\.\d+)`).FindStringSubmatch(string(gomod))
	if wantMatch == nil {
		t.Fatal("go.mod has no 'go X.Y' directive — cannot derive the required toolchain")
	}
	want := wantMatch[1]

	readme, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatalf("reading README.md: %v", err)
	}
	matches := regexp.MustCompile(`Go (\d+\.\d+) or later`).FindAllStringSubmatch(string(readme), -1)
	if len(matches) == 0 {
		t.Fatal("README.md has no 'Go X.Y or later' prerequisite — the requirement " +
			"may have been reworded. Restore it or update internal/release/consistency_test.go.")
	}
	for _, m := range matches {
		if m[1] != want {
			t.Errorf("Go requirement drift: README.md says 'Go %s or later', but go.mod "+
				"requires Go %s. Update README.md prerequisites to match go.mod.", m[1], want)
		}
	}
}
