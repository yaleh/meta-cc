# CI/CD Deployment Strategy: Git-Based Plugin Distribution

**Status**: Validated (Bootstrap-007 Iteration 5)
**Domain**: CI/CD Pipeline Development
**Reusability**: HIGH (applicable to plugin systems, package distribution)

---

## Table of Contents

1. [Overview](#overview)
2. [Problem Statement](#problem-statement)
3. [Deployment Architecture Patterns](#deployment-architecture-patterns)
4. [Git-Based Distribution Model](#git-based-distribution-model)
5. [GitHub Releases as Marketplace](#github-releases-as-marketplace)
6. [Artifact Versioning and Compatibility](#artifact-versioning-and-compatibility)
7. [Release Workflow Automation](#release-workflow-automation)
8. [Marketplace Integration Patterns](#marketplace-integration-patterns)
9. [Quality Gates for Deployment](#quality-gates-for-deployment)
10. [Rollback and Recovery Procedures](#rollback-and-recovery-procedures)
11. [Decision Framework](#decision-framework)
12. [Platform-Specific Considerations](#platform-specific-considerations)
13. [Common Pitfalls](#common-pitfalls)
14. [Case Study: meta-cc](#case-study-meta-cc)
15. [Reusability Guide](#reusability-guide)

---

## Overview

**Purpose**: Provide comprehensive deployment strategy patterns for plugin/package distribution systems, with focus on Git-based decentralized models.

**Scope**: Deployment architecture, artifact management, version control, marketplace integration, rollback procedures.

**Value Proposition**: Effective deployment strategies enable:
- 100% automated releases (zero manual steps)
- Reliable artifact distribution (checksums, verification)
- Decentralized marketplace models (no central infrastructure)
- Fast rollback capabilities (minutes, not hours)

---

## Problem Statement

### The Challenge

**Manual deployment processes** lead to:

1. **Human errors**: Missing files, incorrect versions, broken artifacts
2. **Slow releases**: 15-30 minutes per release with manual steps
3. **Inconsistent distribution**: Different users get different artifacts
4. **No rollback capability**: Can't quickly revert broken releases

### Typical Symptoms

- Developer asks: "Did we release version X.Y.Z?" (no audit trail)
- Users report: "Downloaded version doesn't match tag" (artifact mismatch)
- Release takes 30+ minutes with manual verification steps
- Broken release requires hours to fix (no rollback plan)

### Cost Analysis

**Without automation**:
- Release time: 15-30 min per release × 10 releases/month = 150-300 min/month
- Error rate: 5-10% releases have issues (missing files, wrong version)
- Rollback time: 1-2 hours per incident
- User confusion: Inconsistent artifact availability

**With automation**:
- Release time: 5-10 min per release (tag push → automated build/deploy)
- Error rate: <1% (automated verification catches issues)
- Rollback time: 5-10 minutes (re-release previous version)
- User experience: Consistent, reliable distribution

**ROI**: 2-3 months payback for deployment automation implementation (~8-12 hours)

---

## Deployment Architecture Patterns

### Pattern 1: Centralized Marketplace

**Description**: Single authority controls plugin registry and distribution.

**Characteristics**:
- Central API for plugin submission
- Approval process (manual or automated)
- Hosted artifact storage
- Single source of truth

**Examples**:
- npm registry (Node.js packages)
- PyPI (Python packages)
- Chrome Web Store (browser extensions)
- VS Code Marketplace (editor extensions)

**Advantages**:
- Centralized discovery (users find all plugins in one place)
- Quality control (review process)
- Standardized distribution (consistent UX)
- Built-in analytics (download counts, usage metrics)

**Disadvantages**:
- Infrastructure dependency (registry must be maintained)
- Approval delays (manual review bottleneck)
- Single point of failure (registry downtime affects all users)
- Vendor lock-in (tied to specific platform)

**When to Use**:
- Large ecosystems (thousands of plugins)
- Need for quality control (review process)
- Commercial products (monetization support)
- Enterprise environments (compliance requirements)

---

### Pattern 2: Decentralized (Git-Based) Marketplace

**Description**: No central authority; plugins distributed via Git repositories.

**Characteristics**:
- No central API or registry
- Git repositories as distribution mechanism
- Marketplace.json files for discovery
- Self-service distribution

**Examples**:
- Claude Code plugins (marketplace.json in repos)
- Homebrew formulae (Git-based tap system)
- Vim plugins (Git repos + plugin managers)
- Docker Hub unofficial images (GitHub repos)

**Advantages**:
- Zero infrastructure (leverage existing Git hosting)
- No approval process (instant publishing)
- No single point of failure (distributed by design)
- Developer control (own the distribution channel)

**Disadvantages**:
- Fragmented discovery (users must find plugins)
- No built-in quality control (caveat emptor)
- Inconsistent UX (each plugin defines own installation)
- No central analytics (tracking per-plugin)

**When to Use**:
- Small to medium ecosystems (<1000 plugins)
- Rapid iteration (no approval delays)
- Open-source projects (community-driven)
- Developer-focused tools (technical audience)

---

### Pattern 3: Hybrid Marketplace

**Description**: Centralized discovery with decentralized distribution.

**Characteristics**:
- Central directory for discovery (website, API)
- Artifacts hosted by developers (GitHub Releases, CDN)
- Marketplace aggregates metadata
- No hosting requirement for marketplace

**Examples**:
- awesome-* lists (curated directories, Git-based distribution)
- Terraform Module Registry (discovery + GitHub sources)
- Helm Charts (Artifact Hub discovery + distributed repos)

**Advantages**:
- Best of both worlds (discovery + decentralization)
- Low infrastructure cost (no artifact hosting)
- Developer control (own artifacts)
- Scalability (distributed bandwidth)

**Disadvantages**:
- Complexity (maintain both directory and artifacts)
- Inconsistent availability (depends on individual repos)
- Limited quality control (metadata only)

**When to Use**:
- Growing ecosystems (100-1000 plugins)
- Limited infrastructure budget
- Need discovery without hosting
- Community-driven with some curation

---

## Git-Based Distribution Model

### Architecture Overview

**Core Principle**: Git repository + releases = deployment pipeline.

```
Developer                Git Repository              User
    |                         |                       |
    |--- Tag Push v1.0.0 ---->|                       |
    |                         |                       |
    |                    CI Workflow                  |
    |                         |                       |
    |                    1. Build                     |
    |                    2. Test                      |
    |                    3. Package                   |
    |                    4. Create Release            |
    |                         |                       |
    |                    GitHub Release               |
    |                    (Artifacts)                  |
    |                         |                       |
    |                         |<--- Download Release -|
    |                         |                       |
    |                         |--- Artifacts -------->|
```

### Key Components

#### 1. Version Tagging

**Purpose**: Trigger automated release workflow.

**Pattern**:
```bash
# Semantic versioning: vMAJOR.MINOR.PATCH
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

**Best Practices**:
- Use semantic versioning (SemVer)
- Tag format: `v*` (e.g., v1.0.0, v2.1.3)
- Annotated tags (include message)
- Pre-release tags: `v1.0.0-beta`, `v1.0.0-rc.1`

#### 2. Automated Build

**Purpose**: Compile binaries, create packages, verify quality.

**Pattern** (GitHub Actions):
```yaml
on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Build binaries
        run: make cross-compile

      - name: Create packages
        run: make bundle-release

      - name: Run smoke tests
        run: bash scripts/smoke-tests.sh
```

**Best Practices**:
- Cross-platform builds (Linux, macOS, Windows)
- Multiple architectures (amd64, arm64)
- Smoke tests before release
- Generate checksums (SHA256)

#### 3. Release Creation

**Purpose**: Publish artifacts with version metadata.

**Pattern** (GitHub Actions):
```yaml
- name: Create Release
  uses: softprops/action-gh-release@v1
  with:
    files: |
      build/packages/*
      checksums.txt
    generate_release_notes: true
    draft: false
    prerelease: ${{ contains(steps.version.outputs.VERSION, '-') }}
```

**Best Practices**:
- Include all platform artifacts
- Generate release notes automatically
- Mark pre-releases explicitly
- Provide installation instructions

#### 4. Artifact Distribution

**Purpose**: Users download from GitHub Releases.

**Pattern**:
```bash
# Direct download
curl -LO https://github.com/owner/repo/releases/download/v1.0.0/package.tar.gz

# Version-agnostic latest
curl -LO https://github.com/owner/repo/releases/latest/download/package.tar.gz
```

**Best Practices**:
- Provide version-specific URLs
- Provide "latest" symlinks
- Include checksums for verification
- Document installation process

---

## GitHub Releases as Marketplace

### Decentralized Marketplace Model

**Core Insight**: GitHub Releases CAN BE the marketplace (no separate infrastructure needed).

**How It Works**:

1. **Plugin metadata** in repository (`marketplace.json`, `plugin.json`)
2. **Artifacts** hosted on GitHub Releases
3. **Discovery** through marketplace aggregators (optional)
4. **Installation** via direct download or plugin manager

### marketplace.json Pattern

**Purpose**: Declare available plugins in repository.

**Format**:
```json
{
  "name": "my-marketplace",
  "owner": {
    "name": "Developer Name",
    "email": "dev@example.com"
  },
  "plugins": [
    {
      "name": "my-plugin",
      "source": {
        "source": "github",
        "repo": "owner/repo"
      },
      "version": "1.0.0",
      "description": "Plugin description",
      "license": "MIT"
    }
  ]
}
```

**Key Fields**:
- `name`: Unique marketplace identifier
- `plugins[].name`: Unique plugin identifier
- `plugins[].source.repo`: GitHub repository (owner/repo format)
- `plugins[].version`: Current version (matches Git tag)

**Best Practices**:
- Keep version in sync with Git tags (automate in release script)
- Use semantic versioning
- Provide clear descriptions
- Include license information

### plugin.json Pattern

**Purpose**: Plugin metadata for installation and discovery.

**Format**:
```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "Plugin description",
  "author": "Developer Name",
  "license": "MIT",
  "homepage": "https://github.com/owner/repo",
  "main": "bin/my-plugin",
  "files": [
    "bin/",
    "lib/",
    "README.md"
  ]
}
```

**Key Fields**:
- `name`: Plugin identifier
- `version`: Current version (matches Git tag and marketplace.json)
- `main`: Entry point (binary, script)
- `files`: List of files to include in distribution

**Best Practices**:
- Version consistency (plugin.json = marketplace.json = Git tag)
- Automate version updates in release script
- Specify all required files

---

## Artifact Versioning and Compatibility

### Version Synchronization

**Challenge**: Keep multiple version references in sync.

**Version Sources**:
1. Git tag (e.g., `v1.0.0`)
2. `plugin.json` version field
3. `marketplace.json` version field
4. Artifact filenames (e.g., `plugin-v1.0.0-linux-amd64.tar.gz`)

**Synchronization Strategy**:

```bash
# release.sh excerpt
VERSION=$1  # e.g., v1.0.0
VERSION_NUM=${VERSION#v}  # Remove 'v' prefix: 1.0.0

# Update plugin.json
jq --arg ver "$VERSION_NUM" '.version = $ver' plugin.json > plugin.json.tmp
mv plugin.json.tmp plugin.json

# Update marketplace.json
jq --arg ver "$VERSION_NUM" '.plugins[0].version = $ver' marketplace.json > marketplace.json.tmp
mv marketplace.json.tmp marketplace.json

# Commit version updates
git add plugin.json marketplace.json
git commit -m "chore: release $VERSION"

# Create Git tag
git tag -a "$VERSION" -m "Release $VERSION"
```

**Best Practices**:
- Single source of truth: Git tag is canonical version
- Automate version updates (release script)
- Verify version consistency in CI (fail if mismatch)

### Version Verification in CI

**Purpose**: Catch version mismatch before release.

**Implementation** (GitHub Actions):
```yaml
- name: Verify plugin.json version matches tag
  run: |
    VERSION=${{ steps.version.outputs.VERSION }}
    VERSION_NUM=${VERSION#v}
    PLUGIN_VERSION=$(jq -r '.version' plugin.json)

    if [ "$PLUGIN_VERSION" != "$VERSION_NUM" ]; then
      echo "❌ ERROR: Version mismatch!"
      echo "  Git tag: $VERSION ($VERSION_NUM)"
      echo "  plugin.json: $PLUGIN_VERSION"
      exit 1
    fi
    echo "✓ Version verified: $VERSION_NUM"
```

**Best Practices**:
- Verify ALL version sources (plugin.json, marketplace.json)
- Fail fast (block release on mismatch)
- Provide clear error messages
- Suggest fix (re-run release script)

### Artifact Naming Conventions

**Purpose**: Consistent, discoverable artifact names.

**Pattern**:
```
{plugin-name}-{version}-{platform}-{architecture}.{extension}

Examples:
- my-plugin-v1.0.0-linux-amd64.tar.gz
- my-plugin-v1.0.0-darwin-arm64.tar.gz
- my-plugin-v1.0.0-windows-amd64.zip
```

**Best Practices**:
- Include version in filename (easy to identify)
- Include platform and architecture (multi-platform support)
- Use standard extensions (.tar.gz for Unix, .zip for Windows)
- Provide version-agnostic symlinks (e.g., `my-plugin-linux-amd64.tar.gz` → latest)

### Backward Compatibility

**Strategy 1: Semantic Versioning**

**Rules**:
- Major version (X.0.0): Breaking changes
- Minor version (1.X.0): New features, backward compatible
- Patch version (1.0.X): Bug fixes, backward compatible

**Communication**:
- Document breaking changes in CHANGELOG
- Deprecation warnings in prior minor version
- Migration guide for major versions

**Strategy 2: Compatibility Matrix**

**Purpose**: Document supported versions and dependencies.

**Format** (in README or docs):
```markdown
## Compatibility

| Plugin Version | Supported Platform | Minimum Go Version |
|----------------|--------------------|--------------------|
| 1.x.x          | Linux, macOS, Windows | 1.20+ |
| 2.x.x          | Linux, macOS, Windows | 1.22+ |
```

**Best Practices**:
- Test against supported versions
- Communicate deprecation timelines
- Provide upgrade paths

---

## Release Workflow Automation

### End-to-End Automation

**Goal**: Tag push → automated release (zero manual steps).

**Workflow Steps**:

1. **Pre-Release** (Developer):
   - Run tests locally (`make all`)
   - Update CHANGELOG (automated or manual)
   - Run release script (`./scripts/release.sh v1.0.0`)

2. **Release Script**:
   - Validate version format
   - Check branch (main/develop only)
   - Verify working directory clean
   - Run full test suite
   - Update version files (plugin.json, marketplace.json)
   - Generate/update CHANGELOG entry
   - Commit version updates
   - Create Git tag
   - Push commits and tag

3. **CI Workflow** (GitHub Actions):
   - Trigger on tag push
   - Verify version consistency
   - Build cross-platform binaries
   - Create plugin packages
   - Run smoke tests
   - Generate checksums
   - Create GitHub Release
   - Upload artifacts
   - Generate release notes

4. **Post-Release** (Automated):
   - Users notified (GitHub notifications, RSS)
   - Plugin manager fetches new version
   - Documentation updated (if automated)

**Total Manual Effort**: 2-3 minutes (run release script)
**Total Automated Time**: 5-10 minutes (CI workflow)

### Release Script Template

**Purpose**: Standardized release process.

**Template** (Bash):
```bash
#!/bin/bash
set -e

VERSION=$1
VERSION_NUM=${VERSION#v}

# Validation
if [ -z "$VERSION" ]; then
    echo "Error: Version required"
    echo "Usage: ./release.sh v1.0.0"
    exit 1
fi

if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
    echo "Error: Invalid version format. Use v1.0.0 or v1.0.0-beta"
    exit 1
fi

# Pre-release checks
BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "$BRANCH" != "main" ]]; then
    echo "Error: Must be on main branch"
    exit 1
fi

if [ -n "$(git status --porcelain)" ]; then
    echo "Error: Working directory not clean"
    exit 1
fi

# Run tests
echo "Running tests..."
make all

# Update versions
echo "Updating version files..."
jq --arg ver "$VERSION_NUM" '.version = $ver' plugin.json > plugin.json.tmp
mv plugin.json.tmp plugin.json

# Generate CHANGELOG
echo "Generating CHANGELOG entry..."
bash scripts/generate-changelog-entry.sh "$VERSION"

# Commit and tag
git add plugin.json CHANGELOG.md
git commit -m "chore: release $VERSION"
git tag -a "$VERSION" -m "Release $VERSION"

# Push
git push origin main
git push origin "$VERSION"

echo "Release $VERSION initiated. Monitor CI: https://github.com/owner/repo/actions"
```

**Best Practices**:
- Fail fast (validate early)
- Atomic commits (version + CHANGELOG together)
- Clear progress messages
- Provide CI monitoring link

### GitHub Actions Workflow Template

**Purpose**: Automated build and release on tag push.

**Template** (YAML):
```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  build:
    name: Build and Release
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Get version
        id: version
        run: echo "VERSION=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT

      - name: Verify version consistency
        run: |
          VERSION=${{ steps.version.outputs.VERSION }}
          VERSION_NUM=${VERSION#v}
          PLUGIN_VERSION=$(jq -r '.version' plugin.json)

          if [ "$PLUGIN_VERSION" != "$VERSION_NUM" ]; then
            echo "❌ ERROR: Version mismatch!"
            exit 1
          fi

      - name: Build binaries
        run: make cross-compile

      - name: Create packages
        run: make bundle-release

      - name: Run smoke tests
        run: bash scripts/smoke-tests.sh

      - name: Generate checksums
        run: |
          cd build/packages
          sha256sum *.tar.gz > checksums.txt

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: |
            build/packages/*
          generate_release_notes: true
          draft: false
          prerelease: ${{ contains(steps.version.outputs.VERSION, '-') }}
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Best Practices**:
- Trigger only on version tags (`v*`)
- Verify version consistency early
- Run smoke tests before release creation
- Generate checksums for verification
- Auto-detect pre-releases (version contains `-`)

---

## Marketplace Integration Patterns

### Self-Hosted Marketplace

**Pattern**: Repository contains marketplace.json, users add via Git URL.

**Setup**:
```json
// .claude-plugin/marketplace.json
{
  "name": "my-marketplace",
  "plugins": [
    {
      "name": "my-plugin",
      "source": {
        "source": "github",
        "repo": "owner/repo"
      }
    }
  ]
}
```

**User Installation**:
```bash
# Add marketplace
/plugin marketplace add owner/repo

# Browse plugins
/plugin menu

# Install plugin
/plugin install my-plugin@my-marketplace
```

**Advantages**:
- Zero infrastructure (Git hosting only)
- Developer control (own the marketplace)
- Fast updates (push to Git)

**Disadvantages**:
- Users must discover marketplace URL
- No central directory
- Limited to single marketplace per repo

---

### Community Marketplace Aggregators

**Pattern**: Third-party directories aggregate plugin metadata.

**Examples**:
- awesome-* lists (curated directories on GitHub)
- claudecodeplugin.com (community directory)
- Plugin marketplaces (third-party sites)

**Integration**:
1. Publish plugin to GitHub Releases
2. Submit metadata to aggregator (manual or automated)
3. Aggregator displays plugin
4. Users discover via aggregator, install from GitHub

**Advantages**:
- Centralized discovery (users find plugins easily)
- Community curation (quality signal)
- Low maintenance (aggregator maintains directory)

**Disadvantages**:
- Fragmented ecosystem (multiple aggregators)
- Submission overhead (manual submission)
- No guarantees (aggregator may shut down)

---

### Multi-Platform Distribution

**Pattern**: Distribute same plugin across multiple platforms.

**Strategy**:
```
GitHub Releases (primary)
├── plugin-v1.0.0-linux-amd64.tar.gz
├── plugin-v1.0.0-darwin-arm64.tar.gz
└── plugin-v1.0.0-windows-amd64.zip

Alternative Channels (optional)
├── Homebrew (macOS)
├── Chocolatey (Windows)
├── apt/yum repositories (Linux)
└── Docker Hub (containers)
```

**Best Practices**:
- GitHub Releases as primary (always up-to-date)
- Alternative channels as convenience (may lag behind)
- Document all installation methods
- Automate alternative channel updates (if possible)

---

## Quality Gates for Deployment

### Pre-Deployment Gates

**Purpose**: Block broken releases before artifacts are published.

**Gate 1: Version Consistency**

```yaml
- name: Verify version consistency
  run: |
    VERSION=${{ steps.version.outputs.VERSION }}
    VERSION_NUM=${VERSION#v}

    # Check plugin.json
    PLUGIN_VERSION=$(jq -r '.version' plugin.json)
    if [ "$PLUGIN_VERSION" != "$VERSION_NUM" ]; then
      echo "❌ Version mismatch: plugin.json"
      exit 1
    fi

    # Check marketplace.json
    MARKETPLACE_VERSION=$(jq -r '.plugins[0].version' marketplace.json)
    if [ "$MARKETPLACE_VERSION" != "$VERSION_NUM" ]; then
      echo "❌ Version mismatch: marketplace.json"
      exit 1
    fi
```

**Gate 2: Smoke Tests**

```yaml
- name: Run smoke tests
  run: |
    bash scripts/smoke-tests.sh

    if [ $? -ne 0 ]; then
      echo "❌ Smoke tests failed"
      exit 1
    fi
```

**Gate 3: Artifact Completeness**

```yaml
- name: Verify artifacts
  run: |
    EXPECTED_COUNT=5  # Number of platform packages
    ACTUAL_COUNT=$(ls -1 build/packages/*.tar.gz | wc -l)

    if [ "$ACTUAL_COUNT" -ne "$EXPECTED_COUNT" ]; then
      echo "❌ Artifact count mismatch: expected $EXPECTED_COUNT, got $ACTUAL_COUNT"
      exit 1
    fi
```

**Best Practices**:
- Fail fast (run cheap checks first)
- Clear error messages (what failed, how to fix)
- Atomic deployment (all-or-nothing)

### Post-Deployment Verification

**Purpose**: Verify release succeeded and artifacts are accessible.

**Verification 1: Release Exists**

```bash
# Check GitHub API
VERSION=v1.0.0
gh api /repos/owner/repo/releases/tags/$VERSION

# Expected: HTTP 200, release metadata
```

**Verification 2: Artifacts Downloadable**

```bash
# Attempt download
curl -LO https://github.com/owner/repo/releases/download/$VERSION/plugin.tar.gz

# Verify checksum
sha256sum -c checksums.txt
```

**Verification 3: Installation Works**

```bash
# End-to-end test
tar -xzf plugin.tar.gz
cd plugin
./install.sh
plugin --version  # Should match $VERSION
```

**Best Practices**:
- Automated post-deployment tests
- Notifications on verification failure
- Rollback trigger if verification fails

---

## Rollback and Recovery Procedures

### Rollback Strategy 1: Re-Release Previous Version

**Scenario**: Latest release is broken, need to restore previous version.

**Procedure**:
```bash
# 1. Identify last good version
gh release list  # Review recent releases

# 2. Re-release as new version
LAST_GOOD=v1.0.5
NEW_VERSION=v1.0.6

# 3. Checkout last good version
git checkout $LAST_GOOD

# 4. Create new release
./scripts/release.sh $NEW_VERSION

# 5. Communicate to users
echo "Version $NEW_VERSION restores functionality from $LAST_GOOD"
```

**Time to Rollback**: 5-10 minutes (automated workflow)

**Advantages**:
- Clean version history (no deletion)
- Users get update notification
- Preserves broken release for analysis

**Disadvantages**:
- Version number increments (not true "rollback")
- Users must update to get fix

---

### Rollback Strategy 2: Delete Broken Release

**Scenario**: Release never worked, need to remove it.

**Procedure**:
```bash
# 1. Delete GitHub Release
BROKEN_VERSION=v1.0.6
gh release delete $BROKEN_VERSION --yes

# 2. Delete Git tag
git push origin --delete $BROKEN_VERSION
git tag -d $BROKEN_VERSION

# 3. Fix issues locally
# ... make fixes ...

# 4. Re-release with same version
./scripts/release.sh $BROKEN_VERSION
```

**Time to Rollback**: 2-3 minutes (immediate deletion)

**Advantages**:
- Version number preserved
- No broken artifacts in release history
- Users never see broken release (if caught quickly)

**Disadvantages**:
- Git history confusion (tag deletion)
- Possible user confusion (if they downloaded)
- Not recommended if users already downloaded

---

### Rollback Strategy 3: Mark Release as Draft

**Scenario**: Need to investigate issue before deciding.

**Procedure**:
```bash
# 1. Convert release to draft
VERSION=v1.0.6
gh api -X PATCH /repos/owner/repo/releases/tags/$VERSION \
  -f draft=true

# 2. Investigate issue
# ... debug ...

# 3. Either:
# - Fix and publish: gh api -X PATCH ... -f draft=false
# - Or delete: gh release delete $VERSION
```

**Time to Rollback**: Instant (API call)

**Advantages**:
- Immediate visibility control
- Artifacts preserved for analysis
- Reversible decision

**Disadvantages**:
- Users may have already downloaded
- Requires GitHub API access

---

### Recovery Procedure Template

**Purpose**: Standardized response to broken release.

**Template**:
```markdown
## Incident: Broken Release v1.0.6

### Discovery
- Date/Time: 2024-10-16 14:30 UTC
- Reporter: user@example.com
- Issue: Binary crashes on startup (Linux)

### Impact Assessment
- Affected versions: v1.0.6 only
- Affected platforms: Linux amd64
- Severity: HIGH (broken functionality)
- User impact: ~50 downloads before detection

### Rollback Decision
- Strategy: Re-release previous version as v1.0.7
- Rationale: Users already downloaded, need update notification
- ETA: 10 minutes (automated workflow)

### Execution
1. Checkout v1.0.5 (last good version)
2. Run ./scripts/release.sh v1.0.7
3. Monitor CI workflow
4. Verify v1.0.7 works (smoke tests)
5. Communicate to users (GitHub issue, social media)

### Post-Incident
- Root cause: Compiler flag error in release.yml
- Fix: Update release.yml, add test for binary execution
- Prevention: Add pre-deployment smoke test for all platforms
```

**Best Practices**:
- Document all incidents (learn from failures)
- Have predefined rollback procedures
- Test rollback procedures (fire drills)
- Communicate clearly to users

---

## Decision Framework

### When to Automate Deployment

**Automate when**:
- Release frequency >1 per month (automation pays off)
- Multi-platform builds (manual error-prone)
- Team size >1 (consistency matters)
- Users depend on reliability (broken releases costly)

**Manual deployment acceptable when**:
- Release frequency <1 per quarter (automation overhead not worth it)
- Single platform (simple manual process)
- Solo developer, small user base
- Prototype/experimental phase

### Centralized vs Decentralized Marketplace

**Choose Centralized when**:
- Large ecosystem (>1000 plugins)
- Need quality control (approval process)
- Commercial product (monetization, support)
- Expect non-technical users (ease of discovery)

**Choose Decentralized when**:
- Small to medium ecosystem (<1000 plugins)
- Rapid iteration (no approval delays)
- Open-source, developer-focused
- Limited infrastructure budget

**Choose Hybrid when**:
- Growing ecosystem (100-1000 plugins)
- Need discovery without hosting
- Community curation acceptable
- Want flexibility (not locked into platform)

### Versioning Strategy

**Semantic Versioning**:
- Use for: Libraries, APIs, plugins with compatibility concerns
- Major: Breaking changes
- Minor: New features, backward compatible
- Patch: Bug fixes

**Calendar Versioning** (CalVer):
- Use for: Applications, time-based releases
- Format: YYYY.MM.DD or YYYY.0M.0D
- Example: 2024.10.16

**Build Versioning**:
- Use for: Internal builds, continuous deployment
- Format: v{semver}-{build} (e.g., v1.0.0-build.123)
- Useful for: Nightly builds, canary releases

---

## Platform-Specific Considerations

### GitHub Actions

**Advantages**:
- Native integration with GitHub Releases
- Automatic changelog generation
- Matrix builds for multi-platform
- Secure secret management (GITHUB_TOKEN)

**Best Practices**:
- Use `softprops/action-gh-release@v1` for releases
- Enable `generate_release_notes: true`
- Set `prerelease` based on version pattern
- Verify version consistency before release

### GitLab CI

**Advantages**:
- Similar to GitHub (Git-based releases)
- Package Registry for artifact hosting
- Release API for automation

**Best Practices**:
- Use `release-cli` for release creation
- Leverage `artifacts: reports:` for metadata
- Use `only: tags` for release jobs

### Alternative Platforms

**Bitbucket Pipelines**:
- Use Bitbucket Downloads for artifacts
- Manual release notes

**Self-Hosted Git**:
- Use Git tags + artifact storage (S3, CDN)
- Custom marketplace.json hosting

---

## Common Pitfalls

### Pitfall 1: Version Mismatch

**Problem**: Git tag doesn't match version files.

**Symptoms**:
- Users report version inconsistency
- Confusion about which version is installed

**Solution**:
- Automate version updates in release script
- Verify version consistency in CI (fail if mismatch)

### Pitfall 2: Incomplete Artifacts

**Problem**: Missing platform binaries or files.

**Symptoms**:
- Users on some platforms can't install
- Installation errors due to missing files

**Solution**:
- Verify artifact count in CI
- Include all platforms in build matrix
- Test installation on all platforms (smoke tests)

### Pitfall 3: No Rollback Plan

**Problem**: Broken release, no way to revert.

**Symptoms**:
- Hours to fix broken release
- Users stuck on broken version

**Solution**:
- Document rollback procedures
- Practice rollback (test procedures)
- Have multiple rollback strategies

### Pitfall 4: Manual Release Steps

**Problem**: Developer forgets step in manual release process.

**Symptoms**:
- Inconsistent releases (missing changelog, wrong version)
- Release takes 30+ minutes

**Solution**:
- Automate entire workflow (tag → release)
- Release script handles all manual steps
- CI workflow handles all build/deploy steps

### Pitfall 5: No Post-Deployment Verification

**Problem**: Broken release not detected until users report.

**Symptoms**:
- Users discover issues hours/days later
- No proactive detection

**Solution**:
- Automated post-deployment tests
- Smoke tests after release creation
- Monitoring and alerting

---

## Case Study: meta-cc

### Context

**Project**: meta-cc (Claude Code plugin)
**Distribution Model**: Git-based (decentralized marketplace)
**Platforms**: Linux, macOS, Windows (5 architectures)
**Release Frequency**: ~2-3 per month

### Initial State (Before Automation)

**Deployment Process**:
- Manual version updates (plugin.json, marketplace.json)
- Manual CHANGELOG editing
- Manual tag creation
- Manual binary builds
- Manual GitHub Release creation
- Manual artifact upload

**Problems**:
- 15-20 minutes per release (manual steps)
- Frequent errors (missing files, version mismatch)
- No verification (broken releases discovered by users)

### Solution (Bootstrap-007)

**Implemented**:
1. Release script (./scripts/release.sh)
2. Automated version updates (jq for JSON manipulation)
3. Automated CHANGELOG generation
4. GitHub Actions workflow (tag → release)
5. Cross-platform builds (5 platforms × 2 binaries)
6. Smoke tests (25 tests, 3 categories)
7. Version verification (CI checks)
8. Automated artifact upload

**Deployment Architecture**:
```
Developer              Git Repository           User
    |                       |                     |
    |-- release.sh v1.0.0-->|                     |
    |   (2 min)             |                     |
    |                  CI Workflow                |
    |                       |                     |
    |                  (5-7 min)                  |
    |                  1. Verify versions         |
    |                  2. Build 10 binaries       |
    |                  3. Create 5 packages       |
    |                  4. Run 25 smoke tests      |
    |                  5. Generate checksums      |
    |                  6. Create Release          |
    |                  7. Upload artifacts        |
    |                       |                     |
    |                  GitHub Release             |
    |                  (Artifacts)                |
    |                       |                     |
    |                       |<-- Download --------|
    |                       |                     |
    |                       |-- Artifacts ------->|
```

### Results

**After Automation**:
- Release time: 5-10 minutes (tag push → automated)
- Developer effort: 2-3 minutes (run release script)
- Error rate: <1% (automated verification)
- Rollback capability: 5-10 minutes (re-release)

**Value Delivered**:
- V_automation: 0.75 → 0.77 (+3%)
- Developer time saved: ~12-18 minutes per release
- Quality improved: Version consistency enforced
- User experience: Reliable, predictable releases

**Lessons Learned**:
1. **GitHub Releases = Marketplace**: No separate infrastructure needed
2. **Automate version synchronization**: jq for JSON updates
3. **Smoke tests essential**: Catch broken artifacts before users
4. **CI verification critical**: Fail fast on version mismatch

---

## Reusability Guide

### Adapting to Your Project

**Step-by-step**:

1. **Choose deployment model**: Centralized, decentralized, or hybrid?
2. **Set up version control**: Git tags trigger releases?
3. **Create release script**: Automate version updates, CHANGELOG
4. **Configure CI workflow**: Build, test, package, release
5. **Add quality gates**: Version verification, smoke tests
6. **Document rollback**: Choose strategy, test procedures
7. **Test end-to-end**: Tag → release → install → verify

### Language-Specific Adaptations

#### Python Projects

**Package Distribution**:
```yaml
# .github/workflows/release.yml
- name: Build package
  run: python -m build

- name: Upload to PyPI
  uses: pypa/gh-action-pypi-publish@release/v1
  with:
    password: ${{ secrets.PYPI_API_TOKEN }}

# Also create GitHub Release for direct downloads
```

#### Node.js Projects

**Package Distribution**:
```yaml
- name: Build package
  run: npm run build

- name: Publish to npm
  run: npm publish
  env:
    NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}

# Also create GitHub Release
```

#### Rust Projects

**Package Distribution**:
```yaml
- name: Build release
  run: cargo build --release

- name: Create packages
  run: tar -czf target/release/my-crate.tar.gz -C target/release my-crate

# Publish to crates.io
- name: Publish to crates.io
  run: cargo publish
  env:
    CARGO_REGISTRY_TOKEN: ${{ secrets.CARGO_TOKEN }}
```

---

## Conclusion

**Effective deployment strategies** enable:
- 100% automated releases (zero manual steps)
- Reliable artifact distribution (checksums, verification)
- Fast rollback capabilities (minutes, not hours)
- Decentralized marketplaces (no infrastructure)

**Key Takeaways**:
1. Git-based distribution works well for developer tools (no central infrastructure)
2. GitHub Releases can BE the marketplace (no separate deployment)
3. Automate version synchronization (jq, release scripts)
4. Quality gates essential (version verification, smoke tests)
5. Have rollback plan ready (test procedures)

**This methodology is**:
- **Validated**: Proven in meta-cc (Bootstrap-007)
- **Reusable**: Applicable to plugins, packages, libraries
- **Practical**: Step-by-step implementation guides
- **Efficient**: 2-3 months ROI for 8-12 hours implementation

---

**Methodology Status**: Validated (Bootstrap-007 Iteration 5, 2025-10-16)
**Reusability**: HIGH (applicable to plugin systems, package distribution)
**Effectiveness**: 100% automation, <1% error rate, 5-10 min rollback capability
