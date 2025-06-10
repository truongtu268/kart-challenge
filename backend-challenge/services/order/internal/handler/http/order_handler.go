package http

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"order-service/internal/service"
)

// OrderHandler handles HTTP requests for order operations
type OrderHandler struct {
	orderService *service.OrderService
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PlaceOrder handles POST /order
func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	var req service.OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Validate that items are provided
	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "At least one item is required",
		})
		return
	}

	// Validate each item
	for i, item := range req.Items {
		if item.ProductID == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_request",
				Message: fmt.Sprintf("Product ID is required for item %d", i+1),
			})
			return
		}
		if item.Quantity <= 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_request",
				Message: fmt.Sprintf("Quantity must be greater than 0 for item %d", i+1),
			})
			return
		}
	}

	order, err := h.orderService.PlaceOrder(c.Request.Context(), req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "place_order_failed"

		// Handle specific error cases
		if contains(err.Error(), "product") && contains(err.Error(), "not found") {
			statusCode = http.StatusBadRequest
			errorCode = "invalid_product"
		} else if contains(err.Error(), "invalid coupon") {
			statusCode = http.StatusBadRequest
			errorCode = "invalid_coupon"
		} else if contains(err.Error(), "validation failed") {
			statusCode = http.StatusBadRequest
			errorCode = "validation_failed"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   errorCode,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, order)
}

// GetOrder handles GET /order/:id (for internal use)
func (h *OrderHandler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid order ID format",
		})
		return
	}

	order, err := h.orderService.GetOrder(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "get_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Order retrieved successfully",
		Data:    order,
	})
}

// ListOrders handles GET /orders (for internal use)
func (h *OrderHandler) ListOrders(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil || offset < 0 {
		offset = 0
	}

	orders, err := h.orderService.ListOrders(c.Request.Context(), int32(limit), int32(offset))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "list_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Orders retrieved successfully",
		Data:    orders,
	})
}

// HealthCheck handles GET /health
func (h *OrderHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "order-service",
		"version": "1.0.0",
	})
}

// Metrics handles GET /metrics
func (h *OrderHandler) Metrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "order-service",
		"metrics": gin.H{
			"uptime": "running",
		},
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
			containsInMiddle(s, substr))))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}