package forge

import (
	"fmt"
	"sync"
	"time"

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

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	Driver        string
	Name          string
	Host          string
	Port          int
	Username      string
	Password      string
	SlowThreshold time.Duration
}

// New creates a new Forge application
func New(config *Config) (*Application, error) {
	app := &Application{
		config:    config,
		server:    fiber.New(),
		validator: validator.New(),
	}

	// Initialize database
	db, err := NewDatabase(&config.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	app.database = db

	// Initialize auth
	auth, err := auth.New(config.Auth)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize auth: %w", err)
	}
	app.auth = auth

	// Initialize mailer
	mailer, err := mailer.New(config.Mailer)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize mailer: %w", err)
	}
	app.mailer = mailer

	// Initialize queue
	queue, err := queue.New(config.Queue.Host, config.Queue.Password, config.Queue.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize queue: %w", err)
	}
	app.queue = queue

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
	app.controllers = append(app.controllers, controller)
}

// Start starts the application
func (app *Application) Start() error {
	// Start queue
	app.queue.Start()

	// Start server
	return app.server.Listen(fmt.Sprintf(":%d", app.config.Server.Port))
}

// Shutdown gracefully shuts down the application
func (app *Application) Shutdown() error {
	// Stop queue
	app.queue.Stop()

	// Unload plugins
	if app.plugins != nil {
		if err := app.plugins.UnloadPlugins(); err != nil {
			return fmt.Errorf("failed to unload plugins: %w", err)
		}
	}

	// Close database
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