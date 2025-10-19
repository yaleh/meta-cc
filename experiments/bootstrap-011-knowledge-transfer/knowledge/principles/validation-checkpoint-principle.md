# Principle: Validation Checkpoints

**Principle Type**: Learning Design
**Domain**: Universal (applies to all learning contexts)
**Status**: Validated
**Extracted From**: Iteration 1 - Day-1 Learning Path

---

## Statement

**Every learning stage must have clear, actionable validation criteria that enable self-assessment without mentor dependency.**

## Rationale

### Why This Matters

1. **Self-Directed Learning**: Contributors can verify progress without waiting for mentor feedback
2. **Consistency**: All learners validate against same objective criteria
3. **Confidence Building**: Clear checkpoints provide psychological wins ("I completed this!")
4. **Gap Identification**: Failed checkpoints immediately reveal knowledge gaps

### Evidence

From iteration 1 analysis:
- Day-1 path with 23 validation checkpoints enables self-paced progression
- Each checkpoint answers: "How do I know I've learned this?"
- Checkbox format provides clear visual progress tracking

## Applications

### In Learning Paths

```markdown
### Validation Checkpoint ✓

Before proceeding, verify:
- [ ] `make all` passes without errors
- [ ] Can run `./binary --help`
- [ ] Understand project structure (cmd/, internal/, docs/)
```

**Key Characteristics**:
- **Actionable**: Can be verified through action (run command, produce output)
- **Objective**: Not subjective ("understand" → clarified as "can explain in one sentence")
- **Checkbox format**: Visual progress tracking

### In Documentation

```markdown
## Prerequisites

Before reading this guide, you should:
- [ ] Have Go 1.21+ installed (`go version`)
- [ ] Be able to run `make all` successfully
- [ ] Have read the README.md
```

### In Onboarding Checklists

```markdown
## Week-1 Checklist

- [ ] Completed Day-1 path (environment setup + first PR)
- [ ] Read architecture docs (docs/architecture/)
- [ ] Can explain parser → analyzer → query flow
- [ ] Delivered first "good first issue" PR
```

## Design Guidelines

### What Makes a Good Validation Checkpoint?

1. **Clear Criteria**:
   - ✅ Good: "`make all` passes without errors"
   - ❌ Bad: "Understand the build system"

2. **Actionable**:
   - ✅ Good: "Can explain meta-cc's purpose in one sentence"
   - ❌ Bad: "Familiar with meta-cc"

3. **Self-Verifiable**:
   - ✅ Good: "Can run `./binary --help` and see command list"
   - ❌ Bad: "Knows all commands" (how to verify?)

4. **Objective**:
   - ✅ Good: "PR submitted with passing tests"
   - ❌ Bad: "Good understanding of contribution process"

### Checkpoint Patterns

**Command Execution**:
```
- [ ] `<command>` produces expected output
- [ ] `<command>` passes without errors
```

**Knowledge Check**:
```
- [ ] Can explain <concept> in one sentence
- [ ] Know where to find <resource>
- [ ] Understand <process> flow
```

**Artifact Production**:
```
- [ ] Created <file/PR/document>
- [ ] Submitted <contribution> with passing tests
- [ ] Documented <findings> in <location>
```

## Trade-offs

**Benefits**:
- ✅ Enables self-directed learning
- ✅ Provides clear progress indicators
- ✅ Reduces mentor dependency
- ✅ Builds learner confidence

**Costs**:
- ❌ Upfront effort to design criteria
- ❌ May feel tedious for experienced learners
- ❌ Requires clarity in learning objectives (forces precision)

## Anti-Patterns

### Vague Checkpoints

```markdown
❌ BAD:
- [ ] Understand the codebase

✅ GOOD:
- [ ] Can explain cmd/ vs. internal/ directory purposes
- [ ] Know where to find CLI commands (cmd/)
- [ ] Know where to find core logic (internal/)
```

### Subjective Checkpoints

```markdown
❌ BAD:
- [ ] Comfortable with git workflow

✅ GOOD:
- [ ] Can create branch, commit, and push
- [ ] Can create PR with proper description
- [ ] Can respond to PR feedback
```

### Untestable Checkpoints

```markdown
❌ BAD:
- [ ] Know everything about parsing

✅ GOOD:
- [ ] Can run parser tests (`go test ./internal/parser/`)
- [ ] Can explain what ToolCall struct represents
- [ ] Can locate parser type definitions
```

## Related Principles

- **Progressive Disclosure Principle**: Checkpoints mark transition between disclosure layers
- **Scaffolding Principle**: Checkpoints verify prerequisite knowledge before building further
- **Self-Directed Learning Principle**: Checkpoints enable autonomy

## Validation

**Evidence of Effectiveness**:
- Day-1 path with 23 checkpoints enables complete self-paced onboarding
- Each checkpoint corresponds to clear learning objective
- No external validation required (contributor can self-assess)

**Future Validation**:
- Test with real contributors
- Measure: % who self-validate correctly vs. mentor assessment
- Refine based on feedback

---

**Principle Extracted**: 2025-10-17 (Iteration 1)
**Source**: Day-1 Learning Path validation checkpoint design
**Iteration**: bootstrap-011-knowledge-transfer/iteration-1
**Status**: Validated (theoretically sound, awaiting empirical validation)
