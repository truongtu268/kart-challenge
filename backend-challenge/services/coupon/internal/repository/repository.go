package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	query "coupon-service/internal/repository/postgres/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// CouponRepository provides database operations for coupons
type CouponRepository struct {
	db      query.DBTX
	queries *query.Queries
}

// NewCouponRepository creates a new coupon repository
func NewCouponRepository(db query.DBTX) *CouponRepository {
	return &CouponRepository{
		db:      db,
		queries: query.New(db),
	}
}

// CreateCoupon creates a new coupon
func (r *CouponRepository) CreateCoupon(ctx context.Context, code string, isValid bool) (*query.Coupon, error) {
	coupon, err := r.queries.CreateCoupon(ctx, query.CreateCouponParams{
		CouponCode: code,
		IsValid:    isValid,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create coupon: %w", err)
	}
	return &coupon, nil
}

// GetCoupon retrieves a coupon by ID
func (r *CouponRepository) GetCoupon(ctx context.Context, id uuid.UUID) (*query.Coupon, error) {
	coupon, err := r.queries.GetCoupon(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("coupon not found")
		}
		return nil, fmt.Errorf("failed to get coupon: %w", err)
	}
	return &coupon, nil
}

// GetCouponByCode retrieves a coupon by code
func (r *CouponRepository) GetCouponByCode(ctx context.Context, code string) (*query.Coupon, error) {
	coupon, err := r.queries.GetCouponByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("coupon not found")
		}
		return nil, fmt.Errorf("failed to get coupon by code: %w", err)
	}
	return &coupon, nil
}

// ListCoupons retrieves coupons with pagination
func (r *CouponRepository) ListCoupons(ctx context.Context, limit, offset int32) ([]query.Coupon, error) {
	coupons, err := r.queries.ListCoupons(ctx, query.ListCouponsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list coupons: %w", err)
	}
	return coupons, nil
}

// ListValidCoupons retrieves valid coupons with pagination
func (r *CouponRepository) ListValidCoupons(ctx context.Context, limit, offset int32) ([]query.Coupon, error) {
	coupons, err := r.queries.ListValidCoupons(ctx, query.ListValidCouponsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list valid coupons: %w", err)
	}
	return coupons, nil
}

// UpdateCoupon updates a coupon
func (r *CouponRepository) UpdateCoupon(ctx context.Context, id uuid.UUID, code string, isValid bool) (*query.Coupon, error) {
	coupon, err := r.queries.UpdateCoupon(ctx, query.UpdateCouponParams{
		ID:         id,
		CouponCode: code,
		IsValid:    isValid,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("coupon not found")
		}
		return nil, fmt.Errorf("failed to update coupon: %w", err)
	}
	return &coupon, nil
}

// UpdateCouponValidity updates only the validity of a coupon
func (r *CouponRepository) UpdateCouponValidity(ctx context.Context, id uuid.UUID, isValid bool) (*query.Coupon, error) {
	coupon, err := r.queries.UpdateCouponValidity(ctx, query.UpdateCouponValidityParams{
		ID:      id,
		IsValid: isValid,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("coupon not found")
		}
		return nil, fmt.Errorf("failed to update coupon validity: %w", err)
	}
	return &coupon, nil
}

// DeleteCoupon deletes a coupon
func (r *CouponRepository) DeleteCoupon(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteCoupon(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete coupon: %w", err)
	}
	return nil
}

// CountCoupons returns the total number of coupons
func (r *CouponRepository) CountCoupons(ctx context.Context) (int64, error) {
	count, err := r.queries.CountCoupons(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count coupons: %w", err)
	}
	return count, nil
}

// CountValidCoupons returns the number of valid coupons
func (r *CouponRepository) CountValidCoupons(ctx context.Context) (int64, error) {
	count, err := r.queries.CountValidCoupons(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count valid coupons: %w", err)
	}
	return count, nil
}

// SearchCouponsByCode searches coupons by code pattern
func (r *CouponRepository) SearchCouponsByCode(ctx context.Context, pattern string, limit, offset int32) ([]query.Coupon, error) {
	coupons, err := r.queries.SearchCouponsByCode(ctx, query.SearchCouponsByCodeParams{
		Column1: pgtype.Text{String: pattern, Valid: true},
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search coupons by code: %w", err)
	}
	return coupons, nil
}

// CouponOrderRepository provides database operations for coupon orders
type CouponOrderRepository struct {
	db      query.DBTX
	queries *query.Queries
}

// NewCouponOrderRepository creates a new coupon order repository
func NewCouponOrderRepository(db query.DBTX) *CouponOrderRepository {
	return &CouponOrderRepository{
		db:      db,
		queries: query.New(db),
	}
}

// CreateCouponOrder creates a new coupon order
func (r *CouponOrderRepository) CreateCouponOrder(ctx context.Context, orderID, couponID uuid.UUID) (*query.CouponOrder, error) {
	couponOrder, err := r.queries.CreateCouponOrder(ctx, query.CreateCouponOrderParams{
		OrderID:  orderID,
		CouponID: couponID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create coupon order: %w", err)
	}
	return &couponOrder, nil
}

// GetCouponOrder retrieves a coupon order by ID
func (r *CouponOrderRepository) GetCouponOrder(ctx context.Context, id uuid.UUID) (*query.CouponOrder, error) {
	couponOrder, err := r.queries.GetCouponOrder(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("coupon order not found")
		}
		return nil, fmt.Errorf("failed to get coupon order: %w", err)
	}
	return &couponOrder, nil
}

// GetCouponOrdersByOrderID retrieves coupon orders by order ID
func (r *CouponOrderRepository) GetCouponOrdersByOrderID(ctx context.Context, orderID uuid.UUID) ([]query.GetCouponOrderByOrderIDRow, error) {
	couponOrders, err := r.queries.GetCouponOrderByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupon orders by order ID: %w", err)
	}
	return couponOrders, nil
}

// GetCouponOrdersByCouponID retrieves coupon orders by coupon ID
func (r *CouponOrderRepository) GetCouponOrdersByCouponID(ctx context.Context, couponID uuid.UUID) ([]query.CouponOrder, error) {
	couponOrders, err := r.queries.GetCouponOrderByCouponID(ctx, couponID)
	if err != nil {
		return nil, fmt.Errorf("failed to get coupon orders by coupon ID: %w", err)
	}
	return couponOrders, nil
}

// CheckCouponUsageInOrder checks if a coupon is used in an order
func (r *CouponOrderRepository) CheckCouponUsageInOrder(ctx context.Context, orderID, couponID uuid.UUID) (bool, error) {
	result, err := r.queries.CheckCouponUsageInOrder(ctx, query.CheckCouponUsageInOrderParams{
		OrderID:  orderID,
		CouponID: couponID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check coupon usage in order: %w", err)
	}
	return result, nil
}

// DeleteCouponOrder deletes a coupon order
func (r *CouponOrderRepository) DeleteCouponOrder(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteCouponOrder(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete coupon order: %w", err)
	}
	return nil
}

// DeleteCouponOrdersByOrderID deletes all coupon orders for an order
func (r *CouponOrderRepository) DeleteCouponOrdersByOrderID(ctx context.Context, orderID uuid.UUID) error {
	err := r.queries.DeleteCouponOrderByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to delete coupon orders by order ID: %w", err)
	}
	return nil
}

// CountCouponOrdersByOrderID counts coupon orders for an order
func (r *CouponOrderRepository) CountCouponOrdersByOrderID(ctx context.Context, orderID uuid.UUID) (int64, error) {
	count, err := r.queries.CountCouponOrdersByOrderID(ctx, orderID)
	if err != nil {
		return 0, fmt.Errorf("failed to count coupon orders by order ID: %w", err)
	}
	return count, nil
}
