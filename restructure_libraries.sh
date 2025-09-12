#!/bin/bash

# Script to restructure all microservices libraries
# This script moves gateway files to root and updates go.mod files

set -e

# List of libraries that need restructuring
LIBRARIES=(
    "filegen"
    "middleware" 
    "scheduling"
    "event"
    "failover"
    "ratelimit"
    "discovery"
    "monitoring"
    "logging"
    "config"
    "communication"
    "messaging"
    "storage"
    "payment"
)

# Function to restructure a single library
restructure_library() {
    local lib_name=$1
    local lib_path="./$lib_name"
    
    echo "Processing library: $lib_name"
    
    # Check if gateway directory exists
    if [ ! -d "$lib_path/gateway" ]; then
        echo "  No gateway directory found, skipping..."
        return
    fi
    
    # Check if manager.go exists in gateway
    if [ ! -f "$lib_path/gateway/manager.go" ]; then
        echo "  No manager.go found in gateway, skipping..."
        return
    fi
    
    # Create new manager.go in root with updated package name
    echo "  Creating manager.go in root..."
    sed "s/package gateway/package $lib_name/g" "$lib_path/gateway/manager.go" > "$lib_path/manager.go"
    
    # Update go.mod to add dependencies if needed
    echo "  Updating go.mod..."
    if [ -f "$lib_path/go.mod" ]; then
        # Check if logrus is already in dependencies
        if ! grep -q "github.com/sirupsen/logrus" "$lib_path/go.mod"; then
            # Add logrus dependency if it's used in the manager
            if grep -q "logrus" "$lib_path/manager.go"; then
                sed -i '/^go /a\\nrequire (\n\tgithub.com/sirupsen/logrus v1.9.3\n)\n\nrequire golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect' "$lib_path/go.mod"
            fi
        fi
        
        # Remove gateway replace directive if it exists
        sed -i '/replace.*gateway/d' "$lib_path/go.mod"
    fi
    
    # Remove old gateway files
    echo "  Removing old gateway files..."
    rm -rf "$lib_path/gateway"
    
    echo "  Completed: $lib_name"
}

# Main execution
echo "Starting library restructuring..."

for lib in "${LIBRARIES[@]}"; do
    restructure_library "$lib"
done

echo "Library restructuring completed!"
echo ""
echo "Summary of changes:"
echo "- Moved manager.go from gateway/ to root directory"
echo "- Updated package names from 'gateway' to library name"
echo "- Updated go.mod files to remove gateway dependencies"
echo "- Removed old gateway directories"
echo ""
echo "Import paths have been changed from:"
echo "  github.com/anasamu/microservices-library-go/{lib}/gateway"
echo "To:"
echo "  github.com/anasamu/microservices-library-go/{lib}"
