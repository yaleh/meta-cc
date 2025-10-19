# Pattern: Progressive Disclosure

**Status**: ✅ Validated (2 uses)
**Domain**: Documentation
**Transferability**: Universal (applies to all complex topics)

---

## Problem

Complex technical topics overwhelm readers when presented all at once. Users with different expertise levels need different depths of information.

**Symptoms**:
- New users bounce off documentation (too complex)
- Dense paragraphs with no entry point
- No clear path from beginner to advanced
- Examples too complex for first-time users

---

## Solution

Structure content in layers, revealing complexity incrementally:

1. **Simple overview first** - What is it? Why care?
2. **Quick start** - Minimal viable example (10 minutes)
3. **Core concepts** - Key ideas with simple explanations
4. **Detailed workflow** - Step-by-step with all options
5. **Advanced topics** - Edge cases, optimization, internals

**Key Principle**: Each layer is independently useful. Reader can stop at any level and have learned something valuable.

---

## Implementation

### Structure Template

```markdown
# Topic Name

**Brief one-liner** - Core value proposition

---

## Quick Start (10 minutes)

Minimal example that works:
- 3-5 steps maximum
- No configuration options
- One happy path
- Working result

---

## What is [Topic]?

Simple explanation:
- Analogy or metaphor
- Core problem it solves
- Key benefit (one sentence)

---

## Core Concepts

Key ideas (3-6 concepts):
- Concept 1: Simple definition + example
- Concept 2: Simple definition + example
- ...

---

## Detailed Guide

Complete reference:
- All options
- Configuration
- Edge cases
- Advanced usage

---

## Reference

Technical details:
- API reference
- Configuration reference
- Troubleshooting
```

### Writing Guidelines

**Layer 1 (Quick Start)**:
- ✅ One path, no branches
- ✅ Copy-paste ready code
- ✅ Working in < 10 minutes
- ❌ No "depending on your setup" qualifiers
- ❌ No advanced options

**Layer 2 (Core Concepts)**:
- ✅ Explain "why" not just "what"
- ✅ One concept per subsection
- ✅ Concrete example for each concept
- ❌ No forward references to advanced topics
- ❌ No API details (save for reference)

**Layer 3 (Detailed Guide)**:
- ✅ All options documented
- ✅ Decision trees for choices
- ✅ Links to reference for details
- ✅ Examples for common scenarios

**Layer 4 (Reference)**:
- ✅ Complete API coverage
- ✅ Alphabetical or categorical organization
- ✅ Brief descriptions (link to guide for concepts)

---

## When to Use

✅ **Use progressive disclosure when**:
- Topic has multiple levels of complexity
- Audience spans from beginners to experts
- Quick start path exists (< 10 min viable example)
- Advanced features are optional, not required

❌ **Don't use when**:
- Topic is inherently simple (< 5 concepts)
- No quick start path (all concepts required)
- Audience is uniformly expert or beginner

---

## Validation

### First Use: BAIME Usage Guide
**Context**: Explaining BAIME framework (complex: iterations, agents, capabilities, value functions)

**Structure**:
1. What is BAIME? (1 paragraph overview)
2. Quick Start (4 steps, 10 minutes)
3. Core Concepts (6 concepts explained simply)
4. Step-by-Step Workflow (detailed 3-phase guide)
5. Specialized Agents (advanced topic)

**Evidence of Success**:
- ✅ Clear entry point for new users
- ✅ Each layer independently useful
- ✅ Complexity introduced incrementally
- ✅ No user feedback yet (baseline), but structure feels right

**Effectiveness**: Unknown (no user testing yet), but pattern emerged naturally from managing complexity

### Second Use: Iteration-1-strategy.md (This Document)
**Context**: Explaining iteration 1 strategy

**Structure**:
1. Objectives (what we're doing)
2. Strategy Decisions (priorities)
3. Execution Plan (detailed steps)
4. Expected Outcomes (results)

**Evidence of Success**:
- ✅ Quick scan gives overview (Objectives)
- ✅ Can stop after Strategy Decisions and understand plan
- ✅ Execution Plan provides full detail for implementers

**Effectiveness**: Pattern naturally applied. Confirms reusability.

---

## Variations

### Variation 1: Tutorial vs Reference
**Tutorial**: Progressive disclosure with narrative flow
**Reference**: Progressive disclosure with random access (clear sections, can jump anywhere)

### Variation 2: Depth vs Breadth
**Depth-first**: Deep dive on one topic before moving to next (better for learning)
**Breadth-first**: Overview of all topics before deep dive (better for scanning)

**Recommendation**: Breadth-first for frameworks, depth-first for specific features

---

## Related Patterns

- **Example-Driven Explanation**: Each layer should have examples (complements progressive disclosure)
- **Multi-Level Content**: Similar concept, focuses on parallel tracks (novice vs expert)
- **Visual Structure**: Helps users navigate between layers (use clear headings, TOC)

---

## Anti-Patterns

❌ **Hiding required information in advanced sections**
- If it's required, it belongs in core concepts or earlier

❌ **Making quick start too complex**
- Quick start should work in < 10 min, no exceptions

❌ **Assuming readers will read sequentially**
- Each layer should be useful independently
- Use cross-references liberally

❌ **No clear boundaries between layers**
- Use headings, whitespace, visual cues to separate layers

---

## Measurement

### Effectiveness Metrics
- **Time to first success**: Users should get working example in < 10 min
- **Completion rate**: % users who finish quick start (target: > 80%)
- **Drop-off points**: Where do users stop reading? (reveals layer effectiveness)
- **Advanced feature adoption**: % users who reach Layer 3+ (target: 20-30%)

### Quality Metrics
- **Layer independence**: Can each layer stand alone? (manual review)
- **Concept density**: Concepts per layer (target: < 7 per layer)
- **Example coverage**: Does each layer have examples? (target: 100%)

---

## Template Application Guidance

### Step 1: Identify Complexity Levels
Map your content to layers:
- What's the simplest path? (Quick Start)
- What concepts are essential? (Core Concepts)
- What options exist? (Detailed Guide)
- What's for experts only? (Reference)

### Step 2: Write Quick Start First
This validates you have a simple path:
- If quick start is > 10 steps, topic may be too complex
- If no quick start possible, reconsider structure

### Step 3: Expand Incrementally
Add layers from simple to complex:
- Core concepts next (builds on quick start)
- Detailed guide (expands core concepts)
- Reference (all remaining details)

### Step 4: Test Transitions
Verify each layer works independently:
- Can reader stop after quick start and have working knowledge?
- Does core concepts add value beyond quick start?
- Can reader skip to reference if already familiar?

---

## Status

**Validation**: ✅ 2 uses (BAIME guide, Iteration 1 strategy)
**Confidence**: High - Pattern emerged naturally twice
**Transferability**: Universal (applies to all complex documentation)
**Recommendation**: Extract to template (done in this iteration)

**Next Steps**:
- Validate in third context (different domain - API docs, troubleshooting guide, etc.)
- Gather user feedback on effectiveness
- Refine metrics based on actual usage data
