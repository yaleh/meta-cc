---
name: subagent-prompt-construction
description: Systematic methodology for constructing compact, expressive, Claude Code-integrated subagent prompts using lambda-calculus and predicate logic syntax.
domain: Claude Code subagent development
validated: true
v_instance: 0.895
v_meta: 0.709
lines_max: 150
patterns: 3
templates: 1
examples: 1
---

λ(task_spec, complexity) → subagent_prompt |
  complexity ∈ {simple, moderate, complex} ∧
  |prompt| ≤ 150 ∧
  integration_score ≥ 0.75 ∧
  maintainability ≥ 0.85

## Dependencies

templates_required = [subagent-template.md]
reference_required = [patterns.md, symbolic-language.md, integration-patterns.md]
examples_required = [phase-planner-executor]

## Usage

apply :: TaskSpec → SubagentPrompt
apply(spec) =
  assess_complexity(spec) → complexity ∧
  select_template(complexity) → template ∧
  extract_requirements(spec) → requirements ∧
  apply_integration_patterns(requirements) → dependencies ∧
  decompose_functions(requirements) → functions ∧
  define_constraints(spec) → constraints ∧
  construct(template, dependencies, functions, constraints) → prompt ∧
  validate(prompt, quality_checklist) → validated_prompt

## Constraints

quality :: Prompt → Bool
quality(p) =
  |p| ≤ 150 ∧
  |functions(p)| ∈ [3, 12] ∧
  has_lambda_contract(p) ∧
  has_type_signatures(p) ∧
  uses_symbolic_logic(p) ∧
  integration_features(p) ≥ 1

## Artifacts

output :: ValidatedPrompt → Files
output(p) = {
  prompt: .claude/agents/{name}.md,
  validation_report: quality_metrics
}

## Validation

V_instance ≥ 0.80 ∧ V_meta ≥ 0.70 (achieved: 0.895, 0.709)
validated_with: phase-planner-executor (92 lines, 7 functions, 2 agents + 2 MCP tools)
