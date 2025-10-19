# Template: Example Walkthrough

**Purpose**: Structured template for creating end-to-end practical examples in documentation
**Based on**: Testing methodology example from BAIME guide
**Validated**: 1 use, ready for reuse

---

## When to Use This Template

✅ **Use for**:
- End-to-end workflow demonstrations
- Real-world use case examples
- Tutorial practical sections
- "How do I accomplish X?" documentation

❌ **Don't use for**:
- Code snippets (use inline examples)
- API reference examples (use API docs format)
- Concept explanations (use concept template)
- Quick tips (use list format)

---

## Template Structure

```markdown
## Practical Example: [Use Case Name]

**Scenario**: [1-2 sentence description of what we're accomplishing]

**Domain**: [Problem domain - testing, CI/CD, etc.]

**Time to complete**: [Estimate]

---

### Context

**Problem**: [What problem are we solving?]

**Goal**: [What we want to achieve]

**Starting state**:
- [Condition 1]
- [Condition 2]
- [Condition 3]

**Success criteria**:
- [Measurable outcome 1]
- [Measurable outcome 2]

---

### Prerequisites

**Required**:
- [Tool/knowledge 1]
- [Tool/knowledge 2]

**Files needed**:
- `[path/to/file]` - [Purpose]

**Setup**:
```bash
[Setup commands if needed]
```

---

### Workflow

#### Phase 1: [Phase Name]

**Objective**: [What this phase accomplishes]

**Step 1**: [Action]

[Explanation of what we're doing]

```[language]
[Code or command]
```

**Output**:
```
[Expected output]
```

**Why this matters**: [Reasoning]

**Step 2**: [Continue pattern]

**Phase 1 Result**: [What we have now]

---

#### Phase 2: [Phase Name]

[Repeat structure for 2-4 phases]

---

#### Phase 3: [Phase Name]

---

### Results

**Outcomes achieved**:
- ✅ [Outcome 1 with metric]
- ✅ [Outcome 2 with metric]
- ✅ [Outcome 3 with metric]

**Before and after comparison**:
| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| [Metric 1] | [Value] | [Value] | [%/x] |
| [Metric 2] | [Value] | [Value] | [%/x] |

**Artifacts created**:
- `[file]` - [Description]
- `[file]` - [Description]

---

### Takeaways

**What we learned**:
1. [Insight 1]
2. [Insight 2]
3. [Insight 3]

**Key patterns observed**:
- [Pattern 1]
- [Pattern 2]

**Next steps**:
- [What to do next]
- [How to extend this example]

---

### Variations

**For different scenarios**:

**Scenario A**: [Variation description]
- Change: [What's different]
- Impact: [How it affects workflow]

**Scenario B**: [Another variation]
- Change: [What's different]
- Impact: [How it affects workflow]

---

### Troubleshooting

**Common issues in this example**:

**Issue 1**: [Problem]
- **Symptoms**: [How to recognize]
- **Cause**: [Why it happens]
- **Solution**: [How to fix]

**Issue 2**: [Continue pattern]

```

---

## Section Guidelines

### Scenario
- **Length**: 1-2 sentences
- **Specificity**: Concrete, not abstract ("Create testing strategy for Go project", not "Use BAIME for testing")
- **Appeal**: Should sound relevant to target audience

### Context
- **Problem statement**: Clear pain point
- **Starting state**: Observable conditions (can be verified)
- **Success criteria**: Measurable (coverage %, time, error rate, etc.)

### Workflow
- **Organization**: By logical phases (2-4 phases)
- **Detail level**: Sufficient to reproduce
- **Code blocks**: Runnable, copy-paste ready
- **Explanations**: "Why" not just "what"

### Results
- **Metrics**: Quantitative when possible
- **Comparison**: Before/after table
- **Artifacts**: List all files created

### Takeaways
- **Insights**: What was learned
- **Patterns**: What emerged from practice
- **Generalization**: How to apply elsewhere

---

## Quality Checklist

**Completeness**:
- [ ] All prerequisites listed
- [ ] Starting state clearly defined
- [ ] Success criteria measurable
- [ ] All phases documented
- [ ] Results quantified
- [ ] Artifacts listed

**Reproducibility**:
- [ ] Commands are copy-paste ready
- [ ] File paths are clear
- [ ] Setup instructions complete
- [ ] Expected outputs shown
- [ ] Tested on clean environment

**Clarity**:
- [ ] Each step has explanation
- [ ] "Why" provided for key decisions
- [ ] Phases logically organized
- [ ] Progression clear (what we have after each phase)

**Realism**:
- [ ] Based on real use case (not toy example)
- [ ] Complexity matches real-world (not oversimplified)
- [ ] Metrics are actual measurements (not estimates)
- [ ] Problems/challenges acknowledged

---

## Example: Testing Methodology Walkthrough

**Actual example from BAIME guide** (simplified):

```markdown
## Practical Example: Testing Methodology

**Scenario**: Developing systematic testing strategy for Go project using BAIME

**Domain**: Software testing
**Time to complete**: 6-8 hours across 3-5 iterations

---

### Context

**Problem**: Ad-hoc testing approach, coverage at 60%, no systematic strategy

**Goal**: Reach 80%+ coverage with reusable testing patterns

**Starting state**:
- Go project with 10K lines
- 60% test coverage
- Mix of unit and integration tests
- No testing standards

**Success criteria**:
- Test coverage ≥ 80%
- Testing patterns documented
- Methodology transferable to other Go projects (≥70%)

---

### Workflow

#### Phase 1: Baseline (Iteration 0)

**Objective**: Measure current state and identify gaps

**Step 1**: Measure coverage
```bash
go test -cover ./...
# Output: coverage: 60.2% of statements
```

**Step 2**: Analyze test quality
- Found 15 untested edge cases
- Identified 3 patterns: table-driven, golden file, integration

**Phase 1 Result**: Baseline established (V_instance=0.40, V_meta=0.20)

---

#### Phase 2: Pattern Codification (Iterations 1-2)

**Objective**: Extract and document testing patterns

**Step 1**: Extract table-driven pattern
```go
// Pattern: Table-driven tests
func TestFunction(t *testing.T) {
    tests := []struct {
        name string
        input int
        want int
    }{
        {"zero", 0, 0},
        {"positive", 5, 25},
        {"negative", -3, 9},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Function(tt.input)
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

**Step 2**: Document 8 testing patterns
**Step 3**: Create test templates

**Phase 2 Result**: Patterns documented, coverage at 72%

---

#### Phase 3: Automation (Iteration 3)

**Objective**: Automate pattern detection and enforcement

**Step 1**: Create coverage analyzer script
**Step 2**: Create test generator tool
**Step 3**: Add pre-commit hooks

**Phase 3 Result**: Coverage at 86%, automated quality gates

---

### Results

**Outcomes achieved**:
- ✅ Coverage: 60% → 86% (+26 percentage points)
- ✅ Methodology: 8 patterns, 3 tools, comprehensive guide
- ✅ Transferability: 89% to other Go projects

**Before and after comparison**:
| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Coverage | 60% | 86% | +26 pp |
| Test generation time | 30 min | 2 min | 15x |
| Pattern consistency | Ad-hoc | Enforced | 100% |

**Artifacts created**:
- `docs/testing-strategy.md` - Complete methodology
- `scripts/coverage-analyzer.sh` - Coverage analysis tool
- `scripts/test-generator.sh` - Test template generator
- `patterns/*.md` - 8 testing patterns

---

### Takeaways

**What we learned**:
1. Table-driven tests are most common pattern (60% of tests)
2. Coverage gaps mostly in error handling paths
3. Automation provides 15x speedup over manual

**Key patterns observed**:
- Progressive coverage improvement (60→72→86)
- Value convergence in 3 iterations (faster than expected)
- Patterns emerged from practice, not designed upfront

**Next steps**:
- Apply to other Go projects to validate 89% transferability claim
- Add mutation testing for quality validation
- Expand pattern library based on new use cases
```

---

## Variations

### Variation 1: Quick Example (< 5 min)

For simple, focused examples:

```markdown
## Example: [Task]

**Task**: [What we're doing]

**Steps**:
1. [Action]
   ```
   [Code]
   ```
2. [Action]
   ```
   [Code]
   ```
3. [Action]
   ```
   [Code]
   ```

**Result**: [What we achieved]
```

### Variation 2: Comparison Example

When showing before/after or comparing approaches:

```markdown
## Example: [Comparison]

**Scenario**: [Context]

### Approach A: [Name]
[Implementation]
**Pros**: [Benefits]
**Cons**: [Drawbacks]

### Approach B: [Name]
[Implementation]
**Pros**: [Benefits]
**Cons**: [Drawbacks]

### Recommendation
[Which to use when]
```

### Variation 3: Error Recovery Example

For troubleshooting documentation:

```markdown
## Example: Recovering from [Error]

**Symptom**: [What user sees]

**Diagnosis**:
1. Check [aspect]
   ```
   [Diagnostic command]
   ```
2. Verify [aspect]

**Solution**:
1. [Fix step]
   ```
   [Fix command]
   ```
2. [Verification step]

**Prevention**: [How to avoid in future]
```

---

## Validation

**Usage**: 1 complete walkthrough (Testing Methodology in BAIME guide)

**Effectiveness**:
- ✅ Clear phases and progression
- ✅ Realistic (based on actual experiment)
- ✅ Quantified results (metrics, before/after)
- ✅ Reproducible (though conceptual, not literal)

**Gaps identified in Iteration 0**:
- ⚠️ Example was conceptual, not literally tested
- ⚠️ Should be more specific (actual commands, actual output)

**Improvements for next use**:
- Make example literally reproducible (test every command)
- Add troubleshooting section specific to example
- Include timing for each phase

---

## Related Templates

- [tutorial-structure.md](tutorial-structure.md) - Practical Example section uses this template
- [concept-explanation.md](concept-explanation.md) - Uses brief examples; walkthrough provides depth

---

**Status**: ✅ Ready for use | Validated in 1 context | Refinement needed for reproducibility
**Maintenance**: Update based on example effectiveness feedback
