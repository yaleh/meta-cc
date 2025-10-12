# Phase 15: MCP Output Control & Tools Standardization

## Quick Links

- **[Implementation Plan](iteration-15-implementation-plan.md)** - Complete TDD implementation plan
- **[Stage 15.3 Summary](stage-15.3-summary.md)** - Tool description simplification summary
- **[Plan Overview](plan.md)** - Original phase planning document

## Phase Overview

**Goal**: Implement MCP output size control and standardize tool parameters to prevent context overflow

**Status**: Ready for implementation

**Timeline**: 2-3 days

**Priority**: High (resolves MCP context overflow, completes Phase 14 enhancements)

## Problem Statement

**Issue**: MCP `query_user_messages` returns ~10.7k tokens due to large session summaries in message content fields, overwhelming Claude's context budget.

**Root Cause**:
1. User messages can contain session summaries (thousands of lines)
2. `jq_filter ".[]"` returns full objects including massive `content` fields
3. `max_output_bytes` only truncates at the end (too late)

**Solution**: Two-layer output control
- Message-level truncation (`max_message_length`, default: 500 chars)
- Content summary mode (`content_summary`, metadata only)

**Expected Result**: 10.7k tokens → 1-2k tokens (81-91% compression)

## Stage Breakdown

### Stage 15.1: MCP Output Size Control (~150 lines)
- Implement `TruncateMessageContent()` function
- Implement `ApplyContentSummary()` function
- Add `max_message_length` and `content_summary` parameters to executor
- Unit tests with ≥90% coverage

**Files**:
- `cmd/mcp-server/filters.go` (new, ~80 lines)
- `cmd/mcp-server/executor.go` (update, ~50 lines)
- `cmd/mcp-server/executor_test.go` (update, ~70 lines)

### Stage 15.2: Standardize MCP Tool Parameters (~100 lines)
- Add new parameters to `StandardToolParameters()`
- Update all 15 tool definitions
- Verify parameter consistency in tests

**Files**:
- `cmd/mcp-server/tools.go` (update)
- `cmd/mcp-server/tools_test.go` (update)

### Stage 15.3: Simplify MCP Tool Descriptions (~50 lines)
- Reduce all tool descriptions to ≤100 characters
- Format: `<action> <object> <scope>`
- Add description length validation tests

**Files**:
- `cmd/mcp-server/tools.go` (update descriptions)
- `cmd/mcp-server/tools_test.go` (new validation tests)

### Stage 15.4: MCP Tool Documentation (~200 lines)
- Create comprehensive `docs/mcp-guide.md`
- Document all parameters, usage scenarios, examples
- Migration guide from deprecated tools
- Best practices and troubleshooting

**Files**:
- `docs/mcp-guide.md` (new, ~400-500 lines)
- `docs/plan.md` (update Phase 15 notes)

## Success Criteria

- ✅ MCP output compression ≥80% (10.7k → 1-2k tokens)
- ✅ All 15 tools support standard parameters
- ✅ All tool descriptions ≤100 characters
- ✅ Unit test coverage ≥85% for cmd/mcp-server
- ✅ Complete MCP tools reference documentation
- ✅ Integration tests pass
- ✅ No regressions in Phase 14 functionality

## Key Files

| File | Purpose | Lines | Status |
|------|---------|-------|--------|
| `iteration-15-implementation-plan.md` | Complete TDD implementation plan | ~1400 | ✅ Created |
| `cmd/mcp-server/filters.go` | Message truncation logic | ~80 | 📝 To implement |
| `cmd/mcp-server/executor.go` | Parameter handling | ~50 update | 📝 To update |
| `cmd/mcp-server/tools.go` | Tool definitions | ~100 update | 📝 To update |
| `docs/mcp-guide.md` | Complete tool reference | ~400-500 | 📝 To create |
| `test-scripts/validate-phase-15.sh` | Integration tests | ~100 | 📝 To create |

## Testing Strategy

### Unit Tests (TDD)
```bash
# Stage 15.1: Output control
go test ./cmd/mcp-server -run TestTruncate -v
go test ./cmd/mcp-server -run TestApplyContentSummary -v

# Stage 15.2: Parameter standardization
go test ./cmd/mcp-server -run TestAllToolsHaveStandard -v

# Stage 15.3: Description validation
go test ./cmd/mcp-server -run TestToolDescriptionLength -v

# Overall coverage
go test ./cmd/mcp-server -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Tests
```bash
# Build and validate
make all
./test-scripts/validate-phase-15.sh

# Verify output size reduction
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"query_user_messages","arguments":{"pattern":".*","max_message_length":500}}}' | ./meta-cc-mcp | wc -c
# Expected: <10000 bytes (vs ~30000 before)
```

## Performance Benchmarks

### Output Size
- **Before**: ~10.7k tokens (query_user_messages)
- **After (max_message_length=500)**: ~1.5k tokens (-86%)
- **After (content_summary=true)**: ~800 tokens (-93%)

### Processing Time
- **Baseline**: ~150ms average query
- **Target**: ~155-160ms (+3-7% overhead)
- **Truncation overhead**: +5ms
- **Summary overhead**: +3ms

## Migration Guide

### For MCP Clients

**Old behavior** (Phase 14):
```json
{
  "name": "query_user_messages",
  "arguments": {"pattern": ".*"}
}
// Returns: ~10.7k tokens
```

**New behavior** (Phase 15, recommended):
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": ".*",
    "max_message_length": 500
  }
}
// Returns: ~1.5k tokens
```

**Alternative** (metadata only):
```json
{
  "name": "query_user_messages",
  "arguments": {
    "pattern": ".*",
    "content_summary": true
  }
}
// Returns: ~800 tokens
```

### Backward Compatibility

Default parameters maintain existing behavior:
- `max_message_length`: 500 (configurable, 0=unlimited)
- `content_summary`: false (full content by default)

Set `max_message_length=0` to disable truncation entirely.

## Dependencies

### Prerequisites
- Phase 14 complete (gojq integration, meta-cc-mcp executable)
- Go 1.23+
- make, jq installed

### Related Phases
- **Phase 14**: MCP architecture refactoring (foundation)
- **Phase 16** (planned): Subagent integration
- **Phase 17** (planned): Advanced analytics

## Quick Start

```bash
# 1. Read the implementation plan
cat iteration-15-implementation-plan.md

# 2. Start with Stage 15.1 (TDD approach)
# Write tests first
vim cmd/mcp-server/executor_test.go

# Run tests (should fail)
go test ./cmd/mcp-server -run TestTruncate -v

# Implement filters.go
vim cmd/mcp-server/filters.go

# Run tests again (should pass)
go test ./cmd/mcp-server -run TestTruncate -v

# 3. Continue with remaining stages
# See iteration-15-implementation-plan.md for details
```

## Resources

- [MCP Protocol Specification](https://spec.modelcontextprotocol.io)
- [jq Manual](https://jqlang.github.io/jq/manual/)
- [Phase 14 Implementation Plan](../14/iteration-14-implementation-plan.md)
- [Project CLAUDE.md](/home/yale/work/meta-cc/CLAUDE.md)

## Timeline

| Day | Morning | Afternoon | Deliverable |
|-----|---------|-----------|-------------|
| 1 | Stage 15.1 implementation | Stage 15.1 tests | Message truncation working |
| 2 | Stage 15.2 parameters | Stage 15.3 descriptions | Tools standardized |
| 3 | Stage 15.4 documentation | Integration tests | Phase complete |

## Notes

- **TDD Required**: Write tests before implementation
- **Stage-Level Testing**: Run `make all` after each stage
- **Coverage Target**: ≥85% for cmd/mcp-server
- **Documentation First**: Keep docs in sync with code

---

**Last Updated**: 2025-10-06

**Status**: ✅ Ready for implementation

**Next Steps**: Begin Stage 15.1 (MCP Output Size Control)
