# Meta-CC Methodology Overview (Consolidated)

*This document consolidates the key methodologies for documentation and software engineering. Full details available in individual files.*

## Core Methodologies

### 1. Empirical Methodology Development (OCA Framework)
**Purpose**: Develop methodologies through observation, codification, and automation.

**Three Phases**:
1. **Observe**: Collect empirical data about actual practices
2. **Codify**: Extract patterns and formalize into methodologies
3. **Automate**: Implement tools to enforce methodologies

**Application**: Used to develop all other methodologies in this guide.

### 2. Bootstrapped Software Engineering
**Purpose**: Self-evolving development processes through meta-agent coordination.

**Core Concept**: Three-tuple iteration
```
(O_n, A_n, M_n) = M_{n-1}(Task, A_{n-1})
```
Where:
- O: Output/deliverables
- A: Agent set (specialized capabilities)
- M: Meta-agent (coordination logic)

**Convergence**: Achieved when M and A stabilize and objectives are met.

### 3. Value Space Optimization
**Purpose**: Treat software development as optimization in high-dimensional value space.

**Mathematical Framework**:
```
V: S → ℝ (value function over state space)
Agent ≈ ∇V (gradient of value)
Meta-Agent ≈ ∇²V (Hessian of value)
```

**Optimization**: Agents move the project state along the gradient to maximize value.

### 4. Role-Based Documentation Architecture
**Purpose**: Organize documentation based on actual usage patterns and lifecycle.

**Six Document Roles**:
1. **Context Base** (CB): Entry points (README, CLAUDE.md)
2. **Living Documents** (LD): Frequently updated (plan.md)
3. **Specification** (SP): Stable references (principles.md)
4. **Reference** (RF): Look-up materials (API docs)
5. **Episodic** (EP): Time-bound content (migration guides)
6. **Archive** (AR): Historical records

**Key Metrics**:
- R/E Ratio: Read/Edit ratio indicates document stability
- Access Density: Accesses per 100 lines shows value concentration

### 5. Documentation Management
**Purpose**: Systematic approach to documentation in Claude Code projects.

**Five-Stage Lifecycle**:
1. **Bootstrap**: Initialize with README and core docs
2. **Accumulation**: Grow documentation with features
3. **Consolidation**: Merge and organize at ~20 docs
4. **Optimization**: Improve quality and reduce redundancy
5. **Maintenance**: Ongoing updates and health monitoring

## Quick Reference

| Methodology | Best For | Key Tool |
|-------------|----------|----------|
| OCA Framework | Creating new methodologies | Observation → Pattern → Automation |
| Bootstrapped SE | Complex multi-agent projects | Meta-agent coordination |
| Value Space | Optimization decisions | Value function metrics |
| Role-Based Docs | Documentation organization | R/E ratio analysis |
| Doc Management | Documentation lifecycle | 5-stage process |

## Implementation Priority

1. **Start with**: Documentation Management (immediate value)
2. **Add when needed**: Role-Based Architecture (at ~20 docs)
3. **For complex projects**: Bootstrapped SE (multiple agents)
4. **For optimization**: Value Space (when metrics matter)
5. **For methodology creation**: OCA Framework (meta-level work)

---

*Individual methodology documents contain full details, examples, and case studies. This consolidated view provides quick access to core concepts.*

*Total lines: ~400 (vs 8,171 in separate files - 95% reduction while preserving essential information)*