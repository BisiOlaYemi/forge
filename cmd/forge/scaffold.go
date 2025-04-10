package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectTemplate represents a project template
type ProjectTemplate struct {
	Name        string
	Description string
	Files       map[string]string
}

// ControllerTemplate represents a controller template
type ControllerTemplate struct {
	Name        string
	Description string
	Methods     []string
}

// ModelTemplate represents a model template
type ModelTemplate struct {
	Name        string
	Description string
	Fields      []string
}

// createNewProject creates a new Forge project
func createNewProject(name string) error {
	// Create project directory
	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create project structure
	dirs := []string{
		filepath.Join(name, "app", "controllers"),
		filepath.Join(name, "app", "models"),
		filepath.Join(name, "app", "services"),
		filepath.Join(name, "config"),
		filepath.Join(name, "database", "migrations"),
		filepath.Join(name, "database", "seeders"),
		filepath.Join(name, "routes"),
		filepath.Join(name, "templates"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create main.go
	mainContent := `package main

import (
	"github.com/forge/framework/pkg/forge"
)

func main() {
	app := forge.New(&forge.Config{
		Port:     "3000",
		Host:     "localhost",
		BasePath: "/",
	})

	// Register controllers
	// app.RegisterController(&UserController{})

	// Start the server
	if err := app.Start(); err != nil {
		panic(err)
	}
}
`

	if err := os.WriteFile(filepath.Join(name, "main.go"), []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}

	// Create forge.yaml
	configContent := `app:
  name: "` + name + `"
  version: "0.1.0"
  description: "A Forge application"

server:
  port: 3000
  host: localhost
  base_path: /

database:
  driver: sqlite
  name: forge.db
`

	if err := os.WriteFile(filepath.Join(name, "forge.yaml"), []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create forge.yaml: %w", err)
	}

	// Create go.mod
	modContent := `module ` + name + `

go 1.21

require (
	github.com/forge/framework v0.1.0
)
`

	if err := os.WriteFile(filepath.Join(name, "go.mod"), []byte(modContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
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
	"github.com/forge/framework/pkg/forge"
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
	return ctx.JSON([]interface{}{})
}

// HandleGet` + strings.TrimSuffix(name, "Controller") + `ByID handles getting a ` + strings.TrimSuffix(name, "Controller") + ` by ID
// @route GET /` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `s/:id
// @desc Get a ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + ` by ID
// @param id path int true "` + strings.TrimSuffix(name, "Controller") + ` ID"
// @response 200 ` + strings.TrimSuffix(name, "Controller") + `
func (c *` + name + `) HandleGet` + strings.TrimSuffix(name, "Controller") + `ByID(ctx *forge.Context) error {
	id := ctx.Param("id")
	return ctx.JSON(map[string]interface{}{"id": id})
}

// HandleCreate` + strings.TrimSuffix(name, "Controller") + ` handles creating a new ` + strings.TrimSuffix(name, "Controller") + `
// @route POST /` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `s
// @desc Create a new ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `
// @body Create` + strings.TrimSuffix(name, "Controller") + `Request
// @response 201 ` + strings.TrimSuffix(name, "Controller") + `
func (c *` + name + `) HandleCreate` + strings.TrimSuffix(name, "Controller") + `(ctx *forge.Context) error {
	return ctx.Status(201).JSON(map[string]interface{}{})
}

// HandleUpdate` + strings.TrimSuffix(name, "Controller") + ` handles updating a ` + strings.TrimSuffix(name, "Controller") + `
// @route PUT /` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `s/:id
// @desc Update a ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `
// @param id path int true "` + strings.TrimSuffix(name, "Controller") + ` ID"
// @body Update` + strings.TrimSuffix(name, "Controller") + `Request
// @response 200 ` + strings.TrimSuffix(name, "Controller") + `
func (c *` + name + `) HandleUpdate` + strings.TrimSuffix(name, "Controller") + `(ctx *forge.Context) error {
	id := ctx.Param("id")
	return ctx.JSON(map[string]interface{}{"id": id})
}

// HandleDelete` + strings.TrimSuffix(name, "Controller") + ` handles deleting a ` + strings.TrimSuffix(name, "Controller") + `
// @route DELETE /` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `s/:id
// @desc Delete a ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `
// @param id path int true "` + strings.TrimSuffix(name, "Controller") + ` ID"
// @response 204
func (c *` + name + `) HandleDelete` + strings.TrimSuffix(name, "Controller") + `(ctx *forge.Context) error {
	id := ctx.Param("id")
	return ctx.Status(204).Send([]byte{})
}
`

	// Create model file
	modelContent := `package models

// ` + strings.TrimSuffix(name, "Controller") + ` represents a ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + ` entity
type ` + strings.TrimSuffix(name, "Controller") + ` struct {
	ID        uint   ` + "`json:\"id\"`" + `
	CreatedAt string ` + "`json:\"created_at\"`" + `
	UpdatedAt string ` + "`json:\"updated_at\"`" + `
}
`

	// Create request/response types
	typesContent := `package controllers

// Create` + strings.TrimSuffix(name, "Controller") + `Request represents the request body for creating a ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `
type Create` + strings.TrimSuffix(name, "Controller") + `Request struct {
	// Add fields here
}

// Update` + strings.TrimSuffix(name, "Controller") + `Request represents the request body for updating a ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `
type Update` + strings.TrimSuffix(name, "Controller") + `Request struct {
	// Add fields here
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
	// Convert name to proper case
	name = strings.ToUpper(name[:1]) + name[1:]

	// Create model file
	modelContent := `package models

// ` + name + ` represents a ` + strings.ToLower(name) + ` entity
type ` + name + ` struct {
	ID        uint   ` + "`json:\"id\" gorm:\"primaryKey\"`" + `
	CreatedAt string ` + "`json:\"created_at\" gorm:\"autoCreateTime\"`" + `
	UpdatedAt string ` + "`json:\"updated_at\" gorm:\"autoUpdateTime\"`" + `
	// Add fields here
}

// TableName returns the table name for the model
func (` + name + `) TableName() string {
	return "` + strings.ToLower(name) + `s"
}
`

	// Create migration file
	migrationContent := `package migrations

import (
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

	// Write files
	if err := os.WriteFile(filepath.Join("app", "models", strings.ToLower(name)+".go"), []byte(modelContent), 0644); err != nil {
		return fmt.Errorf("failed to create model file: %w", err)
	}

	if err := os.WriteFile(filepath.Join("database", "migrations", strings.ToLower(name)+"_migration.go"), []byte(migrationContent), 0644); err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}

	fmt.Printf("Generated model: %s\n", name)
	return nil
} 