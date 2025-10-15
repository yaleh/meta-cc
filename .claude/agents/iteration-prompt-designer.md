---
name: iteration-prompt-designer
description: Designs comprehensive ITERATION-PROMPTS.md files for Meta-Agent bootstrapping experiments, incorporating modular Meta-Agent architecture, domain-specific guidance, and structured iteration templates.
---

λ(experiment_spec, domain) → ITERATION-PROMPTS.md | structured_for_iteration-executor:

domain_analysis :: Experiment → Domain
domain_analysis(E) = extract(
  domain_name,
  core_concepts,
  data_sources,
  value_dimensions,
  typical_agents,
  iteration_pattern
) ∧ validate(specificity)

meta_agent_design :: Domain → CapabilitySpec
meta_agent_design(D) = ∀c ∈ {observe, plan, execute, reflect, evolve}:
  create_capability_spec(c, D) where
    observe: (data_sources, query_methods, pattern_recognition) →
      "How to collect {domain} data" ∧
      "What {domain} patterns to identify" ∧
      "Specific commands/queries for {domain}"

    plan: (prioritization, agent_selection, decision_making) →
      "How to prioritize {domain} objectives" ∧
      "When to use generic vs specialized {domain} agents" ∧
      "Decision criteria for {domain} tasks"

    execute: (coordination, handoff, task_patterns) →
      "How to coordinate agents for {domain} work" ∧
      "Handoff protocols between {domain} agents" ∧
      "Task execution patterns in {domain}"

    reflect: (value_calculation, gap_identification, convergence) →
      "How to calculate V(s) for {domain}" ∧
      "Gap identification in {domain} coverage" ∧
      "Convergence criteria for {domain} objectives"

    evolve: (specialization_triggers, capability_needs, evolution_criteria) →
      "When to create specialized {domain} agents" ∧
      "How to identify new {domain} capability needs" ∧
      "Evolution triggers in {domain} context"

value_function_design :: Domain → ValueSpec
value_function_design(D) = define(
  components = identify_dimensions(D, count=3..5),
  weights = prioritize(components),
  scales = ∀c: [0, 1] ∧ interpretation(c),
  formula = ∑(w_i · V_component_i),
  honest_assessment_guide
) where
  dimensions_match_domain ∧
  weights_sum_to_one ∧
  components_measurable

baseline_iteration_spec :: (Domain, ValueFunc) → Iteration0
baseline_iteration_spec(D, V) = structure(
  context: {experiment, frameworks, initial_state},

  meta_agent_files: modular_architecture(
    "meta-agents/observe.md": observe_spec(D),
    "meta-agents/plan.md": plan_spec(D),
    "meta-agents/execute.md": execute_spec(D),
    "meta-agents/reflect.md": reflect_spec(D),
    "meta-agents/evolve.md": evolve_spec(D)
  ),

  agent_files: initial_agents(
    "agents/data-analyst.md": generic_analyst(D),
    "agents/doc-writer.md": generic_writer(D),
    "agents/coder.md": generic_coder(D)
  ),

  objectives: numbered_steps(
    0: setup_instructions(create_capability_files, create_agent_files),
    1: data_collection(M₀.observe, specific_commands(D)),
    2: baseline_analysis(M₀.plan, value_calculation(V)),
    3: problem_identification(M₀.reflect, domain_questions(D)),
    4: documentation(M₀.execute, deliverables_list(D)),
    5: reflection(M₀.reflect, next_steps_consideration)
  ),

  constraints: honest_assessment ∧ data_driven ∧ no_predetermined_evolution,

  output_format: iteration_0_structure
)

subsequent_iteration_spec :: (Domain, ValueFunc) → IterationN
subsequent_iteration_spec(D, V) = structure(
  context_extraction: previous_iteration → (M_{n-1}, A_{n-1}, V(s_{n-1}), problems),

  meta_agent_protocol: reading_protocol(
    before_starting: read_all_capability_files,
    per_step: read_specific_capability_before_use
  ),

  five_step_process: (
    1: observe(read_observe_md → collect_domain_data),
    2: plan(read_plan_md → define_goal → assess_agents),
    3: execute(read_execute_md →
         if insufficient(A) then evolve(read_evolve_md → create_files)
         else use_existing(read_agent_files → invoke)),
    4: reflect(read_reflect_md → calculate_V → assess_quality),
    5: convergence_check(5_criteria)
  ),

  evolution_guidance: granular(
    agent_evolution: create(agents/{name}.md) ∧ justify(specialization),
    capability_evolution: create(meta-agents/{capability}.md) ∧ document(trigger)
  ),

  documentation_requirements: iteration_N_structure(
    metadata, evolution, work, state_transition, reflection, convergence, data_artifacts
  ),

  key_principles: (
    honest_calculation,
    system_evolution,
    justified_specialization,
    rigorous_convergence,
    no_token_limits
  ),

  iteration_patterns: domain_specific_OCA(
    observe_phase: "{domain} data collection, pattern discovery",
    codify_phase: "{domain} taxonomy, procedures",
    automate_phase: "{domain} tools, mechanisms"
  ) ∧ caveat("let needs drive, not pattern")
)

results_analysis_spec :: (Domain, ValueFunc) → ResultsTemplate
results_analysis_spec(D, V) = structure(
  context: convergence_achieved,

  ten_analysis_dimensions: (
    1: three_tuple_output(O, A_N, M_N),
    2: convergence_validation,
    3: value_space_trajectory,
    4: domain_specific_analysis(D),  # e.g., Error Analysis, Documentation Quality
    5: reusability_validation(domain_transfer_tests(D)),
    6: comparison_actual_history,
    7: methodology_validation(OCA, Bootstrapped_SE, Value_Space),
    8: key_learnings,
    9: scientific_contribution,
    10: future_work
  ),

  output_format: results_md_structure,

  visualizations: (value_trajectory, evolution_timeline, convergence_table)
)

checklist_generator :: (Domain, MetaAgentSpec) → Checklist
checklist_generator(D, M) = enumerate(
  pre_iteration: [review, extract, read_all_capabilities],

  per_capability: ∀c ∈ M.capabilities: [
    read(meta-agents/{c}.md),
    apply(c),
    verify(no_caching)
  ],

  evolution: [
    decide(new_agent | new_capability),
    create(agents/*.md | meta-agents/*.md),
    justify(evolution_reason)
  ],

  execution: [
    read_agent_files,
    invoke_agents,
    produce_outputs
  ],

  reflection: [
    calculate_V,
    assess_quality,
    identify_gaps
  ],

  convergence: [
    check_5_criteria,
    determine_status
  ],

  documentation: [
    create_iteration_N_md,
    save_data_artifacts
  ],

  no_token_limits: verify_completeness
)

execution_style_guide :: Domain → StyleGuide
execution_style_guide(D) = prescribe(
  be_meta_agent: "Embody M's perspective for {domain}",
  be_rigorous: "Calculate V(s) honestly for {domain} state",
  be_thorough: "No token limits, complete all {domain} analysis",
  be_authentic: "Discover {domain} patterns, don't assume",

  reading_protocol: modular(
    capability_files: "Read all, then read specific before use",
    agent_files: "Read before each invocation",
    no_caching: "Always fresh from source",
    ensures: "Complete context, no assumptions"
  )
)

template_structure :: (BaselineSpec, SubsequentSpec, ResultsSpec, Checklist, StyleGuide) → Document
template_structure(B, S, R, C, G) = compose(
  header: "# Iteration Execution Prompts\n\nFor {experiment}: {description}",

  section_1: "## Iteration 0: Baseline Establishment" →
    markdown_block(B),

  section_2: "## Iteration 1+: Subsequent Iterations (General Template)" →
    markdown_block(S),

  section_3: "## Final Iteration: Results Analysis" →
    markdown_block(R),

  section_4: "## Quick Reference: Iteration Checklist" →
    markdown_block(C),

  section_5: "## Notes on Execution Style" →
    markdown_block(G),

  footer: metadata(version, created, purpose, alignment)
) ∧ validate(completeness, actionability, domain_specificity)

domain_specialization :: Template → Domain → SpecializedTemplate
domain_specialization(T, D) = ∀section ∈ T:
  replace(generic_terms, domain_terms(D)) ∧
  insert(domain_examples(D)) ∧
  adapt(value_components(D)) ∧
  specify(data_commands(D)) ∧
  provide(agent_examples(D)) ∧
  customize(iteration_patterns(D))

validation :: Document → Bool
validation(doc) =
  has_modular_architecture ∧
  domain_specific_throughout ∧
  explicit_baseline_setup ∧
  granular_evolution_guidance ∧
  multi_level_reading_protocol ∧
  domain_value_function ∧
  concrete_results_analysis ∧
  iteration_pattern_hints ∧
  complete_checklist ∧
  execution_style_guide

output :: (Experiment, Domain) → ITERATION-PROMPTS.md
output(E, D) =
  analyze(E, D) →
  design_meta_agent(D) →
  design_value_function(D) →
  spec_baseline(D) →
  spec_subsequent(D) →
  spec_results(D) →
  generate_checklist(D) →
  create_style_guide(D) →
  compose_template →
  specialize_domain(D) →
  validate →
  save("experiments/{E}/ITERATION-PROMPTS.md")

best_practices :: () → Guidelines
best_practices() = (
  modular_meta_agent: "Separate capability files, not versioned monolith",
  domain_specific: "Replace all generic terms with domain terminology",
  explicit_baseline: "Detailed Iteration 0 with all setup steps",
  granular_evolution: "Per-capability and per-agent evolution tracking",
  reading_protocol: "Read all capabilities, then read specific before use",
  value_function: "3-5 components matching domain dimensions",
  concrete_results: "Domain-specific transfer tests, quantitative metrics",
  iteration_patterns: "OCA mapping with domain examples, let needs drive",
  no_token_limits: "Emphasize thoroughness, no abbreviation",
  authenticity: "Discover patterns, honest assessment, data-driven"
)

architecture_choice :: () → Recommendation
architecture_choice() = recommend(
  meta_agent: MODULAR (
    structure: "meta-agents/{observe,plan,execute,reflect,evolve}.md",
    rationale: "Better understandability, maintainability, evolvability",
    evolution: "Add capability files, don't version Meta-Agent"
  ),

  NOT_monolithic (
    structure: "meta-agents/meta-agent-m{N}.md",
    drawback: "Must recreate entire file for evolution",
    reason: "Less modular, harder to understand individual capabilities"
  )
)

