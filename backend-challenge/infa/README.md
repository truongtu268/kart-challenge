# Kart Challenge Infrastructure

This directory contains the infrastructure configuration for deploying the Kart Challenge microservices to a Kubernetes cluster using Kind (Kubernetes in Docker).

## Architecture Overview

The infrastructure consists of:

- **API Gateway** (NGINX) - Routes requests to appropriate services
- **Product Service** - Manages product catalog
- **Coupon Service** - Handles coupon validation and usage
- **Order Service** - Processes customer orders
- **PostgreSQL Database** - Shared database with separate schemas for each service

```
┌─────────────────┐
│   Frontend      │
└─────────┬───────┘
          │
┌─────────▼───────┐
│  API Gateway    │ :8080
│    (NGINX)      │
└─────────┬───────┘
          │
    ┌─────┼─────┐
    │     │     │
┌───▼──┐ ┌▼───┐ ┌▼─────┐
│Product│ │Coupon│ │Order │
│:8080  │ │:8082 │ │:8081 │
└───┬──┘ └┬───┘ └┬─────┘
    │     │      │
    └─────┼──────┘
          │
    ┌─────▼─────┐
    │PostgreSQL │
    │   :5432   │
    └───────────┘
```

## Prerequisites

Before deploying, ensure you have the following installed:

### Required Tools

1. **Docker** - Container runtime
   ```bash
   # macOS
   brew install docker
   # Or download Docker Desktop from https://www.docker.com/products/docker-desktop
   ```

2. **Kind** - Kubernetes in Docker
   ```bash
   # macOS
   brew install kind
   # Or download from https://kind.sigs.k8s.io/docs/user/quick-start/#installation
   ```

3. **kubectl** - Kubernetes CLI
   ```bash
   # macOS
   brew install kubectl
   # Or download from https://kubernetes.io/docs/tasks/tools/install-kubectl/
   ```

4. **Go** - For building services (if not using pre-built images)
   ```bash
   # macOS
   brew install go
   ```

### Verify Installation

```bash
# Check Docker
docker --version
docker info

# Check Kind
kind --version

# Check kubectl
kubectl version --client

# Check Go
go version
```

## Quick Start

### 1. Deploy Everything

```bash
# Navigate to infrastructure directory
cd infa

# Deploy all services
./deploy.sh
```

This script will:
- Create a Kind cluster
- Build Docker images for all services
- Deploy PostgreSQL database
- Deploy all microservices
- Deploy API Gateway
- Run database migrations
- Show deployment status

### 2. Access the API

Once deployed, the API Gateway will be available at:
- **API Gateway**: http://localhost:8080
- **Health Check**: http://localhost:8080/health

### 3. Test the Services

```bash
# Test API Gateway
curl http://localhost:8080/health

# Test Product Service
curl http://localhost:8080/api/product

# Test Coupon Service
curl http://localhost:8080/api/coupon/validate/SAVE10

# Test Order Service
curl -X POST http://localhost:8080/api/order \
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
  }'
```

## Deployment Commands

### Deploy Services

```bash
# Full deployment
./deploy.sh deploy

# Or just
./deploy.sh
```

### Check Status

```bash
# Show deployment status
./deploy.sh status

# Test deployment
./deploy.sh test
```

### Cleanup

```bash
# Delete the Kind cluster
./deploy.sh cleanup
```

### Help

```bash
# Show available commands
./deploy.sh help
```

## Manual Deployment

If you prefer to deploy manually:

### 1. Create Kind Cluster

```bash
kind create cluster --config=kind-config.yaml --name=kart-challenge
```

### 2. Build and Load Images

```bash
# Build product service
cd ../services/product
docker build -t product-service:latest .
kind load docker-image product-service:latest --name=kart-challenge

# Build coupon service
cd ../coupon
docker build -t coupon-service:latest .
kind load docker-image coupon-service:latest --name=kart-challenge

# Build order service
cd ../order
docker build -t order-service:latest .
kind load docker-image order-service:latest --name=kart-challenge
```

### 3. Deploy to Kubernetes

```bash
cd ../../infa

# Apply all manifests
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/product-service.yaml
kubectl apply -f k8s/coupon-service.yaml
kubectl apply -f k8s/order-service.yaml
kubectl apply -f k8s/api-gateway.yaml
```

### 4. Wait for Deployment

```bash
# Wait for all pods to be ready
kubectl wait --for=condition=available --timeout=300s deployment --all -n kart-challenge
```

## Configuration

### Kind Cluster Configuration

The Kind cluster is configured with:
- 1 control-plane node
- 2 worker nodes
- Port forwarding for API Gateway (8080:30080)
- Custom networking configuration

### Service Configuration

Each service is configured with:
- **Replicas**: 2 for high availability
- **Resource Limits**: CPU and memory constraints
- **Health Checks**: Liveness and readiness probes
- **Environment Variables**: Service-specific configuration

### Database Configuration

PostgreSQL is configured with:
- **Persistent Storage**: 1Gi PVC
- **Multiple Databases**: Separate database for each service
- **Health Checks**: PostgreSQL-specific health checks
- **Initialization Scripts**: Create databases and users

## API Gateway Routes

The NGINX API Gateway routes requests as follows:

| Path | Target Service | Description |
|------|----------------|-------------|
| `/api/product/*` | Product Service | Product catalog operations |
| `/api/coupon/*` | Coupon Service | Coupon validation and usage |
| `/api/order/*` | Order Service | Order placement and retrieval |
| `/internal/*` | Various | Internal service-to-service communication |
| `/health` | API Gateway | Health check endpoint |
| `/` | API Gateway | API information |

### CORS Configuration

The API Gateway is configured with CORS headers to allow frontend access:
- **Origins**: `*` (all origins)
- **Methods**: `GET, POST, PUT, DELETE, OPTIONS`
- **Headers**: Standard headers including `Authorization`

### Rate Limiting

The API Gateway includes rate limiting:
- **Rate**: 10 requests per second per IP
- **Burst**: 20 requests
- **Zone**: 10MB memory allocation

## Monitoring and Debugging

### View Logs

```bash
# API Gateway logs
kubectl logs -f deployment/api-gateway -n kart-challenge

# Product service logs
kubectl logs -f deployment/product-service -n kart-challenge

# Coupon service logs
kubectl logs -f deployment/coupon-service -n kart-challenge

# Order service logs
kubectl logs -f deployment/order-service -n kart-challenge

# PostgreSQL logs
kubectl logs -f deployment/postgres -n kart-challenge
```

### Check Pod Status

```bash
# List all pods
kubectl get pods -n kart-challenge

# Describe a specific pod
kubectl describe pod <pod-name> -n kart-challenge

# Get pod events
kubectl get events -n kart-challenge --sort-by='.lastTimestamp'
```

### Port Forwarding

```bash
# Forward API Gateway port
kubectl port-forward service/api-gateway 8080:80 -n kart-challenge

# Forward individual service ports
kubectl port-forward service/product-service 8080:8080 -n kart-challenge
kubectl port-forward service/coupon-service 8082:8082 -n kart-challenge
kubectl port-forward service/order-service 8081:8081 -n kart-challenge

# Forward PostgreSQL port
kubectl port-forward service/postgres 5432:5432 -n kart-challenge
```

### Database Access

```bash
# Connect to PostgreSQL
kubectl exec -it deployment/postgres -n kart-challenge -- psql -U kartuser -d kart_db

# Or connect to specific service database
kubectl exec -it deployment/postgres -n kart-challenge -- psql -U productuser -d product_db
kubectl exec -it deployment/postgres -n kart-challenge -- psql -U couponuser -d coupon_db
kubectl exec -it deployment/postgres -n kart-challenge -- psql -U orderuser -d order_db
```

## Troubleshooting

### Common Issues

1. **Pods stuck in Pending state**
   ```bash
   # Check node resources
   kubectl describe nodes
   
   # Check pod events
   kubectl describe pod <pod-name> -n kart-challenge
   ```

2. **Services not accessible**
   ```bash
   # Check service endpoints
   kubectl get endpoints -n kart-challenge
   
   # Check service configuration
   kubectl describe service <service-name> -n kart-challenge
   ```

3. **Database connection issues**
   ```bash
   # Check PostgreSQL logs
   kubectl logs deployment/postgres -n kart-challenge
   
   # Test database connectivity
   kubectl exec -it deployment/postgres -n kart-challenge -- pg_isready
   ```

4. **Image pull issues**
   ```bash
   # Rebuild and reload images
   cd ../services/product
   docker build -t product-service:latest .
   kind load docker-image product-service:latest --name=kart-challenge
   
   # Restart deployment
   kubectl rollout restart deployment/product-service -n kart-challenge
   ```

### Reset Deployment

```bash
# Delete all resources
kubectl delete namespace kart-challenge

# Or delete entire cluster
kind delete cluster --name=kart-challenge

# Then redeploy
./deploy.sh
```

## Development Workflow

### Making Changes to Services

1. **Modify service code**
2. **Rebuild image**:
   ```bash
   cd services/<service-name>
   docker build -t <service-name>:latest .
   kind load docker-image <service-name>:latest --name=kart-challenge
   ```
3. **Restart deployment**:
   ```bash
   kubectl rollout restart deployment/<service-name> -n kart-challenge
   ```

### Adding New Services

1. **Create service directory** under `services/`
2. **Add Dockerfile** for the service
3. **Create Kubernetes manifests** in `infa/k8s/`
4. **Update deployment script** to include new service
5. **Update API Gateway configuration** if needed

## Production Considerations

For production deployment, consider:

1. **Security**:
   - Use proper secrets management
   - Enable TLS/SSL
   - Implement proper authentication
   - Network policies

2. **Scalability**:
   - Horizontal Pod Autoscaler
   - Cluster autoscaling
   - Load balancing

3. **Monitoring**:
   - Prometheus metrics
   - Grafana dashboards
   - Log aggregation
   - Distributed tracing

4. **Backup**:
   - Database backups
   - Persistent volume snapshots
   - Configuration backups

5. **CI/CD**:
   - Automated testing
   - Image scanning
   - Deployment pipelines
   - GitOps workflows

## Contributing

When contributing to the infrastructure:

1. Test changes locally with Kind
2. Update documentation
3. Follow Kubernetes best practices
4. Ensure backward compatibility
5. Add appropriate resource limits
6. Include health checks

## License

This infrastructure configuration is part of the Kart Challenge project.