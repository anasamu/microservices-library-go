#!/bin/bash

# Script to fix provider dependencies in go.mod files

set -e

echo "Fixing provider dependencies..."

# Fix AI provider dependencies
providers=("anthropic" "deepseek" "google" "openai" "xai")

for provider in "${providers[@]}"; do
    go_mod_file="ai/providers/$provider/go.mod"
    if [ -f "$go_mod_file" ]; then
        echo "Fixing $go_mod_file..."
        # Update the types dependency to use the proper pseudo-version
        sed -i 's|github.com/anasamu/microservices-library-go/ai/types v0.0.0|github.com/anasamu/microservices-library-go/ai/types v0.0.0-00010101000000-000000000000|g' "$go_mod_file"
        echo "Updated $go_mod_file"
    fi
done

echo "Provider dependencies fixed!"
