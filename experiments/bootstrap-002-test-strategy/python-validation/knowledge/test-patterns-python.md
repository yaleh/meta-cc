# Python Test Patterns - Adapted from Bootstrap-002

**Source**: Bootstrap-002 Test Strategy Development (Go)
**Target**: Python/pytest
**Adaptation**: 6 patterns adapted, 30% average modification

---

## Pattern 1: Unit Test Pattern (10% adaptation)

**Purpose**: Test a single function in isolation

**Python Structure**:
```python
def test_function_name_scenario():
    """Test description."""
    # Setup
    input_data = create_test_input()

    # Execute
    result = function_under_test(input_data)

    # Assert
    assert result == expected_value
    assert result.property == expected_property
```

**Adaptation from Go**:
- `func Test` → `def test_`
- `t.Fatalf` → `assert` with message
- `t.Errorf` → `assert`

---

## Pattern 2: Table-Driven Test Pattern (30% adaptation)

**Purpose**: Test multiple scenarios with same logic

**Python Structure**:
```python
@pytest.mark.parametrize("input,expected,should_error", [
    ("valid_input", "expected_output", False),
    ("invalid_input", None, True),
    ("edge_case", "special_output", False),
])
def test_function(input, expected, should_error):
    """Test with various inputs."""
    if should_error:
        with pytest.raises(ValueError):
            function_under_test(input)
    else:
        result = function_under_test(input)
        assert result == expected
```

**Adaptation from Go**:
- `[]struct{}` → `@pytest.mark.parametrize` decorator
- Test struct → tuple list
- `t.Run(tt.name)` → pytest handles naming automatically

---

## Pattern 3: Mock/Stub Pattern (45% adaptation)

**Purpose**: Isolate code from dependencies

**Python Structure**:
```python
from unittest.mock import patch, mock_open

def test_with_file_mock():
    """Test file I/O with mocking."""
    file_content = "test content"

    with patch('pathlib.Path.exists', return_value=True), \
         patch('builtins.open', mock_open(read_data=file_content)):
        result = function_that_reads_file('dummy.txt')
        assert result == expected_output
```

**Adaptation from Go**:
- Interface mocking → `unittest.mock.patch`
- `httptest` → `mock_open` for files
- Additional mocking may be needed (e.g., `Path.exists()`)

---

## Pattern 4: Error Path Test Pattern (25% adaptation)

**Purpose**: Test error handling

**Python Structure**:
```python
def test_function_raises_error():
    """Test error is raised for invalid input."""
    with pytest.raises(ValueError, match="error message pattern"):
        function_under_test(invalid_input)
```

**Adaptation from Go**:
- `if err != nil` checks → `with pytest.raises(ExceptionType)`
- Error message matching more explicit

---

## Pattern 5: Test Fixture Pattern (35% adaptation)

**Purpose**: Reusable test setup/data

**Python Structure**:
```python
@pytest.fixture
def sample_data():
    """Provide sample data for tests."""
    return [
        {"id": 1, "name": "Test 1"},
        {"id": 2, "name": "Test 2"},
    ]

@pytest.fixture
def analyzer(sample_data):
    """Provide configured analyzer."""
    return Analyzer(sample_data)

def test_with_fixture(analyzer):
    """Test using fixture."""
    result = analyzer.process()
    assert len(result) == 2
```

**Adaptation from Go**:
- Helper functions → `@pytest.fixture` decorator
- Explicit setup → pytest lifecycle
- Dependency injection via fixture parameters

---

## Pattern 8: Integration Test Pattern (35% adaptation)

**Purpose**: Test end-to-end workflows

**Python Structure**:
```python
def test_integration_workflow(tmp_path):
    """Test complete workflow with real files."""
    # Setup: Create temp file
    test_file = tmp_path / "test.json"
    test_file.write_text('{"data": "value"}')

    # Execute: Full workflow
    result = process_file(str(test_file))

    # Assert: End-to-end verification
    assert result.success
    assert result.data == "value"
    # pytest handles cleanup automatically
```

**Adaptation from Go**:
- `os.CreateTemp` → `tmp_path` pytest fixture
- `defer cleanup` → pytest handles cleanup
- Path handling uses `pathlib`

---

## Quick Reference

| Go Pattern | Python Equivalent | Adaptation % |
|------------|-------------------|--------------|
| `func TestX(t *testing.T)` | `def test_x():` | 10% |
| `[]struct{...}` | `@pytest.mark.parametrize` | 30% |
| Interface mocking | `unittest.mock.patch` | 45% |
| `if err != nil` | `pytest.raises` | 25% |
| Helper functions | `@pytest.fixture` | 35% |
| `os.CreateTemp` | `tmp_path` fixture | 35% |

**Average Adaptation**: 30%
