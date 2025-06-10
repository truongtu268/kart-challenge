# Kart Challenge - Root Makefile
# This Makefile orchestrates the entire microservices deployment

# Variables
PROJECT_NAME=kart-challenge
CLUSTER_NAME=kart-challenge
NAMESPACE=kart-challenge
API_GATEWAY_PORT=8080

# Service directories
SERVICES_DIR=services
PRODUCT_SERVICE=$(SERVICES_DIR)/product
COUPON_SERVICE=$(SERVICES_DIR)/coupon
ORDER_SERVICE=$(SERVICES_DIR)/order
INFRA_DIR=infa

# Docker images
PRODUCT_IMAGE=product-service:latest
COUPON_IMAGE=coupon-service:latest
ORDER_IMAGE=order-service:latest

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Prerequisites check
.PHONY: check-prereqs
check-prereqs: ## Check if all prerequisites are installed
	@echo "$(BLUE)[INFO]$(NC) Checking prerequisites..."
	@command -v docker >/dev/null 2>&1 || { echo "$(RED)[ERROR]$(NC) Docker is not installed"; exit 1; }
	@command -v kind >/dev/null 2>&1 || { echo "$(RED)[ERROR]$(NC) Kind is not installed"; exit 1; }
	@command -v kubectl >/dev/null 2>&1 || { echo "$(RED)[ERROR]$(NC) kubectl is not installed"; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "$(RED)[ERROR]$(NC) Go is not installed"; exit 1; }
	@docker info >/dev/null 2>&1 || { echo "$(RED)[ERROR]$(NC) Docker is not running"; exit 1; }
	@echo "$(GREEN)[SUCCESS]$(NC) All prerequisites are met"

# Development targets
.PHONY: deps
deps: ## Install dependencies for all services
	@echo "$(BLUE)[INFO]$(NC) Installing dependencies..."
	@cd $(PRODUCT_SERVICE) && go mod download && go mod tidy
	@cd $(COUPON_SERVICE) && go mod download && go mod tidy
	@cd $(ORDER_SERVICE) && go mod download && go mod tidy
	@echo "$(GREEN)[SUCCESS]$(NC) Dependencies installed"

.PHONY: generate
generate: ## Generate code for all services (SQLC)
	@echo "$(BLUE)[INFO]$(NC) Generating code..."
	@cd $(PRODUCT_SERVICE) && sqlc generate
	@cd $(COUPON_SERVICE) && sqlc generate
	@cd $(ORDER_SERVICE) && sqlc generate
	@echo "$(GREEN)[SUCCESS]$(NC) Code generation completed"

.PHONY: build
build: ## Build all service binaries
	@echo "$(BLUE)[INFO]$(NC) Building service binaries..."
	@cd $(PRODUCT_SERVICE) && go build -o bin/product-service cmd/server/main.go
	@cd $(COUPON_SERVICE) && go build -o bin/coupon-service cmd/server/main.go
	@cd $(ORDER_SERVICE) && go build -o bin/order-service cmd/server/main.go
	@echo "$(GREEN)[SUCCESS]$(NC) All binaries built"

.PHONY: test
test: ## Run tests for all services
	@echo "$(BLUE)[INFO]$(NC) Running tests..."
	@cd $(PRODUCT_SERVICE) && go test -v ./...
	@cd $(COUPON_SERVICE) && go test -v ./...
	@cd $(ORDER_SERVICE) && go test -v ./...
	@echo "$(GREEN)[SUCCESS]$(NC) All tests passed"

.PHONY: lint
lint: ## Run linter for all services
	@echo "$(BLUE)[INFO]$(NC) Running linter..."
	@cd $(PRODUCT_SERVICE) && golangci-lint run
	@cd $(COUPON_SERVICE) && golangci-lint run
	@cd $(ORDER_SERVICE) && golangci-lint run
	@echo "$(GREEN)[SUCCESS]$(NC) Linting completed"

.PHONY: fmt
fmt: ## Format code for all services
	@echo "$(BLUE)[INFO]$(NC) Formatting code..."
	@cd $(PRODUCT_SERVICE) && go fmt ./...
	@cd $(COUPON_SERVICE) && go fmt ./...
	@cd $(ORDER_SERVICE) && go fmt ./...
	@echo "$(GREEN)[SUCCESS]$(NC) Code formatting completed"

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker images for all services
	@echo "$(BLUE)[INFO]$(NC) Building Docker images..."
	@cd $(PRODUCT_SERVICE) && docker build -t $(PRODUCT_IMAGE) .
	@cd $(COUPON_SERVICE) && docker build -t $(COUPON_IMAGE) .
	@cd $(ORDER_SERVICE) && docker build -t $(ORDER_IMAGE) .
	@echo "$(GREEN)[SUCCESS]$(NC) All Docker images built"

.PHONY: docker-push
docker-push: ## Push Docker images to registry (configure registry first)
	@echo "$(BLUE)[INFO]$(NC) Pushing Docker images..."
	@docker push $(PRODUCT_IMAGE)
	@docker push $(COUPON_IMAGE)
	@docker push $(ORDER_IMAGE)
	@echo "$(GREEN)[SUCCESS]$(NC) All Docker images pushed"

# Kind cluster management
.PHONY: cluster-create
cluster-create: check-prereqs ## Create Kind cluster
	@echo "$(BLUE)[INFO]$(NC) Creating Kind cluster..."
	@if kind get clusters | grep -q "^$(CLUSTER_NAME)$$"; then \
		echo "$(YELLOW)[WARNING]$(NC) Cluster '$(CLUSTER_NAME)' already exists"; \
	else \
		kind create cluster --config=$(INFRA_DIR)/kind-config.yaml --name=$(CLUSTER_NAME); \
		echo "$(GREEN)[SUCCESS]$(NC) Kind cluster created"; \
	fi

.PHONY: cluster-delete
cluster-delete: ## Delete Kind cluster
	@echo "$(BLUE)[INFO]$(NC) Deleting Kind cluster..."
	@kind delete cluster --name=$(CLUSTER_NAME)
	@echo "$(GREEN)[SUCCESS]$(NC) Kind cluster deleted"

.PHONY: cluster-status
cluster-status: ## Show cluster status
	@echo "$(BLUE)[INFO]$(NC) Cluster status:"
	@kubectl cluster-info --context kind-$(CLUSTER_NAME)
	@echo ""
	@echo "Nodes:"
	@kubectl get nodes
	@echo ""
	@echo "Pods in $(NAMESPACE):"
	@kubectl get pods -n $(NAMESPACE) 2>/dev/null || echo "Namespace $(NAMESPACE) not found"

# Kubernetes deployment
.PHONY: k8s-deploy
k8s-deploy: cluster-create docker-build ## Deploy all services to Kubernetes
	@echo "$(BLUE)[INFO]$(NC) Deploying to Kubernetes..."
	@cd $(INFRA_DIR) && ./deploy.sh deploy
	@echo "$(GREEN)[SUCCESS]$(NC) Deployment completed"

.PHONY: k8s-undeploy
k8s-undeploy: ## Remove all services from Kubernetes
	@echo "$(BLUE)[INFO]$(NC) Removing services from Kubernetes..."
	@kubectl delete namespace $(NAMESPACE) --ignore-not-found=true
	@echo "$(GREEN)[SUCCESS]$(NC) Services removed"

.PHONY: k8s-status
k8s-status: ## Show Kubernetes deployment status
	@cd $(INFRA_DIR) && ./deploy.sh status

.PHONY: k8s-logs
k8s-logs: ## Show logs for all services
	@echo "$(BLUE)[INFO]$(NC) Service logs:"
	@echo ""
	@echo "=== API Gateway Logs ==="
	@kubectl logs -l app=api-gateway -n $(NAMESPACE) --tail=20 || true
	@echo ""
	@echo "=== Product Service Logs ==="
	@kubectl logs -l app=product-service -n $(NAMESPACE) --tail=20 || true
	@echo ""
	@echo "=== Coupon Service Logs ==="
	@kubectl logs -l app=coupon-service -n $(NAMESPACE) --tail=20 || true
	@echo ""
	@echo "=== Order Service Logs ==="
	@kubectl logs -l app=order-service -n $(NAMESPACE) --tail=20 || true

# Port forwarding
.PHONY: port-forward
port-forward: ## Forward API Gateway port to localhost
	@echo "$(BLUE)[INFO]$(NC) Forwarding API Gateway port..."
	@echo "$(GREEN)[SUCCESS]$(NC) API Gateway available at http://localhost:$(API_GATEWAY_PORT)"
	@kubectl port-forward service/api-gateway $(API_GATEWAY_PORT):80 -n $(NAMESPACE)

.PHONY: port-forward-db
port-forward-db: ## Forward PostgreSQL port to localhost
	@echo "$(BLUE)[INFO]$(NC) Forwarding PostgreSQL port..."
	@echo "$(GREEN)[SUCCESS]$(NC) PostgreSQL available at localhost:5432"
	@kubectl port-forward service/postgres 5432:5432 -n $(NAMESPACE)

# Testing
.PHONY: test-api
test-api: ## Test API endpoints
	@echo "$(BLUE)[INFO]$(NC) Testing API endpoints..."
	@echo "Testing API Gateway health..."
	@curl -s http://localhost:$(API_GATEWAY_PORT)/health | jq . || echo "Health check failed"
	@echo ""
	@echo "Testing Product Service..."
	@curl -s http://localhost:$(API_GATEWAY_PORT)/api/product | jq . || echo "Product service test failed"
	@echo ""
	@echo "Testing Coupon Service..."
	@curl -s http://localhost:$(API_GATEWAY_PORT)/api/coupon/validate/SAVE10 | jq . || echo "Coupon service test failed"
	@echo ""
	@echo "Testing Order Service..."
	@curl -s -X POST http://localhost:$(API_GATEWAY_PORT)/api/order \
		-H "Content-Type: application/json" \
		-d '{
			"customer_name": "Test User",
			"customer_email": "test@example.com",
			"items": [
				{
					"product_id": "550e8400-e29b-41d4-a716-446655440000",
					"quantity": 1
				}
			],
			"coupon_code": "SAVE10"
		}' | jq . || echo "Order service test failed"

# Database management
.PHONY: db-connect
db-connect: ## Connect to PostgreSQL database
	@echo "$(BLUE)[INFO]$(NC) Connecting to PostgreSQL..."
	@kubectl exec -it deployment/postgres -n $(NAMESPACE) -- psql -U kartuser -d kart_db

.PHONY: db-migrate
db-migrate: ## Run database migrations
	@echo "$(BLUE)[INFO]$(NC) Running database migrations..."
	@cd $(INFRA_DIR) && ./deploy.sh migrate

# Development workflow
.PHONY: dev-setup
dev-setup: deps generate ## Setup development environment
	@echo "$(GREEN)[SUCCESS]$(NC) Development environment setup completed"

.PHONY: dev-start
dev-start: k8s-deploy ## Start everything for development
	@echo "$(GREEN)[SUCCESS]$(NC) Development environment started"
	@echo "$(BLUE)[INFO]$(NC) API Gateway will be available at http://localhost:$(API_GATEWAY_PORT)"
	@echo "$(BLUE)[INFO]$(NC) Run 'make port-forward' in another terminal to access the API"

.PHONY: dev-stop
dev-stop: k8s-undeploy ## Stop development environment
	@echo "$(GREEN)[SUCCESS]$(NC) Development environment stopped"

.PHONY: dev-restart
dev-restart: dev-stop dev-start ## Restart development environment
	@echo "$(GREEN)[SUCCESS]$(NC) Development environment restarted"

# Service-specific targets
.PHONY: restart-product
restart-product: ## Restart product service
	@echo "$(BLUE)[INFO]$(NC) Restarting product service..."
	@cd $(PRODUCT_SERVICE) && docker build -t $(PRODUCT_IMAGE) .
	@kind load docker-image $(PRODUCT_IMAGE) --name=$(CLUSTER_NAME)
	@kubectl rollout restart deployment/product-service -n $(NAMESPACE)
	@kubectl rollout status deployment/product-service -n $(NAMESPACE)
	@echo "$(GREEN)[SUCCESS]$(NC) Product service restarted"

.PHONY: restart-coupon
restart-coupon: ## Restart coupon service
	@echo "$(BLUE)[INFO]$(NC) Restarting coupon service..."
	@cd $(COUPON_SERVICE) && docker build -t $(COUPON_IMAGE) .
	@kind load docker-image $(COUPON_IMAGE) --name=$(CLUSTER_NAME)
	@kubectl rollout restart deployment/coupon-service -n $(NAMESPACE)
	@kubectl rollout status deployment/coupon-service -n $(NAMESPACE)
	@echo "$(GREEN)[SUCCESS]$(NC) Coupon service restarted"

.PHONY: restart-order
restart-order: ## Restart order service
	@echo "$(BLUE)[INFO]$(NC) Restarting order service..."
	@cd $(ORDER_SERVICE) && docker build -t $(ORDER_IMAGE) .
	@kind load docker-image $(ORDER_IMAGE) --name=$(CLUSTER_NAME)
	@kubectl rollout restart deployment/order-service -n $(NAMESPACE)
	@kubectl rollout status deployment/order-service -n $(NAMESPACE)
	@echo "$(GREEN)[SUCCESS]$(NC) Order service restarted"

.PHONY: restart-gateway
restart-gateway: ## Restart API gateway
	@echo "$(BLUE)[INFO]$(NC) Restarting API gateway..."
	@kubectl rollout restart deployment/api-gateway -n $(NAMESPACE)
	@kubectl rollout status deployment/api-gateway -n $(NAMESPACE)
	@echo "$(GREEN)[SUCCESS]$(NC) API gateway restarted"

# Cleanup
.PHONY: clean
clean: ## Clean build artifacts and Docker images
	@echo "$(BLUE)[INFO]$(NC) Cleaning up..."
	@rm -rf $(PRODUCT_SERVICE)/bin/
	@rm -rf $(COUPON_SERVICE)/bin/
	@rm -rf $(ORDER_SERVICE)/bin/
	@docker rmi $(PRODUCT_IMAGE) 2>/dev/null || true
	@docker rmi $(COUPON_IMAGE) 2>/dev/null || true
	@docker rmi $(ORDER_IMAGE) 2>/dev/null || true
	@echo "$(GREEN)[SUCCESS]$(NC) Cleanup completed"

.PHONY: clean-all
clean-all: clean cluster-delete ## Clean everything including Kind cluster
	@echo "$(GREEN)[SUCCESS]$(NC) Complete cleanup finished"

# All-in-one targets
.PHONY: setup
setup: check-prereqs dev-setup ## Complete setup for development
	@echo "$(GREEN)[SUCCESS]$(NC) Complete setup finished"
	@echo "$(BLUE)[INFO]$(NC) Next steps:"
	@echo "  1. Run 'make dev-start' to start all services"
	@echo "  2. Run 'make port-forward' to access the API"
	@echo "  3. Run 'make test-api' to test the endpoints"

.PHONY: deploy
deploy: setup k8s-deploy ## Complete deployment pipeline
	@echo "$(GREEN)[SUCCESS]$(NC) Complete deployment finished"
	@echo "$(BLUE)[INFO]$(NC) Run 'make port-forward' to access the API at http://localhost:$(API_GATEWAY_PORT)"

# Quick commands
.PHONY: start
start: dev-start ## Alias for dev-start

.PHONY: stop
stop: dev-stop ## Alias for dev-stop

.PHONY: restart
restart: dev-restart ## Alias for dev-restart

.PHONY: status
status: k8s-status ## Alias for k8s-status

.PHONY: logs
logs: k8s-logs ## Alias for k8s-logs