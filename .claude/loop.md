# Meta-cc Continuous Development Loop

Read `.quay/config.yml` for provider and gate configuration. Use the quay-native task board
at /home/yale/work/meta-cc/tasks/ as the canonical task store.

## Cycle

1. DRAIN: read pending directives (label:directive, status:todo), disposition each.
2. SELECT: pick the highest-priority ready task. Exclude label:human-steered.
3. ISOLATE: git worktree add off master HEAD for the selected task.
4. BUILD: implement the task in the worktree. TDD where applicable.
5. GATE: run configured gates. All must pass.
6. LAND: merge worktree to master, mark task done, record evidence.

Use the quay MCP tools (task_list, task_get, task_write) to interact with the task board.
Use the installed workflows (.claude/workflows/) for structured pipeline execution.

Stop when .halt exists. Otherwise continue.
