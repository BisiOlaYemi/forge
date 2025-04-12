package forge

import (
	"reflect"
	"strings"

	"github.com/gofiber/fiber/v2"
)


type Controller struct {
	app *Application
}


type RouteMetadata struct {
	Method      string
	Path        string
	Description string
	RequestBody interface{}
	Response    interface{}
}

type HandlerFunc func(*Context) error

func (c *Controller) RegisterRoutes(router fiber.Router) {
	t := reflect.TypeOf(c)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if strings.HasPrefix(method.Name, "Handle") {
			c.registerRoute(router, method)
		}
	}
}

func (c *Controller) registerRoute(router fiber.Router, method reflect.Method) {
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

func (c *Controller) SetApplication(app *Application) {
	c.app = app
}

func (c *Controller) App() *Application {
	return c.app
} 