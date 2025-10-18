package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jphacks/os_2522/backend/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds database configuration
type Config struct {
	Driver   string
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDB creates a new database connection
func NewDB(config *Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch config.Driver {
	case "postgres":
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode,
		)
		dialector = postgres.Open(dsn)
	case "sqlite":
		// For development/testing
		dialector = sqlite.Open(config.DBName)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	// Configure logger
	logLevel := logger.Silent
	if os.Getenv("DB_LOG") == "true" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		&repository.PersonEntity{},
		&repository.FaceEntity{},
		&repository.EncounterEntity{},
		&repository.JobEntity{},
	)

	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// GetConfigFromEnv reads database configuration from environment variables
func GetConfigFromEnv() *Config {
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "sqlite" // Default to SQLite for development
	}

	config := &Config{
		Driver: driver,
	}

	if driver == "postgres" {
		config.Host = getEnv("DB_HOST", "localhost")
		config.Port = getEnv("DB_PORT", "5432")
		config.User = getEnv("DB_USER", "postgres")
		config.Password = getEnv("DB_PASSWORD", "")
		config.DBName = getEnv("DB_NAME", "arsome_api")
		config.SSLMode = getEnv("DB_SSLMODE", "disable")
	} else {
		config.DBName = getEnv("DB_NAME", "arsome_api.db")
	}

	return config
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
