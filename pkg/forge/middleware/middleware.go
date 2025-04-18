package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/BisiOlaYemi/forge/pkg/forge"
)


func RequestLogger() forge.MiddlewareFunc {
	return func(next forge.HandlerFunc) forge.HandlerFunc {
		return func(ctx *forge.Context) error {
			start := time.Now()

			
			method := ctx.Method()
			path := ctx.Path()

			
			err := next(ctx)

			
			duration := time.Since(start)

			
			if err != nil {
				ctx.App().Logger().Error("[%s] %s - %v - %s", method, path, err, duration)
			} else {
				ctx.App().Logger().Info("[%s] %s - %d - %s", method, path, ctx.Response().StatusCode(), duration)
			}

			return err
		}
	}
}


func Recover() forge.MiddlewareFunc {
	return func(next forge.HandlerFunc) forge.HandlerFunc {
		return func(ctx *forge.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					var ok bool
					if err, ok = r.(error); !ok {
						err = forge.NewAppError("Internal Server Error", 500).
							WithDetail("panic", r)
					}

					
					ctx.App().Logger().Error("Recovered from panic: %v", r)
				}
			}()

			return next(ctx)
		}
	}
}


func RequireAuth() forge.MiddlewareFunc {
	return func(next forge.HandlerFunc) forge.HandlerFunc {
		return func(ctx *forge.Context) error {
			
			token := ctx.Get("Authorization")
			if token == "" {
				return forge.ErrUnauthorized
			}

			
			auth := ctx.App().Auth()
			if auth == nil {
				ctx.App().Logger().Error("Auth is not initialized")
				return forge.ErrInternalError.WithDetail("message", "Authentication system not initialized")
			}

			
			claims, err := auth.ValidateToken(token)
			if err != nil {
				return forge.ErrUnauthorized.WithError(err)
			}

			
			ctx.Locals("user_id", claims["sub"])
			ctx.Locals("claims", claims)

			
			return next(ctx)
		}
	}
}

//  CORS headers
func CORS(options forge.CORSConfig) forge.MiddlewareFunc {
	return func(next forge.HandlerFunc) forge.HandlerFunc {
		return func(ctx *forge.Context) error {
			
			ctx.Set("Access-Control-Allow-Origin", options.AllowOrigins)
			ctx.Set("Access-Control-Allow-Methods", options.AllowMethods)
			ctx.Set("Access-Control-Allow-Headers", options.AllowHeaders)

			
			if ctx.Method() == "OPTIONS" {
				return ctx.SendStatus(204)
			}

			
			return next(ctx)
		}
	}
}

// RateLimiterConfig configures the rate limiting behavior
type RateLimiterConfig struct {
	
	Max int
	
	Window time.Duration
	KeyFunc func(*forge.Context) string
	SkipFunc func(*forge.Context) bool
}

func RateLimit(config RateLimiterConfig) forge.MiddlewareFunc {
	if config.Max <= 0 {
		config.Max = 100
	}
	if config.Window <= 0 {
		config.Window = time.Minute
	}
	if config.KeyFunc == nil {
		config.KeyFunc = func(ctx *forge.Context) string {
			return ctx.IP()
		}
	}

	type windowEntry struct {
		timestamp time.Time
		count     int
	}

	type client struct {
		windows      []windowEntry
		lastAccessed time.Time
		mu           sync.Mutex
	}

	clients := struct {
		data map[string]*client
		mu   sync.RWMutex
	}{
		data: make(map[string]*client),
	}

	// Start a cleanup goroutine
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()

			clients.mu.Lock()
			for key, c := range clients.data {
				
				if now.Sub(c.lastAccessed) > 30*time.Minute {
					delete(clients.data, key)
				}
			}
			clients.mu.Unlock()
		}
	}()

	return func(next forge.HandlerFunc) forge.HandlerFunc {
		return func(ctx *forge.Context) error {
			
			if config.SkipFunc != nil && config.SkipFunc(ctx) {
				return next(ctx)
			}

			
			key := config.KeyFunc(ctx)
			now := time.Now()

			
			clients.mu.RLock()
			c, exists := clients.data[key]
			clients.mu.RUnlock()

			if !exists {
				clients.mu.Lock()
				
				if c, exists = clients.data[key]; !exists {
					c = &client{
						windows: make([]windowEntry, 0, 10),
					}
					clients.data[key] = c
				}
				clients.mu.Unlock()
			}

			
			c.mu.Lock()
			defer c.mu.Unlock()

			
			c.lastAccessed = now

			
			cutoff := now.Add(-config.Window)
			windowStart := 0

			for i, window := range c.windows {
				if window.timestamp.After(cutoff) {
					windowStart = i
					break
				}
			}

			if windowStart > 0 {
				c.windows = c.windows[windowStart:]
			}

			
			count := 0
			for _, window := range c.windows {
				count += window.count
			}

			
			if count >= config.Max {
				ctx.Set("Retry-After", fmt.Sprintf("%d", int(config.Window.Seconds())))
				ctx.Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.Max))
				ctx.Set("X-RateLimit-Remaining", "0")
				ctx.Set("X-RateLimit-Reset", fmt.Sprintf("%d", now.Add(config.Window).Unix()))

				return forge.NewAppError("Rate limit exceeded", 429)
			}

			
			if len(c.windows) > 0 && now.Sub(c.windows[len(c.windows)-1].timestamp) < time.Second {
				
				c.windows[len(c.windows)-1].count++
			} else {
				
				c.windows = append(c.windows, windowEntry{
					timestamp: now,
					count:     1,
				})
			}

			remaining := config.Max - count - 1
			ctx.Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.Max))
			ctx.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			ctx.Set("X-RateLimit-Reset", fmt.Sprintf("%d", now.Add(config.Window).Unix()))

			
			return next(ctx)
		}
	}
}

// Timeout sets a timeout for the request
func Timeout(duration time.Duration) forge.MiddlewareFunc {
	return func(next forge.HandlerFunc) forge.HandlerFunc {
		return func(ctx *forge.Context) error {
			
			done := make(chan error)

			
			go func() {
				done <- next(ctx)
			}()

			
			select {
			case err := <-done:
				return err
			case <-time.After(duration):
				return forge.NewAppError("Request timeout", 408)
			}
		}
	}
}
