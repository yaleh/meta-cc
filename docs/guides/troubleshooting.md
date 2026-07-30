# meta-cc Troubleshooting Guide

This document provides solutions to common issues.

## Installation Issues

### MCP binary not found

**Symptoms**:
```
command not found: meta-cc-mcp
```

**Solution**:
```bash
# Check if binary is installed
which meta-cc-mcp
ls -l ~/.local/bin/meta-cc-mcp

# Add ~/.local/bin to PATH if missing
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Permission denied

**Symptoms**:
```
permission denied: ~/.local/bin/meta-cc-mcp
```

**Solution**:
```bash
# Add executable permission
chmod +x ~/.local/bin/meta-cc-mcp
```

## Session File Location Issues

### "failed to locate session file"

**Symptoms**:
```
MCP error: failed to locate session file: no session file found
```

**Possible causes**:
1. Current directory is not the Claude Code project root
2. Session file does not exist

**Solution**:
```bash
# Check if session files exist for current project
ls ~/.claude/projects/
```

Make sure you run Claude Code from the project directory so the MCP server can locate session files.

### Codex sessions not found

**Symptoms**: MCP tools return no sessions for a Codex project.

**Possible causes**:
1. The canonical Codex root does not match where Codex actually writes.
   meta-cc resolves it as `META_CC_CODEX_ROOT` → `CODEX_HOME` → `~/.codex`
   (see [Discovery roots](../reference/codex-history-model.md#discovery-roots-and-database-compatibility-dir-069)).
2. The queried `working_dir` does not match the `cwd` recorded in the
   Codex thread (SQLite `threads` table) or rollout `session_meta`.

**Solution**:
```bash
ROOT="${META_CC_CODEX_ROOT:-${CODEX_HOME:-$HOME/.codex}}"
ls "$ROOT"/state_*.sqlite 2>/dev/null        # highest N wins when compatible
find "$ROOT/sessions" -name '*.jsonl' | tail # rollout-only fallback source
```

If no compatible `state_N.sqlite` exists, meta-cc lists sessions from the
rollout trees directly — rollouts missing a recorded `cwd` are skipped with
a warning, so cross-project leakage is never inferred.

## MCP Server Issues

### "unknown source type: package" error

**Symptoms**:
```
MCP error -32603: failed to get capability: unknown source type: package
```

**Root Cause**: Fixed in v0.26.6. Update to the latest version.

**Solution**: Reinstall via plugin marketplace or download the latest release archive.

## Slash Commands Issues

### Slash Commands not visible

**Possible causes**:
1. File location incorrect (should be in `~/.claude/commands/`)
2. Frontmatter format error
3. Claude Code not reloaded

**Solution**:
```bash
# Check file location
ls ~/.claude/commands/meta.md ~/.claude/commands/prompt-*.md

# Check frontmatter format
head -n 10 ~/.claude/commands/meta.md

# Restart Claude Code
# Close and reopen Claude Code
```

### Slash Commands execution failed

**Symptoms**:
```
Error executing command: ...
```

**Solution**: Restart Claude Code after installation, then check the MCP server is running (Settings → MCP).

## Output Issues

### Empty or malformed output

**Possible causes**:
1. Session file empty or malformed
2. MCP server not running

**Solution**:
```bash
# Check session file content
head ~/.claude/projects/<hash>/<session-id>.jsonl

# Verify MCP server status in Claude Code (Settings → MCP)
```

## Performance Issues

### Slow MCP queries

**Possible causes**:
1. Session file too large (Turn count > 1000)
2. Large result sets without limit

**Solution**:
```bash
# Check session file size
wc -l ~/.claude/projects/<hash>/<session-id>.jsonl
```

Use the `limit` parameter in MCP queries to restrict result size when needed.

## Debugging Tips

### Validate JSONL format

```bash
# Check JSONL file format
jq . ~/.claude/projects/<hash>/<session-id>.jsonl | head -n 50
```

## Getting Help

If the above solutions don't work, please:

1. **View project documentation**: [README.md](../../README.md)
2. **Submit an Issue**: [GitHub Issues](https://github.com/yaleh/meta-cc/issues)
3. **Check Claude Code documentation**: [Official Documentation](https://docs.claude.com/en/docs/claude-code)

When submitting an Issue, please include:
- meta-cc version (from plugin info or release archive name)
- Complete error message
- Session file size (`wc -l <session-file>`)
- Operating system and version
