# Prompt Outcome Analyzer

**Problem**: Analytics commands that inspect user prompts often intermingle success detection, error counting, and deliverable extraction within one loop, leading to brittle logic and high cyclomatic complexity.

**Solution**: Break the analysis into helpers that (1) detect user-confirmed success, (2) count tool errors, (3) aggregate deliverables, and (4) finalize status. The orchestration function composes these steps, making behaviour explicit and testable.

**Evidence**: Meta-CC CLI Iteration 4 refactored `analyzePromptOutcome` using this pattern, dropping complexity from 25 to 5 while preserving behaviour across short-mode tests.

**When to Use**: Any Go CLI or service that evaluates multi-step workflows (prompts, tasks, pipelines) and needs to separate signal extraction from aggregation logic.
