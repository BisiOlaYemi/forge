package forge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MicroserviceConfig represents the configuration for a microservice
type MicroserviceConfig struct {
	Name        string
	Description string
	Port        int
	WithDB      bool
	WithAuth    bool
	WithCache   bool
	WithQueue   bool
}

// DefaultMicroserviceConfig returns a default configuration for a microservice
func DefaultMicroserviceConfig() *MicroserviceConfig {
	return &MicroserviceConfig{
		Name:        "ms",
		Description: "A Forge microservice",
		Port:        8080,
		WithDB:      true,
		WithAuth:    false,
		WithCache:   false,
		WithQueue:   false,
	}
}

func CreateNewProject(projectName string) {
	fmt.Println("Creating new Forge project:", projectName)
	os.MkdirAll(projectName+"/controllers", os.ModePerm)
	os.MkdirAll(projectName+"/models", os.ModePerm)
	os.MkdirAll(projectName+"/routes", os.ModePerm)
	os.MkdirAll(projectName+"/services", os.ModePerm)
	fmt.Println("Project scaffold created.")
}

func GenerateController(name string, service string) {
	var path string
	if service != "" {
		path = fmt.Sprintf("services/%s/controllers", service)
	} else {
		path = "controllers"
	}
	os.MkdirAll(path, os.ModePerm)
	controllerFile := fmt.Sprintf("%s/%sController.go", path, name)
	content := fmt.Sprintf(`package controllers

import "fmt"

func %sController() {
	fmt.Println("%s controller logic here")
}`, name, name)
	os.WriteFile(controllerFile, []byte(content), 0644)
	fmt.Println("Controller created at:", controllerFile)
}

func GenerateModel(name string, service string) {
	var path string
	if service != "" {
		path = fmt.Sprintf("services/%s/models", service)
	} else {
		path = "models"
	}
	os.MkdirAll(path, os.ModePerm)
	modelFile := fmt.Sprintf("%s/%s.go", path, name)
	content := fmt.Sprintf(`package models

// %s represents the model structure
type %s struct {
	ID   int
	Name string
}`, name, name)
	os.WriteFile(modelFile, []byte(content), 0644)
	fmt.Println("Model created at:", modelFile)
}

func CreateMicroservice(serviceName string) {
	base := fmt.Sprintf("services/%s", serviceName)
	folders := []string{
		base + "/controllers",
		base + "/models",
		base + "/routes",
		base + "/config",
	}
	for _, folder := range folders {
		os.MkdirAll(folder, os.ModePerm)
	}
	mainFile := base + "/main.go"
	mainContent := fmt.Sprintf(`package main

import "fmt"

func main() {
	fmt.Println("%s microservice started")
}`, serviceName)
	os.WriteFile(mainFile, []byte(mainContent), 0644)
	fmt.Println("Microservice scaffold created at:", base)
}

// CreateMicroserviceProject creates a new microservice project with a comprehensive structure
func CreateMicroserviceProject(config *MicroserviceConfig) error {
	if config == nil {
		config = DefaultMicroserviceConfig()
	}

	name := config.Name
	if name == "" {
		return fmt.Errorf("microservice name cannot be empty")
	}

	// Create the base directory
	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("failed to create microservice directory: %w", err)
	}

	// Create the directory structure
	dirs := []string{
		filepath.Join(name, "api"),
		filepath.Join(name, "api", "handlers"),
		filepath.Join(name, "api", "middleware"),
		filepath.Join(name, "internal", "models"),
		filepath.Join(name, "internal", "services"),
		filepath.Join(name, "internal", "repositories"),
		filepath.Join(name, "pkg", "logger"),
		filepath.Join(name, "config"),
		filepath.Join(name, "cmd", name),
	}

	if config.WithDB {
		dirs = append(dirs, filepath.Join(name, "internal", "database"))
		dirs = append(dirs, filepath.Join(name, "migrations"))
	}

	if config.WithCache {
		dirs = append(dirs, filepath.Join(name, "internal", "cache"))
	}

	if config.WithQueue {
		dirs = append(dirs, filepath.Join(name, "internal", "queue"))
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create the main.go file
	mainContent := generateMicroserviceMainFile(config)
	if err := os.WriteFile(filepath.Join(name, "cmd", name, "main.go"), []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}

	// Create config file
	configContent := generateMicroserviceConfigFile(config)
	if err := os.WriteFile(filepath.Join(name, "config", "config.yaml"), []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create config.yaml: %w", err)
	}

	// Create Dockerfile
	dockerfileContent := generateDockerfile(config)
	if err := os.WriteFile(filepath.Join(name, "Dockerfile"), []byte(dockerfileContent), 0644); err != nil {
		return fmt.Errorf("failed to create Dockerfile: %w", err)
	}

	// Create docker-compose.yml
	dockerComposeContent := generateDockerCompose(config)
	if err := os.WriteFile(filepath.Join(name, "docker-compose.yml"), []byte(dockerComposeContent), 0644); err != nil {
		return fmt.Errorf("failed to create docker-compose.yml: %w", err)
	}

	// Create go.mod file
	modContent := fmt.Sprintf(`module github.com/%s

go 1.23

require (
	github.com/BisiOlaYemi/forge v0.0.0-20250410105738-69dbba69f7f0
	github.com/gofiber/fiber/v2 v2.52.6
)
`, name)
	if err := os.WriteFile(filepath.Join(name, "go.mod"), []byte(modContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}

	// Create sample handler
	handlerContent := generateSampleHandler(config)
	if err := os.WriteFile(filepath.Join(name, "api", "handlers", "health.go"), []byte(handlerContent), 0644); err != nil {
		return fmt.Errorf("failed to create sample handler: %w", err)
	}

	// Create README.md
	readmeContent := generateMicroserviceReadme(config)
	if err := os.WriteFile(filepath.Join(name, "README.md"), []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	fmt.Printf("Created new Forge microservice: %s\n", name)
	return nil
}

func generateMicroserviceMainFile(config *MicroserviceConfig) string {
	return fmt.Sprintf(`package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BisiOlaYemi/forge/pkg/forge"
)

func main() {
	// Create a new Forge application
	app, err := forge.New(&forge.Config{
		Name:        "%s",
		Version:     "1.0.0",
		Description: "%s",
		Server: forge.ServerConfig{
			Host:     "0.0.0.0",
			Port:     %d,
			BasePath: "/api",
		},
		%s
	})
	if err != nil {
		log.Fatalf("Failed to create application: %%v", err)
	}

	// Configure API routes
	app.Get().Get("/health", func(c *forge.Context) error {
		return c.JSON(map[string]string{
			"status":  "ok",
			"service": "%s",
			"version": "1.0.0",
		})
	})

	// Register API handlers
	// TODO: Add your handlers here

	// Handle graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit
		fmt.Println("Shutting down server...")
		if err := app.Shutdown(); err != nil {
			log.Fatalf("Error during shutdown: %%v", err)
		}
	}()

	// Start the server
	fmt.Printf("Server starting on http://0.0.0.0:%d/api\n", %d)
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start server: %%v", err)
	}
}
`, 
	config.Name, 
	config.Description, 
	config.Port,
	generateConfigOptions(config),
	config.Name,
	config.Port,
	config.Port)
}

func generateConfigOptions(config *MicroserviceConfig) string {
	options := ""
	
	if config.WithDB {
		options += `Database: forge.DatabaseConfig{
			Driver: "sqlite",
			Name:   "forge.db",
			// Uncomment these for production use
			// Driver:   "postgres",  
			// Host:     "db",
			// Port:     5432,
			// Username: "postgres",
			// Password: "postgres",
			// Name:     "forge",
		},`
	}
	
	return options
}

func generateMicroserviceConfigFile(config *MicroserviceConfig) string {
	return fmt.Sprintf(`# %s Microservice Configuration

# Service Settings
service:
  name: "%s"
  version: "1.0.0"
  description: "%s"
  environment: "development"
  debug: true

# Server Configuration
server:
  host: "0.0.0.0"
  port: %d
  base_path: "/api"
  read_timeout: 10s
  write_timeout: 10s
  idle_timeout: 120s

%s
`, 
	config.Name, 
	config.Name, 
	config.Description, 
	config.Port,
	generateAdditionalConfig(config))
}

func generateAdditionalConfig(config *MicroserviceConfig) string {
	var additionalConfig string
	
	if config.WithDB {
		additionalConfig += `# Database Configuration
database:
  driver: "sqlite"
  name: "forge.db"
  # For production:
  # driver: "postgres"
  # host: "db"
  # port: 5432
  # username: "postgres"
  # password: "postgres"
  # name: "forge"
  max_open_conns: 20
  max_idle_conns: 5
  conn_max_life: 300s

`
	}
	
	if config.WithCache {
		additionalConfig += `# Cache Configuration
cache:
  driver: "redis"
  host: "cache"
  port: 6379
  prefix: "forge:"
  ttl: 3600s

`
	}
	
	if config.WithQueue {
		additionalConfig += `# Queue Configuration
queue:
  driver: "redis"
  host: "queue"
  port: 6379
  db: 1

`
	}
	
	if config.WithAuth {
		additionalConfig += `# Authentication Configuration
auth:
  jwt:
    secret: "change-this-to-a-secure-secret-in-production"
    expiration: 86400s # 24 hours
    refresh_expiration: 604800s # 7 days

`
	}
	
	return additionalConfig
}

func generateDockerfile(config *MicroserviceConfig) string {
	return `# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o service ./cmd/` + config.Name + `

# Final stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/service .
COPY --from=builder /app/config ./config

RUN chmod +x service

EXPOSE ` + fmt.Sprintf("%d", config.Port) + `

ENTRYPOINT ["./service"]
`
}

func generateDockerCompose(config *MicroserviceConfig) string {
	services := `version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "` + fmt.Sprintf("%d:%d", config.Port, config.Port) + `"
    restart: unless-stopped
`

	if config.WithDB {
		services += `    depends_on:
      - db
    environment:
      - DB_HOST=db
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=postgres
      - DB_NAME=forge

  db:
    image: postgres:14-alpine
    volumes:
      - postgres-data:/var/lib/postgresql/data
    environment:
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_USER=postgres
      - POSTGRES_DB=forge
    ports:
      - "5432:5432"
`
	}

	if config.WithCache {
		services += `
  cache:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
`
	}

	volumes := `
volumes:`

	if config.WithDB {
		volumes += `
  postgres-data:`
	}

	if config.WithCache {
		volumes += `
  redis-data:`
	}

	if !config.WithDB && !config.WithCache {
		volumes = ""
	}

	return services + volumes
}

func generateSampleHandler(config *MicroserviceConfig) string {
	return `package handlers

import (
	"github.com/BisiOlaYemi/forge/pkg/forge"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	forge.Controller
}

// HandleGetHealth handles GET /health requests
func (h *HealthHandler) HandleGetHealth(ctx *forge.Context) error {
	return ctx.JSON(map[string]interface{}{
		"status":  "ok",
		"service": "` + config.Name + `",
		"version": "1.0.0",
	})
}
`
}

func generateMicroserviceReadme(config *MicroserviceConfig) string {
	return fmt.Sprintf(`# %s Microservice

%s

## Getting Started

### Running Locally

1. Start the service:

   ` + "```" + `bash
   go run cmd/%s/main.go
   ` + "```" + `

2. The service will be available at http://localhost:%d/api

### Using Docker

1. Build and start the service:

   ` + "```" + `bash
   docker-compose up --build
   ` + "```" + `

2. The service will be available at http://localhost:%d/api

## API Endpoints

- **Health Check**: GET /api/health

## Project Structure

` + "```" + `
%s/
├── api/              # API layer
│   ├── handlers/     # HTTP request handlers
│   └── middleware/   # HTTP middleware
├── cmd/              # Application entry points
│   └── %s/           # Main service executable
├── config/           # Configuration files
├── internal/         # Private application code
│   ├── models/       # Data models
│   ├── services/     # Business logic
│   └── repositories/ # Data access layer
└── pkg/              # Public libraries
    └── logger/       # Logging utilities
` + "```" + `

## Configuration

Configuration is managed through the ` + "```" + `config/config.yaml` + "```" + ` file and environment variables.

## Development

This service is built with the Forge Framework, which provides a modern Go web application architecture.

`, 
	strings.ToTitle(config.Name),
	config.Description,
	config.Name,
	config.Port,
	config.Port,
	config.Name,
	config.Name)
}
