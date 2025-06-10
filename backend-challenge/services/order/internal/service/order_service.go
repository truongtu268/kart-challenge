package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"order-service/internal/repository"
	query "order-service/internal/repository/postgres/sqlc"
)

// OrderService provides business logic for order operations
type OrderService struct {
	orderRepo     *repository.OrderRepository
	orderItemRepo *repository.OrderItemRepository
	productAPIURL string
	couponAPIURL  string
}

// NewOrderService creates a new order service
func NewOrderService(orderRepo *repository.OrderRepository, orderItemRepo *repository.OrderItemRepository, productAPIURL, couponAPIURL string) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		productAPIURL: productAPIURL,
		couponAPIURL:  couponAPIURL,
	}
}

// OrderRequest represents the request to create an order
type OrderRequest struct {
	CouponCode *string     `json:"couponCode,omitempty"`
	Items      []OrderItem `json:"items" binding:"required,min=1"`
}

// OrderItem represents an item in an order request
type OrderItem struct {
	ProductID string `json:"productId" binding:"required"`
	Quantity  int32  `json:"quantity" binding:"required,min=1"`
}

// OrderResponse represents an order response
type OrderResponse struct {
	ID       string                `json:"id"`
	Items    []OrderItemResponse   `json:"items"`
	Products []ProductResponse     `json:"products"`
}

// OrderItemResponse represents an order item in the response
type OrderItemResponse struct {
	ProductID string `json:"productId"`
	Quantity  int32  `json:"quantity"`
}

// ProductResponse represents a product from the product service
type ProductResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

// CouponValidationRequest represents a coupon validation request
type CouponValidationRequest struct {
	CouponCode string    `json:"coupon_code"`
	OrderID    uuid.UUID `json:"order_id"`
}

// CouponValidationResponse represents a coupon validation response
type CouponValidationResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

// CouponUsageRequest represents a coupon usage request
type CouponUsageRequest struct {
	OrderID  uuid.UUID `json:"order_id"`
	CouponID uuid.UUID `json:"coupon_id"`
}

// PlaceOrder creates a new order with validation
func (s *OrderService) PlaceOrder(ctx context.Context, req OrderRequest) (*OrderResponse, error) {
	// Validate products and get product details
	products, err := s.validateAndGetProducts(ctx, req.Items)
	if err != nil {
		return nil, fmt.Errorf("product validation failed: %w", err)
	}

	// Calculate total amount
	totalAmount := decimal.Zero
	for i, item := range req.Items {
		product := products[i]
		itemTotal := decimal.NewFromFloat(product.Price).Mul(decimal.NewFromInt32(item.Quantity))
		totalAmount = totalAmount.Add(itemTotal)
	}

	discountAmount := decimal.Zero
	finalAmount := totalAmount

	// Validate coupon if provided
	if req.CouponCode != nil && *req.CouponCode != "" {
		// Create a temporary order ID for coupon validation
		tempOrderID := uuid.New()
		isValid, err := s.validateCoupon(ctx, *req.CouponCode, tempOrderID)
		if err != nil {
			return nil, fmt.Errorf("coupon validation failed: %w", err)
		}
		if !isValid {
			return nil, fmt.Errorf("invalid coupon code")
		}

		// Apply discount (example: 10% discount)
		discountAmount = totalAmount.Mul(decimal.NewFromFloat(0.1))
		finalAmount = totalAmount.Sub(discountAmount)
	}

	// Create order
	order, err := s.orderRepo.CreateOrder(ctx, req.CouponCode, totalAmount, discountAmount, finalAmount, "confirmed")
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Create order items
	orderItems := make([]OrderItemResponse, len(req.Items))
	for i, item := range req.Items {
		product := products[i]
		unitPrice := decimal.NewFromFloat(product.Price)
		totalPrice := unitPrice.Mul(decimal.NewFromInt32(item.Quantity))

		_, err := s.orderItemRepo.CreateOrderItem(ctx, order.ID, item.ProductID, item.Quantity, unitPrice, totalPrice)
		if err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}

		orderItems[i] = OrderItemResponse{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// Use coupon if provided and valid
	if req.CouponCode != nil && *req.CouponCode != "" {
		err = s.useCoupon(ctx, order.ID, *req.CouponCode)
		if err != nil {
			// Log error but don't fail the order
			fmt.Printf("Warning: failed to use coupon: %v\n", err)
		}
	}

	return &OrderResponse{
		ID:       order.ID.String(),
		Items:    orderItems,
		Products: products,
	}, nil
}

// validateAndGetProducts validates that all products exist and returns their details
func (s *OrderService) validateAndGetProducts(ctx context.Context, items []OrderItem) ([]ProductResponse, error) {
	products := make([]ProductResponse, len(items))

	for i, item := range items {
		product, err := s.getProduct(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product %s not found: %w", item.ProductID, err)
		}
		products[i] = *product
	}

	return products, nil
}

// getProduct retrieves product details from the product service
func (s *OrderService) getProduct(ctx context.Context, productID string) (*ProductResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/product/%s", s.productAPIURL, productID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("product not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product service returned status %d", resp.StatusCode)
	}

	var product ProductResponse
	err = json.NewDecoder(resp.Body).Decode(&product)
	if err != nil {
		return nil, fmt.Errorf("failed to decode product response: %w", err)
	}

	return &product, nil
}

// validateCoupon validates a coupon code with the coupon service
func (s *OrderService) validateCoupon(ctx context.Context, couponCode string, orderID uuid.UUID) (bool, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/api/v1/coupons/validate", s.couponAPIURL)

	reqBody := CouponValidationRequest{
		CouponCode: couponCode,
		OrderID:    orderID,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to validate coupon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, nil // Invalid coupon
	}

	var validationResp CouponValidationResponse
	err = json.NewDecoder(resp.Body).Decode(&validationResp)
	if err != nil {
		return false, fmt.Errorf("failed to decode validation response: %w", err)
	}

	return validationResp.Valid, nil
}

// useCoupon uses a coupon with the coupon service
func (s *OrderService) useCoupon(ctx context.Context, orderID uuid.UUID, couponCode string) error {
	// First get the coupon by code to get its ID
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/api/v1/coupons/code/%s", s.couponAPIURL, couponCode)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get coupon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get coupon by code")
	}

	// Parse the response to get coupon ID
	var couponResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&couponResp)
	if err != nil {
		return fmt.Errorf("failed to decode coupon response: %w", err)
	}

	couponID, err := uuid.Parse(couponResp.Data.ID)
	if err != nil {
		return fmt.Errorf("invalid coupon ID: %w", err)
	}

	// Use the coupon
	useURL := fmt.Sprintf("%s/api/v1/coupons/use", s.couponAPIURL)
	useReqBody := CouponUsageRequest{
		OrderID:  orderID,
		CouponID: couponID,
	}

	jsonBody, err := json.Marshal(useReqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal use request: %w", err)
	}

	useReq, err := http.NewRequestWithContext(ctx, "POST", useURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create use request: %w", err)
	}
	useReq.Header.Set("Content-Type", "application/json")

	useResp, err := client.Do(useReq)
	if err != nil {
		return fmt.Errorf("failed to use coupon: %w", err)
	}
	defer useResp.Body.Close()

	if useResp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to use coupon, status: %d", useResp.StatusCode)
	}

	return nil
}

// GetOrder retrieves an order by ID
func (s *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*query.Order, error) {
	return s.orderRepo.GetOrder(ctx, id)
}

// ListOrders retrieves orders with pagination
func (s *OrderService) ListOrders(ctx context.Context, limit, offset int32) ([]query.Order, error) {
	return s.orderRepo.ListOrders(ctx, limit, offset)
}