# Iteration Execution Prompts

This document provides structured prompts for executing each iteration of the bootstrap-008-code-review experiment.

**Two-Layer Architecture**:
- **Instance Layer**: Agents perform comprehensive code review of internal/ package
- **Meta Layer**: Meta-Agent observes review patterns and extracts methodology

---

## Iteration 0: Baseline Establishment

```markdown
# Execute Iteration 0: Baseline Establishment

## Context
I'm starting the bootstrap-008-code-review experiment. I've reviewed:
- experiments/bootstrap-008-code-review/plan.md
- experiments/bootstrap-008-code-review/README.md
- experiments/bootstrap-008-code-review/BOOTSTRAP-007-INHERITANCE.md
- The three methodology frameworks (OCA, Bootstrapped SE, Value Space Optimization)

## Current State
- Meta-Agent: M₀ (6 capabilities inherited from Bootstrap-007: observe, plan, execute, reflect, evolve, api-design-orchestrator)
- Agent Set: A₀ (15 agents inherited from Bootstrap-007: 3 generic + 12 specialized)
- Target: internal/ package (~15,000 lines Go code across 6 modules)

## Inherited State from Bootstrap-007

**IMPORTANT**: This experiment starts with the converged state from Bootstrap-007, NOT from scratch.

**Meta-Agent capability files (ALREADY EXIST)**:
- experiments/bootstrap-008-code-review/meta-agents/observe.md (validated)
- experiments/bootstrap-008-code-review/meta-agents/plan.md (validated)
- experiments/bootstrap-008-code-review/meta-agents/execute.md (validated)
- experiments/bootstrap-008-code-review/meta-agents/reflect.md (validated)
- experiments/bootstrap-008-code-review/meta-agents/evolve.md (validated)
- experiments/bootstrap-008-code-review/meta-agents/api-design-orchestrator.md (adaptable)

**Agent prompt files (ALREADY EXIST - 15 agents)**:
- Generic (3): data-analyst.md, doc-writer.md, coder.md
- From Bootstrap-001 (2): doc-generator.md, search-optimizer.md
- From Bootstrap-003 (3): error-classifier.md, recovery-advisor.md, root-cause-analyzer.md
- From Bootstrap-006 (7): agent-audit-executor.md, agent-documentation-enhancer.md,
  agent-parameter-categorizer.md, agent-quality-gate-installer.md, agent-schema-refactorer.md,
  agent-validation-builder.md, api-evolution-planner.md

**CRITICAL EXECUTION PROTOCOL**:
- All capability files and agent files ALREADY EXIST (inherited from Bootstrap-007)
- Before embodying Meta-Agent capabilities, ALWAYS read the relevant capability file first
- Before invoking ANY agent, ALWAYS read its prompt file first
- These files contain validated capabilities/agents; adapt them to code review context
- Never assume capabilities - always read from the source files

## Iteration 0 Objectives

Execute baseline establishment:

0. **Setup** (Verify inherited state):
   - **VERIFY META-AGENT CAPABILITY FILES EXIST** (inherited from Bootstrap-007):
     - ✓ meta-agents/observe.md: Data collection, pattern discovery (validated)
     - ✓ meta-agents/plan.md: Prioritization, agent selection (validated)
     - ✓ meta-agents/execute.md: Agent coordination, task execution (validated)
     - ✓ meta-agents/reflect.md: Value calculation (V_instance, V_meta), gap analysis (validated)
     - ✓ meta-agents/evolve.md: Agent creation criteria, methodology extraction (validated)
     - ✓ meta-agents/api-design-orchestrator.md: Domain orchestration (adaptable to code review)
   - **VERIFY INITIAL AGENT PROMPT FILES EXIST** (inherited from Bootstrap-007):
     - ✓ agents/data-analyst.md, doc-writer.md, coder.md (generic agents)
     - ✓ agents/doc-generator.md, search-optimizer.md (from Bootstrap-001)
     - ✓ agents/error-classifier.md, recovery-advisor.md, root-cause-analyzer.md (from Bootstrap-003)
     - ✓ agents/agent-*.md (7 agents from Bootstrap-006)
   - **NO NEED TO CREATE NEW FILES** - all files inherited and ready to use
   - **ADAPTATION NOTE**: Capability files and agent files are generic enough to apply to code review;
     read them to understand their validated approaches, then apply to code review context

1. **Codebase Analysis** (M₀.observe):
   - **READ** meta-agents/observe.md (code review observation strategies)
   - Analyze internal/ package structure:
     - List all modules and files
     - Count lines of code per module
     - Identify test coverage per module
     - Examine code complexity (cyclomatic complexity if available)
     - Review historical edit patterns (most-edited files)
   - Target modules:
     - parser/ (~3,500 lines): Session history JSONL parsing
     - analyzer/ (~2,800 lines): Pattern detection algorithms
     - query/ (~3,200 lines): Query engine implementation
     - validation/ (~2,500 lines): API schema validation
     - tools/ (~1,800 lines): Tool registry and definitions
     - capabilities/ (~1,200 lines): Capability management
   - Review existing quality measures:
     - Test coverage reports
     - Linting results (go vet, golint if available)
     - Build/test history from session data

2. **Current Review Process Assessment** (M₀.observe + M₀.plan):
   - **READ** meta-agents/observe.md (observation strategies)
   - **READ** meta-agents/plan.md (assessment frameworks)
   - What review process currently exists?
     - Manual ad-hoc reviews (baseline)
     - Existing linting: gofmt, go vet (minimal)
     - Test coverage: 70-85% (target: 80%+)
     - Security scanning: None
     - Style enforcement: None beyond gofmt
   - What review artifacts exist?
     - Code comments (quality varies)
     - godoc documentation (coverage varies)
     - Test files (presence varies)
   - What gaps exist?
     - No systematic review checklist
     - No issue tracking for code quality
     - No automated style checking
     - No security scanning
     - No complexity monitoring

3. **Baseline Metrics Calculation** (M₀.plan + data-analyst):
   - **READ** meta-agents/plan.md (prioritization strategies)
   - **READ** agents/data-analyst.md
   - Invoke data-analyst to calculate V_instance(s₀):
     - V_issue_detection: Estimate current issue finding rate
       - Historical bug rate: 6.06% error rate from session data
       - Test coverage gaps: Modules with <80% coverage
       - Manual review baseline: ~30% (estimate based on ad-hoc reviews)
       - Calculate: issues_found / total_actual_issues
     - V_false_positive: Estimate current false positive rate
       - Manual review noise: ~30% (estimate based on vague recommendations)
       - Calculate: 1 - (false_positives / total_issues_reported) = 0.70
     - V_actionability: Assess current recommendation quality
       - Manual review actionability: ~50% (many vague suggestions)
       - Calculate: actionable_recommendations / total_recommendations = 0.50
     - V_learning: Assess current pattern documentation
       - Patterns documented: ~20% (minimal knowledge capture)
       - Calculate: patterns_documented / patterns_identified = 0.20
     - **V_instance(s₀) = 0.3×0.30 + 0.3×0.70 + 0.2×0.50 + 0.2×0.20 = 0.44**
   - Calculate V_meta(s₀):
     - V_completeness: 0.00 (no methodology yet)
     - V_effectiveness: 0.00 (nothing to test)
     - V_reusability: 0.00 (nothing to transfer)
     - **V_meta(s₀) = 0.00**

4. **Gap Identification** (M₀.reflect):
   - **READ** meta-agents/reflect.md (gap analysis process)
   - What code review capabilities are missing?
     - No systematic review process
     - No issue classification taxonomy
     - No automated checking (beyond basic linting)
     - No security scanning
     - No style guide enforcement
   - What review aspects need coverage?
     - Correctness (bugs, edge cases, error handling)
     - Maintainability (complexity, duplication, coupling)
     - Readability (naming, structure, comments)
     - Go best practices (idioms, patterns)
     - Security (input validation, vulnerabilities)
     - Performance (efficiency, memory, concurrency)
   - What methodology components are needed?
     - Review process framework
     - Issue classification taxonomy
     - Review decision criteria
     - Automation strategies
     - Tool selection guidelines
     - Prioritization frameworks

5. **Initial Agent Applicability Assessment** (M₀.plan):
   - **READ** meta-agents/plan.md (agent selection strategies)
   - Which inherited agents are directly applicable to code review?
     - ⭐⭐⭐ agent-quality-gate-installer: Linting rules, pre-commit hooks
     - ⭐⭐ agent-audit-executor: Code consistency audits
     - ⭐⭐ agent-documentation-enhancer: Comment quality improvement
     - ⭐ error-classifier: Classify code issues (bugs, smells, anti-patterns)
     - ⭐ recovery-advisor: Recommend fixes and refactorings
     - ⭐ root-cause-analyzer: Analyze issue root causes
     - ⭐ agent-validation-builder: Review validation logic
     - data-analyst: Analyze code metrics
     - coder: Write custom linters, automation scripts
   - Which inherited agents need adaptation for code review?
     - Most agents: Apply to code domain instead of API/CI/CD domain
     - Read agent files to understand capabilities, adapt prompts contextually
   - What new specialized agents might be needed?
     - code-reviewer: Systematic code review execution (likely needed)
     - security-scanner: Vulnerability detection (may be needed)
     - style-checker: Style guide enforcement (may be needed)
     - best-practice-advisor: Go idioms and patterns (may be needed)
   - **NOTE**: Don't create new agents yet - just identify potential needs

6. **Documentation** (M₀.execute + doc-writer):
   - **READ** meta-agents/execute.md (coordination strategies)
   - **READ** agents/doc-writer.md
   - Invoke doc-writer to create iteration-0.md:
     - M₀ state: 6 capabilities inherited from Bootstrap-007 (observe, plan, execute, reflect, evolve, api-design-orchestrator)
     - A₀ state: 15 agents inherited from Bootstrap-007 (3 generic + 12 specialized)
     - Agent applicability assessment (which agents useful for code review)
     - Codebase analysis summary (6 modules, ~15K lines)
     - Current review process assessment (manual ad-hoc, minimal automation)
     - Calculated V_instance(s₀) = 0.44 and V_meta(s₀) = 0.00
     - Gap analysis (missing review process, taxonomy, automation, methodology)
     - Reflection on next steps and agent reuse strategy
   - Save data artifacts:
     - data/s0-codebase-structure.yaml (module breakdown, line counts, test coverage)
     - data/s0-metrics.json (calculated V_instance and V_meta values)
     - data/s0-review-gaps.yaml (identified gaps in review process)
     - data/s0-agent-applicability.yaml (inherited agents and their code review applicability)
   - Initialize knowledge structure:
     - knowledge/INDEX.md (empty knowledge catalog, will be populated in subsequent iterations)
     - knowledge/patterns/ (directory for domain-specific patterns)
     - knowledge/principles/ (directory for universal principles)
     - knowledge/templates/ (directory for reusable templates)
     - knowledge/best-practices/ (directory for context-specific practices)

7. **Reflection** (M₀.reflect + M₀.evolve):
   - **READ** meta-agents/reflect.md (reflection process)
   - **READ** meta-agents/evolve.md (methodology extraction readiness)
   - Is data collection complete? What additional data might be needed?
   - Are M₀ capabilities sufficient for baseline? (Yes, core capabilities adequate)
   - What should be the focus of Iteration 1?
     - Likely: Manual code review of parser/ module (largest module, ~3,500 lines)
     - Or: Setup automated linting and static analysis first
     - Decision based on OCA framework: Start with Observe phase (manual review)
   - Which inherited agents will be most useful in Iteration 1?
     - Likely: Need new code-reviewer agent (systematic review execution)
     - Reuse: data-analyst (code metrics), doc-writer (report generation)
   - Methodology extraction readiness:
     - Note patterns observed in existing code
     - Identify review decision points that could become methodology
     - Prepare for pattern documentation in subsequent iterations

## Constraints
- Do NOT pre-decide what agents to create next
- Do NOT assume the review process or evolution path
- Let the codebase data and gaps guide next steps
- Be honest about current review state (minimal automation, manual ad-hoc)
- Calculate V(s₀) based on actual observations, not target values
- Remember two layers: concrete review (instance) + methodology (meta)

## Output Format
Create iteration-0.md following this structure:
- Iteration metadata (number, date, duration)
- M₀ state documentation (6 capabilities inherited from Bootstrap-007)
- A₀ state documentation (15 agents inherited: 3 generic + 12 specialized)
- Agent applicability assessment (which agents useful for code review domain)
- Codebase analysis (6 modules, line counts, test coverage, complexity)
- Current review process assessment (manual ad-hoc, minimal automation)
- Value calculation (V_instance(s₀) = 0.44, V_meta(s₀) = 0.00)
- Gap identification (missing review process, taxonomy, automation, methodology)
- Reflection on next steps and agent reuse strategy
- Data artifacts saved to data/ directory
```

---

## Iteration 1+: Subsequent Iterations (General Template)

```markdown
# Execute Iteration N: [To be determined by Meta-Agent]

## Context from Previous Iteration

Review the previous iteration file: experiments/bootstrap-008-code-review/iteration-[N-1].md

Extract:
- Current Meta-Agent state: M_{N-1}
- Current Agent Set: A_{N-1}
- Current review state: V_instance(s_{N-1})
- Current methodology state: V_meta(s_{N-1})
- Problems identified
- Reflection notes on what's needed next

## Two-Layer Execution Protocol

**Layer 1 (Instance)**: Agents perform concrete code review
**Layer 2 (Meta)**: Meta-Agent observes and extracts methodology

Throughout iteration:
- Agents focus on concrete tasks (review code, identify issues, recommend fixes)
- Meta-Agent observes review work and identifies patterns for methodology

## Meta-Agent Decision Process

**BEFORE STARTING**: Read relevant Meta-Agent capability files:
- **READ** meta-agents/observe.md (for observation strategies)
- **READ** meta-agents/plan.md (for planning and decisions)
- **READ** meta-agents/execute.md (for coordination)
- **READ** meta-agents/reflect.md (for evaluation)
- **READ** meta-agents/evolve.md (for evolution assessment)

As M_{N-1}, follow the five-capability process:

### 1. OBSERVE (M.observe)
- **READ** meta-agents/observe.md for code review observation strategies
- Review previous iteration outputs (iteration-[N-1].md)
- Examine code review state:
  - What modules have been reviewed? (if any)
  - What issues have been identified? (if any)
  - What patterns have been observed? (if any)
  - What review automation exists? (if any)
- Identify gaps:
  - What modules still need review?
  - What review aspects are missing? (correctness, maintainability, readability, etc.)
  - What automation is needed?
  - What methodology patterns are emerging?
- **Methodology observation**:
  - What patterns emerged in previous review work?
  - What review decisions were made and why?
  - What principles can be extracted?

### 2. PLAN (M.plan)
- **READ** meta-agents/plan.md for prioritization and agent selection
- Based on observations, what is the primary goal for this iteration?
  - Examples:
    - "Manual code review of parser/ module"
    - "Build issue classification taxonomy"
    - "Implement automated linting rules"
    - "Review query/ and validation/ modules"
    - "Transfer test to cmd/ package"
- What capabilities are needed to achieve this goal?
- **Agent Assessment**:
  - Are current agents (A_{N-1}) sufficient for this goal?
  - Can inherited agents handle code review? (error-classifier, audit-executor, etc.)
  - Or do we need specialized `code-reviewer` for systematic review?
  - Do we need `security-scanner` for vulnerability detection?
  - Do we need `style-checker` for style enforcement?
- **Methodology Planning**:
  - What patterns should be documented this iteration?
  - What review decisions will inform methodology?

### 3. EXECUTE (M.execute)
- **READ** meta-agents/execute.md for coordination and pattern observation
- Decision point: Should I create a new specialized agent?

**IF current agents are insufficient:**
- **EVOLVE** (M.evolve): Create new specialized agent
  - **READ** meta-agents/evolve.md for agent creation criteria
  - Examples of specialized agents:
    - `code-reviewer`: Systematic code review execution
      - Capabilities: Read code, identify issues, categorize by type, recommend fixes
      - Why needed: Generic agents insufficient for comprehensive review
    - `security-scanner`: Vulnerability detection
      - Capabilities: Run gosec, analyze results, identify security issues
      - Why needed: Specialized security knowledge required
    - `style-checker`: Style guide enforcement
      - Capabilities: Check naming, structure, patterns beyond gofmt
      - Why needed: Consistent style beyond basic formatting
    - `best-practice-advisor`: Go idioms and patterns
      - Capabilities: Identify non-idiomatic code, recommend Go best practices
      - Why needed: Language-specific expertise required
  - Define agent name and specialization domain
  - Document capabilities the new agent provides
  - Explain why inherited agents are insufficient
  - **CREATE AGENT PROMPT FILE**: Write agents/{agent-name}.md
    - Include: agent role, code review-specific capabilities, input/output format
    - Include: specific instructions for this iteration's task
    - Include: Go domain knowledge (idioms, best practices, anti-patterns)
  - Add to agent set: A_N = A_{N-1} ∪ {new_agent}

**Agent Invocation** (specialized or inherited):
- **READ agent prompt file** before invocation: agents/{agent-name}.md
- Invoke agent to execute concrete review work:
  - Review code module by module
  - Identify issues (bugs, smells, anti-patterns, security, performance)
  - Categorize issues by type and severity
  - Recommend fixes and improvements
  - Generate review reports
- Produce iteration outputs (review reports, issue catalogs, recommendations)

**Methodology Extraction** (M.evolve):
- **OBSERVE agent work patterns**:
  - How did agent organize review process?
  - What issue types were identified?
  - What decision criteria were used (when to flag, when to ignore)?
  - What prioritization logic was applied?
- **EXTRACT patterns for methodology**:
  - Document review decision frameworks
  - Build issue classification taxonomy
  - Identify reusable review patterns
  - Note principles that emerge
  - Add to methodology documentation

**ELSE use inherited agents:**
- **READ agent prompt file** from agents/{agent-name}.md
- Invoke appropriate agents from A_{N-1}
- Execute planned review work
- Observe for methodology patterns

**CRITICAL EXECUTION PROTOCOL**:
1. ALWAYS read capability files before embodying Meta-Agent capabilities
2. ALWAYS read agent prompt file before each agent invocation
3. Do NOT cache instructions across iterations - always read from files
4. Capability files may be updated between iterations - get latest from files
5. Never assume capabilities - always verify from source files

### 4. REFLECT (M.reflect)
- **READ** meta-agents/reflect.md for evaluation process
- **Evaluate Instance Layer** (Concrete Review):
  - What modules were reviewed this iteration?
  - What issues were identified?
  - Calculate new V_instance(s_N):
    - V_issue_detection: issues_found / total_actual_issues
      - Count issues from review reports
      - Estimate total from complexity + coverage gaps + historical bugs
    - V_false_positive: 1 - (false_positives / total_issues_reported)
      - Validate flagged issues
      - Count false positives (issues that aren't real problems)
    - V_actionability: actionable_recommendations / total_recommendations
      - Assess if recommendations are specific, implementable, justified
      - Count actionable vs total recommendations
    - V_learning: patterns_documented / patterns_identified
      - Count patterns in knowledge/ directory
      - Count patterns observed during review
    - **V_instance(s_N) = 0.3×V_issue_det + 0.3×V_false_pos + 0.2×V_action + 0.2×V_learn**
  - Calculate change: ΔV_instance = V_instance(s_N) - V_instance(s_{N-1})
  - Are review objectives met? What gaps remain?

- **Evaluate Meta Layer** (Methodology):
  - What patterns were extracted this iteration?
  - Calculate new V_meta(s_N):
    - V_completeness: documented_patterns / total_patterns
      - Required: Review process, issue taxonomy, decision criteria, automation strategies, tool guidelines, prioritization
      - Count documented vs required (7 minimum)
    - V_effectiveness: 1 - (review_time_with_methodology / review_time_baseline)
      - Measure review time if methodology used
      - Compare to baseline (~8-10 hours for 15K lines)
      - Later iterations: Transfer test to cmd/ package
    - V_reusability: successful_transfers / transfer_attempts
      - Test methodology on different codebase (cmd/ package)
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
    details: "If no new review specialization needed"

  instance_value_threshold:
    question: "Is V_instance(s_N) ≥ 0.80 (review quality)?"
    V_instance(s_N): [calculated value]
    threshold_met: [Yes/No]
    components:
      V_issue_detection: [value] "≥0.70 target"
      V_false_positive: [value] "≥0.80 target"
      V_actionability: [value] "≥0.80 target"
      V_learning: [value] "≥0.75 target"

  meta_value_threshold:
    question: "Is V_meta(s_N) ≥ 0.80 (methodology quality)?"
    V_meta(s_N): [calculated value]
    threshold_met: [Yes/No]
    components:
      V_completeness: [value] "≥0.90 target"
      V_effectiveness: [value] "≥0.80 target (5x speedup)"
      V_reusability: [value] "≥0.70 target (70% transfer)"

  instance_objectives:
    all_modules_reviewed: [Yes/No] "parser, analyzer, query, validation, tools, capabilities"
    issue_catalog_complete: [Yes/No]
    recommendations_actionable: [Yes/No]
    automation_implemented: [Yes/No]
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
  - O = {review reports + methodology documentation}
  - A_N = {final agent set with specializations}
  - M_N = {final meta-agent capabilities}

**IF NOT CONVERGED:**
- Identify what's needed for next iteration
- Note: Focus on instance (review) OR meta (methodology) as needed
- Continue to Iteration N+1

## Documentation Requirements

Create experiments/bootstrap-008-code-review/iteration-N.md with:

### 1. Iteration Metadata
```yaml
iteration: N
date: YYYY-MM-DD
duration: ~X hours
status: [completed/converged]
layers:
  instance: "Code review work performed this iteration"
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

  inherited_baseline: "15 agents from Bootstrap-007 (3 generic + 12 specialized)"

  IF evolved (new agent created):
    new_agents:
      - name: agent_name
        specialization: code_review_domain_area
        capabilities: [list]
        creation_reason: "Why were inherited 15 agents insufficient?"
        justification: "What gap exists that no inherited agent can fill?"
        prompt_file: "agents/{agent-name}.md"

  IF unchanged:
    status: "A_N = A_{N-1} (inherited agents sufficient for this iteration)"

  IF reused_inherited:
    reused_agents:
      - agent_name: task_performed
        adaptation_notes: "How inherited agent was adapted to code review context"

agents_invoked_this_iteration:
  - agent_name: task_performed
    source: [inherited / newly_created]
```

### 4. Instance Work Executed (Concrete Review)
- What modules were reviewed?
  - parser/ (if reviewed)
  - analyzer/ (if reviewed)
  - query/ (if reviewed)
  - validation/ (if reviewed)
  - tools/ (if reviewed)
  - capabilities/ (if reviewed)
- What issues were identified?
  - Bugs (correctness issues)
  - Code smells (maintainability issues)
  - Readability issues (naming, structure, comments)
  - Security vulnerabilities
  - Performance issues
  - Go best practice violations
- What outputs were produced?
  - Review reports (per-module)
  - Issue catalog (categorized, prioritized)
  - Improvement recommendations (actionable)
- Summary of concrete deliverables

### 5. Meta Work Executed (Methodology Extraction)
- What patterns were observed in review work?
  - Review decision patterns
  - Issue classification patterns
  - Prioritization logic
  - Fix recommendation patterns
- What methodology content was documented?
  - Review process frameworks
  - Issue taxonomies
  - Decision criteria
  - Automation strategies

**Knowledge Artifacts Created**:

Organize extracted knowledge into appropriate categories:

- **Patterns** (domain-specific): knowledge/patterns/{pattern-name}.md
  - Specific solutions to recurring problems in code review
  - Example: "Comment-First Review Pattern", "Module-By-Module Review Pattern"
  - Format: Problem, Context, Solution, Consequences, Examples

- **Principles** (universal): knowledge/principles/{principle-name}.md
  - Fundamental truths or rules discovered
  - Example: "False Positive Minimization Principle", "Actionability Principle"
  - Format: Statement, Rationale, Evidence, Applications

- **Templates** (reusable): knowledge/templates/{template-name}.{md|yaml|json}
  - Concrete implementations ready for reuse
  - Example: "Code Review Checklist Template", "Issue Report Template"
  - Format: Template file + usage documentation

- **Best Practices** (context-specific): knowledge/best-practices/{topic}.md
  - Recommended approaches for specific contexts
  - Example: "Go Error Handling Best Practices", "Go Naming Best Practices"
  - Format: Context, Recommendation, Justification, Trade-offs

- **Methodology** (project-wide): docs/methodology/{methodology-name}.md
  - Comprehensive guides for reuse across projects
  - Example: "Code Review Methodology for Go Projects"
  - Format: Complete methodology with decision frameworks

**Knowledge Index Update**:
- Update knowledge/INDEX.md with:
  - New knowledge entries
  - Links to iteration where extracted
  - Domain tags (code-review, go, security, etc.)
  - Validation status (proposed, validated, refined)

### 6. State Transition

**Instance Layer** (Review State):
```yaml
s_{N-1} → s_N (Code Review):
  changes:
    - modules_reviewed: [list]
    - issues_identified: [count]
    - recommendations_made: [count]

  metrics:
    V_issue_detection: [value] (was: [previous])
    V_false_positive: [value] (was: [previous])
    V_actionability: [value] (was: [previous])
    V_learning: [value] (was: [previous])

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
  - Instance learnings (code review insights)
  - Meta learnings (methodology insights)
- What worked well?
- What challenges were encountered?
- What is needed next?
  - For review completion
  - For methodology completion

### 8. Convergence Check
[Use the convergence criteria structure above]

### 9. Data Artifacts

**Ephemeral Data** (iteration-specific, saved to data/):
- data/iteration-N-metrics.json (V_instance, V_meta calculations)
- data/iteration-N-review-state.yaml (modules reviewed, issues found)
- data/iteration-N-methodology.yaml (extracted patterns)
- data/iteration-N-artifacts/ (review reports, issue catalogs)

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
   - Instance: Perform concrete code review
   - Meta: Extract methodology patterns from review work

2. **Be Honest**: Calculate V(s_N) based on actual state
   - Don't inflate instance values (review quality)
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
   - Instance convergence: Review quality + completeness
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

## Code Review Domain-Specific Patterns

Based on OCA framework, code review iterations may follow:

### Observe Phase (Iterations 0-2)
- Iteration 0: Baseline establishment, codebase analysis
- Iteration 1: Manual code review of parser/ and analyzer/ modules
- Iteration 2: Manual code review of query/ and validation/ modules, pattern identification
- **Methodology**: Observe review patterns, identify issue types

### Codify Phase (Iterations 3-4)
- Iteration 3: Build issue classification taxonomy, document review patterns
- Iteration 4: Create review checklist, decision frameworks, automation strategies
- **Methodology**: Extract review decision frameworks, codify patterns

### Automate Phase (Iterations 5-6)
- Iteration 5: Implement linting rules, configure static analysis, install pre-commit hooks
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
- [ ] **READ** meta-agents/observe.md (code review observation strategies)
- [ ] Analyze review state (modules reviewed, issues found, patterns observed)
- [ ] Identify gaps (modules not reviewed, review aspects missing, methodology gaps)
- [ ] Observe patterns (for methodology extraction)

**Plan Phase**:
- [ ] **READ** meta-agents/plan.md (prioritization, agent selection)
- [ ] Define iteration goal (instance layer: module review, issue identification, automation)
- [ ] Assess agent sufficiency (inherited sufficient OR need specialized?)
- [ ] Plan methodology extraction (what patterns to document?)

**Execute Phase**:
- [ ] **READ** meta-agents/execute.md (coordination, pattern observation)
- [ ] **IF NEW AGENT NEEDED**:
  - [ ] **READ** meta-agents/evolve.md (agent creation criteria)
  - [ ] Create agent prompt file: agents/{agent-name}.md
  - [ ] Document specialization reason
- [ ] **READ** agent prompt file(s) before invocation: agents/{agent-name}.md
- [ ] Invoke agents to perform code review
- [ ] Observe review work for methodology patterns
- [ ] Extract patterns and update methodology documentation

**Reflect Phase**:
- [ ] **READ** meta-agents/reflect.md (evaluation process)
- [ ] Calculate V_instance(s_N) (review quality):
  - [ ] V_issue_detection (issues_found / total_actual_issues)
  - [ ] V_false_positive (1 - false_positives / total_issues_reported)
  - [ ] V_actionability (actionable_recommendations / total_recommendations)
  - [ ] V_learning (patterns_documented / patterns_identified)
- [ ] Calculate V_meta(s_N) (methodology quality):
  - [ ] V_completeness (documented_patterns / total_patterns)
  - [ ] V_effectiveness (1 - review_time_with_methodology / review_time_baseline)
  - [ ] V_reusability (successful_transfers / transfer_attempts)
- [ ] Assess quality honestly (don't inflate values)
- [ ] Identify gaps (review + methodology)

**Convergence Check**:
- [ ] M_N == M_{N-1}? (meta-agent stable)
- [ ] A_N == A_{N-1}? (agent set stable)
- [ ] V_instance(s_N) ≥ 0.80? (review quality threshold)
- [ ] V_meta(s_N) ≥ 0.80? (methodology quality threshold)
- [ ] Instance objectives complete? (all modules reviewed, issue catalog complete)
- [ ] Meta objectives complete? (methodology documented, transfer test successful)
- [ ] Determine: CONVERGED or NOT_CONVERGED

**Documentation**:
- [ ] Create iteration-N.md with:
  - [ ] Metadata (iteration, date, duration, status)
  - [ ] M evolution (evolved or unchanged)
  - [ ] A evolution (new agents or unchanged)
  - [ ] Instance work (modules reviewed, issues found, recommendations made)
  - [ ] Meta work (patterns extracted)
  - [ ] Knowledge artifacts created (list all new knowledge)
  - [ ] State transition (V_instance, V_meta calculated)
  - [ ] Reflection (learnings, next steps)
  - [ ] Convergence check (criteria evaluated)
- [ ] Save data artifacts to data/:
  - [ ] iteration-N-metrics.json
  - [ ] iteration-N-review-state.yaml
  - [ ] iteration-N-methodology.yaml
  - [ ] iteration-N-artifacts/ (review reports, issue catalogs)
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
- [ ] Document three-tuple: (O={review reports+methodology}, A_N, M_N)
- [ ] Validate methodology alignment (OCA, Bootstrapped SE, Value Space)

---

## Notes on Execution Style

**Be the Meta-Agent**: When executing iterations, embody M's perspective:
- Think through the observe-plan-execute-reflect-evolve cycle
- Read capability files to understand M's strategies
- Make explicit decisions about agent creation
- Justify why specialization is needed
- Extract methodology patterns from review work
- Track both V_instance and V_meta

**Be Domain-Aware**: Code review has specific concerns:
- Go language idioms and best practices
- Code correctness (bugs, edge cases, error handling)
- Code maintainability (complexity, duplication, coupling)
- Code readability (naming, structure, comments)
- Security vulnerabilities (input validation, etc.)
- Performance issues (efficiency, memory, concurrency)
- Test coverage and quality

**Be Rigorous**: Calculate values honestly
- V_instance based on actual review state (not aspirational)
- V_meta based on actual methodology coverage (not desired)
- Don't force convergence prematurely
- Don't skip methodology extraction to focus only on review
- Let data and needs drive the process

**Be Thorough**: Document decisions and reasoning
- Save intermediate data (metrics, review reports, issue catalogs)
- Show your work (calculations, analysis)
- Make evolution path traceable
- **NO TOKEN LIMITS**: Complete all steps fully, never abbreviate

**Be Two-Layer Aware**: Always work on both layers
- Instance layer: Perform concrete code review
- Meta layer: Extract reusable methodology
- Don't neglect either layer
- Methodology extraction is as important as review execution

**Be Authentic**: This is a real experiment
- Discover code review patterns, don't assume them
- Create agents based on need, not predetermined plan
- Extract methodology from actual review work, not theory
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
**Purpose**: Guide authentic execution of bootstrap-008-code-review experiment with dual-layer architecture

**Key Innovation**: Dual value functions (V_instance + V_meta) enable simultaneous optimization of concrete code review and reusable methodology.
