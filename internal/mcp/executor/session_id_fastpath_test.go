package executor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestQuerySessionContent_Codex_SessionIDReadsOnlyThatThread is the DIR-030
// end-to-end proof that session_id is an exact-thread selector distinct
// from scope="session" (most recent): with 2 valid Codex sessions in the
// project, requesting session_id=<the older one> must return only that
// session's content, even though scope="project" would normally include
// both, and even though scope="session" would pick the newer one.
func TestQuerySessionContent_Codex_SessionIDReadsOnlyThatThread(t *testing.T) {
	projectPath := setupCodexMultiSessionFixtureProject(t, 2)

	result, err := handleQuerySessionContent(NewToolExecutor(), "project", map[string]interface{}{
		"role":        "user",
		"provider":    "codex",
		"working_dir": projectPath,
		"session_id":  "good-session-a",
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1, "session_id must restrict to exactly the requested session, not every session in project scope")

	entry, ok := result.Entries[0].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "good-session-a", entry["session_id"])
}

// TestQuerySessionContent_Codex_SessionIDUnknownReturnsActionableError
// proves an unknown session_id fails closed with an error rather than a
// silently-empty result.
func TestQuerySessionContent_Codex_SessionIDUnknownReturnsActionableError(t *testing.T) {
	projectPath := setupCodexMultiSessionFixtureProject(t, 1)

	_, err := handleQuerySessionContent(NewToolExecutor(), "project", map[string]interface{}{
		"role":        "user",
		"provider":    "codex",
		"working_dir": projectPath,
		"session_id":  "does-not-exist",
	})
	require.Error(t, err)
}

// TestQuerySessionSignals_Codex_SessionIDFastPath proves session_id also
// wires through query_session_signals (not just query_session_content),
// via the shared dispatchProviderQuery path.
func TestQuerySessionSignals_Codex_SessionIDFastPath(t *testing.T) {
	projectPath := setupCodexMultiSessionFixtureProject(t, 2)

	result, err := handleQuerySessionSignals(NewToolExecutor(), "project", map[string]interface{}{
		"type":        "timestamps",
		"provider":    "codex",
		"working_dir": projectPath,
		"session_id":  "good-session-b",
	})
	require.NoError(t, err)
	for _, entry := range result.Entries {
		m, ok := entry.(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "good-session-b", m["session_id"])
	}
}

// TestQuerySessionContent_Claude_SessionIDRespectsCWDBoundary is the DIR-030
// cross-project leak regression proof for the default/claude provider path:
// a session_id that exists on disk but belongs to a DIFFERENT project than
// the caller's working_dir must not return that session's content, just
// because the ID happened to match somewhere on disk. This mirrors
// TestBuildForSession_CWDBoundaryExcludesCrossProjectSession's guarantee
// for the codex/all path — before the fix, the default/claude path used
// locator.FromSessionID directly with zero comparison against working_dir,
// so this test failed (returned the cross-project content) prior to the
// fix in ExecuteQueryForSession.
func TestQuerySessionContent_Claude_SessionIDRespectsCWDBoundary(t *testing.T) {
	_, sessionID := setupClaudeSessionFixtureProject(t, "hello world")

	otherProject := t.TempDir()

	_, err := handleQuerySessionContent(NewToolExecutor(), "project", map[string]interface{}{
		"role":        "user",
		"working_dir": otherProject,
		"session_id":  sessionID,
	})
	require.Error(t, err, "a session_id belonging to a different project must not be readable via a mismatched working_dir")
}

// TestQuerySessionSignals_Claude_SessionIDRespectsCWDBoundary is the
// query_session_signals companion to the content-tool boundary regression
// above, proving the fix applies consistently across the shared
// dispatchProviderQuery path (not just query_session_content).
func TestQuerySessionSignals_Claude_SessionIDRespectsCWDBoundary(t *testing.T) {
	_, sessionID := setupClaudeSessionFixtureProject(t, "hello world")

	otherProject := t.TempDir()

	_, err := handleQuerySessionSignals(NewToolExecutor(), "project", map[string]interface{}{
		"type":        "timestamps",
		"working_dir": otherProject,
		"session_id":  sessionID,
	})
	require.Error(t, err, "a session_id belonging to a different project must not be readable via a mismatched working_dir")
}

// TestQuerySessionContent_Claude_SessionIDSameProjectStillWorks proves the
// cwd-boundary fix above does not spuriously exclude a session_id that
// genuinely belongs to the caller's own working_dir/project — an
// overly-broad fix could break this.
func TestQuerySessionContent_Claude_SessionIDSameProjectStillWorks(t *testing.T) {
	projectPath, sessionID := setupClaudeSessionFixtureProject(t, "hello world")

	result, err := handleQuerySessionContent(NewToolExecutor(), "project", map[string]interface{}{
		"role":        "user",
		"working_dir": projectPath,
		"session_id":  sessionID,
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1)
	entry, ok := result.Entries[0].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, sessionID, entry["sessionId"])
}

// TestQuerySessionContent_All_SessionIDRespectsCWDBoundaryForClaudeSession
// proves the fix in providerrecords.BuildForSession also closes the same
// leak for provider="all": before the fix, BuildForSession's cwd-boundary
// check only applied to conversation.ProviderCodex sessions, so a Claude
// session reached via provider="all" (which also iterates the Claude
// sub-provider) was NOT boundary-checked. All four provider variants
// (omitted/claude/codex/all) must behave consistently for this AC.
func TestQuerySessionContent_All_SessionIDRespectsCWDBoundaryForClaudeSession(t *testing.T) {
	_, sessionID := setupClaudeSessionFixtureProject(t, "hello world")

	otherProject := t.TempDir()

	_, err := handleQuerySessionContent(NewToolExecutor(), "project", map[string]interface{}{
		"role":        "user",
		"provider":    "all",
		"working_dir": otherProject,
		"session_id":  sessionID,
	})
	require.Error(t, err, "a session_id belonging to a different project must not be readable via provider=\"all\" with a mismatched working_dir")
}
