# Build the server
build:
	go build -o bin/server ./cmd/server

# Build the init command
build-init:
	go build -o bin/init ./cmd/init

# Run the server
run:
	go run ./cmd/server

# Run the initialization command
init-coupons:
	go run ./cmd/init

# Import uncompressed couponbase1 file directly into PostgreSQL
import-couponbase1:
	@echo "Importing couponbase1 file into PostgreSQL..."
	@docker cp ./cmd/init/data/couponbase1 coupon-postgres:/tmp/couponbase1
	@docker exec coupon-postgres psql -U postgres -d coupon_service -c "CREATE TABLE temp_1_coupons (coupon_code TEXT); COPY temp_1_coupons FROM '/tmp/couponbase1' WITH (FORMAT text);"
	@docker exec coupon-postgres rm -f /tmp/couponbase1
	@echo "Import completed couponbase1 successfully!"

import-couponbase2:
	@echo "Importing couponbase2 file into PostgreSQL..."
	@docker cp ./cmd/init/data/couponbase2 coupon-postgres:/tmp/couponbase2
	@docker exec coupon-postgres psql -U postgres -d coupon_service -c "CREATE TABLE temp_2_coupons (coupon_code TEXT); COPY temp_2_coupons FROM '/tmp/couponbase2' WITH (FORMAT text);"
	@docker exec coupon-postgres rm -f /tmp/couponbase2
	@echo "Import completed couponbase2 successfully!"

import-couponbase3:
	@echo "Importing couponbase3 file into PostgreSQL..."
	@docker cp ./cmd/init/data/couponbase3 coupon-postgres:/tmp/couponbase3
	@docker exec coupon-postgres psql -U postgres -d coupon_service -c "CREATE TABLE temp_3_coupons (coupon_code TEXT); COPY temp_3_coupons FROM '/tmp/couponbase3' WITH (FORMAT text);"
	@docker exec coupon-postgres rm -f /tmp/couponbase3
	@echo "Import completed couponbase3 successfully!"

# Run tests
test:
	go test ./...

# Clean build artifactscc
clean:
	rm -rf bin/

# Install dependencies
deps:
	go mod download
	go mod tidy

# Database migrations up
migrate-up:
	migrate -path db/migrations -database "postgres://postgres:postgres@localhost:5432/coupon_service?sslmode=disable" up

# Database migrations down
migrate-down:
	migrate -path db/migrations -database "postgres://postgres:postgres@localhost:5432/coupon_service?sslmode=disable" down

# Start database with docker-compose
db-up:
	docker-compose up -d postgres

# Stop database
db-down:
	docker-compose down

# Wait for database to be ready
wait-db:
	@echo "Waiting for database to be ready..."
	@for i in {1..30}; do \
		if pg_isready -h localhost -p 5432 -U postgres > /dev/null 2>&1; then \
			echo "Database is ready!"; \
			break; \
		fi; \
		echo "Waiting for database... ($$i/30)"; \
		sleep 2; \
	done

import-coupons: import-couponbase1 import-couponbase2 import-couponbase3

# Full setup: start db, wait for it, run migrations, init coupons
setup: db-up wait-db migrate-up import-coupons

# Easy setup using script
setup-script:
	./scripts/init-coupons.sh

# Make script executable
setup-permissions:
	chmod +x scripts/init-coupons.sh

.PHONY: build build-init run init-coupons import-couponbase1 test clean deps migrate-up migrate-down db-up db-down wait-db setup setup-script setup-permissions import-couponbase1 import-couponbase2 import-couponbase3 import-coupons create-unique_coupons

#CREATE table unique_coupon (
#    coupon_code text primary key,
#    count integer DEFAULT 1
#);

#INSERT INTO unique_coupon (coupon_code)
#select temp_1_coupons.coupon_code from temp_1_coupons
#ON CONFLICT (coupon_code)
#    DO UPDATE SET count = unique_coupon.count + 1;