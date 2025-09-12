#!/bin/bash

# Script to run go mod tidy on all modules

set -e

echo "Running go mod tidy on all modules..."

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

# Run go mod tidy on each module
for module in "${modules[@]}"; do
    if [ -d "$module" ]; then
        echo "Running go mod tidy in $module..."
        cd "$module"
        go mod tidy
        cd ..
        echo "Completed $module"
    else
        echo "Warning: Directory $module not found, skipping..."
    fi
done

echo "All modules have been tidied!"
