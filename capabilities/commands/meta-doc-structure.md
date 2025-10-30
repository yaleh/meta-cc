---
name: meta-doc-structure
description: Validate documentation structure compliance with DRY principles, Progressive Disclosure, Task-Oriented organization, and size guidelines.
keywords: documentation, structure, validation, DRY, progressive-disclosure, task-oriented, size-limits
category: diagnostics
---

λ(scope) → structure_validation_report | ∀doc_structure ∈ {project_documentation}:

scope :: project

analyze :: Project → StructureValidation
analyze(P) = collect(structure) ∧ validate(dry_principle) ∧ validate(progressive_disclosure) ∧ validate(task_organization) ∧ validate(size_guidelines) ∧ classify(issues)

collect :: Project → DocumentationStructure
collect(P) = {
  core_docs: {
    claude: read("CLAUDE.md"),
    readme: read("README.md"),
    plan: read("docs/core/plan.md"),
    principles: read("docs/core/principles.md"),
    doc_map: read("docs/DOCUMENTATION_MAP.md") if exists
  },

  directory_structure: {
    root_docs: glob("*.md"),
    docs_dirs: glob("docs/*/"),
    guides_structure: glob("docs/guides/**/*.md"),
    reference_structure: glob("docs/reference/**/*.md"),
    tutorials_structure: glob("docs/tutorials/**/*.md"),
    core_docs: glob("docs/core/**/*.md")
  },

  doc_metadata: {
    file_sizes: for each file in glob("**/*.md"): get_file_size(file),
    file_line_counts: for each file in glob("**/*.md"): wc -l file,
    last_modified: for each file in glob("**/*.md"): stat -c %Y file
  }
}

validate_dry_principle :: DocumentationStructure → DRYViolations
validate_dry_principle(D) = {
  # Detect duplicate content across documentation files
  content_hashes: for each file in glob("**/*.md"): {
    file_path: file,
    content_hash: hash(strip_whitespace(read(file)))
  },

  duplicates: group_by(content_hashes, content_hash) where count > 1,

  cross_references: {
    # Check if content is referenced in multiple places without proper linking
    for each file in glob("**/*.md"):
      extract_key_concepts(file) → check_if_mentioned_elsewhere_without_linking
  }
}

validate_progressive_disclosure :: DocumentationStructure → DisclosureViolations
validate_progressive_disclosure(D) = {
  # Validate README → guides → reference hierarchy
  readme_content: read("README.md"),
  guides_exist: exists("docs/guides/"),
  reference_exists: exists("docs/reference/"),

  hierarchy_check: {
    # README should link to guides
    readme_links_to_guides: contains_links_to(readme_content, "docs/guides/"),

    # Guides should link to reference
    guides_link_to_reference: for each guide in glob("docs/guides/**/*.md"):
      contains_links_to(read(guide), "docs/reference/"),

    # Reference should be comprehensive but not duplicated in guides
    reference_completeness: check_reference_completeness("docs/reference/")
  },

  content_depth_check: {
    # README should be high-level overview
    readme_line_limit: line_count("README.md") <= 350,

    # Guide files should be detailed but not too long
    guide_size_check: for each guide in glob("docs/guides/**/*.md"):
      line_count(guide) <= 800,

    # Reference files can be longer but should be well-structured
    reference_size_check: for each ref in glob("docs/reference/**/*.md"):
      line_count(ref) <= 1200
  }
}

validate_task_organization :: DocumentationStructure → OrganizationViolations
validate_task_organization(D) = {
  # Check if docs/guides/ structure follows task-oriented organization
  guides_dirs: glob("docs/guides/*/"),

  task_oriented_check: {
    # Directory names should reflect tasks or user goals
    meaningful_names: for each dir in guides_dirs:
      is_task_oriented_name(dir),

    # Each guide directory should have clear purpose
    purpose_documentation: for each dir in guides_dirs:
      exists(join(dir, "README.md")) or
      has_clear_purpose_in_directory_name(dir)
  },

  # Check for orphaned or unlinked documentation
  orphaned_docs: {
    all_docs: glob("docs/**/*.md"),
    linked_docs: extract_all_links_from(glob("**/*.md")),
    unlinked: all_docs - linked_docs
  }
}

validate_size_guidelines :: DocumentationStructure → SizeViolations
validate_size_guidelines(D) = {
  # Validate document sizes per methodology guidelines
  methodology_limits: {
    context_base: 300,      # CLAUDE.md, README.md
    living_doc: 600,        # plan.md, frequently updated docs
    specification: "∞",     # ADRs, stable references
    reference: 800,         # API references, command docs
    episodic: "∞",          # Phase docs, temporary docs
    archive: "∞"            # Historical docs
  },

  size_violations: {
    oversized_docs: for each file in glob("**/*.md"):
      if line_count(file) > get_size_limit_for_doc_type(file):
        {
          file: file,
          actual_lines: line_count(file),
          limit: get_size_limit_for_doc_type(file),
          suggested_action: suggest_refactoring(file)
        }
  },

  size_distribution: {
    by_type: group_docs_by_type(glob("**/*.md")),
    avg_size_by_type: for each type in by_type:
      avg(line_count for doc in type)
  }
}

classify :: ValidationResults → IssueReport
classify(V) = {
  by_severity: {
    critical: [
      V.dry_principle.duplicates where similarity > 0.9,
      V.size_guidelines.size_violations where excess > 200%,
      V.task_organization.orphaned_docs where count > 10
    ],

    high: [
      V.progressive_disclosure.hierarchy_check where missing_links,
      V.dry_principle.duplicates where similarity > 0.7,
      V.size_guidelines.size_violations where excess > 100%
    ],

    medium: [
      V.progressive_disclosure.content_depth_check where limits_exceeded,
      V.task_organization.task_oriented_check where names_not_meaningful,
      V.size_guidelines.size_violations where excess > 50%
    ],

    low: [
      V.dry_principle.cross_references where missing_links,
      V.progressive_disclosure.content_depth_check where near_limits,
      V.task_organization.orphaned_docs where count <= 10
    ]
  },

  by_principle: {
    "DRY Principle": group_issues(V.dry_principle.*),
    "Progressive Disclosure": group_issues(V.progressive_disclosure.*),
    "Task-Oriented Organization": group_issues(V.task_organization.*),
    "Size Guidelines": group_issues(V.size_guidelines.*)
  }
}

output :: IssueReport → Report
output(I) = {
  executive_summary: {
    total_issues: count(I.by_severity.*),
    critical_issues: count(I.by_severity.critical),
    structure_health_score: (1 - total_issues / total_checks) * 100,
    compliance_status: if count(I.by_severity.critical) == 0 then
      "✅ STRUCTURE COMPLIANT - Documentation follows methodology guidelines"
    else
      "❌ NON-COMPLIANT - Critical structure issues must be fixed"
  },

  critical_issues: if count(I.by_severity.critical) > 0 then {
    title: "🚨 CRITICAL: Documentation Structure Violations",
    description: "These must be fixed to comply with documentation methodology",
    issues: for each issue in I.by_severity.critical:
      format_structure_issue(issue) where format = "{principle}: {description} ({file}:{lines})"
  } else null,

  high_priority_issues: if count(I.by_severity.high) > 0 then {
    title: "⚠️  HIGH PRIORITY: Structure Improvements Needed",
    description: "Address these to improve documentation quality",
    issues: format_issues(I.by_severity.high)
  } else null,

  medium_priority_issues: if count(I.by_severity.medium) > 0 then {
    title: "⚡ MEDIUM: Structure Optimization Opportunities",
    description: "Consider addressing during next documentation update",
    issues: format_issues(I.by_severity.medium)
  } else null,

  by_principle_report: {
    title: "Structure Validation by Principle",
    format: "table",
    data: for each (principle, issues) in I.by_principle:
      {
        principle: principle,
        issues_found: count(issues),
        critical: count(filter(issues, severity="critical")),
        status: if count(issues) == 0 then "✅ Compliant" else "❌ Issues found"
      }
  },

  detailed_analysis: {
    duplicate_content: {
      count: count(I.by_principle."DRY Principle" where type="duplicate"),
      examples: sample(I.by_principle."DRY Principle" where type="duplicate", 3),
      recommendation: "Use cross-references instead of duplicating content"
    },

    hierarchy_violations: {
      count: count(I.by_principle."Progressive Disclosure" where type="hierarchy"),
      examples: sample(I.by_principle."Progressive Disclosure" where type="hierarchy", 3),
      recommendation: "Follow README → guides → reference structure"
    },

    task_organization_issues: {
      count: count(I.by_principle."Task-Oriented Organization" where type="organization"),
      examples: sample(I.by_principle."Task-Oriented Organization" where type="organization", 3),
      recommendation: "Organize guides around user tasks, not features"
    },

    size_violations: {
      count: count(I.by_principle."Size Guidelines" where type="size"),
      examples: sample(I.by_principle."Size Guidelines" where type="size", 3),
      recommendation: "Split large documents or refactor content"
    }
  },

  structure_checklist: {
    title: "Documentation Structure Checklist",
    description: "Follow this checklist to ensure compliance with methodology",
    steps: [
      {step: "Validate DRY principle - no duplicate content", priority: 1},
      {step: "Ensure Progressive Disclosure hierarchy (README → guides → reference)", priority: 2},
      {step: "Organize guides around user tasks, not features", priority: 3},
      {step: "Check document sizes against methodology limits", priority: 4},
      {step: "Verify all documents are properly linked", priority: 5}
    ]
  },

  recommendations: {
    immediate_actions: generate_fix_commands(I.by_severity.critical),

    quarterly_assessment: [
      "Run '/meta doc-structure' for documentation health check",
      "Review document sizes and refactor oversized documents",
      "Check for duplicate content and consolidate",
      "Verify Progressive Disclosure hierarchy compliance"
    ],

    automation: "Consider adding doc-structure check to CI/CD pipeline",

    periodic_review: "Run '/meta doc-structure' after major documentation updates"
  }
} where ¬execute(fix_commands)

implementation_notes:
- read documentation files using Read tool
- use glob to discover directory structure
- calculate file sizes and line counts
- detect duplicate content using hash comparison
- validate hierarchy by checking cross-references
- classify issues by severity based on impact
- provide actionable recommendations with specific file references

constraints:
- comprehensive: check all structure principles (DRY, Progressive Disclosure, Task-Oriented, Size)
- fast: complete validation in <20s for typical project
- actionable: provide file:line:issue format for easy fixing
- severity_aware: Critical (structural) > High (hierarchy) > Medium (organization) > Low (optimization)
- methodology_aligned: follows Documentation Management Methodology v5.0
- non_destructive: read-only analysis, no automatic fixes
- ordered_recommendations: suggest fix order (critical first, then dependent issues)
