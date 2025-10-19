# Pattern Extraction Rules

**Purpose**: Guidelines for identifying and extracting reusable patterns from BAIME experiment artifacts.

**Version**: 1.0
**Created**: 2025-10-19

---

## What is a Pattern?

**Definition**: A reusable solution to a recurring problem in a specific context.

**Components**:
1. **Name**: Clear, descriptive identifier (e.g., "Extract Method", "Characterization Tests")
2. **Context**: When/where the pattern applies
3. **Problem**: What issue it solves
4. **Solution**: Step-by-step approach
5. **Example**: Concrete code example (before/after)
6. **Evidence**: Validation data from experiment (applications count, success rate, impact metrics)

---

## Where to Find Patterns

### Primary Sources

1. **results.md**:
   - Look for: "Patterns" section (usually in "Domain Results" or "Learnings")
   - Format: Often enumerated lists or subsections
   - Example markers: "### Pattern:", "Pattern:", "Pattern 1:", numbered lists

2. **iteration reports** (`iterations/*.md`):
   - Look for: "Pattern Extraction" sections, "What Worked Well" sections
   - Context: Patterns emerge during execution, documented in reflection
   - Example markers: "Pattern used:", "Applied pattern:", "New pattern discovered"

3. **templates** (`knowledge/templates/*.md`):
   - Look for: Step-by-step processes that solve recurring problems
   - Note: Templates are themselves patterns in documented form

### Secondary Sources

4. **Code changes** (if available in experiment):
   - Look for: Repeated refactoring techniques, common code structures
   - Extract: Abstract the technique from specific implementation

5. **Problem lists** ("Problems and Gaps" sections):
   - Look for: Recurring issues that were solved
   - Extract: The solution as a pattern

---

## Identification Criteria

### A knowledge item is a PATTERN if:

✅ **Reusable**: Applied ≥2 times in experiment OR clearly applicable to other scenarios
✅ **Specific**: Concrete steps, not vague advice
✅ **Validated**: Evidence of success (metrics, test results, applications count)
✅ **Problem-solving**: Addresses a clear problem or challenge
✅ **Context-bound**: Specifies when/where to apply

### NOT a pattern if:

❌ **One-time solution**: Used once, specific to unique scenario
❌ **Vague guideline**: "Write good code" (too abstract)
❌ **Unvalidated**: No evidence it works
❌ **Trivial**: Obvious solution with no insight (e.g., "read the documentation")

---

## Extraction Process

### Step 1: Scan for Pattern Markers

**Commands**:
```bash
# Find pattern sections in results.md
grep -n "Pattern" experiments/[experiment]/results.md

# Find pattern applications in iterations
grep -rn "Pattern:" experiments/[experiment]/iterations/

# List templates (often patterns in disguise)
ls experiments/[experiment]/knowledge/templates/
```

**Expected Output**: Line numbers where patterns are mentioned

---

### Step 2: Read and Classify

For each potential pattern:

1. **Read full description** (context, problem, solution)
2. **Check evidence**:
   - How many times applied? (≥2 preferred)
   - Success rate? (100% ideal, ≥80% acceptable)
   - Quantified impact? (metrics, time savings, quality improvement)
3. **Classify**:
   - **Core pattern**: Central to methodology, high impact, frequently used
   - **Supporting pattern**: Helpful but not critical, occasional use
   - **Domain-specific**: Only applies to specific domain (e.g., refactoring-only)
   - **Universal**: Applies across domains (e.g., TDD, incremental commits)

---

### Step 3: Extract Components

For each pattern, extract:

**Required**:
- **Name**: From source (use exact name if available)
- **Context**: "When to use" (from pattern description or inferred)
- **Problem**: "What issue it solves" (from pattern description)
- **Solution**: Step-by-step (numbered list, ≥3 steps)
- **Validation**: Evidence (applications count, success rate, metrics)

**Optional but Recommended**:
- **Example**: Code/text before/after (from iterations or results.md)
- **Safety notes**: Precautions, common mistakes
- **Metrics**: Quantified impact (time saved, quality improved)
- **Transferability**: Language/domain independence assessment

---

### Step 4: Format Consistently

**Template**:
```markdown
### Pattern: [Name]

**Context**: [When to use this pattern]

**Problem**: [What issue does it solve?]

**Solution**:
1. [Step 1]
2. [Step 2]
3. [Step 3]
[...additional steps]

**Example**:
```[language]
// Before
[code or text showing problem]

// After
[code or text showing solution]
```

**Validation**:
- **Applications**: [Count] times in [Experiment]
- **Success Rate**: [Percentage]% ([successful]/[total])
- **Impact**: [Quantified metrics]

**Transferability**: [Universal / Language-specific / Domain-specific]

**Safety**: [Precautions, common mistakes to avoid]
```

---

## Pattern Categories

### By Domain

1. **Refactoring Patterns**: Extract Method, Simplify Conditionals, Remove Duplication
2. **Testing Patterns**: Characterization Tests, TDD Workflow, Test Fixture Setup
3. **Error Handling Patterns**: Sentinel Errors, Context Preservation, Error Wrapping
4. **Workflow Patterns**: Incremental Commits, Safety Checklist, Automated Validation

### By Abstraction Level

1. **Low-Level** (code-specific): Extract Method, Inline Variable
2. **Mid-Level** (process-specific): TDD Workflow, Safety Checklist
3. **High-Level** (principle-based): Incremental Safety, Behavior Preservation

### By Transferability

1. **Universal**: Apply to any language/domain (e.g., TDD, incremental commits)
2. **Language-Family**: Apply to specific language family (e.g., OOP refactorings)
3. **Domain-Specific**: Apply to specific domain only (e.g., Go-specific complexity thresholds)

---

## Common Pattern Sources

### From Iteration Reports

**Markers**:
- "Pattern used:"
- "Applied [Pattern Name]"
- "Technique:"
- "Approach:"
- "What worked well:" (often describes patterns)

**Example**:
> "Applied Extract Method pattern twice, reducing complexity by 70%"

**Extract**: Extract Method pattern with validation data (2 applications, 70% reduction)

---

### From results.md

**Markers**:
- "## Patterns Discovered"
- "### Pattern: [Name]"
- Enumerated lists under "Patterns" heading
- "Reusable approach:" sections

**Example**:
> "### Pattern: Characterization Tests
> Used to document legacy behavior before refactoring. Applied 9 times with 100% success preventing regressions."

**Extract**: Characterization Tests pattern with complete evidence

---

### From Templates

**Markers**:
- Templates themselves are often patterns
- Look for: Step-by-step processes, checklists, workflows

**Example**:
> Template: `refactoring-safety-checklist.md` (276 lines)

**Extract**: "Safety Checklist" pattern (process for zero-regression refactoring)

---

## Validation Checklist

Before finalizing extracted patterns:

- [ ] **Name** is clear and descriptive?
- [ ] **Context** specifies when to use?
- [ ] **Problem** clearly stated?
- [ ] **Solution** has ≥3 concrete steps?
- [ ] **Example** provided (code or text)?
- [ ] **Validation** includes quantified evidence?
- [ ] **Transferability** assessed (universal/specific)?
- [ ] **Source references** documented (file, line numbers)?

---

## Common Pitfalls

### Pitfall 1: Extracting Non-Patterns

**Issue**: Extracting guidelines that don't meet pattern criteria

**Example**: "Write good tests" (too vague, no concrete steps)

**Fix**: Look for specific techniques with step-by-step solutions

---

### Pitfall 2: Missing Validation Data

**Issue**: Pattern without evidence it works

**Example**: "Use Extract Method" (no applications count, success rate, or metrics)

**Fix**: Cross-reference with iteration reports for validation data

---

### Pitfall 3: Incomplete Steps

**Issue**: Solution has vague or missing steps

**Example**:
> Solution:
> 1. Identify code to extract
> 2. Extract it
> 3. Test

**Fix**: Add specifics:
> Solution:
> 1. Identify cohesive code block (5-10 lines, single responsibility)
> 2. Write test for extracted behavior if not covered
> 3. Extract to new function with descriptive name
> 4. Run all tests (must pass 100%)
> 5. Commit with descriptive message

---

### Pitfall 4: Confusing Patterns and Principles

**Pattern**: Concrete solution with steps (e.g., "Extract Method: 1. Identify block, 2. Write test, 3. Extract")

**Principle**: Abstract guideline (e.g., "Behavior Preservation: Maintain exact original behavior during refactoring")

**Fix**: If it has concrete steps → Pattern. If it's a guiding value → Principle.

---

## Time Estimates

| Task | Time Estimate |
|------|---------------|
| Scan for patterns (grep, skim) | 5-10 min |
| Read and classify (per pattern) | 2-3 min |
| Extract components (per pattern) | 3-5 min |
| Format consistently (per pattern) | 2-3 min |
| **Total per pattern** | **7-11 min** |
| **For 8 patterns** | **56-88 min** |

**Optimization**: Extract in batches (read all, then classify all, then extract all) for efficiency

---

## Examples

### Example 1: Well-Extracted Pattern

```markdown
### Pattern: Extract Method

**Context**: Functions with cyclomatic complexity >8, multiple responsibilities

**Problem**: High complexity makes code hard to understand, test, and maintain

**Solution**:
1. Identify cohesive code block (5-10 lines, single responsibility)
2. Write test for extracted behavior if not already covered
3. Extract to new function with descriptive, intention-revealing name
4. Run all tests (must pass 100%)
5. Commit with message: "refactor: extract [function name]"

**Example**:
```go
// Before (complexity: 10)
func calculate() int {
    // 10 lines of complex logic mixing concerns
}

// After (complexity: 3)
func calculate() int {
    data := collectData()
    return processData(data)
}
```

**Validation**:
- **Applications**: 2 times in Bootstrap-004
- **Success Rate**: 100% (2/2 successful)
- **Impact**: -43% to -70% complexity reduction

**Transferability**: Universal (applies to all languages)

**Safety**: Always write tests before extracting; run tests after each extraction
```

**Why this is good**:
✅ Clear name and context
✅ Specific problem stated
✅ Concrete 5-step solution
✅ Before/after code example
✅ Quantified validation data
✅ Transferability assessed
✅ Safety notes included

---

### Example 2: Poorly-Extracted Pattern (Don't do this)

```markdown
### Pattern: Good Testing

**Context**: Testing

**Problem**: Tests are important

**Solution**:
1. Write tests
2. Make them good

**Example**: (none)

**Validation**: Tests worked

**Transferability**: Universal
```

**Why this is bad**:
❌ Vague name ("Good Testing")
❌ Vague context ("Testing" - when specifically?)
❌ Vague problem ("important" - what specific issue?)
❌ Vague solution (no concrete steps)
❌ No example
❌ No quantified validation data
❌ No specifics

---

## Quick Reference

**Pattern Identification**:
- ✅ Reusable (≥2 applications)
- ✅ Specific steps
- ✅ Validated evidence
- ✅ Problem-solving
- ✅ Context-bound

**Pattern Sources**:
1. results.md → "Patterns" section
2. iterations/*.md → "Pattern used" markers
3. templates/ → Step-by-step processes

**Extraction Steps**:
1. Scan for markers (grep)
2. Read and classify
3. Extract components (name, context, problem, solution, validation)
4. Format consistently (use template)
5. Validate (checklist)

**Time**: ~10 min per pattern, ~80 min for 8 patterns

---

**Version**: 1.0
**Last Updated**: 2025-10-19
**Validated**: Bootstrap-005 Iteration 0-1 (8 patterns extracted successfully)
