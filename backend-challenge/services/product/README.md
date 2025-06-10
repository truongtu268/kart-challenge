# Product Service

A RESTful API service for managing products, built with Go, Gin, and PostgreSQL.

## Features

- Create, read, update, and delete products
- Product listing with pagination
- Product search by name
- Input validation and error handling
- Database migrations
- Docker support
- Health check endpoint

## API Endpoints

### OpenAPI Specification Routes
- `GET /api/v1/product` - List products with pagination
- `POST /api/v1/product` - Create a new product

### Additional CRUD Routes
- `GET /api/v1/products` - List products with pagination and search
- `POST /api/v1/products` - Create a new product
- `GET /api/v1/products/:id` - Get a product by ID
- `PUT /api/v1/products/:id` - Update a product
- `DELETE /api/v1/products/:id` - Delete a product
- `GET /health` - Health check

## Product Schema

```json
{
  "id": "uuid",
  "name": "string",
  "price": "number",
  "description": "string",
  "created_at": "string",
  "updated_at": "string"
}
```

## Prerequisites

- Go 1.21+
- PostgreSQL 12+
- Docker (optional)
- SQLC
- golang-migrate

## Setup

### 1. Install Development Dependencies

```bash
make install-dev
```

### 2. Generate Database Code

```bash
make sqlc-generate
```

### 3. Set Environment Variables

```bash
export DATABASE_URL="postgres://username:password@localhost:5432/product_db?sslmode=disable"
export PORT=8080
export LOG_LEVEL=info
```

### 4. Run Database Migrations

```bash
make migrate-up
```

### 5. Build and Run

```bash
make run
```

## Development

### Using Air for Hot Reload

```bash
make dev
```

### Running Tests

```bash
make test
```

### Building for Production

```bash
make build
```

## Docker

### Build Docker Image

```bash
make docker-build
```

### Run with Docker

```bash
make docker-run
```

## Database Schema

The service uses a PostgreSQL database with the following table:

```sql
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## Architecture

The service follows a clean architecture pattern:

- **Handler Layer**: HTTP request handling and response formatting
- **Service Layer**: Business logic and validation
- **Repository Layer**: Database operations and data access
- **Config**: Application configuration management

## Error Handling

The API returns consistent error responses:

```json
{
  "error": "error_code",
  "message": "Human readable error message"
}
```

## Success Responses

Successful operations return:

```json
{
  "message": "Operation completed successfully",
  "data": { /* response data */ }
}
```

## Query Parameters

### List Products
- `limit`: Number of products to return (default: 10, max: 100)
- `offset`: Number of products to skip (default: 0)
- `search`: Search products by name (optional)

Example: `GET /api/v1/products?limit=20&offset=0&search=pizza`

## Contributing

1. Follow the existing code structure and patterns
2. Add tests for new features
3. Update documentation as needed
4. Use `make sqlc-generate` after modifying SQL queries
5. Run `make test` before submitting changes