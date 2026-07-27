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
