# Subagent Prompt Template

## Usage

Copy this template and fill in the placeholders following the guidelines in reference/patterns.md.

## Template Structure

```markdown
---
name: {agent_name}
description: {one_line_task_description}
---

λ({input_params}) → {outputs} | {constraints}

## Dependencies (optional, if using Claude Code features)

agents_required :: [AgentType]
agents_required = [{agent1}, {agent2}, ...]

mcp_tools_required :: [ToolName]
mcp_tools_required = [{tool1}, {tool2}, ...]

skills_required :: [SkillName]
skills_required = [{skill1}, {skill2}, ...]

## Core Logic

{function_name_1} :: {InputType} → {OutputType}
{function_name_1}({params}) = {definition}

{function_name_2} :: {InputType} → {OutputType}
{function_name_2}({params}) = {definition}

...

## Execution Flow

{main_function} :: {MainInput} → {MainOutput}
{main_function}({params}) =
  {step_1} →
  {step_2} →
  ...
  {result}

## Constraints (optional)

constraints :: {ContextType} → Bool
constraints({ctx}) =
  {constraint_1} ∧ {constraint_2} ∧ ...

## Output (optional)

output :: {ResultType} → {Artifacts}
output({result}) = {artifact_definitions}
```

## Complexity Guidelines

| Complexity | Lines | Functions | Use For |
|------------|-------|-----------|---------|
| Simple | 30-60 | 3-5 | Single-purpose agents |
| Moderate | 60-120 | 5-8 | Multi-step workflows |
| Complex | 120-150 | 8-12 | Orchestration agents |

**Hard limit**: 150 lines

## Quality Checklist

### Before Construction
- [ ] Clear single purpose defined
- [ ] Input/output types specified
- [ ] Dependencies identified (agents/MCP/skills)
- [ ] Complexity assessed (simple/moderate/complex)

### During Construction
- [ ] Lambda contract defined
- [ ] Dependencies section (if applicable)
- [ ] Type signatures for all functions
- [ ] Symbolic logic used for constraints
- [ ] Function count: 3-12
- [ ] Line count: ≤150

### After Construction
- [ ] Compactness: Within target range
- [ ] Integration: Uses ≥1 Claude Code feature (if applicable)
- [ ] Clarity: Easy to understand flow
- [ ] Completeness: All cases handled
- [ ] Constraints: Explicitly defined
