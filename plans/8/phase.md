# Phase 8: Query Foundation & Integration Improvements

## Overview

**Phase Name**: Query Foundation & Integration Improvements

**Goal**: Implement core `query` command capabilities (Stage 8.1-8.4) AND update existing integrations to leverage Phase 8 (Stage 8.5-8.7)

**Code Budget**:
- Core Implementation (8.1-8.4): ~400 lines (Go code)
- Integration Updates (8.5-8.7): ~250 lines (configuration/documentation)
- **Total**: ~650 lines

**Priority**: High (core query capability + immediate practical improvements)

**Status**: ✅ Stages 8.1-8.4 Completed, Stages 8.5-8.7 Planned

## Stage Breakdown

### Core Query Implementation (Completed)

#### Stage 8.1: Query Command Framework ✅
- **Objective**: Establish `query` command structure
- **Code**: ~100 lines
- **Deliverables**: `cmd/query.go` with routing

#### Stage 8.2: Query Tools Command ✅
- **Objective**: Implement `query tools` with filtering
- **Code**: ~120 lines
- **Deliverables**: `cmd/query_tools.go`, tool filtering, sorting

#### Stage 8.3: Query User-Messages Command ✅
- **Objective**: Implement `query user-messages` with regex
- **Code**: ~100 lines
- **Deliverables**: `cmd/query_messages.go`, regex pattern matching

#### Stage 8.4: Enhanced Filter Engine ✅
- **Objective**: Support `--where` syntax
- **Code**: ~80 lines
- **Deliverables**: Enhanced `internal/filter/` package

### Integration Improvements (New - Planned)

#### Stage 8.5: Update Slash Commands for Phase 8
- **Objective**: Update existing Slash Commands to use Phase 8 capabilities
- **Code**: ~50 lines (configuration changes)
- **Time**: 15-30 minutes
- **Deliverables**:
  - Update `/meta-timeline` to use `query tools --limit`
  - Verify `/meta-stats` already optimal
  - Avoid context overflow in large sessions

#### Stage 8.6: Update @meta-coach Documentation
- **Objective**: Enable @meta-coach to use Phase 8 query commands
- **Code**: ~80 lines (documentation)
- **Time**: 20-30 minutes
- **Deliverables**:
  - Add Phase 8 query capabilities section
  - Document iterative analysis pattern
  - Add best practices for large sessions
  - Include Phase 8 examples in coaching scenarios

#### Stage 8.7: Create New Query-Focused Slash Commands
- **Objective**: Create specialized Slash Commands for quick queries
- **Code**: ~120 lines (2 new commands)
- **Time**: 30-45 minutes
- **Deliverables**:
  - `/meta-query-tools [tool] [status] [limit]` - Quick tool query
  - `/meta-query-messages [pattern] [limit]` - Message search

## Architecture

### Core Query Flow (Stage 8.1-8.4)
```
meta-cc query <type> [filters] [options]
              │
              ├─ tools          → Query tool calls (8.2)
              ├─ user-messages  → Query user messages (8.3)
              └─ [future: sessions, errors]

┌─────────────┐    ┌──────────┐    ┌─────────┐    ┌────────┐    ┌────────┐
│   Locator   │───→│  Parser  │───→│ Querier │───→│ Filter │───→│ Output │
└─────────────┘    └──────────┘    └─────────┘    └────────┘    └────────┘
```

### Integration Layer (Stage 8.5-8.7)
```
User Interface Layer:
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│ Slash Commands   │  │   @meta-coach    │  │  New Commands    │
│ (Updated)        │  │   (Enhanced)     │  │  (Created)       │
│                  │  │                  │  │                  │
│ /meta-timeline   │  │ Phase 8 aware    │  │ /meta-query-*    │
│ /meta-stats      │  │ Iterative mode   │  │                  │
└────────┬─────────┘  └────────┬─────────┘  └────────┬─────────┘
         │                     │                     │
         └─────────────────────┼─────────────────────┘
                               ↓
                    ┌──────────────────────┐
                    │  Phase 8 Query API   │
                    │  query tools         │
                    │  query user-messages │
                    └──────────────────────┘
```

## Key Features

### Core Query Capabilities (8.1-8.4)
- ✅ Flexible tool call queries with filtering
- ✅ User message search with regex
- ✅ Sorting and limiting
- ✅ Enhanced `--where` filter syntax
- ✅ Pagination support

### Integration Improvements (8.5-8.7)
- 📋 Existing commands use Phase 8 (avoid context overflow)
- 📋 @meta-coach leverages new query capabilities
- 📋 Quick query commands for common tasks
- 📋 Better user experience with specialized commands

## Usage Examples

### Core Query Commands (8.1-8.4)
```bash
# Query tool calls
meta-cc query tools --status error --limit 20
meta-cc query tools --tool Bash --sort-by timestamp
meta-cc query tools --where "tool=Edit,status=error"

# Query user messages
meta-cc query user-messages --match "fix.*bug"
meta-cc query user-messages --match "error|warning" --limit 10
```

### Updated Slash Commands (8.5)
```bash
# /meta-timeline now uses Phase 8
/meta-timeline          # Uses query tools --limit 50 (no overflow)
/meta-timeline 100      # Custom limit
```

### Enhanced @meta-coach (8.6)
```
@meta-coach 分析我的工作流

# Now uses:
# - query tools --limit 100 (efficient)
# - Iterative analysis pattern
# - Targeted queries (no context overflow)
```

### New Quick Commands (8.7)
```bash
# Quick tool query
/meta-query-tools Bash                    # All Bash calls
/meta-query-tools "" error                # All errors
/meta-query-tools Edit error 10           # Last 10 Edit errors

# Quick message search
/meta-query-messages "Phase 8"            # Find mentions
/meta-query-messages "error|bug"          # Regex search
/meta-query-messages "fix.*bug" 20        # Complex regex
```

## Implementation Priority

### Must Do (Stage 8.1-8.4) ✅
Core query infrastructure - **COMPLETED**

### High Priority (Stage 8.5-8.6) 📋
- Stage 8.5: Update Slash Commands (15-30 min)
  - Critical: Prevents context overflow in large sessions
  - Low risk: Minimal changes, high impact

- Stage 8.6: Update @meta-coach (20-30 min)
  - Important: Enables better coaching
  - Demonstrates Phase 8 value

### Medium Priority (Stage 8.7) 📋
- Stage 8.7: New Quick Commands (30-45 min)
  - Nice to have: Improves UX
  - Can be deferred if time-constrained

## Testing Strategy

### Stage 8.5 Testing
```bash
# Test updated /meta-timeline
/meta-timeline          # Default limit
/meta-timeline 20       # Custom limit
# Verify: No context overflow in large sessions
```

### Stage 8.6 Testing
```
@meta-coach 分析我的工作流
# Verify: Uses query commands
# Verify: Demonstrates iterative pattern
```

### Stage 8.7 Testing
```bash
/meta-query-tools Bash
/meta-query-tools "" error
/meta-query-messages "Phase 8"
# Verify: Clear output, helpful tips
```

### Integration Testing
```bash
# Test all commands in a large session (>500 turns)
/meta-stats              # Should work (already optimal)
/meta-timeline 100       # Should use Phase 8
@meta-coach 帮我优化      # Should use iterative pattern
/meta-query-tools Bash   # Should work efficiently
```

## Success Metrics

### Core Implementation (8.1-8.4) ✅
- ✅ All unit tests pass
- ✅ `query tools` and `query user-messages` work
- ✅ Filtering, sorting, limiting functional
- ✅ Performance < 100ms for typical sessions

### Integration Success (8.5-8.7)
- 📋 No context overflow in sessions >500 turns
- 📋 @meta-coach uses Phase 8 iterative pattern
- 📋 New commands provide clear, helpful output
- 📋 Users can perform common queries without CLI knowledge

## Dependencies

### Prerequisites
- ✅ Phase 0-7 completed (all infrastructure ready)
- ✅ Stage 8.1-8.4 completed (query commands available)
- ✅ `meta-cc` binary in PATH
- ✅ `jq` installed (for Slash Commands)

### Stage Dependencies
- Stage 8.5: Depends on 8.2 (query tools)
- Stage 8.6: Depends on 8.2, 8.3 (query tools, messages)
- Stage 8.7: Depends on 8.2, 8.3 (query tools, messages)

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Context overflow in large sessions | High | ✅ Stage 8.5 updates commands to use `--limit` |
| @meta-coach not using Phase 8 | Medium | 📋 Stage 8.6 adds documentation and examples |
| Users don't discover query commands | Low | 📋 Stage 8.7 provides easy-to-use Slash Commands |
| Backward compatibility | Low | All changes are additive, old commands still work |

## Deliverables Checklist

### Core Implementation (8.1-8.4) ✅
- ✅ `cmd/query.go` - Command framework
- ✅ `cmd/query_tools.go` - Tool query implementation
- ✅ `cmd/query_messages.go` - Message query implementation
- ✅ `internal/filter/` enhancements - WHERE syntax
- ✅ Unit tests for all new code
- ✅ Integration tests passing

### Integration Updates (8.5-8.7) 📋
- 📋 `.claude/commands/meta-timeline.md` - Updated to use Phase 8
- 📋 `.claude/agents/meta-coach.md` - Phase 8 documentation added
- 📋 `.claude/commands/meta-query-tools.md` - New command created
- 📋 `.claude/commands/meta-query-messages.md` - New command created
- 📋 `README.md` or `docs/examples-usage.md` - Usage examples updated

## Documentation Updates

### Files to Update
1. **README.md** - Add Phase 8 query examples
2. **docs/examples-usage.md** - Add quick query command guide
3. **docs/plan.md** - Update Phase 8 description

### New Documentation
- `plans/8/stage-8.5.md` - Slash Commands update plan ✅
- `plans/8/stage-8.6.md` - @meta-coach update plan ✅
- `plans/8/stage-8.7.md` - New commands plan ✅
- `plans/8/phase.md` - This overview ✅

## Next Steps

### Immediate (Stage 8.5-8.7)
1. ✅ Plan created (this document)
2. 📋 Execute Stage 8.5 (15-30 min): Update Slash Commands
3. 📋 Execute Stage 8.6 (20-30 min): Update @meta-coach
4. 📋 Execute Stage 8.7 (30-45 min): Create new commands
5. 📋 Test all integrations
6. 📋 Update main documentation

### Future (Phase 9+)
- Phase 9: Context-Length Management (pagination, chunking)
- Phase 10: Advanced Query (aggregation, time-series)
- Phase 11: Unix Composability (streaming, exit codes)

## Related Documentation

- **Implementation Plan**: `plans/8/phase-8-implementation-plan.md`
- **Stage 8.5 Plan**: `plans/8/stage-8.5.md`
- **Stage 8.6 Plan**: `plans/8/stage-8.6.md`
- **Stage 8.7 Plan**: `plans/8/stage-8.7.md`
- **Integration Proposal**: `/tmp/meta-cc-integration-improvement-proposal.md`
- **Main Plan**: `docs/plan.md`
