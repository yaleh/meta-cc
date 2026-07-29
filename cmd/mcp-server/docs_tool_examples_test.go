package main

// DIR-075: normative documentation must present only the currently registered
// MCP tool surface as executable guidance. The consolidated query tools
// (query_session_content, query_session_signals, query_file_activity,
// query_sessions) replaced the legacy specialized query tools; removed names
// such as query_tools, query_user_messages, query_tool_errors, and
// query_token_usage may survive only inside explicitly labeled migration or
// historical material (see legacyAllowedDocs), never in current examples.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/yaleh/meta-cc/internal/mcp/tools"
)

var removedToolNames = []string{
	"query_tools",
	"query_user_messages",
	"query_tool_errors",
	"query_token_usage",
}

// legacyAllowedDocs may mention removed tool names because they are explicitly
// labeled migration/historical material. Each value is a marker string the
// file must contain, so the label cannot silently rot away while the file
// keeps teaching removed tools as current.
var legacyAllowedDocs = map[string]string{
	"docs/guides/mcp-query-tools.md":            "Migration from Previous Tools",
	"docs/guides/mcp-v2-migration.md":           "SUPERSEDED",
	"docs/guides/migration-to-unified-query.md": "SUPERSEDED",
	"docs/guides/unified-query-api.md":          "SUPERSEDED",
	"docs/examples/query-cookbook.md":           "SUPERSEDED",
	"docs/examples/mcp-query-cookbook.md":       "SUPERSEDED",
	// Legacy names appear only inside historical session-record fixture
	// output (e.g. old cmd/query_tools.go file paths in sample records),
	// not as executable guidance.
	"docs/examples/multi-file-jsonl-queries.md": "historical session-record fixtures",
}

func docsRepoRoot(t *testing.T) string {
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
		t.Fatalf("repo root has no go.mod (from %s): %v", wd, err)
	}
	return root
}

func normativeDocFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	for _, dir := range []string{"docs/guides", "docs/tutorials", "docs/examples"} {
		entries, err := os.ReadDir(filepath.Join(root, dir))
		if err != nil {
			t.Fatalf("reading %s: %v", dir, err)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				files = append(files, filepath.Join(dir, e.Name()))
			}
		}
	}
	return append(files, "docs/reference/jsonl-schema.md")
}

// TestNormativeDocsKeepRemovedToolsOutOfCurrentGuidance fails when a normative
// doc (guide, tutorial, example, or the JSONL schema reference) mentions a
// removed tool name without being explicitly labeled migration/historical
// material.
func TestNormativeDocsKeepRemovedToolsOutOfCurrentGuidance(t *testing.T) {
	root := docsRepoRoot(t)
	for _, rel := range normativeDocFiles(t, root) {
		body, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("reading %s: %v", rel, err)
		}
		text := string(body)
		if marker, allowed := legacyAllowedDocs[rel]; allowed {
			if !strings.Contains(text, marker) {
				t.Errorf("%s is allowlisted for legacy tool names but no longer carries its label marker %q", rel, marker)
			}
			continue
		}
		for _, name := range removedToolNames {
			if strings.Contains(text, name) {
				t.Errorf("%s mentions removed tool %q outside a labeled migration/historical context; rewrite the example with query_session_content / query_session_signals / query_file_activity / query_sessions (mapping: docs/guides/mcp-query-tools.md)", rel, name)
			}
		}
	}
}

var mcpToolCallPatterns = []*regexp.Regexp{
	// JavaScript-style calls: query_session_signals({...}), get_session_stats().
	// No whitespace allowed before the paren so prose like "Stage 2 query (if
	// implemented)" does not match.
	regexp.MustCompile(`\b((?:query|get_session|get_timeline|get_tech_debt|get_work_patterns|inspect_session|execute_stage2|analyze_|quality_scan|cleanup_temp_files)[a-z0-9_]*)\(`),
	// JSON-RPC examples: "name": "query_session_signals"
	regexp.MustCompile(`"name"\s*:\s*"((?:query|get_session|get_timeline|get_tech_debt|get_work_patterns|inspect_session|execute_stage2|analyze_|quality_scan|cleanup_temp_files)[a-z0-9_]*)"`),
}

// TestNormativeDocExamplesOnlyCallRegisteredTools validates representative
// executable examples (JS-style tool calls and JSON-RPC "name" fields)
// against the live tool schemas, so docs cannot drift back to any removed
// tool, not just the four names hard-listed above.
func TestNormativeDocExamplesOnlyCallRegisteredTools(t *testing.T) {
	root := docsRepoRoot(t)
	index := tools.BuildToolSchemaIndex()
	for _, rel := range normativeDocFiles(t, root) {
		if _, allowed := legacyAllowedDocs[rel]; allowed {
			continue // labeled historical material may show removed calls
		}
		body, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("reading %s: %v", rel, err)
		}
		text := string(body)
		for _, re := range mcpToolCallPatterns {
			for _, m := range re.FindAllStringSubmatch(text, -1) {
				name := m[1]
				if _, err := tools.GetToolSchemaByName(index, name); err != nil {
					t.Errorf("%s shows an executable example calling %q, which is not a registered MCP tool (see docs/guides/mcp-query-tools.md for current equivalents)", rel, name)
				}
			}
		}
	}
}
