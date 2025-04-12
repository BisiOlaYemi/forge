package main

import (
	"fmt"
	"log"

	"github.com/BisiOlaYemi/forge/pkg/forge"
)

type HelloController struct {
	forge.Controller
}

// HandleGetHello handles GET /hello requests
// @route GET /hello
// @desc Get a hello message
// @response 200 object
func (c *HelloController) HandleGetHello(ctx *forge.Context) error {
	return ctx.JSON(forge.H{
		"message": "Hello from Forge!",
	})
}

func main() {
	app, err := forge.New(&forge.Config{
		Name:        "Hello Forge",
		Version:     "1.0.0",
		Description: "A simple Hello World example",
		Server: forge.ServerConfig{
			Host:     "localhost",
			Port:     3000,
			BasePath: "/",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	app.RegisterController(&HelloController{})

	fmt.Println("Server starting on http://localhost:3000")
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 