# Order Service

The Order Service is a microservice responsible for managing customer orders in the Kart Challenge e-commerce platform. It handles order placement, validation, and retrieval operations.

## Features

- **Order Management**: Create and retrieve orders
- **Product Validation**: Verify product information before order placement
- **Coupon Integration**: Validate and apply coupon codes
- **Database Integration**: PostgreSQL with SQLC for type-safe queries
- **Health Checks**: Built-in health and metrics endpoints
- **Observability**: Structured logging with Zap and OpenTelemetry tracing

## API Endpoints

### Public Endpoints

- `POST /order` - Place a new order
- `GET /health` - Health check endpoint
- `GET /metrics` - Metrics endpoint

### Internal Endpoints

- `GET /internal/orders` - List orders (with pagination)
- `GET /internal/orders/:id` - Get order by ID

## API Documentation

### Place Order

```http
POST /order
Content-Type: application/json

{
  "customer_name": "John Doe",
  "customer_email": "john@example.com",
  "items": [
    {
      "product_id": "550e8400-e29b-41d4-a716-446655440000",
      "quantity": 2
    }
  ],
  "coupon_code": "SAVE10"
}
```

**Response:**
```json
{
  "order_id": "123e4567-e89b-12d3-a456-426614174000",
  "status": "confirmed",
  "total_amount": 19.98,
  "discount_amount": 2.00,
  "final_amount": 17.98
}
```

### Get Order

```http
GET /internal/orders/123e4567-e89b-12d3-a456-426614174000
```

**Response:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "customer_name": "John Doe",
  "customer_email": "john@example.com",
  "status": "confirmed",
  "total_amount": 19.98,
  "discount_amount": 2.00,
  "final_amount": 17.98,
  "items": [
    {
      "product_id": "550e8400-e29b-41d4-a716-446655440000",
      "quantity": 2,
      "unit_price": 9.99,
      "total_price": 19.98
    }
  ],
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

## Prerequisites

- Go 1.23 or later
- PostgreSQL 15+
- Docker and Docker Compose (for containerized deployment)
- SQLC (for code generation)
- golang-migrate (for database migrations)

## Installation

### Local Development

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd services/order
   ```

2. **Install dependencies:**
   ```bash
   make deps
   ```

3. **Generate code:**
   ```bash
   make generate
   ```

4. **Start the database:**
   ```bash
   make db-up
   ```

5. **Run migrations:**
   ```bash
   make db-migrate
   ```

6. **Start the service:**
   ```bash
   make dev
   ```

### Docker Deployment

1. **Build and run with Docker Compose:**
   ```bash
   make docker-run
   ```

2. **Stop services:**
   ```bash
   make docker-down
   ```

## Configuration

The service can be configured using environment variables or a YAML configuration file.

### Environment Variables

- `DATABASE_URL` - PostgreSQL connection string
- `ORDER_SERVER_HOST` - Server host (default: 0.0.0.0)
- `ORDER_SERVER_PORT` - Server port (default: 8081)
- `ORDER_SERVICES_PRODUCT_API` - Product service URL
- `ORDER_SERVICES_COUPON_API` - Coupon service URL
- `ORDER_LOG_LEVEL` - Log level (debug, info, warn, error)
- `ORDER_LOG_FORMAT` - Log format (json, console)

### Configuration File

Create a `config.yaml` file:

```yaml
database:
  host: localhost
  port: 5433
  user: orderuser
  password: orderpass
  name: order_db
  sslmode: disable

server:
  host: 0.0.0.0
  port: 8081

log:
  level: info
  format: json

services:
  product_api: http://localhost:8080
  coupon_api: http://localhost:8082
```

## Database Schema

The service uses two main tables:

### Orders Table
- `id` (UUID, Primary Key)
- `customer_name` (VARCHAR)
- `customer_email` (VARCHAR)
- `status` (VARCHAR)
- `total_amount` (DECIMAL)
- `discount_amount` (DECIMAL)
- `final_amount` (DECIMAL)
- `coupon_code` (VARCHAR, Optional)
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

### Order Items Table
- `id` (UUID, Primary Key)
- `order_id` (UUID, Foreign Key)
- `product_id` (UUID)
- `quantity` (INTEGER)
- `unit_price` (DECIMAL)
- `total_price` (DECIMAL)
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

## Development

### Available Make Commands

```bash
make help          # Show available commands
make dev           # Run in development mode
make build         # Build binary
make test          # Run tests
make test-coverage # Run tests with coverage
make lint          # Run linter
make fmt           # Format code
make generate      # Generate SQLC code
make db-up         # Start database
make db-migrate    # Run migrations
make docker-build  # Build Docker image
make test-api      # Test API endpoints
```

### Code Generation

This project uses SQLC for generating type-safe Go code from SQL queries:

```bash
make generate
```

### Database Migrations

Migrations are located in `db/migrations/` and can be run using:

```bash
make db-migrate      # Apply migrations
make db-migrate-down # Rollback migrations
make db-reset        # Reset database
```

## Testing

### Unit Tests

```bash
make test
```

### API Testing

```bash
make test-api
```

### Manual Testing

1. **Health Check:**
   ```bash
   curl http://localhost:8081/health
   ```

2. **Place Order:**
   ```bash
   curl -X POST http://localhost:8081/order \
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

## Integration with Other Services

### Product Service
The order service validates product information by calling the Product Service API before placing orders.

### Coupon Service
The order service validates and applies coupon codes by integrating with the Coupon Service API.

## Monitoring and Observability

- **Health Checks**: `/health` endpoint for service health monitoring
- **Metrics**: `/metrics` endpoint for Prometheus metrics
- **Logging**: Structured JSON logging with Zap
- **Tracing**: OpenTelemetry integration for distributed tracing

## Error Handling

The service returns appropriate HTTP status codes and error messages:

- `400 Bad Request` - Invalid request data
- `404 Not Found` - Order not found
- `422 Unprocessable Entity` - Invalid product or coupon
- `500 Internal Server Error` - Server errors

## Security Considerations

- Input validation on all endpoints
- SQL injection prevention through parameterized queries
- Non-root user in Docker container
- Health check endpoints for monitoring

## Contributing

1. Follow Go coding standards
2. Write tests for new features
3. Update documentation
4. Run linting and formatting before committing

## License

This project is part of the Kart Challenge backend services.