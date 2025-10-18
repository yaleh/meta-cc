# Cross-Language Adaptation Guide

**Version**: 1.0
**Created**: 2025-10-18 (Iteration 5)
**Experiment**: Bootstrap-002 Test Strategy Development
**Purpose**: Enable methodology transfer from Go to Python, JavaScript, Rust, Java

---

## Overview

This guide provides adaptation instructions for transferring the test strategy methodology to different programming languages and test frameworks. The methodology's core principles (coverage-driven workflow, pattern library, automation) remain universal, but implementation details vary.

**Adaptation Effort Estimates**:
- **Go → Rust**: 10-15% modification (similar tooling, type systems)
- **Go → Java**: 15-25% modification (JUnit patterns differ)
- **Go → Python**: 25-35% modification (pytest differs significantly)
- **Go → JavaScript**: 30-40% modification (async patterns, jest differs)

---

## Universal Components (0% Adaptation)

These transfer unchanged to **any language**:

### 1. Coverage-Driven Workflow (8 Steps)
```
1. Run coverage analysis
2. Identify gaps (prioritize by category: error-handling > business-logic > utility)
3. Select appropriate test pattern
4. Generate test scaffold
5. Implement test
6. Run tests and verify
7. Check coverage improvement
8. Iterate until target reached
```

**Transfer**: ✅ 100% - Workflow is language-agnostic

---

### 2. Priority Matrix (P1-P4 Categorization)
```
P1: Error handling, validation (target: 80-90% coverage)
P2: Business logic, core algorithms (target: 75-85% coverage)
P3: Utilities, helpers (target: 60-70% coverage)
P4: Infrastructure, logging (target: best effort)
```

**Transfer**: ✅ 100% - Categorization applies universally

---

### 3. Quality Standards Checklist
```
- Test pass rate ≥95%
- No flaky tests (<5% tolerance)
- Clear test names (describe scenario)
- DRY principle (test helpers, fixtures)
- Fast execution (<2 min for full suite)
- Isolated tests (no interdependencies)
- Maintainable (avoid brittle selectors, magic numbers)
```

**Transfer**: ✅ 100% - Best practices apply to all languages

---

## Pattern Library Adaptation

### Pattern 1: Unit Test Pattern

**Go (testing package)**:
```go
func TestFunctionName_Scenario(t *testing.T) {
    input := createInput()
    result, err := FunctionUnderTest(input)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

**Python (pytest)**:
```python
def test_function_name_scenario():
    input_data = create_input()
    result = function_under_test(input_data)
    assert result == expected, f"got {result}, want {expected}"
```

**JavaScript (jest)**:
```javascript
test('functionName scenario', () => {
    const input = createInput();
    const result = functionUnderTest(input);
    expect(result).toBe(expected);
});
```

**Rust (built-in)**:
```rust
#[test]
fn test_function_name_scenario() {
    let input = create_input();
    let result = function_under_test(input).unwrap();
    assert_eq!(result, expected);
}
```

**Adaptation Effort**: 5-10% (syntax only, concept identical)

---

### Pattern 2: Table-Driven Test Pattern

**Go (testing package)**:
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
        wantErr  bool
    }{
        {name: "valid", input: valid, expected: result, wantErr: false},
        {name: "invalid", input: invalid, expected: nil, wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            if !tt.wantErr && result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

**Python (pytest with parametrize)**:
```python
@pytest.mark.parametrize("input_data,expected,should_raise", [
    (valid_input, expected_result, False),
    (invalid_input, None, True),
])
def test_function(input_data, expected, should_raise):
    if should_raise:
        with pytest.raises(Exception):
            function(input_data)
    else:
        result = function(input_data)
        assert result == expected
```

**JavaScript (jest with test.each)**:
```javascript
test.each([
    ['valid', validInput, expectedResult, false],
    ['invalid', invalidInput, null, true],
])('function %s', (name, input, expected, shouldThrow) => {
    if (shouldThrow) {
        expect(() => functionUnderTest(input)).toThrow();
    } else {
        const result = functionUnderTest(input);
        expect(result).toEqual(expected);
    }
});
```

**Rust (using rstest or similar)**:
```rust
#[rstest]
#[case("valid", valid_input(), expected_result())]
#[case("invalid", invalid_input(), should_error())]
fn test_function(#[case] name: &str, #[case] input: Input, #[case] expected: Result<Output, Error>) {
    let result = function(input);
    assert_eq!(result, expected);
}
```

**Adaptation Effort**: 15-25% (framework-specific parametrization syntax)

---

### Pattern 4: Error Path Pattern

**Go (testing package)**:
```go
func TestValidation_Errors(t *testing.T) {
    tests := []struct {
        name      string
        input     InputType
        wantError string
    }{
        {name: "empty", input: empty, wantError: "cannot be empty"},
        {name: "invalid format", input: bad, wantError: "invalid format"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Validate(tt.input)
            if err == nil {
                t.Fatal("expected error, got nil")
            }
            if !strings.Contains(err.Error(), tt.wantError) {
                t.Errorf("error = %v, want %v", err, tt.wantError)
            }
        })
    }
}
```

**Python (pytest)**:
```python
@pytest.mark.parametrize("input_data,expected_error", [
    (empty_input, "cannot be empty"),
    (bad_format, "invalid format"),
])
def test_validation_errors(input_data, expected_error):
    with pytest.raises(ValidationError, match=expected_error):
        validate(input_data)
```

**JavaScript (jest)**:
```javascript
test.each([
    ['empty', emptyInput, 'cannot be empty'],
    ['invalid format', badFormat, 'invalid format'],
])('validation errors - %s', (name, input, expectedError) => {
    expect(() => validate(input))
        .toThrow(expectedError);
});
```

**Rust (built-in)**:
```rust
#[test]
fn test_validation_errors() {
    let result = validate(empty_input());
    assert!(result.is_err());
    assert!(result.unwrap_err().to_string().contains("cannot be empty"));
}
```

**Adaptation Effort**: 10-20% (error handling conventions differ)

---

### Pattern 5: Test Helper Pattern

**Go (testing package)**:
```go
func createTestFixture(t *testing.T) *Fixture {
    t.Helper()
    return &Fixture{
        Field1: "value1",
        Field2: "value2",
    }
}

func TestWithFixture(t *testing.T) {
    fixture := createTestFixture(t)
    result := UseFixture(fixture)
    // assertions
}
```

**Python (pytest fixtures)**:
```python
@pytest.fixture
def test_fixture():
    return Fixture(field1="value1", field2="value2")

def test_with_fixture(test_fixture):
    result = use_fixture(test_fixture)
    # assertions
```

**JavaScript (jest beforeEach or factory)**:
```javascript
function createTestFixture() {
    return {
        field1: 'value1',
        field2: 'value2'
    };
}

test('with fixture', () => {
    const fixture = createTestFixture();
    const result = useFixture(fixture);
    // assertions
});
```

**Rust (common module or helper)**:
```rust
fn create_test_fixture() -> Fixture {
    Fixture {
        field1: "value1".to_string(),
        field2: "value2".to_string(),
    }
}

#[test]
fn test_with_fixture() {
    let fixture = create_test_fixture();
    let result = use_fixture(fixture);
    // assertions
}
```

**Adaptation Effort**: 20-30% (fixture patterns vary significantly)

---

### Pattern 6: Dependency Injection Pattern

**Go (interfaces)**:
```go
type FileReader interface {
    Read(path string) ([]byte, error)
}

type MockReader struct {
    Data []byte
    Err  error
}

func (m *MockReader) Read(path string) ([]byte, error) {
    return m.Data, m.Err
}

func TestWithMock(t *testing.T) {
    mock := &MockReader{Data: []byte("test")}
    result := ProcessFile(mock, "path")
    // assertions
}
```

**Python (unittest.mock or dependency injection)**:
```python
from unittest.mock import Mock

def test_with_mock():
    mock_reader = Mock()
    mock_reader.read.return_value = b"test"

    result = process_file(mock_reader, "path")
    # assertions
    mock_reader.read.assert_called_once_with("path")
```

**JavaScript (jest.mock or manual mocking)**:
```javascript
const mockReader = {
    read: jest.fn().mockResolvedValue('test')
};

test('with mock', async () => {
    const result = await processFile(mockReader, 'path');
    expect(mockReader.read).toHaveBeenCalledWith('path');
    // assertions
});
```

**Rust (trait mocking with mockall or manual)**:
```rust
#[cfg(test)]
mod tests {
    use super::*;

    struct MockReader {
        data: Vec<u8>,
    }

    impl FileReader for MockReader {
        fn read(&self, path: &str) -> Result<Vec<u8>, Error> {
            Ok(self.data.clone())
        }
    }

    #[test]
    fn test_with_mock() {
        let mock = MockReader { data: b"test".to_vec() };
        let result = process_file(&mock, "path").unwrap();
        // assertions
    }
}
```

**Adaptation Effort**: 25-35% (mocking patterns differ significantly)

---

## Coverage Tools Adaptation

### Go → Python
```bash
# Go
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Python (pytest-cov)
pytest --cov=src --cov-report=term-missing
coverage report
```

**Tool Modifications**:
- Coverage analyzer: Parse `coverage report` output instead of `go tool cover`
- Threshold logic: Same
- Categorization: Function name patterns similar
- **Adaptation Effort**: 30-40% (output format parsing)

---

### Go → JavaScript
```bash
# Go
go test -coverprofile=coverage.out ./...

# JavaScript (jest)
npm test -- --coverage
npx nyc report --reporter=text
```

**Tool Modifications**:
- Coverage analyzer: Parse `nyc` or `jest` JSON output
- Threshold logic: Same
- Categorization: Function name patterns similar
- **Adaptation Effort**: 35-45% (JSON parsing, async awareness)

---

### Go → Rust
```bash
# Go
go test -coverprofile=coverage.out ./...

# Rust (tarpaulin or llvm-cov)
cargo tarpaulin --out Stdout
cargo llvm-cov --summary-only
```

**Tool Modifications**:
- Coverage analyzer: Parse tarpaulin output
- Threshold logic: Same
- Categorization: Function name patterns similar (Result<T, E> for error handling)
- **Adaptation Effort**: 20-30% (similar tooling philosophy)

---

## Test Generator Adaptation

### Template Modifications by Language

**Python (pytest)**:
- Replace `func Test...` with `def test_...`
- Replace `t *testing.T` with pytest fixtures
- Replace `t.Run` with `@pytest.mark.parametrize`
- Replace assertions with pytest assertions
- **Adaptation Effort**: 40-50% (syntax and framework differ)

**JavaScript (jest)**:
- Replace `func Test...` with `test('...', () => {})`
- Replace struct tables with `test.each([[...]])`
- Replace assertions with `expect(...).toBe(...)`
- Add async/await for Promise-based code
- **Adaptation Effort**: 45-55% (async patterns, syntax differ)

**Rust (built-in testing)**:
- Replace `func Test...` with `#[test] fn test_...`
- Replace struct tables with `#[rstest]` or manual loops
- Replace assertions with `assert_eq!`, `assert!`
- Handle Result<T, E> pattern with `.unwrap()` or `.expect()`
- **Adaptation Effort**: 30-40% (type system stricter, syntax differs)

---

## Workflow Automation Adaptation

### CI/CD Integration

**GitHub Actions (Go)**:
```yaml
- name: Run tests
  run: go test -v -coverprofile=coverage.out ./...
- name: Check coverage
  run: |
    total=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$total < 80" | bc -l) )); then exit 1; fi
```

**GitHub Actions (Python)**:
```yaml
- name: Run tests
  run: pytest --cov=src --cov-fail-under=80
```

**GitHub Actions (JavaScript)**:
```yaml
- name: Run tests
  run: npm test -- --coverage --coverageThreshold='{"global":{"lines":80}}'
```

**Adaptation Effort**: 15-25% (CI syntax similar, tool invocation differs)

---

## Summary: Adaptation Effort by Language

| Language | Workflow | Patterns | Tools | CI/CD | Overall |
|----------|----------|----------|-------|-------|---------|
| **Go → Rust** | 0% | 15% | 25% | 15% | **10-15%** |
| **Go → Java** | 0% | 20% | 30% | 20% | **15-25%** |
| **Go → Python** | 0% | 25% | 40% | 20% | **25-35%** |
| **Go → JavaScript** | 0% | 30% | 45% | 25% | **30-40%** |

**Key Insight**: Workflow (0%) and core principles (priority matrix, quality standards) transfer perfectly. Implementation details (patterns, tools) require more adaptation.

---

## V_reusability Score Calculation

**For V_reusability = 0.80 (15-40% modification threshold)**:

- **Same language (Go)**: 5% modification → V_reusability = 0.95
- **Similar language (Rust)**: 12% modification → V_reusability = 0.88
- **Different paradigm (Python)**: 30% modification → V_reusability = 0.80 ✅
- **Different paradigm + async (JavaScript)**: 35% modification → V_reusability = 0.75

**Conclusion**: Methodology achieves V_reusability ≥ 0.80 for Go, Rust, Java, Python. JavaScript at 0.75 (still good, but async patterns add complexity).

---

## Recommendations for Cross-Language Transfer

1. **Start with Workflow**: Implement the 8-step coverage-driven workflow first (0% adaptation)
2. **Adapt Priority Matrix**: Use same P1-P4 categorization (0% adaptation)
3. **Translate Patterns**: Start with Patterns 1-2 (unit, table-driven) - 15-25% effort
4. **Build Tools Incrementally**: Start with manual process, automate later - 30-45% effort
5. **Validate Quality Standards**: Same checklist applies (0% adaptation)

**Expected Timeline**:
- Same language: 2-4 hours (mostly tool setup)
- Similar language: 4-8 hours (pattern translation)
- Different paradigm: 8-12 hours (pattern + tool adaptation)

---

**Conclusion**: Methodology is highly transferable (0-40% adaptation depending on target language). Core principles (workflow, prioritization, quality standards) are universal. Implementation details (pattern syntax, tool parsing) require language-specific work but follow same structure.
