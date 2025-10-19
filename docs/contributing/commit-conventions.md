# Commit Message Conventions

This project uses **Conventional Commits** to enable automated CHANGELOG generation and maintain a clear commit history.

## Format

```
<type>(<scope>): <subject>

[optional body]

[optional footer]
```

### Type (Required)

The type indicates the nature of the change:

- **feat**: A new feature (→ "Added" in CHANGELOG)
- **fix**: A bug fix (→ "Fixed" in CHANGELOG)
- **docs**: Documentation changes (→ "Changed" in CHANGELOG)
- **refactor**: Code refactoring without feature changes (→ "Changed" in CHANGELOG)
- **perf**: Performance improvements (→ "Improved" in CHANGELOG)
- **test**: Adding or updating tests (→ "Changed" in CHANGELOG)
- **chore**: Maintenance tasks, dependency updates (→ "Changed" in CHANGELOG)
- **style**: Code style changes (formatting, etc.)
- **build**: Build system or external dependency changes
- **ci**: CI/CD configuration changes

### Scope (Optional)

The scope specifies what part of the codebase is affected:

```
feat(mcp): add new query tool
fix(cli): correct output format
docs(readme): update installation steps
refactor(parser): simplify error handling
```

Common scopes:
- `mcp` - MCP server
- `cli` - CLI tool
- `parser` - Session parser
- `query` - Query engine
- `agents` - Agent definitions
- `docs` - Documentation
- `tests` - Test suite

### Subject (Required)

The subject is a short description of the change:

- Use imperative mood: "add feature" not "added feature"
- Don't capitalize the first letter
- No period at the end
- Maximum 72 characters

### Body (Optional)

The body provides additional context:

- Separate from subject with a blank line
- Explain **what** and **why**, not **how**
- Wrap at 72 characters
- Use bullet points for multiple items

### Footer (Optional)

The footer references issues, breaking changes, or co-authors:

```
BREAKING CHANGE: API endpoint changed from /v1 to /v2
Fixes #123
Closes #456

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## Examples

### Good Examples

```
feat: add CHANGELOG automation script

Implements automatic CHANGELOG generation from conventional commits.
Removes manual editing step from release process.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

```
fix(mcp): correct session locator environment variable check

Fixed inverted condition that prevented session-only mode from working.
Now correctly checks CC_SESSION_ID and CC_PROJECT_HASH.

Fixes #42
```

```
docs: update installation guide with automated steps
```

```
refactor(parser): simplify error extraction logic
```

### Bad Examples

```
# ❌ No type prefix
Updated the README file

# ❌ Capitalized subject
feat: Add new feature

# ❌ Period at end
fix: correct the bug.

# ❌ Past tense
feat: added CHANGELOG automation

# ❌ Too vague
chore: fixes
```

## CHANGELOG Mapping

The automated CHANGELOG generation maps commit types to sections:

| Commit Type | CHANGELOG Section | Example |
|-------------|------------------|---------|
| `feat` | Added | New features and capabilities |
| `fix` | Fixed | Bug fixes |
| `docs` | Changed | Documentation updates |
| `refactor` | Changed | Code improvements |
| `perf` | Improved | Performance enhancements |
| `test` | Changed | Test additions/updates |
| `chore` | Changed | Maintenance and dependencies |
| `style` | Changed | Code formatting |
| `build` | Changed | Build system updates |
| `ci` | Changed | CI/CD configuration |

## Automated CHANGELOG Generation

When you run `./scripts/release.sh vX.Y.Z`:

1. Script calls `generate-changelog-entry.sh`
2. Parses commits since last release tag
3. Groups commits by type
4. Generates CHANGELOG entry in "Keep a Changelog" format
5. Inserts entry into CHANGELOG.md
6. Commits with version updates

### Manual CHANGELOG Editing

While CHANGELOG entries are auto-generated, you may want to:

- Add additional context to entries
- Group related changes under phase/feature headers
- Add migration guides for breaking changes
- Include technical details or examples

To do this:

1. Run release script to generate initial entry
2. Edit CHANGELOG.md to add context
3. Commit additional changes

## Tips for Writing Good Commits

### Be Descriptive

```
# ❌ Too vague
fix: bug fix

# ✓ Descriptive
fix(parser): handle empty session files without crashing
```

### Use Scopes for Clarity

```
# ❌ No scope (unclear what changed)
feat: add new command

# ✓ With scope (clear location)
feat(cli): add query-conversation command
```

### Explain Why, Not How

```
# ❌ Describes implementation
fix: change if condition to !opts.SessionOnly

# ✓ Describes problem and solution
fix(mcp): correct session-only mode environment variable detection

Fixed inverted condition that prevented proper session detection.
```

### Keep Commits Focused

Each commit should represent a single logical change:

```
# ❌ Multiple unrelated changes
feat: add feature X, fix bug Y, update docs

# ✓ Separate commits
feat: add feature X
fix: correct bug Y
docs: update feature X documentation
```

## Enforcement

- **Pre-commit hooks**: (Optional) Install hooks to validate commit messages
- **CI validation**: (Future) Add CI check for conventional commit format
- **Code review**: Reviewers should check commit message quality

## References

- [Conventional Commits Specification](https://www.conventionalcommits.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Semantic Versioning](https://semver.org/)

## Questions?

If you're unsure about commit message formatting:

1. Check recent commits: `git log --oneline -20`
2. See examples in this guide
3. Ask in pull request reviews

Remember: Good commit messages make the CHANGELOG better for everyone!
