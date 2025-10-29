#!/usr/bin/env bash
# validate-skill.sh - Validate skill structure and meta-objective compliance

set -euo pipefail

SKILL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INVENTORY_DIR="$SKILL_DIR/inventory"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Validation results
ERRORS=0
WARNINGS=0

echo "=== Skill Validation Report ==="
echo ""

# 1. Directory structure validation
echo "1. Directory Structure:"
required_dirs=("templates" "examples" "reference" "reference/case-studies" "scripts" "inventory")
for dir in "${required_dirs[@]}"; do
    if [[ -d "$SKILL_DIR/$dir" ]]; then
        echo -e "  ${GREEN}✅${NC} $dir/"
    else
        echo -e "  ${RED}❌${NC} $dir/ (missing)"
        ERRORS=$((ERRORS + 1))
    fi
done
echo ""

# 2. Required files validation
echo "2. Required Files:"
required_files=("SKILL.md" "templates/subagent-template.md" "examples/phase-planner-executor.md")
for file in "${required_files[@]}"; do
    if [[ -f "$SKILL_DIR/$file" ]]; then
        echo -e "  ${GREEN}✅${NC} $file"
    else
        echo -e "  ${RED}❌${NC} $file (missing)"
        ERRORS=$((ERRORS + 1))
    fi
done
echo ""

# 3. Compactness validation
echo "3. Compactness Constraints:"

if [[ -f "$SKILL_DIR/SKILL.md" ]]; then
    skill_lines=$(wc -l < "$SKILL_DIR/SKILL.md")
    if [[ $skill_lines -le 40 ]]; then
        echo -e "  ${GREEN}✅${NC} SKILL.md: $skill_lines lines (≤40)"
    else
        echo -e "  ${RED}❌${NC} SKILL.md: $skill_lines lines (exceeds 40 by $(($skill_lines - 40)))"
        ERRORS=$((ERRORS + 1))
    fi
fi

for file in "$SKILL_DIR"/examples/*.md; do
    if [[ -f "$file" ]]; then
        lines=$(wc -l < "$file")
        basename=$(basename "$file")
        if [[ $lines -le 150 ]]; then
            echo -e "  ${GREEN}✅${NC} examples/$basename: $lines lines (≤150)"
        else
            echo -e "  ${YELLOW}⚠️${NC}  examples/$basename: $lines lines (exceeds 150 by $(($lines - 150)))"
            WARNINGS=$((WARNINGS + 1))
        fi
    fi
done
echo ""

# 4. Lambda contract validation
echo "4. Lambda Contract:"
if [[ -f "$SKILL_DIR/SKILL.md" ]]; then
    if grep -q "^λ(" "$SKILL_DIR/SKILL.md"; then
        echo -e "  ${GREEN}✅${NC} Lambda contract found"
    else
        echo -e "  ${RED}❌${NC} Lambda contract missing"
        ERRORS=$((ERRORS + 1))
    fi
fi
echo ""

# 5. Reference files validation
echo "5. Reference Documentation:"
reference_files=("patterns.md" "integration-patterns.md" "symbolic-language.md")
for file in "${reference_files[@]}"; do
    if [[ -f "$SKILL_DIR/reference/$file" ]]; then
        lines=$(wc -l < "$SKILL_DIR/reference/$file")
        echo -e "  ${GREEN}✅${NC} reference/$file ($lines lines)"
    else
        echo -e "  ${YELLOW}⚠️${NC}  reference/$file (missing)"
        WARNINGS=$((WARNINGS + 1))
    fi
done
echo ""

# 6. Case studies validation
echo "6. Case Studies:"
case_study_count=$(find "$SKILL_DIR/reference/case-studies" -name "*.md" 2>/dev/null | wc -l)
if [[ $case_study_count -gt 0 ]]; then
    echo -e "  ${GREEN}✅${NC} $case_study_count case study file(s) found"
else
    echo -e "  ${YELLOW}⚠️${NC}  No case studies found"
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# 7. Scripts validation
echo "7. Automation Scripts:"
script_count=$(find "$SKILL_DIR/scripts" -name "*.sh" -o -name "*.py" 2>/dev/null | wc -l)
if [[ $script_count -ge 4 ]]; then
    echo -e "  ${GREEN}✅${NC} $script_count script(s) found (≥4)"
else
    echo -e "  ${YELLOW}⚠️${NC}  $script_count script(s) found (target: ≥4)"
    WARNINGS=$((WARNINGS + 1))
fi

# List scripts
for script in "$SKILL_DIR"/scripts/*.{sh,py}; do
    if [[ -f "$script" ]]; then
        basename=$(basename "$script")
        echo "    - $basename"
    fi
done
echo ""

# 8. Meta-objective compliance (from config.json if available)
echo "8. Meta-Objective Compliance:"

config_file="$SKILL_DIR/experiment-config.json"
if [[ -f "$config_file" ]]; then
    echo -e "  ${GREEN}✅${NC} experiment-config.json found"

    # Check V_meta and V_instance
    v_meta=$(grep -oP '"v_meta":\s*\K[0-9.]+' "$config_file" || echo "0")
    v_instance=$(grep -oP '"v_instance":\s*\K[0-9.]+' "$config_file" || echo "0")

    echo "    V_meta: $v_meta (target: ≥0.75)"
    echo "    V_instance: $v_instance (target: ≥0.80)"

    if (( $(echo "$v_instance >= 0.80" | bc -l) )); then
        echo -e "    ${GREEN}✅${NC} V_instance meets threshold"
    else
        echo -e "    ${YELLOW}⚠️${NC}  V_instance below threshold"
        WARNINGS=$((WARNINGS + 1))
    fi

    if (( $(echo "$v_meta >= 0.75" | bc -l) )); then
        echo -e "    ${GREEN}✅${NC} V_meta meets threshold"
    else
        echo -e "    ${YELLOW}⚠️${NC}  V_meta below threshold (near convergence)"
        WARNINGS=$((WARNINGS + 1))
    fi
else
    echo -e "  ${YELLOW}⚠️${NC}  experiment-config.json not found"
    WARNINGS=$((WARNINGS + 1))
fi
echo ""

# Summary
echo "=== Validation Summary ==="
echo ""
if [[ $ERRORS -eq 0 ]]; then
    echo -e "${GREEN}✅ All critical validations passed${NC}"
else
    echo -e "${RED}❌ $ERRORS critical error(s) found${NC}"
fi

if [[ $WARNINGS -gt 0 ]]; then
    echo -e "${YELLOW}⚠️  $WARNINGS warning(s) found${NC}"
fi
echo ""

# Exit code
if [[ $ERRORS -gt 0 ]]; then
    exit 1
else
    exit 0
fi
