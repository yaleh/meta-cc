# meta-cc

Meta-Cognition tool for Claude Code - analyze session history for workflow optimization.

## Features

- 🔍 Parse Claude Code session history (JSONL format)
- 📊 Statistical analysis of tool usage and errors
- 🎯 Pattern detection for workflow optimization
- 🚀 Zero dependencies - single binary deployment

## Installation

### From Source

```bash
git clone https://github.com/yale/meta-cc.git
cd meta-cc
make build
```

### Cross-Platform Binaries

```bash
# Build for all supported platforms
make cross-compile

# Binaries will be in build/ directory:
# - build/meta-cc-linux-amd64
# - build/meta-cc-linux-arm64
# - build/meta-cc-darwin-amd64
# - build/meta-cc-darwin-arm64
# - build/meta-cc-windows-amd64.exe
```

## Usage

```bash
# Show help
./meta-cc --help

# Show version
./meta-cc --version

# Global options
./meta-cc --session <session-id>    # Specify session ID
./meta-cc --project <path>          # Specify project path
./meta-cc --output json|md|csv      # Output format
```

## Claude Code Integration

meta-cc provides deep integration with Claude Code, allowing you to analyze session metadata directly within your conversation using Slash Commands.

### Installation Steps

#### 1. Install meta-cc CLI Tool

**Option A: Build from source**

```bash
# Clone the repository
git clone https://github.com/yale/meta-cc.git
cd meta-cc

# Build the binary
make build

# Install to system path
sudo cp meta-cc /usr/local/bin/meta-cc
sudo chmod +x /usr/local/bin/meta-cc

# Verify installation
meta-cc --version
```

**Option B: Download pre-compiled binary** (coming soon)

```bash
# Download latest release (Linux x64)
curl -L https://github.com/yale/meta-cc/releases/latest/download/meta-cc-linux-amd64 -o meta-cc

# Install to system path
sudo mv meta-cc /usr/local/bin/meta-cc
sudo chmod +x /usr/local/bin/meta-cc

# Verify installation
meta-cc --version
```

#### 2. Configure Slash Commands

Slash Commands are already included in the project (`.claude/commands/` directory).

**Project-level Slash Commands** (recommended):

```bash
# Slash Commands are already in the project
ls .claude/commands/
# meta-stats.md
# meta-errors.md

# No additional configuration needed - just open in Claude Code
```

**Personal-level Slash Commands** (optional, available globally):

```bash
# Copy to personal Slash Commands directory
mkdir -p ~/.claude/commands
cp .claude/commands/meta-stats.md ~/.claude/commands/
cp .claude/commands/meta-errors.md ~/.claude/commands/

# Now available in all Claude Code projects
```

#### 3. Run Integration Tests

```bash
# Run integration test script
bash tests/integration/slash_commands_test.sh
```

Expected output:
```
=== meta-cc Slash Commands Integration Test ===

[1/5] Checking meta-cc installation...
✅ meta-cc installed: /usr/local/bin/meta-cc

[2/5] Checking Slash Command files...
✅ Slash Command files exist

[3/5] Testing meta-cc parse stats...
✅ meta-cc parse stats executed successfully

[4/5] Testing meta-cc analyze errors...
✅ meta-cc analyze errors executed successfully

[5/5] Testing meta-cc parse extract...
✅ meta-cc parse extract executed successfully

=== All tests passed ✅ ===
```

### Available Slash Commands

#### `/meta-stats` - Session Statistics

Display statistical information about the current Claude Code session.

**Usage**:
```
/meta-stats
```

**Output example**:
```markdown
# Session Statistics

- **Total Turns**: 245
- **Tool Calls**: 853
- **Error Count**: 0
- **Error Rate**: 0.00%
- **Session Duration**: 3h 42m

## Tool Usage Frequency

| Tool | Count | Percentage |
|------|-------|------------|
| Bash | 320 | 37.5% |
| Read | 198 | 23.2% |
| Edit | 156 | 18.3% |
```

**Use cases**:
- Quickly understand session overview
- Check for tool usage anomalies (high error rate)
- Evaluate session efficiency
- Discover tool usage patterns

#### `/meta-errors` - Error Pattern Analysis

Analyze error patterns in the current session, detecting repeated errors.

**Usage**:
```
/meta-errors              # Use default window (20 turns)
/meta-errors 50           # Analyze last 50 turns
/meta-errors 100          # Analyze last 100 turns
```

**Output example**:
```markdown
## Error Data Extraction

Detected 12 error tool calls.

## Error Pattern Analysis (window size: 20)

# Error Pattern Analysis

Found 2 error pattern(s):

## Pattern 1: Bash

- **Type**: repeated_error
- **Occurrences**: 5 times
- **Signature**: `a3f2b1c4d5e6f7g8`
- **Error**: command not found: xyz

### Context

- **First Occurrence**: 2025-10-02T10:00:00.000Z
- **Last Occurrence**: 2025-10-02T10:15:00.000Z
- **Time Span**: 900 seconds (15.0 minutes)
- **Affected Turns**: 5

---

## Optimization Recommendations

Based on detected error patterns, consider the following:

1. **Check root cause of repeated errors**
2. **Use Claude Code Hooks to prevent errors**
3. **Adjust workflow**
```

**Use cases**:
- Identify repeated errors to avoid redundant debugging
- Discover workflow bottlenecks (frequent failures)
- Get optimization recommendations (hooks, alternatives, prompt improvements)
- Focus on recent errors (using window parameter)

### Troubleshooting

#### Issue 1: "❌ Error: meta-cc not installed or not in PATH"

**Cause**: meta-cc binary not installed or not in system PATH.

**Solution**:
```bash
# Check if meta-cc is installed
which meta-cc

# If not found, install meta-cc
cd /path/to/meta-cc
make build
sudo cp meta-cc /usr/local/bin/meta-cc
sudo chmod +x /usr/local/bin/meta-cc

# Verify installation
meta-cc --version
```

#### Issue 2: "failed to locate session file"

**Cause**: meta-cc cannot find the current session's JSONL file.

**Solution**:
```bash
# Option 1: Check environment variables (Claude Code may provide)
echo $CC_SESSION_ID
echo $CC_PROJECT_HASH

# Option 2: Manually specify session file
meta-cc parse stats --session <session-id>

# Option 3: Check if session file exists
ls ~/.claude/projects/
```

#### Issue 3: Slash Commands not visible or unavailable

**Cause**: Slash Command files in wrong location or Claude Code not loaded.

**Solution**:
```bash
# Check if files exist
ls .claude/commands/meta-stats.md
ls .claude/commands/meta-errors.md

# Restart Claude Code
# Close and reopen the project

# Check file format (frontmatter must be correct)
head -n 10 .claude/commands/meta-stats.md
```

#### Issue 4: Garbled output or errors

**Cause**: meta-cc version mismatch or incorrect command parameters.

**Solution**:
```bash
# Check meta-cc version
meta-cc --version

# Ensure using latest version
cd /path/to/meta-cc
git pull
make build
sudo cp meta-cc /usr/local/bin/meta-cc

# Test commands manually
meta-cc parse stats --output md
meta-cc analyze errors --window 20 --output md
```

#### Issue 5: Permission errors

**Cause**: meta-cc doesn't have permission to read session files.

**Solution**:
```bash
# Check session file permissions
ls -l ~/.claude/projects/

# Ensure current user has read permission
chmod -R u+r ~/.claude/projects/
```

### Advanced Usage

#### Combining Slash Commands and CLI

```bash
# Step 1: Quick view in Claude Code using /meta-stats
# /meta-stats

# Step 2: If high error rate found, analyze with /meta-errors
# /meta-errors

# Step 3: Export detailed error data for deep analysis
meta-cc parse extract --type tools --filter "status=error" --output csv > errors.csv

# Step 4: Generate complete report
meta-cc parse stats --output md > session-report.md
meta-cc analyze errors --output md >> session-report.md
```

#### Creating Custom Slash Commands

Based on meta-cc, you can create custom Slash Commands:

**Example: `.claude/commands/meta-tool-usage.md`**

```markdown
---
name: meta-tool-usage
description: Display usage details for a specific tool
allowed_tools: [Bash]
argument-hint: [tool-name]
---

```bash
TOOL_NAME=${1:-Bash}
meta-cc parse extract --type tools --filter "tool=$TOOL_NAME" --output md
```
```

**Usage**:
```
/meta-tool-usage Bash
/meta-tool-usage Read
```

### Environment Variables

meta-cc supports the following environment variables (if provided by Claude Code):

- `CC_SESSION_ID`: Current session ID
- `CC_PROJECT_HASH`: Project path hash

**Check environment variables**:
```bash
# Check in Slash Command
echo "Session ID: $CC_SESSION_ID"
echo "Project Hash: $CC_PROJECT_HASH"
```

If these environment variables are unavailable, meta-cc will automatically fall back to working directory inference.

### Reference Documentation

- [Claude Code Slash Commands Official Documentation](https://docs.claude.com/en/docs/claude-code/slash-commands)
- [meta-cc Command Reference](#usage)
- [meta-cc Technical Proposal](./docs/proposals/meta-cognition-proposal.md)
- [Troubleshooting Guide](./docs/troubleshooting.md)

## Development

### Prerequisites

- Go 1.21 or later
- Make (optional, for build automation)

### Build

```bash
# Using Make
make build

# Or using go directly
go build -o meta-cc
```

### Test

```bash
# Run all tests
make test

# Run with coverage
make test-coverage
# Open coverage.html in browser
```

### Available Make Targets

```bash
make build           # Build for current platform
make test            # Run tests
make test-coverage   # Run tests with coverage report
make clean           # Remove build artifacts
make install         # Install to GOPATH/bin
make cross-compile   # Build for all platforms
make deps            # Download and tidy dependencies
make help            # Show help message
```

## Supported Platforms

- Linux (amd64, arm64)
- macOS (amd64, arm64/Apple Silicon)
- Windows (amd64)

## Project Structure

```
meta-cc/
├── cmd/              # Command definitions (Cobra)
├── internal/         # Internal packages
│   └── testutil/    # Test utilities
├── pkg/              # Public packages
├── tests/            # Test files and fixtures
└── docs/             # Documentation
```

## License

MIT
