package forge

import (
	"fmt"
	"sync"
	"time"

	"github.com/forge/framework/pkg/forge/auth"
	"github.com/forge/framework/pkg/forge/mailer"
	"github.com/forge/framework/pkg/forge/plugin"
	"github.com/forge/framework/pkg/forge/queue"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var validate = validator.New()

// Application represents a Forge application
type Application struct {
	config     *Config
	app        *fiber.App
	controllers []Controller
	database   *Database
	auth       *auth.JWTManager
	queue      *queue.Queue
	mailer     *mailer.Mailer
	plugins    *plugin.Manager
	mu         sync.RWMutex
}

// Config represents application configuration
type Config struct {
	Name     string         `mapstructure:"name"`
	Port     int            `mapstructure:"port"`
	Root     string         `mapstructure:"root"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Queue    QueueConfig    `mapstructure:"queue"`
	Mailer   MailerConfig   `mapstructure:"mailer"`
	Plugins  PluginsConfig  `mapstructure:"plugins"`
}

type DatabaseConfig struct {
	Driver        string `mapstructure:"driver"`
	Name          string `mapstructure:"name"`
	SlowThreshold int    `mapstructure:"slow_threshold"`
}

type AuthConfig struct {
	SecretKey     string        `mapstructure:"secret_key"`
	TokenDuration time.Duration `mapstructure:"token_duration"`
}

type QueueConfig struct {
	Host     string `mapstructure:"host"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type MailerConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	From        string `mapstructure:"from"`
	TemplateDir string `mapstructure:"template_dir"`
}

type PluginsConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	PluginDir  string `mapstructure:"plugin_dir"`
}

// New creates a new Forge application
func New(config *Config) (*Application, error) {
	if config == nil {
		config = &Config{
			Name: "forge-app",
			Port: 3000,
			Root: ".",
			Database: DatabaseConfig{
				Driver:        "sqlite",
				Name:          "forge.db",
				SlowThreshold: 200,
			},
			Auth: AuthConfig{
				SecretKey:     "your-secret-key",
				TokenDuration: 24 * time.Hour,
			},
			Queue: QueueConfig{
				Host:     "localhost:6379",
				Password: "",
				DB:       0,
			},
			Mailer: MailerConfig{
				Host:        "smtp.gmail.com",
				Port:        587,
				TemplateDir: "templates/email",
			},
			Plugins: PluginsConfig{
				Enabled:   true,
				PluginDir: "plugins",
			},
		}
	}

	app := &Application{
		config:     config,
		app:        fiber.New(),
		controllers: make([]Controller, 0),
	}

	// Initialize database
	db, err := NewDatabase(config.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	app.database = db

	// Initialize auth
	app.auth = auth.NewJWTManager(config.Auth.SecretKey, config.Auth.TokenDuration)

	// Initialize queue
	q, err := queue.New(config.Queue.Host, config.Queue.Password, config.Queue.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize queue: %w", err)
	}
	app.queue = q

	// Initialize mailer
	m, err := mailer.New(mailer.Config{
		Host:        config.Mailer.Host,
		Port:        config.Mailer.Port,
		Username:    config.Mailer.Username,
		Password:    config.Mailer.Password,
		From:        config.Mailer.From,
		TemplateDir: config.Mailer.TemplateDir,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize mailer: %w", err)
	}
	app.mailer = m

	// Initialize plugins
	if config.Plugins.Enabled {
		app.plugins = plugin.NewManager(app, config.Plugins.PluginDir)
		if err := app.plugins.LoadPlugins(); err != nil {
			return nil, fmt.Errorf("failed to load plugins: %w", err)
		}
	}

	return app, nil
}

// RegisterController registers a controller
func (app *Application) RegisterController(controller Controller) {
	app.mu.Lock()
	defer app.mu.Unlock()
	app.controllers = append(app.controllers, controller)
}

// Start starts the application
func (app *Application) Start() error {
	// Start queue
	app.queue.Start()

	// Start server
	return app.app.Listen(fmt.Sprintf(":%d", app.config.Port))
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
	return app.app.Shutdown()
}

// DB returns the database connection
func (app *Application) DB() *gorm.DB {
	return app.database.DB
}

// Auth returns the JWT manager
func (app *Application) Auth() *auth.JWTManager {
	return app.auth
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
	return a.app.Group(prefix)
}

// Use adds middleware to the application
func (a *Application) Use(middleware ...interface{}) {
	a.app.Use(middleware...)
}

// Get returns the underlying Fiber app
func (a *Application) Get() *fiber.App {
	return a.app
} 