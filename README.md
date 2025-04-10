# Forge

Forge is a modern, full-stack web framework for Go — designed to combine developer happiness, performance, and structure.

## Features

- **Type-Safe Request Handling**: Struct-based binding & validation (like FastAPI)
- **Auto-generated Swagger Docs**: OpenAPI docs from your handlers
- **Modular MVC Architecture**: Controllers, Services, Models (like NestJS)
- **CLI Scaffolding**: `forge make:controller`, `make:model`, `serve`, etc.
- **Go-Level Performance**: Fiber/Gin speed under the hood
- **Built-in Auth**: CLI generator for login/register flow
- **Extensible Plugins**: File uploads, RBAC, jobs, and more coming
- **Full-Stack Ready**: With template support or HTMX/SPA integration

## Installation

```bash
go install github.com/forge/framework/cmd/forge@latest
```

## Quick Start

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

## Project Structure

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

## Example Controller

```go
package controllers

import (
	"github.com/forge/framework/pkg/forge"
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

	// TODO: Implement actual login logic
	return ctx.JSON(forge.H{
		"message": "Welcome",
	})
}
```

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

- `forge new [name]`: Create a new Forge project
- `forge make:controller [name]`: Generate a new controller
- `forge make:model [name]`: Generate a new model
- `forge serve`: Start the development server with hot reload
- `forge db:migrate`: Run database migrations
- `forge doc:generate`: Generate OpenAPI documentation

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 