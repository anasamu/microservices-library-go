#!/bin/bash

echo "Tidying go.mod files..."

# Main module
echo "Tidying main module..."
go mod tidy

# Gateway
echo "Tidying gateway..."
cd gateway && go mod tidy && cd ..

# Providers
echo "Tidying PostgreSQL provider..."
cd providers/postgresql && go mod tidy && cd ../..

echo "Tidying Kafka provider..."
cd providers/kafka && go mod tidy && cd ../..

echo "Tidying NATS provider..."
cd providers/nats && go mod tidy && cd ../..

# Examples
echo "Tidying examples..."
cd examples && go mod tidy && cd ..

# Tests
echo "Tidying unit tests..."
cd test/unit && go mod tidy && cd ../..

echo "Tidying integration tests..."
cd test/integration && go mod tidy && cd ../..

echo "Tidying completed successfully!"
