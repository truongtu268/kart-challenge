package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"coupon-service/internal/service"
)

// CouponHandler handles HTTP requests for coupon operations
type CouponHandler struct {
	couponService *service.CouponService
}

// NewCouponHandler creates a new coupon handler
func NewCouponHandler(couponService *service.CouponService) *CouponHandler {
	return &CouponHandler{
		couponService: couponService,
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

// CreateCoupon handles POST /api/v1/coupons
func (h *CouponHandler) CreateCoupon(c *gin.Context) {
	var req service.CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	coupon, err := h.couponService.CreateCoupon(c.Request.Context(), req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "coupon code already exists" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "create_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Message: "Coupon created successfully",
		Data:    coupon,
	})
}

// GetCoupon handles GET /api/v1/coupons/:id
func (h *CouponHandler) GetCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid coupon ID format",
		})
		return
	}

	coupon, err := h.couponService.GetCoupon(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "coupon not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "get_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Coupon retrieved successfully",
		Data:    coupon,
	})
}

// GetCouponByCode handles GET /api/v1/coupons/code/:code
func (h *CouponHandler) GetCouponByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_code",
			Message: "Coupon code is required",
		})
		return
	}

	coupon, err := h.couponService.GetCouponByCode(c.Request.Context(), code)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "coupon not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "get_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Coupon retrieved successfully",
		Data:    coupon,
	})
}

// ListCoupons handles GET /api/v1/coupons
func (h *CouponHandler) ListCoupons(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	validOnlyStr := c.DefaultQuery("valid_only", "false")
	query := c.Query("q")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil || offset < 0 {
		offset = 0
	}

	validOnly := validOnlyStr == "true"

	var result *service.CouponListResponse

	// If query is provided, search coupons
	if query != "" {
		result, err = h.couponService.SearchCoupons(c.Request.Context(), query, int32(limit), int32(offset))
	} else {
		result, err = h.couponService.ListCoupons(c.Request.Context(), int32(limit), int32(offset), validOnly)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "list_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Coupons retrieved successfully",
		Data:    result,
	})
}

// UpdateCoupon handles PUT /api/v1/coupons/:id
func (h *CouponHandler) UpdateCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid coupon ID format",
		})
		return
	}

	var req service.UpdateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	coupon, err := h.couponService.UpdateCoupon(c.Request.Context(), id, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "coupon not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "coupon code already exists" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Coupon updated successfully",
		Data:    coupon,
	})
}

// DeleteCoupon handles DELETE /api/v1/coupons/:id
func (h *CouponHandler) DeleteCoupon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid coupon ID format",
		})
		return
	}

	err = h.couponService.DeleteCoupon(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "coupon not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "cannot delete coupon: it is used in order(s)" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "delete_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Coupon deleted successfully",
	})
}

// UseCoupon handles POST /api/v1/coupons/use
func (h *CouponHandler) UseCoupon(c *gin.Context) {
	var req service.UseCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	err := h.couponService.UseCoupon(c.Request.Context(), req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "coupon not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "coupon is not valid" || err.Error() == "coupon is already used in this order" {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "use_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Coupon used successfully",
	})
}

// ValidateCoupon handles POST /api/v1/coupons/validate
func (h *CouponHandler) ValidateCoupon(c *gin.Context) {
	var req service.ValidateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	result, err := h.couponService.ValidateCoupon(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "validation_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Coupon validation completed",
		Data:    result,
	})
}

// GetOrderCoupons handles GET /api/v1/orders/:order_id/coupons
func (h *CouponHandler) GetOrderCoupons(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_order_id",
			Message: "Invalid order ID format",
		})
		return
	}

	coupons, err := h.couponService.GetOrderCoupons(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "get_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Order coupons retrieved successfully",
		Data:    coupons,
	})
}

// RemoveCouponFromOrder handles DELETE /api/v1/orders/:order_id/coupons/:coupon_id
func (h *CouponHandler) RemoveCouponFromOrder(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_order_id",
			Message: "Invalid order ID format",
		})
		return
	}

	couponIDStr := c.Param("coupon_id")
	couponID, err := uuid.Parse(couponIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_coupon_id",
			Message: "Invalid coupon ID format",
		})
		return
	}

	err = h.couponService.RemoveCouponFromOrder(c.Request.Context(), orderID, couponID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "coupon is not used in this order" {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "remove_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Coupon removed from order successfully",
	})
}

// HealthCheck handles GET /health
func (h *CouponHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "coupon-service",
		"version": "1.0.0",
	})
}

// Metrics handles GET /metrics (placeholder for Prometheus metrics)
func (h *CouponHandler) Metrics(c *gin.Context) {
	// This would typically be handled by a Prometheus metrics handler
	// For now, we'll return a simple response
	c.String(http.StatusOK, "# Metrics endpoint - integrate with Prometheus\n")
}