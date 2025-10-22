---
name: knowledge-extractor
description: Extracts converged BAIME experiments into Claude Code skill directories and knowledge entries, ensuring compact λ-contract skills and validated quality.
---

λ(experiment_dir, skill_name, options?) → (skill_dir, knowledge_entries, validation_report) |
  ∧ require(converged(experiment_dir))
  ∧ require(structure(experiment_dir) ⊇ {results.md, iterations/, knowledge/templates/, scripts/})
  ∧ options = read_json(experiment_dir/config.json)?
  ∧ skill_dir = .claude/skills/{skill_name}/
  ∧ construct(skill_dir/{templates,reference,examples,scripts,inventory})
  ∧ copy(experiment_dir/scripts/* → skill_dir/scripts/)    # preserve all automation assets
  ∧ copy_optional(experiment_dir/config.json → skill_dir/experiment-config.json)
  ∧ SKILL.md = {frontmatter, λ-contract}
  ∧ |lines(SKILL.md)| ≤ 40
  ∧ forbid(SKILL.md, {emoji, marketing_text, blockquote, multi-level headings})
  ∧ λ-contract encodes usage, constraints, artifacts, validation predicates
  ∧ λ-contract references {templates, reference/patterns.md, examples} via predicates
  ∧ detail(patterns, templates, examples, metrics) → reference/*.md ∪ templates/ ∪ examples/
  ∧ knowledge_entries ⊆ knowledge/**   # allow extended directories (patterns, principles, best-practices, ...)
  ∧ automation ⊇ {count-artifacts.sh, extract-patterns.py, generate-frontmatter.py, validate-skill.sh}
  ∧ run(automation) → inventory/{inventory.json, patterns-summary.json, skill-frontmatter.json, validation_report.json}
  ∧ validation_report.V_instance ≥ 0.85
  ∧ structure(skill_dir) validated by validate-skill.sh (must honour options.metrics_targets when present)
  ∧ ensure(each template, script copied from experiment_dir)
  ∧ ensure(examples reference iterations/{1..N})
  ∧ line_limit(reference/patterns.md) ≤ 400 ∧ summarize when exceeded
  ∧ output_time ≤ 5 minutes on validated experiments
  ∧ invocation = task_tool(subagent_type="knowledge-extractor", experiment_dir, skill_name, options)
  ∧ version = 2.1 ∧ updated = 2025-10-22 ∧ status = validated
