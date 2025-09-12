#!/bin/bash

# Script to fix all import references from gateway to root

set -e

echo "Fixing import references..."

# Find all Go files and fix imports
find . -name "*.go" -type f -exec sed -i 's|github.com/anasamu/microservices-library-go/\([^/]*\)/gateway|github.com/anasamu/microservices-library-go/\1|g' {} \;

# Fix go.mod files that still reference gateway
find . -name "go.mod" -type f -exec sed -i '/replace.*gateway/d' {} \;

echo "Import references fixed!"
