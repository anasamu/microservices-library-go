#!/bin/bash

# Script untuk menjalankan go mod tidy di semua modul microservices-library-go
# Usage: ./tidy-all.sh [options]
# Options:
#   -c, --cache-only    Only tidy cache module
#   -a, --all          Tidy all modules (default)
#   -h, --help         Show this help message

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default options
CACHE_ONLY=false
SHOW_HELP=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--cache-only)
            CACHE_ONLY=true
            shift
            ;;
        -a|--all)
            CACHE_ONLY=false
            shift
            ;;
        -h|--help)
            SHOW_HELP=true
            shift
            ;;
        *)
            echo "Unknown option $1"
            exit 1
            ;;
    esac
done

# Show help
if [ "$SHOW_HELP" = true ]; then
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -c, --cache-only    Only tidy cache module"
    echo "  -a, --all          Tidy all modules (default)"
    echo "  -h, --help         Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                 # Tidy all modules"
    echo "  $0 --cache-only    # Only tidy cache module"
    echo "  $0 -c              # Only tidy cache module"
    exit 0
fi

echo -e "${BLUE}🧹 Go Mod Tidy Script for Microservices Library${NC}"
echo "=================================================="

# Function to run go mod tidy in a directory
run_tidy() {
    local dir=$1
    local module_name=$2
    
    if [ -f "$dir/go.mod" ]; then
        echo -e "${YELLOW}📦 Tidying $module_name...${NC}"
        cd "$dir"
        
        # Run go mod tidy and capture output
        if go mod tidy 2>&1; then
            echo -e "${GREEN}✅ Completed $module_name${NC}"
            return 0
        else
            echo -e "${RED}❌ Failed to tidy $module_name${NC}"
            return 1
        fi
        
        cd - > /dev/null
    else
        echo -e "${YELLOW}⚠️  No go.mod found in $dir, skipping...${NC}"
        return 0
    fi
}

# Function to tidy cache module
tidy_cache() {
    echo -e "${BLUE}🗂️  Processing Cache Module${NC}"
    echo "------------------------"
    
    # Array of cache directories that contain go.mod files
    cache_directories=(
        "cache:Main Cache Module"
        "cache/gateway:Cache Gateway"
        "cache/types:Cache Types"
        "cache/providers/redis:Redis Provider"
        "cache/providers/memory:Memory Provider"
        "cache/providers/memcache:Memcache Provider"
        "cache/examples:Cache Examples"
    )
    
    local failed=0
    
    for entry in "${cache_directories[@]}"; do
        IFS=':' read -r dir name <<< "$entry"
        if ! run_tidy "$dir" "$name"; then
            failed=$((failed + 1))
        fi
    done
    
    return $failed
}

# Function to get module name from go.mod
get_module_name() {
    local dir=$1
    if [ -f "$dir/go.mod" ]; then
        # Extract module name from go.mod file
        local module_name
        module_name=$(grep "^module " "$dir/go.mod" | head -1 | cut -d' ' -f2)
        if [ -n "$module_name" ]; then
            echo "$module_name"
        else
            echo "$(basename "$dir")"
        fi
    else
        echo "$(basename "$dir")"
    fi
}

# Function to tidy all modules
tidy_all() {
    echo -e "${BLUE}🗂️  Processing All Modules (Dynamic Discovery)${NC}"
    echo "----------------------------------------------"
    
    # Find all go.mod files in the project
    local go_mod_files
    go_mod_files=$(find . -name "go.mod" -type f -not -path "*/vendor/*" | sort)
    
    local failed=0
    local total=0
    local skipped=0
    
    while IFS= read -r go_mod_file; do
        total=$((total + 1))
        dir=$(dirname "$go_mod_file")
        module_name=$(get_module_name "$dir")
        
        # Skip if it's a vendor directory (additional safety check)
        if [[ "$dir" == *"/vendor/"* ]]; then
            echo -e "${YELLOW}⚠️  Skipping vendor directory: $dir${NC}"
            skipped=$((skipped + 1))
            continue
        fi
        
        if ! run_tidy "$dir" "$module_name"; then
            failed=$((failed + 1))
        fi
    done <<< "$go_mod_files"
    
    echo ""
    echo -e "${BLUE}📊 Summary:${NC}"
    echo "   Total modules found: $((total + skipped))"
    echo "   Modules processed: $total"
    echo "   Modules skipped: $skipped"
    echo -e "   ${GREEN}Successful: $((total - failed))${NC}"
    if [ $failed -gt 0 ]; then
        echo -e "   ${RED}Failed: $failed${NC}"
    fi
    
    return $failed
}

# Main execution
main() {
    local start_time
    start_time=$(date +%s)
    
    if [ "$CACHE_ONLY" = true ]; then
        if tidy_cache; then
            echo ""
            echo -e "${GREEN}🎉 Cache module tidy completed successfully!${NC}"
        else
            echo ""
            echo -e "${RED}❌ Some cache modules failed to tidy${NC}"
            exit 1
        fi
    else
        if tidy_all; then
            echo ""
            echo -e "${GREEN}🎉 All modules tidied successfully!${NC}"
        else
            echo ""
            echo -e "${RED}❌ Some modules failed to tidy${NC}"
            exit 1
        fi
    fi
    
    local end_time
    end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    echo -e "${BLUE}⏱️  Total time: ${duration}s${NC}"
}

# Run main function
main
