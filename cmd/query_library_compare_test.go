package cmd

import (
	"bytes"
	"strings"
	"testing"

	internalOutput "github.com/yaleh/meta-cc/internal/output"
	querypkg "github.com/yaleh/meta-cc/internal/query"
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
