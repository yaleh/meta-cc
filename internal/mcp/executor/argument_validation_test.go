package executor

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yaleh/meta-cc/internal/config"
	"github.com/yaleh/meta-cc/internal/mcp/tools"
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

// dispatchableTools returns the union of every tool reachable through either
// dispatch path: the special-tool registry (ExecuteSpecialTool) and the query
// pipeline registry. DIR-079: schema/runtime drift historically produced
// declared-but-ignored and executable-but-undeclared arguments (DIR-023/041/
// 045/048); these gates pin both dispatch paths to the declared schema.
func dispatchableTools() map[string]bool {
	names := make(map[string]bool, len(specialToolRegistry)+len(queryHandlerRegistry))
	for name := range specialToolRegistry {
		names[name] = true
	}
	for name := range queryHandlerRegistry {
		names[name] = true
	}
	return names
}

// TestDriftGate_ValidationRunsBeforeEveryDispatchPath asserts that runtime
// argument validation executes before EVERY dispatch path — both the special
// registry and the query pipeline registry — so an undeclared parameter can
// never reach a handler regardless of which path serves the tool. Adding a
// tool to either registry without validation coverage fails deterministically
// (validation returns before any handler runs, so no fixtures are required).
func TestDriftGate_ValidationRunsBeforeEveryDispatchPath(t *testing.T) {
	exec := NewToolExecutor()
	dispatchable := dispatchableTools()
	require.NotEmpty(t, dispatchable, "no tools registered — registries look broken")

	for name := range dispatchable {
		t.Run(name, func(t *testing.T) {
			_, err := exec.ExecuteTool(&config.Config{}, name, map[string]interface{}{
				"__drift_probe_undeclared_param": true,
			})
			require.Error(t, err, "tool %q must reject an undeclared parameter before dispatch", name)
			require.Contains(t, err.Error(), "unknown parameter(s): __drift_probe_undeclared_param",
				"tool %q must validate args before dispatch", name)
		})
	}
}

// TestDriftGate_RegistryAndSchemaAgree asserts that the set of dispatchable
// tools exactly equals the declared schema tools — catching both
// executable-but-undeclared tools (registered with no schema, so validation
// can't run) and declared-but-not-executable tools (a schema with no handler,
// so the tool is advertised but can never run).
func TestDriftGate_RegistryAndSchemaAgree(t *testing.T) {
	defs := tools.GetToolDefinitions()
	require.NotEmpty(t, defs, "tool registry returned no definitions")

	schemaTools := make(map[string]bool, len(defs))
	for _, def := range defs {
		schemaTools[def.Name] = true
	}
	dispatchable := dispatchableTools()

	for name := range dispatchable {
		if !schemaTools[name] {
			t.Errorf("executable-but-undeclared: tool %q is registered for dispatch but has no schema in tools.GetToolDefinitions()", name)
		}
	}
	for name := range schemaTools {
		if !dispatchable[name] {
			t.Errorf("declared-but-not-executable: tool %q has a schema but is registered in neither dispatch registry", name)
		}
	}
}

// TestDriftGate_SchemaParamsAreDeclaredExecutable pins the concrete historical
// drift cases: parameters that were once declared-but-ignored or
// executable-but-undeclared must now be declared on exactly the tools whose
// pipeline actually consumes them, so a regression to inert declaration fails.
func TestDriftGate_SchemaParamsAreDeclaredExecutable(t *testing.T) {
	index := tools.BuildToolSchemaIndex()

	// output_mode is consumed only by the four pipeline-backed tools
	// (DIR-023) and must be declared on each.
	for _, toolName := range []string{"query_session_content", "query_session_signals", "query_file_activity", "query_sessions"} {
		s, err := tools.GetToolSchemaByName(index, toolName)
		require.NoError(t, err)
		_, ok := s.Properties["output_mode"]
		require.True(t, ok, "%s must declare output_mode (consumed by pipeline.BuildResponse)", toolName)
	}

	// output_mode must NOT be advertised on analysis tools, which never reach
	// the response pipeline (DIR-048).
	for _, toolName := range []string{"analyze_bugs", "analyze_errors", "quality_scan", "get_work_patterns", "get_timeline", "get_tech_debt"} {
		s, err := tools.GetToolSchemaByName(index, toolName)
		require.NoError(t, err)
		_, ok := s.Properties["output_mode"]
		require.False(t, ok, "%s must not declare output_mode (never consumed)", toolName)
	}
}
