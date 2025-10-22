# Code Refactoring BAIME Results

## Experiment A — MCP Server (cmd/mcp-server)

| Iteration | Focus | V_instance | V_meta | Evidence |
|-----------|-------|------------|--------|----------|
| 0 | Baseline calibration | 0.42 | 0.18 | iterations/iteration-0.md |
| 1 | Executor command builder | 0.83 | 0.50 | iterations/iteration-1.md |
| 2 | JQ filter decomposition & metrics automation | 0.92 | 0.67 | iterations/iteration-2.md |
| 3 | Coverage & methodology integration | 0.93 | 0.80 | iterations/iteration-3.md |

**Convergence**: Iteration 3 (dual value ≥0.80).

Key assets:
- Metrics targets: `metrics-mcp`
- Automation scripts: `scripts/capture-mcp-metrics.sh`, `scripts/new-iteration-doc.sh`
- Patterns captured: builder map decomposition, pipeline config struct, helper specialization, jq pipeline segmentation

## Experiment B — CLI Refactor (cmd)

| Iteration | Focus | V_instance | V_meta | Evidence |
|-----------|-------|------------|--------|----------|
| 0 | Baseline & architecture survey | 0.36 | 0.22 | experiments/meta-cc-cli-refactor/iterations/iteration-0.md |
| 1 | Sandbox locator & harness | 0.70 | 0.46 | experiments/meta-cc-cli-refactor/iterations/iteration-1.md |
| 2 | Query pipeline staging | 0.74 | 0.58 | experiments/meta-cc-cli-refactor/iterations/iteration-2.md |
| 3 | Filter engine & validation subcommand | 0.77 | 0.72 | experiments/meta-cc-cli-refactor/iterations/iteration-3.md |
| 4 | Conversation & prompt modularization | 0.84 | 0.82 | experiments/meta-cc-cli-refactor/iterations/iteration-4.md |

**Convergence**: Iteration 4.

Key assets:
- Metrics targets: `metrics-cli`, `metrics-mcp`
- Automation scripts: `scripts/capture-cli-metrics.sh`
- New patterns: conversation turn pipeline, prompt outcome analyzer, documentation templates

Refer to `.claude/experiments/meta-cc-cli-refactor/` for CLI-specific iterations and `iterations/` for MCP server history.
