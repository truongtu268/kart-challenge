package http

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"product-service/internal/service"
)

// Router handles HTTP routing for the product service
type Router struct {
	engine         *gin.Engine
	productHandler *ProductHandler
}

// NewRouter creates a new HTTP router
func NewRouter(productService *service.ProductService) *Router {
	// Set Gin mode based on environment
	gin.SetMode(gin.ReleaseMode) // Change to gin.DebugMode for development

	engine := gin.New()
	productHandler := NewProductHandler(productService)

	router := &Router{
		engine:         engine,
		productHandler: productHandler,
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
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.engine.Use(cors.New(config))
}

// setupRoutes configures routes for the router
func (r *Router) setupRoutes() {
	// Health check endpoint
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "product-service",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	// OpenAPI specification routes
	r.engine.GET("/product", r.productHandler.ListProducts)          // GET /product
	r.engine.GET("/product/:productId", r.productHandler.GetProduct) // GET /product/{productId}
}
