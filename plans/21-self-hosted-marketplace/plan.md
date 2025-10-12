# Phase 21: Self-Hosted Marketplace - TDD Implementation Plan

## Phase Overview

**Objective**: Create Claude Code plugin marketplace configuration to enable one-click installation via `/plugin install yaleh/meta-cc`.

**Code Volume**: ~200 lines | **Priority**: High | **Status**: Planning

**Dependencies**:
- Phase 20 (Plugin Packaging & Release - Complete)
- Existing GitHub Release workflow with plugin packages
- Existing plugin.json with complete metadata

**Deliverables**:
- `.claude-plugin/marketplace.json` marketplace listing configuration
- Updated documentation with marketplace installation instructions
- Marketing materials and screenshots demonstrating plugin capabilities
- Marketplace listing validation and testing

---

## Phase Objectives

### Core Problems

**Problem 1: Manual Download Friction**
- Current: Users download GitHub Release packages manually
- Impact: Multi-step process reduces discoverability and adoption
- Need: One-command installation via Claude Code plugin marketplace

**Problem 2: Marketplace Discoverability**
- Current: No marketplace listing, users must find GitHub repository
- Missing: Plugin marketplace entry with rich metadata and screenshots
- Need: Searchable marketplace listing with compelling description

**Problem 3: Installation Instructions Complexity**
- Current: Platform-specific download commands in README
- Missing: Unified marketplace installation method
- Need: Single command (`/plugin install meta-cc`) for all platforms

**Problem 4: Visual Demonstration Gap**
- Current: Text-only documentation
- Missing: Screenshots and GIFs demonstrating plugin capabilities
- Need: Visual assets showcasing key features (meta-coach, meta-viz, etc.)

### Solution Architecture

```
Phase 21 Implementation Strategy:

1. Marketplace Metadata (Stage 21.1)
   - Create .claude-plugin/marketplace.json with complete metadata
   - Configure GitHub Release asset references
   - Define plugin categories and tags
   - Document component inventory (slash commands, subagents, MCP tools)

2. Installation Documentation (Stage 21.2)
   - Update README.md with marketplace installation as primary method
   - Create docs/marketplace-listing.md with marketing copy
   - Add installation badges and visual indicators
   - Document /plugin install workflow

3. Visual Demonstrations (Stage 21.3)
   - Create installation demo (GIF or short video)
   - Capture feature screenshots (meta-coach analysis, meta-viz dashboard)
   - Organize assets in docs/screenshots/ directory
   - Update documentation to reference visual materials

4. Validation & Testing (Stage 21.4)
   - Validate marketplace.json format against Claude Code schema
   - Test /plugin marketplace add yaleh/meta-cc command
   - Verify /plugin install meta-cc workflow
   - Update CHANGELOG with Phase 21 changes
```

### Design Principles

1. **GitHub Release Integration**: Marketplace points to existing GitHub Release assets (no duplication)
2. **Rich Metadata**: Comprehensive plugin description with feature highlights
3. **Visual First**: Screenshots and GIFs demonstrate capabilities immediately
4. **Backward Compatible**: Existing installation methods continue to work
5. **Self-Hosted**: Repository-based marketplace (no external hosting)

---

## Success Criteria

**Functional Acceptance**:
- ✅ `.claude-plugin/marketplace.json` validates against Claude Code schema
- ✅ `/plugin marketplace add yaleh/meta-cc` command succeeds
- ✅ `/plugin install meta-cc` installs plugin successfully
- ✅ Marketplace listing displays plugin metadata correctly
- ✅ Screenshots and GIFs demonstrate key features

**Integration Acceptance**:
- ✅ GitHub Release assets referenced correctly
- ✅ Plugin version syncs with marketplace metadata
- ✅ Installation verification passes
- ✅ README.md prioritizes marketplace installation

**Code Quality**:
- ✅ Total code: ~200 lines (within Phase 21 budget)
  - Stage 21.1: ~80 lines (marketplace metadata)
  - Stage 21.2: ~60 lines (documentation)
  - Stage 21.3: ~30 lines (asset organization, minimal code)
  - Stage 21.4: ~30 lines (validation, changelog)
- ✅ Each stage ≤ 200 lines
- ✅ All tests pass (format validation)

---

## Stage 21.1: Marketplace Metadata

### Objective

Create `.claude-plugin/marketplace.json` with complete plugin metadata, GitHub Release references, and component inventory for marketplace listing.

### Acceptance Criteria

- [ ] `.claude-plugin/marketplace.json` created with complete metadata
- [ ] Plugin name, description, author, license specified
- [ ] GitHub Release asset references configured correctly
- [ ] Component inventory includes counts (10 slash commands, 3 subagents, 14 MCP tools)
- [ ] Version number syncs with plugin.json (0.13.0)
- [ ] Categories and tags defined for discoverability
- [ ] Installation instructions reference /plugin install command
- [ ] JSON format validates against schema

### TDD Approach

**Test File**: `tests/marketplace_metadata_test.sh` (~40 lines)

```bash
#!/bin/bash
# Test functions:
# - test_marketplace_json_exists - Verify .claude-plugin/marketplace.json exists
# - test_marketplace_json_valid - Validate JSON syntax
# - test_required_fields - Verify all required fields present
# - test_version_sync - Ensure version matches plugin.json
# - test_github_release_urls - Validate asset URL format
# - test_component_counts - Verify component inventory accuracy
```

**Test Strategy**:
1. Verify `.claude-plugin/marketplace.json` exists and is valid JSON
2. Check all required fields (name, version, description, repository, assets)
3. Validate version synchronization with plugin.json (0.13.0)
4. Verify GitHub Release asset URLs are well-formed
5. Validate component counts (10 commands, 3 subagents, 14 tools)
6. Test category and tag structure

**Implementation Files**:

1. `.claude-plugin/marketplace.json` (~70 lines)

```json
{
  "name": "meta-cc",
  "displayName": "Meta-CC: Workflow Analysis for Claude Code",
  "version": "0.13.0",
  "description": "Meta-Cognition tool for analyzing Claude Code session history. Provides workflow optimization insights, error pattern detection, and productivity analytics through 10 slash commands, 3 subagents, and 14 MCP tools.",
  "longDescription": "meta-cc transforms Claude Code session logs into actionable workflow insights. Analyze tool usage patterns, detect error repetitions, track file access history, and optimize your development workflow with comprehensive metacognitive analysis.\n\nKey Features:\n- 📊 Session analytics with detailed statistics\n- 🔍 Error pattern detection and root cause analysis\n- 🎯 Workflow optimization recommendations\n- 📈 Visual dashboards with ASCII charts\n- 🤖 AI-powered coaching with @meta-coach subagent\n- 🔗 Seamless MCP integration (14 query tools)\n- 💡 Prompt optimization based on successful patterns\n\nPerfect for developers seeking to understand and improve their Claude Code workflows.",
  "author": {
    "name": "Yale Huang",
    "email": "yaleh@ieee.org",
    "url": "https://github.com/yaleh"
  },
  "license": "MIT",
  "homepage": "https://github.com/yaleh/meta-cc",
  "repository": {
    "type": "git",
    "url": "https://github.com/yaleh/meta-cc"
  },
  "documentation": "https://github.com/yaleh/meta-cc/blob/develop/README.md",
  "keywords": [
    "workflow-analysis",
    "session-history",
    "productivity",
    "metacognition",
    "analytics",
    "optimization",
    "error-detection",
    "coaching"
  ],
  "categories": [
    "Development Tools",
    "Analytics",
    "Productivity"
  ],
  "screenshots": [
    "docs/screenshots/meta-coach-analysis.png",
    "docs/screenshots/meta-viz-dashboard.png",
    "docs/screenshots/installation-demo.gif"
  ],
  "components": {
    "slash_commands": 10,
    "subagents": 3,
    "mcp_tools": 14
  },
  "platforms": [
    "linux-amd64",
    "linux-arm64",
    "darwin-amd64",
    "darwin-arm64",
    "windows-amd64"
  ],
  "assets": {
    "type": "github-release",
    "repository": "yaleh/meta-cc",
    "pattern": "meta-cc-plugin-v{version}-{platform}.tar.gz",
    "latest": "https://github.com/yaleh/meta-cc/releases/latest"
  },
  "installation": {
    "command": "/plugin install yaleh/meta-cc",
    "requirements": {
      "claude-code": ">=1.0.0"
    }
  },
  "links": {
    "issues": "https://github.com/yaleh/meta-cc/issues",
    "changelog": "https://github.com/yaleh/meta-cc/blob/develop/CHANGELOG.md",
    "documentation": "https://github.com/yaleh/meta-cc/blob/develop/docs/",
    "examples": "https://github.com/yaleh/meta-cc/blob/develop/docs/examples-usage.md"
  }
}
```

2. `tests/marketplace_metadata_test.sh` (~40 lines)

```bash
#!/bin/bash
# meta-cc marketplace metadata validation tests

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

pass() {
    echo -e "${GREEN}✓${NC} $1"
}

fail() {
    echo -e "${RED}✗${NC} $1"
    exit 1
}

# Test 1: marketplace.json exists
test_marketplace_json_exists() {
    if [ -f .claude-plugin/marketplace.json ]; then
        pass "marketplace.json exists"
    else
        fail "marketplace.json not found"
    fi
}

# Test 2: Valid JSON syntax
test_marketplace_json_valid() {
    if jq empty .claude-plugin/marketplace.json 2>/dev/null; then
        pass "marketplace.json is valid JSON"
    else
        fail "marketplace.json has invalid JSON syntax"
    fi
}

# Test 3: Required fields present
test_required_fields() {
    REQUIRED_FIELDS="name version description repository assets installation"
    for field in $REQUIRED_FIELDS; do
        if jq -e ".$field" .claude-plugin/marketplace.json >/dev/null 2>&1; then
            pass "Required field present: $field"
        else
            fail "Missing required field: $field"
        fi
    done
}

# Test 4: Version sync with plugin.json
test_version_sync() {
    MARKETPLACE_VERSION=$(jq -r '.version' .claude-plugin/marketplace.json)
    PLUGIN_VERSION=$(jq -r '.version' plugin.json)

    if [ "$MARKETPLACE_VERSION" = "$PLUGIN_VERSION" ]; then
        pass "Version synchronized: $MARKETPLACE_VERSION"
    else
        fail "Version mismatch: marketplace=$MARKETPLACE_VERSION, plugin=$PLUGIN_VERSION"
    fi
}

# Test 5: Component counts accuracy
test_component_counts() {
    SLASH_COMMANDS=$(jq -r '.components.slash_commands' .claude-plugin/marketplace.json)
    SUBAGENTS=$(jq -r '.components.subagents' .claude-plugin/marketplace.json)
    MCP_TOOLS=$(jq -r '.components.mcp_tools' .claude-plugin/marketplace.json)

    if [ "$SLASH_COMMANDS" = "10" ] && [ "$SUBAGENTS" = "3" ] && [ "$MCP_TOOLS" = "14" ]; then
        pass "Component counts correct: $SLASH_COMMANDS commands, $SUBAGENTS subagents, $MCP_TOOLS tools"
    else
        fail "Component counts incorrect: got $SLASH_COMMANDS/$SUBAGENTS/$MCP_TOOLS, expected 10/3/14"
    fi
}

# Run all tests
echo "Running marketplace metadata validation tests..."
test_marketplace_json_exists
test_marketplace_json_valid
test_required_fields
test_version_sync
test_component_counts

echo ""
echo -e "${GREEN}All marketplace metadata tests passed!${NC}"
```

### File Changes

**New Files**:
- `.claude-plugin/marketplace.json` (+70 lines)
- `tests/marketplace_metadata_test.sh` (+40 lines)

**Total**: ~110 lines (exceeds 80-line target by 30 lines, acceptable for comprehensive metadata)

### Test Commands

```bash
# Run Stage 21.1 tests
bash tests/marketplace_metadata_test.sh

# Validate JSON syntax
jq empty .claude-plugin/marketplace.json

# Verify version sync
diff <(jq -r '.version' plugin.json) <(jq -r '.version' .claude-plugin/marketplace.json)

# Check required fields
jq -e '.name, .version, .description, .repository, .assets' .claude-plugin/marketplace.json

# Validate component counts
jq '.components' .claude-plugin/marketplace.json
```

### Testing Protocol

**After Implementation**:
1. Run `bash tests/marketplace_metadata_test.sh`
2. Validate JSON syntax with `jq`
3. Verify version synchronization
4. Check GitHub Release asset URL format
5. Validate component inventory accuracy
6. **HALT if validation fails after 2 fix attempts**

### Dependencies

None (foundation stage)

### Estimated Time

1 hour (110 lines implementation + tests)

---

## Stage 21.2: Installation Documentation Updates

### Objective

Update README.md and create marketplace-listing.md with marketplace installation instructions as the primary installation method.

### Acceptance Criteria

- [ ] README.md updated with marketplace installation as recommended method
- [ ] `docs/marketplace-listing.md` created with marketing copy
- [ ] Installation badges added to README (marketplace, GitHub release)
- [ ] `/plugin install` workflow documented clearly
- [ ] Existing installation methods preserved (backward compatibility)
- [ ] Cross-references between docs verified
- [ ] Links to screenshots included

### Documentation Changes

**1. README.md Updates** (~30 lines changed/added)

```markdown
## Installation

### Marketplace Installation (Recommended)

Install meta-cc directly from the Claude Code plugin marketplace:

```
/plugin marketplace add yaleh/meta-cc
/plugin install meta-cc
```

That's it! The plugin will be installed with all components (CLI, MCP server, slash commands, and subagents).

**What's included:**
- ✅ `meta-cc` CLI tool
- ✅ `meta-cc-mcp` MCP server (auto-configured)
- ✅ 10 slash commands (`.claude/commands/`)
- ✅ 3 subagents (`.claude/agents/`)
- ✅ 14 MCP query tools (auto-available)

### Alternative: Plugin Package Installation

For manual installation or if marketplace is unavailable, download platform-specific packages:

[Existing plugin package installation instructions...]

### Alternative: Individual Binaries

[Existing individual binary installation instructions...]
```

**2. docs/marketplace-listing.md** (~120 lines, new)

```markdown
# meta-cc - Workflow Analysis for Claude Code

![License](https://img.shields.io/github/license/yaleh/meta-cc)
![Version](https://img.shields.io/github/v/release/yaleh/meta-cc)
![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-blue)

**Transform your Claude Code session logs into actionable workflow insights.**

## What is meta-cc?

meta-cc is a metacognition tool that analyzes your Claude Code session history to help you understand and optimize your development workflow. By parsing session logs, detecting patterns, and providing AI-powered recommendations, meta-cc turns raw data into productivity insights.

## Key Features

### 📊 Comprehensive Analytics
- **Session Statistics**: Detailed metrics on tool usage, errors, and workflow patterns
- **Error Detection**: Identify repetitive errors and anti-patterns
- **File Access Tracking**: Understand which files are accessed most frequently
- **Time Series Analysis**: Track productivity metrics over time (hourly/daily/weekly)

### 🎯 Workflow Optimization
- **Pattern Recognition**: Detect common tool sequences and workflow bottlenecks
- **Prompt Analysis**: Learn from your most successful prompts
- **Quality Scoring**: Assess response quality and iteration efficiency
- **Habit Insights**: Discover productivity patterns and areas for improvement

### 🤖 AI-Powered Coaching
- **@meta-coach**: Interactive subagent providing personalized workflow recommendations
- **Context-Aware**: Analyzes your specific project history for tailored advice
- **Multi-Turn Conversations**: Deep dive into specific workflow aspects

### 📈 Visual Dashboards
- **ASCII Charts**: Terminal-friendly visualizations of metrics
- **Timeline Views**: Project evolution over time
- **Focus Analysis**: Attention distribution across files and tasks

## Components

### 10 Slash Commands
- `/meta-stats` - Quick session statistics
- `/meta-errors` - Error pattern analysis
- `/meta-timeline` - Project evolution timeline
- `/meta-viz` - Visual analytics dashboard
- `/meta-habits` - Productivity habit insights
- `/meta-quality-scan` - Quality assessment with scorecard
- `/meta-focus-analyzer` - Attention pattern analysis
- `/meta-guide` - Intelligent guidance and recommendations
- `/meta-next` - Generate next-step prompts
- `/meta-prompt` - Refine prompts using historical patterns

### 3 Subagents
- **@meta-coach** - Comprehensive workflow analysis and coaching
- **@meta-query** - Complex query orchestration with Unix pipelines
- **@project-planner** - Project planning assistance

### 14 MCP Query Tools
Programmatic access to session data for autonomous analysis:
- `get_session_stats` - Session-level metrics
- `query_tools` - Tool call filtering
- `query_user_messages` - User input pattern analysis
- `query_assistant_messages` - Response quality assessment
- `query_conversation` - Full turn analysis
- `query_files` - File operation statistics
- `query_context` - Error context extraction
- `query_tool_sequences` - Workflow pattern detection
- `query_file_access` - File access history
- `query_project_state` - Project evolution tracking
- `query_successful_prompts` - High-quality prompt patterns
- `query_tools_advanced` - SQL-like query expressions
- `query_time_series` - Time-based metric analysis
- `cleanup_temp_files` - Temporary file management

## Installation

```
/plugin marketplace add yaleh/meta-cc
/plugin install meta-cc
```

Restart Claude Code to activate all components.

## Quick Start

After installation, try these commands:

1. **Get session overview**:
   ```
   /meta-stats
   ```

2. **Analyze errors**:
   ```
   /meta-errors
   ```

3. **Interactive coaching**:
   ```
   @meta-coach analyze my workflow
   ```

4. **Visual dashboard**:
   ```
   /meta-viz
   ```

## Use Cases

### For Solo Developers
- Understand your coding patterns and improve efficiency
- Identify repetitive errors and learn from mistakes
- Track productivity trends over time

### For Team Leads
- Analyze team workflow patterns
- Identify common pain points across projects
- Share best practices based on successful patterns

### For Learning
- Review your Claude Code learning journey
- Track skill progression over time
- Optimize your prompting strategies

## Platform Support

- **Linux**: x86_64, ARM64
- **macOS**: Intel, Apple Silicon
- **Windows**: x86_64 (via Git Bash)

## Documentation

- [Complete Documentation](https://github.com/yaleh/meta-cc/blob/develop/docs/)
- [Installation Guide](https://github.com/yaleh/meta-cc/blob/develop/docs/installation.md)
- [Examples & Usage](https://github.com/yaleh/meta-cc/blob/develop/docs/examples-usage.md)
- [Troubleshooting](https://github.com/yaleh/meta-cc/blob/develop/docs/troubleshooting.md)

## Links

- [GitHub Repository](https://github.com/yaleh/meta-cc)
- [Issue Tracker](https://github.com/yaleh/meta-cc/issues)
- [Changelog](https://github.com/yaleh/meta-cc/blob/develop/CHANGELOG.md)

## License

MIT License - see [LICENSE](https://github.com/yaleh/meta-cc/blob/develop/LICENSE) for details.

## Author

Yale Huang ([@yaleh](https://github.com/yaleh))
```

**3. README.md Badge Updates** (~10 lines)

Add marketplace installation badge at the top:

```markdown
[![Plugin Marketplace](https://img.shields.io/badge/Claude_Code-Plugin_Marketplace-blue)](https://github.com/yaleh/meta-cc)
```

### File Changes

**Modified Files**:
- `README.md` (~30 lines changed)

**New Files**:
- `docs/marketplace-listing.md` (+120 lines)

**Total**: ~150 lines (exceeds 60-line target by 90 lines due to comprehensive marketing copy)

### Verification Commands

```bash
# Verify README installation section updated
grep -A5 "Marketplace Installation" README.md

# Check marketplace-listing.md exists and is complete
test -f docs/marketplace-listing.md && wc -l docs/marketplace-listing.md

# Verify badge added
grep "Plugin_Marketplace" README.md

# Check cross-references
grep -r '\[.*\](docs/marketplace-listing.md)' README.md CLAUDE.md
```

### Testing Protocol

**After Implementation**:
1. Review README.md for clarity and accuracy
2. Verify marketplace-listing.md reads well (marketing perspective)
3. Test all links in marketplace-listing.md
4. Verify installation command syntax
5. Check cross-references between documents
6. **HALT if documentation quality is poor after 2 revisions**

### Dependencies

- Stage 21.1 (marketplace.json)

### Estimated Time

1 hour (150 lines documentation)

---

## Stage 21.3: Visual Demonstration Assets

### Objective

Create screenshots and GIFs demonstrating plugin installation and key features (meta-coach, meta-viz) for marketplace listing.

### Acceptance Criteria

- [ ] Installation demo GIF created (showing `/plugin install` workflow)
- [ ] meta-coach feature screenshot captured (showing analysis output)
- [ ] meta-viz dashboard screenshot captured (showing ASCII charts)
- [ ] Assets organized in `docs/screenshots/` directory
- [ ] README.md and marketplace-listing.md reference screenshots
- [ ] Image file sizes optimized (<500KB each)
- [ ] Asset list documented

### Visual Asset Creation

**Required Assets**:

1. **installation-demo.gif** (~5-10 seconds)
   - Show: `/plugin marketplace add yaleh/meta-cc` command
   - Show: `/plugin install meta-cc` command output
   - Show: Installation success message

2. **meta-coach-analysis.png**
   - Show: `@meta-coach analyze my workflow` interaction
   - Capture: Comprehensive analysis output with recommendations

3. **meta-viz-dashboard.png**
   - Show: `/meta-viz` command output
   - Capture: ASCII charts and visual dashboard

**Asset Organization**:

```
docs/screenshots/
├── installation-demo.gif       # Installation workflow demo
├── meta-coach-analysis.png     # @meta-coach feature demo
├── meta-viz-dashboard.png      # /meta-viz dashboard demo
└── README.md                   # Asset inventory
```

**Implementation File**: `docs/screenshots/README.md` (~30 lines)

```markdown
# meta-cc Screenshots and Demos

This directory contains visual assets for documentation and marketplace listing.

## Assets

### Installation Demo
- **File**: `installation-demo.gif`
- **Description**: Demonstrates `/plugin install` workflow from start to finish
- **Duration**: ~5-10 seconds
- **Size**: <500KB

### Feature Demonstrations

#### meta-coach Subagent
- **File**: `meta-coach-analysis.png`
- **Description**: Shows @meta-coach providing workflow analysis and recommendations
- **Dimensions**: 1200x800 (approximate)
- **Size**: <400KB

#### meta-viz Dashboard
- **File**: `meta-viz-dashboard.png`
- **Description**: Shows ASCII charts and visual analytics from /meta-viz command
- **Dimensions**: 1200x800 (approximate)
- **Size**: <400KB

## Usage

Referenced in:
- `README.md` (installation section)
- `docs/marketplace-listing.md` (feature showcase)
- `.claude-plugin/marketplace.json` (screenshots array)

## Creating Screenshots

### Installation Demo (GIF)
```bash
# Using asciinema + agg (for GIF conversion)
asciinema rec installation-demo.cast
agg installation-demo.cast installation-demo.gif
```

### Feature Screenshots (PNG)
1. Run command in Claude Code (`@meta-coach` or `/meta-viz`)
2. Capture terminal output (screenshot tool or Claude Code export)
3. Optimize image size (imagemagick, pngquant, etc.)
4. Save to `docs/screenshots/`

## Optimization

```bash
# Optimize PNG files
pngquant --quality=65-80 *.png

# Optimize GIF files (if needed)
gifsicle -O3 --colors 256 installation-demo.gif -o installation-demo-opt.gif
```
```

### File Changes

**New Files**:
- `docs/screenshots/README.md` (+30 lines)
- `docs/screenshots/installation-demo.gif` (binary asset, ~300-500KB)
- `docs/screenshots/meta-coach-analysis.png` (binary asset, ~300-400KB)
- `docs/screenshots/meta-viz-dashboard.png` (binary asset, ~300-400KB)

**Modified Files**:
- `README.md` (~10 lines added to reference screenshots)
- `docs/marketplace-listing.md` (~10 lines added to reference screenshots)

**Total Code**: ~50 lines (README.md only, binary assets don't count toward code limit)

### Asset Creation Commands

```bash
# Create screenshots directory
mkdir -p docs/screenshots

# Create placeholder README
cat > docs/screenshots/README.md << 'EOF'
[README content as above]
EOF

# TODO: Manual asset creation
# 1. Record installation demo with asciinema/terminal recorder
# 2. Capture meta-coach screenshot in Claude Code
# 3. Capture meta-viz screenshot in Claude Code
# 4. Optimize file sizes
# 5. Commit to repository
```

### Testing Protocol

**After Implementation**:
1. Verify all assets created and properly sized (<500KB each)
2. Check image quality (readable text, clear visuals)
3. Verify references in README.md and marketplace-listing.md
4. Test that marketplace.json references screenshot paths correctly
5. Validate images load correctly in GitHub preview
6. **HALT if asset quality is poor after 2 iterations**

### Dependencies

- Stage 21.2 (documentation updates)

### Estimated Time

1.5 hours (asset creation + organization + documentation)

---

## Stage 21.4: Validation and Testing

### Objective

Validate marketplace.json format, test marketplace installation commands, and update CHANGELOG with Phase 21 completion.

### Acceptance Criteria

- [ ] marketplace.json format validated against Claude Code schema
- [ ] `/plugin marketplace add yaleh/meta-cc` command tested
- [ ] `/plugin install meta-cc` installation workflow verified
- [ ] Version numbers consistent across all files
- [ ] CHANGELOG.md updated with Phase 21 changes
- [ ] All documentation cross-references verified
- [ ] Phase 21 completion documented in docs/plan.md

### Validation and Testing

**Test File**: `tests/marketplace_validation_test.sh` (~50 lines)

```bash
#!/bin/bash
# meta-cc marketplace installation validation tests

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass() {
    echo -e "${GREEN}✓${NC} $1"
}

fail() {
    echo -e "${RED}✗${NC} $1"
    exit 1
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Test 1: marketplace.json format validation
test_marketplace_format() {
    echo "Testing marketplace.json format..."

    # Validate JSON syntax
    if ! jq empty .claude-plugin/marketplace.json 2>/dev/null; then
        fail "marketplace.json has invalid JSON syntax"
    fi

    # Check required fields
    REQUIRED="name version description repository assets installation"
    for field in $REQUIRED; do
        if ! jq -e ".$field" .claude-plugin/marketplace.json >/dev/null 2>&1; then
            fail "Missing required field: $field"
        fi
    done

    pass "marketplace.json format valid"
}

# Test 2: Version consistency
test_version_consistency() {
    echo "Testing version consistency..."

    MARKETPLACE_VERSION=$(jq -r '.version' .claude-plugin/marketplace.json)
    PLUGIN_VERSION=$(jq -r '.version' plugin.json)

    if [ "$MARKETPLACE_VERSION" != "$PLUGIN_VERSION" ]; then
        fail "Version mismatch: marketplace=$MARKETPLACE_VERSION, plugin=$PLUGIN_VERSION"
    fi

    pass "Version consistent: $MARKETPLACE_VERSION"
}

# Test 3: Asset references
test_asset_references() {
    echo "Testing asset references..."

    # Check that screenshots exist
    SCREENSHOTS=$(jq -r '.screenshots[]' .claude-plugin/marketplace.json 2>/dev/null)
    for screenshot in $SCREENSHOTS; do
        if [ ! -f "$screenshot" ]; then
            warn "Screenshot not found: $screenshot (may need to be created)"
        else
            pass "Screenshot found: $screenshot"
        fi
    done
}

# Test 4: Documentation cross-references
test_documentation_links() {
    echo "Testing documentation cross-references..."

    # Check that marketplace-listing.md exists
    if [ ! -f docs/marketplace-listing.md ]; then
        fail "docs/marketplace-listing.md not found"
    fi

    # Check README references marketplace
    if ! grep -q "Marketplace Installation" README.md; then
        fail "README.md does not reference marketplace installation"
    fi

    pass "Documentation cross-references valid"
}

# Test 5: CHANGELOG updated
test_changelog_updated() {
    echo "Testing CHANGELOG update..."

    if ! grep -q "Phase 21" CHANGELOG.md; then
        warn "CHANGELOG.md may not include Phase 21 changes"
    else
        pass "CHANGELOG.md includes Phase 21 changes"
    fi
}

# Run all tests
echo "Running marketplace validation tests..."
echo ""

test_marketplace_format
test_version_consistency
test_asset_references
test_documentation_links
test_changelog_updated

echo ""
echo -e "${GREEN}Marketplace validation complete!${NC}"
echo ""
echo "Manual verification steps:"
echo "1. Test: /plugin marketplace add yaleh/meta-cc"
echo "2. Test: /plugin install meta-cc"
echo "3. Verify installation completes successfully"
echo "4. Test slash commands and subagents work"
```

**CHANGELOG.md Updates** (~25 lines)

```markdown
## [0.14.0] - 2025-10-11

### Added (Phase 21)
- Plugin marketplace configuration (.claude-plugin/marketplace.json)
- Marketplace listing with rich metadata and component inventory
- Marketing documentation (docs/marketplace-listing.md)
- Visual demonstration assets (installation demo, feature screenshots)
- Marketplace installation as recommended method in README
- Installation badges for marketplace and GitHub releases

### Changed
- README.md now prioritizes marketplace installation
- Documentation references updated to include marketplace workflow
- Installation guide expanded with /plugin install instructions

### Improved
- Plugin discoverability via Claude Code marketplace
- One-command installation experience
- Visual documentation with screenshots and GIFs

### Fixed
- None (documentation-only phase)
```

**docs/plan.md Updates** (~10 lines)

Update Phase 21 status from "待开始" to "✅ 已完成":

```markdown
## Phase 21: 自托管插件市场（Self-Hosted Marketplace）

**目标**：创建 Claude Code 插件市场配置，支持一键安装
**代码量**：~200 行 | **优先级**：高 | **状态**：✅ 已完成

[Keep existing stage descriptions...]

**工作量**：~4h | ~200 lines

详细计划见 `plans/21/plan.md`
```

### File Changes

**New Files**:
- `tests/marketplace_validation_test.sh` (+50 lines)

**Modified Files**:
- `CHANGELOG.md` (+25 lines)
- `docs/plan.md` (+10 lines changed)

**Total**: ~85 lines (exceeds 30-line target by 55 lines due to comprehensive validation)

### Test Commands

```bash
# Run marketplace validation tests
bash tests/marketplace_validation_test.sh

# Validate JSON format
jq empty .claude-plugin/marketplace.json

# Check version consistency
diff <(jq -r '.version' plugin.json) <(jq -r '.version' .claude-plugin/marketplace.json)

# Verify documentation
grep -q "Marketplace Installation" README.md && echo "README updated"
test -f docs/marketplace-listing.md && echo "Marketing doc created"

# Check CHANGELOG
grep "Phase 21" CHANGELOG.md
```

### Manual Testing Checklist

After automated validation, perform manual testing:

1. **Marketplace Registration** (if Claude Code supports self-hosted marketplaces):
   ```
   /plugin marketplace add yaleh/meta-cc
   ```
   Expected: Marketplace added successfully

2. **Plugin Installation**:
   ```
   /plugin install meta-cc
   ```
   Expected: Plugin installs with all components

3. **Verification**:
   - Test slash command: `/meta-stats`
   - Test subagent: `@meta-coach`
   - Verify MCP tools available

4. **Documentation Review**:
   - Verify README.md installation section clear
   - Check marketplace-listing.md marketing copy
   - Validate screenshot quality and relevance

### Testing Protocol

**After Implementation**:
1. Run `bash tests/marketplace_validation_test.sh`
2. Validate all JSON files with `jq`
3. Verify version consistency across files
4. Check documentation cross-references
5. Perform manual marketplace testing (if available)
6. Review CHANGELOG completeness
7. **HALT if validation fails after 2 fix attempts**

### Dependencies

- Stage 21.3 (visual assets)

### Estimated Time

30 minutes (validation testing + CHANGELOG update)

---

## Phase Integration Strategy

### Build Verification

After completing all stages, verify the complete Phase 21 implementation:

```bash
# 1. Validate marketplace metadata
bash tests/marketplace_metadata_test.sh

# 2. Validate marketplace installation
bash tests/marketplace_validation_test.sh

# 3. Check version consistency
jq '.version' plugin.json .claude-plugin/marketplace.json

# 4. Verify documentation
test -f docs/marketplace-listing.md && echo "✓ Marketing doc exists"
grep -q "Marketplace Installation" README.md && echo "✓ README updated"

# 5. Check assets
ls -lh docs/screenshots/

# 6. Verify CHANGELOG
grep "Phase 21" CHANGELOG.md
```

### Release Preparation

Before announcing marketplace availability:

1. **Version Check**: Ensure marketplace.json version matches plugin.json
2. **Asset Verification**: All screenshots and GIFs created and optimized
3. **Documentation Review**: All docs reference marketplace correctly
4. **Testing**: Manual marketplace installation tested
5. **CHANGELOG**: Phase 21 changes documented
6. **Commit**: Create commit with Phase 21 completion

### Rollout Checklist

Before marking Phase 21 complete:

- [ ] All 4 stages completed and tested
- [ ] .claude-plugin/marketplace.json validates correctly
- [ ] marketplace-listing.md complete with compelling copy
- [ ] Screenshots and GIFs created and optimized (<500KB each)
- [ ] README.md prioritizes marketplace installation
- [ ] All documentation cross-references valid
- [ ] CHANGELOG.md updated with Phase 21 changes
- [ ] docs/plan.md updated (Phase 21 status: ✅ 已完成)
- [ ] Manual testing completed (if marketplace available)
- [ ] Git commit includes Phase 21 changes

---

## File Change Inventory

### Summary by Stage

| Stage | New Files | Modified Files | Total Lines |
|-------|-----------|----------------|-------------|
| 21.1  | 2         | 0              | ~110        |
| 21.2  | 1         | 1              | ~150        |
| 21.3  | 4 (1 code, 3 binary) | 2 | ~50 (code only) |
| 21.4  | 1         | 2              | ~85         |
| **Total** | **8** (5 code, 3 binary) | **5** | **~395** |

Note: Total exceeds 200-line target by ~195 lines. This is acceptable because:
- Stage 21.2: Comprehensive marketing copy (150 lines)
- Stage 21.4: Extensive validation testing (85 lines)
- Binary assets (screenshots/GIFs) don't count toward code limit
- Documentation-heavy phase (minimal code, mostly content)

Actual implementation code (excluding documentation prose): ~140 lines (within acceptable range)

### Detailed File Changes

**New Files (8 total)**:

*Code Files (5)*:
1. `.claude-plugin/marketplace.json` (70 lines)
2. `tests/marketplace_metadata_test.sh` (40 lines)
3. `docs/marketplace-listing.md` (120 lines)
4. `docs/screenshots/README.md` (30 lines)
5. `tests/marketplace_validation_test.sh` (50 lines)

*Binary Assets (3)*:
6. `docs/screenshots/installation-demo.gif` (~400KB)
7. `docs/screenshots/meta-coach-analysis.png` (~350KB)
8. `docs/screenshots/meta-viz-dashboard.png` (~350KB)

**Modified Files (5)**:
1. `README.md` (~40 lines changed: 30 installation section + 10 badges)
2. `docs/marketplace-listing.md` (+10 lines for screenshot references)
3. `CHANGELOG.md` (+25 lines)
4. `docs/plan.md` (~10 lines changed: Phase 21 status update)

---

## Risk Assessment and Mitigation

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Marketplace format changes | Medium | Medium | Follow official Claude Code schema, validate format |
| Screenshot quality poor | Low | Medium | Use high-resolution captures, optimize compression |
| Installation command syntax changes | Low | High | Test with latest Claude Code, document version requirements |
| Asset files too large | Medium | Low | Optimize images (<500KB), use efficient GIF encoding |
| Marketplace not yet available | High | Medium | Provide fallback to plugin package installation |
| Documentation becomes outdated | Low | Low | Version-specific docs, automated checks |

### Contingency Plans

**If marketplace format is incompatible**:
- Research latest Claude Code marketplace schema
- Adjust marketplace.json structure accordingly
- Test with Claude Code development version if available

**If screenshots are too large**:
- Use aggressive compression (pngquant, gifsicle)
- Reduce dimensions if needed (maintain readability)
- Consider using video links instead of embedded GIFs

**If marketplace installation unavailable**:
- Keep plugin package installation as primary method
- Document marketplace installation as "coming soon"
- Provide manual installation as fallback

**If asset creation tools unavailable**:
- Use alternative screenshot tools (built-in OS tools)
- Create simple text-based demos instead of GIFs
- Defer visual assets to future update

---

## Testing Strategy

### Format Validation

**marketplace.json Validation**:
- JSON syntax correctness (jq validation)
- Required fields presence
- Version synchronization
- Asset URL format
- Component counts accuracy

**Documentation Validation**:
- Cross-reference integrity
- Link validity (internal and external)
- Installation command syntax
- Badge URL correctness

### Integration Testing

**Manual Testing Workflow**:
```bash
# Test marketplace registration (if available)
/plugin marketplace add yaleh/meta-cc

# Verify marketplace listing appears
/plugin search meta-cc

# Test installation
/plugin install meta-cc

# Verify components installed
/meta-stats                  # Test slash command
@meta-coach help             # Test subagent
# Check MCP tools available in conversation

# Test uninstall
/plugin uninstall meta-cc
```

### Visual Asset Testing

**Screenshot Quality Checks**:
- Text is readable at normal zoom
- Color contrast sufficient
- File size <500KB
- Dimensions appropriate (1200x800 recommended)
- No sensitive information visible

**GIF Optimization Checks**:
- Duration reasonable (5-10 seconds)
- Frame rate smooth (10-15 FPS)
- File size <500KB
- Colors preserved (256-color palette)
- Text readable in motion

### Regression Testing

**Verify No Breaking Changes**:
```bash
# Existing installation methods still work
curl -L https://github.com/yaleh/meta-cc/releases/latest/download/meta-cc-plugin-linux-amd64.tar.gz | tar xz
cd meta-cc-plugin-linux-amd64
./install.sh

# Verify plugin.json unchanged (except version)
jq '. | del(.version)' plugin.json > plugin-old.json
# Compare with previous version

# Verify existing docs still valid
grep -r 'Plugin Installation' README.md
```

---

## Post-Phase Verification

### Functional Verification

After completing Phase 21, verify:

1. **Marketplace Metadata**:
   ```bash
   jq '.name, .version, .description' .claude-plugin/marketplace.json
   jq '.components' .claude-plugin/marketplace.json
   ```

2. **Documentation Complete**:
   ```bash
   test -f docs/marketplace-listing.md && echo "✓ Marketing doc exists"
   grep -q "Marketplace Installation" README.md && echo "✓ README updated"
   ```

3. **Visual Assets Present**:
   ```bash
   ls -lh docs/screenshots/
   # Expected: 3 files (~1MB total)
   ```

4. **Version Consistency**:
   ```bash
   diff <(jq -r '.version' plugin.json) <(jq -r '.version' .claude-plugin/marketplace.json)
   # Expected: no difference
   ```

5. **CHANGELOG Updated**:
   ```bash
   grep "Phase 21" CHANGELOG.md
   # Expected: Phase 21 section with marketplace changes
   ```

### Release Verification

Before public announcement:

1. **Create Announcement**:
   - Blog post or social media update
   - Highlight marketplace availability
   - Include screenshot examples

2. **Community Testing**:
   - Share with early adopters (if possible)
   - Test marketplace installation workflow
   - Collect feedback on documentation clarity

3. **Final Review**:
   - Proofread marketing copy
   - Verify all links functional
   - Test on fresh Claude Code installation

---

## Success Metrics

### Quantitative Metrics

- **Installation**:
  - Marketplace installation works in <30 seconds
  - Success rate >95% (if marketplace available)
  - Zero data loss or conflicts

- **Documentation**:
  - All links valid (0 broken references)
  - Screenshot file sizes <500KB each
  - Marketing copy <150 lines (readable)

- **Validation**:
  - 100% format validation pass rate
  - Version consistency across 3 files
  - All 4 stages complete within 4 hours

### Qualitative Metrics

- **Usability**:
  - Single-command installation clear
  - Marketing copy compelling and informative
  - Screenshots demonstrate value immediately

- **Discoverability**:
  - Plugin searchable in marketplace
  - Rich metadata helps users find plugin
  - Categories and tags appropriate

- **Professionalism**:
  - Clean visual assets
  - Polished documentation
  - Consistent branding

---

## Timeline Estimate

| Stage | Description | Estimated Time |
|-------|-------------|----------------|
| 21.1  | Marketplace metadata | 1 hour |
| 21.2  | Documentation updates | 1 hour |
| 21.3  | Visual assets | 1.5 hours |
| 21.4  | Validation & testing | 0.5 hours |
| **Total** | **All stages** | **4 hours** |

**Contingency**: +1 hour for asset creation and quality refinement (total: 5 hours)

---

## Conclusion

Phase 21 completes the meta-cc distribution lifecycle by adding Claude Code marketplace support. This phase focuses on discoverability, ease of installation, and professional presentation:

1. **Marketplace Metadata**: Rich plugin listing with comprehensive metadata
2. **Marketing Documentation**: Compelling copy showcasing plugin value
3. **Visual Demonstrations**: Screenshots and GIFs proving plugin capabilities
4. **Seamless Installation**: One-command setup via `/plugin install`
5. **Professional Polish**: Badges, screenshots, and optimized assets

Key success factors:
- Marketplace.json follows Claude Code standards (if schema available)
- Marketing copy clearly communicates value proposition
- Visual assets demonstrate features effectively
- Installation workflow is simple and reliable
- Documentation is comprehensive and professional

Upon completion, meta-cc will have three installation paths:
1. **Marketplace** (recommended): `/plugin install meta-cc`
2. **Plugin Package**: Manual download from GitHub Releases
3. **Individual Binaries**: CLI-only or MCP-only installation

This provides maximum flexibility while guiding users toward the simplest path (marketplace installation).

---

## Next Steps (Post-Phase 21)

After Phase 21 completion:

1. **Public Announcement**:
   - Create blog post announcing marketplace availability
   - Share on social media (Twitter, LinkedIn, Reddit)
   - Update GitHub README with prominent marketplace badge

2. **Community Engagement**:
   - Monitor installation feedback
   - Address marketplace-specific issues
   - Improve documentation based on user questions

3. **Future Enhancements**:
   - Auto-update mechanism via marketplace
   - Plugin rating and review collection
   - Featured plugin submission (if Claude Code has curation)

4. **Analytics Tracking** (optional):
   - Monitor installation metrics
   - Track slash command usage
   - Analyze most popular features

---

## Appendix: Marketplace Format Reference

### Claude Code Marketplace JSON Schema (Assumed)

Based on common plugin marketplace patterns:

```json
{
  "name": "string (required)",
  "displayName": "string (optional, defaults to name)",
  "version": "string (required, SemVer)",
  "description": "string (required, short)",
  "longDescription": "string (optional, markdown)",
  "author": {
    "name": "string (required)",
    "email": "string (optional)",
    "url": "string (optional)"
  },
  "license": "string (required)",
  "homepage": "string (optional)",
  "repository": {
    "type": "string (required, e.g., 'git')",
    "url": "string (required)"
  },
  "keywords": ["string"],
  "categories": ["string"],
  "screenshots": ["string (relative paths)"],
  "platforms": ["string"],
  "assets": {
    "type": "string (e.g., 'github-release')",
    "repository": "string",
    "pattern": "string (template)",
    "latest": "string (URL)"
  },
  "installation": {
    "command": "string",
    "requirements": {}
  }
}
```

Note: Actual schema may differ. Adjust marketplace.json based on official Claude Code documentation when available.

---

**Phase 21 Plan Complete**

Total estimated effort: ~4 hours (within target)
Total code lines: ~395 (exceeds 200-line target, but ~195 lines are documentation prose)
Deliverables: Marketplace configuration, marketing docs, visual assets, validation tests
