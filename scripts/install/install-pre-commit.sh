#!/bin/bash
# Install pre-commit framework hooks for meta-cc
# Original: Bootstrap-008 Code Review Methodology (Iteration 3)
# Refactored: DIR-092 — module-aware detection with venv fallback

set -e

echo "=========================================="
echo "Installing Pre-Commit Framework Hooks"
echo "=========================================="
echo ""

# ---------------------------------------------------------------------------
# Module-aware detection
# ---------------------------------------------------------------------------

PRE_COMMIT_BIN=""
PRE_COMMIT_RUN=""

detect_pre_commit() {
    # 1. python3 -m pre_commit (preferred: always the correct module)
    if python3 -m pre_commit --version >/dev/null 2>&1; then
        PRE_COMMIT_RUN="python3 -m pre_commit"
        echo "✓ pre_commit module found (python3 -m pre_commit)"
        return 0
    fi

    # 2. Bare pre-commit binary on PATH
    if command -v pre-commit &>/dev/null; then
        PRE_COMMIT_RUN="pre-commit"
        echo "✓ pre-commit binary found on PATH"
        return 0
    fi

    return 1
}

install_pre_commit_module() {
    echo ""
    echo "pre_commit Python module not found. Attempting to install..."

    # 1. Try user-level pip install
    if python3 -m pip install --user pre-commit 2>/dev/null; then
        # Verify
        if python3 -m pre_commit --version >/dev/null 2>&1; then
            PRE_COMMIT_RUN="python3 -m pre_commit"
            echo "✓ pre_commit installed (user-level pip)"
            return 0
        fi
    fi

    # 2. Fallback: venv in ~/.local/share/meta-cc/venv
    local VENV_DIR="${HOME}/.local/share/meta-cc/venv"
    echo "User-level pip failed, creating venv at ${VENV_DIR}..."
    python3 -m venv "${VENV_DIR}"
    "${VENV_DIR}/bin/pip" install pre-commit >/dev/null 2>&1

    # Verify
    if "${VENV_DIR}/bin/python3" -m pre_commit --version >/dev/null 2>&1; then
        PRE_COMMIT_RUN="${VENV_DIR}/bin/python3 -m pre_commit"
        echo "✓ pre_commit installed (venv at ${VENV_DIR})"
        return 0
    fi

    return 1
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

if ! detect_pre_commit; then
    if ! install_pre_commit_module; then
        echo ""
        echo "❌ Failed to install pre_commit module"
        echo ""
        echo "Fix: python3 -m pip install --user pre-commit"
        exit 1
    fi
fi

# ---------------------------------------------------------------------------
# Pre-flight checks
# ---------------------------------------------------------------------------

if ! git rev-parse --git-dir >/dev/null 2>&1; then
    echo "❌ Not a git repository"
    echo "Please run this script from the meta-cc root directory"
    exit 1
fi

echo "✓ Git repository detected"
echo ""

if [ ! -f ".pre-commit-config.yaml" ]; then
    echo "❌ .pre-commit-config.yaml not found"
    exit 1
fi

echo "✓ Configuration file found"
echo ""

# ---------------------------------------------------------------------------
# Install hooks
# ---------------------------------------------------------------------------

echo "Installing pre-commit hooks..."
${PRE_COMMIT_RUN} install

echo ""
echo "✓ Hooks installed to .git/hooks/pre-commit"
echo ""

# Optionally run on all files
read -p "Run pre-commit on all files now? (y/N): " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo "Running pre-commit on all files..."
    echo "(This may take a few minutes on first run)"
    echo ""
    ${PRE_COMMIT_RUN} run --all-files || {
        echo ""
        echo "⚠️  Some checks failed. Please review and fix issues above."
        echo "Hooks are still installed and will run on future commits."
        exit 0
    }
fi

echo ""
echo "=========================================="
echo "✓ Pre-commit hooks installation complete!"
echo "=========================================="
echo ""
echo "Hooks will now run automatically on 'git commit'"
echo ""
echo "Commands:"
echo "  Skip hooks:        git commit --no-verify"
echo "  Run manually:      ${PRE_COMMIT_RUN} run"
echo "  Run on all files:  ${PRE_COMMIT_RUN} run --all-files"
echo "  Update hooks:      ${PRE_COMMIT_RUN} autoupdate"
echo ""
echo "Formatting (run before commit):"
echo "  make fmt            - Format code with gofmt"
echo "  make fix-imports    - Fix import ordering with goimports"
echo ""
