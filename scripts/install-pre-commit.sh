#!/bin/bash
# Install pre-commit framework hooks for meta-cc
# Generated during Bootstrap-008 Code Review Methodology (Iteration 3)

set -e

echo "=========================================="
echo "Installing Pre-Commit Framework Hooks"
echo "=========================================="
echo ""

# Check if pre-commit is installed
if ! command -v pre-commit &> /dev/null; then
    echo "❌ pre-commit not found"
    echo ""
    echo "Please install pre-commit:"
    echo "  macOS:   brew install pre-commit"
    echo "  Linux:   pip install pre-commit"
    echo "  Windows: pip install pre-commit"
    echo ""
    exit 1
fi

echo "✓ pre-commit found: $(pre-commit --version)"
echo ""

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo "❌ Not a git repository"
    echo "Please run this script from the meta-cc root directory"
    exit 1
fi

echo "✓ Git repository detected"
echo ""

# Check if .pre-commit-config.yaml exists
if [ ! -f ".pre-commit-config.yaml" ]; then
    echo "❌ .pre-commit-config.yaml not found"
    exit 1
fi

echo "✓ Configuration file found"
echo ""

# Install hooks
echo "Installing pre-commit hooks..."
pre-commit install

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
    pre-commit run --all-files || {
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
echo "  Run manually:      pre-commit run"
echo "  Run on all files:  pre-commit run --all-files"
echo "  Update hooks:      pre-commit autoupdate"
echo ""
