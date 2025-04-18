package forge

import (
	"github.com/gofiber/fiber/v2"
)

type Context struct {
	*fiber.Ctx
	app *Application
}

type H map[string]interface{}

func NewContext(c *fiber.Ctx, app *Application) *Context {
	return &Context{
		Ctx: c,
		app: app,
	}
}

func (c *Context) App() *Application {
	return c.app
}

func (c *Context) JSON(data interface{}) error {
	return c.Ctx.JSON(data)
}

func (c *Context) Bind(v interface{}) error {
	return c.Ctx.BodyParser(v)
}

func (c *Context) Validate(v interface{}) error {
	return validate.Struct(v)
}

func (c *Context) Param(name string) string {
	return c.Ctx.Params(name)
}

func (c *Context) Query(key string) string {
	return c.Ctx.Query(key)
}

func (c *Context) Header(key string) string {
	return c.Ctx.Get(key)
}

func (c *Context) SetHeader(key, value string) {
	c.Ctx.Set(key, value)
}

func (c *Context) Status(code int) *Context {
	c.Ctx.Status(code)
	return c
}