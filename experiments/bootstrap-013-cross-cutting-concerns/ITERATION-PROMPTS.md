# Iteration Execution Prompts

This document provides structured prompts for executing each iteration of the bootstrap-013-cross-cutting-concerns experiment.

**Two-Layer Architecture**:
- **Instance Layer**: Agents standardize cross-cutting concerns across meta-cc codebase
- **Meta Layer**: Meta-Agent observes pattern standardization work and extracts methodology

---

## Iteration 0: Baseline Establishment

```markdown
# Execute Iteration 0: Baseline Establishment

## Context
I'm starting the bootstrap-013-cross-cutting-concerns experiment. I've reviewed:
- experiments/bootstrap-013-cross-cutting-concerns/plan.md
- experiments/bootstrap-013-cross-cutting-concerns/README.md
- experiments/EXPERIMENTS-OVERVIEW.md (Bootstrap-013 specification)
- The three methodology frameworks (OCA, Bootstrapped SE, Value Space Optimization)

## Current State
- Meta-Agent: M₀ (5 capabilities inherited from Bootstrap-003: observe, plan, execute, reflect, evolve)
- Agent Set: A₀ (3 generic agents inherited from Bootstrap-003: data-analyst.md, doc-writer.md, coder.md)
- Target: All modules with logging, error handling, config (~5,000 lines across internal/ and cmd/)

## Inherited State from Bootstrap-003

**IMPORTANT**: This experiment starts with the converged state from Bootstrap-003, NOT from scratch.

**Meta-Agent capability files (ALREADY EXIST)**:
- experiments/bootstrap-013-cross-cutting-concerns/meta-agents/observe.md (validated)
- experiments/bootstrap-013-cross-cutting-concerns/meta-agents/plan.md (validated)
- experiments/bootstrap-013-cross-cutting-concerns/meta-agents/execute.md (validated)
- experiments/bootstrap-013-cross-cutting-concerns/meta-agents/reflect.md (validated)
- experiments/bootstrap-013-cross-cutting-concerns/meta-agents/evolve.md (validated)

**Agent prompt files (ALREADY EXIST - 3 generic agents)**:
- agents/data-analyst.md (data analysis and metrics)
- agents/doc-writer.md (documentation generation)
- agents/coder.md (code implementation)

**CRITICAL EXECUTION PROTOCOL**:
- All capability files and agent files ALREADY EXIST (inherited from Bootstrap-003)
- Before embodying Meta-Agent capabilities, ALWAYS read the relevant capability file first
- Before invoking ANY agent, ALWAYS read its prompt file first
- These files contain validated capabilities/agents; adapt them to cross-cutting concerns context
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
     - ✓ agents/data-analyst.md (generic agent for data analysis)
     - ✓ agents/doc-writer.md (generic agent for documentation)
     - ✓ agents/coder.md (generic agent for code implementation)
   - **NO NEED TO CREATE NEW FILES** - all files inherited and ready to use
   - **ADAPTATION NOTE**: Capability files and agent files are generic enough to apply to cross-cutting concerns;
     read them to understand their validated approaches, then apply to pattern standardization context

1. **Codebase Pattern Analysis** (M₀.observe):
   - **READ** meta-agents/observe.md (pattern observation strategies)
   - Analyze cross-cutting concerns across codebase:
     - **Logging patterns**:
       - grep -r "log\\." internal/ cmd/ (identify all log statements)
       - Catalog logging styles: fmt.Printf, log.Printf, custom logger
       - Analyze log levels used: debug, info, warn, error
       - Identify inconsistencies: missing context, varying formats
       - Count files with logging: total vs. consistent style
     - **Error handling patterns**:
       - grep -r "if err != nil" internal/ cmd/ (identify all error checks)
       - Catalog error patterns: wrap, return, log, ignore
       - Analyze error context: sufficient vs. insufficient
       - Identify inconsistencies: bare returns, lost context
       - Count error handling styles per module
     - **Configuration patterns**:
       - grep -r "os.Getenv" "viper" "config" internal/ cmd/ (identify config access)
       - Catalog config styles: environment vars, config files, flags
       - Analyze config validation: present vs. absent
       - Identify inconsistencies: hardcoded values, missing defaults
       - Count configuration approaches per module
     - **Target modules** (~5,000 lines):
       - internal/parser/ (~3,500 lines): Core parsing logic
       - internal/analyzer/ (~2,800 lines): Pattern detection
       - internal/query/ (~3,200 lines): Query engine
       - cmd/mcp/ (~2,500 lines): MCP server
       - cmd/meta-cc/ (~1,800 lines): CLI tool
   - Use meta-cc query tools for additional insights:
     - meta-cc query-files --threshold 10 (high-access files likely to have patterns)
     - meta-cc query-tools --tool Edit | grep "error\|log\|config" (pattern evolution)
     - meta-cc query-user-messages --pattern "log|error|config" (developer pain points)

2. **Pattern Inventory Creation** (M₀.observe + data-analyst):
   - **READ** meta-agents/observe.md (observation strategies)
   - **READ** agents/data-analyst.md
   - Invoke data-analyst to create pattern inventory:
     - **Logging patterns found**:
       - List all unique logging styles with file locations
       - Count occurrences of each pattern
       - Identify most common vs. most inconsistent
       - Example: "fmt.Printf used in 15 files, log.Printf in 8 files, custom logger in 3 files"
     - **Error handling patterns found**:
       - List all unique error handling approaches
       - Count occurrences (wrap, return, log+return, ignore)
       - Identify inconsistencies and anti-patterns
       - Example: "Wrapped errors in 60% of cases, bare returns in 40%"
     - **Configuration patterns found**:
       - List all unique config access methods
       - Count occurrences (env vars, files, flags, hardcoded)
       - Identify missing validation and defaults
       - Example: "os.Getenv in 20 places, 12 without defaults"
     - **Consistency metrics**:
       - Calculate pattern consistency per module
       - Identify modules with highest/lowest consistency
       - Estimate overall consistency baseline (~30-40% expected)

3. **Baseline Metrics Calculation** (M₀.plan + data-analyst):
   - **READ** meta-agents/plan.md (prioritization strategies)
   - **READ** agents/data-analyst.md
   - Invoke data-analyst to calculate V_instance(s₀):
     - V_consistency: Uniform patterns across codebase
       - Calculate: files_with_standard_pattern / total_files_with_concern
       - Per concern: logging, error handling, config
       - Overall: weighted average across concerns
       - Baseline estimate: ~0.30-0.40 (low, manual ad-hoc patterns)
     - V_maintainability: Easy to update patterns
       - Assess: How easy to change logging format? Error wrapping style?
       - Estimate: centralized=1.0, scattered=0.0
       - Baseline estimate: ~0.20-0.30 (scattered, hard to change)
     - V_enforcement: Automated pattern checking
       - Count: automated checks / total pattern types
       - Current state: gofmt only (not pattern-aware)
       - Baseline estimate: ~0.10 (minimal automation)
     - V_documentation: Patterns well-documented
       - Count: documented patterns / total patterns identified
       - Current state: implicit patterns, no formal docs
       - Baseline estimate: ~0.05-0.10 (virtually undocumented)
     - **V_instance(s₀) = 0.4×V_consistency + 0.3×V_maintainability + 0.2×V_enforcement + 0.1×V_documentation**
     - Expected: V_instance(s₀) ≈ 0.25-0.35 (low baseline, significant improvement opportunity)
   - Calculate V_meta(s₀):
     - V_completeness: 0.00 (no methodology yet)
     - V_effectiveness: 0.00 (nothing to test)
     - V_reusability: 0.00 (nothing to transfer)
     - **V_meta(s₀) = 0.00**

4. **Gap Identification** (M₀.reflect):
   - **READ** meta-agents/reflect.md (gap analysis process)
   - What standardization capabilities are missing?
     - No pattern extraction process
     - No convention definition framework
     - No automated enforcement (linters)
     - No code generation templates
     - No migration planning approach
   - What patterns need standardization?
     - **Logging**: Unified logger interface, structured logging, log levels
     - **Error handling**: Consistent wrapping, context preservation, error types
     - **Configuration**: Centralized config, validation, defaults, documentation
     - **Additional concerns** (discovered during analysis):
       - Resource cleanup (defer patterns)
       - Concurrency patterns (sync, channels)
       - Testing patterns (mocks, fixtures)
   - What methodology components are needed?
     - Pattern extraction framework
     - Convention definition process
     - Linter generation strategy
     - Template creation approach
     - Migration planning methodology
     - Documentation standards

5. **Initial Agent Applicability Assessment** (M₀.plan):
   - **READ** meta-agents/plan.md (agent selection strategies)
   - Which inherited agents are directly applicable to cross-cutting concerns?
     - ⭐⭐⭐ data-analyst: Analyze pattern occurrences, calculate metrics
     - ⭐⭐ coder: Implement linters, create templates, write migration scripts
     - ⭐ doc-writer: Document patterns and conventions
   - What new specialized agents might be needed?
     - pattern-extractor: Identify and catalog existing patterns in codebase
       - Capabilities: AST parsing, pattern matching, categorization
       - Why needed: Systematic pattern discovery beyond grep
     - convention-definer: Define standard patterns for concerns
       - Capabilities: Analyze patterns, propose conventions, document standards
       - Why needed: Expert judgment for best-practice selection
     - linter-generator: Generate custom linters for pattern enforcement
       - Capabilities: go/analysis framework, AST rules, golangci-lint integration
       - Why needed: Automated enforcement requires specialized knowledge
     - template-creator: Create code generation templates
       - Capabilities: Template design, code generation, usage documentation
       - Why needed: Systematic pattern application
     - migration-planner: Plan migration from ad-hoc to systematic patterns
       - Capabilities: Impact analysis, sequencing, risk assessment
       - Why needed: Safe, incremental migration strategy
   - **NOTE**: Don't create new agents yet - just identify potential needs

6. **Documentation** (M₀.execute + doc-writer):
   - **READ** meta-agents/execute.md (coordination strategies)
   - **READ** agents/doc-writer.md
   - Invoke doc-writer to create iteration-0.md:
     - M₀ state: 5 capabilities inherited from Bootstrap-003 (observe, plan, execute, reflect, evolve)
     - A₀ state: 3 generic agents inherited from Bootstrap-003 (data-analyst, doc-writer, coder)
     - Agent applicability assessment (which agents useful for pattern standardization)
     - Pattern analysis summary (logging, error handling, config across ~5,000 lines)
     - Pattern inventory (all discovered patterns with counts and locations)
     - Current consistency assessment (low, ~30-40% baseline)
     - Calculated V_instance(s₀) ≈ 0.25-0.35 and V_meta(s₀) = 0.00
     - Gap analysis (missing pattern extraction, conventions, enforcement, templates, methodology)
     - Reflection on next steps and agent needs
   - Save data artifacts:
     - data/s0-pattern-inventory.yaml (all patterns found: logging, errors, config)
     - data/s0-pattern-locations.yaml (file locations for each pattern)
     - data/s0-consistency-metrics.json (calculated metrics per module and concern)
     - data/s0-metrics.json (calculated V_instance and V_meta values)
     - data/s0-gaps.yaml (identified gaps in standardization and methodology)
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
     - Additional AST analysis for deeper pattern understanding?
     - Historical analysis of pattern evolution in git history?
   - Are M₀ capabilities sufficient for baseline? (Yes, core capabilities adequate)
   - What should be the focus of Iteration 1?
     - Likely: Pattern extraction for logging (most pervasive concern)
     - Or: Define standard conventions for all concerns first
     - Decision based on OCA framework: Start with Observe phase (pattern extraction)
   - Which agents will be needed in Iteration 1?
     - Likely: Need specialized pattern-extractor for systematic discovery
     - Reuse: data-analyst (metrics), doc-writer (documentation)
   - Methodology extraction readiness:
     - Note pattern discovery approaches used
     - Identify decision points that could become methodology
     - Prepare for methodology documentation in subsequent iterations

## Constraints
- Do NOT pre-decide what agents to create next
- Do NOT assume the standardization approach or evolution path
- Let the pattern data and gaps guide next steps
- Be honest about current consistency state (low baseline expected)
- Calculate V(s₀) based on actual observations, not target values
- Remember two layers: concrete standardization (instance) + methodology (meta)

## Output Format
Create iteration-0.md following this structure:
- Iteration metadata (number, date, duration)
- M₀ state documentation (5 capabilities inherited from Bootstrap-003)
- A₀ state documentation (3 generic agents inherited from Bootstrap-003)
- Agent applicability assessment (which agents useful for pattern standardization)
- Pattern analysis (logging, error handling, config across ~5,000 lines)
- Pattern inventory (all discovered patterns with counts and locations)
- Consistency metrics (V_consistency, V_maintainability, V_enforcement, V_documentation)
- Value calculation (V_instance(s₀) ≈ 0.25-0.35, V_meta(s₀) = 0.00)
- Gap identification (missing pattern extraction, conventions, enforcement, methodology)
- Reflection on next steps and agent needs
- Data artifacts saved to data/ directory
```

---

## Iteration 1+: Subsequent Iterations (General Template)

```markdown
# Execute Iteration N: [To be determined by Meta-Agent]

## Context from Previous Iteration

Review the previous iteration file: experiments/bootstrap-013-cross-cutting-concerns/iteration-[N-1].md

Extract:
- Current Meta-Agent state: M_{N-1}
- Current Agent Set: A_{N-1}
- Current standardization state: V_instance(s_{N-1})
- Current methodology state: V_meta(s_{N-1})
- Problems identified
- Reflection notes on what's needed next

## Two-Layer Execution Protocol

**Layer 1 (Instance)**: Agents perform concrete pattern standardization
**Layer 2 (Meta)**: Meta-Agent observes and extracts methodology

Throughout iteration:
- Agents focus on concrete tasks (extract patterns, define conventions, implement linters, create templates, migrate code)
- Meta-Agent observes standardization work and identifies patterns for methodology

## Meta-Agent Decision Process

**BEFORE STARTING**: Read relevant Meta-Agent capability files:
- **READ** meta-agents/observe.md (for observation strategies)
- **READ** meta-agents/plan.md (for planning and decisions)
- **READ** meta-agents/execute.md (for coordination)
- **READ** meta-agents/reflect.md (for evaluation)
- **READ** meta-agents/evolve.md (for evolution assessment)

As M_{N-1}, follow the five-capability process:

### 1. OBSERVE (M.observe)
- **READ** meta-agents/observe.md for pattern observation strategies
- Review previous iteration outputs (iteration-[N-1].md)
- Examine standardization state:
  - What patterns have been extracted? (if any)
  - What conventions have been defined? (if any)
  - What enforcement mechanisms exist? (if any)
  - What code has been migrated? (if any)
- Identify gaps:
  - What concerns still need pattern extraction?
  - What conventions are missing or incomplete?
  - What enforcement automation is needed?
  - What migration work remains?
- **Methodology observation**:
  - What patterns emerged in previous standardization work?
  - What decisions were made and why?
  - What principles can be extracted?

### 2. PLAN (M.plan)
- **READ** meta-agents/plan.md for prioritization and agent selection
- Based on observations, what is the primary goal for this iteration?
  - Examples:
    - "Extract logging patterns and define standard convention"
    - "Implement custom linter for error handling enforcement"
    - "Create code generation templates for configuration"
    - "Migrate internal/parser/ to standardized patterns"
    - "Document pattern library and usage guidelines"
- What capabilities are needed to achieve this goal?
- **Agent Assessment**:
  - Are current agents (A_{N-1}) sufficient for this goal?
  - Is generic `coder` enough to write linters?
  - Or do we need specialized `linter-generator` for go/analysis expertise?
  - Do we need `pattern-extractor` for AST-based discovery?
  - Do we need `convention-definer` for standard definition?
  - Do we need `migration-planner` for safe code updates?
- **Methodology Planning**:
  - What patterns should be documented this iteration?
  - What standardization decisions will inform methodology?

### 3. EXECUTE (M.execute)
- **READ** meta-agents/execute.md for coordination and pattern observation
- Decision point: Should I create a new specialized agent?

**IF current agents are insufficient:**
- **EVOLVE** (M.evolve): Create new specialized agent
  - **READ** meta-agents/evolve.md for agent creation criteria
  - Examples of specialized agents:
    - `pattern-extractor`: Identify existing patterns in codebase
      - Capabilities: AST parsing (go/ast), pattern matching, categorization, inventory
      - Why needed: Systematic pattern discovery beyond simple grep
    - `convention-definer`: Define standard patterns for concerns
      - Capabilities: Analyze patterns, research best practices, propose conventions, document standards
      - Why needed: Expert judgment for selecting best patterns from alternatives
    - `linter-generator`: Generate custom linters for pattern enforcement
      - Capabilities: go/analysis framework knowledge, AST rule writing, golangci-lint integration
      - Why needed: Specialized expertise in Go static analysis tooling
    - `template-creator`: Create code generation templates
      - Capabilities: Template design (text/template), code generation, usage documentation
      - Why needed: Systematic approach to reusable code patterns
    - `migration-planner`: Plan migration from ad-hoc to systematic patterns
      - Capabilities: Impact analysis, dependency ordering, risk assessment, rollback planning
      - Why needed: Safe, incremental migration strategy across large codebase
  - Define agent name and specialization domain
  - Document capabilities the new agent provides
  - Explain why inherited agents are insufficient
  - **CREATE AGENT PROMPT FILE**: Write agents/{agent-name}.md
    - Include: agent role, cross-cutting concerns-specific capabilities, input/output format
    - Include: specific instructions for this iteration's task
    - Include: Go domain knowledge (AST, go/analysis, best practices)
  - Add to agent set: A_N = A_{N-1} ∪ {new_agent}

**Agent Invocation** (specialized or inherited):
- **READ agent prompt file** before invocation: agents/{agent-name}.md
- Invoke agent to execute concrete standardization work:
  - Extract patterns from codebase (AST analysis, grep, manual review)
  - Define conventions (research, propose, document standards)
  - Implement enforcement (custom linters, golangci-lint rules, pre-commit hooks)
  - Create templates (code generation, reusable patterns)
  - Migrate code (systematic pattern replacement, testing)
- Produce iteration outputs (pattern library, conventions, linters, templates, migrated code)

**Methodology Extraction** (M.evolve):
- **OBSERVE agent work patterns**:
  - How did agent organize pattern extraction process?
  - What criteria were used to select standard conventions?
  - How were linters designed and implemented?
  - What migration strategy was chosen?
- **EXTRACT patterns for methodology**:
  - Document standardization decision frameworks
  - Build pattern selection criteria
  - Identify reusable standardization patterns
  - Note principles that emerge
  - Add to methodology documentation

**ELSE use inherited agents:**
- **READ agent prompt file** from agents/{agent-name}.md
- Invoke appropriate agents from A_{N-1}
- Execute planned standardization work
- Observe for methodology patterns

**CRITICAL EXECUTION PROTOCOL**:
1. ALWAYS read capability files before embodying Meta-Agent capabilities
2. ALWAYS read agent prompt file before each agent invocation
3. Do NOT cache instructions across iterations - always read from files
4. Capability files may be updated between iterations - get latest from files
5. Never assume capabilities - always verify from source files

### 4. REFLECT (M.reflect)
- **READ** meta-agents/reflect.md for evaluation process
- **Evaluate Instance Layer** (Concrete Standardization):
  - What standardization work was completed this iteration?
  - Calculate new V_instance(s_N):
    - V_consistency: files_with_standard_pattern / total_files_with_concern
      - Count files now using standard patterns (per concern: logging, errors, config)
      - Calculate overall consistency across all concerns
    - V_maintainability: ease_of_pattern_update
      - Assess: Centralized (1.0) vs. scattered (0.0) pattern definitions
      - Count: files that would need changes to update pattern
      - Calculate: 1 - (files_to_change / total_files)
    - V_enforcement: automated_checks / total_pattern_types
      - Count: linters implemented for each concern
      - Count: pre-commit hooks enforcing patterns
      - Calculate: enforcement_coverage / total_concerns
    - V_documentation: documented_patterns / total_patterns_identified
      - Count: patterns with formal documentation
      - Count: patterns with usage examples
      - Calculate: documentation_coverage
    - **V_instance(s_N) = 0.4×V_consistency + 0.3×V_maintainability + 0.2×V_enforcement + 0.1×V_documentation**
  - Calculate change: ΔV_instance = V_instance(s_N) - V_instance(s_{N-1})
  - Are standardization objectives met? What gaps remain?

- **Evaluate Meta Layer** (Methodology):
  - What patterns were extracted this iteration?
  - Calculate new V_meta(s_N):
    - V_completeness: documented_methodology_components / required_components
      - Required: Pattern extraction framework, convention definition process, linter generation strategy, template creation approach, migration planning methodology
      - Count documented vs required (5 minimum)
    - V_effectiveness: 1 - (standardization_time_with_methodology / standardization_time_baseline)
      - Measure standardization time if methodology used
      - Compare to baseline (~40-60 hours for 5K lines ad-hoc)
      - Later iterations: Transfer test to different codebase
    - V_reusability: successful_transfers / transfer_attempts
      - Test methodology on different codebase (e.g., cmd/ vs internal/)
      - Assess if methodology applies with <20% modification
    - **V_meta(s_N) = 0.4×V_completeness + 0.3×V_effectiveness + 0.3×V_reusability**
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
    details: "If no new pattern standardization specialization needed"

  instance_value_threshold:
    question: "Is V_instance(s_N) ≥ 0.80 (standardization quality)?"
    V_instance(s_N): [calculated value]
    threshold_met: [Yes/No]
    components:
      V_consistency: [value] "≥0.80 target (80% consistency)"
      V_maintainability: [value] "≥0.80 target (easy updates)"
      V_enforcement: [value] "≥0.80 target (80% automated)"
      V_documentation: [value] "≥0.80 target (well-documented)"

  meta_value_threshold:
    question: "Is V_meta(s_N) ≥ 0.80 (methodology quality)?"
    V_meta(s_N): [calculated value]
    threshold_met: [Yes/No]
    components:
      V_completeness: [value] "≥0.90 target (all components)"
      V_effectiveness: [value] "≥0.75 target (3x speedup)"
      V_reusability: [value] "≥0.75 target (75% transfer)"

  instance_objectives:
    patterns_extracted: [Yes/No] "All concerns analyzed"
    conventions_defined: [Yes/No] "Standards documented"
    enforcement_implemented: [Yes/No] "Linters active"
    templates_created: [Yes/No] "Code generation ready"
    migration_complete: [Yes/No] "80% of code standardized"
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
  - O = {pattern library + linters + templates + migrated code + methodology documentation}
  - A_N = {final agent set with specializations}
  - M_N = {final meta-agent capabilities}

**IF NOT CONVERGED:**
- Identify what's needed for next iteration
- Note: Focus on instance (standardization) OR meta (methodology) as needed
- Continue to Iteration N+1

## Documentation Requirements

Create experiments/bootstrap-013-cross-cutting-concerns/iteration-N.md with:

### 1. Iteration Metadata
```yaml
iteration: N
date: YYYY-MM-DD
duration: ~X hours
status: [completed/converged]
layers:
  instance: "Standardization work performed this iteration"
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

  inherited_baseline: "3 generic agents from Bootstrap-003 (data-analyst, doc-writer, coder)"

  IF evolved (new agent created):
    new_agents:
      - name: agent_name
        specialization: pattern_standardization_domain_area
        capabilities: [list]
        creation_reason: "Why were inherited 3 agents insufficient?"
        justification: "What gap exists that no inherited agent can fill?"
        prompt_file: "agents/{agent-name}.md"

  IF unchanged:
    status: "A_N = A_{N-1} (inherited agents sufficient for this iteration)"

  IF reused_inherited:
    reused_agents:
      - agent_name: task_performed
        adaptation_notes: "How inherited agent was adapted to standardization context"

agents_invoked_this_iteration:
  - agent_name: task_performed
    source: [inherited / newly_created]
```

### 4. Instance Work Executed (Concrete Standardization)
- What standardization work was performed?
  - Pattern extraction (which concerns analyzed?)
  - Convention definition (which standards documented?)
  - Enforcement implementation (which linters created?)
  - Template creation (which templates developed?)
  - Code migration (which modules migrated?)
- What outputs were produced?
  - Pattern library (documented patterns)
  - Convention documents (standards)
  - Linters (custom rules)
  - Templates (code generation)
  - Migrated code (standardized modules)
- Summary of concrete deliverables

### 5. Meta Work Executed (Methodology Extraction)
- What patterns were observed in standardization work?
  - Pattern extraction approaches
  - Convention selection criteria
  - Linter design decisions
  - Template design patterns
  - Migration strategies
- What methodology content was documented?
  - Pattern extraction frameworks
  - Convention definition processes
  - Linter generation strategies
  - Template creation approaches
  - Migration planning methodologies

**Knowledge Artifacts Created**:

Organize extracted knowledge into appropriate categories:

- **Patterns** (domain-specific): knowledge/patterns/{pattern-name}.md
  - Specific solutions to recurring problems in cross-cutting concerns
  - Example: "Structured Logging Pattern", "Error Context Preservation Pattern"
  - Format: Problem, Context, Solution, Consequences, Examples

- **Principles** (universal): knowledge/principles/{principle-name}.md
  - Fundamental truths or rules discovered
  - Example: "Consistency Over Perfection Principle", "Incremental Migration Principle"
  - Format: Statement, Rationale, Evidence, Applications

- **Templates** (reusable): knowledge/templates/{template-name}.{go|md|yaml}
  - Concrete implementations ready for reuse
  - Example: "Structured Logger Template", "Error Wrapper Template", "Config Validation Template"
  - Format: Template file + usage documentation

- **Best Practices** (context-specific): knowledge/best-practices/{topic}.md
  - Recommended approaches for specific contexts
  - Example: "Go Logging Best Practices", "Go Error Handling Best Practices"
  - Format: Context, Recommendation, Justification, Trade-offs

- **Methodology** (project-wide): docs/methodology/{methodology-name}.md
  - Comprehensive guides for reuse across projects
  - Example: "Cross-Cutting Concerns Standardization Methodology"
  - Format: Complete methodology with decision frameworks

**Knowledge Index Update**:
- Update knowledge/INDEX.md with:
  - New knowledge entries
  - Links to iteration where extracted
  - Domain tags (logging, error-handling, config, patterns, etc.)
  - Validation status (proposed, validated, refined)

### 6. State Transition

**Instance Layer** (Standardization State):
```yaml
s_{N-1} → s_N (Pattern Standardization):
  changes:
    - patterns_extracted: [list]
    - conventions_defined: [list]
    - enforcement_added: [list]
    - templates_created: [list]
    - code_migrated: [modules list]

  metrics:
    V_consistency: [value] (was: [previous])
    V_maintainability: [value] (was: [previous])
    V_enforcement: [value] (was: [previous])
    V_documentation: [value] (was: [previous])

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
  - Instance learnings (pattern standardization insights)
  - Meta learnings (methodology insights)
- What worked well?
- What challenges were encountered?
- What is needed next?
  - For standardization completion
  - For methodology completion

### 8. Convergence Check
[Use the convergence criteria structure above]

### 9. Data Artifacts

**Ephemeral Data** (iteration-specific, saved to data/):
- data/iteration-N-metrics.json (V_instance, V_meta calculations)
- data/iteration-N-standardization-state.yaml (patterns, conventions, enforcement status)
- data/iteration-N-methodology.yaml (extracted patterns)
- data/iteration-N-artifacts/ (linters, templates, migrated code)

**Permanent Knowledge** (cumulative, saved to knowledge/):
- knowledge/patterns/{pattern-name}.md (domain-specific patterns)
- knowledge/principles/{principle-name}.md (universal principles)
- knowledge/templates/{template-name}.{go|md|yaml} (reusable templates)
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
   - Instance: Perform concrete pattern standardization
   - Meta: Extract methodology patterns from standardization work

2. **Be Honest**: Calculate V(s_N) based on actual state
   - Don't inflate instance values (standardization quality)
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
   - Instance convergence: Standardization quality + completeness
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

## Cross-Cutting Concerns Domain-Specific Patterns

Based on OCA framework, pattern standardization iterations may follow:

### Observe Phase (Iterations 0-2)
- Iteration 0: Baseline establishment, pattern inventory creation
- Iteration 1: Pattern extraction for logging (AST analysis, categorization)
- Iteration 2: Pattern extraction for error handling and configuration
- **Methodology**: Observe pattern discovery approaches, identify classification frameworks

### Codify Phase (Iterations 3-4)
- Iteration 3: Define standard conventions for all concerns (logging, errors, config)
- Iteration 4: Implement custom linters and enforcement automation
- **Methodology**: Extract convention selection criteria, document linter generation strategies

### Automate Phase (Iterations 5-6)
- Iteration 5: Create code generation templates, migrate internal/ modules
- Iteration 6: Migrate cmd/ modules, transfer test methodology, convergence
- **Methodology**: Document template creation approaches, validate migration strategies

**Caveat**: Let actual needs drive the sequence, not this expected pattern. The pattern is a hint, not a prescription.
```

---

## Quick Reference: Iteration Checklist

For each iteration N ≥ 1, ensure you:

**Preparation**:
- [ ] Review previous iteration (iteration-[N-1].md)
- [ ] Extract current state (M_{N-1}, A_{N-1}, V_instance(s_{N-1}), V_meta(s_{N-1}))

**Observe Phase**:
- [ ] **READ** meta-agents/observe.md (pattern observation strategies)
- [ ] Analyze standardization state (patterns, conventions, enforcement, migration)
- [ ] Identify gaps (concerns not standardized, conventions missing, enforcement gaps)
- [ ] Observe patterns (for methodology extraction)

**Plan Phase**:
- [ ] **READ** meta-agents/plan.md (prioritization, agent selection)
- [ ] Define iteration goal (instance layer: pattern extraction, convention definition, linter implementation, migration)
- [ ] Assess agent sufficiency (inherited sufficient OR need specialized?)
- [ ] Plan methodology extraction (what patterns to document?)

**Execute Phase**:
- [ ] **READ** meta-agents/execute.md (coordination, pattern observation)
- [ ] **IF NEW AGENT NEEDED**:
  - [ ] **READ** meta-agents/evolve.md (agent creation criteria)
  - [ ] Create agent prompt file: agents/{agent-name}.md
  - [ ] Document specialization reason
- [ ] **READ** agent prompt file(s) before invocation: agents/{agent-name}.md
- [ ] Invoke agents to perform standardization work
- [ ] Observe standardization work for methodology patterns
- [ ] Extract patterns and update methodology documentation

**Reflect Phase**:
- [ ] **READ** meta-agents/reflect.md (evaluation process)
- [ ] Calculate V_instance(s_N) (standardization quality):
  - [ ] V_consistency (files_with_standard_pattern / total_files_with_concern)
  - [ ] V_maintainability (ease of pattern updates, centralization)
  - [ ] V_enforcement (automated_checks / total_pattern_types)
  - [ ] V_documentation (documented_patterns / total_patterns_identified)
- [ ] Calculate V_meta(s_N) (methodology quality):
  - [ ] V_completeness (documented_components / required_components)
  - [ ] V_effectiveness (1 - standardization_time_with_methodology / baseline_time)
  - [ ] V_reusability (successful_transfers / transfer_attempts)
- [ ] Assess quality honestly (don't inflate values)
- [ ] Identify gaps (standardization + methodology)

**Convergence Check**:
- [ ] M_N == M_{N-1}? (meta-agent stable)
- [ ] A_N == A_{N-1}? (agent set stable)
- [ ] V_instance(s_N) ≥ 0.80? (standardization quality threshold)
- [ ] V_meta(s_N) ≥ 0.80? (methodology quality threshold)
- [ ] Instance objectives complete? (patterns extracted, conventions defined, enforcement active, migration complete)
- [ ] Meta objectives complete? (methodology documented, transfer test successful)
- [ ] Determine: CONVERGED or NOT_CONVERGED

**Documentation**:
- [ ] Create iteration-N.md with:
  - [ ] Metadata (iteration, date, duration, status)
  - [ ] M evolution (evolved or unchanged)
  - [ ] A evolution (new agents or unchanged)
  - [ ] Instance work (patterns, conventions, linters, templates, migrated code)
  - [ ] Meta work (patterns extracted)
  - [ ] Knowledge artifacts created (list all new knowledge)
  - [ ] State transition (V_instance, V_meta calculated)
  - [ ] Reflection (learnings, next steps)
  - [ ] Convergence check (criteria evaluated)
- [ ] Save data artifacts to data/:
  - [ ] iteration-N-metrics.json
  - [ ] iteration-N-standardization-state.yaml
  - [ ] iteration-N-methodology.yaml
  - [ ] iteration-N-artifacts/ (linters, templates, migrated code)
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
- [ ] Document three-tuple: (O={pattern library+linters+templates+migrated code+methodology}, A_N, M_N)
- [ ] Validate methodology alignment (OCA, Bootstrapped SE, Value Space)

---

## Notes on Execution Style

**Be the Meta-Agent**: When executing iterations, embody M's perspective:
- Think through the observe-plan-execute-reflect-evolve cycle
- Read capability files to understand M's strategies
- Make explicit decisions about agent creation
- Justify why specialization is needed
- Extract methodology patterns from standardization work
- Track both V_instance and V_meta

**Be Domain-Aware**: Cross-cutting concerns have specific characteristics:
- **Logging**: Structured logging, log levels, context propagation, performance
- **Error handling**: Error wrapping, context preservation, error types, recovery
- **Configuration**: Validation, defaults, environment-specific, documentation
- **Go best practices**: Idioms, patterns, anti-patterns, standard library usage
- **AST manipulation**: go/ast, go/analysis, static analysis tooling
- **Code generation**: text/template, code templates, usage documentation

**Be Rigorous**: Calculate values honestly
- V_instance based on actual standardization state (not aspirational)
- V_meta based on actual methodology coverage (not desired)
- Don't force convergence prematurely
- Don't skip methodology extraction to focus only on standardization
- Let data and needs drive the process

**Be Thorough**: Document decisions and reasoning
- Save intermediate data (patterns, conventions, linters, templates)
- Show your work (calculations, analysis)
- Make evolution path traceable
- **NO TOKEN LIMITS**: Complete all steps fully, never abbreviate

**Be Two-Layer Aware**: Always work on both layers
- Instance layer: Perform concrete pattern standardization
- Meta layer: Extract reusable methodology
- Don't neglect either layer
- Methodology extraction is as important as standardization execution

**Be Authentic**: This is a real experiment
- Discover standardization patterns, don't assume them
- Create agents based on need, not predetermined plan
- Extract methodology from actual standardization work, not theory
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
**Purpose**: Guide authentic execution of bootstrap-013-cross-cutting-concerns experiment with dual-layer architecture

**Key Innovation**: Dual value functions (V_instance + V_meta) enable simultaneous optimization of concrete pattern standardization and reusable methodology.
