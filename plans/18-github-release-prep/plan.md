# Phase 18: GitHub Release Preparation - Development Plan

## Phase Overview

**Objective**: Establish complete open-source release infrastructure with automated CI/CD, binary distribution, and community guidelines to prepare for v1.0 public release.

**Background & Problems**:
- **Problem 1**: No LICENSE file - GitHub shows "No license", users cannot legally use/fork the project
- **Problem 2**: No CI/CD pipeline - PRs not automatically tested, cross-platform builds are manual
- **Problem 3**: No automated releases - README promises "pre-compiled binary (coming soon)" but not implemented
- **Problem 4**: Missing community documentation - No CONTRIBUTING.md, CODE_OF_CONDUCT.md, SECURITY.md
- **Problem 5**: No Issue/PR templates - Inconsistent bug reports and contribution workflows

**Success Criteria**:
- MIT LICENSE file added and recognized by GitHub
- CI/CD workflows operational (test + lint + release)
- Automated binary releases for 5 platforms (Linux amd64/arm64, macOS amd64/arm64, Windows amd64)
- Complete community documentation (CONTRIBUTING, CODE_OF_CONDUCT, SECURITY)
- Issue/PR templates functional
- README updated with badges and installation instructions
- Repository settings configured per open-source best practices
- `make all` passes after all stages
- Phase 18 marked complete in docs/plan.md

**Technical Constraints**:
- Code limit per phase: ≤1,250 lines (per Phase 18 estimate in docs/plan.md)
- Code limit per stage: ≤500 lines
- Test-driven development (TDD) where applicable
- Zero breaking changes to existing functionality
- All CI/CD must use GitHub Actions

---

## Stage 1: Open Source Licensing and Compliance

### Objective

Add LICENSE file, SECURITY policy, and third-party notices to ensure legal compliance and establish vulnerability reporting process.

### Acceptance Criteria

- [ ] LICENSE file created with MIT License text
- [ ] Copyright year set to 2025
- [ ] Copyright holder: "meta-cc contributors"
- [ ] GitHub automatically detects and displays MIT License
- [ ] SECURITY.md created with vulnerability reporting process
- [ ] NOTICE file created (if third-party dependencies require attribution)
- [ ] All third-party dependencies are MIT-compatible

### TDD Approach

**No automated tests needed** (documentation files). Manual verification:
- Check GitHub repository settings show "MIT License"
- Validate LICENSE text against official MIT template
- Verify SECURITY.md contains contact information

### Implementation Tasks

1. **Create LICENSE file** (~20 lines):
   ```
   MIT License

   Copyright (c) 2025 meta-cc contributors

   Permission is hereby granted, free of charge, to any person obtaining a copy
   of this software and associated documentation files (the "Software"), to deal
   in the Software without restriction...
   ```

2. **Create SECURITY.md** (~30 lines):
   ```markdown
   # Security Policy

   ## Supported Versions

   | Version | Supported          |
   | ------- | ------------------ |
   | 0.x.x   | :white_check_mark: |

   ## Reporting a Vulnerability

   Please report security vulnerabilities to: [email/contact method]

   Expected response time: 48 hours
   ```

3. **Create NOTICE file** (if needed, ~10 lines):
   - List third-party dependencies requiring attribution
   - Check go.mod for dependencies with specific license requirements

### File Changes

**New Files**:
- `LICENSE` (~20 lines)
- `SECURITY.md` (~30 lines)
- `NOTICE` (~10 lines, if needed)

**Modified Files**:
- None

### Verification Commands

```bash
# Verify LICENSE file exists
ls -la LICENSE

# Check file contents
cat LICENSE | grep "MIT License"
cat LICENSE | grep "2025"

# Verify SECURITY.md
cat SECURITY.md | grep "Reporting a Vulnerability"

# Check GitHub recognition (manual: visit repository page)
```

### Dependencies

None (foundation stage)

---

## Stage 2: Contributing Guidelines and Community Standards

### Objective

Create comprehensive contribution guidelines and code of conduct to establish clear community standards and development workflows.

### Acceptance Criteria

- [ ] CONTRIBUTING.md created with complete development workflow
- [ ] CODE_OF_CONDUCT.md created following Contributor Covenant 2.1
- [ ] README.md updated with "Contributing" section linking to CONTRIBUTING.md
- [ ] Development setup instructions clear and complete
- [ ] Code style and testing requirements documented
- [ ] Commit message and PR workflow defined
- [ ] Documentation readable in ≤5 minutes

### Implementation Tasks

1. **Create CONTRIBUTING.md** (~250 lines):
   ```markdown
   # Contributing to meta-cc

   ## Development Setup

   Prerequisites:
   - Go 1.21 or later
   - make
   - golangci-lint (optional but recommended)

   Setup:
   ```bash
   git clone https://github.com/[user]/meta-cc.git
   cd meta-cc
   make all  # Run lint, test, and build
   ```

   ## Code Style

   - Follow standard Go conventions
   - Run `make lint` before committing
   - Use `gofmt` for formatting

   ## Testing Requirements

   - Test coverage ≥80%
   - Run `make test` before submitting PR
   - Add tests for new features

   ## Commit Message Format

   Follow Conventional Commits:
   ```
   type(scope): description

   [optional body]

   [optional footer]
   ```

   Types: feat, fix, docs, test, refactor, chore

   ## Pull Request Process

   1. Fork the repository
   2. Create feature branch: `git checkout -b feature/my-feature`
   3. Make changes and test: `make all`
   4. Commit with clear message
   5. Push and create PR
   6. Wait for CI checks and review
   ```

2. **Create CODE_OF_CONDUCT.md** (~50 lines):
   - Use Contributor Covenant 2.1 template
   - Add contact information for enforcement
   - Define expected behavior and unacceptable conduct

3. **Update README.md** (~10 lines addition):
   ```markdown
   ## Contributing

   We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for:
   - Development setup
   - Code style guidelines
   - Testing requirements
   - Pull request process

   Please also read our [Code of Conduct](CODE_OF_CONDUCT.md).
   ```

### File Changes

**New Files**:
- `CONTRIBUTING.md` (~250 lines)
- `CODE_OF_CONDUCT.md` (~50 lines)

**Modified Files**:
- `README.md` (~10 lines added)

### Verification Commands

```bash
# Verify files exist
ls -la CONTRIBUTING.md CODE_OF_CONDUCT.md

# Check content structure
grep "Development Setup" CONTRIBUTING.md
grep "Testing Requirements" CONTRIBUTING.md
grep "Pull Request Process" CONTRIBUTING.md
grep "Contributor Covenant" CODE_OF_CONDUCT.md

# Verify README links
grep "CONTRIBUTING.md" README.md
grep "CODE_OF_CONDUCT.md" README.md

# Test documentation readability (manual review)
```

### Dependencies

None (independent of Stage 1)

---

## Stage 3: GitHub Templates and Configuration

### Objective

Create Issue and Pull Request templates to standardize community contributions and bug reports.

### Acceptance Criteria

- [ ] Bug report template functional
- [ ] Feature request template functional
- [ ] Pull request template displays checklist
- [ ] Templates use GitHub YAML format
- [ ] Required fields properly marked
- [ ] Issue creation loads templates automatically
- [ ] PR creation shows template automatically

### Implementation Tasks

1. **Create .github/ISSUE_TEMPLATE/bug_report.yml** (~70 lines):
   ```yaml
   name: Bug Report
   description: Report a bug or unexpected behavior
   labels: ["bug"]
   body:
     - type: markdown
       attributes:
         value: |
           Thanks for reporting a bug! Please fill out the sections below.

     - type: input
       id: version
       attributes:
         label: Version
         description: meta-cc version (run `meta-cc --version`)
         placeholder: "v0.11.1"
       validations:
         required: true

     - type: textarea
       id: description
       attributes:
         label: Bug Description
         description: Clear description of what went wrong
       validations:
         required: true

     - type: textarea
       id: reproduction
       attributes:
         label: Steps to Reproduce
         description: Step-by-step instructions
         placeholder: |
           1. Run command '...'
           2. Observe error '...'
       validations:
         required: true

     - type: textarea
       id: expected
       attributes:
         label: Expected Behavior
         description: What should have happened?
       validations:
         required: true

     - type: textarea
       id: environment
       attributes:
         label: Environment
         description: OS, Go version, etc.
         placeholder: |
           - OS: Ubuntu 22.04
           - Go: 1.21
       validations:
         required: true
   ```

2. **Create .github/ISSUE_TEMPLATE/feature_request.yml** (~50 lines):
   ```yaml
   name: Feature Request
   description: Suggest a new feature or enhancement
   labels: ["enhancement"]
   body:
     - type: textarea
       id: description
       attributes:
         label: Feature Description
         description: What would you like to see added?
       validations:
         required: true

     - type: textarea
       id: use-case
       attributes:
         label: Use Case
         description: Why is this feature needed?
       validations:
         required: true

     - type: textarea
       id: alternatives
       attributes:
         label: Alternatives Considered
         description: Other approaches you've thought about
       validations:
         required: false
   ```

3. **Create .github/ISSUE_TEMPLATE/config.yml** (~10 lines):
   ```yaml
   blank_issues_enabled: false
   contact_links:
     - name: Question or Discussion
       url: https://github.com/[user]/meta-cc/discussions
       about: Ask questions or discuss ideas
   ```

4. **Create .github/PULL_REQUEST_TEMPLATE.md** (~40 lines):
   ```markdown
   ## Description

   <!-- Briefly describe your changes -->

   ## Related Issue

   <!-- Link to related issue (e.g., Fixes #123) -->

   ## Changes Made

   - [ ] List key changes

   ## Testing Checklist

   - [ ] `make test` passes
   - [ ] `make lint` passes
   - [ ] Added/updated tests for new functionality
   - [ ] Test coverage maintained ≥80%

   ## Documentation

   - [ ] Updated relevant documentation
   - [ ] Added code comments where needed
   - [ ] Updated CHANGELOG.md (if applicable)

   ## Additional Notes

   <!-- Any other context or information -->
   ```

### File Changes

**New Files**:
- `.github/ISSUE_TEMPLATE/bug_report.yml` (~70 lines)
- `.github/ISSUE_TEMPLATE/feature_request.yml` (~50 lines)
- `.github/ISSUE_TEMPLATE/config.yml` (~10 lines)
- `.github/PULL_REQUEST_TEMPLATE.md` (~40 lines)

**Modified Files**:
- None

### Verification Commands

```bash
# Verify directory structure
ls -la .github/ISSUE_TEMPLATE/
ls -la .github/PULL_REQUEST_TEMPLATE.md

# Check YAML syntax
cat .github/ISSUE_TEMPLATE/bug_report.yml | grep "name:"
cat .github/ISSUE_TEMPLATE/feature_request.yml | grep "name:"

# Verify PR template
cat .github/PULL_REQUEST_TEMPLATE.md | grep "Testing Checklist"

# Manual verification: Create test issue/PR on GitHub
```

### Dependencies

None (independent stage)

---

## Stage 4: CI/CD Pipeline Implementation

### Objective

Implement GitHub Actions workflows for automated testing, linting, and release builds across multiple platforms.

### Acceptance Criteria

- [ ] CI workflow runs on every push and PR
- [ ] Tests execute on Linux, macOS, Windows
- [ ] Tests run against Go 1.21 and 1.22
- [ ] golangci-lint enforced
- [ ] Test coverage uploaded to Codecov
- [ ] Release workflow triggers on version tags
- [ ] Cross-platform binaries built automatically (5 platforms)
- [ ] GitHub Release created with binaries and release notes
- [ ] All workflow files pass YAML validation

### Implementation Tasks

1. **Create .github/workflows/ci.yml** (~120 lines):
   ```yaml
   name: CI

   on:
     push:
       branches: [ main, develop ]
     pull_request:
       branches: [ main, develop ]

   jobs:
     test:
       name: Test
       strategy:
         matrix:
           os: [ubuntu-latest, macos-latest, windows-latest]
           go: ['1.21', '1.22']
       runs-on: ${{ matrix.os }}

       steps:
         - name: Checkout code
           uses: actions/checkout@v4

         - name: Setup Go
           uses: actions/setup-go@v4
           with:
             go-version: ${{ matrix.go }}

         - name: Cache Go modules
           uses: actions/cache@v3
           with:
             path: ~/go/pkg/mod
             key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
             restore-keys: |
               ${{ runner.os }}-go-

         - name: Install dependencies
           run: go mod download

         - name: Run tests
           run: make test

         - name: Run linter
           if: matrix.os == 'ubuntu-latest' && matrix.go == '1.22'
           run: make lint

         - name: Upload coverage
           if: matrix.os == 'ubuntu-latest' && matrix.go == '1.22'
           uses: codecov/codecov-action@v3
           with:
             files: ./coverage.out
             flags: unittests
             name: codecov-umbrella

     lint:
       name: Lint
       runs-on: ubuntu-latest
       steps:
         - name: Checkout code
           uses: actions/checkout@v4

         - name: Setup Go
           uses: actions/setup-go@v4
           with:
             go-version: '1.22'

         - name: Run golangci-lint
           uses: golangci/golangci-lint-action@v3
           with:
             version: latest
             args: --timeout=5m
   ```

2. **Create .github/workflows/release.yml** (~150 lines):
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
           with:
             fetch-depth: 0

         - name: Setup Go
           uses: actions/setup-go@v4
           with:
             go-version: '1.22'

         - name: Get version
           id: version
           run: echo "VERSION=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT

         - name: Build binaries
           run: |
             mkdir -p build

             # Linux amd64
             GOOS=linux GOARCH=amd64 go build -o build/meta-cc-linux-amd64 ./cmd/meta-cc

             # Linux arm64
             GOOS=linux GOARCH=arm64 go build -o build/meta-cc-linux-arm64 ./cmd/meta-cc

             # macOS amd64
             GOOS=darwin GOARCH=amd64 go build -o build/meta-cc-darwin-amd64 ./cmd/meta-cc

             # macOS arm64
             GOOS=darwin GOARCH=arm64 go build -o build/meta-cc-darwin-arm64 ./cmd/meta-cc

             # Windows amd64
             GOOS=windows GOARCH=amd64 go build -o build/meta-cc-windows-amd64.exe ./cmd/meta-cc

             # MCP server binaries
             GOOS=linux GOARCH=amd64 go build -o build/meta-cc-mcp-linux-amd64 ./cmd/mcp-server
             GOOS=darwin GOARCH=amd64 go build -o build/meta-cc-mcp-darwin-amd64 ./cmd/mcp-server
             GOOS=darwin GOARCH=arm64 go build -o build/meta-cc-mcp-darwin-arm64 ./cmd/mcp-server
             GOOS=windows GOARCH=amd64 go build -o build/meta-cc-mcp-windows-amd64.exe ./cmd/mcp-server

         - name: Generate checksums
           run: |
             cd build
             sha256sum * > checksums.txt

         - name: Create Release
           uses: softprops/action-gh-release@v1
           with:
             files: build/*
             generate_release_notes: true
             draft: false
             prerelease: ${{ contains(steps.version.outputs.VERSION, '-') }}
           env:
             GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
   ```

3. **Update Makefile** (~30 lines addition):
   ```makefile
   # Cross-compile target for local testing
   .PHONY: cross-compile
   cross-compile:
   	@echo "Building cross-platform binaries..."
   	@mkdir -p build
   	GOOS=linux GOARCH=amd64 go build -o build/meta-cc-linux-amd64 ./cmd/meta-cc
   	GOOS=linux GOARCH=arm64 go build -o build/meta-cc-linux-arm64 ./cmd/meta-cc
   	GOOS=darwin GOARCH=amd64 go build -o build/meta-cc-darwin-amd64 ./cmd/meta-cc
   	GOOS=darwin GOARCH=arm64 go build -o build/meta-cc-darwin-arm64 ./cmd/meta-cc
   	GOOS=windows GOARCH=amd64 go build -o build/meta-cc-windows-amd64.exe ./cmd/meta-cc
   	@echo "Binaries built in build/"

   # Test coverage with output
   .PHONY: test-coverage
   test-coverage:
   	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
   	go tool cover -html=coverage.out -o coverage.html
   ```

### File Changes

**New Files**:
- `.github/workflows/ci.yml` (~120 lines)
- `.github/workflows/release.yml` (~150 lines)

**Modified Files**:
- `Makefile` (~30 lines added)

### Verification Commands

```bash
# Validate YAML syntax
yamllint .github/workflows/ci.yml
yamllint .github/workflows/release.yml

# Test Makefile targets locally
make cross-compile
ls -la build/

# Verify binaries built
file build/meta-cc-linux-amd64
file build/meta-cc-darwin-arm64
file build/meta-cc-windows-amd64.exe

# Test coverage generation
make test-coverage
ls -la coverage.out coverage.html

# Manual verification: Push to GitHub and check Actions tab
```

### Dependencies

- Stage 2 (CONTRIBUTING.md defines testing requirements)

---

## Stage 5: Release Automation Scripting

### Objective

Create release automation scripts and establish semantic versioning strategy for streamlined version management.

### Acceptance Criteria

- [ ] scripts/release.sh created and executable
- [ ] Script validates current branch (main/develop)
- [ ] Script runs full test suite before release
- [ ] Script prompts for CHANGELOG.md update
- [ ] Script creates git tag
- [ ] Script pushes tag to remote (triggering GitHub Actions)
- [ ] CHANGELOG.md updated with release format guidelines
- [ ] Semantic versioning strategy documented
- [ ] Release process tested end-to-end

### Implementation Tasks

1. **Create scripts/release.sh** (~120 lines):
   ```bash
   #!/bin/bash
   set -e

   # Usage: ./scripts/release.sh v1.0.0

   VERSION=$1

   if [ -z "$VERSION" ]; then
       echo "Error: Version required"
       echo "Usage: ./scripts/release.sh v1.0.0"
       exit 1
   fi

   # Validate version format
   if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
       echo "Error: Invalid version format. Use v1.0.0 or v1.0.0-beta"
       exit 1
   fi

   # Check current branch
   BRANCH=$(git rev-parse --abbrev-ref HEAD)
   if [[ "$BRANCH" != "main" && "$BRANCH" != "develop" ]]; then
       echo "Error: Must be on main or develop branch (current: $BRANCH)"
       exit 1
   fi

   # Check working directory clean
   if [ -n "$(git status --porcelain)" ]; then
       echo "Error: Working directory not clean. Commit or stash changes."
       exit 1
   fi

   echo "=== Release $VERSION ==="
   echo ""

   # Run full test suite
   echo "Running tests..."
   make all
   echo "✓ Tests passed"
   echo ""

   # Prompt for CHANGELOG update
   echo "Please update CHANGELOG.md with release notes for $VERSION"
   echo "Press Enter when ready to continue, or Ctrl+C to abort..."
   read

   # Verify CHANGELOG was updated
   if ! grep -q "$VERSION" CHANGELOG.md; then
       echo "Warning: $VERSION not found in CHANGELOG.md"
       echo "Continue anyway? (y/N)"
       read -r response
       if [[ ! "$response" =~ ^[Yy]$ ]]; then
           echo "Aborted"
           exit 1
       fi
   fi

   # Create tag
   echo "Creating tag $VERSION..."
   git tag -a "$VERSION" -m "Release $VERSION"
   echo "✓ Tag created"
   echo ""

   # Push tag
   echo "Pushing tag to remote..."
   git push origin "$VERSION"
   echo "✓ Tag pushed"
   echo ""

   echo "=== Release $VERSION Complete ==="
   echo ""
   echo "GitHub Actions will now:"
   echo "  1. Build cross-platform binaries"
   echo "  2. Create GitHub Release"
   echo "  3. Upload binaries"
   echo ""
   echo "Monitor progress: https://github.com/[user]/meta-cc/actions"
   ```

2. **Update CHANGELOG.md** (~50 lines additions):
   ```markdown
   # Changelog

   All notable changes to the meta-cc project will be documented in this file.

   The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
   and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

   ## [Unreleased]

   ### Added
   - Feature description

   ### Changed
   - Change description

   ### Fixed
   - Bug fix description

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

   ## [v0.11.1-formalization] - 2025-10-03

   [... existing changelog content ...]
   ```

3. **Create docs/release-process.md** (~30 lines):
   ```markdown
   # Release Process

   ## Prerequisites

   - Maintainer with push access to main branch
   - All tests passing (`make all`)
   - CHANGELOG.md updated with release notes

   ## Steps

   1. **Prepare Release**:
      ```bash
      # Ensure on main branch
      git checkout main
      git pull origin main

      # Update CHANGELOG.md
      vim CHANGELOG.md
      git commit -am "docs: update CHANGELOG for vX.Y.Z"
      git push
      ```

   2. **Execute Release**:
      ```bash
      ./scripts/release.sh v1.0.0
      ```

      This will:
      - Run full test suite
      - Prompt for CHANGELOG verification
      - Create and push git tag
      - Trigger GitHub Actions release workflow

   3. **Verify Release**:
      - Check [GitHub Actions](https://github.com/[user]/meta-cc/actions)
      - Verify binaries on [Releases page](https://github.com/[user]/meta-cc/releases)
      - Test download and installation

   ## Troubleshooting

   - **Build fails**: Check GitHub Actions logs
   - **Tag already exists**: Delete tag (`git tag -d vX.Y.Z && git push --delete origin vX.Y.Z`)
   - **Binary missing**: Check release.yml workflow configuration
   ```

### File Changes

**New Files**:
- `scripts/release.sh` (~120 lines)
- `docs/release-process.md` (~30 lines)

**Modified Files**:
- `CHANGELOG.md` (~50 lines added)

### Verification Commands

```bash
# Make script executable
chmod +x scripts/release.sh

# Test script validation (dry run)
./scripts/release.sh
# Expected: Error message about version required

./scripts/release.sh invalid-version
# Expected: Error message about version format

# Test CHANGELOG format
grep "Semantic Versioning" CHANGELOG.md
grep "Release Process" CHANGELOG.md

# Manual verification: Run release script on test tag
# git checkout -b test-release
# ./scripts/release.sh v0.99.0-test
```

### Dependencies

- Stage 4 (CI/CD workflows must exist for release.yml to trigger)

---

## Stage 6: Documentation Enhancement

### Objective

Enhance README.md with badges, improve installation instructions, and add visual elements for professional presentation.

### Acceptance Criteria

- [ ] README.md includes CI badge
- [ ] README.md includes coverage badge
- [ ] README.md includes license badge
- [ ] README.md includes release version badge
- [ ] Installation section updated with binary download instructions
- [ ] Download links for all 5 platforms provided
- [ ] All badges clickable and functional
- [ ] README renders correctly on GitHub
- [ ] Documentation meets professional open-source standards

### Implementation Tasks

1. **Update README.md badges section** (~20 lines):
   ```markdown
   # meta-cc

   [![CI](https://github.com/[user]/meta-cc/actions/workflows/ci.yml/badge.svg)](https://github.com/[user]/meta-cc/actions)
   [![Coverage](https://codecov.io/gh/[user]/meta-cc/branch/main/graph/badge.svg)](https://codecov.io/gh/[user]/meta-cc)
   [![License](https://img.shields.io/github/license/[user]/meta-cc)](LICENSE)
   [![Release](https://img.shields.io/github/v/release/[user]/meta-cc)](https://github.com/[user]/meta-cc/releases)
   [![Go Version](https://img.shields.io/github/go-mod/go-version/[user]/meta-cc)](go.mod)

   Meta-Cognition tool for Claude Code - analyze session history for workflow optimization.
   ```

2. **Update installation section** (~60 lines):
   ```markdown
   ## Installation

   ### Option 1: Download Pre-compiled Binary (Recommended)

   Download the latest release for your platform:

   #### Linux (x86_64)
   ```bash
   curl -L https://github.com/[user]/meta-cc/releases/latest/download/meta-cc-linux-amd64 -o meta-cc
   chmod +x meta-cc
   sudo mv meta-cc /usr/local/bin/
   ```

   #### Linux (ARM64)
   ```bash
   curl -L https://github.com/[user]/meta-cc/releases/latest/download/meta-cc-linux-arm64 -o meta-cc
   chmod +x meta-cc
   sudo mv meta-cc /usr/local/bin/
   ```

   #### macOS (Intel)
   ```bash
   curl -L https://github.com/[user]/meta-cc/releases/latest/download/meta-cc-darwin-amd64 -o meta-cc
   chmod +x meta-cc
   sudo mv meta-cc /usr/local/bin/
   ```

   #### macOS (Apple Silicon)
   ```bash
   curl -L https://github.com/[user]/meta-cc/releases/latest/download/meta-cc-darwin-arm64 -o meta-cc
   chmod +x meta-cc
   sudo mv meta-cc /usr/local/bin/
   ```

   #### Windows (PowerShell)
   ```powershell
   Invoke-WebRequest -Uri "https://github.com/[user]/meta-cc/releases/latest/download/meta-cc-windows-amd64.exe" -OutFile "meta-cc.exe"
   # Move to a directory in your PATH
   ```

   [View all releases →](https://github.com/[user]/meta-cc/releases)

   ### Option 2: Install from Source

   ```bash
   git clone https://github.com/[user]/meta-cc.git
   cd meta-cc
   make install
   ```

   Requirements:
   - Go 1.21 or later
   - make
   ```

3. **Add verification section** (~10 lines):
   ```markdown
   ### Verify Installation

   ```bash
   meta-cc --version
   meta-cc --help
   ```

   Expected output:
   ```
   meta-cc version v1.0.0
   ```
   ```

### File Changes

**New Files**:
- None

**Modified Files**:
- `README.md` (~90 lines modified/added)

### Verification Commands

```bash
# Verify badge syntax
grep "shields.io" README.md
grep "github.com/.*/actions/workflows/ci.yml" README.md

# Check installation instructions
grep "curl -L" README.md
grep "releases/latest/download" README.md

# Count platform installation sections
grep -c "####" README.md
# Expected: 5 (5 platforms)

# Render README locally (if using grip or similar)
grip README.md

# Manual verification: View README on GitHub after push
```

### Dependencies

- Stage 4 (CI/CD badges require workflows to exist)
- Stage 5 (release download links require release.yml)

---

## Stage 7: Repository Configuration and Final Setup

### Objective

Configure GitHub repository settings, branch protection, and finalize open-source infrastructure.

### Acceptance Criteria

- [ ] Repository description set
- [ ] Repository topics configured
- [ ] GitHub Actions enabled
- [ ] Branch protection rules active on main
- [ ] Required status checks configured
- [ ] Repository features configured (Issues, Discussions)
- [ ] All settings align with open-source best practices
- [ ] Phase 18 marked complete in docs/plan.md

### Implementation Tasks

1. **Repository Settings** (Manual configuration on GitHub):
   ```
   Description: "Meta-Cognition tool for Claude Code - analyze session history for workflow optimization"

   Website: https://github.com/[user]/meta-cc

   Topics:
   - go
   - claude-code
   - meta-cognition
   - cli
   - mcp
   - workflow-analysis
   - developer-tools
   - code-analysis

   Features:
   ✓ Issues
   ✓ Discussions (optional)
   ✗ Projects (not needed initially)
   ✗ Wiki (documentation in docs/)
   ```

2. **Branch Protection Rules** (main branch):
   ```
   Branch name pattern: main

   Protect matching branches:
   ✓ Require a pull request before merging
     ✓ Require approvals: 1
     ✓ Dismiss stale pull request approvals when new commits are pushed
   ✓ Require status checks to pass before merging
     ✓ Require branches to be up to date before merging
     Required status checks:
       - test (ubuntu-latest, 1.22)
       - lint
   ✓ Require conversation resolution before merging
   ✗ Require signed commits (optional)
   ✓ Require linear history
   ✓ Do not allow bypassing the above settings

   Rules applied to administrators:
   ✗ Allow force pushes (never)
   ✗ Allow deletions (never)
   ```

3. **Actions Permissions**:
   ```
   Settings → Actions → General

   Actions permissions:
   ○ Allow all actions and reusable workflows

   Workflow permissions:
   ○ Read and write permissions
   ✓ Allow GitHub Actions to create and approve pull requests
   ```

4. **Update docs/plan.md** (~10 lines):
   ```markdown
   ## Project Status

   **项目状态**：
   - ✅ **Phase 0-9 已完成**（核心查询 + 上下文管理）
   - ✅ **Phase 14 已完成**（架构重构 + MCP 独立可执行文件）
   - ✅ **Phase 15 已完成**（MCP 输出控制 + 工具标准化）
   - ✅ **Phase 16 已完成**（混合输出模式 + 无截断 + 可配置阈值）
   - ✅ **Phase 17 已完成**（Subagent 形式化实现）
   - ✅ **Phase 18 已完成**（GitHub Release 准备）
   ```

5. **Create docs/github-setup.md** (~40 lines):
   ```markdown
   # GitHub Repository Setup

   This document records the GitHub repository configuration for meta-cc.

   ## Repository Settings

   - **Description**: Meta-Cognition tool for Claude Code - analyze session history for workflow optimization
   - **Topics**: go, claude-code, meta-cognition, cli, mcp, workflow-analysis, developer-tools
   - **License**: MIT
   - **Features**: Issues ✓, Discussions (optional)

   ## Branch Protection (main)

   - Require PR with 1 approval
   - Require status checks: test, lint
   - Require branches up to date
   - Require conversation resolution
   - Linear history enforced
   - No force push, no deletion

   ## GitHub Actions

   - Workflow permissions: Read and write
   - Auto-approve PRs: Enabled

   ## Access Control

   - Admins cannot bypass branch protection
   - Require status checks for all users

   ## Release Process

   See [docs/release-process.md](release-process.md)
   ```

### File Changes

**New Files**:
- `docs/github-setup.md` (~40 lines)

**Modified Files**:
- `docs/plan.md` (~10 lines: mark Phase 18 complete)

### Verification Commands

```bash
# Verify documentation updates
grep "Phase 18 已完成" docs/plan.md

# Check GitHub setup doc
cat docs/github-setup.md | grep "Branch Protection"

# Manual verification checklist:
# 1. Visit GitHub repository settings
# 2. Verify description and topics
# 3. Check branch protection rules
# 4. Verify Actions permissions
# 5. Create test PR to verify status checks
# 6. Attempt force push to verify protection
```

### Dependencies

- Stage 4 (CI/CD workflows must exist for status checks)
- All previous stages (complete infrastructure before configuring repository)

---

## Phase-Level Integration

### Cross-Stage Integration Points

1. **Stage 1 → Stage 2**: LICENSE file referenced in CONTRIBUTING.md
2. **Stage 2 → Stage 3**: CONTRIBUTING.md referenced in PR template
3. **Stage 2 → Stage 4**: Testing requirements from CONTRIBUTING.md enforced by CI
4. **Stage 3 → Stage 7**: Issue/PR templates referenced in repository settings
5. **Stage 4 → Stage 5**: Release workflow triggered by release script
6. **Stage 4 → Stage 6**: CI badge requires workflow to exist
7. **Stage 5 → Stage 7**: Release process referenced in repository documentation

### Integration Test Scenarios

1. **Complete Release Flow**:
   ```bash
   # Scenario: Developer creates v1.0.0 release
   git checkout main
   ./scripts/release.sh v1.0.0
   # Expected: Tag created, GitHub Actions triggered, binaries uploaded
   ```

2. **PR Workflow**:
   ```bash
   # Scenario: External contributor submits PR
   git checkout -b feature/new-feature
   # Make changes
   git commit -m "feat: add new feature"
   git push origin feature/new-feature
   # Create PR on GitHub
   # Expected: CI runs, lint checks, PR template displayed
   ```

3. **Issue Creation**:
   ```
   # Scenario: User reports bug
   Navigate to GitHub Issues → New Issue → Bug Report
   # Expected: Template loaded with required fields
   ```

4. **Binary Download**:
   ```bash
   # Scenario: User downloads and installs binary
   curl -L https://github.com/[user]/meta-cc/releases/latest/download/meta-cc-linux-amd64 -o meta-cc
   chmod +x meta-cc
   ./meta-cc --version
   # Expected: Version displayed
   ```

---

## Testing Strategy

### Automated Testing

**Stage-Level Tests** (where applicable):
- Stage 4: GitHub Actions workflow validation (YAML lint)
- Stage 5: Release script validation (version format, branch checks)

**Integration Tests**:
- Manual verification of complete release flow
- Manual verification of CI/CD pipeline execution

### Manual Verification Checklist

**Stage 1**:
- [ ] GitHub displays MIT License
- [ ] SECURITY.md visible in repository

**Stage 2**:
- [ ] CONTRIBUTING.md renders correctly
- [ ] CODE_OF_CONDUCT.md follows Contributor Covenant

**Stage 3**:
- [ ] Issue templates load on issue creation
- [ ] PR template displays on PR creation

**Stage 4**:
- [ ] CI workflow runs on push
- [ ] CI workflow runs on PR
- [ ] Lint job passes
- [ ] Test jobs pass on all platforms
- [ ] Release workflow triggers on tag push
- [ ] Binaries uploaded to release

**Stage 5**:
- [ ] Release script validates input
- [ ] Release script runs tests
- [ ] Release script creates tag
- [ ] Release script pushes tag

**Stage 6**:
- [ ] All badges display correctly
- [ ] Badge links functional
- [ ] Installation instructions clear
- [ ] Download links work

**Stage 7**:
- [ ] Repository description set
- [ ] Topics searchable
- [ ] Branch protection enforces CI
- [ ] PR requires approval
- [ ] Force push blocked

---

## Risk Mitigation

### Potential Risks

1. **GitHub Actions quotas**: Public repos have limited free minutes
   - **Mitigation**: Optimize workflows, cache dependencies, limit matrix size

2. **Release workflow failures**: Build errors on specific platforms
   - **Mitigation**: Test cross-compile locally first, add detailed error logging

3. **Badge URLs incorrect**: Badges show 404 or incorrect status
   - **Mitigation**: Test badge URLs before committing, use GitHub username placeholder

4. **Branch protection too strict**: Blocks legitimate changes
   - **Mitigation**: Allow administrators to bypass temporarily if needed

5. **Binary size too large**: GitHub release asset limit is 2GB
   - **Mitigation**: Go binaries typically <50MB, well under limit

### Testing Failure Protocol

- If Stage verification fails → **FIX** before proceeding to next stage
- If CI workflow fails → **DEBUG** locally with `act` (GitHub Actions local runner)
- If release workflow fails → **ROLLBACK** tag, fix workflow, retry

---

## Success Metrics

### Functional Metrics

- [ ] LICENSE file present and recognized by GitHub
- [ ] CI/CD workflows operational (test + lint + release)
- [ ] Automated releases produce 5 platform binaries
- [ ] CONTRIBUTING.md complete with development workflow
- [ ] CODE_OF_CONDUCT.md present
- [ ] Issue/PR templates functional
- [ ] README badges all functional
- [ ] Repository settings configured
- [ ] Branch protection active

### Quality Metrics

- [ ] All CI checks pass on main branch
- [ ] Test coverage maintained ≥80%
- [ ] All linters pass
- [ ] Documentation professional and complete
- [ ] Release process tested end-to-end

### Completion Criteria

- [ ] `make all` passes
- [ ] GitHub repository fully configured
- [ ] Test release created successfully
- [ ] All documentation complete
- [ ] Phase 18 marked complete in docs/plan.md

---

## Code Change Summary

**Total Code Changes** (within ≤1,250 line limit):
- Stage 1: ~60 lines (LICENSE + SECURITY.md + NOTICE)
- Stage 2: ~310 lines (CONTRIBUTING.md + CODE_OF_CONDUCT.md + README update)
- Stage 3: ~170 lines (Issue/PR templates)
- Stage 4: ~300 lines (CI/CD workflows + Makefile updates)
- Stage 5: ~200 lines (release.sh + CHANGELOG updates + release-process.md)
- Stage 6: ~90 lines (README badges + installation updates)
- Stage 7: ~50 lines (github-setup.md + plan.md update)
- **Total: ~1,180 lines** (within limit)

**File Breakdown**:
- New files: 15
- Modified files: 4 (README.md, CHANGELOG.md, Makefile, docs/plan.md)

---

## Next Steps After Phase 18

**Immediate Actions**:
1. Create v1.0.0 release using new infrastructure
2. Monitor GitHub Actions for any workflow issues
3. Gather community feedback on contribution process
4. Add project to relevant directories (Awesome Claude Code, etc.)

**Phase 19 Candidates**:
- **GitHub Pages Documentation**: Auto-generated docs from godoc
- **Homebrew Tap**: Package manager distribution for macOS/Linux
- **Docker Images**: Containerized distribution
- **Performance Benchmarking**: Continuous performance monitoring in CI

---

**Plan Version**: 1.0
**Created**: 2025-10-07
**Estimated Effort**: 6.5 hours (as per docs/plan.md estimate)
**Dependencies**: None (can start immediately)
