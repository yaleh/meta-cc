# Phase 11: Unix Composability (Unix 工具可组合性)

## Overview

This phase optimizes output formats and CLI design to enhance Unix pipeline support, enabling meta-cc to compose seamlessly with standard Unix tools like jq, grep, awk, and other shell utilities.

## Quick Start

### Implementation Order

1. **Stage 11.1**: JSONL Streaming Output (TDD: tests first)
2. **Stage 11.2**: Exit Code Standardization (TDD: tests first)
3. **Stage 11.3**: stderr/stdout Separation (TDD: tests first)
4. **Stage 11.4**: Documentation - Cookbook and Composability Guide

### After Each Stage

```bash
# Stage 11.1: Test streaming output
go test ./internal/output -run TestStreaming
./meta-cc query tools --stream | head -5

# Stage 11.2: Test exit codes
go test ./internal/output -run TestExitCodes
./meta-cc query tools --where "tool='NonExistent'" && echo "Found" || echo "Not found (exit $?)"

# Stage 11.3: Test stderr/stdout separation
go test ./internal/output -run TestStderrStdout
./meta-cc query tools 2>/dev/null | wc -l  # Only data on stdout
./meta-cc query tools >/dev/null            # Only logs on stderr

# Stage 11.4: Validate documentation examples
cd docs && bash -x cookbook.md  # Run cookbook examples
```

### After Phase Completion

```bash
# Run integration tests
./tests/integration/unix_composability_test.sh

# Verify Unix pipeline workflows
./meta-cc query tools --stream | jq '.ToolName' | sort | uniq -c
./meta-cc stats aggregate --group-by tool --output json | jq '.[] | select(.error_rate > 0.1)'
./meta-cc query tools --where "status='error'" | grep -i "permission"

# Update README.md
# Add "Unix Composability" section with pipeline examples
```

## Deliverables

### Commands

- `meta-cc query tools --stream` - JSONL streaming output
- Exit codes: 0 (success), 1 (error), 2 (no results)
- All commands: Logs to stderr, data to stdout

### Examples

```bash
# Streaming for pipeline processing
meta-cc query tools --stream | jq -c 'select(.Status == "error")'

# Exit codes for scripting
if meta-cc query tools --where "tool='Bash' AND status='error'"; then
  echo "Found Bash errors!"
fi

# stderr/stdout separation
meta-cc query tools --output json 2>debug.log | jq '.[] | .ToolName'
```

### Documentation

- `docs/cookbook.md`: 10+ practical analysis patterns
- `docs/cli-composability.md`: jq/grep/awk integration examples

### Testing

- Unit tests: `internal/output/stream_test.go`, `internal/output/exitcode_test.go`
- Integration tests: `tests/integration/unix_composability_test.sh`
- Example validation: All cookbook examples are executable and verified

## Code Budget

- Stage 11.1: JSONL Streaming (~50 lines)
- Stage 11.2: Exit Code Standardization (~30 lines)
- Stage 11.3: stderr/stdout Separation (~40 lines)
- Stage 11.4: Documentation (~80 lines markdown + examples)
- **Total**: ~200 lines (within target)

## Success Criteria

- All unit tests pass (100%)
- Integration tests pass
- `--stream` flag works with all query commands
- Exit codes match Unix conventions
- All logs go to stderr, all data to stdout
- Cookbook has 10+ working examples
- No regressions (all existing tests pass)
- Pipeline workflows validated: `meta-cc | jq`, `meta-cc | grep`, `meta-cc | awk`

## Stage Checklist

- [ ] Stage 11.1: JSONL Streaming Output
  - [ ] Add `--stream` flag to query commands
  - [ ] Implement `internal/output/stream.go`
  - [ ] Support streaming for JSON output format
  - [ ] Validate JSONL format (one JSON object per line)
  - [ ] Unit tests pass

- [ ] Stage 11.2: Exit Code Standardization
  - [ ] Define exit codes in `internal/output/exitcode.go`
  - [ ] Update all commands to return standard codes
  - [ ] 0 = success, 1 = error, 2 = no results
  - [ ] Handle graceful degradation (partial results)
  - [ ] Unit tests pass

- [ ] Stage 11.3: stderr/stdout Separation
  - [ ] Audit all commands for output destinations
  - [ ] Move logs to stderr (progress, debug, warnings)
  - [ ] Ensure data only on stdout (JSON, Markdown, TSV)
  - [ ] Update formatters to respect separation
  - [ ] Unit tests pass

- [ ] Stage 11.4: Cookbook Documentation
  - [ ] Create `docs/cookbook.md` with 10+ patterns
  - [ ] Create `docs/cli-composability.md` with tool examples
  - [ ] Validate all examples are executable
  - [ ] Add pipeline workflow diagrams
  - [ ] Update README.md with composability section

- [ ] Integration Testing
  - [ ] Pipeline workflows pass
  - [ ] Exit code behavior validated
  - [ ] Real-world scenarios tested

- [ ] Documentation
  - [ ] README.md updated
  - [ ] Usage examples added
  - [ ] Integration guide updated

## Integration Points

Phase 11 builds on:
- **Phase 8-10**: All query and stats commands
- **Phase 9**: Output formatting infrastructure

Phase 11 provides:
- **Unix composability**: Seamless pipeline integration
- **Scripting support**: Predictable exit codes
- **Clean I/O**: Proper separation of logs and data
- **Best practices**: Cookbook and integration patterns

## Implementation Status

**Phase 11 状态**: 📝 Planning Complete, Ready for Implementation

**Next Action**: Begin Stage 11.1 TDD - write tests for JSONL streaming
