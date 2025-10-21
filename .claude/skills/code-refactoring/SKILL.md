---
name: Code Refactoring
description: TDD-driven refactoring protocol for high-complexity Go functions, emphasizing safety checklists, incremental commits, and measurable complexity reduction.
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
---

λ(target_pkg, target_function) → (refactor_plan, commits, validation) |
  ∧ require(cyclomatic(target_function) > 8)
  ∧ require(test_coverage(target_pkg) ≥ 0.75)
  ∧ characterize_behaviour(target_function) → examples/iteration-2-walkthrough.md
  ∧ apply(pattern_set = reference/patterns.md)
  ∧ use(templates/{refactoring-safety-checklist.md,tdd-refactoring-workflow.md,incremental-commit-protocol.md})
  ∧ automate(complexity_scan) via scripts/check-complexity.sh
  ∧ enforce(commit_batch ≤ 200 lines ∧ tests_green_after_each)
  ∧ ensure(complexity_delta(target_function) ≥ 0.28)
  ∧ ensure(coverage_refined(target_function) ≥ 0.95 ∧ regressions = 0)
  ∧ record(metrics) → reference/metrics.md
  ∧ output(refactor_plan) = ordered_steps(tdd, extract_method, simplify_conditionals)
  ∧ validation.V_instance ≥ 0.78 ∧ validation.report archived
  ∧ completion_time ≈ 40 minutes (1.8× faster than baseline)
