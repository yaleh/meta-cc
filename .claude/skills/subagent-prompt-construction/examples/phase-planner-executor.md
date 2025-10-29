# Example: phase-planner-executor (Orchestration Pattern)

**Metrics**: 92 lines | 2 agents + 2 MCP tools | Integration: 0.75 | V_instance: 0.895 ✅

**Demonstrates**: Agent composition, MCP integration, error handling, progress tracking, TDD compliance

## Prompt Structure

```markdown
---
name: phase-planner-executor
description: Plans and executes new development phases end-to-end
---

λ(feature_spec, todo_ref?) → (plan, execution_report, status) | TDD ∧ code_limits

agents_required = [project-planner, stage-executor]
mcp_tools_required = [mcp__meta-cc__query_tool_errors, mcp__meta-cc__query_summaries]
```

## Function Decomposition (7 functions)

```
parse_feature :: FeatureSpec → Requirements
parse_feature(spec) = extract(objectives, scope, constraints) ∧ identify(deliverables)

generate_plan :: Requirements → Plan
generate_plan(req) = agent(project-planner, "${req.objectives}...") → plan

execute_stage :: (Plan, StageNumber) → StageResult
execute_stage(plan, n) = agent(stage-executor, plan.stages[n].description) → result

quality_check :: StageResult → QualityReport
quality_check(result) = test_coverage(result) ≥ 0.80 ∧ all_tests_pass(result)

error_analysis :: Execution → ErrorReport
error_analysis(exec) = mcp::query_tool_errors(limit: 20) → recent_errors ∧ categorize

progress_tracking :: [StageResult] → ProgressReport
progress_tracking(results) = completed / |results| → percentage

execute_phase :: FeatureSpec → PhaseReport (main)
execute_phase(spec) =
  req = parse_feature(spec) →
  plan = generate_plan(req) →
  ∀stage_num ∈ [1..|plan.stages|]:
    result = execute_stage(plan, stage_num) →
    if result.status == "error" then error_analysis(result) → return
  report(plan, results, quality_check, progress_tracking)
```

## Constraints

```
constraints :: PhaseExecution → Bool
constraints(exec) =
  ∀stage ∈ exec.plan.stages:
    |code(stage)| ≤ 200 ∧ |test(stage)| ≤ 200 ∧ coverage(stage) ≥ 0.80 ∧
  |code(exec.phase)| ≤ 500 ∧ tdd_compliance(exec)
```

## Integration Patterns

**Agent Composition**:
```
agent(project-planner, "Create plan for: ${objectives}") → plan
agent(stage-executor, "Execute: ${stage.description}") → result
```

**MCP Integration**:
```
mcp::query_tool_errors(limit: 20) → recent_errors
mcp::query_summaries() → summaries
```

## Validation Results

| Metric | Target | Actual | Status |
|--------|--------|--------|--------|
| Lines | ≤150 | 92 | ✅ |
| Functions | 5-8 | 7 | ✅ |
| Integration Score | ≥0.50 | 0.75 | ✅ |
| Compactness | ≥0.30 | 0.387 | ✅ |

**Source**: `/home/yale/work/meta-cc/.claude/agents/phase-planner-executor.md`
**Analysis**: `reference/case-studies/phase-planner-executor-analysis.md`
