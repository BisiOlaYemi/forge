# Forge

Forge is a modern, full-stack web framework for Go — designed to combine developer happiness, performance, and structure.

## Features

- **Type-Safe Request Handling**: Struct-based binding & validation (like FastAPI)
- **Auto-generated Swagger Docs**: OpenAPI docs from your handlers
- **Modular MVC Architecture**: Controllers, Services, Models (like NestJS)
- **Microservices Support**: First-class tools for building distributed systems
- **CLI Scaffolding**: `forge make:controller`, `make:model`, `make:microservice`, etc.
- **Go-Level Performance**: Fiber/Gin speed under the hood
- **Built-in Auth**: CLI generator for login/register flow
- **Extensible Plugins**: File uploads, RBAC, jobs, and more coming
- **Full-Stack Ready**: With template support or HTMX/SPA integration

## Implementation Details

### Why Fiber?

Forge is built on top of the [Fiber](https://github.com/gofiber/fiber) web framework for several reasons:

1. **Performance**: Fiber is built on top of [fasthttp](https://github.com/valyala/fasthttp), which is significantly faster than Go's standard net/http package
2. **Express-like API**: Familiar API design for developers coming from Node.js/Express
3. **Middleware ecosystem**: Rich middleware ecosystem that we can leverage
4. **Low memory footprint**: Optimized for minimal memory usage and high concurrency

While we could have built directly on Go's standard library, we chose Fiber to provide better performance and developer experience. The Forge framework abstracts away most Fiber-specific details, allowing you to work with a clean, consistent API.

## Forge vs Fiber: Why Choose Forge?

While Forge is built on top of Fiber for its performance benefits, it offers several significant advantages:

1. **Convention over Configuration Architecture**
   - Opinionated MVC structure with clear separation of concerns
   - Controller-based routing with automatic route generation
   - Standardized project layout for sustainable development

2. **Powerful Middleware System**
   - Express.js-like middleware with next() handler functionality
   - Controller-level middleware for route-specific handling
   - Middleware groups for sharing behavior across controllers
   - Built-in middleware for common tasks (logging, auth, rate limiting)

3. **Full-Stack Development Framework**
   - Complete solution beyond just HTTP handling
   - Database integration with GORM (ORM)
   - Authentication system with JWT
   - Background job processing with queues
   - Mailing capabilities
   - Extensible plugin architecture

4. **Dual Architecture Support**
   - Monolithic applications with MVC pattern
   - Microservices with modern containerized structure
   - Shared tools and patterns across both architectures

5. **Developer-Friendly Tooling**
   - CLI for scaffolding new projects, controllers, models, and microservices
   - Hot reloading for rapid development
   - Automatic OpenAPI documentation generation

## Installation

```bash
go install github.com/BisiOlaYemi/forge/cmd/forge@latest
```

## Quick Start

### Monolithic Application

Create a new Forge project:

```bash
forge new myapp
cd myapp
```

Generate a controller:

```bash
forge make:controller User
```

Generate a model:

```bash
forge make:model Post --migration
```

Start the development server:

```bash
forge serve
```

### Microservice Application

Create a new microservice:

```bash
forge make:microservice user-service
cd user-service
```

Optional flags for microservice creation:
- `--with-db`: Include database integration
- `--with-auth`: Include authentication support
- `--with-cache`: Include Redis cache integration
- `--with-queue`: Include task queue support

Sample with options:

```bash
forge make:microservice payment-service --with-db --with-auth
```

## Architecture Options

Forge supports two primary architectural patterns:

### 1. Monolithic Architecture

Best for:
- Smaller teams and projects
- Rapid prototyping
- Applications with simpler domains

Structure:
```
myapp/
├── app/
│   ├── controllers/      # Route handlers
│   ├── services/         # Business logic
│   ├── models/           # DB schemas
├── config/               # App/env config
├── database/             # Migrations/seeders
├── routes/               # Route groups
├── templates/            # Optional views
├── forge.yaml            # Project config
└── main.go
```

### 2. Microservice Architecture

Best for:
- Larger teams and projects
- Complex domain boundaries
- Scalable, distributed systems

Structure:
```
service-name/
├── api/              # API layer
│   ├── handlers/     # HTTP request handlers
│   └── middleware/   # HTTP middleware
├── cmd/              # Application entry points
│   └── service-name/ # Main service executable
├── config/           # Configuration files
├── internal/         # Private application code
│   ├── models/       # Data models
│   ├── services/     # Business logic
│   └── repositories/ # Data access layer
└── pkg/              # Public libraries
    └── logger/       # Logging utilities
```

## Sample Controller

```go
package controllers

import (
	"github.com/BisiOlaYemi/forge/pkg/forge"
)

// UserController handles user-related requests
type UserController struct {
	forge.Controller
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// HandlePostLogin handles user login
// @route POST /login
// @desc Authenticate a user
// @body LoginRequest
// @response 200 { message: string }
func (c *UserController) HandlePostLogin(ctx *forge.Context) error {
	var req LoginRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.Status(400).JSON(forge.H{
			"error": "Invalid request body",
		})
	}

	// Implement actual login logic
	return ctx.JSON(forge.H{
		"message": "Welcome",
	})
}
```

## Routing and Controllers

Forge uses a convention-based approach to routing inspired by Ruby on Rails and Laravel. Controllers and their methods automatically map to HTTP routes.

### Controller Naming Convention

Controllers should be named with the `Controller` suffix:

```go
// UserController -> maps to "/user" route prefix
type UserController struct {
	forge.Controller
}

// AuthController -> maps to "/auth" route prefix
type AuthController struct {
	forge.Controller
}
```

### Method Naming Convention

Controller methods should follow this pattern:

```
Handle[HTTP Method][Action]
```

For example:

```go
// HandleGetUsers maps to GET /user
func (c *UserController) HandleGetUsers(ctx *forge.Context) error {
    // ...continue with implementation inside this wrapper
}

// HandlePostUser maps to POST /user
func (c *UserController) HandlePostUser(ctx *forge.Context) error {
    // ...continue with implementation inside this wrapper
}

// HandlePutUserById maps to PUT /user/:id
func (c *UserController) HandlePutUserById(ctx *forge.Context) error {
    id := ctx.Param("id")
    // ...continue with implementation inside this wrapper
}

// HandleDeleteUser maps to DELETE /user
func (c *UserController) HandleDeleteUser(ctx *forge.Context) error {
    // ...continue with implementation inside this wrapper
}
```

### Special Path Rules

1. `ById` in the method name automatically maps to the path pattern with `:id` parameter
2. For nested resources, use camel case: `HandleGetUserPosts` maps to GET /user/posts

### Registering Controllers

To register a controller with your Forge application:

```go
app.RegisterController(&UserController{})
app.RegisterController(&AuthController{})
```

### Complete Example: Auth Controller

Here's an example of a complete authentication controller:

```go
package controllers

import (
	"github.com/BisiOlaYemi/forge/pkg/forge"
)

// AuthController handles authentication-related requests
type AuthController struct {
	forge.Controller
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// HandlePostRegister handles user registration
// Maps to: POST /auth/register
func (c *AuthController) HandlePostRegister(ctx *forge.Context) error {
	var req RegisterRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.Status(400).JSON(forge.H{
			"error": "Invalid request body",
		})
	}
	
	if err := ctx.Validate(&req); err != nil {
		return ctx.Status(400).JSON(forge.H{
			"error": "Validation failed",
			"details": err.Error(),
		})
	}
	
	// Add user registration logic here...
	
	return ctx.Status(201).JSON(forge.H{
		"message": "User registered successfully",
	})
}

// HandlePostLogin handles user login
// Maps to: POST /auth/login
func (c *AuthController) HandlePostLogin(ctx *forge.Context) error {
	var req LoginRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.Status(400).JSON(forge.H{
			"error": "Invalid request body",
		})
	}
	
	if err := ctx.Validate(&req); err != nil {
		return ctx.Status(400).JSON(forge.H{
			"error": "Validation failed", 
			"details": err.Error(),
		})
	}
	
	// user authentication logic here...
	
	token := "sample-jwt-token" 
	
	return ctx.JSON(forge.H{
		"token": token,
		"message": "Login successful",
	})
}
```

### Starting the Server

You can impmentment start Forge application server using any of these methods:

```go
// Option 1: Using Start (recommended)
if err := app.Start(); err != nil {
    log.Fatalf("Failed to start server: %v", err)
}

// Option 2: Using Listen (Fiber-style)
if err := app.Listen(":3000"); err != nil {
    log.Fatalf("Failed to start server: %v", err)
}

// Option 3: Using Serve (net/http-style)
if err := app.Serve(); err != nil {
    log.Fatalf("Failed to start server: %v", err)
}
```

## Middleware System

Forge provides a powerful middleware system inspired by Express.js. Middleware functions have access to the request/response cycle and can:

- Execute any code
- Make changes to the request and response objects
- End the request-response cycle
- Call the next middleware in the stack

### Defining Middleware

```go
// Simple middleware function
func LoggingMiddleware(next forge.HandlerFunc) forge.HandlerFunc {
    return func(ctx *forge.Context) error {
        start := time.Now()
        
        // Call the next handler in the chain
        err := next(ctx)
        
        // Log after the request is processed
        duration := time.Since(start)
        ctx.App().Logger().Info("Request processed in %s", duration)
        
        return err
    }
}
```

### Using Middleware

Middleware can be applied at multiple levels:

#### 1. Controller-level middleware

```go
// Apply middleware to a controller
userController := &UserController{}
userController.Use(middleware.RequestLogger(), middleware.RequireAuth())

// Register the controller
app.RegisterController(userController)
```

#### 2. Controller group middleware

```go
// Create a group with shared middleware
api := (&forge.Controller{}).Group("/api")
api.Use(middleware.Recover(), middleware.RequestLogger())

// Add controllers to the group
api.Add(&UserController{})
api.Add(&ProductController{})

// Register all controllers in the group
api.Register(app)
```

#### 3. Global middleware

```go
// Apply middleware to all routes
app.Use(middleware.Recover(), middleware.RequestLogger())
```

### Built-in Middleware

Forge comes with several built-in middleware functions:

- `middleware.RequestLogger()` - Logs request information and timing
- `middleware.Recover()` - Catches panics and converts them to errors
- `middleware.RequireAuth()` - Handles authentication checks
- `middleware.CORS(options)` - Configures CORS headers
- `middleware.RateLimit(limit)` - Limits request rates
- `middleware.Timeout(duration)` - Sets a timeout for request handling

## CORS Configuration

Forge includes built-in CORS support. Configure it in your application:

```go
app, err := forge.New(&forge.Config{
    
    CORS: forge.CORSConfig{
        AllowOrigins:     "http://localhost:3000,https://ffg.com",
        AllowMethods:     "GET,POST,PUT,DELETE",
        AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
        AllowCredentials: true,
        MaxAge:           86400, 
    },
})
```

If not specified, Forge uses a permissive default CORS configuration that allows all origins.

## Database Integration

Forge uses GORM for database operations. Here's an example model:

```go
package models

// User represents a user entity
type User struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Email     string `json:"email" gorm:"uniqueIndex"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt string `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for the model
func (User) TableName() string {
	return "users"
}
```

## Configuration

Configure your application in `forge.yaml`:

```yaml
app:
  name: "myapp"
  version: "0.1.0"
  description: "A Forge application"

server:
  port: 3000
  host: localhost
  base_path: /

database:
  driver: sqlite
  name: forge.db
```

## CLI Commands

- `forge new [name]`: Create a new monolithic Forge project
- `forge make:controller [name]`: Generate a new controller
- `forge make:model [name]`: Generate a new model
- `forge make:microservice [name]`: Generate a new microservice project
- `forge serve`: Start the development server with hot reload
- `forge db:migrate`: Run database migrations
- `forge doc:generate`: Generate OpenAPI documentation

## Microservices with Forge

Forge provides first-class support for building microservices with a modern, production-ready structure:

### Features

- **Containerization**: Docker and docker-compose configurations included
- **API-First Design**: Structured API handlers and middleware
- **Configuration Management**: Environment-based configuration
- **Health Checks**: Built-in health check endpoint
- **Modern Project Layout**: Following Go best practices for project structure

### Development Workflow

1. Create a new microservice: `forge make:microservice my-service`
2. Add your business logic in the internal/services directory
3. Expose your API endpoints in the api/handlers directory
4. Run locally: `go run cmd/my-service/main.go`
5. Deploy with Docker: `docker-compose up --build`

## Contributing

Contributions are welcome! Raise any noticed issue and Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.