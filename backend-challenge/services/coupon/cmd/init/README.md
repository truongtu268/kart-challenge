# Coupon Data Initialization

This command initializes the coupon database with valid promo codes based on the validation rules specified in the project requirements.

## Overview

The initialization process:

1. Reads three coupon files from the local data directory:
   - `./data/couponbase1.gz`
   - `./data/couponbase2.gz` 
   - `./data/couponbase3.gz`

2. Extracts and processes each file to find potential promo codes

3. Validates promo codes according to the rules:
   - Must be 8-10 characters long
   - Must appear in at least 2 of the 3 files
   - Must contain only uppercase letters and numbers

4. Stores valid promo codes in the database

## Usage

### Prerequisites

1. Ensure PostgreSQL database is running
2. Run database migrations first:
   ```bash
   make migrate-up
   ```

### Running the Initialization

```bash
# Using make command
make init-coupons

# Or run directly with go
go run ./cmd/init

# Or build and run
go build -o bin/init ./cmd/init
./bin/init
```

### Configuration

The command uses the same configuration as the main server. You can configure it using environment variables:

```bash
# Database configuration
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=coupon_service
export DB_SSLMODE=disable

# Logging configuration
export LOG_LEVEL=info
export LOG_FORMAT=json
```

### Full Setup

To set up everything from scratch:

```bash
# Start database, run migrations, and initialize coupons
make setup
```

## Implementation Details

### File Processing

- Reads gzipped files from local data directory
- Decompresses files using gzip reader
- Scans content line by line to extract potential coupon codes
- Uses regex pattern `\b[A-Z0-9]{8,10}\b` to match 8-10 character alphanumeric codes

### Validation Logic

- Extracts all potential codes from each file
- Counts occurrences across all three files
- Only codes appearing in 2+ files are considered valid
- Stores valid codes in the database with `is_valid=true`

### Error Handling

- Continues processing even if individual coupon creation fails
- Logs warnings for failed coupon insertions (e.g., duplicates)
- Fails fast on critical errors (database connection, file reading)

## Example Output

```
{"level":"info","ts":1234567890,"msg":"Starting coupon data initialization"}
{"level":"info","ts":1234567890,"msg":"Reading coupon file","path":"./data/couponbase1.gz"}
{"level":"info","ts":1234567890,"msg":"Extracted potential coupons from file","count":1250}
{"level":"info","ts":1234567890,"msg":"Reading coupon file","path":"./data/couponbase2.gz"}
{"level":"info","ts":1234567890,"msg":"Extracted potential coupons from file","count":1180}
{"level":"info","ts":1234567890,"msg":"Reading coupon file","path":"./data/couponbase3.gz"}
{"level":"info","ts":1234567890,"msg":"Extracted potential coupons from file","count":1320}
{"level":"info","ts":1234567890,"msg":"Found valid coupons","count":45}
{"level":"info","ts":1234567890,"msg":"Coupon data initialization completed successfully"}
```