package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ProjectTemplate struct {
	Name        string
	Description string
	Files       map[string]string
}

type ControllerTemplate struct {
	Name        string
	Description string
	Methods     []string
}

type ModelTemplate struct {
	Name        string
	Description string
	Fields      []string
}

func createNewProject(name string) error {

	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	dirs := []string{
		filepath.Join(name, "app", "controllers"),
		filepath.Join(name, "app", "models"),
		filepath.Join(name, "app", "services"),
		filepath.Join(name, "config"),
		filepath.Join(name, "database", "migrations"),
		filepath.Join(name, "database", "seeders"),
		filepath.Join(name, "routes"),
		filepath.Join(name, "templates"),
		filepath.Join(name, "storage", "logs"),
		filepath.Join(name, "storage", "uploads"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create main.go
	mainContent := `package main

import (
	"fmt"
	"log"

	"github.com/BisiOlaYemi/forge/pkg/forge"
)

func main() {
	// Create a new Forge application
	app, err := forge.New(&forge.Config{
		Name:        "` + name + `",
		Version:     "1.0.0",
		Description: "A Forge application",
		Server: forge.ServerConfig{
			Host:     "localhost",
			Port:     3000,
			BasePath: "/",
		},
		Database: forge.DatabaseConfig{
			Driver: "sqlite",  // Choose from: sqlite, mysql, postgres, sqlserver
			Name:   "forge.db",
			// Uncomment these for other database types
			// Host:     "localhost",
			// Port:     3306,  // MySQL: 3306, PostgreSQL: 5432, SQL Server: 1433
			// Username: "forge_user",
			// Password: "forge_password",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Register controllers
	// app.RegisterController(&UserController{})

	// Start the server
	fmt.Printf("Server starting on http://localhost:3000\n")
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
`

	if err := os.WriteFile(filepath.Join(name, "main.go"), []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}

	// Create config/forge.yaml with comprehensive configuration
	configContent := `# Forge Framework Configuration

# Application Settings
app:
  name: "` + name + `"
  version: "1.0.0"
  description: "A powerful web application built with Forge Framework"
  environment: "development" 
  debug: true
  timezone: "UTC"
  secret_key: "change-this-to-your-secure-secret-key"
  log_level: "info" 

# Server Configuration
server:
  host: "localhost"
  port: 3000
  base_path: "/"
  read_timeout: 10s
  write_timeout: 10s
  idle_timeout: 120s

# Database Configuration
database:
  # Main database connection
  default:
    driver: "sqlite" 
    name: "forge.db"
    # Uncomment these for other database types
    # host: "localhost"
    # port: 3306  
    # username: "forge_user"
    # password: "forge_password"
    # ssl_mode: "disable" 
    # charset: "utf8mb4"
    # timezone: "Local"
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_life: 3600s 
    slow_threshold: 200ms
    log_level: "info" 
    debug: false

# Authentication Configuration
auth:
  jwt:
    secret_key: "change-this-to-your-personal-jwt-secret-key"
    expiration: 86400 
    refresh_expiration: 604800 
    signing_method: "HS256" 

# View Configuration
view:
  engine: "go-template" # go-template, jet
  directory: "templates"
  extension: ".gohtml"
  cache: true
`

	if err := os.WriteFile(filepath.Join(name, "config", "forge.yaml"), []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create forge.yaml: %w", err)
	}

	// Create go.mod
	modContent := `module ` + name + `

go 1.21

require (
	github.com/BisiOlaYemi/forge v0.0.0-20250410105738-69dbba69f7f0
)
`

	if err := os.WriteFile(filepath.Join(name, "go.mod"), []byte(modContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}

	// Create a basic README.md
	readmeContent := `# ` + name + `

A web application built with Forge Framework.

## Getting Started

1. Run the development server:
   
   ` + "```" + `bash
   forge serve
   ` + "```" + `

2. Open [http://localhost:3000](http://localhost:3000) in your browser.

## Database Configuration

This project uses SQLite by default, which requires no additional setup. To use other databases:

1. Edit the database configuration in ` + "`config/forge.yaml`" + `
2. Choose from: sqlite, mysql, postgres, sqlserver
3. Provide connection details as required

## Creating Controllers and Models

Generate new controllers:

` + "```" + `bash
forge generate controller User
` + "```" + `

Generate new models:

` + "```" + `bash
forge generate model User
` + "```" + `

## Learn More

To learn more about Forge Framework, check out the documentation at [Forge Framework Documentation](https://github.com/BisiOlaYemi/forge).
`

	if err := os.WriteFile(filepath.Join(name, "README.md"), []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	fmt.Printf("Created new Forge project: %s\n", name)
	return nil
}

// generateController generates a new controller
func generateController(name string) error {
	// Convert name to proper case
	name = strings.ToUpper(name[:1]) + name[1:]
	if !strings.HasSuffix(name, "Controller") {
		name += "Controller"
	}

	// Create controller file
	controllerContent := `package controllers

import (
	"github.com/BisiOlaYemi/forge/pkg/forge"
)

// ` + name + ` handles requests related to ` + strings.TrimSuffix(name, "Controller") + `
type ` + name + ` struct {
	forge.Controller
}

// HandleGet` + strings.TrimSuffix(name, "Controller") + ` handles getting a ` + strings.TrimSuffix(name, "Controller") + `
// @route GET /` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `s
// @desc Get all ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `s
// @response 200 []` + strings.TrimSuffix(name, "Controller") + `
func (c *` + name + `) HandleGet` + strings.TrimSuffix(name, "Controller") + `s(ctx *forge.Context) error {
	var items []interface{}
	if err := c.App.DB().Find(&items).Error; err != nil {
		return ctx.Status(500).JSON(map[string]string{"error": err.Error()})
	}
	return ctx.JSON(items)
}

// HandleGet` + strings.TrimSuffix(name, "Controller") + `ByID handles getting a ` + strings.TrimSuffix(name, "Controller") + ` by ID
// @route GET /` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `s/:id
// @desc Get a ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + ` by ID
// @param id path int true "` + strings.TrimSuffix(name, "Controller") + ` ID"
// @response 200 ` + strings.TrimSuffix(name, "Controller") + `
func (c *` + name + `) HandleGet` + strings.TrimSuffix(name, "Controller") + `ByID(ctx *forge.Context) error {
	id := ctx.Param("id")
	var item interface{}
	if err := c.App.DB().First(&item, id).Error; err != nil {
		return ctx.Status(404).JSON(map[string]string{"error": "Not found"})
	}
	return ctx.JSON(item)
}

// HandleCreate` + strings.TrimSuffix(name, "Controller") + ` handles creating a new ` + strings.TrimSuffix(name, "Controller") + `
// @route POST /` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `s
// @desc Create a new ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `
// @body Create` + strings.TrimSuffix(name, "Controller") + `Request
// @response 201 ` + strings.TrimSuffix(name, "Controller") + `
func (c *` + name + `) HandleCreate` + strings.TrimSuffix(name, "Controller") + `(ctx *forge.Context) error {
	var req Create` + strings.TrimSuffix(name, "Controller") + `Request
	
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(map[string]string{"error": err.Error()})
	}
	
	// Create record (replace with your model)
	item := map[string]interface{}{"name": req.Name}
	
	if err := c.App.DB().Create(&item).Error; err != nil {
		return ctx.Status(500).JSON(map[string]string{"error": err.Error()})
	}
	
	return ctx.Status(201).JSON(item)
}

// HandleUpdate` + strings.TrimSuffix(name, "Controller") + ` handles updating a ` + strings.TrimSuffix(name, "Controller") + `
// @route PUT /` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `s/:id
// @desc Update a ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `
// @param id path int true "` + strings.TrimSuffix(name, "Controller") + ` ID"
// @body Update` + strings.TrimSuffix(name, "Controller") + `Request
// @response 200 ` + strings.TrimSuffix(name, "Controller") + `
func (c *` + name + `) HandleUpdate` + strings.TrimSuffix(name, "Controller") + `(ctx *forge.Context) error {
	id := ctx.Param("id")
	
	var req Update` + strings.TrimSuffix(name, "Controller") + `Request
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(map[string]string{"error": err.Error()})
	}
	
	// Update record (replace with your model)
	var item interface{}
	if err := c.App.DB().First(&item, id).Error; err != nil {
		return ctx.Status(404).JSON(map[string]string{"error": "Not found"})
	}
	
	// Update fields based on request
	
	if err := c.App.DB().Save(&item).Error; err != nil {
		return ctx.Status(500).JSON(map[string]string{"error": err.Error()})
	}
	
	return ctx.JSON(item)
}

// HandleDelete` + strings.TrimSuffix(name, "Controller") + ` handles deleting a ` + strings.TrimSuffix(name, "Controller") + `
// @route DELETE /` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `s/:id
// @desc Delete a ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `
// @param id path int true "` + strings.TrimSuffix(name, "Controller") + ` ID"
// @response 204
func (c *` + name + `) HandleDelete` + strings.TrimSuffix(name, "Controller") + `(ctx *forge.Context) error {
	id := ctx.Param("id")
	
	// Delete record (replace with your model)
	var item interface{}
	if err := c.App.DB().First(&item, id).Error; err != nil {
		return ctx.Status(404).JSON(map[string]string{"error": "Not found"})
	}
	
	if err := c.App.DB().Delete(&item).Error; err != nil {
		return ctx.Status(500).JSON(map[string]string{"error": err.Error()})
	}
	
	return ctx.Status(204).Send([]byte{})
}
`

	modelContent := `package models

import (
	"time"
)

// ` + strings.TrimSuffix(name, "Controller") + ` represents a ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + ` entity
type ` + strings.TrimSuffix(name, "Controller") + ` struct {
	ID        uint      ` + "`json:\"id\" gorm:\"primaryKey\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\" gorm:\"autoCreateTime\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\" gorm:\"autoUpdateTime\"`" + `
	// Add custom fields here
}
`

	// Create request/response types
	typesContent := `package controllers

// Create` + strings.TrimSuffix(name, "Controller") + `Request represents the request body for creating a ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `
type Create` + strings.TrimSuffix(name, "Controller") + `Request struct {
	Name string ` + "`json:\"name\" validate:\"required\"`" + `
	// Add other fields here
}

// Update` + strings.TrimSuffix(name, "Controller") + `Request represents the request body for updating a ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `
type Update` + strings.TrimSuffix(name, "Controller") + `Request struct {
	Name string ` + "`json:\"name\" validate:\"required\"`" + `
	// Add other fields here
}
`

	// Write files
	if err := os.WriteFile(filepath.Join("app", "controllers", strings.ToLower(name)+".go"), []byte(controllerContent), 0644); err != nil {
		return fmt.Errorf("failed to create controller file: %w", err)
	}

	if err := os.WriteFile(filepath.Join("app", "models", strings.ToLower(strings.TrimSuffix(name, "Controller"))+".go"), []byte(modelContent), 0644); err != nil {
		return fmt.Errorf("failed to create model file: %w", err)
	}

	if err := os.WriteFile(filepath.Join("app", "controllers", strings.ToLower(name)+"_types.go"), []byte(typesContent), 0644); err != nil {
		return fmt.Errorf("failed to create types file: %w", err)
	}

	fmt.Printf("Generated controller: %s\n", name)
	return nil
}

// generateModel generates a new model
func generateModel(name string) error {
	name = strings.ToUpper(name[:1]) + name[1:]
	modelContent := `package models

import (
	"time"
)

// ` + name + ` represents a ` + strings.ToLower(name) + ` entity
type ` + name + ` struct {
	ID        uint      ` + "`json:\"id\" gorm:\"primaryKey\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\" gorm:\"autoCreateTime\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\" gorm:\"autoUpdateTime\"`" + `
	// Add your custom fields here
	
	// Example fields (uncomment and modify as needed):
	// Name        string    ` + "`json:\"name\" gorm:\"size:255;not null\"`" + `
	// Description string    ` + "`json:\"description\" gorm:\"type:text\"`" + `
	// Status      string    ` + "`json:\"status\" gorm:\"size:50;default:'active'\"`" + `
	// Amount      float64   ` + "`json:\"amount\" gorm:\"type:decimal(10,2);default:0\"`" + `
	// IsActive    bool      ` + "`json:\"is_active\" gorm:\"default:true\"`" + `
	// ExpiresAt   time.Time ` + "`json:\"expires_at\" gorm:\"index\"`" + `
}

// TableName overrides the table name
func (` + name + `) TableName() string {
	return "` + strings.ToLower(name) + `s"
}

// BeforeCreate hook called before record creation
func (m *` + name + `) BeforeCreate() error {
	// Add custom validation or data preparation logic here
	return nil
}
`

	// Create migration file
	migrationContent := `package migrations

import (
	"` + getCurrentModuleName() + `/app/models"
	"gorm.io/gorm"
)

// Create` + name + `Table creates the ` + strings.ToLower(name) + `s table
func Create` + name + `Table(db *gorm.DB) error {
	return db.AutoMigrate(&models.` + name + `{})
}

// Drop` + name + `Table drops the ` + strings.ToLower(name) + `s table
func Drop` + name + `Table(db *gorm.DB) error {
	return db.Migrator().DropTable(&models.` + name + `{})
}
`

	// Create a repository file for database operations
	repositoryContent := `package repositories

import (
	"` + getCurrentModuleName() + `/app/models"
	"gorm.io/gorm"
)

// ` + name + `Repository provides database operations for ` + name + ` model
type ` + name + `Repository struct {
	DB *gorm.DB
}

// New` + name + `Repository creates a new repository instance
func New` + name + `Repository(db *gorm.DB) *` + name + `Repository {
	return &` + name + `Repository{
		DB: db,
	}
}

// Create inserts a new ` + name + ` record
func (r *` + name + `Repository) Create(` + strings.ToLower(name) + ` *models.` + name + `) error {
	return r.DB.Create(` + strings.ToLower(name) + `).Error
}

// FindByID retrieves a ` + name + ` by ID
func (r *` + name + `Repository) FindByID(id uint) (*models.` + name + `, error) {
	var ` + strings.ToLower(name) + ` models.` + name + `
	err := r.DB.First(&` + strings.ToLower(name) + `, id).Error
	return &` + strings.ToLower(name) + `, err
}

// FindAll retrieves all ` + name + ` records
func (r *` + name + `Repository) FindAll() ([]models.` + name + `, error) {
	var ` + strings.ToLower(name) + `s []models.` + name + `
	err := r.DB.Find(&` + strings.ToLower(name) + `s).Error
	return ` + strings.ToLower(name) + `s, err
}

// Update updates a ` + name + ` record
func (r *` + name + `Repository) Update(` + strings.ToLower(name) + ` *models.` + name + `) error {
	return r.DB.Save(` + strings.ToLower(name) + `).Error
}

// Delete removes a ` + name + ` record
func (r *` + name + `Repository) Delete(id uint) error {
	return r.DB.Delete(&models.` + name + `{}, id).Error
}

// Count returns the total number of ` + name + ` records
func (r *` + name + `Repository) Count() (int64, error) {
	var count int64
	err := r.DB.Model(&models.` + name + `{}).Count(&count).Error
	return count, err
}

// Custom queries can be added below
`

	// Write files
	if err := os.WriteFile(filepath.Join("app", "models", strings.ToLower(name)+".go"), []byte(modelContent), 0644); err != nil {
		return fmt.Errorf("failed to create model file: %w", err)
	}

	if err := os.WriteFile(filepath.Join("database", "migrations", strings.ToLower(name)+"_migration.go"), []byte(migrationContent), 0644); err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}

	// Create repositories directory if it doesn't exist
	if err := os.MkdirAll(filepath.Join("app", "repositories"), 0755); err != nil {
		return fmt.Errorf("failed to create repositories directory: %w", err)
	}

	if err := os.WriteFile(filepath.Join("app", "repositories", strings.ToLower(name)+"_repository.go"), []byte(repositoryContent), 0644); err != nil {
		return fmt.Errorf("failed to create repository file: %w", err)
	}

	fmt.Printf("Generated model: %s\n", name)
	return nil
}

// Helper function to get the current Go module name
func getCurrentModuleName() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "app"
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) >= 2 && parts[0] == "module" {
			return parts[1]
		}
	}

	return "app"
}
