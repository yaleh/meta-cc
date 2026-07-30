package executor

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/config"
)

func TestExecuteTool_AnalysisToolsRejectUndeclaredOutputMode(t *testing.T) {
	for _, toolName := range []string{
		"analyze_bugs",
		"analyze_errors",
		"quality_scan",
		"get_work_patterns",
		"get_timeline",
		"get_tech_debt",
	} {
		t.Run(toolName, func(t *testing.T) {
			_, err := NewToolExecutor().ExecuteTool(&config.Config{}, toolName, map[string]interface{}{
				"output_mode": "inline",
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), "unknown parameter(s): output_mode")
		})
	}
}

func TestExecuteTool_SpecialToolRejectsUndeclaredArgument(t *testing.T) {
	_, err := NewToolExecutor().ExecuteTool(&config.Config{}, "cleanup_temp_files", map[string]interface{}{
		"undeclared": true,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown parameter(s): undeclared")
}

func TestExecuteTool_QuerySessionsAcceptsOutputModes(t *testing.T) {
	projectPath, _ := setupClaudeSessionFixtureProject(t, "output mode fixture")

	for _, mode := range []string{"auto", "inline", "file_ref"} {
		t.Run(mode, func(t *testing.T) {
			output, err := NewToolExecutor().ExecuteTool(&config.Config{}, "query_sessions", map[string]interface{}{
				"working_dir": projectPath,
				"output_mode": mode,
			})
			require.NoError(t, err)

			var envelope struct {
				Mode string `json:"mode"`
			}
			require.NoError(t, json.Unmarshal([]byte(output), &envelope))
			if mode == "auto" {
				require.Contains(t, []string{"inline", "file_ref"}, envelope.Mode)
			} else {
				require.Equal(t, mode, envelope.Mode)
			}
		})
	}
}

func TestExecuteTool_QuerySessionsRejectsInvalidOutputMode(t *testing.T) {
	projectPath, _ := setupClaudeSessionFixtureProject(t, "invalid output mode fixture")

	_, err := NewToolExecutor().ExecuteTool(&config.Config{}, "query_sessions", map[string]interface{}{
		"working_dir": projectPath,
		"output_mode": "bogus",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "output_mode must be auto, inline, or file_ref")
}
