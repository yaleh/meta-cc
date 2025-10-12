# Phase 18 Release Infrastructure Test Plan

**Purpose**: Validate end-to-end release infrastructure before v1.0.0

**Prerequisites**:
- ✅ Phase 18 stages 18.1-18.7 complete
- ✅ All files created and verified
- ✅ `make all` passing

---

## Test 1: CI Workflow Validation

### Objective
Verify GitHub Actions CI workflow runs successfully on pull requests

### Steps

1. **Create test branch and make trivial change**:
   ```bash
   cd /home/yale/work/meta-cc
   git checkout -b test-ci-workflow
   echo "# CI Test" > .github/CI_TEST.md
   git add .github/CI_TEST.md
   git commit -m "test: verify CI workflow triggers correctly"
   git push origin test-ci-workflow
   ```

2. **Create Pull Request**:
   - Navigate to GitHub repository
   - Click "Compare & pull request"
   - Title: "test: CI workflow validation"
   - Description: "Testing Phase 18 CI infrastructure"
   - Create PR

3. **Monitor CI Execution**:
   - Go to PR "Checks" tab
   - Verify workflow "CI" appears and runs
   - Expected jobs:
     * `test` (matrix: ubuntu/macos/windows × go 1.21/1.22)
     * `lint` (ubuntu × go 1.22)

4. **Verify Results**:
   - [ ] All test jobs pass (6 jobs total: 3 OS × 2 Go versions)
   - [ ] Lint job passes
   - [ ] Codecov report uploaded (if configured)
   - [ ] PR shows green checkmark

5. **Cleanup**:
   ```bash
   # Close PR without merging
   gh pr close <PR_NUMBER>

   # Delete test branch
   git checkout main
   git branch -D test-ci-workflow
   git push origin --delete test-ci-workflow

   # Remove test file
   rm .github/CI_TEST.md
   ```

### Expected Issues
- First CI run may require GitHub Actions to be enabled in repository settings
- golangci-lint may fail if not installed in CI environment (check .github/workflows/ci.yml)
- Codecov token may need to be configured in repository secrets

### Success Criteria
- ✅ CI workflow triggers automatically on PR
- ✅ All test matrices complete successfully
- ✅ Lint checks pass
- ✅ Workflow completes in <10 minutes

---

## Test 2: Release Workflow Validation

### Objective
Verify automated binary release works end-to-end

### Steps

1. **Ensure working directory is clean**:
   ```bash
   cd /home/yale/work/meta-cc
   git status
   # Should show "nothing to commit, working tree clean"
   ```

2. **Update CHANGELOG.md for test release**:
   ```bash
   cat >> CHANGELOG.md << 'EOF'

## [v0.99.0-test] - 2025-10-07

### Added
- Test release for Phase 18 infrastructure validation

EOF

   git add CHANGELOG.md
   git commit -m "docs: add v0.99.0-test to CHANGELOG"
   git push origin main
   ```

3. **Execute release script**:
   ```bash
   ./scripts/release.sh v0.99.0-test
   ```

   **Expected prompts**:
   - Script validates branch (should be on main)
   - Runs `make all` (should pass)
   - Prompts to verify CHANGELOG.md updated (press Enter)
   - Creates git tag `v0.99.0-test`
   - Pushes tag to remote

4. **Monitor Release Workflow**:
   ```bash
   # Open GitHub Actions page
   gh run list --workflow=release.yml

   # Or visit manually:
   # https://github.com/[user]/meta-cc/actions/workflows/release.yml
   ```

5. **Verify GitHub Release**:
   - Navigate to repository Releases page
   - Find "v0.99.0-test" release
   - Verify release notes auto-generated
   - Check assets uploaded:
     * [ ] meta-cc-linux-amd64
     * [ ] meta-cc-linux-arm64
     * [ ] meta-cc-darwin-amd64
     * [ ] meta-cc-darwin-arm64
     * [ ] meta-cc-windows-amd64.exe
     * [ ] meta-cc-mcp-linux-amd64
     * [ ] meta-cc-mcp-darwin-amd64
     * [ ] meta-cc-mcp-darwin-arm64
     * [ ] meta-cc-mcp-windows-amd64.exe
     * [ ] checksums.txt

6. **Test Binary Download and Execution**:
   ```bash
   # Download binary for current platform
   cd /tmp
   curl -L https://github.com/[user]/meta-cc/releases/download/v0.99.0-test/meta-cc-linux-amd64 -o meta-cc-test
   chmod +x meta-cc-test

   # Test execution
   ./meta-cc-test --version
   ./meta-cc-test --help

   # Expected: Version should be v0.99.0-test

   # Cleanup
   rm meta-cc-test
   cd /home/yale/work/meta-cc
   ```

7. **Cleanup Test Release**:
   ```bash
   # Delete GitHub release
   gh release delete v0.99.0-test --yes

   # Delete local tag
   git tag -d v0.99.0-test

   # Delete remote tag
   git push origin :refs/tags/v0.99.0-test

   # Remove CHANGELOG test entry
   git revert HEAD --no-edit
   git push origin main
   ```

### Expected Issues
- First release may need write permissions configured for GITHUB_TOKEN
- Binary builds may timeout on slow runners (increase timeout in workflow)
- Checksums generation may fail if sha256sum not available

### Success Criteria
- ✅ Release script completes without errors
- ✅ GitHub Release created automatically
- ✅ All 9 binaries uploaded successfully
- ✅ Downloaded binary executes and shows correct version
- ✅ Release workflow completes in <15 minutes

---

## Test 3: Documentation and Badge Verification

### Objective
Verify README badges, installation instructions, and documentation

### Steps

1. **Check README Badges**:
   - Open repository homepage on GitHub
   - Verify badges display:
     * [ ] CI badge (green = passing)
     * [ ] Coverage badge (shows percentage)
     * [ ] License badge (MIT)
     * [ ] Release badge (shows latest version)
     * [ ] Go Version badge (shows go.mod version)

2. **Click each badge** to verify links work:
   - CI badge → GitHub Actions CI workflow runs
   - Coverage badge → Codecov project page
   - License badge → LICENSE file
   - Release badge → GitHub Releases page
   - Go badge → go.mod file

3. **Test Installation Instructions**:

   **Linux x86_64**:
   ```bash
   # Copy command from README and test
   cd /tmp
   curl -L https://github.com/[user]/meta-cc/releases/latest/download/meta-cc-linux-amd64 -o meta-cc
   chmod +x meta-cc
   ./meta-cc --version
   rm meta-cc
   ```

   **macOS ARM64** (if on Mac):
   ```bash
   cd /tmp
   curl -L https://github.com/[user]/meta-cc/releases/latest/download/meta-cc-darwin-arm64 -o meta-cc
   chmod +x meta-cc
   ./meta-cc --version
   rm meta-cc
   ```

4. **Verify Documentation Links**:
   - README → CONTRIBUTING.md (works)
   - README → CODE_OF_CONDUCT.md (works)
   - README → SECURITY.md (works)
   - README → docs/ (all links work)

5. **Check Issue/PR Templates**:
   - Navigate to Issues → New Issue
   - Verify templates appear:
     * [ ] Bug Report
     * [ ] Feature Request
   - Create test issue using Bug Report template
   - Verify all fields load correctly
   - Close issue without submitting

   - Create test PR (from test branch)
   - Verify PR template loads with checklist
   - Close PR without merging

### Expected Issues
- Badges may show "unknown" until first workflow runs
- Coverage badge may not work until Codecov is configured
- Badge URLs may need GitHub username placeholder updated

### Success Criteria
- ✅ All badges display and link correctly
- ✅ Installation instructions work on tested platforms
- ✅ Documentation links functional
- ✅ Issue/PR templates load correctly

---

## Test 4: Cross-Platform Build Verification

### Objective
Verify `make cross-compile` works locally before release

### Steps

1. **Run cross-compile target**:
   ```bash
   cd /home/yale/work/meta-cc
   make cross-compile
   ```

2. **Verify binaries created**:
   ```bash
   ls -lh build/
   ```

   Expected files:
   - meta-cc-linux-amd64
   - meta-cc-linux-arm64
   - meta-cc-darwin-amd64
   - meta-cc-darwin-arm64
   - meta-cc-windows-amd64.exe

   All files should be >5MB

3. **Check file types**:
   ```bash
   file build/meta-cc-linux-amd64
   # Expected: ELF 64-bit LSB executable, x86-64

   file build/meta-cc-darwin-arm64
   # Expected: Mach-O 64-bit arm64 executable

   file build/meta-cc-windows-amd64.exe
   # Expected: PE32+ executable (console) x86-64
   ```

4. **Test local binary execution** (if on Linux amd64):
   ```bash
   ./build/meta-cc-linux-amd64 --version
   ./build/meta-cc-linux-amd64 --help
   ```

5. **Cleanup**:
   ```bash
   rm -rf build/
   ```

### Success Criteria
- ✅ All 5 binaries build successfully
- ✅ Binaries are correct architecture/OS
- ✅ Local binary executes correctly
- ✅ Build completes in <2 minutes

---

## Test 5: Branch Protection Verification

### Objective
Verify branch protection rules work as expected

### Steps

1. **Attempt direct push to main** (should fail):
   ```bash
   git checkout main
   echo "test" > test.txt
   git add test.txt
   git commit -m "test: branch protection"
   git push origin main
   ```

   **Expected**: Push rejected (branch protection enabled)

   If push succeeds, branch protection needs to be configured.

2. **Reset test commit**:
   ```bash
   git reset --hard HEAD~1
   rm test.txt
   ```

3. **Verify PR requirements**:
   - Create test PR
   - Attempt to merge before CI completes
   - **Expected**: Merge button disabled until CI passes
   - **Expected**: "1 approval required" shown

4. **Test force push protection**:
   ```bash
   git push --force origin main
   ```

   **Expected**: Force push rejected

### Success Criteria
- ✅ Direct push to main blocked
- ✅ CI checks required before merge
- ✅ PR approval required
- ✅ Force push blocked

---

## Overall Success Criteria

All tests must pass before proceeding to v1.0.0 release:

- [ ] Test 1: CI workflow working
- [ ] Test 2: Release workflow creating binaries
- [ ] Test 3: Documentation and badges correct
- [ ] Test 4: Cross-platform builds successful
- [ ] Test 5: Branch protection enforced

---

## Troubleshooting Guide

### Issue: CI workflow not triggering
**Solution**:
1. Check Settings → Actions → General
2. Enable "Allow all actions and reusable workflows"
3. Set workflow permissions to "Read and write permissions"

### Issue: Release workflow fails to upload binaries
**Solution**:
1. Check workflow permissions include "contents: write"
2. Verify GITHUB_TOKEN has correct scopes
3. Check release.yml uses `softprops/action-gh-release@v1`

### Issue: Badges show "unknown"
**Solution**:
1. Wait for first workflow run to complete
2. Update badge URLs with correct GitHub username
3. For coverage badge, configure Codecov token in repository secrets

### Issue: Binary download fails
**Solution**:
1. Check release is not marked as "Draft"
2. Verify binary uploaded to correct release tag
3. Check file permissions on uploaded assets

### Issue: golangci-lint fails in CI
**Solution**:
1. Check `.github/workflows/ci.yml` uses `golangci/golangci-lint-action@v3`
2. Verify `--timeout=5m` is set
3. Run `make lint` locally to reproduce issue

---

## Post-Test Actions

After all tests pass:

1. **Document Test Results**:
   - Create `plans/18/TEST_RESULTS.md` with outcomes
   - Note any issues encountered and solutions

2. **Update Documentation**:
   - Fix any badge URLs that were wrong
   - Correct installation instructions if needed
   - Update troubleshooting guide with new issues

3. **Prepare for v1.0.0**:
   - Review CHANGELOG.md for completeness
   - Update version numbers if needed
   - Plan release announcement

4. **Mark Phase 18 Verified**:
   - Update `plans/18/EXECUTION_STATUS.md`
   - Add "✅ Infrastructure Tested" status
   - Ready for production release

---

**Test Plan Version**: 1.0
**Created**: 2025-10-07
**Last Updated**: 2025-10-07
**Status**: Ready to Execute
