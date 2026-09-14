package executor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestQuerySessionContent_Codex_OneCorruptSessionDoesNotEraseOthers is the
// DIR-030 fixture-based proof of the "one corrupt rollout does not erase
// valid results from other sessions" contract: a project-scope query
// across 2 valid Codex sessions plus 1 session whose rollout file is
// missing must still return the 2 valid sessions' results, plus exactly
// one warning naming the corrupt session.
func TestQuerySessionContent_Codex_OneCorruptSessionDoesNotEraseOthers(t *testing.T) {
	projectPath := setupCodexMultiSessionFixtureProject(t, 2)

	e := NewToolExecutor()
	result, err := handleQuerySessionContent(e, "project", map[string]interface{}{
		"role":        "user",
		"provider":    "codex",
		"working_dir": projectPath,
	})
	require.NoError(t, err, "a corrupt session must not abort the whole project query")
	require.Len(t, result.Entries, 2, "expected one user-message record per valid session (2), corrupt session excluded")

	require.Len(t, result.Warnings, 1, "expected exactly one warning for the corrupt session")
	require.Contains(t, result.Warnings[0], "corrupt-session")
}
