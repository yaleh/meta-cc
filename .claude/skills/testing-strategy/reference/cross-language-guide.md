# Cross-Language Test Strategy Adaptation

**Version**: 2.0
**Source**: Bootstrap-002 Test Strategy Development
**Last Updated**: 2025-10-18

This document provides guidance for adapting test patterns and methodology to different programming languages and frameworks.

---

## Transferability Overview

### Universal Concepts (100% Transferable)

The following concepts apply to ALL languages:

1. **Coverage-Driven Workflow**: Analyze → Prioritize → Test → Verify
2. **Priority Matrix**: P1 (error handling) → P4 (infrastructure)
3. **Pattern-Based Testing**: Structured approaches to common scenarios
4. **Table-Driven Approach**: Multiple scenarios with shared logic
5. **Error Path Testing**: Systematic edge case coverage
6. **Dependency Injection**: Mock external dependencies
7. **Quality Standards**: Test structure and best practices
8. **TDD Cycle**: Red-Green-Refactor

### Language-Specific Elements (Require Adaptation)

1. **Syntax and Imports**: Language-specific
2. **Testing Framework APIs**: Different per ecosystem
3. **Coverage Tool Commands**: Language-specific tools
4. **Mock Implementation**: Different mocking libraries
5. **Build/Run Commands**: Different toolchains

---

## Go → Python Adaptation

### Transferability: 80-90%

### Testing Framework Mapping

| Go Concept | Python Equivalent |
|------------|------------------|
| `testing` package | `unittest` or `pytest` |
| `t.Run()` subtests | `pytest` parametrize or `unittest` subtests |
| `t.Helper()` | `pytest` fixtures |
| `t.Cleanup()` | `pytest` fixtures with yield or `unittest` tearDown |
| Table-driven tests | `@pytest.mark.parametrize` |

### Pattern Adaptations

#### Pattern 1: Unit Test

**Go**:
```go
func TestFunction(t *testing.T) {
    result := Function(input)
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

**Python (pytest)**:
```python
def test_function():
    result = function(input)
    assert result == expected, f"got {result}, want {expected}"
```

**Python (unittest)**:
```python
class TestFunction(unittest.TestCase):
    def test_function(self):
        result = function(input)
        self.assertEqual(result, expected)
```

#### Pattern 2: Table-Driven Test

**Go**:
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    int
        expected int
    }{
        {"case1", 1, 2},
        {"case2", 2, 4},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Function(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

**Python (pytest)**:
```python
@pytest.mark.parametrize("input,expected", [
    (1, 2),
    (2, 4),
])
def test_function(input, expected):
    result = function(input)
    assert result == expected
```

**Python (unittest)**:
```python
class TestFunction(unittest.TestCase):
    def test_cases(self):
        cases = [
            ("case1", 1, 2),
            ("case2", 2, 4),
        ]
        for name, input, expected in cases:
            with self.subTest(name=name):
                result = function(input)
                self.assertEqual(result, expected)
```

#### Pattern 6: Dependency Injection (Mocking)

**Go**:
```go
type Executor interface {
    Execute(args Args) (Result, error)
}

type MockExecutor struct {
    Results map[string]Result
}

func (m *MockExecutor) Execute(args Args) (Result, error) {
    return m.Results[args.Key], nil
}
```

**Python (unittest.mock)**:
```python
from unittest.mock import Mock, MagicMock

def test_process():
    mock_executor = Mock()
    mock_executor.execute.return_value = expected_result

    result = process_data(mock_executor)

    assert result == expected
    mock_executor.execute.assert_called_once()
```

**Python (pytest-mock)**:
```python
def test_process(mocker):
    mock_executor = mocker.Mock()
    mock_executor.execute.return_value = expected_result

    result = process_data(mock_executor)

    assert result == expected
```

### Coverage Tools

**Go**:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out
```

**Python (pytest-cov)**:
```bash
pytest --cov=package --cov-report=term
pytest --cov=package --cov-report=html
pytest --cov=package --cov-report=term-missing
```

**Python (coverage.py)**:
```bash
coverage run -m pytest
coverage report
coverage html
```

---

## Go → JavaScript/TypeScript Adaptation

### Transferability: 75-85%

### Testing Framework Mapping

| Go Concept | JavaScript/TypeScript Equivalent |
|------------|--------------------------------|
| `testing` package | Jest, Mocha, Vitest |
| `t.Run()` subtests | `describe()` / `it()` blocks |
| Table-driven tests | `test.each()` (Jest) |
| Mocking | Jest mocks, Sinon |
| Coverage | Jest built-in, nyc/istanbul |

### Pattern Adaptations

#### Pattern 1: Unit Test

**Go**:
```go
func TestFunction(t *testing.T) {
    result := Function(input)
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

**JavaScript (Jest)**:
```javascript
test('function returns expected result', () => {
    const result = functionUnderTest(input);
    expect(result).toBe(expected);
});
```

**TypeScript (Jest)**:
```typescript
describe('functionUnderTest', () => {
    it('returns expected result', () => {
        const result = functionUnderTest(input);
        expect(result).toBe(expected);
    });
});
```

#### Pattern 2: Table-Driven Test

**Go**:
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    int
        expected int
    }{
        {"case1", 1, 2},
        {"case2", 2, 4},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Function(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

**JavaScript/TypeScript (Jest)**:
```typescript
describe('functionUnderTest', () => {
    test.each([
        ['case1', 1, 2],
        ['case2', 2, 4],
    ])('%s: input %i should return %i', (name, input, expected) => {
        const result = functionUnderTest(input);
        expect(result).toBe(expected);
    });
});
```

**Alternative with object syntax**:
```typescript
describe('functionUnderTest', () => {
    test.each([
        { name: 'case1', input: 1, expected: 2 },
        { name: 'case2', input: 2, expected: 4 },
    ])('$name', ({ input, expected }) => {
        const result = functionUnderTest(input);
        expect(result).toBe(expected);
    });
});
```

#### Pattern 6: Dependency Injection (Mocking)

**Go**:
```go
type MockExecutor struct {
    Results map[string]Result
}
```

**JavaScript (Jest)**:
```javascript
const mockExecutor = {
    execute: jest.fn((args) => {
        return results[args.key];
    })
};

test('processData uses executor', () => {
    const result = processData(mockExecutor, testData);

    expect(result).toBe(expected);
    expect(mockExecutor.execute).toHaveBeenCalledWith(testData);
});
```

**TypeScript (Jest)**:
```typescript
const mockExecutor: Executor = {
    execute: jest.fn((args: Args): Result => {
        return results[args.key];
    })
};
```

### Coverage Tools

**Jest (built-in)**:
```bash
jest --coverage
jest --coverage --coverageReporters=html
jest --coverage --coverageReporters=text-summary
```

**nyc (for Mocha)**:
```bash
nyc mocha
nyc report --reporter=html
nyc report --reporter=text-summary
```

---

## Go → Rust Adaptation

### Transferability: 70-80%

### Testing Framework Mapping

| Go Concept | Rust Equivalent |
|------------|----------------|
| `testing` package | Built-in `#[test]` |
| `t.Run()` subtests | `#[test]` functions |
| Table-driven tests | Loop or macro |
| Error handling | `Result<T, E>` assertions |
| Mocking | `mockall` crate |

### Pattern Adaptations

#### Pattern 1: Unit Test

**Go**:
```go
func TestFunction(t *testing.T) {
    result := Function(input)
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

**Rust**:
```rust
#[test]
fn test_function() {
    let result = function(input);
    assert_eq!(result, expected);
}
```

#### Pattern 2: Table-Driven Test

**Go**:
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    int
        expected int
    }{
        {"case1", 1, 2},
        {"case2", 2, 4},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Function(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

**Rust**:
```rust
#[test]
fn test_function() {
    let tests = vec![
        ("case1", 1, 2),
        ("case2", 2, 4),
    ];

    for (name, input, expected) in tests {
        let result = function(input);
        assert_eq!(result, expected, "test case: {}", name);
    }
}
```

**Rust (using rstest crate)**:
```rust
use rstest::rstest;

#[rstest]
#[case(1, 2)]
#[case(2, 4)]
fn test_function(#[case] input: i32, #[case] expected: i32) {
    let result = function(input);
    assert_eq!(result, expected);
}
```

#### Pattern 4: Error Path Testing

**Go**:
```go
func TestFunction_Error(t *testing.T) {
    _, err := Function(invalidInput)
    if err == nil {
        t.Error("expected error, got nil")
    }
}
```

**Rust**:
```rust
#[test]
fn test_function_error() {
    let result = function(invalid_input);
    assert!(result.is_err(), "expected error");
}

#[test]
#[should_panic(expected = "invalid input")]
fn test_function_panic() {
    function_that_panics(invalid_input);
}
```

### Coverage Tools

**tarpaulin**:
```bash
cargo tarpaulin --out Html
cargo tarpaulin --out Lcov
```

**llvm-cov (nightly)**:
```bash
cargo +nightly llvm-cov --html
cargo +nightly llvm-cov --text
```

---

## Adaptation Checklist

When adapting test methodology to a new language:

### Phase 1: Map Core Concepts

- [ ] Identify language testing framework (unittest, pytest, Jest, etc.)
- [ ] Map test structure (functions vs classes vs methods)
- [ ] Map assertion style (if/error vs assert vs expect)
- [ ] Map test organization (subtests, parametrize, describe/it)
- [ ] Map mocking approach (interfaces vs dependency injection vs mocks)

### Phase 2: Adapt Patterns

- [ ] Translate Pattern 1 (Unit Test) to target language
- [ ] Translate Pattern 2 (Table-Driven) to target language
- [ ] Translate Pattern 4 (Error Path) to target language
- [ ] Identify language-specific patterns (e.g., decorator tests in Python)
- [ ] Document language-specific gotchas

### Phase 3: Adapt Tools

- [ ] Identify coverage tool (coverage.py, Jest, tarpaulin, etc.)
- [ ] Create coverage gap analyzer script for target language
- [ ] Create test generator script for target language
- [ ] Adapt automation workflow to target toolchain

### Phase 4: Adapt Workflow

- [ ] Update coverage generation commands
- [ ] Update test execution commands
- [ ] Update IDE/editor integration
- [ ] Update CI/CD pipeline
- [ ] Document language-specific workflow

### Phase 5: Validate

- [ ] Apply methodology to sample project
- [ ] Measure effectiveness (time per test, coverage increase)
- [ ] Document lessons learned
- [ ] Refine patterns based on feedback

---

## Language-Specific Considerations

### Python

**Strengths**:
- `pytest` parametrize is excellent for table-driven tests
- Fixtures provide powerful setup/teardown
- `unittest.mock` is very flexible

**Challenges**:
- Dynamic typing can hide errors caught at compile time in Go
- Coverage tools sometimes struggle with decorators
- Import-time code execution complicates testing

**Tips**:
- Use type hints to catch errors early
- Use `pytest-cov` for coverage
- Use `pytest-mock` for simpler mocking
- Test module imports separately

### JavaScript/TypeScript

**Strengths**:
- Jest has excellent built-in mocking
- `test.each` is natural for table-driven tests
- TypeScript adds compile-time type safety

**Challenges**:
- Async/Promise handling adds complexity
- Module mocking can be tricky
- Coverage of TypeScript types vs runtime code

**Tips**:
- Use TypeScript for better IDE support and type safety
- Use Jest's `async/await` test support
- Use `ts-jest` for TypeScript testing
- Mock at module boundaries, not implementation details

### Rust

**Strengths**:
- Built-in testing framework is simple and fast
- Compile-time guarantees reduce need for some tests
- `Result<T, E>` makes error testing explicit

**Challenges**:
- Less mature test tooling ecosystem
- Mocking requires more setup (mockall crate)
- Lifetime and ownership can complicate test data

**Tips**:
- Use `rstest` for parametrized tests
- Use `mockall` for mocking traits
- Use integration tests (`tests/` directory) for public API
- Use unit tests for internal logic

---

## Effectiveness Across Languages

### Expected Methodology Transfer

| Language | Pattern Transfer | Tool Adaptation | Overall Transfer |
|----------|-----------------|----------------|-----------------|
| **Python** | 95% | 80% | 80-90% |
| **JavaScript/TypeScript** | 90% | 75% | 75-85% |
| **Rust** | 85% | 70% | 70-80% |
| **Java** | 90% | 80% | 80-85% |
| **C#** | 90% | 85% | 85-90% |
| **Ruby** | 85% | 75% | 75-80% |

### Time to Adapt

| Activity | Estimated Time |
|----------|---------------|
| Map core concepts | 2-3 hours |
| Adapt patterns | 3-4 hours |
| Create automation tools | 4-6 hours |
| Validate on sample project | 2-3 hours |
| Document adaptations | 1-2 hours |
| **Total** | **12-18 hours** |

---

**Source**: Bootstrap-002 Test Strategy Development
**Framework**: BAIME (Bootstrapped AI Methodology Engineering)
**Status**: Production-ready, validated through 4 iterations
