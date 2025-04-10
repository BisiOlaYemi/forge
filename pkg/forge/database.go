package forge

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database represents the database connection
type Database struct {
	DB *gorm.DB
}

// NewDatabase creates a new database connection
func NewDatabase(config *DatabaseConfig) (*Database, error) {
	if config == nil {
		config = &DatabaseConfig{
			Driver:        "sqlite",
			Name:          "forge.db",
			SlowThreshold: 200 * time.Millisecond,
		}
	}

	// Create database directory if it doesn't exist
	dbDir := filepath.Dir(config.Name)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Configure GORM logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             config.SlowThreshold,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Create database connection
	var dialector gorm.Dialector
	switch config.Driver {
	case "sqlite":
		dialector = sqlite.Open(config.Name)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	// Connect to database
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Database{DB: db}, nil
}

// AutoMigrate automatically migrates the database schema
func (d *Database) AutoMigrate(models ...interface{}) error {
	return d.DB.AutoMigrate(models...)
}

// Create creates a new record
func (d *Database) Create(value interface{}) error {
	return d.DB.Create(value).Error
}

// First finds the first record matching the condition
func (d *Database) First(dest interface{}, cond ...interface{}) error {
	return d.DB.First(dest, cond...).Error
}

// Find finds all records matching the condition
func (d *Database) Find(dest interface{}, cond ...interface{}) error {
	return d.DB.Find(dest, cond...).Error
}

// Update updates a record
func (d *Database) Update(value interface{}) error {
	return d.DB.Save(value).Error
}

// Delete deletes a record
func (d *Database) Delete(value interface{}) error {
	return d.DB.Delete(value).Error
}

// Where creates a query with the given condition
func (d *Database) Where(query interface{}, args ...interface{}) *gorm.DB {
	return d.DB.Where(query, args...)
}

// Transaction executes a function within a transaction
func (d *Database) Transaction(fc func(tx *gorm.DB) error) error {
	return d.DB.Transaction(fc)
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
} 