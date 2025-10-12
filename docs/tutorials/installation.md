# Installation Guide

## Quick Install (Recommended)

### Linux (x86_64)
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-linux-amd64.tar.gz | tar xz
cd meta-cc-plugin-linux-amd64
./install.sh
```

### Linux (ARM64)
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-linux-arm64.tar.gz | tar xz
cd meta-cc-plugin-linux-arm64
./install.sh
```

### macOS (Intel)
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-darwin-amd64.tar.gz | tar xz
cd meta-cc-plugin-darwin-amd64
./install.sh
```

### macOS (Apple Silicon)
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-darwin-arm64.tar.gz | tar xz
cd meta-cc-plugin-darwin-arm64
./install.sh
```

### Windows (x86_64)

**Using Git Bash (Recommended):**
```bash
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-windows-amd64.tar.gz | tar xz
cd meta-cc-plugin-windows-amd64
./install.sh
```

**Manual Download:**
1. Download `meta-cc-plugin-windows-amd64.tar.gz` from [GitHub Releases](https://github.com/yaleh/meta-cc/releases/latest)
2. Extract the archive using 7-Zip or similar tool
3. Open Git Bash in the extracted directory
4. Run `./install.sh`

## Manual Installation

If the automated installer fails, follow these steps:

### 1. Download Binaries

**Linux (x86_64):**
```bash
wget https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-linux-amd64
wget https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-mcp-linux-amd64
```

**macOS (Apple Silicon):**
```bash
wget https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-darwin-arm64
wget https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-mcp-darwin-arm64
```

**Windows (x86_64):**
```bash
wget https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-windows-amd64.exe
wget https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-mcp-windows-amd64.exe
```

### 2. Install Binaries

**Linux/macOS:**
```bash
mkdir -p ~/.local/bin
mv meta-cc-<platform> ~/.local/bin/meta-cc
mv meta-cc-mcp-<platform> ~/.local/bin/meta-cc-mcp
chmod +x ~/.local/bin/meta-cc ~/.local/bin/meta-cc-mcp
```

**Windows:**
```bash
mkdir -p ~/.local/bin
mv meta-cc-windows-amd64.exe ~/.local/bin/meta-cc.exe
mv meta-cc-mcp-windows-amd64.exe ~/.local/bin/meta-cc-mcp.exe
```

### 3. Install Claude Code Files

```bash
# Download plugin package for your platform
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-<platform>.tar.gz | tar xz
cd meta-cc-plugin-<platform>

# Copy integration files
mkdir -p ~/.claude/commands ~/.claude/agents
cp .claude/commands/* ~/.claude/commands/
cp .claude/agents/* ~/.claude/agents/
```

### 4. Configure MCP

Edit `~/.claude/mcp.json` and add the meta-cc server:

```json
{
  "mcpServers": {
    "meta-cc": {
      "command": "meta-cc-mcp",
      "args": [],
      "disabled": false
    }
  }
}
```

If you already have other MCP servers configured, add the `"meta-cc"` entry to the existing `"mcpServers"` object.

## Verification

After installation, verify the setup:

```bash
# Check binary version
meta-cc --version

# Check binary location
which meta-cc

# Test MCP server binary
meta-cc-mcp --version
```

**In Claude Code:**

1. **Test Slash Command**: Type `/meta-stats` and press Enter
2. **Test Subagent**: Type `@meta-coach` in a new conversation
3. **Test MCP Tools**: In conversation, ask "What are my recent tool usage patterns?"

## Troubleshooting

### Binary not found

**Issue**: `meta-cc: command not found`

**Solution**: Add `~/.local/bin` to PATH:

```bash
# For bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# For zsh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# For fish
fish_add_path ~/.local/bin
```

**Windows (Git Bash)**:
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bash_profile
source ~/.bash_profile
```

### MCP server not connecting

**Issue**: MCP server fails to start or times out

**Solutions**:

1. **Check MCP logs** in Claude Code settings (Settings → MCP)
2. **Verify binary is executable**:
   ```bash
   ls -l ~/.local/bin/meta-cc-mcp
   chmod +x ~/.local/bin/meta-cc-mcp
   ```
3. **Test MCP server manually**:
   ```bash
   meta-cc-mcp
   # Should start and wait for JSON-RPC messages
   # Press Ctrl+C to exit
   ```
4. **Check MCP configuration**:
   ```bash
   cat ~/.claude/mcp.json
   # Verify meta-cc entry exists and is valid JSON
   ```

### Slash commands not working

**Issue**: Slash commands not recognized in Claude Code

**Solutions**:

1. **Restart Claude Code** after installation
2. **Verify command files exist**:
   ```bash
   ls -l ~/.claude/commands/meta-*
   ```
3. **Check command permissions**:
   ```bash
   chmod +r ~/.claude/commands/meta-*.md
   ```
4. **Check Claude Code settings** to ensure slash commands are enabled

### Subagents not working

**Issue**: Subagent `@meta-coach` not recognized

**Solutions**:

1. **Restart Claude Code** after installation
2. **Verify agent files exist**:
   ```bash
   ls -l ~/.claude/agents/meta-*
   ```
3. **Check agent JSON syntax**:
   ```bash
   jq empty ~/.claude/agents/meta-coach.json
   ```

### Installation fails on macOS

**Issue**: macOS blocks execution due to Gatekeeper

**Solutions**:

1. **Allow unsigned binary**:
   ```bash
   xattr -d com.apple.quarantine ~/.local/bin/meta-cc
   xattr -d com.apple.quarantine ~/.local/bin/meta-cc-mcp
   ```
2. **Or use System Settings**:
   - Go to System Settings → Privacy & Security
   - Allow the binary to run

### Permission denied errors

**Issue**: Permission errors during installation

**Solutions**:

1. **Ensure write permissions**:
   ```bash
   mkdir -p ~/.local/bin ~/.claude/commands ~/.claude/agents
   chmod u+w ~/.local/bin ~/.claude
   ```
2. **Check disk space**:
   ```bash
   df -h ~
   ```
3. **Run without sudo** (installation should not require root)

### Windows-specific issues

**Issue**: Installation fails on Windows

**Solutions**:

1. **Use Git Bash** (not PowerShell or CMD)
2. **Check PATH in Git Bash**:
   ```bash
   echo $PATH | tr ':' '\n' | grep local
   ```
3. **Verify .exe extensions**:
   ```bash
   ls -l ~/.local/bin/meta-cc.exe
   ```

## Uninstallation

To remove meta-cc:

### Using uninstall script

```bash
cd meta-cc-plugin-<platform>
./uninstall.sh
```

### Manual uninstallation

```bash
# Remove binaries
rm ~/.local/bin/meta-cc ~/.local/bin/meta-cc-mcp

# Remove Claude Code files
rm -rf ~/.claude/commands/meta-*
rm -rf ~/.claude/agents/meta-*

# Manually edit ~/.claude/mcp.json to remove meta-cc server
```

**Note**: Uninstallation preserves `~/.claude/mcp.json` to avoid breaking other MCP servers. You must manually remove the `"meta-cc"` entry from the `"mcpServers"` object.

## Upgrading

To upgrade to a newer version:

1. **Download new version** using the Quick Install commands above
2. **Run install.sh** - it will overwrite existing binaries
3. **Restart Claude Code** to load the new version

The installer preserves your MCP configuration and existing settings.

## Platform-Specific Notes

### Linux

- **Distributions**: Tested on Ubuntu 22.04+, Debian 11+, Fedora 38+
- **Dependencies**: None (statically compiled binaries)
- **systemd**: Not required (MCP server runs on-demand)

### macOS

- **Versions**: Tested on macOS 12 (Monterey) and later
- **Gatekeeper**: See "Installation fails on macOS" troubleshooting
- **Homebrew**: Not required (standalone binaries)

### Windows

- **Requirements**: Git Bash (part of Git for Windows)
- **PowerShell**: Not supported (use Git Bash)
- **WSL**: Not required (native Windows binaries)

## Getting Help

If you encounter issues not covered in this guide:

1. **Check existing issues**: [GitHub Issues](https://github.com/yaleh/meta-cc/issues)
2. **Create new issue**: Include:
   - Operating system and version
   - Installation method used
   - Complete error messages
   - Output of `meta-cc --version` (if binary runs)
3. **Community support**: See [Discussions](https://github.com/yaleh/meta-cc/discussions)

## Next Steps

After successful installation:

1. **Read the documentation**: [Getting Started](../../README.md)
2. **Try slash commands**: `/meta-stats`, `/meta-errors`, `/meta-timeline`
3. **Explore subagents**: `@meta-coach` for workflow analysis
4. **Learn MCP tools**: See [MCP Guide](../guides/mcp.md)
