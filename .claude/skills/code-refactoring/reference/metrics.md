# Metrics Playbook

- **Cyclomatic Complexity**: capture with `gocyclo cmd/mcp-server` or `make metrics-mcp`; target runtime hotspots ≤ 10 post-refactor.
- **Test Coverage**: rely on `make metrics-mcp` (71.4% achieved); aim for +1% delta per iteration when feasible.
- **Value Functions**: calculate V_instance and V_meta per iteration; see iterations/iteration-*.md for formulas and evidence.
- **Artifacts**: store snapshots under `build/methodology/` with ISO timestamps for audit trails.
