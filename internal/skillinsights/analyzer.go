package skillinsights

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/yaleh/meta-cc/internal/parser"
	"github.com/yaleh/meta-cc/internal/types"
)

var skillReadPattern = regexp.MustCompile(`(?i)(?:^|/)([^/\s]+)/(?:SKILL\.md)$`)

type analyzer struct {
	knownSkills map[string]bool
}

type sessionFacts struct {
	file       string
	provider   string
	sessionID  string
	entries    []types.SessionEntry
	toolErrors map[string]int
	skillUses  []SkillUse
	missed     []MissedSkill
	misuses    []PossibleMisuse
	firstTime  time.Time
	lastTime   time.Time
	parseErr   error
}

type turn struct {
	idx       int
	entry     types.SessionEntry
	timestamp time.Time
	role      string
	text      string
	tools     []toolCall
	results   []toolResult
}

type toolCall struct {
	id    string
	name  string
	input map[string]interface{}
}

type toolResult struct {
	id      string
	content string
	isError bool
}

func Analyze(opts Options) (Report, error) {
	if opts.MaxSamples <= 0 {
		opts.MaxSamples = 10
	}
	a := analyzer{knownSkills: discoverKnownSkills(opts)}
	files, skipped, err := collectFiles(opts)
	if err != nil {
		return Report{}, err
	}

	report := Report{
		GeneratedAt:  time.Now(),
		Since:        opts.Since,
		FilesSkipped: skipped,
	}

	for _, file := range files {
		facts := a.analyzeFile(file, opts.Since)
		report.FilesScanned++
		if facts.parseErr != nil {
			report.FilesWithErrors++
			if len(report.UnreadableSamples) < opts.MaxSamples {
				report.UnreadableSamples = append(report.UnreadableSamples, FileError{File: file, Error: facts.parseErr.Error()})
			}
			continue
		}
		if len(facts.entries) == 0 {
			continue
		}
		report.SessionsScanned++
		report.SkillUses = append(report.SkillUses, facts.skillUses...)
		report.MissedSkills = append(report.MissedSkills, facts.missed...)
		report.PossibleMisuses = append(report.PossibleMisuses, facts.misuses...)
		for tool, count := range facts.toolErrors {
			addToolError(&report, tool, count)
		}
	}

	report.SkillStats = buildSkillStats(report.SkillUses)
	report.ScenarioStats = buildScenarioStats(report.SkillUses)
	sortReport(&report)
	return report, nil
}

func collectFiles(opts Options) ([]string, int, error) {
	roots := opts.Roots
	if len(roots) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, 0, err
		}
		roots = []string{
			filepath.Join(home, ".claude", "projects"),
			filepath.Join(home, ".codex", "sessions"),
		}
	}

	var files []string
	skipped := 0
	for _, root := range roots {
		if root == "" {
			continue
		}
		info, err := os.Stat(root)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			if strings.HasSuffix(root, ".jsonl") && keepByTime(info.ModTime(), opts.Since) && keepBySize(info.Size(), opts.MaxFileMB) {
				files = append(files, root)
			} else {
				skipped++
			}
			continue
		}
		err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".jsonl") {
				return nil
			}
			info, statErr := d.Info()
			if statErr == nil {
				if !keepByTime(info.ModTime(), opts.Since) || !keepBySize(info.Size(), opts.MaxFileMB) {
					skipped++
					return nil
				}
			}
			files = append(files, path)
			return nil
		})
		if err != nil {
			return nil, skipped, err
		}
	}
	sort.Slice(files, func(i, j int) bool {
		ai, _ := os.Stat(files[i])
		aj, _ := os.Stat(files[j])
		if ai == nil || aj == nil {
			return files[i] < files[j]
		}
		return ai.ModTime().After(aj.ModTime())
	})
	if opts.MaxFiles > 0 && len(files) > opts.MaxFiles {
		skipped += len(files) - opts.MaxFiles
		files = files[:opts.MaxFiles]
	}
	return files, skipped, nil
}

func keepByTime(mod time.Time, since time.Time) bool {
	return since.IsZero() || mod.After(since) || mod.Equal(since)
}

func keepBySize(size int64, maxFileMB int64) bool {
	if maxFileMB <= 0 {
		return true
	}
	return size <= maxFileMB*1024*1024
}

func (a analyzer) analyzeFile(file string, since time.Time) sessionFacts {
	facts := sessionFacts{
		file:       file,
		provider:   providerFromPath(file),
		toolErrors: map[string]int{},
	}
	entries, err := parser.NewSessionParser(file).ParseEntries()
	if err != nil {
		facts.parseErr = err
		return facts
	}
	for _, entry := range entries {
		ts := parseTime(entry.Timestamp)
		if !since.IsZero() && !ts.IsZero() && ts.Before(since) {
			continue
		}
		if facts.sessionID == "" {
			facts.sessionID = entry.SessionID
		}
		if facts.firstTime.IsZero() || (!ts.IsZero() && ts.Before(facts.firstTime)) {
			facts.firstTime = ts
		}
		if ts.After(facts.lastTime) {
			facts.lastTime = ts
		}
		facts.entries = append(facts.entries, entry)
	}
	if len(facts.entries) == 0 {
		return facts
	}

	turns := buildTurns(facts.entries)
	seenSkillInSession := map[string]bool{}
	for i := range turns {
		t := turns[i]
		for _, r := range t.results {
			if r.isError {
				tool := toolNameForResult(turns, r.id)
				facts.toolErrors[tool]++
			}
		}
		if t.role != "assistant" {
			continue
		}
		uses := a.detectSkillUses(facts, turns, i)
		for _, use := range uses {
			key := use.Skill + "|" + use.TriggerType + "|" + use.Evidence
			if seenSkillInSession[key] {
				continue
			}
			seenSkillInSession[key] = true
			use.ErrorCount = countErrorsNearby(turns, i, 8)
			use.Effective = effectiveness(turns, i)
			facts.skillUses = append(facts.skillUses, use)
		}
	}
	facts.missed = a.detectMissedSkills(facts, turns)
	facts.misuses = detectMisuses(facts.skillUses)
	return facts
}

func buildTurns(entries []types.SessionEntry) []turn {
	var out []turn
	for i, entry := range entries {
		if entry.Message == nil {
			continue
		}
		t := turn{
			idx:       i,
			entry:     entry,
			timestamp: parseTime(entry.Timestamp),
			role:      entry.Message.Role,
		}
		if t.role == "" {
			t.role = entry.Type
		}
		var texts []string
		for _, block := range entry.Message.Content {
			switch block.Type {
			case "text":
				if block.Text != "" {
					texts = append(texts, block.Text)
				}
			case "tool_use":
				if block.ToolUse != nil {
					t.tools = append(t.tools, toolCall{
						id:    block.ToolUse.ID,
						name:  block.ToolUse.Name,
						input: block.ToolUse.Input,
					})
				}
			case "tool_result":
				if block.ToolResult != nil {
					t.results = append(t.results, toolResult{
						id:      block.ToolResult.ToolUseID,
						content: block.ToolResult.Content,
						isError: block.ToolResult.IsError || block.ToolResult.Status == "error" || block.ToolResult.Error != "",
					})
					if block.ToolResult.Content != "" {
						texts = append(texts, block.ToolResult.Content)
					}
				}
			}
		}
		t.text = strings.Join(texts, "\n")
		out = append(out, t)
	}
	return out
}

func (a analyzer) detectSkillUses(facts sessionFacts, turns []turn, idx int) []SkillUse {
	t := turns[idx]
	scenario := classifyScenario(recentUserText(turns, idx), t.text, t.tools)
	base := SkillUse{
		Scenario:  scenario,
		Timestamp: t.timestamp,
		SessionID: facts.sessionID,
		Provider:  facts.provider,
		File:      facts.file,
	}
	var uses []SkillUse

	for _, call := range t.tools {
		if skill, ok := a.skillFromToolCall(call); ok {
			use := base
			use.Skill = skill
			use.TriggerType = "inferred_tool"
			use.Confidence = 0.65
			use.Evidence = truncate(call.name+" "+formatToolInput(call.input), 220)
			uses = append(uses, use)
		}
		if skill, ok := a.skillFromSkillRead(call); ok {
			use := base
			use.Skill = skill
			use.TriggerType = "skill_file_read"
			use.Confidence = 1.0
			use.Evidence = truncate(call.name+" "+formatToolInput(call.input), 220)
			uses = append(uses, use)
		}
	}

	for _, skill := range a.skillsFromText(t.text) {
		use := base
		use.Skill = skill
		use.TriggerType = "assistant_declared"
		use.Confidence = 0.85
		use.Evidence = truncate(t.text, 220)
		uses = append(uses, use)
	}
	for _, skill := range a.skillsFromText(recentUserText(turns, idx)) {
		use := base
		use.Skill = skill
		use.TriggerType = "explicit_user"
		use.Confidence = 0.8
		use.Evidence = truncate(recentUserText(turns, idx), 220)
		uses = append(uses, use)
	}
	return mergeSkillUses(uses)
}

func (a analyzer) skillFromToolCall(call toolCall) (string, bool) {
	name := strings.ToLower(call.name)
	input := strings.ToLower(formatToolInput(call.input))
	switch {
	case strings.Contains(name, "mcp__meta_cc") || strings.Contains(input, "meta-cc-insights"):
		return a.allowSkill("meta-cc-insights")
	case strings.Contains(name, "mcp__paseo") || strings.Contains(input, "paseo"):
		return a.allowSkill("paseo")
	case strings.Contains(name, "browser") || strings.Contains(name, "chrome") || strings.Contains(input, "playwright"):
		return a.allowSkill("agent-browser")
	case strings.Contains(name, "web") || strings.Contains(name, "search"):
		return a.allowSkill("agent-reach")
	case strings.Contains(name, "mcp__archguard"):
		return a.allowSkill("cs-arch")
	case strings.Contains(name, "mcp__") && strings.Contains(name, "lark"):
		return a.allowSkill("lark-shared")
	default:
		return "", false
	}
}

func (a analyzer) skillFromSkillRead(call toolCall) (string, bool) {
	if !isReadTool(call.name) {
		return "", false
	}
	if isExecTool(call.name) && !execCommandReadsFile(call.input) {
		return "", false
	}
	for _, value := range inputPathCandidates(call.input) {
		matches := skillReadPattern.FindStringSubmatch(value)
		if len(matches) >= 2 {
			return a.allowSkill(cleanSkillName(matches[1]))
		}
	}
	return "", false
}

func isReadTool(name string) bool {
	lower := strings.ToLower(name)
	return lower == "read" || lower == "view" || isExecTool(lower)
}

func isExecTool(name string) bool {
	lower := strings.ToLower(name)
	return lower == "functions.exec_command" || lower == "exec_command"
}

func execCommandReadsFile(input map[string]interface{}) bool {
	for _, cmd := range appendStringValues(nil, input["cmd"]) {
		if shellCommandReadsFile(cmd) {
			return true
		}
	}
	return false
}

func shellCommandReadsFile(cmd string) bool {
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return false
	}
	readCommands := map[string]bool{
		"awk":  true,
		"bat":  true,
		"cat":  true,
		"head": true,
		"less": true,
		"more": true,
		"nl":   true,
		"sed":  true,
		"tail": true,
	}
	for _, field := range fields {
		token := strings.Trim(field, "'\"")
		if token == "" || strings.Contains(token, "=") {
			continue
		}
		token = strings.TrimPrefix(token, "command")
		token = strings.TrimPrefix(token, "\\")
		token = filepath.Base(token)
		return readCommands[token]
	}
	return false
}

func (a analyzer) skillsFromText(text string) []string {
	lower := strings.ToLower(text)
	var found []string
	for skill := range a.knownSkills {
		if !strings.Contains(skill, "-") && !strings.Contains(skill, ":") {
			continue
		}
		if containsSkillName(lower, strings.ToLower(skill)) {
			found = append(found, skill)
		}
	}
	if strings.Contains(lower, " skill") || strings.Contains(lower, "skills") || strings.Contains(lower, "技能") {
		for _, token := range strings.Fields(strings.NewReplacer("`", " ", "$", " ", "，", " ", "。", " ", "：", " ", ":", " ").Replace(text)) {
			if looksLikeSkill(token) {
				if skill, ok := a.allowSkill(cleanSkillName(token)); ok {
					found = append(found, skill)
				}
			}
		}
	}
	return uniqueStrings(found)
}

func containsSkillName(text, skill string) bool {
	return strings.Contains(text, "$"+skill) ||
		strings.Contains(text, "`"+skill+"`") ||
		strings.Contains(text, " "+skill+" ") ||
		strings.Contains(text, "/"+skill+"/") ||
		strings.Contains(text, skill+" skill") ||
		strings.Contains(text, skill+" 技能")
}

func looksLikeSkill(s string) bool {
	s = strings.Trim(s, ".,;:()[]{}<>\"'")
	if len(s) < 3 || !strings.Contains(s, "-") {
		return false
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == ':' {
			continue
		}
		return false
	}
	return true
}

func cleanSkillName(s string) string {
	s = strings.Trim(s, ".,;:()[]{}<>\"'")
	return strings.TrimSpace(s)
}

func mergeSkillUses(in []SkillUse) []SkillUse {
	bySkill := map[string]SkillUse{}
	for _, use := range in {
		if use.Skill == "" {
			continue
		}
		current, ok := bySkill[use.Skill]
		if !ok || use.Confidence > current.Confidence {
			bySkill[use.Skill] = use
		}
	}
	out := make([]SkillUse, 0, len(bySkill))
	for _, use := range bySkill {
		out = append(out, use)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Timestamp.Equal(out[j].Timestamp) {
			return out[i].Skill < out[j].Skill
		}
		return out[i].Timestamp.Before(out[j].Timestamp)
	})
	return out
}

func (a analyzer) detectMissedSkills(facts sessionFacts, turns []turn) []MissedSkill {
	var missed []MissedSkill
	for i := range turns {
		t := turns[i]
		if t.role != "user" || len(t.results) > 0 {
			continue
		}
		text := t.text
		scenario := classifyScenario(text, "", nil)
		expected, reason, ok := expectedSkillForScenario(scenario, text)
		if !ok {
			continue
		}
		if a.skillUsedNear(turns, i, expected, 6) {
			continue
		}
		errCount := countErrorsNearby(turns, i, 8)
		if errCount == 0 && !strongMissTrigger(text, scenario) {
			continue
		}
		missed = append(missed, MissedSkill{
			ExpectedSkill: expected,
			Scenario:      scenario,
			Reason:        reason,
			Confidence:    missedConfidence(errCount, text),
			Timestamp:     t.timestamp,
			SessionID:     facts.sessionID,
			Provider:      facts.provider,
			File:          facts.file,
			Evidence:      truncate(text, 260),
			ErrorCount:    errCount,
		})
	}
	return missed
}

func detectMisuses(uses []SkillUse) []PossibleMisuse {
	var out []PossibleMisuse
	for _, use := range uses {
		if use.Effective != "ineffective" {
			continue
		}
		if use.Confidence < 0.65 {
			continue
		}
		out = append(out, PossibleMisuse{
			Skill:      use.Skill,
			Scenario:   use.Scenario,
			Reason:     "skill 使用后附近仍出现多次工具错误或失败结果，需要人工复核是否场景不匹配、执行顺序不对或缺少后续验证",
			Confidence: 0.55,
			Timestamp:  use.Timestamp,
			SessionID:  use.SessionID,
			Provider:   use.Provider,
			File:       use.File,
			Evidence:   use.Evidence,
		})
	}
	return out
}

func expectedSkillForScenario(scenario, text string) (string, string, bool) {
	lower := strings.ToLower(text)
	switch scenario {
	case "session_reflection":
		return "meta-cc-insights", "用户在问历史模式、频繁错误、工具/skill 使用或工作流反思", true
	case "code_review":
		if strings.Contains(lower, "pr") || strings.Contains(lower, "review") || strings.Contains(text, "审查") {
			return "open-code-review", "代码审查场景应该优先使用专门 review skill", true
		}
	case "frontend_qa":
		return "frontend-skill", "前端/UI/截图/浏览器验证场景应该使用前端或浏览器相关 skill", true
	case "debugging":
		return "cs-issue", "反复报错或修 bug 场景应该进入 issue/debugging skill 流程", true
	case "documentation":
		return "cs-docs-neat", "文档整理/编写场景应该使用文档 skill 约束结构和输出", true
	case "agent_orchestration":
		return "paseo", "创建、调度或评估子 agent 场景应该使用 paseo 相关 skill", true
	}
	if strings.Contains(lower, "github") || strings.Contains(lower, "internet") || strings.Contains(text, "联网") {
		return "agent-reach", "联网/GitHub/实时信息场景应该使用检索 skill 或代理策略", true
	}
	return "", "", false
}

func strongMissTrigger(text, scenario string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "skill") ||
		strings.Contains(text, "技能") ||
		strings.Contains(text, "频繁") ||
		strings.Contains(lower, "review") ||
		scenario == "session_reflection"
}

func missedConfidence(errCount int, text string) float64 {
	score := 0.45
	if errCount > 0 {
		score += 0.2
	}
	if errCount > 2 {
		score += 0.15
	}
	if strings.Contains(strings.ToLower(text), "skill") || strings.Contains(text, "技能") {
		score += 0.15
	}
	if score > 0.9 {
		return 0.9
	}
	return score
}

func (a analyzer) skillUsedNear(turns []turn, idx int, skill string, window int) bool {
	start := idx - window
	if start < 0 {
		start = 0
	}
	end := idx + window
	if end >= len(turns) {
		end = len(turns) - 1
	}
	for i := start; i <= end; i++ {
		t := turns[i]
		for _, call := range t.tools {
			if got, ok := a.skillFromToolCall(call); ok && got == skill {
				return true
			}
			if got, ok := a.skillFromSkillRead(call); ok && got == skill {
				return true
			}
		}
		for _, got := range a.skillsFromText(t.text) {
			if got == skill {
				return true
			}
		}
	}
	return false
}

func discoverKnownSkills(opts Options) map[string]bool {
	known := map[string]bool{}
	for _, skill := range opts.KnownSkills {
		if skill = cleanSkillName(skill); skill != "" {
			known[skill] = true
		}
	}
	roots := opts.SkillRoots
	if len(roots) == 0 {
		if home, err := os.UserHomeDir(); err == nil {
			roots = []string{
				filepath.Join(home, ".codex", "skills"),
				filepath.Join(home, ".agents", "skills"),
				filepath.Join(home, ".codex", "skills", ".system"),
				filepath.Join(home, ".nacos-cli", "skill-sync", "profiles", "default", "skill-repo"),
				filepath.Join(home, ".codex", "plugins", "cache"),
			}
		}
	}
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || d.Name() != "SKILL.md" {
				return nil
			}
			name := filepath.Base(filepath.Dir(path))
			if name != "" {
				known[name] = true
			}
			return nil
		})
	}
	for _, fallback := range []string{"meta-cc-insights", "paseo", "agent-browser", "agent-reach", "frontend-skill", "open-code-review", "cs-issue", "cs-feat", "cs-refactor", "cs-code-review", "cs-arch", "cs-docs-neat", "lark-shared"} {
		known[fallback] = true
	}
	return known
}

func (a analyzer) allowSkill(skill string) (string, bool) {
	skill = cleanSkillName(skill)
	if skill == "" {
		return "", false
	}
	if a.knownSkills[skill] {
		return skill, true
	}
	return "", false
}

func classifyScenario(userText, assistantText string, tools []toolCall) string {
	text := strings.ToLower(userText + "\n" + assistantText)
	for _, call := range tools {
		text += "\n" + strings.ToLower(call.name+" "+formatToolInput(call.input))
	}
	switch {
	case containsAny(text, "session", "history", "timeline", "pattern", "reflect", "workflow", "meta-cc", "会话", "历史", "反思", "模式", "频繁", "skills", "skill"):
		return "session_reflection"
	case containsAny(text, "review", "pr", "审查", "代码评审"):
		return "code_review"
	case containsAny(text, "bug", "error", "stack trace", "failing", "debug", "报错", "修复", "失败"):
		return "debugging"
	case containsAny(text, "frontend", "ui", "screenshot", "browser", "playwright", "css", "页面", "截图", "前端"):
		return "frontend_qa"
	case containsAny(text, "doc", "readme", "guide", "教程", "文档"):
		return "documentation"
	case containsAny(text, "agent", "subagent", "paseo", "schedule", "worktree", "代理", "子任务"):
		return "agent_orchestration"
	case containsAny(text, "refactor", "重构", "清理"):
		return "refactor"
	case containsAny(text, "implement", "feature", "build", "add ", "实现", "新增", "开发"):
		return "feature_dev"
	default:
		return "general"
	}
}

func containsAny(text string, needles ...string) bool {
	for _, needle := range needles {
		if containsToken(text, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

func containsToken(text, needle string) bool {
	if needle == "" {
		return false
	}
	if hasCJK(needle) || strings.Contains(needle, " ") {
		return strings.Contains(text, needle)
	}
	start := 0
	for {
		idx := strings.Index(text[start:], needle)
		if idx < 0 {
			return false
		}
		pos := start + idx
		end := pos + len(needle)
		if isTokenBoundary(text, pos-1) && isTokenBoundary(text, end) {
			return true
		}
		start = end
	}
}

func isTokenBoundary(s string, idx int) bool {
	if idx < 0 || idx >= len(s) {
		return true
	}
	ch := s[idx]
	return !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-')
}

func hasCJK(s string) bool {
	for _, r := range s {
		if (r >= '\u4e00' && r <= '\u9fff') || (r >= '\u3400' && r <= '\u4dbf') {
			return true
		}
	}
	return false
}

func recentUserText(turns []turn, idx int) string {
	var parts []string
	for i := idx - 1; i >= 0 && len(parts) < 3; i-- {
		if turns[i].role == "user" && len(turns[i].results) == 0 && strings.TrimSpace(turns[i].text) != "" {
			parts = append(parts, turns[i].text)
		}
	}
	return strings.Join(parts, "\n")
}

func countErrorsNearby(turns []turn, idx int, window int) int {
	end := idx + window
	if end >= len(turns) {
		end = len(turns) - 1
	}
	count := 0
	for i := idx; i <= end; i++ {
		for _, result := range turns[i].results {
			if result.isError || looksLikeError(result.content) {
				count++
			}
		}
	}
	return count
}

func effectiveness(turns []turn, idx int) string {
	errors := countErrorsNearby(turns, idx, 8)
	if errors >= 3 {
		return "ineffective"
	}
	if errors == 0 && hasProgressAfter(turns, idx, 10) {
		return "effective"
	}
	return "unknown"
}

func hasProgressAfter(turns []turn, idx int, window int) bool {
	end := idx + window
	if end >= len(turns) {
		end = len(turns) - 1
	}
	for i := idx + 1; i <= end; i++ {
		text := strings.ToLower(turns[i].text)
		if containsAny(text, "done", "completed", "fixed", "verified", "success", "已完成", "修复", "验证", "通过") {
			return true
		}
	}
	return false
}

func looksLikeError(text string) bool {
	lower := strings.ToLower(text)
	return containsAny(lower, "error", "failed", "exception", "permission denied", "no such file", "exit code", "panic", "timeout", "报错", "失败")
}

func toolNameForResult(turns []turn, id string) string {
	if id == "" {
		return "unknown"
	}
	for _, t := range turns {
		for _, call := range t.tools {
			if call.id == id {
				if call.name == "" {
					return "unknown"
				}
				return call.name
			}
		}
	}
	return "unknown"
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.000Z07:00"}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, s); err == nil {
			return ts
		}
	}
	return time.Time{}
}

func providerFromPath(path string) string {
	switch {
	case strings.Contains(path, "/.codex/"):
		return "codex"
	case strings.Contains(path, "/.claude/"):
		return "claude"
	default:
		return "unknown"
	}
}

func addToolError(report *Report, tool string, count int) {
	if tool == "" {
		tool = "unknown"
	}
	for i := range report.ToolErrorStats {
		if report.ToolErrorStats[i].Tool == tool {
			report.ToolErrorStats[i].Errors += count
			return
		}
	}
	report.ToolErrorStats = append(report.ToolErrorStats, ToolErrorStat{Tool: tool, Errors: count})
}

func buildSkillStats(uses []SkillUse) []SkillStat {
	stats := map[string]*SkillStat{}
	scenarios := map[string]map[string]int{}
	for _, use := range uses {
		if use.Skill == "" {
			continue
		}
		stat := stats[use.Skill]
		if stat == nil {
			stat = &SkillStat{Skill: use.Skill}
			stats[use.Skill] = stat
		}
		stat.Uses++
		switch use.Effective {
		case "effective":
			stat.Effective++
		case "ineffective":
			stat.Ineffective++
		default:
			stat.Unknown++
		}
		if scenarios[use.Skill] == nil {
			scenarios[use.Skill] = map[string]int{}
		}
		scenarios[use.Skill][use.Scenario]++
	}
	out := make([]SkillStat, 0, len(stats))
	for skill, stat := range stats {
		if stat.Effective+stat.Ineffective > 0 {
			stat.EffectRate = float64(stat.Effective) / float64(stat.Effective+stat.Ineffective)
		}
		stat.TopScenarios = topCounts(scenarios[skill], 3)
		out = append(out, *stat)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Uses == out[j].Uses {
			return out[i].Skill < out[j].Skill
		}
		return out[i].Uses > out[j].Uses
	})
	return out
}

func buildScenarioStats(uses []SkillUse) []ScenarioSkillStat {
	counts := map[string]int{}
	for _, use := range uses {
		if use.Skill == "" || use.Scenario == "" {
			continue
		}
		counts[use.Scenario+"|"+use.Skill]++
	}
	out := make([]ScenarioSkillStat, 0, len(counts))
	for key, count := range counts {
		parts := strings.SplitN(key, "|", 2)
		out = append(out, ScenarioSkillStat{Scenario: parts[0], Skill: parts[1], Uses: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Scenario == out[j].Scenario {
			if out[i].Uses == out[j].Uses {
				return out[i].Skill < out[j].Skill
			}
			return out[i].Uses > out[j].Uses
		}
		return out[i].Scenario < out[j].Scenario
	})
	return out
}

func topCounts(counts map[string]int, limit int) []Count {
	out := make([]Count, 0, len(counts))
	for name, count := range counts {
		out = append(out, Count{Name: name, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count == out[j].Count {
			return out[i].Name < out[j].Name
		}
		return out[i].Count > out[j].Count
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func sortReport(report *Report) {
	sort.Slice(report.ToolErrorStats, func(i, j int) bool {
		if report.ToolErrorStats[i].Errors == report.ToolErrorStats[j].Errors {
			return report.ToolErrorStats[i].Tool < report.ToolErrorStats[j].Tool
		}
		return report.ToolErrorStats[i].Errors > report.ToolErrorStats[j].Errors
	})
	sort.Slice(report.SkillUses, func(i, j int) bool {
		return report.SkillUses[i].Timestamp.After(report.SkillUses[j].Timestamp)
	})
	sort.Slice(report.MissedSkills, func(i, j int) bool {
		if report.MissedSkills[i].ErrorCount == report.MissedSkills[j].ErrorCount {
			return report.MissedSkills[i].Timestamp.After(report.MissedSkills[j].Timestamp)
		}
		return report.MissedSkills[i].ErrorCount > report.MissedSkills[j].ErrorCount
	})
	sort.Slice(report.PossibleMisuses, func(i, j int) bool {
		return report.PossibleMisuses[i].Timestamp.After(report.PossibleMisuses[j].Timestamp)
	})
}

func uniqueStrings(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		s = cleanSkillName(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func truncate(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

func formatToolInput(input map[string]interface{}) string {
	if len(input) == 0 {
		return ""
	}
	data, err := json.Marshal(input)
	if err != nil {
		return fmt.Sprint(input)
	}
	return string(data)
}

func inputPathCandidates(input map[string]interface{}) []string {
	keys := []string{"file_path", "path", "filename", "file", "uri"}
	var out []string
	for _, key := range keys {
		if value, ok := input[key]; ok {
			out = appendStringValues(out, value)
		}
	}
	if len(out) == 0 {
		out = appendStringValues(out, input)
	}
	return out
}

func appendStringValues(out []string, value interface{}) []string {
	switch v := value.(type) {
	case string:
		if v != "" {
			out = append(out, v)
		}
	case []string:
		for _, item := range v {
			if item != "" {
				out = append(out, item)
			}
		}
	case []interface{}:
		for _, item := range v {
			out = appendStringValues(out, item)
		}
	case map[string]interface{}:
		for _, item := range v {
			out = appendStringValues(out, item)
		}
	}
	return out
}
