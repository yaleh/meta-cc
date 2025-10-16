# Pipeline Script Tests

Unit tests for CI/CD pipeline bash scripts using the Bats (Bash Automated Testing System) framework.

## Prerequisites

Install Bats:

```bash
# macOS
brew install bats-core

# Linux (Ubuntu/Debian)
sudo apt-get install bats

# Or via npm
npm install -g bats
```

## Running Tests

Run all script tests:

```bash
bats tests/scripts/*.bats
```

Run specific test file:

```bash
bats tests/scripts/test-track-metrics.bats
```

Run with verbose output:

```bash
bats -t tests/scripts/*.bats
```

## Test Files

- `test-track-metrics.bats` - Tests for metrics tracking (CSV storage)
- `test-check-performance-regression.bats` - Tests for regression detection
- `test-smoke-tests.bats` - Tests for smoke testing framework

## Writing New Tests

Follow Bats conventions:

```bash
#!/usr/bin/env bats

setup() {
    # Run before each test
    export TEST_DIR="$(mktemp -d)"
}

teardown() {
    # Run after each test
    rm -rf "$TEST_DIR"
}

@test "description of what is being tested" {
    run command_to_test

    [ "$status" -eq 0 ]  # Check exit code
    [[ "$output" =~ "expected pattern" ]]  # Check output
}
```

## Test Coverage

These tests validate:

1. **Input validation** - Correct handling of invalid arguments
2. **File operations** - CSV creation, appending, trimming
3. **Calculations** - Moving average, regression percentage
4. **Exit codes** - Proper status codes for different scenarios
5. **Output formatting** - Expected messages and confirmations

## CI Integration

Tests run automatically in GitHub Actions workflow:

```yaml
- name: Run pipeline script tests
  run: |
    sudo apt-get update && sudo apt-get install -y bats
    bats tests/scripts/*.bats
```

## Local Development

Run tests before committing changes to pipeline scripts:

```bash
# Quick check
make test-scripts  # (if Makefile target added)

# Or directly
bats tests/scripts/*.bats
```
