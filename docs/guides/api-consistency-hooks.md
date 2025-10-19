# Git Hooks for API Consistency

This guide explains how to use git hooks to automatically validate MCP API consistency before commits.

---

## Overview

The pre-commit hook for API consistency runs `meta-cc validate-api` before each git commit to ensure MCP tool definitions follow established conventions.

**Benefits**:
- Catches API convention violations early (before commit)
- Prevents inconsistent tool definitions from entering repository
- Enforces naming, parameter ordering, and description standards automatically
- Reduces manual code review burden

**Note**: This hook is separate from the plugin version management hook documented in [Git Hooks](git-hooks.md).

---

## Installation

### Automatic Installation (Recommended)

Run the installation script:

```bash
./scripts/install-consistency-hooks.sh
```

This will:
1. Check if git repository exists
2. Build `validate-api` binary if missing
3. Back up existing pre-commit hook (if any)
4. Copy hook to `.git/hooks/pre-commit`
5. Make hook executable
6. Test hook with sample validation

### Manual Installation

1. Ensure `validate-api` binary exists:
   ```bash
   go build -o validate-api ./cmd/validate-api
   ```

2. Copy hook template:
   ```bash
   cp scripts/pre-commit.sample .git/hooks/pre-commit-api-validation
   chmod +x .git/hooks/pre-commit-api-validation
   ```

3. Link or merge with existing pre-commit hook:
   ```bash
   # If no existing hook
   ln -s pre-commit-api-validation .git/hooks/pre-commit

   # If existing hook, add validation call
   echo "./path/to/pre-commit-api-validation" >> .git/hooks/pre-commit
   ```

---

## Hook Behavior

### What It Does

1. Detects if `cmd/mcp-server/tools.go` was modified in staged changes
2. Runs `meta-cc validate-api --fast` if tools.go changed
3. **If violations found**: Blocks commit, shows errors
4. **If no violations**: Allows commit to proceed
5. **If tools.go not changed**: Skips validation, allows commit

### Example (Passing Validation)

```bash
$ git commit -m "Add query_project_timeline tool"

===========================================
Pre-Commit Hook: API Consistency Validation
===========================================

Detected changes to cmd/mcp-server/tools.go
Running validation...

API Consistency Validation: PASSED

✓ All checks passed
✓ Commit allowed
```

### Example (Failing Validation)

```bash
$ git commit -m "Add retrieve_data tool"

===========================================
Pre-Commit Hook: API Consistency Validation
===========================================

Detected changes to cmd/mcp-server/tools.go
Running validation...

API Consistency Validation: FAILED

✗ retrieve_data: Naming pattern violation
  Tool name should use query_* prefix
  Suggestion: Rename to query_data
  Severity: ERROR

✗ retrieve_data: Description format violation
  Missing "Default scope:" in description
  Expected template: "<Action> <object>. Default scope: <X>."
  Severity: ERROR

❌ Commit blocked. Fix violations before committing.

To bypass (not recommended):
  git commit --no-verify
```

---

## Bypassing Hook (Not Recommended)

In exceptional cases, you can bypass the hook:

```bash
git commit --no-verify -m "WIP: Experimental API changes"
```

**Warning**: Only use `--no-verify` for:
- Work-in-progress commits (that won't be pushed)
- Emergency hotfixes (fix violations immediately after)
- Testing hook behavior (development only)

**Always fix violations** before pushing to remote repository.

---

## Disabling Hook

### Temporarily Disable

```bash
# Rename hook
mv .git/hooks/pre-commit .git/hooks/pre-commit.disabled

# Re-enable
mv .git/hooks/pre-commit.disabled .git/hooks/pre-commit
```

### Permanently Remove

```bash
rm .git/hooks/pre-commit
```

To reinstall later, run `./scripts/install-consistency-hooks.sh`.

---

## Troubleshooting

### Hook Not Running

**Issue**: Commit proceeds without validation even though tools.go changed

**Possible Causes**:
1. Hook not executable
2. Hook not installed
3. Hook script has errors

**Solutions**:
```bash
# Check if hook exists
ls -l .git/hooks/pre-commit

# Make executable
chmod +x .git/hooks/pre-commit

# Test hook manually
.git/hooks/pre-commit

# Reinstall
./scripts/install-consistency-hooks.sh
```

### Hook Failing Incorrectly

**Issue**: Hook reports violations but tools.go wasn't changed

**Cause**: Hook detecting unrelated file changes or stale staged files

**Solution**:
```bash
# Check what's staged
git diff --cached --name-only

# Unstage unintended files
git reset HEAD <file>

# Review hook logic
cat .git/hooks/pre-commit
```

### validate-api Binary Missing

**Issue**: Hook fails with "validate-api: command not found"

**Cause**: Binary not built or not in PATH

**Solution**:
```bash
# Build binary
go build -o validate-api ./cmd/validate-api

# Verify binary exists
ls -l validate-api

# Reinstall hook (auto-builds binary)
./scripts/install-consistency-hooks.sh
```

### Hook Slow

**Issue**: Hook takes too long to run (> 5 seconds)

**Cause**: Running full validation instead of fast mode, or large tools.go file

**Solution**:
1. Ensure hook uses `--fast` flag:
   ```bash
   grep "\-\-fast" .git/hooks/pre-commit
   ```

2. If not using `--fast`, edit hook:
   ```bash
   vim .git/hooks/pre-commit
   # Change: ./validate-api
   # To:     ./validate-api --fast
   ```

3. For very large files, consider validation as separate step

### False Positives

**Issue**: Hook reports violations that don't seem like real problems

**Cause**: Convention rules may be strict or parser limitations

**Solution**:
1. Review API conventions:
   ```bash
   cat experiments/bootstrap-006-api-design/data/api-parameter-convention.md
   ```

2. Check if tool definition follows conventions:
   - Naming: `query_*`, `get_*`, `list_*`, `cleanup_*`
   - Description: Includes "Default scope: X"
   - Parameter order: Tier 1 (required) → Tier 2 (filtering) → Tier 3 (range) → Tier 4 (output)

3. If convention is unclear or outdated, discuss with team before bypassing

---

## Advanced Configuration

### Custom Validation File

Edit `.git/hooks/pre-commit` to validate different file:

```bash
# Change this line:
if git diff --cached --name-only | grep -q "cmd/mcp-server/tools.go"; then

# To:
if git diff --cached --name-only | grep -q "path/to/other/api-file.go"; then
```

### Additional Checks

Add more validation steps to the hook:

```bash
#!/bin/bash

# API consistency validation
if git diff --cached --name-only | grep -q "cmd/mcp-server/tools.go"; then
    ./validate-api --fast || exit 1
fi

# Additional checks
go fmt ./cmd/mcp-server/... || exit 1
go vet ./cmd/mcp-server/... || exit 1
```

### Quiet Mode

Reduce hook output:

```bash
# Edit hook to use --quiet flag
./validate-api --fast --quiet || exit 1
```

### JSON Output for Logging

Log validation results:

```bash
# Edit hook
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
./validate-api --fast --json > ".git/hooks/validation-${TIMESTAMP}.json" || exit 1
```

---

## Integration with CI/CD

The validation hook can be integrated into CI/CD pipelines:

### GitHub Actions Example

```yaml
name: API Consistency Check

on: [pull_request]

jobs:
  validate-api:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build validate-api
        run: go build -o validate-api ./cmd/validate-api

      - name: Validate MCP API
        run: ./validate-api --file cmd/mcp-server/tools.go --json
```

### GitLab CI Example

```yaml
validate-api:
  stage: test
  script:
    - go build -o validate-api ./cmd/validate-api
    - ./validate-api --file cmd/mcp-server/tools.go --json
  only:
    changes:
      - cmd/mcp-server/tools.go
```

---

## Comparison: Hook vs Manual Validation

| Aspect | Git Hook | Manual Validation |
|--------|----------|-------------------|
| **Trigger** | Automatic (on commit) | Manual invocation |
| **Speed** | Fast (~1 second) | Fast (~1 second) |
| **Coverage** | 100% (can't commit without) | Varies (easy to forget) |
| **Bypass** | Use `--no-verify` | N/A |
| **Best For** | Enforcement | Development, debugging |

---

## Best Practices

### For Developers

1. **Run validation manually** before staging:
   ```bash
   ./validate-api --fast
   ```

2. **Fix violations immediately** when hook blocks commit

3. **Don't bypass** unless absolutely necessary (WIP commits, emergencies)

4. **Review validation output** to understand violations

5. **Learn conventions** to avoid violations in future

### For Team Leads

1. **Require hook installation** for all team members

2. **Document conventions clearly** (link from hook output)

3. **Update validation rules** as API evolves

4. **Review bypassed commits** in code review

5. **Integrate into CI** as backup enforcement

---

## Hook Implementation Details

### File Locations

- **Hook template**: `scripts/pre-commit.sample` (tracked in git)
- **Active hook**: `.git/hooks/pre-commit` (not tracked, installed locally)
- **Installation script**: `scripts/install-consistency-hooks.sh`

### Detection Logic

```bash
# Check if tools.go was modified in staged changes
if git diff --cached --name-only | grep -q "cmd/mcp-server/tools.go"; then
    TOOLS_CHANGED=true
fi
```

### Validation Execution

```bash
# Run validate-api with fast mode
if ! ./validate-api --fast; then
    echo "❌ Commit blocked. Fix violations before committing."
    exit 1
fi
```

### Exit Codes

- `0`: Validation passed, allow commit
- `1`: Validation failed, block commit

---

## See Also

- [validate-api CLI Reference](../reference/cli.md#validate-api) - Complete command documentation
- [Git Hooks (Plugin Version Management)](git-hooks.md) - Plugin version bump hooks
- [API Parameter Convention](../../experiments/bootstrap-006-api-design/data/api-parameter-convention.md) - Convention specification
- [MCP Guide](mcp.md) - MCP tool reference
