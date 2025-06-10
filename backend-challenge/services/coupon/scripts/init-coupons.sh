#!/bin/bash

# Coupon Data Initialization Script
# This script sets up the database and initializes coupon data

set -e  # Exit on any error

echo "🚀 Starting coupon service initialization..."

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose is required but not installed."
    exit 1
fi

# Check if migrate is available
if ! command -v migrate &> /dev/null; then
    echo "⚠️  migrate CLI not found. Installing..."
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
fi

# Start PostgreSQL database
echo "📦 Starting PostgreSQL database..."
docker-compose up -d postgres

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
sleep 10

# Check if database is ready
echo "🔍 Checking database connection..."
max_attempts=30
attempt=1

while [ $attempt -le $max_attempts ]; do
    if docker-compose exec -T postgres pg_isready -U postgres -d coupon_service > /dev/null 2>&1; then
        echo "✅ Database is ready!"
        break
    fi
    
    if [ $attempt -eq $max_attempts ]; then
        echo "❌ Database failed to start after $max_attempts attempts"
        exit 1
    fi
    
    echo "⏳ Attempt $attempt/$max_attempts - Database not ready yet, waiting..."
    sleep 2
    attempt=$((attempt + 1))
done

# Run database migrations
echo "🔄 Running database migrations..."
make migrate-up

# Initialize coupon data from local files
echo "💾 Initializing coupon data from local files..."
make init-coupons

echo "🎉 Coupon service initialization completed successfully!"
echo ""
echo "📋 Next steps:"
echo "   • Run 'make run' to start the coupon service"
echo "   • Check logs with 'docker-compose logs postgres'"
echo "   • Stop database with 'make db-down'"