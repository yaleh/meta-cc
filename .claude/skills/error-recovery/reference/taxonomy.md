# Error Classification Taxonomy

**Version**: 2.0
**Source**: Bootstrap-003 Error Recovery Methodology
**Last Updated**: 2025-10-18
**Coverage**: 95.4% of observed errors
**Categories**: 13 complete categories

This taxonomy classifies errors systematically for effective recovery and prevention.

---

## Overview

This taxonomy is:
- **MECE** (Mutually Exclusive, Collectively Exhaustive): 95.4% coverage
- **Actionable**: Each category has clear recovery paths
- **Observable**: Each category has detectable symptoms
- **Universal**: 85-90% applicable to other software projects

**Automation Coverage**: 23.7% of errors preventable with 3 automated tools

---

## 13 Error Categories

### Category 1: Build/Compilation Errors (15.0%)

**Definition**: Syntax errors, type mismatches, import issues preventing compilation

**Examples**:
- `cmd/root.go:4:2: "fmt" imported and not used`
- `undefined: someFunction`
- `cannot use x (type int) as type string`

**Common Causes**:
- Unused imports after refactoring
- Type mismatches from incomplete changes
- Missing function definitions
- Syntax errors

**Detection Pattern**: `*.go:[line]:[col]: [error message]`

**Prevention**:
- Pre-commit linting (gofmt, golangci-lint)
- IDE real-time syntax checking
- Incremental compilation

**Recovery**: Fix syntax/type issue, retry `go build`

**Automation Potential**: Medium

---

### Category 2: Test Failures (11.2%)

**Definition**: Unit or integration test assertions that fail during execution

**Examples**:
- `--- FAIL: TestLoadFixture (0.00s)`
- `Fixture content should contain 'sequence' field`
- `FAIL	github.com/project/package	0.003s`

**Common Causes**:
- Test fixture data mismatch
- Assertion failures from code changes
- Missing test data files
- Incorrect expected values

**Detection Pattern**: `--- FAIL:`, `FAIL\t`, assertion messages

**Prevention**:
- Run tests before commit
- Update test fixtures with code changes
- Test-driven development (TDD)

**Recovery**: Update test expectations or fix code

**Automation Potential**: Low (requires understanding test intent)

---

### Category 3: File Not Found (18.7%) ⚠️ AUTOMATABLE

**Definition**: Attempts to access non-existent files or directories

**Examples**:
- `File does not exist.`
- `wc: /path/to/file: No such file or directory`
- `File does not exist. Did you mean file.md?`

**Common Causes**:
- Typos in file paths
- Files moved or deleted
- Incorrect working directory
- Case sensitivity issues

**Detection Pattern**: `File does not exist`, `No such file or directory`

**Prevention**:
- **Automation: `validate-path.sh`** ✅ (prevents 65.2% of category 3 errors)
- Validate paths before file operations
- Use autocomplete for paths
- Check file existence first

**Recovery**: Correct file path, create missing file, or change directory

**Automation Potential**: **HIGH** ✅

---

### Category 4: File Size Exceeded (6.3%) ⚠️ AUTOMATABLE

**Definition**: Attempted to read files exceeding token limit

**Examples**:
- `File content (46892 tokens) exceeds maximum allowed tokens (25000)`
- `File too large to read in single operation`

**Common Causes**:
- Reading large generated files without pagination
- Reading entire JSON files
- Reading log files without limiting lines

**Detection Pattern**: `exceeds maximum allowed tokens`, `File too large`

**Prevention**:
- **Automation: `check-file-size.sh`** ✅ (prevents 100% of category 4 errors)
- Pre-check file size before reading
- Use offset/limit parameters
- Use grep/head/tail instead of full Read

**Recovery**: Use Read with offset/limit, or use grep

**Automation Potential**: **FULL** ✅

---

### Category 5: Write Before Read (5.2%) ⚠️ AUTOMATABLE

**Definition**: Attempted to Write/Edit a file without reading it first

**Examples**:
- `File has not been read yet. Read it first before writing to it.`

**Common Causes**:
- Forgetting to read file before edit
- Reading wrong file, editing intended file
- Session context lost
- Workflow error

**Detection Pattern**: `File has not been read yet`

**Prevention**:
- **Automation: `check-read-before-write.sh`** ✅ (prevents 100% of category 5 errors)
- Always Read before Write/Edit
- Use Edit instead of Write for existing files
- Check read history

**Recovery**: Read the file, then retry Write/Edit

**Automation Potential**: **FULL** ✅

---

### Category 6: Command Not Found (3.7%)

**Definition**: Bash commands that don't exist or aren't in PATH

**Examples**:
- `/bin/bash: line 1: meta-cc: command not found`
- `command not found: gofmt`

**Common Causes**:
- Binary not built yet
- Binary not in PATH
- Typo in command name
- Required tool not installed

**Detection Pattern**: `command not found`

**Prevention**:
- Build before running commands
- Verify tool installation
- Use absolute paths for project binaries

**Recovery**: Build binary, install tool, or correct command

**Automation Potential**: Medium

---

### Category 7: JSON Parsing Errors (6.0%)

**Definition**: Malformed JSON or schema mismatches

**Examples**:
- `json: cannot unmarshal string into Go struct field`
- `invalid character '}' looking for beginning of value`

**Common Causes**:
- Schema changes without updating code
- Malformed JSON in test fixtures
- Type mismatches
- Missing or extra commas/braces

**Detection Pattern**: `json:`, `unmarshal`, `invalid character`

**Prevention**:
- Validate JSON with jq before use
- Use JSON schema validation
- Test JSON fixtures with actual code

**Recovery**: Fix JSON structure or update schema

**Automation Potential**: Medium

---

### Category 8: Request Interruption (2.2%)

**Definition**: User manually interrupted tool execution

**Examples**:
- `[Request interrupted by user for tool use]`
- `Command aborted before execution`

**Common Causes**:
- User realized mistake mid-execution
- User wants to change approach
- Long-running command needs stopping

**Detection Pattern**: `interrupted by user`, `aborted before execution`

**Prevention**: Not applicable (user decision)

**Recovery**: Not needed (intentional)

**Automation Potential**: N/A

---

### Category 9: MCP Server Errors (17.1%)

**Definition**: Errors from Model Context Protocol tool integrations

**Subcategories**:
- 9a. Connection Errors (server unavailable)
- 9b. Timeout Errors (query exceeds time limit)
- 9c. Query Errors (invalid parameters)
- 9d. Data Errors (unexpected format)

**Examples**:
- `MCP server connection failed`
- `Query timeout after 30s`
- `Invalid parameter: status must be 'error' or 'success'`

**Common Causes**:
- MCP server not running
- Network issues
- Query too broad
- Invalid parameters
- Schema changes

**Prevention**:
- Check MCP server status before queries
- Use pagination for large queries
- Validate query parameters
- Handle connection errors gracefully

**Recovery**: Restart MCP server, optimize query, or fix parameters

**Automation Potential**: Medium

---

### Category 10: Permission Denied (0.7%)

**Definition**: Insufficient permissions to access file or execute command

**Examples**:
- `Permission denied: /path/to/file`
- `Operation not permitted`

**Common Causes**:
- File permissions too restrictive
- Directory not writable
- User doesn't own file

**Detection Pattern**: `Permission denied`, `Operation not permitted`

**Prevention**:
- Verify permissions before operations
- Use appropriate user context
- Avoid modifying system files

**Recovery**: Change permissions (chmod/chown)

**Automation Potential**: Low

---

### Category 11: Empty Command String (1.1%)

**Definition**: Bash tool invoked with empty or whitespace-only command

**Examples**:
- `/bin/bash: line 1: : command not found`

**Common Causes**:
- Variable expansion to empty string
- Conditional command construction error
- Copy-paste error

**Detection Pattern**: `/bin/bash: line 1: : command not found`

**Prevention**:
- Validate command strings are non-empty
- Check variable values
- Use bash -x to debug

**Recovery**: Provide valid command string

**Automation Potential**: High

---

### Category 12: Go Module Already Exists (0.4%)

**Definition**: Attempted `go mod init` when go.mod already exists

**Examples**:
- `go: /path/to/go.mod already exists`

**Common Causes**:
- Forgot to check for existing go.mod
- Re-running initialization script

**Detection Pattern**: `go.mod already exists`

**Prevention**:
- Check for go.mod existence before init
- Idempotent scripts

**Recovery**: No action needed

**Automation Potential**: Full

---

### Category 13: String Not Found (Edit Errors) (3.2%)

**Definition**: Edit tool attempts to replace non-existent string

**Examples**:
- `String to replace not found in file.`
- `String: {old content} not found`

**Common Causes**:
- File changed since last inspection (stale old_string)
- Whitespace differences (tabs vs spaces)
- Line ending differences (LF vs CRLF)
- Copy-paste errors

**Detection Pattern**: `String to replace not found in file`

**Prevention**:
- Re-read file immediately before Edit
- Use exact string copies
- Include sufficient context in old_string
- Verify file hasn't changed

**Recovery**:
1. Re-read file to get current content
2. Locate target section
3. Copy exact current string
4. Retry Edit with correct old_string

**Automation Potential**: High

---

## Uncategorized Errors (4.6%)

**Remaining**: 61 errors

**Breakdown**:
- Low-frequency unique errors: ~35 errors (2.6%)
- Rare edge cases: ~15 errors (1.1%)
- Other tool-specific errors: ~11 errors (0.8%)

These occur too infrequently (<0.5% each) to warrant dedicated categories.

---

## Automation Summary

**Automated Prevention Available**:
| Category | Errors | Tool | Coverage |
|----------|--------|------|----------|
| File Not Found | 250 (18.7%) | `validate-path.sh` | 65.2% |
| File Size Exceeded | 84 (6.3%) | `check-file-size.sh` | 100% |
| Write Before Read | 70 (5.2%) | `check-read-before-write.sh` | 100% |
| **Total Automated** | **317 (23.7%)** | **3 tools** | **Weighted avg** |

**Automation Speedup**: 20.9x for automated categories

---

## Transferability

**Universal Categories** (90-100% transferable):
- Build/Compilation Errors
- Test Failures
- File Not Found
- File Size Limits
- Permission Denied
- Empty Command

**Portable Categories** (70-90% transferable):
- Command Not Found
- JSON Parsing
- String Not Found

**Context-Specific Categories** (40-70% transferable):
- Write Before Read (Claude Code specific)
- Request Interruption (AI assistant specific)
- MCP Server Errors (MCP-enabled systems)
- Go Module Exists (Go-specific)

**Overall Transferability**: ~85-90%

---

## Usage

### For Developers

1. **Error occurs** → Match to category using detection pattern
2. **Review common causes** → Identify root cause
3. **Apply prevention** → Check if automated tool available
4. **Execute recovery** → Follow category-specific steps

### For Tool Builders

1. **High automation potential** → Prioritize Categories 3, 4, 5, 11, 12
2. **Medium automation** → Consider Categories 6, 7, 9
3. **Low automation** → Manual handling for Categories 2, 8, 10

### For Project Adaptation

1. **Start with universal categories** (1-7, 10, 11, 13)
2. **Adapt context-specific** (8, 9, 12)
3. **Monitor uncategorized** → Create new categories if patterns emerge

---

**Source**: Bootstrap-003 Error Recovery Methodology
**Framework**: BAIME (Bootstrapped AI Methodology Engineering)
**Status**: Production-ready, validated with 1336 errors
**Coverage**: 95.4% (converged)
