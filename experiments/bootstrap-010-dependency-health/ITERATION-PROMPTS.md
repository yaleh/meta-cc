# Iteration Execution Prompts

This document provides structured prompts for executing each iteration of the bootstrap-010-dependency-health experiment.

**Two-Layer Architecture**:
- **Instance Layer**: Agents audit and update meta-cc project dependencies
- **Meta Layer**: Meta-Agent observes dependency management patterns and extracts methodology

---

## Iteration 0: Baseline Establishment

```markdown
# Execute Iteration 0: Baseline Establishment

## Context
I'm starting the bootstrap-010-dependency-health experiment. I've reviewed:
- experiments/bootstrap-010-dependency-health/plan.md
- experiments/bootstrap-010-dependency-health/README.md
- experiments/bootstrap-010-dependency-health/BOOTSTRAP-003-INHERITANCE.md
- The three methodology frameworks (OCA, Bootstrapped SE, Value Space Optimization)

## Current State
- Meta-Agent: M₀ (5 capabilities inherited from Bootstrap-003: observe, plan, execute, reflect, evolve)
- Agent Set: A₀ (6 agents inherited from Bootstrap-003: 3 generic + 3 specialized)
- Target: go.mod and all dependencies (~20 direct dependencies)

## Inherited State from Bootstrap-003

**IMPORTANT**: This experiment starts with the converged state from Bootstrap-003, NOT from scratch.

**Meta-Agent capability files (ALREADY EXIST)**:
- experiments/bootstrap-010-dependency-health/meta-agents/observe.md (validated)
- experiments/bootstrap-010-dependency-health/meta-agents/plan.md (validated)
- experiments/bootstrap-010-dependency-health/meta-agents/execute.md (validated)
- experiments/bootstrap-010-dependency-health/meta-agents/reflect.md (validated)
- experiments/bootstrap-010-dependency-health/meta-agents/evolve.md (validated)

**Agent prompt files (ALREADY EXIST - 6 agents)**:
- Generic (3): data-analyst.md, doc-writer.md, coder.md
- From Bootstrap-003 (3): error-classifier.md, recovery-advisor.md, root-cause-analyzer.md

**CRITICAL EXECUTION PROTOCOL**:
- All capability files and agent files ALREADY EXIST (inherited from Bootstrap-003)
- Before embodying Meta-Agent capabilities, ALWAYS read the relevant capability file first
- Before invoking ANY agent, ALWAYS read its prompt file first
- These files contain validated capabilities/agents; adapt them to dependency management context
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
     - ✓ agents/data-analyst.md (generic agent)
     - ✓ agents/doc-writer.md (generic agent)
     - ✓ agents/coder.md (generic agent)
     - ✓ agents/error-classifier.md (specialized agent from Bootstrap-003)
     - ✓ agents/recovery-advisor.md (specialized agent from Bootstrap-003)
     - ✓ agents/root-cause-analyzer.md (specialized agent from Bootstrap-003)
   - **NO NEED TO CREATE NEW FILES** - all files inherited and ready to use
   - **ADAPTATION NOTE**: Capability files and agent files are generic enough to apply to dependency management;
     read them to understand their validated approaches, then apply to dependency management context

1. **Dependency Landscape Analysis** (M₀.observe):
   - **READ** meta-agents/observe.md (dependency observation strategies)
   - Analyze go.mod structure:
     - Parse go.mod file
     - Count direct vs indirect dependencies
     - Build dependency graph using: `go list -m -json all`
     - Identify top-level dependencies (~20 direct)
     - Map transitive dependency tree
   - Analyze dependency characteristics:
     - go.sum (checksum inventory for all dependencies)
     - Dependency versions (semantic versioning analysis)
     - Release dates (identify stale dependencies)
     - Maintenance status (last commit, issue activity)
   - Review dependency update history:
     - Query: `meta-cc query-files --file go.mod` (dependency update patterns)
     - Git log analysis: `git log --all --oneline -- go.mod`
     - Historical update frequency and triggers
   - Identify critical dependencies:
     - Core functionality dependencies (MCP server, CLI)
     - Testing dependencies
     - Build/tooling dependencies
     - Security-sensitive dependencies

2. **Current Dependency Health Assessment** (M₀.observe + M₀.plan):
   - **READ** meta-agents/observe.md (observation strategies)
   - **READ** meta-agents/plan.md (assessment frameworks)
   - What dependency management process currently exists?
     - Manual go get updates (ad-hoc baseline)
     - Periodic go mod tidy (cleanup only)
     - No systematic vulnerability scanning
     - No license compliance checking
     - No automated update strategy
   - What dependency artifacts exist?
     - go.mod (dependency declarations)
     - go.sum (checksums for verification)
     - vendor/ directory (if vendoring enabled)
   - What gaps exist?
     - No vulnerability monitoring (CVE tracking)
     - No license inventory or compliance checks
     - No dependency freshness tracking
     - No update testing automation
     - No bloat detection (unnecessary dependencies)
     - No dependency health dashboard

3. **Baseline Security Scan** (M₀.plan + data-analyst):
   - **READ** meta-agents/plan.md (prioritization strategies)
   - **READ** agents/data-analyst.md
   - Invoke data-analyst to perform initial security assessment:
     - Run vulnerability scanner: `govulncheck ./...` (Go official tool)
     - Query GitHub Advisory Database for known CVEs
     - Query OSV (Open Source Vulnerabilities) database
     - Check deps.dev API for dependency insights
     - Identify high/medium/low severity vulnerabilities
     - Count vulnerable dependencies
     - Assess exploitability and impact

4. **Baseline Freshness Assessment** (M₀.plan + data-analyst):
   - **READ** agents/data-analyst.md
   - Invoke data-analyst to assess dependency staleness:
     - Compare current versions vs latest available:
       - Use: `go list -m -u all` (check for updates)
       - Categorize: up-to-date vs outdated vs severely outdated
     - Calculate age metrics:
       - Time since last update (per dependency)
       - Major version lag (how many majors behind)
       - Minor/patch version lag
     - Identify abandoned dependencies:
       - Last commit > 2 years
       - No recent releases
       - High issue count with no responses

5. **Baseline License Compliance Check** (M₀.plan + data-analyst):
   - **READ** agents/data-analyst.md
   - Invoke data-analyst to assess license compliance:
     - Extract licenses from dependencies:
       - Use go-licenses tool or manual LICENSE file parsing
       - SPDX identifier extraction
     - Categorize licenses:
       - Permissive (MIT, Apache, BSD)
       - Copyleft (GPL, LGPL, MPL)
       - Proprietary or restrictive
       - Unknown or missing licenses
     - Check compatibility:
       - meta-cc license: [identify project license]
       - Dependency license compatibility matrix
       - Identify license conflicts

6. **Baseline Metrics Calculation** (M₀.plan + data-analyst):
   - **READ** meta-agents/plan.md (prioritization strategies)
   - **READ** agents/data-analyst.md
   - Invoke data-analyst to calculate V_instance(s₀):
     - V_security: Vulnerability-free dependencies
       - Critical vulnerabilities: [count] (weight: -1.0 each)
       - High vulnerabilities: [count] (weight: -0.5 each)
       - Medium vulnerabilities: [count] (weight: -0.2 each)
       - Calculate: max(0, 1 - vulnerability_penalty) where penalty = Σ(severity_weight × count)
       - **Baseline estimate**: ~0.60 (assume 1-2 medium vulnerabilities typical)
     - V_freshness: Up-to-date dependencies
       - Up-to-date dependencies: [count] / [total]
       - Outdated dependencies: [count] (versions > 6 months old)
       - Severely outdated: [count] (versions > 1 year old)
       - Calculate: (up_to_date + 0.5×outdated) / total
       - **Baseline estimate**: ~0.50 (typical staleness in established projects)
     - V_stability: Tested, compatible versions
       - Test suite pass rate: 100% (baseline)
       - Breaking changes introduced: 0 (baseline)
       - Rollback incidents: 0 (no updates yet)
       - Calculate: test_pass_rate × (1 - breaking_change_rate)
       - **Baseline estimate**: 1.00 (no changes yet, tests passing)
     - V_license: License compliance
       - Compatible licenses: [count] / [total]
       - Unknown licenses: [count]
       - Conflicting licenses: [count]
       - Calculate: compatible / total
       - **Baseline estimate**: ~0.85 (assume most dependencies permissive)
     - **V_instance(s₀) = 0.4×0.60 + 0.3×0.50 + 0.2×1.00 + 0.1×0.85 = 0.625**
   - Calculate V_meta(s₀):
     - V_completeness: 0.00 (no methodology yet)
     - V_effectiveness: 0.00 (nothing to test)
     - V_reusability: 0.00 (nothing to transfer)
     - **V_meta(s₀) = 0.00**

7. **Gap Identification** (M₀.reflect):
   - **READ** meta-agents/reflect.md (gap analysis process)
   - What dependency management capabilities are missing?
     - No systematic vulnerability scanning
     - No automated update workflow
     - No license compliance tracking
     - No dependency bloat detection
     - No update testing strategy
     - No rollback procedures
   - What dependency aspects need coverage?
     - Security (vulnerability scanning, CVE tracking, patch urgency)
     - Freshness (version tracking, update recommendations, deprecation notices)
     - Stability (compatibility testing, breaking change detection, rollback)
     - License (compliance checking, conflict detection, policy enforcement)
     - Bloat (unnecessary dependency detection, tree analysis)
     - Automation (CI/CD integration, auto-update PRs, alerting)
   - What methodology components are needed?
     - Vulnerability assessment framework
     - Update strategy (when to update, testing requirements)
     - License compliance policy
     - Bloat detection criteria
     - Update workflow (manual vs automated)
     - Rollback procedures

8. **Initial Agent Applicability Assessment** (M₀.plan):
   - **READ** meta-agents/plan.md (agent selection strategies)
   - Which inherited agents are directly applicable to dependency management?
     - ⭐⭐⭐ data-analyst: Analyze dependency metrics, vulnerabilities, freshness
     - ⭐⭐ error-classifier: Classify dependency issues (vulnerabilities, conflicts, staleness)
     - ⭐⭐ root-cause-analyzer: Analyze why dependencies became stale or vulnerable
     - ⭐ recovery-advisor: Recommend dependency update strategies
     - ⭐ coder: Write automation scripts (update workflows, scanners)
     - doc-writer: Document dependency management methodology
   - Which inherited agents need adaptation for dependency management?
     - error-classifier: Adapt to dependency issue types (CVEs, license conflicts, staleness)
     - root-cause-analyzer: Adapt to dependency root causes (abandoned projects, breaking changes)
     - recovery-advisor: Adapt to dependency recovery (safe update paths, testing strategies)
     - Read agent files to understand capabilities, adapt prompts contextually
   - What new specialized agents might be needed?
     - dependency-analyzer: Parse go.mod, build dependency graph, analyze tree (likely needed)
     - vulnerability-scanner: Query CVE databases, assess risks, prioritize fixes (likely needed)
     - update-advisor: Recommend safe upgrade paths, test dependencies (likely needed)
     - license-checker: Ensure license compatibility, SPDX analysis (may be needed)
     - bloat-detector: Identify unnecessary dependencies (may be needed)
     - compatibility-tester: Test dependency updates against test suite (may be needed)
   - **NOTE**: Don't create new agents yet - just identify potential needs

9. **Documentation** (M₀.execute + doc-writer):
   - **READ** meta-agents/execute.md (coordination strategies)
   - **READ** agents/doc-writer.md
   - Invoke doc-writer to create iteration-0.md:
     - M₀ state: 5 capabilities inherited from Bootstrap-003 (observe, plan, execute, reflect, evolve)
     - A₀ state: 6 agents inherited from Bootstrap-003 (3 generic + 3 specialized)
     - Agent applicability assessment (which agents useful for dependency management)
     - Dependency landscape analysis (go.mod, ~20 direct dependencies, dependency tree)
     - Current dependency health assessment (ad-hoc updates, no scanning, minimal automation)
     - Security scan results (vulnerabilities found, severity distribution)
     - Freshness assessment (up-to-date vs outdated vs abandoned)
     - License compliance check (license inventory, compatibility analysis)
     - Calculated V_instance(s₀) = 0.625 and V_meta(s₀) = 0.00
     - Gap analysis (missing scanning, automation, methodology)
     - Reflection on next steps and agent reuse strategy
   - Save data artifacts:
     - data/s0-dependency-inventory.yaml (all dependencies, versions, licenses)
     - data/s0-dependency-graph.json (dependency tree from go list)
     - data/s0-vulnerabilities.yaml (security scan results)
     - data/s0-freshness-report.yaml (version comparison, staleness analysis)
     - data/s0-license-inventory.yaml (license catalog, compatibility matrix)
     - data/s0-metrics.json (calculated V_instance and V_meta values)
     - data/s0-dependency-gaps.yaml (identified gaps in dependency management)
     - data/s0-agent-applicability.yaml (inherited agents and their dependency management applicability)
   - Initialize knowledge structure:
     - knowledge/INDEX.md (empty knowledge catalog, will be populated in subsequent iterations)
     - knowledge/patterns/ (directory for domain-specific patterns)
     - knowledge/principles/ (directory for universal principles)
     - knowledge/templates/ (directory for reusable templates)
     - knowledge/best-practices/ (directory for context-specific practices)

10. **Reflection** (M₀.reflect + M₀.evolve):
    - **READ** meta-agents/reflect.md (reflection process)
    - **READ** meta-agents/evolve.md (methodology extraction readiness)
    - Is data collection complete? What additional data might be needed?
    - Are M₀ capabilities sufficient for baseline? (Yes, core capabilities adequate)
    - What should be the focus of Iteration 1?
      - Likely: Vulnerability remediation (address critical/high CVEs)
      - Or: Dependency graph analysis and bloat detection
      - Or: Setup automated vulnerability scanning
      - Decision based on OCA framework: Start with Observe phase (manual analysis)
    - Which inherited agents will be most useful in Iteration 1?
      - Likely: Need new dependency-analyzer agent (parse go.mod, build graph)
      - Likely: Need new vulnerability-scanner agent (CVE scanning, risk assessment)
      - Reuse: data-analyst (metrics), error-classifier (issue categorization)
    - Methodology extraction readiness:
      - Note patterns observed in dependency analysis
      - Identify update decision points that could become methodology
      - Prepare for pattern documentation in subsequent iterations

## Constraints
- Do NOT pre-decide what agents to create next
- Do NOT assume the dependency management process or evolution path
- Let the dependency data and gaps guide next steps
- Be honest about current dependency health (likely some vulnerabilities, some staleness)
- Calculate V(s₀) based on actual observations, not target values
- Remember two layers: concrete dependency management (instance) + methodology (meta)
- Expected baseline: V_instance(s₀) ≈ 0.60-0.70 (low baseline acceptable and expected)

## Output Format
Create iteration-0.md following this structure:
- Iteration metadata (number, date, duration)
- M₀ state documentation (5 capabilities inherited from Bootstrap-003)
- A₀ state documentation (6 agents inherited: 3 generic + 3 specialized)
- Agent applicability assessment (which agents useful for dependency management domain)
- Dependency landscape analysis (go.mod, dependency count, graph structure)
- Current dependency health assessment (vulnerabilities, freshness, licenses)
- Security scan results (CVE findings, severity distribution)
- Freshness assessment (version comparison, staleness metrics)
- License compliance check (license inventory, compatibility)
- Value calculation (V_instance(s₀) = 0.625, V_meta(s₀) = 0.00)
- Gap identification (missing scanning, automation, testing, methodology)
- Reflection on next steps and agent reuse strategy
- Data artifacts saved to data/ directory
```

---

## Iteration 1+: Subsequent Iterations (General Template)

```markdown
# Execute Iteration N: [To be determined by Meta-Agent]

## Context from Previous Iteration

Review the previous iteration file: experiments/bootstrap-010-dependency-health/iteration-[N-1].md

Extract:
- Current Meta-Agent state: M_{N-1}
- Current Agent Set: A_{N-1}
- Current dependency health state: V_instance(s_{N-1})
- Current methodology state: V_meta(s_{N-1})
- Problems identified
- Reflection notes on what's needed next

## Two-Layer Execution Protocol

**Layer 1 (Instance)**: Agents perform concrete dependency management
**Layer 2 (Meta)**: Meta-Agent observes and extracts methodology

Throughout iteration:
- Agents focus on concrete tasks (scan vulnerabilities, update dependencies, test compatibility)
- Meta-Agent observes dependency work and identifies patterns for methodology

## Meta-Agent Decision Process

**BEFORE STARTING**: Read relevant Meta-Agent capability files:
- **READ** meta-agents/observe.md (for observation strategies)
- **READ** meta-agents/plan.md (for planning and decisions)
- **READ** meta-agents/execute.md (for coordination)
- **READ** meta-agents/reflect.md (for evaluation)
- **READ** meta-agents/evolve.md (for evolution assessment)

As M_{N-1}, follow the five-capability process:

### 1. OBSERVE (M.observe)
- **READ** meta-agents/observe.md for dependency observation strategies
- Review previous iteration outputs (iteration-[N-1].md)
- Examine dependency health state:
  - What vulnerabilities were addressed? (if any)
  - What dependencies were updated? (if any)
  - What patterns were observed? (if any)
  - What automation was implemented? (if any)
- Identify gaps:
  - What vulnerabilities remain unaddressed?
  - What dependencies are still outdated?
  - What license issues persist?
  - What methodology patterns are emerging?
- **Methodology observation**:
  - What patterns emerged in previous dependency work?
  - What update decisions were made and why?
  - What principles can be extracted?

### 2. PLAN (M.plan)
- **READ** meta-agents/plan.md for prioritization and agent selection
- Based on observations, what is the primary goal for this iteration?
  - Examples:
    - "Address critical/high severity vulnerabilities"
    - "Update stale dependencies (>1 year old)"
    - "Build dependency graph analyzer"
    - "Implement automated vulnerability scanning"
    - "Test dependency updates against test suite"
    - "Develop license compliance policy"
    - "Transfer test to npm/pip project"
- What capabilities are needed to achieve this goal?
- **Agent Assessment**:
  - Are current agents (A_{N-1}) sufficient for this goal?
  - Can inherited agents handle dependency management? (error-classifier, data-analyst)
  - Or do we need specialized `dependency-analyzer` for go.mod parsing?
  - Do we need `vulnerability-scanner` for CVE scanning?
  - Do we need `update-advisor` for safe upgrade paths?
  - Do we need `license-checker` for SPDX analysis?
  - Do we need `compatibility-tester` for update testing?
- **Methodology Planning**:
  - What patterns should be documented this iteration?
  - What update decisions will inform methodology?

### 3. EXECUTE (M.execute)
- **READ** meta-agents/execute.md for coordination and pattern observation
- Decision point: Should I create a new specialized agent?

**IF current agents are insufficient:**
- **EVOLVE** (M.evolve): Create new specialized agent
  - **READ** meta-agents/evolve.md for agent creation criteria
  - Examples of specialized agents:
    - `dependency-analyzer`: Parse go.mod, build dependency graph
      - Capabilities: Parse go.mod/go.sum, extract dependencies, build tree, analyze relationships
      - Why needed: Generic agents insufficient for Go module analysis
    - `vulnerability-scanner`: Query CVE databases, assess risks
      - Capabilities: Run govulncheck, query GitHub Advisory Database, OSV, deps.dev API
      - Why needed: Specialized security knowledge and API integration required
    - `update-advisor`: Recommend safe upgrade paths
      - Capabilities: Analyze version compatibility, test updates, recommend strategies
      - Why needed: Update decision framework requires domain expertise
    - `license-checker`: Ensure license compatibility
      - Capabilities: Extract SPDX identifiers, check compatibility matrix, detect conflicts
      - Why needed: Legal compliance expertise required
    - `bloat-detector`: Identify unnecessary dependencies
      - Capabilities: Analyze import usage, detect unused dependencies, recommend removal
      - Why needed: Dependency optimization requires specialized analysis
    - `compatibility-tester`: Test dependency updates
      - Capabilities: Run test suite, detect breaking changes, validate compatibility
      - Why needed: Safe update validation requires systematic testing
  - Define agent name and specialization domain
  - Document capabilities the new agent provides
  - Explain why inherited agents are insufficient
  - **CREATE AGENT PROMPT FILE**: Write agents/{agent-name}.md
    - Include: agent role, dependency management-specific capabilities, input/output format
    - Include: specific instructions for this iteration's task
    - Include: Go modules knowledge (go.mod syntax, semantic versioning, dependency resolution)
  - Add to agent set: A_N = A_{N-1} ∪ {new_agent}

**Agent Invocation** (specialized or inherited):
- **READ agent prompt file** before invocation: agents/{agent-name}.md
- Invoke agent to execute concrete dependency work:
  - Scan vulnerabilities (govulncheck, CVE databases)
  - Analyze dependency graph (direct vs transitive)
  - Assess freshness (version comparison, staleness)
  - Check license compliance (SPDX, compatibility)
  - Detect bloat (unused dependencies)
  - Test updates (compatibility, breaking changes)
  - Generate dependency reports
- Produce iteration outputs (scan reports, update recommendations, test results)

**Methodology Extraction** (M.evolve):
- **OBSERVE agent work patterns**:
  - How did agent organize dependency analysis?
  - What vulnerability severity criteria were used?
  - What update decision criteria were applied (when to update, when to wait)?
  - What testing strategies were employed?
- **EXTRACT patterns for methodology**:
  - Document vulnerability assessment frameworks
  - Build dependency issue taxonomy (vulnerabilities, staleness, license, bloat)
  - Identify reusable update strategies
  - Note principles that emerge
  - Add to methodology documentation

**ELSE use inherited agents:**
- **READ agent prompt file** from agents/{agent-name}.md
- Invoke appropriate agents from A_{N-1}
- Execute planned dependency work
- Observe for methodology patterns

**CRITICAL EXECUTION PROTOCOL**:
1. ALWAYS read capability files before embodying Meta-Agent capabilities
2. ALWAYS read agent prompt file before each agent invocation
3. Do NOT cache instructions across iterations - always read from files
4. Capability files may be updated between iterations - get latest from files
5. Never assume capabilities - always verify from source files

### 4. REFLECT (M.reflect)
- **READ** meta-agents/reflect.md for evaluation process
- **Evaluate Instance Layer** (Concrete Dependency Management):
  - What vulnerabilities were addressed this iteration?
  - What dependencies were updated?
  - What tests were conducted?
  - Calculate new V_instance(s_N):
    - V_security: Vulnerability-free dependencies
      - Count critical/high/medium vulnerabilities
      - Calculate penalty: Σ(severity_weight × count)
      - Score: max(0, 1 - penalty) where critical=-1.0, high=-0.5, medium=-0.2
      - Target: 1.0 (zero vulnerabilities)
    - V_freshness: Up-to-date dependencies
      - Count up-to-date vs outdated vs severely outdated
      - Score: (up_to_date + 0.5×outdated) / total
      - Target: 0.85+ (>85% dependencies up-to-date)
    - V_stability: Tested, compatible versions
      - Test suite pass rate after updates
      - Breaking changes introduced
      - Score: test_pass_rate × (1 - breaking_change_rate)
      - Target: 1.0 (100% tests pass, no breaking changes)
    - V_license: License compliance
      - Compatible licenses / total dependencies
      - Unknown or conflicting licenses count
      - Score: compatible / total
      - Target: 0.95+ (>95% compatible licenses)
    - **V_instance(s_N) = 0.4×V_security + 0.3×V_freshness + 0.2×V_stability + 0.1×V_license**
  - Calculate change: ΔV_instance = V_instance(s_N) - V_instance(s_{N-1})
  - Are dependency objectives met? What gaps remain?

- **Evaluate Meta Layer** (Methodology):
  - What patterns were extracted this iteration?
  - Calculate new V_meta(s_N):
    - V_completeness: documented_patterns / total_patterns
      - Required: Vulnerability assessment, update strategy, license policy, bloat detection, automation workflow, testing procedures
      - Count documented vs required (6 minimum)
      - Score: documented / 6
    - V_effectiveness: 1 - (time_with_methodology / time_baseline)
      - Measure dependency management time if methodology used
      - Compare to baseline (~4-8 hours for 20 dependencies manual audit)
      - Later iterations: Transfer test to npm/pip project
      - Score: efficiency gain ratio
    - V_reusability: successful_transfers / transfer_attempts
      - Test methodology on different ecosystem (npm, pip, cargo)
      - Assess if methodology applies with <20% modification
      - Score: transfer success rate
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
    details: "If no new dependency specialization needed"

  instance_value_threshold:
    question: "Is V_instance(s_N) ≥ 0.80 (dependency health)?"
    V_instance(s_N): [calculated value]
    threshold_met: [Yes/No]
    components:
      V_security: [value] "≥0.90 target (minimal vulnerabilities)"
      V_freshness: [value] "≥0.85 target (>85% up-to-date)"
      V_stability: [value] "≥1.00 target (100% tests pass)"
      V_license: [value] "≥0.95 target (>95% compatible)"

  meta_value_threshold:
    question: "Is V_meta(s_N) ≥ 0.80 (methodology quality)?"
    V_meta(s_N): [calculated value]
    threshold_met: [Yes/No]
    components:
      V_completeness: [value] "≥0.85 target (5/6 patterns documented)"
      V_effectiveness: [value] "≥0.75 target (4x speedup)"
      V_reusability: [value] "≥0.80 target (80% transfer success)"

  instance_objectives:
    all_vulnerabilities_addressed: [Yes/No] "Critical/high CVEs fixed"
    dependencies_updated: [Yes/No] ">85% dependencies up-to-date"
    license_compliance_achieved: [Yes/No] ">95% compatible licenses"
    automation_implemented: [Yes/No] "CI/CD scanning integrated"
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
  - O = {dependency reports + methodology documentation}
  - A_N = {final agent set with specializations}
  - M_N = {final meta-agent capabilities}

**IF NOT CONVERGED:**
- Identify what's needed for next iteration
- Note: Focus on instance (dependency health) OR meta (methodology) as needed
- Continue to Iteration N+1

## Documentation Requirements

Create experiments/bootstrap-010-dependency-health/iteration-N.md with:

### 1. Iteration Metadata
```yaml
iteration: N
date: YYYY-MM-DD
duration: ~X hours
status: [completed/converged]
layers:
  instance: "Dependency management work performed this iteration"
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
        specialization: dependency_management_domain_area
        capabilities: [list]
        creation_reason: "Why were inherited 6 agents insufficient?"
        justification: "What gap exists that no inherited agent can fill?"
        prompt_file: "agents/{agent-name}.md"

  IF unchanged:
    status: "A_N = A_{N-1} (inherited agents sufficient for this iteration)"

  IF reused_inherited:
    reused_agents:
      - agent_name: task_performed
        adaptation_notes: "How inherited agent was adapted to dependency management context"

agents_invoked_this_iteration:
  - agent_name: task_performed
    source: [inherited / newly_created]
```

### 4. Instance Work Executed (Concrete Dependency Management)
- What vulnerabilities were addressed?
  - Critical vulnerabilities fixed (if any)
  - High vulnerabilities fixed (if any)
  - Medium vulnerabilities fixed (if any)
- What dependencies were updated?
  - Updated dependencies list (name, old version, new version)
  - Update strategy (direct update, major version jump, etc.)
  - Breaking changes handled
- What tests were conducted?
  - Test suite pass rate before/after updates
  - Compatibility tests run
  - Rollback tests (if applicable)
- What automation was implemented?
  - CI/CD integration (if any)
  - Automated scanning (if any)
  - Auto-update workflows (if any)
- What outputs were produced?
  - Vulnerability reports
  - Dependency update reports
  - License compliance reports
  - Automation scripts
- Summary of concrete deliverables

### 5. Meta Work Executed (Methodology Extraction)
- What patterns were observed in dependency work?
  - Vulnerability assessment patterns
  - Update decision patterns
  - Testing strategies
  - Rollback procedures
- What methodology content was documented?
  - Vulnerability assessment frameworks
  - Update decision criteria
  - License compliance policies
  - Automation strategies

**Knowledge Artifacts Created**:

Organize extracted knowledge into appropriate categories:

- **Patterns** (domain-specific): knowledge/patterns/{pattern-name}.md
  - Specific solutions to recurring problems in dependency management
  - Example: "Vulnerability Severity Triage Pattern", "Safe Update Path Pattern"
  - Format: Problem, Context, Solution, Consequences, Examples

- **Principles** (universal): knowledge/principles/{principle-name}.md
  - Fundamental truths or rules discovered
  - Example: "Security First Principle", "Test Before Update Principle"
  - Format: Statement, Rationale, Evidence, Applications

- **Templates** (reusable): knowledge/templates/{template-name}.{md|yaml|json}
  - Concrete implementations ready for reuse
  - Example: "Vulnerability Report Template", "Update Checklist Template"
  - Format: Template file + usage documentation

- **Best Practices** (context-specific): knowledge/best-practices/{topic}.md
  - Recommended approaches for specific contexts
  - Example: "Go Dependency Update Best Practices", "License Compliance Best Practices"
  - Format: Context, Recommendation, Justification, Trade-offs

- **Methodology** (project-wide): docs/methodology/{methodology-name}.md
  - Comprehensive guides for reuse across projects
  - Example: "Dependency Health Management Methodology for Go Projects"
  - Format: Complete methodology with decision frameworks

**Knowledge Index Update**:
- Update knowledge/INDEX.md with:
  - New knowledge entries
  - Links to iteration where extracted
  - Domain tags (dependency-management, security, licensing, go, npm, pip)
  - Validation status (proposed, validated, refined)

### 6. State Transition

**Instance Layer** (Dependency Health State):
```yaml
s_{N-1} → s_N (Dependency Health):
  changes:
    - vulnerabilities_fixed: [count]
    - dependencies_updated: [count]
    - tests_conducted: [count]

  metrics:
    V_security: [value] (was: [previous])
    V_freshness: [value] (was: [previous])
    V_stability: [value] (was: [previous])
    V_license: [value] (was: [previous])

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
  - Instance learnings (dependency management insights)
  - Meta learnings (methodology insights)
- What worked well?
- What challenges were encountered?
- What is needed next?
  - For dependency health completion
  - For methodology completion

### 8. Convergence Check
[Use the convergence criteria structure above]

### 9. Data Artifacts

**Ephemeral Data** (iteration-specific, saved to data/):
- data/iteration-N-metrics.json (V_instance, V_meta calculations)
- data/iteration-N-dependency-state.yaml (vulnerabilities, updates, tests)
- data/iteration-N-methodology.yaml (extracted patterns)
- data/iteration-N-artifacts/ (scan reports, update reports, test results)

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
   - Instance: Perform concrete dependency management
   - Meta: Extract methodology patterns from dependency work

2. **Be Honest**: Calculate V(s_N) based on actual state
   - Don't inflate instance values (dependency health)
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
   - Instance convergence: Dependency health + completeness
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

## Dependency Management Domain-Specific Guidance

Based on OCA framework, dependency management iterations may follow:

### Observe Phase (Iterations 0-2)
- Iteration 0: Baseline establishment, dependency inventory, security scan
- Iteration 1: Vulnerability analysis, dependency graph analysis, license audit
- Iteration 2: Freshness assessment, bloat detection, pattern identification
- **Methodology**: Observe dependency management patterns, identify issue types

### Codify Phase (Iterations 3-4)
- Iteration 3: Build vulnerability assessment framework, update decision criteria
- Iteration 4: Create dependency issue taxonomy, document update strategies
- **Methodology**: Extract update decision frameworks, codify patterns

### Automate Phase (Iterations 5-6)
- Iteration 5: Implement automated scanning (CI/CD integration), auto-update workflows
- Iteration 6: Transfer test to npm/pip project, validate methodology, convergence
- **Methodology**: Document automation strategies, validate transferability

**Caveat**: Let actual needs drive the sequence, not this expected pattern. The pattern is a hint, not a prescription.

## Domain-Specific Tools and Data Sources

### Vulnerability Scanning
- **govulncheck**: Official Go vulnerability checker
  - Command: `govulncheck ./...`
  - Output: Vulnerability report with severity, CVE ID, affected packages
- **GitHub Advisory Database**: CVE database for GitHub packages
  - API: `https://api.github.com/advisories`
  - Query by ecosystem (Go), package name
- **OSV (Open Source Vulnerabilities)**: Cross-ecosystem vulnerability database
  - API: `https://api.osv.dev/v1/query`
  - Query by package name, version
- **deps.dev API**: Dependency insights (Google)
  - API: `https://api.deps.dev/v3alpha/systems/go/packages/{package}`
  - Provides: versions, advisories, licenses

### Dependency Analysis
- **go list**: Go module dependency listing
  - Command: `go list -m -json all` (all dependencies as JSON)
  - Command: `go list -m -u all` (check for updates)
  - Output: Dependency graph with versions
- **go mod graph**: Dependency graph visualization
  - Command: `go mod graph`
  - Output: Dependency relationships (edges)
- **go mod why**: Explain dependency requirement
  - Command: `go mod why <package>`
  - Output: Dependency chain explaining requirement

### License Compliance
- **go-licenses**: Extract licenses from Go dependencies
  - Tool: `github.com/google/go-licenses`
  - Command: `go-licenses csv <module>`
  - Output: CSV with package, version, license
- **SPDX**: Software Package Data Exchange (license identifiers)
  - Standard: SPDX license identifiers (MIT, Apache-2.0, GPL-3.0, etc.)
  - Compatibility matrix: Check license compatibility

### Freshness and Maintenance
- **GitHub API**: Repository metadata
  - Last commit date, release date, issue count, contributor activity
  - API: `https://api.github.com/repos/{owner}/{repo}`
- **Go Package Index**: Package metadata
  - pkg.go.dev provides version history, import by count
  - API: `https://proxy.golang.org/{module}/@v/list`

### Update Testing
- **go test**: Run test suite
  - Command: `go test ./...` (all packages)
  - Verify: Test pass rate before/after update
- **go build**: Build verification
  - Command: `go build ./...`
  - Verify: Build succeeds after update

## Expected Agents for Dependency Management

Based on domain analysis, expect these specialized agents to emerge:

1. **dependency-analyzer** (likely Iteration 1)
   - Parse go.mod, go.sum
   - Build dependency graph (direct vs transitive)
   - Analyze dependency relationships

2. **vulnerability-scanner** (likely Iteration 1-2)
   - Run govulncheck
   - Query GitHub Advisory Database, OSV
   - Assess vulnerability severity and exploitability
   - Prioritize remediation

3. **update-advisor** (likely Iteration 2-3)
   - Analyze version compatibility
   - Recommend safe upgrade paths
   - Detect breaking changes (major version bumps)
   - Plan update strategy (incremental vs batch)

4. **license-checker** (likely Iteration 2-3)
   - Extract SPDX identifiers from dependencies
   - Check license compatibility matrix
   - Detect conflicting licenses
   - Generate compliance report

5. **bloat-detector** (likely Iteration 3-4)
   - Analyze import statements
   - Detect unused dependencies
   - Recommend dependency removal
   - Calculate dependency weight

6. **compatibility-tester** (likely Iteration 3-4)
   - Run test suite before/after updates
   - Detect breaking changes
   - Validate rollback procedures
   - Generate compatibility report

**Note**: These are expected based on domain analysis, but let actual needs drive agent creation.
```

---

## Quick Reference: Iteration Checklist

For each iteration N ≥ 1, ensure you:

**Preparation**:
- [ ] Review previous iteration (iteration-[N-1].md)
- [ ] Extract current state (M_{N-1}, A_{N-1}, V_instance(s_{N-1}), V_meta(s_{N-1}))

**Observe Phase**:
- [ ] **READ** meta-agents/observe.md (dependency observation strategies)
- [ ] Analyze dependency health state (vulnerabilities, freshness, licenses)
- [ ] Identify gaps (unaddressed vulnerabilities, outdated dependencies, methodology gaps)
- [ ] Observe patterns (for methodology extraction)

**Plan Phase**:
- [ ] **READ** meta-agents/plan.md (prioritization, agent selection)
- [ ] Define iteration goal (instance layer: vulnerability fixes, updates, automation)
- [ ] Assess agent sufficiency (inherited sufficient OR need specialized?)
- [ ] Plan methodology extraction (what patterns to document?)

**Execute Phase**:
- [ ] **READ** meta-agents/execute.md (coordination, pattern observation)
- [ ] **IF NEW AGENT NEEDED**:
  - [ ] **READ** meta-agents/evolve.md (agent creation criteria)
  - [ ] Create agent prompt file: agents/{agent-name}.md
  - [ ] Document specialization reason
- [ ] **READ** agent prompt file(s) before invocation: agents/{agent-name}.md
- [ ] Invoke agents to perform dependency management
- [ ] Observe dependency work for methodology patterns
- [ ] Extract patterns and update methodology documentation

**Reflect Phase**:
- [ ] **READ** meta-agents/reflect.md (evaluation process)
- [ ] Calculate V_instance(s_N) (dependency health):
  - [ ] V_security (vulnerability-free dependencies)
  - [ ] V_freshness (up-to-date dependencies)
  - [ ] V_stability (tested, compatible versions)
  - [ ] V_license (license compliance)
- [ ] Calculate V_meta(s_N) (methodology quality):
  - [ ] V_completeness (documented_patterns / total_patterns)
  - [ ] V_effectiveness (1 - time_with_methodology / time_baseline)
  - [ ] V_reusability (successful_transfers / transfer_attempts)
- [ ] Assess quality honestly (don't inflate values)
- [ ] Identify gaps (dependency health + methodology)

**Convergence Check**:
- [ ] M_N == M_{N-1}? (meta-agent stable)
- [ ] A_N == A_{N-1}? (agent set stable)
- [ ] V_instance(s_N) ≥ 0.80? (dependency health threshold)
- [ ] V_meta(s_N) ≥ 0.80? (methodology quality threshold)
- [ ] Instance objectives complete? (vulnerabilities fixed, dependencies updated, automation implemented)
- [ ] Meta objectives complete? (methodology documented, transfer test successful)
- [ ] Determine: CONVERGED or NOT_CONVERGED

**Documentation**:
- [ ] Create iteration-N.md with:
  - [ ] Metadata (iteration, date, duration, status)
  - [ ] M evolution (evolved or unchanged)
  - [ ] A evolution (new agents or unchanged)
  - [ ] Instance work (vulnerabilities fixed, dependencies updated, tests conducted)
  - [ ] Meta work (patterns extracted)
  - [ ] Knowledge artifacts created (list all new knowledge)
  - [ ] State transition (V_instance, V_meta calculated)
  - [ ] Reflection (learnings, next steps)
  - [ ] Convergence check (criteria evaluated)
- [ ] Save data artifacts to data/:
  - [ ] iteration-N-metrics.json
  - [ ] iteration-N-dependency-state.yaml
  - [ ] iteration-N-methodology.yaml
  - [ ] iteration-N-artifacts/ (scan reports, update reports, test results)
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
- [ ] Perform reusability validation (transfer test to npm/pip/cargo)
- [ ] Document three-tuple: (O={dependency reports+methodology}, A_N, M_N)
- [ ] Validate methodology alignment (OCA, Bootstrapped SE, Value Space)

---

## Notes on Execution Style

**Be the Meta-Agent**: When executing iterations, embody M's perspective:
- Think through the observe-plan-execute-reflect-evolve cycle
- Read capability files to understand M's strategies
- Make explicit decisions about agent creation
- Justify why specialization is needed
- Extract methodology patterns from dependency work
- Track both V_instance and V_meta

**Be Domain-Aware**: Dependency management has specific concerns:
- Security vulnerabilities (CVEs, exploitability, severity)
- Dependency freshness (up-to-date vs outdated vs abandoned)
- License compliance (SPDX, compatibility, conflicts)
- Dependency bloat (unnecessary dependencies, tree optimization)
- Update safety (breaking changes, compatibility, rollback)
- Ecosystem differences (Go modules, npm, pip, cargo)

**Be Rigorous**: Calculate values honestly
- V_instance based on actual dependency health (not aspirational)
- V_meta based on actual methodology coverage (not desired)
- Don't force convergence prematurely
- Don't skip methodology extraction to focus only on dependency work
- Let data and needs drive the process

**Be Thorough**: Document decisions and reasoning
- Save intermediate data (scan reports, update reports, test results)
- Show your work (calculations, analysis)
- Make evolution path traceable
- **NO TOKEN LIMITS**: Complete all steps fully, never abbreviate

**Be Two-Layer Aware**: Always work on both layers
- Instance layer: Perform concrete dependency management
- Meta layer: Extract reusable methodology
- Don't neglect either layer
- Methodology extraction is as important as dependency work

**Be Authentic**: This is a real experiment
- Discover dependency management patterns, don't assume them
- Create agents based on need, not predetermined plan
- Extract methodology from actual dependency work, not theory
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
**Purpose**: Guide authentic execution of bootstrap-010-dependency-health experiment with dual-layer architecture

**Key Innovation**: Dual value functions (V_instance + V_meta) enable simultaneous optimization of concrete dependency health management and reusable methodology.
