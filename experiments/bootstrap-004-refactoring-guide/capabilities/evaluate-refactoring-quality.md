# Capability: Evaluate Refactoring Quality

## Purpose
Calculate dual-layer value functions (V_instance and V_meta) using rigorous rubrics and concrete evidence.

## When to Use
- End of each iteration
- When assessing convergence
- When comparing progress across iterations

## Inputs
- Current iteration metrics (from collect-refactoring-data)
- Baseline metrics
- Refactoring log
- Session analysis data
- Knowledge artifacts

## Outputs
- V_instance calculation with component breakdown
- V_meta calculation with rubric assessments
- Evidence documentation for all scores
- Gap enumeration

## V_instance Calculation Protocol

### Component 1: V_code_quality (0.3 weight)

**Rubric Application**:
- 1.0: ≥30% complexity reduction, zero duplication, zero static warnings
- 0.8: 20-29% complexity reduction, <5% duplication, <3 warnings
- 0.6: 10-19% complexity reduction, <10% duplication, <10 warnings
- 0.4: 5-9% complexity reduction, <20% duplication, <20 warnings
- 0.2: 1-4% complexity reduction, some progress
- 0.0: No measurable improvement

**Evidence Required**:
- Complexity metrics: baseline vs current
- Duplication metrics: blocks removed
- Static analysis: warnings fixed
- Specific file/function examples

**Calculation**:
```
complexity_score = apply_rubric(complexity_reduction_%)
duplication_score = apply_rubric(duplication_reduction_%)
static_score = apply_rubric(warnings_reduction_%)
V_code_quality = (complexity_score + duplication_score + static_score) / 3
```

### Component 2: V_maintainability (0.3 weight)

**Rubric Application**:
- 1.0: ≥85% coverage, perfect cohesion, complete documentation
- 0.8: 75-84% coverage, good cohesion, >90% documentation
- 0.6: 65-74% coverage, acceptable cohesion, >75% documentation
- 0.4: 55-64% coverage, some cohesion issues, >50% documentation
- 0.2: 45-54% coverage, poor cohesion, <50% documentation
- 0.0: <45% coverage or significant maintainability issues

**Evidence Required**:
- Coverage report with percentages
- Module structure assessment
- Documentation coverage count
- Cohesion examples

**Calculation**:
```
coverage_score = current_coverage / 85% (capped at 1.0)
cohesion_score = subjective_assessment_using_rubric
documentation_score = documented_public_funcs / total_public_funcs
V_maintainability = (coverage_score + cohesion_score + documentation_score) / 3
```

### Component 3: V_safety (0.2 weight)

**Rubric Application**:
- 1.0: 100% tests pass, all steps verified, perfect git discipline
- 0.8: 100% tests pass, >95% steps verified, good git discipline
- 0.6: 100% tests pass, >90% steps verified, acceptable git discipline
- 0.4: 98-99% tests pass, some verification gaps
- 0.2: 95-97% tests pass, significant verification gaps
- 0.0: <95% tests pass or broken tests

**Evidence Required**:
- Test execution logs
- Git commit history
- Verification checklist
- Rollback incidents

**Calculation**:
```
test_pass_rate = passing_tests / total_tests
verification_rate = verified_steps / total_steps
git_discipline = subjective_assessment_using_rubric
V_safety = (test_pass_rate + verification_rate + git_discipline) / 3
```

### Component 4: V_effort (0.2 weight)

**Rubric Application**:
- 1.0: ≥10x speedup, >90% automation, minimal rework
- 0.8: 7-9x speedup, 75-90% automation, <10% rework
- 0.6: 5-6x speedup, 60-75% automation, <20% rework
- 0.4: 3-4x speedup, 40-60% automation, <30% rework
- 0.2: 2x speedup, <40% automation, significant rework
- 0.0: No speedup or slower than ad-hoc

**Evidence Required**:
- Time tracking data
- Automation tool usage
- Rework/rollback count
- Efficiency metrics

**Calculation**:
```
efficiency_ratio = baseline_time_estimate / actual_time
automation_rate = automated_checks / total_checks
rework_rate = 1 - (rework_steps / total_steps)
V_effort = apply_rubric(efficiency_ratio, automation_rate, rework_rate)
```

### Final V_instance

```
V_instance = 0.3 × V_code_quality + 0.3 × V_maintainability + 0.2 × V_safety + 0.2 × V_effort
```

## V_meta Calculation Protocol

**CRITICAL**: Evaluate V_meta INDEPENDENTLY of V_instance. Meta layer assesses methodology quality, not task success.

### Component 1: V_completeness (0.4 weight)

**Phase Assessment** (each phase: 0.25 weight):

**Detection Phase**:
- Exceptional (1.0): Comprehensive taxonomy (8+ categories), automated tools, prioritization, validated
- Strong (0.75): Good taxonomy (5-7 categories), semi-automated, basic prioritization
- Acceptable (0.5): Basic taxonomy (3-4 categories), manual, ad-hoc prioritization
- Weak (0.25): Minimal taxonomy (1-2 categories), inconsistent
- Missing (0.0): No systematic detection

**Planning Phase**:
- Exceptional (1.0): Comprehensive patterns (10+ types), safety protocols, sequencing, rollback, validated
- Strong (0.75): Good patterns (6-9 types), safety guidelines, basic sequencing
- Acceptable (0.5): Basic patterns (3-5 types), some safety
- Weak (0.25): Minimal patterns (1-2 types), limited safety
- Missing (0.0): No systematic planning

**Execution Phase**:
- Exceptional (1.0): Detailed recipes, TDD integration, continuous verification, git discipline, automation
- Strong (0.75): Good guidance, test requirements, verification steps
- Acceptable (0.5): Basic steps, some test coverage
- Weak (0.25): Minimal guidance, inconsistent testing
- Missing (0.0): No systematic execution

**Verification Phase**:
- Exceptional (1.0): Multi-layer validation, automated regression, quality gates, rollback triggers
- Strong (0.75): Good validation, some automation, clear criteria
- Acceptable (0.5): Basic validation, manual checks, informal criteria
- Weak (0.25): Minimal validation, inconsistent
- Missing (0.0): No systematic verification

**Evidence Required**:
- Artifacts in knowledge/ directory
- Capability coverage
- Pattern catalog
- Automation tools

**Calculation**:
```
V_completeness = (detection_score + planning_score + execution_score + verification_score) / 4
```

### Component 2: V_effectiveness (0.3 weight)

**Quality Improvement** (0.33):
- Exceptional (1.0): Consistent gains (≥3 examples), quantified, before/after metrics
- Strong (0.75): Good gains (2 examples), measurable
- Acceptable (0.5): Some gains (1 example), qualitative
- Weak (0.25): Unclear impact
- Missing (0.0): No demonstrated improvement

**Safety Record** (0.33):
- Exceptional (1.0): Zero breaking changes, 100% test pass, clean rollback, documented
- Strong (0.75): Minimal issues (<5%), strong discipline
- Acceptable (0.5): Some issues (5-10%), acceptable coverage
- Weak (0.25): Frequent issues (>10%), poor discipline
- Missing (0.0): No safety tracking or major breakages

**Efficiency Gains** (0.33):
- Exceptional (1.0): ≥10x speedup demonstrated, high automation, minimal rework
- Strong (0.75): 5-9x speedup, good automation
- Acceptable (0.5): 3-4x speedup, some automation
- Weak (0.25): <3x speedup, minimal automation
- Missing (0.0): No efficiency improvement

**Evidence Required**:
- Quantitative metrics
- Before/after examples
- Time tracking
- Safety incident log

**Calculation**:
```
V_effectiveness = (quality_improvement_score + safety_record_score + efficiency_gains_score) / 3
```

### Component 3: V_reusability (0.3 weight)

**Language Independence** (0.33):
- Exceptional (1.0): Applies to 5+ languages, language-agnostic documented
- Strong (0.75): Applies to 3-4 languages, mostly transferable
- Acceptable (0.5): Applies to 2 languages, some adaptation
- Weak (0.25): Mostly language-specific, limited transfer
- Missing (0.0): Completely language-specific

**Codebase Generality** (0.33):
- Exceptional (1.0): Diverse codebases (CLI, library, web, embedded), codebase-agnostic
- Strong (0.75): 2-3 codebase types, good generality
- Acceptable (0.5): 1-2 codebase types, some adaptation
- Weak (0.25): Mostly specific to one type
- Missing (0.0): Completely codebase-specific

**Abstraction Quality** (0.33):
- Exceptional (1.0): Universal principles extracted, minimal context-specific, clear adaptation guides
- Strong (0.75): Good principles, some context elements, basic guidance
- Acceptable (0.5): Mixed principles and specifics, limited guidance
- Weak (0.25): Mostly context-specific, unclear principles
- Missing (0.0): No abstraction, purely instance-specific

**Evidence Required**:
- Transferability analysis
- Cross-language examples
- Adaptation guidelines
- Principle extraction

**Calculation**:
```
V_reusability = (language_independence_score + codebase_generality_score + abstraction_quality_score) / 3
```

### Final V_meta

```
V_meta = 0.4 × V_completeness + 0.3 × V_effectiveness + 0.3 × V_reusability
```

## Bias Avoidance Protocol

**CRITICAL**: Apply these checks before finalizing scores:

1. **Seek Disconfirming Evidence**:
   - For each score ≥0.7: What evidence contradicts this?
   - Lower score if counterevidence exists

2. **Enumerate Gaps Explicitly**:
   - List missing elements in each rubric category
   - Don't gloss over weaknesses

3. **Ground in Concrete Evidence**:
   - Every score must cite specific artifacts
   - Avoid vague assessments like "seems good"

4. **Challenge High Scores**:
   - Scores ≥0.8 require exceptional evidence
   - Ask: "What would make this 1.0? Why isn't it there?"

5. **Independent Layer Evaluation**:
   - Do NOT let V_instance influence V_meta
   - Evaluate methodology quality separately from task success

## Output Format

Create two files:

### `data/iteration-N/value-instance.md`

```markdown
# V_instance Calculation - Iteration N

## Component Scores

### V_code_quality = [score]

**Complexity Reduction**:
- Baseline average: [value]
- Current average: [value]
- Reduction: [%]
- Rubric score: [0.0-1.0]

**Duplication Elimination**:
- Baseline blocks: [count]
- Current blocks: [count]
- Reduction: [%]
- Rubric score: [0.0-1.0]

**Static Analysis**:
- Baseline warnings: [count]
- Current warnings: [count]
- Reduction: [%]
- Rubric score: [0.0-1.0]

**Component Score**: ([complexity] + [duplication] + [static]) / 3 = [score]

**Evidence**: [List specific files, functions, metrics]

### V_maintainability = [score]

[Similar detailed breakdown]

### V_safety = [score]

[Similar detailed breakdown]

### V_effort = [score]

[Similar detailed breakdown]

## Final V_instance

V_instance = 0.3×[code_quality] + 0.3×[maintainability] + 0.2×[safety] + 0.2×[effort]
V_instance = **[final score]**

## Evidence Summary

[Comprehensive list of all evidence supporting scores]

## Gaps Identified

[Explicit list of weaknesses and missing elements]
```

### `data/iteration-N/value-meta.md`

```markdown
# V_meta Calculation - Iteration N

## Component Scores

### V_completeness = [score]

**Detection Phase** = [score]
- Taxonomy coverage: [assessment]
- Automation: [assessment]
- Prioritization: [assessment]
- Evidence: [artifacts]
- Gaps: [list]

**Planning Phase** = [score]
[Similar breakdown]

**Execution Phase** = [score]
[Similar breakdown]

**Verification Phase** = [score]
[Similar breakdown]

**Component Score**: ([detection] + [planning] + [execution] + [verification]) / 4 = [score]

### V_effectiveness = [score]

[Similar detailed breakdown with evidence]

### V_reusability = [score]

[Similar detailed breakdown with transferability analysis]

## Final V_meta

V_meta = 0.4×[completeness] + 0.3×[effectiveness] + 0.3×[reusability]
V_meta = **[final score]**

## Evidence Summary

[Independent evidence - NOT derived from V_instance]

## Methodology Gaps

[Explicit list of methodology weaknesses]

## Disconfirming Evidence

[Evidence that challenged high scores]
```

## Success Criteria

- Both value functions calculated with evidence
- All rubrics applied rigorously
- Gaps enumerated explicitly
- Scores grounded in concrete data
- Independent evaluation maintained
- Bias avoidance protocol followed

## Convergence Thresholds

- V_instance ≥ 0.75
- V_meta ≥ 0.70
- Must be sustained for 2 consecutive iterations
