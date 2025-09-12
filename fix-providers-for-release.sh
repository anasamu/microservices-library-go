#!/bin/bash

# Script to fix provider modules for release

set -e

echo "Fixing provider modules for release..."

# Fix AI provider modules
providers=("anthropic" "deepseek" "google" "openai" "xai")

for provider in "${providers[@]}"; do
    go_mod_file="ai/providers/$provider/go.mod"
    if [ -f "$go_mod_file" ]; then
        echo "Fixing $go_mod_file..."
        
        # Remove replace directive
        sed -i '/^replace /d' "$go_mod_file"
        
        # Update types dependency to use v1.0.0
        sed -i 's|github.com/anasamu/microservices-library-go/ai/types v0.0.0-20250912212654-08af9e89ff53|github.com/anasamu/microservices-library-go/ai/types v1.0.0|g' "$go_mod_file"
        
        echo "Updated $go_mod_file"
    fi
done

echo "Provider modules fixed for release!"
