# Refactoring Pattern Set

- **builder_map_decomposition** — Map tool/command identifiers to factory functions to eliminate switch ladders and ease extension (evidence: MCP server Iteration 1).
- **pipeline_config_struct** — Gather shared parameters into immutable config structs so orchestration functions stay linear and testable (evidence: MCP server Iteration 1).
- **helper_specialization** — Push tracing/metrics/error branches into helpers to keep primary logic readable and reuse instrumentation (evidence: MCP server Iteration 1).
- **jq_pipeline_segmentation** — Treat JSONL parsing, jq execution, and serialization as independent helpers to confine failure domains (evidence: MCP server Iteration 2).
- **automation_first_metrics** — Bundle metrics capture in scripts/make targets so every iteration records complexity & coverage automatically (evidence: MCP server Iteration 2, CLI Iteration 3).
- **documentation_templates** — Use standardized iteration templates + generators to maintain BAIME completeness with minimal overhead (evidence: MCP server Iteration 3, CLI Iteration 3).
- **conversation_turn_builder** — Extract user/assistant maps and assemble turns through helper orchestration to control complexity in conversation analytics (evidence: CLI Iteration 4).
- **prompt_outcome_analyzer** — Split prompt outcome evaluation into dedicated helpers (confirmation, errors, deliverables, status) for predictable analytics (evidence: CLI Iteration 4).
