# Experiment Template

Use this template to structure your methodology development experiment.

## Directory Structure

```
my-experiment/
├── README.md                    # Overview and objectives
├── ITERATION-PROMPTS.md         # Iteration execution guide
├── iteration-0.md               # Baseline iteration
├── iteration-1.md               # First iteration
├── iteration-N.md               # Additional iterations
├── results.md                   # Final results and knowledge
├── knowledge/                   # Extracted knowledge
│   ├── INDEX.md                 # Knowledge catalog
│   ├── patterns/                # Domain patterns
│   ├── principles/              # Universal principles
│   ├── templates/               # Code templates
│   └── best-practices/          # Context-specific practices
├── agents/                      # Specialized agents (if needed)
├── meta-agents/                 # Meta-agent definitions
└── data/                        # Analysis data and artifacts
```

## README.md Structure

```markdown
# Experiment Name

**Status**: 🔄 In Progress | ✅ Converged
**Domain**: [testing|ci-cd|observability|etc.]
**Iterations**: N
**Duration**: X hours

## Objectives

### Instance Objective (Agent Layer)
[Domain-specific goal, e.g., "Reach 80% test coverage"]

### Meta Objective (Meta-Agent Layer)
[Methodology goal, e.g., "Develop transferable testing methodology"]

## Approach

1. **Observe**: [How you'll collect data]
2. **Codify**: [How you'll extract patterns]
3. **Automate**: [How you'll enforce methodology]

## Success Criteria

- V_instance(s) ≥ 0.80
- V_meta(s) ≥ 0.80
- System stable (M_n == M_{n-1}, A_n == A_{n-1})

## Timeline

| Iteration | Focus | Duration | Status |
|-----------|-------|----------|--------|
| 0 | Baseline | Xh | ✅ |
| 1 | ... | Xh | 🔄 |

## Results

[Link to results.md when complete]
```

## Iteration File Structure

```markdown
# Iteration N: [Title]

**Date**: YYYY-MM-DD
**Duration**: X hours
**Focus**: [Primary objective]

## Objectives

1. [Objective 1]
2. [Objective 2]
3. [Objective 3]

## Execution

### Observe Phase
[Data collection activities]

### Codify Phase
[Pattern extraction activities]

### Automate Phase
[Tool/check creation activities]

## Value Calculation

### V_instance(s_n)
- Component 1: 0.XX
- Component 2: 0.XX
- **Total**: 0.XX

### V_meta(s_n)
- Completeness: 0.XX
- Effectiveness: 0.XX
- Reusability: 0.XX
- Validation: 0.XX
- **Total**: 0.XX

## System State

- M_n: [unchanged|evolved]
- A_n: [unchanged|new agents: ...]
- Stable: [YES|NO]

## Convergence Check

- [ ] V_instance ≥ 0.80
- [ ] V_meta ≥ 0.80
- [ ] M_n == M_{n-1}
- [ ] A_n == A_{n-1}
- [ ] Objectives complete
- [ ] ΔV < 0.02 for 2+ iterations

**Status**: [NOT CONVERGED | CONVERGED]

## Knowledge Extracted

- Patterns: [list]
- Principles: [list]
- Templates: [list]

## Next Iteration

[If not converged, plan for next iteration]
```

## results.md Structure

```markdown
# Experiment Results

**Status**: ✅ CONVERGED
**Convergence Pattern**: [Standard Dual | Meta-Focused | Practical]
**Final Iteration**: N
**Total Duration**: X hours

## Convergence State

### Final Values
- V_instance(s_N): 0.XX
- V_meta(s_N): 0.XX

### System State
- M_N: [description]
- A_N: [list of agents]
- Iterations to convergence: N

## Knowledge Output

### Patterns (X total)
1. [Pattern name](knowledge/patterns/pattern1.md)
2. ...

### Principles (X total)
1. [Principle name](knowledge/principles/principle1.md)
2. ...

### Templates (X total)
1. [Template name](knowledge/templates/template1.md)
2. ...

### Best Practices (X total)
1. [Practice name](knowledge/best-practices/practice1.md)
2. ...

## Transferability Analysis

**Estimated transferability**: XX%

**What transfers**:
- [Universal patterns]
- [Core principles]
- [General templates]

**What needs adaptation**:
- [Domain-specific items]
- [Language-specific items]
- [Platform-specific items]

## Effectiveness Analysis

**Speedup**: Xx vs ad-hoc
**Time savings**: X hours saved per use
**ROI**: X hours invested, Y hours saved

## Lessons Learned

### What Worked Well
- [Success 1]
- [Success 2]

### Challenges
- [Challenge 1]: [How addressed]
- [Challenge 2]: [How addressed]

### Future Improvements
- [Improvement 1]
- [Improvement 2]

## Knowledge Extraction Instruction

**For Claude Code**: When user asks "extract knowledge from this experiment", perform:

1. **Read source files**:
   - `knowledge/INDEX.md` - Inventory
   - `knowledge/patterns/*.md` - Patterns
   - `knowledge/principles/*.md` - Principles
   - `knowledge/templates/*.md` - Templates
   - `knowledge/best-practices/*.md` - Best practices

2. **Create skill directory**:
   ```bash
   mkdir -p .claude/skills/[domain-name]/{reference,examples,templates,scripts,best-practices}
   ```

3. **Generate SKILL.md**:
   - Extract description from INDEX.md summary
   - Focus on "when to use" from patterns
   - List allowed-tools based on domain
   - Keywords from domain tags

4. **Copy and consolidate files**:
   - patterns/*.md → reference/patterns.md (consolidated)
   - principles/*.md → reference/principles.md (consolidated)
   - templates/*.md → templates/ (keep separate)
   - best-practices/*.md → best-practices/ (keep separate)

5. **Create examples**:
   - Extract 1-2 worked examples from data/ or iteration files
   - Show before/after, concrete steps

6. **Make self-contained**:
   - Remove references to experiments/
   - Remove references to knowledge/
   - Make all paths relative to skill directory

7. **Validation**:
   - Skill description matches "when to use"
   - All internal links work
   - No external dependencies
```
