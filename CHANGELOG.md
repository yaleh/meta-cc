# Changelog

All notable changes to the meta-cc project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.33.0] - 2025-10-20


### Changed

- Refactoring: remove environment variable detection for session location (#4)
- Documentation: add Chinese BAIME intro page with English translations
- Documentation: add BAIME introductory tutorial page with overview and stats

### Fixed

- update script for consolidated plugin structure
- restore version field in golangci config
- remove obsolete plugin.json verification step

## [0.30.2] - 2025-10-19


## [0.30.1] - 2025-10-19


## [0.30.0] - 2025-10-19


### Changed

- Maintenance: release v0.30.0
- Maintenance: remove marketplace.json backup file

### Fixed

- update smoke tests for separate meta-cc-skills plugin

## [0.30.0] - 2025-10-19


### Changed

- Maintenance: remove marketplace.json backup file

## [Unreleased]

## [2.0.0] - 2025-10-25

### Added

- **jq-based Query Interface** (Phase 25): Complete refactoring with native jq integration
  - **QueryExecutor**: Streaming jq execution with file processing pipeline
  - **15 MCP Tools**: Unified 3-layer architecture
    - **Core Tools** (2): `query`, `query_raw` - Unified interface with jq filtering
    - **Convenience Tools** (10): `query_tool_errors`, `query_token_usage`, `query_conversation_flow`, `query_user_messages`, `query_tools`, `query_file_snapshots`, `query_timestamps`, `query_summaries`, `query_tool_blocks`, `query_system_errors`
    - **Utility Tools** (3): `cleanup_temp_files`, `list_capabilities`, `get_capability`
  - **Hybrid Output Mode**: Auto-switches between inline (<8KB) and file_ref (≥8KB)
  - **No Limits by Default**: Returns all results, relies on hybrid mode
  - **Standard Parameters**: Consistent interface across all tools (scope, jq_filter, stats_only, etc.)
  - See [MCP Query Tools Reference](docs/guides/mcp-query-tools.md)

- **Comprehensive Documentation**:
  - [MCP Query Tools Reference](docs/guides/mcp-query-tools.md) - Complete tool documentation (20 tools)
  - [MCP Query Cookbook](docs/examples/mcp-query-cookbook.md) - 25+ practical examples
  - [MCP v2.0 Migration Guide](docs/guides/mcp-v2-migration.md) - Upgrade from v1.x
  - [Unified Query API Guide](docs/guides/unified-query-api.md) - Query architecture
  - [Frequent JSONL Queries](docs/examples/frequent-jsonl-queries.md) - Updated with MCP tool mappings

- **Performance Benchmarks**:
  - QueryExecutor benchmarks for streaming execution
  - Expression compilation benchmarks
  - Memory usage profiling for different result sizes
  - Hybrid output mode benchmarks (inline vs file_ref)

- **Schema Standardization**: All output fields now use consistent snake_case
  - Matches JSONL source format 100%
  - ToolCall struct updated: `ToolName` → `tool_name`, `SessionID` → `session_id`, etc.
  - Complete field mapping documented in migration guide

### Changed

- **MCP Architecture**: Consolidated 16+ specialized tools into 20 tools with unified interface
  - Core `query` tool replaces 6 removed tools (query_context, query_tools_advanced, etc.)
  - 8 new convenience tools for common queries
  - 7 legacy tools remain for backward compatibility
- **Query Interface**: jq-based filtering replaces custom filter parameters
  - Maximum flexibility with native jq expressions
  - Composable queries with jq_filter and jq_transform
  - Standard parameters across all tools
- **Output Mode**: Hybrid output mode replaces fixed-size pagination
  - Automatic inline vs file_ref decision based on result size
  - Default threshold: 8KB (configurable via inline_threshold_bytes or META_CC_INLINE_THRESHOLD)
  - No need for explicit limit parameters
- **Documentation**: Updated all guides to reference v2.0 query interface

### Removed

- **5 Legacy MCP Tools**: Removed to simplify architecture and eliminate backward compatibility code
  - `query_tool_sequences` - Use `query` with jq filtering for tool sequences
  - `query_file_access` - Use `query_file_snapshots` or `query` with jq filtering
  - `get_session_stats` - Use `query_token_usage` with stats_only=true
  - `query_project_state` - Use `query` with custom jq expressions
  - `query_successful_prompts` - Use `query` with jq filtering for quality analysis
- **Backward Compatibility Code**: All adapter and compatibility layer code removed
  - OutputModeLegacy mode removed
  - Legacy parameter mapping removed
  - FilterMessages and FilterMessagesSummary functions removed

### Breaking Changes

- **Tool Removal**: 5 legacy tools removed (see Removed section)
  - Use `query` with jq filtering as replacement for removed tools
  - Use convenience tools for common patterns
  - Use `query_raw` for maximum flexibility
- **Field Names**: All output fields changed from PascalCase/camelCase to snake_case
  - Old: `{"ToolName": "Read", "SessionID": "abc"}`
  - New: `{"tool_name": "Read", "session_id": "abc"}`
- **Double JQ Application Bug Fixed**: All Phase 25 tools (12 tools) now correctly apply jq filters only once
  - Previously: Convenience tools applied jq twice, causing "expected an object but got: array" errors
  - Now: Executor recognizes Phase 25 tools and skips redundant jq application
- **Migration Required**: See [MCP v2.0 Migration Guide](docs/guides/mcp-v2-migration.md) for detailed instructions

### Migration

- **Upgrade Path**:
  1. Review [MCP v2.0 Migration Guide](docs/guides/mcp-v2-migration.md) for tool migration table
  2. Replace removed tools with `query` + jq filtering
  3. Use convenience tools for common patterns
  4. Test with hybrid output mode (no explicit limits needed)
  5. Optionally migrate legacy tools to new unified interface
- **Backward Compatibility**: 7 legacy tools remain functional during transition
- **Tool Migration Examples**: 25+ examples in [MCP Query Cookbook](docs/examples/mcp-query-cookbook.md)
- **Tool mapping table**: Direct equivalents for all 16 tools
- See [docs/guides/migration-to-unified-query.md](docs/guides/migration-to-unified-query.md)

### Performance

- **No degradation**: Unified query uses same optimized extraction logic
- **Potential improvements**: Avoids duplicate parsing for combined queries
- **Hybrid output mode**: Automatic handling of large results (inline ≤8KB, file_ref >8KB)

### Documentation

- 3 new comprehensive guides (~500 lines total)
- 10+ practical query examples in cookbook
- Updated all existing documentation to reference unified query
- Complete API reference with all parameters documented

---

## [0.33.0] - 2025-10-20

### Added
- **Skills Distribution** - 15 validated methodologies now packaged with plugin
  - Testing Strategy (3.1x speedup, 89% transferable)
  - CI/CD Optimization (2.5-3.5x speedup)
  - Error Recovery (95.4% error coverage)
  - Dependency Health (6x speedup)
  - Knowledge Transfer (3-8x ramp-up reduction)
  - Technical Debt Management (4.5x speedup)
  - Code Refactoring (28% complexity reduction)
  - Cross-Cutting Concerns (60-75% faster diagnosis)
  - Observability Instrumentation (23-46x speedup)
  - API Design (82.5% transferable)
  - Methodology Bootstrapping (10-50x speedup)
  - Agent Prompt Evolution (5x performance gap detection)
  - Baseline Quality Assessment (40-50% iteration reduction)
  - Rapid Convergence (40-60% time reduction)
  - Retrospective Validation (40-60% time reduction)
- **Agent Manifest Updates** - All 5 agents now declared in plugin.json
  - iteration-executor
  - iteration-prompt-designer
  - knowledge-extractor
  - project-planner (existing)
  - stage-executor (existing)

### Changed
- **Build Pipeline** - Updated to include skills directory in releases
  - scripts/sync-plugin-files.sh: Copy skills to dist/
  - Makefile: Support skills in sync-plugin-files target
  - .github/workflows/release.yml: Package skills in releases
- **Installation** - Enhanced to deploy skills to ~/.claude/skills/
- **Package Size** - Increased by ~1.5MB (10-11MB → 12MB, ~15% increase)

### Improved
- **User Experience** - Zero-friction access to validated methodologies
- **Documentation** - Added comprehensive skills overview to README.md
- **Testing** - Added 4 new smoke tests for skills and agents verification

## [0.26.8] - 2025-10-12

### Changed
- **Documentation Simplification**
  - README.md drastically simplified from 1909 lines → 275 lines (85% reduction)
  - New users can now understand the project in < 2 minutes (vs ~15 min before)
  - Better navigation with clear 3-tier documentation hierarchy

### Added
- **New Reference Documentation**
  - `docs/cli-reference.md` (506 lines) - Complete CLI command reference
  - `docs/jsonl-reference.md` (524 lines) - JSONL output format and jq patterns
  - `docs/features.md` (699 lines) - Advanced features overview

### Improved
- **Documentation Structure**
  - README: Quick start and overview (public-facing)
  - Reference docs: Complete technical documentation
  - CLAUDE.md: Development entry point (internal)
- **User Experience**
  - Advanced users find detailed docs easily via clear navigation
  - Developers have clear separation from public documentation
  - Comprehensive documentation index for all user types

## [0.26.7] - 2025-10-12

### Fixed
- **Version Synchronization**
  - Aligned plugin.json and marketplace.json versions with git tags
  - Improved release process to prevent version drift

## [0.26.6] - 2025-10-12

### Fixed
- **MCP Server Capability Loading**
  - Fixed "unknown source type: package" error when using package-based capability sources
  - Implemented `readPackageCapability` function with proper cache validation and fallback
  - Package sources now correctly download, extract, and cache `.tar.gz` capability packages
- **Session-Only Mode**
  - Fixed inverted condition in session locator that prevented environment variable detection
  - `--session-only` flag now correctly checks `CC_SESSION_ID` and `CC_PROJECT_HASH` environment variables
  - MCP tools with `scope: "session"` parameter now work correctly
- **Test Suite**
  - Updated `TestParseStatsCommand_JSON` and `TestParseStatsCommand_Markdown` to use `--session-only` flag
  - Tests now correctly reflect environment variable detection requirements

### Technical Details
- **Root Cause**: Session locator had inverted logic (`!opts.SessionOnly` instead of `opts.SessionOnly`)
- **Impact**: MCP commands with `scope: "session"` failed to locate session files
- **Files Changed**:
  - `cmd/mcp-server/capabilities.go` - Added package source handling
  - `internal/locator/locator.go:25` - Fixed session-only condition
  - `cmd/parse_test.go` - Updated tests for correct behavior

### Documentation
- Enhanced troubleshooting guide with MCP server issues section
- Added environment variables troubleshooting with design rationale
- Documented upgrade path from older versions

## [0.24.0] - 2025-10-11

### Fixed
- **Technical Debt Resolution**
  - Clean up test validation debt by removing false-positive bug references
  - Update Go dependencies to resolve infrastructure debt (5 packages updated)
  - Verify JSONL output format for query errors with improved test coverage

### Changed
- **Internal Improvements**
  - Enhanced MCP server builder reliability
  - Improved JSON output format validation
  - Refactored test assertions for better clarity

### Security
- Updated dependencies to latest stable versions for improved security

## [0.20.0] - 2025-10-10

### Added
- **Phase 21: Self-Hosted Marketplace**
  - Stage 21.1: Plugin marketplace configuration (.claude-plugin/marketplace.json)
  - Stage 21.2: Marketing documentation (docs/marketplace-listing.md)
  - Stage 21.3: Visual demonstration structure (docs/screenshots/)
  - Stage 21.4: Marketplace validation and testing
- Plugin marketplace configuration:
  - `.claude-plugin/marketplace.json` with rich metadata and component inventory
  - Comprehensive plugin description with feature highlights
  - GitHub Release asset references for all platforms
  - Installation command documentation (/plugin install yaleh/meta-cc)

## [Unreleased]

### Added

#### Phase 22: Unified Meta Command & Multi-Source Capability Discovery

**Phase 22.1-22.7: Core Multi-Source System**
- **Unified `/meta` command** with natural language intent matching
  - Semantic keyword scoring for capability selection
  - Natural language interface: `/meta "show errors"`, `/meta "quality check"`, etc.
  - Automatic capability discovery from configured sources
- **Multi-source capability loading system**
  - Local directory support (immediate reflection, no cache)
  - GitHub repository support (smart caching via jsDelivr CDN)
  - Priority-based merging (left = highest priority)
  - Environment variable configuration: `META_CC_CAPABILITY_SOURCES`
- **New MCP tools**:
  - `list_capabilities()` - Get capability index from all sources
  - `get_capability(name)` - Retrieve complete capability content
- **Frontmatter metadata** for all 13 existing slash commands
  - Structured metadata: name, description, keywords, category
  - Enables semantic capability discovery and matching
- **Capability development guide** (docs/capabilities-guide.md)
  - Local development workflow with real-time reflection
  - Publishing to GitHub repositories
  - Integration with MCP tools
  - Community contribution guidelines

**Phase 22.8-22.11: jsDelivr CDN Integration & Default Source Changes**
- **jsDelivr CDN Integration**: GitHub capabilities now load via jsDelivr CDN (cdn.jsdelivr.net)
  - Avoids GitHub raw API rate limiting
  - Improved performance and reliability with global CDN
  - Smart caching (1h branches, 7d tags)
- **Branch/Tag Specification**: Use `@` symbol to specify versions
  - Format: `owner/repo@branch/subdir`
  - Examples: `yaleh/meta-cc@v1.0.0/commands`, `yaleh/meta-cc@develop/commands`
  - Supports branches, tags, and commit hashes
- **Enhanced Error Handling**:
  - Exponential backoff retry for 5xx server errors (3 attempts: 1s, 2s, 4s)
  - Fallback to stale cache on network failure (up to 7 days)
  - Clear, actionable error messages for 404 and network errors

**Phase 22.8-22.10: Capability Package Distribution** (NEW)
- **Build tooling** for capability packages
  - `make bundle-capabilities` creates `.tar.gz` packages
  - GitHub Actions automatically uploads packages to releases
  - Package format: `commands/` and `agents/` directories
- **Package source support**:
  - MCP recognizes `.tar.gz` URLs and local paths
  - Automatic download and extraction
  - Format: `export META_CC_CAPABILITY_SOURCES="/path/to/caps.tar.gz"`
  - Mix package files with other source types
- **Package cache strategy**:
  - Release packages (`/releases/`): 7-day cache (immutable)
  - Custom packages: 1-hour cache (mutable)
  - Cache directory: `~/.capabilities-cache/packages/<hash>/`
  - Cache metadata: `.meta-cc-cache.json` tracks TTL and versions
  - Automatic cleanup of expired cache (>7 days)
- **Intelligent fallback**:
  - Primary: Package file (if configured)
  - Fallback 1: GitHub raw access
  - Fallback 2: Stale cache (up to 7 days old)
  - Clear error messages guide users to solutions
- **Benefits**:
  - **Offline-friendly**: Download once, cache locally
  - **Reliable**: No CDN dependencies, no rate limits
  - **Fast**: No network calls after initial download
  - **Simple**: Standard tar.gz format, easy to verify and debug

### Changed

**Phase 22.1-22.7**:
- Total MCP tools increased from 14 to 16 (added 2 capability discovery tools)
- Documentation updated with unified `/meta` command usage
- CLAUDE.md expanded with multi-source configuration guide
- README.md includes capability development quick start

**Phase 22.8-22.11**:
- **Default Source Changed**: Capabilities now load from `yaleh/meta-cc@main/commands` by default
  - Zero-configuration deployment for production users
  - Local development requires: `export META_CC_CAPABILITY_SOURCES="commands"`
- **Cache Strategy Enhanced**:
  - Branches: 1-hour cache (mutable, changes frequently)
  - Tags: 7-day cache (immutable, stable versions)
  - Local sources: No cache (always fresh)
- **GitHub URL Format**: raw.githubusercontent.com → cdn.jsdelivr.net/gh
- **Documentation Comprehensively Updated**:
  - CLAUDE.md: Added CDN and caching section, updated multi-source configuration
  - README.md: Added zero-configuration setup, version pinning, and branch specification
  - docs/capabilities-guide.md: Updated local development workflow, added branch testing section
  - CHANGELOG.md: Added breaking changes and migration guide

### Breaking Changes

⚠️ **Default Capability Source Changed**: If you were relying on the default local source behavior, you must now explicitly configure it.

**Impact**:
- **Production users**: No changes needed - automatic GitHub loading via CDN (zero-configuration)
- **Local developers**: Must set `export META_CC_CAPABILITY_SOURCES="commands"` for local development

**Old behavior (Phase 22.1-22.7)**:
```bash
# Capabilities loaded from local directory by default
# No explicit configuration needed for local development
```

**New behavior (Phase 22.8+)**:
```bash
# Default: Load from GitHub via jsDelivr CDN
# Explicit configuration required for local development
export META_CC_CAPABILITY_SOURCES="commands"
```

### Migration Guide

#### For Local Development

**Before Phase 22.8** (implicit local loading):
```bash
# No configuration needed
# Capabilities automatically loaded from commands/ directory
```

**After Phase 22.8** (explicit configuration required):
```bash
# REQUIRED for local development
export META_CC_CAPABILITY_SOURCES="commands"

# Or add to your shell profile
echo 'export META_CC_CAPABILITY_SOURCES="commands"' >> ~/.bashrc
source ~/.bashrc
```

#### For Production Deployment

**Before Phase 22.8** (manual configuration):
```bash
# Required environment variable
export META_CC_CAPABILITY_SOURCES=".claude/commands"
```

**After Phase 22.8** (zero-configuration):
```bash
# No configuration needed (uses GitHub default)
# Capabilities automatically loaded from yaleh/meta-cc@main/commands via jsDelivr CDN

# OR explicitly set for custom source
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@v1.0.0/commands"
```

#### Version Pinning (Recommended)

Pin to specific release for production stability:

```bash
# Pin to specific release tag (7-day cache, immutable)
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@v1.0.0/commands"

# Benefits:
# - 7-day cache (faster loading, reduced CDN requests)
# - Immutable (no unexpected changes)
# - Explicit version control
# - Predictable behavior
```

#### Testing Against Branches

Test capabilities from feature branches:

```bash
# Test from develop branch
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@develop/commands"

# Test from feature branch
export META_CC_CAPABILITY_SOURCES="yaleh/meta-cc@feature/new-capability/commands"

# Cache: 1 hour (changes propagate within 1 hour)
```

#### Multi-Source Configuration

Combine local and GitHub sources:

```bash
# Local has highest priority (left = highest)
export META_CC_CAPABILITY_SOURCES="~/dev/my-caps:commands:yaleh/meta-cc@main/commands"

# Benefits:
# - Local capabilities override GitHub versions
# - Test changes before submitting PR
# - Mix local development with production capabilities
```

### Technical Details

**jsDelivr CDN URLs**:
```
# Old format (Phase 22.1-22.7)
https://raw.githubusercontent.com/yaleh/meta-cc/main/commands/meta-errors.md

# New format (Phase 22.8+)
https://cdn.jsdelivr.net/gh/yaleh/meta-cc@main/commands/meta-errors.md
```

**Cache Behavior**:
- Branches (e.g., `@main`, `@develop`): 1-hour TTL
- Tags (e.g., `@v1.0.0`): 7-day TTL
- Local sources: No cache (immediate reflection)

**Network Resilience**:
- 5xx errors: Exponential backoff retry (1s, 2s, 4s)
- Network unreachable: Fallback to stale cache (up to 7 days)
- 404 errors: Clear error messages with troubleshooting steps

### Upgrade Checklist

For smooth migration to Phase 22.8+:

- [ ] **Local developers**: Set `export META_CC_CAPABILITY_SOURCES="commands"` in shell profile
- [ ] **Production users**: Review zero-configuration behavior (default GitHub loading)
- [ ] **CI/CD pipelines**: Update environment variable configuration if needed
- [ ] **Documentation**: Update team documentation with new default behavior
- [ ] **Version pinning**: Consider pinning to release tags for stability (`@v1.0.0`)
- [ ] **Testing**: Verify capability loading works in all environments

#### Phase 21: Self-Hosted Marketplace
  - Stage 21.1: Plugin marketplace configuration (.claude-plugin/marketplace.json)
  - Stage 21.2: Marketing documentation (docs/marketplace-listing.md)
  - Stage 21.3: Visual demonstration structure (docs/screenshots/)
  - Stage 21.4: Marketplace validation and testing
- Plugin marketplace configuration:
  - `.claude-plugin/marketplace.json` with rich metadata and component inventory
  - Comprehensive plugin description with feature highlights
  - GitHub Release asset references for all platforms
  - Installation command documentation (/plugin install yaleh/meta-cc)
- Marketing documentation:
  - `docs/marketplace-listing.md` with compelling feature showcase
  - Visual asset structure in `docs/screenshots/` directory
  - Installation badges for marketplace and GitHub releases
- Marketplace validation:
  - `tests/marketplace_validation_test.sh` for format and consistency checks
  - Version synchronization verification across plugin.json and marketplace.json
  - Documentation cross-reference validation

### Changed
- README.md now prioritizes marketplace installation as recommended method
- Installation guide expanded with /plugin install workflow
- Documentation references updated to include marketplace installation
- Component inventory documented (10 slash commands, 3 subagents, 14 MCP tools)

### Improved
- Plugin discoverability via Claude Code marketplace
- One-command installation experience
- Professional visual documentation structure
- Installation workflow streamlined

- **Phase 20: Plugin Packaging & Release**
  - Stage 20.1: Plugin structure definition with plugin.json manifest
  - Stage 20.2: Automated installation script with platform detection and MCP configuration merging
  - Stage 20.3: Multi-platform GitHub Release workflow enhancements
  - Stage 20.4: Comprehensive installation documentation and testing checklists
- Plugin packaging structure:
  - `plugin.json` with metadata, dependencies, and platform definitions
  - `.claude/lib/mcp-config.json` template for MCP server configuration
  - Enhanced `install.sh` with platform detection (Linux, macOS, Windows)
  - New `uninstall.sh` for clean component removal
- Multi-platform plugin packages (5 platforms):
  - linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
  - Each package includes binaries, Claude Code integration files, and install/uninstall scripts
- Installation documentation:
  - `docs/installation.md` with platform-specific instructions and troubleshooting
  - `tests/PLUGIN_VERIFICATION.md` for comprehensive pre-release testing
- GitHub Release workflow enhancements:
  - Automated plugin package creation for all platforms
  - Checksum generation and verification
  - Auto-generated release notes

### Changed
- **Slash Command Enhancements**: Integrated new message query capabilities
  - `/meta-timeline` now includes assistant response analysis and conversation latency metrics
  - `/meta-coach` now includes interaction quality analysis with response time, tool efficiency, and satisfaction metrics
- **Installation Process**: Enhanced with platform detection and safe MCP configuration merging
  - `install.sh` now detects OS and architecture automatically
  - MCP configuration merged safely without overwriting existing servers
  - Post-install verification ensures correct setup
- **Release Workflow**: Now creates plugin packages instead of bare binaries
  - Each release includes 5 platform-specific plugin packages
  - One-command installation: `curl -L <url> | tar xz && ./install.sh`
  - Checksums provided for integrity verification

## [v0.14.0] - 2025-10-09

### Added
- **Phase 19: Assistant Message and Conversation Query**
  - Stage 19.1: AssistantMessage serialization support
  - Stage 19.2: `query-assistant-messages` CLI command with regex search
  - Stage 19.3: `query-conversation` CLI command with role filter support
  - Stage 19.4: MCP integration with `query_assistant_messages` and `query_conversation` tools
  - Stage 19.5: Comprehensive documentation updates
- New MCP tools (14 total, up from 12):
  - `mcp__meta_cc__query_assistant_messages` - Search assistant responses with regex
  - `mcp__meta_cc__query_conversation` - Search conversation messages with optional role filter
- CLI commands:
  - `meta-cc query assistant-messages --pattern <regex>` - Search assistant messages
  - `meta-cc query conversation --pattern <regex> [--role user|assistant]` - Search conversation

### Changed
- Documentation updated to reflect 14 MCP tools (previously 12)
- All message query tools now use hybrid output mode for large results
- Enhanced conversation analysis capabilities with role-based filtering

### Technical Details
- Supports regex pattern matching across assistant responses
- Conversation query can filter by role (user, assistant, or both)
- Full hybrid output mode support with configurable thresholds
- Consistent parameter naming with existing message query tools

## [v0.13.0] - 2025-10-09

### Added
- **Bundled Release Artifacts**: Platform-specific bundles (.tar.gz) containing binaries, slash commands, subagents, and installation script
  - 5 platform bundles: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
  - Each bundle includes: meta-cc CLI, meta-cc-mcp server, 8 slash commands, 3 subagents
  - One-command installation: `curl -L <bundle-url> | tar xz && ./install.sh`
  - Total release artifacts: 16 files (10 binaries + 5 bundles + checksums)
- **Installation Script**: `scripts/install.sh` for automated setup of binaries and Claude Code integration files
- **Makefile Target**: `bundle-release` for creating platform bundles (requires VERSION=vX.Y.Z)
- **Slash Command**: `/meta-next` for workflow continuation suggestions

### Changed
- **GitHub Actions**: Updated to build Linux ARM64 MCP server binary and generate platform bundles
- **Documentation**: Updated installation guides with bundle installation instructions
- **Phase 16 Migration Guide**: Clarified `inline_threshold_bytes` vs `max_output_bytes` differences and migration strategy

### Fixed
- Improved Phase 16 migration documentation with clear comparison table and usage examples

## [v0.12.1] - 2025-10-08

### Changed
- **Meta-insight to Meta-cc Renaming**: Comprehensive refactor to update all references from the old `mcp_meta_insight` / `meta‑insight` namespace to the new `mcp_meta_cc` / `meta‑cc` across documentation, agents, commands, and tests. This ensures consistency with the renamed meta-cc component and prevents confusion from stale meta-insight references.

## [v0.12.0] - 2025-10-08

### Added
- Configured remote repository and release process.

## Release Process

To create a new release:

1. Update CHANGELOG.md with version and release notes
2. Run `./scripts/release.sh v1.0.0`
3. Monitor GitHub Actions for build completion
4. Verify binaries on GitHub Releases page

## Versioning Strategy

- **v0.x.x**: Beta releases (pre-1.0)
- **v1.0.0**: First stable release
- **v1.x.0**: Minor version (new features, backward compatible)
- **v1.0.x**: Patch version (bug fixes only)
- **v1.0.0-beta.1**: Pre-release tags

---

## [v0.11.1-formalization] - 2025-10-03

### Changed
- **Agent Formalization**: Replaced verbose natural language content with compact lambda calculus formal specifications
  - 5 agents formalized: `doc-updater`, `prompt-suggester`, `pattern-analyzer`, `prompt-refiner`, `meta-coach`
  - **92% overall size reduction** (3074 → 244 lines) while preserving **100% behavioral semantics**
  - Individual reductions:
    - `meta-coach.md`: 1092 → 47 lines (96% reduction)
    - `prompt-refiner.md`: 604 → 51 lines (92% reduction)
    - `pattern-analyzer.md`: 543 → 50 lines (91% reduction)
    - `prompt-suggester.md`: 441 → 50 lines (89% reduction)
    - `doc-updater.md`: 394 → 46 lines (88% reduction)
  - All frontmatter YAML headers preserved exactly
  - Zero regressions (all tests pass)

### Added
- Formal specification documentation in `plans/11/`:
  - `agent-formalization-inventory.md`: Stage 1 inventory analysis
  - `agent-formalization-design.md`: Stage 2 design with semantic mapping
  - `agent-formalization-replacement-prompt.md`: Refined replacement strategy
- `.claude/agents/FORMALIZATION_SUMMARY.md`: Complete formalization results and methodology

## [v0.11.0] - 2025-10-03

### Added
- **Phase 11: Unix Composability** - Comprehensive MCP server piping and CLI integration
  - Stage 11.1: Tool output format standardization (JSON Lines, CSV, JSON)
  - Stage 11.2: Unix pipeline integration (`meta-cc ... | jq`, `| grep`)
  - Stage 11.3: stderr/stdout separation for clean piping
  - Stage 11.4: Cookbook documentation with 15+ practical recipes
- `docs/cookbook.md`: Real-world usage patterns and workflows
- `--format` flag support across all commands (json, jsonl, csv, markdown, table)

### Changed
- All CLI commands now support structured output for pipeline composition
- MCP server responses now use standardized schemas
- Error messages redirected to stderr (data to stdout)

## [v0.10.0] - 2025-10-02

### Added
- **Phase 10: Advanced Query & Aggregation** - SQL-like filtering and time-series analysis
  - Stage 10.1: `query-tools-advanced` with WHERE clause support
  - Stage 10.2: Aggregation engine (`aggregate-stats` by tool/status/uuid)
  - Stage 10.3: Time-series analysis (`query-time-series` with hourly/daily/weekly buckets)
  - Stage 10.4: File-level statistics (`query-files` with operation counts and error rates)
- Advanced MCP tools:
  - `mcp__meta_cc__query_tools_advanced`
  - `mcp__meta_cc__aggregate_stats`
  - `mcp__meta_cc__query_time_series`
  - `mcp__meta_cc__query_files`

### Changed
- Enhanced query engine with SQL-like expression parsing
- Aggregation pipeline with GROUP BY and metric calculation

## [v0.9.0] - 2025-10-01

### Added
- **Phase 9: Context Query & Workflow Patterns**
  - Stage 9.1: Error context extraction (`query-context` with temporal windows)
  - Stage 9.2: Tool sequence detection (`query-tool-sequences` for workflow patterns)
  - Stage 9.3: File access history (`query-file-access` for read/edit/write tracking)
- MCP server tools:
  - `mcp__meta_cc__query_context`
  - `mcp__meta_cc__query_tool_sequences`
  - `mcp__meta_cc__query_file_access`

### Changed
- Enhanced pattern detection with sequence analysis
- File operation tracking across session history

## [v0.8.0] - 2025-09-30

### Added
- **Phase 8: Enhanced Query & Pagination**
  - Stage 8.1-8.3: Tool extraction with pagination (`extract-tools --limit`)
  - Stage 8.4-8.6: Advanced tool querying (`query-tools --tool=X --status=Y`)
  - Stage 8.7-8.9: Message search with regex (`query-user-messages --pattern`)
  - Stage 8.10-8.12: Project state and successful prompt analysis
- MCP server tools:
  - `mcp__meta_cc__extract_tools`
  - `mcp__meta_cc__query_tools`
  - `mcp__meta_cc__query_user_messages`
  - `mcp__meta_cc__query_project_state`
  - `mcp__meta_cc__query_successful_prompts`

### Changed
- All query commands now support pagination and filtering
- Enhanced Markdown output with section anchors

## [v0.7.0] - 2025-09-28

### Added
- **Phase 7: MCP Server Implementation**
  - Stage 7.1-7.3: MCP server foundation with stdio transport
  - Stage 7.4-7.6: Tool exposure (`get_session_stats`, `analyze_errors`, `extract_tools`)
  - Stage 7.7-7.9: Integration testing and error handling
- MCP server binary: `meta-cc-mcp`
- MCP configuration documentation in `docs/mcp-setup.md`

### Changed
- Refactored internal packages for MCP tool integration
- Enhanced JSON schema output for MCP compatibility

## [v0.6.0] - 2025-09-25

### Added
- **Phase 6: Slash Commands & Subagent Integration**
  - Stage 6.1-6.3: Slash command implementation (`/meta-stats`, `/meta-errors`, `/meta-timeline`)
  - Stage 6.4-6.6: `@meta-coach` subagent with conversational analysis
  - Stage 6.7-6.9: Integration testing and documentation
- Slash commands: `/meta-stats`, `/meta-errors`, `/meta-timeline`, `/meta-query-tools`
- Subagent: `@meta-coach` for guided workflow optimization

### Changed
- Enhanced CLI output formatting for slash command integration
- Subagent prompts with context-aware recommendations

## [v0.5.0] - 2025-09-20

### Added
- **Phase 5: Error Pattern Analysis**
  - Stage 5.1-5.3: Error signature extraction and clustering
  - Stage 5.4-5.6: Frequency analysis and recurrence detection
  - Stage 5.7-5.9: Root cause identification and recommendations
- `meta-cc analyze errors` command with pattern detection
- Error clustering by signature and temporal analysis

### Changed
- Enhanced error analysis with contextual tool usage correlation

## [v0.4.0] - 2025-09-15

### Added
- **Phase 4: Timeline Visualization**
  - Stage 4.1-4.3: Chronological tool execution timeline
  - Stage 4.4-4.6: Error event correlation and visualization
  - Stage 4.7-4.9: Markdown timeline rendering
- `meta-cc analyze timeline` command with temporal analysis
- ASCII timeline visualization with error markers

### Changed
- Timeline output includes tool durations and error context

## [v0.3.0] - 2025-09-10

### Added
- **Phase 3: Semantic Integration** (Optional)
  - Enhanced query capabilities with semantic filters
  - Cross-session pattern detection

### Changed
- Improved query performance with indexed lookups

## [v0.2.0] - 2025-09-05

### Added
- **Phase 2: Index Optimization** (Optional)
  - SQLite indexing for cross-session queries
  - Advanced query commands: `meta-cc query sessions`, `meta-cc query tools`
  - Session database with efficient lookups

### Changed
- Optimized large session file parsing with streaming

## [v0.1.0] - 2025-09-01

### Added
- **Phase 1: Core Parser (MVP)**
  - JSONL session history parser
  - Basic commands:
    - `meta-cc parse extract --session <uuid>`: Extract structured data
    - `meta-cc parse stats --session <uuid>`: Tool usage statistics
    - `meta-cc analyze errors --session <uuid>`: Error pattern detection
  - JSON, Markdown, and table output formats
  - Session auto-detection from current directory
  - Initial documentation and README

### Infrastructure
- Go project setup with Cobra CLI framework
- Unit test foundation with 70%+ coverage
- GitHub repository initialization

---

**Legend**:
- `[vX.Y.Z]` - Release version
- `Added` - New features
- `Changed` - Changes to existing functionality
- `Deprecated` - Soon-to-be removed features
- `Removed` - Removed features
- `Fixed` - Bug fixes
- `Security` - Security improvements
