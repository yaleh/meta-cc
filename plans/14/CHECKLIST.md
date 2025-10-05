# Phase 14 Implementation Checklist

## Pre-Implementation

- [ ] Review `iteration-14-implementation-plan.md` thoroughly
- [ ] Understand breaking changes in `MIGRATION.md`
- [ ] Set up development branch: `git checkout -b feature/phase-14`
- [ ] Ensure Phase 13 is committed and working
- [ ] Create baseline test outputs for comparison

---

## Stage 14.1: Pipeline Abstraction Layer

### Implementation
- [ ] Create `pkg/pipeline/session.go`
- [ ] Implement `NewSessionPipeline()` constructor
- [ ] Implement `Load()` method (session location + JSONL parsing)
- [ ] Implement `ExtractToolCalls()` method
- [ ] Implement `ExtractUserMessages()` method
- [ ] Implement `BuildTurnIndex()` method with caching
- [ ] Add `SessionPath()` and `EntryCount()` helper methods

### Testing
- [ ] Create `pkg/pipeline/session_test.go`
- [ ] Test: Load via session ID
- [ ] Test: Load via project path
- [ ] Test: Load via auto-detect
- [ ] Test: Error handling (missing session)
- [ ] Test: Extract tool calls
- [ ] Test: Extract user messages
- [ ] Test: Build turn index (idempotency)
- [ ] Run: `go test ./pkg/pipeline -v`
- [ ] Verify: Coverage ≥90%

### Validation
- [ ] Pipeline loads real session successfully
- [ ] Extracted data matches parser output
- [ ] Turn index correctly built
- [ ] All unit tests pass

---

## Stage 14.2: Simplify errors Command

### Implementation
- [ ] Create `cmd/query_errors.go`
- [ ] Define `ErrorEntry` struct
- [ ] Implement `queryErrorsCmd` Cobra command
- [ ] Implement `runQueryErrors()` function
- [ ] Implement `generateErrorSignature()` helper
- [ ] Use SessionPipeline for session loading
- [ ] Extract errors from tool calls
- [ ] Sort errors by timestamp
- [ ] Apply pagination if specified
- [ ] Format output via `output.Format()`

### Cleanup
- [ ] Update `cmd/analyze.go` - remove `analyzeErrorsCmd`
- [ ] Remove `analyze errors` from command tree
- [ ] Keep other analyze subcommands (sequences, file-churn, idle)

### Testing
- [ ] Create `cmd/query_errors_test.go`
- [ ] Test: `generateErrorSignature()` short error
- [ ] Test: `generateErrorSignature()` long error (truncation)
- [ ] Test: `generateErrorSignature()` whitespace normalization
- [ ] Test: Integration test with session containing errors
- [ ] Run: `go test ./cmd -run TestQueryErrors -v`
- [ ] Verify: 5 errors → 5 results
- [ ] Verify: Sorted by timestamp
- [ ] Verify: Signatures correct format

### Validation
- [ ] `meta-cc query errors` works
- [ ] `meta-cc analyze errors` removed (returns error)
- [ ] Error signatures use `{tool}:{error[:50]}` format
- [ ] Output sorted deterministically
- [ ] Code reduced from 317 → ~80 lines

### Documentation
- [ ] Update `.claude/commands/meta-errors.md` (use jq aggregation)
- [ ] Add migration note to README.md
- [ ] Update command help text

---

## Stage 14.3: Output Sorting Standardization

### Implementation
- [ ] Create `pkg/output/sort.go`
- [ ] Implement `SortByTimestamp()` for multiple types
- [ ] Implement `SortByTurnSequence()` for multiple types
- [ ] Implement `SortByUUID()` for tool calls
- [ ] Implement `DefaultSort()` (delegates to SortByTimestamp)

### Testing
- [ ] Create `pkg/output/sort_test.go`
- [ ] Test: SortByTimestamp with 3 unsorted items
- [ ] Test: SortByTurnSequence with 3 unsorted items
- [ ] Test: SortByUUID with 3 unsorted items
- [ ] Test: Determinism (sort 100 random items twice, verify identical)
- [ ] Run: `go test ./pkg/output -run TestSort -v`
- [ ] Verify: All sorting functions work
- [ ] Verify: Idempotent (running twice produces same result)

### Integration
- [ ] Update `cmd/query_tools.go` - add `output.SortByTimestamp(tools)`
- [ ] Update `cmd/query_messages.go` - add `output.SortByTurnSequence(messages)`
- [ ] Update `cmd/query_errors.go` - add `output.SortByTimestamp(errors)`
- [ ] Update `cmd/parse.go` - verify stats sorting
- [ ] Update `cmd/analyze_sequences.go` - add sorting

### Validation
- [ ] Run same query twice: `meta-cc query tools --limit 50`
- [ ] Verify outputs are identical (byte-for-byte)
- [ ] Test all query commands for determinism
- [ ] Code added: ~50 lines (source) + ~60 lines (tests)

---

## Stage 14.4: Code Deduplication

### Helper Function
- [ ] Add `getGlobalOptions()` to `cmd/root.go`

### Refactor: parse stats
- [ ] Update `cmd/parse.go` - refactor `runParseStats()`
- [ ] Use SessionPipeline instead of direct locator/parser
- [ ] Verify behavior unchanged
- [ ] Test: `go test ./cmd -run TestParseStats -v`
- [ ] Code reduced: ~170 → ~60 lines (-65%)

### Refactor: query tools
- [ ] Update `cmd/query_tools.go` - refactor `runQueryTools()`
- [ ] Use SessionPipeline
- [ ] Add sorting call
- [ ] Verify behavior unchanged
- [ ] Test: `go test ./cmd -run TestQueryTools -v`
- [ ] Code reduced: ~307 → ~80 lines (-74%)

### Refactor: query messages
- [ ] Update `cmd/query_messages.go` - refactor `runQueryMessages()`
- [ ] Use SessionPipeline
- [ ] Add sorting call
- [ ] Verify behavior unchanged
- [ ] Test: `go test ./cmd -run TestQueryMessages -v`
- [ ] Code reduced: ~280 → ~70 lines (-75%)

### Refactor: analyze sequences
- [ ] Update `cmd/analyze_sequences.go` - refactor `runAnalyzeSequences()`
- [ ] Use SessionPipeline
- [ ] Add sorting call
- [ ] Verify behavior unchanged
- [ ] Test: `go test ./cmd -run TestAnalyzeSequences -v`
- [ ] Code reduced: ~120 → ~50 lines (-58%)

### Refactor: analyze file-churn
- [ ] Update `cmd/analyze_file_churn.go` - refactor `runAnalyzeFileChurn()`
- [ ] Use SessionPipeline
- [ ] Verify behavior unchanged
- [ ] Test: `go test ./cmd -run TestAnalyzeFileChurn -v`

### Validation
- [ ] Verify total code reduction ≥60%
- [ ] Run: `wc -l cmd/*.go pkg/pipeline/*.go`
- [ ] Target: ≤460 lines total (340 commands + 120 pipeline)
- [ ] No duplicate session location code
- [ ] No duplicate JSONL parsing code
- [ ] No duplicate output formatting code

---

## Integration Testing

### Unit Tests
- [ ] Run all tests: `go test ./... -v`
- [ ] Verify 0 failures
- [ ] Check coverage: `go test ./... -coverprofile=coverage.out`
- [ ] Verify coverage ≥80%

### Phase 14 Validation Script
- [ ] Create `test-scripts/validate-phase-14.sh`
- [ ] Test 1: Pipeline functionality
- [ ] Test 2: query errors command
- [ ] Test 3: Output determinism
- [ ] Test 4: Code size verification
- [ ] Test 5: Behavior equivalence
- [ ] Run: `./test-scripts/validate-phase-14.sh`
- [ ] Verify: All tests pass

### Real-World Validation
- [ ] Test with meta-cc project session
- [ ] Test with NarrativeForge project session
- [ ] Test with claude-tmux project session
- [ ] Verify deterministic output across all
- [ ] Verify no errors

### Slash Commands
- [ ] Test `/meta-stats` in Claude Code
- [ ] Test updated `/meta-errors` in Claude Code
- [ ] Test `/meta-timeline` in Claude Code
- [ ] Verify all commands work
- [ ] Verify output is correct

### MCP Server
- [ ] Test `get_session_stats` tool
- [ ] Test `query_tools` tool (verify deterministic)
- [ ] Test `query_errors` tool (new name)
- [ ] Verify `analyze_errors` removed or renamed
- [ ] All MCP tools return sorted data

---

## Documentation

### README.md
- [ ] Add "Phase 14 Breaking Changes" section
- [ ] Document `analyze errors` → `query errors` migration
- [ ] Add jq examples for error aggregation
- [ ] Update "Architecture" section (mention Pipeline)
- [ ] Update code size statistics

### CHANGELOG.md
- [ ] Add Phase 14 entry
- [ ] List breaking changes
- [ ] List improvements
- [ ] Note code size reduction

### Slash Commands
- [ ] Update `.claude/commands/meta-errors.md`
- [ ] Test updated command
- [ ] Verify jq aggregation works

### Migration Guide
- [ ] Verify `MIGRATION.md` is complete
- [ ] Add troubleshooting section if needed
- [ ] Test all migration examples

---

## Code Quality

### Linting
- [ ] Run: `go fmt ./...`
- [ ] Run: `go vet ./...`
- [ ] Fix any warnings

### Code Review
- [ ] Pipeline abstraction is clean
- [ ] No duplicate code across commands
- [ ] Error handling is consistent
- [ ] Sorting is applied correctly
- [ ] Tests are comprehensive

### Performance
- [ ] Benchmark: `time meta-cc query tools --limit 1000`
- [ ] Verify: <250ms (target: ~235ms, +2% from Phase 13)
- [ ] No memory leaks
- [ ] No performance regressions

---

## Pre-Commit

### Final Checks
- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Validation script passes
- [ ] Code coverage ≥80%
- [ ] Documentation complete
- [ ] Breaking changes documented
- [ ] Migration guide complete

### Code Size Verification
- [ ] Run: `wc -l cmd/*.go pkg/pipeline/*.go`
- [ ] Verify: Total ≤460 lines
- [ ] Verify: Net reduction ≥60% from Phase 13
- [ ] Commands average ≤68 lines each

### Git
- [ ] Review changes: `git diff`
- [ ] Stage changes: `git add .`
- [ ] Verify no unintended changes
- [ ] Check for TODO comments

---

## Commit & Push

### Commit Message Template
```
refactor(phase-14): architecture refactoring with pipeline abstraction

Phase 14: Architecture Refactoring
- Add SessionPipeline abstraction (pkg/pipeline/session.go)
- Simplify errors command: analyze errors → query errors
- Standardize output sorting (deterministic results)
- Deduplicate 345 lines across 5 commands (-72%)

Breaking Changes:
- `meta-cc analyze errors` removed, use `meta-cc query errors`
- Error signature format changed (SHA256 → readable)
- `--window` parameter removed (use jq for windowing)
- Output now deterministically sorted

Code Size:
- Before: 1194 lines (5 commands)
- After: 460 lines (340 commands + 120 pipeline)
- Reduction: 734 lines (-61.5%)

Tests:
- Unit tests: 47/47 passing
- Coverage: 85%+
- Integration: validation script passing

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

### Commands
- [ ] Commit: `git commit -m "..."`
- [ ] Push: `git push origin feature/phase-14`
- [ ] Create PR (if needed)

---

## Post-Implementation

### Verification
- [ ] Pull latest changes
- [ ] Clean build: `go clean && go build`
- [ ] Run full test suite
- [ ] Test in fresh environment

### Documentation
- [ ] Update online docs (if applicable)
- [ ] Notify team of breaking changes
- [ ] Publish migration guide

### Monitoring
- [ ] Monitor for issues
- [ ] Address user feedback
- [ ] Fix any bugs quickly

---

## Rollback Plan (If Needed)

### Revert Steps
- [ ] Checkout Phase 13: `git checkout feature/phase-13`
- [ ] Rebuild: `go build -o meta-cc`
- [ ] Test: `./meta-cc --version`
- [ ] Document rollback reason

### Communication
- [ ] Notify users of rollback
- [ ] Document issues encountered
- [ ] Plan fixes for Phase 14 retry

---

## Success Metrics

### Quantitative
- ✅ Code reduction: ≥60% (target: -72%)
- ✅ Test coverage: ≥80%
- ✅ Unit tests: 100% passing
- ✅ Integration tests: 100% passing
- ✅ Performance: ≤250ms (target: ~235ms)
- ✅ No regressions

### Qualitative
- ✅ Code is more maintainable
- ✅ Responsibilities are clearer
- ✅ Output is deterministic
- ✅ Pipeline pattern is reusable
- ✅ Documentation is complete

### User Impact
- ✅ Migration path is clear
- ✅ Breaking changes are documented
- ✅ jq examples work correctly
- ✅ Slash Commands updated
- ✅ MCP Server works

---

## Notes

**Estimated Time**: 2-3 days

**Key Risks**:
- Breaking changes may affect users
- Pipeline bugs could break all commands
- Code size underestimation

**Mitigation**:
- Comprehensive tests
- Migration guide
- Incremental refactoring (one command at a time)

**Ready to Begin**: ✅

---

**Checklist Version**: 1.0
**Last Updated**: 2025-10-05
**Phase**: 14 (Architecture Refactoring)
