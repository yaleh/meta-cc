# Iteration Execution Prompts

This document provides structured prompts for executing each iteration of the bootstrap-007-cicd-pipeline experiment.

**Two-Layer Architecture**:
- **Instance Layer**: Agents build concrete CI/CD pipeline for meta-cc
- **Meta Layer**: Meta-Agent observes agents and extracts methodology

---

## Iteration 0: Baseline Establishment

```markdown
# Execute Iteration 0: Baseline Establishment

## Context
I'm starting the bootstrap-007-cicd-pipeline experiment. I've reviewed:
- experiments/bootstrap-007-cicd-pipeline/plan.md
- experiments/bootstrap-007-cicd-pipeline/README.md
- The three methodology frameworks (OCA, Bootstrapped SE, Value Space Optimization)

## Current State
- Meta-Agent: M₀ (6 capabilities inherited from Bootstrap-006: observe, plan, execute, reflect, evolve, api-design-orchestrator)
- Agent Set: A₀ (15 agents inherited from Bootstrap-006: 3 generic + 12 specialized)
- Project: meta-cc with manual build/release process

## Inherited State from Bootstrap-006

**IMPORTANT**: This experiment starts with the converged state from Bootstrap-006, NOT from scratch.

**Meta-Agent capability files (ALREADY EXIST)**:
- experiments/bootstrap-007-cicd-pipeline/meta-agents/observe.md (validated)
- experiments/bootstrap-007-cicd-pipeline/meta-agents/plan.md (validated)
- experiments/bootstrap-007-cicd-pipeline/meta-agents/execute.md (validated)
- experiments/bootstrap-007-cicd-pipeline/meta-agents/reflect.md (validated)
- experiments/bootstrap-007-cicd-pipeline/meta-agents/evolve.md (validated)
- experiments/bootstrap-007-cicd-pipeline/meta-agents/api-design-orchestrator.md (adaptable)

**Agent prompt files (ALREADY EXIST - 15 agents)**:
- Generic (3): data-analyst.md, doc-writer.md, coder.md
- From Bootstrap-001 (2): doc-generator.md, search-optimizer.md
- From Bootstrap-003 (3): error-classifier.md, recovery-advisor.md, root-cause-analyzer.md
- From Bootstrap-006 (7): agent-audit-executor.md, agent-documentation-enhancer.md,
  agent-parameter-categorizer.md, agent-quality-gate-installer.md, agent-schema-refactorer.md,
  agent-validation-builder.md, api-evolution-planner.md

**CRITICAL EXECUTION PROTOCOL**:
- All capability files and agent files ALREADY EXIST (inherited from Bootstrap-006)
- Before embodying Meta-Agent capabilities, ALWAYS read the relevant capability file first
- Before invoking ANY agent, ALWAYS read its prompt file first
- These files contain validated capabilities/agents; adapt them to CI/CD context
- Never assume capabilities - always read from the source files

## Iteration 0 Objectives

Execute baseline establishment:

0. **Setup** (Verify inherited state):
   - **VERIFY META-AGENT CAPABILITY FILES EXIST** (inherited from Bootstrap-006):
     - ✓ meta-agents/observe.md: Data collection, pattern discovery (validated)
     - ✓ meta-agents/plan.md: Prioritization, agent selection (validated)
     - ✓ meta-agents/execute.md: Agent coordination, task execution (validated)
     - ✓ meta-agents/reflect.md: Value calculation (V_instance, V_meta), gap analysis (validated)
     - ✓ meta-agents/evolve.md: Agent creation criteria, methodology extraction (validated)
     - ✓ meta-agents/api-design-orchestrator.md: Domain orchestration (adaptable to CI/CD)
   - **VERIFY INITIAL AGENT PROMPT FILES EXIST** (inherited from Bootstrap-006):
     - ✓ agents/data-analyst.md, doc-writer.md, coder.md (generic agents)
     - ✓ agents/doc-generator.md, search-optimizer.md (from Bootstrap-001)
     - ✓ agents/error-classifier.md, recovery-advisor.md, root-cause-analyzer.md (from Bootstrap-003)
     - ✓ agents/agent-*.md (7 agents from Bootstrap-006)
   - **NO NEED TO CREATE NEW FILES** - all files inherited and ready to use
   - **ADAPTATION NOTE**: Capability files and agent files are generic enough to apply to CI/CD;
     read them to understand their validated approaches, then apply to build infrastructure

1. **Build Infrastructure Analysis** (M₀.observe):
   - **READ** meta-agents/observe.md (CI/CD observation strategies)
   - Read Makefile (192 lines, 22 targets):
     - Analyze build targets: all, build, build-cli, build-mcp
     - Review test targets: test (short), test-all (full), test-coverage
     - Examine quality targets: lint, fmt, vet
     - Study release targets: cross-compile, bundle-release
   - Read scripts/release.sh (113 lines):
     - Analyze manual release steps (13 steps)
     - Identify automation opportunities
     - Note quality gates and validation
   - Read scripts/install-hooks.sh (49 lines):
     - Review git hook setup
     - Understand version management automation
   - Query CHANGELOG.md patterns:
     - Release note structure
     - Version tracking format

2. **Baseline Metrics Calculation** (M₀.plan + data-analyst):
   - **READ** meta-agents/plan.md (prioritization strategies)
   - **READ** agents/data-analyst.md
   - Invoke data-analyst to calculate V_instance(s₀):
     - V_automation: Count automated vs manual tasks
       - Automated: build, test, lint (local)
       - Manual: release, deploy, cross-platform verification, version bump
       - Calculate: automated_tasks / total_tasks
     - V_reliability: Estimate current reliability
       - No CI data, estimate based on manual process (~60%)
     - V_speed: Measure baseline time
       - Manual release: ~15-20 minutes
       - Target: ≤5 minutes
       - Calculate: 1 - (15 / baseline=15) = 0.50
     - V_observability: Assess current monitoring
       - Local logs only, no CI instrumentation (~30%)
     - **V_instance(s₀) = 0.3×V_auto + 0.3×V_rel + 0.2×V_speed + 0.2×V_obs**
   - Calculate V_meta(s₀):
     - V_completeness: 0.00 (no methodology yet)
     - V_effectiveness: 0.00 (nothing to test)
     - V_reusability: 0.00 (nothing to transfer)
     - **V_meta(s₀) = 0.00**

3. **Gap Identification** (M₀.reflect):
   - **READ** meta-agents/reflect.md (gap analysis process)
   - What CI/CD automation is missing?
     - No GitHub Actions workflows
     - No automated release process
     - No cross-platform CI verification
     - No automated quality gates
   - What quality gates need definition?
     - Test coverage threshold enforcement
     - Linting requirement enforcement
     - Cross-platform build verification
     - CHANGELOG validation
   - What deployment automation is needed?
     - GitHub Releases creation
     - Artifact uploading
     - Plugin marketplace sync
     - Version management integration

4. **Documentation** (M₀.execute + doc-writer):
   - **READ** meta-agents/execute.md (coordination strategies)
   - **READ** agents/doc-writer.md
   - Invoke doc-writer to create iteration-0.md:
     - M₀ state: 6 capabilities inherited from Bootstrap-006 (observe, plan, execute, reflect, evolve, api-design-orchestrator)
     - A₀ state: 15 agents inherited from Bootstrap-006 (3 generic + 12 specialized)
     - Note which inherited agents may be useful for CI/CD (e.g., agent-quality-gate-installer)
     - Infrastructure analysis summary
     - Calculated V_instance(s₀) and V_meta(s₀)
     - Gap analysis and problem statement
     - Reflection on next steps and potential agent reuse
   - Save data artifacts:
     - data/s0-infrastructure.yaml (Makefile analysis, script analysis)
     - data/s0-metrics.json (calculated values)
     - data/automation-opportunities.yaml (identified gaps)

5. **Reflection** (M₀.reflect + M₀.evolve):
   - **READ** meta-agents/reflect.md (reflection process)
   - **READ** meta-agents/evolve.md (methodology extraction readiness)
   - Is data collection complete? What additional data might be needed?
   - Are M₀ capabilities sufficient for baseline? (Yes, core capabilities adequate)
   - What should be the focus of Iteration 1?
     - Likely: Implement basic CI workflow (build + test + lint)
     - Or: Design GitHub Actions structure
   - Methodology extraction readiness:
     - Note patterns observed in existing infrastructure
     - Identify design decisions that could become methodology

## Constraints
- Do NOT pre-decide what agents to create next
- Do NOT assume the CI/CD design or evolution path
- Let the infrastructure data and gaps guide next steps
- Be honest about current automation state
- Calculate V(s₀) based on actual observations, not target values
- Remember two layers: concrete pipeline (instance) + methodology (meta)

## Output Format
Create iteration-0.md following this structure:
- Iteration metadata (number, date, duration)
- M₀ state documentation (6 capabilities inherited from Bootstrap-006)
- A₀ state documentation (15 agents inherited: 3 generic + 12 specialized)
- Assessment of inherited agent applicability to CI/CD domain
- Build infrastructure analysis (Makefile, scripts, git hooks)
- Value calculation (V_instance(s₀), V_meta(s₀))
- Gap identification (missing automation, quality gates, deployment)
- Reflection on next steps and agent reuse strategy
- Data artifacts saved to data/ directory
```

---

## Iteration 1+: Subsequent Iterations (General Template)

```markdown
# Execute Iteration N: [To be determined by Meta-Agent]

## Context from Previous Iteration

Review the previous iteration file: experiments/bootstrap-007-cicd-pipeline/iteration-[N-1].md

Extract:
- Current Meta-Agent state: M_{N-1}
- Current Agent Set: A_{N-1}
- Current pipeline state: V_instance(s_{N-1})
- Current methodology state: V_meta(s_{N-1})
- Problems identified
- Reflection notes on what's needed next

## Two-Layer Execution Protocol

**Layer 1 (Instance)**: Agents build concrete CI/CD pipeline
**Layer 2 (Meta)**: Meta-Agent observes and extracts methodology

Throughout iteration:
- Agents focus on concrete tasks (write workflows, define gates, implement automation)
- Meta-Agent observes agent work and identifies patterns for methodology

## Meta-Agent Decision Process

**BEFORE STARTING**: Read relevant Meta-Agent capability files:
- **READ** meta-agents/observe.md (for observation strategies)
- **READ** meta-agents/plan.md (for planning and decisions)
- **READ** meta-agents/execute.md (for coordination)
- **READ** meta-agents/reflect.md (for evaluation)
- **READ** meta-agents/evolve.md (for evolution assessment)

As M_{N-1}, follow the five-capability process:

### 1. OBSERVE (M.observe)
- **READ** meta-agents/observe.md for CI/CD observation strategies
- Review previous iteration outputs (iteration-[N-1].md)
- Examine pipeline state:
  - What workflows exist? (if any)
  - What quality gates are defined? (if any)
  - What automation is deployed? (if any)
  - What CI metrics are available? (if any)
- Identify gaps:
  - What automation is still missing?
  - What quality gates need definition?
  - What deployment steps are manual?
- **Methodology observation**:
  - What patterns emerged in previous agent work?
  - What design decisions were made and why?
  - What principles can be extracted?

### 2. PLAN (M.plan)
- **READ** meta-agents/plan.md for prioritization and agent selection
- Based on observations, what is the primary goal for this iteration?
  - Examples:
    - "Implement CI workflow (build + test + lint)"
    - "Define quality gates and thresholds"
    - "Automate release process"
    - "Add cross-platform builds"
- What capabilities are needed to achieve this goal?
- **Agent Assessment**:
  - Are current agents (A_{N-1}) sufficient for this goal?
  - Is generic `coder` enough to write GitHub Actions?
  - Or do we need specialized `pipeline-designer` for orchestration?
- **Methodology Planning**:
  - What patterns should be documented this iteration?
  - What design decisions will inform methodology?

### 3. EXECUTE (M.execute)
- **READ** meta-agents/execute.md for coordination and pattern observation
- Decision point: Should I create a new specialized agent?

**IF current agents are insufficient:**
- **EVOLVE** (M.evolve): Create new specialized agent
  - **READ** meta-agents/evolve.md for agent creation criteria
  - Examples of specialized agents:
    - `pipeline-designer`: Design GitHub Actions workflows, stage orchestration
    - `quality-gate-definer`: Define gates, thresholds, policies
    - `deployment-strategist`: Design deployment strategies
    - `rollback-planner`: Design rollback and recovery procedures
  - Define agent name and specialization domain
  - Document capabilities the new agent provides
  - Explain why generic agents are insufficient
  - **CREATE AGENT PROMPT FILE**: Write agents/{agent-name}.md
    - Include: agent role, CI/CD-specific capabilities, input/output format
    - Include: specific instructions for this iteration's task
    - Include: CI/CD domain knowledge (GitHub Actions syntax, best practices)
  - Add to agent set: A_N = A_{N-1} ∪ {new_agent}

**Agent Invocation** (specialized or generic):
- **READ agent prompt file** before invocation: agents/{agent-name}.md
- Invoke agent to execute concrete pipeline work:
  - Write GitHub Actions workflow files
  - Define quality gate configurations
  - Implement deployment automation
  - Create monitoring/observability
- Produce iteration outputs (workflow files, configs, scripts)

**Methodology Extraction** (M.evolve):
- **OBSERVE agent work patterns**:
  - How did agent organize workflow stages?
  - What decisions were made about parallelization?
  - How were thresholds determined?
  - What deployment strategy was chosen?
- **EXTRACT patterns for methodology**:
  - Document design decisions and rationale
  - Identify reusable patterns
  - Note principles that emerge
  - Add to methodology documentation

**ELSE use existing agents:**
- **READ agent prompt file** from agents/{agent-name}.md
- Invoke appropriate agents from A_{N-1}
- Execute planned pipeline work
- Observe for methodology patterns

**CRITICAL EXECUTION PROTOCOL**:
1. ALWAYS read capability files before embodying Meta-Agent capabilities
2. ALWAYS read agent prompt file before each agent invocation
3. Do NOT cache instructions across iterations - always read from files
4. Capability files may be updated between iterations - get latest from files
5. Never assume capabilities - always verify from source files

### 4. REFLECT (M.reflect)
- **READ** meta-agents/reflect.md for evaluation process
- **Evaluate Instance Layer** (Concrete Pipeline):
  - What pipeline components were built this iteration?
  - Calculate new V_instance(s_N):
    - V_automation: What % of tasks are now automated?
    - V_reliability: What's the CI success rate? (if CI exists)
    - V_speed: How long does pipeline take?
    - V_observability: What % of stages are instrumented?
    - **V_instance(s_N) = 0.3×V_auto + 0.3×V_rel + 0.2×V_speed + 0.2×V_obs**
  - Calculate change: ΔV_instance = V_instance(s_N) - V_instance(s_{N-1})
  - Are pipeline objectives met? What gaps remain?

- **Evaluate Meta Layer** (Methodology):
  - What patterns were extracted this iteration?
  - Calculate new V_meta(s_N):
    - V_completeness: What % of patterns are documented?
    - V_effectiveness: Can we estimate speedup on transfer yet?
    - V_reusability: Have we tested transfer? (later iterations)
    - **V_meta(s_N) = 0.4×V_complete + 0.3×V_effective + 0.3×V_reusable**
  - Calculate change: ΔV_meta = V_meta(s_N) - V_meta(s_{N-1})
  - Is methodology sufficiently documented? What patterns are missing?

- **Honest Assessment**:
  - Don't inflate values to meet targets
  - Calculate based on actual state
  - Identify real gaps, not aspirational ones

### 5. CHECK CONVERGENCE

Evaluate convergence criteria:

```yaml
convergence_check:
  meta_agent_stable:
    question: "Did M gain new capabilities this iteration?"
    M_N == M_{N-1}: [Yes/No]
    details: "M capabilities are modular, rarely change"

  agent_set_stable:
    question: "Were new agents created this iteration?"
    A_N == A_{N-1}: [Yes/No]
    details: "If no new pipeline specialization needed"

  instance_value_threshold:
    question: "Is V_instance(s_N) ≥ 0.80 (pipeline quality)?"
    V_instance(s_N): [calculated value]
    threshold_met: [Yes/No]
    components:
      V_automation: [value] "≥0.90 target"
      V_reliability: [value] "≥0.95 target"
      V_speed: [value] "≥0.67 target (≤5 min)"
      V_observability: [value] "=1.00 target"

  meta_value_threshold:
    question: "Is V_meta(s_N) ≥ 0.80 (methodology quality)?"
    V_meta(s_N): [calculated value]
    threshold_met: [Yes/No]
    components:
      V_completeness: [value] "≥0.90 target"
      V_effectiveness: [value] "≥0.80 target (3x speedup)"
      V_reusability: [value] "≥0.65 target (65% transfer)"

  instance_objectives:
    ci_workflow_deployed: [Yes/No]
    quality_gates_enforced: [Yes/No]
    release_automated: [Yes/No]
    observability_implemented: [Yes/No]
    all_objectives_met: [Yes/No]

  meta_objectives:
    methodology_documented: [Yes/No]
    patterns_extracted: [Yes/No]
    transfer_tests_conducted: [Yes/No]
    all_objectives_met: [Yes/No]

  diminishing_returns:
    ΔV_instance_current: [current change]
    ΔV_meta_current: [current change]
    interpretation: "Is improvement marginal?"

convergence_status: [CONVERGED / NOT_CONVERGED]
```

**IF CONVERGED:**
- Stop iteration process
- Proceed to results analysis
- Document three-tuple: (O, A_N, M_N)
  - O = {pipeline artifacts, methodology documentation}
  - A_N = {final agent set with specializations}
  - M_N = {final meta-agent capabilities}

**IF NOT CONVERGED:**
- Identify what's needed for next iteration
- Note: Focus on instance (pipeline) OR meta (methodology) as needed
- Continue to Iteration N+1

## Documentation Requirements

Create experiments/bootstrap-007-cicd-pipeline/iteration-N.md with:

### 1. Iteration Metadata
```yaml
iteration: N
date: YYYY-MM-DD
duration: ~X hours
status: [completed/converged]
layers:
  instance: "Pipeline component built this iteration"
  meta: "Methodology patterns extracted this iteration"
```

### 2. Meta-Agent Evolution (if applicable)
```yaml
M_{N-1} → M_N:
  evolution: [evolved / unchanged]
  
  IF evolved:
    new_capabilities: [list new capabilities added]
    evolution_reason: "Why was this capability needed?"
    evolution_trigger: "What problem triggered this?"
    capability_file_created: "meta-agents/{new-capability}.md"

  IF unchanged:
    status: "M_N = M_{N-1} (no evolution, core capabilities sufficient)"
```

### 3. Agent Set Evolution (if applicable)
```yaml
A_{N-1} → A_N:
  evolution: [evolved / unchanged / reused_inherited]

  inherited_baseline: "15 agents from Bootstrap-006 (3 generic + 12 specialized)"

  IF evolved (new agent created):
    new_agents:
      - name: agent_name
        specialization: ci_cd_domain_area
        capabilities: [list]
        creation_reason: "Why were inherited 15 agents insufficient?"
        justification: "What gap exists that no inherited agent can fill?"
        prompt_file: "agents/{agent-name}.md"

  IF unchanged:
    status: "A_N = A_{N-1} (inherited agents sufficient for this iteration)"

  IF reused_inherited:
    reused_agents:
      - agent_name: task_performed
        adaptation_notes: "How inherited agent was adapted to CI/CD context"

agents_invoked_this_iteration:
  - agent_name: task_performed
    source: [inherited / newly_created]
```

### 4. Instance Work Executed (Concrete Pipeline)
- What pipeline components were built?
  - GitHub Actions workflows created/modified
  - Quality gates defined/implemented
  - Deployment automation added
  - Monitoring/observability configured
- What outputs were produced?
  - Workflow files: .github/workflows/*.yml
  - Configuration files
  - Scripts and automation
- Summary of concrete deliverables

### 5. Meta Work Executed (Methodology Extraction)
- What patterns were observed in agent work?
  - Design decisions made
  - Principles that emerged
  - Reusable patterns identified
- What methodology content was documented?
  - Patterns added to methodology
  - Frameworks refined
  - Decision criteria codified

### 6. State Transition

**Instance Layer** (Pipeline State):
```yaml
s_{N-1} → s_N (CI/CD Pipeline):
  changes:
    - description of pipeline changes

  metrics:
    V_automation: [value] (was: [previous])
    V_reliability: [value] (was: [previous])
    V_speed: [value] (was: [previous])
    V_observability: [value] (was: [previous])

  value_function:
    V_instance(s_N): [calculated]
    V_instance(s_{N-1}): [previous]
    ΔV_instance: [change]
    percentage: +X.X%
```

**Meta Layer** (Methodology State):
```yaml
methodology_{N-1} → methodology_N:
  changes:
    - patterns extracted
    - principles documented
    - frameworks refined

  metrics:
    V_completeness: [value] (was: [previous])
    V_effectiveness: [value] (was: [previous])
    V_reusability: [value] (was: [previous])

  value_function:
    V_meta(s_N): [calculated]
    V_meta(s_{N-1}): [previous]
    ΔV_meta: [change]
    percentage: +X.X%
```

### 7. Reflection
- What was learned this iteration?
  - Instance learnings (pipeline construction)
  - Meta learnings (methodology insights)
- What worked well?
- What challenges were encountered?
- What is needed next?
  - For pipeline completion
  - For methodology completion

### 8. Convergence Check
[Use the convergence criteria structure above]

### 9. Data Artifacts
Save artifacts to data/ directory:
- data/iteration-N-metrics.json (V_instance, V_meta calculations)
- data/iteration-N-pipeline-state.yaml (CI/CD state)
- data/iteration-N-methodology.yaml (extracted patterns)
- data/iteration-N-artifacts/ (workflow files, configs)

Reference data files in iteration document.

## Key Principles

1. **Two-Layer Awareness**: Always work on both layers
   - Instance: Build concrete pipeline components
   - Meta: Extract methodology patterns from agent work

2. **Be Honest**: Calculate V(s_N) based on actual state
   - Don't inflate instance values (pipeline quality)
   - Don't inflate meta values (methodology quality)

3. **Let System Evolve**: Don't force predetermined paths
   - Create agents based on need, not plan
   - Extract patterns that actually emerge
   - Don't fabricate methodology to meet targets

4. **Justify Specialization**: Only create agents when generic insufficient
   - Document clear reason for specialization
   - Explain what generic agent couldn't do

5. **Document Evolution**: Clearly explain WHY M or A evolved
   - What triggered agent creation?
   - What capability gap was identified?

6. **Check Convergence Rigorously**: Evaluate both layers
   - Instance convergence: Pipeline quality + stability
   - Meta convergence: Methodology quality + completeness

7. **Stop When Done**: If converged, don't force more iterations
   - Both V_instance ≥ 0.80 AND V_meta ≥ 0.80
   - System stable (M_N = M_{N-1}, A_N = A_{N-1})

8. **No Token Limits**: Complete all steps thoroughly
   - Do NOT skip steps due to perceived token limits
   - Do NOT abbreviate data collection or analysis
   - Do NOT summarize when full details are needed
   - Complete ALL steps regardless of length

9. **Capability Files Required**: Always read before use
   - Meta-Agent files: meta-agents/{observe,plan,execute,reflect,evolve}.md
   - Agent files: agents/{agent-name}.md
   - Read: ALWAYS read capability file before embodying capability
   - Read: ALWAYS read agent file before agent invocation
   - Update: Rarely update capability files (stable architecture)
   - Update: Create new agent files when agents specialize

## CI/CD Domain-Specific Patterns

Based on OCA framework, CI/CD iterations may follow:

### Observe Phase (Iterations 0-1)
- Analyze existing build infrastructure (Makefile, scripts)
- Identify manual processes and automation gaps
- Study release workflows and quality checks
- **Methodology**: Observe infrastructure patterns

### Codify Phase (Iterations 2-3)
- Implement CI workflows (GitHub Actions)
- Define quality gates and thresholds
- Create deployment automation
- **Methodology**: Extract pipeline design patterns from implementations

### Automate Phase (Iterations 4-5)
- Add advanced automation (multi-platform, parallel builds)
- Implement monitoring and observability
- Refine quality gates based on data
- **Methodology**: Document automation decision frameworks

**Caveat**: Let actual needs drive the sequence, not this expected pattern. The pattern is a hint, not a prescription.
```

---

## Quick Reference: Iteration Checklist

For each iteration N ≥ 1, ensure you:

**Preparation**:
- [ ] Review previous iteration (iteration-[N-1].md)
- [ ] Extract current state (M_{N-1}, A_{N-1}, V_instance(s_{N-1}), V_meta(s_{N-1}))

**Observe Phase**:
- [ ] **READ** meta-agents/observe.md (CI/CD observation strategies)
- [ ] Analyze pipeline state (workflows, gates, automation)
- [ ] Identify gaps (missing automation, quality gates, deployment)
- [ ] Observe patterns (for methodology extraction)

**Plan Phase**:
- [ ] **READ** meta-agents/plan.md (prioritization, agent selection)
- [ ] Define iteration goal (instance layer: concrete pipeline component)
- [ ] Assess agent sufficiency (generic sufficient OR need specialized?)
- [ ] Plan methodology extraction (what patterns to document?)

**Execute Phase**:
- [ ] **READ** meta-agents/execute.md (coordination, pattern observation)
- [ ] **IF NEW AGENT NEEDED**:
  - [ ] **READ** meta-agents/evolve.md (agent creation criteria)
  - [ ] Create agent prompt file: agents/{agent-name}.md
  - [ ] Document specialization reason
- [ ] **READ** agent prompt file(s) before invocation: agents/{agent-name}.md
- [ ] Invoke agents to build pipeline components
- [ ] Observe agent work for methodology patterns
- [ ] Extract patterns and update methodology documentation

**Reflect Phase**:
- [ ] **READ** meta-agents/reflect.md (evaluation process)
- [ ] Calculate V_instance(s_N) (pipeline quality):
  - [ ] V_automation (automated tasks / total tasks)
  - [ ] V_reliability (successful builds / total builds)
  - [ ] V_speed (1 - current_time / baseline_time)
  - [ ] V_observability (instrumented stages / total stages)
- [ ] Calculate V_meta(s_N) (methodology quality):
  - [ ] V_completeness (documented patterns / total patterns)
  - [ ] V_effectiveness (speedup on transfer tests)
  - [ ] V_reusability (successful transfers / attempts)
- [ ] Assess quality honestly (don't inflate values)
- [ ] Identify gaps (pipeline + methodology)

**Convergence Check**:
- [ ] M_N == M_{N-1}? (meta-agent stable)
- [ ] A_N == A_{N-1}? (agent set stable)
- [ ] V_instance(s_N) ≥ 0.80? (pipeline quality threshold)
- [ ] V_meta(s_N) ≥ 0.80? (methodology quality threshold)
- [ ] Instance objectives complete? (pipeline deployed)
- [ ] Meta objectives complete? (methodology documented)
- [ ] Determine: CONVERGED or NOT_CONVERGED

**Documentation**:
- [ ] Create iteration-N.md with:
  - [ ] Metadata (iteration, date, duration, status)
  - [ ] M evolution (evolved or unchanged)
  - [ ] A evolution (new agents or unchanged)
  - [ ] Instance work (pipeline components built)
  - [ ] Meta work (patterns extracted)
  - [ ] State transition (V_instance, V_meta calculated)
  - [ ] Reflection (learnings, next steps)
  - [ ] Convergence check (criteria evaluated)
- [ ] Save data artifacts to data/:
  - [ ] iteration-N-metrics.json
  - [ ] iteration-N-pipeline-state.yaml
  - [ ] iteration-N-methodology.yaml
  - [ ] iteration-N-artifacts/ (workflow files, configs)

**Quality Assurance**:
- [ ] **NO TOKEN LIMITS**: Verify all steps completed fully without abbreviation
- [ ] Capability files read before use
- [ ] Agent files read before invocation
- [ ] Honest value calculations (not inflated)
- [ ] Both layers addressed (instance + meta)

**If CONVERGED**:
- [ ] Create results.md with 10-point analysis
- [ ] Perform reusability validation (transfer tests)
- [ ] Compare with actual history (hypothetical)
- [ ] Document three-tuple: (O={pipeline+methodology}, A_N, M_N)
- [ ] Validate methodology alignment (OCA, Bootstrapped SE, Value Space)

---

## Notes on Execution Style

**Be the Meta-Agent**: When executing iterations, embody M's perspective:
- Think through the observe-plan-execute-reflect-evolve cycle
- Read capability files to understand M's strategies
- Make explicit decisions about agent creation
- Justify why specialization is needed
- Extract methodology patterns from agent work
- Track both V_instance and V_meta

**Be Domain-Aware**: CI/CD has specific concerns:
- GitHub Actions syntax and best practices
- Multi-platform builds (cross-compilation)
- Quality gate thresholds (coverage, linting)
- Deployment strategies (releases, artifacts)
- Rollback and recovery procedures
- Observability patterns (monitoring, alerting)

**Be Rigorous**: Calculate values honestly
- V_instance based on actual pipeline state (not aspirational)
- V_meta based on actual methodology coverage (not desired)
- Don't force convergence prematurely
- Don't skip methodology extraction to focus only on pipeline
- Let data and needs drive the process

**Be Thorough**: Document decisions and reasoning
- Save intermediate data (metrics, configs, workflows)
- Show your work (calculations, analysis)
- Make evolution path traceable
- **NO TOKEN LIMITS**: Complete all steps fully, never abbreviate

**Be Two-Layer Aware**: Always work on both layers
- Instance layer: Build concrete CI/CD pipeline
- Meta layer: Extract reusable methodology
- Don't neglect either layer
- Methodology extraction is as important as pipeline implementation

**Be Authentic**: This is a real experiment
- Discover CI/CD patterns, don't assume them
- Create agents based on need, not predetermined plan
- Extract methodology from actual agent work, not theory
- Stop when truly converged (both layers), not at target iteration count

**Capability File Protocol**:
- **Meta-Agent capabilities**: meta-agents/{observe,plan,execute,reflect,evolve}.md
  - **ALWAYS** read capability file before embodying that capability
  - Modular architecture: each capability in separate file
  - Rarely change (stable core capabilities)
  - Never assume capability details - always read from file
- **Agent files**: agents/{agent-name}.md
  - **ALWAYS** read agent file before invocation
  - Create files for new specialized agents
  - Update files as agents evolve or requirements change
  - Never assume agent instructions - always read from file
- **Reading ensures**:
  - Complete context and all details captured
  - No assumptions about capabilities or processes
  - Latest updates and refinements incorporated
  - Explicit rather than implicit execution

---

**Document Version**: 1.0
**Created**: 2025-10-16
**Last Updated**: 2025-10-16
**Purpose**: Guide authentic execution of bootstrap-007-cicd-pipeline experiment with dual-layer architecture

**Key Innovation**: Dual value functions (V_instance + V_meta) enable simultaneous optimization of concrete pipeline and reusable methodology.
