# Claude Code Integration Patterns

## Overview

Formal patterns for integrating Claude Code features (subagents, MCP tools, skills) into subagent prompts.

**Key Innovation**: Symbolic syntax for feature references enables compact, maintainable subagent definitions with explicit dependency tracking.

---

## Pattern 1: Subagent Composition

### Syntax

```
agent(type, description) :: Context → Output
```

### Semantics

```
agent(type, desc) =
  invoke_task_tool(
    subagent_type: type,
    prompt: desc
  ) ∧ await_completion → output
```

### Declaration

```
agents_required :: [AgentType]
agents_required = [agent1, agent2, ...]
```

### Usage Examples

**Simple Invocation**:
```
agents_required = [project-planner]

generate_plan :: Requirements → Plan
generate_plan(req) =
  agent(project-planner,
    "Create implementation plan for: ${req.objectives}"
  ) → plan
```

**With Context Interpolation**:
```
agents_required = [stage-executor]

execute_stage :: (Plan, N) → Result
execute_stage(plan, n) =
  stage = plan.stages[n] →
  agent(stage-executor,
    "Execute Stage ${n} using TDD:\n" +
    stage.description + "\n" +
    "Acceptance criteria:\n" + stage.criteria
  ) → result
```

**Sequential Composition**:
```
agents_required = [project-planner, stage-executor]

orchestrate :: Spec → Results
orchestrate(spec) =
  plan = agent(project-planner, spec.description) →
  results = [] →
  ∀stage ∈ plan.stages:
    result = agent(stage-executor, stage.spec) →
    results = results + [result] →
  results
```

### Best Practices

1. **Declare Dependencies**: Always list in `agents_required`
2. **Complete Context**: Pass all necessary context in description
3. **String Interpolation**: Use `${var}` for dynamic values
4. **Error Handling**: Check agent output status
5. **Validation**: Validate agent results before proceeding

### Anti-Patterns

**Missing Declaration**:
```
# Bad - undeclared dependency
generate_plan :: Req → Plan
generate_plan(req) =
  agent(project-planner, req.description)  # Not in agents_required!
```

**Incomplete Context**:
```
# Bad - missing critical context
agent(stage-executor, "Execute stage 1")  # Which stage? What criteria?

# Good - complete context
agent(stage-executor,
  "Execute Stage 1: ${stage.description}\n" +
  "Criteria: ${stage.acceptance_criteria}")
```

---

## Pattern 2: MCP Tool Usage

### Syntax

```
mcp::tool_name(params) :: → Data
```

### Semantics

```
mcp::tool_name(p) =
  direct_invocation(tool_name, p) ∧ handle_result → data
```

### Declaration

```
mcp_tools_required :: [ToolName]
mcp_tools_required = [tool1, tool2, ...]
```

### Usage Examples

**Query with Limit**:
```
mcp_tools_required = [mcp__meta-cc__query_tool_errors]

error_analysis :: Execution → ErrorReport
error_analysis(exec) =
  mcp::query_tool_errors(limit: 20) → recent_errors ∧
  categorize(recent_errors) →
  suggest_fixes(recent_errors) →
  report(errors, fixes)
```

**Multiple Tools**:
```
mcp_tools_required = [
  mcp__meta-cc__query_tool_errors,
  mcp__meta-cc__query_summaries,
  mcp__meta-cc__query_token_usage
]

comprehensive_analysis :: Session → Report
comprehensive_analysis(session) =
  errors = mcp::query_tool_errors(limit: 50) →
  summaries = mcp::query_summaries() →
  tokens = mcp::query_token_usage() →
  correlate(errors, summaries, tokens) →
  report(findings)
```

**Filtered Query**:
```
mcp_tools_required = [mcp__meta-cc__query_tools]

bash_errors :: → ErrorList
bash_errors() =
  mcp::query_tools(
    tool: "Bash",
    status: "error"
  ) → errors ∧
  filter_recent(errors) →
  errors
```

### Best Practices

1. **Use Full Names**: Complete tool name (e.g., `mcp__meta-cc__query_tool_errors`)
2. **Declare Dependencies**: List in `mcp_tools_required`
3. **Filter Early**: Use limit/filter parameters to reduce data
4. **Handle Empty**: Check for empty results
5. **Error Handling**: Handle tool errors gracefully

### Anti-Patterns

**Unbounded Query**:
```
# Bad - no limit, could return huge dataset
mcp::query_tools() → all_tools  # Potentially thousands of records

# Good - bounded query
mcp::query_tools(limit: 100) → recent_tools
```

**Missing Error Handling**:
```
# Bad - assumes tool succeeds
errors = mcp::query_tool_errors()

# Good - handle errors
result = mcp::query_tool_errors() →
if empty(result) then
  return no_errors_report
else
  analyze(result)
```

---

## Pattern 3: Skill Reference

### Syntax

```
skill(name) :: Context → Result
```

### Semantics

```
skill(name) =
  invoke_skill_tool(command: name) ∧ await_completion → result
```

### Declaration

```
skills_required :: [SkillName]
skills_required = [skill1, skill2, ...]
```

### Usage Examples

**Load Guidelines**:
```
skills_required = [testing-strategy]

generate_tests :: Code → Tests
generate_tests(code) =
  guidelines = skill(testing-strategy) →
  extract_patterns(guidelines) →
  apply_patterns(code, patterns) →
  tests
```

**Multiple Skills**:
```
skills_required = [testing-strategy, code-refactoring]

improve_code :: Code → ImprovedCode
improve_code(code) =
  test_guidelines = skill(testing-strategy) →
  refactor_guidelines = skill(code-refactoring) →
  analysis = analyze(code, test_guidelines, refactor_guidelines) →
  improvements = generate_improvements(analysis) →
  apply(improvements, code)
```

**Conditional Skill Usage**:
```
skills_required = [error-recovery, testing-strategy]

fix_failing_test :: Test → FixedTest
fix_failing_test(test) =
  if has_errors(test) then
    guidelines = skill(error-recovery) →
    apply_fixes(test, guidelines)
  else
    guidelines = skill(testing-strategy) →
    improve_test(test, guidelines)
```

### Best Practices

1. **Exact Names**: Use exact skill name (case-sensitive)
2. **Declare Dependencies**: List in `skills_required`
3. **Extract Patterns**: Don't use raw skill content, extract patterns
4. **Apply Systematically**: Follow skill guidelines methodically
5. **Validate Results**: Check improvements against skill criteria

### Anti-Patterns

**Undeclared Skill**:
```
# Bad - skill not declared
improve :: Code → Better
improve(code) =
  skill(testing-strategy) → guidelines  # Not in skills_required!
```

**Unused Declaration**:
```
# Bad - declared but never used
skills_required = [testing-strategy, code-refactoring]

improve :: Code → Better
improve(code) = analyze(code)  # Skills not used!
```

---

## Pattern 4: Resource Loading

### Syntax

```
read(path) :: Path → Content
```

### Semantics

```
read(p) =
  load_file(p) ∧ parse_content → content
```

### Usage Examples

**Load Previous State**:
```
load_context :: IterationNumber → Context
load_context(n) =
  prev = read(f"iteration_{n-1}.md") →
  parse(prev) →
  context
```

**Load Multiple Files**:
```
gather_requirements :: Spec → Requirements
gather_requirements(spec) =
  plan = read("docs/core/plan.md") →
  principles = read("docs/core/principles.md") →
  todo = read("TODO.md") →
  merge(plan, principles, todo) →
  requirements
```

**Conditional Loading**:
```
load_baseline :: Experiment → Maybe Baseline
load_baseline(exp) =
  if exists(f"${exp.dir}/baseline.md") then
    read(f"${exp.dir}/baseline.md") → Just(baseline)
  else
    Nothing
```

### Best Practices

1. **Check Existence**: Verify file exists before reading
2. **Relative Paths**: Use relative paths from workspace root
3. **Parse Results**: Don't use raw file content, parse it
4. **Error Handling**: Handle missing files gracefully

### Anti-Patterns

**Assume Existence**:
```
# Bad - assumes file exists
baseline = read("baseline.md")  # Crashes if missing

# Good - check first
baseline = if exists("baseline.md") then
  read("baseline.md")
else
  default_baseline
```

---

## Combined Patterns

### Orchestration with All Features

```
λ(task_spec) → comprehensive_result |
  all_features_used ∧ quality_maintained

agents_required = [planner, executor]
mcp_tools_required = [query_tool_errors]
skills_required = [testing-strategy]

orchestrate :: TaskSpec → Result
orchestrate(spec) =
  # Load context
  context = read("docs/core/plan.md") →

  # Get skill guidelines
  test_guidelines = skill(testing-strategy) →

  # Plan with agent
  plan = agent(planner,
    "Create plan for: ${spec.objective}\n" +
    "Context: ${context}\n" +
    "Guidelines: ${test_guidelines}"
  ) →

  # Execute stages
  results = [] →
  ∀stage ∈ plan.stages:
    result = agent(executor, stage.spec) →
    if error(result) then
      # Query recent errors via MCP
      errors = mcp::query_tool_errors(limit: 10) →
      suggest_fixes(errors, result) →
    results = results + [result] →

  # Final report
  report(plan, results)
```

### Analysis with MCP Tools

```
λ(query_spec) → analysis_report | data_comprehensive

mcp_tools_required = [
  query_tool_errors,
  query_token_usage,
  query_summaries
]
skills_required = [error-recovery]

analyze :: QuerySpec → Report
analyze(spec) =
  # Gather data via MCP
  errors = mcp::query_tool_errors(limit: 100) →
  tokens = mcp::query_token_usage() →
  summaries = mcp::query_summaries() →

  # Load skill for analysis patterns
  patterns = skill(error-recovery) →

  # Correlate and analyze
  correlations = correlate(errors, tokens, summaries) →
  insights = extract_insights(correlations, patterns) →
  recommendations = generate_recommendations(insights) →

  report(correlations, insights, recommendations)
```

---

## Dependency Management

### Declaration Order

```
# Standard order in prompt
agents_required :: [AgentType]
agents_required = [...]

mcp_tools_required :: [ToolName]
mcp_tools_required = [...]

skills_required :: [SkillName]
skills_required = [...]
```

### Dependency Graph

```
validate_dependencies :: Prompt → Bool
validate_dependencies(prompt) =
  ∀agent ∈ agents_used(prompt):
    agent ∈ prompt.agents_required ∧
  ∀tool ∈ tools_used(prompt):
    tool ∈ prompt.mcp_tools_required ∧
  ∀skill ∈ skills_used(prompt):
    skill ∈ prompt.skills_required
```

### Minimal Dependencies

**Good - Minimal**:
```
agents_required = [project-planner]  # Only what's used
```

**Bad - Excessive**:
```
agents_required = [
  project-planner,
  stage-executor,
  iteration-executor,
  ...15 more agents...
]  # Most never used!
```

---

## Integration Metrics

### Integration Score

**Formula**:
```
integration_score = features_used / applicable_features
```

Where applicable_features depends on agent type:
- Orchestration: 4 (agents, MCP, skills, resources)
- Analysis: 3 (MCP, skills, resources)
- Enhancement: 3 (skills, agents, resources)

### Target Scores

- **High integration**: ≥0.75 (3+ features)
- **Moderate**: ≥0.50 (2 features)
- **Low**: ≥0.25 (1 feature)

### Example Calculation

**phase-planner-executor**:
```
agents_used = 2 (project-planner, stage-executor)
mcp_tools_used = 2 (query_tool_errors, query_summaries)
skills_used = 0
resources_used = 0

features_used = 2 (agents, MCP)
applicable_features = 4 (orchestration agent)

integration_score = 2/4 = 0.50  # Moderate
```

**Improved version would add**:
```
skills_required = [testing-strategy]
# Use in plan generation

integration_score = 3/4 = 0.75  # High ✅
```

---

## References

- **Validated Example**: .claude/agents/phase-planner-executor.md
- **Pattern Catalog**: reference/patterns.md
- **Symbolic Language**: reference/symbolic-language.md
- **Claude Code Docs**:
  - [Subagents](https://docs.claude.com/en/docs/claude-code/subagents)
  - [Skills](https://docs.claude.com/en/docs/claude-code/skills)
  - [MCP Integration](https://docs.claude.com/en/docs/claude-code/mcp)
