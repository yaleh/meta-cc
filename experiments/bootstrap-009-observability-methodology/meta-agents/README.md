# Meta-Agent Capability Files

This directory contains the individual capability files for the Meta-Agent M₀ used in the bootstrap-003-error-recovery experiment.

---

## Structure

Instead of a single `meta-agent-m0.md` file, each Meta-Agent capability is documented in its own file:

```
meta-agents/
├── observe.md       # M.observe - Data collection and pattern recognition
├── plan.md          # M.plan - Strategy formulation and agent selection
├── execute.md       # M.execute - Agent coordination and task execution
├── reflect.md       # M.reflect - Evaluation and learning
└── evolve.md        # M.evolve - System adaptation and growth
```

---

## Why Separate Files?

**Benefits**:
1. **Modularity**: Each capability is independently documented and updatable
2. **Clarity**: Easier to focus on one capability at a time
3. **Evolution**: New capabilities can be added without modifying existing files
4. **Reusability**: Capabilities can be referenced independently
5. **Maintainability**: Changes to one capability don't affect others

**Usage Protocol**:
- **Before iteration**: Read ALL capability files to understand M₀
- **During capability use**: Read specific capability file before using it
- **After evolution**: Create new capability file (if M evolves)

---

## Capability Summary

### observe.md (9.2 KB)
**Purpose**: Gather error data and identify patterns

**Key Sections**:
- Data collection strategies (error history, tool-specific errors, patterns)
- Pattern recognition (frequency, category, tool, impact patterns)
- Gap identification (detection, diagnosis, recovery, prevention gaps)
- Observation protocol (step-by-step process)
- Output format and metrics

### plan.md (4.6 KB)
**Purpose**: Formulate strategy and select agents

**Key Sections**:
- Strategic planning process (assess state, define goals, prioritize)
- Agent selection strategy (decision tree, criteria)
- Planning output format
- Integration with other capabilities

### execute.md (4.8 KB)
**Purpose**: Coordinate agents to perform work

**Key Sections**:
- Execution protocol (preparation, invocation, monitoring)
- Agent coordination (handoff, agent-specific patterns)
- Output collection
- Integration with other capabilities

### reflect.md (6.1 KB)
**Purpose**: Evaluate quality and calculate value

**Key Sections**:
- Value calculation (V(s) formula and components)
- Quality evaluation criteria
- Gap identification
- Convergence checking (formal criteria)
- Learning and insights
- Reflection output format

### evolve.md (9.2 KB)
**Purpose**: Adapt agent set and Meta-Agent capabilities

**Key Sections**:
- Agent evolution (when, how, common specialized agents)
- Meta-Agent evolution (when, how, examples)
- Evolution decision framework
- Evolution tracking
- Anti-patterns to avoid
- Success criteria and philosophy

---

## Reading Order

### For First Iteration (Iteration 0)

1. Read **ALL files** first to understand M₀ completely:
   ```
   1. observe.md  - Understand data collection approach
   2. plan.md     - Understand strategy formulation
   3. execute.md  - Understand agent coordination
   4. reflect.md  - Understand evaluation process
   5. evolve.md   - Understand evolution criteria
   ```

2. Then during iteration, read **specific file** before using capability:
   ```
   - Before observing → Read observe.md
   - Before planning → Read plan.md
   - Before executing → Read execute.md
   - Before reflecting → Read reflect.md
   - When considering evolution → Read evolve.md
   ```

### For Subsequent Iterations

1. Re-read ALL files (capabilities may have been updated)
2. Follow same protocol: read specific file before using capability

---

## Evolution

### If Meta-Agent Evolves (M_N ≠ M_{N-1})

**Scenario 1**: New capability added
- Create new file: `meta-agents/{new-capability}.md`
- Document complete capability specification
- Update other files if they reference it
- Example: `triage_by_severity.md` if severity-based error triage needed

**Scenario 2**: Existing capability enhanced
- Update the relevant capability file
- Document what changed and why
- Increment version number in file
- Example: `observe.md` updated with real-time error monitoring

### Expected Evolution

Based on bootstrap-001-doc-methodology:
- **Most likely**: M₀ remains stable (these 5 capabilities sufficient)
- **Possible**: 1 new capability added if novel coordination pattern emerges
- **Rare**: Multiple capabilities added (indicates M₀ was insufficient)

**Historical precedent**:
- bootstrap-001: M₀ remained unchanged through all 3 iterations
- Only 5 core capabilities needed for convergence

---

## File Sizes

```
observe.md:  9.2 KB (most detailed - data collection is complex)
plan.md:     4.6 KB (focused - strategy is clear)
execute.md:  4.8 KB (structured - coordination is systematic)
reflect.md:  6.1 KB (detailed - evaluation needs rigor)
evolve.md:   9.2 KB (comprehensive - evolution requires careful justification)
```

**Total**: ~34 KB for complete M₀ specification

---

## Integration

These capability files integrate with:

**Agent prompt files** (in `../agents/`):
- Each agent has its own prompt file
- Agents are invoked by M.execute capability
- New agents created by M.evolve capability

**Iteration execution** (via `../ITERATION-PROMPTS.md`):
- Provides detailed execution instructions
- References these capability files
- Enforces reading protocol

**Experiment plan** (via `../plan.md`):
- Defines overall experiment framework
- These capabilities implement the framework
- Value function optimization through capabilities

---

## Quick Reference

### Capability Dependencies

```
observe → plan → execute → reflect → evolve
   ↑                           ↓
   └───────────────────────────┘
   (feedback loop for next iteration)
```

### When to Read

```
Starting iteration    → Read ALL files
Before observing      → Read observe.md
Before planning       → Read plan.md
Before executing      → Read execute.md
Before reflecting     → Read reflect.md
Considering evolution → Read evolve.md
```

### Common Patterns

```yaml
typical_iteration:
  1. Read all capability files
  2. observe: Collect error data
  3. plan: Define iteration goal, select agents
  4. execute: Coordinate agents to do work
  5. reflect: Calculate V(s), check convergence
  6. evolve: Create agents if needed (rare: add M capability)
  7. Document iteration results
  8. If not converged: goto step 1 (next iteration)
```

---

## Version History

**Version 0.0** (2025-10-14):
- Initial capability files created
- 5 core capabilities documented
- Error recovery domain specialization

**Future versions**:
- Will be documented here as capabilities evolve
- Each file tracks its own version history

---

## References

**Experiment Files**:
- [../plan.md](../plan.md) - Experiment design
- [../ITERATION-PROMPTS.md](../ITERATION-PROMPTS.md) - Execution guide
- [../README.md](../README.md) - Experiment overview

**Methodology Frameworks**:
- [Empirical Methodology Development](../../../docs/methodology/empirical-methodology-development.md)
- [Bootstrapped Software Engineering](../../../docs/methodology/bootstrapped-software-engineering.md)
- [Value Space Optimization](../../../docs/methodology/value-space-optimization.md)

---

**Created**: 2025-10-14
**Experiment**: bootstrap-003-error-recovery
**Meta-Agent Version**: M₀ (v0.0)
**Total Capabilities**: 5 (observe, plan, execute, reflect, evolve)
