# Iteration Execution Prompts

This document provides structured prompts for executing each iteration of the bootstrap-009-observability-methodology experiment.

**Two-Layer Architecture**:
- **Instance Layer**: Agents add comprehensive observability instrumentation to meta-cc MCP server
- **Meta Layer**: Meta-Agent observes instrumentation patterns and extracts observability methodology

---

## Iteration 0: Baseline Establishment

```markdown
# Execute Iteration 0: Baseline Establishment

## Context
I'm starting the bootstrap-009-observability-methodology experiment. I've reviewed:
- experiments/bootstrap-009-observability-methodology/plan.md
- experiments/bootstrap-009-observability-methodology/README.md
- experiments/EXPERIMENTS-OVERVIEW.md (Bootstrap-009 section)
- The three methodology frameworks (OCA, Bootstrapped SE, Value Space Optimization)

## Current State
- Meta-Agent: M₀ (5 capabilities inherited from Bootstrap-003: observe, plan, execute, reflect, evolve)
- Agent Set: A₀ (6 agents inherited from Bootstrap-003: 3 generic + 3 specialized)
- Target: cmd/mcp/ server and internal/ modules (~2,000 lines to instrument)

## Inherited State from Bootstrap-003

**IMPORTANT**: This experiment starts with the converged state from Bootstrap-003, NOT from scratch.

**Meta-Agent capability files (ALREADY EXIST)**:
- experiments/bootstrap-009-observability-methodology/meta-agents/observe.md (validated)
- experiments/bootstrap-009-observability-methodology/meta-agents/plan.md (validated)
- experiments/bootstrap-009-observability-methodology/meta-agents/execute.md (validated)
- experiments/bootstrap-009-observability-methodology/meta-agents/reflect.md (validated)
- experiments/bootstrap-009-observability-methodology/meta-agents/evolve.md (validated)

**Agent prompt files (ALREADY EXIST - 6 agents)**:
- Generic (3): data-analyst.md, doc-writer.md, coder.md
- From Bootstrap-003 (3): error-classifier.md, recovery-advisor.md, root-cause-analyzer.md

**CRITICAL EXECUTION PROTOCOL**:
- All capability files and agent files ALREADY EXIST (inherited from Bootstrap-003)
- Before embodying Meta-Agent capabilities, ALWAYS read the relevant capability file first
- Before invoking ANY agent, ALWAYS read its prompt file first
- These files contain validated capabilities/agents; adapt them to observability context
- Never assume capabilities - always read from the source files

## Iteration 0 Objectives

Execute baseline establishment:

0. **Setup** (Verify inherited state):
   - **VERIFY META-AGENT CAPABILITY FILES EXIST** (inherited from Bootstrap-003):
     - ✓ meta-agents/observe.md: Data collection, pattern discovery (validated)
     - ✓ meta-agents/plan.md: Prioritization, agent selection (validated)
     - ✓ meta-agents/execute.md: Agent coordination, task execution (validated)
     - ✓ meta-agents/reflect.md: Value calculation (V_instance, V_meta), gap analysis (validated)
     - ✓ meta-agents/evolve.md: Agent creation criteria, methodology extraction (validated)
   - **VERIFY INITIAL AGENT PROMPT FILES EXIST** (inherited from Bootstrap-003):
     - ✓ agents/data-analyst.md, doc-writer.md, coder.md (generic agents)
     - ✓ agents/error-classifier.md, recovery-advisor.md, root-cause-analyzer.md (from Bootstrap-003)
   - **NO NEED TO CREATE NEW FILES** - all files inherited and ready to use
   - **ADAPTATION NOTE**: Capability files and agent files are generic enough to apply to observability;
     read them to understand their validated approaches, then apply to observability context

1. **Codebase Analysis** (M₀.observe):
   - **READ** meta-agents/observe.md (observability assessment strategies)
   - Analyze current observability state:
     - List all modules in cmd/mcp/ and internal/
     - Count lines of code per module
     - Identify critical code paths (main, handlers, query engine, parser)
     - Examine existing logging statements (if any)
     - Review error handling patterns
     - Check for existing metrics or tracing
   - Target modules for instrumentation:
     - cmd/mcp/server.go: MCP server main
     - cmd/mcp/tools.go: 16 MCP tool handlers
     - cmd/mcp/executor.go: Query execution
     - internal/parser/: Session history parsing
     - internal/analyzer/: Pattern detection
     - internal/query/: Query engine
   - Critical paths to instrument (90% coverage target):
     - Tool invocation paths (all 16 tools)
     - Query execution paths (parser → analyzer → output)
     - Error handling paths (query failures, parse errors)
     - Performance-critical sections (large file processing)

2. **Current Observability Assessment** (M₀.observe + M₀.plan):
   - **READ** meta-agents/observe.md (observation strategies)
   - **READ** meta-agents/plan.md (assessment frameworks)
   - What observability currently exists?
     - Logging: Ad-hoc fmt.Printf or log statements? (baseline)
     - Metrics: None (need instrumentation)
     - Tracing: None (need design)
     - Structured logging: None (need structured format)
     - Dashboards: None (need operational visibility)
   - What observability artifacts exist?
     - Error messages in code (unstructured)
     - Debug statements (inconsistent)
     - No operational metrics
   - What gaps exist?
     - No structured logging framework
     - No metrics instrumentation (RED, USE, Four Golden Signals)
     - No distributed tracing
     - No operational dashboards
     - No alerting rules
     - No log aggregation strategy

3. **Baseline Metrics Calculation** (M₀.plan + data-analyst):
   - **READ** meta-agents/plan.md (prioritization strategies)
   - **READ** agents/data-analyst.md
   - Invoke data-analyst to calculate V_instance(s₀):
     - V_coverage: Estimate current observability coverage
       - Critical paths with logging: ~10% (minimal ad-hoc logging)
       - Critical paths with metrics: 0% (no metrics)
       - Critical paths with tracing: 0% (no tracing)
       - Calculate: instrumented_paths / total_critical_paths = 0.10
     - V_actionability: Assess diagnostic effectiveness
       - Time to diagnose issues: ~hours (manual log inspection)
       - Error context available: ~20% (minimal context)
       - Performance bottleneck visibility: 0% (no metrics)
       - Calculate: diagnostic_speed_factor = 0.20
     - V_performance: Assess observability overhead
       - Logging overhead: Minimal (few logs, but unoptimized)
       - Metrics overhead: 0% (none exist)
       - Estimated overhead if instrumented: ~5% (acceptable baseline)
       - Calculate: 1 - (overhead / acceptable_threshold) = 0.95
     - V_consistency: Assess pattern consistency
       - Logging format consistency: ~10% (ad-hoc formats)
       - Naming conventions: 0% (no conventions)
       - Metric standards: 0% (no metrics)
       - Calculate: consistent_patterns / total_patterns = 0.10
     - **V_instance(s₀) = 0.3×0.10 + 0.3×0.20 + 0.2×0.95 + 0.2×0.10 = 0.30**
   - Calculate V_meta(s₀):
     - V_completeness: 0.00 (no methodology yet)
     - V_effectiveness: 0.00 (nothing to test)
     - V_reusability: 0.00 (nothing to transfer)
     - **V_meta(s₀) = 0.00**

4. **Gap Identification** (M₀.reflect):
   - **READ** meta-agents/reflect.md (gap analysis process)
   - What observability capabilities are missing?
     - No structured logging framework (need: structured JSON logs)
     - No metrics instrumentation (need: RED metrics, USE metrics)
     - No distributed tracing (need: trace IDs, spans)
     - No operational dashboards (need: key metrics visualization)
     - No alerting rules (need: actionable alerts)
   - What instrumentation aspects need coverage?
     - Logging levels: DEBUG, INFO, WARN, ERROR (need standards)
     - Log context: Request IDs, user context, timing (need enrichment)
     - Metrics types: Counters, gauges, histograms (need implementation)
     - Metric dimensions: Tool name, error type, duration (need definition)
     - Trace spans: Request flow, query execution (need design)
     - Dashboard panels: Throughput, latency, errors, saturation (need creation)
   - What methodology components are needed?
     - Observability strategy framework
     - Logging standards and patterns
     - Metrics selection criteria (RED, USE, Four Golden Signals)
     - Tracing design patterns
     - Dashboard design principles
     - Alerting rule definition process

5. **Data Collection** (M₀.observe):
   - **READ** meta-agents/observe.md (data collection strategies)
   - Use meta-cc CLI to gather observability-related data:
     - meta-cc query-user-messages --pattern "log|error|debug|trace|metric|monitor"
       - How often are observability issues mentioned?
       - What kind of visibility problems occur?
     - meta-cc query-tools --status error
       - What errors need better visibility?
       - What error patterns exist?
     - meta-cc query-files --threshold 5
       - What high-touch files need logging?
       - What modules are error-prone?
   - Analyze git log:
     - Search for logging-related commits
     - Identify logging evolution patterns
     - Note existing observability attempts

6. **Initial Agent Applicability Assessment** (M₀.plan):
   - **READ** meta-agents/plan.md (agent selection strategies)
   - Which inherited agents are directly applicable to observability?
     - ⭐⭐ error-classifier: Classify error types for logging categorization
     - ⭐ root-cause-analyzer: Identify what needs instrumentation for diagnosis
     - ⭐ data-analyst: Analyze existing log patterns, metrics needs
     - coder: Write instrumentation code (logging, metrics)
     - doc-writer: Document observability standards
   - Which inherited agents need adaptation for observability?
     - error-classifier: Apply to log message classification
     - root-cause-analyzer: Use to identify instrumentation gaps
     - Read agent files to understand capabilities, adapt prompts contextually
   - What new specialized agents might be needed?
     - log-analyzer: Analyze existing log statements, identify patterns (likely needed)
     - metric-designer: Design meaningful metrics (RED, USE, Four Golden Signals) (likely needed)
     - trace-architect: Design distributed tracing strategy (may be needed)
     - dashboard-builder: Create operational dashboard specifications (may be needed)
     - alert-definer: Define actionable alert rules (may be needed)
   - **NOTE**: Don't create new agents yet - just identify potential needs

7. **Documentation** (M₀.execute + doc-writer):
   - **READ** meta-agents/execute.md (coordination strategies)
   - **READ** agents/doc-writer.md
   - Invoke doc-writer to create iteration-0.md:
     - M₀ state: 5 capabilities inherited from Bootstrap-003 (observe, plan, execute, reflect, evolve)
     - A₀ state: 6 agents inherited from Bootstrap-003 (3 generic + 3 specialized)
     - Agent applicability assessment (which agents useful for observability)
     - Codebase analysis summary (cmd/mcp/ + internal/ modules, ~2,000 lines to instrument)
     - Current observability assessment (minimal ad-hoc logging, no metrics/tracing)
     - Calculated V_instance(s₀) = 0.30 and V_meta(s₀) = 0.00
     - Gap analysis (missing logging framework, metrics, tracing, dashboards, methodology)
     - Reflection on next steps and agent reuse strategy
   - Save data artifacts:
     - data/s0-codebase-structure.yaml (module breakdown, line counts, critical paths)
     - data/s0-metrics.json (calculated V_instance and V_meta values)
     - data/s0-observability-gaps.yaml (identified gaps in observability)
     - data/s0-agent-applicability.yaml (inherited agents and their observability applicability)
     - data/s0-error-patterns.jsonl (error patterns needing visibility)
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
   - Are M₀ capabilities sufficient for baseline? (Yes, core capabilities adequate)
   - What should be the focus of Iteration 1?
     - Likely: Analyze existing logging patterns, design structured logging framework
     - Or: Design metrics instrumentation strategy (RED, USE metrics)
     - Decision based on OCA framework: Start with Observe phase (analyze existing state)
   - Which inherited agents will be most useful in Iteration 1?
     - Likely: Need new log-analyzer agent (systematic log pattern analysis)
     - Reuse: data-analyst (pattern analysis), coder (implementation)
   - Methodology extraction readiness:
     - Note patterns observed in existing code
     - Identify instrumentation decision points that could become methodology
     - Prepare for pattern documentation in subsequent iterations

## Constraints
- Do NOT pre-decide what agents to create next
- Do NOT assume the instrumentation approach or evolution path
- Let the codebase data and gaps guide next steps
- Be honest about current observability state (minimal, ad-hoc, inconsistent)
- Calculate V(s₀) based on actual observations, not target values
- Remember two layers: concrete instrumentation (instance) + methodology (meta)

## Output Format
Create iteration-0.md following this structure:
- Iteration metadata (number, date, duration)
- M₀ state documentation (5 capabilities inherited from Bootstrap-003)
- A₀ state documentation (6 agents inherited: 3 generic + 3 specialized)
- Agent applicability assessment (which agents useful for observability domain)
- Codebase analysis (cmd/mcp/ + internal/ modules, critical paths, line counts)
- Current observability assessment (minimal ad-hoc logging, no metrics/tracing)
- Value calculation (V_instance(s₀) = 0.30, V_meta(s₀) = 0.00)
- Gap identification (missing logging framework, metrics, tracing, dashboards, methodology)
- Reflection on next steps and agent reuse strategy
- Data artifacts saved to data/ directory
```

---

## Iteration 1+: Subsequent Iterations (General Template)

```markdown
# Execute Iteration N: [To be determined by Meta-Agent]

## Context from Previous Iteration

Review the previous iteration file: experiments/bootstrap-009-observability-methodology/iteration-[N-1].md

Extract:
- Current Meta-Agent state: M_{N-1}
- Current Agent Set: A_{N-1}
- Current observability state: V_instance(s_{N-1})
- Current methodology state: V_meta(s_{N-1})
- Problems identified
- Reflection notes on what's needed next

## Two-Layer Execution Protocol

**Layer 1 (Instance)**: Agents perform concrete observability instrumentation
**Layer 2 (Meta)**: Meta-Agent observes and extracts methodology

Throughout iteration:
- Agents focus on concrete tasks (add logging, define metrics, implement tracing)
- Meta-Agent observes instrumentation work and identifies patterns for methodology

## Meta-Agent Decision Process

**BEFORE STARTING**: Read relevant Meta-Agent capability files:
- **READ** meta-agents/observe.md (for observation strategies)
- **READ** meta-agents/plan.md (for planning and decisions)
- **READ** meta-agents/execute.md (for coordination)
- **READ** meta-agents/reflect.md (for evaluation)
- **READ** meta-agents/evolve.md (for evolution assessment)

As M_{N-1}, follow the five-capability process:

### 1. OBSERVE (M.observe)
- **READ** meta-agents/observe.md for observability assessment strategies
- Review previous iteration outputs (iteration-[N-1].md)
- Examine observability state:
  - What instrumentation has been added? (if any)
  - What logging patterns have been implemented? (if any)
  - What metrics have been defined? (if any)
  - What tracing design exists? (if any)
- Identify gaps:
  - What critical paths still need instrumentation?
  - What observability aspects are missing? (logging, metrics, tracing, dashboards, alerts)
  - What consistency issues exist?
  - What methodology patterns are emerging?
- **Methodology observation**:
  - What patterns emerged in previous instrumentation work?
  - What instrumentation decisions were made and why?
  - What principles can be extracted?

### 2. PLAN (M.plan)
- **READ** meta-agents/plan.md for prioritization and agent selection
- Based on observations, what is the primary goal for this iteration?
  - Examples:
    - "Analyze existing log patterns and design structured logging framework"
    - "Design and implement RED metrics for MCP tools"
    - "Add distributed tracing instrumentation to query engine"
    - "Create operational dashboard specifications"
    - "Define actionable alerting rules"
- What capabilities are needed to achieve this goal?
- **Agent Assessment**:
  - Are current agents (A_{N-1}) sufficient for this goal?
  - Can inherited agents handle observability? (error-classifier, data-analyst, etc.)
  - Or do we need specialized `log-analyzer` for systematic pattern analysis?
  - Do we need `metric-designer` for metrics selection (RED, USE, Four Golden Signals)?
  - Do we need `trace-architect` for tracing design?
  - Do we need `dashboard-builder` for dashboard specifications?
  - Do we need `alert-definer` for alerting rules?
- **Methodology Planning**:
  - What patterns should be documented this iteration?
  - What instrumentation decisions will inform methodology?

### 3. EXECUTE (M.execute)
- **READ** meta-agents/execute.md for coordination and pattern observation
- Decision point: Should I create a new specialized agent?

**IF current agents are insufficient:**
- **EVOLVE** (M.evolve): Create new specialized agent
  - **READ** meta-agents/evolve.md for agent creation criteria
  - Examples of specialized agents:
    - `log-analyzer`: Systematic log pattern analysis
      - Capabilities: Parse logs, identify patterns, classify messages, recommend standards
      - Why needed: Generic agents insufficient for log-specific analysis
    - `metric-designer`: Metrics selection and design
      - Capabilities: Apply RED/USE frameworks, define metric dimensions, recommend dashboards
      - Why needed: Specialized metrics knowledge required
    - `trace-architect`: Distributed tracing design
      - Capabilities: Design trace spans, define trace context, recommend instrumentation points
      - Why needed: Tracing-specific expertise required
    - `dashboard-builder`: Operational dashboard design
      - Capabilities: Design dashboard layouts, select key metrics, define alert thresholds
      - Why needed: Operational visibility expertise required
    - `alert-definer`: Alerting rule definition
      - Capabilities: Define actionable alerts, set thresholds, prevent alert fatigue
      - Why needed: Alerting best practices required
  - Define agent name and specialization domain
  - Document capabilities the new agent provides
  - Explain why inherited agents are insufficient
  - **CREATE AGENT PROMPT FILE**: Write agents/{agent-name}.md
    - Include: agent role, observability-specific capabilities, input/output format
    - Include: specific instructions for this iteration's task
    - Include: observability domain knowledge (logging levels, metric types, tracing patterns)
  - Add to agent set: A_N = A_{N-1} ∪ {new_agent}

**Agent Invocation** (specialized or inherited):
- **READ agent prompt file** before invocation: agents/{agent-name}.md
- Invoke agent to execute concrete instrumentation work:
  - Add structured logging to critical paths
  - Define and implement metrics (RED, USE, Four Golden Signals)
  - Design and add distributed tracing
  - Create operational dashboard specifications
  - Define actionable alerting rules
- Produce iteration outputs (instrumented code, logging standards, metrics definitions, dashboards)

**Methodology Extraction** (M.evolve):
- **OBSERVE agent work patterns**:
  - How did agent organize instrumentation process?
  - What logging levels were chosen and why?
  - What metric types were selected (counter, gauge, histogram)?
  - What decision criteria were used (when to log, what to measure)?
  - What prioritization logic was applied?
- **EXTRACT patterns for methodology**:
  - Document instrumentation decision frameworks
  - Build logging pattern library
  - Identify reusable metrics patterns
  - Note principles that emerge
  - Add to methodology documentation

**ELSE use inherited agents:**
- **READ agent prompt file** from agents/{agent-name}.md
- Invoke appropriate agents from A_{N-1}
- Execute planned instrumentation work
- Observe for methodology patterns

**CRITICAL EXECUTION PROTOCOL**:
1. ALWAYS read capability files before embodying Meta-Agent capabilities
2. ALWAYS read agent prompt file before each agent invocation
3. Do NOT cache instructions across iterations - always read from files
4. Capability files may be updated between iterations - get latest from files
5. Never assume capabilities - always verify from source files

### 4. REFLECT (M.reflect)
- **READ** meta-agents/reflect.md for evaluation process
- **Evaluate Instance Layer** (Concrete Instrumentation):
  - What instrumentation was added this iteration?
  - What critical paths are now observable?
  - Calculate new V_instance(s_N):
    - V_coverage: instrumented_paths / total_critical_paths
      - Count instrumented paths (logging + metrics + tracing)
      - Estimate total critical paths (~90% target)
    - V_actionability: diagnostic_speed_factor
      - Estimate time to diagnose issues (hours → minutes with good logs/metrics)
      - Assess error context availability (minimal → rich context)
      - Calculate improvement: 1 - (new_diagnosis_time / baseline_diagnosis_time)
    - V_performance: 1 - (overhead / acceptable_threshold)
      - Measure logging overhead (should be <5%)
      - Measure metrics overhead (should be <3%)
      - Measure tracing overhead (should be <2%)
      - Acceptable threshold: 10% total
    - V_consistency: consistent_patterns / total_patterns
      - Count consistent logging patterns (structured format, naming conventions)
      - Count consistent metric patterns (naming, labels)
      - Assess overall pattern consistency
    - **V_instance(s_N) = 0.3×V_coverage + 0.3×V_actionability + 0.2×V_performance + 0.2×V_consistency**
  - Calculate change: ΔV_instance = V_instance(s_N) - V_instance(s_{N-1})
  - Are instrumentation objectives met? What gaps remain?

- **Evaluate Meta Layer** (Methodology):
  - What patterns were extracted this iteration?
  - Calculate new V_meta(s_N):
    - V_completeness: documented_patterns / total_patterns
      - Required: Logging standards, metrics framework, tracing design, dashboard principles, alerting rules, instrumentation strategy
      - Count documented vs required (6 minimum)
    - V_effectiveness: 1 - (instrumentation_time_with_methodology / instrumentation_time_baseline)
      - Measure instrumentation time if methodology used
      - Compare to baseline (~40-60 hours for 2,000 lines)
      - Later iterations: Transfer test to cmd/ package
    - V_reusability: successful_transfers / transfer_attempts
      - Test methodology on different codebase (cmd/ package or other services)
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
    details: "If no new instrumentation specialization needed"

  instance_value_threshold:
    question: "Is V_instance(s_N) ≥ 0.80 (observability quality)?"
    V_instance(s_N): [calculated value]
    threshold_met: [Yes/No]
    components:
      V_coverage: [value] "≥0.85 target (90% critical paths)"
      V_actionability: [value] "≥0.75 target (minutes vs hours)"
      V_performance: [value] "≥0.80 target (<10% overhead)"
      V_consistency: [value] "≥0.80 target (consistent patterns)"

  meta_value_threshold:
    question: "Is V_meta(s_N) ≥ 0.80 (methodology quality)?"
    V_meta(s_N): [calculated value]
    threshold_met: [Yes/No]
    components:
      V_completeness: [value] "≥0.90 target"
      V_effectiveness: [value] "≥0.80 target (5x speedup)"
      V_reusability: [value] "≥0.85 target (90% transfer)"

  instance_objectives:
    logging_instrumented: [Yes/No] "Structured logging in 90% of critical paths"
    metrics_implemented: [Yes/No] "RED/USE metrics for all MCP tools"
    tracing_added: [Yes/No] "Distributed tracing in query engine"
    dashboards_created: [Yes/No] "Operational dashboards specified"
    alerts_defined: [Yes/No] "Actionable alerting rules defined"
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
  - O = {instrumented code + methodology documentation}
  - A_N = {final agent set with specializations}
  - M_N = {final meta-agent capabilities}

**IF NOT CONVERGED:**
- Identify what's needed for next iteration
- Note: Focus on instance (instrumentation) OR meta (methodology) as needed
- Continue to Iteration N+1

## Documentation Requirements

Create experiments/bootstrap-009-observability-methodology/iteration-N.md with:

### 1. Iteration Metadata
```yaml
iteration: N
date: YYYY-MM-DD
duration: ~X hours
status: [completed/converged]
layers:
  instance: "Observability instrumentation work performed this iteration"
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

  inherited_baseline: "6 agents from Bootstrap-003 (3 generic + 3 specialized)"

  IF evolved (new agent created):
    new_agents:
      - name: agent_name
        specialization: observability_domain_area
        capabilities: [list]
        creation_reason: "Why were inherited 6 agents insufficient?"
        justification: "What gap exists that no inherited agent can fill?"
        prompt_file: "agents/{agent-name}.md"

  IF unchanged:
    status: "A_N = A_{N-1} (inherited agents sufficient for this iteration)"

  IF reused_inherited:
    reused_agents:
      - agent_name: task_performed
        adaptation_notes: "How inherited agent was adapted to observability context"

agents_invoked_this_iteration:
  - agent_name: task_performed
    source: [inherited / newly_created]
```

### 4. Instance Work Executed (Concrete Instrumentation)
- What instrumentation was added?
  - Structured logging (if added)
  - Metrics instrumentation (if added)
  - Distributed tracing (if added)
  - Dashboard specifications (if created)
  - Alerting rules (if defined)
- What critical paths are now observable?
  - Tool invocation paths
  - Query execution paths
  - Error handling paths
  - Performance-critical sections
- What outputs were produced?
  - Instrumented code (logging, metrics, tracing)
  - Logging standards documentation
  - Metrics definitions (RED, USE)
  - Dashboard specifications
  - Alerting rules
- Summary of concrete deliverables

### 5. Meta Work Executed (Methodology Extraction)
- What patterns were observed in instrumentation work?
  - Logging decision patterns
  - Metrics selection patterns
  - Instrumentation prioritization logic
  - Dashboard design patterns
- What methodology content was documented?
  - Logging standards framework
  - Metrics selection criteria
  - Tracing design patterns
  - Dashboard design principles

**Knowledge Artifacts Created**:

Organize extracted knowledge into appropriate categories:

- **Patterns** (domain-specific): knowledge/patterns/{pattern-name}.md
  - Specific solutions to recurring problems in observability
  - Example: "Structured Logging Pattern", "RED Metrics Pattern", "Trace Context Propagation Pattern"
  - Format: Problem, Context, Solution, Consequences, Examples

- **Principles** (universal): knowledge/principles/{principle-name}.md
  - Fundamental truths or rules discovered
  - Example: "Low Overhead Principle", "Actionable Metrics Principle", "Context Enrichment Principle"
  - Format: Statement, Rationale, Evidence, Applications

- **Templates** (reusable): knowledge/templates/{template-name}.{md|yaml|json}
  - Concrete implementations ready for reuse
  - Example: "Logging Configuration Template", "Metrics Dashboard Template", "Alert Rule Template"
  - Format: Template file + usage documentation

- **Best Practices** (context-specific): knowledge/best-practices/{topic}.md
  - Recommended approaches for specific contexts
  - Example: "Go Logging Best Practices", "MCP Server Metrics Best Practices", "Trace Sampling Best Practices"
  - Format: Context, Recommendation, Justification, Trade-offs

- **Methodology** (project-wide): docs/methodology/{methodology-name}.md
  - Comprehensive guides for reuse across projects
  - Example: "Observability Methodology for Go Services"
  - Format: Complete methodology with decision frameworks

**Knowledge Index Update**:
- Update knowledge/INDEX.md with:
  - New knowledge entries
  - Links to iteration where extracted
  - Domain tags (observability, logging, metrics, tracing, etc.)
  - Validation status (proposed, validated, refined)

### 6. State Transition

**Instance Layer** (Observability State):
```yaml
s_{N-1} → s_N (Observability):
  changes:
    - instrumentation_added: [list]
    - critical_paths_covered: [list]
    - metrics_defined: [count]

  metrics:
    V_coverage: [value] (was: [previous])
    V_actionability: [value] (was: [previous])
    V_performance: [value] (was: [previous])
    V_consistency: [value] (was: [previous])

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
  - Instance learnings (observability insights)
  - Meta learnings (methodology insights)
- What worked well?
- What challenges were encountered?
- What is needed next?
  - For instrumentation completion
  - For methodology completion

### 8. Convergence Check
[Use the convergence criteria structure above]

### 9. Data Artifacts

**Ephemeral Data** (iteration-specific, saved to data/):
- data/iteration-N-metrics.json (V_instance, V_meta calculations)
- data/iteration-N-observability-state.yaml (instrumentation added, coverage)
- data/iteration-N-methodology.yaml (extracted patterns)
- data/iteration-N-artifacts/ (instrumented code, logging configs, metrics definitions)

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
   - Instance: Perform concrete observability instrumentation
   - Meta: Extract methodology patterns from instrumentation work

2. **Be Honest**: Calculate V(s_N) based on actual state
   - Don't inflate instance values (observability quality)
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
   - Instance convergence: Observability quality + coverage
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

## Observability Domain-Specific Patterns

Based on OCA framework, observability iterations may follow:

### Observe Phase (Iterations 0-2)
- Iteration 0: Baseline establishment, current observability assessment
- Iteration 1: Analyze existing log patterns, design structured logging framework
- Iteration 2: Design metrics strategy (RED, USE, Four Golden Signals), pattern identification
- **Methodology**: Observe instrumentation patterns, identify logging/metrics types

### Codify Phase (Iterations 3-4)
- Iteration 3: Implement structured logging and metrics instrumentation
- Iteration 4: Design distributed tracing, create dashboard specifications, document patterns
- **Methodology**: Extract instrumentation decision frameworks, codify patterns

### Automate Phase (Iterations 5-6)
- Iteration 5: Define alerting rules, create operational dashboards
- Iteration 6: Transfer test to cmd/ package, validate methodology, convergence
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
- [ ] **READ** meta-agents/observe.md (observability assessment strategies)
- [ ] Analyze observability state (instrumentation added, coverage, patterns)
- [ ] Identify gaps (critical paths not instrumented, observability aspects missing)
- [ ] Observe patterns (for methodology extraction)

**Plan Phase**:
- [ ] **READ** meta-agents/plan.md (prioritization, agent selection)
- [ ] Define iteration goal (instance layer: add logging, metrics, tracing, dashboards, alerts)
- [ ] Assess agent sufficiency (inherited sufficient OR need specialized?)
- [ ] Plan methodology extraction (what patterns to document?)

**Execute Phase**:
- [ ] **READ** meta-agents/execute.md (coordination, pattern observation)
- [ ] **IF NEW AGENT NEEDED**:
  - [ ] **READ** meta-agents/evolve.md (agent creation criteria)
  - [ ] Create agent prompt file: agents/{agent-name}.md
  - [ ] Document specialization reason
- [ ] **READ** agent prompt file(s) before invocation: agents/{agent-name}.md
- [ ] Invoke agents to perform instrumentation
- [ ] Observe instrumentation work for methodology patterns
- [ ] Extract patterns and update methodology documentation

**Reflect Phase**:
- [ ] **READ** meta-agents/reflect.md (evaluation process)
- [ ] Calculate V_instance(s_N) (observability quality):
  - [ ] V_coverage (instrumented_paths / total_critical_paths)
  - [ ] V_actionability (diagnostic_speed_factor)
  - [ ] V_performance (1 - overhead / acceptable_threshold)
  - [ ] V_consistency (consistent_patterns / total_patterns)
- [ ] Calculate V_meta(s_N) (methodology quality):
  - [ ] V_completeness (documented_patterns / total_patterns)
  - [ ] V_effectiveness (1 - instrumentation_time_with_methodology / baseline_time)
  - [ ] V_reusability (successful_transfers / transfer_attempts)
- [ ] Assess quality honestly (don't inflate values)
- [ ] Identify gaps (instrumentation + methodology)

**Convergence Check**:
- [ ] M_N == M_{N-1}? (meta-agent stable)
- [ ] A_N == A_{N-1}? (agent set stable)
- [ ] V_instance(s_N) ≥ 0.80? (observability quality threshold)
- [ ] V_meta(s_N) ≥ 0.80? (methodology quality threshold)
- [ ] Instance objectives complete? (logging, metrics, tracing, dashboards, alerts instrumented)
- [ ] Meta objectives complete? (methodology documented, transfer test successful)
- [ ] Determine: CONVERGED or NOT_CONVERGED

**Documentation**:
- [ ] Create iteration-N.md with:
  - [ ] Metadata (iteration, date, duration, status)
  - [ ] M evolution (evolved or unchanged)
  - [ ] A evolution (new agents or unchanged)
  - [ ] Instance work (instrumentation added, critical paths covered)
  - [ ] Meta work (patterns extracted)
  - [ ] Knowledge artifacts created (list all new knowledge)
  - [ ] State transition (V_instance, V_meta calculated)
  - [ ] Reflection (learnings, next steps)
  - [ ] Convergence check (criteria evaluated)
- [ ] Save data artifacts to data/:
  - [ ] iteration-N-metrics.json
  - [ ] iteration-N-observability-state.yaml
  - [ ] iteration-N-methodology.yaml
  - [ ] iteration-N-artifacts/ (instrumented code, configs, dashboards)
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
- [ ] Perform reusability validation (transfer test to cmd/ package)
- [ ] Document three-tuple: (O={instrumented code+methodology}, A_N, M_N)
- [ ] Validate methodology alignment (OCA, Bootstrapped SE, Value Space)

---

## Notes on Execution Style

**Be the Meta-Agent**: When executing iterations, embody M's perspective:
- Think through the observe-plan-execute-reflect-evolve cycle
- Read capability files to understand M's strategies
- Make explicit decisions about agent creation
- Justify why specialization is needed
- Extract methodology patterns from instrumentation work
- Track both V_instance and V_meta

**Be Domain-Aware**: Observability has specific concerns:
- Logging frameworks (structured logging, log levels, context enrichment)
- Metrics types (counters, gauges, histograms)
- Metrics frameworks (RED: Rate, Errors, Duration; USE: Utilization, Saturation, Errors; Four Golden Signals)
- Distributed tracing (trace IDs, spans, context propagation)
- Dashboard design (key metrics, layout, visualization)
- Alerting (actionable alerts, thresholds, alert fatigue prevention)
- Performance overhead (logging, metrics, tracing costs)

**Be Rigorous**: Calculate values honestly
- V_instance based on actual observability state (not aspirational)
- V_meta based on actual methodology coverage (not desired)
- Don't force convergence prematurely
- Don't skip methodology extraction to focus only on instrumentation
- Let data and needs drive the process

**Be Thorough**: Document decisions and reasoning
- Save intermediate data (metrics, instrumented code, dashboard specs)
- Show your work (calculations, analysis)
- Make evolution path traceable
- **NO TOKEN LIMITS**: Complete all steps fully, never abbreviate

**Be Two-Layer Aware**: Always work on both layers
- Instance layer: Perform concrete observability instrumentation
- Meta layer: Extract reusable methodology
- Don't neglect either layer
- Methodology extraction is as important as instrumentation execution

**Be Authentic**: This is a real experiment
- Discover observability patterns, don't assume them
- Create agents based on need, not predetermined plan
- Extract methodology from actual instrumentation work, not theory
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
**Purpose**: Guide authentic execution of bootstrap-009-observability-methodology experiment with dual-layer architecture

**Key Innovation**: Dual value functions (V_instance + V_meta) enable simultaneous optimization of concrete observability instrumentation and reusable methodology.
