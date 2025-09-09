# Event Sourcing Library Makefile

.PHONY: help build test test-unit test-integration test-coverage clean tidy examples lint

# Default target
help:
	@echo "Available targets:"
	@echo "  build          - Build all packages"
	@echo "  test           - Run all tests"
	@echo "  test-unit      - Run unit tests only"
	@echo "  test-integration - Run integration tests (requires database)"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  examples       - Build and run examples"
	@echo "  clean          - Clean build artifacts"
	@echo "  tidy           - Tidy all go.mod files"
	@echo "  lint           - Run linter"

# Build all packages
build:
	@echo "Building event sourcing library..."
	@go build ./...
	@echo "Build completed successfully"

# Run all tests
test: test-unit

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	@cd test/unit && go test -v ./...

# Run integration tests (requires database setup)
test-integration:
	@echo "Running integration tests..."
	@echo "Note: Integration tests require PostgreSQL running on localhost:5432"
	@echo "Database: eventsourcing_test, User: postgres, Password: password"
	@cd test/integration && go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@cd test/unit && go test -v -coverprofile=coverage.out ./...
	@cd test/unit && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: test/unit/coverage.html"

# Build and run examples
examples:
	@echo "Building examples..."
	@cd examples && go build -o example example.go
	@echo "Running examples..."
	@cd examples && ./example

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@find . -name "*.exe" -delete
	@find . -name "*.exe~" -delete
	@find . -name "*.test" -delete
	@find . -name "*.out" -delete
	@find . -name "*.html" -delete
	@find . -name "coverage.out" -delete
	@find . -name "coverage.html" -delete
	@rm -f examples/example
	@echo "Clean completed"

# Tidy all go.mod files
tidy:
	@echo "Tidying go.mod files..."
	@go mod tidy
	@cd gateway && go mod tidy
	@cd providers/postgresql && go mod tidy
	@cd providers/kafka && go mod tidy
	@cd providers/nats && go mod tidy
	@cd examples && go mod tidy
	@cd test/unit && go mod tidy
	@cd test/integration && go mod tidy
	@echo "Tidying completed"

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Setup development environment
setup:
	@echo "Setting up development environment..."
	@go mod download
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development environment setup completed"

# Docker commands for testing
docker-postgres:
	@echo "Starting PostgreSQL for testing..."
	@docker run --name eventsourcing-postgres -e POSTGRES_PASSWORD=password -e POSTGRES_DB=eventsourcing_test -p 5432:5432 -d postgres:13
	@echo "PostgreSQL started. Connection: localhost:5432, Database: eventsourcing_test, User: postgres, Password: password"

docker-kafka:
	@echo "Starting Kafka for testing..."
	@docker run --name eventsourcing-kafka -p 9092:9092 -d confluentinc/cp-kafka:latest
	@echo "Kafka started on localhost:9092"

docker-nats:
	@echo "Starting NATS for testing..."
	@docker run --name eventsourcing-nats -p 4222:4222 -d nats:latest
	@echo "NATS started on localhost:4222"

docker-stop:
	@echo "Stopping test containers..."
	@docker stop eventsourcing-postgres eventsourcing-kafka eventsourcing-nats 2>/dev/null || true
	@docker rm eventsourcing-postgres eventsourcing-kafka eventsourcing-nats 2>/dev/null || true
	@echo "Test containers stopped"

# Full test suite with Docker
test-full: docker-stop docker-postgres
	@sleep 5  # Wait for PostgreSQL to start
	@$(MAKE) test-integration
	@$(MAKE) docker-stop

# CI/CD pipeline
ci: tidy lint test-unit
	@echo "CI pipeline completed successfully"
