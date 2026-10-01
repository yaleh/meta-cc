package executor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/config"
)

// This file is the regression suite for the path-normalization defect that
// opened task gap-path-normalization-bypassed-in-query-sessions-filter. Two
// path-normalization disciplines coexist in this codebase: the session corpus
// is KEYED by locator.PathToHash, which resolves symlinks, while some
// comparison inputs were only Cleaned by filepath.Abs. query_sessions wrote
// that Abs result straight into filter.CWD and compared it — never passing
// through PathToHash — so a symlinked ALIAS of a project path returned a
// silently empty listing (no error, no warning) while the physical path
// returned the real sessions. On this host /home/yale -> /data/home/yale, so
// the same repository is reachable as either form.
//
// The regression is asserted as a PROPERTY, not as the two symptoms
// originally reported: for every path-taking entry point, the alias and
// physical forms of the same project must return the same result. The
// reported defect (query_sessions, default and explicit-cwd arms) is one row
// of that table; the Codex adjudication is another (see the bottom of this
// file).

// symlinkAlias creates a symlink that resolves to real and returns the alias
// path — the shape of the real-world trigger (a symlinked prefix such as
// /home -> /data/home).
func symlinkAlias(t *testing.T, real string) string {
	t.Helper()
	alias := filepath.Join(t.TempDir(), "project-alias")
	require.NoError(t, os.Symlink(real, alias))
	return alias
}

// sessionIDs extracts the session_id of every entry so two listings can be
// compared by content without depending on the exact rendered shape.
func sessionIDs(t *testing.T, entries []interface{}) []string {
	t.Helper()
	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		m, ok := e.(map[string]interface{})
		require.True(t, ok, "unexpected entry shape: %#v", e)
		id, ok := m["session_id"].(string)
		require.True(t, ok, "entry has no session_id: %#v", m)
		ids = append(ids, id)
	}
	return ids
}

// TestQuerySessions_SymlinkedAliasMatchesPhysical is the direct regression for
// the reported defect: query_sessions with a symlinked alias working_dir must
// return the same sessions as the physical path — on BOTH arms (the default
// arm, which derives filter.CWD from working_dir, and the explicit-cwd arm,
// where the caller passes cwd directly).
func TestQuerySessions_SymlinkedAliasMatchesPhysical(t *testing.T) {
	projectPath, sessionID := setupClaudeSessionFixtureProject(t, "alias probe")
	alias := symlinkAlias(t, projectPath)

	// Guard: the physical path must produce a non-empty listing, otherwise
	// "both empty" would satisfy an equivalence assertion vacuously.
	physical, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{"working_dir": projectPath})
	require.NoError(t, err)
	require.Equal(t, []string{sessionID}, sessionIDs(t, physical.Entries),
		"the physical path must list the fixture session; the equivalence below is vacuous if it does not")

	t.Run("default arm (working_dir only)", func(t *testing.T) {
		aliased, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{"working_dir": alias})
		require.NoError(t, err)
		require.Equal(t, sessionIDs(t, physical.Entries), sessionIDs(t, aliased.Entries),
			"a symlinked alias working_dir must list the same sessions as the physical path, not a silent empty result")
	})

	t.Run("explicit-cwd arm (cwd passed directly)", func(t *testing.T) {
		// working_dir is held constant as the alias; only cwd varies — the
		// variable-isolating probe from the task's evidence section.
		viaPhysicalCWD, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
			"working_dir": alias,
			"cwd":         projectPath,
		})
		require.NoError(t, err)
		require.Equal(t, []string{sessionID}, sessionIDs(t, viaPhysicalCWD.Entries),
			"a physical cwd must match even when working_dir is an alias")

		viaAliasCWD, err := handleQuerySessions(NewToolExecutor(), "project", map[string]interface{}{
			"working_dir": alias,
			"cwd":         alias,
		})
		require.NoError(t, err)
		require.Equal(t, sessionIDs(t, viaPhysicalCWD.Entries), sessionIDs(t, viaAliasCWD.Entries),
			"an alias cwd must match the same sessions as the physical cwd")
	})
}

// TestPathTakingEntryPoints_AliasEqualsPhysical is the general property the
// regression is filed under: for EVERY path-taking entry point, the alias and
// physical forms of the same project must return the same result. The two
// query_sessions arms are rows 1-2. query_session_content, analyze_errors and
// get_session_directory were alias-safe at filing time because they funnel the
// path through locator.PathToHash; this table pins that so a future entry point
// that compares raw paths is caught here rather than as a silent empty result.
func TestPathTakingEntryPoints_AliasEqualsPhysical(t *testing.T) {
	projectPath, _ := setupClaudeSessionFixtureProject(t, "alias property")
	alias := symlinkAlias(t, projectPath)
	cfg := &config.Config{}

	// "output_mode": "inline" keeps the rendered response from embedding the
	// per-call temp file path a file_ref response carries — the comparison must
	// be over the RESULT, not over a name that differs between any two calls.
	tests := []struct {
		name    string
		tool    string
		pathArg string // which argument receives the project path
		extra   map[string]interface{}
	}{
		{name: "query_sessions/working_dir", tool: "query_sessions", pathArg: "working_dir",
			extra: map[string]interface{}{"output_mode": "inline"}},
		// working_dir is held constant (physical) and only cwd varies.
		{name: "query_sessions/cwd", tool: "query_sessions", pathArg: "cwd",
			extra: map[string]interface{}{"working_dir": projectPath, "output_mode": "inline"}},
		{name: "query_session_content/working_dir", tool: "query_session_content", pathArg: "working_dir",
			extra: map[string]interface{}{"role": "user", "pattern": ".*", "output_mode": "inline"}},
		{name: "analyze_errors/working_dir", tool: "analyze_errors", pathArg: "working_dir"},
		{name: "get_session_directory/working_dir", tool: "get_session_directory", pathArg: "working_dir",
			extra: map[string]interface{}{"scope": "project", "provider": "claude"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			argsFor := func(p string) map[string]interface{} {
				args := map[string]interface{}{}
				for k, v := range tc.extra {
					args[k] = v
				}
				args[tc.pathArg] = p
				return args
			}

			physical, err := NewToolExecutor().ExecuteTool(cfg, tc.tool, argsFor(projectPath))
			require.NoError(t, err, "the physical-path form must execute")
			require.NotEmpty(t, physical, "the physical-path result must be non-empty so the equivalence is not vacuous")

			aliased, err := NewToolExecutor().ExecuteTool(cfg, tc.tool, argsFor(alias))
			require.NoError(t, err, "the alias-path form must execute (a loud error is not equivalence)")
			require.Equal(t, physical, aliased,
				"%s must return the same result for a symlinked alias and the physical path", tc.tool)
		})
	}
}

// TestCodexPathTakingEntryPoints_AliasEqualsPhysical adjudicates the "Open
// item (NOT evidenced)" from the task's Proposal and pins the verdict as a
// regression.
//
// The open item asked whether the Codex-only discovery path — which reaches
// rawfiles.NewRegistry via internal/mcp/query/stage.go's resolveProjectPath —
// is alias-safe, since Codex rollouts are not stored under the Claude <hash>
// layout and so may not pass through locator.PathToHash.
//
// VERDICT: it is the SAME defect. resolveProjectPath only Abs'd (Cleaned) its
// input, and rawfiles.SelectCodexFiles matches sessions with a RAW cwd
// comparison (providerrecords.FilterSessionsForScope), so an aliased
// working_dir filtered every Codex session out and surfaced the loud
// "no codex sessions found for project <alias>" while the physical path
// returned the corpus. resolveProjectPath now resolves symlinks; this test
// proves both Codex path-taking entry points agree across the two forms.
func TestCodexPathTakingEntryPoints_AliasEqualsPhysical(t *testing.T) {
	projectPath := setupCodexRolloutFixtureProject(t, "rollout-legacy-sample.jsonl")
	alias := symlinkAlias(t, projectPath)
	cfg := &config.Config{}

	for _, tool := range []string{"get_session_directory", "get_session_metadata"} {
		t.Run(tool, func(t *testing.T) {
			argsFor := func(p string) map[string]interface{} {
				return map[string]interface{}{"scope": "project", "provider": "codex", "working_dir": p}
			}

			physical, err := NewToolExecutor().ExecuteTool(cfg, tool, argsFor(projectPath))
			require.NoError(t, err, "the physical-path form must execute")
			require.NotEmpty(t, physical, "the physical-path result must be non-empty so the equivalence is not vacuous")

			aliased, err := NewToolExecutor().ExecuteTool(cfg, tool, argsFor(alias))
			require.NoError(t, err,
				"the Codex %s path must resolve an aliased working_dir, not fail with 'no codex sessions found for project'", tool)
			require.Equal(t, physical, aliased,
				"Codex %s must return the same result for a symlinked alias and the physical path", tool)
		})
	}
}
