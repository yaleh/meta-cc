# Stage 1: Agent Formalization Inventory

**Date**: 2025-10-03
**Branch**: feature/agent-formalization
**Total Files**: 14 agent .md files

---

## File Statistics

| File | Lines | Status | Category | Reduction Potential |
|------|-------|--------|----------|---------------------|
| meta-coach.md | 1092 | 🔴 NEEDS REPLACEMENT | Verbose | ~85% (1092 → ~165) |
| prompt-refiner.md | 604 | 🔴 NEEDS REPLACEMENT | Verbose | ~87% (604 → ~80) |
| pattern-analyzer.md | 543 | 🔴 NEEDS REPLACEMENT | Verbose | ~82% (543 → ~98) |
| prompt-suggester.md | 441 | 🔴 NEEDS REPLACEMENT | Verbose | ~78% (441 → ~97) |
| doc-updater.md | 394 | 🟡 PARTIAL REPLACEMENT | Mixed | ~77% (394 → ~90) |
| architecture-advisor.md | 59 | ✅ ALREADY COMPACT | Formal | Keep as-is |
| stage-executor.md | 51 | ✅ ALREADY COMPACT | Formal | Keep as-is |
| prompt-distiller.md | 37 | ✅ ALREADY COMPACT | Formal | Keep as-is |
| simple-phase-planner.md | 30 | ✅ ALREADY COMPACT | Formal | Keep as-is |
| test-runner-fixer.md | 26 | ✅ ALREADY COMPACT | Formal | Keep as-is |
| simple-phase-executor.md | 25 | ✅ ALREADY COMPACT | Formal | Keep as-is |
| phase-verifier-and-fixer.md | 24 | ✅ ALREADY COMPACT | Formal | Keep as-is |
| git-committer.md | 20 | ✅ ALREADY COMPACT | Formal | Keep as-is |
| project-planner.md | 16 | ✅ ALREADY COMPACT | Formal | Keep as-is |
| **TOTAL** | **3362** | | | **~60%** (3362 → ~1347) |

---

## Classification Summary

### ✅ Already Compact (8 files, 288 lines total)
Files that already use pure formal specifications with minimal prose:
1. **project-planner.md** (16 lines) - Pure lambda calculus, reference template
2. **git-committer.md** (20 lines) - Operator notation (Φ, Ψ, Γ, Δ, Τ)
3. **phase-verifier-and-fixer.md** (24 lines) - Sequential verification spec
4. **simple-phase-executor.md** (25 lines) - Functional composition spec
5. **test-runner-fixer.md** (26 lines) - Compact notation with cleanup
6. **simple-phase-planner.md** (30 lines) - Constraint-based spec
7. **prompt-distiller.md** (37 lines) - DSL transformation spec
8. **stage-executor.md** (51 lines) - Comprehensive but compact
9. **architecture-advisor.md** (59 lines) - Full formal spec

**Action**: Keep as-is or minor refinement only

### 🟡 Partial Replacement (1 file, 394 lines)
Files with formal specs but significant verbose prose:
1. **doc-updater.md** (394 lines)
   - Lines 1-46: ✅ Formal spec (already compact)
   - Lines 47-394: 🔴 Verbose prose (347 lines)
   - Target: Keep formal spec, condense prose into comments/examples

**Action**: Replace prose sections, keep formal spec core

### 🔴 Needs Replacement (5 files, 2680 lines total)
Files with excessive verbose content that can be formalized:
1. **meta-coach.md** (1092 lines)
   - Lines 1-47: ✅ Formal spec header
   - Lines 48-1092: 🔴 Verbose documentation (1045 lines)
   - Includes: examples, analysis tools, Phase 10/11 guides, methodology

2. **prompt-refiner.md** (604 lines)
   - Lines 1-50: ✅ Formal spec header
   - Lines 51-604: 🔴 Verbose prose (554 lines)
   - Includes: workflow, examples, templates, best practices

3. **pattern-analyzer.md** (543 lines)
   - Lines 1-50: ✅ Formal spec header
   - Lines 51-543: 🔴 Verbose documentation (493 lines)
   - Includes: usage guides, examples, artifact templates

4. **prompt-suggester.md** (441 lines)
   - Lines 1-50: ✅ Formal spec header
   - Lines 51-441: 🔴 Verbose prose (391 lines)
   - Includes: methodology, examples, templates

5. **doc-updater.md** (394 lines - also in partial category)
   - Lines 1-46: ✅ Formal spec
   - Lines 47-394: 🔴 Prose and examples (347 lines)

**Action**: Full content replacement - encode all prose semantics in formal notation

---

## Semantic Extraction Analysis

### Common Patterns Found

**1. Workflow Sequences** (All 5 verbose files)
- Current: Natural language step-by-step
- Formal: `step1 → step2 → step3` or `f ∘ g ∘ h`
- Example: "First read, then analyze, finally report" → `read → analyze → report`

**2. Constraints and Validation** (All files)
- Current: Bullet points with "must", "should", "cannot"
- Formal: Logic operators `∀x: constraint(x)`, `∃y: property(y)`
- Example: "Must have at least 3 files" → `|files| ≥ 3 ∨ error`

**3. Conditional Logic** (prompt-refiner, pattern-analyzer)
- Current: If-then prose
- Formal: Guards or implication `condition ⇒ action`
- Example: "If score < 0.3 then reject" → `score < 0.3 ⇒ reject(P)`

**4. Type Definitions** (All files)
- Current: Prose descriptions of data structures
- Formal: Set notation or type definitions
- Example: "Recommendation has file, line, severity" → `type Recommendation = {file: string, line: nat, severity: Level}`

**5. Examples and Use Cases** (Verbose files)
- Current: 100-300 lines of prose examples
- Formal: Can be encoded as test cases or inline examples in formal spec
- Strategy: Either formalize as invariants or omit (redundant with formal spec)

---

## Formalization Strategy by File

### meta-coach.md (1092 → ~165 lines, 85% reduction)

**Current Structure**:
- Formal spec header (47 lines) ✅
- Analysis Tools section (50 lines) - Bash command examples
- Cross-Project Analysis (30 lines)
- Query Tools section (80 lines)
- Phase 10 Advanced Analysis (150 lines)
- Phase 11 Unix Composability (155 lines)
- Coaching Methodology (80 lines)
- Example Sessions (200 lines)
- Best Practices (300 lines)

**Formalization Plan**:
```
λ(session_history, user_query) → coaching_guidance:

analysis_tools :: Session → Tool_Set
analysis_tools(S) = {
  stats: parse_stats(S),
  errors: analyze_errors(S, window),
  tools: query_tools(S, filters),
  cross_project: compare(S, other_projects),
  phase10: {filters, aggregation, time_series, files},
  phase11: {streaming, exit_codes, pipelines}
}

coach_methodology :: Query → Response
coach_methodology(Q) = listen → gather_data → analyze → reflect → recommend → implement

workflow :: Coaching_Session → Outcome
workflow(C) = {
  Phase1: listen(user_intent) ∧ ask_clarifying_questions,
  Phase2: gather_data(session_stats ∪ error_patterns ∪ tool_usage),
  Phase3: analyze(identify_patterns ∧ measure_efficiency ∧ detect_anti_patterns),
  Phase4: reflect(present_findings ∧ encourage_discovery),
  Phase5: recommend(tiered_suggestions ∧ actionable_steps),
  Phase6: implement(provide_code ∨ guide_setup)
}

phase10_capabilities :: Advanced_Analysis
phase10_capabilities = {
  filtering: SQL_like(WHERE, AND, OR, IN, BETWEEN, LIKE, REGEXP),
  aggregation: group_by(tool ∨ status ∨ uuid) ∧ metrics(count, error_rate),
  time_series: bucket(hour ∨ day ∨ week) ∧ analyze(trends),
  files: track(operations) ∧ identify(hotspots)
}

phase11_capabilities :: Unix_Composability
phase11_capabilities = {
  streaming: JSONL_output(--stream),
  exit_codes: {0: success, 1: error, 2: no_results},
  io_separation: data→stdout ∧ logs→stderr,
  pipelines: compatible(jq ∪ grep ∪ awk ∪ sed)
}

constraints:
- data_driven: ∀recommendation → backed_by(session_intelligence)
- actionable: ∀suggestion → implementable ∧ tested
- pedagogical: guide_discovery > prescribe_solutions
- iterative: measure → change → measure → adapt

output :: Coaching_Session → Comprehensive_Report
output(C) = insights ∧ recommendations(tiered) ∧ implementation_guidance ∧ follow_up
```

**Size**: ~165 lines (85% reduction from 1092)

---

### prompt-refiner.md (604 → ~80 lines, 87% reduction)

**Current Structure**:
- Formal spec header (50 lines) ✅
- Understanding Phase prose (80 lines)
- Enrichment Phase prose (60 lines)
- Quality Checklist examples (100 lines)
- Scoring methodology examples (80 lines)
- Refinement templates (150 lines)
- Best practices (84 lines)

**Formalization Plan**:
```
λ(vague_prompt, context) → refined_prompt:

workflow :: Prompt → Refined_Prompt
workflow(P) = understand(P) → enrich(context) → check_quality(P) → score(P) → refine(P) → validate(result)

understand :: Prompt → Intent_Analysis
understand(P) = {
  literal: extract_text(P),
  inferred: deduce_goal(P),
  ambiguities: identify_unclear(P),
  gaps: detect_missing(P)
}

enrich :: Intent → Contextual_Data
enrich(I) = query(project_files) ∧ retrieve(successful_patterns) ∧ analyze(recent_history) ∧ reference(proven_workflows)

quality_elements :: Checklist
quality_elements = [clear_goal, context, constraints, acceptance, deliverables]

scoring :: Prompt → Score
scoring(P) = Σ(present(elements)) / 5 where:
  excellent: [0.9, 1.0]
  good: [0.7, 0.9)
  fair: [0.5, 0.7)
  poor: [0.3, 0.5)
  very_poor: [0, 0.3)

refine :: (Intent, Context, Patterns) → Structured_Prompt
refine(I, C, P) = apply_template ∧ inject_context ∧ define_constraints ∧ specify_acceptance ∧ list_deliverables

template_structure :: Standard
template_structure = {
  header: action_verb + target + purpose,
  goal: measurable ∧ specific,
  scope: included ∪ excluded,
  constraints: technical ∩ resource ∩ quality,
  deliverables: files ∪ artifacts ∪ reports,
  acceptance: testable ∧ verifiable ∧ quantitative
}

validation :: Refined_Prompt → Quality_Gate
validation(R) = verify(intent_preserved) ∧ check(clarity_improved) ∧ measure(completeness_increased)

constraints:
- preserve_intent: refined ⊇ original_intent
- enhance_clarity: ambiguity(refined) < ambiguity(original)
- add_structure: completeness(refined) > completeness(original)
- actionable: ready_to_execute ∧ no_clarification_needed

output :: Refinement → Report
output(R) = {
  original_score: score(before),
  refined_prompt: improved_version,
  refined_score: score(after),
  improvement_analysis: delta_breakdown,
  usage_instructions: how_to_apply
}
```

**Size**: ~80 lines (87% reduction from 604)

---

### pattern-analyzer.md (543 → ~98 lines, 82% reduction)

**Current Structure**:
- Formal spec header (50 lines) ✅
- Collection methodology (60 lines)
- Pattern identification (80 lines)
- Artifact generation (120 lines)
- Examples and templates (233 lines)

**Formalization Plan**:
```
λ(session_data, threshold) → automation_artifacts:

collect :: Session → Raw_Patterns
collect(S) = extract(turns ∪ tools ∪ errors ∪ sequences) ∧ compute(frequencies) ∧ identify(correlations)

identify :: Raw_Patterns → Classified_Patterns
identify(P) = cluster(similar) ∧ measure(frequency) ∧ classify(type) ∧ score(priority) ∧ filter(threshold)

pattern_types :: Pattern → Category
pattern_types(P) = {
  command: freq(P) ≥ 3 ∧ structure(imperative) ∧ deterministic,
  query: freq(P) ≥ 5 ∧ structure(interrogative) ∧ data_retrieval,
  validation: freq(P) ≥ 3 ∧ intent(verify) ∧ boolean_result,
  sequence: freq(chain) ≥ 5 ∧ ordered(tools) ∧ repeatable,
  workflow: freq(W) ≥ 3 ∧ multi_step ∧ contextual
}

artifact_mapping :: Pattern → Artifact_Type
artifact_mapping(P) = {
  slash_command: deterministic(P) ∧ query_like ∧ ¬complex_reasoning,
  subagent: conversational ∧ contextual_decisions ∧ multi_step ∧ requires_LLM,
  hook: validation ∧ preventive ∧ event_triggered ∧ non_interactive
}

frequency_tiers :: Pattern → Priority
frequency_tiers(P) = {
  high: occurrences ≥ 10,
  medium: 5 ≤ occurrences < 10,
  low: 3 ≤ occurrences < 5,
  ignore: occurrences < 3
}

generate :: Pattern → Artifact
generate(P) = {
  template: select_template(artifact_type(P)),
  parameters: extract_variations(P),
  documentation: generate_usage_guide(P),
  roi: estimate(time_saved ∧ effort_to_create)
}

roi_calculation :: Artifact → ROI_Score
roi_calculation(A) = {
  time_saved_per_use: avg_duration(manual) - avg_duration(automated),
  frequency: occurrences_per_week(pattern),
  creation_effort: estimated_hours(development),
  payback_period: creation_effort / (time_saved * frequency),
  decision: payback_period < 2_weeks ⇒ recommend(A)
}

constraints:
- data_driven: ∀artifact → backed_by(frequency ≥ threshold)
- actionable: ∀artifact → ready_to_deploy ∧ tested
- prioritized: ordered_by(frequency ∧ impact ∧ roi)
- roi_justified: creation_effort < cumulative_time_saved

output :: Analysis → Recommendations
output(A) = {
  patterns: identified_patterns(sorted_by_frequency),
  artifacts: generated_artifacts(ready_to_use),
  recommendations: prioritized_list(with_roi),
  implementation_guide: step_by_step_deployment
}
```

**Size**: ~98 lines (82% reduction from 543)

---

### prompt-suggester.md (441 → ~97 lines, 78% reduction)

**Current Structure**:
- Formal spec header (50 lines) ✅
- Gathering intelligence (60 lines)
- Analysis methodology (70 lines)
- Prioritization logic (80 lines)
- Suggestion templates (181 lines)

**Formalization Plan**:
```
λ(session_context, project_state) → prioritized_suggestions:

gather :: Session → Intelligence
gather(S) = {
  recent: query(last_N_turns(20)),
  state: analyze(project_files ∪ git_status ∪ todo_list),
  workflows: retrieve(successful_patterns),
  history: assess(trajectory ∧ blockers ∧ velocity)
}

analyze :: Intelligence → Insights
analyze(I) = {
  intent_trajectory: trace(user_goals) → {progressing, stuck, exploring},
  incomplete_tasks: identify(unfinished ∧ mentioned),
  blockers: detect(repeated_errors ∨ stuck_patterns ∨ uncertainty),
  session_health: measure(progress_rate ∧ error_rate ∧ tool_efficiency)
}

intent_states :: User_Messages → State
intent_states(M) = {
  progressing: clear_direction ∧ consistent_focus ∧ forward_momentum,
  stuck: repeated_questions ∨ uncertainty_signals ∨ circular_patterns,
  exploring: diverse_topics ∧ low_commitment ∧ open_ended_queries
}

prioritize :: Insights → Ranked_Suggestions
prioritize(I) = rank_by(urgency ∧ continuity ∧ success_probability ∧ proven_pattern_match)

priority_levels :: Suggestion → Priority
priority_levels(S) = {
  high: blocks_progress ∨ (natural_continuation ∧ proven_pattern ∧ high_value),
  medium: important ∧ ¬blocking ∧ partial_match ∧ medium_value,
  low: optional ∨ exploratory ∨ context_switch_required ∨ low_value
}

suggestion_quality :: Prompt → Quality_Score
suggestion_quality(P) = score({
  clear_goal: specific_verb ∧ concrete_target ∧ measurable_outcome,
  rich_context: links_to(project_state) ∧ explains(why) ∧ shows(trajectory),
  defined_constraints: boundaries ∧ limits ∧ quality_requirements,
  testable_acceptance: verifiable ∧ quantitative ∧ unambiguous,
  explicit_deliverables: files ∧ artifacts ∧ reports ∧ commits
})

recommend :: Suggestion → Complete_Prompt
recommend(S) = {
  prompt: structured_template(S),
  rationale: justify_with_data(session_intelligence),
  workflow: predict_tool_sequence(S),
  success_rate: estimate(based_on_historical_patterns),
  effort: estimate_time(S),
  impact: estimate_value(S)
}

suggestion_count :: Context → Count
suggestion_count(C) = {
  standard: 5_suggestions,
  stuck: 3_suggestions(focused_on_unblocking),
  exploring: 7_suggestions(diverse_options)
}

constraints:
- data_driven: ∀recommendation → backed_by(session_data ∧ proven_patterns)
- actionable: ∀prompt → complete ∧ ready_to_use ∧ ¬needs_clarification
- prioritized: ordered_by(urgency ∧ continuity ∧ success_probability)
- respectful: user_chooses ∧ ¬prescriptive ∧ explain_reasoning

output :: Suggestion_Set → Formatted_Response
output(S) = {
  analysis: session_summary ∧ trajectory_assessment,
  suggestions: top_N(ranked_prompts),
  rationale: data_driven_justification(∀suggestion),
  next_steps: recommended_action ∧ alternatives
}
```

**Size**: ~97 lines (78% reduction from 441)

---

### doc-updater.md (394 → ~90 lines, 77% reduction)

**Current Structure**:
- Formal spec header (46 lines) ✅ Keep
- Purpose and behavior (30 lines) - Can formalize
- Workflow description (50 lines) - Can formalize
- Examples and use cases (180 lines) - Can condense to inline examples
- Best practices (88 lines) - Can formalize as constraints

**Formalization Plan**:
```
λ(files, changes) → updated_docs:

workflow :: Documentation_Task → Updated_Files
workflow(T) = plan → track_start → batch_edit → verify → commit → track_complete

plan :: Change_Set → Edit_Plan
plan(C) = {
  scope: identify_targets(C) | scope ≤ 5_files,
  edits: group_by_file(C) ∧ optimize_sequence,
  validation: define_checks(syntax ∧ accuracy ∧ completeness),
  rollback: prepare_restore_points
}

batch_edit :: Edit_Plan → Edited_Files
batch_edit(P) = ∀file ∈ P.scope:
  read(file) → apply_edits(P.edits[file]) → verify_syntax(file) → verify_content(file)

track :: Progress → Todo_State
track(P) = {
  start: TodoWrite([{task: "in_progress"}]),
  milestones: TodoWrite([completed_tasks]) ∀ checkpoint,
  complete: TodoWrite([all: "completed"])
}

commit :: Edited_Files → Git_Commit
commit(E) = stage(E) → generate_message(E) → git_commit → verify_clean_status

message_template :: Changes → Commit_Message
message_template(C) = {
  type: "docs" ∨ "refactor",
  scope: affected_area(C),
  summary: concise_description(C),
  body: detailed_changes(C) ∧ rationale(C)
}

optimization_patterns :: Workflow_Statistics
optimization_patterns = {
  edit_batching: consecutive_edits ≈ 2.5 (optimal),
  progress_tracking: TodoWrite @ {start, milestones, end},
  file_focus: |scope| ≤ 5 (reduce_context_switching),
  zero_errors: pre_validate ∧ post_verify ∧ rollback_on_fail
}

constraints:
- batching: |consecutive_edits| ≥ 2 (efficiency_gain)
- validation: syntax_check ∧ content_accuracy ∧ no_information_loss
- atomicity: all_succeed ∨ rollback_all
- tracking: progress_visible ∀ user ∀ stage

success_metrics :: Execution → Quality
success_metrics(E) = {
  accuracy: no_syntax_errors ∧ content_preserved ∧ intent_fulfilled,
  efficiency: edit_batching_optimal ∧ minimal_file_switches,
  visibility: todos_updated ∀ milestone,
  reliability: zero_regressions ∧ git_clean
}

examples :: Common_Use_Cases
examples = {
  readme_update: update(README.md, new_phase_features),
  multi_doc_sync: batch_edit([README, guide, api_docs], consistent_terminology),
  agent_enhancement: update(agent.md, new_capabilities ∪ examples)
}

output :: Documentation_Update → Summary
output(D) = {
  files_modified: list(affected_files),
  changes_applied: count(edits),
  validation_results: all_checks_passed,
  commit_hash: git_sha(commit)
}
```

**Size**: ~90 lines (77% reduction from 394)

---

## Already Compact Files (Keep As-Is)

These files are already optimally compact and use pure formal specifications:

### project-planner.md (16 lines) ✅
- Pure lambda calculus with constraints
- Reference template for other agents
- **Action**: Keep exactly as-is (reference standard)

### git-committer.md (20 lines) ✅
- Operator notation (Φ, Ψ, Γ, Δ, Τ)
- Execution sequence
- **Action**: Keep as-is

### phase-verifier-and-fixer.md (24 lines) ✅
- Sequential verification specification
- **Action**: Keep as-is

### simple-phase-executor.md (25 lines) ✅
- Functional composition
- Termination conditions
- **Action**: Keep as-is

### test-runner-fixer.md (26 lines) ✅
- Compact notation with cleanup phases
- **Action**: Keep as-is

### simple-phase-planner.md (30 lines) ✅
- Constraint-based specification
- **Action**: Keep as-is

### prompt-distiller.md (37 lines) ✅
- DSL transformation spec
- **Action**: Keep as-is

### stage-executor.md (51 lines) ✅
- Comprehensive but still compact
- **Action**: Keep as-is or very minor refinement

### architecture-advisor.md (59 lines) ✅
- Full formal specification
- **Action**: Keep as-is

---

## Summary Statistics

### Files Requiring Replacement (5 files)
- meta-coach.md: 1092 → 165 lines (85% reduction)
- prompt-refiner.md: 604 → 80 lines (87% reduction)
- pattern-analyzer.md: 543 → 98 lines (82% reduction)
- prompt-suggester.md: 441 → 97 lines (78% reduction)
- doc-updater.md: 394 → 90 lines (77% reduction)

**Subtotal**: 3074 → 530 lines (83% average reduction)

### Files Keeping As-Is (9 files)
**Subtotal**: 288 lines (no change)

### Project Total
- **Before**: 3362 lines
- **After**: 818 lines (530 + 288)
- **Overall Reduction**: 76% (2544 lines removed)

---

## Formalization Patterns Identified

### 1. Workflow Encoding
**Pattern**: Multi-step processes
**Formal**: `step1 → step2 → step3` or function composition `f ∘ g ∘ h`

### 2. Constraint Definition
**Pattern**: "Must", "should", "cannot" statements
**Formal**: Logic guards `∀x: property(x)`, `constraint ⇒ action`

### 3. Type Specifications
**Pattern**: Data structure descriptions
**Formal**: Set notation `type T = {field1: Type1, field2: Type2}`

### 4. Conditional Logic
**Pattern**: If-then-else prose
**Formal**: Guards, pattern matching, or implication `condition ⇒ action`

### 5. Quality Metrics
**Pattern**: Success criteria descriptions
**Formal**: Quantitative expressions `metric ≥ threshold`, `score ∈ [range]`

### 6. Categorization
**Pattern**: Lists of types or categories
**Formal**: Discriminated unions, set membership `type ∈ {A, B, C}`

---

## Semantic Preservation Strategy

For each verbose file, the following semantic elements will be preserved in formal notation:

1. **Input/Output Signatures**: All function signatures preserved as `λ(inputs) → outputs`
2. **Constraints**: All "must", "should", "cannot" → logic operators
3. **Workflows**: All step-by-step processes → sequential operators
4. **Examples**: Essential examples → inline within formal spec or as test cases
5. **Edge Cases**: Error handling, special cases → guards and conditionals
6. **Quality Metrics**: Success criteria → quantitative expressions

**Zero Information Loss Guarantee**: Every behavioral semantic in the verbose prose will be encoded in the formal specification.

---

## Next Steps (Stage 2)

For each of the 5 files requiring replacement:
1. Create complete formal specification encoding all semantics
2. Show before/after comparison
3. Provide semantic mapping (line-by-line correspondence)
4. Calculate exact size reduction
5. **Wait for human approval** before Stage 3

---

## Risk Assessment

### Low Risk (9 files)
Files already compact - no changes needed

### Medium Risk (5 files)
Files with extensive prose but clear formal spec headers
- Risk: Potential semantic loss during compression
- Mitigation: Detailed semantic mapping + human review

### Critical Success Factors
1. ✅ Human review of all before/after comparisons (Stage 2)
2. ✅ Semantic equivalence verification
3. ✅ Git backup before replacement (Stage 3)
4. ✅ Rollback capability if issues detected

---

**Stage 1 Complete**: Inventory and semantic extraction finished. Ready for Stage 2 design phase.
