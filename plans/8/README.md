# Phase 8: Query Foundation

## Overview

This phase implements the `meta-cc query` command group to provide flexible, user-friendly data retrieval from Claude Code sessions.

## Quick Start

### Implementation Order

1. **Stage 8.1**: Query command framework (TDD: tests first)
2. **Stage 8.2**: Query tools command (TDD: tests first)
3. **Stage 8.3**: Query user-messages command (TDD: tests first)
4. **Stage 8.4**: Enhanced filter engine (TDD: tests first)

### After Each Stage

```bash
# Run unit tests
go test ./cmd -run TestQuery
go test ./internal/filter -run TestWhere

# Run all tests
go test ./...

# Build and test manually
go build -o meta-cc
./meta-cc query --help
```

### After Phase Completion

```bash
# Run integration tests
./tests/integration/query_commands_test.sh

# Verify with real projects
./meta-cc query tools --status error
./meta-cc query user-messages --match "error"

# Update README.md
# Add "Query Commands" section
```

## Deliverables

### Commands

- `meta-cc query tools [--status STATUS] [--tool TOOL] [--where CONDITION] [--limit N] [--sort-by FIELD]`
- `meta-cc query user-messages [--match PATTERN] [--limit N] [--sort-by FIELD]`

### Examples

```bash
# Query tool calls
meta-cc query tools --status error --limit 20
meta-cc query tools --tool Bash --sort-by timestamp
meta-cc query tools --where "tool=Edit,status=error"

# Query user messages
meta-cc query user-messages --match "fix.*bug"
meta-cc query user-messages --match "error|warning" --limit 10
```

### Testing

- Unit tests: `cmd/query_test.go`, `cmd/query_tools_test.go`, `cmd/query_messages_test.go`
- Integration tests: `tests/integration/query_commands_test.sh`
- Real-world validation: 3 projects (meta-cc, NarrativeForge, claude-tmux)

## Documentation

See [phase-8-implementation-plan.md](./phase-8-implementation-plan.md) for complete details:
- TDD test scenarios
- Implementation code snippets
- Acceptance criteria
- Integration testing
- README.md updates

## Code Budget

- Core Implementation (8.1-8.4): ~400 lines
  - Stage 8.1: ~100 lines (framework)
  - Stage 8.2: ~120 lines (query tools)
  - Stage 8.3: ~100 lines (query messages)
  - Stage 8.4: ~80 lines (filter enhancement)
- Integration Updates (8.5-8.7): ~250 lines (config/docs)
- MCP Integration (8.8-8.9): ~120 lines
- Context Query Extensions (8.10-8.11): ~280 lines
  - Stage 8.10: ~180 lines (context queries)
  - Stage 8.11: ~100 lines (workflow patterns)
- **Total**: ~1050 lines

## Success Criteria

- ✅ All unit tests pass (100%)
- ✅ Integration tests pass
- ✅ Query performance < 100ms for typical sessions
- ✅ Works with all verified projects
- ✅ README.md updated
- ✅ Code coverage ≥ 80%
