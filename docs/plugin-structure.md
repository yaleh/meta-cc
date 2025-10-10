# Plugin Structure

meta-cc follows the Claude Code plugin standard with the following structure:

```
meta-cc/
├── plugin.json              # Plugin metadata and manifest
├── install.sh               # Installation script
├── uninstall.sh             # Uninstallation script
├── bin/                     # Binaries (meta-cc, meta-cc-mcp)
├── .claude/
│   ├── commands/            # Slash commands
│   ├── agents/              # Subagent definitions
│   └── lib/
│       ├── mcp-config.json  # MCP configuration template
│       └── meta-utils.sh    # Shared utilities
└── docs/                    # Documentation
```
