package forge

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"unicode"

	"github.com/BisiOlaYemi/forge/pkg/forge/auth"
	"github.com/BisiOlaYemi/forge/pkg/forge/logger"
	"github.com/BisiOlaYemi/forge/pkg/forge/mailer"
	"github.com/BisiOlaYemi/forge/pkg/forge/plugin"
	"github.com/BisiOlaYemi/forge/pkg/forge/queue"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiblogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
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
	logger      *logger.Logger
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
	CORS        CORSConfig
	LogLevel    string
}

type ServerConfig struct {
	Host     string
	Port     int
	BasePath string
}


type CORSConfig struct {
	AllowOrigins     string `yaml:"allow_origins"`
	AllowMethods     string `yaml:"allow_methods"`
	AllowHeaders     string `yaml:"allow_headers"`
	AllowCredentials bool   `yaml:"allow_credentials"`
	ExposeHeaders    string `yaml:"expose_headers"`
	MaxAge           int    `yaml:"max_age"`
}


func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
		MaxAge:           86400,
	}
}

func New(config *Config) (*Application, error) {
	fiberConfig := fiber.Config{
		AppName:      config.Name,
		ErrorHandler: defaultErrorHandler,
	}

	app := &Application{
		config:    config,
		server:    fiber.New(fiberConfig),
		validator: validator.New(),
	}

	// Configure logger
	logLevel := logger.LevelInfo
	if config.LogLevel != "" {
		logLevel = logger.ParseLevel(config.LogLevel)
	}

	log := logger.New(logger.Config{
		Level: logLevel,
	})
	log.Info("Initializing Forge application: %s v%s", config.Name, config.Version)
	app.logger = log

	
	app.server.Use(recover.New())
	app.server.Use(fiblogger.New())

	
	corsConfig := config.CORS
	if corsConfig.AllowOrigins == "" {
		corsConfig = DefaultCORSConfig()
	}
	app.server.Use(cors.New(cors.Config{
		AllowOrigins:     corsConfig.AllowOrigins,
		AllowMethods:     corsConfig.AllowMethods,
		AllowHeaders:     corsConfig.AllowHeaders,
		AllowCredentials: corsConfig.AllowCredentials,
		ExposeHeaders:    corsConfig.ExposeHeaders,
		MaxAge:           corsConfig.MaxAge,
	}))

	if config.Database.Driver != "" {
		log.Info("Initializing database connection: %s", config.Database.Driver)
		db, err := NewDatabase(&config.Database)
		if err != nil {
			log.Error("Failed to initialize database: %v", err)
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
		app.database = db
		log.Info("Database connection established")
	}

	if config.Auth.SecretKey != "" {
		log.Info("Initializing authentication")
		auth, err := auth.New(config.Auth)
		if err != nil {
			log.Error("Failed to initialize auth: %v", err)
			return nil, fmt.Errorf("failed to initialize auth: %w", err)
		}
		app.auth = auth
		log.Info("Authentication initialized")
	}

	if config.Mailer.Host != "" {
		log.Info("Initializing mailer")
		mailer, err := mailer.New(config.Mailer)
		if err != nil {
			log.Error("Failed to initialize mailer: %v", err)
			return nil, fmt.Errorf("failed to initialize mailer: %w", err)
		}
		app.mailer = mailer
		log.Info("Mailer initialized")
	}

	if config.Queue.Host != "" {
		log.Info("Initializing message queue")
		queue, err := queue.New(config.Queue.Host, config.Queue.Password, config.Queue.DB)
		if err != nil {
			log.Error("Failed to initialize queue: %v", err)
			return nil, fmt.Errorf("failed to initialize queue: %w", err)
		}
		app.queue = queue
		log.Info("Message queue initialized")
	}

	log.Info("Loading plugins")
	plugins := plugin.NewManager(app, "plugins")
	if err := plugins.LoadPlugins(); err != nil {
		log.Error("Failed to load plugins: %v", err)
		return nil, fmt.Errorf("failed to load plugins: %w", err)
	}
	app.plugins = plugins
	log.Info("Plugins loaded successfully")

	app.server.Get("/", func(c *fiber.Ctx) error {
		return c.Type("html").SendString(`
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<title>🔥 Forge</title>
				<style>
					body { font-family: sans-serif; text-align: center; padding: 50px; background-color:rgb(11, 10, 10); }
					h1 { font-size: 2.5em; color:rgb(209, 219, 231); }
					p { font-size: 1.2em; color: #fff; }
					a { color: #007BFF; text-decoration: none; }
					a:hover { text-decoration: underline; }
				</style>
			</head>
			<body>
				<h1>Welcome to 🔥 Forge</h1>
				<p>Built with love by <strong>Yemi Ogunrinde</strong></p>
				<p>Version: <strong>` + config.Version + `</strong></p>
			</body>
			</html>
		`)
	})

	log.Info("Forge application initialized successfully")
	return app, nil
}


func defaultErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	// Retrieve the custom status code if it's a fiber.*Error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error":   true,
		"message": err.Error(),
	})
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

	
	controllerName := controllerType.Elem().Name()
	controllerBaseName := strings.TrimSuffix(controllerName, "Controller")
	basePath := "/" + strings.ToLower(controllerBaseName)

	for i := 0; i < controllerType.NumMethod(); i++ {
		method := controllerType.Method(i)

		
		if !strings.HasPrefix(method.Name, "Handle") {
			continue
		}

		
		routeInfo := parseRouteFromMethodName(method.Name, basePath)

		
		handler := createHandlerFunc(method, controllerValue)

		// Route is Registered with the fiber app
		switch routeInfo.HTTPMethod {
		case "GET":
			app.server.Get(routeInfo.Path, handler)
		case "POST":
			app.server.Post(routeInfo.Path, handler)
		case "PUT":
			app.server.Put(routeInfo.Path, handler)
		case "DELETE":
			app.server.Delete(routeInfo.Path, handler)
		case "PATCH":
			app.server.Patch(routeInfo.Path, handler)
		case "OPTIONS":
			app.server.Options(routeInfo.Path, handler)
		case "HEAD":
			app.server.Head(routeInfo.Path, handler)
		}
	}
}


type RouteInfo struct {
	HTTPMethod string
	Path       string
}


func parseRouteFromMethodName(methodName string, basePath string) RouteInfo {
	
	actionName := strings.TrimPrefix(methodName, "Handle")

	
	httpMethod := "GET"

	
	for _, method := range []string{"Get", "Post", "Put", "Delete", "Patch", "Options", "Head"} {
		if strings.HasPrefix(actionName, method) {
			httpMethod = strings.ToUpper(method)
			actionName = strings.TrimPrefix(actionName, method)
			break
		}
	}

	
	if actionName != "" {
		
		var path strings.Builder
		for i, r := range actionName {
			if i > 0 && r >= 'A' && r <= 'Z' {
				path.WriteRune('-')
			}
			path.WriteRune(unicode.ToLower(r))
		}

		actionPath := path.String()

		
		if actionPath == "index" || actionPath == "" {
			return RouteInfo{
				HTTPMethod: httpMethod,
				Path:       basePath,
			}
		}

		// Special case for "By" patterns like GetUserById -> /users/:id
		if strings.Contains(actionPath, "by-id") {
			return RouteInfo{
				HTTPMethod: httpMethod,
				Path:       fmt.Sprintf("%s/:id", basePath),
			}
		}

		return RouteInfo{
			HTTPMethod: httpMethod,
			Path:       fmt.Sprintf("%s/%s", basePath, actionPath),
		}
	}

	return RouteInfo{
		HTTPMethod: httpMethod,
		Path:       basePath,
	}
}


func createHandlerFunc(method reflect.Method, controllerValue reflect.Value) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := &Context{Ctx: c}
		result := method.Func.Call([]reflect.Value{controllerValue, reflect.ValueOf(ctx)})
		if len(result) > 0 && !result[0].IsNil() {
			if err, ok := result[0].Interface().(error); ok {
				return err
			}
		}
		return nil
	}
}


func (app *Application) Start() error {
	if app.queue != nil {
		app.queue.Start()
	}

	return app.server.Listen(fmt.Sprintf("%s:%d", app.config.Server.Host, app.config.Server.Port))
}


func (app *Application) Listen(addr string) error {
	if app.queue != nil {
		app.queue.Start()
	}

	return app.server.Listen(addr)
}

// Serve is an alias for Start to provide a more familiar API to users coming from net/http
func (app *Application) Serve() error {
	return app.Start()
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

func (app *Application) Logger() *logger.Logger {
	return app.logger
}

func (app *Application) WithLogField(key string, value interface{}) *logger.Logger {
	return app.logger.WithField(key, value)
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
