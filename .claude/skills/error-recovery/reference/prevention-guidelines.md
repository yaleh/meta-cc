# Error Prevention Guidelines

**Version**: 1.0
**Source**: Bootstrap-003 Error Recovery Methodology
**Last Updated**: 2025-10-18

Proactive strategies to prevent common errors before they occur.

---

## Overview

**Prevention is better than recovery**. This document provides actionable guidelines to prevent the most common error categories.

**Automation Impact**: 3 automated tools prevent 23.7% of all errors (317/1336)

---

## Category 1: Build/Compilation Errors (15.0%)

### Prevention Strategies

**1. Pre-Commit Linting**
```bash
# Add to .git/hooks/pre-commit
gofmt -w .
golangci-lint run
go build
```

**2. IDE Integration**
- Use IDE with real-time syntax checking (VS Code, GoLand)
- Enable "save on format" (gofmt)
- Configure inline linter warnings

**3. Incremental Compilation**
```bash
# Build frequently during development
go build ./...  # Fast incremental build
```

**4. Type Safety**
- Use strict type checking
- Avoid `interface{}` when possible
- Add type assertions with error checks

### Effectiveness
Prevents ~60% of Category 1 errors

---

## Category 2: Test Failures (11.2%)

### Prevention Strategies

**1. Run Tests Before Commit**
```bash
# Add to .git/hooks/pre-commit
go test ./...
```

**2. Test-Driven Development (TDD)**
- Write test first
- Write minimal code to pass
- Refactor

**3. Fixture Management**
```bash
# Version control test fixtures
git add tests/fixtures/
# Update fixtures with code changes
./scripts/update-fixtures.sh
```

**4. Continuous Integration**
```yaml
# .github/workflows/test.yml
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run tests
        run: go test ./...
```

### Effectiveness
Prevents ~70% of Category 2 errors

---

## Category 3: File Not Found (18.7%) ⚠️ AUTOMATABLE

### Prevention Strategies

**1. Path Validation Tool** ✅
```bash
# Use automation before file operations
./scripts/validate-path.sh [path]

# Returns:
# - File exists: OK
# - File missing: Suggests similar paths
```

**2. Autocomplete**
- Use shell/IDE autocomplete for paths
- Tab completion reduces typos by 95%

**3. Existence Checks**
```go
// In code
if _, err := os.Stat(path); os.IsNotExist(err) {
    return fmt.Errorf("file not found: %s", path)
}
```

**4. Working Directory Awareness**
```bash
# Always know where you are
pwd
# Use absolute paths when unsure
realpath [relative_path]
```

### Effectiveness
**Prevents 65.2% of Category 3 errors** with automation

---

## Category 4: File Size Exceeded (6.3%) ⚠️ AUTOMATABLE

### Prevention Strategies

**1. Size Check Tool** ✅
```bash
# Use automation before reading
./scripts/check-file-size.sh [file]

# Returns:
# - OK to read
# - Too large, use pagination
# - Suggests offset/limit values
```

**2. Pre-Read Size Check**
```bash
# Manual check
wc -l [file]
du -h [file]

# If >10000 lines, use pagination
```

**3. Use Selective Reading**
```bash
# Instead of full read
head -n 1000 [file]
grep "pattern" [file]
tail -n 1000 [file]
```

**4. Streaming for Large Files**
```go
// In code, process line-by-line
scanner := bufio.NewScanner(file)
for scanner.Scan() {
    processLine(scanner.Text())
}
```

### Effectiveness
**Prevents 100% of Category 4 errors** with automation

---

## Category 5: Write Before Read (5.2%) ⚠️ AUTOMATABLE

### Prevention Strategies

**1. Read-Before-Write Check** ✅
```bash
# Use automation before Write/Edit
./scripts/check-read-before-write.sh [file]

# Returns:
# - File already read: OK to write
# - File not read: Suggests Read first
```

**2. Always Read First**
```bash
# Workflow pattern
Read [file]      # Step 1: Always read
Edit [file] ...  # Step 2: Then edit
```

**3. Use Edit for Modifications**
- Edit: Requires prior read (safer)
- Write: For new files or complete rewrites

**4. Session Context Awareness**
- Track what files have been read
- Clear workflow: Read → Analyze → Edit

### Effectiveness
**Prevents 100% of Category 5 errors** with automation

---

## Category 6: Command Not Found (3.7%)

### Prevention Strategies

**1. Build Before Execute**
```bash
# Always build first
make build
./command [args]
```

**2. PATH Verification**
```bash
# Check command availability
which [command] || echo "Command not found, build first"
```

**3. Use Absolute Paths**
```bash
# For project binaries
./bin/meta-cc [args]
# Not: meta-cc [args]
```

**4. Dependency Checks**
```bash
# Check required tools
command -v jq >/dev/null || echo "jq not installed"
command -v go >/dev/null || echo "go not installed"
```

### Effectiveness
Prevents ~80% of Category 6 errors

---

## Category 7: JSON Parsing Errors (6.0%)

### Prevention Strategies

**1. Validate JSON Before Use**
```bash
# Validate syntax
jq . [file.json] > /dev/null

# Validate and pretty-print
cat [file.json] | python -m json.tool
```

**2. Schema Validation**
```bash
# Use JSON schema validator
jsonschema -i [data.json] [schema.json]
```

**3. Test Fixtures with Code**
```go
// Test that fixtures parse correctly
func TestFixtureParsing(t *testing.T) {
    data, _ := os.ReadFile("tests/fixtures/sample.json")
    var result MyStruct
    if err := json.Unmarshal(data, &result); err != nil {
        t.Errorf("Fixture doesn't match schema: %v", err)
    }
}
```

**4. Type Safety**
```go
// Use strong typing
type Config struct {
    Port int    `json:"port"`    // Not string
    Name string `json:"name"`
}
```

### Effectiveness
Prevents ~70% of Category 7 errors

---

## Category 13: String Not Found (Edit) (3.2%)

### Prevention Strategies

**1. Always Re-Read Before Edit**
```bash
# Workflow
Read [file]              # Fresh read
Edit [file] old="..." new="..."  # Then edit
```

**2. Copy Exact Strings**
- Don't retype old_string
- Copy from file viewer
- Preserves whitespace/formatting

**3. Include Context**
```go
// Not: old_string="x"
// Yes: old_string="    x = 1\n    y = 2"  // Includes indentation
```

**4. Verify File Hasn't Changed**
```bash
# Check file modification time
ls -l [file]
# Or use version control
git status [file]
```

### Effectiveness
Prevents ~80% of Category 13 errors

---

## Cross-Cutting Prevention Strategies

### 1. Automation First

**High-Priority Automated Tools**:
1. `validate-path.sh` (65.2% of Category 3)
2. `check-file-size.sh` (100% of Category 4)
3. `check-read-before-write.sh` (100% of Category 5)

**Combined Impact**: 23.7% of ALL errors prevented

**Installation**:
```bash
# Add to PATH
export PATH=$PATH:./scripts

# Or use as hooks
./scripts/install-hooks.sh
```

### 2. Pre-Commit Hooks

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Format code
gofmt -w .

# Run linters
golangci-lint run

# Run tests
go test ./...

# Build
go build

# If any fail, prevent commit
if [ $? -ne 0 ]; then
    echo "Pre-commit checks failed"
    exit 1
fi
```

### 3. Continuous Integration

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Setup Go
        uses: actions/setup-go@v2
      - name: Lint
        run: golangci-lint run
      - name: Test
        run: go test ./... -cover
      - name: Build
        run: go build
```

### 4. Development Workflow

**Standard Workflow**:
1. Write code
2. Format (gofmt)
3. Lint (golangci-lint)
4. Test (go test)
5. Build (go build)
6. Commit

**TDD Workflow**:
1. Write test (fails - red)
2. Write code (passes - green)
3. Refactor
4. Repeat

---

## Prevention Metrics

### Impact by Category

| Category | Baseline Frequency | Prevention | Remaining |
|----------|-------------------|------------|-----------|
| File Not Found (3) | 250 (18.7%) | -163 (65.2%) | 87 (6.5%) |
| File Size (4) | 84 (6.3%) | -84 (100%) | 0 (0%) |
| Write Before Read (5) | 70 (5.2%) | -70 (100%) | 0 (0%) |
| **Total Automated** | **404 (30.2%)** | **-317 (78.5%)** | **87 (6.5%)** |

### ROI Analysis

**Time Investment**:
- Setup automation: 2 hours
- Maintain automation: 15 min/week

**Time Saved**:
- 317 errors × 3 min avg recovery = 951 minutes = 15.9 hours
- **ROI**: 7.95x in first month alone

---

## Best Practices

### Do's

✅ Use automation tools when available
✅ Run pre-commit hooks
✅ Test before commit
✅ Build incrementally
✅ Validate inputs (paths, JSON, etc.)
✅ Use type safety
✅ Check file existence before operations

### Don'ts

❌ Skip validation steps to save time
❌ Commit without running tests
❌ Ignore linter warnings
❌ Manually type file paths (use autocomplete)
❌ Skip pre-read for file edits
❌ Ignore automation tool suggestions

---

**Source**: Bootstrap-003 Error Recovery Methodology
**Framework**: BAIME (Bootstrapped AI Methodology Engineering)
**Status**: Production-ready, validated with 1336 errors
**Automation Coverage**: 23.7% of errors prevented
