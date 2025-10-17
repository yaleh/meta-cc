# Iteration Execution Prompts

This document provides structured prompts for executing each iteration of the bootstrap-012-technical-debt experiment.

**Two-Layer Architecture**:
- **Instance Layer**: Agents quantify and prioritize technical debt in meta-cc codebase
- **Meta Layer**: Meta-Agent observes debt measurement patterns and extracts methodology

---

## Iteration 0: Baseline Establishment

```markdown
# Execute Iteration 0: Baseline Establishment

## Context
I'm starting the bootstrap-012-technical-debt experiment. I've reviewed:
- experiments/bootstrap-012-technical-debt/plan.md
- experiments/bootstrap-012-technical-debt/README.md
- experiments/EXPERIMENTS-OVERVIEW.md (Bootstrap-012 specification)
- The three methodology frameworks (OCA, Bootstrapped SE, Value Space Optimization)

## Current State
- Meta-Agent: M₀ (5 core capabilities inherited from bootstrap-003: observe, plan, execute, reflect, evolve)
- Agent Set: A₀ (3 generic agents inherited from bootstrap-003: data-analyst, doc-writer, coder)
- Target: All `internal/` and `cmd/` modules (~5,000 lines)

## Inherited State from Bootstrap-003

**IMPORTANT**: This experiment starts with the converged generic agent set from Bootstrap-003, NOT from scratch.

**Meta-Agent capability files (ALREADY EXIST)**:
- experiments/bootstrap-012-technical-debt/meta-agents/observe.md (validated)
- experiments/bootstrap-012-technical-debt/meta-agents/plan.md (validated)
- experiments/bootstrap-012-technical-debt/meta-agents/execute.md (validated)
- experiments/bootstrap-012-technical-debt/meta-agents/reflect.md (validated)
- experiments/bootstrap-012-technical-debt/meta-agents/evolve.md (validated)

**Agent prompt files (ALREADY EXIST - 3 generic agents)**:
- experiments/bootstrap-012-technical-debt/agents/data-analyst.md
- experiments/bootstrap-012-technical-debt/agents/doc-writer.md
- experiments/bootstrap-012-technical-debt/agents/coder.md

**CRITICAL EXECUTION PROTOCOL**:
- All capability files and agent files ALREADY EXIST (inherited from bootstrap-003)
- Before embodying Meta-Agent capabilities, ALWAYS read the relevant capability file first
- Before invoking ANY agent, ALWAYS read its prompt file first
- These files contain validated capabilities/agents; adapt them to technical debt context
- Never assume capabilities - always read from the source files

## Iteration 0 Objectives

Execute baseline establishment:

0. **Setup** (Verify inherited state):
   - **VERIFY META-AGENT CAPABILITY FILES EXIST** (inherited from bootstrap-003):
     - ✓ meta-agents/observe.md: Data collection, pattern discovery (validated)
     - ✓ meta-agents/plan.md: Prioritization, agent selection (validated)
     - ✓ meta-agents/execute.md: Agent coordination, task execution (validated)
     - ✓ meta-agents/reflect.md: Value calculation (V_instance, V_meta), gap analysis (validated)
     - ✓ meta-agents/evolve.md: Agent creation criteria, methodology extraction (validated)
   - **VERIFY INITIAL AGENT PROMPT FILES EXIST** (inherited from bootstrap-003):
     - ✓ agents/data-analyst.md (generic agent)
     - ✓ agents/doc-writer.md (generic agent)
     - ✓ agents/coder.md (generic agent)
   - **NO NEED TO CREATE NEW FILES** - all files inherited and ready to use
   - **ADAPTATION NOTE**: Capability files and agent files are generic enough to apply to technical debt;
     read them to understand their validated approaches, then apply to debt measurement context

1. **Codebase Inventory** (M₀.observe):
   - **READ** meta-agents/observe.md (technical debt observation strategies)
   - Analyze target codebase structure:
     - internal/ modules:
       - parser/ (~800 lines): JSONL parsing, session data extraction
       - analyzer/ (~1,200 lines): Pattern detection, data aggregation
       - query/ (~1,500 lines): Query engine, filtering, sorting
       - validation/ (~600 lines): Schema validation, input checking
       - tools/ (~400 lines): Tool definitions and registry
       - capabilities/ (~500 lines): Capability management
     - cmd/ modules:
       - meta-cc/ (~500 lines): CLI entry point
       - mcp/ (~500 lines): MCP server implementation
   - Total: ~5,000 lines of Go code
   - Identify high-change files from meta-cc session data:
     - Query: `meta-cc query-files --threshold 20` (high-change = debt risk)
     - Expected hotspots: tools.go, plan.md, query engine modules
   - Identify error-prone files:
     - Query: `meta-cc query-tools --status error --tool Edit` (error-prone edits)
     - Look for files with repeated edit failures

2. **Technical Debt Data Collection** (M₀.observe + data-analyst):
   - **READ** meta-agents/observe.md (observation strategies)
   - **READ** agents/data-analyst.md
   - Invoke data-analyst to collect baseline debt metrics:

     **Code Complexity Metrics**:
     - Run `gocyclo -over 10 ./internal ./cmd` (cyclomatic complexity)
     - Identify functions with complexity >15 (high debt risk)
     - Calculate average complexity per module

     **Code Duplication Metrics**:
     - Run `dupl -threshold 50 ./internal ./cmd` (duplicate code detection)
     - Identify duplicated code blocks >50 tokens
     - Calculate duplication percentage per module

     **Static Analysis Metrics**:
     - Run `staticcheck ./...` (static analysis)
     - Count issues by severity (error, warning, info)
     - Categorize by issue type (unused code, inefficient patterns, etc.)

     **Test Coverage Metrics**:
     - Run `go test -coverprofile=coverage.out ./...`
     - Run `go tool cover -func=coverage.out` (per-function coverage)
     - Identify modules with <80% coverage (debt indicator)

     **Change Frequency Metrics** (from meta-cc session data):
     - Query: `meta-cc query-files --threshold 20`
     - Identify files with >20 edits (change hotspots)
     - Correlate high-change with low test coverage (debt accumulation pattern)

     **Error Pattern Metrics**:
     - Query: `meta-cc query-conversation --pattern "fix|bug|issue|problem"`
     - Count bug-fix conversations
     - Identify modules mentioned in bug discussions (quality debt)

3. **Baseline Debt Quantification** (M₀.plan + data-analyst):
   - **READ** meta-agents/plan.md (prioritization strategies)
   - **READ** agents/data-analyst.md
   - Invoke data-analyst to calculate V_instance(s₀):
     - V_measurement: Accuracy of debt quantification
       - Baseline: 0.30 (basic metrics collected, no SQALE index, no holistic view)
       - Metrics: (complexity + duplication + coverage + static_analysis) / 4
       - Coverage: 4/10 possible debt dimensions (missing: documentation debt, architecture debt, performance debt, security debt, dependency debt, process debt)
       - Calculate: 0.3 (partial coverage, no prioritization framework)
     - V_prioritization: Quality of debt ranking
       - Baseline: 0.20 (no prioritization framework, unclear which debt matters most)
       - No value/effort analysis
       - No business impact assessment
       - No paydown ROI calculation
       - Calculate: 0.2 (ad-hoc, subjective prioritization)
     - V_tracking: Visibility of debt trends
       - Baseline: 0.10 (no historical tracking, no trend analysis)
       - No baseline comparison
       - No debt accumulation rate tracking
       - No paydown progress monitoring
       - Calculate: 0.1 (one-time snapshot only)
     - V_actionability: Clarity of paydown strategies
       - Baseline: 0.40 (some tools suggest fixes, but no systematic paydown plan)
       - staticcheck provides some guidance
       - No paydown roadmap
       - No effort estimation
       - Calculate: 0.4 (tool suggestions exist, but incomplete)
     - **V_instance(s₀) = 0.3×0.30 + 0.3×0.20 + 0.2×0.10 + 0.2×0.40 = 0.25**
   - Calculate V_meta(s₀):
     - V_completeness: 0.00 (no methodology yet)
     - V_effectiveness: 0.00 (nothing to test)
     - V_reusability: 0.00 (nothing to transfer)
     - **V_meta(s₀) = 0.00**

4. **Debt Hotspot Identification** (M₀.reflect):
   - **READ** meta-agents/reflect.md (gap analysis process)
   - What are the highest-debt areas?
     - Modules with complexity >15 AND coverage <80%
     - Files with >20 edits AND error-prone edit history
     - Modules with high duplication AND high change frequency
   - What debt types are most prevalent?
     - Code complexity debt (cyclomatic complexity)
     - Code duplication debt (duplicated logic)
     - Test coverage debt (insufficient tests)
     - Code smell debt (staticcheck issues)
     - Change frequency debt (fragile code)
   - What debt dimensions are missing?
     - Documentation debt (missing docs, outdated docs)
     - Architecture debt (coupling, cohesion issues)
     - Performance debt (inefficient algorithms)
     - Security debt (vulnerabilities, unsafe patterns)
     - Dependency debt (outdated deps, unnecessary deps)
     - Process debt (manual processes, missing automation)

5. **Gap Identification** (M₀.reflect):
   - **READ** meta-agents/reflect.md (gap analysis process)
   - What technical debt capabilities are missing?
     - No SQALE methodology implementation (industry standard)
     - No debt categorization framework (code vs. architecture vs. process)
     - No value/effort prioritization (which debt to pay down first?)
     - No paydown roadmap (sequencing, dependencies)
     - No prevention checklist (avoid new debt)
     - No debt trend tracking (is debt growing or shrinking?)
   - What debt aspects need coverage?
     - Measurement: SQALE index, code smells, technical debt ratio
     - Prioritization: Value/effort matrix, business impact, risk assessment
     - Tracking: Historical baselines, debt accumulation rate, paydown velocity
     - Actionability: Paydown strategies, effort estimation, ROI calculation
   - What methodology components are needed?
     - Debt measurement framework (SQALE, code smells)
     - Debt prioritization matrix (value vs. effort)
     - Debt tracking system (trend analysis)
     - Paydown strategy templates (refactoring patterns)
     - Prevention guidelines (best practices to avoid debt)

6. **Initial Agent Applicability Assessment** (M₀.plan):
   - **READ** meta-agents/plan.md (agent selection strategies)
   - Which inherited agents are directly applicable to technical debt?
     - ⭐⭐⭐ data-analyst: Collect metrics, analyze trends, calculate debt indices
     - ⭐⭐ coder: Write custom debt analysis scripts, automation tools
     - ⭐ doc-writer: Document debt reports, paydown plans, prevention guides
   - Which inherited agents need adaptation for technical debt?
     - All agents: Apply to technical debt domain instead of error recovery
     - Read agent files to understand capabilities, adapt prompts contextually
   - What new specialized agents might be needed?
     - debt-quantifier: Calculate SQALE index, technical debt ratio, code smells
     - hotspot-identifier: Find high-debt areas (complexity + change frequency)
     - impact-analyzer: Assess debt impact on velocity and quality
     - paydown-strategist: Prioritize debt by value/effort ratio
     - trend-tracker: Track debt accumulation/paydown over time
     - prevention-advisor: Suggest practices to prevent new debt
   - **NOTE**: Don't create new agents yet - just identify potential needs

7. **Documentation** (M₀.execute + doc-writer):
   - **READ** meta-agents/execute.md (coordination strategies)
   - **READ** agents/doc-writer.md
   - Invoke doc-writer to create iteration-0.md:
     - M₀ state: 5 capabilities inherited from bootstrap-003 (observe, plan, execute, reflect, evolve)
     - A₀ state: 3 agents inherited from bootstrap-003 (data-analyst, doc-writer, coder)
     - Agent applicability assessment (which agents useful for technical debt)
     - Codebase inventory (internal/ + cmd/ = ~5,000 lines)
     - Technical debt data collection summary (complexity, duplication, coverage, static analysis, change frequency, error patterns)
     - Calculated V_instance(s₀) = 0.25 and V_meta(s₀) = 0.00
     - Debt hotspot identification (high-debt modules and files)
     - Gap analysis (missing debt capabilities and methodology)
     - Reflection on next steps and agent reuse strategy
   - Save data artifacts:
     - data/s0-codebase-inventory.yaml (module breakdown, line counts)
     - data/s0-debt-metrics-raw.json (gocyclo, dupl, staticcheck, coverage outputs)
     - data/s0-debt-hotspots.yaml (identified high-debt areas)
     - data/s0-metrics.json (calculated V_instance and V_meta values)
     - data/s0-gap-analysis.yaml (identified gaps in debt measurement)
     - data/s0-agent-applicability.yaml (inherited agents and their debt applicability)
   - Initialize knowledge structure:
     - knowledge/INDEX.md (empty knowledge catalog, will be populated in subsequent iterations)
     - knowledge/patterns/ (directory for domain-specific patterns)
     - knowledge/principles/ (directory for universal principles)
     - knowledge/templates/ (directory for reusable templates)
     - knowledge/best-practices/ (directory for context-specific practices)

8. **Reflection** (M₀.reflect + M₀.evolve):
   - **READ** meta-agents/reflect.md (reflection process)
   - **READ** meta-agents/evolve.md (methodology extraction readiness)
   - Is data collection complete? What additional data might be needed?
     - SQALE index calculation (needs implementation)
     - Code smell detection (beyond staticcheck)
     - Architecture quality metrics (coupling, cohesion)
   - Are M₀ capabilities sufficient for baseline? (Yes, core capabilities adequate)
   - What should be the focus of Iteration 1?
     - Likely: Implement SQALE methodology for debt quantification
     - Or: Build debt prioritization matrix (value vs. effort)
     - Decision based on OCA framework: Start with Observe phase (comprehensive measurement)
   - Which inherited agents will be most useful in Iteration 1?
     - Likely: Need new debt-quantifier agent (SQALE implementation)
     - Reuse: data-analyst (metrics aggregation), doc-writer (report generation)
   - Methodology extraction readiness:
     - Note patterns observed in debt metrics
     - Identify measurement decision points that could become methodology
     - Prepare for pattern documentation in subsequent iterations

## Constraints
- Do NOT pre-decide what debt metrics to use
- Do NOT assume the debt prioritization framework
- Let the codebase data and metrics guide next steps
- Be honest about current debt state (low baseline expected)
- Calculate V(s₀) based on actual observations, not target values
- Remember two layers: concrete debt quantification (instance) + methodology (meta)

## Output Format
Create iteration-0.md following this structure:
- Iteration metadata (number, date, duration)
- M₀ state documentation (5 capabilities inherited from bootstrap-003)
- A₀ state documentation (3 agents inherited: data-analyst, doc-writer, coder)
- Agent applicability assessment (which agents useful for technical debt domain)
- Codebase inventory (internal/ + cmd/ modules, ~5,000 lines)
- Technical debt data collection (complexity, duplication, coverage, static analysis, change frequency, error patterns)
- Value calculation (V_instance(s₀) = 0.25, V_meta(s₀) = 0.00)
- Debt hotspot identification (high-debt areas)
- Gap identification (missing debt capabilities and methodology)
- Reflection on next steps and agent reuse strategy
- Data artifacts saved to data/ directory
```

---

## Iteration 1+: Subsequent Iterations (General Template)

```markdown
# Execute Iteration N: [To be determined by Meta-Agent]

## Context from Previous Iteration

Review the previous iteration file: experiments/bootstrap-012-technical-debt/iteration-[N-1].md

Extract:
- Current Meta-Agent state: M_{N-1}
- Current Agent Set: A_{N-1}
- Current debt state: V_instance(s_{N-1})
- Current methodology state: V_meta(s_{N-1})
- Debt problems identified
- Reflection notes on what's needed next

## Two-Layer Execution Protocol

**Layer 1 (Instance)**: Agents perform concrete technical debt quantification and prioritization
**Layer 2 (Meta)**: Meta-Agent observes and extracts debt measurement methodology

Throughout iteration:
- Agents focus on concrete tasks (measure debt, prioritize paydown, create roadmap)
- Meta-Agent observes debt measurement work and identifies patterns for methodology

## Meta-Agent Decision Process

**BEFORE STARTING**: Read relevant Meta-Agent capability files:
- **READ** meta-agents/observe.md (for observation strategies)
- **READ** meta-agents/plan.md (for planning and decisions)
- **READ** meta-agents/execute.md (for coordination)
- **READ** meta-agents/reflect.md (for evaluation)
- **READ** meta-agents/evolve.md (for evolution assessment)

As M_{N-1}, follow the five-capability process:

### 1. OBSERVE (M.observe)
- **READ** meta-agents/observe.md for technical debt observation strategies
- Review previous iteration outputs (iteration-[N-1].md)
- Examine debt measurement state:
  - What debt metrics have been collected? (if any)
  - What debt dimensions have been measured? (if any)
  - What patterns have been observed? (if any)
  - What paydown strategies have been identified? (if any)
- Identify gaps:
  - What debt dimensions still need measurement?
  - What measurement aspects are missing? (SQALE, code smells, trend tracking)
  - What prioritization frameworks are needed?
  - What methodology patterns are emerging?
- **Methodology observation**:
  - What patterns emerged in previous debt measurement work?
  - What measurement decisions were made and why?
  - What principles can be extracted?

### 2. PLAN (M.plan)
- **READ** meta-agents/plan.md for prioritization and agent selection
- Based on observations, what is the primary goal for this iteration?
  - Examples:
    - "Implement SQALE methodology for technical debt index"
    - "Build debt prioritization matrix (value vs. effort)"
    - "Create paydown roadmap with sequencing and dependencies"
    - "Implement debt trend tracking system"
    - "Develop prevention checklist and best practices"
- What capabilities are needed to achieve this goal?
- **Agent Assessment**:
  - Are current agents (A_{N-1}) sufficient for this goal?
  - Can inherited agents handle debt quantification? (data-analyst for metrics, coder for tooling)
  - Or do we need specialized `debt-quantifier` for SQALE implementation?
  - Do we need `hotspot-identifier` for high-debt area detection?
  - Do we need `impact-analyzer` for debt impact assessment?
  - Do we need `paydown-strategist` for prioritization?
  - Do we need `trend-tracker` for debt trend monitoring?
  - Do we need `prevention-advisor` for prevention guidelines?
- **Methodology Planning**:
  - What patterns should be documented this iteration?
  - What measurement decisions will inform methodology?

### 3. EXECUTE (M.execute)
- **READ** meta-agents/execute.md for coordination and pattern observation
- Decision point: Should I create a new specialized agent?

**IF current agents are insufficient:**
- **EVOLVE** (M.evolve): Create new specialized agent
  - **READ** meta-agents/evolve.md for agent creation criteria
  - Examples of specialized agents:
    - `debt-quantifier`: Calculate SQALE index, technical debt ratio, code smells
      - Capabilities: SQALE implementation, code smell detection, debt ratio calculation
      - Why needed: Generic agents lack SQALE domain expertise
    - `hotspot-identifier`: Find high-debt areas (complexity + change frequency)
      - Capabilities: Combine metrics, identify correlations, rank hotspots
      - Why needed: Multi-dimensional analysis requires specialized logic
    - `impact-analyzer`: Assess debt impact on velocity and quality
      - Capabilities: Velocity impact modeling, quality degradation analysis
      - Why needed: Business impact assessment requires domain knowledge
    - `paydown-strategist`: Prioritize debt by value/effort ratio
      - Capabilities: Value/effort matrix, ROI calculation, sequencing
      - Why needed: Prioritization framework requires strategic thinking
    - `trend-tracker`: Track debt accumulation/paydown over time
      - Capabilities: Baseline comparison, trend analysis, velocity calculation
      - Why needed: Historical tracking requires specialized data structures
    - `prevention-advisor`: Suggest practices to prevent new debt
      - Capabilities: Best practice recommendations, prevention checklist
      - Why needed: Prevention requires domain-specific knowledge
  - Define agent name and specialization domain
  - Document capabilities the new agent provides
  - Explain why inherited agents are insufficient
  - **CREATE AGENT PROMPT FILE**: Write agents/{agent-name}.md
    - Include: agent role, technical debt-specific capabilities, input/output format
    - Include: specific instructions for this iteration's task
    - Include: technical debt domain knowledge (SQALE, code smells, prioritization frameworks)
  - Add to agent set: A_N = A_{N-1} ∪ {new_agent}

**Agent Invocation** (specialized or inherited):
- **READ agent prompt file** before invocation: agents/{agent-name}.md
- Invoke agent to execute concrete debt measurement work:
  - Measure technical debt (SQALE, code smells, debt ratio)
  - Identify debt hotspots (high-debt areas)
  - Assess debt impact (velocity, quality degradation)
  - Prioritize debt (value/effort matrix, ROI)
  - Track debt trends (accumulation, paydown velocity)
  - Create prevention guidelines (best practices)
- Produce iteration outputs (debt reports, prioritization matrix, paydown roadmap)

**Methodology Extraction** (M.evolve):
- **OBSERVE agent work patterns**:
  - How did agent organize debt measurement process?
  - What debt metrics were prioritized?
  - What decision criteria were used (when to measure, when to prioritize)?
  - What prioritization logic was applied?
- **EXTRACT patterns for methodology**:
  - Document debt measurement frameworks (SQALE implementation)
  - Build debt categorization taxonomy (code vs. architecture vs. process)
  - Identify reusable prioritization patterns (value/effort matrix)
  - Note principles that emerge
  - Add to methodology documentation

**ELSE use inherited agents:**
- **READ agent prompt file** from agents/{agent-name}.md
- Invoke appropriate agents from A_{N-1}
- Execute planned debt measurement work
- Observe for methodology patterns

**CRITICAL EXECUTION PROTOCOL**:
1. ALWAYS read capability files before embodying Meta-Agent capabilities
2. ALWAYS read agent prompt file before each agent invocation
3. Do NOT cache instructions across iterations - always read from files
4. Capability files may be updated between iterations - get latest from files
5. Never assume capabilities - always verify from source files

### 4. REFLECT (M.reflect)
- **READ** meta-agents/reflect.md for evaluation process
- **Evaluate Instance Layer** (Concrete Debt Measurement):
  - What debt dimensions were measured this iteration?
  - What debt hotspots were identified?
  - Calculate new V_instance(s_N):
    - V_measurement: Accuracy of debt quantification
      - Count debt dimensions measured (SQALE, code smells, coverage, complexity, duplication, change frequency, architecture, performance, security, dependencies)
      - Calculate: measured_dimensions / total_dimensions (target: 10 dimensions)
      - Assess SQALE implementation quality (if implemented)
      - Calculate: 0.0-1.0 based on coverage and accuracy
    - V_prioritization: Quality of debt ranking
      - Assess value/effort matrix completeness
      - Check ROI calculation quality
      - Validate business impact assessment
      - Calculate: prioritization_quality (0.0-1.0)
    - V_tracking: Visibility of debt trends
      - Check baseline comparison availability
      - Assess trend analysis capability
      - Validate paydown velocity tracking
      - Calculate: tracking_capability (0.0-1.0)
    - V_actionability: Clarity of paydown strategies
      - Count actionable recommendations in paydown roadmap
      - Assess effort estimation quality
      - Validate sequencing and dependency analysis
      - Calculate: actionable_recommendations / total_recommendations
    - **V_instance(s_N) = 0.3×V_measure + 0.3×V_prior + 0.2×V_track + 0.2×V_action**
  - Calculate change: ΔV_instance = V_instance(s_N) - V_instance(s_{N-1})
  - Are debt measurement objectives met? What gaps remain?

- **Evaluate Meta Layer** (Methodology):
  - What patterns were extracted this iteration?
  - Calculate new V_meta(s_N):
    - V_completeness: documented_patterns / total_patterns
      - Required: SQALE methodology, code smell taxonomy, prioritization framework, tracking system, paydown strategies, prevention guidelines
      - Count documented vs required (6 minimum)
    - V_effectiveness: 1 - (debt_measurement_time_with_methodology / debt_measurement_time_baseline)
      - Measure debt measurement time if methodology used
      - Compare to baseline (~8-10 hours for 5K lines manual)
      - Later iterations: Transfer test to different codebase
    - V_reusability: successful_transfers / transfer_attempts
      - Test methodology on different codebase (e.g., different Go project)
      - Assess if methodology applies with <20% modification
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
    details: "If no new debt measurement specialization needed"

  instance_value_threshold:
    question: "Is V_instance(s_N) ≥ 0.80 (debt management quality)?"
    V_instance(s_N): [calculated value]
    threshold_met: [Yes/No]
    components:
      V_measurement: [value] "≥0.80 target"
      V_prioritization: [value] "≥0.80 target"
      V_tracking: [value] "≥0.75 target"
      V_actionability: [value] "≥0.85 target"

  meta_value_threshold:
    question: "Is V_meta(s_N) ≥ 0.80 (methodology quality)?"
    V_meta(s_N): [calculated value]
    threshold_met: [Yes/No]
    components:
      V_completeness: [value] "≥0.90 target"
      V_effectiveness: [value] "≥0.75 target (4x speedup)"
      V_reusability: [value] "≥0.75 target (75% transfer)"

  instance_objectives:
    all_debt_dimensions_measured: [Yes/No] "SQALE, code smells, coverage, complexity, duplication, change frequency, architecture, performance, security, dependencies"
    prioritization_matrix_complete: [Yes/No]
    paydown_roadmap_created: [Yes/No]
    prevention_checklist_created: [Yes/No]
    trend_tracking_implemented: [Yes/No]
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
  - O = {debt reports + methodology documentation}
  - A_N = {final agent set with specializations}
  - M_N = {final meta-agent capabilities}

**IF NOT CONVERGED:**
- Identify what's needed for next iteration
- Note: Focus on instance (debt) OR meta (methodology) as needed
- Continue to Iteration N+1

## Documentation Requirements

Create experiments/bootstrap-012-technical-debt/iteration-N.md with:

### 1. Iteration Metadata
```yaml
iteration: N
date: YYYY-MM-DD
duration: ~X hours
status: [completed/converged]
layers:
  instance: "Technical debt measurement work performed this iteration"
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

  inherited_baseline: "3 agents from bootstrap-003 (data-analyst, doc-writer, coder)"

  IF evolved (new agent created):
    new_agents:
      - name: agent_name
        specialization: technical_debt_domain_area
        capabilities: [list]
        creation_reason: "Why were inherited 3 agents insufficient?"
        justification: "What gap exists that no inherited agent can fill?"
        prompt_file: "agents/{agent-name}.md"

  IF unchanged:
    status: "A_N = A_{N-1} (inherited agents sufficient for this iteration)"

  IF reused_inherited:
    reused_agents:
      - agent_name: task_performed
        adaptation_notes: "How inherited agent was adapted to technical debt context"

agents_invoked_this_iteration:
  - agent_name: task_performed
    source: [inherited / newly_created]
```

### 4. Instance Work Executed (Concrete Debt Measurement)
- What debt dimensions were measured?
  - SQALE index (if implemented)
  - Code smells (if detected)
  - Cyclomatic complexity (if analyzed)
  - Code duplication (if measured)
  - Test coverage (if calculated)
  - Change frequency (if tracked)
  - Architecture quality (if assessed)
  - Performance issues (if identified)
  - Security vulnerabilities (if scanned)
  - Dependency health (if checked)
- What debt hotspots were identified?
  - High-complexity modules
  - High-duplication areas
  - Low-coverage files
  - High-change-frequency hotspots
- What outputs were produced?
  - Debt reports (per-module)
  - Prioritization matrix (value vs. effort)
  - Paydown roadmap (sequencing, dependencies)
  - Prevention checklist (best practices)
  - Trend tracking dashboard (if implemented)
- Summary of concrete deliverables

### 5. Meta Work Executed (Methodology Extraction)
- What patterns were observed in debt measurement work?
  - Measurement decision patterns
  - Prioritization logic patterns
  - Tracking approaches
  - Prevention strategies
- What methodology content was documented?
  - SQALE methodology implementation
  - Debt categorization taxonomy
  - Prioritization frameworks
  - Tracking system design
  - Prevention guidelines

**Knowledge Artifacts Created**:

Organize extracted knowledge into appropriate categories:

- **Patterns** (domain-specific): knowledge/patterns/{pattern-name}.md
  - Specific solutions to recurring problems in technical debt management
  - Example: "SQALE-Based Debt Quantification Pattern", "Value-Effort Prioritization Pattern"
  - Format: Problem, Context, Solution, Consequences, Examples

- **Principles** (universal): knowledge/principles/{principle-name}.md
  - Fundamental truths or rules discovered
  - Example: "Pay High-Value Low-Effort Debt First Principle", "Prevent Before Paydown Principle"
  - Format: Statement, Rationale, Evidence, Applications

- **Templates** (reusable): knowledge/templates/{template-name}.{md|yaml|json}
  - Concrete implementations ready for reuse
  - Example: "Debt Prioritization Matrix Template", "Paydown Roadmap Template"
  - Format: Template file + usage documentation

- **Best Practices** (context-specific): knowledge/best-practices/{topic}.md
  - Recommended approaches for specific contexts
  - Example: "Go Code Complexity Best Practices", "Test Coverage Best Practices"
  - Format: Context, Recommendation, Justification, Trade-offs

- **Methodology** (project-wide): docs/methodology/{methodology-name}.md
  - Comprehensive guides for reuse across projects
  - Example: "Technical Debt Quantification Methodology for Software Projects"
  - Format: Complete methodology with decision frameworks

**Knowledge Index Update**:
- Update knowledge/INDEX.md with:
  - New knowledge entries
  - Links to iteration where extracted
  - Domain tags (technical-debt, sqale, prioritization, etc.)
  - Validation status (proposed, validated, refined)

### 6. State Transition

**Instance Layer** (Debt Measurement State):
```yaml
s_{N-1} → s_N (Technical Debt):
  changes:
    - debt_dimensions_measured: [list]
    - debt_hotspots_identified: [list]
    - prioritization_matrix_created: [Yes/No]
    - paydown_roadmap_created: [Yes/No]

  metrics:
    V_measurement: [value] (was: [previous])
    V_prioritization: [value] (was: [previous])
    V_tracking: [value] (was: [previous])
    V_actionability: [value] (was: [previous])

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
  - Instance learnings (debt measurement insights)
  - Meta learnings (methodology insights)
- What worked well?
- What challenges were encountered?
- What is needed next?
  - For debt measurement completion
  - For methodology completion

### 8. Convergence Check
[Use the convergence criteria structure above]

### 9. Data Artifacts

**Ephemeral Data** (iteration-specific, saved to data/):
- data/iteration-N-metrics.json (V_instance, V_meta calculations)
- data/iteration-N-debt-state.yaml (debt dimensions measured, hotspots identified)
- data/iteration-N-methodology.yaml (extracted patterns)
- data/iteration-N-artifacts/ (debt reports, prioritization matrix, paydown roadmap)

**Permanent Knowledge** (cumulative, saved to knowledge/):
- knowledge/patterns/{pattern-name}.md (domain-specific patterns)
- knowledge/principles/{principle-name}.md (universal principles)
- knowledge/templates/{template-name}.{md|yaml|json} (reusable templates)
- knowledge/best-practices/{topic}.md (context-specific practices)
- knowledge/INDEX.md (knowledge catalog with iteration links)

**Project-Wide Methodology** (saved to docs/methodology/):
- docs/methodology/{methodology-name}.md (comprehensive reusable guides)

**Knowledge Organization Principles**:
1. **Separation**: Distinguish ephemeral data from permanent knowledge
2. **Categorization**: Use appropriate knowledge category (pattern/principle/template/best-practice/methodology)
3. **Indexing**: Always update knowledge/INDEX.md when adding knowledge
4. **Linking**: Link knowledge to source iteration for traceability
5. **Validation**: Track validation status (proposed → validated → refined)

Reference data files and knowledge artifacts in iteration document.

## Key Principles

1. **Two-Layer Awareness**: Always work on both layers
   - Instance: Perform concrete technical debt quantification
   - Meta: Extract methodology patterns from debt measurement work

2. **Be Honest**: Calculate V(s_N) based on actual state
   - Don't inflate instance values (debt management quality)
   - Don't inflate meta values (methodology quality)

3. **Let System Evolve**: Don't force predetermined paths
   - Create agents based on need, not plan
   - Extract patterns that actually emerge
   - Don't fabricate methodology to meet targets

4. **Justify Specialization**: Only create agents when inherited insufficient
   - Document clear reason for specialization
   - Explain what inherited agent couldn't do

5. **Document Evolution**: Clearly explain WHY M or A evolved
   - What triggered agent creation?
   - What capability gap was identified?

6. **Check Convergence Rigorously**: Evaluate both layers
   - Instance convergence: Debt management quality + completeness
   - Meta convergence: Methodology quality + transferability

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

## Technical Debt Domain-Specific Patterns

Based on OCA framework, technical debt iterations may follow:

### Observe Phase (Iterations 0-2)
- Iteration 0: Baseline establishment, codebase inventory, initial metrics
- Iteration 1: SQALE methodology implementation, comprehensive debt measurement
- Iteration 2: Debt hotspot identification, pattern analysis, trend tracking setup
- **Methodology**: Observe debt patterns, identify measurement frameworks

### Codify Phase (Iterations 3-4)
- Iteration 3: Build debt categorization taxonomy, prioritization framework
- Iteration 4: Create paydown roadmap, prevention checklist, best practices
- **Methodology**: Extract debt management frameworks, codify patterns

### Automate Phase (Iterations 5-6)
- Iteration 5: Implement automated debt tracking, trend monitoring dashboard
- Iteration 6: Transfer test to different codebase, validate methodology, convergence
- **Methodology**: Document automation strategies, validate transferability

**Caveat**: Let actual needs drive the sequence, not this expected pattern. The pattern is a hint, not a prescription.
```

---

## Quick Reference: Iteration Checklist

For each iteration N ≥ 1, ensure you:

**Preparation**:
- [ ] Review previous iteration (iteration-[N-1].md)
- [ ] Extract current state (M_{N-1}, A_{N-1}, V_instance(s_{N-1}), V_meta(s_{N-1}))

**Observe Phase**:
- [ ] **READ** meta-agents/observe.md (technical debt observation strategies)
- [ ] Analyze debt measurement state (dimensions measured, hotspots identified, patterns observed)
- [ ] Identify gaps (dimensions not measured, measurement aspects missing, methodology gaps)
- [ ] Observe patterns (for methodology extraction)

**Plan Phase**:
- [ ] **READ** meta-agents/plan.md (prioritization, agent selection)
- [ ] Define iteration goal (instance layer: debt measurement, prioritization, tracking)
- [ ] Assess agent sufficiency (inherited sufficient OR need specialized?)
- [ ] Plan methodology extraction (what patterns to document?)

**Execute Phase**:
- [ ] **READ** meta-agents/execute.md (coordination, pattern observation)
- [ ] **IF NEW AGENT NEEDED**:
  - [ ] **READ** meta-agents/evolve.md (agent creation criteria)
  - [ ] Create agent prompt file: agents/{agent-name}.md
  - [ ] Document specialization reason
- [ ] **READ** agent prompt file(s) before invocation: agents/{agent-name}.md
- [ ] Invoke agents to perform debt measurement
- [ ] Observe debt measurement work for methodology patterns
- [ ] Extract patterns and update methodology documentation

**Reflect Phase**:
- [ ] **READ** meta-agents/reflect.md (evaluation process)
- [ ] Calculate V_instance(s_N) (debt management quality):
  - [ ] V_measurement (measured_dimensions / total_dimensions)
  - [ ] V_prioritization (prioritization_quality)
  - [ ] V_tracking (tracking_capability)
  - [ ] V_actionability (actionable_recommendations / total_recommendations)
- [ ] Calculate V_meta(s_N) (methodology quality):
  - [ ] V_completeness (documented_patterns / total_patterns)
  - [ ] V_effectiveness (1 - measurement_time_with / measurement_time_baseline)
  - [ ] V_reusability (successful_transfers / transfer_attempts)
- [ ] Assess quality honestly (don't inflate values)
- [ ] Identify gaps (debt + methodology)

**Convergence Check**:
- [ ] M_N == M_{N-1}? (meta-agent stable)
- [ ] A_N == A_{N-1}? (agent set stable)
- [ ] V_instance(s_N) ≥ 0.80? (debt management quality threshold)
- [ ] V_meta(s_N) ≥ 0.80? (methodology quality threshold)
- [ ] Instance objectives complete? (all debt dimensions measured, prioritization matrix complete)
- [ ] Meta objectives complete? (methodology documented, transfer test successful)
- [ ] Determine: CONVERGED or NOT_CONVERGED

**Documentation**:
- [ ] Create iteration-N.md with:
  - [ ] Metadata (iteration, date, duration, status)
  - [ ] M evolution (evolved or unchanged)
  - [ ] A evolution (new agents or unchanged)
  - [ ] Instance work (debt measured, hotspots identified, prioritization created)
  - [ ] Meta work (patterns extracted)
  - [ ] Knowledge artifacts created (list all new knowledge)
  - [ ] State transition (V_instance, V_meta calculated)
  - [ ] Reflection (learnings, next steps)
  - [ ] Convergence check (criteria evaluated)
- [ ] Save data artifacts to data/:
  - [ ] iteration-N-metrics.json
  - [ ] iteration-N-debt-state.yaml
  - [ ] iteration-N-methodology.yaml
  - [ ] iteration-N-artifacts/ (debt reports, prioritization matrix, paydown roadmap)
- [ ] Save knowledge artifacts:
  - [ ] Create/update knowledge files in appropriate categories
  - [ ] Update knowledge/INDEX.md with new entries
  - [ ] Link knowledge to iteration source
  - [ ] Tag knowledge by domain
  - [ ] Mark validation status

**Quality Assurance**:
- [ ] **NO TOKEN LIMITS**: Verify all steps completed fully without abbreviation
- [ ] Capability files read before use
- [ ] Agent files read before invocation
- [ ] Honest value calculations (not inflated)
- [ ] Both layers addressed (instance + meta)

**If CONVERGED**:
- [ ] Create results.md with 10-point analysis
- [ ] Perform reusability validation (transfer test to different codebase)
- [ ] Document three-tuple: (O={debt reports+methodology}, A_N, M_N)
- [ ] Validate methodology alignment (OCA, Bootstrapped SE, Value Space)

---

## Notes on Execution Style

**Be the Meta-Agent**: When executing iterations, embody M's perspective:
- Think through the observe-plan-execute-reflect-evolve cycle
- Read capability files to understand M's strategies
- Make explicit decisions about agent creation
- Justify why specialization is needed
- Extract methodology patterns from debt measurement work
- Track both V_instance and V_meta

**Be Domain-Aware**: Technical debt has specific concerns:
- SQALE methodology (industry standard for debt quantification)
- Code smells (maintainability issues, anti-patterns)
- Technical debt ratio (debt / total development cost)
- Value/effort prioritization (ROI-driven paydown)
- Debt accumulation rate (is debt growing or shrinking?)
- Prevention strategies (avoid new debt)
- Go language specifics (gocyclo, staticcheck, dupl)

**Be Rigorous**: Calculate values honestly
- V_instance based on actual debt measurement state (not aspirational)
- V_meta based on actual methodology coverage (not desired)
- Don't force convergence prematurely
- Don't skip methodology extraction to focus only on debt measurement
- Let data and needs drive the process

**Be Thorough**: Document decisions and reasoning
- Save intermediate data (metrics, debt reports, prioritization matrices)
- Show your work (calculations, analysis)
- Make evolution path traceable
- **NO TOKEN LIMITS**: Complete all steps fully, never abbreviate

**Be Two-Layer Aware**: Always work on both layers
- Instance layer: Perform concrete technical debt quantification
- Meta layer: Extract reusable methodology
- Don't neglect either layer
- Methodology extraction is as important as debt measurement

**Be Authentic**: This is a real experiment
- Discover debt patterns, don't assume them
- Create agents based on need, not predetermined plan
- Extract methodology from actual debt measurement work, not theory
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
**Created**: 2025-10-17
**Last Updated**: 2025-10-17
**Purpose**: Guide authentic execution of bootstrap-012-technical-debt experiment with dual-layer architecture

**Key Innovation**: Dual value functions (V_instance + V_meta) enable simultaneous optimization of concrete technical debt quantification and reusable debt management methodology.
