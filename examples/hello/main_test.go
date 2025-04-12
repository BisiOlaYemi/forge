package main

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/BisiOlaYemi/forge/pkg/forge"
	"github.com/stretchr/testify/assert"
)

func TestHelloEndpoint(t *testing.T) {
	app, err := forge.New(&forge.Config{
		Name:        "Hello Forge Test",
		Version:     "1.0.0",
		Description: "Test instance of Hello World",
		Server: forge.ServerConfig{
			Host:     "localhost",
			Port:     3000,
			BasePath: "/",
		},
	})
	assert.NoError(t, err)

	
	app.RegisterController(&HelloController{})

	
	req := httptest.NewRequest("GET", "/hello", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	assert.NoError(t, err)
	assert.Equal(t, "Hello from Forge!", response["message"])
} 