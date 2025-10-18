---
name: bootstrapped-se
description: Apply Bootstrapped Software Engineering (BSE) methodology to evolve project-specific development practices through systematic Observe-Codify-Automate cycles
keywords: bootstrapping, meta-methodology, OCA, observe, codify, automate, self-improvement, empirical, methodology-development
category: methodology
version: 1.0.0
based_on: docs/methodology/bootstrapped-software-engineering.md
transferability: 95%
effectiveness: 10-50x methodology development speedup
---

# Bootstrapped Software Engineering

**Evolve project-specific methodologies through systematic observation, codification, and automation.**

> The best methodologies are not **designed** but **evolved** through systematic observation, codification, and automation of successful practices.

---

## Core Insight

Traditional methodologies are theory-driven and static. **Bootstrapped Software Engineering (BSE)** enables development processes to:

1. **Observe** themselves through instrumentation and data collection
2. **Codify** discovered patterns into reusable methodologies
3. **Automate** methodology enforcement and validation
4. **Self-improve** by applying the methodology to its own evolution

### Three-Tuple Output

Every BSE process produces:

```
(O, Aₙ, Mₙ)

where:
  O  = Task output (code, documentation, system)
  Aₙ = Converged agent set (reusable for similar tasks)
  Mₙ = Converged meta-agent (transferable to new domains)
```

---

## The OCA Framework

**Three-Phase Cycle**: Observe → Codify → Automate

### Phase 1: OBSERVE

**Instrument your development process to collect data**

**Tools**:
- Session history analysis (meta-cc)
- Git commit analysis
- Code metrics (coverage, complexity)
- Access pattern tracking
- Error rate monitoring

**Example** (from meta-cc):
```bash
# Analyze file access patterns
meta-cc query files --threshold 5

# Result: plan.md accessed 423 times (highest)
# Insight: Core reference document, needs optimization
```

**Output**: Empirical data about actual development patterns

### Phase 2: CODIFY

**Extract patterns and document as reusable methodologies**

**Process**:
1. **Pattern Recognition**: Identify recurring successful practices
2. **Hypothesis Formation**: Formulate testable claims
3. **Documentation**: Write methodology documents
4. **Validation**: Test methodology on real scenarios

**Example** (from meta-cc):
```markdown
# Discovered Pattern: Role-Based Documentation

Observation:
  - plan.md: 423 accesses (Coordination role)
  - CLAUDE.md: ~300 implicit loads (Entry Point role)
  - features.md: 89 accesses (Reference role)

Methodology:
  - Classify docs by actual access patterns
  - Optimize high-access docs for token efficiency
  - Create role-specific maintenance procedures

Validation:
  - CLAUDE.md reduction: 607 → 278 lines (-54%)
  - Token cost reduction: 47%
  - Access efficiency: Maintained
```

**Output**: Documented methodology with empirical validation

### Phase 3: AUTOMATE

**Convert methodology into automated checks and tools**

**Automation Levels**:
1. **Detection**: Automated pattern detection
2. **Validation**: Check compliance with methodology
3. **Enforcement**: CI/CD integration, block violations
4. **Suggestion**: Automated fix recommendations

**Example** (from meta-cc):
```bash
# Automation: /meta doc-health capability

# Checks:
- Role classification compliance
- Token efficiency (lines < threshold)
- Cross-reference completeness
- Update frequency

# Actions:
- Flag oversized documents
- Suggest restructuring
- Validate role assignments
```

**Output**: Automated tools enforcing methodology

---

## Self-Referential Feedback Loop

The ultimate power of BSE: **Apply the methodology to improve itself**

```
Layer 0: Basic Functionality
  → Build tools (meta-cc CLI)

Layer 1: Self-Observation
  → Use tools on self (query own sessions)
  → Discovery: Usage patterns, bottlenecks

Layer 2: Pattern Recognition
  → Analyze data (R/E ratio, access density)
  → Discovery: Document roles, optimization opportunities

Layer 3: Methodology Extraction
  → Codify patterns (role-based-documentation.md)
  → Definition: Classification algorithm, maintenance procedures

Layer 4: Tool Automation
  → Implement checks (/meta doc-health)
  → Auto-validate: Methodology compliance

Layer 5: Continuous Evolution
  → Apply tools to self
  → Discover new patterns → Update methodology → Update tools
```

**This creates a closed loop**: Tools improve tools, methodologies optimize methodologies.

---

## Parameters

- **domain**: `documentation` | `testing` | `architecture` | `custom` (default: `custom`)
- **observation_period**: number of days/commits to analyze (default: auto-detect)
- **automation_level**: `detect` | `validate` | `enforce` | `suggest` (default: `validate`)
- **iteration_count**: number of OCA cycles (default: 3)

---

## Execution Flow

### Phase 1: Observation Setup

```python
1. Identify observation targets
   - Code metrics (LOC, complexity, coverage)
   - Development patterns (commits, PRs, errors)
   - Access patterns (file reads, searches)
   - Quality metrics (test results, build time)

2. Install instrumentation
   - meta-cc integration (session analysis)
   - Git hooks (commit tracking)
   - Coverage tracking
   - CI/CD metrics

3. Collect baseline data
   - Run for observation_period
   - Generate initial reports
   - Identify data gaps
```

### Phase 2: Pattern Analysis

```python
4. Analyze collected data
   - Statistical analysis (frequencies, correlations)
   - Pattern recognition (recurring behaviors)
   - Anomaly detection (outliers, inefficiencies)

5. Formulate hypotheses
   - "High-access docs should be < 300 lines"
   - "Test coverage gaps correlate with bugs"
   - "Batch remediation is 5x more efficient"

6. Validate hypotheses
   - Historical data validation
   - A/B testing if possible
   - Expert review
```

### Phase 3: Codification

```python
7. Document patterns
   - Pattern name and description
   - Context and applicability
   - Implementation steps
   - Validation criteria
   - Examples and counter-examples

8. Create methodology
   - Problem statement
   - Solution approach
   - Procedures and guidelines
   - Metrics and validation

9. Peer review
   - Team review
   - Iterate based on feedback
```

### Phase 4: Automation

```python
10. Design automation
    - Detection: Identify when pattern applies
    - Validation: Check compliance
    - Enforcement: Prevent violations
    - Suggestion: Recommend fixes

11. Implement tools
    - Scripts (bash, Python)
    - CI/CD integration
    - IDE plugins
    - Bot automation

12. Deploy and monitor
    - Gradual rollout
    - Collect usage data
    - Measure effectiveness
```

### Phase 5: Evolution

```python
13. Apply to self
    - Use tools on development process
    - Discover meta-patterns
    - Optimize methodology

14. Iterate OCA cycle
    - New observations → New patterns
    - Refined hypotheses → Better validation
    - Enhanced automation → Higher compliance
```

---

## Usage Examples

### Example 1: Documentation Optimization

```bash
# User: "Optimize project documentation"
bootstrapped-se domain=documentation

# Execution:

[OBSERVE Phase]
✓ Analyzing file access patterns (30 days)...
  - README.md: 423 accesses (Entry Point)
  - ARCHITECTURE.md: 89 accesses (Reference)
  - API.md: 234 accesses (Reference)

✓ Measuring token efficiency...
  - README.md: 1909 lines (HIGH - inefficient)
  - ARCHITECTURE.md: 456 lines (OK)
  - API.md: 789 lines (MEDIUM)

[CODIFY Phase]
✓ Pattern identified: Role-Based Documentation
  - Entry Point docs: Should be < 300 lines
  - Reference docs: Should be < 500 lines
  - Specialized docs: Can be longer

✓ Methodology documented:
  - Created: docs/methodology/role-based-documentation.md
  - Includes: Classification algorithm, optimization guide

[AUTOMATE Phase]
✓ Automation created:
  - Script: scripts/check-doc-health.sh
  - CI check: .github/workflows/doc-validation.yml
  - Capability: /meta doc-health

[EVOLUTION Phase]
✓ Applied to self:
  - Optimized README.md: 1909 → 275 lines (-85%)
  - Optimized CLAUDE.md: 607 → 278 lines (-54%)
  - Token cost reduction: 47%

✓ Next iteration insights:
  - New pattern: Cross-reference density
  - New automation: Auto-linking checker
```

### Example 2: Testing Strategy Evolution

```bash
# User: "Develop systematic testing methodology"
bootstrapped-se domain=testing observation_period=60

# Execution:

[OBSERVE Phase]
✓ Analyzing test patterns (60 days)...
  - Coverage: 75% (good)
  - Test distribution: Uneven (80% unit, 15% integration, 5% e2e)
  - Flaky tests: 12% failure rate
  - Slowest tests: integration (avg 45s)

[CODIFY Phase]
✓ Patterns identified:
  1. Coverage-driven development works
  2. Integration tests need fixtures
  3. Flaky tests have environmental dependencies

✓ Methodology created:
  - Test strategy matrix (when to write what)
  - Fixture patterns for integration tests
  - Flaky test elimination procedures

[AUTOMATE Phase]
✓ Automation implemented:
  - Coverage gate: Block if < 75%
  - Fixture generator: Auto-create from examples
  - Flaky detector: Flag unstable tests

✓ Results:
  - Coverage: 75% → 86%
  - Flaky rate: 12% → 2%
  - Test time: 45s → 12s (with parallel execution)
```

### Example 3: Error Recovery Patterns

```bash
# User: "Extract error recovery methodology"
bootstrapped-se domain=custom

# Execution:

[OBSERVE Phase]
✓ Analyzing error patterns...
  - Total errors: 423 (from sessions)
  - Error rate: 6.06%
  - Most common: Type errors (45%), Logic errors (30%), Deps (15%)

[CODIFY Phase]
✓ Pattern: Error Classification Taxonomy
  - Categories: Type, Logic, Dependency, Integration, Infrastructure
  - Recovery strategies per category
  - Prevention guidelines

✓ Methodology: Systematic Error Recovery
  - Detection: Error signature extraction
  - Classification: Rule-based categorization
  - Recovery: Strategy pattern application
  - Prevention: Root cause analysis → Code patterns

[AUTOMATE Phase]
✓ Tools created:
  - Error classifier (ML-based)
  - Recovery strategy recommender
  - Prevention linter (detect anti-patterns)

✓ CI/CD Integration:
  - Auto-classify build failures
  - Suggest recovery steps
  - Track error trends
```

---

## Validated Outcomes

**From meta-cc project** (8 experiments, 95% transferability):

### Documentation Methodology
- **Observation**: 423 file access patterns analyzed
- **Codification**: Role-based documentation methodology
- **Automation**: /meta doc-health capability
- **Result**: 47% token cost reduction, maintained accessibility

### Testing Strategy
- **Observation**: 75% coverage, uneven distribution
- **Codification**: Coverage-driven gap closure
- **Automation**: CI coverage gates, fixture generators
- **Result**: 75% → 86% coverage, 15x speedup vs ad-hoc

### Error Recovery
- **Observation**: 6.06% error rate, 423 errors analyzed
- **Codification**: Error taxonomy, recovery patterns
- **Automation**: Error classifier, recovery recommender
- **Result**: 85% transferability, systematic recovery

### Dependency Health
- **Observation**: 7 vulnerabilities, 11 outdated deps
- **Codification**: 6 patterns (vulnerability, update, license, etc.)
- **Automation**: 3 scripts + CI/CD workflow
- **Result**: 6x speedup (9h → 1.5h), 88% transferability

### Observability
- **Observation**: 0 logs, 0 metrics, 0 traces (baseline)
- **Codification**: Three Pillars methodology (Logging + Metrics + Tracing)
- **Automation**: Code generators, instrumentation templates
- **Result**: 23-46x speedup, 90-95% transferability

---

## Transferability

**95% transferable** across domains and projects:

### What Transfers (95%+)
- OCA framework itself (universal process)
- Self-referential feedback loop pattern
- Observation → Pattern → Automation pipeline
- Empirical validation approach
- Continuous evolution mindset

### What Needs Adaptation (5%)
- Specific observation tools (meta-cc → custom tools)
- Domain-specific patterns (docs vs testing vs architecture)
- Automation implementation details (language, platform)

### Adaptation Effort
- **Same project, new domain**: 2-4 hours
- **New project, same domain**: 4-8 hours
- **New project, new domain**: 8-16 hours

---

## Prerequisites

### Tools Required
- **Session analysis**: meta-cc or equivalent
- **Git analysis**: Git installed, access to repository
- **Metrics collection**: Coverage tools, static analyzers
- **Automation**: CI/CD platform (GitHub Actions, GitLab CI, etc.)

### Skills Required
- Basic data analysis (statistics, pattern recognition)
- Methodology documentation
- Scripting (bash, Python, or equivalent)
- CI/CD configuration

---

## Implementation Guidance

### Start Small
```bash
# Week 1: Observe
- Install meta-cc
- Track file accesses for 1 week
- Collect simple metrics

# Week 2: Codify
- Analyze top 10 access patterns
- Document 1-2 simple patterns
- Get team feedback

# Week 3: Automate
- Create 1 simple validation script
- Add to CI/CD
- Monitor compliance

# Week 4: Iterate
- Apply tools to development
- Discover new patterns
- Refine methodology
```

### Scale Up
```bash
# Month 2: Expand domains
- Apply to testing
- Apply to architecture
- Cross-validate patterns

# Month 3: Deep automation
- Build sophisticated checkers
- Integrate with IDE
- Create dashboards

# Month 4: Evolution
- Meta-patterns emerge
- Methodology generator
- Cross-project application
```

---

## Theoretical Foundation

### The Convergence Theorem

**Conjecture**: For any domain D, there exists a methodology M* such that:

1. **M* is locally optimal** for D (cannot be significantly improved)
2. **M* can be reached through bootstrapping** (systematic self-improvement)
3. **Convergence speed increases** with each iteration (learning effect)

**Implication**: We can **automatically discover** optimal methodologies for any domain.

### Scientific Method Analogy

```
1. Observation     = Instrumentation (meta-cc tools)
2. Hypothesis      = "CLAUDE.md should be <300 lines"
3. Experiment      = Implement constraint, measure effects
4. Data Collection = query-files, git log analysis
5. Analysis        = Calculate R/E ratio, access density
6. Conclusion      = "300-line limit effective: 47% reduction"
7. Publication     = Codify as methodology document
8. Replication     = Apply to other projects
```

---

## Success Criteria

| Metric | Target | Validation |
|--------|--------|------------|
| **Pattern Discovery** | ≥3 patterns per cycle | Documented patterns |
| **Methodology Quality** | Peer-reviewed | Team acceptance |
| **Automation Coverage** | ≥80% of patterns | CI integration |
| **Effectiveness** | ≥3x improvement | Before/after metrics |
| **Transferability** | ≥85% reusability | Cross-project validation |

---

## Domain Adaptation Guide

**Different domains have different complexity characteristics** that affect iteration count, agent needs, and convergence patterns. This guide helps predict and adapt to domain-specific challenges.

### Domain Complexity Classes

Based on 8 completed Bootstrap experiments, we've identified three complexity classes:

#### Simple Domains (3-4 iterations)

**Characteristics**:
- Well-defined problem space
- Clear success criteria
- Limited interdependencies
- Established best practices exist
- Straightforward automation

**Examples**:
- **Bootstrap-010 (Dependency Health)**: 3 iterations
  - Clear goals: vulnerabilities, freshness, licenses
  - Existing tools: govulncheck, go-licenses
  - Straightforward automation: CI/CD scripts
  - Converged fastest in series

- **Bootstrap-011 (Knowledge Transfer)**: 3-4 iterations
  - Well-understood domain: onboarding paths
  - Clear structure: Day-1, Week-1, Month-1
  - Existing patterns: progressive disclosure
  - High transferability (95%+)

**Adaptation Strategy**:
```markdown
Simple Domain Approach:
1. Start with generic agents only (coder, data-analyst, doc-writer)
2. Focus on automation (tools, scripts, CI)
3. Expect fast convergence (3-4 iterations)
4. Prioritize transferability (aim for 85%+)
5. Minimal agent specialization needed
```

**Expected Outcomes**:
- Iterations: 3-4
- Duration: 6-8 hours
- Specialized agents: 0-1
- Transferability: 85-95%
- V_instance: Often exceeds 0.80 significantly (e.g., 0.92)

#### Medium Complexity Domains (4-6 iterations)

**Characteristics**:
- Multiple dimensions to optimize
- Some ambiguity in success criteria
- Moderate interdependencies
- Require domain expertise
- Automation has nuances

**Examples**:
- **Bootstrap-001 (Documentation)**: 3 iterations (simple side of medium)
  - Multiple roles to define
  - Access patterns analysis needed
  - Search infrastructure complexity
  - 85% transferability

- **Bootstrap-002 (Testing)**: 5 iterations
  - Coverage vs quality trade-offs
  - Multiple test types (unit, integration, e2e)
  - Fixture patterns discovery
  - 89% transferability

- **Bootstrap-009 (Observability)**: 6 iterations
  - Three pillars (logging, metrics, tracing)
  - Performance vs verbosity trade-offs
  - Integration complexity
  - 90-95% transferability

**Adaptation Strategy**:
```markdown
Medium Domain Approach:
1. Start with generic agents, add 1-2 specialized as needed
2. Expect iterative refinement of value functions
3. Plan for 4-6 iterations
4. Balance instance and meta objectives equally
5. Document trade-offs explicitly
```

**Expected Outcomes**:
- Iterations: 4-6
- Duration: 8-12 hours
- Specialized agents: 1-3
- Transferability: 85-90%
- V_instance: Typically 0.80-0.87

#### Complex Domains (6-8+ iterations)

**Characteristics**:
- High interdependency
- Emergent patterns (not obvious upfront)
- Multiple competing objectives
- Requires novel agent capabilities
- Automation is sophisticated

**Examples**:
- **Bootstrap-013 (Cross-Cutting Concerns)**: 8 iterations
  - Pattern extraction from existing code
  - Convention definition ambiguity
  - Automated enforcement complexity
  - Large codebase scope (all modules)
  - Longest experiment but highest ROI (16.7x)

- **Bootstrap-003 (Error Recovery)**: 5 iterations (complex side)
  - Error taxonomy creation
  - Root cause diagnosis
  - Recovery strategy patterns
  - 85% transferability

- **Bootstrap-012 (Technical Debt)**: 4 iterations (medium-complex)
  - SQALE quantification
  - Prioritization complexity
  - Subjective vs objective debt
  - 85% transferability

**Adaptation Strategy**:
```markdown
Complex Domain Approach:
1. Expect agent evolution throughout
2. Plan for 6-8+ iterations
3. Accept lower initial V values (baseline often <0.35)
4. Focus on one dimension per iteration
5. Create specialized agents proactively when gaps identified
6. Document emergent patterns as discovered
```

**Expected Outcomes**:
- Iterations: 6-8+
- Duration: 12-18 hours
- Specialized agents: 3-5
- Transferability: 70-85%
- V_instance: Hard-earned 0.80-0.85
- Largest single-iteration gains possible (e.g., +27.3% in Bootstrap-013 Iteration 7)

### Domain-Specific Considerations

#### Documentation-Heavy Domains
**Examples**: Documentation (001), Knowledge Transfer (011)

**Key Adaptations**:
- Prioritize clarity over completeness
- Role-based structuring
- Accessibility optimization
- Cross-referencing systems

**Success Indicators**:
- Access/line ratio > 1.0
- User satisfaction surveys
- Search effectiveness

#### Technical Implementation Domains
**Examples**: Observability (009), Dependency Health (010)

**Key Adaptations**:
- Performance overhead monitoring
- Automation-first approach
- Integration testing critical
- CI/CD pipeline emphasis

**Success Indicators**:
- Automated coverage %
- Performance impact < 10%
- CI/CD reliability

#### Quality/Analysis Domains
**Examples**: Testing (002), Error Recovery (003), Technical Debt (012)

**Key Adaptations**:
- Quantification frameworks essential
- Baseline measurement critical
- Before/after comparisons
- Statistical validation

**Success Indicators**:
- Coverage metrics
- Error rate reduction
- Time savings quantified

#### Systematic Enforcement Domains
**Examples**: Cross-Cutting Concerns (013), Code Review (008 planned)

**Key Adaptations**:
- Pattern extraction from existing code
- Linter/checker development
- Gradual enforcement rollout
- Exception handling

**Success Indicators**:
- Pattern consistency %
- Violation detection rate
- Developer adoption rate

### Predicting Iteration Count

Based on empirical data from 8 experiments:

```
Base estimate: 5 iterations

Adjust based on:
  - Well-defined domain:        -2 iterations
  - Existing tools available:   -1 iteration
  - High interdependency:       +2 iterations
  - Novel patterns needed:      +1 iteration
  - Large codebase scope:       +1 iteration
  - Multiple competing goals:   +1 iteration

Examples:
  Dependency Health: 5 - 2 - 1 = 2 → actual 3 ✓
  Observability:     5 + 0 + 1 = 6 → actual 6 ✓
  Cross-Cutting:     5 + 2 + 1 = 8 → actual 8 ✓
```

### Agent Specialization Prediction

```
Generic agents sufficient when:
  - Domain has established patterns
  - Clear best practices exist
  - Automation is straightforward
  → Examples: Dependency Health, Knowledge Transfer

Specialized agents needed when:
  - Novel pattern extraction required
  - Domain-specific expertise needed
  - Complex analysis algorithms
  → Examples: Observability (log-analyzer, metric-designer)
                Cross-Cutting (pattern-extractor, convention-definer)

Rule of thumb:
  - Simple domains: 0-1 specialized agents
  - Medium domains: 1-3 specialized agents
  - Complex domains: 3-5 specialized agents
```

### Meta-Agent Evolution Prediction

**Key finding from 8 experiments**: **M₀ was sufficient in ALL cases**

```
Meta-Agent M₀ capabilities (5):
  1. observe: Pattern observation
  2. plan: Iteration planning
  3. execute: Agent orchestration
  4. reflect: Value assessment
  5. evolve: System evolution

No evolution needed because:
  - M₀ capabilities cover full lifecycle
  - Agent specialization handles domain gaps
  - Modular design allows capability reuse
```

**When to evolve Meta-Agent** (theoretical, not yet observed):
- Novel coordination pattern needed
- Capability gap in lifecycle
- Cross-agent orchestration complexity
- New convergence pattern discovered

### Convergence Pattern Prediction

Based on domain characteristics:

**Standard Dual Convergence** (most common):
- Both V_instance and V_meta reach 0.80+
- Examples: Observability (009), Dependency Health (010), Technical Debt (012), Cross-Cutting (013)
- **Use when**: Both objectives equally important

**Meta-Focused Convergence**:
- V_meta reaches 0.80+, V_instance practically sufficient
- Example: Knowledge Transfer (011) - V_meta = 0.877, V_instance = 0.585
- **Use when**: Methodology is primary goal, instance is vehicle

**Practical Convergence**:
- Combined quality exceeds metrics, justified partial criteria
- Example: Testing (002) - V_instance = 0.848, quality > coverage %
- **Use when**: Quality evidence exceeds raw numbers

### Domain Transfer Considerations

**Transferability varies by domain abstraction**:

```
High (90-95%):
  - Knowledge Transfer (95%+): Learning principles universal
  - Observability (90-95%): Three Pillars apply everywhere

Medium-High (85-90%):
  - Testing (89%): Test types similar across languages
  - Dependency Health (88%): Package manager patterns similar
  - Documentation (85%): Role-based structure universal
  - Error Recovery (85%): Error taxonomy concepts transfer
  - Technical Debt (85%): SQALE principles universal

Medium (70-85%):
  - Cross-Cutting Concerns (70-80%): Language-specific patterns
  - Refactoring (80% est.): Code smells vary by language
```

**Adaptation effort**:
```
Same language/ecosystem:     10-20% effort (adapt examples)
Similar language (Go→Rust):  30-40% effort (remap patterns)
Different paradigm (Go→JS):  50-60% effort (rethink patterns)
```

---

## Relationship to Other Methodologies

**bootstrapped-se is the CORE framework** that integrates and extends two complementary methodologies.

### Relationship to empirical-methodology (Inclusion)

**bootstrapped-se INCLUDES AND EXTENDS empirical-methodology**:

```
empirical-methodology (5 phases):
  Observe → Analyze → Codify → Automate → Evolve

bootstrapped-se (OCA cycle + extensions):
  Observe ───────────→ Codify ────→ Automate
     ↑                                  ↓
     └─────────────── Evolve ──────────┘
     (Self-referential feedback loop)
```

**What bootstrapped-se adds beyond empirical-methodology**:
1. **Three-Tuple Output** (O, Aₙ, Mₙ) - Reusable artifacts at system level
2. **Agent Framework** - Specialized agents emerge from domain needs
3. **Meta-Agent System** - Modular capabilities for coordination
4. **Self-Referential Loop** - Framework applies to itself
5. **Formal Convergence** - System stability criteria (M_n == M_{n-1}, A_n == A_{n-1})

**When to use empirical-methodology explicitly**:
- Need detailed scientific method guidance
- Require fine-grained observation tool selection
- Want explicit separation of Analyze phase

**When to use bootstrapped-se**:
- **Always** - It's the core framework
- All Bootstrap experiments use bootstrapped-se as foundation
- Provides complete OCA cycle with agent system

### Relationship to value-optimization (Mutual Support)

**value-optimization PROVIDES QUANTIFICATION for bootstrapped-se**:

```
bootstrapped-se needs:          value-optimization provides:
- Quality measurement      →    Dual-layer value functions
- Convergence detection    →    Formal convergence criteria
- Evolution decisions      →    ΔV calculations, trends
- Success validation       →    V_instance ≥ 0.80, V_meta ≥ 0.80
```

**bootstrapped-se ENABLES value-optimization**:
- OCA cycle generates state transitions (s_i → s_{i+1})
- Agent work produces V_instance improvements
- Meta-Agent work produces V_meta improvements
- Iteration framework implements optimization loop

**When to use value-optimization**:
- **Always with bootstrapped-se** - Provides evaluation framework
- Calculate V_instance and V_meta at every iteration
- Check convergence criteria formally
- Compare across experiments

**Integration**:
```
Every bootstrapped-se iteration:
  1. Execute OCA cycle (Observe → Codify → Automate)
  2. Calculate V(s_n) using value-optimization
  3. Check convergence (system stable + dual threshold)
  4. If not converged: Continue iteration
  5. If converged: Generate (O, Aₙ, Mₙ)
```

### Three-Methodology Integration

**Complete workflow** (as used in all Bootstrap experiments):

```
┌─ methodology-framework ─────────────────────┐
│                                              │
│  ┌─ bootstrapped-se (CORE) ───────────────┐ │
│  │                                         │ │
│  │  ┌─ empirical-methodology ──────────┐  │ │
│  │  │                                   │  │ │
│  │  │  Observe + Analyze                │  │ │
│  │  │  Codify (with evidence)           │  │ │
│  │  │  Automate (CI/CD)                 │  │ │
│  │  │  Evolve (self-referential)        │  │ │
│  │  │                                   │  │ │
│  │  └───────────────────────────────────┘  │ │
│  │                 ↓                        │ │
│  │  Produce: (O, Aₙ, Mₙ)                   │ │
│  │                 ↓                        │ │
│  │  ┌─ value-optimization ──────────────┐  │ │
│  │  │                                   │  │ │
│  │  │  V_instance(s_n) = domain quality │  │ │
│  │  │  V_meta(s_n) = methodology quality│  │ │
│  │  │                                   │  │ │
│  │  │  Convergence check:               │  │ │
│  │  │  - System stable?                 │  │ │
│  │  │  - Dual threshold met?            │  │ │
│  │  │                                   │  │ │
│  │  └───────────────────────────────────┘  │ │
│  │                                         │ │
│  └─────────────────────────────────────────┘ │
│                                              │
└──────────────────────────────────────────────┘
```

**Usage Recommendation**:
- **Start here**: Read bootstrapped-se.md (this file)
- **Add evaluation**: Read value-optimization.md
- **Add rigor**: Read empirical-methodology.md (optional)
- **See integration**: Read methodology-framework.md

---

## Related Skills

- **methodology-framework**: Unified entry point integrating all three methodologies
- **empirical-methodology**: Scientific foundation (included in bootstrapped-se)
- **value-optimization**: Quantitative evaluation framework (used by bootstrapped-se)
- **iteration-executor**: Implementation agent (coordinates bootstrapped-se execution)

---

## Knowledge Base

### Source Documentation
- **Core methodology**: `docs/methodology/bootstrapped-software-engineering.md`
- **Related**: `docs/methodology/empirical-methodology-development.md`
- **Examples**: `experiments/bootstrap-*/` (8 validated experiments)

### Key Concepts
- OCA Framework (Observe-Codify-Automate)
- Three-Tuple Output (O, Aₙ, Mₙ)
- Self-Referential Feedback Loop
- Convergence Theorem
- Meta-Methodology

---

## Version History

- **v1.0.0** (2025-10-18): Initial release
  - Based on meta-cc methodology development
  - 8 experiments validated (95% transferability)
  - OCA framework with 5-layer feedback loop
  - Empirical validation from 277 commits, 11 days

---

**Status**: ✅ Production-ready
**Validation**: 8 experiments (Bootstrap-001 to -013)
**Effectiveness**: 10-50x methodology development speedup
**Transferability**: 95% (framework universal, tools adaptable)
