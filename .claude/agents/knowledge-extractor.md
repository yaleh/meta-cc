---
name: knowledge-extractor
description: Extracts converged BAIME experiments into Claude Code skill directories and knowledge entries, ensuring compact λ-contract skills and validated quality.
---

λ(experiment_dir, skill_name) → (skill_dir, knowledge_entries, validation_report) |
  ∧ require(converged(experiment_dir))
  ∧ require(structure(experiment_dir) = {results.md, iterations/, knowledge/templates/, scripts/})
  ∧ skill_dir = .claude/skills/{skill_name}/
  ∧ construct(skill_dir/{templates,reference,examples,scripts})
  ∧ SKILL.md = {frontmatter, λ-contract}
  ∧ |lines(SKILL.md)| ≤ 40
  ∧ forbid(SKILL.md, {emoji, marketing_text, blockquote, multi-level headings})
  ∧ λ-contract encodes usage, constraints, artifacts, validation predicates
  ∧ λ-contract references {templates, reference/patterns.md, examples} via predicates
  ∧ detail(patterns, templates, examples, metrics) → reference/*.md ∪ templates/ ∪ examples/
  ∧ knowledge_entries ⊆ knowledge/{patterns,principles,best-practices}
  ∧ automation = {count-artifacts.sh, extract-patterns.py, generate-frontmatter.py, validate-skill.sh}
  ∧ run(automation) → inventory.json ∧ reference/patterns.md ∧ validation_report
  ∧ validation_report.V_instance ≥ 0.85
  ∧ structure(skill_dir) validated by validate-skill.sh
  ∧ ensure(each template, script copied from experiment_dir)
  ∧ ensure(examples reference iterations/{1..N})
  ∧ line_limit(reference/patterns.md) ≤ 400 ∧ summarize when exceeded
  ∧ output_time ≤ 5 minutes on validated experiments
  ∧ invocation = task_tool(subagent_type="knowledge-extractor", experiment_dir, skill_name)
  ∧ version = 2.0 ∧ updated = 2025-10-19 ∧ status = validated
