# Example: Applying Documentation Patterns

**Context**: Demonstrate how to apply the three core documentation patterns (Progressive Disclosure, Example-Driven Explanation, Problem-Solution Structure) to improve documentation quality.

**Objective**: Show concrete before/after examples of pattern application.

---

## Pattern 1: Progressive Disclosure

### Problem
Documentation that presents all complexity at once overwhelms readers.

### Bad Example (Before)

```markdown
# Value Functions

V_instance = (Accuracy + Completeness + Usability + Maintainability) / 4
V_meta = (Completeness + Effectiveness + Reusability + Validation) / 4

Accuracy measures technical correctness including link validity, command
syntax, example functionality, and concept precision. Completeness evaluates
user need coverage, edge case handling, prerequisite clarity, and example
sufficiency. Usability assesses navigation intuitiveness, example concreteness,
jargon definition, and progressive disclosure application. Maintainability
examines modular structure, automated validation, version tracking, and
update ease.

V_meta Completeness measures lifecycle phase coverage (needs analysis,
strategy, execution, validation, maintenance), pattern catalog completeness,
template library completeness, and automation tool completeness...
```

**Issues**:
- All details dumped at once
- No clear progression (simple → complex)
- Reader overwhelmed immediately
- No logical entry point

### Good Example (After - Progressive Disclosure Applied)

```markdown
# Value Functions

BAIME uses two value functions to assess quality:
- **V_instance**: Documentation quality (how good is this doc?)
- **V_meta**: Methodology quality (how good is this methodology?)

Both range from 0.0 to 1.0. Target: ≥0.80 for production-ready.

## V_instance (Documentation Quality)

**Simple Formula**: Average of 4 components
- Accuracy: Is it correct?
- Completeness: Does it cover all user needs?
- Usability: Is it easy to use?
- Maintainability: Is it easy to maintain?

**Example**:
If Accuracy=0.75, Completeness=0.85, Usability=0.80, Maintainability=0.85:
V_instance = (0.75 + 0.85 + 0.80 + 0.85) / 4 = 0.8125 ≈ 0.82 ✅

### Component Details

**Accuracy (0.0-1.0)**: Technical correctness
- All links work?
- Commands run as documented?
- Examples realistic and tested?
- Concepts explained correctly?

**Completeness (0.0-1.0)**: User need coverage
- All questions answered?
- Edge cases covered?
- Prerequisites clear?
- Examples sufficient?

... (continue with other components)

## V_meta (Methodology Quality)

(Similar progressive structure: simple → detailed)
```

**Improvements**:
1. ✅ Start with "what" (2 value functions)
2. ✅ Simple explanation before formula
3. ✅ Example before detailed components
4. ✅ Details deferred to subsections
5. ✅ Reader can stop at any level

**Result**: Readers grasp concept quickly, dive deeper as needed.

---

## Pattern 2: Example-Driven Explanation

### Problem
Abstract concepts without concrete examples don't stick.

### Bad Example (Before)

```markdown
# Template Reusability

Templates are designed for cross-domain transferability with minimal
adaptation overhead. The parameterization strategy enables domain-agnostic
structure preservation while accommodating context-specific content variations.
Template instantiation follows a substitution-based approach where placeholders
are replaced with domain-specific values while maintaining structural integrity.
```

**Issues**:
- Abstract jargon ("transferability", "parameterization", "substitution-based")
- No concrete example
- Reader can't visualize usage
- Unclear benefit

### Good Example (After - Example-Driven Applied)

```markdown
# Template Reusability

Templates work across different documentation types with minimal changes.

**Example**: Tutorial Structure Template

**Generic Template** (domain-agnostic):
```
## What is [FEATURE_NAME]?
[FEATURE_NAME] is a [CATEGORY] that [PRIMARY_BENEFIT].

## When to Use [FEATURE_NAME]
Use [FEATURE_NAME] when:
- [USE_CASE_1]
- [USE_CASE_2]
```

**Applied to Testing** (domain-specific):
```
## What is Table-Driven Testing?
Table-Driven Testing is a testing pattern that reduces code duplication.

## When to Use Table-Driven Testing
Use Table-Driven Testing when:
- Testing multiple input/output combinations
- Reducing test code duplication
```

**Applied to Error Handling** (different domain):
```
## What is Sentinel Error Pattern?
Sentinel Error Pattern is an error handling approach that enables error checking.

## When to Use Sentinel Error Pattern
Use Sentinel Error Pattern when:
- Need to distinguish specific error types
- Callers need to handle errors differently
```

**Key Insight**: Same template structure, different domain content.
~90% structure preserved, ~10% adaptation for domain specifics.
```

**Improvements**:
1. ✅ Concept stated clearly first
2. ✅ Immediate concrete example (Testing)
3. ✅ Second example shows transferability (Error Handling)
4. ✅ Explicit benefit (90% reuse)
5. ✅ Reader sees exactly how to use template

**Result**: Readers understand concept through examples, not abstraction.

---

## Pattern 3: Problem-Solution Structure

### Problem
Documentation organized around features, not user problems.

### Bad Example (Before - Feature-Centric)

```markdown
# FAQ Command

The FAQ command displays frequently asked questions.

## Syntax
`/meta "faq"`

## Options
- No options available

## Output
Returns FAQ entries in markdown format

## Implementation
Uses MCP query_user_messages tool with pattern matching

## See Also
- /meta "help"
- Documentation guide
```

**Issues**:
- Organized around command features
- Doesn't address user problems
- Unclear when to use
- No problem-solving context

### Good Example (After - Problem-Solution Structure)

```markdown
# Troubleshooting: Finding Documentation Quickly

## Problem: "I have a question but don't know where to look"

**Symptoms**:
- Need quick answer to common question
- Don't want to read full documentation
- Searching docs takes too long

**Diagnosis**:
You need FAQ-style quick reference.

**Solution**: Use FAQ command
```bash
/meta "faq"
```

**What You'll Get**:
- 10-15 most common questions
- Concise answers
- Links to detailed docs

**Example**:
```
Q: How do I query error tool calls?
A: Use: get_session_stats() with status="error" filter
   See: docs/guides/mcp.md#error-analysis
```

**When This Works**:
- ✅ Question is common (covered in FAQ)
- ✅ Need quick answer (not deep dive)
- ✅ General question (not project-specific)

**When This Doesn't Work**:
- ❌ Complex debugging (use /meta "analyze errors" instead)
- ❌ Need comprehensive guide (read full docs)
- ❌ Project-specific issue (analyze your session data)

**Alternative Solutions**:
- Full search: `/meta "search [topic]"`
- Error analysis: `/meta "analyze errors"`
- Documentation: Browse docs/ directory
```

**Improvements**:
1. ✅ Starts with user problem
2. ✅ Symptoms → Diagnosis → Solution flow
3. ✅ Concrete example of output
4. ✅ Clear when to use / not use
5. ✅ Alternative solutions for edge cases

**Result**: Users find solutions to their problems, not feature descriptions.

---

## Combining Patterns

### Example: BAIME Troubleshooting Section

**Context**: Create troubleshooting guide for BAIME methodology using all 3 patterns.

**Approach**:
1. **Problem-Solution** structure overall
2. **Progressive Disclosure** within each problem (simple → complex)
3. **Example-Driven** for each solution

### Result

```markdown
# BAIME Troubleshooting

## Problem 1: "Iterations aren't converging" (Simple Problem First)

**Symptoms**:
- Value scores stagnant (∆V < 0.05 for 2+ iterations)
- Gap to threshold not closing
- Unclear what to improve

**Diagnosis**: Insufficient gap analysis or wrong priorities

**Solution 1: Analyze Gap Components** (Simple Solution First)

Break down V_instance gap by component:
- Accuracy gap: -0.10 → Focus on technical correctness
- Completeness gap: -0.05 → Add missing sections
- Usability gap: -0.15 → Improve examples and navigation
- Maintainability gap: 0.00 → No action needed

**Example**: (Concrete Application)
```
Iteration 2: V_instance = 0.70
Target: V_instance = 0.80
Gap: -0.10

Components:
- Accuracy: 0.75 (gap -0.05)
- Completeness: 0.60 (gap -0.20) ← CRITICAL
- Usability: 0.70 (gap -0.10)
- Maintainability: 0.75 (gap -0.05)

**Conclusion**: Prioritize Completeness (largest gap)
**Action**: Add second domain example (+0.15 Completeness expected)
```

**Advanced**: (Detailed Solution - Progressive Disclosure)
If simple gap analysis doesn't reveal priorities:
1. Calculate ROI for each improvement (∆V / hours)
2. Identify critical path items (must-have vs nice-to-have)
3. Use Tier system (Tier 1 mandatory, Tier 2 high-value, Tier 3 defer)

... (continue with more problems, each following same pattern)

## Problem 2: "System keeps evolving (M_n ≠ M_{n-1})" (Complex Problem Later)

**Symptoms**:
- Capabilities changing every iteration
- Agents being added/removed
- System feels unstable

**Diagnosis**: Domain complexity or insufficient specialization

**Solution**: Evaluate whether evolution is necessary

... (continues)
```

**Pattern Application**:
1. ✅ **Problem-Solution**: Organized around problems users face
2. ✅ **Progressive Disclosure**: Simple problems first, simple solutions before advanced
3. ✅ **Example-Driven**: Every solution has concrete example

**Result**: Users quickly find and solve their specific problems.

---

## Pattern Selection Guide

### When to Use Progressive Disclosure

**Use When**:
- Topic is complex (multiple layers of detail)
- Target audience has mixed expertise (beginners + experts)
- Concept builds on prerequisite knowledge
- Risk of overwhelming readers

**Example Scenarios**:
- Tutorial documentation (start simple, add complexity)
- Concept explanations (definition → details → edge cases)
- Architecture guides (overview → components → interactions)

**Don't Use When**:
- Topic is simple (single concept, no layers)
- Audience is uniform (all experts or all beginners)
- Reference documentation (users need quick lookup)

### When to Use Example-Driven

**Use When**:
- Explaining abstract concepts
- Demonstrating patterns or templates
- Teaching methodology or workflow
- Showing before/after improvements

**Example Scenarios**:
- Pattern documentation (concept + example)
- Template guides (structure + application)
- Methodology tutorials (theory + practice)

**Don't Use When**:
- Concept is self-explanatory
- Examples would be contrived
- Pure reference documentation (API, CLI)

### When to Use Problem-Solution

**Use When**:
- Creating troubleshooting guides
- Documenting error handling
- Addressing user pain points
- FAQ sections

**Example Scenarios**:
- Troubleshooting guides (symptom → solution)
- Error recovery documentation
- FAQ sections
- Debugging guides

**Don't Use When**:
- Documenting features (use feature-centric)
- Tutorial walkthroughs (use progressive disclosure)
- Concept explanations (use example-driven)

---

## Validation

### How to Know Patterns Are Working

**Progressive Disclosure**:
- ✅ Readers can stop at any level and understand
- ✅ Beginners aren't overwhelmed
- ✅ Experts can skip to advanced sections
- ✅ TOC shows clear hierarchy

**Example-Driven**:
- ✅ Every abstract concept has concrete example
- ✅ Examples realistic and tested
- ✅ Readers say "I see how to use this"
- ✅ Examples vary (simple → complex)

**Problem-Solution**:
- ✅ Users find their problem quickly
- ✅ Solutions actionable (can apply immediately)
- ✅ Alternative solutions for edge cases
- ✅ Users say "This solved my problem"

### Common Mistakes

**Progressive Disclosure**:
- ❌ Starting with complex details
- ❌ No clear progression (jumping between levels)
- ❌ Advanced topics mixed with basics

**Example-Driven**:
- ❌ Abstract explanation without example
- ❌ Contrived or unrealistic examples
- ❌ Single example (doesn't show variations)

**Problem-Solution**:
- ❌ Organized around features, not problems
- ❌ Solutions not actionable
- ❌ Missing "when to use / not use"

---

## Conclusion

**Key Takeaways**:
1. **Progressive Disclosure** reduces cognitive load (simple → complex)
2. **Example-Driven** makes abstract concepts concrete
3. **Problem-Solution** matches user mental model (problems, not features)

**Pattern Combinations**:
- Troubleshooting: Problem-Solution + Progressive Disclosure + Example-Driven
- Tutorial: Progressive Disclosure + Example-Driven
- Reference: Example-Driven (no progressive disclosure needed)

**Validation**:
- Test patterns on target audience
- Measure user success (can they find solutions?)
- Iterate based on feedback

**Next Steps**:
- Apply patterns to your documentation
- Validate with users
- Refine based on evidence
