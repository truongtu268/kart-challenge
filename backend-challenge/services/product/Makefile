.PHONY: build run test clean docker-build docker-run sqlc-generate migrate-up migrate-down

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=product-service
BINARY_UNIX=$(BINARY_NAME)_unix

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v cmd/server/main.go

# Run the application
run:
	$(GOBUILD) -o $(BINARY_NAME) -v cmd/server/main.go
	./$(BINARY_NAME)

# Test the application
test:
	$(GOTEST) -v ./...

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v cmd/server/main.go

# Docker commands
docker-build:
	docker build -t $(BINARY_NAME) .

docker-run:
	docker run -p 8080:8080 $(BINARY_NAME)

# SQLC commands
sqlc-generate:
	sqlc generate

# Database migration commands (requires golang-migrate)
migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down

# Development commands
dev:
	air

# Install development dependencies
install-dev:
	go install github.com/cosmtrek/air@latest
	go install github.com/kyleconroy/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest