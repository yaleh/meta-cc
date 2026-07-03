# meta-cc Capability Awareness: Claude and Codex

This guide explains how the `meta-cc-insights` skill enables both Claude Code and OpenAI Codex to be continuously aware of meta-cc's 21 MCP analysis tools.

## Claude Code (Plugin Skill)

For Claude Code, the skill lives in `~/.claude/skills/meta-cc-insights/SKILL.md`.

The mechanism relies on Claude Code's **progressive disclosure**:
- The frontmatter `name` and `description` are always present in the agent's context
- The full body is only loaded when the `description` matches the task at hand

The `description` uses **trigger scenario language**:
> "When you repeat similar errors, need to reflect on recent work, want to understand your tool use patterns, or need to query session history, use the meta-cc insights tools for self-reflection and analysis."

This avoids spamming the agent with all 21 tools every time, yet still makes them discoverable exactly when needed.

## Codex (Skills Directory + AGENTS.md Anchor)

For Codex, the skill lives in `~/.codex/skills/meta-cc-insights/SKILL.md`, exactly the same file as Claude Code's skill.

In addition, the `AGENTS.md` anchor explicitly calls out the meta-cc capability, which helps Codex discover the skill even when the skill discovery heuristics might miss it.

Codex's skill mechanism is still evolving, so we use multiple overlapping signals to maximize discovery:
1. The skill file in `~/.codex/skills/`
2. The anchor in `AGENTS.md`
3. The MCP tools registered in `.codex-mcp.json`

## Install

Run `scripts/install/install-skills.sh` to install to both Claude Code and Codex.

To verify installation:
- Claude: `ls -la ~/.claude/skills/meta-cc-insights/SKILL.md`
- Codex: `ls -la ~/.codex/skills/meta-cc-insights/SKILL.md`

## Maintenance

When new MCP tools are added to meta-cc:
1. Add them to `plugin-src/skills/meta-cc-insights/SKILL.md`
2. Run `scripts/hooks/validate-skill-tools.sh` to verify
3. Run `make push` to run all checks before committing

The `validate-skill-tools.sh` script ensures the skill stays in sync with `GetToolDefinitions()` in `internal/mcp/tools/tools.go`.
