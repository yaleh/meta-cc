# TDD Workflow and Coverage-Driven Development

**Version**: 2.0
**Source**: Bootstrap-002 Test Strategy Development
**Last Updated**: 2025-10-18

This document describes the Test-Driven Development (TDD) workflow and coverage-driven testing approach.

---

## Coverage-Driven Workflow

### Step 1: Generate Coverage Report

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out > coverage-by-func.txt
```

### Step 2: Identify Gaps

**Option A: Use automation tool**
```bash
./scripts/analyze-coverage-gaps.sh coverage.out --top 15
```

**Option B: Manual analysis**
```bash
# Find low-coverage functions
go tool cover -func=coverage.out | grep "^github.com" | awk '$NF < 60.0'

# Find zero-coverage functions
go tool cover -func=coverage.out | grep "0.0%"
```

### Step 3: Prioritize Targets

**Decision Tree**:
```
Is function critical to core functionality?
├─ YES: Is it error handling or validation?
│  ├─ YES: Priority 1 (80%+ coverage target)
│  └─ NO: Is it business logic?
│     ├─ YES: Priority 2 (75%+ coverage)
│     └─ NO: Priority 3 (60%+ coverage)
└─ NO: Is it infrastructure/initialization?
   ├─ YES: Priority 4 (test if easy, skip if hard)
   └─ NO: Priority 5 (skip)
```

**Priority Matrix**:
| Category | Target Coverage | Priority | Time/Test |
|----------|----------------|----------|-----------|
| Error Handling | 80-90% | P1 | 15 min |
| Business Logic | 75-85% | P2 | 12 min |
| CLI Handlers | 70-80% | P2 | 12 min |
| Integration | 70-80% | P3 | 20 min |
| Utilities | 60-70% | P3 | 8 min |
| Infrastructure | Best effort | P4 | 25 min |

### Step 4: Select Pattern

**Pattern Selection Decision Tree**:
```
What are you testing?
├─ CLI command with flags?
│  ├─ Multiple flag combinations? → Pattern 8 (Global Flag)
│  ├─ Integration test needed? → Pattern 7 (CLI Command)
│  └─ Command execution? → Pattern 7 (CLI Command)
├─ Error paths?
│  ├─ Multiple error scenarios? → Pattern 4 (Error Path) + Pattern 2 (Table-Driven)
│  └─ Single error case? → Pattern 4 (Error Path)
├─ Unit function?
│  ├─ Multiple inputs? → Pattern 2 (Table-Driven)
│  └─ Single input? → Pattern 1 (Unit Test)
├─ External dependency?
│  └─ → Pattern 6 (Dependency Injection)
└─ Integration flow?
   └─ → Pattern 3 (Integration Test)
```

### Step 5: Generate Test

**Option A: Use automation tool**
```bash
./scripts/generate-test.sh FunctionName --pattern PATTERN --scenarios N
```

**Option B: Manual from template**
- Copy pattern template from patterns.md
- Adapt to function signature
- Fill in test data

### Step 6: Implement Test

1. Fill in TODO comments
2. Add test data (inputs, expected outputs)
3. Customize assertions
4. Add edge cases

### Step 7: Verify Coverage Impact

```bash
# Run tests
go test ./package/...

# Generate new coverage
go test -coverprofile=new_coverage.out ./...

# Compare
echo "Old coverage:"
go tool cover -func=coverage.out | tail -1

echo "New coverage:"
go tool cover -func=new_coverage.out | tail -1

# Show improved functions
diff <(go tool cover -func=coverage.out) <(go tool cover -func=new_coverage.out) | grep "^>"
```

### Step 8: Track Metrics

**Per Test Batch**:
- Pattern(s) used
- Time spent (actual)
- Coverage increase achieved
- Issues encountered

**Example Log**:
```
Date: 2025-10-18
Batch: Validation error paths (4 tests)
Pattern: Error Path + Table-Driven
Time: 50 min (estimated 60 min) → 17% faster
Coverage: internal/validation 57.9% → 75.2% (+17.3%)
Total coverage: 72.3% → 73.5% (+1.2%)
Efficiency: 0.3% per test
Issues: None
Lessons: Table-driven error tests very efficient
```

---

## Red-Green-Refactor TDD Cycle

### Overview

The classic TDD cycle consists of three phases:

1. **Red**: Write a failing test
2. **Green**: Write minimal code to make it pass
3. **Refactor**: Improve code while keeping tests green

### Phase 1: Red (Write Failing Test)

**Goal**: Define expected behavior through a test that fails

```go
func TestValidateEmail_ValidFormat(t *testing.T) {
    // Write test BEFORE implementation exists
    email := "user@example.com"

    err := ValidateEmail(email)  // Function doesn't exist yet

    if err != nil {
        t.Errorf("ValidateEmail(%s) returned error: %v", email, err)
    }
}
```

**Run test**:
```bash
$ go test ./...
# Compilation error: ValidateEmail undefined
```

**Checklist for Red Phase**:
- [ ] Test clearly describes expected behavior
- [ ] Test compiles (stub function if needed)
- [ ] Test fails for the right reason
- [ ] Failure message is clear

### Phase 2: Green (Make It Pass)

**Goal**: Write simplest possible code to make test pass

```go
func ValidateEmail(email string) error {
    // Minimal implementation
    if !strings.Contains(email, "@") {
        return fmt.Errorf("invalid email: missing @")
    }
    return nil
}
```

**Run test**:
```bash
$ go test ./...
PASS
```

**Checklist for Green Phase**:
- [ ] Test passes
- [ ] Implementation is minimal (no over-engineering)
- [ ] No premature optimization
- [ ] All existing tests still pass

### Phase 3: Refactor (Improve Code)

**Goal**: Improve code quality without changing behavior

```go
func ValidateEmail(email string) error {
    // Refactor: Use regex for proper validation
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(email) {
        return fmt.Errorf("invalid email format: %s", email)
    }
    return nil
}
```

**Run tests**:
```bash
$ go test ./...
PASS  # All tests still pass after refactoring
```

**Checklist for Refactor Phase**:
- [ ] Code is more readable
- [ ] Duplication eliminated
- [ ] All tests still pass
- [ ] No new functionality added

---

## TDD for New Features

### Example: Add Email Validation Feature

**Iteration 1: Basic Structure**

1. **Red**: Test for valid email
```go
func TestValidateEmail_ValidFormat(t *testing.T) {
    err := ValidateEmail("user@example.com")
    if err != nil {
        t.Errorf("valid email rejected: %v", err)
    }
}
```

2. **Green**: Minimal implementation
```go
func ValidateEmail(email string) error {
    if !strings.Contains(email, "@") {
        return fmt.Errorf("invalid email")
    }
    return nil
}
```

3. **Refactor**: Extract constant
```go
const emailPattern = "@"

func ValidateEmail(email string) error {
    if !strings.Contains(email, emailPattern) {
        return fmt.Errorf("invalid email")
    }
    return nil
}
```

**Iteration 2: Add Edge Cases**

1. **Red**: Test for empty email
```go
func TestValidateEmail_Empty(t *testing.T) {
    err := ValidateEmail("")
    if err == nil {
        t.Error("empty email should be invalid")
    }
}
```

2. **Green**: Add empty check
```go
func ValidateEmail(email string) error {
    if email == "" {
        return fmt.Errorf("email cannot be empty")
    }
    if !strings.Contains(email, "@") {
        return fmt.Errorf("invalid email")
    }
    return nil
}
```

3. **Refactor**: Use regex
```go
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) error {
    if email == "" {
        return fmt.Errorf("email cannot be empty")
    }
    if !emailRegex.MatchString(email) {
        return fmt.Errorf("invalid email format")
    }
    return nil
}
```

**Iteration 3: Add More Cases**

Convert to table-driven test:

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid", "user@example.com", false},
        {"empty", "", true},
        {"no @", "userexample.com", true},
        {"no domain", "user@", true},
        {"no user", "@example.com", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail(%s) error = %v, wantErr %v",
                    tt.email, err, tt.wantErr)
            }
        })
    }
}
```

---

## TDD for Bug Fixes

### Workflow

1. **Reproduce bug with test** (Red)
2. **Fix bug** (Green)
3. **Refactor if needed** (Refactor)
4. **Verify bug doesn't regress** (Test stays green)

### Example: Fix Nil Pointer Bug

**Step 1: Write failing test that reproduces bug**

```go
func TestProcessData_NilInput(t *testing.T) {
    // This currently crashes with nil pointer
    _, err := ProcessData(nil)

    if err == nil {
        t.Error("ProcessData(nil) should return error, not crash")
    }
}
```

**Run test**:
```bash
$ go test ./...
panic: runtime error: invalid memory address or nil pointer dereference
FAIL
```

**Step 2: Fix the bug**

```go
func ProcessData(input *Input) (Result, error) {
    // Add nil check
    if input == nil {
        return Result{}, fmt.Errorf("input cannot be nil")
    }

    // Original logic...
    return result, nil
}
```

**Run test**:
```bash
$ go test ./...
PASS
```

**Step 3: Add more edge cases**

```go
func TestProcessData_ErrorCases(t *testing.T) {
    tests := []struct {
        name    string
        input   *Input
        wantErr bool
        errMsg  string
    }{
        {
            name:    "nil input",
            input:   nil,
            wantErr: true,
            errMsg:  "cannot be nil",
        },
        {
            name:    "empty input",
            input:   &Input{},
            wantErr: true,
            errMsg:  "empty",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := ProcessData(tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("ProcessData() error = %v, wantErr %v", err, tt.wantErr)
            }

            if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
                t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
            }
        })
    }
}
```

---

## Integration with Coverage-Driven Development

TDD and coverage-driven approaches complement each other:

### Pure TDD (New Feature Development)

**When**: Building new features from scratch

**Workflow**: Red → Green → Refactor (repeat)

**Focus**: Design through tests, emergent architecture

### Coverage-Driven (Existing Codebase)

**When**: Improving test coverage of existing code

**Workflow**: Analyze coverage → Prioritize → Write tests → Verify

**Focus**: Systematic gap closure, efficiency

### Hybrid Approach (Recommended)

**For new features**:
1. Use TDD to drive design
2. Track coverage as you go
3. Use coverage tools to identify blind spots

**For existing code**:
1. Use coverage-driven to systematically add tests
2. Apply TDD for any refactoring
3. Apply TDD for bug fixes

---

## Best Practices

### Do's

✅ Write test before code (for new features)
✅ Keep Red phase short (minutes, not hours)
✅ Make smallest possible change to get to Green
✅ Refactor frequently
✅ Run all tests after each change
✅ Commit after each successful Red-Green-Refactor cycle

### Don'ts

❌ Skip the Red phase (writing tests for existing working code is not TDD)
❌ Write multiple tests before making them pass
❌ Write too much code in Green phase
❌ Refactor while tests are red
❌ Skip Refactor phase
❌ Ignore test failures

---

## Common Challenges

### Challenge 1: Test Takes Too Long to Write

**Symptom**: Spending 30+ minutes on single test

**Causes**:
- Testing too much at once
- Complex setup required
- Unclear requirements

**Solutions**:
- Break into smaller tests
- Create test helpers for setup
- Clarify requirements before writing test

### Challenge 2: Can't Make Test Pass Without Large Changes

**Symptom**: Green phase requires extensive code changes

**Causes**:
- Test is too ambitious
- Existing code not designed for testability
- Missing intermediate steps

**Solutions**:
- Write smaller test
- Refactor existing code first (with existing tests passing)
- Add intermediate tests to build up gradually

### Challenge 3: Tests Pass But Coverage Doesn't Improve

**Symptom**: Writing tests but coverage metrics don't increase

**Causes**:
- Testing already-covered code paths
- Tests not exercising target functions
- Indirect coverage already exists

**Solutions**:
- Check per-function coverage: `go tool cover -func=coverage.out`
- Focus on 0% coverage functions
- Use coverage tools to identify true gaps

---

**Source**: Bootstrap-002 Test Strategy Development
**Framework**: BAIME (Bootstrapped AI Methodology Engineering)
**Status**: Production-ready, validated through 4 iterations
