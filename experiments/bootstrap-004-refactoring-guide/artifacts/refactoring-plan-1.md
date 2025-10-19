# Refactoring Plan - Iteration 1

**Created**: 2025-10-19
**Targets**: 2 high-priority refactorings
**Estimated Duration**: 75 minutes

---

## Target 1: Eliminate buildContext* Duplication

### Location
**File**: `internal/query/context.go`
**Lines**: 83-100 (`buildContextBefore`) and 103-120 (`buildContextAfter`)
**Duplication Severity**: HIGH (95% identical code)

### Code Smell Identified
**Type**: Duplicated Code (CS-001)
**Impact**:
- Violates DRY principle
- Double maintenance burden (changes must be synchronized)
- 18 lines of duplicated logic
- Error-prone (easy to update one and forget the other)

### Current Implementation Analysis

**buildContextBefore**:
```go
func buildContextBefore(entries []parser.SessionEntry, errorTurn, window int, turnIndex map[string]int) []TurnPreview {
	var previews []TurnPreview
	for _, entry := range entries {
		if !entry.IsMessage() { continue }
		turn := turnIndex[entry.UUID]
		if turn >= errorTurn || turn < errorTurn-window { continue }  // ← ONLY DIFFERENCE
		previews = append(previews, buildTurnPreview(entry, turn))
	}
	return previews
}
```

**buildContextAfter**:
```go
func buildContextAfter(entries []parser.SessionEntry, errorTurn, window int, turnIndex map[string]int) []TurnPreview {
	var previews []TurnPreview
	for _, entry := range entries {
		if !entry.IsMessage() { continue }
		turn := turnIndex[entry.UUID]
		if turn <= errorTurn || turn > errorTurn+window { continue }  // ← ONLY DIFFERENCE
		previews = append(previews, buildTurnPreview(entry, turn))
	}
	return previews
}
```

**Key Observation**: Only difference is the conditional logic in line 92 vs 112.

### Refactoring Technique
**Extract Method with Direction Parameter**

Create a unified function that accepts a direction parameter:
```go
func buildContextWindow(entries []parser.SessionEntry, errorTurn, window int,
                       turnIndex map[string]int, direction string) []TurnPreview
```

### Incremental Steps

#### Step 1: Verify Test Coverage
**Action**: Examine existing tests for `buildContextBefore` and `buildContextAfter`
**Safety**: Ensure edge cases are covered (empty entries, boundary conditions, window=0)
**Checkpoint**: Run `go test -v ./internal/query/context_test.go`

#### Step 2: Extract Common Logic
**Action**: Create `buildContextWindow` function with direction logic
**Implementation**:
```go
// buildContextWindow builds context in specified direction from error turn
func buildContextWindow(entries []parser.SessionEntry, errorTurn, window int,
                       turnIndex map[string]int, direction string) []TurnPreview {
	var previews []TurnPreview

	for _, entry := range entries {
		if !entry.IsMessage() {
			continue
		}

		turn := turnIndex[entry.UUID]

		// Direction-specific filtering
		var skip bool
		if direction == "before" {
			skip = turn >= errorTurn || turn < errorTurn-window
		} else { // "after"
			skip = turn <= errorTurn || turn > errorTurn+window
		}

		if skip {
			continue
		}

		previews = append(previews, buildTurnPreview(entry, turn))
	}

	return previews
}
```
**Safety**: Run tests immediately after adding this function
**Checkpoint**: `go test ./internal/query/...`

#### Step 3: Refactor buildContextBefore
**Action**: Replace implementation with call to `buildContextWindow`
```go
func buildContextBefore(entries []parser.SessionEntry, errorTurn, window int, turnIndex map[string]int) []TurnPreview {
	return buildContextWindow(entries, errorTurn, window, turnIndex, "before")
}
```
**Safety**: Run tests to ensure no behavior change
**Checkpoint**: `go test -v ./internal/query/context_test.go`

#### Step 4: Refactor buildContextAfter
**Action**: Replace implementation with call to `buildContextWindow`
```go
func buildContextAfter(entries []parser.SessionEntry, errorTurn, window int, turnIndex map[string]int) []TurnPreview {
	return buildContextWindow(entries, errorTurn, window, turnIndex, "after")
}
```
**Safety**: Run full test suite
**Checkpoint**: `go test ./internal/query/...`

#### Step 5: Final Verification
**Action**: Run comprehensive tests and check behavior preservation
**Commands**:
```bash
go test -v ./internal/query/...
go test -cover ./internal/query/...
dupl -threshold 15 internal/query/context.go
```

### Safety Verification Plan
- [ ] All tests pass before starting
- [ ] Tests pass after each incremental step
- [ ] No new compiler errors introduced
- [ ] Coverage maintained or improved
- [ ] Duplication reduced (verify with dupl)
- [ ] Git commit after each successful step

### Expected Metrics Improvement
**Before**:
- Duplication: 1 clone group (18 lines)
- Functions: 2 nearly identical

**After**:
- Duplication: 0 clone groups (eliminated)
- Functions: 1 unified + 2 thin wrappers
- Lines saved: ~15-18 lines
- Maintainability: Single source of truth

### Estimated Effort
**Planned**: 30 minutes
**Breakdown**:
- Test verification: 5 minutes
- Extract method: 10 minutes
- Refactor wrappers: 10 minutes
- Final verification: 5 minutes

---

## Target 2: Simplify calculateSequenceTimeSpan

### Location
**File**: `internal/query/sequences.go`
**Lines**: 214-242
**Cyclomatic Complexity**: 11 (target: 7)

### Code Smell Identified
**Type**: Complex Function (CS-002)
**Impact**:
- Nested loops (2 levels deep)
- Multiple conditional branches
- Mixes time calculation with data traversal
- Hard to understand at first glance
- Difficult to test edge cases independently
- O(n*m) complexity (potential performance concern)

### Current Implementation Analysis

```go
func calculateSequenceTimeSpan(occurrences []types.SequenceOccurrence, entries []parser.SessionEntry, toolCalls []toolCallWithTurn) int {
	if len(occurrences) == 0 { return 0 }  // Branch 1

	var minTs, maxTs int64

	for _, occ := range occurrences {              // Loop 1
		for _, tc := range toolCalls {             // Loop 2 (nested)
			if tc.turn == occ.StartTurn || tc.turn == occ.EndTurn {  // Branch 2
				ts := getToolCallTimestamp(entries, tc.uuid)
				if minTs == 0 || ts < minTs {      // Branch 3
					minTs = ts
				}
				if ts > maxTs {                    // Branch 4
					maxTs = ts
				}
			}
		}
	}

	if minTs == 0 || maxTs == 0 { return 0 }       // Branch 5

	return int((maxTs - minTs) / 60)
}
```

**Complexity Contributors**:
- 1 early return check
- 2 nested loops
- 3 conditional branches inside loops
- 1 final validation check

### Refactoring Technique
**Extract Method + Simplify Logic**

Split into:
1. Timestamp collection (extract to helper)
2. Min/max calculation (simplified logic)
3. Span calculation (final step)

### Incremental Steps

#### Step 1: Verify Test Coverage
**Action**: Check tests for `calculateSequenceTimeSpan`
**Safety**: Ensure edge cases are tested:
- Empty occurrences
- Single occurrence
- Multiple occurrences
- Missing timestamps (minTs=0, maxTs=0)
**Checkpoint**: `go test -v ./internal/query/sequences_test.go -run TestCalculateSequenceTimeSpan` (if exists)

#### Step 2: Extract Timestamp Lookup Helper
**Action**: Create helper to find timestamp for a turn
**Implementation**:
```go
// findTimestampForTurn finds the timestamp for a specific turn
func findTimestampForTurn(entries []parser.SessionEntry, toolCalls []toolCallWithTurn, turn int) int64 {
	for _, tc := range toolCalls {
		if tc.turn == turn {
			return getToolCallTimestamp(entries, tc.uuid)
		}
	}
	return 0
}
```
**Safety**: Add unit test for this helper if needed
**Checkpoint**: `go test ./internal/query/...`

#### Step 3: Simplify Main Function
**Action**: Refactor to use helper and simplify logic
**Implementation**:
```go
func calculateSequenceTimeSpan(occurrences []types.SequenceOccurrence, entries []parser.SessionEntry, toolCalls []toolCallWithTurn) int {
	if len(occurrences) == 0 {
		return 0
	}

	// Collect all relevant timestamps
	var timestamps []int64

	for _, occ := range occurrences {
		// Get timestamps for start and end of each occurrence
		startTs := findTimestampForTurn(entries, toolCalls, occ.StartTurn)
		endTs := findTimestampForTurn(entries, toolCalls, occ.EndTurn)

		if startTs > 0 {
			timestamps = append(timestamps, startTs)
		}
		if endTs > 0 && endTs != startTs {
			timestamps = append(timestamps, endTs)
		}
	}

	if len(timestamps) == 0 {
		return 0
	}

	// Find min and max
	minTs := timestamps[0]
	maxTs := timestamps[0]
	for _, ts := range timestamps[1:] {
		if ts < minTs {
			minTs = ts
		}
		if ts > maxTs {
			maxTs = ts
		}
	}

	return int((maxTs - minTs) / 60)
}
```
**Complexity Reduction**: Should reduce from 11 to ~7
**Safety**: Run tests after each change
**Checkpoint**: `go test -v ./internal/query/sequences_test.go`

#### Step 4: Performance Verification
**Action**: Ensure performance is maintained or improved
**Note**: New implementation is O(n) for timestamp collection + O(m) for finding, vs O(n*m) nested loops
**Expected**: Same or better performance

#### Step 5: Final Verification
**Action**: Run comprehensive tests
**Commands**:
```bash
go test -v ./internal/query/...
gocyclo internal/query/sequences.go
go test -bench=. ./internal/query/... -benchmem (if benchmarks exist)
```

### Safety Verification Plan
- [ ] All tests pass before starting
- [ ] Helper function tested independently
- [ ] Tests pass after refactoring main function
- [ ] Complexity measured with gocyclo (should be ≤7)
- [ ] No performance regression
- [ ] Git commit after successful refactoring

### Expected Metrics Improvement
**Before**:
- Cyclomatic complexity: 11
- Lines: 29
- Nested loops: 2 levels
- Conditional branches: 5

**After**:
- Cyclomatic complexity: ~7 (36% reduction)
- Lines: ~35 (reorganized, not reduced)
- Nested loops: 0 (eliminated)
- Conditional branches: ~3
- Clarity: HIGH (separated concerns)

### Estimated Effort
**Planned**: 45 minutes
**Breakdown**:
- Test verification: 10 minutes
- Extract helper: 10 minutes
- Refactor main function: 15 minutes
- Performance verification: 5 minutes
- Final verification: 5 minutes

---

## Overall Safety Protocol

### Before Starting
1. ✅ Ensure all tests pass: `go test ./internal/query/...`
2. ✅ Check git status is clean: `git status`
3. ✅ Create feature branch: `git checkout -b refactor/iteration-1`

### During Refactoring
1. ✅ Make ONE small change at a time
2. ✅ Run tests after EACH change
3. ✅ Git commit after each successful step
4. ✅ If tests fail → rollback immediately
5. ✅ Never proceed with failing tests

### After Completing
1. ✅ Run full test suite
2. ✅ Measure new metrics (complexity, duplication, coverage)
3. ✅ Compare with baseline
4. ✅ Document patterns observed
5. ✅ Commit final state

### Rollback Procedure
```bash
# If a step fails
git diff                    # Review changes
git restore <file>          # Restore specific file
# OR
git reset --hard HEAD       # Reset all changes (last resort)

# If tests fail after commit
git revert HEAD             # Revert last commit
```

---

## Expected Timeline

| Task | Estimated Time | Checkpoint |
|------|----------------|------------|
| Target 1: buildContext* duplication | 30 min | Tests pass |
| Target 2: calculateSequenceTimeSpan | 45 min | Complexity ≤7 |
| **Total** | **75 min** | All tests pass |

---

## Success Criteria

### Target 1 Success
- ✅ Duplication eliminated (18 lines reduced)
- ✅ All tests pass
- ✅ Single source of truth established
- ✅ No behavior changes

### Target 2 Success
- ✅ Complexity reduced from 11 to ≤7
- ✅ All tests pass
- ✅ Logic simplified and clarified
- ✅ No performance regression

### Overall Iteration Success
- ✅ Both targets completed
- ✅ 100% test pass rate
- ✅ Metrics improved (complexity down, duplication down)
- ✅ Patterns documented
- ✅ Ready for methodology codification

---

**Status**: ✅ Plan Complete - Ready for Execution
**Next**: Execute Target 1
