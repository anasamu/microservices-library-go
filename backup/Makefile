# Backup Microservice Library Makefile

.PHONY: help build test clean examples

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build all modules
	@echo "Building backup library..."
	@cd gateway && go build
	@cd providers/local && go build
	@cd providers/s3 && go build
	@cd providers/gcs && go build
	@echo "Build completed!"

test: ## Run all tests
	@echo "Running unit tests..."
	@cd test/unit && go test -v
	@echo "Running integration tests..."
	@cd test/integration && go test -v
	@echo "All tests completed!"

test-unit: ## Run unit tests only
	@cd test/unit && go test -v

test-integration: ## Run integration tests only
	@cd test/integration && go test -v

examples: ## Build and run examples
	@echo "Building examples..."
	@cd examples && go build -o backup-example main.go
	@echo "Running backup example..."
	@cd examples && ./backup-example

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@find . -name "*.exe" -delete
	@find . -name "backup-example" -delete
	@find . -name "*.test" -delete
	@echo "Clean completed!"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@cd gateway && go mod download
	@cd providers/local && go mod download
	@cd providers/s3 && go mod download
	@cd providers/gcs && go mod download
	@cd test/unit && go mod download
	@cd test/integration && go mod download
	@cd test/mocks && go mod download
	@cd examples && go mod download
	@echo "Dependencies downloaded!"

fmt: ## Format all Go code
	@echo "Formatting code..."
	@go fmt ./...
	@cd gateway && go fmt
	@cd providers/local && go fmt
	@cd providers/s3 && go fmt
	@cd providers/gcs && go fmt
	@cd test/unit && go fmt
	@cd test/integration && go fmt
	@cd test/mocks && go fmt
	@cd examples && go fmt
	@echo "Code formatted!"

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...
	@cd gateway && go vet
	@cd providers/local && go vet
	@cd providers/s3 && go vet
	@cd providers/gcs && go vet
	@cd test/unit && go vet
	@cd test/integration && go vet
	@cd test/mocks && go vet
	@cd examples && go vet
