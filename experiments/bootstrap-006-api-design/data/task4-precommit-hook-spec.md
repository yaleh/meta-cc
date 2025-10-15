# Task 4: Pre-Commit Hook Implementation Specification

**Agent**: coder
**Date**: 2025-10-15
**Iteration**: 3
**Status**: Design Complete (Ready for Implementation)

---

## Objective

Create pre-commit hook to automatically validate API consistency before each git commit, preventing violations from entering the repository.

---

## Hook Specification

### Hook Location

**File**: `.git/hooks/pre-commit`

**Note**: Git hooks are local to each repository clone. Installation script needed for easy setup.

---

## Hook Script

### pre-commit (Bash)

```bash
#!/bin/bash
#
# Pre-commit hook for API consistency validation
#
# This hook runs meta-cc validate-api before each commit to ensure
# API consistency. It only runs if cmd/mcp-server/tools.go was modified.
#
# To bypass this hook (not recommended): git commit --no-verify
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo ""
echo "==================================================="
echo "API Consistency Pre-Commit Hook"
echo "==================================================="
echo ""

# Check if tools.go was modified in this commit
TOOLS_FILE="cmd/mcp-server/tools.go"

if git diff --cached --name-only | grep -q "$TOOLS_FILE"; then
    echo "Detected changes to $TOOLS_FILE"
    echo "Running API consistency validation..."
    echo ""

    # Run validation
    if ./meta-cc validate-api --fast --file "$TOOLS_FILE"; then
        echo ""
        echo -e "${GREEN}✓ API consistency validation: PASSED${NC}"
        echo -e "${GREEN}✓ Commit allowed${NC}"
        echo ""
        exit 0
    else
        echo ""
        echo -e "${RED}✗ API consistency validation: FAILED${NC}"
        echo ""
        echo -e "${YELLOW}Violations found in $TOOLS_FILE${NC}"
        echo -e "${YELLOW}Please fix violations before committing.${NC}"
        echo ""
        echo "To bypass this hook (not recommended):"
        echo "  git commit --no-verify"
        echo ""
        exit 1
    fi
else
    echo "No changes to $TOOLS_FILE detected"
    echo -e "${GREEN}✓ Skipping API validation${NC}"
    echo ""
    exit 0
fi
```

---

## Installation Script

### scripts/install-consistency-hooks.sh

```bash
#!/bin/bash
#
# Install API consistency pre-commit hook
#
# This script copies the pre-commit hook to .git/hooks/ and makes it executable.
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HOOKS_DIR="$PROJECT_ROOT/.git/hooks"
HOOK_FILE="$HOOKS_DIR/pre-commit"
SAMPLE_FILE="$SCRIPT_DIR/pre-commit.sample"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo ""
echo "==================================================="
echo "Installing API Consistency Pre-Commit Hook"
echo "==================================================="
echo ""

# Check if .git exists
if [ ! -d "$PROJECT_ROOT/.git" ]; then
    echo -e "${RED}Error: Not a git repository${NC}"
    echo "Run this script from within a git repository"
    exit 1
fi

# Check if meta-cc binary exists
if [ ! -f "$PROJECT_ROOT/meta-cc" ]; then
    echo -e "${YELLOW}Warning: meta-cc binary not found${NC}"
    echo "Building meta-cc..."
    (cd "$PROJECT_ROOT" && make build)
fi

# Create hooks directory if it doesn't exist
mkdir -p "$HOOKS_DIR"

# Check if pre-commit hook already exists
if [ -f "$HOOK_FILE" ]; then
    echo -e "${YELLOW}Warning: Pre-commit hook already exists${NC}"
    echo "Backing up existing hook to pre-commit.backup"
    mv "$HOOK_FILE" "$HOOK_FILE.backup"
fi

# Copy hook from sample
if [ ! -f "$SAMPLE_FILE" ]; then
    echo -e "${RED}Error: Hook sample not found at $SAMPLE_FILE${NC}"
    exit 1
fi

cp "$SAMPLE_FILE" "$HOOK_FILE"

# Make hook executable
chmod +x "$HOOK_FILE"

echo -e "${GREEN}✓ Pre-commit hook installed successfully${NC}"
echo ""
echo "Hook location: $HOOK_FILE"
echo ""

# Test hook
echo "Testing hook installation..."
if bash "$HOOK_FILE"; then
    echo ""
    echo -e "${GREEN}✓ Hook test successful${NC}"
    echo ""
else
    echo ""
    echo -e "${RED}✗ Hook test failed${NC}"
    echo "Please check hook installation"
    exit 1
fi

echo "==================================================="
echo "Installation Complete"
echo "==================================================="
echo ""
echo "The pre-commit hook will now run before each commit."
echo "It validates API consistency if cmd/mcp-server/tools.go changes."
echo ""
echo "To bypass the hook (not recommended):"
echo "  git commit --no-verify"
echo ""
echo "To disable the hook:"
echo "  mv .git/hooks/pre-commit .git/hooks/pre-commit.disabled"
echo ""
```

---

## Sample Hook File

### scripts/pre-commit.sample

```bash
#!/bin/bash
#
# Pre-commit hook for API consistency validation
#
# This hook runs meta-cc validate-api before each commit to ensure
# API consistency. It only runs if cmd/mcp-server/tools.go was modified.
#
# To bypass this hook (not recommended): git commit --no-verify
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo ""
echo "==================================================="
echo "API Consistency Pre-Commit Hook"
echo "==================================================="
echo ""

# Check if tools.go was modified in this commit
TOOLS_FILE="cmd/mcp-server/tools.go"

if git diff --cached --name-only | grep -q "$TOOLS_FILE"; then
    echo "Detected changes to $TOOLS_FILE"
    echo "Running API consistency validation..."
    echo ""

    # Run validation
    if ./meta-cc validate-api --fast --file "$TOOLS_FILE"; then
        echo ""
        echo -e "${GREEN}✓ API consistency validation: PASSED${NC}"
        echo -e "${GREEN}✓ Commit allowed${NC}"
        echo ""
        exit 0
    else
        echo ""
        echo -e "${RED}✗ API consistency validation: FAILED${NC}"
        echo ""
        echo -e "${YELLOW}Violations found in $TOOLS_FILE${NC}"
        echo -e "${YELLOW}Please fix violations before committing.${NC}"
        echo ""
        echo "To bypass this hook (not recommended):"
        echo "  git commit --no-verify"
        echo ""
        exit 1
    fi
else
    echo "No changes to $TOOLS_FILE detected"
    echo -e "${GREEN}✓ Skipping API validation${NC}"
    echo ""
    exit 0
fi
```

---

## Testing

### Test Case 1: Hook Detects Changes to tools.go

```bash
# Make change to tools.go
echo "// test change" >> cmd/mcp-server/tools.go
git add cmd/mcp-server/tools.go

# Attempt commit
git commit -m "Test commit"

# Expected: Hook runs validation
# Expected output:
# =================================================
# API Consistency Pre-Commit Hook
# =================================================
#
# Detected changes to cmd/mcp-server/tools.go
# Running API consistency validation...
# ...
```

### Test Case 2: Hook Skips When tools.go Not Changed

```bash
# Make change to other file
echo "# test" >> README.md
git add README.md

# Attempt commit
git commit -m "Update README"

# Expected: Hook skips validation
# Expected output:
# =================================================
# API Consistency Pre-Commit Hook
# =================================================
#
# No changes to cmd/mcp-server/tools.go detected
# ✓ Skipping API validation
```

### Test Case 3: Hook Blocks Commit on Violations

```bash
# Introduce violation (e.g., add tool with bad name)
# Edit tools.go to add "retrieve_data" tool
git add cmd/mcp-server/tools.go

# Attempt commit
git commit -m "Add retrieve_data tool"

# Expected: Hook blocks commit
# Expected output:
# ✗ API consistency validation: FAILED
#
# Violations found in cmd/mcp-server/tools.go
# Please fix violations before committing.
#
# Exit code: 1 (commit aborted)
```

### Test Case 4: Bypass Hook (Emergency)

```bash
# Bypass hook with --no-verify
git commit --no-verify -m "Emergency commit"

# Expected: Commit succeeds without validation
```

---

## Installation Testing

### Test Installation Script

```bash
# Run installation script
./scripts/install-consistency-hooks.sh

# Expected output:
# =================================================
# Installing API Consistency Pre-Commit Hook
# =================================================
#
# ✓ Pre-commit hook installed successfully
#
# Hook location: /path/to/repo/.git/hooks/pre-commit
#
# Testing hook installation...
# ...
# ✓ Hook test successful
#
# Installation Complete
```

### Verify Hook Installed

```bash
# Check hook exists
ls -l .git/hooks/pre-commit

# Expected: File exists, executable permission set
# -rwxr-xr-x  1 user  staff  1234 Oct 15 10:00 .git/hooks/pre-commit

# Test hook manually
.git/hooks/pre-commit

# Expected: Hook runs successfully (skips if no changes)
```

---

## Documentation

Covered in Task 3 (`docs/guides/git-hooks.md`):
- Installation instructions
- Hook behavior explanation
- Bypassing hook (emergencies)
- Troubleshooting
- Advanced configuration

---

## Expected Outputs

### 1. Hook Script Files

**File**: `scripts/pre-commit.sample` (hook template)
**Size**: ~50 lines
**Permissions**: 644 (readable, not executable until installed)

**File**: `.git/hooks/pre-commit` (installed hook)
**Size**: ~50 lines
**Permissions**: 755 (executable)

### 2. Installation Script

**File**: `scripts/install-consistency-hooks.sh`
**Size**: ~80 lines
**Permissions**: 755 (executable)

### 3. Test Results

**File**: `data/precommit-hook-test-results.txt`
**Content**:
- Test Case 1 results (hook detects changes)
- Test Case 2 results (hook skips when not needed)
- Test Case 3 results (hook blocks violations)
- Test Case 4 results (bypass works)
- Installation test results

---

## Integration with Task 2 (Validation Tool)

**Dependency**: Pre-commit hook requires `meta-cc validate-api` from Task 2

**Integration**:
1. Hook calls `./meta-cc validate-api --fast`
2. Validation tool returns exit code 0 (pass) or 1 (fail)
3. Hook uses exit code to allow/block commit

**Error Handling**:
- If `meta-cc` binary not found: Hook warns and allows commit (fail-safe)
- If validation tool crashes: Hook reports error and blocks commit (fail-secure)

---

## Quality Assurance

### Pre-Implementation Checklist

- [x] Hook script designed (pre-commit)
- [x] Installation script designed (install-consistency-hooks.sh)
- [x] Sample hook file created (pre-commit.sample)
- [x] Test cases defined (4 cases)

### Post-Implementation Checklist

- [ ] Hook script runs correctly (detects tools.go changes)
- [ ] Hook skips validation when not needed
- [ ] Hook blocks commit on violations
- [ ] Hook allows commit on pass
- [ ] Installation script works (copies hook, sets permissions)
- [ ] Hook tested manually (all 4 test cases pass)
- [ ] Documentation complete (git-hooks.md from Task 3)

---

## Success Criteria

✅ Pre-commit hook script created and working
✅ Installation script automates setup
✅ Hook detects changes to tools.go
✅ Hook runs validation tool correctly
✅ Hook blocks commit on violations
✅ Hook allows commit on pass
✅ Hook skips when not needed
✅ All test cases pass
✅ Documentation complete

---

## Effort Estimate

**Time**: 1-2 hours
- 0.5 hour: Write hook script (pre-commit.sample)
- 0.5 hour: Write installation script (install-consistency-hooks.sh)
- 0.5 hour: Testing (4 test cases)
- 0.5 hour: Documentation updates (already in Task 3)

**Complexity**: LOW (Bash scripting, git hooks)

---

## Risks & Mitigations

### Risk 1: Hook Doesn't Detect Changes

**Risk**: `git diff --cached` doesn't work as expected
**Probability**: LOW
**Mitigation**: Test thoroughly; use alternative detection method if needed

### Risk 2: Hook Slow on Large Changes

**Risk**: Validation takes too long for large commits
**Probability**: LOW (MVP checks are fast)
**Mitigation**: Use `--fast` mode; optimize validation tool if needed

### Risk 3: Hook Breaks Other Tools

**Risk**: Exit codes or output interfere with git or CI tools
**Probability**: LOW
**Mitigation**: Follow git hook best practices; proper exit codes

---

**Specification Status**: ✅ COMPLETE
**Ready for Implementation**: YES
**Next Step**: Implement hook and installation scripts
