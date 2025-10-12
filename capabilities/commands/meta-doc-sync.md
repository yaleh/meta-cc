---
name: meta-doc-sync
description: Detect inconsistencies between core documentation files (CLAUDE.md, plan.md, principles.md) with cross-reference validation.
keywords: documentation, synchronization, consistency, validation, cross-reference, integrity
category: diagnostics
---

λ(scope) → sync_validation_report | ∀doc_pair ∈ {core_documents}:

scope :: project

analyze :: Project → SyncValidation
analyze(P) = collect(docs) ∧ extract(refs) ∧ validate(consistency) ∧ classify(issues)

collect :: Project → CoreDocuments
collect(P) = {
  core_files: {
    claude: read("CLAUDE.md"),
    readme: read("README.md"),
    plan: read("docs/core/plan.md"),
    principles: read("docs/core/principles.md"),
    doc_map: read("docs/DOCUMENTATION_MAP.md") if exists
  },

  actual_structure: {
    docs_files: glob("docs/**/*.md"),
    plans_dirs: glob("plans/*/"),
    plans_readmes: glob("plans/*/README.md")
  },

  extracted_metadata: {
    claude_links: extract_all_links(core_files.claude),
    readme_links: extract_all_links(core_files.readme),
    plan_phases: extract_phase_info(core_files.plan),
    principles_constraints: extract_constraints(core_files.principles)
  }
}

extract :: CoreDocuments → ReferenceMapping
extract(C) = {
  claude_references: {
    quick_links: parse_section(C.core_files.claude, "Quick Links"),
    faq_references: parse_section(C.core_files.claude, "FAQ"),
    command_references: parse_section(C.core_files.claude, "Development Commands"),
    doc_references: extract_markdown_links(C.core_files.claude)
  },

  plan_references: {
    phases: parse_phases(C.core_files.plan) where {
      phase_number: integer,
      phase_name: string,
      status: "🚧 In Progress" | "✅ Complete" | "📋 Planned",
      stages: list[stage_info],
      dependencies: list[phase_number]
    },
    plans_dir_refs: extract_directory_references(C.core_files.plan, pattern="plans/\\d+")
  },

  principles_references: {
    constraints: parse_constraints(C.core_files.principles) where {
      code_limits: {phase: integer, stage: integer},
      test_coverage: percentage,
      methodology: list[string]
    },
    architecture_refs: extract_links_to(C.core_files.principles, pattern="architecture/")
  },

  readme_references: {
    doc_links: extract_markdown_links(C.core_files.readme),
    getting_started: parse_section(C.core_files.readme, "Getting Started"),
    quick_links: parse_section(C.core_files.readme, "Quick Links") if exists
  },

  doc_map_inventory: if C.core_files.doc_map then
    extract_documented_files(C.core_files.doc_map)
  else null
}

validate :: ReferenceMapping → ValidationResults
validate(R) = {
  claude_to_plan: {
    quick_links_valid: for each link in R.claude_references.quick_links:
      verify_link_target(link, "docs/core/plan.md"),

    phase_references_match: for each phase_ref in extract_phase_mentions(claude):
      verify_phase_exists_in_plan(phase_ref, R.plan_references.phases),

    current_status_sync: compare(
      extract_current_status(claude),
      get_latest_phase_status(R.plan_references.phases)
    )
  },

  claude_to_principles: {
    code_limit_consistency: compare(
      extract_code_limits(claude),
      R.principles_references.constraints.code_limits
    ),

    test_coverage_consistency: compare(
      extract_test_requirements(claude),
      R.principles_references.constraints.test_coverage
    ),

    methodology_references: for each methodology in extract_methodologies(claude):
      verify_referenced_in_principles(methodology, principles)
  },

  plan_to_plans_dirs: {
    phase_directory_sync: for each phase in R.plan_references.phases:
      check_directory_exists(phase.number, actual_structure.plans_dirs),

    phase_status_consistency: for each phase in R.plan_references.phases:
      if directory_exists(phase.number) then
        compare_status(
          phase.status_in_plan,
          read_status_from_directory(phase.number)
        )
  },

  plan_to_principles: {
    constraints_referenced: verify_all_constraints_mentioned(
      R.plan_references,
      R.principles_references.constraints
    ),

    dependency_validation: for each phase in R.plan_references.phases:
      verify_dependencies_respect_principles(phase.dependencies)
  },

  readme_to_core_docs: {
    quick_links_valid: if R.readme_references.quick_links then
      validate_links_point_to_core_docs(R.readme_references.quick_links),

    getting_started_current: verify_getting_started_matches_plan(
      R.readme_references.getting_started,
      R.plan_references.phases[0]
    )
  },

  doc_map_completeness: if R.doc_map_inventory then {
    all_docs_listed: for each doc in actual_structure.docs_files:
      verify_listed_in_map(doc, R.doc_map_inventory),

    no_phantom_entries: for each entry in R.doc_map_inventory:
      verify_file_exists(entry, actual_structure.docs_files)
  } else null
}

classify :: ValidationResults → IssueReport
classify(V) = {
  by_severity: {
    critical: [
      V.claude_to_plan.quick_links_valid where broken,
      V.claude_to_principles.code_limit_consistency where mismatch,
      V.plan_to_plans_dirs.phase_directory_sync where missing
    ],

    high: [
      V.claude_to_plan.phase_references_match where mismatch,
      V.plan_to_plans_dirs.phase_status_consistency where inconsistent,
      V.claude_to_principles.test_coverage_consistency where mismatch,
      V.readme_to_core_docs.quick_links_valid where broken
    ],

    medium: [
      V.claude_to_plan.current_status_sync where outdated,
      V.claude_to_principles.methodology_references where missing,
      V.doc_map_completeness.all_docs_listed where missing
    ],

    low: [
      V.doc_map_completeness.no_phantom_entries where phantom,
      V.plan_to_principles.constraints_referenced where not_mentioned
    ]
  },

  by_document_pair: {
    "CLAUDE.md ↔ plan.md": group_issues([
      V.claude_to_plan.quick_links_valid,
      V.claude_to_plan.phase_references_match,
      V.claude_to_plan.current_status_sync
    ]),

    "CLAUDE.md ↔ principles.md": group_issues([
      V.claude_to_principles.code_limit_consistency,
      V.claude_to_principles.test_coverage_consistency,
      V.claude_to_principles.methodology_references
    ]),

    "plan.md ↔ plans/*/": group_issues([
      V.plan_to_plans_dirs.phase_directory_sync,
      V.plan_to_plans_dirs.phase_status_consistency
    ]),

    "plan.md ↔ principles.md": group_issues([
      V.plan_to_principles.constraints_referenced,
      V.plan_to_principles.dependency_validation
    ]),

    "README.md ↔ core docs": group_issues([
      V.readme_to_core_docs.quick_links_valid,
      V.readme_to_core_docs.getting_started_current
    ]),

    "DOCUMENTATION_MAP.md ↔ actual files": if V.doc_map_completeness then
      group_issues([
        V.doc_map_completeness.all_docs_listed,
        V.doc_map_completeness.no_phantom_entries
      ])
    else null
  },

  by_issue_type: {
    broken_cross_references: filter(all_issues, type="broken_link"),
    content_mismatches: filter(all_issues, type="mismatch"),
    missing_synchronization: filter(all_issues, type="outdated"),
    structural_inconsistencies: filter(all_issues, type="missing_entity")
  }
}

output :: IssueReport → Report
output(I) = {
  executive_summary: {
    total_issues: count(I.by_severity.*),
    critical_issues: count(I.by_severity.critical),
    sync_health_score: (1 - total_issues / total_checks) * 100,
    pre_merge_status: if count(I.by_severity.critical) == 0 then
      "✅ SAFE TO MERGE - Core docs are synchronized"
    else
      "❌ BLOCKED - Critical sync issues must be fixed"
  },

  critical_issues: if count(I.by_severity.critical) > 0 then {
    title: "🚨 CRITICAL: Core Documentation Out of Sync",
    description: "These must be fixed before merging or releasing",
    issues: for each issue in I.by_severity.critical:
      format_sync_issue(issue) where format = "{doc1} ↔ {doc2}: {description} ({location})"
  } else null,

  high_priority_issues: if count(I.by_severity.high) > 0 then {
    title: "⚠️  HIGH PRIORITY: Documentation Inconsistencies",
    description: "Address these before next release",
    issues: format_issues(I.by_severity.high)
  } else null,

  medium_priority_issues: if count(I.by_severity.medium) > 0 then {
    title: "⚡ MEDIUM: Minor Synchronization Issues",
    description: "Consider fixing during next documentation update",
    issues: format_issues(I.by_severity.medium)
  } else null,

  by_document_pair_report: {
    title: "Synchronization Status by Document Pair",
    format: "table",
    data: for each (pair, issues) in I.by_document_pair:
      {
        document_pair: pair,
        issues_found: count(issues),
        critical: count(filter(issues, severity="critical")),
        status: if count(issues) == 0 then "✅ Synchronized" else "❌ Issues found"
      }
  },

  detailed_issues: {
    broken_cross_references: {
      count: count(I.by_issue_type.broken_cross_references),
      examples: sample(I.by_issue_type.broken_cross_references, 5),
      common_causes: [
        "File moved without updating references",
        "Section renamed without updating Quick Links",
        "Phase number changed without updating cross-references"
      ]
    },

    content_mismatches: {
      count: count(I.by_issue_type.content_mismatches),
      examples: sample(I.by_issue_type.content_mismatches, 5),
      common_causes: [
        "Code limits updated in principles.md but not CLAUDE.md",
        "Test coverage changed without updating FAQ",
        "Phase status updated in plan.md but not plans/N/README.md"
      ]
    },

    missing_synchronization: {
      count: count(I.by_issue_type.missing_synchronization),
      examples: sample(I.by_issue_type.missing_synchronization, 5),
      common_causes: [
        "Current status outdated in CLAUDE.md",
        "README.md doesn't reflect latest getting started steps",
        "DOCUMENTATION_MAP.md missing newly added files"
      ]
    }
  },

  sync_checklist: {
    title: "Synchronization Checklist",
    description: "Follow this order to fix sync issues efficiently",
    steps: prioritize_sync_order([
      {step: "Fix broken cross-references in CLAUDE.md Quick Links", priority: 1},
      {step: "Update code limits and constraints consistency", priority: 2},
      {step: "Sync phase status between plan.md and plans/N/ directories", priority: 3},
      {step: "Update CLAUDE.md current status to match latest phase", priority: 4},
      {step: "Refresh DOCUMENTATION_MAP.md to include all docs", priority: 5}
    ])
  },

  recommendations: {
    immediate_actions: generate_fix_commands(I.by_severity.critical),

    pre_merge_workflow: if count(I.by_severity.critical) > 0 then
      "Run '/meta doc-sync' before every merge to catch sync issues",

    pre_release_validation: [
      "Verify all core docs are synchronized",
      "Check phase status consistency",
      "Update DOCUMENTATION_MAP.md",
      "Run '/meta doc-links' to ensure link integrity"
    ],

    automation: "Consider adding doc-sync check to pre-commit or CI/CD pipeline",

    periodic_review: "Run '/meta doc-sync' after completing each project phase"
  }
} where ¬execute(fix_commands)

implementation_notes:
- read CLAUDE.md, README.md, docs/core/plan.md, docs/core/principles.md using Read tool
- parse markdown to extract Quick Links, FAQ sections, phase tables
- use regex patterns to find phase references (e.g., "Phase N", "plans/N/")
- glob docs/ and plans/ directories to verify actual structure
- compare extracted metadata between documents for consistency
- classify issues by severity based on impact (critical = broken links in core docs)
- provide actionable recommendations with specific file:section references

constraints:
- comprehensive: check all core document pairs
- fast: complete validation in <15s for typical project
- actionable: provide file:section:issue format for easy fixing
- severity_aware: Critical (broken refs) > High (mismatches) > Medium (outdated) > Low (missing mentions)
- pre_merge_safe: only critical issues block merge
- non_destructive: read-only analysis, no automatic fixes
- ordered_recommendations: suggest fix order (critical first, then dependent issues)
