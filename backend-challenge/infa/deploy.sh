#!/bin/bash

# Kart Challenge Deployment Script
# This script deploys all services to a Kind Kubernetes cluster

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CLUSTER_NAME="kart-challenge"
NAMESPACE="kart-challenge"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Kind is installed
    if ! command -v kind &> /dev/null; then
        log_error "Kind is not installed. Please install Kind first."
        log_info "Installation: https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
        exit 1
    fi
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    
    # Check if docker is running
    if ! docker info &> /dev/null; then
        log_error "Docker is not running. Please start Docker first."
        exit 1
    fi
    
    log_success "All prerequisites are met"
}

create_cluster() {
    log_info "Creating Kind cluster..."
    
    # Check if cluster already exists
    if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
        log_warning "Cluster '${CLUSTER_NAME}' already exists"
        read -p "Do you want to delete and recreate it? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            log_info "Deleting existing cluster..."
            kind delete cluster --name "$CLUSTER_NAME"
        else
            log_info "Using existing cluster"
            return 0
        fi
    fi
    
    # Create cluster with config
    kind create cluster --config="$SCRIPT_DIR/kind-config.yaml" --name="$CLUSTER_NAME"
    
    # Wait for cluster to be ready
    log_info "Waiting for cluster to be ready..."
    kubectl wait --for=condition=Ready nodes --all --timeout=300s
    
    log_success "Kind cluster created successfully"
}

build_images() {
    log_info "Building Docker images..."
    
    # Build product service
    log_info "Building product service image..."
    cd "$PROJECT_ROOT/services/product"
    docker build -t product-service:latest .
    kind load docker-image product-service:latest --name="$CLUSTER_NAME"
    
    # Build coupon service
    log_info "Building coupon service image..."
    cd "$PROJECT_ROOT/services/coupon"
    docker build -t coupon-service:latest .
    kind load docker-image coupon-service:latest --name="$CLUSTER_NAME"
    
    # Build order service
    log_info "Building order service image..."
    cd "$PROJECT_ROOT/services/order"
    docker build -t order-service:latest .
    kind load docker-image order-service:latest --name="$CLUSTER_NAME"
    
    cd "$SCRIPT_DIR"
    log_success "All images built and loaded successfully"
}

deploy_services() {
    log_info "Deploying services to Kubernetes..."
    
    # Apply namespace
    kubectl apply -f "$SCRIPT_DIR/k8s/namespace.yaml"
    
    # Deploy PostgreSQL
    log_info "Deploying PostgreSQL..."
    kubectl apply -f "$SCRIPT_DIR/k8s/postgres.yaml"
    
    # Wait for PostgreSQL to be ready
    log_info "Waiting for PostgreSQL to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/postgres -n "$NAMESPACE"
    
    # Deploy services
    log_info "Deploying product service..."
    kubectl apply -f "$SCRIPT_DIR/k8s/product-service.yaml"
    
    log_info "Deploying coupon service..."
    kubectl apply -f "$SCRIPT_DIR/k8s/coupon-service.yaml"
    
    log_info "Deploying order service..."
    kubectl apply -f "$SCRIPT_DIR/k8s/order-service.yaml"
    
    # Deploy API Gateway
    log_info "Deploying API Gateway..."
    kubectl apply -f "$SCRIPT_DIR/k8s/api-gateway.yaml"
    
    # Wait for all deployments to be ready
    log_info "Waiting for all services to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/product-service -n "$NAMESPACE"
    kubectl wait --for=condition=available --timeout=300s deployment/coupon-service -n "$NAMESPACE"
    kubectl wait --for=condition=available --timeout=300s deployment/order-service -n "$NAMESPACE"
    kubectl wait --for=condition=available --timeout=300s deployment/api-gateway -n "$NAMESPACE"
    
    log_success "All services deployed successfully"
}

run_migrations() {
    log_info "Running database migrations..."
    
    # Get postgres pod name
    POSTGRES_POD=$(kubectl get pods -n "$NAMESPACE" -l app=postgres -o jsonpath='{.items[0].metadata.name}')
    
    if [ -z "$POSTGRES_POD" ]; then
        log_error "PostgreSQL pod not found"
        return 1
    fi
    
    # Run migrations for each service
    log_info "Running product service migrations..."
    kubectl exec -n "$NAMESPACE" "$POSTGRES_POD" -- psql -U productuser -d product_db -c "
        CREATE TABLE IF NOT EXISTS products (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            name VARCHAR(255) NOT NULL,
            description TEXT,
            price DECIMAL(10,2) NOT NULL,
            category VARCHAR(100),
            image_url VARCHAR(500),
            is_available BOOLEAN DEFAULT true,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        
        INSERT INTO products (id, name, description, price, category, image_url) VALUES
        ('550e8400-e29b-41d4-a716-446655440000', 'Margherita Pizza', 'Classic pizza with tomato sauce, mozzarella, and basil', 12.99, 'Pizza', 'https://example.com/margherita.jpg'),
        ('550e8400-e29b-41d4-a716-446655440001', 'Pepperoni Pizza', 'Pizza with tomato sauce, mozzarella, and pepperoni', 14.99, 'Pizza', 'https://example.com/pepperoni.jpg'),
        ('550e8400-e29b-41d4-a716-446655440002', 'Caesar Salad', 'Fresh romaine lettuce with Caesar dressing and croutons', 8.99, 'Salad', 'https://example.com/caesar.jpg')
        ON CONFLICT (id) DO NOTHING;
    "
    
    log_info "Running coupon service migrations..."
    kubectl exec -n "$NAMESPACE" "$POSTGRES_POD" -- psql -U couponuser -d coupon_db -c "
        CREATE TABLE IF NOT EXISTS coupons (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            coupon_code VARCHAR(50) UNIQUE NOT NULL,
            is_valid BOOLEAN DEFAULT true,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        
        INSERT INTO coupons (coupon_code, is_valid) VALUES
        ('SAVE10', true),
        ('WELCOME20', true),
        ('FIFTYOFF', true),
        ('HAPPYHRS', true)
        ON CONFLICT (coupon_code) DO NOTHING;
    "
    
    log_info "Running order service migrations..."
    kubectl exec -n "$NAMESPACE" "$POSTGRES_POD" -- psql -U orderuser -d order_db -c "
        CREATE TABLE IF NOT EXISTS orders (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            customer_name VARCHAR(255) NOT NULL,
            customer_email VARCHAR(255) NOT NULL,
            status VARCHAR(50) DEFAULT 'pending',
            total_amount DECIMAL(10,2) NOT NULL,
            discount_amount DECIMAL(10,2) DEFAULT 0,
            final_amount DECIMAL(10,2) NOT NULL,
            coupon_code VARCHAR(50),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        
        CREATE TABLE IF NOT EXISTS order_items (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
            product_id UUID NOT NULL,
            quantity INTEGER NOT NULL,
            unit_price DECIMAL(10,2) NOT NULL,
            total_price DECIMAL(10,2) NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    "
    
    log_success "Database migrations completed"
}

show_status() {
    log_info "Deployment Status:"
    echo
    
    # Show cluster info
    echo "Cluster: $CLUSTER_NAME"
    echo "Namespace: $NAMESPACE"
    echo
    
    # Show pods status
    echo "Pods:"
    kubectl get pods -n "$NAMESPACE"
    echo
    
    # Show services status
    echo "Services:"
    kubectl get services -n "$NAMESPACE"
    echo
    
    # Show API Gateway URL
    echo "API Gateway URL: http://localhost:8080"
    echo
    
    # Show useful commands
    echo "Useful commands:"
    echo "  kubectl get pods -n $NAMESPACE"
    echo "  kubectl logs -f deployment/api-gateway -n $NAMESPACE"
    echo "  kubectl logs -f deployment/product-service -n $NAMESPACE"
    echo "  kubectl logs -f deployment/coupon-service -n $NAMESPACE"
    echo "  kubectl logs -f deployment/order-service -n $NAMESPACE"
    echo "  kubectl port-forward service/api-gateway 8080:80 -n $NAMESPACE"
    echo
}

test_deployment() {
    log_info "Testing deployment..."
    
    # Wait a bit for services to be fully ready
    sleep 10
    
    # Test API Gateway health
    log_info "Testing API Gateway health..."
    if curl -f http://localhost:8080/health &> /dev/null; then
        log_success "API Gateway is healthy"
    else
        log_warning "API Gateway health check failed, but this might be normal during startup"
    fi
    
    # Test product service through gateway
    log_info "Testing product service..."
    if curl -f http://localhost:8080/api/product &> /dev/null; then
        log_success "Product service is accessible"
    else
        log_warning "Product service test failed, but this might be normal during startup"
    fi
}

cleanup() {
    log_info "Cleaning up..."
    
    read -p "Do you want to delete the Kind cluster? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        kind delete cluster --name="$CLUSTER_NAME"
        log_success "Cluster deleted"
    else
        log_info "Cluster preserved"
    fi
}

# Main execution
main() {
    case "${1:-deploy}" in
        "deploy")
            check_prerequisites
            create_cluster
            build_images
            deploy_services
            run_migrations
            show_status
            test_deployment
            ;;
        "cleanup")
            cleanup
            ;;
        "status")
            show_status
            ;;
        "test")
            test_deployment
            ;;
        "help")
            echo "Usage: $0 [deploy|cleanup|status|test|help]"
            echo "  deploy  - Deploy all services (default)"
            echo "  cleanup - Delete the Kind cluster"
            echo "  status  - Show deployment status"
            echo "  test    - Test the deployment"
            echo "  help    - Show this help message"
            ;;
        *)
            log_error "Unknown command: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"