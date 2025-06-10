package service

import (
	"context"
	"fmt"
	"product-service/internal/repository"
	"product-service/internal/repository/postgres/sqlc"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ProductService provides business logic for product operations
type ProductService struct {
	productRepo *repository.ProductRepository
}

// NewProductService creates a new product service
func NewProductService(productRepo *repository.ProductRepository) *ProductService {
	return &ProductService{
		productRepo: productRepo,
	}
}

// ProductResponse represents a product response matching OpenAPI spec
type ProductResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(ctx context.Context, id string) (*ProductResponse, error) {
	// Parse string ID to UUID
	uuidID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	// Convert to pgtype.UUID
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(uuidID); err != nil {
		return nil, fmt.Errorf("failed to convert UUID: %w", err)
	}

	product, err := s.productRepo.GetProduct(ctx, pgUUID)
	if err != nil {
		return nil, err
	}

	return s.toProductResponse(&product), nil
}

// ListProducts retrieves all products
func (s *ProductService) ListProducts(ctx context.Context) ([]ProductResponse, error) {
	products, err := s.productRepo.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	productResponses := make([]ProductResponse, len(products))
	for i, product := range products {
		// Convert ListProductsRow to GetProductRow for consistency
		getProductRow := &sqlc.GetProductRow{
			ID:       product.ID,
			Name:     product.Name,
			Price:    product.Price,
			Category: product.Category,
		}
		productResponses[i] = *s.toProductResponse(getProductRow)
	}

	return productResponses, nil
}

// toProductResponse converts a product entity to response
func (s *ProductService) toProductResponse(product *sqlc.GetProductRow) *ProductResponse {
	// Convert pgtype.UUID to string
	var id string
	if product.ID.Valid {
		uuidVal := uuid.UUID(product.ID.Bytes)
		id = uuidVal.String()
	}

	// Convert pgtype.Text to string
	category := ""
	if product.Category.Valid {
		category = product.Category.String
	}

	// Convert pgtype.Numeric to float64
	var price float64
	if product.Price.Valid {
		if priceFloat, err := strconv.ParseFloat(product.Price.Int.String(), 64); err == nil {
			price = priceFloat
		}
	}

	return &ProductResponse{
		ID:       id,
		Name:     product.Name,
		Price:    price,
		Category: category,
	}
}
