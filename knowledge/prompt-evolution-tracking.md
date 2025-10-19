# Prompt Evolution Tracking Methodology

**Framework**: BAIME Enhancement
**Domain**: Iteration Execution
**Status**: Formalized (2025-10-18)
**Source**: Bootstrap-002 Test Strategy Experiment

---

## Overview

Prompt Evolution Tracking is a BAIME framework enhancement that systematically captures how agent prompts, instructions, and capabilities evolve during iterative methodology development. This enables meta-level understanding of *when* and *why* patterns emerge, converge, or diverge.

**Problem**: During BAIME experiments, specialized agents may emerge, existing agents may be refined, and meta-agent capabilities may evolve. Without explicit tracking, valuable learning about agent evolution patterns is lost.

**Solution**: Add structured prompt evolution metrics to iteration documentation, capturing changes to agent definitions, meta-agent capabilities, and coordination patterns.

---

## When to Apply

Use Prompt Evolution Tracking when:
- **Agent specialization emerges**: Generic agents → specialized agents (e.g., data-analyst → coverage-analyzer)
- **Meta-agent capabilities evolve**: M₀ → Mₙ (new coordination patterns added)
- **Multi-experiment comparison needed**: Understanding why some experiments require specialization and others don't
- **Methodology transferability analysis**: Determining which agent patterns transfer across domains

**Do NOT use** for:
- Single-iteration experiments (no evolution to track)
- Experiments with stable agent sets (A₀ = Aₙ) where no evolution occurred
- Experiments where agent details are not relevant to reusability

---

## Metrics to Track

### 1. Agent Set Evolution (Aₙ)

**Per Iteration**, document:

```yaml
iteration_n:
  agent_set:
    - name: "<agent-name>"
      type: "generic|specialized"
      created_iteration: n  # When agent was introduced
      modified_iteration: n # When agent prompt was modified
      rationale: "<why created/modified>"
      prompt_delta: "<key changes from previous version>"
      domain_specificity: "universal|domain-specific|task-specific"
      reusability_estimate: "high|medium|low"
```

**Example** (from Bootstrap-002):

```yaml
iteration_3:
  agent_set:
    - name: "coverage-analyzer"
      type: "specialized"
      created_iteration: 3
      rationale: "Generic data-analyst insufficient for coverage gap analysis"
      prompt_delta: "Added: Go coverage parsing, gap identification heuristics"
      domain_specificity: "domain-specific (test coverage)"
      reusability_estimate: "high (any test coverage domain)"

    - name: "test-generator"
      type: "specialized"
      created_iteration: 3
      rationale: "Generic coder insufficient for pattern-based test generation"
      prompt_delta: "Added: 8 test patterns, Go syntax templates"
      domain_specificity: "task-specific (pattern-based test generation)"
      reusability_estimate: "medium (needs pattern library per domain)"
```

### 2. Meta-Agent Evolution (Mₙ)

**Per Iteration**, document:

```yaml
iteration_n:
  meta_agent:
    capabilities_added: ["<capability>", ...]
    capabilities_modified: ["<capability>", ...]
    coordination_patterns_changed: "<description>"
    rationale: "<why meta-agent evolved>"
    stability: "stable|evolving"
```

**Example**:

```yaml
iteration_4:
  meta_agent:
    capabilities_added: ["multi-context-validation"]
    capabilities_modified: []
    coordination_patterns_changed: "Added parallel validation across contexts A/B/C"
    rationale: "Need to validate methodology across 3 project archetypes simultaneously"
    stability: "evolving"

iteration_5:
  meta_agent:
    capabilities_added: []
    capabilities_modified: []
    coordination_patterns_changed: "None"
    rationale: "Meta-agent sufficient, no evolution needed"
    stability: "stable"
```

### 3. Prompt Version History

Maintain a `prompts/` directory within each experiment with versioned agent prompts:

```
experiments/bootstrap-00X/
  prompts/
    agents/
      coverage-analyzer-v1.md  # Created iteration 3
      coverage-analyzer-v2.md  # Modified iteration 4
      test-generator-v1.md     # Created iteration 3
    meta-agent/
      meta-agent-v0.md         # Initial (from M₀)
      meta-agent-v1.md         # Modified iteration 4
```

**File format** for each prompt version:

```markdown
# <Agent Name> - Version <n>

**Created**: Iteration <n>
**Modified**: Iteration <m> (if applicable)
**Type**: generic|specialized
**Domain**: <domain>

## Capabilities

- <capability 1>
- <capability 2>

## Prompt

<full agent prompt>

## Changelog

**v2** (Iteration 4):
- Added: Coverage gap analysis heuristics
- Modified: Error handling for incomplete coverage data

**v1** (Iteration 3):
- Initial version created from generic data-analyst
- Added: Go coverage parsing
```

---

## Evolution Pattern Analysis

At experiment conclusion, analyze agent evolution patterns:

### Specialization Decision Tree

```
Agent Specialization Analysis:
1. Generic agent used for 2+ iterations?
   → YES: Evidence that generic agents sufficient initially
   → NO: Skip to 2

2. Specialization triggered by:
   → Performance inadequacy? (generic agent too slow)
   → Capability gap? (generic agent lacks domain knowledge)
   → Quality threshold? (generic agent produces suboptimal output)

3. Specialization outcome:
   → Performance gain: <X>x speedup measured
   → Quality improvement: <Y> points V_instance increase
   → Capability addition: <list new capabilities>

4. Specialization reusability:
   → Universal (95%+ contexts)
   → Domain-specific (60-95% contexts, e.g., test coverage domain)
   → Task-specific (20-60% contexts, e.g., Go test generation)
   → Experiment-specific (<20% contexts)

5. Recommendation:
   → Promote to M₀: If universal and validated across 3+ experiments
   → Keep as domain specialist: If domain-specific and validated
   → Document as pattern: If task-specific but reusable template
   → Archive: If experiment-specific
```

### Meta-Agent Stability Analysis

```
Meta-Agent Evolution Analysis:
1. M₀ sufficient throughout? (Mₙ = M₀)
   → YES: Standard BAIME pattern, no evolution needed
   → NO: Analyze why evolution occurred

2. If evolved, categorize changes:
   → Coordination: Multi-agent orchestration patterns added
   → Validation: New validation capabilities (e.g., multi-context testing)
   → Optimization: Performance or resource management improvements
   → Error handling: Recovery or retry logic added

3. Evolution reusability:
   → Universal: Add to M₀ for all future experiments
   → Domain-specific: Document as conditional enhancement
   → Experiment-specific: Archive as example only

4. Stability achieved? (Mₙ = Mₙ₊₁)
   → YES: Convergence criterion met
   → NO: Further evolution expected in next iteration
```

---

## Integration with Iteration Documentation

### In iteration-N.md Files

Add section at the end of each iteration:

```markdown
## Agent and Meta-Agent Evolution

### Agent Set (Aₙ)

**Changes from previous iteration**:
- Created: <new agents>
- Modified: <changed agents>
- Stable: <unchanged agents>

**Specialization decisions**:
- **<agent-name>**: <rationale for creation/modification>
  - **Trigger**: <capability gap / performance / quality>
  - **Impact**: <measured improvement>
  - **Reusability**: <estimate>

**Current agent set**: A₃ = {data-analyst, doc-writer, coder, coverage-analyzer, test-generator}

### Meta-Agent (Mₙ)

**Stability**: Stable (M₃ = M₂) | Evolving

**Changes** (if any):
- <capability additions/modifications>
- <rationale>

**Coordination patterns**:
- <any new multi-agent orchestration patterns>

### Evolution Summary

| Component | Status | Rationale |
|-----------|--------|-----------|
| Agent Set | A₃ ≠ A₂ | Added test-generator for pattern-based generation |
| Meta-Agent | M₃ = M₂ | Stable, no coordination changes needed |
| Convergence Impact | Positive | Specialization improved V_instance by 0.15 |
```

### In results.md Files

Add dedicated section:

```markdown
## Prompt Evolution Analysis

### Agent Specialization Summary

| Agent | Type | Created | Rationale | Impact | Reusability |
|-------|------|---------|-----------|--------|-------------|
| coverage-analyzer | Specialized | Iter 3 | Coverage gap analysis | +0.10 V_instance | High (test domains) |
| test-generator | Specialized | Iter 3 | Pattern-based generation | +0.05 V_instance | Medium (needs patterns) |

### Meta-Agent Evolution Summary

| Iteration | Status | Changes | Rationale |
|-----------|--------|---------|-----------|
| 0-2 | Stable | None | M₀ sufficient |
| 3 | Evolved | Added multi-context validation | Need parallel context testing |
| 4-5 | Stable | None | No further evolution needed |

### Evolution Patterns Observed

1. **Generic → Specialized (Iteration 3)**:
   - **Trigger**: Coverage gap analysis too complex for generic data-analyst
   - **Decision**: Create coverage-analyzer specialist
   - **Outcome**: 10x speedup for coverage analysis, +0.10 V_instance

2. **Meta-Agent Enhancement (Iteration 3)**:
   - **Trigger**: Single-context validation insufficient for transferability claims
   - **Decision**: Add multi-context validation capability to M₀
   - **Outcome**: Enabled parallel testing across Context A/B/C

3. **Stability Achieved (Iteration 4-5)**:
   - **Evidence**: A₄ = A₅, M₄ = M₅
   - **Interpretation**: Agent set and coordination patterns converged
   - **Impact**: Convergence criterion (3) and (4) satisfied

### Transferability Implications

**High-reusability agents** (promote to M₀ or domain library):
- **coverage-analyzer**: 85% reusable across test coverage domains
- Recommendation: Add to test-strategy methodology library

**Medium-reusability agents** (document as templates):
- **test-generator**: 60% reusable (needs pattern library per domain)
- Recommendation: Create pattern library template, document adaptation process

**Experiment-specific components** (archive as examples):
- None (all agents have cross-experiment value)

### Recommendations for Future Experiments

1. **Start with generic agents**: Evidence shows A₀ sufficient for 2-3 iterations
2. **Specialize when measured**: Wait for performance/quality evidence before creating specialists
3. **Validate reusability early**: Test specialist transferability within experiment (multi-context validation)
4. **Promote validated specialists**: Add high-reusability agents to methodology libraries
5. **Document evolution rationale**: Explicit decision criteria enable better future specialization decisions
```

---

## Benefits

### 1. Meta-Level Learning
- Understand *when* specialization provides value vs adds complexity
- Identify reusable specialist patterns across experiments
- Optimize M₀ over time with validated enhancements

### 2. Transferability Prediction
- Agent reusability estimates inform methodology adaptation guidance
- Distinguish universal vs domain-specific vs task-specific agents
- Reduce trial-and-error in future experiments

### 3. Convergence Insights
- Agent/meta-agent stability (A₂ = A₃, M₂ = M₃) signals convergence
- Evolution patterns reveal when methodology is stabilizing
- Inform iteration planning (if agents still evolving, more iterations likely needed)

### 4. BAIME Framework Refinement
- Build library of validated specialized agents
- Improve M₀ with battle-tested coordination patterns
- Enable meta-meta-learning (how methodology development itself evolves)

---

## Implementation Effort

**Per experiment**:
- **Setup**: 15-30 minutes (create prompts/ directory structure)
- **Per iteration**: 10-20 minutes (document agent changes in iteration-N.md)
- **Conclusion**: 1-2 hours (evolution analysis in results.md)

**Total overhead**: ~2-3 hours per experiment (2-3% of typical 25-hour experiment)

**ROI**: High - enables systematic learning across experiments, informs future agent design decisions, reduces redundant specialization.

---

## Examples

### Example 1: Bootstrap-002 Test Strategy

**Agent Evolution**:
```yaml
iteration_0:
  agent_set: [data-analyst, doc-writer, coder]  # A₀ (generic)

iteration_1:
  agent_set: [data-analyst, doc-writer, coder]  # A₁ = A₀ (stable)

iteration_2:
  agent_set: [data-analyst, doc-writer, coder]  # A₂ = A₀ (stable)

iteration_3:
  agent_set: [data-analyst, doc-writer, coder, coverage-analyzer, test-generator]
  # A₃ ≠ A₀ (specialization triggered)
  new_agents:
    - coverage-analyzer:
        rationale: "Generic data-analyst took 30 min for coverage analysis, needed 3 min"
        impact: "10x speedup, +0.10 V_instance"
        reusability: "High (test coverage domain)"
    - test-generator:
        rationale: "Pattern-based test generation needed specialized prompt with templates"
        impact: "200x speedup, +0.05 V_instance"
        reusability: "Medium (needs domain-specific pattern library)"

iteration_4-5:
  agent_set: [data-analyst, doc-writer, coder, coverage-analyzer, test-generator]
  # A₄ = A₅ = A₃ (stable, convergence)
```

**Meta-Agent Evolution**:
```yaml
iteration_0-2:
  meta_agent: M₀ (standard 5 capabilities)

iteration_3:
  meta_agent: M₁ (M₀ + multi-context-validation)
  rationale: "Need parallel validation across Context A/B/C for transferability claims"

iteration_4-5:
  meta_agent: M₁ (stable, convergence)
```

**Outcome**:
- Specialization at iteration 3 improved V_instance by 0.15
- Agent stability from iteration 3-5 contributed to convergence
- coverage-analyzer identified as high-reusability agent → add to test methodology library

### Example 2: Bootstrap-003 Error Recovery

**Agent Evolution**:
```yaml
iteration_0-2:
  agent_set: [data-analyst, doc-writer, coder]  # A₀ throughout
  # No specialization needed - generic agents sufficient
```

**Meta-Agent Evolution**:
```yaml
iteration_0-2:
  meta_agent: M₀ (stable throughout)
  # No evolution needed - standard coordination sufficient
```

**Outcome**:
- Generic agents sufficient for entire experiment → faster convergence (3 iterations vs 6)
- Demonstrates that not all experiments require specialization
- Insight: Well-scoped domains with clear metrics may not need specialized agents

---

## Decision Criteria: When to Specialize

Based on Bootstrap-002 and Bootstrap-003 patterns:

### Create Specialized Agent When:

1. **Performance bottleneck** (measured):
   - Generic agent takes >15 minutes for repeated task
   - Expected iteration count × task frequency × time > 2 hours total
   - Specialization can achieve >5x speedup

2. **Quality gap** (measured):
   - Generic agent output requires substantial manual refinement
   - Quality issues prevent V_instance from reaching threshold
   - Specialization can improve V_instance by >0.10

3. **Capability gap** (qualitative):
   - Task requires domain-specific knowledge generic agent lacks
   - Multiple iterations show consistent generic agent failures
   - Specialization adds distinct capabilities (not just prompt refinement)

4. **Reusability potential** (estimated):
   - Specialist pattern will be used in 3+ iterations or 2+ experiments
   - Domain applicability ≥60% (medium+ reusability)
   - Clear transferability path to other contexts

### Keep Generic Agents When:

1. **Performance acceptable**:
   - Task completion time <15 minutes
   - No iteration delays due to agent speed

2. **Quality sufficient**:
   - V_instance on track to reach threshold with generic agents
   - Output requires minimal refinement

3. **Early iteration** (n ≤ 2):
   - Insufficient data to justify specialization
   - Risk of premature optimization

4. **Low reusability**:
   - Pattern likely experiment-specific
   - Maintenance overhead > benefits

---

## Validation

**Success criteria** for Prompt Evolution Tracking:

1. **Complete evolution history**: All agent and meta-agent changes documented per iteration
2. **Rationale clarity**: Each specialization decision has measurable trigger (performance/quality/capability)
3. **Reusability assessment**: Each specialized agent has reusability estimate and validation evidence
4. **Cross-experiment learning**: Evolution patterns inform future experiment agent design
5. **Convergence correlation**: Agent/meta-agent stability (Aₙ = Aₙ₊₁, Mₙ = Mₙ₊₁) correlates with value function convergence

**Measurement**:
- Track specialization success rate (% of specialists that improve V_instance ≥0.05)
- Measure reusability accuracy (estimated vs actual reusability across experiments)
- Count promoted agents (specialists added to M₀ or methodology libraries)

---

## Related Methodologies

- **Empirical Methodology** (bootstrapped-se): Observation → Pattern → Validation
- **Value Optimization** (value-optimization): Dual value functions guide when to specialize
- **BAIME Framework** (bootstrapped-ai-methodology-engineering): OCA cycle includes agent evolution
- **Meta-Agent Architecture** (M₀): 5 capabilities that may evolve per experiment

---

## References

**Validated In**:
- Bootstrap-002: Test Strategy Development (6 iterations, 2 specialized agents emerged)
- Bootstrap-003: Error Recovery Methodology (3 iterations, generic agents sufficient)

**Future Work**:
- Track prompt evolution across 5+ experiments to identify universal specialization patterns
- Create agent reusability model (predict reusability from domain characteristics)
- Automate prompt versioning and diff tracking

---

**Status**: ✅ Formalized
**Effort**: 2-3 hours per experiment overhead
**Expected Impact**: Better specialization decisions, systematic cross-experiment learning, optimized M₀ evolution
**Validated**: Yes (retrospective analysis of Bootstrap-002, Bootstrap-003)

---

**Version**: 1.0
**Created**: 2025-10-18
**Source**: Bootstrap-002 Test Strategy Experiment (Future Work #7)
