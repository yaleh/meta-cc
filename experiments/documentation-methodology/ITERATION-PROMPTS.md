# Documentation Methodology - Iteration Prompts

This document provides structured prompts for developing a systematic documentation methodology for Claude Code projects through the BAIME framework (Bootstrapped AI Methodology Engineering).

## Experiment Overview

**Domain**: Documentation Management for Claude Code projects

**Dual Objectives**:
- **Meta Layer**: Develop transferable documentation methodology (patterns, templates, validation, maintenance)
- **Instance Layer**: Produce high-quality documentation for meta-cc project (README updates, BAIME usage guide)

**Convergence Targets**:
- V_instance ≥ 0.80 (documentation quality for meta-cc deliverables)
- V_meta ≥ 0.80 (methodology quality for reuse across Claude Code projects)
- Stability: Both values remain ≥ 0.80 across consecutive iterations

---

## Architecture Overview

### Meta-Agent System (Modular Capabilities)

The meta-agent system orchestrates the documentation methodology lifecycle through specialized capabilities:

**Files**:
- `capabilities/doc-collect.md` - Data collection patterns for documentation needs
- `capabilities/doc-strategy.md` - Strategy formation for documentation approaches
- `capabilities/doc-execute.md` - Execution patterns for writing documentation
- `capabilities/doc-evaluate.md` - Evaluation rubrics and validation patterns
- `capabilities/doc-converge.md` - Convergence assessment and iteration planning

**Lifecycle Protocol**:
1. **Read all capabilities before starting iteration** to understand full methodology
2. **Read specific capability before each phase** for detailed execution guidance
3. **Update capabilities** only when retrospective evidence demonstrates gaps or improvements needed

### Agent System (Specialized Executors)

Domain-specific agents execute documentation tasks:

**Files**:
- `agents/doc-writer.md` - Technical documentation writing patterns
- `agents/doc-validator.md` - Documentation quality validation and testing
- `agents/doc-organizer.md` - Structure and organization patterns
- `agents/content-analyzer.md` - Content gap analysis and completeness checking

### System State Tracking

**Files**:
- `system-state.md` - Current methodology state, patterns discovered, value scores
- `iteration-log.md` - Chronological record of iterations and decisions
- `data/` - Evidence collected (user needs, documentation gaps, validation results)

### Knowledge Organization

**Directories**:
- `patterns/` - Documentation patterns extracted from iteration work
- `templates/` - Reusable documentation templates created
- `principles/` - Universal documentation principles discovered
- `best-practices/` - Context-specific best practices documented
- `methodology/` - Project-wide reusable knowledge for documentation systems

**Index**: `knowledge-index.md` maps patterns to iterations, domains, and validation status

---

## Value Function Design

### V_instance: Documentation Quality Score

**Purpose**: Measure quality of specific documentation deliverables for meta-cc project

**Formula**:
```
V_instance = (Accuracy × Completeness × Usability × Maintainability) / 4
```

**Components**:

1. **Accuracy** (weight: 0.25)
   - Technical correctness: Installation steps work, commands execute correctly
   - Up-to-date information: Reflects current codebase state
   - No broken links or references
   - Scoring: 0.0 (major errors) → 1.0 (completely accurate)

2. **Completeness** (weight: 0.25)
   - Coverage of user needs: Addresses identified documentation gaps
   - End-to-end workflows: Complete task flows documented
   - Edge cases covered: Error handling, troubleshooting included
   - Scoring: 0.0 (major gaps) → 1.0 (comprehensive coverage)

3. **Usability** (weight: 0.25)
   - Clarity: Clear, concise writing appropriate for target audience
   - Navigation: Easy to find information, logical structure
   - Examples: Concrete examples and code snippets provided
   - Accessibility: Appropriate for skill level (beginner to advanced)
   - Scoring: 0.0 (confusing) → 1.0 (highly usable)

4. **Maintainability** (weight: 0.25)
   - Modularity: Easy to update sections independently
   - Consistency: Follows project conventions and style
   - Automation-friendly: Validation and testing possible
   - Version tracking: Clear what changed and why
   - Scoring: 0.0 (hard to maintain) → 1.0 (easy to maintain)

**Evaluation Method**: Evidence-based assessment using:
- Manual testing of documented procedures
- Peer review against rubric criteria
- User feedback (if available)
- Automated validation (link checking, command testing)

### V_meta: Methodology Quality Score

**Purpose**: Measure quality of documentation methodology for reuse across Claude Code projects

**Formula**:
```
V_meta = (Completeness × Effectiveness × Reusability × Validation) / 4
```

**Components**:

1. **Completeness** (weight: 0.25)
   - Lifecycle coverage: All phases addressed (needs analysis → writing → validation → maintenance)
   - Pattern catalog: Documentation patterns identified and documented
   - Template library: Reusable templates created
   - Automation tools: Scripts/workflows for validation and maintenance
   - Scoring: 0.0 (major gaps) → 1.0 (comprehensive methodology)

2. **Effectiveness** (weight: 0.25)
   - Problem resolution: Methodology addresses identified documentation challenges
   - Efficiency gains: Measurable speedup vs ad-hoc approach
   - Quality improvement: Higher documentation quality achieved
   - Scoring: 0.0 (ineffective) → 1.0 (highly effective)

3. **Reusability** (weight: 0.25)
   - Generalizability: Patterns applicable beyond meta-cc project
   - Adaptation effort: Low effort to apply to new Claude Code projects
   - Domain independence: Core principles work across project types
   - Clear guidance: Methodology provides actionable steps for new contexts
   - Scoring: 0.0 (project-specific) → 1.0 (highly reusable)

4. **Validation** (weight: 0.25)
   - Empirical grounding: Patterns validated through actual documentation work
   - Metrics defined: Clear quality and progress metrics
   - Retrospective testing: Methodology assessed against historical documentation needs
   - Quality gates: Automated validation mechanisms established
   - Scoring: 0.0 (unvalidated) → 1.0 (empirically validated)

**Evaluation Method**: Rubric-based assessment using:
- Pattern extraction validation: Confirm patterns emerge from evidence
- Cross-project applicability testing: Apply to different documentation scenarios
- Methodology completeness checklist: Verify all lifecycle phases covered
- Automation effectiveness: Test validation tools and workflows

---

## Baseline Iteration (Iteration 0)

### Context

**Starting Point**:
- Meta-cc project has existing documentation structure (docs/, README.md)
- No systematic documentation methodology in place
- Specific gaps identified: Plugin installation, BAIME usage guide
- Ad-hoc documentation creation process

**Expectation**:
Low baseline values expected and acceptable. This iteration establishes measurement baselines and initial system state. V_instance and V_meta may be 0.20-0.40 as we're starting from scratch on methodology.

### System Setup

**Objective**: Create modular architecture for documentation methodology development

**Steps**:

1. **Create capability files** (meta-agent system):
   ```bash
   # Create capabilities/ directory structure
   mkdir -p capabilities

   # Initialize lifecycle capabilities
   touch capabilities/doc-collect.md      # Data collection patterns
   touch capabilities/doc-strategy.md     # Strategy formation
   touch capabilities/doc-execute.md      # Execution patterns
   touch capabilities/doc-evaluate.md     # Evaluation rubrics
   touch capabilities/doc-converge.md     # Convergence assessment
   ```

2. **Create agent files** (specialized executors):
   ```bash
   # Create agents/ directory structure
   mkdir -p agents

   # Initialize domain agents
   touch agents/doc-writer.md          # Writing patterns
   touch agents/doc-validator.md       # Validation patterns
   touch agents/doc-organizer.md       # Organization patterns
   touch agents/content-analyzer.md    # Gap analysis patterns
   ```

3. **Create system state files**:
   ```bash
   # Initialize tracking files
   touch system-state.md          # Current methodology state
   touch iteration-log.md         # Chronological log

   # Create data directory
   mkdir -p data
   ```

4. **Create knowledge organization structure**:
   ```bash
   # Create knowledge directories
   mkdir -p patterns templates principles best-practices methodology

   # Initialize index
   touch knowledge-index.md
   ```

**Initial Content**: Each capability and agent file should contain:
- Purpose statement
- Placeholder for patterns/guidance
- Note: "Populated during iteration cycles based on empirical evidence"

### Objectives

**Iteration 0 Goals**:

1. **Collect baseline data** on documentation needs and current state
2. **Establish baseline value scores** (V_instance_0 and V_meta_0)
3. **Identify initial problems** with current documentation approach
4. **Document initial system state**
5. **Create first documentation deliverable** (even if low quality) to establish measurement baseline

### Execution Steps

#### Phase 1: Data Collection

**Read**: `capabilities/doc-collect.md` (will be empty initially - use general data collection principles)

**Tasks**:

1. **Analyze current documentation state**:
   - Survey existing meta-cc documentation (README.md, docs/)
   - Identify documentation structure and conventions
   - Note what works well and what's missing
   - Save findings to `data/current-state-analysis.md`

2. **Identify documentation gaps**:
   - Review meta-cc plugin: What installation instructions are needed?
   - Review BAIME framework: What usage guidance is missing?
   - Analyze user journey: What questions would new users have?
   - Save gap analysis to `data/documentation-gaps.md`

3. **Gather user needs evidence**:
   - Review existing issues/questions (if available)
   - Analyze similar projects' documentation approaches
   - Identify target audiences (new users, developers, methodology users)
   - Save to `data/user-needs.md`

4. **Assess documentation tools available**:
   - Markdown capabilities
   - Link validation tools
   - Command testing approaches
   - Save to `data/tool-inventory.md`

**Output**: Evidence files in `data/` directory documenting current state and needs

#### Phase 2: Strategy Formation

**Read**: `capabilities/doc-strategy.md` (will be empty initially - use general strategy principles)

**Tasks**:

1. **Prioritize documentation deliverables**:
   - Based on gap analysis, prioritize: README.md updates vs BAIME guide
   - Decide which to tackle first in Iteration 0
   - Document rationale in `iteration-log.md`

2. **Define initial documentation approach**:
   - Choose structure for first deliverable
   - Identify examples/templates to reference
   - Plan sections and content flow
   - Document in `iteration-log.md`

3. **Establish quality criteria**:
   - Define what "good enough for Iteration 0" means
   - Set realistic targets given baseline context
   - Document criteria in `iteration-log.md`

**Output**: Strategy documented in `iteration-log.md`

#### Phase 3: Execution

**Read**: `capabilities/doc-execute.md` (will be empty initially - use general execution principles)

**Tasks**:

1. **Write first documentation deliverable**:
   - Choose one deliverable (e.g., README.md plugin installation section)
   - Write initial version following planned structure
   - Include examples and code snippets
   - Commit to repository

2. **Document writing process**:
   - Note challenges encountered
   - Record decisions made
   - Identify patterns that emerged
   - Save observations to `data/writing-observations.md`

3. **Perform basic validation**:
   - Test installation instructions manually
   - Check links
   - Review for clarity
   - Document validation results in `data/validation-results.md`

**Output**:
- First documentation deliverable committed
- Process observations documented in `data/`

#### Phase 4: Evaluation

**Read**: `capabilities/doc-evaluate.md` (will be empty initially - use value function rubrics)

**Tasks**:

1. **Calculate V_instance_0**:
   - **Accuracy**: Test documented procedures, check technical correctness
     - Evidence: Do installation steps work? Are commands correct?
     - Score: _____ (0.0-1.0)
     - Justification: _____

   - **Completeness**: Assess coverage of user needs
     - Evidence: What gaps remain? Are workflows complete?
     - Score: _____ (0.0-1.0)
     - Justification: _____

   - **Usability**: Evaluate clarity and navigation
     - Evidence: Is it easy to find information? Are examples clear?
     - Score: _____ (0.0-1.0)
     - Justification: _____

   - **Maintainability**: Check modularity and consistency
     - Evidence: Will this be easy to update? Follows conventions?
     - Score: _____ (0.0-1.0)
     - Justification: _____

   - **V_instance_0 = (Accuracy + Completeness + Usability + Maintainability) / 4**
   - **Final Score**: _____

2. **Calculate V_meta_0**:
   - **Completeness**: Assess methodology coverage
     - Evidence: What lifecycle phases have initial patterns?
     - Score: _____ (0.0-1.0)
     - Justification: _____

   - **Effectiveness**: Evaluate problem resolution
     - Evidence: Did approach address documentation challenges?
     - Score: _____ (0.0-1.0)
     - Justification: _____

   - **Reusability**: Check generalizability
     - Evidence: Could this approach work for other projects?
     - Score: _____ (0.0-1.0)
     - Justification: _____

   - **Validation**: Assess empirical grounding
     - Evidence: Are patterns validated through actual work?
     - Score: _____ (0.0-1.0)
     - Justification: _____

   - **V_meta_0 = (Completeness + Effectiveness + Reusability + Validation) / 4**
   - **Final Score**: _____

3. **Document evaluation**:
   - Save detailed evaluation to `system-state.md`
   - Include evidence for each score
   - Note areas for improvement

**Output**: Baseline value scores documented in `system-state.md`

#### Phase 5: Problem Identification

**Tasks**:

1. **Identify documentation challenges encountered**:
   - What was difficult about writing documentation?
   - What took longer than expected?
   - What quality issues emerged?
   - Document in `system-state.md`

2. **Identify methodology gaps**:
   - What guidance would have been helpful?
   - What patterns could be extracted?
   - What tools/automation would help?
   - Document in `system-state.md`

3. **Prioritize problems for next iteration**:
   - Rank by impact on V_instance and V_meta
   - Choose top 2-3 problems to address in Iteration 1
   - Document in `system-state.md`

**Output**: Problem inventory in `system-state.md`

#### Phase 6: Initial System State Documentation

**Tasks**:

1. **Update `system-state.md`**:
   ```markdown
   # System State After Iteration 0

   ## Value Scores
   - V_instance_0: [score] ([date])
   - V_meta_0: [score] ([date])

   ## Methodology Components
   - Capabilities: [list with status]
   - Agents: [list with status]
   - Patterns: [count and summary]
   - Templates: [count and summary]

   ## Problems Identified
   1. [Problem 1]
   2. [Problem 2]
   3. [Problem 3]

   ## Priorities for Next Iteration
   - [Priority 1]
   - [Priority 2]
   ```

2. **Update `iteration-log.md`**:
   ```markdown
   # Iteration Log

   ## Iteration 0 ([date])

   ### Objectives
   - Establish baseline measurements
   - Create first documentation deliverable
   - Identify initial problems

   ### Execution Summary
   - Deliverable created: [name]
   - Data collected: [summary]
   - Patterns observed: [summary]

   ### Results
   - V_instance_0: [score]
   - V_meta_0: [score]
   - Key learnings: [summary]

   ### Next Iteration Focus
   - [Focus area 1]
   - [Focus area 2]
   ```

3. **Create initial knowledge index**:
   ```markdown
   # Knowledge Index

   ## Patterns
   (None yet - to be extracted in subsequent iterations)

   ## Templates
   (None yet - to be created based on patterns)

   ## Principles
   (None yet - to be discovered through iterations)
   ```

**Output**: Complete initial system state documentation

### Honest Assessment Reminder

**Key Principles for Iteration 0**:

1. **Low scores are expected and acceptable**
   - This is baseline measurement, not final product
   - V_instance_0 and V_meta_0 may be 0.20-0.40
   - Purpose: Establish starting point and improvement trajectory

2. **Evidence-based scoring**
   - Ground each score component in concrete evidence
   - Document what works and what doesn't
   - Avoid aspirational or inflated scores

3. **Focus on learning**
   - Primary goal: Understand the documentation problem space
   - Secondary goal: Create measurement baseline
   - Tertiary goal: Produce initial deliverable (quality less important)

4. **No predetermined evolution**
   - Don't plan ahead what patterns "should" emerge
   - Let patterns reveal themselves through actual work
   - Methodology evolves based on problems encountered, not theory

---

## Subsequent Iterations (Iterations 1+)

### Context Extraction

**Before Starting Each Iteration**:

1. **Read previous iteration state**:
   - Review `system-state.md` for current value scores and methodology status
   - Review `iteration-log.md` for recent history and decisions
   - Understand what problems were identified
   - Note what priorities were set for this iteration

2. **Read all capabilities** (full lifecycle overview):
   - `capabilities/doc-collect.md`
   - `capabilities/doc-strategy.md`
   - `capabilities/doc-execute.md`
   - `capabilities/doc-evaluate.md`
   - `capabilities/doc-converge.md`

   **Purpose**: Understand complete methodology before starting work

3. **Understand current baseline**:
   - V_instance from previous iteration: _____
   - V_meta from previous iteration: _____
   - Target this iteration: Improve both scores

### Lifecycle Protocol

**Capability Reading Protocol**:
- **All capabilities before iteration start**: Full methodology overview
- **Specific capability before each phase**: Detailed execution guidance
- **Capability updates**: Only when retrospective evidence demonstrates necessity

### Iteration Cycle

Each iteration follows the Observe-Codify-Automate (OCA) lifecycle:

#### Phase 1: Data Collection (Observe)

**Read**: `capabilities/doc-collect.md` before starting this phase

**Objectives**:
- Gather evidence about documentation needs and current state
- Collect data on problems identified in previous iteration
- Build empirical foundation for strategy decisions

**Tasks**:

1. **Review problem priorities from previous iteration**:
   - What specific problems are we addressing?
   - What evidence do we need to understand these problems?
   - Document focus in `iteration-log.md`

2. **Collect targeted data**:
   - Based on priorities, gather relevant evidence
   - Examples:
     - If problem is "incomplete workflows", analyze existing docs for workflow gaps
     - If problem is "unclear examples", review example quality in current docs
     - If problem is "hard to maintain", assess how last doc update went
   - Save evidence to `data/iteration-N-evidence.md`

3. **Analyze patterns in collected data**:
   - What recurring issues appear?
   - What successful approaches are evident?
   - What gaps remain?
   - Document analysis in `data/iteration-N-analysis.md`

**Output**: Evidence files in `data/` documenting findings

**Anti-patterns to Avoid**:
- ❌ Collecting data without clear connection to identified problems
- ❌ Relying on assumptions instead of empirical evidence
- ❌ Skipping data collection because "we already know the solution"

#### Phase 2: Strategy Formation (Codify)

**Read**: `capabilities/doc-strategy.md` before starting this phase

**Objectives**:
- Define approach for addressing identified problems
- Make evidence-based decisions about documentation work
- Plan concrete actions for this iteration

**Tasks**:

1. **Synthesize data into insights**:
   - What does the evidence tell us?
   - What patterns are emerging?
   - What strategies would address identified problems?
   - Document in `iteration-log.md`

2. **Define iteration strategy**:
   - What specific documentation work will we do?
   - What methodology improvements will we make?
   - What patterns will we extract and codify?
   - Document strategy in `iteration-log.md`

3. **Set concrete targets**:
   - What V_instance improvements are we targeting?
   - What V_meta improvements are we targeting?
   - What specific deliverables will we produce?
   - Document targets in `iteration-log.md`

**Output**: Strategy documented in `iteration-log.md`

**Strategy Formation Principles**:
- Evidence-driven: Strategy emerges from data, not theory
- Focused: Address top 2-3 problems, not everything
- Measurable: Clear success criteria defined
- Realistic: Achievable within iteration scope

#### Phase 3: Execution (Automate/Execute)

**Read**: `capabilities/doc-execute.md` before starting this phase

**Objectives**:
- Execute planned documentation work
- Apply and validate patterns
- Create/update deliverables

**Tasks**:

1. **Execute documentation work**:
   - Write/update documentation based on strategy
   - Apply patterns from methodology
   - Follow quality criteria
   - Commit deliverables to repository

2. **Document execution process**:
   - Record challenges encountered
   - Note successful approaches
   - Capture new patterns observed
   - Save to `data/iteration-N-execution.md`

3. **Validate deliverables**:
   - Test documented procedures
   - Check links and references
   - Review for clarity and completeness
   - Document validation results in `data/iteration-N-validation.md`

**Output**:
- Updated documentation deliverables
- Process observations in `data/`

**Execution Best Practices**:
- Follow methodology patterns consistently
- Test as you write (installation steps, commands, examples)
- Capture observations in real-time
- Don't deviate from planned strategy without documented reason

#### Phase 4: Evaluation

**Read**: `capabilities/doc-evaluate.md` before starting this phase

**Objectives**:
- Calculate V_instance and V_meta for this iteration
- Provide evidence-based assessment of quality
- Identify gaps and improvement areas

**Tasks**:

1. **Calculate V_instance_N** (use rubric from value function section):

   **Accuracy** (0.0-1.0):
   - Evidence: [Test results, technical review findings]
   - Score: _____
   - Justification: _____

   **Completeness** (0.0-1.0):
   - Evidence: [Gap analysis, coverage assessment]
   - Score: _____
   - Justification: _____

   **Usability** (0.0-1.0):
   - Evidence: [Clarity review, navigation testing, example quality]
   - Score: _____
   - Justification: _____

   **Maintainability** (0.0-1.0):
   - Evidence: [Update effort, consistency check, automation support]
   - Score: _____
   - Justification: _____

   **V_instance_N**: _____ (average of four components)
   **Improvement from previous**: _____ (ΔV_instance)

2. **Calculate V_meta_N** (use rubric from value function section):

   **Completeness** (0.0-1.0):
   - Evidence: [Lifecycle coverage, pattern catalog size, templates created]
   - Score: _____
   - Justification: _____

   **Effectiveness** (0.0-1.0):
   - Evidence: [Problem resolution, efficiency gains, quality improvements]
   - Score: _____
   - Justification: _____

   **Reusability** (0.0-1.0):
   - Evidence: [Generalizability, adaptation effort, domain independence]
   - Score: _____
   - Justification: _____

   **Validation** (0.0-1.0):
   - Evidence: [Empirical grounding, metrics defined, retrospective testing]
   - Score: _____
   - Justification: _____

   **V_meta_N**: _____ (average of four components)
   **Improvement from previous**: _____ (ΔV_meta)

3. **Document evaluation**:
   - Update `system-state.md` with new scores
   - Include detailed evidence for each component
   - Highlight improvements and remaining gaps
   - Save to `system-state.md`

**Output**: Updated value scores in `system-state.md`

**Honest Assessment Principles**:

1. **Seek disconfirming evidence**:
   - Actively look for what doesn't work
   - Test edge cases and error scenarios
   - Don't just confirm what you hope is true

2. **Enumerate gaps explicitly**:
   - List specific missing pieces
   - Quantify coverage percentages
   - Identify quality issues honestly

3. **Ground scores in concrete evidence**:
   - Reference specific test results
   - Cite examples of issues found
   - Use measurable criteria where possible

4. **Challenge high scores**:
   - If scoring > 0.90, provide strong justification
   - High scores require exceptional quality
   - Most iterations show incremental improvement (0.05-0.15)

5. **Avoid anti-patterns**:
   - ❌ "Looks good to me" without testing
   - ❌ Scoring based on effort rather than results
   - ❌ Inflating scores because of time pressure
   - ❌ Comparing to "could be worse" instead of objective criteria

#### Phase 5: Convergence Check

**Read**: `capabilities/doc-converge.md` before starting this phase

**Objectives**:
- Assess progress toward convergence
- Decide on next iteration focus or declare convergence
- Update methodology components based on learnings

**Tasks**:

1. **Check convergence criteria**:

   **Instance Layer Convergence**:
   - V_instance_N ≥ 0.80? _____
   - Stable across last 2 iterations? _____
   - All critical documentation gaps addressed? _____
   - Status: [Converged / Not Converged / Approaching]

   **Meta Layer Convergence**:
   - V_meta_N ≥ 0.80? _____
   - Stable across last 2 iterations? _____
   - Methodology covers full documentation lifecycle? _____
   - Patterns validated and reusable? _____
   - Status: [Converged / Not Converged / Approaching]

   **Overall Convergence**:
   - Both layers ≥ 0.80 and stable? _____
   - Decision: [Continue / Converged]

2. **Identify remaining problems** (if not converged):
   - What gaps prevent reaching 0.80?
   - What quality issues remain?
   - What methodology improvements needed?
   - Document in `system-state.md`

3. **Prioritize next iteration** (if not converged):
   - Rank problems by impact on value scores
   - Choose top 2-3 for next iteration
   - Document priorities in `system-state.md`

4. **Extract and codify patterns**:
   - Review `data/iteration-N-*.md` for recurring patterns
   - Extract validated patterns to `patterns/`
   - Update capabilities with new guidance
   - Create/update templates as needed
   - **Principle**: Only codify patterns with empirical evidence from actual work

5. **Update methodology artifacts**:
   - Update capability files with new patterns/guidance
   - Create/update templates in `templates/`
   - Document principles in `principles/`
   - Update `knowledge-index.md`
   - **Trigger for updates**: Retrospective evidence demonstrates gap or improvement opportunity
   - **Anti-trigger**: Pattern matching or anticipatory design

**Output**:
- Convergence decision in `system-state.md`
- Updated methodology components (if evidence supports changes)
- Priorities for next iteration (if not converged)

**Pattern Extraction Principles**:

1. **Evidence-based extraction**:
   - Pattern must appear in actual iteration work
   - Must have solved a real problem
   - Must be validated through application
   - Document evidence in pattern file

2. **Avoid premature generalization**:
   - Don't extract pattern from single occurrence
   - Wait for pattern to recur 2-3 times
   - Validate applicability across contexts

3. **Clear documentation**:
   - Pattern name and purpose
   - Context where applicable
   - Evidence of effectiveness
   - Examples from iterations

**Methodology Evolution Guidance**:

**When to evolve methodology** (add patterns, update capabilities):
- ✅ Retrospective evidence shows gap: "We struggled with X and wish we had guidance"
- ✅ Pattern recurred 2-3 times across iterations
- ✅ Attempted alternatives and this approach proved superior
- ✅ Improvement quantifiable in value scores

**When NOT to evolve methodology**:
- ❌ Pattern matching: "Other projects do it this way"
- ❌ Anticipatory design: "We might need this later"
- ❌ Theoretical completeness: "Methodology should cover this"
- ❌ Single occurrence: "This worked once so let's generalize"

**Evolution validation**:
- Document necessity: Why is this change needed?
- Document evidence: What data supports this change?
- Measure improvement: Did value scores improve after change?

#### Phase 6: Iteration Summary

**Tasks**:

1. **Update `iteration-log.md`**:
   ```markdown
   ## Iteration N ([date])

   ### Objectives
   - [Objective 1]
   - [Objective 2]

   ### Strategy
   - [Strategy summary]

   ### Execution Summary
   - Deliverables created/updated: [list]
   - Patterns applied: [list]
   - Challenges encountered: [summary]

   ### Results
   - V_instance_N: [score] (Δ [change])
   - V_meta_N: [score] (Δ [change])
   - Key learnings: [summary]

   ### Convergence Status
   - Instance layer: [status]
   - Meta layer: [status]
   - Overall: [Continue / Converged]

   ### Next Iteration Focus
   - [Focus area 1]
   - [Focus area 2]
   ```

2. **Update `system-state.md`** with current state after iteration

3. **Commit all changes**:
   - Documentation deliverables
   - System state files
   - Methodology updates
   - Data/evidence files

**Output**: Complete iteration record

---

## Knowledge Organization

### Pattern Extraction

**Process**:

1. **During iteration**: Observe and document patterns in `data/iteration-N-*.md`
2. **During convergence check**: Review observations for recurring patterns
3. **Extract to patterns/**: When pattern validated and reusable

**Pattern File Template**:
```markdown
# [Pattern Name]

## Purpose
[What this pattern accomplishes]

## Context
[When this pattern applies]

## Problem
[What problem this pattern solves]

## Solution
[How to apply this pattern]

## Evidence
[Data from iterations showing effectiveness]
- Iteration [N]: [Example]
- Iteration [M]: [Example]

## Examples
[Concrete examples from meta-cc documentation work]

## Related Patterns
[Links to related patterns]

## Validation Status
- First observed: Iteration [N]
- Validated across: [contexts/iterations]
- Effectiveness: [metric/evidence]
```

### Template Creation

**Trigger**: Pattern recurs 3+ times and has clear structure

**Template File Structure**:
```markdown
# [Template Name]

## Purpose
[What this template is for]

## When to Use
[Context and applicability]

## Template

[Actual template content with placeholders]

## Instructions
[How to fill out template]

## Examples
[Filled examples from meta-cc work]

## Source Pattern
[Link to pattern this emerged from]
```

### Principle Documentation

**Trigger**: Universal insight discovered through iterations

**Principle File Structure**:
```markdown
# [Principle Name]

## Statement
[Clear principle statement]

## Rationale
[Why this principle matters]

## Evidence
[Data supporting this principle]

## Application
[How to apply this principle]

## Counter-examples
[When this principle doesn't apply]

## Related Principles
[Links to related principles]
```

### Knowledge Index

**Purpose**: Map knowledge artifacts to iterations, domains, validation status

**Structure**:
```markdown
# Knowledge Index

## Patterns
- [Pattern Name] - [File] - Extracted: Iter [N] - Validated: [status]

## Templates
- [Template Name] - [File] - Created: Iter [N] - Usage: [count]

## Principles
- [Principle Name] - [File] - Discovered: Iter [N] - Domain: [scope]

## Best Practices
- [Practice Name] - [File] - Context: [domain] - Validated: [status]

## Cross-References
- [Pattern A] relates to [Pattern B]: [relationship]
- [Template X] implements [Pattern Y]
- [Principle P] supports [Pattern Q]

## Validation Status Legend
- ✅ Validated: Applied successfully 3+ times
- 🔄 In Progress: Applied 1-2 times, monitoring
- 📋 Proposed: Extracted but not yet validated
```

---

## Results Analysis Template

**Context**: Use this template after convergence is achieved (both V_instance and V_meta ≥ 0.80 and stable)

### 1. Convergence Validation

**Value Score Trajectory**:
```
Iteration | V_instance | V_meta | Notes
----------|------------|--------|-------
0         | [score]    | [score]| Baseline
1         | [score]    | [score]| [key change]
2         | [score]    | [score]| [key change]
...       | ...        | ...    | ...
N         | [score]    | [score]| Final
```

**Convergence Criteria Met**:
- ✅/❌ V_instance ≥ 0.80 for 2+ consecutive iterations
- ✅/❌ V_meta ≥ 0.80 for 2+ consecutive iterations
- ✅/❌ All critical documentation gaps addressed
- ✅/❌ Methodology covers full documentation lifecycle
- ✅/❌ Patterns validated and reusable

**Final Scores**:
- V_instance_final: _____ (iteration ___)
- V_meta_final: _____ (iteration ___)

### 2. Instance Layer Results

**Documentation Deliverables Completed**:
1. [Deliverable 1] - [Status] - [Location]
2. [Deliverable 2] - [Status] - [Location]
3. [Deliverable 3] - [Status] - [Location]

**Quality Assessment** (final V_instance breakdown):
- Accuracy: _____ - [Evidence]
- Completeness: _____ - [Evidence]
- Usability: _____ - [Evidence]
- Maintainability: _____ - [Evidence]

**Key Improvements**:
- [Improvement 1]: [Before] → [After]
- [Improvement 2]: [Before] → [After]

**User Impact**:
- [How documentation helps new users]
- [How documentation supports developers]
- [How documentation enables methodology adoption]

### 3. Meta Layer Results

**Methodology Components Developed**:

**Capabilities**:
- `doc-collect.md`: [Description] - [Status]
- `doc-strategy.md`: [Description] - [Status]
- `doc-execute.md`: [Description] - [Status]
- `doc-evaluate.md`: [Description] - [Status]
- `doc-converge.md`: [Description] - [Status]

**Agents**:
- `doc-writer.md`: [Description] - [Status]
- `doc-validator.md`: [Description] - [Status]
- `doc-organizer.md`: [Description] - [Status]
- `content-analyzer.md`: [Description] - [Status]

**Patterns Extracted**: _____ total
1. [Pattern name] - [Brief description] - [Validation status]
2. [Pattern name] - [Brief description] - [Validation status]

**Templates Created**: _____ total
1. [Template name] - [Purpose] - [Usage count]
2. [Template name] - [Purpose] - [Usage count]

**Principles Discovered**: _____ total
1. [Principle name] - [Brief statement]
2. [Principle name] - [Brief statement]

**Automation Tools Developed**:
- [Tool 1]: [Description] - [Impact]
- [Tool 2]: [Description] - [Impact]

### 4. Effectiveness Analysis

**Efficiency Gains**:
- Time to create documentation: [baseline] → [with methodology]
- Speedup factor: _____x
- Effort reduction: _____%

**Quality Improvements**:
- Documentation quality: [baseline V_instance] → [final V_instance]
- Improvement: +_____ (____%)

**Problem Resolution**:
- Initial problems identified: [count]
- Problems resolved: [count]
- Resolution rate: _____%

**Key Problems Solved**:
1. [Problem]: [How methodology solved it]
2. [Problem]: [How methodology solved it]

### 5. Reusability Assessment

**Generalizability**:
- Domain-specific components: [count/percentage]
- Project-agnostic components: [count/percentage]
- Reusability score: _____ (V_meta Reusability component)

**Adaptation Effort Estimate**:
- To apply to new Claude Code project: [effort estimate]
- To adapt to different domain: [effort estimate]

**Transferability Analysis**:
- Universal patterns (work anywhere): [count]
- Claude Code specific patterns: [count]
- Meta-cc specific patterns: [count]

**Reuse Potential**:
- High (ready to use as-is): [pattern list]
- Medium (minor adaptation needed): [pattern list]
- Low (significant customization required): [pattern list]

### 6. Methodology Validation

**Empirical Grounding**:
- Patterns validated through actual work: [count/percentage]
- Retrospective testing performed: ✅/❌
- Metrics and quality gates defined: ✅/❌

**Lifecycle Coverage**:
- Data collection: ✅/❌ - [Coverage assessment]
- Strategy formation: ✅/❌ - [Coverage assessment]
- Execution: ✅/❌ - [Coverage assessment]
- Evaluation: ✅/❌ - [Coverage assessment]
- Convergence: ✅/❌ - [Coverage assessment]

**Quality Gates Established**:
1. [Gate name]: [Description] - [Automated: Yes/No]
2. [Gate name]: [Description] - [Automated: Yes/No]

**Validation Mechanisms**:
- [Mechanism 1]: [Description]
- [Mechanism 2]: [Description]

### 7. Key Learnings

**What Worked Well**:
1. [Learning 1]
2. [Learning 2]
3. [Learning 3]

**What Was Challenging**:
1. [Challenge 1] - [How addressed]
2. [Challenge 2] - [How addressed]

**Surprises and Insights**:
1. [Insight 1]
2. [Insight 2]

**Methodology Evolution**:
- Number of methodology updates: _____
- Most impactful change: [Description]
- Evolution trigger: [What drove changes]

### 8. Documentation Domain Insights

**Documentation Patterns Discovered**:
1. [Pattern]: [Description and impact]
2. [Pattern]: [Description and impact]

**Claude Code Specific Findings**:
- [Finding 1]: [Implication for Claude Code documentation]
- [Finding 2]: [Implication for Claude Code documentation]

**Best Practices Identified**:
1. [Practice]: [Context and rationale]
2. [Practice]: [Context and rationale]

### 9. Knowledge Catalog

**Organized Knowledge**:
- Patterns directory: [file count] patterns
- Templates directory: [file count] templates
- Principles directory: [file count] principles
- Best practices directory: [file count] practices
- Methodology directory: [file count] reusable artifacts

**Knowledge Index**:
- Total cross-references: _____
- Validation coverage: _____%

**Most Valuable Artifacts**:
1. [Artifact]: [Why valuable]
2. [Artifact]: [Why valuable]

### 10. Future Recommendations

**For meta-cc Project**:
- [Recommendation 1]
- [Recommendation 2]

**For Documentation Methodology**:
- [Recommendation 1]
- [Recommendation 2]

**For BAIME Framework**:
- [Observation 1]
- [Observation 2]

---

## Execution Guidance

### Meta-Agent Perspective

**Role**: You are the meta-agent orchestrating the documentation methodology development for the meta-cc project.

**Responsibilities**:
1. **Dual-layer focus**: Simultaneously develop methodology AND produce documentation deliverables
2. **Evidence-based evolution**: Let patterns emerge from actual work, don't pre-plan
3. **Honest evaluation**: Rigorous, unbiased assessment using value function rubrics
4. **Systematic knowledge capture**: Extract patterns, create templates, document principles

**Approach**:
- Embody the methodology you're developing
- Make decisions based on data collected
- Validate patterns through application
- Measure progress against dual value functions

### Rigor and Thoroughness

**Evaluation Rigor**:
- Always calculate both V_instance and V_meta
- Provide concrete evidence for each component score
- Seek disconfirming evidence actively
- Challenge high scores with extra scrutiny

**No Token Limits**:
- Complete thorough analysis without cutting corners
- Document all evidence collected
- Provide comprehensive justifications for scores
- Don't skip steps to save tokens

**Completeness**:
- Cover all lifecycle phases each iteration
- Update all tracking files (system-state.md, iteration-log.md)
- Extract patterns when evidence supports it
- Maintain knowledge index consistently

### Authenticity and Discovery

**Discover, Don't Assume**:
- Don't import documentation patterns from other projects without validation
- Don't assume what "good documentation" looks like without evidence
- Let patterns emerge from actual documentation work on meta-cc
- Validate assumptions through testing and user perspective

**Honest Assessment Protocol**:

**1. Seek Disconfirming Evidence**:
- Actively test what could go wrong
- Try documented procedures in fresh environment
- Look for gaps and unclear instructions
- Challenge your own assumptions

**2. Enumerate Gaps Explicitly**:
- List specific missing pieces: "Missing: error handling for X"
- Quantify coverage: "Covers 7/10 identified use cases"
- Identify quality issues: "Example unclear because Y"

**3. Ground Scores in Concrete Evidence**:
- Reference specific tests: "Installation steps tested on clean VM"
- Cite examples: "Link checking found 3 broken links"
- Use measurable criteria: "12/15 workflows documented = 80% completeness"

**4. Challenge High Scores**:
- If scoring > 0.90: Requires exceptional quality with strong evidence
- If scoring > 0.85: Must demonstrate near-complete coverage
- If scoring > 0.80: Should show comprehensive quality across all dimensions
- Most iterations show incremental improvement (0.05-0.15 per iteration)

**5. Avoid Anti-patterns**:
- ❌ "Looks good to me" → ✅ "Tested X, Y, Z; found issues A, B"
- ❌ Scoring based on effort → ✅ Scoring based on measurable results
- ❌ Inflating scores due to time pressure → ✅ Honest assessment regardless of iteration count
- ❌ "Better than nothing" comparisons → ✅ Objective criteria from rubrics
- ❌ Assuming completeness → ✅ Systematically checking coverage

### Independent Dual-Layer Assessment

**Instance Layer Evaluation**:
- Focus: Quality of meta-cc documentation deliverables
- Method: Test procedures, check accuracy, assess usability
- Independence: Don't let methodology quality influence documentation quality scores

**Meta Layer Evaluation**:
- Focus: Quality of documentation methodology for reuse
- Method: Assess completeness, effectiveness, reusability, validation
- Independence: Don't let documentation deliverable quality influence methodology scores

**Both Must Meet Threshold**:
- Convergence requires BOTH V_instance ≥ 0.80 AND V_meta ≥ 0.80
- Can't declare convergence if one layer succeeds but other doesn't
- Stability required: Both scores ≥ 0.80 for 2+ consecutive iterations

### Methodology Evolution Principles

**Evidence-Driven Evolution**:
- Update capabilities when retrospective evidence shows gaps
- Extract patterns when they recur 2-3 times
- Create templates when structure is validated
- Document principles when insights are universal

**Avoid Premature Generalization**:
- Don't extract patterns from single occurrence
- Don't create templates without validation
- Don't document principles without evidence
- Don't evolve methodology based on theory

**Validation Requirements**:
- Pattern must have solved real problem in iteration work
- Must be validated across multiple contexts
- Effectiveness must be measurable
- Reusability must be demonstrated

---

## Quick Reference

### Iteration Checklist

**Every Iteration**:
- [ ] Read `system-state.md` (previous results)
- [ ] Read all capabilities (methodology overview)
- [ ] **Phase 1: Data Collection** - Read `doc-collect.md`, gather evidence
- [ ] **Phase 2: Strategy** - Read `doc-strategy.md`, plan approach
- [ ] **Phase 3: Execution** - Read `doc-execute.md`, create deliverables
- [ ] **Phase 4: Evaluation** - Read `doc-evaluate.md`, calculate V_instance and V_meta
- [ ] **Phase 5: Convergence** - Read `doc-converge.md`, assess progress
- [ ] **Phase 6: Summary** - Update `iteration-log.md` and `system-state.md`
- [ ] Extract patterns if evidence supports (2+ occurrences)
- [ ] Update knowledge index
- [ ] Commit all changes

### Value Function Quick Reference

**V_instance = (Accuracy + Completeness + Usability + Maintainability) / 4**
- Accuracy: Technical correctness, working procedures
- Completeness: Coverage of user needs, workflows
- Usability: Clarity, navigation, examples
- Maintainability: Modularity, consistency, automation

**V_meta = (Completeness + Effectiveness + Reusability + Validation) / 4**
- Completeness: Lifecycle coverage, pattern catalog
- Effectiveness: Problem resolution, efficiency gains
- Reusability: Generalizability, adaptation effort
- Validation: Empirical grounding, metrics, testing

**Convergence**: Both ≥ 0.80 and stable for 2+ iterations

### File Organization Quick Reference

**System Files**:
- `system-state.md` - Current methodology state and value scores
- `iteration-log.md` - Chronological iteration history
- `knowledge-index.md` - Map of all knowledge artifacts

**Capabilities** (meta-agent lifecycle):
- `capabilities/doc-collect.md` - Data collection guidance
- `capabilities/doc-strategy.md` - Strategy formation guidance
- `capabilities/doc-execute.md` - Execution patterns
- `capabilities/doc-evaluate.md` - Evaluation rubrics
- `capabilities/doc-converge.md` - Convergence assessment

**Agents** (domain executors):
- `agents/doc-writer.md` - Writing patterns
- `agents/doc-validator.md` - Validation patterns
- `agents/doc-organizer.md` - Organization patterns
- `agents/content-analyzer.md` - Gap analysis patterns

**Knowledge**:
- `patterns/` - Extracted patterns
- `templates/` - Reusable templates
- `principles/` - Universal principles
- `best-practices/` - Context-specific practices
- `methodology/` - Project-wide reusable knowledge

**Data**:
- `data/iteration-N-evidence.md` - Evidence collected
- `data/iteration-N-analysis.md` - Data analysis
- `data/iteration-N-execution.md` - Execution observations
- `data/iteration-N-validation.md` - Validation results

---

## End of Iteration Prompts Document

This document provides the complete framework for developing documentation methodology through BAIME. Follow the iteration cycle systematically, maintain honest dual-layer evaluation, and let the methodology evolve based on empirical evidence from actual documentation work on the meta-cc project.

**Remember**: The goal is twofold:
1. **Instance**: Produce high-quality documentation for meta-cc (README updates, BAIME guide)
2. **Meta**: Develop transferable documentation methodology for Claude Code projects

Both objectives are equally important. Success means both V_instance and V_meta reach ≥ 0.80 through evidence-based, systematic methodology development.
