# Template: Tutorial Structure

**Purpose**: Structured template for creating comprehensive technical tutorials
**Based on**: Progressive disclosure pattern + BAIME usage guide
**Validated**: 1 use (BAIME guide), ready for reuse

---

## When to Use This Template

✅ **Use for**:
- Complex frameworks or systems
- Topics requiring multiple levels of understanding
- Audiences with mixed expertise (beginners to experts)
- Topics where quick start is possible (< 10 min example)

❌ **Don't use for**:
- Simple how-to guides (< 5 steps)
- API reference documentation
- Quick tips or cheat sheets

---

## Template Structure

```markdown
# [Topic Name]

**[One-sentence description]** - [Core value proposition]

---

## Table of Contents

- [What is [Topic]?](#what-is-topic)
- [When to Use [Topic]](#when-to-use-topic)
- [Prerequisites](#prerequisites)
- [Core Concepts](#core-concepts)
- [Quick Start](#quick-start)
- [Step-by-Step Workflow](#step-by-step-workflow)
- [Advanced Topics](#advanced-topics) (if applicable)
- [Practical Example](#practical-example)
- [Troubleshooting](#troubleshooting)
- [Next Steps](#next-steps)

---

## What is [Topic]?

[2-3 paragraphs explaining the topic]

**Paragraph 1**: Integration/components
- What methodologies/tools does it integrate?
- How do they work together?

**Paragraph 2**: Key innovation
- What problem does it solve?
- How is it different from alternatives?

**Paragraph 3** (optional): Proof points
- Results from real usage
- Examples of applications

### Why [Topic]?

**Problem**: [Describe the pain point]

**Solution**: [Topic] provides systematic approach with:
- ✅ [Benefit 1 with metric]
- ✅ [Benefit 2 with metric]
- ✅ [Benefit 3 with metric]
- ✅ [Benefit 4 with metric]

### [Topic] in Action

**Example Results**:
- **[Domain 1]**: [Metric], [Transferability]
- **[Domain 2]**: [Metric], [Transferability]
- **[Domain 3]**: [Metric], [Transferability]

---

## When to Use [Topic]

### Use [Topic] For

✅ **[Category 1]** for:
- [Use case 1]
- [Use case 2]
- [Use case 3]

✅ **When you need**:
- [Need 1]
- [Need 2]
- [Need 3]

### Don't Use [Topic] For

❌ [Anti-pattern 1]
❌ [Anti-pattern 2]
❌ [Anti-pattern 3]

---

## Prerequisites

### Required

1. **[Tool/knowledge 1]**
   - [Installation/setup link]
   - Verify: [How to check it's working]

2. **[Tool/knowledge 2]**
   - [Setup instructions or reference]

3. **[Context requirement]**
   - [What the reader needs to have]
   - [How to measure current state]

### Recommended

- **[Optional tool/knowledge 1]**
  - [Why it helps]
  - [How to get it]

- **[Optional tool/knowledge 2]**
  - [Why it helps]
  - [Link to documentation]

---

## Core Concepts

**[Number] key concepts you need to understand**:

### 1. [Concept Name]

**Definition**: [1-2 sentence explanation]

**Why it matters**: [Practical reason]

**Example**:
```
[Code or conceptual example]
```

### 2. [Concept Name]

[Repeat structure]

### [3-6 total concepts]

---

## Quick Start

**Goal**: [What reader will accomplish] in 10 minutes

### Step 1: [Action]

[Brief instruction]

```bash
[Code block if applicable]
```

**Expected result**: [What should happen]

### Step 2: [Action]

[Continue for 3-5 steps maximum]

### Step 3: [Action]

### Step 4: [Action]

---

## Step-by-Step Workflow

**Complete guide** organized by phases or stages:

### Phase 1: [Phase Name]

**Purpose**: [What this phase accomplishes]

**Steps**:

1. **[Step name]**
   - [Detailed instructions]
   - **Why**: [Rationale]
   - **Example**: [If applicable]

2. **[Step name]**
   - [Continue pattern]

**Output**: [What you have after this phase]

### Phase 2: [Phase Name]

[Repeat structure for 2-4 phases]

### Phase 3: [Phase Name]

---

## [Advanced Topics] (Optional)

**For experienced users** who want to customize or extend:

### [Advanced Topic 1]

[Explanation]

### [Advanced Topic 2]

[Explanation]

---

## Practical Example

**Real-world walkthrough**: [Domain/use case]

### Context

[What problem we're solving]

### Setup

[Starting state]

### Execution

**Step 1**: [Action]
```
[Code/example]
```

**Result**: [Outcome]

**Step 2**: [Continue pattern]

### Outcome

[What we achieved]

[Metrics or concrete results]

---

## Troubleshooting

**Common issues and solutions**:

### Issue 1: [Problem description]

**Symptoms**:
- [Symptom 1]
- [Symptom 2]

**Cause**: [Root cause]

**Solution**:
```
[Fix or workaround]
```

### Issue 2: [Repeat structure for 5-7 common issues]

---

## Next Steps

**After mastering the basics**:

1. **[Next learning path]**
   - [Link to advanced guide]
   - [What you'll learn]

2. **[Complementary topic]**
   - [Link to related documentation]
   - [How it connects]

3. **[Community/support]**
   - [Where to ask questions]
   - [How to contribute]

**Further reading**:
- [Link 1]: [Description]
- [Link 2]: [Description]
- [Link 3]: [Description]

---

**Status**: [Version] | [Date] | [Maintenance status]
```

---

## Content Guidelines

### What is [Topic]? Section
- **Length**: 3-5 paragraphs
- **Tone**: Accessible, not overly technical
- **Include**: Problem statement, solution overview, proof points
- **Avoid**: Implementation details (save for later sections)

### Core Concepts Section
- **Count**: 3-6 concepts (7+ is too many)
- **Each concept**: Definition + why it matters + example
- **Order**: Most fundamental to most advanced
- **Examples**: Concrete, not abstract

### Quick Start Section
- **Time limit**: Must be completable in < 10 minutes
- **Steps**: 3-5 maximum
- **Complexity**: One happy path, no branching
- **Outcome**: Working example, not full understanding

### Step-by-Step Workflow Section
- **Organization**: By phases or logical groupings
- **Detail level**: Complete (all options, all decisions)
- **Examples**: Throughout, not just at end
- **Cross-references**: Link to concepts and troubleshooting

### Practical Example Section
- **Realism**: Based on actual use case, not toy example
- **Completeness**: End-to-end, showing all steps
- **Metrics**: Quantify outcomes when possible
- **Context**: Explain why this example matters

### Troubleshooting Section
- **Coverage**: 5-7 common issues
- **Structure**: Symptoms → Cause → Solution
- **Evidence**: Based on real problems (user feedback or anticipated)
- **Links**: Cross-reference to relevant sections

---

## Adaptation Guide

### For Simple Topics (< 5 concepts)
- **Omit**: Advanced Topics section
- **Combine**: Core Concepts + Quick Start
- **Simplify**: Step-by-Step Workflow (single section, not phases)

### For API Documentation
- **Omit**: Practical Example (use code examples instead)
- **Expand**: Core Concepts (one per major API concept)
- **Add**: API Reference section after Step-by-Step

### For Process Documentation
- **Omit**: Quick Start (processes don't always have quick paths)
- **Expand**: Step-by-Step Workflow (detailed process maps)
- **Add**: Decision trees for complex choices

---

## Quality Checklist

Before publishing, verify:

**Structure**:
- [ ] Table of contents present with working links
- [ ] All required sections present (What is, When to Use, Prerequisites, Core Concepts, Quick Start, Workflow, Example, Troubleshooting, Next Steps)
- [ ] Progressive disclosure (simple → complex)
- [ ] Clear section boundaries (headings, whitespace)

**Content**:
- [ ] Core concepts have examples (100%)
- [ ] Quick start is < 10 minutes
- [ ] Step-by-step workflow is complete (no "TBD" placeholders)
- [ ] Practical example is realistic and complete
- [ ] Troubleshooting covers 5+ issues

**Usability**:
- [ ] Links work (use validation tool)
- [ ] Code blocks have syntax highlighting
- [ ] Examples are copy-paste ready
- [ ] No broken forward references

**Accuracy**:
- [ ] Technical details verified (test examples)
- [ ] Metrics are current and accurate
- [ ] Links point to correct resources
- [ ] Prerequisites are complete and correct

---

## Example Usage

**Input**: Need to create tutorial for "API Design Methodology"

**Step 1**: Copy template

**Step 2**: Fill in topic-specific content
- What is API Design? → Explain methodology
- When to Use → API design scenarios
- Core Concepts → 5-6 API design principles
- Quick Start → Design first API in 10 min
- Workflow → Full design process
- Example → Real API design walkthrough
- Troubleshooting → Common API design problems

**Step 3**: Verify with checklist

**Step 4**: Validate links and examples

**Step 5**: Publish

---

## Validation

**First Use**: BAIME Usage Guide
- **Structure match**: 95% (omitted some optional sections)
- **Effectiveness**: Created comprehensive guide (V_instance = 0.66)
- **Learning**: Pattern worked well, validated structure

**Transferability**: Expected 90%+ (universal tutorial structure)

**Next Validation**: Apply to different domain (API docs, troubleshooting guide, etc.)

---

## Related Templates

- [concept-explanation.md](concept-explanation.md) - Template for explaining individual concepts
- [example-walkthrough.md](example-walkthrough.md) - Template for practical examples
- [progressive-disclosure pattern](../patterns/progressive-disclosure.md) - Underlying pattern

---

**Status**: ✅ Ready for use | Validated in 1 context | High confidence
**Maintenance**: Update based on usage feedback
