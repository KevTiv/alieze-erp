# Simple Makefile for a Go project

# Build the application
all: build test

build:
	@echo "Building..."


	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go &
	@npm install --prefer-offline --no-fund --prefix ./frontend
	@npm run dev --prefix ./frontend
# Create DB container
docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Running all tests..."
	@go test ./... -v -count=1

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	@go test ./... -race -v -count=1

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out -v -count=1
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run specific module tests
# Example: make test-module MODULE=internal/database
test-module:
	@echo "Running tests for module: $(MODULE)"
	@go test ./$(MODULE) -v -count=1

# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v -count=1

# Run database-specific tests with Docker
db-test:
	@echo "Running database tests with Docker PostgreSQL..."
	@docker-compose up -d psql_bp
	@sleep 5
	@go test ./internal/database -v -count=1
	@docker-compose down

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

.PHONY: all build run test test-race coverage test-module clean watch docker-run docker-down itest db-test
