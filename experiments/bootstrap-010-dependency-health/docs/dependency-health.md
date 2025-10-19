# Dependency Health Management

Automated dependency health monitoring and management for Go projects.

**Pattern**: CI/CD Automation Integration (Pattern 5)
**Source**: Iteration 2 automation pattern, Iteration 3 implementation
**Status**: Implemented

---

## Overview

This dependency health management system provides:

1. **Automated vulnerability scanning** (govulncheck)
2. **License compliance checking** (go-licenses)
3. **Dependency freshness monitoring** (go list)
4. **Interactive update workflow** (scripts)
5. **CI/CD integration** (GitHub Actions)

**Benefits**:
- **Early detection**: Catch vulnerabilities before deployment
- **Continuous monitoring**: Weekly scheduled scans
- **Fast updates**: 6x speedup over manual process (1.5h vs 9h)
- **Safe updates**: Test-driven update workflow with rollback
- **Compliance**: Automated license policy enforcement

---

## CI/CD Automation

### GitHub Actions Workflow

Location: `.github/workflows/dependency-health.yml`

**Triggers**:
- Push to main branch
- Pull requests
- Weekly schedule (Monday 9am UTC)
- Manual dispatch

**Jobs**:

1. **security-scan**: Run govulncheck, fail on vulnerabilities
2. **license-compliance**: Check licenses, fail on prohibited licenses
3. **dependency-freshness**: Report outdated dependencies
4. **summary**: Generate combined health report

**Artifacts**:
- `vulnerability-report` (govulncheck output)
- `license-report` (licenses.csv)
- `dependency-freshness-report` (outdated dependencies)

Retention: 90 days

**Example Output**:

```
Dependency Health Summary

| Check                  | Status      |
|------------------------|-------------|
| Vulnerability Scan     | ✅ Pass     |
| License Compliance     | ✅ Pass     |
| Dependency Freshness   | ⚠️ Updates  |
```

---

## Automation Scripts

All scripts located in `scripts/` directory.

### 1. check-deps.sh

**Purpose**: Run all dependency health checks locally

**Usage**:
```bash
./scripts/check-deps.sh
```

**Checks Performed**:
1. Vulnerability scan (govulncheck)
2. License compliance (go-licenses)
3. Dependency freshness (go list -m -u all)
4. Go module tidy check
5. Optional: Test suite

**Output**:
```
======================================
Dependency Health Check
======================================

-----------------------------------
1. Vulnerability Scan (govulncheck)
-----------------------------------
✓ No vulnerabilities found

-----------------------------------
2. License Compliance (go-licenses)
-----------------------------------
✓ All 18 dependencies compliant

-----------------------------------
3. Dependency Freshness
-----------------------------------
⚠ 3 outdated dependencies

Outdated dependencies:
golang.org/x/sync v0.8.0 [v0.10.0]
golang.org/x/sys v0.26.0 [v0.28.0]
golang.org/x/term v0.25.0 [v0.27.0]

Run 'scripts/update-deps.sh' to update

-----------------------------------
4. Go Module Tidy Check
-----------------------------------
✓ go.mod and go.sum are tidy

======================================
Summary
======================================

Passed:   3
Warnings: 1
Failed:   0

Dependency health check passed with WARNINGS
```

**Exit Codes**:
- `0`: All checks passed (may have warnings)
- `1`: One or more checks failed

**Prerequisites**:
```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/google/go-licenses@latest
```

---

### 2. update-deps.sh

**Purpose**: Interactive dependency update with testing and rollback

**Usage**:
```bash
./scripts/update-deps.sh
```

**Workflow**:

1. **List outdated dependencies**
   ```
   Found 3 outdated dependencies:
   golang.org/x/sync v0.8.0 [v0.10.0]
   golang.org/x/sys v0.26.0 [v0.28.0]
   golang.org/x/term v0.25.0 [v0.27.0]
   ```

2. **Confirm update**
   ```
   Do you want to update all dependencies? (y/N): y
   ```

3. **Establish baseline**
   ```
   Baseline: 15 test packages passed
   ```

4. **Apply updates**
   ```
   Dependencies updated
   go.mod tidied
   ```

5. **Run tests**
   ```
   After update: 15 test packages passed
   ```

6. **Compare results**
   ```
   No regressions detected!
   Update successful
   ```

7. **Run vulnerability scan**
   ```
   No vulnerabilities found
   ```

**Features**:
- **Baseline testing**: Establishes test pass count before update
- **Regression detection**: Compares before/after test results
- **Automatic rollback**: Reverts on test failure (with confirmation)
- **Backup/restore**: Saves go.mod and go.sum before update
- **Vulnerability scan**: Runs govulncheck after update

**Rollback Criteria**:
- Test packages that passed before now fail
- User confirms rollback when prompted

**Example (with regression)**:
```
REGRESSION DETECTED!
Tests that passed before now fail after update

Rollback to previous versions? (Y/n): Y
Rolling back...
Rolled back to previous dependency versions
```

---

### 3. generate-licenses.sh

**Purpose**: Generate THIRD_PARTY_LICENSES file

**Usage**:
```bash
./scripts/generate-licenses.sh
```

**Output Files**:
1. `THIRD_PARTY_LICENSES` - Full license texts
2. `licenses.csv` - Summary CSV

**Example Output**:
```
======================================
Third-Party License Generator
======================================

Collecting licenses to /tmp/tmp.XXXXXXXXXX...
Licenses collected
Found 18 dependency licenses
Generating THIRD_PARTY_LICENSES...

Successfully generated THIRD_PARTY_LICENSES
File size: 245 KiB
Includes 18 dependency licenses

Generating licenses.csv...
Summary CSV generated

License distribution:
     10 Apache-2.0
      5 BSD-3-Clause
      2 MIT
      1 ISC

License generation complete!

Files created:
  - THIRD_PARTY_LICENSES (full license texts)
  - licenses.csv (summary CSV)
```

**Format (THIRD_PARTY_LICENSES)**:
```markdown
# Third-Party Licenses

This file contains the licenses of all third-party dependencies used in this project.

## Dependency Licenses

---

## 1. github.com/google/go-cmp

```
[Full license text]
```

---

## 2. golang.org/x/sync

```
[Full license text]
```

...
```

**Prerequisites**:
```bash
go install github.com/google/go-licenses@latest
```

---

## Dependency Update Workflow

**Recommended workflow** (manual updates):

### Weekly Review

1. **Check CI status**
   - Review GitHub Actions runs
   - Check for vulnerability alerts
   - Review outdated dependency warnings

2. **Run local health check**
   ```bash
   ./scripts/check-deps.sh
   ```

3. **If updates available, run interactive update**
   ```bash
   ./scripts/update-deps.sh
   ```

4. **Generate updated license file**
   ```bash
   ./scripts/generate-licenses.sh
   ```

5. **Commit changes**
   ```bash
   git add go.mod go.sum THIRD_PARTY_LICENSES licenses.csv
   git commit -m "chore: update dependencies and licenses"
   ```

### Emergency Vulnerability Response

1. **CI fails with vulnerability**
   - GitHub Actions posts comment on PR
   - Review govulncheck report artifact

2. **Assess severity**
   - HIGH/CRITICAL: Immediate action
   - MEDIUM/LOW: Schedule for next update cycle

3. **Apply targeted update**
   ```bash
   # Update specific dependency
   go get -u github.com/vulnerable/package@latest
   go mod tidy

   # Run tests
   go test ./...

   # Verify fix
   govulncheck ./...
   ```

4. **Commit fix**
   ```bash
   git add go.mod go.sum
   git commit -m "fix: patch vulnerability in github.com/vulnerable/package"
   ```

---

## License Policy

**Allowed Licenses**:
- Apache-2.0
- MIT
- BSD-2-Clause
- BSD-3-Clause
- ISC

**Review Required**:
- MPL-2.0 (Mozilla Public License)
- LGPL-2.1 (Lesser GPL)
- LGPL-3.0

**Prohibited Licenses**:
- GPL-2.0 (GNU General Public License)
- GPL-3.0
- AGPL-3.0 (Affero GPL)
- SSPL (Server Side Public License)
- Commons-Clause

**Policy Enforcement**:
- CI fails on prohibited licenses
- Manual review required for "Review Required" licenses

---

## Metrics

### Automation Effectiveness

**Before Automation** (manual process):
- Time: ~9 hours
- Tasks:
  - Manual govulncheck: 0.5h
  - Manual license check: 0.5h
  - Dependency updates: 6h
  - Testing/validation: 1.5h
  - Documentation: 0.5h

**After Automation** (CI + scripts):
- Time: ~1.5 hours (6x speedup)
- Tasks:
  - Review CI reports: 0.5h
  - Apply updates (scripts): 0.5h
  - Review test results: 0.25h
  - Merge and document: 0.25h

**Key Improvements**:
- **6x speedup**: 9h → 1.5h
- **100% scan coverage**: Every PR scanned
- **Faster detection**: Weekly scheduled scans (vs manual quarterly)
- **Safer updates**: Test-driven workflow with rollback

---

## Troubleshooting

### Tool Installation Issues

**govulncheck not found**:
```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
```

**go-licenses not found**:
```bash
go install github.com/google/go-licenses@latest
```

**Tools not in PATH**:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
# Add to ~/.bashrc or ~/.zshrc for persistence
```

### CI Failures

**Vulnerability scan failed**:
1. Download artifact: `vulnerability-report`
2. Review govulncheck output
3. Apply updates: `go get -u <package>@latest`
4. Re-run CI

**License compliance failed**:
1. Download artifact: `license-report`
2. Review licenses.csv
3. Remove or replace prohibited dependencies
4. Re-run CI

**Tests failing after update**:
1. Run `./scripts/update-deps.sh` locally
2. Script will detect regression and offer rollback
3. Investigate specific test failures
4. Apply targeted fixes

### Script Issues

**Permission denied**:
```bash
chmod +x scripts/*.sh
```

**go mod tidy diff warning**:
```bash
go mod tidy
git add go.mod go.sum
```

---

## References

### Patterns

- **Pattern 1**: Vulnerability Assessment (data/s1-vulnerability-analysis.yaml)
- **Pattern 3**: License Compliance (data/s1-license-compliance-report.yaml)
- **Pattern 5**: CI/CD Automation (data/iteration-2-automation-pattern.yaml)
- **Pattern 6**: Update Testing (data/iteration-2-testing-pattern.yaml)

### Principles

- **Security-First**: Patch HIGH/CRITICAL immediately
- **Batch Remediation**: Group related fixes (5x+ efficiency)
- **Test-Before-Update**: Baseline comparison prevents regressions
- **Policy-Driven**: Explicit license policy enables automation
- **Platform-Context**: Prioritize based on deployment context

### Tools

- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) - Go vulnerability scanner
- [go-licenses](https://github.com/google/go-licenses) - License compliance checker
- [Dependabot](https://docs.github.com/en/code-security/dependabot) - Automated dependency updates (optional)

---

## Next Steps

### Optional Enhancements

1. **Dependabot Integration**
   - Create `.github/dependabot.yml`
   - Configure auto-merge for patch updates
   - Require manual review for minor/major

2. **Notification Integration**
   - Slack/Teams alerts on CI failures
   - GitHub issues for HIGH severity vulnerabilities
   - Weekly summary reports

3. **Dashboard**
   - Dependency health metrics visualization
   - Trend analysis (vulnerability count over time)
   - Update lag metrics

---

**Status**: Implemented (Iteration 3)
**Version**: 1.0
**Last Updated**: 2025-10-17
