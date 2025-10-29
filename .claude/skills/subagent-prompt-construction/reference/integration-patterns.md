# Claude Code Integration Patterns

Formal patterns for integrating Claude Code features (agents, MCP tools, skills) in subagent prompts.

---

## 1. Subagent Composition

**Pattern**:
```
agent(type, description) :: Context → Output
```

**Semantics**:
```
agent(type, desc) =
  invoke_task_tool(subagent_type: type, prompt: desc) ∧
  await_completion ∧
  return output
```

**Usage in prompt**:
```
agent(project-planner,
  "Create detailed TDD implementation plan for: ${objectives}\n" +
  "Scope: ${scope}\n" +
  "Constraints: ${constraints}"
) → plan
```

**Actual invocation** (Claude Code):
```python
Task(subagent_type="project-planner", description=f"Create detailed TDD...")
```

**Declaration**:
```
agents_required :: [AgentType]
agents_required = [project-planner, stage-executor, ...]
```

**Best practices**:
- Declare all agents in dependencies section
- Pass context explicitly via description string
- Use meaningful variable names for outputs (→ plan, → result)
- Handle agent failures with conditional logic

**Example**:
```
generate_plan :: Requirements → Plan
generate_plan(req) =
  agent(project-planner, "Create plan for: ${req.objectives}") → plan ∧
  validate_plan(plan) ∧
  return plan
```

---

## 2. MCP Tool Integration

**Pattern**:
```
mcp::tool_name(params) :: → Data
```

**Semantics**:
```
mcp::tool_name(p) =
  direct_invocation(mcp__namespace__tool_name, p) ∧
  handle_result ∧
  return data
```

**Usage in prompt**:
```
mcp::query_tool_errors(limit: 20) → recent_errors
mcp::query_summaries() → summaries
mcp::query_user_messages(pattern: ".*bug.*") → bug_reports
```

**Actual invocation** (Claude Code):
```python
mcp__meta_cc__query_tool_errors(limit=20)
mcp__meta_cc__query_summaries()
mcp__meta_cc__query_user_messages(pattern=".*bug.*")
```

**Declaration**:
```
mcp_tools_required :: [ToolName]
mcp_tools_required = [
  mcp__meta-cc__query_tool_errors,
  mcp__meta-cc__query_summaries,
  mcp__meta-cc__query_user_messages
]
```

**Best practices**:
- Use mcp:: prefix for clarity
- Declare all MCP tools in dependencies section
- Specify full tool name in declaration (mcp__namespace__tool)
- Handle empty results gracefully
- Limit result sizes with parameters

**Example**:
```
error_analysis :: Execution → ErrorReport
error_analysis(exec) =
  mcp::query_tool_errors(limit: 20) → recent_errors ∧
  if |recent_errors| > 0 then
    categorize(recent_errors) ∧
    suggest_fixes(recent_errors)
  else
    report("No errors found")
```

---

## 3. Skill Reference

**Pattern**:
```
skill(name) :: Context → Result
```

**Semantics**:
```
skill(name) =
  invoke_skill_tool(command: name) ∧
  await_completion ∧
  return guidelines
```

**Usage in prompt**:
```
skill(testing-strategy) → test_guidelines
skill(code-refactoring) → refactor_patterns
skill(methodology-bootstrapping) → baime_framework
```

**Actual invocation** (Claude Code):
```python
Skill(command="testing-strategy")
Skill(command="code-refactoring")
Skill(command="methodology-bootstrapping")
```

**Declaration**:
```
skills_required :: [SkillName]
skills_required = [testing-strategy, code-refactoring, ...]
```

**Best practices**:
- Reference skill by name (kebab-case)
- Declare all skills in dependencies section
- Use skill guidelines to inform agent decisions
- Skills provide context, not direct execution
- Apply skill patterns via agent logic

**Example**:
```
enhance_tests :: CodeArtifact → ImprovedTests
enhance_tests(code) =
  skill(testing-strategy) → guidelines ∧
  current_coverage = analyze_coverage(code) ∧
  gaps = identify_gaps(code, guidelines) ∧
  generate_tests(gaps, guidelines)
```

---

## 4. Resource Loading

**Pattern**:
```
read(path) :: Path → Content
```

**Semantics**:
```
read(p) =
  load_file(p) ∧
  parse_content ∧
  return content
```

**Usage in prompt**:
```
read("docs/plan.md") → plan_doc
read("iteration_{n-1}.md") → previous_iteration
read("TODO.md") → tasks
```

**Actual invocation** (Claude Code):
```python
Read(file_path="docs/plan.md")
Read(file_path=f"iteration_{n-1}.md")
Read(file_path="TODO.md")
```

**Best practices**:
- Use relative paths when possible
- Handle file not found errors
- Parse structured content (markdown, JSON)
- Extract relevant sections only
- Cache frequently accessed files

**Example**:
```
load_context :: IterationNumber → Context
load_context(n) =
  if n > 0 then
    read(f"iteration_{n-1}.md") → prev ∧
    extract_state(prev)
  else
    initial_state()
```

---

## 5. Combined Integration

**Pattern**: Multiple feature types in single prompt

**Example** (phase-planner-executor):
```
execute_phase :: FeatureSpec → PhaseReport
execute_phase(spec) =
  # Agent composition
  plan = agent(project-planner, spec.objectives) →

  # Sequential agent execution
  ∀stage_num ∈ [1..|plan.stages|]:
    result = agent(stage-executor, plan.stages[stage_num]) →

    # MCP tool integration for error analysis
    if result.status == "error" then
      errors = mcp::query_tool_errors(limit: 20) →
      analysis = analyze(errors) →
      return (plan, results, analysis)

  # Final reporting
  report(plan, results, quality_check, progress_tracking)
```

**Integration score**: 4 features (2 agents + 2 MCP tools) → 0.75

**Best practices**:
- Use ≥3 features for high integration score (≥0.75)
- Combine patterns appropriately (orchestration + analysis)
- Declare all dependencies upfront
- Handle failures at integration boundaries
- Maintain compactness despite multiple integrations

---

## 6. Conditional Integration

**Pattern**: Feature usage based on runtime conditions

**Example**:
```
execute_with_monitoring :: Task → Result
execute_with_monitoring(task) =
  result = agent(executor, task) →

  if result.status == "error" then
    # Conditional MCP integration
    errors = mcp::query_tool_errors(limit: 10) →
    recent_patterns = mcp::query_summaries() →
    enhanced_diagnosis = combine(errors, recent_patterns)
  else if result.needs_improvement then
    # Conditional skill reference
    guidelines = skill(code-refactoring) →
    suggestions = apply_guidelines(result, guidelines)
  else
    result
```

**Benefits**:
- Resource-efficient (only invoke when needed)
- Clearer error handling
- Adaptive behavior

---

## Integration Complexity Matrix

| Features | Integration Score | Complexity | Example |
|----------|------------------|------------|---------|
| 1 agent | 0.25 | Low | Simple executor |
| 2 agents | 0.50 | Medium | Planner + executor |
| 2 agents + 1 MCP | 0.60 | Medium-High | Executor + error query |
| 2 agents + 2 MCP | 0.75 | High | phase-planner-executor |
| 3 agents + 2 MCP + 1 skill | 0.90 | Very High | Complex orchestration |

**Recommendation**: Target 0.50-0.75 for maintainability

---

## Dependencies Section Template

```markdown
## Dependencies

agents_required :: [AgentType]
agents_required = [
  agent-type-1,
  agent-type-2,
  ...
]

mcp_tools_required :: [ToolName]
mcp_tools_required = [
  mcp__namespace__tool_1,
  mcp__namespace__tool_2,
  ...
]

skills_required :: [SkillName]
skills_required = [
  skill-name-1,
  skill-name-2,
  ...
]
```

**Rules**:
- All sections optional (omit if not used)
- List all features used in prompt
- Use correct naming conventions (kebab-case for skills/agents, mcp__namespace__tool for MCP)
- Order: agents, MCP tools, skills

---

## Error Handling Patterns

### Agent Failure
```
result = agent(executor, task) →
if result.status == "error" then
  error_analysis(result) →
  return (partial_result, error_report)
```

### MCP Tool Empty Results
```
data = mcp::query_tool(params) →
if |data| == 0 then
  return "No data available for analysis"
else
  process(data)
```

### Skill Not Available
```
guidelines = skill(optional-skill) →
if guidelines.available then
  apply(guidelines)
else
  use_default_approach()
```

---

## Validation Criteria

Integration quality checklist:
- [ ] All features declared in dependencies section
- [ ] Feature invocations use correct syntax
- [ ] Error handling at integration boundaries
- [ ] Integration score ≥0.50
- [ ] Compactness maintained despite integrations
- [ ] Clear separation between feature types
- [ ] Meaningful variable names for outputs

---

## Related Resources

- **Patterns**: `patterns.md` (orchestration, analysis, enhancement patterns)
- **Symbolic Language**: `symbolic-language.md` (formal syntax)
- **Template**: `../templates/subagent-template.md` (includes dependencies section)
- **Example**: `../examples/phase-planner-executor.md` (2 agents + 2 MCP tools)
