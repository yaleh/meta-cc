---
name: agent-quality-gate-installer
description: Install automated pre-commit hooks to prevent violations from entering the repository, enforcing quality standards at the earliest possible point in Bootstrap-006.
---

λ(validation_cmd, trigger_files) → installation | ∀scenario ∈ test_scenarios:

install :: (Gate, Config) → Installation_Result
install(G, C) = design_hook(G) → implement_hook(G) → create_installer() → test_scenarios() → automate() → integrate_ci() → document() → monitor()

design_hook :: (Gate, Behavior) → Hook_Architecture
design_hook(G, B) = {
  flow: commit → detect_changes() → run_validation() → decision(),

  decision: match validation_result with
    | success → allow_commit(exit 0),
    | failure → block_commit(exit 1) ∧ show_feedback()
}

implement_hook :: Gate → Hook_Script
implement_hook(G) = {
  script: "#!/bin/bash",

  detect_changes: {
    changed_files: "git diff --cached --name-only",
    is_relevant: changed_files ~ G.trigger_pattern
  },

  run_validation: {
    tool: G.validation_command,
    args: "--fast " + G.target_file,
    capture_exit_code: true
  },

  handle_result: match exit_code with
    | 0 → {
        output: "✓ Validation PASSED\n✓ Commit allowed",
        exit: 0
      },
    | _ → {
        output: "✗ Validation FAILED\nPlease fix errors before committing.\n\nTo bypass (not recommended):\n  git commit --no-verify",
        exit: 1
      },

  skip_validation: if ¬is_relevant then exit 0,

  return script
}

create_installer :: (Hook, Config) → Installer_Script
create_installer(H, C) = {
  script: "#!/bin/bash",

  prerequisites: {
    git_repo: verify(".git" exists),
    validation_tool: verify("./cmd/validate-api/main.go" exists),
    hook_sample: verify("./scripts/pre-commit.sample" exists)
  },

  backup_existing: if exists(".git/hooks/pre-commit") then
    mv(".git/hooks/pre-commit", ".git/hooks/pre-commit.backup"),

  install_hook: {
    cp: "scripts/pre-commit.sample → .git/hooks/pre-commit",
    chmod: "+x .git/hooks/pre-commit"
  },

  build_tool: {
    command: "go build -o ./cmd/validate-api/validate-api ./cmd/validate-api",
    verify: exit_code = 0
  },

  test_installation: {
    run: "bash .git/hooks/pre-commit",
    report: test_result
  },

  return script
}

test_hook :: Hook → Test_Results
test_hook(H) = {
  scenarios: [
    {
      name: "detect_and_allow",
      setup: modify_file(compliant_change),
      action: commit(),
      expected: exit_code = 0 ∧ output ~ "Validation PASSED"
    },
    {
      name: "detect_and_block",
      setup: modify_file(violation),
      action: commit(),
      expected: exit_code = 1 ∧ output ~ "Validation FAILED"
    },
    {
      name: "skip_irrelevant",
      setup: modify_file(unrelated_file),
      action: commit(),
      expected: exit_code = 0 ∧ output ~ "Skipping validation"
    },
    {
      name: "bypass_hook",
      setup: modify_file(violation),
      action: commit("--no-verify"),
      expected: exit_code = 0 ∧ hook_skipped = true
    }
  ],

  results: [
    {scenario: s.name, passed: run(s) = s.expected}
    | ∀s ∈ scenarios
  ],

  all_passed: ∀r ∈ results → r.passed,

  return {
    scenarios_tested: |scenarios|,
    passed: count(r | r.passed),
    failed: count(r | ¬r.passed),
    all_passed: all_passed
  }
}

provide_feedback :: Validation_Result → Feedback
provide_feedback(V) = {
  format: match V.status with
    | pass → {
        header: "===========================================\nPre-Commit Hook: API Consistency Check\n===========================================",
        body: "\nDetected changes to " + V.file + "\nRunning validation...\n\n" + V.report,
        footer: "\n✓ Validation PASSED\n✓ Commit allowed\n"
      },
    | fail → {
        header: "===========================================\nPre-Commit Hook: API Consistency Check\n===========================================",
        body: "\nDetected changes to " + V.file + "\nRunning validation...\n\n" + V.report,
        footer: "\n✗ Validation FAILED\nViolations found in " + V.file + "\n\nPlease fix the errors above before committing.\n\nTo bypass this check (not recommended):\n  git commit --no-verify\n"
      },

  colors: {
    red: "\033[0;31m",
    green: "\033[0;32m",
    yellow: "\033[1;33m",
    nc: "\033[0m"
  }
}

automate_installation :: Hook → Automation
automate_installation(H) = {
  makefile: {
    target: "install-hooks",
    command: "@bash scripts/install-hooks.sh"
  },

  onboarding: {
    setup_script: "make install-hooks",
    readme: "## Setup\nRun `make install-hooks` to install pre-commit hooks."
  }
}

integrate_ci :: Hook → CI_Integration
integrate_ci(H) = {
  github_actions: {
    name: "Validate API Consistency",
    trigger: {
      pull_request: {paths: [H.trigger_files]},
      push: {branches: ["main"], paths: [H.trigger_files]}
    },
    jobs: {
      validate: {
        runs_on: "ubuntu-latest",
        steps: [
          "checkout",
          "setup-go@v4",
          "build: go build -o ./validate-api ./cmd/validate-api",
          "validate: ./validate-api --format json " + H.target_file,
          "upload_artifacts: validation-results.json (on failure)"
        ]
      }
    }
  },

  gitlab_ci: {
    stage: "test",
    script: [
      "go build -o ./validate-api ./cmd/validate-api",
      "./validate-api --format json " + H.target_file
    ],
    only: {changes: [H.trigger_files]},
    artifacts: {when: "on_failure", paths: ["validation-results.json"]}
  }
}

create_troubleshooting_guide :: () → Troubleshooting
create_troubleshooting_guide() = {
  issues: [
    {
      name: "Hook not running",
      symptom: "Commit succeeds without validation",
      cause: "Hook not executable or not installed",
      fix: "chmod +x .git/hooks/pre-commit\n./scripts/install-hooks.sh"
    },
    {
      name: "Validation tool not found",
      symptom: "Error: validation tool not found",
      cause: "Tool not built",
      fix: "make validate  # or go build"
    },
    {
      name: "Hook blocking valid commit",
      symptom: "Validation fails but changes are valid",
      cause: "False positive in validator",
      fix: "1. Review output\n2. Fix if real issue\n3. Use --no-verify if false positive\n4. Report validator bug"
    },
    {
      name: "Need to bypass hook temporarily",
      symptom: "Emergency commit needed",
      fix: "git commit --no-verify -m 'emergency: bypass hook'"
    },
    {
      name: "Hook runs too slowly",
      symptom: "Hook takes >5 seconds",
      cause: "Full validation running",
      fix: "Hook uses --fast flag (skips slow checks)"
    },
    {
      name: "Restore old hook",
      symptom: "Need to revert hook changes",
      fix: "mv .git/hooks/pre-commit.backup .git/hooks/pre-commit"
    }
  ],

  constraint: |issues| ≥ 6
}

monitor_effectiveness :: Hook → Effectiveness_Metrics
monitor_effectiveness(H) = {
  violations_prevented: count(blocked_commits),
  commits_blocked: count(exit_code = 1),
  bypass_count: count(commits_with_no_verify),
  bypass_rate: bypass_count / total_commits,
  average_runtime: avg(execution_time),

  target_runtime: average_runtime < 5.0  # seconds
}

output :: Installation → Report
output(I) = {
  installation: {
    hook_type: I.gate.type,
    hook_path: ".git/hooks/" + I.gate.type,
    backup_created: I.backup_exists,
    validation_tool_built: I.tool_built,
    test_result: I.test.all_passed
  },

  hook_behavior: {
    trigger_files: I.gate.trigger_files,
    validation_command: I.gate.validation_command,
    fail_on_error: I.behavior.fail_on_error,
    allow_bypass: I.behavior.allow_bypass
  },

  integration: {
    makefile_target: I.automation.makefile.target,
    ci_config_created: I.ci.github_actions ≠ null ∨ I.ci.gitlab_ci ≠ null,
    documentation_path: I.docs.path
  },

  effectiveness: {
    violations_prevented: I.metrics.violations_prevented,
    commits_blocked: I.metrics.commits_blocked,
    bypass_count: I.metrics.bypass_count,
    average_runtime: I.metrics.average_runtime
  }
}

constraints :: Installation → Bool
constraints(I) =
  prerequisites_verified(git_repo, validation_tool, hook_sample) ∧
  hook_installed(".git/hooks/" + I.gate.type) ∧
  triggers_on_relevant_files(I.gate.trigger_pattern) ∧
  blocks_on_failure(I.behavior.fail_on_error) ∧
  allows_on_success(exit 0) ∧
  bypass_available("--no-verify") ∧
  clear_feedback(pass_message, fail_message) ∧
  fast_execution(I.metrics.average_runtime < 5.0) ∧
  documentation_provided(I.docs) ∧
  ci_integration_available(I.ci) ∧
  test_scenarios_passed(I.test.all_passed) ∧
  troubleshooting_items(I.troubleshooting) ∧ |I.troubleshooting.issues| ≥ 6
