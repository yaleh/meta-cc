package skillinsights

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

func WriteMarkdown(w io.Writer, report Report, sampleLimit int) {
	if sampleLimit <= 0 {
		sampleLimit = 8
	}
	fmt.Fprintf(w, "# Skill Insights MVP\n\n")
	fmt.Fprintf(w, "- Generated: %s\n", formatTime(report.GeneratedAt))
	if !report.Since.IsZero() {
		fmt.Fprintf(w, "- Since: %s\n", formatTime(report.Since))
	}
	fmt.Fprintf(w, "- Files scanned: %d\n", report.FilesScanned)
	if report.FilesSkipped > 0 {
		fmt.Fprintf(w, "- Files skipped by filters: %d\n", report.FilesSkipped)
	}
	fmt.Fprintf(w, "- Sessions scanned: %d\n", report.SessionsScanned)
	fmt.Fprintf(w, "- Skill uses detected: %d\n", len(report.SkillUses))
	fmt.Fprintf(w, "- Missed skill candidates: %d\n", len(report.MissedSkills))
	fmt.Fprintf(w, "- Possible misuses: %d\n", len(report.PossibleMisuses))
	if report.FilesWithErrors > 0 {
		fmt.Fprintf(w, "- Files with parse/read errors: %d\n", report.FilesWithErrors)
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "## 高频 Skills")
	if len(report.SkillStats) == 0 {
		fmt.Fprintln(w, "No skill use detected.")
	} else {
		fmt.Fprintln(w, "| Skill | Uses | Effective | Ineffective | Unknown | Effect rate | Top scenarios |")
		fmt.Fprintln(w, "| --- | ---: | ---: | ---: | ---: | ---: | --- |")
		for _, stat := range report.SkillStats {
			fmt.Fprintf(w, "| %s | %d | %d | %d | %d | %.0f%% | %s |\n",
				esc(stat.Skill), stat.Uses, stat.Effective, stat.Ineffective, stat.Unknown,
				stat.EffectRate*100, esc(formatCounts(stat.TopScenarios)))
		}
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "## 场景 x Skill")
	if len(report.ScenarioStats) == 0 {
		fmt.Fprintln(w, "No scenario-skill pairs detected.")
	} else {
		fmt.Fprintln(w, "| Scenario | Skill | Uses |")
		fmt.Fprintln(w, "| --- | --- | ---: |")
		for _, stat := range report.ScenarioStats {
			fmt.Fprintf(w, "| %s | %s | %d |\n", esc(stat.Scenario), esc(stat.Skill), stat.Uses)
		}
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "## 疑似有效 Skill 使用样例")
	writeUseSamples(w, report.SkillUses, "effective", sampleLimit)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "## 疑似用错或无效")
	if len(report.PossibleMisuses) == 0 {
		fmt.Fprintln(w, "No possible misuse detected by the MVP rules.")
	} else {
		fmt.Fprintln(w, "| Time | Skill | Scenario | Reason | Evidence | File |")
		fmt.Fprintln(w, "| --- | --- | --- | --- | --- | --- |")
		for i, misuse := range report.PossibleMisuses {
			if i >= sampleLimit {
				break
			}
			fmt.Fprintf(w, "| %s | %s | %s | %s | %s | %s |\n",
				formatTime(misuse.Timestamp), esc(misuse.Skill), esc(misuse.Scenario),
				esc(misuse.Reason), esc(misuse.Evidence), esc(shortFile(misuse.File)))
		}
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "## 本应使用 Skill 但疑似遗漏")
	if len(report.MissedSkills) == 0 {
		fmt.Fprintln(w, "No missed skill candidates detected by the MVP rules.")
	} else {
		fmt.Fprintln(w, "| Time | Expected skill | Scenario | Errors nearby | Confidence | Reason | Evidence | File |")
		fmt.Fprintln(w, "| --- | --- | --- | ---: | ---: | --- | --- | --- |")
		for i, missed := range report.MissedSkills {
			if i >= sampleLimit {
				break
			}
			fmt.Fprintf(w, "| %s | %s | %s | %d | %.2f | %s | %s | %s |\n",
				formatTime(missed.Timestamp), esc(missed.ExpectedSkill), esc(missed.Scenario),
				missed.ErrorCount, missed.Confidence, esc(missed.Reason),
				esc(missed.Evidence), esc(shortFile(missed.File)))
		}
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "## 高频工具错误")
	if len(report.ToolErrorStats) == 0 {
		fmt.Fprintln(w, "No tool errors detected.")
	} else {
		fmt.Fprintln(w, "| Tool | Errors |")
		fmt.Fprintln(w, "| --- | ---: |")
		for i, stat := range report.ToolErrorStats {
			if i >= sampleLimit {
				break
			}
			fmt.Fprintf(w, "| %s | %d |\n", esc(stat.Tool), stat.Errors)
		}
	}
	if len(report.UnreadableSamples) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "## 解析失败样例")
		for _, sample := range report.UnreadableSamples {
			fmt.Fprintf(w, "- `%s`: %s\n", sample.File, sample.Error)
		}
	}
}

func writeUseSamples(w io.Writer, uses []SkillUse, target string, limit int) {
	count := 0
	fmt.Fprintln(w, "| Time | Skill | Scenario | Trigger | Confidence | Evidence | File |")
	fmt.Fprintln(w, "| --- | --- | --- | --- | ---: | --- | --- |")
	for _, use := range uses {
		if use.Effective != target {
			continue
		}
		if count >= limit {
			break
		}
		count++
		fmt.Fprintf(w, "| %s | %s | %s | %s | %.2f | %s | %s |\n",
			formatTime(use.Timestamp), esc(use.Skill), esc(use.Scenario),
			esc(use.TriggerType), use.Confidence, esc(use.Evidence), esc(shortFile(use.File)))
	}
	if count == 0 {
		fmt.Fprintln(w, "| | | | | | No effective samples detected by the MVP rules. | |")
	}
}

func formatCounts(counts []Count) string {
	var parts []string
	for _, count := range counts {
		parts = append(parts, fmt.Sprintf("%s:%d", count.Name, count.Count))
	}
	return strings.Join(parts, ", ")
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}

func shortFile(path string) string {
	if path == "" {
		return ""
	}
	parent := filepath.Base(filepath.Dir(path))
	base := filepath.Base(path)
	if parent == "." || parent == "/" {
		return base
	}
	return filepath.Join(parent, base)
}

func esc(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
