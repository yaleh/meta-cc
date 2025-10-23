package cmd

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	internalOutput "github.com/yaleh/meta-cc/internal/output"
	"github.com/yaleh/meta-cc/internal/parser"
	querypkg "github.com/yaleh/meta-cc/internal/query"
	pkgoutput "github.com/yaleh/meta-cc/pkg/output"
)

func TestQueryToolsLibraryMatchesCLI(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())

	projectPath := t.TempDir()
	sessionID := "library-cli-tools"
	fixture := `{"type":"assistant","timestamp":"2025-10-02T10:00:00Z","uuid":"uuid-1","sessionId":"` + sessionID + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Bash","input":{"command":"ls"}}]}}
{"type":"user","timestamp":"2025-10-02T10:00:01Z","uuid":"uuid-2","sessionId":"` + sessionID + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"file.txt"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:00:02Z","uuid":"uuid-3","sessionId":"` + sessionID + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-2","name":"Read","input":{"file_path":"/tmp/file.txt"}}]}}
{"type":"user","timestamp":"2025-10-02T10:00:03Z","uuid":"uuid-4","sessionId":"` + sessionID + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-2","content":"data"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixture)

	cliOutput := runCLICommand(t, "query", "tools", "--project", projectPath, "--limit", "1", "--output", "jsonl")

	libraryOpts := querypkg.ToolsQueryOptions{
		Pipeline:   toPipelineOptions(GlobalOptions{ProjectPath: projectPath}),
		Limit:      1,
		Offset:     0,
		SortBy:     "",
		Reverse:    false,
		Status:     "",
		Tool:       "",
		Where:      "",
		Expression: "",
	}

	results, err := querypkg.RunToolsQuery(libraryOpts)
	if err != nil {
		t.Fatalf("library query failed: %v", err)
	}

	formatted, err := internalOutput.FormatOutput(results, "jsonl")
	if err != nil {
		t.Fatalf("failed to format library output: %v", err)
	}

	if strings.TrimSpace(cliOutput) != strings.TrimSpace(formatted) {
		t.Fatalf("library output does not match CLI output.\nCLI: %s\nLibrary: %s", cliOutput, formatted)
	}
}

func TestQueryUserMessagesLibraryMatchesCLI(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())

	projectPath := t.TempDir()
	sessionID := "library-cli-messages"
	fixture := `{"type":"user","timestamp":"2025-10-02T10:00:00Z","uuid":"uuid-1","sessionId":"` + sessionID + `","message":{"role":"user","content":[{"type":"text","text":"Investigate bug"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:00:01Z","uuid":"uuid-2","sessionId":"` + sessionID + `","message":{"role":"assistant","content":[{"type":"text","text":"Sure"}]}}
{"type":"user","timestamp":"2025-10-02T10:00:02Z","uuid":"uuid-3","sessionId":"` + sessionID + `","message":{"role":"user","content":[{"type":"text","text":"Fix applied"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixture)

	cliOutput := runCLICommand(t, "query", "user-messages", "--project", projectPath, "--pattern", "Fix", "--with-context", "1", "--output", "jsonl")

	libraryOpts := querypkg.UserMessagesQueryOptions{
		Pipeline: toPipelineOptions(GlobalOptions{ProjectPath: projectPath}),
		Pattern:  "Fix",
		Context:  1,
		Limit:    0,
		Offset:   0,
		SortBy:   "",
		Reverse:  false,
	}

	messages, err := querypkg.RunUserMessagesQuery(libraryOpts)
	if err != nil {
		t.Fatalf("library query failed: %v", err)
	}

	formatted, err := internalOutput.FormatOutput(messages, "jsonl")
	if err != nil {
		t.Fatalf("failed to format library output: %v", err)
	}

	if strings.TrimSpace(cliOutput) != strings.TrimSpace(formatted) {
		t.Fatalf("library output does not match CLI output.\nCLI: %s\nLibrary: %s", cliOutput, formatted)
	}
}

func TestQueryToolsAdvancedWhereLikeMatchesCLI(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())

	projectPath := t.TempDir()
	sessionID := "library-cli-advanced"
	fixture := `{"type":"assistant","timestamp":"2025-10-02T10:00:00Z","uuid":"uuid-1","sessionId":"` + sessionID + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"meta-cc-run","input":{"command":"meta"}}]}}
{"type":"user","timestamp":"2025-10-02T10:00:01Z","uuid":"uuid-2","sessionId":"` + sessionID + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"ok"}]}}
`

	writeSessionFixture(t, projectPath, sessionID, fixture)

	cliOutput := runCLICommand(t, "query", "tools", "--project", projectPath, "--where", "tool LIKE 'meta-cc%'")

	opts := querypkg.ToolsQueryOptions{
		Pipeline: toPipelineOptions(GlobalOptions{ProjectPath: projectPath}),
		Where:    "tool LIKE 'meta-cc%'",
	}

	results, err := querypkg.RunToolsQuery(opts)
	if err != nil {
		t.Fatalf("library query failed: %v", err)
	}

	formatted, err := internalOutput.FormatOutput(results, "jsonl")
	if err != nil {
		t.Fatalf("failed to format library output: %v", err)
	}

	if strings.TrimSpace(cliOutput) != strings.TrimSpace(formatted) {
		t.Fatalf("library output does not match CLI output.\nCLI: %s\nLibrary: %s", cliOutput, formatted)
	}
}

func TestSessionStatsLibraryMatchesCLI(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())

	projectPath := t.TempDir()
	sessionID := "library-cli-stats"
	fixture := `{"type":"user","timestamp":"2025-10-02T10:00:00Z","uuid":"uuid-1","sessionId":"` + sessionID + `","message":{"role":"user","content":[{"type":"text","text":"run analysis"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:01:00Z","uuid":"uuid-2","sessionId":"` + sessionID + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Bash","input":{"command":"ls"}}]}}
{"type":"user","timestamp":"2025-10-02T10:01:30Z","uuid":"uuid-3","sessionId":"` + sessionID + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"file.txt"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixture)

	cliOutput := runCLICommand(t, "parse", "stats", "--project", projectPath, "--output", "jsonl")

	entries, toolCalls := loadEntriesAndToolCalls(t, projectPath)
	stats := querypkg.BuildSessionStats(entries, toolCalls)
	formatted, err := internalOutput.FormatOutput(stats, "jsonl")
	if err != nil {
		t.Fatalf("failed to format stats: %v", err)
	}

	if strings.TrimSpace(cliOutput) != strings.TrimSpace(formatted) {
		t.Fatalf("library session stats mismatch. CLI: %s, Library: %s", cliOutput, formatted)
	}
}

func TestTimeSeriesLibraryMatchesCLI(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())

	projectPath := t.TempDir()
	sessionID := "library-cli-timeseries"
	fixture := `{"type":"assistant","timestamp":"2025-10-02T10:00:00Z","uuid":"uuid-1","sessionId":"` + sessionID + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Bash","input":{"command":"ls"}}]}}
{"type":"user","timestamp":"2025-10-02T10:05:00Z","uuid":"uuid-2","sessionId":"` + sessionID + `","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool-1","content":"file.txt"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixture)

	cliOutput := runCLICommand(t, "stats", "time-series", "--project", projectPath, "--metric", "tool-calls", "--interval", "hour", "--output", "jsonl")

	_, toolCalls := loadEntriesAndToolCalls(t, projectPath)
	points, err := querypkg.AnalyzeTimeSeries(toolCalls, "tool-calls", "hour", "")
	if err != nil {
		t.Fatalf("AnalyzeTimeSeries failed: %v", err)
	}
	formatted, err := internalOutput.FormatOutput(points, "jsonl")
	if err != nil {
		t.Fatalf("failed to format timeseries: %v", err)
	}

	if strings.TrimSpace(cliOutput) != strings.TrimSpace(formatted) {
		t.Fatalf("library time series mismatch. CLI: %s, Library: %s", cliOutput, formatted)
	}
}

func TestProjectStateLibraryMatchesCLI(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())

	projectPath := t.TempDir()
	sessionID := "library-cli-project_state"
	fixture := `{"type":"assistant","timestamp":"2025-10-02T10:00:00Z","uuid":"uuid-1","sessionId":"` + sessionID + `","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/tmp/file.txt"}}]}}
{"type":"assistant","timestamp":"2025-10-02T10:01:00Z","uuid":"uuid-2","sessionId":"` + sessionID + `","message":{"role":"assistant","content":[{"type":"text","text":"Completed phase"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixture)

	cliOutput := runCLICommand(t, "query", "project-state", "--project", projectPath, "--output", "jsonl")

	entries, _ := loadEntriesAndToolCalls(t, projectPath)
	state := querypkg.BuildProjectState(entries, querypkg.ProjectStateOptions{IncludeIncomplete: true})
	formatted, err := pkgoutput.FormatJSONL([]interface{}{state})
	if err != nil {
		t.Fatalf("failed to format project state: %v", err)
	}

	if !equalJSON(cliOutput, formatted) {
		t.Fatalf("library project state mismatch. CLI: %s, Library: %s", cliOutput, formatted)
	}
}

func TestSuccessfulPromptsLibraryMatchesCLI(t *testing.T) {
	t.Setenv("META_CC_PROJECTS_ROOT", t.TempDir())

	projectPath := t.TempDir()
	sessionID := "library-cli-prompts"
	fixture := `{"type":"user","timestamp":"2025-10-02T10:00:00Z","uuid":"uuid-1","sessionId":"` + sessionID + `","message":{"role":"user","content":[{"type":"text","text":"Please fix bug"}]}}
{"type":"assistant","timestamp":"2025-10-02T10:01:00Z","uuid":"uuid-2","sessionId":"` + sessionID + `","message":{"role":"assistant","content":[{"type":"text","text":"Completed task"}]}}
`
	writeSessionFixture(t, projectPath, sessionID, fixture)

	cliOutput := runCLICommand(t, "query", "successful-prompts", "--project", projectPath)

	entries, _ := loadEntriesAndToolCalls(t, projectPath)
	result := querypkg.BuildSuccessfulPrompts(entries, 0.0, 0)
	items := make([]interface{}, 0, len(result.Prompts))
	for _, prompt := range result.Prompts {
		items = append(items, prompt)
	}
	formatted, err := pkgoutput.FormatJSONL(items)
	if err != nil {
		t.Fatalf("failed to format prompts: %v", err)
	}

	if strings.TrimSpace(cliOutput) != strings.TrimSpace(formatted) {
		t.Fatalf("library successful prompts mismatch. CLI: %s, Library: %s", cliOutput, formatted)
	}
}

func loadEntriesAndToolCalls(t *testing.T, projectPath string) ([]parser.SessionEntry, []parser.ToolCall) {
	opts := GlobalOptions{ProjectPath: projectPath}
	pipe := NewSessionPipeline(opts)
	if err := pipe.Load(LoadOptions{AutoDetect: true}); err != nil {
		t.Fatalf("failed to load pipeline: %v", err)
	}

	entries := pipe.GetEntries()
	toolCalls := pipe.ExtractToolCalls()
	return entries, toolCalls
}

func equalJSON(a, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)

	var ma, mb interface{}
	if err := json.Unmarshal([]byte(a), &ma); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &mb); err != nil {
		return false
	}
	return reflect.DeepEqual(ma, mb)
}

func runCLICommand(t *testing.T, args ...string) string {
	resetCLIState()

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs(args)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("CLI command failed: %v\noutput: %s", err, buf.String())
	}

	return strings.TrimSpace(buf.String())
}

func resetCLIState() {
	sessionID = ""
	projectPath = ""
	sessionOnly = false

	queryToolsStatus = ""
	queryToolsTool = ""
	queryToolsWhere = ""
	queryToolsFilter = ""

	queryLimit = 0
	queryOffset = 0
	querySortBy = ""
	queryReverse = false
	queryStream = false

	queryMessagesPattern = ""
	queryMessagesContext = 0

	rootCmd.PersistentFlags().Set("session", "")
	rootCmd.PersistentFlags().Set("project", "")
	rootCmd.PersistentFlags().Set("session-only", "false")
}
