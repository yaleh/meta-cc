# Subagent Prompt Patterns

Core patterns for constructing Claude Code subagent prompts.

---

## Pattern 1: Orchestration Agent

**Use case**: Coordinate multiple subagents for complex workflows

**Structure**:
```
orchestrate :: Task → Result
orchestrate(task) =
  plan = agent(planner, task.spec) →
  ∀stage ∈ plan.stages:
    result = agent(executor, stage) →
    validate(result) →
  aggregate(results)
```

**Example**: phase-planner-executor
- Coordinates project-planner and stage-executor
- Sequential stage execution with validation
- Error detection and recovery
- Progress tracking

**When to use**:
- Multi-step workflows requiring planning
- Need to coordinate 2+ specialized agents
- Sequential stages with dependencies
- Error handling between stages critical

**Complexity**: Moderate to Complex (60-150 lines)

---

## Pattern 2: Analysis Agent

**Use case**: Analyze data via MCP tools and generate insights

**Structure**:
```
analyze :: Query → Report
analyze(query) =
  data = mcp::query_tool(query.params) →
  patterns = extract_patterns(data) →
  insights = generate_insights(patterns) →
  report(patterns, insights)
```

**Example**: error-analyzer (hypothetical)
- Query tool errors via MCP
- Categorize error patterns
- Suggest fixes
- Generate analysis report

**When to use**:
- Need to query session data
- Pattern extraction from data
- Insight generation from analysis
- Reporting on historical data

**Complexity**: Simple to Moderate (30-90 lines)

---

## Pattern 3: Enhancement Agent

**Use case**: Apply skill guidelines to improve artifacts

**Structure**:
```
enhance :: Artifact → ImprovedArtifact
enhance(artifact) =
  guidelines = skill(domain-skill) →
  analysis = analyze(artifact, guidelines) →
  improvements = generate(analysis) →
  apply(improvements, artifact)
```

**Example**: code-refactorer (hypothetical)
- Load refactoring skill guidelines
- Analyze code against guidelines
- Generate improvement suggestions
- Apply or report improvements

**When to use**:
- Systematic artifact improvement
- Apply established skill patterns
- Need consistent quality standards
- Repeatable enhancement process

**Complexity**: Moderate (60-120 lines)

---

## Pattern 4: Validation Agent

**Use case**: Validate artifacts against criteria

**Structure**:
```
validate :: Artifact → ValidationReport
validate(artifact) =
  criteria = load_criteria() →
  results = check_all(artifact, criteria) →
  report(passes, failures, warnings)
```

**Example**: quality-checker (hypothetical)
- Load quality criteria
- Check code standards, tests, coverage
- Generate pass/fail report
- Provide remediation suggestions

**When to use**:
- Pre-commit checks
- Quality gates
- Compliance validation
- Systematic artifact verification

**Complexity**: Simple to Moderate (30-90 lines)

---

## Pattern Selection Guide

| Need | Pattern | Complexity | Integration |
|------|---------|------------|-------------|
| Coordinate agents | Orchestration | High | 2+ agents |
| Query & analyze data | Analysis | Medium | MCP tools |
| Improve artifacts | Enhancement | Medium | Skills |
| Check compliance | Validation | Low | Skills optional |
| Multi-step workflow | Orchestration | High | Agents + MCP |
| One-step analysis | Analysis | Low | MCP only |

---

## Common Anti-Patterns

### ❌ Flat Structure (No Decomposition)
```
# Bad - 100 lines of inline logic
λ(input) → output | step1 ∧ step2 ∧ ... ∧ step50
```

**Fix**: Decompose into 5-10 functions

### ❌ Verbose Natural Language
```
# Bad
"First, we need to validate the input. Then, we should..."
```

**Fix**: Use symbolic logic and function composition

### ❌ Missing Dependencies
```
# Bad - calls agents without declaring
agent(mystery-agent, ...) → result
```

**Fix**: Explicit dependencies section

### ❌ Unclear Constraints
```
# Bad - vague requirements
"Make sure code quality is good"
```

**Fix**: Explicit predicates (coverage ≥ 0.80)

---

## Pattern Composition

Patterns can be composed for complex workflows:

**Orchestration + Analysis**:
```
orchestrate(task) =
  plan = agent(planner, task) →
  ∀stage ∈ plan.stages:
    result = agent(executor, stage) →
    if result.status == "error" then
      analysis = analyze_errors(result) → # Analysis pattern
      return (partial_results, analysis)
  aggregate(results)
```

**Enhancement + Validation**:
```
enhance(artifact) =
  improved = apply_skill(artifact) →
  validation = validate(improved) → # Validation pattern
  if validation.passes then improved else retry(validation.issues)
```

---

## Quality Metrics

### Compactness
- Simple: ≤60 lines (score ≥0.60)
- Moderate: ≤90 lines (score ≥0.40)
- Complex: ≤150 lines (score ≥0.00)

**Formula**: `1 - (lines / 150)`

### Integration
- High: 3+ features (score ≥0.75)
- Moderate: 2 features (score ≥0.50)
- Low: 1 feature (score ≥0.25)

**Formula**: `features_used / applicable_features`

### Clarity
- Clear structure (0-1 subjective)
- Obvious flow (0-1 subjective)
- Self-documenting (0-1 subjective)

**Target**: All ≥0.80

---

## Validation Checklist

Before using a pattern:
- [ ] Pattern matches use case
- [ ] Complexity appropriate (simple/moderate/complex)
- [ ] Dependencies identified
- [ ] Function count: 3-12
- [ ] Line count: ≤150
- [ ] Integration score: ≥0.50
- [ ] Constraints explicit
- [ ] Example reviewed

---

## Related Resources

- **Integration Patterns**: `integration-patterns.md` (agent/MCP/skill syntax)
- **Symbolic Language**: `symbolic-language.md` (operators, quantifiers)
- **Template**: `../templates/subagent-template.md` (reusable structure)
- **Examples**: `../examples/phase-planner-executor.md` (orchestration)
- **Case Studies**: `case-studies/phase-planner-executor-analysis.md` (detailed)
