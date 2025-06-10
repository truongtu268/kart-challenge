package repository

import (
	"context"
	"database/sql"
	"fmt"

	query "product-service/internal/repository/postgres/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

// ProductRepository provides database operations for products
type ProductRepository struct {
	db      query.DBTX
	queries *query.Queries
}

// NewProductRepository creates a new product repository
func NewProductRepository(db query.DBTX) *ProductRepository {
	return &ProductRepository{
		db:      db,
		queries: query.New(db),
	}
}

// GetProduct retrieves a product by ID
func (r *ProductRepository) GetProduct(ctx context.Context, id pgtype.UUID) (query.GetProductRow, error) {
	product, err := r.queries.GetProduct(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return query.GetProductRow{}, fmt.Errorf("product not found")
		}
		return query.GetProductRow{}, fmt.Errorf("failed to get product: %w", err)
	}
	return product, nil
}

// ListProducts retrieves all products
func (r *ProductRepository) ListProducts(ctx context.Context) ([]query.ListProductsRow, error) {
	products, err := r.queries.ListProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	return products, nil
}
