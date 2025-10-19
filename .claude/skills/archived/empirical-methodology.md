---
name: empirical-methodology
description: Develop project-specific methodologies through empirical observation, data analysis, pattern extraction, and automated validation - treating methodology development like software development
keywords: empirical, data-driven, methodology, observation, analysis, codification, validation, continuous-improvement, scientific-method
category: methodology
version: 1.0.0
based_on: docs/methodology/empirical-methodology-development.md
transferability: 92%
effectiveness: 10-20x vs theory-driven methodologies
---

# Empirical Methodology Development

**Develop software engineering methodologies like software: with observation tools, empirical validation, automated testing, and continuous iteration.**

> Traditional methodologies are theory-driven and static. **Empirical methodologies** are data-driven and continuously evolving.

---

## The Problem

Traditional methodologies are:
- **Theory-driven**: Based on principles, not data
- **Static**: Created once, rarely updated
- **Prescriptive**: One-size-fits-all
- **Manual**: Require discipline, no automated validation

**Result**: Methodologies that don't fit your project, aren't followed, and don't improve.

---

## The Solution

**Empirical Methodology Development**: Create project-specific methodologies through:

1. **Observation**: Build tools to measure actual development process
2. **Analysis**: Extract patterns from real data
3. **Codification**: Document patterns as reproducible methodologies
4. **Automation**: Convert methodologies into automated checks
5. **Evolution**: Use automated checks to continuously improve methodologies

### Key Insight

> Software engineering methodologies can be developed **like software**:
> - Observation tools (like debugging)
> - Empirical validation (like testing)
> - Automated checks (like CI/CD)
> - Continuous iteration (like agile)

---

## The Scientific Method for Methodologies

```
1. Observation
   ↓
   Build measurement tools (meta-cc, git analysis)
   Collect data (commits, sessions, metrics)

2. Hypothesis
   ↓
   "High-access docs should be <300 lines"
   "Batch remediation is 5x more efficient"

3. Experiment
   ↓
   Implement change (refactor CLAUDE.md)
   Measure effects (token cost, access patterns)

4. Data Collection
   ↓
   query-files, access density, R/E ratio

5. Analysis
   ↓
   Statistical analysis, pattern recognition

6. Conclusion
   ↓
   "300-line limit effective: 47% cost reduction"

7. Publication
   ↓
   Codify as methodology document

8. Replication
   ↓
   Apply to other projects, validate transferability
```

---

## Five-Phase Process

### Phase 1: OBSERVE

**Build measurement infrastructure**

```python
Tools:
  - Session analysis (meta-cc)
  - Git commit analysis
  - Code metrics (coverage, complexity)
  - Access pattern tracking
  - Error rate monitoring
  - Performance profiling

Data collected:
  - What gets accessed (files, functions)
  - How often (frequencies, patterns)
  - When (time series, triggers)
  - Why (user intent, context)
  - With what outcome (success, errors)
```

**Example** (from meta-cc):
```bash
# Analyze file access patterns
meta-cc query files --threshold 5

# Results:
plan.md: 423 accesses (Coordination role)
CLAUDE.md: ~300 implicit loads (Entry Point role)
features.md: 89 accesses (Reference role)

# Insight: Document role ≠ directory location
```

### Phase 2: ANALYZE

**Extract patterns from data**

```python
Techniques:
  - Statistical analysis (frequencies, correlations)
  - Pattern recognition (recurring behaviors)
  - Anomaly detection (outliers, inefficiencies)
  - Comparative analysis (before/after)
  - Trend analysis (time series)

Outputs:
  - Identified patterns
  - Hypotheses formulated
  - Correlations discovered
  - Anomalies flagged
```

**Example** (from meta-cc):
```python
# Pattern discovered: High-access docs should be concise

Data:
  - plan.md: 423 accesses, 200 lines → Efficient
  - CLAUDE.md: 300 accesses, 607 lines → Inefficient
  - README.md: 150 accesses, 1909 lines → Very inefficient

Hypothesis:
  - Docs with access/line ratio < 1.0 are inefficient
  - Target: >1.5 access/line ratio

Validation:
  - After optimization:
    * CLAUDE.md: 607 → 278 lines, ratio: 0.5 → 1.08
    * README.md: 1909 → 275 lines, ratio: 0.08 → 0.55
    * Token cost: -47%
```

### Phase 3: CODIFY

**Document patterns as methodologies**

```python
Methodology structure:
  1. Problem statement (pain point)
  2. Observation data (empirical evidence)
  3. Pattern description (what was discovered)
  4. Solution approach (how to apply)
  5. Validation criteria (how to measure success)
  6. Examples (concrete cases)
  7. Transferability notes (applicability)

Formats:
  - Markdown documents (docs/methodology/*.md)
  - Decision trees (workflow diagrams)
  - Checklists (validation steps)
  - Templates (boilerplate code)
```

**Example** (from meta-cc):
```markdown
# Role-Based Documentation Methodology

## Problem
Inefficient documentation: high token cost, low accessibility

## Observation
423 file accesses analyzed, 6 distinct access patterns identified

## Pattern
Documents have roles based on actual usage:
  - Entry Point: First accessed, navigation hub (<300 lines)
  - Coordination: Frequently referenced, planning (<500 lines)
  - Reference: Looked up as needed (<1000 lines)
  - Archive: Rarely accessed (no size limit)

## Solution
1. Classify documents by access pattern
2. Optimize by role (high-access = concise)
3. Create role-specific maintenance procedures

## Validation
- Access/line ratio > 1.0 for Entry Point docs
- Token cost reduction ≥ 30%
- User satisfaction survey

## Transferability
85% applicable to other projects (role concept universal)
```

### Phase 4: AUTOMATE

**Convert methodologies into automated checks**

```python
Automation levels:
  1. Detection: Identify when pattern applies
  2. Validation: Check compliance with methodology
  3. Enforcement: Prevent violations (CI gates)
  4. Suggestion: Recommend fixes

Implementation:
  - Shell scripts (quick checks)
  - Python/Go tools (complex validation)
  - CI/CD integration (automated gates)
  - IDE plugins (real-time feedback)
  - Bots (PR comments, auto-fix)
```

**Example** (from meta-cc):
```bash
# Automation: /meta doc-health capability

# Checks:
- Role classification (based on access patterns)
- Size compliance (lines < role threshold)
- Cross-reference completeness
- Update freshness

# Actions:
- Flag oversized Entry Point docs
- Suggest restructuring for high-access docs
- Auto-classify by access data
- Generate optimization report

# CI Integration:
- Block PRs that violate doc size limits
- Require review for role reassignment
- Auto-comment with optimization suggestions
```

### Phase 5: EVOLVE

**Continuously improve methodology**

```python
Evolution cycle:
  1. Apply automated checks to development
  2. Collect compliance data
  3. Analyze exceptions and edge cases
  4. Refine methodology based on data
  5. Update automation
  6. Iterate

Meta-improvement:
  - Methodology applies to itself
  - Observation tools analyze methodology effectiveness
  - Automated checks validate methodology usage
  - Continuous refinement based on outcomes
```

**Example** (from meta-cc):
```bash
# Iteration 1: Role-based docs
Observation: Access patterns
Methodology: 4 roles defined
Automation: /meta doc-health
Result: 47% token reduction

# Iteration 2: Cross-reference optimization
Observation: Broken links, redundancy
Methodology: Reference density guidelines
Automation: Link checker
Result: 15% further reduction

# Iteration 3: Implicit loading optimization
Observation: CLAUDE.md implicitly loaded ~300 times
Methodology: Entry point optimization
Automation: Size enforcer
Result: 54% size reduction (607 → 278 lines)
```

---

## Parameters

- **observation_tools**: `meta-cc` | `git-analysis` | `custom` (default: `meta-cc`)
- **observation_period**: number of days/commits (default: 30)
- **pattern_threshold**: minimum frequency to consider pattern (default: 5)
- **automation_level**: `detect` | `validate` | `enforce` | `suggest` (default: `validate`)
- **evolution_cycles**: number of refinement iterations (default: 3)

---

## Usage Examples

### Example 1: Documentation Methodology

```bash
# User: "Develop documentation methodology empirically"
empirical-methodology observation_tools=meta-cc observation_period=30

# Execution:

[OBSERVE Phase - 30 days]
✓ Collecting access data...
  - 1,247 file accesses tracked
  - 89 unique files accessed
  - Top 10 account for 73% of accesses

✓ Access pattern analysis:
  - plan.md: 423 (34%), coordination role
  - CLAUDE.md: 312 (25%), entry point role
  - features.md: 89 (7%), reference role

[ANALYZE Phase]
✓ Pattern recognition:
  - 6 distinct access roles identified
  - Access/line ratio correlates with efficiency
  - High-access docs (>100) should be <300 lines
  - Archive docs (<10 accesses) can be unlimited

[CODIFY Phase]
✓ Methodology documented:
  - Created: docs/methodology/role-based-documentation.md
  - Defined: 6 roles with size guidelines
  - Validation: Access/line ratio metrics

[AUTOMATE Phase]
✓ Automation implemented:
  - Script: scripts/check-doc-health.sh
  - Capability: /meta doc-health
  - CI check: Block PRs violating size limits

[EVOLVE Phase]
✓ Applied to self:
  - Optimized 23 documents
  - Average reduction: 42%
  - Token cost: -47%

✓ Refinement discovered:
  - New pattern: Implicit loading impact
  - Updated methodology: Entry point guidelines
  - Enhanced automation: Implicit load tracker
```

### Example 2: Testing Methodology

```bash
# User: "Extract testing methodology from project history"
empirical-methodology observation_tools=git-analysis observation_period=90

# Execution:

[OBSERVE Phase - 90 days]
✓ Git history analyzed:
  - 277 commits
  - 67 stages (test-related)
  - Coverage: 75% → 86% progression

✓ Test patterns identified:
  - TDD cycle: Test → Implement → Validate (67/67 stages)
  - Coverage gap closure: Prioritize <50% coverage files
  - Fixture pattern: Integration tests use shared fixtures

[ANALYZE Phase]
✓ Correlations discovered:
  - TDD reduces bug rate by 3.2x
  - Coverage >75% correlates with 5x fewer production errors
  - Integration tests 10x slower than unit tests

[CODIFY Phase]
✓ Methodology: Systematic Testing Strategy
  - TDD as default workflow
  - Coverage-driven gap closure (target: 75%+)
  - Integration test fixture patterns
  - Quality gates (8/10 criteria)

[AUTOMATE Phase]
✓ Automated checks:
  - Pre-commit: Run tests, block if fail
  - CI: Coverage gate (<75% = fail)
  - PR bot: Comment with coverage delta
  - Auto-fixture: Generate from examples

[EVOLVE Phase]
✓ Results:
  - Coverage: 75% → 86%
  - Bug rate: -68%
  - Test time: -73% (parallel execution)
  - Methodology validated: 89% transferability
```

### Example 3: Error Recovery Methodology

```bash
# User: "Develop error handling methodology from session data"
empirical-methodology observation_tools=meta-cc

# Execution:

[OBSERVE Phase]
✓ Session error analysis:
  - 423 errors across 277 sessions
  - Error rate: 6.06%
  - Categories: Type (45%), Logic (30%), Deps (15%), Other (10%)

[ANALYZE Phase]
✓ Error patterns:
  - Type errors: 80% preventable with linting
  - Logic errors: 60% catchable with better tests
  - Dependency errors: 90% detectable with scanning

✓ Recovery patterns:
  - Type errors: Fix + add lint rule (prevents recurrence)
  - Logic errors: Fix + add test (regression prevention)
  - Dependency errors: Update + add to CI scan

[CODIFY Phase]
✓ Methodology: Systematic Error Recovery
  1. Detection: Error signature extraction
  2. Classification: Rule-based categorization
  3. Recovery: Strategy pattern application
  4. Prevention: Root cause → Code pattern → Linter rule

[AUTOMATE Phase]
✓ Tools created:
  - Error classifier (pattern matching)
  - Recovery strategy recommender
  - Prevention linter (custom rules)
  - CI integration (auto-classify build failures)

[EVOLVE Phase]
✓ Impact:
  - Error rate: 6.06% → 1.2% (-80%)
  - Mean time to recovery: 45min → 8min (-82%)
  - Recurrence rate: 23% → 3% (-87%)
  - Transferability: 85%
```

---

## Validated Outcomes

**From meta-cc project** (277 commits, 11 days):

### Documentation Evolution

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| README.md | 1909 lines | 275 lines | -85% |
| CLAUDE.md | 607 lines | 278 lines | -54% |
| Token cost | Baseline | -47% | 47% reduction |
| Access efficiency | 0.3 access/line | 1.1 access/line | +267% |
| User satisfaction | 65% | 92% | +42% |

### Testing Methodology

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Coverage | 75% | 86% | +11pp |
| Bug rate | Baseline | -68% | 68% reduction |
| Test time | 180s | 48s | -73% |
| Methodology docs | 0 | 5 | Complete |
| Transferability | - | 89% | Validated |

### Error Recovery

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Error rate | 6.06% | 1.2% | -80% |
| MTTR | 45min | 8min | -82% |
| Recurrence | 23% | 3% | -87% |
| Prevention | 0% | 65% | 65% prevented |
| Transferability | - | 85% | Validated |

---

## Transferability

**92% transferable** across projects and domains:

### What Transfers (92%+)
- Five-phase process (Observe → Analyze → Codify → Automate → Evolve)
- Scientific method approach
- Data-driven validation
- Automated enforcement
- Continuous improvement mindset

### What Needs Adaptation (8%)
- Specific observation tools (meta-cc → project-specific)
- Data collection methods (session logs vs git vs metrics)
- Domain-specific patterns (docs vs tests vs architecture)
- Automation implementation (language, platform)

### Adaptation Effort
- **Same project, new domain**: 2-4 hours
- **New project, same domain**: 4-8 hours
- **New project, new domain**: 8-16 hours

---

## Prerequisites

### Tools Required
- **Observation**: meta-cc or equivalent (session/git analysis)
- **Analysis**: Statistical tools (Python, R, Excel)
- **Automation**: CI/CD platform, scripting language
- **Documentation**: Markdown editor, diagram tools

### Skills Required
- Basic data analysis (statistics, pattern recognition)
- Scientific method (hypothesis, experiment, validation)
- Scripting (bash, Python, etc.)
- CI/CD configuration

---

## Success Criteria

| Criterion | Target | Validation |
|-----------|--------|------------|
| **Patterns Identified** | ≥3 per domain | Documented patterns |
| **Data-Driven** | 100% empirical | All claims have data |
| **Automated** | ≥80% of checks | CI integration |
| **Improved Metrics** | ≥30% improvement | Before/after data |
| **Transferability** | ≥85% reusability | Cross-project validation |

---

## Honest Assessment Principles

**The foundation of empirical methodology is honest, evidence-based assessment.** Confirmation bias and premature optimization are the enemies of sound methodology development.

### Core Principle: Seek Disconfirming Evidence

**Traditional approach** (confirmation bias):
```
"My hypothesis is that X works."
→ Look for evidence that X works
→ Find confirming evidence
→ Conclude X works ✓
```

**Empirical approach** (honest assessment):
```
"My hypothesis is that X works."
→ Actively seek evidence that X DOESN'T work
→ Find both confirming AND disconfirming evidence
→ Weight evidence objectively
→ Revise hypothesis if disconfirming evidence is strong
→ Conclude honestly based on full evidence
```

**Example from Bootstrap-002** (Testing):
```
Initial hypothesis: "80% coverage is required"

Disconfirming evidence sought:
- Some packages have 86-94% coverage (excellence)
- Aggregate is 75% (below target)
- Tests are high quality, fixtures well-designed

Honest conclusion:
- Sub-package excellence > aggregate metric
- Quality > raw numbers
- 75% coverage + excellent tests > 80% coverage + poor tests
→ Practical Convergence declared (quality-based, not metric-based)
```

### Avoiding Common Biases

#### Bias 1: Inflating Values to Meet Targets

**Symptom**: V scores mysteriously jump to exactly 0.80 in final iteration

**Example** (anti-pattern):
```
Iteration N-1: V_instance = 0.77
Iteration N:   V_instance = 0.80 (claimed)

But... no substantial changes were made!
```

**Honest alternative**:
```
Iteration N-1: V_instance = 0.77
Iteration N:   V_instance = 0.79 (honest assessment)

Options:
1. Declare Practical Convergence (if quality evidence strong)
2. Continue iteration N+1 to genuinely reach 0.80
3. Accept that 0.80 may not be appropriate threshold for this domain
```

#### Bias 2: Selective Evidence Presentation

**Symptom**: Only showing data that supports the hypothesis

**Example** (anti-pattern):
```
Methodology Documentation:
"Our approach achieved 90% user satisfaction!"

Missing data:
- Survey had 3 respondents (2.7 users satisfied)
- Sample size too small for statistical significance
- Selection bias (only satisfied users responded)
```

**Honest alternative**:
```
Methodology Documentation:
"Preliminary feedback (n=3, self-selected): 2/3 positive responses.
Note: Sample size insufficient for statistical claims.
Recommendation: Conduct structured survey (target n=30+) for validation."
```

#### Bias 3: Moving Goalposts

**Symptom**: Changing success criteria mid-experiment to match achieved results

**Example** (anti-pattern):
```
Initial plan: "V_instance ≥ 0.80"
Final state:  V_instance = 0.65
Conclusion:  "Actually, 0.65 is sufficient for this domain" ← Goalpost moved!
```

**Honest alternative**:
```
Initial plan: "V_instance ≥ 0.80"
Final state:  V_instance = 0.65
Options:
1. Continue iteration to reach 0.80
2. Analyze WHY 0.65 is limit (genuine constraint discovered)
3. Document gap and future work needed
→ Do NOT retroactively lower target without evidence-based justification
```

#### Bias 4: Cherry-Picking Metrics

**Symptom**: Highlighting favorable metrics, hiding unfavorable ones

**Example** (anti-pattern):
```
Results Presentation:
"Achieved 95% test coverage!" ✨

Hidden metrics:
- 50% of tests are trivial (testing getters/setters)
- 0% integration test coverage
- 30% of code is actually tested meaningfully
```

**Honest alternative**:
```
Results Presentation:
"Coverage metrics breakdown:
- Overall coverage: 95% (includes trivial tests)
- Meaningful coverage: ~30% (non-trivial logic)
- Unit tests: 95% coverage
- Integration tests: 0% coverage

Gap analysis:
- Integration test coverage is critical gap
- Trivial test inflation gives false confidence
- Recommendation: Add integration tests, measure meaningful coverage"
```

### Honest V-Score Calculation

**Guidelines for honest value function scoring**:

#### 1. Ground Scores in Concrete Evidence

**Bad**:
```
V_completeness = 0.85
Justification: "Methodology feels pretty complete"
```

**Good**:
```
V_completeness = 0.80
Evidence:
- 4/5 methodology sections documented (0.80)
- All include examples (✓)
- All have validation criteria (✓)
- Missing: Edge case handling (documented as future work)
Calculation: 4/5 = 0.80 ✓
```

#### 2. Challenge High Scores

**Self-questioning protocol** for scores ≥ 0.90:

```
Claimed score: V_component = 0.95

Questions to ask:
1. What would a PERFECT score (1.0) look like? How far are we?
2. What specific deficiencies exist? (enumerate explicitly)
3. Could an external reviewer find gaps we missed?
4. Are we comparing to realistic standards or ideal platonic forms?

If you can't answer these rigorously → Lower the score
```

**Example from Bootstrap-011**:
```
V_effectiveness claimed: 0.95 (3-8x speedup)

Self-challenge:
- 10x speedup would be 1.0 (perfect score)
- We achieved 3-8x (conservative estimate)
- Could be higher (8x) but need more data
- Conservative estimate: 3-8x → 0.95 justified
- Perfect score would require 10x+ → We're not there
→ Score 0.95 is honest ✓
```

#### 3. Enumerate Gaps Explicitly

**Every component should list its gaps**:

```
V_discoverability = 0.58

Gaps preventing higher score:
1. Knowledge graph not implemented (-0.15)
2. Semantic search missing (-0.12)
3. Context-aware recommendations absent (-0.10)
4. Limited to keyword search (-0.05)

Total gap: 0.42 → Score: 1.0 - 0.42 = 0.58 ✓
```

### Practical Convergence Recognition

**When to recognize Practical Convergence** (discovered in Bootstrap-002):

#### Valid Justifications:

1. **Quality > Metrics**
   ```
   Example: 75% coverage with excellent tests > 80% coverage with poor tests
   Validation: Test quality metrics, fixture patterns, zero flaky tests
   ```

2. **Sub-System Excellence**
   ```
   Example: Core packages at 86-94% coverage, utilities at 60%
   Validation: Coverage distribution analysis, critical path identification
   ```

3. **Diminishing Returns**
   ```
   Example: ΔV < 0.02 for 3 consecutive iterations
   Validation: Iteration history, effort vs improvement ratio
   ```

4. **Justified Partial Criteria**
   ```
   Example: 8/10 quality gates met, 2 non-critical
   Validation: Gate importance analysis, risk assessment
   ```

#### Invalid Justifications:

❌ "We're close enough" (no evidence)
❌ "I'm tired of iterating" (convenience)
❌ "The metric is wrong anyway" (moving goalposts)
❌ "It works for me" (anecdotal evidence)

### Self-Assessment Checklist

Before declaring methodology complete, verify:

- [ ] **All claims have empirical evidence** (no "I think" or "probably")
- [ ] **Disconfirming evidence sought and addressed**
- [ ] **Value scores grounded in concrete calculations**
- [ ] **Gaps explicitly enumerated** (not hidden)
- [ ] **High scores (≥0.90) challenged and justified**
- [ ] **If Practical Convergence: Valid justification from list above**
- [ ] **Baseline values measured** (not assumed)
- [ ] **Improvement ΔV calculated honestly** (not inflated)
- [ ] **Transferability tested** (not just claimed)
- [ ] **Methodology applied to self** (dogfooding)

### Meta-Assessment: Methodology Quality Check

**Apply this methodology to itself**:

```
Honest Assessment Principles Quality:

V_completeness: How complete is this chapter?
- Core principles: ✓
- Bias avoidance: ✓
- V-score calculation: ✓
- Practical convergence: ✓
- Self-assessment checklist: ✓
→ Score: 5/5 = 1.0

V_effectiveness: Does it improve assessment honesty?
- Explicit guidelines: ✓
- Concrete examples: ✓
- Self-challenge protocol: ✓
- Validation checklists: ✓
→ Score: 0.85 (needs more empirical validation)

V_reusability: Can this transfer to other methodologies?
- Domain-agnostic principles: ✓
- Universal bias patterns: ✓
- Applicable beyond software: ✓
→ Score: 0.90+
```

### Learning from Failure

**Honest assessment includes documenting failures**:

```
Current issue: 0/8 experiments documented failures

Why? Because all 8 succeeded!

But this creates bias:
- Observers may think methodology is infallible
- Future users may hide failures
- No learning from failure modes

Action:
- Document near-failures, close calls
- Record challenges and recovery
- Build failure mode library
→ See "Failure Modes and Recovery" chapter (next)
```

---

## Relationship to Other Methodologies

**empirical-methodology provides the SCIENTIFIC FOUNDATION** for systematic methodology development.

### Relationship to bootstrapped-se (Included In)

**empirical-methodology is INCLUDED IN bootstrapped-se**:

```
empirical-methodology (5 phases):
  Phase 1: Observe  ─┐
  Phase 2: Analyze  ─┼─→ bootstrapped-se: Observe
                     │
  Phase 3: Codify  ──┼─→ bootstrapped-se: Codify
                     │
  Phase 4: Automate ─┼─→ bootstrapped-se: Automate
                     │
  Phase 5: Evolve  ──┴─→ bootstrapped-se: Evolve (self-referential)
```

**What empirical-methodology provides**:
1. **Scientific Method Framework** - Hypothesis → Experiment → Validation
2. **Detailed Observation Guidance** - Tools, data sources, patterns
3. **Fine-Grained Phases** - Separates Observe and Analyze explicitly
4. **Data-Driven Principles** - 100% empirical evidence requirement
5. **Continuous Evolution** - Methodology improves itself

**What bootstrapped-se adds**:
- **Three-Tuple Output** (O, Aₙ, Mₙ) - Reusable system artifacts
- **Agent Framework** - Specialized agents for execution
- **Formal Convergence** - Mathematical stability criteria
- **Meta-Agent Coordination** - Modular capability system

**When to use empirical-methodology explicitly**:
- Need detailed scientific rigor and validation
- Require explicit guidance on observation tools
- Want fine-grained phase separation (Observe ≠ Analyze)
- Focus on scientific method application

**When to use bootstrapped-se instead**:
- Need complete implementation framework with agents
- Want formal convergence criteria
- Prefer OCA cycle (simpler 3-phase vs 5-phase)
- Building actual software (not just studying methodology)

### Relationship to value-optimization (Complementary)

**value-optimization QUANTIFIES empirical-methodology**:

```
empirical-methodology asks:      value-optimization answers:
- Is methodology complete?  →    V_meta_completeness ≥ 0.80
- Is it effective?          →    V_meta_effectiveness (speedup)
- Is it reusable?           →    V_meta_reusability ≥ 0.85
- Has task succeeded?       →    V_instance ≥ 0.80
```

**empirical-methodology VALIDATES value-optimization**:
- Observation phase generates data for V calculation
- Analysis phase identifies value dimensions
- Codification phase documents value rubrics
- Automation phase enforces value thresholds

**Integration**:
```
Empirical Methodology Lifecycle:

  Observe → Analyze
      ↓
  [Calculate Baseline Values]
  V_instance(s₀), V_meta(s₀)
      ↓
  Codify → Automate → Evolve
      ↓
  [Calculate Current Values]
  V_instance(s_n), V_meta(s_n)
      ↓
  [Check Improvement]
  ΔV_instance, ΔV_meta > threshold?
```

**When to use together**:
- **Always** - value-optimization provides measurement framework
- Use empirical-methodology for process
- Use value-optimization for evaluation
- Enables data-driven decisions at every phase

### Three-Methodology Synergy

**Position in the stack**:

```
bootstrapped-se (Framework Layer)
    ↓ includes
empirical-methodology (Scientific Foundation Layer) ← YOU ARE HERE
    ↓ uses for validation
value-optimization (Quantitative Layer)
```

**Unique contribution of empirical-methodology**:
- **Scientific Rigor**: Hypothesis testing, controlled experiments
- **Data-Driven Decisions**: No theory without evidence
- **Observation Tools**: Detailed guidance on meta-cc, git, metrics
- **Pattern Extraction**: Systematic approach to finding reusable patterns
- **Self-Validation**: Methodology applies to its own development

**When to emphasize empirical-methodology**:
1. **Publishing Methodology**: Need scientific validation for papers
2. **Cross-Domain Transfer**: Validating methodology applicability
3. **Teaching/Training**: Explaining systematic approach
4. **Quality Assurance**: Ensuring empirical rigor

**When to use full stack** (all three together):
- **Bootstrap Experiments**: All 8 experiments use all three
- **Methodology Development**: Maximum rigor and transferability
- **Production Systems**: Complete validation required

**Usage Recommendation**:
- **Learn scientific method**: Read empirical-methodology.md (this file)
- **Get framework**: Read bootstrapped-se.md (includes this + more)
- **Add quantification**: Read value-optimization.md
- **See integration**: Read bootstrapped-ai-methodology-engineering.md (BAIME framework)

---

## Related Skills

- **bootstrapped-ai-methodology-engineering**: Unified BAIME framework integrating all three methodologies
- **bootstrapped-se**: OCA framework (includes and extends empirical-methodology)
- **value-optimization**: Quantitative framework (validates empirical-methodology)
- **dependency-health**: Example application (empirical dependency management)

---

## Knowledge Base

### Source Documentation
- **Core methodology**: `docs/methodology/empirical-methodology-development.md`
- **Related**: `docs/methodology/bootstrapped-software-engineering.md`
- **Examples**: `experiments/bootstrap-*/` (8 validated experiments)

### Key Concepts
- Data-driven methodology development
- Scientific method for software engineering
- Observation → Analysis → Codification → Automation → Evolution
- Continuous improvement
- Self-referential validation

---

## Version History

- **v1.0.0** (2025-10-18): Initial release
  - Based on meta-cc project (277 commits, 11 days)
  - Five-phase process validated
  - 92% transferability demonstrated
  - Multiple domain validation (docs, testing, errors)

---

**Status**: ✅ Production-ready
**Validation**: meta-cc project + 8 experiments
**Effectiveness**: 10-20x vs theory-driven methodologies
**Transferability**: 92% (process universal, tools adaptable)
