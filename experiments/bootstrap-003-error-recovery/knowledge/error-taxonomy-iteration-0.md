# Error Classification Taxonomy - Iteration 0

**Version**: 0.1 (Baseline)
**Date**: 2025-10-18
**Coverage**: 10 categories, 79.1% of observed errors

---

## Taxonomy Principles

- **MECE**: Mutually Exclusive, Collectively Exhaustive (goal, not yet achieved)
- **Actionable**: Each category has clear recovery path
- **Observable**: Each category has detectable error signatures
- **Hierarchical**: Top-level categories with sub-categories

---

## Error Categories (10)

### 1. Build/Compilation Errors

**Definition**: Errors from Go compiler (go build, go test) indicating syntax or type issues

**Sub-categories**:
- 1.1 Syntax errors (unused imports, undefined variables)
- 1.2 Type mismatches
- 1.3 Module conflicts (go.mod already exists)
- 1.4 Import path errors (internal package access)

**Frequency**: 15.0% of total errors (200/1336)

**Impact**: Blocking - prevents code execution

**Detection Signatures**:
- `# github.com/yale/meta-cc/...`
- `go: /home/yale/work/meta-cc/go.mod already exists`
- `package command-line-arguments`
- `.go:[line]:[col]: [error message]`

**Common Causes**:
- Syntax errors in Go code
- Unused imports/variables
- Internal package access violations
- Module initialization conflicts

**Example**:
```
# github.com/yale/meta-cc/cmd
cmd/root.go:4:2: "fmt" imported and not used
```

---

### 2. Test Failures

**Definition**: Go test failures indicating code does not meet test expectations

**Sub-categories**:
- 2.1 Assertion failures
- 2.2 Missing test fixtures
- 2.3 Test setup errors
- 2.4 Build failures in tests

**Frequency**: 11.2% of total errors (150/1336)

**Impact**: Blocking - indicates code quality issues

**Detection Signatures**:
- `--- FAIL: TestName`
- `FAIL\tpackage\ttime`
- Test assertion messages

**Common Causes**:
- Code changes breaking tests
- Missing or incorrect test fixtures
- Test expectations not updated
- Build errors in test code

**Example**:
```
--- FAIL: TestLoadFixture (0.00s)
    fixtures_test.go:34: Fixture content should contain 'sequence' field
FAIL
FAIL	github.com/yale/meta-cc/internal/testutil	0.003s
```

---

### 3. File Not Found Errors

**Definition**: File access errors due to non-existent files or incorrect paths

**Sub-categories**:
- 3.1 Read tool - file does not exist
- 3.2 Bash command - no such file or directory
- 3.3 Incorrect file path
- 3.4 File deleted or moved

**Frequency**: 18.7% of total errors (250/1336)

**Impact**: Blocking - cannot proceed without file

**Detection Signatures**:
- `<tool_use_error>File does not exist.</tool_use_error>`
- `no such file or directory`
- `wc: /path/to/file: No such file or directory`
- `cannot access '/path/to/file'`

**Common Causes**:
- Typo in file path
- File not created yet (workflow order issue)
- File deleted or moved
- Working directory confusion

**Example**:
```
<tool_use_error>File does not exist.</tool_use_error>
```

---

### 4. File Content Size Exceeded

**Definition**: Read tool errors when file size exceeds token limits

**Sub-categories**:
- 4.1 File exceeds token limit (25000 tokens)
- 4.2 Missing pagination parameters

**Frequency**: 1.5% of total errors (20/1336)

**Impact**: Recoverable - use offset/limit parameters

**Detection Signatures**:
- `File content (N tokens) exceeds maximum allowed tokens (25000)`
- `Please use offset and limit parameters`

**Common Causes**:
- Large files (>25000 tokens)
- Attempting to read entire file without pagination
- No limit parameter specified

**Example**:
```
File content (46892 tokens) exceeds maximum allowed tokens (25000).
Please use offset and limit parameters to read specific portions of the file
```

---

### 5. Write Before Read Errors

**Definition**: Write tool errors due to Claude Code safety constraint requiring Read before Write

**Sub-categories**:
- 5.1 Attempting to overwrite existing file without reading
- 5.2 Write to Edit tool confusion

**Frequency**: 3.0% of total errors (40/1336)

**Impact**: Recoverable - read file first, then write

**Detection Signatures**:
- `<tool_use_error>File has not been read yet. Read it first before writing to it.</tool_use_error>`

**Common Causes**:
- Claude Code safety constraint not followed
- Workflow error (should use Edit for existing files)
- Forgetting to read file first

**Example**:
```
<tool_use_error>File has not been read yet. Read it first before writing to it.</tool_use_error>
```

---

### 6. Command Not Found Errors

**Definition**: Bash errors when command/binary is not in PATH or not installed

**Sub-categories**:
- 6.1 Binary not built yet
- 6.2 Binary not in PATH
- 6.3 Missing dependencies
- 6.4 Typo in command name

**Frequency**: 3.7% of total errors (50/1336)

**Impact**: Blocking - command unavailable

**Detection Signatures**:
- `command not found: [command]`
- `/bin/bash: line 1: [command]: command not found`

**Common Causes**:
- Binary not built yet (need to run `make build`)
- Binary not installed to system PATH
- Missing dependencies
- Typo in command name

**Example**:
```
/bin/bash: line 1: meta-cc: command not found
```

---

### 7. JSON Parsing Errors

**Definition**: Bash errors from jq when parsing invalid JSON or incorrect filter syntax

**Sub-categories**:
- 7.1 Invalid JSON input
- 7.2 Incorrect jq filter syntax
- 7.3 Empty or null data
- 7.4 Type mismatch in jq filter

**Frequency**: 6.0% of total errors (80/1336)

**Impact**: Blocking - data processing fails

**Detection Signatures**:
- `parse error: Invalid numeric literal at line N, column M`
- `jq: error: Cannot index array with string`
- `jq: error: [error message]`

**Common Causes**:
- Invalid JSON from upstream command
- Incorrect jq filter syntax
- Empty output piped to jq
- Type assumptions in jq filter incorrect

**Example**:
```
parse error: Invalid numeric literal at line 1, column 8
```

---

### 8. Request Interruption

**Definition**: Task tool errors when user interrupts agent execution

**Sub-categories**:
- 8.1 User-initiated interruption
- 8.2 Claude Code workflow interruption

**Frequency**: 2.2% of total errors (30/1336)

**Impact**: User action - not a true error

**Detection Signatures**:
- `[Request interrupted by user for tool use]`
- `Command aborted before execution`

**Common Causes**:
- User manually interrupted execution
- User requested different approach
- Claude Code internal workflow change

**Example**:
```
[Request interrupted by user for tool use]
```

---

### 9. MCP Integration Errors

**Definition**: Errors from MCP tool queries (meta-cc, meta-insight, etc.)

**Sub-categories**:
- 9.1 MCP server not running/reachable
- 9.2 Query syntax errors
- 9.3 Missing capabilities
- 9.4 Timeout errors
- 9.5 Data format errors

**Frequency**: 17.1% of total errors (228/1336)

**Impact**: Varies (some queries fail gracefully, others block)

**Detection Signatures**:
- Tool name prefix: `mcp__`
- Various error messages from MCP tools

**Common Causes**:
- MCP server not running
- Incorrect query parameters
- Missing or outdated capabilities
- Network/IPC issues
- Data format mismatches

**Example**:
```
[MCP error messages vary widely]
```

---

### 10. Permission Denied Errors

**Definition**: Bash errors when insufficient permissions for file/command operations

**Sub-categories**:
- 10.1 File permission denied
- 10.2 Sudo without interactive terminal
- 10.3 Directory access denied

**Frequency**: 0.7% of total errors (10/1336)

**Impact**: Blocking - cannot perform privileged operation

**Detection Signatures**:
- `permission denied`
- `sudo: a terminal is required to read the password`
- `sudo: a password is required`

**Common Causes**:
- Insufficient file permissions
- Sudo in non-interactive environment
- Directory not readable/writable

**Example**:
```
sudo: a terminal is required to read the password; either use the -S option to read from standard input or configure an askpass helper
```

---

## Taxonomy Statistics

| Category | Count | % | Detection | Recoverability |
|----------|-------|---|-----------|----------------|
| 1. Build/Compilation | 200 | 15.0% | ✅ High | Manual fix |
| 2. Test Failures | 150 | 11.2% | ✅ High | Manual fix |
| 3. File Not Found | 250 | 18.7% | ✅ High | Path correction |
| 4. File Size Exceeded | 20 | 1.5% | ✅ High | Use pagination |
| 5. Write Before Read | 40 | 3.0% | ✅ High | Read first |
| 6. Command Not Found | 50 | 3.7% | ✅ High | Install/build |
| 7. JSON Parsing | 80 | 6.0% | ✅ High | Fix JSON/jq |
| 8. Request Interruption | 30 | 2.2% | ✅ High | N/A (user action) |
| 9. MCP Integration | 228 | 17.1% | ⚠️ Medium | Varies |
| 10. Permission Denied | 10 | 0.7% | ✅ High | Fix permissions |
| **Uncategorized** | 278 | 20.9% | ❌ Low | Unknown |
| **TOTAL** | 1336 | 100% | - | - |

**Coverage**: 79.1% (1058/1336 errors categorized)

---

## Gaps and Limitations (Iteration 0)

1. **Incomplete coverage**: 20.9% of errors uncategorized
2. **Broad MCP category**: Need sub-categorization for 17.1% of errors
3. **No severity levels**: All errors treated equally
4. **No frequency tracking**: Static snapshot, no trend analysis
5. **Manual detection**: No automated error classification yet
6. **Missing categories**: Likely additional error types not yet observed

---

## Next Iteration Goals

1. Expand to 12+ categories by analyzing uncategorized errors
2. Add severity levels (Critical, High, Medium, Low)
3. Break down MCP Integration category into sub-categories
4. Add automated detection patterns (regex, signatures)
5. Achieve >90% coverage
6. Add error frequency trends over time

---

**Version History**:
- v0.1 (2025-10-18): Initial baseline taxonomy with 10 categories
