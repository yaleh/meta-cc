package skillinsights

import (
	"strings"
	"testing"
	"time"
)

func TestWriteMarkdownEscapesTableContent(t *testing.T) {
	report := Report{
		GeneratedAt: time.Date(2026, 7, 3, 1, 2, 0, 0, time.UTC),
		SkillUses: []SkillUse{
			{
				Skill:       "paseo|unsafe",
				Scenario:    "agent_orchestration",
				TriggerType: "skill_file_read",
				Confidence:  1.0,
				Effective:   "effective",
				Evidence:    "line1\nline2 | pipe",
			},
		},
		SkillStats: []SkillStat{
			{
				Skill:        "paseo|unsafe",
				Uses:         1,
				Effective:    1,
				EffectRate:   1,
				TopScenarios: []Count{{Name: "agent_orchestration", Count: 1}},
			},
		},
		ScenarioStats: []ScenarioSkillStat{{Scenario: "agent_orchestration", Skill: "paseo|unsafe", Uses: 1}},
	}

	var b strings.Builder
	WriteMarkdown(&b, report, 3)
	got := b.String()

	if !strings.Contains(got, "paseo\\|unsafe") {
		t.Fatalf("expected pipe to be escaped, got:\n%s", got)
	}
	if strings.Contains(got, "line1\nline2") {
		t.Fatalf("expected evidence newline to be flattened, got:\n%s", got)
	}
}

func TestWriteMarkdownEmptyReportAndUnreadableSamples(t *testing.T) {
	report := Report{
		GeneratedAt:       time.Date(2026, 7, 3, 1, 2, 0, 0, time.UTC),
		FilesWithErrors:   1,
		UnreadableSamples: []FileError{{File: "/tmp/session.jsonl", Error: "bad json"}},
	}

	var b strings.Builder
	WriteMarkdown(&b, report, 0)
	got := b.String()

	for _, want := range []string{
		"No skill use detected.",
		"No scenario-skill pairs detected.",
		"No possible misuse detected by the MVP rules.",
		"No missed skill candidates detected by the MVP rules.",
		"No tool errors detected.",
		"解析失败样例",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected markdown to contain %q, got:\n%s", want, got)
		}
	}
}

func TestShortFile(t *testing.T) {
	if got := shortFile("/tmp/project/session.jsonl"); got != "project/session.jsonl" {
		t.Fatalf("unexpected short file: %q", got)
	}
	if got := shortFile("session.jsonl"); got != "session.jsonl" {
		t.Fatalf("unexpected short file: %q", got)
	}
	if got := shortFile(""); got != "" {
		t.Fatalf("unexpected empty short file: %q", got)
	}
}
