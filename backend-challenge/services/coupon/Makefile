# Makefile for Coupon Service

.PHONY: help build run test clean docker-build docker-run docker-stop migrate-up migrate-down deps

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application locally"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  docker-stop  - Stop Docker Compose services"
	@echo "  migrate-up   - Run database migrations up"
	@echo "  migrate-down - Run database migrations down"
	@echo "  deps         - Download dependencies"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"

# Application name
APP_NAME=coupon-service

# Database connection for migrations
DB_URL=postgres://postgres:postgres@localhost:5432/coupon_service?sslmode=disable

# Build the application
build: generate
	@echo "Building $(APP_NAME)..."
	go build -o bin/$(APP_NAME) cmd/server/main.go

# Run the application locally
run:
	@echo "Running $(APP_NAME)..."
	go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f $(APP_NAME)

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run

# Docker targets
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME) .

docker-run:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

docker-stop:
	@echo "Stopping Docker Compose services..."
	docker-compose down

docker-logs:
	@echo "Showing Docker Compose logs..."
	docker-compose logs -f

# Database migration targets (requires migrate CLI)
migrate-up:
	@echo "Running migrations up..."
	migrate -path db/migrations -database "$(DB_URL)" up

migrate-down:
	@echo "Running migrations down..."
	migrate -path db/migrations -database "$(DB_URL)" down

migrate-force:
	@echo "Forcing migration version..."
	@read -p "Enter version: " version; \
	migrate -path db/migrations -database "$(DB_URL)" force $$version

migrate-version:
	@echo "Current migration version:"
	migrate -path db/migrations -database "$(DB_URL)" version

# SQLC code generation
generate:
	@echo "Generating SQLC code..."
	sqlc generate

# Development helpers
dev-setup: install-tools generate
	@echo "Setting up development environment..."
	docker-compose up postgres -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 5
	make migrate-up
	@echo "Development environment ready!"

dev-reset:
	@echo "Resetting development environment..."
	make docker-stop
	docker-compose down -v
	make dev-setup

# Health check
health:
	@echo "Checking service health..."
	curl -f http://localhost:8080/health || echo "Service is not running"

# Install tools
install-tools:
	@echo "Installing development tools..."
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest