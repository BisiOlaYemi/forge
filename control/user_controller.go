package control

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

// LoginResponse represents the login response
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// User represents a user entity
type User struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// HandlePostLogin handles user login
// @route POST /login
// @desc Authenticate a user
// @body LoginRequest
// @response 200 LoginResponse
func (c *UserController) HandlePostLogin(ctx *forge.Context) error {
	var req LoginRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.Status(400).JSON(forge.H{
			"error": "Invalid request body",
		})
	}

	// login logic
	response := LoginResponse{
		Token: "dummy-token",
		User: User{
			ID:    1,
			Email: req.Email,
			Name:  "Go Forge",
		},
	}

	return ctx.JSON(response)
}

// HandleGetUser handles getting a user by ID
// @route GET /users/:id
// @desc Get a user by ID
// @param id path int true "User ID"
// @response 200 User
func (c *UserController) HandleGetUser(ctx *forge.Context) error {
	
	user := User{
		ID:    1,
		Email: "fgo@example.com",
		Name:  "Go Forge",
	}

	return ctx.JSON(user)
} 