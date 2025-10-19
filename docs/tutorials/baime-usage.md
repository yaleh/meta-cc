# BAIME Usage Guide

**BAIME (Bootstrapped AI Methodology Engineering)** - A systematic framework for developing and validating software engineering methodologies through observation, codification, and automation.

---

## Table of Contents

- [What is BAIME?](#what-is-baime)
- [When to Use BAIME](#when-to-use-baime)
- [Prerequisites](#prerequisites)
- [Core Concepts](#core-concepts)
- [Frequently Asked Questions](#frequently-asked-questions)
- [Quick Start](#quick-start)
- [Step-by-Step Workflow](#step-by-step-workflow)
- [Specialized Agents](#specialized-agents)
- [Practical Example](#practical-example)
- [Troubleshooting](#troubleshooting)
- [Next Steps](#next-steps)

---

## What is BAIME?

BAIME integrates three complementary methodologies optimized for LLM-based development:

1. **OCA Cycle** (Observe-Codify-Automate) - Core iterative framework
2. **Empirical Validation** - Scientific method and data-driven decisions
3. **Value Optimization** - Dual-layer value functions for quantitative evaluation

**Key Innovation**: BAIME treats methodology development like software development—with empirical observation, automated testing, continuous iteration, and quantitative metrics.

### Why BAIME?

**Problem**: Ad-hoc methodology development is slow, subjective, and hard to validate.

**Solution**: BAIME provides systematic approach with:
- ✅ **Rapid convergence**: Typically 3-7 iterations, 6-15 hours
- ✅ **Empirical validation**: Data-driven evidence, not opinions
- ✅ **High transferability**: 70-95% reusable across projects
- ✅ **Proven results**: 100% success rate across 8 experiments, 10-50x speedup

### BAIME in Action

**Example Results**:
- **Testing Strategy**: 15x speedup, 89% transferability
- **CI/CD Pipeline**: 2.5-3.5x speedup, 91.7% pattern validation
- **Error Recovery**: 95.4% error coverage, 3 iterations
- **Documentation System**: 47% token cost reduction, 85% reduction in redundancy
- **Knowledge Transfer**: 3-8x onboarding speedup

---

## When to Use BAIME

### Use BAIME For

✅ **Creating systematic methodologies** for:
- Testing strategies
- CI/CD pipelines
- Error handling patterns
- Observability systems
- Dependency management
- Documentation systems
- Knowledge transfer processes
- Technical debt management
- Cross-cutting concerns

✅ **When you need**:
- Empirical validation with data
- Iterative methodology evolution
- Quantitative quality metrics
- Transferable best practices
- Rapid convergence (hours to days, not weeks)

### Don't Use BAIME For

❌ **One-time ad-hoc tasks** without reusability goals
❌ **Trivial processes** (<100 lines of code/docs)
❌ **Established standards** that fully solve your problem

---

## Prerequisites

### Required

1. **meta-cc plugin installed** and configured
   - See [Installation Guide](installation.md)
   - Verify: `/meta "show stats"` works

2. **Claude Code** environment
   - Access to Task tool for subagent invocation

3. **Project with need for methodology**
   - Have a specific domain in mind (testing, CI/CD, etc.)
   - Able to measure current state (baseline)

### Recommended

- **Familiarity with meta-cc** basic features
- **Understanding of your domain** (e.g., if developing testing methodology, know testing basics)
- **Git repository** for tracking methodology evolution

---

## Core Concepts

### Understanding Value Functions

BAIME uses **dual-layer value functions** to measure quality at two independent levels:

#### V_instance: Domain-Specific Quality

Measures the quality of your specific deliverables:

- **Purpose**: Assess whether your domain work is high-quality
- **Examples**:
  - Testing methodology: Test coverage percentage, test maintainability
  - CI/CD pipeline: Build time, deployment success rate, quality gate coverage
  - Documentation: Completeness, accuracy, usability
- **Characteristics**: Domain-dependent, specific to your work

#### V_meta: Methodology Quality

Measures the quality of the methodology itself:

- **Purpose**: Assess whether your methodology is reusable and effective
- **Components**:
  - **Completeness**: All necessary patterns, templates, tools exist
  - **Effectiveness**: Methodology improves quality and efficiency
  - **Reusability**: Works across projects with minimal adaptation
  - **Validation**: Empirically tested and proven effective
- **Characteristics**: Domain-independent, universal assessment

#### Convergence Requirement

**Both must reach ≥ 0.80** for methodology to be complete:

- V_instance ≥ 0.80: Domain work is production-ready
- V_meta ≥ 0.80: Methodology is reusable
- If only one converges, keep iterating

---

### The OCA Cycle

Each iteration follows the **Observe-Codify-Automate** cycle:

```
Observe → Codify → Automate → Evaluate
   ↓                              ↓
   ← ← ← ← ← Iterate ← ← ← ← ← ←
```

#### Phase 1: Observe

**Goal**: Collect empirical data about current state

**Activities**:
- Read previous iteration results
- Measure baseline (Iteration 0) or current state
- Identify problems and patterns
- Gather evidence about what's working/not working

**Output**: Data artifacts documenting observations

#### Phase 2: Codify

**Goal**: Extract patterns and create reusable structures

**Activities**:
- Form strategy based on evidence
- Extract recurring patterns into documented forms
- Create templates for common structures
- Prioritize improvements based on impact

**Output**: Patterns, templates, strategy documentation

#### Phase 3: Automate

**Goal**: Build tools to improve efficiency and consistency

**Activities**:
- Create automation scripts (validators, generators, analyzers)
- Implement quality gates
- Build CI integration
- Execute planned improvements

**Output**: Working tools, improved deliverables

#### Phase 4: Evaluate

**Goal**: Measure progress and assess convergence

**Activities**:
- Calculate V_instance and V_meta scores
- Provide evidence for each component
- Identify remaining gaps
- Check convergence criteria

**Output**: Value scores, gap analysis, convergence decision

---

### Meta-Agent and Specialized Agents

#### Meta-Agent

The **meta-agent orchestrates** the entire BAIME process:

**Responsibilities**:
- Read lifecycle capabilities before each phase (fresh, no caching)
- Execute OCA cycle systematically
- Track system state evolution (M_n, A_n, s_n)
- Coordinate specialized agents when needed
- Make evidence-based evolution decisions

**Key Behavior**: Reads capabilities fresh each iteration to incorporate latest guidance

#### Specialized Agents

**Domain-specific executors** created when evidence shows need:

**When created**:
- Generic approach insufficient (demonstrated, not assumed)
- Task recurs 3+ times with similar structure
- Clear expected improvement from specialization

**Examples**:
- `test-generator`: Creates tests following validated patterns
- `validator-agent`: Checks deliverables against quality criteria
- `knowledge-extractor`: Transforms experiment into reusable methodology

**Key Principle**: Agents evolve based on retrospective evidence (not anticipatory design)

---

### Capabilities and System State

#### Capabilities

**Modular guidance files** for each OCA lifecycle phase:

- `capabilities/collect.md` - Data collection patterns
- `capabilities/strategy.md` - Strategy formation guidance
- `capabilities/execute.md` - Execution patterns
- `capabilities/evaluate.md` - Evaluation rubrics
- `capabilities/converge.md` - Convergence assessment

**Evolution**:
- Start empty (placeholders) in Iteration 0
- Evolve when patterns recur 2-3 times
- Based on retrospective evidence (not speculation)
- Read fresh each phase (no caching)

#### System State

**Tracked components** across iterations:

- **M_n**: Methodology components (capabilities, patterns, templates)
- **A_n**: Agent system (specialized agents)
- **s_n**: Current state (deliverables, artifacts, value scores)
- **V(s_n)**: Dual value functions (V_instance, V_meta)

**State transition**: s_{n-1} → s_n documents evolution

---

### Convergence Criteria

Methodology is **complete and production-ready** when all four conditions met:

#### 1. Dual Threshold

- ✅ V_instance ≥ 0.80 (domain goals achieved)
- ✅ V_meta ≥ 0.80 (methodology quality high)

#### 2. System Stability

- ✅ M_n == M_{n-1} (no methodology changes)
- ✅ A_n == A_{n-1} (no agent evolution)
- ✅ Stable for 2+ consecutive iterations

#### 3. Objectives Complete

- ✅ All planned work finished
- ✅ No critical gaps remaining

#### 4. Diminishing Returns

- ✅ ΔV_instance < 0.02 for 2+ iterations
- ✅ ΔV_meta < 0.02 for 2+ iterations

**Note**: If system evolves (new agent/capability), stability clock resets. Evolution must be validated in next iteration before convergence.

---

## Frequently Asked Questions

### General Questions

#### What exactly is BAIME and how is it different from other methodologies?

BAIME (Bootstrapped AI Methodology Engineering) is a meta-methodology for developing domain-specific methodologies through empirical observation and iteration. Unlike traditional methodologies that are designed upfront, BAIME creates methodologies through practice:

- **Traditional approach**: Design methodology → Apply → Hope it works
- **BAIME approach**: Observe patterns → Extract methodology → Validate → Iterate

Key differentiators:
- Dual-layer value functions measure both deliverable quality AND methodology quality
- Evidence-driven evolution (not anticipatory design)
- Quantitative convergence criteria (≥0.80 thresholds)
- Specialized subagents for consistent execution

#### When should I use BAIME vs just following existing best practices?

**Use BAIME when**:
- No established methodology fully fits your domain
- You need methodology customized to your project constraints
- You want empirically validated patterns, not borrowed practices
- You need to measure and prove methodology effectiveness

**Use existing practices when**:
- Industry-standard methodology already solves your problem
- Team already trained on established framework
- Project timeline doesn't allow methodology development
- Problem domain is simple and well-understood

**Use both**: Start with BAIME to develop baseline, then integrate proven external practices in later iterations.

#### How long does a typical BAIME experiment take?

**Typical timeline**:
- **Iteration 0** (Baseline): 2-4 hours
- **Iterations 1-N**: 3-6 hours each
- **Total**: 10-30 hours over 3-7 iterations
- **Knowledge extraction**: 2-4 hours post-convergence

**Time factors**:
- Domain complexity (testing < CI/CD < architecture)
- Baseline quality (higher baseline → fewer iterations)
- Team familiarity with BAIME (improves with practice)
- Automation investment (upfront cost, ongoing savings)

**ROI**: 10-50x speedup on future work justifies investment. A 20-hour methodology development that saves 10 hours per month pays off in month 2.

#### What if my value scores aren't improving between iterations?

**Diagnostic steps**:

1. **Check if addressing root problems**:
   - Review problem identification from previous iteration
   - Are you solving symptoms vs causes?
   - Example: Low test coverage may be due to unclear testing strategy, not lack of tests

2. **Verify evidence quality**:
   - Is data collection comprehensive?
   - Are you making evidence-based decisions?
   - Review data artifacts - do they support your strategy?

3. **Assess scope**:
   - Trying to fix too many things?
   - Focus on top 2-3 highest-impact problems
   - Better to solve 2 problems well than 5 problems poorly

4. **Challenge your scoring**:
   - Are scores honest (vs inflated)?
   - Seek disconfirming evidence
   - Compare against rubric, not "could be worse"

5. **Consider system evolution**:
   - Do you need specialized agent for recurring complex task?
   - Would new capability help structure repeated work?
   - Evolution requires evidence of insufficiency (not speculation)

**If still stuck after 2-3 iterations**: Re-examine value function definitions. May need to adjust components or convergence targets.

### Usage Questions

#### Can I use BAIME for [specific domain]?

BAIME works for **any software engineering domain where**:
- ✅ You can measure quality objectively
- ✅ Patterns emerge from practice
- ✅ Work involves 100+ lines of code/docs
- ✅ Results will be reused (methodology has value)

**Proven domains** (8 successful experiments):
- Testing strategy
- CI/CD pipelines
- Error recovery
- Observability instrumentation
- Dependency management
- Documentation systems
- Knowledge transfer
- Technical debt management

**Untested but promising**:
- API design
- Database migration
- Performance optimization
- Security review processes
- Code review workflows

**Probably not suitable**:
- One-time tasks (no reusability)
- Trivial processes (<1 hour total work)
- Domains with perfect existing solutions

#### Do I need the meta-cc plugin to use BAIME?

**For full BAIME workflow**: Yes, meta-cc provides:
- Session history analysis (understanding past work)
- MCP tools for querying patterns
- Specialized subagents (iteration-executor, knowledge-extractor)
- `/meta` command for quick insights

**Without meta-cc**: You can still apply BAIME principles:
- Manual OCA cycle execution
- Self-tracked value functions
- Evidence collection through notes/logs
- Pattern extraction through reflection

**Recommendation**: Use meta-cc. The 5-minute installation saves hours of manual tracking and provides empirical data for better decisions.

#### How do I know when to create a specialized agent?

**Create specialized agent when** (all three conditions):

1. **Evidence of insufficiency**:
   - Generic approach tried and struggled
   - Task complexity consistently high
   - Errors or quality issues recurring

2. **Pattern recurrence**:
   - Task performed 3+ times across iterations
   - Similar structure each time
   - Clear enough to codify

3. **Expected improvement**:
   - Can articulate what agent will do better
   - Have evidence from past attempts
   - Benefit justifies creation cost

**Don't create agent when**:
- Task only done 1-2 times (insufficient evidence)
- Generic approach working fine
- Speculation about future need (wait for evidence)

**Example**: In testing methodology, created `test-generator` agent after:
- Iteration 0-1: Manually wrote tests (worked but slow)
- Iteration 2: Pattern clear (fixture → arrange → act → assert)
- Iteration 3: Created agent, 3x speedup validated

### Technical Questions

#### What's the difference between capabilities and agents?

**Capabilities** (meta-agent lifecycle phases):
- **Purpose**: Guide meta-agent through OCA cycle phases
- **Content**: Patterns, guidelines, checklists for each phase
- **Location**: `capabilities/` directory (e.g., `capabilities/collect.md`)
- **Evolution**: Based on retrospective evidence (start as placeholders)
- **Example**: Strategy formation capability contains prioritization patterns

**Agents** (specialized executors):
- **Purpose**: Execute specific domain tasks
- **Content**: Domain expertise, task-specific workflows
- **Location**: `agents/` directory (e.g., `agents/test-generator.md`)
- **Evolution**: Created when evidence shows insufficiency
- **Example**: Test generator agent creates tests following patterns

**Analogy**:
- Capabilities = "How to think about the work" (meta-level)
- Agents = "How to do the work" (execution-level)

**Both**:
- Start as placeholders (empty files)
- Evolve based on evidence (not anticipatory design)
- Read fresh each time (no caching)

#### How do capabilities evolve during iterations?

**Evolution trigger**: Retrospective evidence of pattern recurrence

**Process**:

1. **Iteration 0-1**: Capabilities are placeholders (empty)
   - Meta-agent works generically
   - Patterns emerge during work

2. **Iteration 2-3**: Evidence accumulates
   - Same problems recur
   - Solutions follow similar patterns
   - Decision points become predictable

3. **Evolution point**: When pattern recurs 2-3 times
   - Extract pattern to relevant capability
   - Document guidance based on what worked
   - Add to capability file

4. **Validation**: Next iteration tests guidance
   - Does following capability improve outcomes?
   - Are value scores higher?
   - Is work more efficient?

**Example**: In CI/CD methodology:
- Iteration 0-1: Strategy capability empty
- Iteration 2: Same prioritization pattern used twice (quality gates > performance > observability)
- Iteration 2 end: Extracted to `strategy.md` capability
- Iteration 3: Following capability saved 30 minutes of decision-making

**Key principle**: Capabilities codify what worked, not what might work.

### Convergence Questions

#### Can I stop before reaching 0.80 thresholds?

**Yes, but understand trade-offs**:

**Stop at V_instance < 0.80**:
- Deliverable is incomplete or lower quality
- May need significant rework for production use
- Methodology validation is weak

**Stop at V_meta < 0.80**:
- Methodology is not fully reusable
- Transferability to other projects questionable
- May be project-specific, not universal

**When early stopping is acceptable**:
- Proof of concept (showing BAIME works for domain)
- Time constraints (better to have 0.70 than nothing)
- Sufficient for current needs (will iterate later)
- Learning exercise (not production use)

**When to push for full convergence**:
- Production deliverable needed
- Methodology will be shared/reused
- Investment in convergence pays off quickly
- Demonstrating BAIME effectiveness

**Recommendation**: Aim for dual convergence. The final iterations often provide the highest-value insights.

#### What if iterations take longer than estimated?

**Common in early BAIME use**:
- First experiment: 20-40 hours (learning BAIME itself)
- Second experiment: 15-25 hours (familiar with process)
- Third+ experiment: 10-20 hours (efficient execution)

**Time optimization strategies**:

1. **Invest in baseline** (Iteration 0):
   - 3-4 hours in Iteration 0 can save 6+ hours overall
   - Higher V_meta_0 (≥0.40) enables rapid convergence

2. **Use specialized subagents**:
   - iteration-executor saves 1-2 hours per iteration
   - knowledge-extractor saves 4-6 hours post-convergence

3. **Time-box template creation**:
   - Set 1.5 hour limit per template
   - Quality over quantity (3 excellent > 5 mediocre)

4. **Batch similar work**:
   - Create all templates together (context switching cost)
   - Run all automation tools together (testing efficiency)

5. **Defer low-ROI items**:
   - Visual aids can wait (2 hours for +0.03 impact)
   - Second example if first validates pattern

**If consistently over time**: Review your value function definitions. May be too ambitious for domain complexity.

---

## Quick Start

### 1. Define Your Domain

Choose the methodology you want to develop:

```
Examples:
- "Develop systematic testing strategy for Go projects"
- "Create CI/CD pipeline methodology with quality gates"
- "Build error recovery patterns for web services"
- "Establish documentation management system"
```

### 2. Establish Baseline

Measure current state in your domain:

```bash
# Example: Testing domain
- Current coverage: 65%
- Test approach: Ad-hoc
- No systematic patterns
- Estimated effort: High

# Example: CI/CD domain
- Build time: 5 minutes
- No quality gates
- Manual releases
- No smoke tests
```

### 3. Set Dual Goals

Define objectives for both layers:

**Instance Goal** (domain-specific):
- "Reach 80% test coverage with systematic strategy"
- "Reduce CI/CD build time to <2 minutes with quality gates"

**Meta Goal** (methodology):
- "Create reusable testing strategy with 85%+ transferability"
- "Develop CI/CD methodology applicable to any Go project"

### 4. Create Experiment Structure

```bash
# Create experiment directory
mkdir -p experiments/my-methodology

# Use iteration-prompt-designer subagent
# (See Specialized Agents section below)
```

### 5. Start Iteration 0

Execute baseline iteration using iteration-executor subagent.

---

## Step-by-Step Workflow

### Phase 0: Experiment Setup

**Goal**: Create experiment structure and iteration prompts

**Steps**:

1. **Create experiment directory**:
   ```bash
   cd your-project
   mkdir -p experiments/my-methodology-name
   cd experiments/my-methodology-name
   ```

2. **Design iteration prompts** (use iteration-prompt-designer subagent):
   ```
   User: "Design ITERATION-PROMPTS.md for [domain] methodology experiment"

   Agent creates:
   - ITERATION-PROMPTS.md (comprehensive iteration guidance)
   - Architecture overview (meta-agent + agents)
   - Value function definitions
   - Baseline iteration steps
   ```

3. **Review and customize**:
   - Adjust value function components for your domain
   - Customize baseline iteration steps
   - Set convergence targets

**Output**: `ITERATION-PROMPTS.md` ready for execution

---

### Phase 1: Iteration 0 (Baseline)

**Goal**: Establish baseline measurements and initial system state

**Steps**:

1. **Execute iteration** (use iteration-executor subagent):
   ```
   User: "Execute Iteration 0 for [domain] methodology using iteration-executor"
   ```

2. **Iteration-executor will**:
   - Create modular architecture (capabilities, agents, system state)
   - Collect baseline data
   - Create first deliverables (low quality expected)
   - Calculate V_instance_0 and V_meta_0 (honest assessment)
   - Identify problems and gaps
   - Generate iteration-0.md documentation

3. **Review baseline results**:
   ```bash
   # Check value scores
   cat system-state.md

   # Review iteration documentation
   cat iteration-0.md

   # Check identified problems
   grep "Problems" system-state.md
   ```

**Expected Baseline**: V_instance: 0.20-0.40, V_meta: 0.15-0.30

**Key Principle**: Low scores are expected and acceptable. This is measurement baseline, not final product.

---

### Phase 2: Iterations 1-N (Evolution)

**Goal**: Iteratively improve both deliverables and methodology until convergence

**For Each Iteration**:

1. **Read system state**:
   ```bash
   cat system-state.md  # Current scores and problems
   cat iteration-log.md # Iteration history
   ```

2. **Execute iteration** (use iteration-executor):
   ```
   User: "Execute Iteration N for [domain] methodology using iteration-executor"
   ```

3. **Iteration-executor follows OCA cycle**:

   **Observe**:
   - Read all capabilities for methodology context
   - Collect data on prioritized problems
   - Gather evidence about current state

   **Codify**:
   - Form strategy based on evidence
   - Plan specific improvements
   - Set iteration targets

   **Execute**:
   - Create/improve deliverables
   - Apply methodology patterns
   - Document execution observations

   **Evaluate**:
   - Calculate V_instance_N and V_meta_N
   - Provide evidence for each score component
   - Identify remaining gaps

   **Converge**:
   - Check convergence criteria
   - Extract patterns (if evidence supports)
   - Update capabilities (if retrospective evidence shows gaps)
   - Prioritize next iteration focus

4. **Review iteration results**:
   ```bash
   cat iteration-N.md      # Complete iteration documentation
   cat system-state.md     # Updated scores and state
   cat iteration-log.md    # Updated history
   ```

5. **Check convergence**:
   - V_instance ≥ 0.80?
   - V_meta ≥ 0.80?
   - Both stable for 2+ iterations?
   - If YES → Converged! Move to Phase 3
   - If NO → Continue to next iteration

**Typical Iteration Count**: 3-7 iterations to convergence

---

### Phase 3: Knowledge Extraction (Post-Convergence)

**Goal**: Transform experiment artifacts into reusable methodology

**Steps**:

1. **Use knowledge-extractor subagent**:
   ```
   User: "Extract methodology from [domain] experiment using knowledge-extractor"
   ```

2. **Knowledge-extractor creates**:
   - Methodology guide (comprehensive documentation)
   - Pattern library (extracted patterns)
   - Template collection (reusable templates)
   - Automation tools (scripts, validators)
   - Best practices (principles discovered)

3. **Package as skill** (optional):
   ```bash
   # Create skill structure
   mkdir -p .claude/skills/my-methodology

   # Copy extracted knowledge
   cp -r patterns templates .claude/skills/my-methodology/

   # Create SKILL.md
   # (See knowledge-extractor output for template)
   ```

**Output**: Reusable methodology ready for other projects

---

## Specialized Agents

BAIME provides three specialized Claude Code subagents:

### iteration-prompt-designer

**Purpose**: Design comprehensive ITERATION-PROMPTS.md for your experiment

**When to use**: At experiment start, before Iteration 0

**Invocation**:
```
Use Task tool with subagent_type="iteration-prompt-designer"

Example:
"Design ITERATION-PROMPTS.md for CI/CD optimization methodology experiment"
```

**What it creates**:
- Modular meta-agent architecture definition
- Domain-specific value function design
- Baseline iteration (Iteration 0) detailed steps
- Subsequent iteration templates
- Evidence-driven evolution guidance

**Time saved**: 2-3 hours of setup work

---

### iteration-executor

**Purpose**: Execute iteration through complete OCA cycle

**When to use**: For each iteration (Iteration 0, 1, 2, ...)

**Invocation**:
```
Use Task tool with subagent_type="iteration-executor"

Example:
"Execute Iteration 2 of testing methodology experiment using iteration-executor"
```

**What it does**:
1. Reads previous iteration state
2. Reads all capability files (fresh, no caching)
3. Executes lifecycle phases:
   - Data Collection (Observe)
   - Strategy Formation (Codify)
   - Work Execution (Automate)
   - Evaluation (Calculate dual values)
   - Convergence Check (Assess progress)
4. Generates complete iteration-N.md documentation
5. Updates system-state.md and iteration-log.md

**Benefits**:
- ✅ Consistent iteration structure
- ✅ Systematic value calculation (reduces bias)
- ✅ Proper convergence evaluation
- ✅ Complete artifact generation
- ✅ Structured execution vs ad-hoc

---

### knowledge-extractor

**Purpose**: Extract and transform converged experiment into reusable methodology

**When to use**: After experiment converges

**Invocation**:
```
Use Task tool with subagent_type="knowledge-extractor"

Example:
"Extract methodology from documentation-management experiment using knowledge-extractor"
```

**What it creates**:
- Methodology guide (user-facing documentation)
- Pattern library (validated patterns)
- Template collection (reusable templates)
- Automation tools (scripts, validators)
- Best practices guide (principles)
- Skill package (optional .claude/skills/ structure)

**Time saved**: 4-6 hours of knowledge organization work

---

## Practical Example

### Example: Developing Testing Methodology

**Domain**: Systematic testing strategy for Go projects

#### Step 1: Setup

```bash
# Create experiment
mkdir -p experiments/testing-methodology
cd experiments/testing-methodology

# Design iteration prompts
# (Use iteration-prompt-designer subagent)
```

Result: `ITERATION-PROMPTS.md` created with:
- Value functions for testing (coverage, quality, maintainability)
- Baseline iteration steps
- Testing-specific guidance

#### Step 2: Iteration 0 (Baseline)

```
User: "Execute Iteration 0 of testing methodology using iteration-executor"
```

**What happens**:

1. **Architecture created**:
   ```
   testing-methodology/
   ├── capabilities/
   │   ├── test-collect.md      (placeholder)
   │   ├── test-strategy.md     (placeholder)
   │   ├── test-execute.md      (placeholder)
   │   ├── test-evaluate.md     (placeholder)
   │   └── test-converge.md     (placeholder)
   ├── agents/
   │   ├── test-generator.md    (placeholder)
   │   └── test-validator.md    (placeholder)
   ├── data/
   ├── patterns/
   ├── templates/
   ├── system-state.md
   ├── iteration-log.md
   └── knowledge-index.md
   ```

2. **Data collected**:
   ```
   data/current-testing-state.md:
   - Current coverage: 65%
   - Test approach: Ad-hoc unit tests
   - No integration test strategy
   - No TDD workflow
   ```

3. **First deliverable created**:
   ```
   # Example: Basic test helper function
   # Quality: Low (intentionally, for baseline)
   ```

4. **Baseline scores calculated**:
   ```
   V_instance_0: 0.35
   - Coverage: 0.40 (65% actual, target 80%)
   - Quality: 0.25 (ad-hoc, no systematic approach)
   - Maintainability: 0.40 (some organization)

   V_meta_0: 0.25
   - Completeness: 0.20 (capabilities empty)
   - Effectiveness: 0.30 (no proven patterns yet)
   - Reusability: 0.20 (project-specific so far)
   - Validation: 0.30 (baseline measurement only)
   ```

5. **Problems identified**:
   - No TDD workflow
   - Coverage gaps unknown
   - Test organization unclear
   - No fixture patterns

**Output**: `iteration-0.md` with complete baseline documentation

#### Step 3: Iteration 1 (First Improvement)

```
User: "Execute Iteration 1 of testing methodology using iteration-executor"
```

**Focused on**: TDD workflow and coverage analysis

**Results**:
- Created TDD workflow pattern
- Built coverage gap analyzer tool
- Improved test organization
- V_instance_1: 0.55 (+0.20)
- V_meta_1: 0.45 (+0.20)

#### Step 4: Iterations 2-3 (Evolution)

Continued iterations until:
- V_instance_3: 0.85
- V_meta_3: 0.83
- Both stable (no major changes in iteration 4)

**Convergence achieved!**

#### Step 5: Knowledge Extraction

```
User: "Extract methodology from testing-methodology experiment using knowledge-extractor"
```

**Created**:
- `methodology/testing-strategy.md` (comprehensive guide)
- 8 validated patterns
- 3 reusable templates
- Coverage analyzer tool
- Test generator script

**Result**: Reusable testing methodology ready for other Go projects

---

### Example 2: Developing Error Recovery Methodology

**Domain**: Systematic error handling and recovery patterns for software systems

**Why This Example**: Demonstrates BAIME applicability to a different domain (error handling vs testing), showing methodology transferability and universal OCA cycle pattern.

#### Step 1: Setup

```bash
# Create experiment
mkdir -p experiments/error-recovery
cd experiments/error-recovery

# Design iteration prompts
# (Use iteration-prompt-designer subagent)
```

Result: `ITERATION-PROMPTS.md` created with:
- Value functions for error recovery (coverage, diagnostic quality, recovery effectiveness)
- Error taxonomy definition
- Recovery pattern identification

#### Step 2: Iteration 0 (Baseline)

```
User: "Execute Iteration 0 of error-recovery methodology using iteration-executor"
```

**What happens**:

1. **Architecture created**:
   ```
   error-recovery/
   ├── capabilities/
   │   ├── error-collect.md       (placeholder)
   │   ├── error-strategy.md      (placeholder)
   │   ├── error-execute.md       (placeholder)
   │   ├── error-evaluate.md      (placeholder)
   │   └── error-converge.md      (placeholder)
   ├── agents/
   │   ├── error-analyzer.md      (placeholder)
   │   └── error-classifier.md    (placeholder)
   ├── data/
   ├── patterns/
   ├── templates/
   ├── system-state.md
   ├── iteration-log.md
   └── knowledge-index.md
   ```

2. **Data collected**:
   ```
   data/error-analysis.md:
   - Historical errors: 1,336 instances analyzed
   - Error handling: Ad-hoc, inconsistent
   - Recovery patterns: None documented
   - MTTD/MTTR: High, no systematic diagnosis
   ```

3. **First deliverable created**:
   ```
   # Initial error taxonomy (13 categories)
   # Quality: Basic classification, no recovery patterns yet
   ```

4. **Baseline scores calculated**:
   ```
   V_instance_0: 0.40
   - Coverage: 0.50 (errors classified, not all types covered)
   - Diagnostic Quality: 0.30 (basic categorization only)
   - Recovery Effectiveness: 0.25 (no systematic recovery)
   - Documentation: 0.55 (taxonomy exists)

   V_meta_0: 0.30
   - Completeness: 0.25 (taxonomy only, no workflows)
   - Effectiveness: 0.35 (classification helpful but limited)
   - Reusability: 0.25 (domain-specific so far)
   - Validation: 0.35 (validated against 1,336 historical errors)
   ```

5. **Problems identified**:
   - No systematic diagnosis workflow
   - No recovery patterns
   - No prevention guidelines
   - Taxonomy incomplete (95.4% coverage, gaps exist)

**Output**: `iteration-0.md` with complete baseline documentation

**Key Difference from Testing Example**: Error Recovery started with rich historical data (1,336 errors), enabling retrospective validation from Iteration 0. This demonstrates how domain characteristics affect baseline quality (V_instance_0 = 0.40 vs Testing's 0.35).

#### Step 3: Iteration 1 (Diagnostic Workflows)

```
User: "Execute Iteration 1 of error-recovery methodology using iteration-executor"
```

**Focused on**: Creating diagnostic workflows and expanding taxonomy

**Results**:
- Created 8 diagnostic workflows (file operations, API calls, data validation, etc.)
- Expanded error taxonomy to 13 categories
- Added contextual logging patterns
- **V_instance_1: 0.62** (+0.22, significant jump due to workflow addition)
- **V_meta_1: 0.50** (+0.20, patterns emerging)

**Pattern Emerged**: Error diagnosis follows consistent structure:
1. Symptom identification
2. Context gathering
3. Root cause analysis
4. Solution selection

#### Step 4: Iteration 2 (Recovery Patterns and Prevention)

```
User: "Execute Iteration 2 of error-recovery methodology using iteration-executor"
```

**Focused on**: Recovery patterns, prevention guidelines, automation

**Results**:
- Documented 5 recovery patterns (retry, fallback, circuit breaker, graceful degradation, fail-fast)
- Created 8 prevention guidelines
- Built 3 automation tools (file path validation, read-before-write check, file size validation)
- **V_instance_2: 0.78** (+0.16, approaching convergence)
- **V_meta_2: 0.72** (+0.22, acceleration due to automation)

**Automation Impact**: Prevention tools covered 23.7% of historical errors, proving methodology effectiveness empirically.

#### Step 5: Iteration 3 (Convergence)

```
User: "Execute Iteration 3 of error-recovery methodology using iteration-executor"
```

**Focused on**: Final validation, cross-language transferability

**Results**:
- Validated patterns across 4 languages (Go, Python, JavaScript, Rust)
- Achieved 95.4% error coverage (1,274/1,336 historical errors)
- Transferability assessment: 85-90% universal patterns
- **V_instance_3: 0.83** (+0.05, exceeded threshold)
- **V_meta_3: 0.85** (+0.13, strong convergence)

**System Stability**: No capability or agent evolution needed (3 iterations stable) - generic OCA cycle sufficient.

**Convergence Status**: ✅ **CONVERGED**
- Both layers > 0.80 ✅
- System stable (M_3 == M_2, A_3 == A_2) ✅
- Objectives complete ✅
- Total time: ~10 hours over 3 iterations

#### Step 6: Knowledge Extraction

```
User: "Extract methodology from error-recovery experiment using knowledge-extractor"
```

**Created**:
- `methodology/error-recovery.md` (comprehensive 13-category taxonomy)
- 8 diagnostic workflows
- 5 recovery patterns
- 8 prevention guidelines
- 3 automation tools (file validation, read-before-write, size validation)
- Retrospective validation report (95.4% historical error coverage)

**Result**: Reusable error recovery methodology with 85-90% transferability across languages/platforms

**Transferability Evidence**:
- Core concepts: 100% universal (error taxonomy, diagnostic workflows)
- Recovery patterns: 95% universal (retry, fallback, circuit breaker work everywhere)
- Automation tools: 60% universal (concepts transfer, implementations vary by language)

---

### Comparing the Two Examples

| Aspect | Testing Methodology | Error Recovery Methodology |
|--------|---------------------|----------------------------|
| **Domain Complexity** | Medium (test strategies, patterns) | High (13 error categories, recovery patterns) |
| **Baseline Data** | Limited (current tests only) | Rich (1,336 historical errors) |
| **V_instance_0** | 0.35 | 0.40 (higher due to historical data) |
| **V_meta_0** | 0.25 | 0.30 (retrospective validation possible) |
| **Iterations to Converge** | 3-4 iterations | 3 iterations (rapid due to data richness) |
| **Total Time** | ~12 hours | ~10 hours (rich baseline enabled efficiency) |
| **Transferability** | 89% (Go projects) | 85-90% (universal, cross-language) |
| **Key Innovation** | TDD workflow, coverage analyzer | Error taxonomy, diagnostic workflows, prevention |
| **System Evolution** | Stable (no agent specialization) | Stable (no agent specialization) |

**Universal Lessons**:
1. **Rich baseline data accelerates convergence** (Error Recovery's 1,336 errors vs Testing's current state)
2. **OCA cycle works across domains** (same structure, different content)
3. **System stability is common** (both examples: no agent evolution needed)
4. **Retrospective validation powerful** (Error Recovery: 95.4% coverage proves methodology)
5. **Automation provides empirical evidence** (23.7% error prevention measurable)

**BAIME Transferability Confirmed**: Same methodology framework produced high-quality results in two distinct domains (testing vs error handling), demonstrating universal applicability.

---

## Troubleshooting

### Issue: Value scores not improving

**Symptoms**: V_instance or V_meta stuck or decreasing across iterations

**Example**:
```
Iteration 0: V_instance = 0.35, V_meta = 0.25
Iteration 1: V_instance = 0.37, V_meta = 0.28  (minimal progress)
Iteration 2: V_instance = 0.34, V_meta = 0.30  (instance decreased!)
```

**Diagnosis**:

**Root Cause 1: Solving symptoms, not problems**
```
❌ Problem identified: "Low test coverage"
❌ Solution attempted: "Write more tests"
❌ Result: Coverage increased but tests are brittle, hard to maintain

✅ Better problem: "No systematic testing strategy"
✅ Better solution: "Create TDD workflow, extract test patterns"
✅ Result: Fewer tests, but higher quality and maintainable
```

**Root Cause 2: Strategy not evidence-based**
```
❌ Strategy: "Let's add integration tests because they seem useful"
❌ Evidence: None (speculation)

✅ Strategy: "Data shows 80% of bugs in API layer, add API tests"
✅ Evidence: Bug analysis from data/bug-analysis.md
```

**Root Cause 3: Scope too broad**
```
❌ Iteration 2 plan: Fix 7 problems (test coverage, CI/CD, docs, errors)
❌ Result: All partially done, none well done

✅ Iteration 2 plan: Fix top 2 problems (test strategy, coverage analysis)
✅ Result: Both fully solved, measurable improvement
```

**Solutions**:
1. **Re-examine problem identification**:
   - Are you solving root causes or symptoms?
   - Review data artifacts - do they support your problem statement?
   - Ask "why" 3 times to find root cause

2. **Verify evidence quality**:
   - Is data collection comprehensive?
   - Do you have concrete measurements?
   - Can you show before/after data?

3. **Narrow focus**:
   - Address top 2-3 highest-impact problems only
   - Better to solve 2 problems completely than 5 partially
   - Defer lower-priority items to next iteration

4. **Re-evaluate strategy**:
   - Is it based on data or assumptions?
   - Review iteration-N-strategy.md for evidence
   - Challenge each planned improvement: "What evidence supports this?"

---

### Issue: Methodology not transferable (low V_meta Reusability)

**Symptoms**: V_meta Reusability component < 0.60 after multiple iterations

**Example**:
```
Iteration 2 evaluation:
- Completeness: 0.70 ✅
- Effectiveness: 0.75 ✅
- Reusability: 0.45 ❌ (blocking convergence)
- Validation: 0.65 ✅
```

**Diagnosis**:

**Problem: Patterns too project-specific**

Before (Low Reusability):
```markdown
# Testing Pattern
1. Create test file in src/api/handlers/__tests__/
2. Import UserModel from "../../models/user"
3. Use Jest expect() assertions
4. Run with npm test
```

After (High Reusability):
```markdown
# Testing Pattern (Parameterized)
1. Create test file adjacent to source: {source_dir}/__tests__/{module}_test{ext}
2. Import module under test: {import_statement}
3. Use test framework assertion: {assertion_method}
4. Run with project test command: {test_command}

Adaptation guide:
- Go: {ext}=.go, {assertion_method}=testing.T methods
- JS: {ext}=.js, {assertion_method}=expect() or assert()
- Python: {ext}=.py, {assertion_method}=unittest assertions
```

**Problem: No abstraction of domain concepts**

Before:
```markdown
# CI/CD Pattern
- Install Go 1.21
- Run go test ./...
- Build with go build -o bin/app
- Check coverage is >80%
```

After (Abstracted):
```markdown
# CI/CD Quality Gate Pattern

Universal steps:
1. Install language runtime (version from project config)
2. Run test suite (project-specific command)
3. Build artifact (project-specific build process)
4. Verify quality threshold (configurable threshold)

Domain-specific implementations:
- Go: {runtime}=Go 1.21+, {test}=go test, {build}=go build
- Node: {runtime}=Node 18+, {test}=npm test, {build}=npm run build
- Python: {runtime}=Python 3.10+, {test}=pytest, {build}=python setup.py
```

**Solutions**:
1. **Extract universal patterns**:
   - Identify what's essential vs project-specific
   - Replace hardcoded values with parameters
   - Document adaptation guide

2. **Create parameterized templates**:
   - Use placeholders: {variable_name}
   - Provide examples for 3+ different contexts
   - Include "How to adapt" section

3. **Test across scenarios**:
   - Apply pattern to different project in same domain
   - Document what needed changing
   - Refine pattern based on adaptation effort

4. **Add abstraction layers**:
   - Layer 1: Universal principle (works anywhere)
   - Layer 2: Domain-specific implementation (testing/CI/CD/etc)
   - Layer 3: Tool-specific details (Jest/pytest/etc)

---

### Issue: Can't reach convergence (stuck at V ~0.70)

**Symptoms**: Multiple iterations without reaching 0.80

**Causes**:
- Unrealistic convergence targets
- Missing critical patterns
- Need specialized agent but using generic approach

**Solutions**:
1. Review value function definitions - are they appropriate?
2. Identify missing methodology components
3. Consider creating specialized agent if problem recurs
4. Re-assess convergence criteria - is 0.80 realistic for this domain?

---

### Issue: Too many iterations (>10)

**Symptoms**: Slow convergence, many iterations needed

**Causes**:
- Insufficient baseline (V_meta_0 < 0.20)
- Not extracting patterns early enough
- Too conservative improvements

**Solutions**:
1. Improve baseline iteration - invest more time in Iteration 0
2. Extract patterns when they recur (don't wait)
3. Make bolder improvements (test larger changes)
4. Use specialized agents earlier

---

### Issue: Premature convergence claims

**Symptoms**: Claiming convergence but quality obviously low

**Causes**:
- Inflated value scores (not honest assessment)
- Comparing to "could be worse" instead of rubrics
- Time pressure leading to rushed evaluation

**Solutions**:
1. Seek disconfirming evidence actively
2. Test deliverables thoroughly
3. Enumerate gaps explicitly
4. Challenge high scores with extra scrutiny
5. Remember: Honest assessment is more valuable than fast convergence

---

## Next Steps

### After Your First BAIME Experiment

1. **Review iteration documentation** - See what worked, what didn't
2. **Extract lessons learned** - Document insights about BAIME process
3. **Apply methodology** - Use created methodology in real work
4. **Share knowledge** - Package as skill or contribute back

### Advanced Topics

- **Baseline Quality Assessment** - Achieve comprehensive baseline (V_meta ≥ 0.40 in Iteration 0) for faster convergence
- **Rapid Convergence** - Techniques for 3-4 iteration methodology development
- **Agent Specialization** - When and how to create specialized agents
- **Retrospective Validation** - Validate methodology against historical data
- **Cross-Domain Transfer** - Apply methodology to different projects

See individual skills for detailed guidance:
- `baseline-quality-assessment`
- `rapid-convergence`
- `agent-prompt-evolution`
- `retrospective-validation`

### Further Reading

- **[Methodology Bootstrapping Skill](../../.claude/skills/methodology-bootstrapping/)** - Complete BAIME reference
- **[Empirical Methodology Development](../methodology/empirical-methodology-development.md)** - Theoretical foundation
- **[Bootstrapped Software Engineering](../methodology/bootstrapped-software-engineering.md)** - BAIME in depth
- **[Example Experiments](../../experiments/)** - Real BAIME experiments to study

### Getting Help

- **Check skill documentation**: `.claude/skills/methodology-bootstrapping/`
- **Review example experiments**: `experiments/bootstrap-*/`
- **Use @meta-coach**: Ask for workflow optimization guidance
- **Read iteration documentation**: See how past experiments evolved

---

## Summary

**BAIME provides**:
- ✅ Systematic framework for methodology development
- ✅ Empirical validation with data-driven decisions
- ✅ Dual-layer value functions for quality measurement
- ✅ Specialized agents for streamlined execution
- ✅ Proven results: 10-50x speedup, 70-95% transferability

**Key workflow**:
1. Define domain and dual goals
2. Design iteration prompts (iteration-prompt-designer)
3. Execute Iteration 0 baseline (iteration-executor)
4. Iterate until convergence (typically 3-7 iterations)
5. Extract knowledge (knowledge-extractor)
6. Apply methodology to real work

**Remember**:
- Start with clear domain and goals
- Low baseline scores are expected
- Honest assessment is crucial
- Evidence-driven evolution (not anticipatory design)
- Convergence requires both V_instance ≥ 0.80 AND V_meta ≥ 0.80

**Ready to start?** Choose your domain, set up your experiment, and begin with Iteration 0!

---

**Document Version**: 1.0 (Iteration 0 Baseline)
**Last Updated**: 2025-10-19
**Status**: Initial version - Will evolve based on user feedback
