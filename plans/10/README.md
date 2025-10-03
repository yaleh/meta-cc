# Phase 10: Advanced Query Capabilities

## Overview

This phase implements advanced filtering, aggregation functions, time series analysis, and file-level statistics to provide richer data analysis capabilities for Claude integration layers.

## Quick Start

### Implementation Order

1. **Stage 10.1**: Advanced Filter Engine (TDD: tests first)
2. **Stage 10.2**: Aggregation Functions (TDD: tests first)
3. **Stage 10.3**: Time Series Analysis (TDD: tests first)
4. **Stage 10.4**: File-Level Statistics (TDD: tests first)

### After Each Stage

```bash
# Run stage-specific unit tests
go test ./internal/filter -run TestAdvanced
go test ./internal/analyzer -run TestAggregate
go test ./internal/analyzer -run TestTimeSeries
go test ./internal/analyzer -run TestFileStats

# Run all tests
go test ./...

# Build and test manually
go build -o meta-cc
./meta-cc query tools --where "tool IN ('Bash','Edit')"
./meta-cc stats aggregate --group-by tool
```

### After Phase Completion

```bash
# Run integration tests
./tests/integration/advanced_query_test.sh

# Verify with real projects
./meta-cc query tools --where "status='error' AND duration>1000"
./meta-cc stats aggregate --group-by tool --metrics "count,error_rate"
./meta-cc stats time-series --metric tool-calls --interval hour
./meta-cc stats files --sort-by error-count --top 10

# Update README.md
# Add "Advanced Query Features" section
```

## Deliverables

### Commands

- `meta-cc query tools --where "EXPRESSION"` (boolean expressions, range queries)
- `meta-cc stats aggregate --group-by FIELD --metrics "METRIC1,METRIC2"`
- `meta-cc stats time-series --metric METRIC --interval INTERVAL`
- `meta-cc stats files --sort-by FIELD --top N`

### Examples

```bash
# Advanced filtering
meta-cc query tools --where "tool IN ('Bash','Edit') AND status='error'"
meta-cc query tools --where "duration > 1000 AND timestamp BETWEEN '2025-10-01' AND '2025-10-03'"
meta-cc query tools --where "NOT (status='success') AND tool LIKE 'meta%'"

# Aggregation
meta-cc stats aggregate --group-by tool --metrics "count,error_rate,avg_duration"
meta-cc stats aggregate --group-by status --metrics "count,percentage"

# Time series
meta-cc stats time-series --metric tool-calls --interval hour --output json
meta-cc stats time-series --metric error-rate --interval day --window 7

# File statistics
meta-cc stats files --sort-by edit-count --top 20
meta-cc stats files --sort-by error-count --filter "errors>3"
```

### Testing

- Unit tests: `internal/filter/advanced_test.go`, `internal/analyzer/aggregate_test.go`, etc.
- Integration tests: `tests/integration/advanced_query_test.sh`
- Real-world validation: 3 projects (meta-cc, NarrativeForge, claude-tmux)

## Documentation

See [plan.md](./plan.md) for complete details:
- TDD test scenarios for each stage
- Implementation code structure
- Acceptance criteria per stage
- Integration with existing commands
- Performance benchmarks

## Code Budget

- Stage 10.1: Advanced Filter Engine (~120 lines)
- Stage 10.2: Aggregation Functions (~100 lines)
- Stage 10.3: Time Series Analysis (~100 lines)
- Stage 10.4: File-Level Statistics (~80 lines)
- **Total**: ~400 lines (within 350-450 line target)

## Success Criteria

- All unit tests pass (100%)
- Integration tests pass
- Query performance < 200ms for complex queries on typical sessions
- Aggregation accuracy verified against manual calculations
- Time series data correctly bucketed by interval
- File statistics match actual file operation counts
- Works with all verified projects
- README.md updated with advanced query documentation

## Stage Checklist

- [ ] Stage 10.1: Advanced Filter Engine
  - [ ] Boolean expression parser (AND, OR, NOT)
  - [ ] Range queries (BETWEEN, >, <, >=, <=)
  - [ ] Set operations (IN, NOT IN)
  - [ ] Pattern matching (LIKE, REGEXP)
  - [ ] Unit tests pass

- [ ] Stage 10.2: Aggregation Functions
  - [ ] `stats aggregate` command framework
  - [ ] Group-by functionality
  - [ ] Metrics: count, sum, avg, min, max, error_rate, percentiles
  - [ ] Unit tests pass

- [ ] Stage 10.3: Time Series Analysis
  - [ ] `stats time-series` command
  - [ ] Time bucketing (hour, day, week)
  - [ ] Metrics over time (tool-calls, errors, duration)
  - [ ] Trend detection (optional)
  - [ ] Unit tests pass

- [ ] Stage 10.4: File-Level Statistics
  - [ ] `stats files` command
  - [ ] File operation tracking (read, edit, write)
  - [ ] File modification frequency
  - [ ] Error correlation by file
  - [ ] Unit tests pass

- [ ] Integration Testing
  - [ ] All stages integrated
  - [ ] End-to-end tests pass
  - [ ] Real project validation

- [ ] Documentation
  - [ ] README.md updated
  - [ ] Usage examples added
  - [ ] Performance benchmarks documented

## Integration Points

Phase 10 builds on:
- **Phase 8**: Query command framework (`query tools`, `query user-messages`)
- **Phase 9**: Output control strategies (pagination, projection, compact formats)

Phase 10 provides:
- **Advanced filtering**: Rich query expressions for precise data retrieval
- **Statistical analysis**: Aggregations and trends for pattern recognition
- **File insights**: File-centric views for debugging and optimization
- **Data for Claude**: High-density structured data for @meta-coach and MCP Server

## Implementation Status

**Phase 10 状态**: 📝 Planning Complete, Ready for Implementation

**Next Action**: Begin Stage 10.1 TDD - write tests for advanced filter engine
