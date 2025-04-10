package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BisiOlaYemi/forge/pkg/forge"
	"github.com/redis/go-redis/v9"
)

// CachePlugin implements a Redis-based caching plugin for Forge
type CachePlugin struct {
	app    *forge.Application
	client *redis.Client
	prefix string
}

// Config holds the Redis configuration
type Config struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	Prefix   string `yaml:"prefix"`
}

// New creates a new cache plugin instance
func New(app *forge.Application, config *Config) (*CachePlugin, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &CachePlugin{
		app:    app,
		client: client,
		prefix: config.Prefix,
	}, nil
}

// Shutdown closes the Redis connection
func (p *CachePlugin) Shutdown() error {
	return p.client.Close()
}

// Set stores a value in the cache with the given TTL
func (p *CachePlugin) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return p.client.Set(ctx, p.prefix+key, data, ttl).Err()
}

// Get retrieves a value from the cache
func (p *CachePlugin) Get(ctx context.Context, key string, value interface{}) error {
	data, err := p.client.Get(ctx, p.prefix+key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return forge.ErrNotFound
		}
		return err
	}

	return json.Unmarshal(data, value)
}

// Delete removes a key from the cache
func (p *CachePlugin) Delete(ctx context.Context, key string) error {
	return p.client.Del(ctx, p.prefix+key).Err()
}

// Clear removes all keys with the plugin's prefix
func (p *CachePlugin) Clear(ctx context.Context) error {
	pattern := p.prefix + "*"
	iter := p.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		if err := p.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}

// Exists checks if a key exists in the cache
func (p *CachePlugin) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := p.client.Exists(ctx, p.prefix+key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// Increment increases the value of a key by 1
func (p *CachePlugin) Increment(ctx context.Context, key string) (int64, error) {
	return p.client.Incr(ctx, p.prefix+key).Result()
}

// Decrement decreases the value of a key by 1
func (p *CachePlugin) Decrement(ctx context.Context, key string) (int64, error) {
	return p.client.Decr(ctx, p.prefix+key).Result()
}

// SetNX sets a value only if the key doesn't exist
func (p *CachePlugin) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	return p.client.SetNX(ctx, p.prefix+key, data, ttl).Result()
}

// GetOrSet retrieves a value from the cache, or sets it using the provided function if not found
func (p *CachePlugin) GetOrSet(ctx context.Context, key string, value interface{}, ttl time.Duration, fn func() (interface{}, error)) error {
	// Try to get the value from cache
	err := p.Get(ctx, key, value)
	if err == nil {
		return nil
	}

	if err != forge.ErrNotFound {
		return err
	}

	// Value not found, call the function to get it
	newValue, err := fn()
	if err != nil {
		return err
	}

	// Store the new value in cache
	if err := p.Set(ctx, key, newValue, ttl); err != nil {
		return err
	}

	// Copy the new value to the provided value interface
	data, err := json.Marshal(newValue)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, value)
} 