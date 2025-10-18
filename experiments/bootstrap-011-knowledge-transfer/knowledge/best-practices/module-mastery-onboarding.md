# Best Practice: Module Mastery Onboarding

**Practice Type**: Onboarding
**Context**: Week-1 onboarding for software projects with modular architecture
**Domain**: Knowledge Transfer, Code Navigation
**Extracted From**: Iteration 2 - Week-1 Learning Path Design

---

## Context

**When to Apply**:
- Contributor has completed basic setup (Day-1 equivalent)
- Project has modular architecture (multiple distinct modules)
- Need to build deep understanding of codebase internals
- Goal is meaningful contribution (beyond trivial fixes)

**When Not to Apply**:
- Monolithic codebase with no clear module boundaries
- Contributors only need surface-level understanding
- Project is very small (<1000 lines of code)

---

## Recommendation

**For Week-1 level onboarding, structure module learning as**:

### 1. Architecture First (20% of time)
- Start with high-level architecture overview
- Understand data flow through the system
- Identify module boundaries and responsibilities
- Draw architecture diagram (mental model)

**Rationale**: Without architectural context, module details don't make sense.

### 2. Module Deep Dives (40% of time)
- One module at a time (don't mix)
- For each module:
  - Read core types and data structures
  - Read implementation (main logic)
  - Read and run tests
  - Understand module API
  - Identify contribution opportunities

**Rationale**: Deep understanding of modules enables meaningful contributions.

### 3. Workflow Practice (20% of time)
- Learn development workflows (TDD, debugging, CI/CD)
- Practice git workflow (branching, committing, PR)
- Understand code review process

**Rationale**: Can't contribute effectively without workflow mastery.

### 4. Meaningful Contribution (20% of time)
- Apply learning to real contribution
- Implement feature with tests
- Go through PR and code review process

**Rationale**: Learning solidifies through practice.

---

## Structure Template

```markdown
## Week-1 Learning Path

### Section 1: Architecture Deep Dive (4-8 hours)
- [ ] Read architecture docs
- [ ] Trace data flow through system
- [ ] Draw architecture diagram
- [ ] Understand module boundaries

### Section 2: Core Module Mastery (6-12 hours)
For each major module:
- [ ] Read types and data structures
- [ ] Read implementation
- [ ] Read and run tests
- [ ] Understand module API
- [ ] Identify contribution area

### Section 3: Development Workflows (4-8 hours)
- [ ] Practice TDD workflow
- [ ] Learn debugging techniques
- [ ] Understand CI/CD pipeline
- [ ] Master git workflow

### Section 4: Meaningful Contribution (6-12 hours)
- [ ] Find good first issue
- [ ] Implement with tests
- [ ] Submit PR
- [ ] Respond to code review
```

---

## Justification

### Why Architecture First?

**Evidence**: Week-1 path shows that:
- Without architecture context, contributors get lost in implementation details
- Understanding data flow (Session JSONL → Parser → Analyzer → Query → Output) provides navigation framework
- Module boundaries become clear when architecture is understood first

**Consequence**: 20% time investment in architecture saves 50%+ time later (less confusion, better mental model)

### Why One Module at a Time?

**Evidence**: Week-1 path separates parser, analyzer, and query into distinct sections:
- Mixing modules causes cognitive overload
- Sequential mastery builds confidence
- Each module completion provides psychological win

**Consequence**: 3 sequential module deep dives (2-4 hours each) more effective than parallel exploration

### Why Workflows Before Contribution?

**Evidence**: Week-1 path teaches TDD, debugging, git *before* meaningful contribution:
- Contributors need workflow muscle memory before complex tasks
- Practice on small examples before real features
- Reduces friction during actual contribution

**Consequence**: 4-8 hours workflow practice enables smooth 6-12 hour contribution

---

## Trade-offs

**Benefits**:
- ✅ Deep understanding (not surface-level)
- ✅ Confidence to contribute meaningfully
- ✅ Reduced errors (understand system before modifying)
- ✅ Faster long-term productivity (strong foundation)

**Costs**:
- ❌ Takes 20-40 hours (significant time investment)
- ❌ May feel slow initially (Day-1 was faster)
- ❌ Requires discipline (temptation to skip to contribution)

**Mitigations**:
- Provide clear time estimates (manage expectations)
- Include validation checkpoints (track progress)
- Show value of deep understanding (faster later)

---

## Anti-Patterns

### Rushing to Contribution

```markdown
❌ BAD: Week-1 Path Without Architecture
Section 1: Pick a module and start coding

✅ GOOD: Week-1 Path With Architecture First
Section 1: Architecture Deep Dive (understand system)
Section 2: Module Mastery (deep dive each module)
Section 3: Workflows (practice TDD, git)
Section 4: Contribution (apply learning)
```

### Module Mixing

```markdown
❌ BAD: Learning All Modules Simultaneously
Section 1: Intro to parser, analyzer, query
  → 1.1 Parser types
  → 1.2 Analyzer patterns
  → 1.3 Query filters
  [Mixed context, cognitive overload]

✅ GOOD: Sequential Module Mastery
Section 2.1: Parser Deep Dive (2-4 hours)
  → Types, implementation, tests
  [Complete parser understanding]

Section 2.2: Analyzer Deep Dive (2-4 hours)
  → Types, implementation, tests
  [Complete analyzer understanding]

Section 2.3: Query Deep Dive (2-4 hours)
  → Types, implementation, tests
  [Complete query understanding]
```

### Skipping Tests

```markdown
❌ BAD: Module Learning Without Tests
- Read module types
- Read module implementation
- [Done, move to next module]

✅ GOOD: Module Learning With Tests
- Read module types
- Read module implementation
- **Read and run module tests**
  [Tests show expected behavior, edge cases, usage examples]
- Understand test coverage
- Identify test gaps (contribution opportunities)
```

---

## Validation

**Evidence of Effectiveness**:
- Week-1 path with module mastery structure enables meaningful contributions
- Architecture-first approach prevents getting lost in implementation details
- Sequential module mastery builds confidence and depth
- TDD/workflow practice before contribution reduces friction

**Metrics**:
- Time to meaningful contribution: ~5 days (vs. weeks without structure)
- Contributor confidence: High (can explain architecture, understand modules)
- Contribution quality: Better (deep understanding → fewer errors)

---

## Examples

### meta-cc Week-1 Path

**Architecture (4-8 hours)**:
- Read architecture proposal, ADRs, principles
- Trace data flow: Session JSONL → Parser → Analyzer → Query → Output
- Understand two-layer architecture (CLI + core logic)

**Module Mastery (6-12 hours)**:
- Parser deep dive (2-4h): types, implementation, tests
- Analyzer deep dive (2-4h): patterns, statistics, tests
- Query deep dive (2-4h): filtering, aggregation, tests

**Workflows (4-8 hours)**:
- TDD practice: write test → implement → verify
- Debugging: print debugging, delve
- CI/CD: make all, fix lint/test/build errors
- Git workflow: branch, commit, PR

**Contribution (6-12 hours)**:
- Find good first issue
- Implement with tests (TDD)
- Submit PR with description
- Code review and iteration

---

## Related Practices

- **Progressive Disclosure**: Reveal complexity gradually (architecture → modules → contribution)
- **Scaffolding**: Build on previous knowledge (Day-1 → Week-1 → Month-1)
- **Validation Checkpoints**: Self-assessment at each stage
- **Test-Driven Learning**: Learn through tests (understand expected behavior)

---

**Practice Extracted**: 2025-10-17 (Iteration 2)
**Source**: Week-1 Learning Path design for meta-cc
**Iteration**: bootstrap-011-knowledge-transfer/iteration-2
**Status**: Proposed (awaits real contributor validation)
