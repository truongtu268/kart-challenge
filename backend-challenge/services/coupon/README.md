# Coupon Service

A microservice for managing coupons and promo codes in the food ordering system.

## Quick Start

### Prerequisites

- Go 1.23+
- Docker and Docker Compose
- PostgreSQL (via Docker)

### Setup

1. **Easy setup with script (recommended):**
   ```bash
   make setup-script
   ```
   This will:
   - Start PostgreSQL database
   - Run database migrations
   - Read and process coupon files from local data directory
   - Initialize valid promo codes in the database

2. **Manual setup:**
   ```bash
   # Start database
   make db-up
   
   # Run migrations
   make migrate-up
   
   # Initialize coupon data
   make init-coupons
   ```

3. **Start the service:**
   ```bash
   make run
   ```

### Coupon Data Initialization

The service includes a command to initialize coupon data based on the project requirements:

- Reads three coupon files from local data directory
- Validates promo codes (8-10 characters, must appear in at least 2 files)
- Stores valid codes in the database

See [`cmd/init/README.md`](cmd/init/README.md) for detailed documentation.

### Required Data Files

Before running the initialization, you need to download the required coupon data files and place them in the `cmd/init/data/` directory:

**Download Links:**
- [couponbase1_unique](https://drive.google.com/file/d/1QjwEJribi6b9T0tzqTstuMvPu4o7ZGTx/view?usp=sharing)
- [couponbase2_unique](https://drive.google.com/file/d/1rSKoSwksdV9lCY7IovBbNK6yGc_ThFOx/view?usp=drive_link)
- [couponbase3_unique](https://drive.google.com/file/d/1b-ErGJWv2HJqPdAfwWJ-hzjFXj6yM3F7/view?usp=sharing)

**Setup Instructions:**
1. Download all three files from the Google Drive links above
2. Place them in the `cmd/init/data/` directory
3. Ensure the files are named exactly as shown (without any extensions)
4. Run the initialization command: `make init-coupons`

**Note:** These files are large (approximately 1GB each) and are stored externally to keep the repository size manageable.

### Available Commands

```bash
make build          # Build the server
make build-init     # Build the init command
make run            # Run the server
make init-coupons   # Initialize coupon data
make test           # Run tests
make clean          # Clean build artifacts
make deps           # Install dependencies
make migrate-up     # Run database migrations
make migrate-down   # Rollback migrations
make db-up          # Start database
make db-down        # Stop database
make setup          # Full setup (db + migrations + init)
make setup-script   # Easy setup using script
```

A comprehensive coupon management service built with Go, PostgreSQL, and SQLC for type-safe database operations.

## Features

- **CRUD Operations**: Create, read, update, and delete coupons
- **Coupon Validation**: Validate coupons for orders
- **Order Integration**: Track coupon usage in orders
- **Search & Pagination**: Search coupons by code with pagination
- **Type-Safe Database Operations**: Using SQLC for generated Go code from SQL queries
- **RESTful API**: Clean HTTP API with proper error handling
- **Database Migrations**: Automated database schema management
- **Health Checks**: Built-in health check endpoints
- **Graceful Shutdown**: Proper server shutdown handling
- **Structured Logging**: Using Zap for structured logging
- **Docker Support**: Complete containerization with Docker Compose

## Architecture

```
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── handler/http/    # HTTP handlers and routing
│   ├── service/         # Business logic layer
│   └── repository/      # Data access layer with SQLC
├── db/
│   ├── migrations/      # Database migration files
│   └── queries/         # SQL queries for SQLC
├── sqlc.yaml           # SQLC configuration
└── docker-compose.yml  # Docker services configuration
```

## Database Schema

### Coupons Table
```sql
CREATE TABLE coupons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coupon_code VARCHAR(255) UNIQUE NOT NULL,
    is_valid BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Coupon Order Table
```sql
CREATE TABLE coupon_order (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    coupon_id UUID NOT NULL REFERENCES coupons(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(order_id, coupon_id)
);
```

## SQLC Integration

This project uses [SQLC](https://sqlc.dev/) to generate type-safe Go code from SQL queries.

### SQLC Configuration

The `sqlc.yaml` file configures SQLC to:
- Read SQL queries from `db/queries/`
- Use database schema from `db/migrations/`
- Generate Go code in `internal/repository/sqlc/`
- Use PostgreSQL with pgx/v5 driver
- Generate JSON tags and interfaces

### SQL Queries

SQL queries are defined in `db/queries/` directory:

- `coupons.sql`: CRUD operations for coupons table
- `coupon_order.sql`: CRUD operations for coupon_order table

### Generated Code

SQLC generates:
- `models.go`: Go structs for database tables
- `coupons.sql.go`: Go functions for coupon queries
- `coupon_order.sql.go`: Go functions for coupon order queries
- `db.go`: Database interface and connection handling
- `querier.go`: Interface for all generated queries

### Code Generation

To generate Go code from SQL queries:

```bash
# Install SQLC
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Generate code
sqlc generate

# Or use make command
make generate
```

## Quick Start

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- Make (optional, for convenience commands)

### Running with Docker Compose

1. **Clone and navigate to the project:**
   ```bash
   cd services/coupon
   ```

2. **Start all services:**
   ```bash
   docker-compose up -d
   ```

3. **Check service health:**
   ```bash
   curl http://localhost:8080/health
   ```

4. **View logs:**
   ```bash
   docker-compose logs -f coupon-service
   ```

### Running Locally

1. **Install dependencies:**
   ```bash
   make install-tools
   ```

2. **Start PostgreSQL:**
   ```bash
   docker-compose up -d postgres
   ```

3. **Generate SQLC code:**
   ```bash
   make generate
   ```

4. **Run migrations:**
   ```bash
   make migrate-up
   ```

5. **Start the service:**
   ```bash
   make run
   ```

## API Endpoints

### Coupon Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/coupons` | Create a new coupon |
| GET | `/api/v1/coupons` | List coupons (with pagination and search) |
| GET | `/api/v1/coupons/:id` | Get coupon by ID |
| GET | `/api/v1/coupons/code/:code` | Get coupon by code |
| PUT | `/api/v1/coupons/:id` | Update coupon |
| DELETE | `/api/v1/coupons/:id` | Delete coupon |

### Coupon Operations

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/coupons/use` | Use coupon in an order |
| POST | `/api/v1/coupons/validate` | Validate coupon for an order |

### Order-Coupon Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/orders/:order_id/coupons` | Get all coupons for an order |
| DELETE | `/api/v1/orders/:order_id/coupons/:coupon_id` | Remove coupon from order |

### System Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/metrics` | Metrics endpoint |

## API Examples

### Create Coupon
```bash
curl -X POST http://localhost:8080/api/v1/coupons \
  -H "Content-Type: application/json" \
  -d '{
    "coupon_code": "SAVE20",
    "is_valid": true
  }'
```

### List Coupons
```bash
# List all coupons
curl "http://localhost:8080/api/v1/coupons?limit=10&offset=0"

# List only valid coupons
curl "http://localhost:8080/api/v1/coupons?valid_only=true"

# Search coupons
curl "http://localhost:8080/api/v1/coupons?q=SAVE"
```

### Validate Coupon
```bash
curl -X POST http://localhost:8080/api/v1/coupons/validate \
  -H "Content-Type: application/json" \
  -d '{
    "coupon_code": "SAVE20",
    "order_id": "123e4567-e89b-12d3-a456-426614174000"
  }'
```

### Use Coupon
```bash
curl -X POST http://localhost:8080/api/v1/coupons/use \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "123e4567-e89b-12d3-a456-426614174000",
    "coupon_id": "456e7890-e89b-12d3-a456-426614174001"
  }'
```

## Configuration

The service can be configured using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `DATABASE_URL` | `postgres://...` | PostgreSQL connection string |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `GIN_MODE` | `release` | Gin framework mode |

## Database Migrations

### Available Commands

```bash
# Run all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Check migration status
make migrate-status

# Force migration version (use with caution)
make migrate-force VERSION=1
```

### Manual Migration Commands

```bash
# Install migrate tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path db/migrations -database "postgres://user:password@localhost:5432/coupon_db?sslmode=disable" up

# Rollback migrations
migrate -path db/migrations -database "postgres://user:password@localhost:5432/coupon_db?sslmode=disable" down 1
```

## Development

### Building

```bash
# Build the application
make build

# Build Docker image
make docker-build
```

### Testing

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run all checks
make check
```

### Development Environment

```bash
# Setup complete development environment
make dev-setup

# Reset development environment
make dev-reset
```

## Docker

### Services

- **coupon-service**: Main application service
- **postgres**: PostgreSQL database
- **migration**: Database migration service
- **pgadmin** (optional): Database administration interface

### Commands

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f coupon-service

# Stop services
docker-compose down

# Rebuild and restart
docker-compose up --build -d
```

## Monitoring

### Health Checks

```bash
# Application health
curl http://localhost:8080/health

# Database connectivity is checked as part of application startup
```

### Logs

The service uses structured logging with Zap. Logs include:
- HTTP request/response details
- Database operation logs
- Error tracking
- Performance metrics

## Production Considerations

1. **Environment Variables**: Set appropriate values for production
2. **Database**: Use managed PostgreSQL service
3. **Logging**: Configure log aggregation
4. **Monitoring**: Add Prometheus metrics
5. **Security**: Implement authentication and authorization
6. **Rate Limiting**: Add rate limiting middleware
7. **CORS**: Configure CORS for your frontend domains

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check PostgreSQL is running
   - Verify connection string
   - Check network connectivity

2. **Migration Failed**
   - Check database permissions
   - Verify migration files
   - Check migration status

3. **SQLC Generation Failed**
   - Verify SQL syntax in query files
   - Check sqlc.yaml configuration
   - Ensure database schema is up to date

4. **Port Already in Use**
   - Change PORT environment variable
   - Stop conflicting services

### Logs

```bash
# View application logs
docker-compose logs coupon-service

# View database logs
docker-compose logs postgres

# Follow logs in real-time
docker-compose logs -f
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `make check` to ensure code quality
6. Submit a pull request

## License

This project is licensed under the MIT License.
