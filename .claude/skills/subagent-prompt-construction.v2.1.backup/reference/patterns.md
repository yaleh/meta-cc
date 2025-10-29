# Subagent Construction Patterns

## Overview

Three validated patterns for subagent construction in Claude Code, extracted from BAIME experiment.

**Status**: Orchestration pattern validated (V_instance = 0.895), Analysis and Enhancement patterns designed.

---

## Pattern 1: Orchestration Agent

**Use Case**: Coordinate multiple subagents to accomplish complex multi-stage tasks.

**When to Use**:
- Need to compose existing agents
- Sequential stage execution required
- Progress tracking needed
- Error handling between stages

**Structure**:
```
λ(task_spec) → aggregated_result | all_stages_complete

agents_required = [planner, executor, ...]
mcp_tools_required = [error_query, ...]

parse :: Spec → Requirements
plan :: Requirements → Plan  (via agent)
execute_stage :: (Plan, N) → StageResult  (via agent)
validate :: StageResult → Quality
aggregate :: [StageResult] → FinalReport

orchestrate :: Spec → Report
orchestrate(spec) =
  req = parse(spec) →
  plan = agent(planner, req) →
  ∀stage ∈ plan.stages:
    result = agent(executor, stage) →
    if error(result) then handle_error →
    validate(result) →
  aggregate(results)
```

**Key Characteristics**:
- 5-8 functions
- Agent composition via `agent(type, desc)`
- Error handling in loop
- Progress tracking
- Aggregation function

**Validated Example**: phase-planner-executor
- 92 lines, 7 functions
- 2 agents (project-planner, stage-executor)
- 2 MCP tools (query_tool_errors, query_summaries)
- V_instance = 0.895

**Anti-Patterns**:
- Inline all logic in main function (no decomposition)
- Missing error handling
- No progress tracking
- Unclear stage termination

---

## Pattern 2: Analysis Agent

**Use Case**: Query MCP tools to extract patterns and generate insights from session data.

**When to Use**:
- Need to analyze session history
- Extract patterns from tool calls
- Generate recommendations
- Identify trends or anomalies

**Structure**:
```
λ(query_spec) → analysis_report | data_sufficient

mcp_tools_required = [query_tool, ...]

parse_query :: QuerySpec → Params
fetch_data :: Params → RawData  (via MCP)
extract_patterns :: RawData → Patterns
generate_insights :: Patterns → Insights
recommend :: Insights → Actions

analyze :: QuerySpec → Report
analyze(spec) =
  params = parse_query(spec) →
  data = mcp::query_tool(params) →
  patterns = extract_patterns(data) →
  insights = generate_insights(patterns) →
  actions = recommend(insights) →
  report(patterns, insights, actions)
```

**Key Characteristics**:
- 4-6 functions
- Heavy MCP tool usage
- Pattern extraction
- Insight generation
- Actionable recommendations

**Example Use Cases**:
- Error pattern analyzer
- Performance bottleneck detector
- Quality metrics analyzer
- Usage pattern reporter

**Anti-Patterns**:
- Query without filtering (data overload)
- Missing pattern categorization
- Insights without context
- No actionable recommendations

---

## Pattern 3: Enhancement Agent

**Use Case**: Apply skill guidelines to improve or refactor artifacts.

**When to Use**:
- Code refactoring needed
- Documentation improvement
- Test coverage enhancement
- Style/quality improvements

**Structure**:
```
λ(artifact_spec) → improved_artifact | quality_improved

skills_required = [domain_skill, ...]

load_artifact :: Spec → Artifact
load_guidelines :: Skill → Guidelines  (via skill)
analyze :: (Artifact, Guidelines) → Issues
generate_improvements :: Issues → Changes
apply :: (Changes, Artifact) → ImprovedArtifact
validate :: ImprovedArtifact → Quality

enhance :: Spec → Enhanced
enhance(spec) =
  artifact = load_artifact(spec) →
  guidelines = skill(domain_skill) →
  issues = analyze(artifact, guidelines) →
  changes = generate_improvements(issues) →
  improved = apply(changes, artifact) →
  validate(improved) →
  improved
```

**Key Characteristics**:
- 5-7 functions
- Skill reference for guidelines
- Analysis phase
- Generation phase
- Validation phase

**Example Use Cases**:
- Code refactorer (using code-refactoring skill)
- Test enhancer (using testing-strategy skill)
- Documentation improver (using documentation skill)
- API designer (using api-design skill)

**Anti-Patterns**:
- Improvements without guidelines
- No before/after validation
- Missing quality metrics
- Unclear improvement criteria

---

## Common Function Decomposition

### Optimal Structure (5-10 functions)

**Parse/Extract (1-2 functions)**:
- Input validation
- Requirement extraction
- Context parsing

**Core Logic (3-5 functions)**:
- Main business logic
- Transformation functions
- Validation functions

**Integration (1-2 functions)**:
- Agent composition
- MCP tool calls
- Skill references

**Output/Reporting (1-2 functions)**:
- Result aggregation
- Report generation
- Artifact creation

### Example Decomposition

```
# 7-function orchestration agent
parse_feature :: FeatureSpec → Requirements
generate_plan :: Requirements → Plan  (agent)
execute_stage :: (Plan, StageNumber) → StageResult  (agent)
quality_check :: StageResult → QualityReport
error_analysis :: Execution → ErrorReport  (MCP)
progress_tracking :: [StageResult] → ProgressReport
execute_phase :: FeatureSpec → PhaseReport  (main)
```

---

## Integration Best Practices

### Agent Composition

**Pattern**:
```
agent(type, description) :: Context → Output
```

**Best Practices**:
- Declare in `agents_required`
- Pass complete context in description
- Use string interpolation for dynamic context
- Handle agent errors explicitly

**Example**:
```
agents_required = [project-planner, stage-executor]

generate_plan :: Requirements → Plan
generate_plan(req) =
  agent(project-planner,
    "Create detailed TDD implementation plan for: ${req.objectives}\n" +
    "Scope: ${req.scope}\n" +
    "Constraints: ${req.constraints}\n" +
    "Code limit: ≤500 lines per phase, ≤200 lines per stage"
  ) → plan ∧
  validate_plan(plan) →
  plan
```

### MCP Tool Usage

**Pattern**:
```
mcp::tool_name(params) :: → Data
```

**Best Practices**:
- Declare in `mcp_tools_required`
- Use full tool name (e.g., `mcp__meta-cc__query_tool_errors`)
- Handle empty results
- Filter data early

**Example**:
```
mcp_tools_required = [
  mcp__meta-cc__query_tool_errors,
  mcp__meta-cc__query_summaries
]

error_analysis :: Execution → ErrorReport
error_analysis(exec) =
  mcp::query_tool_errors(limit: 20) → recent_errors ∧
  categorize(recent_errors) →
  suggest_fixes(recent_errors) →
  report(errors, fixes)
```

### Skill Reference

**Pattern**:
```
skill(name) :: Context → Result
```

**Best Practices**:
- Declare in `skills_required`
- Use skill name exactly
- Extract relevant patterns
- Apply guidelines systematically

**Example**:
```
skills_required = [testing-strategy]

generate_tests :: Code → Tests
generate_tests(code) =
  guidelines = skill(testing-strategy) →
  extract_patterns(guidelines) →
  apply_to_code(code, patterns) →
  tests
```

---

## Compactness Techniques

### 1. Symbolic Logic

**Verbose**:
```
if x is valid and y is complete and z is not empty then...
```

**Compact**:
```
valid(x) ∧ complete(y) ∧ ¬empty(z) → ...
```

### 2. Function Composition

**Verbose**:
```
temp1 = step1(input)
temp2 = step2(temp1)
result = step3(temp2)
```

**Compact**:
```
step1(input) → step2 → step3 → result
```

### 3. Predicate Constraints

**Verbose**:
```
Check that all stages have code ≤200 lines
Check that all stages have tests ≤200 lines
Check that coverage ≥80%
```

**Compact**:
```
∀stage ∈ stages:
  |code(stage)| ≤ 200 ∧
  |test(stage)| ≤ 200 ∧
  coverage(stage) ≥ 0.80
```

### 4. Type Signatures

**Clear and Compact**:
```
parse_feature :: FeatureSpec → Requirements
generate_plan :: Requirements → Plan
execute_stage :: (Plan, StageNumber) → StageResult
```

---

## Quality Metrics

### Compactness

**Formula**: `1 - (lines / 150)`

**Targets**:
- Simple: ≥0.70 (≤45 lines)
- Moderate: ≥0.50 (≤75 lines)
- Complex: ≥0.30 (≤105 lines)

### Integration

**Formula**: `features_used / applicable_features`

**Targets**:
- High integration: ≥0.75 (3+ features)
- Moderate: ≥0.50 (2 features)
- Low: ≥0.25 (1 feature)

### Maintainability

**Components**:
- Clear structure (0-1)
- Easy to modify (0-1)
- Well-documented (0-1)

**Target**: ≥0.85

---

## Pattern Selection Guide

### Decision Tree

```
Is task orchestration of multiple agents?
├─ Yes → Use Orchestration Pattern
└─ No → Is task data analysis?
    ├─ Yes → Use Analysis Pattern
    └─ No → Is task artifact improvement?
        ├─ Yes → Use Enhancement Pattern
        └─ No → Create custom pattern (use base template)
```

### Pattern Comparison

| Aspect | Orchestration | Analysis | Enhancement |
|--------|---------------|----------|-------------|
| **Primary Integration** | Agents | MCP Tools | Skills |
| **Function Count** | 5-8 | 4-6 | 5-7 |
| **Typical Lines** | 80-120 | 60-90 | 70-100 |
| **Complexity** | Moderate-Complex | Simple-Moderate | Moderate |
| **Error Handling** | Per-stage | Data validation | Quality checks |
| **Validation Status** | ✅ Validated | 🎯 Designed | 🎯 Designed |

---

## References

- **Validated Example**: .claude/agents/phase-planner-executor.md
- **Template**: templates/subagent-template.md
- **Symbolic Language**: reference/symbolic-language.md
- **Integration Patterns**: reference/integration-patterns.md
