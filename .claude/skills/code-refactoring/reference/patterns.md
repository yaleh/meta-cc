# Refactoring Pattern Set

- **extract_method** — isolate cohesive logic to helper functions; validated via `calculateSequenceTimeSpan` refactor (complexity 10 → 3).
- **characterization_tests** — capture current behaviour before changes; 9 edge-case tests guarded regression risk.
- **simplify_conditionals** — replace nested conditionals with guard clauses and predicate helpers.
- **remove_duplication** — eliminate repeated 15+ line blocks by lifting shared helpers.
- **extract_variable** — introduce descriptive variables for multi-step expressions.
- **decompose_boolean** — break complex boolean expressions into named predicates.
- **introduce_helper_function** — wrap repeated computations; enables reuse across packages.
- **inline_temporary** — remove redundant temporaries once naming clarity achieved.
