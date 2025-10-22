# Refactoring Pattern Set

- **builder_map_decomposition** — replace large dispatcher switches with map-driven builder functions (iteration-1.md).
- **pipeline_config_struct** — centralize shared parameter extraction into immutable config structs to shrink orchestration functions (iteration-1.md).
- **helper_specialization** — move logging/metrics branches into dedicated helpers to keep main flow linear (iteration-1.md).
- **jq_pipeline_segmentation** — split parsing, execution, and encoding into helpers to reduce panic surfaces and simplify testing (iteration-2.md).
- **automation_first_metrics** — codify scripts/make targets for complexity & coverage snapshots; treat metrics as part of the refactor (iteration-2.md, iteration-3.md).
- **documentation_templates** — generate iteration logs from templates to maintain V_meta_completeness ≥ 0.8 (iteration-3.md).
- **conversation_turn_builder** — extract user/assistant maps and assemble turns via helper orchestration to keep conversation queries readable (cli iteration-3.md).
- **prompt_outcome_analyzer** — split outcome scanning into confirmation/error/deliverable helpers to evaluate user prompts predictably (cli iteration-3.md).
