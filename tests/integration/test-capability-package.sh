#!/bin/bash
# Integration test for capability package creation and extraction

set -e

BUILD_DIR="build"
PACKAGE="capabilities-latest.tar.gz"
TEST_DIR="/tmp/test-capability-package-$$"

echo "=== Testing Capability Package Creation ==="

# Clean previous builds
make clean-capabilities >/dev/null 2>&1 || true

# Build package
echo "Building capability package..."
make bundle-capabilities

# Verify package exists
if [ ! -f "$BUILD_DIR/$PACKAGE" ]; then
    echo "ERROR: Package not created"
    exit 1
fi
echo "✓ Package created: $BUILD_DIR/$PACKAGE"

# Extract package
echo "Extracting package to test directory..."
mkdir -p "$TEST_DIR"
tar -xzf "$BUILD_DIR/$PACKAGE" -C "$TEST_DIR"

# Verify structure
if [ ! -d "$TEST_DIR/commands" ]; then
    echo "ERROR: commands/ directory not found in package"
    rm -rf "$TEST_DIR"
    exit 1
fi
echo "✓ Package structure valid"

# Count files
COMMAND_COUNT=$(find "$TEST_DIR/commands" -name "*.md" 2>/dev/null | wc -l)
if [ "$COMMAND_COUNT" -lt 1 ]; then
    echo "ERROR: No capability files found in package"
    rm -rf "$TEST_DIR"
    exit 1
fi
echo "✓ Found $COMMAND_COUNT capability files"

# Verify frontmatter in sample file
SAMPLE_FILE=$(find "$TEST_DIR/commands" -name "*.md" | head -1)
if [ -f "$SAMPLE_FILE" ]; then
    if ! grep -q "^---$" "$SAMPLE_FILE"; then
        echo "WARNING: Sample file missing frontmatter: $SAMPLE_FILE"
    else
        echo "✓ Sample file has valid frontmatter"
    fi
fi

# Test package re-extraction (idempotent)
echo "Testing re-extraction..."
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"
tar -xzf "$BUILD_DIR/$PACKAGE" -C "$TEST_DIR"
REEXTRACT_COUNT=$(find "$TEST_DIR/commands" -name "*.md" 2>/dev/null | wc -l)
if [ "$REEXTRACT_COUNT" -ne "$COMMAND_COUNT" ]; then
    echo "ERROR: Re-extraction produced different file count"
    rm -rf "$TEST_DIR"
    exit 1
fi
echo "✓ Re-extraction successful (idempotent)"

# Clean up
rm -rf "$TEST_DIR"
echo "✓ All tests passed"
echo ""
echo "Package Details:"
echo "  Path: $BUILD_DIR/$PACKAGE"
echo "  Size: $(du -h $BUILD_DIR/$PACKAGE | cut -f1)"
echo "  Files: $(tar -tzf $BUILD_DIR/$PACKAGE | wc -l)"
echo "  Capabilities: $COMMAND_COUNT"
