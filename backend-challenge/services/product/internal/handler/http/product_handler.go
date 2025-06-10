package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"product-service/internal/service"
)

// ProductHandler handles HTTP requests for product operations
type ProductHandler struct {
	productService *service.ProductService
}

// NewProductHandler creates a new product handler
func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
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

// ListProducts handles GET /product
func (h *ProductHandler) ListProducts(c *gin.Context) {
	products, err := h.productService.ListProducts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve products",
		})
		return
	}

	c.JSON(http.StatusOK, products)
}

// GetProduct handles GET /product/{productId}
func (h *ProductHandler) GetProduct(c *gin.Context) {
	productID := c.Param("productId")
	if productID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Product ID is required",
		})
		return
	}

	product, err := h.productService.GetProduct(c.Request.Context(), productID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "product not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, ErrorResponse{
			Error:   "get_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, product)
}
