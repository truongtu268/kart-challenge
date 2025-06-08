package service

import (
	"context"
	"coupon-service/internal/repository/postgres/sqlc"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"coupon-service/internal/repository"
)

// CouponService provides business logic for coupon operations
type CouponService struct {
	couponRepo      *repository.CouponRepository
	couponOrderRepo *repository.CouponOrderRepository
}

// NewCouponService creates a new coupon service
func NewCouponService(couponRepo *repository.CouponRepository, couponOrderRepo *repository.CouponOrderRepository) *CouponService {
	return &CouponService{
		couponRepo:      couponRepo,
		couponOrderRepo: couponOrderRepo,
	}
}

// CreateCouponRequest represents the request to create a coupon
type CreateCouponRequest struct {
	CouponCode string `json:"coupon_code" binding:"required,min=3,max=50"`
	IsValid    bool   `json:"is_valid"`
}

// UpdateCouponRequest represents the request to update a coupon
type UpdateCouponRequest struct {
	CouponCode string `json:"coupon_code" binding:"required,min=3,max=50"`
	IsValid    bool   `json:"is_valid"`
}

// UseCouponRequest represents the request to use a coupon
type UseCouponRequest struct {
	OrderID  uuid.UUID `json:"order_id" binding:"required"`
	CouponID uuid.UUID `json:"coupon_id" binding:"required"`
}

// ValidateCouponRequest represents the request to validate a coupon
type ValidateCouponRequest struct {
	CouponCode string    `json:"coupon_code" binding:"required"`
	OrderID    uuid.UUID `json:"order_id" binding:"required"`
}

// CouponResponse represents a coupon response
type CouponResponse struct {
	ID         uuid.UUID `json:"id"`
	CouponCode string    `json:"coupon_code"`
	IsValid    bool      `json:"is_valid"`
	CreatedAt  string    `json:"created_at"`
	UpdatedAt  string    `json:"updated_at"`
}

// CouponListResponse represents a paginated list of coupons
type CouponListResponse struct {
	Coupons []CouponResponse `json:"coupons"`
	Total   int64            `json:"total"`
	Limit   int32            `json:"limit"`
	Offset  int32            `json:"offset"`
}

// ValidationResult represents the result of coupon validation
type ValidationResult struct {
	Valid   bool            `json:"valid"`
	Message string          `json:"message"`
	Coupon  *CouponResponse `json:"coupon,omitempty"`
}

// CreateCoupon creates a new coupon
func (s *CouponService) CreateCoupon(ctx context.Context, req CreateCouponRequest) (*CouponResponse, error) {
	// Validate coupon code format
	if err := s.validateCouponCode(req.CouponCode); err != nil {
		return nil, err
	}

	// Check if coupon code already exists
	existingCoupon, err := s.couponRepo.GetCouponByCode(ctx, req.CouponCode)
	if err == nil && existingCoupon != nil {
		return nil, fmt.Errorf("coupon code already exists")
	}

	coupon, err := s.couponRepo.CreateCoupon(ctx, req.CouponCode, req.IsValid)
	if err != nil {
		return nil, fmt.Errorf("failed to create coupon: %w", err)
	}

	return s.toCouponResponse(coupon), nil
}

// GetCoupon retrieves a coupon by ID
func (s *CouponService) GetCoupon(ctx context.Context, id uuid.UUID) (*CouponResponse, error) {
	coupon, err := s.couponRepo.GetCoupon(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toCouponResponse(coupon), nil
}

// GetCouponByCode retrieves a coupon by code
func (s *CouponService) GetCouponByCode(ctx context.Context, code string) (*CouponResponse, error) {
	coupon, err := s.couponRepo.GetCouponByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	return s.toCouponResponse(coupon), nil
}

// ListCoupons retrieves coupons with pagination
func (s *CouponService) ListCoupons(ctx context.Context, limit, offset int32, validOnly bool) (*CouponListResponse, error) {
	var coupons []sqlc.Coupon
	var total int64
	var err error

	if validOnly {
		coupons, err = s.couponRepo.ListValidCoupons(ctx, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to list valid coupons: %w", err)
		}
		total, err = s.couponRepo.CountValidCoupons(ctx)
	} else {
		coupons, err = s.couponRepo.ListCoupons(ctx, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to list coupons: %w", err)
		}
		total, err = s.couponRepo.CountCoupons(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to count coupons: %w", err)
	}

	couponResponses := make([]CouponResponse, len(coupons))
	for i, coupon := range coupons {
		couponResponses[i] = *s.toCouponResponse(&coupon)
	}

	return &CouponListResponse{
		Coupons: couponResponses,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

// UpdateCoupon updates a coupon
func (s *CouponService) UpdateCoupon(ctx context.Context, id uuid.UUID, req UpdateCouponRequest) (*CouponResponse, error) {
	// Validate coupon code format
	if err := s.validateCouponCode(req.CouponCode); err != nil {
		return nil, err
	}

	// Check if the coupon exists
	existingCoupon, err := s.couponRepo.GetCoupon(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if the new coupon code is already used by another coupon
	if existingCoupon.CouponCode != req.CouponCode {
		couponByCode, err := s.couponRepo.GetCouponByCode(ctx, req.CouponCode)
		if err == nil && couponByCode != nil && couponByCode.ID != id {
			return nil, fmt.Errorf("coupon code already exists")
		}
	}

	coupon, err := s.couponRepo.UpdateCoupon(ctx, id, req.CouponCode, req.IsValid)
	if err != nil {
		return nil, fmt.Errorf("failed to update coupon: %w", err)
	}

	return s.toCouponResponse(coupon), nil
}

// DeleteCoupon deletes a coupon
func (s *CouponService) DeleteCoupon(ctx context.Context, id uuid.UUID) error {
	// Check if coupon exists
	_, err := s.couponRepo.GetCoupon(ctx, id)
	if err != nil {
		return err
	}

	// Check if coupon is used in any orders
	couponOrders, err := s.couponOrderRepo.GetCouponOrdersByCouponID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check coupon usage: %w", err)
	}

	if len(couponOrders) > 0 {
		return fmt.Errorf("cannot delete coupon: it is used in %d order(s)", len(couponOrders))
	}

	err = s.couponRepo.DeleteCoupon(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete coupon: %w", err)
	}

	return nil
}

// UseCoupon applies a coupon to an order
func (s *CouponService) UseCoupon(ctx context.Context, req UseCouponRequest) error {
	// Check if coupon exists and is valid
	coupon, err := s.couponRepo.GetCoupon(ctx, req.CouponID)
	if err != nil {
		return fmt.Errorf("coupon not found: %w", err)
	}

	if !coupon.IsValid {
		return fmt.Errorf("coupon is not valid")
	}

	// Check if coupon is already used in this order
	alreadyUsed, err := s.couponOrderRepo.CheckCouponUsageInOrder(ctx, req.OrderID, req.CouponID)
	if err != nil {
		return fmt.Errorf("failed to check coupon usage: %w", err)
	}

	if alreadyUsed {
		return fmt.Errorf("coupon is already used in this order")
	}

	// Create coupon order record
	_, err = s.couponOrderRepo.CreateCouponOrder(ctx, req.OrderID, req.CouponID)
	if err != nil {
		return fmt.Errorf("failed to use coupon: %w", err)
	}

	return nil
}

// ValidateCoupon validates a coupon for an order
func (s *CouponService) ValidateCoupon(ctx context.Context, req ValidateCouponRequest) (*ValidationResult, error) {
	if len(req.CouponCode) < 8 || len(req.CouponCode) > 10 {
		return &ValidationResult{
			Valid:   false,
			Message: "Coupon wrong format",
		}, nil
	}

	// Get coupon by code
	coupon, err := s.couponRepo.GetCouponByCode(ctx, req.CouponCode)
	if err != nil {
		return &ValidationResult{
			Valid:   false,
			Message: "Coupon not found",
		}, nil
	}

	// Check if coupon is valid
	if !coupon.IsValid {
		return &ValidationResult{
			Valid:   false,
			Message: "Coupon is not valid",
			Coupon:  s.toCouponResponse(coupon),
		}, nil
	}

	// Check if coupon is already used in this order
	alreadyUsed, err := s.couponOrderRepo.CheckCouponUsageInOrder(ctx, req.OrderID, coupon.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check coupon usage: %w", err)
	}

	if alreadyUsed {
		return &ValidationResult{
			Valid:   false,
			Message: "Coupon is already used in this order",
			Coupon:  s.toCouponResponse(coupon),
		}, nil
	}

	return &ValidationResult{
		Valid:   true,
		Message: "Coupon is valid and can be used",
		Coupon:  s.toCouponResponse(coupon),
	}, nil
}

// SearchCoupons searches coupons by code pattern
func (s *CouponService) SearchCoupons(ctx context.Context, query string, limit, offset int32) (*CouponListResponse, error) {
	if strings.TrimSpace(query) == "" {
		return s.ListCoupons(ctx, limit, offset, false)
	}

	// Add wildcards for pattern matching
	pattern := "%" + strings.ToUpper(query) + "%"

	coupons, err := s.couponRepo.SearchCouponsByCode(ctx, pattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search coupons: %w", err)
	}

	// For search, we don't have an exact count, so we estimate
	total := int64(len(coupons))
	if len(coupons) == int(limit) {
		// If we got a full page, there might be more
		total = int64(offset) + int64(limit) + 1
	}

	couponResponses := make([]CouponResponse, len(coupons))
	for i, coupon := range coupons {
		couponResponses[i] = *s.toCouponResponse(&coupon)
	}

	return &CouponListResponse{
		Coupons: couponResponses,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

// GetOrderCoupons retrieves all coupons used in an order
func (s *CouponService) GetOrderCoupons(ctx context.Context, orderID uuid.UUID) ([]CouponResponse, error) {
	couponOrders, err := s.couponOrderRepo.GetCouponOrdersByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order coupons: %w", err)
	}

	couponResponses := make([]CouponResponse, len(couponOrders))
	for i, co := range couponOrders {
		couponResponses[i] = CouponResponse{
			ID:         co.CouponID,
			CouponCode: co.CouponCode,
			IsValid:    co.IsValid,
			CreatedAt:  co.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:  co.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return couponResponses, nil
}

// RemoveCouponFromOrder removes a coupon from an order
func (s *CouponService) RemoveCouponFromOrder(ctx context.Context, orderID, couponID uuid.UUID) error {
	// Check if the coupon is used in the order
	isUsed, err := s.couponOrderRepo.CheckCouponUsageInOrder(ctx, orderID, couponID)
	if err != nil {
		return fmt.Errorf("failed to check coupon usage: %w", err)
	}

	if !isUsed {
		return fmt.Errorf("coupon is not used in this order")
	}

	// Remove all coupon orders for this order and coupon combination
	// Note: This is a simplified approach. In a real system, you might want to
	// remove only specific coupon order records
	err = s.couponOrderRepo.DeleteCouponOrdersByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to remove coupon from order: %w", err)
	}

	return nil
}

// Helper methods

func (s *CouponService) validateCouponCode(code string) error {
	code = strings.TrimSpace(code)
	if len(code) < 3 {
		return fmt.Errorf("coupon code must be at least 3 characters long")
	}
	if len(code) > 50 {
		return fmt.Errorf("coupon code must be at most 50 characters long")
	}
	// Add more validation rules as needed
	return nil
}

func (s *CouponService) toCouponResponse(coupon *sqlc.Coupon) *CouponResponse {
	return &CouponResponse{
		ID:         coupon.ID,
		CouponCode: coupon.CouponCode,
		IsValid:    coupon.IsValid,
		CreatedAt:  coupon.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  coupon.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}
