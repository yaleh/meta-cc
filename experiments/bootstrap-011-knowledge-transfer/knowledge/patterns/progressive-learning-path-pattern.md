# Pattern: Progressive Learning Path

**Pattern Type**: Learning Path Design
**Domain**: Knowledge Transfer, Onboarding
**Context**: New contributors need structured guidance from setup to first contribution
**Extracted From**: Iteration 1 - Day-1 Learning Path Design

---

## Problem

New contributors face information overload when exploring a codebase. Without structured guidance:
- They spend weeks randomly exploring instead of days with focused learning
- Critical concepts are discovered late or missed entirely
- Validation of progress is unclear
- Time to first contribution is unpredictable and long

## Context

**When to apply**:
- Onboarding new team members to a codebase
- Teaching complex systems with multiple concepts
- Need for self-paced, self-validated learning
- Limited mentor availability

**When not to apply**:
- Very small projects (<100 lines of code)
- One-off scripts with no team
- Contributors already familiar with the codebase

## Solution

Design learning paths as progressive stages with:

### 1. Time-Boxed Sections
```
Section 1: Environment Setup (1-2 hours)
   ↓
Section 2: Understanding Concepts (1-2 hours)
   ↓
Section 3: Code Exploration (1-2 hours)
   ↓
Section 4: First Contribution (2-4 hours)
```

**Key Principle**: Progressive disclosure - reveal complexity gradually

### 2. Clear Learning Objectives Per Section

Each section must answer:
- **What will you learn?** (objectives)
- **How will you know you learned it?** (validation checkpoints)
- **What can you do after this section?** (capabilities gained)

Example:
```markdown
### Section 1: Environment Setup
**Learning Objectives**:
- Install all dependencies
- Run test suite successfully
- Build project from source

**Validation Checkpoint**:
- [ ] `make all` passes without errors
- [ ] Can run `./binary --help`
```

### 3. Scaffolding Structure

Build knowledge incrementally:
- **Foundation first**: Setup → Understanding → Exploration → Contribution
- **Each section enables next**: Can't contribute without understanding; can't understand without setup
- **Prerequisite chains explicit**: "Before Section 2, complete Section 1"

### 4. Validation Checkpoints

Every section has self-assessment:
- **Checkbox format** for easy tracking
- **Clear criteria** (not subjective)
- **Actionable validation** (run command, explain concept, produce artifact)

Example:
```markdown
**Checkpoint**: Can run `./meta-cc --help` and see command list.
```

### 5. Time Estimates

Provide realistic time ranges:
- **Per section**: 1-2 hours, 2-4 hours
- **Total path**: 4-8 hours for Day-1
- **Account for variance**: Different experience levels take different times

### 6. Exit Criteria

Define clear "done" state:
- What artifacts were produced?
- What capabilities were gained?
- What can you do now that you couldn't before?

Example for Day-1:
```
**Success Criteria**:
- Working dev environment
- Basic understanding of project
- First PR submitted
```

## Structure Template

```markdown
# {Stage} Learning Path: {Project} {Role}

## Overview
- **Objective**: [One-sentence goal]
- **Time**: [X-Y hours]
- **Prerequisites**: [Required knowledge]
- **Success Criteria**: [Clear exit criteria]

## Section 1: {Topic} ({time estimate})
### Learning Objectives
- [What will you learn?]

### Steps
1. [Action to take]
2. [Action to take]
...

### Validation Checkpoint
- [ ] [Clear criteria]
- [ ] [Clear criteria]

## Section 2: ...
[Repeat structure]

## {Stage} Complete!
**What You Accomplished**:
- [Summary of capabilities gained]

**Next Steps**:
- [Link to next learning path]
```

## Consequences

**Benefits**:
- ✅ Reduces onboarding time from weeks to days
- ✅ Enables self-paced learning (no mentor dependency)
- ✅ Provides clear progress indicators
- ✅ Consistent onboarding experience across contributors
- ✅ Easy to maintain (structured sections)

**Trade-offs**:
- ❌ Upfront effort to design paths
- ❌ Requires maintenance when project changes
- ❌ Less flexibility for experienced contributors (may skip sections)
- ❌ One-size-fits-all may not fit all learning styles

**Mitigations**:
- Allow experienced contributors to skip validated sections
- Provide "fast track" alternative for experts
- Regular review and update of learning paths

## Examples

### Day-1 Path (meta-cc)
- **Time**: 4-8 hours
- **Sections**: 4 (Setup, Understanding, Exploration, Contribution)
- **Objectives**: 20 total learning objectives
- **Checkpoints**: 23 validation checkpoints
- **Outcome**: First PR submitted

### Week-1 Path (anticipated)
- **Time**: 20-40 hours
- **Sections**: ~6 (Architecture, Core Modules, Workflows, Feature)
- **Outcome**: Meaningful feature delivered

### Month-1 Path (anticipated)
- **Time**: 80-160 hours
- **Sections**: ~8 (Advanced Topics, Complex Feature, Mentoring)
- **Outcome**: Significant feature + mentoring capability

## Related Patterns

- **Spaced Repetition Pattern**: Review key concepts across multiple paths
- **Validation Checkpoint Pattern**: Clear self-assessment at each stage
- **Progressive Disclosure Pattern**: Reveal complexity gradually
- **Scaffolding Pattern**: Build on previous knowledge

## Validation Status

- **Status**: Proposed (extracted from Iteration 1)
- **Tested**: Not yet (awaits real contributor usage)
- **Next Step**: Validate with actual contributors, gather feedback, refine

---

**Pattern Extracted**: 2025-10-17 (Iteration 1)
**Source**: Day-1 Learning Path design for meta-cc
**Iteration**: bootstrap-011-knowledge-transfer/iteration-1
