# Automate Evidence Capture

**Principle**: Every iteration should capture complexity and coverage metrics via a single command to keep BAIME evaluations trustworthy.

**Implementation**: Iteration 2 introduced `scripts/capture-mcp-metrics.sh`, later surfaced through `make metrics-mcp` (iteration-3.md). Running the target emits timestamped gocyclo and coverage reports under `build/methodology/`.

**Benefit**: Raises V_meta_effectiveness by eliminating manual data gathering and preventing stale metrics.
