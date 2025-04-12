package forge

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/BisiOlaYemi/forge/pkg/forge/auth"
	"github.com/BisiOlaYemi/forge/pkg/forge/mailer"
	"github.com/BisiOlaYemi/forge/pkg/forge/plugin"
	"github.com/BisiOlaYemi/forge/pkg/forge/queue"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var validate = validator.New()

type Application struct {
	config      *Config
	server      *fiber.App
	validator   *validator.Validate
	database    *Database
	auth        *auth.Auth
	mailer      *mailer.Mailer
	queue       *queue.Queue
	plugins     *plugin.Manager
	mu          sync.RWMutex
	controllers []interface{}
}

type Config struct {
	Name        string
	Version     string
	Description string
	Server      ServerConfig
	Database    DatabaseConfig
	Auth        auth.Config
	Mailer      mailer.Config
	Queue       queue.Config
}

type ServerConfig struct {
	Host     string
	Port     int
	BasePath string
}

func New(config *Config) (*Application, error) {
	app := &Application{
		config:    config,
		server:    fiber.New(),
		validator: validator.New(),
	}

	if config.Database.Driver != "" {
		db, err := NewDatabase(&config.Database)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
		app.database = db
	}

	if config.Auth.SecretKey != "" {
		auth, err := auth.New(config.Auth)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize auth: %w", err)
		}
		app.auth = auth
	}

	if config.Mailer.Host != "" {
		mailer, err := mailer.New(config.Mailer)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize mailer: %w", err)
		}
		app.mailer = mailer
	}

	if config.Queue.Host != "" {
		queue, err := queue.New(config.Queue.Host, config.Queue.Password, config.Queue.DB)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize queue: %w", err)
		}
		app.queue = queue
	}

	plugins := plugin.NewManager(app, "plugins")
	if err := plugins.LoadPlugins(); err != nil {
		return nil, fmt.Errorf("failed to load plugins: %w", err)
	}
	app.plugins = plugins

	return app, nil
}

func (app *Application) GetConfig() interface{} {
	return app.config
}

func (app *Application) GetDB() interface{} {
	return app.database.DB
}

func (app *Application) GetAuth() interface{} {
	return app.auth
}

func (app *Application) GetQueue() interface{} {
	return app.queue
}

func (app *Application) GetMailer() interface{} {
	return app.mailer
}

func (app *Application) RegisterController(controller interface{}) {
	app.mu.Lock()
	defer app.mu.Unlock()

	if c, ok := controller.(interface{ SetApplication(*Application) }); ok {
		c.SetApplication(app)
	}

	app.controllers = append(app.controllers, controller)

	controllerType := reflect.TypeOf(controller)
	controllerValue := reflect.ValueOf(controller)

	for i := 0; i < controllerType.NumMethod(); i++ {
		method := controllerType.Method(i)
		if strings.HasPrefix(method.Name, "Handle") {
			httpMethod := strings.TrimPrefix(method.Name, "Handle")
			if strings.HasPrefix(httpMethod, "Get") {
				path := "/" + strings.ToLower(strings.TrimPrefix(httpMethod, "Get"))
				app.server.Get(path, func(c *fiber.Ctx) error {
					ctx := &Context{Ctx: c}
					result := method.Func.Call([]reflect.Value{controllerValue, reflect.ValueOf(ctx)})
					if len(result) > 0 && !result[0].IsNil() {
						if err, ok := result[0].Interface().(error); ok {
							return err
						}
					}
					return nil
				})
			} else if strings.HasPrefix(httpMethod, "Post") {
				path := "/" + strings.ToLower(strings.TrimPrefix(httpMethod, "Post"))
				app.server.Post(path, func(c *fiber.Ctx) error {
					ctx := &Context{Ctx: c}
					result := method.Func.Call([]reflect.Value{controllerValue, reflect.ValueOf(ctx)})
					if len(result) > 0 && !result[0].IsNil() {
						if err, ok := result[0].Interface().(error); ok {
							return err
						}
					}
					return nil
				})
			}
		}
	}
}

func (app *Application) registerWelcomeRoute() {
	app.server.Get("/", func(c *fiber.Ctx) error {
		html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
			<head>
				<title>Welcome to %s</title>
				<style>
					body { font-family: Arial, sans-serif; text-align: center; padding: 60px; background: #f8f9fa; }
					h1 { color: #343a40; }
					a { color: #007bff; text-decoration: none; font-weight: bold; }
				</style>
			</head>
			<body>
				<h1>🚀 Welcome to %s</h1>
				<p>Your Forge application is up and running!</p>
				<p><a href="/api/docs">View API Documentation</a></p>
			</body>
		</html>
		`, app.config.Name, app.config.Name)

		return c.Type("html").SendString(html)
	})
}

func (app *Application) Start() error {

	if app.queue != nil {
		app.queue.Start()
	}

	return app.server.Listen(fmt.Sprintf(":%d", app.config.Server.Port))
}

func (app *Application) Shutdown() error {
	if app.queue != nil {
		app.queue.Stop()
	}

	if app.plugins != nil {
		if err := app.plugins.UnloadPlugins(); err != nil {
			return fmt.Errorf("failed to unload plugins: %w", err)
		}
	}

	if app.database != nil {
		if err := app.database.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}

	return app.server.Shutdown()
}

func (app *Application) DB() *gorm.DB {
	return app.database.DB
}

func (app *Application) Auth() *auth.JWTManager {
	return app.auth.JWTManager
}

func (app *Application) Queue() *queue.Queue {
	return app.queue
}

func (app *Application) Mailer() *mailer.Mailer {
	return app.mailer
}

func (app *Application) Plugins() *plugin.Manager {
	return app.plugins
}

func (a *Application) Group(prefix string) fiber.Router {
	return a.server.Group(prefix)
}

func (a *Application) Use(middleware ...interface{}) {
	a.server.Use(middleware...)
}

func (a *Application) Get() *fiber.App {
	return a.server
}

func (app *Application) Test(req *http.Request) (*http.Response, error) {
	return app.server.Test(req)
}
