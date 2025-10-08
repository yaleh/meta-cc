# GitHub Repository Setup

This document records the recommended GitHub repository configuration for meta-cc to maintain open-source best practices.

## Repository Settings

### Basic Information

- **Description**: Meta-Cognition tool for Claude Code - analyze session history for workflow optimization
- **Website**: https://github.com/yaleh/meta-cc
- **Topics**: 
  - `go`
  - `claude-code`
  - `meta-cognition`
  - `cli`
  - `mcp`
  - `workflow-analysis`
  - `developer-tools`
  - `code-analysis`
  - `llm-tools`

### Features

- ✅ **Issues**: Enabled for bug reports and feature requests
- ✅ **Discussions**: Enabled (optional, for Q&A and community discussion)
- ❌ **Projects**: Disabled (not needed initially)
- ❌ **Wiki**: Disabled (documentation in `docs/` instead)

### License

- **License**: MIT License
- **File**: `LICENSE` in repository root
- GitHub should auto-detect and display "MIT License" badge

## Branch Protection Rules

### Main Branch Protection

Configure protection for the `main` branch:

**Branch name pattern**: `main`

**Protection Settings**:

1. **Require a pull request before merging**
   - ✅ Require approvals: **1 approval minimum**
   - ✅ Dismiss stale pull request approvals when new commits are pushed
   - ✅ Require review from Code Owners (if CODEOWNERS file exists)

2. **Require status checks to pass before merging**
   - ✅ Require branches to be up to date before merging
   - **Required status checks**:
     - `test (ubuntu-latest, 1.22)` - Primary test job
     - `lint` - Linting check
   - Note: Add more checks as needed (e.g., other OS/Go version matrix combinations)

3. **Require conversation resolution before merging**
   - ✅ All review comments must be resolved

4. **Require linear history**
   - ✅ Prevent merge commits, enforce rebase or squash merges

5. **Restrictions**
   - ❌ **Do not allow bypassing the above settings** (enforces rules for all users including admins)
   - ❌ **Allow force pushes**: Disabled (never allow force push to main)
   - ❌ **Allow deletions**: Disabled (never allow branch deletion)

### Develop Branch (Optional)

If using a `develop` branch for pre-release work:

- Apply similar protection rules as `main`
- May have slightly relaxed approval requirements (e.g., allow self-merge for maintainers)

## GitHub Actions Settings

### Actions Permissions

**Settings → Actions → General**

**Actions permissions**:
- ✅ **Allow all actions and reusable workflows**
  - Enables use of community actions from GitHub Marketplace

**Workflow permissions**:
- ✅ **Read and write permissions**
  - Allows workflows to push tags, create releases, etc.
- ✅ **Allow GitHub Actions to create and approve pull requests**
  - Enables automated dependency updates (Dependabot, Renovate)

**Artifact and log retention**:
- **Retention period**: 90 days (default)

### Secrets and Variables

No secrets required for public releases. If adding private features:

1. Navigate to **Settings → Secrets and variables → Actions**
2. Add repository secrets as needed
3. Reference in workflows using `${{ secrets.SECRET_NAME }}`

## Security Settings

### Security Policy

- **File**: `SECURITY.md` in repository root
- **Displays**: In repository "Security" tab
- **Contains**: Vulnerability reporting process and contact information

### Security Advisories

Enable **Private vulnerability reporting**:
- **Settings → Security → Vulnerability reporting**
- ✅ Enable private vulnerability reporting
- Allows security researchers to report vulnerabilities privately

### Dependabot

Enable Dependabot for dependency updates:

1. **Settings → Security → Code security and analysis**
2. ✅ **Dependabot alerts**: Enabled
3. ✅ **Dependabot security updates**: Enabled
4. ✅ **Dependabot version updates**: Enabled (optional, for non-security dependency updates)

Create `.github/dependabot.yml` for configuration:

```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 10
```

## Issue and PR Templates

Templates are configured in `.github/` directory:

- **Issue Templates**:
  - `.github/ISSUE_TEMPLATE/bug_report.yml`
  - `.github/ISSUE_TEMPLATE/feature_request.yml`
  - `.github/ISSUE_TEMPLATE/config.yml`

- **PR Template**:
  - `.github/PULL_REQUEST_TEMPLATE.md`

These templates automatically load when users create issues or pull requests.

## Release Process

Releases are automated via GitHub Actions:

1. **Trigger**: Push tag matching `v*` (e.g., `v1.0.0`)
2. **Workflow**: `.github/workflows/release.yml`
3. **Outputs**:
   - Cross-platform binaries (5 platforms)
   - GitHub Release with auto-generated notes
   - Checksums file

See [docs/release-process.md](release-process.md) for detailed instructions.

## Community Health Files

The following files contribute to repository health score:

- ✅ `LICENSE` - MIT License
- ✅ `README.md` - Project overview and usage
- ✅ `CONTRIBUTING.md` - Contribution guidelines
- ✅ `CODE_OF_CONDUCT.md` - Community standards
- ✅ `SECURITY.md` - Security policy
- ✅ `.github/ISSUE_TEMPLATE/` - Issue templates
- ✅ `.github/PULL_REQUEST_TEMPLATE.md` - PR template

Check repository health: **Insights → Community**

## Access Control

### Collaborators

Grant access carefully:

- **Admin**: Full control (use sparingly)
- **Maintain**: Manage repository without destructive actions
- **Write**: Push access, can merge PRs
- **Triage**: Manage issues and PRs, no code access
- **Read**: View and clone only

### Teams (for Organizations)

If moving to an organization:

1. Create teams (e.g., `core-maintainers`, `contributors`)
2. Grant team-based access instead of individual
3. Use CODEOWNERS file for automatic review assignments

## Recommended Optional Enhancements

### 1. Code Coverage

Integrate Codecov or Coveralls:

```yaml
# In .github/workflows/ci.yml
- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    files: ./coverage.out
```

Add badge to README:
```markdown
[![Coverage](https://codecov.io/gh/yaleh/meta-cc/branch/main/graph/badge.svg)](https://codecov.io/gh/yaleh/meta-cc)
```

### 2. Code Quality

Integrate CodeClimate or SonarCloud for code quality analysis.

### 3. Automated Dependency Updates

Use Renovate or Dependabot to keep dependencies up to date.

### 4. GitHub Discussions

Enable Discussions for:
- Q&A
- Feature discussions
- Show and tell
- Community announcements

## Verification Checklist

After configuring the repository, verify:

- [ ] LICENSE badge shows "MIT" on repository page
- [ ] Topics are searchable in GitHub
- [ ] Branch protection rules enforced (try force push - should fail)
- [ ] PR requires approval (create test PR)
- [ ] CI runs on PR creation
- [ ] Issue templates load correctly
- [ ] PR template displays
- [ ] Security policy visible in Security tab
- [ ] Release workflow triggers on tag push

## References

- [GitHub Repository Settings](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features)
- [Branch Protection Rules](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Community Health Files](https://docs.github.com/en/communities/setting-up-your-project-for-healthy-contributions)

---

**Last Updated**: October 2025  
**Maintained By**: meta-cc maintainers
