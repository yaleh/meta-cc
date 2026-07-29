# Installation Guide

## Method 1: Claude Code Plugin Marketplace

Install meta-cc directly from within Claude Code:

```bash
/plugin marketplace add yaleh/meta-cc
/plugin install meta-cc
```

Then restart Claude Code. The plugin system handles everything:
- Installs slash commands (`/prompt-find`, `/prompt-list`, `/prompt-show`)
- Configures the MCP server automatically via `.mcp.json` (no manual `claude mcp add` needed)

## Method 1b: Codex Plugin Marketplace (preferred for Codex CLI 0.145+)

Codex CLI 0.145 and later ships its own plugin manager (`codex plugin ...`),
which reads the **same** marketplace/plugin manifest format meta-cc already
publishes for Claude Code. This is the preferred way to install meta-cc into
Codex — it installs the MCP server, skills, and version metadata together
and keeps them in one place, instead of hand-copying files.

```bash
# Download and extract a release bundle for your platform (see Method 2 below
# for the platform-specific archive URLs), then, from the extracted directory:
codex plugin marketplace add .
codex plugin add meta-cc@meta-cc-marketplace
```

You can also point `codex plugin marketplace add` at a git checkout of this
repository (`codex plugin marketplace add /path/to/meta-cc` or
`codex plugin marketplace add yaleh/meta-cc`) — the marketplace manifest at
the repository root resolves the plugin from `plugin-src/`.

**A brand-new Codex session is required after installing or updating the
plugin.** Codex resolves plugin-provided skills and MCP servers when a
session starts; an already-running session cannot hot-load them.

### Verifying the install

```bash
codex plugin list --json
```

Look for `"pluginId": "meta-cc@meta-cc-marketplace"` with `"installed": true`
and `"enabled": true`. Then check the resolved MCP registration:

```bash
codex mcp list
```

This should show exactly **one** `meta-cc` entry. If you also manually ran
`codex mcp add meta-cc -- ...` (see the minimal fallback below) in the same
`CODEX_HOME`, remove it first with `codex mcp remove meta-cc` — running both
the plugin and a separate global MCP registration at once creates two active
`meta-cc` tool servers, which Codex will present as duplicates.

### Upgrading

Download the newer release archive over the same local path you originally
pointed `codex plugin marketplace add` at (or `git pull` if you used a git
checkout), then re-run:

```bash
codex plugin add meta-cc@meta-cc-marketplace
```

This is Codex's supported refresh operation for a local marketplace. As of
Codex CLI 0.146.0, `codex plugin add` is **destructive**: it materializes the
new version and removes the previously cached version *before* any post-install
verification runs. There is therefore no automatic rollback — if the new
version turns out to be inconsistent, the old cache is already gone.
`codex plugin marketplace upgrade` is only needed for marketplaces added from a
remote Git URL.

For repository user-scope installs, run `make install-user` followed by
`make install-user-codex`. The second target performs the supported refresh and
then verifies that every discovery surface agrees on one version: the discovery
metadata (`plugin list`), both cached manifests, the source manifests, all three
skills, and the MCP binary's **self-reported** version (queried over MCP
`initialize`, not just its executability).

- If verification succeeds, the cache is aligned and the next session loads it
  cleanly — no dangling old-version path remains to warn about.
- If the Codex CLI lacks the plugin-manager commands, the target fails visibly
  *before* any destructive call, so the previous cache is left untouched.
- If post-install verification fails, the target exits non-zero with an error
  naming the missing/inconsistent artifact and an explicit manual recovery
  command (`codex plugin add meta-cc@meta-cc-marketplace --json`, i.e. re-run
  `make install-user-codex`). It performs **no automatic rollback**, because
  Codex has already removed the old cache.

**Start a new Codex session afterward** to pick up the updated skills/MCP
server. A running session keeps using whatever was resolved at its start and
does not hot-reload plugin upgrades.

### Uninstalling

```bash
codex plugin remove meta-cc@meta-cc-marketplace
codex plugin marketplace remove meta-cc-marketplace   # optional: drop the marketplace source too
```

### Minimal fallback: MCP server only

If you only want the MCP tools in Codex (no skills, no plugin manifest
tracking), register the packaged binary directly as a global MCP server:

```bash
codex mcp add meta-cc -- /absolute/path/to/meta-cc-mcp
```

Verify with `codex mcp list` (again, exactly one `meta-cc` entry), and
restart Codex. To remove it: `codex mcp remove meta-cc`. Do not combine this
with the plugin-manager install above in the same `CODEX_HOME` — pick one,
to avoid duplicate registrations.

## Method 2: Archive Install for Claude Code and Codex

For Claude Code, or for a Codex CLI older than 0.145 (no `codex plugin`
subcommand), download a platform-specific release archive and run the
included installer. For Codex CLI 0.145+, prefer Method 1b above — it uses
the same archive but drives it through Codex's own plugin manager instead of
copying files by hand.

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

The archive installer:
- Copies the `meta-cc-mcp` binary to `~/.local/bin/`
- Copies slash commands to `~/.claude/commands/`
- Copies Codex skills to `~/.codex/skills/`
- Copies Codex plugin metadata to `~/.codex/plugins/meta-cc/`
- Automatically merges Claude Code MCP server configuration into `~/.claude/mcp.json`
- Includes `.codex-mcp.json` for Codex MCP configuration

Use temp or custom destinations when testing:

```bash
INSTALL_DIR=/tmp/bin CLAUDE_DIR=/tmp/claude CODEX_HOME=/tmp/codex ./install.sh
```

Use host flags when you only want one integration:

```bash
INSTALL_CLAUDE=0 ./install.sh  # Codex files only
INSTALL_CODEX=0 ./install.sh   # Claude Code files only
```

## Manual Installation

If the automated installer fails, follow these steps:

### 1. Download Archive

```bash
# Download plugin package for your platform
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-<platform>.tar.gz | tar xz
cd meta-cc-plugin-<platform>
```

### 2. Install Binary

**Linux/macOS:**
```bash
mkdir -p ~/.local/bin
cp bin/meta-cc-mcp ~/.local/bin/meta-cc-mcp
chmod +x ~/.local/bin/meta-cc-mcp
```

**Windows:**
```bash
mkdir -p ~/.local/bin
cp bin/meta-cc-mcp.exe ~/.local/bin/meta-cc-mcp.exe
```

### 3. Install Claude Code and Codex Files

The archive uses a flat layout with `commands/` and `skills/` at the top level:

```bash
mkdir -p ~/.claude/commands
mkdir -p ~/.codex/skills

# Copy slash commands
cp commands/* ~/.claude/commands/

# Copy Codex skills
cp -R skills/* ~/.codex/skills/
```

### 4. Configure MCP for Claude Code

The archive includes a `.mcp.json` file. If you have `jq` installed, merge it automatically:

```bash
jq -s '.[0] * .[1]' ~/.claude/mcp.json .mcp.json > /tmp/mcp-merged.json && mv /tmp/mcp-merged.json ~/.claude/mcp.json
```

Otherwise, manually add to `~/.claude/mcp.json`:

```json
{
  "mcpServers": {
    "meta-cc": {
      "command": "meta-cc-mcp",
      "args": []
    }
  }
}
```

If you already have other MCP servers configured, add the `"meta-cc"` entry to the existing `"mcpServers"` object.

### 5. Configure MCP for Codex

Prefer Codex CLI's own plugin manager over manual copying (see Method 1b).
If your Codex install genuinely lacks `codex plugin`/`codex mcp` (CLI older
than 0.145), or the plugin manager path fails for some other reason, copy
the files into your Codex plugin location by hand and use `.codex-mcp.json`
as the MCP server template:

```bash
mkdir -p ~/.codex/plugins/meta-cc
cp -R .codex-plugin ~/.codex/plugins/meta-cc/
cp .codex-mcp.json ~/.codex/plugins/meta-cc/
```

Or register just the MCP server, without any plugin manifest, using the
minimal fallback from Method 1b:

```bash
codex mcp add meta-cc -- /absolute/path/to/meta-cc-mcp
```

Codex session discovery reads JSONL transcripts from `$CODEX_HOME/sessions` when `CODEX_HOME` is set, otherwise from `~/.codex/sessions`. meta-cc normalizes Codex `response_item` and `event_msg` records into the same internal message/tool schema used for Claude Code, so the common MCP tools work across both hosts.

## Verification

After installation, verify the setup:

```bash
# Check binary location
which meta-cc-mcp

# Verify binary is executable
ls -l ~/.local/bin/meta-cc-mcp
```

**In Claude Code:**

1. **Test MCP Tools**: In conversation, ask "What are my recent tool usage patterns?"
2. **Test Slash Commands**: Type `/prompt-list` and press Enter

**In Codex (plugin-manager install):**

1. **Check plugin state**: `codex plugin list --json` — `meta-cc@meta-cc-marketplace` should show `"installed": true` and `"enabled": true`
2. **Check the resolved MCP server**: `codex mcp list` — expect exactly one `meta-cc` entry
3. **Start a new Codex session** (a running session will not see a newly installed plugin)
4. **Test MCP Tools**: Ask "What are my recent tool usage patterns?"
5. **Test Skills**: Ask Codex to use the `prompt-list` skill

Useful smoke-test prompts for either host:

```text
Find user messages mentioning "refactor"
Which tools do I use most often?
Show token usage for recent assistant turns
```

## Troubleshooting

### Binary not found

**Issue**: `meta-cc-mcp: command not found`

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
   ls ~/.claude/commands/prompt-*.md
   ```
3. **Check command permissions**:
   ```bash
   chmod +r ~/.claude/commands/prompt-find.md
   ```
4. **Check Claude Code settings** to ensure slash commands are enabled

### Codex skills not working

**Issue**: Codex does not see `prompt-find`, `prompt-list`, or `prompt-show`

**Solutions**:

1. **Start a brand-new Codex session** after installation or upgrade. An
   already-running session resolved its plugins/skills at startup and will
   not pick up a change made afterward.
2. **Plugin-manager install**: confirm the plugin is actually enabled and
   find where its skills were unpacked:
   ```bash
   codex plugin list --json | jq '.installed[] | select(.pluginId=="meta-cc@meta-cc-marketplace")'
   ```
   If `"enabled"` is `false`, re-run `codex plugin add meta-cc@meta-cc-marketplace`.
3. **Manual/archive install**: verify skill files exist:
   ```bash
   ls ~/.codex/skills/prompt-*/SKILL.md
   ```
4. **Use a custom Codex home during tests**:
   ```bash
   CODEX_HOME=/tmp/codex ./install-skills.sh
   ```

### Codex reports two `meta-cc` MCP servers

**Issue**: `codex mcp list` shows more than one active entry named `meta-cc`.

**Cause**: The plugin-manager install (Method 1b) and the minimal
`codex mcp add meta-cc -- ...` fallback were both applied in the same
`CODEX_HOME`. Each registers its own MCP server, and Codex does not merge
them.

**Solution**: Pick one path. To keep the plugin-manager install, remove the
manual registration:
```bash
codex mcp remove meta-cc
```
To keep the minimal install instead, remove the plugin:
```bash
codex plugin remove meta-cc@meta-cc-marketplace
```
Then start a new Codex session.

### Codex sessions not found

**Issue**: MCP tools return no sessions for a Codex project.

**Solutions**:

1. **Check the Codex session root**:
   ```bash
   echo "${CODEX_HOME:-$HOME/.codex}"
   find "${CODEX_HOME:-$HOME/.codex}/sessions" -name '*.jsonl' | tail
   ```
2. **Pass the project explicitly** when asking through MCP:
   ```json
   {"working_dir": "/absolute/path/to/project", "scope": "project"}
   ```
3. **Verify the transcript references the project path**. Codex stores sessions in date directories rather than Claude Code's project-hash directories, so meta-cc matches Codex project sessions by transcript content.

### Installation fails on macOS

**Issue**: macOS blocks execution due to Gatekeeper

**Solutions**:

1. **Allow unsigned binary**:
   ```bash
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
   mkdir -p ~/.local/bin ~/.claude/commands
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
   ls -l ~/.local/bin/meta-cc-mcp.exe
   ```

## Uninstallation

To remove meta-cc:

### Using uninstall script

```bash
cd meta-cc-plugin-<platform>
./uninstall.sh
```

The uninstall script removes the binary, Claude Code slash commands, Codex skills/plugin files (manual-install layout only), and the `meta-cc` entry from `~/.claude/mcp.json`.

### Uninstalling a plugin-manager Codex install

If you installed via Method 1b, use Codex's own commands instead of deleting files by hand:

```bash
codex plugin remove meta-cc@meta-cc-marketplace
codex plugin marketplace remove meta-cc-marketplace   # optional
# or, for the minimal codex mcp add fallback:
codex mcp remove meta-cc
```

Start a new Codex session afterward to confirm the tools/skills are gone.

### Manual uninstallation

```bash
# Remove binary
rm ~/.local/bin/meta-cc-mcp

# Remove Claude Code files
rm ~/.claude/commands/prompt-find.md
rm ~/.claude/commands/prompt-list.md
rm ~/.claude/commands/prompt-show.md

# Remove Codex files (manual/archive install layout)
rm -rf ~/.codex/skills/prompt-find ~/.codex/skills/prompt-list ~/.codex/skills/prompt-show
rm -rf ~/.codex/plugins/meta-cc

# Remove meta-cc from MCP configuration
jq 'del(.mcpServers["meta-cc"])' ~/.claude/mcp.json > /tmp/mcp.json && mv /tmp/mcp.json ~/.claude/mcp.json
```

## Upgrading

**Claude Code / manual archive install:**

1. **Download new version** using the Quick Install commands above
2. **Run install.sh** - it will overwrite existing binaries
3. **Restart Claude Code** to load the new version

The installer preserves your MCP configuration and existing settings.

**Codex plugin-manager install (Method 1b):**

1. Download the newer archive over the same local path (or `git pull` a git checkout)
2. Re-run `codex plugin add meta-cc@meta-cc-marketplace` to resolve and install the updated version
3. Confirm with `codex plugin list --json` that the version bumped
4. **Start a new Codex session** — a running session keeps the version it started with

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
   - Output of `meta-cc-mcp --version` (if binary runs)
3. **Community support**: See [Discussions](https://github.com/yaleh/meta-cc/discussions)

## Next Steps

After successful installation:

1. **Read the documentation**: [Getting Started](../../README.md)
2. **Browse prompts**: `/prompt-list` to see saved prompts, `/prompt-find <keywords>` to search
3. **Learn MCP tools**: See [MCP Guide](../guides/mcp.md)
4. **Ask naturally**: "Show me my recent tool errors" or "What are my work patterns?"
