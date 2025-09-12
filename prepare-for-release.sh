#!/bin/bash

# Script to prepare modules for release by removing replace directives
# and updating internal dependencies to use proper version references

set -e

echo "Preparing modules for release..."

# List of all modules
modules=(
    "ai"
    "auth" 
    "backup"
    "cache"
    "chaos"
    "circuitbreaker"
    "communication"
    "config"
    "database"
    "discovery"
    "event"
    "failover"
    "filegen"
    "logging"
    "messaging"
    "middleware"
    "monitoring"
    "payment"
    "ratelimit"
    "scheduling"
    "storage"
)

# Function to update go.mod file for release
update_go_mod_for_release() {
    local module=$1
    local go_mod_file="$module/go.mod"
    
    if [ ! -f "$go_mod_file" ]; then
        echo "Warning: $go_mod_file not found, skipping..."
        return
    fi
    
    echo "Processing $go_mod_file..."
    
    # Create a backup
    cp "$go_mod_file" "$go_mod_file.backup"
    
    # Remove all replace directives
    sed -i '/^replace /d' "$go_mod_file"
    
    # Update internal dependencies to use v1.0.0
    # This handles dependencies like github.com/anasamu/microservices-library-go/ai/types
    sed -i 's|github.com/anasamu/microservices-library-go/\([^/]*\)/types v0.0.0-00010101000000-000000000000|github.com/anasamu/microservices-library-go/\1/types v1.0.0|g' "$go_mod_file"
    sed -i 's|github.com/anasamu/microservices-library-go/\([^/]*\)/types v0.0.0|github.com/anasamu/microservices-library-go/\1/types v1.0.0|g' "$go_mod_file"
    
    # Update other internal dependencies
    sed -i 's|github.com/anasamu/microservices-library-go/\([^/]*\)/providers/\([^ ]*\) v0.0.0-00010101000000-000000000000|github.com/anasamu/microservices-library-go/\1/providers/\2 v1.0.0|g' "$go_mod_file"
    
    # Clean up any empty require blocks
    sed -i '/^require ($/,/^)$/ { /^require ($/! { /^)$/! { /^$/d } } }' "$go_mod_file"
    
    echo "Updated $go_mod_file"
}

# Function to restore go.mod files from backup
restore_go_mod_files() {
    echo "Restoring go.mod files from backup..."
    for module in "${modules[@]}"; do
        local go_mod_file="$module/go.mod"
        local backup_file="$go_mod_file.backup"
        
        if [ -f "$backup_file" ]; then
            mv "$backup_file" "$go_mod_file"
            echo "Restored $go_mod_file"
        fi
    done
}

# Function to clean up backup files
cleanup_backups() {
    echo "Cleaning up backup files..."
    for module in "${modules[@]}"; do
        local backup_file="$module/go.mod.backup"
        if [ -f "$backup_file" ]; then
            rm "$backup_file"
        fi
    done
}

# Main execution
case "${1:-prepare}" in
    "prepare")
        echo "Preparing modules for release..."
        for module in "${modules[@]}"; do
            update_go_mod_for_release "$module"
        done
        echo "Modules prepared for release. Run 'go mod tidy' in each module directory."
        echo "To restore original files, run: $0 restore"
        ;;
    "restore")
        restore_go_mod_files
        cleanup_backups
        echo "Original go.mod files restored."
        ;;
    "cleanup")
        cleanup_backups
        echo "Backup files cleaned up."
        ;;
    *)
        echo "Usage: $0 [prepare|restore|cleanup]"
        echo "  prepare: Remove replace directives and update dependencies for release"
        echo "  restore: Restore original go.mod files from backup"
        echo "  cleanup: Remove backup files"
        exit 1
        ;;
esac
