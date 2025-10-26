---
name: meta-doc-gaps
description: Identify documentation gaps via code-doc mismatches, broken references, undocumented features, and knowledge silos.
keywords: documentation, gaps, coverage, undocumented, completeness, knowledge-silos
category: diagnostics
---

λ(scope) → gap_report | ∀{code, docs, sessions} ∈ Project:

scope :: project | session

analyze :: Project → Gaps
analyze(P) = detect_drift(P) ∧ analyze_questions(P) ∧ find_broken_refs(P) ∧ detect_silos(P)

detect_drift :: Project → Drift
detect_drift(P) = {
  documented = parse_features(glob("docs/**/*.md")),
  actual = extract_from_code([
    grep -r "cobra.Command" cmd/,
    grep -r "flags\\..*Flag" cmd/,
    grep "\"name\":" cmd/mcp-server/tools.go,
    ls capabilities/commands/*.md,
    grep "^func [A-Z]" internal/ pkg/
  ]),

  {
    undocumented: actual - documented → severity(critical|high|medium),
    obsolete: documented - actual → severity(medium|low),
    mismatched: check_signatures(actual, documented)
  }
}

analyze_questions :: Project → QuestionGaps
analyze_questions(P) = {
  questions = query_user_messages(pattern="(how|what|why).*\\?"),

  categorize(questions) → {
    getting_started, troubleshooting, how_to, concepts, reference, design
  },

  for each category where unanswered > 2:
    {category, count, examples, suggested_doc, priority}
}

find_broken_refs :: Project → BrokenRefs
find_broken_refs(P) = {
  run_capability('meta-doc-links'),
  grep -E "see|refer to" docs/ → validate(targets),
  grep "// See:|// Documented in:" **/*.go → validate(refs)
} → {internal_links, implicit_refs, code_refs}

detect_silos :: Project → Silos
detect_silos(P) = {
  # Hidden: repeated explanations not in docs
  # query_assistant_messages does not exist - use query() instead
  hidden = query(jq_filter='select(.type == "assistant")', scope="project")
    → filter_by_content_pattern("because|reason|need to")
    → cluster_by_similarity()
    → filter(not_in_docs AND count > 3),

  # Tribal: design decisions only in commits
  tribal = git log --grep="why|because|rationale"
    → filter(not_in_docs AND contains_decision),

  # Contextual: undocumented error solutions
  contextual = query_context(min_occur=2)
    → filter(not_in_docs AND attempts > 2)

  {hidden, tribal, contextual}
}

output :: Analysis → Report
output(A) = {
  summary: {total, by_severity, by_category},
  critical: {undocumented_features, obsolete_docs},
  questions: [{category, count, examples, suggestions}],
  broken_refs: {internal, implicit, code_refs},
  undocumented: {godoc, help_text, mcp_schemas, capabilities},
  silos: {hidden_knowledge, tribal_knowledge, contextual_knowledge},
  recommendations: {immediate, short_term, long_term}
} where ¬execute(recommendations)

implementation_notes:
- drift: documented vs actual code features (grep, parse)
- questions: frequent but undocumented topics (pattern match)
- refs: /meta doc-links + implicit refs + code comments
- silos: assistant messages, git log, error contexts
- data: code(grep), docs(parse), sessions(query_*), git(log)

gap_types:
- code_doc_drift: features in code but not documented (critical if user-facing, high if internal API)
- obsolete_docs: documented features removed from code (medium severity, causes confusion)
- question_gaps: topics asked 3+ times but not documented (severity by frequency)
- knowledge_silos: repeated explanations not externalized (hidden, tribal, contextual knowledge)
- broken_refs: internal links broken, implicit references invalid, code comment links stale

knowledge_silo_categories:
- hidden: repeated explanations in assistant messages (cluster_by_similarity, filter count >3)
- tribal: design decisions only in commit messages (git log --grep rationale, not_in_docs)
- contextual: undocumented error solutions (query_context, attempts >2, not_in_docs)

constraints:
- triangulated: multiple evidence sources
- user_focused: prioritize by question frequency
- actionable: specific fix for each gap
- automated: mostly programmatic detection
