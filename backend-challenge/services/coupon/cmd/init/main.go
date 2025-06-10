package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"coupon-service/internal/repository"
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
	Log struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
	} `mapstructure:"log"`
}

var couponFilePaths = []string{
	"/Users/macbookpro2017/Documents/source/test/kart-challenge/backend-challenge/services/coupon/cmd/init/data/couponbase1.gz",
	"/Users/macbookpro2017/Documents/source/test/kart-challenge/backend-challenge/services/coupon/cmd/init/data/couponbase2.gz",
	"/Users/macbookpro2017/Documents/source/test/kart-challenge/backend-challenge/services/coupon/cmd/init/data/couponbase3.gz",
}

func main() {
	// Load configuration
	config := loadConfig()

	// Initialize logger
	logger := initLogger(config.Log.Level, config.Log.Format)
	defer logger.Sync()

	logger.Info("Starting coupon data initialization")

	// Connect to database
	db, err := connectDB(config)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize repository
	couponRepo := repository.NewCouponRepository(db)

	// Process coupon files using temporary tables
	err = processeCouponFiles(logger, db, couponRepo)
	if err != nil {
		logger.Fatal("Failed to process coupon files", zap.Error(err))
	}

	logger.Info("Coupon data initialization completed successfully")
}

func loadConfig() *Config {
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.name", "coupon_service")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")

	// Environment variables
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.name", "DB_NAME")
	viper.BindEnv("database.sslmode", "DB_SSLMODE")
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
		return nil, fmt.Errorf("error while creating connection to the database: %w", err)
	}

	return connPool, nil
}

func processeCouponFiles(logger *zap.Logger, db *pgxpool.Pool, couponRepo *repository.CouponRepository) error {
	ctx := context.Background()
	tempTableNames := make([]string, 0, len(couponFilePaths))

	// Process each file and create temporary tables
	for i, filePath := range couponFilePaths {
		// Create temporary table name based on file
		tempTableName := fmt.Sprintf("temp_coupons_%d", i+1)
		tempTableNames = append(tempTableNames, tempTableName)

		// Create temporary table
		err := createTempTable(ctx, db, tempTableName, logger)
		if err != nil {
			return fmt.Errorf("failed to create temp table %s: %w", tempTableName, err)
		}

		logger.Info("Importing coupons from .gz file", zap.String("path", filePath))

		// Read and process .gz file programmatically
		count, err := importGzFileToTable(ctx, db, filePath, tempTableName, logger)
		if err != nil {
			return fmt.Errorf("failed to import .gz file %s to temp table %s: %w", filePath, tempTableName, err)
		}

		logger.Info("Completed importing file",
			zap.String("path", filePath),
			zap.Int("total_imported", count))
	}

	// Find and insert valid coupons (appearing in at least 2 temp tables)
	err := insertValidCouponsFromTempTables(ctx, db, tempTableNames, couponRepo, logger)
	if err != nil {
		return fmt.Errorf("failed to insert valid coupons: %w", err)
	}

	// Clean up temporary tables
	for _, tempTableName := range tempTableNames {
		err = dropTempTable(ctx, db, tempTableName, logger)
		if err != nil {
			logger.Warn("Failed to drop temp table", zap.String("table", tempTableName), zap.Error(err))
		}
	}

	return nil
}

// createTempTable creates a temporary table for storing coupons from a file
func createTempTable(ctx context.Context, db *pgxpool.Pool, tableName string, logger *zap.Logger) error {
	// Drop table if it exists first
	dropQuery := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	_, err := db.Exec(ctx, dropQuery)
	if err != nil {
		return fmt.Errorf("failed to drop existing temp table: %w", err)
	}

	// Create the table without PRIMARY KEY to allow duplicates
	query := fmt.Sprintf(`
		CREATE TABLE %s (
			coupon_code VARCHAR(10)
		)
	`, tableName)

	_, err = db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create temp table: %w", err)
	}

	logger.Info("Created temporary table", zap.String("table", tableName))
	return nil
}

// insertValidCouponsFromTempTables finds coupons that appear in at least 2 temp tables and inserts them into the main coupons table
func insertValidCouponsFromTempTables(ctx context.Context, db *pgxpool.Pool, tempTableNames []string, couponRepo *repository.CouponRepository, logger *zap.Logger) error {
	if len(tempTableNames) < 2 {
		return fmt.Errorf("need at least 2 temp tables to find valid coupons")
	}

	// Build query to find coupons that appear in at least 2 tables
	unionQueries := make([]string, len(tempTableNames))
	for i, tableName := range tempTableNames {
		unionQueries[i] = fmt.Sprintf("SELECT coupon_code FROM %s", tableName)
	}

	query := fmt.Sprintf(`
		SELECT coupon_code, COUNT(*) as file_count
		FROM (
			%s
		) AS all_coupons
		GROUP BY coupon_code
		HAVING COUNT(*) >= 2
	`, strings.Join(unionQueries, " UNION ALL "))

	rows, err := db.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query valid coupons: %w", err)
	}
	defer rows.Close()

	validCount := 0
	for rows.Next() {
		var couponCode string
		var fileCount int

		err := rows.Scan(&couponCode, &fileCount)
		if err != nil {
			return fmt.Errorf("failed to scan coupon row: %w", err)
		}

		// Insert valid coupon into main table
		_, err = couponRepo.CreateCoupon(ctx, couponCode, true)
		if err != nil {
			logger.Warn("Failed to create coupon", zap.String("code", couponCode), zap.Error(err))
			continue
		}

		validCount++
		logger.Debug("Created valid coupon", zap.String("code", couponCode), zap.Int("file_count", fileCount))
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating coupon rows: %w", err)
	}

	logger.Info("Inserted valid coupons", zap.Int("count", validCount))
	return nil
}

// dropTempTable drops a temporary table
func dropTempTable(ctx context.Context, db *pgxpool.Pool, tableName string, logger *zap.Logger) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)

	_, err := db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop temp table: %w", err)
	}

	logger.Info("Dropped temporary table", zap.String("table", tableName))
	return nil
}

// importGzFileToTable reads a .gz file and imports its contents to a temporary table
func importGzFileToTable(ctx context.Context, db *pgxpool.Pool, filePath, tableName string, logger *zap.Logger) (int, error) {
	// Open the .gz file
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return 0, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Read lines and insert in batches
	scanner := bufio.NewScanner(gzReader)
	batch := make([]string, 0, 1000) // Process in batches of 1000
	totalCount := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 {
			batch = append(batch, line)
		}

		// Insert batch when it reaches the limit
		if len(batch) >= 1000 {
			count, err := insertBatch(ctx, db, tableName, batch)
			if err != nil {
				return totalCount, fmt.Errorf("failed to insert batch: %w", err)
			}
			totalCount += count
			batch = batch[:0] // Reset batch
		}
	}

	// Insert remaining items in batch
	if len(batch) > 0 {
		count, err := insertBatch(ctx, db, tableName, batch)
		if err != nil {
			return totalCount, fmt.Errorf("failed to insert final batch: %w", err)
		}
		totalCount += count
	}

	if err := scanner.Err(); err != nil {
		return totalCount, fmt.Errorf("error reading file: %w", err)
	}

	return totalCount, nil
}

// insertBatch inserts a batch of coupon codes into the temporary table
func insertBatch(ctx context.Context, db *pgxpool.Pool, tableName string, codes []string) (int, error) {
	if len(codes) == 0 {
		return 0, nil
	}

	// Build the VALUES clause
	valuesClause := make([]string, len(codes))
	args := make([]interface{}, len(codes))
	for i, code := range codes {
		valuesClause[i] = fmt.Sprintf("($%d)", i+1)
		args[i] = code
	}

	query := fmt.Sprintf("INSERT INTO %s (coupon_code) VALUES %s", tableName, strings.Join(valuesClause, ", "))

	_, err := db.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	return len(codes), nil
}
