package skillinsights

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestSkillFromSkillReadUsesStructuredPathInput(t *testing.T) {
	a := analyzer{knownSkills: map[string]bool{"paseo": true}}

	got, ok := a.skillFromSkillRead(toolCall{
		name: "Read",
		input: map[string]interface{}{
			"file_path": "/Users/example/.codex/skills/paseo/SKILL.md",
		},
	})

	if !ok {
		t.Fatal("expected skill file read to be detected")
	}
	if got != "paseo" {
		t.Fatalf("expected paseo, got %q", got)
	}
}

func TestSkillFromSkillReadSupportsExecCommandSedPath(t *testing.T) {
	a := analyzer{knownSkills: map[string]bool{"meta-cc-insights": true}}

	got, ok := a.skillFromSkillRead(toolCall{
		name: "functions.exec_command",
		input: map[string]interface{}{
			"cmd": "sed -n '1,120p' plugin-src/skills/meta-cc-insights/SKILL.md",
		},
	})

	if !ok {
		t.Fatal("expected exec command reading SKILL.md to be detected")
	}
	if got != "meta-cc-insights" {
		t.Fatalf("expected meta-cc-insights, got %q", got)
	}
}

func TestSkillFromSkillReadIgnoresExecCommandNonReadPathMentions(t *testing.T) {
	a := analyzer{knownSkills: map[string]bool{"paseo": true}}
	cases := []string{
		"git diff plugin-src/skills/paseo/SKILL.md",
		"rm plugin-src/skills/paseo/SKILL.md",
	}

	for _, cmd := range cases {
		got, ok := a.skillFromSkillRead(toolCall{
			name: "functions.exec_command",
			input: map[string]interface{}{
				"cmd": cmd,
			},
		})
		if ok {
			t.Fatalf("did not expect %q to be detected as skill read, got %q", cmd, got)
		}
	}
}

func TestCollectFilesMaxFileMBZeroMeansNoLimit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "large.jsonl")
	if err := os.WriteFile(path, []byte(`{"type":"user","message":{"role":"user","content":"hello"}}`+"\n"), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	files, skipped, err := collectFiles(Options{
		Roots:     []string{dir},
		MaxFileMB: 0,
		Since:     time.Time{},
	})
	if err != nil {
		t.Fatalf("collect files: %v", err)
	}
	if skipped != 0 {
		t.Fatalf("expected no skipped files, got %d", skipped)
	}
	if len(files) != 1 || files[0] != path {
		t.Fatalf("expected %q, got %#v", path, files)
	}
}

func TestTruncateKeepsUTF8Valid(t *testing.T) {
	got := truncate(strings.Repeat("审", 100), 73)
	if !utf8.ValidString(got) {
		t.Fatalf("truncate returned invalid UTF-8: %q", got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("expected ellipsis suffix, got %q", got)
	}
}

func TestBuildSkillStatsEffectRateAndScenarios(t *testing.T) {
	stats := buildSkillStats([]SkillUse{
		{Skill: "paseo", Scenario: "agent_orchestration", Effective: "effective"},
		{Skill: "paseo", Scenario: "agent_orchestration", Effective: "ineffective"},
		{Skill: "paseo", Scenario: "session_reflection", Effective: "unknown"},
	})

	if len(stats) != 1 {
		t.Fatalf("expected one stat, got %#v", stats)
	}
	if stats[0].Uses != 3 || stats[0].Effective != 1 || stats[0].Ineffective != 1 || stats[0].Unknown != 1 {
		t.Fatalf("unexpected counts: %#v", stats[0])
	}
	if stats[0].EffectRate != 0.5 {
		t.Fatalf("expected effect rate 0.5, got %v", stats[0].EffectRate)
	}
	if len(stats[0].TopScenarios) == 0 || stats[0].TopScenarios[0].Name != "agent_orchestration" {
		t.Fatalf("unexpected top scenarios: %#v", stats[0].TopScenarios)
	}
}

func TestAnalyzeDetectsSkillFileReadFromJSONL(t *testing.T) {
	dir := t.TempDir()
	skillRoot := filepath.Join(dir, "skills", "paseo")
	if err := os.MkdirAll(skillRoot, 0755); err != nil {
		t.Fatalf("mkdir skill root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillRoot, "SKILL.md"), []byte("# paseo\n"), 0644); err != nil {
		t.Fatalf("write skill: %v", err)
	}

	sessionPath := filepath.Join(dir, "session.jsonl")
	session := strings.Join([]string{
		`{"type":"user","timestamp":"2026-07-03T00:00:00Z","sessionId":"s1","message":{"role":"user","content":"Use paseo skill"}}`,
		fmt.Sprintf(`{"type":"assistant","timestamp":"2026-07-03T00:00:01Z","sessionId":"s1","message":{"role":"assistant","content":[{"type":"tool_use","id":"tool_1","name":"Read","input":{"file_path":%q}}]}}`, filepath.Join(skillRoot, "SKILL.md")),
	}, "\n")
	if err := os.WriteFile(sessionPath, []byte(session+"\n"), 0644); err != nil {
		t.Fatalf("write session: %v", err)
	}

	report, err := Analyze(Options{
		Roots:       []string{sessionPath},
		SkillRoots:  []string{filepath.Join(dir, "skills")},
		MaxFileMB:   25,
		MaxSamples:  5,
		KnownSkills: []string{"paseo"},
	})
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}

	if len(report.SkillUses) != 1 {
		t.Fatalf("expected one skill use, got %#v", report.SkillUses)
	}
	use := report.SkillUses[0]
	if use.Skill != "paseo" || use.TriggerType != "skill_file_read" || use.Confidence != 1.0 {
		t.Fatalf("unexpected skill use: %#v", use)
	}
}

func TestAnalyzeReportsMissedSkillCandidate(t *testing.T) {
	dir := t.TempDir()
	sessionPath := filepath.Join(dir, "session.jsonl")
	session := `{"type":"user","timestamp":"2026-07-03T00:00:00Z","sessionId":"s1","message":{"role":"user","content":"我总是频繁遇到同一个 skill 相关错误，帮我看看历史模式"}}`
	if err := os.WriteFile(sessionPath, []byte(session+"\n"), 0644); err != nil {
		t.Fatalf("write session: %v", err)
	}

	report, err := Analyze(Options{
		Roots:       []string{sessionPath},
		KnownSkills: []string{"meta-cc-insights"},
		MaxFileMB:   25,
	})
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}

	if len(report.MissedSkills) != 1 {
		t.Fatalf("expected one missed skill, got %#v", report.MissedSkills)
	}
	missed := report.MissedSkills[0]
	if missed.ExpectedSkill != "meta-cc-insights" || missed.Scenario != "session_reflection" {
		t.Fatalf("unexpected missed skill: %#v", missed)
	}
}

func TestAnalyzeToolErrorsAndMisuse(t *testing.T) {
	dir := t.TempDir()
	sessionPath := filepath.Join(dir, "session.jsonl")
	lines := []string{
		`{"type":"user","timestamp":"2026-07-03T00:00:00Z","sessionId":"s1","message":{"role":"user","content":"please use paseo skill"}}`,
		`{"type":"assistant","timestamp":"2026-07-03T00:00:01Z","sessionId":"s1","message":{"role":"assistant","content":[{"type":"text","text":"I will use paseo skill."},{"type":"tool_use","id":"tool_1","name":"mcp__paseo__create_agent","input":{"prompt":"paseo"}}]}}`,
		`{"type":"user","timestamp":"2026-07-03T00:00:02Z","sessionId":"s1","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool_1","content":"first error","is_error":true}]}}`,
		`{"type":"user","timestamp":"2026-07-03T00:00:03Z","sessionId":"s1","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool_1","content":"second error","is_error":true}]}}`,
		`{"type":"user","timestamp":"2026-07-03T00:00:04Z","sessionId":"s1","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool_1","content":"third error","is_error":true}]}}`,
	}
	if err := os.WriteFile(sessionPath, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		t.Fatalf("write session: %v", err)
	}

	report, err := Analyze(Options{
		Roots:       []string{sessionPath},
		KnownSkills: []string{"paseo"},
		MaxFileMB:   25,
	})
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}

	if len(report.ToolErrorStats) != 1 {
		t.Fatalf("expected one tool error stat, got %#v", report.ToolErrorStats)
	}
	if report.ToolErrorStats[0].Tool != "mcp__paseo__create_agent" || report.ToolErrorStats[0].Errors != 3 {
		t.Fatalf("unexpected tool errors: %#v", report.ToolErrorStats)
	}
	if len(report.PossibleMisuses) == 0 {
		t.Fatalf("expected possible misuse, got none")
	}
}

func TestCollectFilesFiltersBySizeAndMaxFiles(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "old.jsonl")
	newPath := filepath.Join(dir, "new.jsonl")
	largePath := filepath.Join(dir, "large.jsonl")
	for path, content := range map[string]string{
		oldPath:   "{}\n",
		newPath:   "{}\n",
		largePath: strings.Repeat("x", 2*1024*1024),
	} {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(oldPath, oldTime, oldTime); err != nil {
		t.Fatalf("chtimes old: %v", err)
	}

	files, skipped, err := collectFiles(Options{
		Roots:     []string{dir},
		Since:     time.Now().Add(-24 * time.Hour),
		MaxFileMB: 1,
		MaxFiles:  1,
	})
	if err != nil {
		t.Fatalf("collect files: %v", err)
	}
	if len(files) != 1 || files[0] != newPath {
		t.Fatalf("expected only %q, got %#v", newPath, files)
	}
	if skipped != 2 {
		t.Fatalf("expected two skipped files, got %d", skipped)
	}
}

func TestParseTimeAndProviderFallbacks(t *testing.T) {
	if parseTime("not a time").IsZero() == false {
		t.Fatal("expected invalid time to produce zero time")
	}
	if providerFromPath("/tmp/.codex/sessions/a.jsonl") != "codex" {
		t.Fatal("expected codex provider")
	}
	if providerFromPath("/tmp/.claude/projects/a.jsonl") != "claude" {
		t.Fatal("expected claude provider")
	}
	if providerFromPath("/tmp/other/a.jsonl") != "unknown" {
		t.Fatal("expected unknown provider")
	}
}

func TestScenarioClassification(t *testing.T) {
	cases := map[string]string{
		"please review this PR":          "code_review",
		"fix failing stack trace":        "debugging",
		"check UI screenshot in browser": "frontend_qa",
		"update README guide":            "documentation",
		"create a subagent worktree":     "agent_orchestration",
		"refactor this package":          "refactor",
		"add new feature":                "feature_dev",
	}
	for text, want := range cases {
		if got := classifyScenario(text, "", nil); got != want {
			t.Fatalf("classifyScenario(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestContainsAnyUsesTokenBoundaryForASCII(t *testing.T) {
	if containsAny("update README guide", "ui") {
		t.Fatal("did not expect ui to match inside guide")
	}
	if !containsAny("check ui layout", "ui") {
		t.Fatal("expected ui token to match")
	}
}

func TestExpectedSkillForScenarioCases(t *testing.T) {
	cases := []struct {
		scenario string
		text     string
		want     string
		ok       bool
	}{
		{scenario: "session_reflection", text: "history", want: "meta-cc-insights", ok: true},
		{scenario: "code_review", text: "review this PR", want: "open-code-review", ok: true},
		{scenario: "frontend_qa", text: "screenshot", want: "frontend-skill", ok: true},
		{scenario: "debugging", text: "bug", want: "cs-issue", ok: true},
		{scenario: "documentation", text: "docs", want: "cs-docs-neat", ok: true},
		{scenario: "agent_orchestration", text: "subagent", want: "paseo", ok: true},
		{scenario: "general", text: "github lookup", want: "agent-reach", ok: true},
		{scenario: "general", text: "plain task", ok: false},
	}

	for _, tc := range cases {
		got, _, ok := expectedSkillForScenario(tc.scenario, tc.text)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("expectedSkillForScenario(%q, %q) = (%q, %v), want (%q, %v)", tc.scenario, tc.text, got, ok, tc.want, tc.ok)
		}
	}
}

func TestSkillFromToolCallMappings(t *testing.T) {
	a := analyzer{knownSkills: map[string]bool{
		"meta-cc-insights": true,
		"paseo":            true,
		"agent-browser":    true,
		"agent-reach":      true,
		"cs-arch":          true,
		"lark-shared":      true,
	}}
	cases := []struct {
		call toolCall
		want string
	}{
		{call: toolCall{name: "mcp__meta_cc__query_tools"}, want: "meta-cc-insights"},
		{call: toolCall{name: "mcp__paseo__create_agent"}, want: "paseo"},
		{call: toolCall{name: "browser_navigate"}, want: "agent-browser"},
		{call: toolCall{name: "web_search"}, want: "agent-reach"},
		{call: toolCall{name: "mcp__archguard__summary"}, want: "cs-arch"},
		{call: toolCall{name: "mcp__lark__doc"}, want: "lark-shared"},
	}
	for _, tc := range cases {
		got, ok := a.skillFromToolCall(tc.call)
		if !ok || got != tc.want {
			t.Fatalf("skillFromToolCall(%q) = (%q, %v), want %q", tc.call.name, got, ok, tc.want)
		}
	}
}

func TestSortReportAndScenarioStats(t *testing.T) {
	report := Report{
		ToolErrorStats: []ToolErrorStat{{Tool: "b", Errors: 1}, {Tool: "a", Errors: 2}},
		SkillUses: []SkillUse{
			{Skill: "b", Scenario: "z", Timestamp: time.Unix(1, 0)},
			{Skill: "a", Scenario: "z", Timestamp: time.Unix(2, 0)},
		},
		MissedSkills: []MissedSkill{
			{ExpectedSkill: "low", ErrorCount: 0, Timestamp: time.Unix(2, 0)},
			{ExpectedSkill: "high", ErrorCount: 2, Timestamp: time.Unix(1, 0)},
		},
		PossibleMisuses: []PossibleMisuse{
			{Skill: "old", Timestamp: time.Unix(1, 0)},
			{Skill: "new", Timestamp: time.Unix(2, 0)},
		},
	}
	sortReport(&report)
	if report.ToolErrorStats[0].Tool != "a" || report.SkillUses[0].Skill != "a" || report.MissedSkills[0].ExpectedSkill != "high" || report.PossibleMisuses[0].Skill != "new" {
		t.Fatalf("unexpected sorted report: %#v", report)
	}

	stats := buildScenarioStats(report.SkillUses)
	if len(stats) != 2 || stats[0].Scenario != "z" {
		t.Fatalf("unexpected scenario stats: %#v", stats)
	}
}

func TestUtilityBranches(t *testing.T) {
	if !looksLikeError("command failed with error") {
		t.Fatal("expected error text to match")
	}
	if looksLikeError("all good") {
		t.Fatal("did not expect success text to match")
	}

	var report Report
	addToolError(&report, "", 1)
	addToolError(&report, "unknown", 2)
	if len(report.ToolErrorStats) != 1 || report.ToolErrorStats[0].Errors != 3 {
		t.Fatalf("unexpected tool error aggregation: %#v", report.ToolErrorStats)
	}

	got := uniqueStrings([]string{" paseo ", "paseo", "", "cs-issue"})
	if len(got) != 2 || got[0] != "paseo" || got[1] != "cs-issue" {
		t.Fatalf("unexpected unique strings: %#v", got)
	}

	if formatToolInput(nil) != "" {
		t.Fatal("expected empty formatted input")
	}
	candidates := inputPathCandidates(map[string]interface{}{
		"nested": map[string]interface{}{"path": "/x/skill/SKILL.md"},
		"list":   []interface{}{"/y/other/SKILL.md"},
	})
	if len(candidates) == 0 {
		t.Fatal("expected recursive path candidates")
	}
}
