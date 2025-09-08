#!/bin/bash

# Tidy script for logging library
echo "Tidying logging library..."

# Go to logging directory
cd "$(dirname "$0")"

# Tidy main module
echo "Tidying main logging module..."
go mod tidy

# Tidy types module
echo "Tidying types module..."
cd types
go mod tidy
cd ..

# Tidy gateway module
echo "Tidying gateway module..."
cd gateway
go mod tidy
cd ..

# Tidy console provider
echo "Tidying console provider..."
cd providers/console
go mod tidy
cd ../..

# Tidy file provider
echo "Tidying file provider..."
cd providers/file
go mod tidy
cd ../..

# Tidy elasticsearch provider
echo "Tidying elasticsearch provider..."
cd providers/elasticsearch
go mod tidy
cd ../..

# Tidy examples
echo "Tidying examples..."
cd examples
go mod tidy
cd ..

echo "Logging library tidied successfully!"
