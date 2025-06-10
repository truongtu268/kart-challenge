# Variables
SERVICE_NAME=order-service
DOCKER_IMAGE=$(SERVICE_NAME):latest
DATABASE_URL=postgres://orderuser:orderpass@localhost:5433/order_db?sslmode=disable

# Default target
.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
.PHONY: dev
dev: ## Run the service in development mode
	DATABASE_URL="$(DATABASE_URL)" go run cmd/server/main.go

.PHONY: build
build: ## Build the service binary
	go build -o bin/$(SERVICE_NAME) cmd/server/main.go

.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Database
.PHONY: db-up
db-up: ## Start database container
	docker-compose up -d order-db

.PHONY: db-down
db-down: ## Stop database container
	docker-compose down order-db

.PHONY: db-migrate
db-migrate: ## Run database migrations
	migrate -path db/migrations -database "$(DATABASE_URL)" up

.PHONY: db-migrate-down
db-migrate-down: ## Rollback database migrations
	migrate -path db/migrations -database "$(DATABASE_URL)" down

.PHONY: db-reset
db-reset: ## Reset database (drop and recreate)
	migrate -path db/migrations -database "$(DATABASE_URL)" drop -f
	migrate -path db/migrations -database "$(DATABASE_URL)" up

# Code generation
.PHONY: generate
generate: ## Generate code (SQLC)
	sqlc generate

# Docker
.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

.PHONY: docker-run
docker-run: ## Run service in Docker
	docker-compose up --build

.PHONY: docker-down
docker-down: ## Stop Docker containers
	docker-compose down

.PHONY: docker-logs
docker-logs: ## Show Docker logs
	docker-compose logs -f $(SERVICE_NAME)

# Linting and formatting
.PHONY: lint
lint: ## Run linter
	golangci-lint run

.PHONY: fmt
fmt: ## Format code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

# Cleanup
.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html
	docker-compose down -v
	docker rmi $(DOCKER_IMAGE) 2>/dev/null || true

# API testing
.PHONY: test-api
test-api: ## Test API endpoints
	@echo "Testing health endpoint..."
	curl -s http://localhost:8081/health | jq .
	@echo "\nTesting place order endpoint..."
	curl -s -X POST http://localhost:8081/order \
		-H "Content-Type: application/json" \
		-d '{
			"customer_name": "John Doe",
			"customer_email": "john@example.com",
			"items": [
				{
					"product_id": "550e8400-e29b-41d4-a716-446655440000",
					"quantity": 2
				}
			],
			"coupon_code": "SAVE10"
		}' | jq .

# Install dependencies
.PHONY: deps
deps: ## Install dependencies
	go mod download
	go mod tidy

# All-in-one commands
.PHONY: setup
setup: deps generate ## Setup development environment

.PHONY: start
start: db-up db-migrate dev ## Start everything for development

.PHONY: stop
stop: docker-down ## Stop all services