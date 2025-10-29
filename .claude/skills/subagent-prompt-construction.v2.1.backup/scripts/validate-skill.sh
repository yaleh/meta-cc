#!/bin/bash
# Validate skill structure and content

SKILL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXIT_CODE=0

echo "Validating skill: subagent-prompt-construction"
echo "=========================================="
echo ""

# Check required directories
echo "Checking directory structure..."
REQUIRED_DIRS=("templates" "reference" "examples" "scripts" "inventory")
for dir in "${REQUIRED_DIRS[@]}"; do
    if [ -d "$SKILL_DIR/$dir" ]; then
        echo "  ✓ $dir/"
    else
        echo "  ✗ $dir/ (missing)"
        EXIT_CODE=1
    fi
done
echo ""

# Check SKILL.md
echo "Checking SKILL.md..."
if [ -f "$SKILL_DIR/SKILL.md" ]; then
    echo "  ✓ SKILL.md exists"

    # Check line count (should be ≤40 for compact skill)
    LINE_COUNT=$(wc -l < "$SKILL_DIR/SKILL.md")
    if [ "$LINE_COUNT" -le 80 ]; then
        echo "  ✓ Line count: $LINE_COUNT (≤80 lines)"
    else
        echo "  ⚠ Line count: $LINE_COUNT (>80 lines, consider compacting)"
    fi

    # Check for lambda contract
    if grep -q "^λ(" "$SKILL_DIR/SKILL.md"; then
        echo "  ✓ Lambda contract present"
    else
        echo "  ✗ Lambda contract missing"
        EXIT_CODE=1
    fi

    # Check for frontmatter
    if head -n 1 "$SKILL_DIR/SKILL.md" | grep -q "^---"; then
        echo "  ✓ Frontmatter present"
    else
        echo "  ✗ Frontmatter missing"
        EXIT_CODE=1
    fi
else
    echo "  ✗ SKILL.md missing"
    EXIT_CODE=1
fi
echo ""

# Check templates
echo "Checking templates..."
TEMPLATE_COUNT=$(find "$SKILL_DIR/templates" -name "*.md" 2>/dev/null | wc -l)
if [ "$TEMPLATE_COUNT" -gt 0 ]; then
    echo "  ✓ $TEMPLATE_COUNT template(s) found"
else
    echo "  ⚠ No templates found"
fi
echo ""

# Check reference docs
echo "Checking reference documentation..."
REFERENCE_COUNT=$(find "$SKILL_DIR/reference" -name "*.md" 2>/dev/null | wc -l)
if [ "$REFERENCE_COUNT" -gt 0 ]; then
    echo "  ✓ $REFERENCE_COUNT reference doc(s) found"

    # Check for key reference files
    KEY_REFS=("patterns.md" "symbolic-language.md" "integration-patterns.md")
    for ref in "${KEY_REFS[@]}"; do
        if [ -f "$SKILL_DIR/reference/$ref" ]; then
            echo "    ✓ $ref"
        else
            echo "    ✗ $ref (missing)"
            EXIT_CODE=1
        fi
    done
else
    echo "  ✗ No reference docs found"
    EXIT_CODE=1
fi
echo ""

# Check examples
echo "Checking examples..."
EXAMPLE_COUNT=$(find "$SKILL_DIR/examples" -name "*.md" 2>/dev/null | wc -l)
if [ "$EXAMPLE_COUNT" -gt 0 ]; then
    echo "  ✓ $EXAMPLE_COUNT example(s) found"
else
    echo "  ⚠ No examples found"
fi
echo ""

# Check scripts
echo "Checking scripts..."
SCRIPT_COUNT=$(find "$SKILL_DIR/scripts" \( -name "*.sh" -o -name "*.py" \) 2>/dev/null | wc -l)
if [ "$SCRIPT_COUNT" -gt 0 ]; then
    echo "  ✓ $SCRIPT_COUNT script(s) found"

    # Check executability
    find "$SKILL_DIR/scripts" -name "*.sh" | while read script; do
        if [ -x "$script" ]; then
            echo "    ✓ $(basename "$script") (executable)"
        else
            echo "    ⚠ $(basename "$script") (not executable)"
        fi
    done
else
    echo "  ⚠ No scripts found"
fi
echo ""

# Summary
echo "=========================================="
if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ Validation PASSED"
else
    echo "❌ Validation FAILED"
fi

exit $EXIT_CODE
