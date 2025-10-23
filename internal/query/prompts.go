package query

import (
	"math"
	"sort"
	"strings"

	"github.com/yaleh/meta-cc/internal/parser"
)

type SuccessfulPromptsResult struct {
	Prompts []SuccessfulPrompt `json:"prompts"`
}

type SuccessfulPrompt struct {
	TurnSequence    int             `json:"turn_sequence"`
	UserPrompt      string          `json:"user_prompt"`
	Context         PromptContext   `json:"context"`
	Outcome         PromptOutcome   `json:"outcome"`
	QualityScore    float64         `json:"quality_score"`
	PatternFeatures PatternFeatures `json:"pattern_features"`
}

type PromptContext struct {
	Phase    string `json:"phase,omitempty"`
	TaskType string `json:"task_type,omitempty"`
}

type PromptOutcome struct {
	Status          string   `json:"status"`
	TurnsToComplete int      `json:"turns_to_complete"`
	ErrorCount      int      `json:"error_count"`
	Deliverables    []string `json:"deliverables,omitempty"`
}

type PatternFeatures struct {
	HasClearGoal          bool `json:"has_clear_goal"`
	HasConstraints        bool `json:"has_constraints"`
	HasAcceptanceCriteria bool `json:"has_acceptance_criteria"`
	HasContext            bool `json:"has_context"`
}

func BuildSuccessfulPrompts(entries []parser.SessionEntry, minQuality float64, limit int) *SuccessfulPromptsResult {
	turnIndex := buildTurnIndex(entries)
	var prompts []SuccessfulPrompt

	for i, entry := range entries {
		if entry.Type != "user" || entry.Message == nil {
			continue
		}

		var promptText strings.Builder
		for _, block := range entry.Message.Content {
			if block.Type == "text" {
				promptText.WriteString(block.Text)
			}
		}
		text := strings.TrimSpace(promptText.String())
		if text == "" {
			continue
		}

		turn := turnIndex[entry.UUID]
		outcome, _ := analyzePromptOutcome(entries, i, turnIndex)
		quality := calculateQualityScore(outcome, text)
		if quality < minQuality {
			continue
		}

		prompts = append(prompts, SuccessfulPrompt{
			TurnSequence:    turn,
			UserPrompt:      text,
			Context:         extractPromptContext(text),
			Outcome:         outcome,
			QualityScore:    quality,
			PatternFeatures: extractPatternFeatures(text),
		})
	}

	sortPrompts(prompts)
	if limit > 0 && len(prompts) > limit {
		prompts = prompts[:limit]
	}

	return &SuccessfulPromptsResult{Prompts: prompts}
}

func analyzePromptOutcome(entries []parser.SessionEntry, index int, turnIndex map[string]int) (PromptOutcome, int) {
	delivered := make(map[string]struct{})
	errorCount := 0
	turns := 0
	status := "in-progress"

	for j := index + 1; j < len(entries); j++ {
		entry := entries[j]
		if entry.Message == nil {
			continue
		}
		turns++

		if entry.Type == "assistant" {
			for _, block := range entry.Message.Content {
				if block.Type == "tool_result" && block.ToolResult != nil {
					if block.ToolResult.IsError {
						errorCount++
					} else if block.ToolResult.Content != "" {
						delivered[block.ToolResult.Content] = struct{}{}
					}
				}
				if block.Type == "text" && strings.Contains(strings.ToLower(block.Text), "completed") {
					status = "completed"
				}
			}
		}
	}

	deliverables := make([]string, 0, len(delivered))
	for item := range delivered {
		deliverables = append(deliverables, item)
	}

	return PromptOutcome{
		Status:          status,
		TurnsToComplete: turns,
		ErrorCount:      errorCount,
		Deliverables:    deliverables,
	}, turns
}

func calculateQualityScore(outcome PromptOutcome, prompt string) float64 {
	score := 1.0
	if outcome.ErrorCount > 0 {
		score *= 0.5
	}
	if outcome.TurnsToComplete > 4 {
		score *= 0.7
	}
	if len(outcome.Deliverables) == 0 {
		score *= 0.8
	}
	if len(prompt) > 300 {
		score *= 0.9
	}
	return math.Min(score, 1.0)
}

func extractPromptContext(prompt string) PromptContext {
	promptLower := strings.ToLower(prompt)
	phase := ""
	if strings.Contains(promptLower, "phase") {
		phase = "phase"
	}
	taskType := ""
	if strings.Contains(promptLower, "refactor") {
		taskType = "refactor"
	} else if strings.Contains(promptLower, "bug") {
		taskType = "bugfix"
	}
	return PromptContext{Phase: phase, TaskType: taskType}
}

func extractPatternFeatures(prompt string) PatternFeatures {
	lower := strings.ToLower(prompt)
	return PatternFeatures{
		HasClearGoal:          strings.Contains(lower, "goal") || strings.Contains(lower, "need"),
		HasConstraints:        strings.Contains(lower, "must") || strings.Contains(lower, "constraint"),
		HasAcceptanceCriteria: strings.Contains(lower, "acceptance") || strings.Contains(lower, "criteria"),
		HasContext:            strings.Contains(lower, "context") || strings.Contains(lower, "background"),
	}
}

func sortPrompts(prompts []SuccessfulPrompt) {
	sort.Slice(prompts, func(i, j int) bool {
		if prompts[i].QualityScore == prompts[j].QualityScore {
			return prompts[i].TurnSequence < prompts[j].TurnSequence
		}
		return prompts[i].QualityScore > prompts[j].QualityScore
	})
}
