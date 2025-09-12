#!/bin/bash

# Script to create git tags for all submodules

set -e

echo "Creating git tags for all submodules..."

# Create tags for AI submodules
echo "Creating tags for AI submodules..."
git tag ai/providers/anthropic/v1.0.0
git tag ai/providers/deepseek/v1.0.0
git tag ai/providers/google/v1.0.0
git tag ai/providers/openai/v1.0.0
git tag ai/providers/xai/v1.0.0
git tag ai/types/v1.0.0

# Create tags for other modules' types submodules
echo "Creating tags for other module types..."
git tag auth/types/v1.0.0
git tag backup/types/v1.0.0
git tag cache/types/v1.0.0
git tag chaos/types/v1.0.0
git tag circuitbreaker/types/v1.0.0
git tag communication/types/v1.0.0
git tag config/types/v1.0.0
git tag database/types/v1.0.0
git tag discovery/types/v1.0.0
git tag event/types/v1.0.0
git tag failover/types/v1.0.0
git tag filegen/types/v1.0.0
git tag logging/types/v1.0.0
git tag messaging/types/v1.0.0
git tag middleware/types/v1.0.0
git tag monitoring/types/v1.0.0
git tag payment/types/v1.0.0
git tag ratelimit/types/v1.0.0
git tag scheduling/types/v1.0.0
git tag storage/types/v1.0.0

echo "All submodule tags created successfully!"
echo "Run 'git push --tags' to push all tags to remote repository."
