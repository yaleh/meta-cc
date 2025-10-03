package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yale/meta-cc/internal/locator"
	"github.com/yale/meta-cc/internal/parser"
)

var (
	successfulPromptsMinQuality float64
)

// querySuccessfulPromptsCmd represents the successful-prompts query command
var querySuccessfulPromptsCmd = &cobra.Command{
	Use:   "successful-prompts",
	Short: "Query successful prompt patterns",
	Long: `Query successful prompt patterns from session history.

This command identifies user prompts that led to successful outcomes, based on:
- Fast completion (few turns)
- No errors during execution
- Clear deliverables
- User confirmation

Example:
  meta-cc query successful-prompts --limit 10 --output json
  meta-cc query successful-prompts --min-quality-score 0.8 --output md`,
	RunE: runQuerySuccessfulPrompts,
}

func init() {
	querySuccessfulPromptsCmd.Flags().Float64Var(&successfulPromptsMinQuality, "min-quality-score", 0.0, "Minimum quality score (0.0-1.0)")

	queryCmd.AddCommand(querySuccessfulPromptsCmd)
}

func runQuerySuccessfulPrompts(cmd *cobra.Command, args []string) error {
	// Locate session file
	loc := locator.NewSessionLocator()
	sessionPath, err := loc.Locate(locator.LocateOptions{
		SessionID:   sessionID,
		ProjectPath: projectPath,
	})
	if err != nil {
		return fmt.Errorf("failed to locate session: %w", err)
	}

	// Read and parse session file
	sessionParser := parser.NewSessionParser(sessionPath)
	entries, err := sessionParser.ParseEntries()
	if err != nil {
		return fmt.Errorf("failed to read session file: %w", err)
	}

	// Build successful prompts result
	result, err := buildSuccessfulPrompts(entries, successfulPromptsMinQuality, queryLimit)
	if err != nil {
		return fmt.Errorf("failed to build successful prompts: %w", err)
	}

	// Output result
	if outputFormat == "md" {
		return outputSuccessfulPromptsMarkdown(cmd, result)
	}

	// JSON output (default)
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// SuccessfulPromptsResult represents successful prompts query result
type SuccessfulPromptsResult struct {
	Prompts []SuccessfulPrompt `json:"prompts"`
}

// SuccessfulPrompt represents a successful prompt with metadata
type SuccessfulPrompt struct {
	TurnSequence    int              `json:"turn_sequence"`
	UserPrompt      string           `json:"user_prompt"`
	Context         PromptContext    `json:"context"`
	Outcome         PromptOutcome    `json:"outcome"`
	QualityScore    float64          `json:"quality_score"`
	PatternFeatures PatternFeatures  `json:"pattern_features"`
}

// PromptContext represents the context when prompt was given
type PromptContext struct {
	Phase    string `json:"phase,omitempty"`
	TaskType string `json:"task_type,omitempty"`
}

// PromptOutcome represents the outcome of the prompt
type PromptOutcome struct {
	Status           string   `json:"status"`
	TurnsToComplete  int      `json:"turns_to_complete"`
	ErrorCount       int      `json:"error_count"`
	Deliverables     []string `json:"deliverables,omitempty"`
}

// PatternFeatures represents structural features of the prompt
type PatternFeatures struct {
	HasClearGoal          bool `json:"has_clear_goal"`
	HasConstraints        bool `json:"has_constraints"`
	HasAcceptanceCriteria bool `json:"has_acceptance_criteria"`
	HasContext            bool `json:"has_context"`
}

func buildSuccessfulPrompts(entries []parser.SessionEntry, minQuality float64, limit int) (*SuccessfulPromptsResult, error) {
	// Build turn index
	turnIndex := buildTurnIndex(entries)

	// Extract user prompts
	var prompts []SuccessfulPrompt

	for i, entry := range entries {
		if entry.Type != "user" || entry.Message == nil {
			continue
		}

		// Extract user prompt content
		var promptText string
		for _, block := range entry.Message.Content {
			if block.Type == "text" {
				promptText += block.Text
			}
		}

		if promptText == "" {
			continue
		}

		turn := turnIndex[entry.UUID]

		// Analyze the outcome of this prompt
		outcome, _ := analyzePromptOutcome(entries, i, turnIndex)

		// Calculate quality score
		qualityScore := calculateQualityScore(outcome, promptText)

		// Skip if below minimum quality
		if qualityScore < minQuality {
			continue
		}

		// Extract context
		context := extractPromptContext(promptText)

		// Extract pattern features
		features := extractPatternFeatures(promptText)

		prompts = append(prompts, SuccessfulPrompt{
			TurnSequence:    turn,
			UserPrompt:      promptText,
			Context:         context,
			Outcome:         outcome,
			QualityScore:    qualityScore,
			PatternFeatures: features,
		})
	}

	// Sort by quality score (descending)
	for i := 0; i < len(prompts); i++ {
		for j := i + 1; j < len(prompts); j++ {
			if prompts[j].QualityScore > prompts[i].QualityScore {
				prompts[i], prompts[j] = prompts[j], prompts[i]
			}
		}
	}

	// Apply limit
	if limit > 0 && len(prompts) > limit {
		prompts = prompts[:limit]
	}

	return &SuccessfulPromptsResult{
		Prompts: prompts,
	}, nil
}

// analyzePromptOutcome analyzes the outcome of a user prompt
func analyzePromptOutcome(entries []parser.SessionEntry, userEntryIdx int, turnIndex map[string]int) (PromptOutcome, int) {
	outcome := PromptOutcome{
		Status:      "unknown",
		ErrorCount:  0,
		Deliverables: []string{},
	}

	startTurn := turnIndex[entries[userEntryIdx].UUID]
	endTurn := startTurn

	// Look ahead for assistant responses and next user message
	for i := userEntryIdx + 1; i < len(entries); i++ {
		entry := entries[i]

		// Stop at next user message
		if entry.Type == "user" {
			endTurn = turnIndex[entry.UUID] - 1

			// Check if user confirmed success
			if entry.Message != nil {
				var content string
				for _, block := range entry.Message.Content {
					if block.Type == "text" {
						content += block.Text
					}
				}

				content = strings.ToLower(content)
				if containsAny(content, []string{"good", "great", "perfect", "thanks", "好的", "很好", "完成", "通过"}) {
					outcome.Status = "success"
				}
			}
			break
		}

		// Count errors
		if entry.Message != nil {
			for _, block := range entry.Message.Content {
				if block.Type == "tool_result" && block.ToolResult != nil {
					if block.ToolResult.Status == "error" || block.ToolResult.Error != "" {
						outcome.ErrorCount++
					}
				}
			}
		}

		// Extract deliverables (files created/modified)
		if entry.Message != nil {
			for _, block := range entry.Message.Content {
				if block.Type == "tool_use" && block.ToolUse != nil {
					if block.ToolUse.Name == "Write" || block.ToolUse.Name == "Edit" {
						if filePath, ok := block.ToolUse.Input["file_path"].(string); ok {
							if !contains(outcome.Deliverables, filePath) {
								outcome.Deliverables = append(outcome.Deliverables, filePath)
							}
						}
					}
				}
			}
		}

		endTurn = turnIndex[entry.UUID]
	}

	// Calculate turns to complete
	outcome.TurnsToComplete = endTurn - startTurn + 1

	// Determine status if not already set
	if outcome.Status == "unknown" {
		if outcome.ErrorCount == 0 && len(outcome.Deliverables) > 0 {
			outcome.Status = "success"
		} else if outcome.ErrorCount > 0 {
			outcome.Status = "partial"
		}
	}

	return outcome, endTurn
}

// calculateQualityScore calculates quality score for a prompt
func calculateQualityScore(outcome PromptOutcome, promptText string) float64 {
	score := 0.0

	// Error rate component (40%)
	if outcome.ErrorCount == 0 {
		score += 0.4
	} else {
		errorRate := float64(outcome.ErrorCount) / float64(outcome.TurnsToComplete)
		score += 0.4 * (1.0 - min(errorRate, 1.0))
	}

	// Speed component (30%)
	if outcome.TurnsToComplete <= 3 {
		score += 0.3
	} else if outcome.TurnsToComplete <= 5 {
		score += 0.2
	} else if outcome.TurnsToComplete <= 10 {
		score += 0.1
	}

	// Deliverable component (20%)
	if len(outcome.Deliverables) > 0 {
		score += 0.2
	}

	// Confirmation component (10%)
	if outcome.Status == "success" {
		score += 0.1
	}

	return score
}

// extractPromptContext extracts context from prompt text
func extractPromptContext(promptText string) PromptContext {
	context := PromptContext{}

	lower := strings.ToLower(promptText)

	// Detect phase
	if strings.Contains(lower, "phase") || strings.Contains(lower, "阶段") {
		context.Phase = "Phase-based development"
	} else if strings.Contains(lower, "stage") {
		context.Phase = "Stage-based development"
	}

	// Detect task type
	taskTypes := map[string][]string{
		"implementation": {"implement", "实现", "添加", "add", "create"},
		"bug_fix":        {"fix", "修复", "bug", "error"},
		"testing":        {"test", "测试", "验证"},
		"refactoring":    {"refactor", "重构", "optimize", "优化"},
		"documentation":  {"doc", "文档", "comment"},
	}

	for taskType, keywords := range taskTypes {
		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				context.TaskType = taskType
				break
			}
		}
		if context.TaskType != "" {
			break
		}
	}

	return context
}

// extractPatternFeatures extracts structural features from prompt
func extractPatternFeatures(promptText string) PatternFeatures {
	features := PatternFeatures{}

	lower := strings.ToLower(promptText)

	// Has clear goal (action verbs)
	actionVerbs := []string{"implement", "add", "fix", "create", "update", "实现", "添加", "修复", "创建"}
	for _, verb := range actionVerbs {
		if strings.Contains(lower, verb) {
			features.HasClearGoal = true
			break
		}
	}

	// Has constraints (mentions limits, requirements)
	constraintKeywords := []string{"constraint", "limit", "requirement", "must", "should", "约束", "限制", "要求"}
	for _, keyword := range constraintKeywords {
		if strings.Contains(lower, keyword) {
			features.HasConstraints = true
			break
		}
	}

	// Has acceptance criteria (mentions tests, verification)
	criteriaKeywords := []string{"test", "verify", "ensure", "validate", "测试", "验证", "确保"}
	for _, keyword := range criteriaKeywords {
		if strings.Contains(lower, keyword) {
			features.HasAcceptanceCriteria = true
			break
		}
	}

	// Has context (mentions phase, stage, background)
	contextKeywords := []string{"phase", "stage", "context", "background", "阶段", "背景", "上下文"}
	for _, keyword := range contextKeywords {
		if strings.Contains(lower, keyword) {
			features.HasContext = true
			break
		}
	}

	return features
}

// Helper functions
func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func outputSuccessfulPromptsMarkdown(cmd *cobra.Command, result *SuccessfulPromptsResult) error {
	var sb strings.Builder

	sb.WriteString("# Successful Prompt Patterns\n\n")
	sb.WriteString(fmt.Sprintf("**Total Prompts**: %d\n\n", len(result.Prompts)))

	for i, prompt := range result.Prompts {
		sb.WriteString(fmt.Sprintf("## Prompt %d (Turn %d, Quality: %.2f)\n\n", i+1, prompt.TurnSequence, prompt.QualityScore))

		// Prompt text
		sb.WriteString("### User Prompt\n\n")
		sb.WriteString(fmt.Sprintf("> %s\n\n", strings.ReplaceAll(prompt.UserPrompt, "\n", "\n> ")))

		// Context
		if prompt.Context.Phase != "" || prompt.Context.TaskType != "" {
			sb.WriteString("### Context\n\n")
			if prompt.Context.Phase != "" {
				sb.WriteString(fmt.Sprintf("- **Phase**: %s\n", prompt.Context.Phase))
			}
			if prompt.Context.TaskType != "" {
				sb.WriteString(fmt.Sprintf("- **Task Type**: %s\n", prompt.Context.TaskType))
			}
			sb.WriteString("\n")
		}

		// Outcome
		sb.WriteString("### Outcome\n\n")
		sb.WriteString(fmt.Sprintf("- **Status**: %s\n", prompt.Outcome.Status))
		sb.WriteString(fmt.Sprintf("- **Turns to Complete**: %d\n", prompt.Outcome.TurnsToComplete))
		sb.WriteString(fmt.Sprintf("- **Error Count**: %d\n", prompt.Outcome.ErrorCount))
		if len(prompt.Outcome.Deliverables) > 0 {
			sb.WriteString(fmt.Sprintf("- **Deliverables**: %s\n", strings.Join(prompt.Outcome.Deliverables, ", ")))
		}
		sb.WriteString("\n")

		// Pattern features
		sb.WriteString("### Pattern Features\n\n")
		sb.WriteString(fmt.Sprintf("- Clear Goal: %v\n", prompt.PatternFeatures.HasClearGoal))
		sb.WriteString(fmt.Sprintf("- Constraints: %v\n", prompt.PatternFeatures.HasConstraints))
		sb.WriteString(fmt.Sprintf("- Acceptance Criteria: %v\n", prompt.PatternFeatures.HasAcceptanceCriteria))
		sb.WriteString(fmt.Sprintf("- Context: %v\n\n", prompt.PatternFeatures.HasContext))

		sb.WriteString("---\n\n")
	}

	fmt.Fprint(cmd.OutOrStdout(), sb.String())
	return nil
}
