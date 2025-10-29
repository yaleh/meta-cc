#!/bin/bash
# Count artifacts in skill directory

SKILL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "Skill: subagent-prompt-construction"
echo "=========================================="
echo ""

# Count templates
TEMPLATE_COUNT=$(find "$SKILL_DIR/templates" -name "*.md" 2>/dev/null | wc -l)
echo "Templates: $TEMPLATE_COUNT"
find "$SKILL_DIR/templates" -name "*.md" 2>/dev/null | sed 's|.*/|  - |'

echo ""

# Count reference docs
REFERENCE_COUNT=$(find "$SKILL_DIR/reference" -name "*.md" 2>/dev/null | wc -l)
echo "Reference Docs: $REFERENCE_COUNT"
find "$SKILL_DIR/reference" -name "*.md" 2>/dev/null | sed 's|.*/|  - |'

echo ""

# Count examples
EXAMPLE_COUNT=$(find "$SKILL_DIR/examples" -name "*.md" 2>/dev/null | wc -l)
echo "Examples: $EXAMPLE_COUNT"
find "$SKILL_DIR/examples" -name "*.md" 2>/dev/null | sed 's|.*/|  - |'

echo ""

# Count scripts
SCRIPT_COUNT=$(find "$SKILL_DIR/scripts" -name "*.sh" -o -name "*.py" 2>/dev/null | wc -l)
echo "Scripts: $SCRIPT_COUNT"
find "$SKILL_DIR/scripts" \( -name "*.sh" -o -name "*.py" \) 2>/dev/null | sed 's|.*/|  - |'

echo ""
echo "=========================================="
echo "Total artifacts: $((TEMPLATE_COUNT + REFERENCE_COUNT + EXAMPLE_COUNT + SCRIPT_COUNT))"
