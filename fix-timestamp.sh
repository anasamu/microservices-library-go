#!/bin/bash

# Script to fix pseudo-versions with correct git timestamp

set -e

echo "Fixing pseudo-versions with correct git timestamp..."

# Get the current commit hash and timestamp
COMMIT_HASH=$(git rev-parse HEAD)
TIMESTAMP=$(git log -1 --format=%ct)
TIMESTAMP_FORMATTED=$(date -u -d @$TIMESTAMP +%Y%m%d%H%M%S)

echo "Using commit hash: $COMMIT_HASH"
echo "Using timestamp: $TIMESTAMP_FORMATTED"

# Create correct pseudo-version
PSEUDO_VERSION="v0.0.0-$TIMESTAMP_FORMATTED-${COMMIT_HASH:0:12}"
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

# Function to update go.mod file with correct pseudo-version
update_go_mod_timestamp() {
    local module=$1
    local go_mod_file="$module/go.mod"
    
    if [ ! -f "$go_mod_file" ]; then
        echo "Warning: $go_mod_file not found, skipping..."
        return
    fi
    
    echo "Processing $go_mod_file..."
    
    # Update internal dependencies to use correct pseudo-version
    sed -i "s|github.com/anasamu/microservices-library-go/\([^/]*\)/types v0.0.0-[0-9]*-[a-f0-9]*|github.com/anasamu/microservices-library-go/\1/types $PSEUDO_VERSION|g" "$go_mod_file"
    
    # Update provider dependencies
    sed -i "s|github.com/anasamu/microservices-library-go/ai/providers/\([^ ]*\) v0.0.0-[0-9]*-[a-f0-9]*|github.com/anasamu/microservices-library-go/ai/providers/\1 $PSEUDO_VERSION|g" "$go_mod_file"
    
    echo "Updated $go_mod_file"
}

# Update main modules
for module in "${modules[@]}"; do
    update_go_mod_timestamp "$module"
done

# Update AI provider modules
providers=("anthropic" "deepseek" "google" "openai" "xai")
for provider in "${providers[@]}"; do
    go_mod_file="ai/providers/$provider/go.mod"
    if [ -f "$go_mod_file" ]; then
        echo "Processing $go_mod_file..."
        sed -i "s|github.com/anasamu/microservices-library-go/ai/types v0.0.0-[0-9]*-[a-f0-9]*|github.com/anasamu/microservices-library-go/ai/types $PSEUDO_VERSION|g" "$go_mod_file"
        echo "Updated $go_mod_file"
    fi
done

echo "All pseudo-versions updated with correct timestamp!"
