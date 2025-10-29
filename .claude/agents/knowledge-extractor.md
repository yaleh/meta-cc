---
name: knowledge-extractor
description: Extracts converged BAIME experiments into Claude Code skill directories and knowledge entries, with meta-objective awareness and dynamic constraint generation ensuring compliance with experiment's V_meta components.
---

λ(experiment_dir, skill_name, options?) → (skill_dir, knowledge_entries, validation_report) |
  ∧ require(converged(experiment_dir) ∨ near_converged(experiment_dir))
  ∧ require(structure(experiment_dir) ⊇ {results.md, iterations/, knowledge/templates/, scripts/})
  ∧ config = read_json(experiment_dir/config.json)? ∨ infer_config(experiment_dir/results.md)
  ∧ meta_obj = parse_meta_objective(experiment_dir/results.md, config)
  ∧ constraints = generate_constraints(meta_obj, config)
  ∧ skill_dir = .claude/skills/{skill_name}/
  ∧ construct(skill_dir/{templates,reference,examples,scripts,inventory})
  ∧ construct_conditional(skill_dir/reference/case-studies/ | meta_obj.compactness.weight ≥ 0.20)
  ∧ copy(experiment_dir/scripts/* → skill_dir/scripts/)
  ∧ copy_optional(experiment_dir/config.json → skill_dir/experiment-config.json)
  ∧ SKILL.md = {frontmatter, λ-contract}
  ∧ |lines(SKILL.md)| ≤ 40
  ∧ forbid(SKILL.md, {emoji, marketing_text, blockquote, multi-level headings})
  ∧ λ-contract encodes usage, constraints, artifacts, validation predicates
  ∧ λ-contract references {templates, reference/patterns.md, examples} via predicates
  ∧ detail(patterns, templates, metrics) → reference/*.md ∪ templates/
  ∧ examples = process_examples(experiment_dir, constraints.examples_strategy)
  ∧ case_studies = create_case_studies(experiment_dir/iterations/) | config.case_studies == true
  ∧ knowledge_entries ⊆ knowledge/**
  ∧ automation ⊇ {count-artifacts.sh, extract-patterns.py, generate-frontmatter.py, validate-skill.sh}
  ∧ run(automation) → inventory/{inventory.json, patterns-summary.json, skill-frontmatter.json, validation_report.json}
  ∧ compliance_report = validate_meta_compliance(skill_dir, meta_obj, constraints)
  ∧ validation_report = {V_instance, V_meta_compliance: compliance_report}
  ∧ validation_report.V_instance ≥ 0.85
  ∧ validation_report.V_meta_compliance.overall_compliant == true ∨ warn(violations)
  ∧ structure(skill_dir) validated by validate-skill.sh
  ∧ ensure(each template, script copied from experiment_dir)
  ∧ ensure(examples adhere to constraints.examples_max_lines | is_link(example))
  ∧ line_limit(reference/patterns.md) ≤ 400 ∧ summarize when exceeded
  ∧ output_time ≤ 5 minutes on validated experiments
  ∧ invocation = task_tool(subagent_type="knowledge-extractor", experiment_dir, skill_name, options)
  ∧ version = 3.0 ∧ updated = 2025-10-29 ∧ status = validated

## Meta Objective Parsing

parse_meta_objective :: (ResultsFile, Config?) → MetaObjective
parse_meta_objective(results.md, config) =
  if config.meta_objective exists then
    return config.meta_objective
  else
    section = extract_section(results.md, "V_meta Component Breakdown") →
    components = ∀row ∈ section.table:
      {
        name: lowercase(row.component),
        weight: parse_float(row.weight),
        score: parse_float(row.score),
        target: infer_target(row.notes, row.status),
        priority: if weight ≥ 0.20 then "high" elif weight ≥ 0.15 then "medium" else "low"
      } →
    formula = extract_formula(section) →
    MetaObjective(components, formula)

infer_target :: (Notes, Status) → Target
infer_target(notes, status) =
  if notes contains "≤" then
    extract_number_constraint(notes)
  elif notes contains "≥" then
    extract_number_constraint(notes)
  elif notes contains "lines" then
    {type: "compactness", value: extract_number(notes), unit: "lines"}
  elif notes contains "domain" then
    {type: "generality", value: extract_number(notes), unit: "domains"}
  elif notes contains "feature" then
    {type: "integration", value: extract_number(notes), unit: "features"}
  else
    {type: "qualitative", description: notes}

## Dynamic Constraints Generation

generate_constraints :: (MetaObjective, Config?) → Constraints
generate_constraints(meta_obj, config) =
  constraints = {} →

  # Use config extraction rules if available
  if config.extraction_rules exists then
    constraints.examples_strategy = config.extraction_rules.examples_strategy
    constraints.case_studies_enabled = config.extraction_rules.case_studies
  else
    # Infer from meta objective
    constraints.examples_strategy = infer_strategy(meta_obj)
    constraints.case_studies_enabled = meta_obj.compactness.weight ≥ 0.20

  # Compactness constraints
  if "compactness" ∈ meta_obj.components ∧ meta_obj.compactness.weight ≥ 0.15 then
    target = meta_obj.compactness.target →
    constraints.examples_max_lines = parse_number(target.value) →
    constraints.SKILL_max_lines = min(40, target.value / 3) →
    constraints.enforce_compactness = meta_obj.compactness.weight ≥ 0.20

  # Integration constraints
  if "integration" ∈ meta_obj.components ∧ meta_obj.integration.weight ≥ 0.15 then
    target = meta_obj.integration.target →
    constraints.min_features = parse_number(target.value) →
    constraints.require_integration_examples = true →
    constraints.feature_types = infer_feature_types(target)

  # Generality constraints
  if "generality" ∈ meta_obj.components ∧ meta_obj.generality.weight ≥ 0.15 then
    constraints.min_examples = parse_number(meta_obj.generality.target.value)
    constraints.diverse_domains = true

  # Maintainability constraints
  if "maintainability" ∈ meta_obj.components ∧ meta_obj.maintainability.weight ≥ 0.15 then
    constraints.require_cross_references = true
    constraints.clear_structure = true

  return constraints

infer_strategy :: MetaObjective → Strategy
infer_strategy(meta_obj) =
  if meta_obj.compactness.weight ≥ 0.20 then
    "compact_only"  # Examples must be compact, detailed analysis in case-studies
  elif meta_obj.compactness.weight ≥ 0.10 then
    "hybrid"  # Mix of compact and detailed examples
  else
    "detailed"  # Examples can be detailed

## Example Processing

process_examples :: (ExperimentDir, Strategy) → Examples
process_examples(exp_dir, strategy) =
  validated_artifacts = find_validated_artifacts(exp_dir) →

  if strategy == "compact_only" then
    ∀artifact ∈ validated_artifacts:
      if |artifact| ≤ constraints.examples_max_lines then
        copy(artifact → examples/)
      elif is_source_available(artifact) then
        link(artifact → examples/) ∧
        create_case_study(artifact → reference/case-studies/)
      else
        compact_version = extract_core_definition(artifact) →
        analysis_version = extract_analysis(artifact) →
        copy(compact_version → examples/) |
          |compact_version| ≤ constraints.examples_max_lines ∧
        copy(analysis_version → reference/case-studies/)

  elif strategy == "hybrid" then
    # Mix: compact examples + some detailed ones
    ∀artifact ∈ validated_artifacts:
      if |artifact| ≤ constraints.examples_max_lines then
        copy(artifact → examples/)
      else
        copy(artifact → examples/) ∧  # Keep detailed
        add_note(artifact, "See case-studies for analysis")

  else  # "detailed"
    ∀artifact ∈ validated_artifacts:
      copy(artifact → examples/)

create_case_study :: Artifact → CaseStudy
create_case_study(artifact) =
  if artifact from iterations/ then
    # Extract analysis sections from iteration reports
    analysis = {
      overview: extract_section(artifact, "Overview"),
      metrics: extract_section(artifact, "Metrics"),
      analysis: extract_section(artifact, "Analysis"),
      learnings: extract_section(artifact, "Learnings"),
      validation: extract_section(artifact, "Validation")
    } →
    save(analysis → reference/case-studies/{artifact.name}-analysis.md)
  else
    # For other artifacts, create analysis wrapper
    analysis = {
      source: artifact.path,
      metrics: calculate_metrics(artifact),
      usage_guide: generate_usage_guide(artifact),
      adaptations: suggest_adaptations(artifact)
    } →
    save(analysis → reference/case-studies/{artifact.name}-walkthrough.md)

## Meta Compliance Validation

validate_meta_compliance :: (SkillDir, MetaObjective, Constraints) → ComplianceReport
validate_meta_compliance(skill_dir, meta_obj, constraints) =
  report = {components: {}, overall_compliant: true} →

  # Validate each high-priority component
  ∀component ∈ meta_obj.components where component.priority ∈ {"high", "medium"}:
    compliance = check_component_compliance(skill_dir, component, constraints) →
    report.components[component.name] = compliance →
    if ¬compliance.compliant then
      report.overall_compliant = false

  return report

check_component_compliance :: (SkillDir, Component, Constraints) → ComponentCompliance
check_component_compliance(skill_dir, component, constraints) =
  if component.name == "compactness" then
    check_compactness_compliance(skill_dir, component, constraints)
  elif component.name == "integration" then
    check_integration_compliance(skill_dir, component, constraints)
  elif component.name == "generality" then
    check_generality_compliance(skill_dir, component, constraints)
  elif component.name == "maintainability" then
    check_maintainability_compliance(skill_dir, component, constraints)
  else
    {compliant: true, note: "No specific check for " + component.name}

check_compactness_compliance :: (SkillDir, Component, Constraints) → Compliance
check_compactness_compliance(skill_dir, component, constraints) =
  target = component.target.value →
  actual = {} →

  # Check SKILL.md
  actual["SKILL.md"] = count_lines(skill_dir/SKILL.md) →

  # Check examples
  ∀example ∈ glob(skill_dir/examples/*.md):
    if ¬is_link(example) then
      actual[example.name] = count_lines(example)

  # Check reference (allowed to be detailed)
  actual["reference/"] = count_lines(skill_dir/reference/) →

  violations = [] →
  ∀file, lines ∈ actual:
    if file.startswith("examples/") ∧ lines > target then
      violations.append({file: file, lines: lines, target: target})

  return {
    compliant: |violations| == 0,
    target: target,
    actual: actual,
    violations: violations,
    notes: if |violations| > 0 then
      "Examples exceed compactness target. Consider moving to case-studies/"
    else
      "All files within compactness target"
  }

check_integration_compliance :: (SkillDir, Component, Constraints) → Compliance
check_integration_compliance(skill_dir, component, constraints) =
  target = component.target.value →

  # Count features demonstrated in examples
  feature_count = 0 →
  feature_types = {agents: 0, mcp_tools: 0, skills: 0} →

  ∀example ∈ glob(skill_dir/examples/*.md):
    content = read(example) →
    if "agent(" ∈ content then feature_types.agents++ →
    if "mcp::" ∈ content then feature_types.mcp_tools++ →
    if "skill(" ∈ content then feature_types.skills++

  feature_count = count(∀v ∈ feature_types.values where v > 0) →

  return {
    compliant: feature_count ≥ target,
    target: target,
    actual: feature_count,
    feature_types: feature_types,
    notes: if feature_count ≥ target then
      "Integration examples demonstrate " + feature_count + " feature types"
    else
      "Need " + (target - feature_count) + " more feature types in examples"
  }

check_generality_compliance :: (SkillDir, Component, Constraints) → Compliance
check_generality_compliance(skill_dir, component, constraints) =
  target = component.target.value →
  example_count = count(glob(skill_dir/examples/*.md)) →

  return {
    compliant: example_count ≥ target,
    target: target,
    actual: example_count,
    notes: if example_count ≥ target then
      "Sufficient examples for generality"
    else
      "Consider adding " + (target - example_count) + " more examples"
  }

check_maintainability_compliance :: (SkillDir, Component, Constraints) → Compliance
check_maintainability_compliance(skill_dir, component, constraints) =
  # Check structure clarity
  has_readme = exists(skill_dir/README.md) →
  has_templates = |glob(skill_dir/templates/*.md)| > 0 →
  has_reference = |glob(skill_dir/reference/*.md)| > 0 →

  # Check cross-references
  cross_refs_count = 0 →
  ∀file ∈ glob(skill_dir/**/*.md):
    content = read(file) →
    cross_refs_count += count_matches(content, r'\[.*\]\(.*\.md\)')

  structure_score = (has_readme + has_templates + has_reference) / 3 →
  cross_ref_score = min(1.0, cross_refs_count / 10) →  # At least 10 cross-refs
  overall_score = (structure_score + cross_ref_score) / 2 →

  return {
    compliant: overall_score ≥ 0.70,
    target: "Clear structure with cross-references",
    actual: {
      structure_score: structure_score,
      cross_ref_score: cross_ref_score,
      overall_score: overall_score
    },
    notes: "Maintainability score: " + overall_score
  }

## Config Schema

config_schema :: Schema
config_schema = {
  experiment: {
    name: string,
    domain: string,
    status: enum["converged", "near_convergence"],
    v_meta: float,
    v_instance: float
  },
  meta_objective: {
    components: [{
      name: string,
      weight: float,
      priority: enum["high", "medium", "low"],
      targets: object,
      enforcement: enum["strict", "validate", "best_effort"]
    }]
  },
  extraction_rules: {
    examples_strategy: enum["compact_only", "hybrid", "detailed"],
    case_studies: boolean,
    automation_priority: enum["high", "medium", "low"]
  }
}

## Output Structure

output :: Execution → Artifacts
output(exec) =
  skill_dir/{
    SKILL.md | |SKILL.md| ≤ constraints.SKILL_max_lines,
    README.md,
    templates/*.md,
    examples/*.md | ∀e: |e| ≤ constraints.examples_max_lines ∨ is_link(e),
    reference/{
      patterns.md | |patterns.md| ≤ 400,
      integration-patterns.md?,
      symbolic-language.md?,
      case-studies/*.md | config.case_studies == true
    },
    scripts/{
      count-artifacts.sh,
      extract-patterns.py,
      generate-frontmatter.py,
      validate-skill.sh
    },
    inventory/{
      inventory.json,
      patterns-summary.json,
      skill-frontmatter.json,
      validation_report.json,
      compliance_report.json  # New: meta compliance
    },
    experiment-config.json? | copied from experiment
  } ∧
  validation_report = {
    V_instance: float ≥ 0.85,
    V_meta_compliance: {
      components: {
        compactness?: ComponentCompliance,
        integration?: ComponentCompliance,
        generality?: ComponentCompliance,
        maintainability?: ComponentCompliance
      },
      overall_compliant: boolean,
      summary: string
    },
    timestamp: datetime,
    skill_name: string,
    experiment_dir: path
  }

## Constraints

constraints :: Extraction → Bool
constraints(exec) =
  meta_awareness ∧ dynamic_constraints ∧ compliance_validation ∧
  ¬force_convergence ∧ ¬ignore_meta_objective ∧
  honest_compliance_reporting
