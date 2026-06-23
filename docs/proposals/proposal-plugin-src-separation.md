# Proposal: plugin-src/ — Source/Deployed Separation for Claude Code Plugin

**Status**: Implemented (plugin-src/ directory exists with bin/, commands/, skills/, marketplace.json; see audit 2026-06-23)
**Date**: 2026-03-09
**Author**: Yale Huang / Claude Code Analysis
**Reviewed by**: Strict architectural review (2026-03-09) — 17 issues identified and addressed

---

## Executive Summary

This proposal introduces `plugin-src/` as the canonical source directory for all Claude Code
plugin artifacts in the meta-cc project, replacing the current scattered arrangement across
`.claude/commands/`, `.claude-plugin/`, and the project root.

The core principle: **source files are edited in `plugin-src/`, runtime always reads from an
installed copy** (local-scope or user-scope). The plugin is never loaded directly from its
own source directory.

---

## 1. Problem Statement

### 1.1 Current Architecture Issues

**Issue 1: No source/deployed separation**

`settings.local.json` registers a marketplace pointing at the project root. The plugin system
snapshots it into a cache (`~/.claude/plugins/cache/.../3.0.0/`) at install time, but the
cache goes stale whenever source files change. There is no explicit "install" step to push
changes to the deployed copy.

**Issue 2: Slash commands loaded twice**

`.claude/commands/*.md` are auto-loaded by Claude Code as project-level slash commands
(standard `.claude/commands/` scanning). The same files are also declared in `plugin.json`
and installed to `~/.claude/commands/` by the plugin system. When working in this project,
each command is available twice under different namespaces.

**Issue 3: plugin.json path references are wrong**

`plugin.json` declares commands as `./.claude/commands/prompt-*.md`. With the current
`source: "./.claude-plugin"`, CLAUDE_PLUGIN_ROOT resolves to
`/home/yale/work/meta-cc/.claude-plugin/`, making those paths invalid. Commands are only
reachable because they are also present at `~/.claude/commands/` from a previous `install.sh`
run.

**Issue 4: Three separate MCP registrations (partially resolved)**

The MCP server `meta-cc` was registered in three places simultaneously:
- `~/.claude.json` project-level `mcpServers` (manual — now removed)
- `~/.claude/mcp.json` user-level (from old `install.sh` run, PATH-based)
- Plugin system cache `.mcp.json` (via plugin, `${CLAUDE_PLUGIN_ROOT}/bin/`)

**Issue 5: Binary version drift**

Three binary files at different paths with different build timestamps:
- `/home/yale/work/meta-cc/bin/meta-cc-mcp` — current build
- `~/.claude/plugins/cache/.../3.0.0/bin/meta-cc-mcp` — snapshot from commit 862acb8
- `~/.local/bin/meta-cc-mcp` — from old `install.sh` run

---

## 2. Constraints and Known Platform Bugs

The following Claude Code platform constraints shape the design.

### 2.1 Plugin Scope Model

Claude Code has **three** distinct scopes. This proposal uses the exact CC terminology:

| CC Scope | Settings file | Gitignored? | Visibility |
|----------|---------------|-------------|------------|
| `local`  | `.claude/settings.local.json` | Yes | This machine only |
| `project` | `.claude/settings.json` | No | All contributors |
| `user`   | `~/.claude/settings.json` | N/A | All projects on machine |

For developer-only plugin activation: **local scope** (not "project scope").
For machine-wide activation across all projects: **user scope**.

### 2.2 `enabledPlugins` in `settings.local.json` is Silently Ignored

**Known bug** (GitHub issues #27247, #25086): `enabledPlugins` in `settings.local.json` is
silently ignored by Claude Code unless `"enabledPlugins": {}` (any value) also exists in
`~/.claude/settings.json`. This is a platform bug with no ETA for a fix.

**Workaround**: `make install-local` must verify the user's `~/.claude/settings.json`
contains an `enabledPlugins` key and add an empty one if absent. Without this check, the
local-scope installation silently does nothing.

### 2.3 Plugin Cache Behavior

Claude Code snapshots the plugin source directory into `~/.claude/plugins/cache/<marketplace>/<plugin>/<version>/`
at install or explicit update time. At runtime, it reads **exclusively from the cache** —
never from the source directory. Updating source files has no effect until the cache is
refreshed.

**Safe cache refresh mechanism** (writing directly to the cache is unsupported and risks
metadata inconsistency):

```bash
# Delete stale cache entry for this plugin/version
rm -rf ~/.claude/plugins/cache/<marketplace>/<plugin>/<version>/
# Claude Code recreates cache from source on next startup or /plugin update
```

`installed_plugins.json` tracks `gitCommitSha` per entry; writing cache files without
updating this JSON leaves metadata inconsistent and may cause CC to re-overwrite the files.

### 2.4 Plugin Cache Grows Without Cleanup

Each version gets its own cache subdirectory and is never automatically removed (GitHub issue
#16453). On this machine, two versions already exist: `2.3.6/` and `3.0.0/`. Any install
mechanism must clean up prior version directories explicitly.

### 2.5 `marketplace.json` `source` Field Must Be a Subdirectory

Schema validation rejects `"source": "."` (project root). Value must start with `./` and
point to a subdirectory — e.g., `"source": "./plugin-src"`. This was the root cause of the
original marketplace loading error.

---

## 3. Proposed Architecture

### 3.1 Core Principle

```
Source (edit here)       Install step            Runtime (read from here)
──────────────────  →   ─────────────    →   ────────────────────────────────
plugin-src/             make install-   →   Cache snapshot in
  commands/*.md           local             ~/.claude/plugins/cache/.../
  .claude-plugin/        make install-  →   Cache snapshot in
  .mcp.json               user              ~/.claude/plugins/cache/.../
  bin/ (gitignore)  ←── make build                   ↑
                                                      │
                         (CC reads cache, never source directory)
```

### 3.2 `plugin-src/` as Deployable Unit

`plugin-src/` contains **all** plugin artifacts:

| File/Dir | Type | Git tracked | Description |
|----------|------|-------------|-------------|
| `commands/*.md` | text | ✓ | Slash command source |
| `.claude-plugin/plugin.json` | text | ✓ | Plugin manifest |
| `.mcp.json` | text | ✓ | MCP server config |
| `bin/meta-cc-mcp` | binary | ✗ (gitignore) | Built by `make build` → `make stage` |

The binary lives in `plugin-src/bin/` because the plugin system copies the **entire**
directory to its cache. `${CLAUDE_PLUGIN_ROOT}/bin/meta-cc-mcp` resolves correctly only if
the binary is present inside the plugin directory at snapshot time.

**Note on `make build` output**: `make build` continues to output to `bin/meta-cc-mcp` at
the project root (preserving compatibility with `test-e2e-mcp` and `make clean`). A separate
`make stage` step copies the binary from `bin/` into `plugin-src/bin/` before the cache is
refreshed.

---

## 4. Directory Structure Changes

### 4.1 Repository Layout (After)

```
meta-cc/
│
├── plugin-src/                         ← NEW: plugin source + deployable unit
│   ├── .claude-plugin/
│   │   └── plugin.json                 ← commands: ["./commands/prompt-*.md"]
│   │                                      mcpServers: "./.mcp.json"
│   ├── .mcp.json                       ← command: "${CLAUDE_PLUGIN_ROOT}/bin/meta-cc-mcp"
│   ├── commands/
│   │   ├── prompt-find.md              ← MOVED from .claude/commands/
│   │   ├── prompt-list.md
│   │   └── prompt-show.md
│   └── bin/                            ← .gitignore; populated by `make stage`
│       └── meta-cc-mcp
│
├── .claude-plugin/
│   └── marketplace.json                ← ONLY marketplace definition
│                                          source: "./plugin-src"
│                                          (plugin.json and mcp.json removed)
│
├── .claude/
│   ├── settings.local.json             ← NOT committed (added to .gitignore)
│   │                                      generated by `make install-local`
│   ├── hooks/                          ← update trigger path to plugin-src/
│   └── (commands/ removed)            ← eliminates double-loading
│
├── bin/                                ← .gitignore; primary make build output
│   └── meta-cc-mcp                     ← used by test-e2e-mcp; staged to plugin-src/bin/
│
├── cmd/                                ← Go source (unchanged)
└── ...
```

### 4.2 What Moves / Changes / Is Generated

| Current location | New location | Action |
|-----------------|--------------|--------|
| `.claude/commands/prompt-*.md` | `plugin-src/commands/prompt-*.md` | Move |
| `.claude-plugin/plugin.json` | `plugin-src/.claude-plugin/plugin.json` | Move + update paths |
| `.claude-plugin/mcp.json` | `plugin-src/.mcp.json` | Move to plugin root |
| `.claude-plugin/marketplace.json` | `.claude-plugin/marketplace.json` | Update `source` field |
| `.claude/settings.local.json` | `.claude/settings.local.json` | Add to `.gitignore`; generate via `make install-local` |
| `.claude/commands/` directory | (removed) | Delete directory |

---

## 5. File Content Changes

### 5.1 `plugin-src/.claude-plugin/plugin.json`

```json
{
  "name": "meta-cc",
  "version": "3.0.0",
  "description": "...",
  "mcpServers": "./.mcp.json",
  "commands": [
    "./commands/prompt-find.md",
    "./commands/prompt-list.md",
    "./commands/prompt-show.md"
  ]
}
```

`commands` paths are relative to the plugin root (`./commands/`), not `./.claude/commands/`.
`mcpServers` points to `./.mcp.json` at the plugin root (official spec location).

### 5.2 `plugin-src/.mcp.json`

```json
{
  "meta-cc": {
    "command": "${CLAUDE_PLUGIN_ROOT}/bin/meta-cc-mcp",
    "args": [],
    "env": {}
  }
}
```

Content identical to current `.claude-plugin/mcp.json`. Moved to plugin root per spec.

### 5.3 `.claude-plugin/marketplace.json`

```json
{
  "name": "meta-cc-marketplace",
  "owner": { ... },
  "description": "...",
  "plugins": [
    {
      "name": "meta-cc",
      "source": "./plugin-src",
      ...
      "commands": [
        "./plugin-src/commands/prompt-find.md",
        "./plugin-src/commands/prompt-list.md",
        "./plugin-src/commands/prompt-show.md"
      ]
    }
  ]
}
```

`source` updated to `"./plugin-src"`. The `commands` field in `marketplace.json` is **display
metadata** for the `/plugin` discovery UI, resolved relative to the marketplace root (project
root). Paths must therefore be `./plugin-src/commands/...`, not `./commands/...`.

The existing `jq` path rewrite in `bundle-release` (`gsub("./.claude/commands/";
"./commands/")`) must be updated or removed as the `marketplace.json` commands paths in the
release bundle will differ from the source.

### 5.4 `.claude/settings.local.json` (generated, not committed)

```json
{
  "permissions": {
    "allow": ["Bash(make:*)", "Bash(go test:*)"]
  },
  "extraKnownMarketplaces": {
    "meta-cc-marketplace": {
      "source": {
        "source": "directory",
        "path": "<ABSOLUTE_PROJECT_ROOT>"
      }
    }
  },
  "enabledPlugins": {
    "meta-cc@meta-cc-marketplace": true
  }
}
```

`<ABSOLUTE_PROJECT_ROOT>` is substituted at generation time using `$(pwd)`. This file must
**not** be committed because the `directory` source type requires an absolute path that
differs per contributor machine.

`make install-local` generates this file and also ensures `~/.claude/settings.json` contains
an `enabledPlugins` key to work around the CC platform bug (Finding 2).

### 5.5 `.gitignore` additions

```gitignore
/plugin-src/bin/
/.claude/settings.local.json
```

---

## 6. Install Scopes

### 6.1 Local Scope (developer, this machine only)

Activation is stored in `.claude/settings.local.json` (gitignored, generated).

```
make build
  → Go build → bin/meta-cc-mcp                    (native platform)

make stage
  → copy bin/meta-cc-mcp → plugin-src/bin/meta-cc-mcp

make install-local
  → runs make build && make stage (if plugin-src/bin/ is absent or stale)
  → generates .claude/settings.local.json with absolute project path
  → ensures ~/.claude/settings.json contains "enabledPlugins" key (bug workaround)
  → purges ALL prior version dirs: rm -rf ~/.claude/plugins/cache/meta-cc-marketplace/meta-cc/
  → Claude Code recreates cache from plugin-src/ on next startup or /plugin update
  → updates installed_plugins.json entry (scope: local, projectPath: <project>)
```

### 6.2 User Scope (all projects on this machine)

Activation stored in `~/.claude/settings.json`. Uses a distinct marketplace name to prevent
collision with local-scope installations.

```
make install-user
  → guard: abort if local-scope is also active (prevent MCP name collision)
  → runs make build && make stage
  → copies plugin-src/ tree → ~/.local/share/meta-cc/
  → generates ~/.local/share/meta-cc/.claude-plugin/marketplace.json
      with source: "." (self-referential; user install is a complete deployed copy)
  → updates ~/.claude/settings.json:
      extraKnownMarketplaces["meta-cc-marketplace"] = { directory: ~/.local/share/meta-cc }
      enabledPlugins["meta-cc@meta-cc-marketplace"] = true
  → purges: rm -rf ~/.claude/plugins/cache/meta-cc-marketplace/meta-cc/
  → Claude Code recreates cache from ~/.local/share/meta-cc/ on next startup
```

**Simultaneous scope guard**: If both local and user scope were active simultaneously, both
would register an MCP server named `meta-cc` (collision) and both would install identically
named slash commands. `make install-user` must check for active local-scope and refuse with
a clear error directing the user to `make uninstall-local` first, and vice versa.

---

## 7. Makefile Changes

### New / Modified Targets

| Target | Description | Change type |
|--------|-------------|-------------|
| `build` | Output to `bin/meta-cc-mcp` | **No change** (preserves test-e2e-mcp compatibility) |
| `stage` | Copy `bin/meta-cc-mcp` → `plugin-src/bin/` | **New** |
| `install-local` | Stage + generate settings + purge cache + install | **New** |
| `install-user` | Stage + copy to `~/.local/share/meta-cc/` + register user marketplace | **New** |
| `uninstall-local` | Purge cache + remove generated `settings.local.json` + remove from `installed_plugins.json` | **New** |
| `uninstall-user` | Remove `~/.local/share/meta-cc/` + remove from `~/.claude/settings.json` + purge cache | **New** |
| `uninstall-legacy` | Remove `~/.local/bin/meta-cc-mcp` + `~/.claude/mcp.json` meta-cc entry (ordered) | **New** |
| `sync-plugin-files` | Update source path from `.claude/commands/` → `plugin-src/commands/` | **Update** |
| `check-plugin-sync` | Update verification paths | **Update** |
| `bundle-release` | Read commands from `plugin-src/commands/`; binary from `build/` cross-compile (unchanged) | **Partial update** |

### Targets Removed or Deprecated

| Target | Action | Reason |
|--------|--------|--------|
| `install` (Go install to PATH) | Remove | Replaced by `install-user` |

### `bundle-release` — Binary Source Unchanged

**Cross-compiled binaries from `build/` remain the source for release packages.**
`plugin-src/bin/` contains only the native build machine binary, which must not be shipped
in multi-platform release packages. The `bundle-release` loop iterates over platforms, taking
`build/meta-cc-mcp-<platform>` for each — this behavior is unchanged.

---

## 8. Scripts and CI Changes

| File | Required change |
|------|----------------|
| `scripts/sync-plugin-files.sh` | Source path: `.claude/commands/` → `plugin-src/commands/` |
| `scripts/install/install.sh` | Refactor to call `make install-user`; retain `lib/mcp-config.json` until refactor complete |
| `scripts/install/uninstall.sh` | Update slash command paths; add cache purge step |
| `.claude/hooks/pre-commit.sh` | Update trigger path from `.claude/commands/` → `plugin-src/` for version-bump logic |
| `.github/workflows/ci.yml` | Update any path references to `.claude/commands/` |
| `.github/workflows/release.yml` | Update `jq` path rewrite for `marketplace.json` commands field; verify `sync-plugin-files.sh` call |
| `scripts/release/bump-plugin-version.sh` | Update to bump version in `plugin-src/.claude-plugin/plugin.json` |

---

## 9. Cleanup of Legacy Artifacts

Cleanup must be performed in dependency order to avoid a broken intermediate state.

### Ordered cleanup (`make uninstall-legacy`)

```bash
# Step 1: Remove PATH-based MCP registration first
jq 'del(.mcpServers["meta-cc"])' ~/.claude/mcp.json > tmp && mv tmp ~/.claude/mcp.json

# Step 2: Remove binary from PATH (after MCP config, not before)
rm -f ~/.local/bin/meta-cc-mcp

# Step 3: Remove user-level commands (plugin system will reinstall from new location)
rm -f ~/.claude/commands/prompt-{find,list,show}.md
```

### Remaining artifacts to evaluate

| Artifact | Action |
|----------|--------|
| `lib/mcp-config.json` | **Retain** until `install.sh` is fully refactored to `make install-user` |
| `dist/` directory | Keep for `bundle-release` intermediate output |
| `~/.claude/plugins/cache/meta-cc-marketplace/meta-cc/2.3.6/` | Purge (stale ghost from old version) |

---

## 10. Developer Workflow (After)

### First-time setup (local scope)

```bash
git clone https://github.com/yaleh/meta-cc && cd meta-cc
make install-local   # build + stage + generate settings + prime cache
# Restart Claude Code → plugin:meta-cc:meta-cc appears under Built-in MCPs
```

### Daily development cycle

```bash
# Edit slash commands — requires re-install to take effect
vim plugin-src/commands/prompt-find.md
make install-local   # stage + purge cache; restart CC to pick up changes

# Edit Go source
vim cmd/mcp-server/tools.go
make build           # → bin/meta-cc-mcp (also used by make test)
make install-local   # → stage to plugin-src/bin/ + purge cache
```

### Machine-wide install (user scope)

```bash
make uninstall-local  # if local scope active; prevents collision
make install-user
# Restart Claude Code → plugin available in all projects
```

### Release

```bash
make bundle-release VERSION=v3.1.0
# Reads commands from plugin-src/commands/
# Reads binaries from build/ (cross-compiled, not from plugin-src/bin/)
```

---

## 11. Trade-offs and Constraints

| Concern | Assessment |
|---------|------------|
| `plugin-src/bin/` mixes source and gitignored binary | Accepted — consistent with existing `bin/` at project root; semantically clear as plugin deployable unit |
| No hot-reload for slash commands | Commands now require `make install-local` + CC restart. `make install-local` is fast; CC restart is the main friction. Considered acceptable given the architectural clarity gained |
| `make install-local` purges cache | Intentional; the only safe way to force CC to re-read updated source. See §2.3 |
| `settings.local.json` no longer committed | Each contributor runs `make install-local` once after clone. Initial friction, but eliminates absolute-path commit bug |
| `enabledPlugins` silent bug workaround | `make install-local` patches `~/.claude/settings.json` automatically; bug is invisible to contributors |
| Local-scope vs user-scope mutual exclusion | Enforced by guards in `install-local` and `install-user`; design avoids MCP name collision |

---

## 12. Files Not in Scope

The following are intentionally unchanged by this proposal:

- `cmd/`, `internal/` — Go source
- `scripts/release/bump-plugin-version.sh` — picks up path change when hooks updated
- `.claude/experiments/` — not part of plugin system
- `~/.claude/settings.json` `hooks` key — user stop-hook, unrelated
- `docs/` — update separately to reflect new paths
