#!/bin/bash

# Script to fix pseudo-versions with actual commit hash

set -e

echo "Fixing pseudo-versions with actual commit hash..."

# Get the current commit hash
COMMIT_HASH=$(git rev-parse HEAD)
echo "Using commit hash: $COMMIT_HASH"

# Create pseudo-version from commit hash
PSEUDO_VERSION="v0.0.0-$(date -u +%Y%m%d%H%M%S)-${COMMIT_HASH:0:12}"
echo "Using pseudo-version: $PSEUDO_VERSION"

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

# Function to update go.mod file with proper pseudo-version
update_go_mod_pseudo_version() {
    local module=$1
    local go_mod_file="$module/go.mod"
    
    if [ ! -f "$go_mod_file" ]; then
        echo "Warning: $go_mod_file not found, skipping..."
        return
    fi
    
    echo "Processing $go_mod_file..."
    
    # Update internal dependencies to use proper pseudo-version
    sed -i "s|github.com/anasamu/microservices-library-go/\([^/]*\)/types v0.0.0-00010101000000-000000000000|github.com/anasamu/microservices-library-go/\1/types $PSEUDO_VERSION|g" "$go_mod_file"
    sed -i "s|github.com/anasamu/microservices-library-go/\([^/]*\)/types v0.0.0|github.com/anasamu/microservices-library-go/\1/types $PSEUDO_VERSION|g" "$go_mod_file"
    
    # Update provider dependencies
    sed -i "s|github.com/anasamu/microservices-library-go/ai/providers/\([^ ]*\) v0.0.0-00010101000000-000000000000|github.com/anasamu/microservices-library-go/ai/providers/\1 $PSEUDO_VERSION|g" "$go_mod_file"
    
    echo "Updated $go_mod_file"
}

# Update main modules
for module in "${modules[@]}"; do
    update_go_mod_pseudo_version "$module"
done

# Update AI provider modules
providers=("anthropic" "deepseek" "google" "openai" "xai")
for provider in "${providers[@]}"; do
    go_mod_file="ai/providers/$provider/go.mod"
    if [ -f "$go_mod_file" ]; then
        echo "Processing $go_mod_file..."
        sed -i "s|github.com/anasamu/microservices-library-go/ai/types v0.0.0-00010101000000-000000000000|github.com/anasamu/microservices-library-go/ai/types $PSEUDO_VERSION|g" "$go_mod_file"
        echo "Updated $go_mod_file"
    fi
done

echo "All pseudo-versions updated!"
