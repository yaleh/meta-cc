# Refactoring Patterns Reference

**Source**: Extracted from Bootstrap-004 Refactoring Guide experiment

**Validation**: 2 refactorings, 10 pattern applications, 100% success rate

---

## Pattern Catalog

### 1. Extract Method

**Context**: Function with complexity >8, multiple responsibilities

**Problem**: Complex function difficult to understand, test, maintain

**Solution**: Extract cohesive code block into named helper function

**Validated**: YES (3 applications, -43% to -70% complexity reduction)

**Steps**:
1. Identify cohesive code block (5-10 lines doing one thing)
2. Ensure tests cover behavior to be extracted (write if missing)
3. Extract block to new function with descriptive name
4. Update original function to call helper
5. Run tests (must pass)
6. Write unit tests for extracted helper
7. Commit

**Example**:
```go
// Before (Complexity 10)
func calculateSequenceTimeSpan(occurrences []Occurrence, entries []SessionEntry, toolCalls []ToolCall) int {
    var timestamps []int64
    for _, occ := range occurrences {
        for _, idx := range occ.TurnIndices {
            if idx < len(entries) {
                if ts := entries[idx].Timestamp; ts > 0 {
                    timestamps = append(timestamps, ts)
                }
            }
        }
    }
    if len(timestamps) == 0 {
        return 0
    }
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

// After (Complexity 3)
func calculateSequenceTimeSpan(occurrences []Occurrence, entries []SessionEntry, toolCalls []ToolCall) int {
    timestamps := collectOccurrenceTimestamps(occurrences, entries, toolCalls)
    return findMinMaxTimestamps(timestamps)
}

func collectOccurrenceTimestamps(occurrences []Occurrence, entries []SessionEntry, toolCalls []ToolCall) []int64 {
    var timestamps []int64
    for _, occ := range occurrences {
        for _, idx := range occ.TurnIndices {
            if idx < len(entries) {
                if ts := entries[idx].Timestamp; ts > 0 {
                    timestamps = append(timestamps, ts)
                }
            }
        }
    }
    return timestamps
}

func findMinMaxTimestamps(timestamps []int64) int {
    if len(timestamps) == 0 {
        return 0
    }
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

**Benefits**:
- Complexity: 10 → 3 (-70%)
- Testability: Each helper testable independently
- Readability: Function name documents intent

**Transferability**: Universal (all languages)

---

### 2. Characterization Tests

**Context**: Legacy code or complex function without adequate tests

**Problem**: Can't refactor safely without knowing exact behavior

**Solution**: Write tests documenting current behavior (even if incorrect)

**Validated**: YES (9 tests, 100% regression prevention)

**Steps**:
1. Run function with typical inputs
2. Observe actual output (use debugger if needed)
3. Write test asserting current behavior
4. Test edge cases: empty input, nil, boundary values
5. Run tests (must all pass with current code)
6. Use tests as safety net during refactoring

**Example**:
```go
// Characterization tests for calculateSequenceTimeSpan
func TestCalculateSequenceTimeSpan_EmptyOccurrences(t *testing.T) {
    result := calculateSequenceTimeSpan(nil, nil, nil)
    assert.Equal(t, 0, result) // Current behavior: returns 0
}

func TestCalculateSequenceTimeSpan_SingleOccurrence(t *testing.T) {
    occurrences := []Occurrence{{TurnIndices: []int{0}}}
    entries := []SessionEntry{{Timestamp: 1000}}
    result := calculateSequenceTimeSpan(occurrences, entries, nil)
    assert.Equal(t, 0, result) // Single timestamp → 0 duration
}

func TestCalculateSequenceTimeSpan_MultipleOccurrences(t *testing.T) {
    occurrences := []Occurrence{
        {TurnIndices: []int{0}},
        {TurnIndices: []int{2}},
    }
    entries := []SessionEntry{
        {Timestamp: 1000},
        {Timestamp: 2000},
        {Timestamp: 5000},
    }
    result := calculateSequenceTimeSpan(occurrences, entries, nil)
    assert.Equal(t, 66, result) // (5000-1000)/60 = 66 minutes
}
```

**Anti-Pattern**: "Fix the tests to match new behavior"
- If tests fail after refactoring, rollback code, not tests!

**Transferability**: Universal (testing concept)

---

### 3. Simplify Conditionals

**Context**: Nested if/else, complex boolean expressions

**Problem**: Hard to read, hard to test all branches

**Solution**: Guard clauses, extract conditions, decompose booleans

**Validated**: Conceptual (not applied in practice, but documented)

**Techniques**:

**Guard Clauses** (early returns):
```go
// Before: Nested
if len(timestamps) > 0 {
    // 20 lines of logic
} else {
    return 0
}

// After: Guard clause
if len(timestamps) == 0 {
    return 0
}
// 20 lines of logic (not nested)
```

**Extract Condition to Variable**:
```go
// Before
if user.Age >= 18 && user.HasLicense && !user.Banned {
    // ...
}

// After
canDrive := user.Age >= 18 && user.HasLicense && !user.Banned
if canDrive {
    // ...
}
```

**Transferability**: Universal

---

### 4. Remove Duplication

**Context**: Duplicate code blocks (detected by `dupl`)

**Problem**: Changes must be made in multiple places, error-prone

**Solution**: Extract to shared helper function

**Validated**: NO (duplication not addressed - acknowledged gap)

**Steps**:
1. Detect duplication: `dupl -threshold 15 <package>/`
2. Identify exact duplicated lines
3. Extract to shared helper
4. Replace first occurrence, run tests
5. Replace second occurrence, run tests
6. Repeat for all occurrences
7. Commit after each replacement (incremental safety)

**Example**:
```go
// Before: Duplication
// File A
if ts > 0 {
    timestamps = append(timestamps, ts)
}

// File B
if ts > 0 {
    timestamps = append(timestamps, ts)
}

// After: Extracted helper
func appendIfValid(timestamps []int64, ts int64) []int64 {
    if ts > 0 {
        return append(timestamps, ts)
    }
    return timestamps
}

// File A: timestamps = appendIfValid(timestamps, ts)
// File B: timestamps = appendIfValid(timestamps, ts)
```

**Transferability**: Universal

---

### 5. Extract Variable

**Context**: Complex expressions or intermediate results

**Problem**: Expression hard to understand, repeated calculations

**Solution**: Extract to named variable

**Validated**: YES (2 applications in refactorings)

**Example**:
```go
// Before
return int((maxTs - minTs) / 60)

// After
durationSeconds := maxTs - minTs
durationMinutes := int(durationSeconds / 60)
return durationMinutes
```

**Transferability**: Universal

---

### 6. Decompose Boolean

**Context**: Complex boolean conditions

**Problem**: Hard to understand intent

**Solution**: Extract sub-conditions to descriptive variables

**Validated**: YES (1 application)

**Example**:
```go
// Before
if idx < len(entries) && entries[idx].Timestamp > 0 {
    // ...
}

// After
indexValid := idx < len(entries)
hasTimestamp := indexValid && entries[idx].Timestamp > 0
if hasTimestamp {
    // ...
}
```

**Transferability**: Universal

---

### 7. Introduce Helper Function

**Context**: Same as Extract Method, but emphasizes helper creation

**Problem**: Main function doing too much

**Solution**: Create focused helper with single responsibility

**Validated**: YES (3 applications, same as Extract Method)

**Benefits**: Reusability, testability, clarity

**Transferability**: Universal

---

### 8. Inline Temporary

**Context**: Unnecessary intermediate variables

**Problem**: Variable used once, adds no clarity

**Solution**: Remove variable, use expression directly

**Validated**: Conceptual (simplification pattern)

**Example**:
```go
// Before
temp := calculate()
return temp

// After
return calculate()
```

**When NOT to use**: If variable name adds clarity

**Transferability**: Universal

---

## Pattern Selection Guide

**Decision Tree**:

1. **Complexity >8?** → Try Extract Method first
2. **No tests?** → Write Characterization Tests first
3. **Nested conditionals?** → Simplify Conditionals
4. **Duplication detected?** → Remove Duplication
5. **Complex expression?** → Extract Variable or Decompose Boolean
6. **Unnecessary variable?** → Inline Temporary

**General Rule**: Extract Method is the workhorse (100% success rate, highest impact)

---

## Success Metrics by Pattern

| Pattern | Applications | Success Rate | Avg Complexity Reduction |
|---------|--------------|--------------|--------------------------|
| Extract Method | 3 | 100% | -56.5% |
| Characterization Tests | 9 tests | 100% | N/A (regression prevention) |
| Extract Variable | 2 | 100% | Minor |
| Decompose Boolean | 1 | 100% | Minor |
| Introduce Helper | 3 | 100% | -56.5% (same as Extract Method) |
| Simplify Conditionals | 0 | N/A | N/A |
| Remove Duplication | 0 | N/A | N/A |
| Inline Temporary | 0 | N/A | N/A |

**Overall**: 10 applications, 100% success rate

---

## Transferability Matrix

| Pattern | Go | Python | JavaScript | Rust | Java | C++ |
|---------|-----|--------|------------|------|------|-----|
| Extract Method | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Characterization Tests | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Simplify Conditionals | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Remove Duplication | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Extract Variable | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Decompose Boolean | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Introduce Helper | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Inline Temporary | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

**All patterns**: 100% language-agnostic (Martin Fowler catalog patterns)

---

**Source**: Bootstrap-004 Refactoring Guide (iterations 0-4)
**Validation**: 2 refactorings, meta-cc project (internal/query package)
**Version**: 1.0
**Date**: 2025-10-19
