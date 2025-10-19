# Bootstrap-006: API Design Methodology

**Status**: In Progress (Iteration 3 Complete) - ⚠️ **RESTRUCTURING REQUIRED**
**Start Date**: 2025-10-15
**Restructuring Date**: 2025-10-15

---

## ⚠️ CRITICAL ARCHITECTURAL ISSUE IDENTIFIED

**Problem**: Iterations 0-3 conflated meta-tasks (methodology development) with instance tasks (concrete API improvements). Agents created **specifications and design documents** instead of executing **concrete API improvements** for meta-agent to observe.

**Impact**: The experiment has been developing methodology documents directly rather than extracting methodology from observed agent work patterns (violates OCA framework's "Observe" step).

**Resolution**: Implement two-layer architecture going forward.

---

## Experiment Objectives (Two-Layer Architecture)

### Meta-Objective (Meta-Agent Layer)

**Goal**: Develop API design methodology through iterative observation, codification, and automation of agent API improvement patterns.

**Deliverables**:
- API design methodology extracted from observed patterns
- Best practices for API usability, consistency, completeness, evolvability
- Specialized agents (if needed) for API design tasks
- Reusable API design three-tuple (M, A, methodology artifacts)

**Success Criteria**:
- Convergence achieved: V(s) ≥ 0.80
- Methodology artifacts complete and transferable
- Meta-agent and agent set stable

### Instance Objective (Agent Layer)

**Goal**: Improve meta-cc MCP server API (16 tools) for usability and consistency.

**Concrete Scope**:
- **Target**: All 16 MCP tools in `cmd/mcp/tools.go` (~2,500 lines)
- **Specific Tasks**:
  1. Fix parameter ordering in `query_tools`, `query_user_messages`, `query_conversation` (3 tools)
  2. Improve error messages in tools with high error rates
  3. Enhance documentation for tools with low usage (query_context, cleanup_temp_files, etc.)
  4. Implement validation tooling (`meta-cc validate-api` command)
  5. Create pre-commit hooks for API consistency

**Success Criteria**:
- Parameter ordering follows tier-based convention (100% compliance)
- Error messages are actionable and clear (V_usability ≥ 0.80)
- All 16 tools have complete documentation (V_completeness ≥ 0.80)
- API consistency enforced via automated tooling (V_consistency ≥ 0.85)

**Expected Agent Work**:
Agents execute concrete API improvements (parameter reordering, error message enhancement, documentation, tooling implementation). Meta-agent observes patterns and codifies them into API design methodology.

---

## Experiment Overview

This experiment applies the bootstrapped software engineering methodology to develop systematic API design practices. The goal is to improve meta-cc MCP server API quality across four dimensions: usability, consistency, completeness, and evolvability.

**Architectural Principle**: Agents work on concrete API improvements; meta-agent observes patterns and extracts methodology.

### Three-Methodology Framework

This experiment integrates three complementary methodologies:

1. **Empirical Methodology Development (OCA Framework)**
   - **Observe**: Collect data about current API usage and patterns
   - **Codify**: Extract design principles and standards from observations
   - **Automate**: Create validation tools and consistency checkers

2. **Bootstrapped Software Engineering**
   - **Three-Tuple**: (Output, Agent Set, Meta-Agent) evolves iteratively
   - **Convergence**: Iterations continue until M_n = M_{n-1}, A_n = A_{n-1}, V(s_n) ≥ 0.80
   - **Reusability**: Final three-tuple transferable to other API design tasks

3. **Value Space Optimization**
   - **Value Function**: V: S → ℝ maps API state to quality score
   - **Agent as Gradient**: Agents optimize ∇V(s) to improve API quality
   - **Meta-Agent as Hessian**: Meta-agent optimizes ∇²V(s) for agent effectiveness

---

## Current State (Iteration 0)

### Baseline Metrics

```yaml
V(s₀): 0.61 / 0.80 (target)

Components:
  V_usability: 0.74 (parameter clarity, error messages, documentation)
  V_consistency: 0.72 (naming patterns, parameter conventions)
  V_completeness: 0.65 (feature coverage, edge cases)
  V_evolvability: 0.22 (versioning, deprecation, migration) ⚠️ CRITICAL

Gap to Target: 0.19

Priority Issue: Evolvability is critically low (0.22)
  - No versioning strategy
  - No deprecation policy
  - No migration support
```

### System Configuration

**Meta-Agent M₀**: 5 capabilities
- observe.md (API data collection)
- plan.md (improvement prioritization)
- execute.md (agent coordination)
- reflect.md (quality evaluation)
- evolve.md (agent specialization triggers)

**Agent Set A₀**: 8 agents
- Generic (4): coder, data-analyst, doc-writer, doc-generator
- Specialized (4): search-optimizer, error-classifier, recovery-advisor, root-cause-analyzer

**Initial State**: Discovered through union search across experiments/bootstrap-*

---

## API Coverage

**16 MCP Tools Analyzed**:

Heavily Used:
- query_tools (105 calls)
- query_user_messages (97 calls)

Frequently Used:
- get_session_stats (48 calls)
- list_capabilities (45 calls)

Moderate Use:
- query_files, query_tool_sequences, query_successful_prompts, get_capability, query_conversation

Low/Rare Use:
- query_file_access, query_time_series, query_project_state, query_assistant_messages
- cleanup_temp_files, query_tools_advanced, query_context (0 calls observed)

---

## Iteration Progress

### Iteration 0: Baseline Establishment ✓

**Status**: Completed
**Date**: 2025-10-15
**Objectives Met**:
- ✓ M₀ capabilities discovered and adapted (5 files)
- ✓ A₀ agents discovered (8 files)
- ✓ API data collected (16 tools, 467 usage records)
- ✓ Value function calculated: V(s₀) = 0.61
- ✓ Problems identified (evolvability critical, naming inconsistencies, error clarity)

**Key Finding**: Evolvability gap (V = 0.22) is the highest-priority issue requiring immediate attention.

**Recommendation for Iteration 1**: Design API versioning and evolution strategy

### Iteration 1: [Planned]

**Proposed Focus**: Address evolvability gap
**Expected Work**:
- Design API versioning strategy
- Define deprecation policy
- Document backward compatibility guidelines
- Create migration path patterns
- Calculate V(s₁)

**Expected ΔV**: +0.096 (if V_evolvability improved from 0.22 to 0.70)

---

## Directory Structure

```
experiments/bootstrap-006-api-design/
├── README.md                    (this file)
├── ITERATION-PROMPTS.md         (execution templates for each iteration)
├── iteration-0.md               (baseline establishment documentation)
├── meta-agents/                 (M₀ capability files)
│   ├── observe.md
│   ├── plan.md
│   ├── execute.md
│   ├── reflect.md
│   └── evolve.md
├── agents/                      (A₀ agent prompt files)
│   ├── coder.md
│   ├── data-analyst.md
│   ├── doc-writer.md
│   ├── doc-generator.md
│   ├── search-optimizer.md
│   ├── error-classifier.md
│   ├── recovery-advisor.md
│   └── root-cause-analyzer.md
└── data/                        (iteration data artifacts)
    ├── s0-api-metrics.yaml
    ├── s0-tool-usage.jsonl
    ├── s0-capabilities-inventory.yaml
    └── s0-agents-inventory.yaml
```

---

## Key Principles

1. **Two-Layer Architecture**: Meta-agent develops methodology by observing agents execute concrete tasks ⚡ **NEW**
2. **Discovery-Driven Evolution**: Initial state discovered through union search, not predetermined
3. **Honest Value Assessment**: V(s) calculated from actual observations, not target values
4. **Modular Meta-Agent**: Separate capability files, not versioned monoliths
5. **Agent Prompt Files**: Every agent has explicit prompt file, read before invocation
6. **No Token Limits**: Complete all analysis steps thoroughly without abbreviation
7. **Needs-Driven Specialization**: Create specialized agents only when generic agents insufficient
8. **Concrete Instance Work**: Agents work on real API code, not methodology documents ⚡ **NEW**

---

## Convergence Criteria

Experiment concludes when ALL criteria met:

1. **Meta-Agent Stable**: M_n = M_{n-1} (no new capabilities)
2. **Agent Set Stable**: A_n = A_{n-1} (no new agents created)
3. **Value Threshold**: V(s_n) ≥ 0.80 (target quality achieved)
4. **Objectives Complete**: All API design methodology tasks finished
5. **Diminishing Returns**: ΔV < 0.05 (minimal improvement per iteration)

**Current Status**: NOT CONVERGED - Iterations 0-3 completed but need architectural correction

---

## Path Forward (Post-Restructuring)

### Analysis of Iterations 0-3

**What Was Done** (Design-focused approach):
- Iteration 0: Baseline establishment, API data collection
- Iteration 1: Evolvability design (versioning strategy, deprecation policy)
- Iteration 2: Consistency design (naming conventions, parameter ordering guidelines)
- Iteration 3: Implementation design (parameter reordering specs, validation tool specs, documentation specs)

**What Was Missing** (Concrete API improvements):
- Agents created **specifications** rather than **implementing** API improvements
- No concrete code changes to `cmd/mcp/tools.go`
- No actual parameter reordering, error message improvements, or documentation enhancements
- Meta-agent had no **concrete agent work** to observe and extract patterns from

**Value of Iterations 0-3**:
- ✅ Valuable design artifacts created (guidelines, conventions, specifications)
- ✅ V(s₃) = 0.80 achieved through **design quality**
- ❌ Missing: Concrete implementation for meta-agent to observe
- ❌ Missing: Actual API improvements to meta-cc MCP server

### Recommended Path Forward

**Option 1: Continue with Iteration 4 (Recommended)**

Execute Iteration 4 with **two-layer architecture**:

**Iteration 4 Objectives**:

**Meta-Agent Work**:
- Observe how agents execute concrete API improvements
- Extract patterns from agent work (e.g., how agents identify parameter tier, how agents improve error messages)
- Codify observations into methodology artifacts (API-DESIGN-METHODOLOGY.md)
- Evaluate if specialized agents are needed

**Agent Work** (Concrete Tasks):
1. **Task 1**: Implement parameter reordering for 3 tools (query_tools, query_user_messages, query_conversation)
   - Edit `cmd/mcp/tools.go` directly
   - Apply tier-based ordering from Iteration 2 guidelines
   - Run tests to verify non-breaking change
   - Deliverable: Working code with improved parameter ordering

2. **Task 2**: Implement validation tool MVP
   - Create `cmd/validate-api/main.go`
   - Implement 3 core checks (naming, parameter ordering, description)
   - Add tests for validation logic
   - Deliverable: Working `meta-cc validate-api` command

3. **Task 3**: Implement pre-commit hook
   - Create `.git/hooks/pre-commit` script
   - Create installation script
   - Test hook functionality
   - Deliverable: Working pre-commit hook that blocks API violations

4. **Task 4**: Enhance documentation for 3 low-usage tools
   - Update docs/guides/mcp.md with better examples for query_context, cleanup_temp_files, query_tools_advanced
   - Add usage patterns and common use cases
   - Deliverable: Improved documentation with examples

**Expected Outcome**:
- Agents produce concrete API improvements (code, tooling, docs)
- Meta-agent observes execution patterns
- Methodology emerges from observation, not predetermined design
- V(s₄) calculated based on **operational improvements**, not just design quality

**Option 2: Restart with Two-Layer Architecture (Alternative)**

If Iterations 0-3 design artifacts are insufficient:
- Archive current experiment as "bootstrap-006-api-design-v1 (design-focused)"
- Start fresh with "bootstrap-006-api-design-v2 (two-layer)"
- Iteration 0: Baseline + identify 5 concrete API improvement tasks
- Iteration 1+: Agents execute tasks, meta-agent observes and codifies

**Recommendation**: **Option 1** (continue with Iteration 4) is preferred because Iterations 1-3 created valuable design artifacts that can guide concrete implementation.

---

## Next Steps

1. **Decide**: Option 1 (continue) vs Option 2 (restart)
2. **If Option 1**: Execute Iteration 4 with two-layer architecture
   - Agents implement parameter reordering, validation tool, pre-commit hook, documentation
   - Meta-agent observes patterns and creates API-DESIGN-METHODOLOGY.md
3. **If Option 2**: Archive current experiment and restart with two-layer design
4. Calculate V(s₄) based on operational improvements
5. Continue iterations until convergence

---

**Experiment Type**: Bootstrapped Software Engineering
**Framework**: OCA + BSE + Value Space Optimization
**Expected Iterations**: 3-5 (based on bootstrap-001 historical data)
**Success Metric**: V(s_final) ≥ 0.80 with stable (M, A)
