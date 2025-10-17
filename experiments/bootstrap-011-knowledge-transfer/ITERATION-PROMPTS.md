# Iteration Execution Prompts

This document provides structured prompts for executing each iteration of the bootstrap-011-knowledge-transfer experiment.

**Two-Layer Architecture**:
- **Instance Layer**: Agents create comprehensive onboarding materials for meta-cc project
- **Meta Layer**: Meta-Agent observes onboarding creation patterns and extracts knowledge transfer methodology

---

## Iteration 0: Baseline Establishment

```markdown
# Execute Iteration 0: Baseline Establishment

## Context
I'm starting the bootstrap-011-knowledge-transfer experiment. I've reviewed:
- experiments/EXPERIMENTS-OVERVIEW.md (Bootstrap-011 specification)
- experiments/bootstrap-011-knowledge-transfer/README.md (if exists)
- The three methodology frameworks (OCA, Bootstrapped SE, Value Space Optimization)

## Current State
- Meta-Agent: M₀ (5 capabilities inherited from Bootstrap-003: observe, plan, execute, reflect, evolve)
- Agent Set: A₀ (6 agents inherited from Bootstrap-003: 3 generic + 3 specialized)
- Target: meta-cc project onboarding materials (Day-1, Week-1, Month-1 paths)

## Inherited State from Bootstrap-003

**IMPORTANT**: This experiment starts with the converged state from Bootstrap-003, NOT from scratch.

**Meta-Agent capability files (ALREADY EXIST)**:
- experiments/bootstrap-011-knowledge-transfer/meta-agents/observe.md (validated)
- experiments/bootstrap-011-knowledge-transfer/meta-agents/plan.md (validated)
- experiments/bootstrap-011-knowledge-transfer/meta-agents/execute.md (validated)
- experiments/bootstrap-011-knowledge-transfer/meta-agents/reflect.md (validated)
- experiments/bootstrap-011-knowledge-transfer/meta-agents/evolve.md (validated)

**Agent prompt files (ALREADY EXIST - 6 agents)**:
- Generic (3): data-analyst.md, doc-writer.md, coder.md
- From Bootstrap-003 (3): error-classifier.md, recovery-advisor.md, root-cause-analyzer.md

**CRITICAL EXECUTION PROTOCOL**:
- All capability files and agent files ALREADY EXIST (inherited from Bootstrap-003)
- Before embodying Meta-Agent capabilities, ALWAYS read the relevant capability file first
- Before invoking ANY agent, ALWAYS read its prompt file first
- These files contain validated capabilities/agents; adapt them to knowledge transfer context
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
   - **ADAPTATION NOTE**: Capability files and agent files are generic enough to apply to knowledge transfer;
     read them to understand their validated approaches, then apply to onboarding context

1. **Current Knowledge State Analysis** (M₀.observe):
   - **READ** meta-agents/observe.md (knowledge transfer observation strategies)
   - Analyze existing meta-cc documentation:
     - Documentation inventory:
       - README.md (quick start)
       - CLAUDE.md (development guide)
       - docs/ (technical documentation)
       - EXPERIMENTS-OVERVIEW.md (experiment catalog)
     - Query session history for knowledge-seeking patterns:
       - meta-cc query-user-messages --pattern "how|what|where|why|explain"
       - Categorize questions by topic (architecture, MCP, testing, workflow, etc.)
       - Identify question frequency → common onboarding pain points
     - Analyze file access patterns:
       - meta-cc query-files --threshold 10 (frequently accessed files = important)
       - Identify "entry point" files (high initial access)
       - Identify "reference" files (repeatedly accessed)
     - Examine workflow patterns:
       - meta-cc query-tool-sequences --min-occurrences 3 (common workflows)
       - Identify typical developer workflows
       - Map workflows to learning stages
     - Git log analysis:
       - Contributor patterns (who works on what)
       - Code ownership (git blame analysis)
       - Evolution patterns (which areas change most)

2. **Current Onboarding Process Assessment** (M₀.observe + M₀.plan):
   - **READ** meta-agents/observe.md (observation strategies)
   - **READ** meta-agents/plan.md (assessment frameworks)
   - What onboarding process currently exists?
     - Manual onboarding: README.md + CLAUDE.md (basic)
     - Documentation navigation: Manual exploration (no guided paths)
     - Code discovery: Manual grep/search (no navigation tools)
     - Expert identification: None (no contributor map)
   - What onboarding artifacts exist?
     - README.md: Installation and quick start
     - CLAUDE.md: Development workflow and common tasks
     - docs/: Technical documentation (scattered)
     - No role-specific paths (contributor vs. user vs. maintainer)
     - No time-based paths (day 1 vs. week 1 vs. month 1)
   - What gaps exist?
     - No learning path guidance (where to start?)
     - No code navigation tools (how to explore codebase?)
     - No expert map (who to ask about what?)
     - No onboarding checklist (how to verify progress?)
     - No context-aware recommendations (right info at right time?)
     - No freshness tracking (is documentation up-to-date?)

3. **Baseline Metrics Calculation** (M₀.plan + data-analyst):
   - **READ** meta-agents/plan.md (prioritization strategies)
   - **READ** agents/data-analyst.md
   - Invoke data-analyst to calculate V_instance(s₀):
     - V_discoverability: How easily can info be found?
       - Current state: Manual search through docs/ and README
       - Search tools: grep, find (basic)
       - Navigation: No structured navigation
       - Estimate baseline: 0.40 (40% discoverability - manual search possible but inefficient)
       - Calculate: (search_success_rate + navigation_ease + tool_availability) / 3
     - V_completeness: All necessary knowledge documented?
       - Documentation coverage analysis: Count documented vs. undocumented topics
       - Core concepts: ~70% documented (parser, analyzer, query documented; workflows less so)
       - Code navigation: ~30% documented (no code map, limited architecture docs)
       - Expert knowledge: ~10% documented (no contributor guide, no expert map)
       - Estimate baseline: 0.37 (37% completeness - major gaps exist)
       - Calculate: (concept_coverage + code_coverage + expert_coverage) / 3
     - V_relevance: Right info at right time?
       - Current state: One-size-fits-all documentation
       - Role awareness: None (no contributor vs. user paths)
       - Time awareness: None (no day-1 vs. week-1 vs. month-1 paths)
       - Context awareness: None (no recommendations based on current task)
       - Estimate baseline: 0.30 (30% relevance - no personalization)
       - Calculate: (role_matching + time_matching + context_matching) / 3
     - V_freshness: Documentation up-to-date?
       - Freshness tracking: None (no last-updated dates)
       - Update frequency: Ad-hoc (updated when remembered)
       - Staleness detection: None (no automated checks)
       - Git integration: Partial (some docs updated with code changes)
       - Estimate baseline: 0.50 (50% freshness - some docs current, others stale)
       - Calculate: (tracked_freshness + update_automation + staleness_detection) / 3
     - **V_instance(s₀) = 0.3×0.40 + 0.3×0.37 + 0.2×0.30 + 0.2×0.50 = 0.39**
   - Calculate V_meta(s₀):
     - V_completeness: 0.00 (no methodology yet)
     - V_effectiveness: 0.00 (nothing to test)
     - V_reusability: 0.00 (nothing to transfer)
     - **V_meta(s₀) = 0.00**

4. **Gap Identification** (M₀.reflect):
   - **READ** meta-agents/reflect.md (gap analysis process)
   - What knowledge transfer capabilities are missing?
     - No guided learning paths (day 1, week 1, month 1)
     - No code navigation tools (architecture map, module explorer)
     - No expert identification (contributor map, code ownership)
     - No bidirectional doc-code links (code → docs, docs → code)
     - No context-aware recommendations (based on user query)
     - No knowledge gap detection (undocumented areas)
   - What onboarding aspects need coverage?
     - Discovery: How to find relevant information quickly
     - Navigation: How to explore codebase systematically
     - Learning: How to progress from beginner to expert
     - Reference: How to find answers to specific questions
     - Contribution: How to make first contribution
     - Maintenance: How to keep knowledge current
   - What methodology components are needed?
     - Knowledge discovery framework
     - Learning path design principles
     - Code navigation strategies
     - Expert identification methods
     - Freshness tracking approaches
     - Documentation organization patterns

5. **Initial Agent Applicability Assessment** (M₀.plan):
   - **READ** meta-agents/plan.md (agent selection strategies)
   - Which inherited agents are directly applicable to knowledge transfer?
     - ⭐⭐⭐ data-analyst: Analyze session history for questions, analyze file access patterns
     - ⭐⭐⭐ doc-writer: Write onboarding guides, learning paths, navigation documentation
     - ⭐⭐ coder: Build navigation tools, freshness tracking scripts, code map generators
     - ⭐ error-classifier: Classify knowledge gaps (categories: architecture, API, workflow)
     - ⭐ recovery-advisor: Recommend learning resources for knowledge gaps
     - root-cause-analyzer: Analyze why information is hard to find (less directly applicable)
   - Which inherited agents need adaptation for knowledge transfer?
     - Most agents: Apply to knowledge domain instead of error domain
     - Read agent files to understand capabilities, adapt prompts contextually
   - What new specialized agents might be needed?
     - learning-path-designer: Design day-1, week-1, month-1 onboarding paths (likely needed)
     - expert-identifier: Identify code ownership and experts via git analysis (may be needed)
     - doc-linker: Create bidirectional links between code and docs (may be needed)
     - navigation-optimizer: Design code navigation strategies (may be needed)
     - knowledge-gap-detector: Identify undocumented areas (may be needed)
     - context-recommender: Suggest relevant docs based on user query (may be needed)
   - **NOTE**: Don't create new agents yet - just identify potential needs

6. **Documentation** (M₀.execute + doc-writer):
   - **READ** meta-agents/execute.md (coordination strategies)
   - **READ** agents/doc-writer.md
   - Invoke doc-writer to create iteration-0.md:
     - M₀ state: 5 capabilities inherited from Bootstrap-003 (observe, plan, execute, reflect, evolve)
     - A₀ state: 6 agents inherited from Bootstrap-003 (3 generic + 3 specialized)
     - Agent applicability assessment (which agents useful for knowledge transfer)
     - Current knowledge state analysis (documentation inventory, question patterns, file access)
     - Current onboarding process assessment (manual, unguided, no tools)
     - Calculated V_instance(s₀) = 0.39 and V_meta(s₀) = 0.00
     - Gap analysis (missing paths, navigation, expert map, context awareness, freshness)
     - Reflection on next steps and agent reuse strategy
   - Save data artifacts:
     - data/s0-documentation-inventory.yaml (existing docs catalog)
     - data/s0-question-patterns.json (user questions from session history)
     - data/s0-file-access-patterns.yaml (frequently accessed files)
     - data/s0-workflow-patterns.json (common tool sequences)
     - data/s0-metrics.json (calculated V_instance and V_meta values)
     - data/s0-knowledge-gaps.yaml (identified gaps in onboarding)
     - data/s0-agent-applicability.yaml (inherited agents and their onboarding applicability)
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
     - Likely: Design Day-1 learning path (quick start for new contributors)
     - Or: Build code navigation map (architecture overview + module explorer)
     - Decision based on OCA framework: Start with Observe phase (analyze learning needs)
   - Which inherited agents will be most useful in Iteration 1?
     - Likely: Need new learning-path-designer agent (systematic path design)
     - Reuse: data-analyst (pattern analysis), doc-writer (guide creation)
   - Methodology extraction readiness:
     - Note patterns observed in existing documentation
     - Identify design decisions that could become methodology
     - Prepare for pattern documentation in subsequent iterations

## Constraints
- Do NOT pre-decide what agents to create next
- Do NOT assume the onboarding structure or evolution path
- Let the data (questions, access patterns, workflows) guide next steps
- Be honest about current state (minimal guidance, manual exploration)
- Calculate V(s₀) based on actual observations, not target values
- Remember two layers: concrete onboarding (instance) + methodology (meta)

## Output Format
Create iteration-0.md following this structure:
- Iteration metadata (number, date, duration)
- M₀ state documentation (5 capabilities inherited from Bootstrap-003)
- A₀ state documentation (6 agents inherited: 3 generic + 3 specialized)
- Agent applicability assessment (which agents useful for knowledge transfer domain)
- Current knowledge state analysis (docs inventory, questions, access patterns, workflows)
- Current onboarding process assessment (manual, unguided, no tools)
- Value calculation (V_instance(s₀) = 0.39, V_meta(s₀) = 0.00)
- Gap identification (missing paths, navigation, expert map, context awareness, freshness)
- Reflection on next steps and agent reuse strategy
- Data artifacts saved to data/ directory
```

---

## Iteration 1+: Subsequent Iterations (General Template)

```markdown
# Execute Iteration N: [To be determined by Meta-Agent]

## Context from Previous Iteration

Review the previous iteration file: experiments/bootstrap-011-knowledge-transfer/iteration-[N-1].md

Extract:
- Current Meta-Agent state: M_{N-1}
- Current Agent Set: A_{N-1}
- Current knowledge transfer state: V_instance(s_{N-1})
- Current methodology state: V_meta(s_{N-1})
- Problems identified
- Reflection notes on what's needed next

## Two-Layer Execution Protocol

**Layer 1 (Instance)**: Agents create concrete onboarding materials
**Layer 2 (Meta)**: Meta-Agent observes and extracts methodology

Throughout iteration:
- Agents focus on concrete tasks (design learning paths, build navigation tools, create guides)
- Meta-Agent observes onboarding work and identifies patterns for methodology

## Meta-Agent Decision Process

**BEFORE STARTING**: Read relevant Meta-Agent capability files:
- **READ** meta-agents/observe.md (for observation strategies)
- **READ** meta-agents/plan.md (for planning and decisions)
- **READ** meta-agents/execute.md (for coordination)
- **READ** meta-agents/reflect.md (for evaluation)
- **READ** meta-agents/evolve.md (for evolution assessment)

As M_{N-1}, follow the five-capability process:

### 1. OBSERVE (M.observe)
- **READ** meta-agents/observe.md for knowledge transfer observation strategies
- Review previous iteration outputs (iteration-[N-1].md)
- Examine onboarding state:
  - What learning paths have been created? (if any)
  - What navigation tools exist? (if any)
  - What expert maps have been built? (if any)
  - What documentation links exist? (if any)
- Identify gaps:
  - What learning stages still need paths? (day 1, week 1, month 1)
  - What navigation tools are missing? (code map, module explorer)
  - What knowledge areas are undocumented?
  - What methodology patterns are emerging?
- **Methodology observation**:
  - What patterns emerged in previous onboarding work?
  - What design decisions were made and why?
  - What principles can be extracted?

### 2. PLAN (M.plan)
- **READ** meta-agents/plan.md for prioritization and agent selection
- Based on observations, what is the primary goal for this iteration?
  - Examples:
    - "Design Day-1 learning path for new contributors"
    - "Build architecture navigation map"
    - "Create expert identification tool (git blame analysis)"
    - "Implement doc-code bidirectional linking"
    - "Design context-aware recommendation system"
    - "Build freshness tracking automation"
- What capabilities are needed to achieve this goal?
- **Agent Assessment**:
  - Are current agents (A_{N-1}) sufficient for this goal?
  - Can inherited agents handle onboarding? (data-analyst for patterns, doc-writer for guides)
  - Or do we need specialized `learning-path-designer` for systematic path design?
  - Do we need `expert-identifier` for git blame analysis?
  - Do we need `doc-linker` for bidirectional linking?
  - Do we need `navigation-optimizer` for code map design?
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
    - `learning-path-designer`: Design day-1, week-1, month-1 onboarding paths
      - Capabilities: Analyze learning stages, design progressive paths, create checklists
      - Why needed: Systematic learning path design requires pedagogical expertise
    - `expert-identifier`: Identify code ownership and experts
      - Capabilities: Git blame analysis, contribution patterns, code ownership mapping
      - Why needed: Automated expert identification requires git analysis expertise
    - `doc-linker`: Create bidirectional doc-code links
      - Capabilities: Parse code, extract concepts, link to docs, generate reverse links
      - Why needed: Automated linking requires code parsing + doc analysis
    - `navigation-optimizer`: Design code navigation strategies
      - Capabilities: Analyze architecture, build module maps, design exploration paths
      - Why needed: Systematic navigation design requires architectural expertise
    - `knowledge-gap-detector`: Identify undocumented areas
      - Capabilities: Code coverage analysis, doc coverage analysis, gap identification
      - Why needed: Automated gap detection requires systematic coverage analysis
    - `context-recommender`: Suggest relevant docs based on query
      - Capabilities: Semantic search, context matching, relevance ranking
      - Why needed: Context-aware recommendations require semantic understanding
  - Define agent name and specialization domain
  - Document capabilities the new agent provides
  - Explain why inherited agents are insufficient
  - **CREATE AGENT PROMPT FILE**: Write agents/{agent-name}.md
    - Include: agent role, knowledge transfer-specific capabilities, input/output format
    - Include: specific instructions for this iteration's task
    - Include: domain knowledge (learning theory, git analysis, semantic search, etc.)
  - Add to agent set: A_N = A_{N-1} ∪ {new_agent}

**Agent Invocation** (specialized or inherited):
- **READ agent prompt file** before invocation: agents/{agent-name}.md
- Invoke agent to execute concrete onboarding work:
  - Design learning paths (progressive, role-based, time-based)
  - Build navigation tools (code maps, module explorers, architecture guides)
  - Create expert maps (contributor guides, code ownership charts)
  - Implement doc-code links (bidirectional, automated)
  - Build recommendation systems (context-aware, semantic)
  - Implement freshness tracking (automated staleness detection)
- Produce iteration outputs (guides, tools, maps, links)

**Methodology Extraction** (M.evolve):
- **OBSERVE agent work patterns**:
  - How did agent organize learning paths?
  - What navigation strategies were used?
  - What expert identification methods were applied?
  - What linking strategies were employed?
- **EXTRACT patterns for methodology**:
  - Document learning path design frameworks
  - Build navigation design principles
  - Identify reusable onboarding patterns
  - Note principles that emerge
  - Add to methodology documentation

**ELSE use inherited agents:**
- **READ agent prompt file** from agents/{agent-name}.md
- Invoke appropriate agents from A_{N-1}
- Execute planned onboarding work
- Observe for methodology patterns

**CRITICAL EXECUTION PROTOCOL**:
1. ALWAYS read capability files before embodying Meta-Agent capabilities
2. ALWAYS read agent prompt file before each agent invocation
3. Do NOT cache instructions across iterations - always read from files
4. Capability files may be updated between iterations - get latest from files
5. Never assume capabilities - always verify from source files

### 4. REFLECT (M.reflect)
- **READ** meta-agents/reflect.md for evaluation process
- **Evaluate Instance Layer** (Concrete Onboarding):
  - What learning paths were created this iteration?
  - What navigation tools were built?
  - What expert maps were created?
  - What links were established?
  - Calculate new V_instance(s_N):
    - V_discoverability: (search_success_rate + navigation_ease + tool_availability) / 3
      - Count navigation tools built (code map, module explorer, etc.)
      - Measure search effectiveness (can users find info quickly?)
      - Tool availability (are discovery tools accessible?)
    - V_completeness: (concept_coverage + code_coverage + expert_coverage) / 3
      - Count documented concepts vs. total concepts
      - Code navigation coverage (architecture, modules, workflows)
      - Expert coverage (contributor map, ownership chart)
    - V_relevance: (role_matching + time_matching + context_matching) / 3
      - Role-based paths (contributor vs. user vs. maintainer)
      - Time-based paths (day 1 vs. week 1 vs. month 1)
      - Context-aware recommendations (based on current task)
    - V_freshness: (tracked_freshness + update_automation + staleness_detection) / 3
      - Freshness tracking implemented (last-updated dates)
      - Update automation (docs updated with code changes)
      - Staleness detection (automated warnings)
    - **V_instance(s_N) = 0.3×V_discover + 0.3×V_complete + 0.2×V_relevant + 0.2×V_fresh**
  - Calculate change: ΔV_instance = V_instance(s_N) - V_instance(s_{N-1})
  - Are onboarding objectives met? What gaps remain?

- **Evaluate Meta Layer** (Methodology):
  - What patterns were extracted this iteration?
  - Calculate new V_meta(s_N):
    - V_completeness: documented_patterns / total_patterns
      - Required: Learning path design, navigation strategies, expert identification, linking approaches, freshness tracking, gap detection (6 minimum)
      - Count documented vs required
    - V_effectiveness: 1 - (onboarding_time_with_methodology / onboarding_time_baseline)
      - Measure onboarding time if methodology used
      - Compare to baseline (~1-2 weeks for project familiarity)
      - Later iterations: Transfer test to different project (e.g., Claude Code plugin projects)
    - V_reusability: successful_transfers / transfer_attempts
      - Test methodology on different codebase (e.g., another Go project)
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
    details: "If no new onboarding specialization needed"

  instance_value_threshold:
    question: "Is V_instance(s_N) ≥ 0.80 (knowledge transfer system quality)?"
    V_instance(s_N): [calculated value]
    threshold_met: [Yes/No]
    components:
      V_discoverability: [value] "≥0.80 target"
      V_completeness: [value] "≥0.90 target (90% coverage)"
      V_relevance: [value] "≥0.75 target"
      V_freshness: [value] "≥0.70 target"

  meta_value_threshold:
    question: "Is V_meta(s_N) ≥ 0.80 (methodology quality)?"
    V_meta(s_N): [calculated value]
    threshold_met: [Yes/No]
    components:
      V_completeness: [value] "≥0.90 target"
      V_effectiveness: [value] "≥0.80 target (5x speedup)"
      V_reusability: [value] "≥0.70 target (70% transfer)"

  instance_objectives:
    day1_path_complete: [Yes/No] "Quick start for new contributors"
    week1_path_complete: [Yes/No] "Deep dive into architecture and modules"
    month1_path_complete: [Yes/No] "Advanced topics and contribution workflow"
    navigation_tools_built: [Yes/No] "Code map, module explorer"
    expert_map_created: [Yes/No] "Contributor guide, code ownership"
    doc_links_established: [Yes/No] "Bidirectional code ↔ docs links"
    freshness_tracking_implemented: [Yes/No] "Automated staleness detection"
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
  - O = {onboarding guides + navigation tools + methodology documentation}
  - A_N = {final agent set with specializations}
  - M_N = {final meta-agent capabilities}

**IF NOT CONVERGED:**
- Identify what's needed for next iteration
- Note: Focus on instance (onboarding) OR meta (methodology) as needed
- Continue to Iteration N+1

## Documentation Requirements

Create experiments/bootstrap-011-knowledge-transfer/iteration-N.md with:

### 1. Iteration Metadata
```yaml
iteration: N
date: YYYY-MM-DD
duration: ~X hours
status: [completed/converged]
layers:
  instance: "Onboarding work performed this iteration"
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
        specialization: knowledge_transfer_domain_area
        capabilities: [list]
        creation_reason: "Why were inherited 6 agents insufficient?"
        justification: "What gap exists that no inherited agent can fill?"
        prompt_file: "agents/{agent-name}.md"

  IF unchanged:
    status: "A_N = A_{N-1} (inherited agents sufficient for this iteration)"

  IF reused_inherited:
    reused_agents:
      - agent_name: task_performed
        adaptation_notes: "How inherited agent was adapted to knowledge transfer context"

agents_invoked_this_iteration:
  - agent_name: task_performed
    source: [inherited / newly_created]
```

### 4. Instance Work Executed (Concrete Onboarding)
- What learning paths were created?
  - Day-1 path (if created)
  - Week-1 path (if created)
  - Month-1 path (if created)
  - Role-specific paths (contributor, user, maintainer)
- What navigation tools were built?
  - Code map (architecture overview)
  - Module explorer (detailed module navigation)
  - Workflow guides (common development workflows)
- What expert maps were created?
  - Contributor guide (who to ask about what)
  - Code ownership chart (git blame analysis)
  - Expertise areas (domain knowledge mapping)
- What doc-code links were established?
  - Code → docs links (inline doc references)
  - Docs → code links (example code snippets)
  - Automated linking tools (if built)
- What freshness tracking was implemented?
  - Last-updated dates (if added)
  - Staleness detection (if automated)
  - Update automation (if implemented)
- Summary of concrete deliverables

### 5. Meta Work Executed (Methodology Extraction)
- What patterns were observed in onboarding work?
  - Learning path design patterns
  - Navigation strategy patterns
  - Expert identification patterns
  - Linking strategy patterns
- What methodology content was documented?
  - Learning path design frameworks
  - Navigation design principles
  - Expert identification methods
  - Freshness tracking approaches

**Knowledge Artifacts Created**:

Organize extracted knowledge into appropriate categories:

- **Patterns** (domain-specific): knowledge/patterns/{pattern-name}.md
  - Specific solutions to recurring problems in knowledge transfer
  - Example: "Progressive Learning Path Pattern", "Code Map Navigation Pattern"
  - Format: Problem, Context, Solution, Consequences, Examples

- **Principles** (universal): knowledge/principles/{principle-name}.md
  - Fundamental truths or rules discovered
  - Example: "Discoverability Principle", "Relevance Principle"
  - Format: Statement, Rationale, Evidence, Applications

- **Templates** (reusable): knowledge/templates/{template-name}.{md|yaml|json}
  - Concrete implementations ready for reuse
  - Example: "Learning Path Template", "Expert Map Template"
  - Format: Template file + usage documentation

- **Best Practices** (context-specific): knowledge/best-practices/{topic}.md
  - Recommended approaches for specific contexts
  - Example: "Onboarding Best Practices", "Code Navigation Best Practices"
  - Format: Context, Recommendation, Justification, Trade-offs

- **Methodology** (project-wide): docs/methodology/{methodology-name}.md
  - Comprehensive guides for reuse across projects
  - Example: "Knowledge Transfer Methodology for Software Projects"
  - Format: Complete methodology with decision frameworks

**Knowledge Index Update**:
- Update knowledge/INDEX.md with:
  - New knowledge entries
  - Links to iteration where extracted
  - Domain tags (knowledge-transfer, onboarding, navigation, etc.)
  - Validation status (proposed, validated, refined)

### 6. State Transition

**Instance Layer** (Onboarding State):
```yaml
s_{N-1} → s_N (Knowledge Transfer):
  changes:
    - learning_paths_created: [list]
    - navigation_tools_built: [list]
    - expert_maps_created: [list]
    - links_established: [count]

  metrics:
    V_discoverability: [value] (was: [previous])
    V_completeness: [value] (was: [previous])
    V_relevance: [value] (was: [previous])
    V_freshness: [value] (was: [previous])

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
  - Instance learnings (onboarding insights)
  - Meta learnings (methodology insights)
- What worked well?
- What challenges were encountered?
- What is needed next?
  - For onboarding completion
  - For methodology completion

### 8. Convergence Check
[Use the convergence criteria structure above]

### 9. Data Artifacts

**Ephemeral Data** (iteration-specific, saved to data/):
- data/iteration-N-metrics.json (V_instance, V_meta calculations)
- data/iteration-N-onboarding-state.yaml (paths created, tools built)
- data/iteration-N-methodology.yaml (extracted patterns)
- data/iteration-N-artifacts/ (guides, maps, tools)

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
   - Instance: Create concrete onboarding materials
   - Meta: Extract methodology patterns from onboarding work

2. **Be Honest**: Calculate V(s_N) based on actual state
   - Don't inflate instance values (onboarding quality)
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
   - Instance convergence: Onboarding quality + completeness
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

## Knowledge Transfer Domain-Specific Patterns

Based on OCA framework, knowledge transfer iterations may follow:

### Observe Phase (Iterations 0-2)
- Iteration 0: Baseline establishment, knowledge state analysis
- Iteration 1: Analyze learning needs from session history, design Day-1 path
- Iteration 2: Design Week-1 path, build code navigation map, pattern identification
- **Methodology**: Observe learning patterns, identify design principles

### Codify Phase (Iterations 3-4)
- Iteration 3: Design Month-1 path, build expert map, document learning path patterns
- Iteration 4: Implement doc-code linking, create navigation framework, codify patterns
- **Methodology**: Extract design frameworks, codify patterns

### Automate Phase (Iterations 5-6)
- Iteration 5: Implement freshness tracking, build context recommender, automate gap detection
- Iteration 6: Transfer test to different project, validate methodology, convergence
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
- [ ] **READ** meta-agents/observe.md (knowledge transfer observation strategies)
- [ ] Analyze onboarding state (paths created, tools built, maps generated)
- [ ] Identify gaps (learning stages missing, navigation tools needed, methodology gaps)
- [ ] Observe patterns (for methodology extraction)

**Plan Phase**:
- [ ] **READ** meta-agents/plan.md (prioritization, agent selection)
- [ ] Define iteration goal (instance layer: learning path, navigation tool, expert map)
- [ ] Assess agent sufficiency (inherited sufficient OR need specialized?)
- [ ] Plan methodology extraction (what patterns to document?)

**Execute Phase**:
- [ ] **READ** meta-agents/execute.md (coordination, pattern observation)
- [ ] **IF NEW AGENT NEEDED**:
  - [ ] **READ** meta-agents/evolve.md (agent creation criteria)
  - [ ] Create agent prompt file: agents/{agent-name}.md
  - [ ] Document specialization reason
- [ ] **READ** agent prompt file(s) before invocation: agents/{agent-name}.md
- [ ] Invoke agents to create onboarding materials
- [ ] Observe onboarding work for methodology patterns
- [ ] Extract patterns and update methodology documentation

**Reflect Phase**:
- [ ] **READ** meta-agents/reflect.md (evaluation process)
- [ ] Calculate V_instance(s_N) (onboarding quality):
  - [ ] V_discoverability (search_success + navigation_ease + tool_availability) / 3
  - [ ] V_completeness (concept_coverage + code_coverage + expert_coverage) / 3
  - [ ] V_relevance (role_matching + time_matching + context_matching) / 3
  - [ ] V_freshness (tracked_freshness + update_automation + staleness_detection) / 3
- [ ] Calculate V_meta(s_N) (methodology quality):
  - [ ] V_completeness (documented_patterns / total_patterns)
  - [ ] V_effectiveness (1 - onboarding_time_with_methodology / onboarding_time_baseline)
  - [ ] V_reusability (successful_transfers / transfer_attempts)
- [ ] Assess quality honestly (don't inflate values)
- [ ] Identify gaps (onboarding + methodology)

**Convergence Check**:
- [ ] M_N == M_{N-1}? (meta-agent stable)
- [ ] A_N == A_{N-1}? (agent set stable)
- [ ] V_instance(s_N) ≥ 0.80? (onboarding quality threshold)
- [ ] V_meta(s_N) ≥ 0.80? (methodology quality threshold)
- [ ] Instance objectives complete? (day-1/week-1/month-1 paths, navigation tools, expert map)
- [ ] Meta objectives complete? (methodology documented, transfer test successful)
- [ ] Determine: CONVERGED or NOT_CONVERGED

**Documentation**:
- [ ] Create iteration-N.md with:
  - [ ] Metadata (iteration, date, duration, status)
  - [ ] M evolution (evolved or unchanged)
  - [ ] A evolution (new agents or unchanged)
  - [ ] Instance work (paths created, tools built, maps generated)
  - [ ] Meta work (patterns extracted)
  - [ ] Knowledge artifacts created (list all new knowledge)
  - [ ] State transition (V_instance, V_meta calculated)
  - [ ] Reflection (learnings, next steps)
  - [ ] Convergence check (criteria evaluated)
- [ ] Save data artifacts to data/:
  - [ ] iteration-N-metrics.json
  - [ ] iteration-N-onboarding-state.yaml
  - [ ] iteration-N-methodology.yaml
  - [ ] iteration-N-artifacts/ (guides, maps, tools)
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
- [ ] Perform reusability validation (transfer test to different project)
- [ ] Document three-tuple: (O={onboarding guides+tools+methodology}, A_N, M_N)
- [ ] Validate methodology alignment (OCA, Bootstrapped SE, Value Space)

---

## Notes on Execution Style

**Be the Meta-Agent**: When executing iterations, embody M's perspective:
- Think through the observe-plan-execute-reflect-evolve cycle
- Read capability files to understand M's strategies
- Make explicit decisions about agent creation
- Justify why specialization is needed
- Extract methodology patterns from onboarding work
- Track both V_instance and V_meta

**Be Domain-Aware**: Knowledge transfer has specific concerns:
- Learning theory (progressive disclosure, scaffolding, etc.)
- Discovery mechanisms (search, navigation, recommendations)
- Expertise identification (git analysis, contribution patterns)
- Documentation organization (categorization, linking, indexing)
- Freshness tracking (staleness detection, update automation)
- Context awareness (role-based, time-based, task-based)

**Be Rigorous**: Calculate values honestly
- V_instance based on actual onboarding state (not aspirational)
- V_meta based on actual methodology coverage (not desired)
- Don't force convergence prematurely
- Don't skip methodology extraction to focus only on onboarding
- Let data and needs drive the process

**Be Thorough**: Document decisions and reasoning
- Save intermediate data (metrics, guides, maps, tools)
- Show your work (calculations, analysis)
- Make evolution path traceable
- **NO TOKEN LIMITS**: Complete all steps fully, never abbreviate

**Be Two-Layer Aware**: Always work on both layers
- Instance layer: Create concrete onboarding materials
- Meta layer: Extract reusable methodology
- Don't neglect either layer
- Methodology extraction is as important as onboarding execution

**Be Authentic**: This is a real experiment
- Discover onboarding patterns, don't assume them
- Create agents based on need, not predetermined plan
- Extract methodology from actual onboarding work, not theory
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
**Purpose**: Guide authentic execution of bootstrap-011-knowledge-transfer experiment with dual-layer architecture

**Key Innovation**: Dual value functions (V_instance + V_meta) enable simultaneous optimization of concrete onboarding materials and reusable knowledge transfer methodology.
