package http

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"coupon-service/internal/service"
)

// Router handles HTTP routing for the coupon service
type Router struct {
	engine        *gin.Engine
	couponHandler *CouponHandler
}

// NewRouter creates a new HTTP router
func NewRouter(couponService *service.CouponService) *Router {
	// Set Gin mode based on environment
	gin.SetMode(gin.ReleaseMode) // Change to gin.DebugMode for development

	engine := gin.New()
	couponHandler := NewCouponHandler(couponService)

	router := &Router{
		engine:        engine,
		couponHandler: couponHandler,
	}

	router.setupMiddleware()
	router.setupRoutes()

	return router
}

// GetEngine returns the Gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// setupMiddleware configures middleware for the router
func (r *Router) setupMiddleware() {
	// Recovery middleware
	r.engine.Use(gin.Recovery())

	// Logger middleware
	r.engine.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\"",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	// CORS middleware
	r.engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Configure appropriately for production
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Request timeout middleware
	r.engine.Use(func(c *gin.Context) {
		// Set a timeout for requests (optional)
		c.Next()
	})
}

// setupRoutes configures all routes for the application
func (r *Router) setupRoutes() {
	// Health check endpoint
	r.engine.GET("/health", r.couponHandler.HealthCheck)

	// Metrics endpoint (for monitoring)
	r.engine.GET("/metrics", r.couponHandler.Metrics)

	// API v1 routes
	v1 := r.engine.Group("/api/v1")
	{
		// Coupon routes
		coupons := v1.Group("/coupons")
		{
			// Basic CRUD operations
			coupons.POST("", r.couponHandler.CreateCoupon)       // Create coupon
			coupons.GET("", r.couponHandler.ListCoupons)         // List coupons with pagination and search
			coupons.GET("/:id", r.couponHandler.GetCoupon)       // Get coupon by ID
			coupons.PUT("/:id", r.couponHandler.UpdateCoupon)    // Update coupon
			coupons.DELETE("/:id", r.couponHandler.DeleteCoupon) // Delete coupon

			// Special operations
			coupons.GET("/code/:code", r.couponHandler.GetCouponByCode) // Get coupon by code
			coupons.POST("/use", r.couponHandler.UseCoupon)             // Use coupon in order
			coupons.POST("/validate", r.couponHandler.ValidateCoupon)   // Validate coupon for order
		}

		// Order-related coupon routes
		orders := v1.Group("/orders")
		{
			orders.GET("/:order_id/coupons", r.couponHandler.GetOrderCoupons)                     // Get all coupons for an order
			orders.DELETE("/:order_id/coupons/:coupon_id", r.couponHandler.RemoveCouponFromOrder) // Remove coupon from order
		}
	}

	// Add a catch-all route for undefined endpoints
	r.engine.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error":   "not_found",
			"message": "The requested endpoint was not found",
			"path":    c.Request.URL.Path,
		})
	})
}

// Additional middleware functions can be added here

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate a simple request ID (in production, use a proper UUID library)
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		c.Header("X-Request-ID", requestID)
		c.Set("RequestID", requestID)
		c.Next()
	}
}

// RateLimitMiddleware adds rate limiting (placeholder implementation)
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implement rate limiting logic here
		// This could use Redis or an in-memory store
		c.Next()
	}
}

// AuthMiddleware adds authentication (placeholder implementation)
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implement authentication logic here
		// This could validate JWT tokens, API keys, etc.
		c.Next()
	}
}
