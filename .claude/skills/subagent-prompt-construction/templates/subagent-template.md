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
