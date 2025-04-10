package forge

import (
	"reflect"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Controller is the base controller type that all controllers should embed
type Controller struct {
	app *Application
}

// Route metadata for OpenAPI documentation
type RouteMetadata struct {
	Method      string
	Path        string
	Description string
	RequestBody interface{}
	Response    interface{}
}

// HandlerFunc is the type for controller action methods
type HandlerFunc func(*Context) error

// RegisterRoutes registers all routes for a controller
func (c *Controller) RegisterRoutes(router fiber.Router) {
	t := reflect.TypeOf(c)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if strings.HasPrefix(method.Name, "Handle") {
			c.registerRoute(router, method)
		}
	}
}

// registerRoute registers a single route based on method name and metadata
func (c *Controller) registerRoute(router fiber.Router, method reflect.Method) {
	// Extract HTTP method and path from method name
	// Example: HandleGetUser -> GET /user
	name := strings.TrimPrefix(method.Name, "Handle")
	parts := splitCamelCase(name)
	
	var httpMethod string
	var path string
	
	if len(parts) > 0 {
		httpMethod = strings.ToUpper(parts[0])
		if len(parts) > 1 {
			path = "/" + strings.ToLower(strings.Join(parts[1:], "/"))
		}
	}

	// Create handler wrapper
	handler := func(ctx *fiber.Ctx) error {
		forgeCtx := NewContext(ctx)
		return method.Func.Call([]reflect.Value{
			reflect.ValueOf(c),
			reflect.ValueOf(forgeCtx),
		})[0].Interface().(error)
	}

	// Register route
	switch httpMethod {
	case "GET":
		router.Get(path, handler)
	case "POST":
		router.Post(path, handler)
	case "PUT":
		router.Put(path, handler)
	case "DELETE":
		router.Delete(path, handler)
	case "PATCH":
		router.Patch(path, handler)
	}
}

// splitCamelCase splits a camelCase string into parts
func splitCamelCase(s string) []string {
	var parts []string
	var current strings.Builder
	
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			parts = append(parts, current.String())
			current.Reset()
		}
		current.WriteRune(r)
	}
	
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	
	return parts
}

// SetApp sets the application instance for the controller
func (c *Controller) SetApp(app *Application) {
	c.app = app
} 