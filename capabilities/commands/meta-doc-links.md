---
name: meta-doc-links
description: Validate documentation link integrity with broken link detection and pre-commit safety checks.
keywords: documentation, links, validation, integrity, broken-links, markdown, pre-commit
category: diagnostics
---

λ(scope) → link_validation_report | ∀link ∈ {markdown_internal_links}:

scope :: project

analyze :: Project → LinkValidation
analyze(P) = collect(docs) ∧ extract(links) ∧ validate(targets) ∧ classify(issues)

collect :: Project → DocumentInventory
collect(P) = {
  core_docs: [
    "CLAUDE.md",
    "README.md",
    "docs/core/plan.md",
    "docs/core/principles.md",
    "docs/DOCUMENTATION_MAP.md"
  ],

  category_docs: {
    guides: glob("docs/guides/*.md"),
    reference: glob("docs/reference/*.md"),
    tutorials: glob("docs/tutorials/*.md"),
    architecture: glob("docs/architecture/**/*.md"),
    methodology: glob("docs/methodology/*.md")
  },

  plans: glob("plans/*/*.md"),

  all_markdown_files: flatten([core_docs, category_docs, plans])
}

extract :: DocumentInventory → LinkCollection
extract(D) = {
  for each file in D.all_markdown_files:
    links: parse_markdown_links(file) where {
      internal_relative: matches("^\\.\\.?/.*\\.md"),
      internal_absolute: matches("^/.*\\.md"),
      anchors: matches("#[a-z0-9-]+"),
      file_with_anchor: matches("\\.md#[a-z0-9-]+")
    },

  link_index: {
    source_file: file,
    target_path: resolve_path(link.href),
    anchor: extract_anchor(link.href),
    line_number: link.line,
    link_text: link.text
  }
}

validate :: LinkCollection → ValidationResults
validate(L) = {
  for each link in L.link_index:
    file_exists: check_file_existence(link.target_path),

    anchor_valid: if link.anchor then
      check_anchor_exists(link.target_path, link.anchor)
    else true,

    status: classify_link_status(file_exists, anchor_valid)
}

classify :: ValidationResults → IssueReport
classify(V) = {
  by_severity: {
    critical: broken_links where source_file in [
      "CLAUDE.md",
      "README.md",
      "docs/core/plan.md",
      "docs/core/principles.md"
    ],

    high: broken_links where source_file in [
      "docs/DOCUMENTATION_MAP.md",
      "docs/guides/*",
      "docs/reference/*"
    ],

    medium: broken_links where source_file in [
      "docs/tutorials/*",
      "docs/architecture/*",
      "plans/*"
    ],

    low: broken_links where source_file in [
      "docs/methodology/*",
      "docs/archive/*"
    ]
  },

  by_issue_type: {
    file_not_found: filter(V, file_exists == false),
    invalid_anchor: filter(V, file_exists == true ∧ anchor_valid == false),
    malformed_path: filter(V, parse_error == true)
  },

  by_source_document: group_by(V.broken_links, source_file)
}

output :: IssueReport → Report
output(I) = {
  executive_summary: {
    total_files_scanned: count(all_markdown_files),
    total_links_checked: count(link_index),
    broken_links_found: count(I.by_severity.*),
    link_health_score: (valid_links / total_links) * 100,
    pre_commit_status: if count(I.by_severity.critical) == 0 then
      "✅ SAFE TO COMMIT - No critical link issues"
    else
      "❌ BLOCKED - Fix critical links before committing"
  },

  critical_issues: if count(I.by_severity.critical) > 0 then {
    title: "🚨 CRITICAL: Broken Links in Core Documents",
    description: "These links MUST be fixed before committing",
    issues: for each link in I.by_severity.critical:
      format_issue(link) where format = "{source_file}:{line_number} → {target_path} ({issue_type})"
  } else null,

  high_priority_issues: if count(I.by_severity.high) > 0 then {
    title: "⚠️  HIGH PRIORITY: Broken Links in Primary Documentation",
    description: "Fix these before merging or releasing",
    issues: format_issues(I.by_severity.high)
  } else null,

  medium_priority_issues: if count(I.by_severity.medium) > 0 then {
    title: "⚡ MEDIUM: Broken Links in Secondary Documentation",
    description: "Consider fixing during next documentation update",
    issues: format_issues(I.by_severity.medium),
    note: "These are acceptable for commits but should be addressed"
  } else null,

  low_priority_issues: if count(I.by_severity.low) > 0 then {
    title: "ℹ️  LOW: Broken Links in Archive/Methodology",
    description: "Optional fixes - low user impact",
    issues: format_issues(I.by_severity.low)
  } else null,

  issue_breakdown: {
    file_not_found: {
      count: count(I.by_issue_type.file_not_found),
      examples: sample(I.by_issue_type.file_not_found, 5),
      common_causes: [
        "File moved without updating links",
        "Typo in file path",
        "Documentation restructuring"
      ]
    },

    invalid_anchors: {
      count: count(I.by_issue_type.invalid_anchor),
      examples: sample(I.by_issue_type.invalid_anchor, 5),
      common_causes: [
        "Heading renamed without updating anchor",
        "Case sensitivity mismatch",
        "Anchor format incorrect"
      ]
    }
  },

  by_document_report: {
    title: "Links by Source Document",
    format: "table",
    data: for each (file, links) in I.by_source_document:
      {
        document: file,
        total_links: count(links),
        broken: count(filter(links, status == "broken")),
        status: if broken == 0 then "✅" else "❌"
      }
  },

  recommendations: {
    immediate_actions: generate_fix_commands(I.by_severity.critical),

    pre_commit_hook: if count(I.by_severity.critical) > 0 then
      "Consider running '/meta doc-links' before each commit",

    automation: "Add this check to CI/CD pipeline to catch issues early",

    documentation_health: if link_health_score < 95 then
      "Link health below 95% - schedule documentation maintenance sprint"
  }
} where ¬execute(fix_commands)

implementation_notes:
- scan all .md files in project root, docs/, plans/ directories
- parse markdown to extract [text](path) and [text](path#anchor) patterns
- resolve relative paths (../, ./) to absolute paths for validation
- check file existence using file system operations
- validate anchors by parsing target markdown headings
- use Read tool to access file contents for anchor validation
- use Glob tool to discover all markdown files
- prioritize by document criticality (core > guides > plans > archive)

constraints:
- fast: complete scan in <10s for typical project (~50 markdown files)
- actionable: provide file:line:target format for easy fixing
- severity_aware: Critical (core docs) > High (guides) > Medium (plans) > Low (archive)
- pre_commit_safe: only critical issues block commits
- comprehensive: check all internal .md links, ignore external URLs
- non_destructive: read-only analysis, no automatic fixes
