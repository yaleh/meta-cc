---
name: meta-doc-health
description: Documentation health check based on access patterns, R/E ratios, file sizes, and role compliance.
keywords: documentation, health, maintenance, metrics, access-patterns, role-compliance
category: diagnostics
---

λ(scope) → doc_health_report | ∀doc ∈ {project_documentation}:

scope :: project | session

analyze :: Project → HealthReport
analyze(P) = collect(data) ∧ classify(roles) ∧ validate(constraints) ∧ recommend(actions)

collect :: Project → Metrics
collect(P) = {
  mcp_meta_cc.query_files(threshold=5, scope=scope),
  git_log_stats(docs/),
  glob("docs/**/*.md") → wc -l
}

classify :: Metrics → Roles
classify(M) = for each file {
  RE = reads / max(edits, 1),
  density = accesses / max(time_span_min, 1),
  
  role = match {
    (path == "CLAUDE.md") → 'context_base',
    (RE > 2.0 AND span > 10k) → 'specification',
    (accesses > 80 AND RE 1.0-1.5 AND span > 10k) → 'living_doc',
    (accesses 30-80 AND RE 1.0-2.0) → 'reference',
    (density > 0.1 OR span < 1k) → 'episodic',
    (accesses < 20 AND span > 10k) → 'archive',
    _ → 'unclassified'
  }
}

validate :: Roles → Violations
validate(R) = {
  limits = {
    context_base: {lines: 300, RE: (0.5, 1.0)},
    living_doc: {lines: 600, RE: (1.0, 1.5), min_access: 50},
    specification: {lines: ∞, RE: (2.0, 5.0)},
    reference: {lines: 800, RE: (1.0, 2.0)},
    episodic: {lines: ∞},
    archive: {lines: ∞}
  },

  for each (file, role) check {
    lines > limits[role].lines → error(size_violation),
    RE ∉ limits[role].RE → warning(re_anomaly),
    role == spec AND edits > reads → warning(unstable_spec),
    role != archive AND accesses < 20 AND span > 10k → info(archival),
    role == archive AND density > 0.001 → info(resurrection)
  }
}

recommend :: Violations → Actions
recommend(V) = {
  critical: V.errors → split_document,
  warnings: V.warnings → review_role | stabilize_spec,
  maintenance: V.info → archive | unarchive
}

output :: Analysis → Report
output(A) = {
  summary: {by_role, health_status},
  critical: {size_violations, split_recommendations},
  warnings: {re_anomalies, unstable_specs},
  maintenance: {archival_candidates, resurrection_candidates},
  top_docs: top_n(10),
  trends: {increased, decreased, stable},
  actions: {high, medium, low}
} where ¬execute(recommendations)

implementation_notes:
- roles: context_base(300), living(600), spec(∞), reference(800), episodic(∞), archive(∞)
- metrics: RE_ratio, access_density, lifecycle
- data: query_files, git log, wc -l
- frequency: monthly or pre-commit

role_definitions:
- context_base: implicitly loaded every request (CLAUDE.md), max 300 lines, RE 0.5-1.0
- living: high-frequency workspace (plan.md, README.md), max 600 lines, RE 1.0-1.5, min 50 accesses
- specification: stable reference (ADRs, proposals), no limit, RE 2.0-5.0+, rarely changed
- reference: on-demand guides (mcp.md, cli.md), max 800 lines, RE 1.0-2.0, 30-80 accesses
- episodic: burst creation then archive (phase plans), no limit during active, RE <0.5, density >0.1/min
- archive: historical docs, no limit, <20 accesses over lifespan, never updated

key_metrics:
- RE_ratio = reads / max(edits, 1): reveals document nature (<0.5 creation, 0.5-1.0 balanced, 1.0-2.0 reference, >2.0 spec)
- access_density = total_accesses / time_span_min: reveals intensity (>0.1 burst, 0.01-0.1 active, 0.001-0.01 normal, <0.001 archive)
- lifecycle: evergreen (>10k min), phase_bound (1k-10k min), sprint_bound (<1k min)

constraints:
- role_based: classify → validate
- evidence_based: ∀violation → ∃threshold
- actionable: prioritized recommendations
- automated: pre-commit hook compatible
