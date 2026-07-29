package analysis_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaleh/meta-cc/internal/analysis"
	"github.com/yaleh/meta-cc/internal/config"
)

// TestService_GetWorkPatterns_OmittedProviderFollowsHost proves the analysis
// dispatch resolves an omitted provider to the process host default
// (DIR-073). setupCodexProviderProject provides a Codex corpus and a
// deliberately missing Claude root, so the host default decides whether the
// call succeeds (codex path) or fails (claude path).
func TestService_GetWorkPatterns_OmittedProviderFollowsHost(t *testing.T) {
	projectPath := setupCodexProviderProject(t)
	svc := analysis.New()

	restore := config.SwapProcessDefault("codex")
	defer restore()

	out, err := svc.GetWorkPatterns(map[string]interface{}{"working_dir": projectPath})
	require.NoError(t, err, "omitted provider under a codex host must analyze the codex corpus")
	require.Contains(t, out, "apply_patch")

	restoreClaude := config.SwapProcessDefault("claude")
	defer restoreClaude()
	_, err = svc.GetWorkPatterns(map[string]interface{}{"working_dir": projectPath})
	require.Error(t, err, "under a claude host the claude corpus path applies (missing here)")
}
