#!/bin/bash
# meta-cc marketplace installation validation tests

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass() {
    echo -e "${GREEN}✓${NC} $1"
}

fail() {
    echo -e "${RED}✗${NC} $1"
    exit 1
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Test 1: marketplace.json format validation
test_marketplace_format() {
    echo "Testing marketplace.json format..."

    # Validate JSON syntax
    if ! jq empty .claude-plugin/marketplace.json 2>/dev/null; then
        fail "marketplace.json has invalid JSON syntax"
    fi

    # Check required fields
    REQUIRED="name version description repository assets installation"
    for field in $REQUIRED; do
        if ! jq -e ".$field" .claude-plugin/marketplace.json >/dev/null 2>&1; then
            fail "Missing required field: $field"
        fi
    done

    pass "marketplace.json format valid"
}

# Test 2: Version consistency
test_version_consistency() {
    echo "Testing version consistency..."

    MARKETPLACE_VERSION=$(jq -r '.version' .claude-plugin/marketplace.json)
    PLUGIN_VERSION=$(jq -r '.version' plugin.json)

    if [ "$MARKETPLACE_VERSION" != "$PLUGIN_VERSION" ]; then
        fail "Version mismatch: marketplace=$MARKETPLACE_VERSION, plugin=$PLUGIN_VERSION"
    fi

    pass "Version consistent: $MARKETPLACE_VERSION"
}

# Test 3: Asset references
test_asset_references() {
    echo "Testing asset references..."

    # Check that screenshots exist
    SCREENSHOTS=$(jq -r '.screenshots[]' .claude-plugin/marketplace.json 2>/dev/null)
    for screenshot in $SCREENSHOTS; do
        if [ ! -f "$screenshot" ]; then
            warn "Screenshot not found: $screenshot (may need to be created)"
        else
            pass "Screenshot found: $screenshot"
        fi
    done
}

# Test 4: Documentation cross-references
test_documentation_links() {
    echo "Testing documentation cross-references..."

    # Check that marketplace-listing.md exists
    if [ ! -f docs/marketplace-listing.md ]; then
        fail "docs/marketplace-listing.md not found"
    fi

    # Check README references marketplace
    if ! grep -q "Marketplace Installation" README.md; then
        fail "README.md does not reference marketplace installation"
    fi

    pass "Documentation cross-references valid"
}

# Test 5: CHANGELOG updated
test_changelog_updated() {
    echo "Testing CHANGELOG update..."

    if ! grep -q "Phase 21" CHANGELOG.md; then
        warn "CHANGELOG.md may not include Phase 21 changes"
    else
        pass "CHANGELOG.md includes Phase 21 changes"
    fi
}

# Run all tests
echo "Running marketplace validation tests..."
echo ""

test_marketplace_format
test_version_consistency
test_asset_references
test_documentation_links
test_changelog_updated

echo ""
echo -e "${GREEN}Marketplace validation complete!${NC}"
echo ""
echo "Manual verification steps:"
echo "1. Test: /plugin marketplace add yaleh/meta-cc"
echo "2. Test: /plugin install meta-cc"
echo "3. Verify installation completes successfully"
echo "4. Test slash commands and subagents work"
