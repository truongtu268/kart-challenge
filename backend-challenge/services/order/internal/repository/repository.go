package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	query "order-service/internal/repository/postgres/sqlc"
)

// OrderRepository handles order data operations
type OrderRepository struct {
	queries *query.Queries
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(queries *query.Queries) *OrderRepository {
	return &OrderRepository{
		queries: queries,
	}
}

// CreateOrder creates a new order
func (r *OrderRepository) CreateOrder(ctx context.Context, couponCode *string, totalAmount, discountAmount, finalAmount decimal.Decimal, status string) (*query.Order, error) {
	var couponCodeText pgtype.Text
	if couponCode != nil {
		couponCodeText = pgtype.Text{String: *couponCode, Valid: true}
	} else {
		couponCodeText = pgtype.Text{Valid: false}
	}
	
	order, err := r.queries.CreateOrder(ctx, query.CreateOrderParams{
		CouponCode:     couponCodeText,
		TotalAmount:    totalAmount,
		DiscountAmount: discountAmount,
		FinalAmount:    finalAmount,
		Status:         status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}
	return &order, nil
}

// GetOrder retrieves an order by ID
func (r *OrderRepository) GetOrder(ctx context.Context, id uuid.UUID) (*query.Order, error) {
	order, err := r.queries.GetOrder(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return &order, nil
}

// ListOrders retrieves orders with pagination
func (r *OrderRepository) ListOrders(ctx context.Context, limit, offset int32) ([]query.Order, error) {
	orders, err := r.queries.ListOrders(ctx, query.ListOrdersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	return orders, nil
}

// UpdateOrderStatus updates the status of an order
func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string) (*query.Order, error) {
	order, err := r.queries.UpdateOrderStatus(ctx, query.UpdateOrderStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}
	return &order, nil
}

// DeleteOrder deletes an order
func (r *OrderRepository) DeleteOrder(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteOrder(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}
	return nil
}

// OrderItemRepository handles order item data operations
type OrderItemRepository struct {
	queries *query.Queries
}

// NewOrderItemRepository creates a new order item repository
func NewOrderItemRepository(queries *query.Queries) *OrderItemRepository {
	return &OrderItemRepository{
		queries: queries,
	}
}

// CreateOrderItem creates a new order item
func (r *OrderItemRepository) CreateOrderItem(ctx context.Context, orderID uuid.UUID, productID string, quantity int32, unitPrice, totalPrice decimal.Decimal) (*query.OrderItem, error) {
	orderItem, err := r.queries.CreateOrderItem(ctx, query.CreateOrderItemParams{
		OrderID:    orderID,
		ProductID:  productID,
		Quantity:   quantity,
		UnitPrice:  unitPrice,
		TotalPrice: totalPrice,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create order item: %w", err)
	}
	return &orderItem, nil
}

// GetOrderItems retrieves all items for an order
func (r *OrderItemRepository) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]query.OrderItem, error) {
	orderItems, err := r.queries.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	return orderItems, nil
}

// GetOrderItem retrieves an order item by ID
func (r *OrderItemRepository) GetOrderItem(ctx context.Context, id uuid.UUID) (*query.OrderItem, error) {
	orderItem, err := r.queries.GetOrderItem(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order item not found")
		}
		return nil, fmt.Errorf("failed to get order item: %w", err)
	}
	return &orderItem, nil
}

// UpdateOrderItem updates an order item
func (r *OrderItemRepository) UpdateOrderItem(ctx context.Context, id uuid.UUID, quantity int32, unitPrice, totalPrice decimal.Decimal) (*query.OrderItem, error) {
	orderItem, err := r.queries.UpdateOrderItem(ctx, query.UpdateOrderItemParams{
		ID:         id,
		Quantity:   quantity,
		UnitPrice:  unitPrice,
		TotalPrice: totalPrice,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update order item: %w", err)
	}
	return &orderItem, nil
}

// DeleteOrderItem deletes an order item
func (r *OrderItemRepository) DeleteOrderItem(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteOrderItem(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete order item: %w", err)
	}
	return nil
}

// DeleteOrderItemsByOrderID deletes all order items for an order
func (r *OrderItemRepository) DeleteOrderItemsByOrderID(ctx context.Context, orderID uuid.UUID) error {
	err := r.queries.DeleteOrderItemsByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to delete order items: %w", err)
	}
	return nil
}