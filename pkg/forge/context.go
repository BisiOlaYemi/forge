package forge

import (
	"github.com/gofiber/fiber/v2"
)

// Context wraps Fiber's context with additional Forge-specific functionality
type Context struct {
	*fiber.Ctx
}

// H is a helper type for JSON responses
type H map[string]interface{}

// NewContext creates a new Forge context from a Fiber context
func NewContext(c *fiber.Ctx) *Context {
	return &Context{Ctx: c}
}

// JSON sends a JSON response
func (c *Context) JSON(data interface{}) error {
	return c.Ctx.JSON(data)
}

// Bind validates and binds the request body to a struct
func (c *Context) Bind(v interface{}) error {
	if err := c.Ctx.BodyParser(v); err != nil {
		return err
	}
	return validate.Struct(v)
}

// Validate performs validation on a struct
func (c *Context) Validate(v interface{}) error {
	return validate.Struct(v)
}

// Param returns a route parameter
func (c *Context) Param(key string) string {
	return c.Ctx.Params(key)
}

// Query returns a query parameter
func (c *Context) Query(key string) string {
	return c.Ctx.Query(key)
}

// Header returns a request header
func (c *Context) Header(key string) string {
	return c.Ctx.Get(key)
}

// SetHeader sets a response header
func (c *Context) SetHeader(key, value string) {
	c.Ctx.Set(key, value)
}

// Status sets the response status code
func (c *Context) Status(code int) *Context {
	c.Ctx.Status(code)
	return c
} 