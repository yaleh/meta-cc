# meta-cc Troubleshooting Guide

This document provides solutions to common issues.

## Installation Issues

### meta-cc not found

**Symptoms**:
```
command not found: meta-cc
```

**Solution**:
```bash
# Confirm meta-cc is installed
which meta-cc

# If not installed, build and install
cd /path/to/meta-cc
make build
sudo cp meta-cc /usr/local/bin/meta-cc
sudo chmod +x /usr/local/bin/meta-cc
```

### Permission denied

**Symptoms**:
```
permission denied: /usr/local/bin/meta-cc
```

**Solution**:
```bash
# Add executable permission
sudo chmod +x /usr/local/bin/meta-cc
```

## Session File Location Issues

### "failed to locate session file"

**Symptoms**:
```
Error: failed to locate session file: no session file found
```

**Possible causes**:
1. Environment variables `CC_SESSION_ID` and `CC_PROJECT_HASH` not set (when using `--session-only`)
2. Current directory is not the Claude Code project root
3. Session file does not exist

**Solution**:
```bash
# Option 1: Manually specify session ID
meta-cc parse stats --session <session-id>

# Option 2: Manually specify project path
meta-cc parse stats --project /path/to/project

# Option 3: Use --session-only with environment variables
export CC_SESSION_ID=<session-id>
export CC_PROJECT_HASH=<project-hash>
meta-cc parse stats --session-only

# Option 4: Check if session file exists
ls ~/.claude/projects/
```

### Environment variables not working

**Symptoms**:
```
Error: session location failed: failed to locate session file: tried session ID, project path, and env vars
```

**Root Cause**: Environment variables `CC_SESSION_ID` and `CC_PROJECT_HASH` are only checked when using `--session-only` flag (by design).

**Solution**:
```bash
# CORRECT: Use --session-only flag with environment variables
export CC_SESSION_ID=<session-id>
export CC_PROJECT_HASH=<project-hash>
meta-cc parse stats --session-only

# INCORRECT: Environment variables without --session-only (will use project-level default)
export CC_SESSION_ID=<session-id>
meta-cc parse stats  # This uses project-level analysis, ignores env vars
```

**Why this design?**
- **Default behavior** (no flags): Project-level analysis using current directory
- **`--session-only` flag**: Session-level analysis using environment variables
- This prevents unintended session-only mode when environment variables are set globally

## MCP Server Issues

### "unknown source type: package" error

**Symptoms**:
```
MCP error -32603: failed to get capability: unknown source type: package
```

**Root Cause**: Fixed in v0.26.6. Update to the latest version.

**Solution**:
```bash
# Update meta-cc to v0.26.6 or later
cd /path/to/meta-cc
git pull
make build
sudo cp meta-cc /usr/local/bin/meta-cc
sudo cp meta-cc-mcp /usr/local/bin/meta-cc-mcp

# Or download from GitHub releases
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-linux-amd64 -o meta-cc
sudo mv meta-cc /usr/local/bin/meta-cc
sudo chmod +x /usr/local/bin/meta-cc
```

### MCP `scope: "session"` not working

**Symptoms**:
```
MCP error -32603: meta-cc error: command failed with exit code (stderr empty)
Command: /path/to/meta-cc --session-only parse stats --output jsonl
```

**Root Cause**: Fixed in v0.26.6. The `--session-only` flag now correctly uses environment variables.

**Verification**:
```bash
# Test session-only mode manually
export CC_SESSION_ID=<session-id>
export CC_PROJECT_HASH=<project-hash>
meta-cc --session-only parse stats --output jsonl

# Should output session statistics
```

**Update to v0.26.6+**: This fix ensures MCP tools with `scope: "session"` parameter work correctly.

## Slash Commands Issues

### Slash Commands not visible

**Possible causes**:
1. File location incorrect (should be in `.claude/commands/`)
2. frontmatter format error
3. Claude Code not reloaded

**Solution**:
```bash
# Check file location
ls .claude/commands/

# Check frontmatter format
head -n 10 .claude/commands/meta-stats.md

# Restart Claude Code
# Close and reopen Claude Code
```

### Slash Commands execution failed

**Symptoms**:
```
Error executing command: ...
```

**Solution**:
```bash
# Manually run command to test
bash -c "$(sed -n '/```bash/,/```/p' .claude/commands/meta-stats.md | grep -v '```')"

# Check meta-cc version
meta-cc --version
```

## Output Issues

### Empty or malformed output

**Possible causes**:
1. Session file empty or malformed
2. meta-cc version too old
3. Incorrect command parameters

**Solution**:
```bash
# Check session file content
head ~/.claude/projects/<hash>/<session-id>.jsonl

# Update meta-cc
cd /path/to/meta-cc
git pull
make build
sudo cp meta-cc /usr/local/bin/meta-cc

# Test command
meta-cc parse stats --output md
```

## Performance Issues

### Slow command execution

**Possible causes**:
1. Session file too large (Turn count > 1000)
2. Window parameter too large

**Solution**:
```bash
# Use window parameter to limit analysis scope
meta-cc analyze errors --window 50

# Check session file size
wc -l ~/.claude/projects/<hash>/<session-id>.jsonl
```

## Debugging Tips

### Enable verbose logging

```bash
# Run command with verbose output
meta-cc parse stats --output md -v
```

### Check intermediate data

```bash
# Extract raw data
meta-cc parse extract --type turns --output json

# Check tool calls
meta-cc parse extract --type tools --output json
```

### Validate JSONL format

```bash
# Check JSONL file format
cat ~/.claude/projects/<hash>/<session-id>.jsonl | jq . | head -n 50
```

## Getting Help

If the above solutions don't work, please:

1. **View project documentation**: [README.md](../../README.md)
2. **Submit an Issue**: [GitHub Issues](https://github.com/yaleh/meta-cc/issues)
3. **Check Claude Code documentation**: [Official Documentation](https://docs.claude.com/en/docs/claude-code)

When submitting an Issue, please include:
- meta-cc version (`meta-cc --version`)
- Complete error message
- Session file size (`wc -l <session-file>`)
- Operating system and version
