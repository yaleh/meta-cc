---
name: Methodology Bootstrapping
description: Apply Bootstrapped AI Methodology Engineering (BAIME) to develop project-specific methodologies through systematic Observe-Codify-Automate cycles with dual-layer value functions (instance quality + methodology quality). Use when creating testing strategies, CI/CD pipelines, error handling patterns, observability systems, or any reusable development methodology. Provides structured framework with convergence criteria, agent coordination, and empirical validation. Validated in 8 experiments with 100% success rate, 4.9 avg iterations, 10-50x speedup vs ad-hoc. Works for testing, CI/CD, error recovery, dependency management, documentation systems, knowledge transfer, technical debt, cross-cutting concerns.
allowed-tools: Read, Grep, Glob, Edit, Write, Bash
---

# Methodology Bootstrapping

**Apply Bootstrapped AI Methodology Engineering (BAIME) to systematically develop and validate software engineering methodologies through observation, codification, and automation.**

> The best methodologies are not designed but evolved through systematic observation, codification, and automation of successful practices.

---

## What is BAIME?

**BAIME (Bootstrapped AI Methodology Engineering)** is a unified framework that integrates three complementary methodologies optimized for LLM-based development:

1. **OCA Cycle** (Observe-Codify-Automate) - Core iterative framework
2. **Empirical Validation** - Scientific method and data-driven decisions
3. **Value Optimization** - Dual-layer value functions for quantitative evaluation

This skill provides the complete BAIME framework for systematic methodology development. The methodology is especially powerful when combined with AI agents (like Claude Code) that can execute the OCA cycle, coordinate specialized agents, and calculate value functions automatically.

**Key Innovation**: BAIME treats methodology development like software development—with empirical observation, automated testing, continuous iteration, and quantitative metrics.

---

## When to Use This Skill

Use this skill when you need to:
- 🎯 **Create systematic methodologies** for testing, CI/CD, error handling, observability, etc.
- 📊 **Validate methodologies empirically** with data-driven evidence
- 🔄 **Evolve practices iteratively** using OCA (Observe-Codify-Automate) cycle
- 📈 **Measure methodology quality** with dual-layer value functions
- 🚀 **Achieve rapid convergence** (typically 3-7 iterations, 6-15 hours)
- 🌍 **Create transferable methodologies** (70-95% reusable across projects)

**Don't use this skill for**:
- ❌ One-time ad-hoc tasks without reusability goals
- ❌ Trivial processes (<100 lines of code/docs)
- ❌ When established industry standards fully solve your problem

---

## Quick Start with BAIME (10 minutes)

### 1. Define Your Domain
Choose what methodology you want to develop using BAIME:
- Testing strategy (15x speedup example)
- CI/CD pipeline (2.5-3.5x speedup example)
- Error recovery patterns (80% error reduction example)
- Observability system (23-46x speedup example)
- Dependency management (6x speedup example)
- Documentation system (47% token cost reduction example)
- Knowledge transfer (3-8x speedup example)
- Technical debt management
- Cross-cutting concerns

### 2. Establish Baseline
Measure current state:
```bash
# Example: Testing domain
- Current coverage: 65%
- Test quality: Ad-hoc
- No systematic approach
- Bug rate: Baseline

# Example: CI/CD domain
- Build time: 5 minutes
- No quality gates
- Manual releases
```

### 3. Set Dual Goals
Define both layers:
- **Instance goal** (domain-specific): "Reach 80% test coverage"
- **Meta goal** (methodology): "Create reusable testing strategy with 85%+ transferability"

### 4. Start Iteration 0
Follow the OCA cycle (see [reference/observe-codify-automate.md](reference/observe-codify-automate.md))

---

## Core Framework

### The OCA Cycle

```
Observe → Codify → Automate
   ↑                    ↓
   └────── Evolve ──────┘
```

**Observe**: Collect empirical data about current practices
- Use meta-cc MCP tools to analyze session history
- Git analysis for commit patterns
- Code metrics (coverage, complexity)
- Access pattern tracking
- Error rate monitoring

**Codify**: Extract patterns and document methodologies
- Pattern recognition from data
- Hypothesis formation
- Documentation as markdown
- Validation with real scenarios

**Automate**: Convert methodologies to automated checks
- Detection: Identify when pattern applies
- Validation: Check compliance
- Enforcement: CI/CD gates
- Suggestion: Automated fix recommendations

**Evolve**: Apply methodology to itself for continuous improvement
- Use tools on development process
- Discover meta-patterns
- Optimize methodology

**Detailed guide**: [reference/observe-codify-automate.md](reference/observe-codify-automate.md)

### Dual-Layer Value Functions

Every iteration calculates two scores:

**V_instance(s)**: Domain-specific task quality
- Example (testing): coverage × quality × stability × performance
- Example (CI/CD): speed × reliability × automation × observability
- Target: ≥0.80

**V_meta(s)**: Methodology transferability quality
- Components: completeness × effectiveness × reusability × validation
- Completeness: Is methodology fully documented?
- Effectiveness: What speedup does it provide?
- Reusability: What % transferable across projects?
- Validation: Is it empirically validated?
- Target: ≥0.80

**Detailed guide**: [reference/dual-value-functions.md](reference/dual-value-functions.md)

### Convergence Criteria

Methodology complete when:
1. ✅ **System stable**: Agent set unchanged for 2+ iterations
2. ✅ **Dual threshold**: V_instance ≥ 0.80 AND V_meta ≥ 0.80
3. ✅ **Objectives complete**: All planned work finished
4. ✅ **Diminishing returns**: ΔV < 0.02 for 2+ iterations

**Alternative patterns**:
- **Meta-Focused Convergence**: V_meta ≥ 0.80, V_instance ≥ 0.55 (when methodology is primary goal)
- **Practical Convergence**: Combined quality exceeds metrics, justified partial criteria

**Detailed guide**: [reference/convergence-criteria.md](reference/convergence-criteria.md)

---

## Three-Layer Architecture

**BAIME** integrates three complementary methodologies into a unified framework:

**Layer 1: Core Framework (OCA Cycle)**
- Observe → Codify → Automate → Evolve
- Three-tuple output: (O, Aₙ, Mₙ)
- Self-referential feedback loop
- Agent coordination

**Layer 2: Scientific Foundation (Empirical Methodology)**
- Empirical observation tools
- Data-driven pattern extraction
- Hypothesis testing
- Scientific validation

**Layer 3: Quantitative Evaluation (Value Optimization)**
- Dual-layer value functions (V_instance + V_meta)
- Convergence mathematics
- Agent as gradient, Meta-Agent as Hessian
- Optimization perspective

**Why "BAIME"?** The framework bootstraps itself—methodologies developed using BAIME can be applied to improve BAIME itself. This self-referential property, combined with AI-agent coordination, makes it uniquely suited for LLM-based development tools.

**Detailed guide**: [reference/three-layer-architecture.md](reference/three-layer-architecture.md)

---

## Proven Results

**Validated in 8 experiments**:
- ✅ 100% success rate (8/8 converged)
- ⏱️ Average: 4.9 iterations, 9.1 hours
- 📈 V_instance average: 0.784 (range: 0.585-0.92)
- 📈 V_meta average: 0.840 (range: 0.83-0.877)
- 🌍 Transferability: 70-95%+
- 🚀 Speedup: 3-46x vs ad-hoc

**Example applications**:
- **Testing strategy**: 15x speedup, 75%→86% coverage ([examples/testing-methodology.md](examples/testing-methodology.md))
- **CI/CD pipeline**: 2.5-3.5x speedup, 91.7% pattern validation ([examples/ci-cd-optimization.md](examples/ci-cd-optimization.md))
- **Error recovery**: 80% error reduction, 85% transferability
- **Observability**: 23-46x speedup, 90-95% transferability
- **Dependency health**: 6x speedup (9h→1.5h), 88% transferability
- **Knowledge transfer**: 3-8x onboarding speedup, 95%+ transferability
- **Documentation**: 47% token cost reduction, 85% transferability
- **Technical debt**: SQALE quantification, 85% transferability

---

## Usage Templates

### Experiment Template
Use [templates/experiment-template.md](templates/experiment-template.md) to structure your methodology development:
- README.md structure
- Iteration prompts
- Knowledge extraction format
- Results documentation

### Iteration Prompt Template
Use [templates/iteration-prompts-template.md](templates/iteration-prompts-template.md) to guide each iteration:
- Iteration N objectives
- OCA cycle execution steps
- Value calculation rubrics
- Convergence checks

---

## Common Pitfalls

❌ **Don't**:
- Use only one methodology layer in isolation (except quick prototyping)
- Predetermine agent evolution path (let specialization emerge from data)
- Force convergence at target iteration count (trust the criteria)
- Inflate value metrics to meet targets (honest assessment critical)
- Skip empirical validation (data-driven decisions only)

✅ **Do**:
- Start with OCA cycle, add evaluation and validation
- Let agent specialization emerge from domain needs
- Trust the convergence criteria (system knows when done)
- Calculate V(s) honestly based on actual state
- Complete all analysis thoroughly before codifying

---

## Related Skills

**Acceleration techniques** (achieve 3-4 iteration convergence):
- [rapid-convergence](../rapid-convergence/SKILL.md) - Fast convergence patterns
- [retrospective-validation](../retrospective-validation/SKILL.md) - Historical data validation
- [baseline-quality-assessment](../baseline-quality-assessment/SKILL.md) - Strong iteration 0

**Supporting skills**:
- [agent-prompt-evolution](../agent-prompt-evolution/SKILL.md) - Track agent specialization

**Domain applications** (ready-to-use methodologies):
- [testing-strategy](../testing-strategy/SKILL.md) - TDD, coverage-driven, fixtures
- [error-recovery](../error-recovery/SKILL.md) - Error taxonomy, recovery patterns
- [ci-cd-optimization](../ci-cd-optimization/SKILL.md) - Quality gates, automation
- [observability-instrumentation](../observability-instrumentation/SKILL.md) - Logging, metrics, tracing
- [dependency-health](../dependency-health/SKILL.md) - Security, freshness, compliance
- [knowledge-transfer](../knowledge-transfer/SKILL.md) - Onboarding, learning paths
- [technical-debt-management](../technical-debt-management/SKILL.md) - SQALE, prioritization
- [cross-cutting-concerns](../cross-cutting-concerns/SKILL.md) - Pattern extraction, enforcement

---

## References

**Core documentation**:
- [Overview](reference/overview.md) - Architecture and philosophy
- [OCA Cycle](reference/observe-codify-automate.md) - Detailed process
- [Value Functions](reference/dual-value-functions.md) - Evaluation framework
- [Convergence Criteria](reference/convergence-criteria.md) - When to stop
- [Three-Layer Architecture](reference/three-layer-architecture.md) - Framework layers

**Quick start**:
- [Quick Start Guide](reference/quick-start-guide.md) - Step-by-step tutorial

**Examples**:
- [Testing Methodology](examples/testing-methodology.md) - Complete walkthrough
- [CI/CD Optimization](examples/ci-cd-optimization.md) - Pipeline example
- [Error Recovery](examples/error-recovery.md) - Error handling example

**Templates**:
- [Experiment Template](templates/experiment-template.md) - Structure your experiment
- [Iteration Prompts](templates/iteration-prompts-template.md) - Guide each iteration

---

**Status**: ✅ Production-ready | BAIME Framework | 8 experiments | 100% success rate | 95% transferable

**Terminology**: This skill implements the **Bootstrapped AI Methodology Engineering (BAIME)** framework. Use "BAIME" when referring to this methodology in documentation, research, or when asking Claude Code for assistance with methodology development.
