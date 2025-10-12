# Unified Meta Command

Complete guide to the `/meta` command - a unified entry point for all meta-cognition capabilities with natural language intent matching.

## Overview

The `/meta` command provides a single interface to access 13+ capabilities for analyzing Claude Code session history. Instead of remembering separate commands, you express your intent in natural language, and the system finds and executes the best matching capability.

**Key Features**:
- **Natural language intent matching**: "show errors" → meta-errors capability
- **Semantic keyword scoring**: Ranks capabilities by relevance
- **Multi-source capability loading**: Local, GitHub, or package files
- **Composite intent detection**: Handles multi-step analysis requests
- **Transparent and discoverable**: Shows matching process and alternatives

## Usage

### Basic Syntax

```
/meta "natural language intent"
```

### Examples

```
/meta "show errors"           → Executes meta-errors
/meta "quality check"         → Executes meta-quality-scan
/meta "visualize timeline"    → Executes meta-timeline
/meta "analyze architecture"  → Executes meta-architecture
/meta "show tech debt"        → Executes meta-tech-debt
/meta "workflow patterns"     → Executes meta-workflow-patterns
/meta "performance analysis"  → Executes meta-performance
```

### Help Mode

```
/meta ""                      → Shows available capabilities
/meta "help"                  → Shows available capabilities
/meta "list"                  → Shows available capabilities
```

## How It Works

### 1. Discovery

Loads capabilities from configured sources:

```
Sources checked:
1. Local directory (if configured)
2. Package files (.tar.gz)
3. GitHub repositories (default: yaleh/meta-cc@main/commands)

Result: Capability index with metadata
```

### 2. Matching

Semantic keyword scoring:

```
Scoring algorithm:
- Name match: +3 points
- Description match: +2 points
- Keywords match: +1 point each
- Category match: +1 point

Threshold: score > 0
```

**Example**:
```
Intent: "show errors"

meta-errors:
  name: "meta-errors" (+3, contains "error")
  keywords: ["error", "debug"] (+1)
  Total: 4 points ✓

meta-quality-scan:
  keywords: ["error"] (+1)
  Total: 1 point

Best match: meta-errors (score: 4)
```

### 3. Reporting

Shows matching process:

```
Discovery: 13 capabilities loaded
Intent: "show errors"

Best match: meta-errors (score: 4)
  - Analyze session errors and debugging guidance
  - Category: debugging
  - Keywords: error, debug, troubleshoot, failure

Executing: meta-errors...
```

### 4. Execution

Loads and executes the capability:

```
1. Fetch capability content (from source)
2. Display capability metadata
3. Interpret and execute capability logic
4. Present results to user
```

## Capability Sources

### Configuration

Set capability sources via environment variable:

```bash
export META_CC_CAPABILITY_SOURCES="source1:source2:source3"
```

**Priority**: Left-to-right (left = highest priority).

### Source Types

#### 1. Local Directories

**Format**: `/path/to/directory`

**Example**:
```bash
export META_CC_CAPABILITY_SOURCES="capabilities/commands"
```

**Use case**: Local development (no cache, immediate reflection).

**Benefits**:
- Immediate changes
- No network dependency
- Fastest iteration

#### 2. Package Files

**Format**: `/path/to/file.tar.gz` or `https://example.com/capabilities.tar.gz`

**Examples**:
```bash
# Local package
export META_CC_CAPABILITY_SOURCES="./capabilities.tar.gz"

# GitHub Release package
export META_CC_CAPABILITY_SOURCES="https://github.com/yaleh/meta-cc/releases/latest/download/capabilities-latest.tar.gz"
```

**Use case**: Offline-friendly distribution, stable versions.

**Cache**:
- Release packages: 7-day cache
- Custom packages: 1-hour cache
- Location: `~/.capabilities-cache/packages/{hash}/`

#### 3. GitHub Repositories

**Format**: `owner/repo@branch/subdir` or `owner/repo/subdir`

**Examples**:
```bash
# Default branch (main)
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc/commands"

# Specific branch
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@develop/commands"

# Specific tag (version pinning)
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@v1.0.0/commands"

# Specific commit
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@abc123def/commands"
```

**Use case**: Production deployment, automatic updates.

**Cache**:
- Branches: 1-hour cache (mutable)
- Tags: 7-day cache (immutable)
- Location: `~/.capabilities-cache/github/{owner}/{repo}/{branch}/{subdir}/`

**CDN**: Uses jsDelivr CDN (https://cdn.jsdelivr.net) for improved performance.

### Multi-Source Configuration

Combine multiple sources for flexibility:

```bash
# Development: Local first, fallback to GitHub
export META_CC_CAPABILITY_SOURCES="~/dev/capabilities:yaleh/meta-cc@main/commands"

# Offline: Package first, fallback to local
export META_CC_CAPABILITY_SOURCES="./capabilities.tar.gz:capabilities/commands"

# Testing: Local test capabilities, then production
export META_CC_CAPABILITY_SOURCES="~/test-caps:yaleh/meta-cc@develop/commands"
```

### Default Source

If not configured:
```
META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@main/commands"
```

## CDN and Caching

### Cache Strategy

| Source Type | Cache TTL | Mutable? | Cache Location |
|-------------|-----------|----------|----------------|
| Local directory | No cache | Yes | N/A |
| Package (release) | 7 days | No | `~/.capabilities-cache/packages/` |
| Package (custom) | 1 hour | Yes | `~/.capabilities-cache/packages/` |
| GitHub (branch) | 1 hour | Yes | `~/.capabilities-cache/github/` |
| GitHub (tag) | 7 days | No | `~/.capabilities-cache/github/` |

### CDN Benefits

GitHub sources use jsDelivr CDN:
- ✅ No GitHub API rate limits
- ✅ Global CDN delivery (faster)
- ✅ Automatic caching with smart TTL

### Network Resilience

Automatic handling of network failures:

- **5xx server errors**: Exponential backoff retry (3 attempts: 1s, 2s, 4s)
- **Network unreachable**: Falls back to stale cache (up to 7 days old)
- **404 errors**: Clear error messages with troubleshooting suggestions

### Cache Metadata

Stored in `~/.capabilities-cache/.meta-cc-cache.json`:
```json
{
  "package_https://github.com/...": {
    "downloaded_at": "2025-10-12T10:00:00Z",
    "ttl_seconds": 604800,
    "url": "https://github.com/..."
  }
}
```

**Cleanup**: Automatic cleanup of expired entries (>7 days).

## Composite Intent Detection

Handles multi-step analysis requests automatically.

### Detection Logic

```
Threshold: ≥2 capabilities with score ≥ max(3, best_score*0.7)

Example:
  Intent: "analyze errors and visualize timeline"

  Matches:
  - meta-errors: score 5
  - meta-timeline: score 4

  Threshold: max(3, 5*0.7) = 3.5
  Both qualify → Composite intent detected
```

### Pipeline Patterns

Automatically infers execution patterns:

- **data_to_viz**: Data analysis → Visualization
- **analysis_to_guidance**: Analysis → Recommendations
- **multi_analysis**: Multiple independent analyses
- **sequential**: Ordered execution

**Example**:
```
Intent: "analyze errors and suggest fixes"

Pattern: analysis_to_guidance
Steps:
  1. meta-errors (data analysis)
  2. meta-guidance (recommendations)

Execution: User can request full pipeline
```

## MCP Tools for Capability Discovery

Programmatic access to capabilities via MCP:

### `list_capabilities()`

Returns capability index from all configured sources.

**Example**:
```json
{
  "capabilities": [
    {
      "name": "meta-errors",
      "description": "Analyze session errors",
      "keywords": ["error", "debug"],
      "category": "debugging",
      "source": "yaleh/meta-cc@main/commands"
    },
    ...
  ]
}
```

### `get_capability(name)`

Retrieves complete capability content.

**Parameters**:
- `name`: Capability name (e.g., "meta-errors")

**Returns**: Full capability markdown content with frontmatter.

## Building Capability Packages

### Create Package

```bash
make bundle-capabilities
```

**Output**: `build/capabilities-latest.tar.gz`

### Verify Package

```bash
tar -tzf build/capabilities-latest.tar.gz | head -20
```

**Expected structure**:
```
meta-errors.md
meta-quality-scan.md
meta-timeline.md
...
```

### Distribute Package

Upload to:
- GitHub Releases
- CDN
- Internal artifact repository

**Users configure**:
```bash
export META_CC_CAPABILITY_SOURCES="https://your-cdn.com/capabilities.tar.gz"
```

## Local Development

### Setup

```bash
export META_CC_CAPABILITY_SOURCES="capabilities/commands"
```

**Benefits**:
- Changes reflect immediately (no cache)
- Fast iteration
- No network dependencies

### Workflow

```bash
# 1. Edit capability
vim capabilities/commands/meta-errors.md

# 2. Test immediately
# In Claude Code: /meta "show errors"

# 3. Iterate
# Changes reflect instantly without cache invalidation
```

### Development Tips

1. **Use local source**: Always set `META_CC_CAPABILITY_SOURCES` for development
2. **Test semantic matching**: Try different intent phrases
3. **Check keyword scoring**: Verify keywords match expected intents
4. **Validate frontmatter**: Ensure name, description, keywords are correct

## Troubleshooting

### No Capabilities Found

**Problem**: `/meta "show errors"` says "No capabilities found".

**Solution**: Check capability sources:
```bash
echo $META_CC_CAPABILITY_SOURCES
# Should show valid source path

# Try default source
unset META_CC_CAPABILITY_SOURCES
# Uses: yaleh/meta-cc@main/commands
```

### Wrong Capability Executed

**Problem**: `/meta "errors"` executes wrong capability.

**Solution**: Improve intent specificity:
```bash
# Too vague
/meta "errors"

# More specific
/meta "show session errors"
/meta "analyze debugging errors"
```

**Or check capability keywords**:
```bash
# View capability index
/meta ""
# Check keywords for each capability
```

### Changes Not Reflecting

**Problem**: Edited capability but behavior unchanged.

**Solution**: Set local source:
```bash
export META_CC_CAPABILITY_SOURCES="capabilities/commands"
```

**Explanation**: Default GitHub source is cached (1-hour TTL).

### Cache Issues

**Problem**: Stale capabilities from cache.

**Solution**: Clear cache:
```bash
rm -rf ~/.capabilities-cache/
# Next /meta command will re-download
```

## Advanced Topics

### Custom Capability Sources

Create your own capability repository:

```bash
# 1. Create capabilities directory
mkdir -p my-capabilities

# 2. Add capability files
vim my-capabilities/meta-custom.md

# 3. Configure source
export META_CC_CAPABILITY_SOURCES="my-capabilities:yaleh/meta-cc@main/commands"
```

**Priority**: `my-capabilities` checked first (overrides default capabilities).

### Version Pinning

Pin to specific version for stability:

```bash
# Production: Use tagged version
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@v1.0.0/commands"

# Development: Use branch
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@develop/commands"
```

### Offline Operation

Download capabilities once, use offline:

```bash
# Download package
wget https://github.com/yaleh/meta-cc/releases/latest/download/capabilities-latest.tar.gz

# Configure local package
export META_CC_CAPABILITY_SOURCES="./capabilities-latest.tar.gz"

# Works offline (cached for 7 days)
```

## See Also

- [Capabilities Guide](../guides/capabilities.md) - Creating custom capabilities
- [Plugin Development](../guides/plugin-development.md) - Plugin workflow
- [MCP Guide](../guides/mcp.md) - MCP tool usage
- [Repository Structure](repository-structure.md) - Directory organization
