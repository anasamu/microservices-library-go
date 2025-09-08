#!/bin/bash

# Script untuk menjalankan go mod tidy di semua direktori cache module
# Usage: ./tidy.sh

set -e  # Exit on any error

echo "🧹 Running go mod tidy for cache module..."

# Array of directories that contain go.mod files
directories=(
    "."
    "gateway"
    "types"
    "providers/redis"
    "providers/memory"
    "providers/memcache"
    "examples"
)

# Function to run go mod tidy in a directory
run_tidy() {
    local dir=$1
    if [ -f "$dir/go.mod" ]; then
        echo "📦 Tidying $dir..."
        cd "$dir"
        go mod tidy
        cd - > /dev/null
        echo "✅ Completed $dir"
    else
        echo "⚠️  No go.mod found in $dir, skipping..."
    fi
}

# Run go mod tidy in each directory
for dir in "${directories[@]}"; do
    run_tidy "$dir"
done

echo ""
echo "🎉 All go mod tidy operations completed successfully!"
echo ""
echo "📋 Summary:"
echo "   - Main cache module: ✅"
echo "   - Gateway: ✅"
echo "   - Types: ✅"
echo "   - Redis provider: ✅"
echo "   - Memory provider: ✅"
echo "   - Memcache provider: ✅"
echo "   - Examples: ✅"
