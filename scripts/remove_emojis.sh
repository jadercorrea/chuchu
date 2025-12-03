#!/bin/bash

# Remove all emojis from Go source files
# Keeps markdown/docs as-is (user can choose to clean those separately)

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "Removing emojis from Go source files..."

# Find all .go files and remove emoji unicode ranges
find "$PROJECT_ROOT" -name "*.go" -type f | while read -r file; do
    # Remove common emojis used in output
    # Using LC_ALL=C for byte-based sed operations
    LC_ALL=C sed -i '' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's/▏//g' \
        -e 's/▌//g' \
        "$file"
done

# Also clean shell scripts in tests/
find "$PROJECT_ROOT/tests" -name "*.sh" -type f | while read -r file; do
    LC_ALL=C sed -i '' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        "$file"
done

# Clean scripts/
find "$PROJECT_ROOT/scripts" -name "*.sh" -type f | while read -r file; do
    LC_ALL=C sed -i '' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        -e 's///g' \
        "$file"
done

echo "Done. Emojis removed from:"
echo "  - All .go files"
echo "  - All .sh files in tests/ and scripts/"
echo ""
echo "Note: Markdown/docs files unchanged (manual cleanup if needed)"
