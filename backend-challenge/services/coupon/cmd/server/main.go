package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/exaring/otelpgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	api "coupon-service/internal/handler/http"
	"coupon-service/internal/repository"
	"coupon-service/internal/service"
)

type Config struct {
	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
		SSLMode  string `mapstructure:"sslmode"`
	} `mapstructure:"database"`
	Server struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"server"`
	Log struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
	} `mapstructure:"log"`
}

func main() {
	// Load configuration
	config := loadConfig()

	// Initialize logger
	logger := initLogger(config.Log.Level, config.Log.Format)
	defer logger.Sync()

	logger.Info("Starting coupon service", zap.String("version", "1.0.0"))

	// Connect to database
	db, err := connectDB(config)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize repositories
	couponRepo := repository.NewCouponRepository(db)
	couponOrderRepo := repository.NewCouponOrderRepository(db)

	// Initialize services
	couponService := service.NewCouponService(couponRepo, couponOrderRepo)

	// Initialize HTTP router
	router := api.NewRouter(couponService)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Server.Port),
		Handler: router.GetEngine(),
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", zap.Int("port", config.Server.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func loadConfig() *Config {
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.name", "coupon_service")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")

	// Environment variables
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.name", "DB_NAME")
	viper.BindEnv("database.sslmode", "DB_SSLMODE")
	viper.BindEnv("server.host", "SERVER_HOST")
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("log.level", "LOG_LEVEL")
	viper.BindEnv("log.format", "LOG_FORMAT")

	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal("Failed to unmarshal config:", err)
	}

	return &config
}

func initLogger(level, format string) *zap.Logger {
	var config zap.Config
	if format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	switch level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err := config.Build()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	return logger
}

func connectDB(config *Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.SSLMode,
	)

	connPool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("Error while creating connection to the database!!")
	}

	if err := otelpgx.RecordStats(connPool); err != nil {
		return nil, fmt.Errorf("unable to record database stats: %w", err)
	}

	return connPool, nil
}

// setupRouter function is no longer needed as we use the http package router
