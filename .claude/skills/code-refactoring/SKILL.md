---
name: Code Refactoring
description: BAIME-aligned refactoring protocol for Go hotspots with automated metrics and documentation.
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
---

λ(target_pkg, target_hotspot) → (refactor_plan, metrics_snapshot, validation_report) |
  ∧ require(cyclomatic(target_hotspot) > 8)
  ∧ require(run("make metrics-mcp"))
  ∧ baseline = results.md ∧ iterations/
  ∧ apply(pattern_set = reference/patterns.md)
  ∧ use(templates/{iteration-template.md,refactoring-safety-checklist.md,tdd-refactoring-workflow.md,incremental-commit-protocol.md})
  ∧ automate(metrics_snapshot) via scripts/{capture-mcp-metrics.sh,count-artifacts.sh}
  ∧ document(knowledge) → knowledge/{patterns,principles,best-practices}
  ∧ ensure(complexity_delta(target_hotspot) ≥ 0.30 ∧ cyclomatic(target_hotspot) ≤ 10)
  ∧ ensure(coverage_delta(target_pkg) ≥ 0.01 ∨ coverage(target_pkg) ≥ 0.70)
  ∧ validation_report = validate-skill.sh → {inventory.json, V_instance ≥ 0.85}
