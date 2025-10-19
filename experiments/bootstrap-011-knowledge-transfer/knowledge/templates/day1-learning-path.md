# Day-1 Learning Path: meta-cc Contributor

**Target Role**: New Contributor
**Prerequisites**: Go basics, git basics, Linux/macOS terminal
**Time Estimate**: 4-8 hours (one work day)
**Success Criteria**: Working dev environment + first trivial contribution committed

---

## Path Overview

### Learning Objectives

By the end of Day-1, you will:
1. ✅ Have a working meta-cc development environment
2. ✅ Understand what meta-cc does and why it exists
3. ✅ Know how to navigate the codebase and documentation
4. ✅ Make and commit your first trivial contribution

### Path Structure

```
Section 1: Environment Setup (1-2 hours)
    ↓
Section 2: Understanding meta-cc (1-2 hours)
    ↓
Section 3: Exploring the Codebase (1-2 hours)
    ↓
Section 4: First Contribution (2-4 hours)
    ↓
Day-1 Complete!
```

### What You'll Learn

- **Setup**: Clone, build, test, and run meta-cc
- **Concepts**: Session history, meta-cognition, MCP server
- **Navigation**: Find your way around code and docs
- **Contribution**: Make a trivial fix and submit PR

---

## Section 1: Environment Setup (1-2 hours)

### Learning Objectives

- Clone meta-cc repository successfully
- Install all dependencies
- Run test suite and verify passing
- Build project from source
- Run meta-cc CLI successfully

### Steps

#### 1.1 Clone Repository (5 min)

```bash
# Clone the repository
git clone https://github.com/yaleh/meta-cc.git
cd meta-cc

# Verify you're in the right place
ls -la  # Should see: README.md, Makefile, cmd/, docs/, etc.
```

**Checkpoint**: You have the meta-cc directory locally.

#### 1.2 Install Dependencies (10-15 min)

```bash
# Check Go version (need Go 1.21+)
go version

# If Go not installed or too old:
# - macOS: brew install go
# - Linux: Download from https://go.dev/dl/

# Install project dependencies
go mod download

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Checkpoint**: `go version` shows 1.21+, `go mod download` succeeds.

#### 1.3 Build and Test (30-60 min)

```bash
# Run the full build + lint + test suite
make all

# This runs:
# - golangci-lint (code quality checks)
# - go test ./... (all tests)
# - go build (compile binary)

# If successful, you'll see:
# ✓ Linting passed
# ✓ All tests passed
# ✓ Binary built: ./meta-cc
```

**Common issues**:
- `golangci-lint not found`: Run `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- Tests fail: Check Go version, try `go mod tidy`
- Build fails: Check error messages, ensure all dependencies downloaded

**Checkpoint**: `make all` passes without errors.

#### 1.4 Run meta-cc CLI (5 min)

```bash
# Run meta-cc to verify it works
./meta-cc --help

# You should see:
# meta-cc - Meta-Cognition for Claude Code
# Usage: meta-cc [command]
# Available Commands: parse, query-tools, query-messages, ...

# Try a simple command
./meta-cc --version
```

**Checkpoint**: Can run `./meta-cc --help` and see command list.

### Section 1 Validation

**Self-Assessment Checklist**:
- [ ] Repository cloned to local machine
- [ ] Go 1.21+ installed and verified
- [ ] `make all` passes without errors
- [ ] Can run `./meta-cc --help`
- [ ] Understand what `make all` does (lint, test, build)

**If all checked**: Proceed to Section 2!

**Estimated time**: 1-2 hours (depending on internet speed and troubleshooting)

---

## Section 2: Understanding meta-cc (1-2 hours)

### Learning Objectives

- Understand what meta-cc does (in one sentence)
- Understand key concepts: session history, meta-cognition
- Know what the MCP server provides
- Navigate to relevant documentation
- Understand the project's purpose

### Steps

#### 2.1 Read Project README (15 min)

```bash
# Open README.md
cat README.md

# Or use your preferred editor/browser
```

**Key questions to answer**:
1. What is meta-cc? (Hint: "Meta-Cognition for Claude Code")
2. What problem does it solve?
3. What are the main features? (CLI, MCP server, slash commands)
4. Who uses it? (Claude Code users wanting to analyze their sessions)

**Checkpoint**: Can explain "meta-cc analyzes Claude Code session history to provide metacognitive insights."

#### 2.2 Understand Key Concepts (30 min)

**Session History**:
- Claude Code stores conversation history in JSONL files
- Location: `~/.claude/projects/<project>/sessions/<session>.jsonl`
- Contains: user messages, assistant messages, tool calls, errors
- meta-cc parses and analyzes this data

**Meta-Cognition**:
- Thinking about your own thinking
- In Claude Code: analyzing HOW you work, not just WHAT you build
- Examples: "What errors do I make repeatedly?", "What workflows do I use most?"

**MCP Server**:
- Model Context Protocol server
- Allows Claude to query session data directly
- Provides tools like `query_tools`, `query_user_messages`, etc.

**Checkpoint**: Can explain what session history is and why meta-cognition is useful.

#### 2.3 Explore Documentation (30-45 min)

Read these key docs (in order):

1. **CLAUDE.md** (15 min):
   - Development workflow
   - Common tasks
   - Quick reference
   - Read sections: "Quick Start", "Common Tasks"

2. **docs/plan.md** (15 min):
   - Project roadmap
   - Phase-by-phase plan
   - Current status
   - Skim to understand overall direction

3. **docs/guides/mcp.md** (15 min):
   - MCP server capabilities
   - Tool reference
   - Usage examples
   - Understand what queries are available

**Checkpoint**: Know where to find:
- Development workflow docs (CLAUDE.md)
- Project roadmap (docs/plan.md)
- MCP reference (docs/guides/mcp.md)

### Section 2 Validation

**Self-Assessment Checklist**:
- [ ] Can explain what meta-cc does in one sentence
- [ ] Understand what session history is
- [ ] Understand what meta-cognition means
- [ ] Know what the MCP server provides
- [ ] Located CLAUDE.md, docs/plan.md, docs/guides/mcp.md
- [ ] Can navigate to relevant docs when needed

**If all checked**: Proceed to Section 3!

**Estimated time**: 1-2 hours (reading and exploration)

---

## Section 3: Exploring the Codebase (1-2 hours)

### Learning Objectives

- Understand high-level project structure
- Know where to find CLI commands, core logic, tests
- Understand the flow: CLI → Internal → Output
- Navigate code using grep/find
- Run a specific test

### Steps

#### 3.1 Project Structure Overview (20 min)

```bash
# View top-level structure
tree -L 1 .

# Key directories:
# cmd/        - CLI commands (entry points)
# internal/   - Core logic (parser, analyzer, query)
# docs/       - Documentation
# tests/      - Test fixtures
# experiments/ - Bootstrap experiments
# .claude/    - Plugin definition (commands, agents)
```

**Mental Model**:
```
User runs CLI command (cmd/)
    ↓
Calls internal logic (internal/)
    ↓
Produces output (pkg/output/)
```

**Checkpoint**: Can list main directories and their purposes.

#### 3.2 Explore CLI Commands (20 min)

```bash
# Look at available commands
ls cmd/

# Key files:
# - root.go: Main command setup
# - parse.go: Parse session files
# - query_tools.go: Query tool usage
# - query_messages.go: Query user messages
# - mcp.go: MCP server

# Read a simple command
cat cmd/parse.go

# Notice pattern:
# 1. Define command (cobra)
# 2. Parse flags
# 3. Call internal logic
# 4. Format output
```

**Checkpoint**: Understand that cmd/ contains CLI entry points.

#### 3.3 Explore Internal Logic (30 min)

```bash
# Look at internal packages
ls internal/

# Key packages:
# - parser/: Parse JSONL session files
# - analyzer/: Analyze parsed data
# - query/: Query engines
# - filter/: SQL-like filtering
# - locator/: Find session files

# Explore parser (most fundamental)
ls internal/parser/
cat internal/parser/types.go

# Notice data structures:
# - ToolCall: Represents a tool invocation
# - Message: Represents user/assistant message
# - Session: Contains all session data
```

**Checkpoint**: Know that internal/ contains core logic, parser/ is fundamental.

#### 3.4 Find Code Using grep (15 min)

Practice finding code:

```bash
# Find all query commands
grep -r "query" cmd/ | grep ".go:"

# Find where ToolCall is defined
grep -r "type ToolCall" .

# Find all test files
find . -name "*_test.go" | head

# Find MCP tool definitions
grep -r "query_tools" .
```

**Checkpoint**: Can use grep/find to locate code.

#### 3.5 Run a Specific Test (15 min)

```bash
# Run tests for parser package
go test ./internal/parser/

# Run a specific test
go test ./internal/parser/ -run TestToolCall

# Run with verbose output
go test -v ./internal/parser/

# Understanding test output:
# - PASS/FAIL for each test
# - Coverage percentage
# - Execution time
```

**Checkpoint**: Can run tests for a specific package.

### Section 3 Validation

**Self-Assessment Checklist**:
- [ ] Know what cmd/, internal/, docs/ contain
- [ ] Understand CLI → Internal → Output flow
- [ ] Can navigate to parser, analyzer, query packages
- [ ] Can use grep to find code
- [ ] Can run tests for a specific package
- [ ] Understand project structure well enough to explore further

**If all checked**: Proceed to Section 4!

**Estimated time**: 1-2 hours (exploration and practice)

---

## Section 4: First Contribution (2-4 hours)

### Learning Objectives

- Find a good first issue
- Make a trivial fix (typo, comment, doc improvement)
- Commit with proper message format
- Run tests to verify fix
- Submit PR

### Steps

#### 4.1 Find a Good First Issue (30 min)

**Options for first contribution**:

1. **Documentation typo/improvement**:
   ```bash
   # Look for typos in docs
   grep -r "occurence" docs/  # Common typo: "occurence" → "occurrence"
   grep -r "seperate" docs/   # Common typo: "seperate" → "separate"

   # Or improve a doc section:
   # - Add missing example
   # - Clarify confusing section
   # - Fix broken link
   ```

2. **Code comment improvement**:
   ```bash
   # Find files with sparse comments
   grep -L "//" cmd/*.go

   # Add helpful comment explaining what a function does
   ```

3. **Test coverage improvement**:
   ```bash
   # Find packages with low test coverage
   go test -cover ./...

   # Add a simple test case
   ```

**For Day-1, choose**: Documentation typo or comment improvement (smallest, lowest risk)

**Checkpoint**: Identified a trivial fix to make (typo, comment, or doc improvement).

#### 4.2 Make the Fix (30-60 min)

Example: Fix documentation typo

```bash
# 1. Create feature branch
git checkout -b fix/documentation-typo

# 2. Make the change
vim docs/plan.md  # Fix typo: "occurence" → "occurrence"

# 3. Verify change
git diff

# 4. Run tests to ensure nothing broke
make all

# 5. If tests pass, you're ready to commit!
```

**Checkpoint**: Made the fix, `make all` passes.

#### 4.3 Commit with Proper Message (15 min)

meta-cc uses conventional commits:

```bash
# Format: <type>(<scope>): <description>
#
# Types: docs, feat, fix, refactor, test
# Scope: area affected (docs, cli, mcp, parser)
# Description: what changed (imperative mood)

# Example commit message:
git add docs/plan.md
git commit -m "docs: fix typo in plan.md (occurence → occurrence)"

# Full message format (if needed):
git commit -m "docs: fix typo in plan.md

Fixed spelling error: 'occurence' → 'occurrence'

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

**Commit Message Guidelines**:
- Type: `docs`, `feat`, `fix`, `refactor`, `test`
- Scope: Optional, e.g., `docs(plan)`, `fix(parser)`
- Description: Imperative mood ("fix typo", not "fixed typo")
- Body: Optional details if needed

**Checkpoint**: Commit made with proper message format.

#### 4.4 Push and Create PR (30-60 min)

```bash
# Push branch to your fork
git push origin fix/documentation-typo

# If you haven't forked yet:
# 1. Go to https://github.com/yaleh/meta-cc
# 2. Click "Fork" button
# 3. Add your fork as remote:
#    git remote add mine https://github.com/YOUR_USERNAME/meta-cc.git
# 4. Push to your fork:
#    git push mine fix/documentation-typo
```

**Create PR on GitHub**:
1. Go to your fork on GitHub
2. Click "Compare & pull request" button
3. Fill in PR description:
   ```markdown
   ## Summary
   Fixed typo in docs/plan.md: "occurence" → "occurrence"

   ## Test Plan
   - [x] Ran `make all` - passes
   - [x] Verified typo fixed in rendered docs

   ## Type of Change
   - [x] Documentation update
   ```

4. Click "Create pull request"

**Checkpoint**: PR submitted with passing tests.

### Section 4 Validation

**Self-Assessment Checklist**:
- [ ] Found a good first issue (typo, comment, or doc improvement)
- [ ] Made the fix in a feature branch
- [ ] Ran `make all` to verify no breakage
- [ ] Committed with proper conventional commit message
- [ ] Pushed branch to fork
- [ ] Created PR with description
- [ ] PR shows passing tests

**If all checked**: Day-1 Complete! 🎉

**Estimated time**: 2-4 hours (depending on issue complexity and familiarity with git/GitHub)

---

## Day-1 Complete! 🎉

### What You Accomplished

**Environment** ✅:
- Cloned meta-cc repository
- Installed dependencies
- Built project successfully
- Ran test suite

**Understanding** ✅:
- Understand what meta-cc does (meta-cognition for Claude Code)
- Know key concepts (session history, meta-cognition, MCP)
- Can navigate documentation

**Code Navigation** ✅:
- Understand project structure (cmd/, internal/, docs/)
- Know where to find CLI commands, core logic, tests
- Can use grep/find to locate code

**Contribution** ✅:
- Made first trivial contribution (typo/comment/doc fix)
- Used proper git workflow (branch, commit, PR)
- Submitted PR with passing tests

### What's Next

**Week-1 Path** (20-40 hours over first week):
- Deep dive into core modules (parser, analyzer, query)
- Understand MCP server architecture
- Learn common development workflows
- Make meaningful contribution (good first issue)
- **Goal**: Merged PR for a small feature

**Where to Go from Here**:
1. While waiting for PR review, explore more of the codebase
2. Read `docs/guides/plugin-development.md` to understand plugin system
3. Try running MCP queries: `./meta-cc mcp --help`
4. Look for "good first issue" labels in GitHub issues
5. Join discussions in existing PRs to learn the codebase

**Questions or Stuck?**:
- Check `CLAUDE.md` for common tasks
- Review `docs/troubleshooting.md` for common issues
- Look at recent PRs for examples
- Ask in PR comments or GitHub discussions

---

## Path Metadata

```yaml
path_type: Day-1
target_role: Contributor
prerequisites:
  - Go basics
  - git basics
  - terminal proficiency

time_estimate:
  minimum: 4 hours
  maximum: 8 hours
  average: 6 hours

sections:
  - name: Environment Setup
    time: 1-2 hours
    objectives: 4
    checkpoints: 4

  - name: Understanding meta-cc
    time: 1-2 hours
    objectives: 5
    checkpoints: 6

  - name: Exploring the Codebase
    time: 1-2 hours
    objectives: 6
    checkpoints: 6

  - name: First Contribution
    time: 2-4 hours
    objectives: 5
    checkpoints: 7

total_objectives: 20
total_checkpoints: 23

success_criteria:
  - working_environment: true
  - basic_understanding: true
  - code_navigation: true
  - first_contribution: true

next_path: Week-1 Learning Path
```

---

**Path Version**: 1.0
**Created**: 2025-10-17
**Validated**: Not yet (to be validated with real contributors)
**Status**: Template ready for use
