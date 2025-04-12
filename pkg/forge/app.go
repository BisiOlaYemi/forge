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

// Application represents the Forge application
type Application struct {
	config     *Config
	server     *fiber.App
	validator  *validator.Validate
	database   *Database
	auth       *auth.Auth
	mailer     *mailer.Mailer
	queue      *queue.Queue
	plugins    *plugin.Manager
	mu         sync.RWMutex
	controllers []interface{}
}

// Config represents the application configuration
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

// ServerConfig represents the server configuration
type ServerConfig struct {
	Host     string
	Port     int
	BasePath string
}



// New creates a new Forge application
func New(config *Config) (*Application, error) {
	app := &Application{
		config:    config,
		server:    fiber.New(),
		validator: validator.New(),
	}

	// Initialize database if configured
	if config.Database.Driver != "" {
		db, err := NewDatabase(&config.Database)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
		app.database = db
	}

	// Initialize auth if configured
	if config.Auth.SecretKey != "" {
		auth, err := auth.New(config.Auth)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize auth: %w", err)
		}
		app.auth = auth
	}

	// Initialize mailer if configured
	if config.Mailer.Host != "" {
		mailer, err := mailer.New(config.Mailer)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize mailer: %w", err)
		}
		app.mailer = mailer
	}

	// Initialize queue if configured
	if config.Queue.Host != "" {
		queue, err := queue.New(config.Queue.Host, config.Queue.Password, config.Queue.DB)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize queue: %w", err)
		}
		app.queue = queue
	}

	// Initialize plugin manager
	plugins := plugin.NewManager(app, "plugins")
	if err := plugins.LoadPlugins(); err != nil {
		return nil, fmt.Errorf("failed to load plugins: %w", err)
	}
	app.plugins = plugins

	return app, nil
}

// GetConfig returns the application configuration
func (app *Application) GetConfig() interface{} {
	return app.config
}

// GetDB returns the database instance
func (app *Application) GetDB() interface{} {
	return app.database.DB
}

// GetAuth returns the auth instance
func (app *Application) GetAuth() interface{} {
	return app.auth
}

// GetQueue returns the queue instance
func (app *Application) GetQueue() interface{} {
	return app.queue
}

// GetMailer returns the mailer instance
func (app *Application) GetMailer() interface{} {
	return app.mailer
}

// RegisterController registers a controller with the application
func (app *Application) RegisterController(controller interface{}) {
	app.mu.Lock()
	defer app.mu.Unlock()

	// Set application instance on the controller
	if c, ok := controller.(interface{ SetApplication(*Application) }); ok {
		c.SetApplication(app)
	}

	app.controllers = append(app.controllers, controller)

	// Get controller type
	controllerType := reflect.TypeOf(controller)
	controllerValue := reflect.ValueOf(controller)

	// Register routes for each method
	for i := 0; i < controllerType.NumMethod(); i++ {
		method := controllerType.Method(i)
		if strings.HasPrefix(method.Name, "Handle") {
			// Extract HTTP method and path from method name
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

// Start starts the application
func (app *Application) Start() error {
	// Start queue if initialized
	if app.queue != nil {
		app.queue.Start()
	}

	// Start server
	return app.server.Listen(fmt.Sprintf(":%d", app.config.Server.Port))
}

// Shutdown gracefully shuts down the application
func (app *Application) Shutdown() error {
	// Stop queue if initialized
	if app.queue != nil {
		app.queue.Stop()
	}

	// Unload plugins if initialized
	if app.plugins != nil {
		if err := app.plugins.UnloadPlugins(); err != nil {
			return fmt.Errorf("failed to unload plugins: %w", err)
		}
	}

	// Close database if initialized
	if app.database != nil {
		if err := app.database.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}

	// Shutdown server
	return app.server.Shutdown()
}

// DB returns the database connection
func (app *Application) DB() *gorm.DB {
	return app.database.DB
}

// Auth returns the JWT manager
func (app *Application) Auth() *auth.JWTManager {
	return app.auth.JWTManager
}

// Queue returns the job queue
func (app *Application) Queue() *queue.Queue {
	return app.queue
}

// Mailer returns the mailer
func (app *Application) Mailer() *mailer.Mailer {
	return app.mailer
}

// Plugins returns the plugin manager
func (app *Application) Plugins() *plugin.Manager {
	return app.plugins
}

// Group creates a new route group
func (a *Application) Group(prefix string) fiber.Router {
	return a.server.Group(prefix)
}

// Use adds middleware to the application
func (a *Application) Use(middleware ...interface{}) {
	a.server.Use(middleware...)
}

// Get returns the underlying Fiber app
func (a *Application) Get() *fiber.App {
	return a.server
}

// Test performs a test request to the application
func (app *Application) Test(req *http.Request) (*http.Response, error) {
	return app.server.Test(req)
} 