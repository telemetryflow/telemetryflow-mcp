// Package cache provides Redis-based caching for the MCP server.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Common errors
var (
	ErrCacheMiss       = errors.New("cache miss")
	ErrCacheDisabled   = errors.New("cache is disabled")
	ErrInvalidKey      = errors.New("invalid cache key")
	ErrSerializeFailed = errors.New("failed to serialize value")
)

// RedisCache provides Redis-based caching functionality.
type RedisCache struct {
	client      *redis.Client
	prefix      string
	defaultTTL  time.Duration
	enabled     bool
	mu          sync.RWMutex
	initialized bool
}

// RedisCacheConfig configures the Redis cache.
type RedisCacheConfig struct {
	// URL is the Redis connection URL (redis://host:port)
	URL string `mapstructure:"url" yaml:"url" json:"url"`
	// Host is the Redis host
	Host string `mapstructure:"host" yaml:"host" json:"host"`
	// Port is the Redis port
	Port int `mapstructure:"port" yaml:"port" json:"port"`
	// Password is the Redis password
	Password string `mapstructure:"password" yaml:"password" json:"password"`
	// DB is the Redis database number
	DB int `mapstructure:"db" yaml:"db" json:"db"`
	// Prefix is the key prefix for all cache keys
	Prefix string `mapstructure:"prefix" yaml:"prefix" json:"prefix"`
	// DefaultTTL is the default TTL for cache entries
	DefaultTTL time.Duration `mapstructure:"default_ttl" yaml:"default_ttl" json:"default_ttl"`
	// Enabled enables or disables the cache
	Enabled bool `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	// PoolSize is the connection pool size
	PoolSize int `mapstructure:"pool_size" yaml:"pool_size" json:"pool_size"`
	// MinIdleConns is the minimum number of idle connections
	MinIdleConns int `mapstructure:"min_idle_conns" yaml:"min_idle_conns" json:"min_idle_conns"`
	// DialTimeout is the connection dial timeout
	DialTimeout time.Duration `mapstructure:"dial_timeout" yaml:"dial_timeout" json:"dial_timeout"`
	// ReadTimeout is the read timeout
	ReadTimeout time.Duration `mapstructure:"read_timeout" yaml:"read_timeout" json:"read_timeout"`
	// WriteTimeout is the write timeout
	WriteTimeout time.Duration `mapstructure:"write_timeout" yaml:"write_timeout" json:"write_timeout"`
}

// DefaultRedisCacheConfig returns default configuration.
func DefaultRedisCacheConfig() *RedisCacheConfig {
	return &RedisCacheConfig{
		Host:         "localhost",
		Port:         6379,
		DB:           0,
		Prefix:       "tfo-mcp:",
		DefaultTTL:   5 * time.Minute,
		Enabled:      true,
		PoolSize:     10,
		MinIdleConns: 2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// NewRedisCache creates a new Redis cache.
func NewRedisCache(cfg *RedisCacheConfig) (*RedisCache, error) {
	if cfg == nil {
		cfg = DefaultRedisCacheConfig()
	}

	if !cfg.Enabled {
		return &RedisCache{
			enabled: false,
		}, nil
	}

	opts := &redis.Options{
		DB:           cfg.DB,
		Password:     cfg.Password,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	// Use URL if provided, otherwise use host:port
	if cfg.URL != "" {
		parsedOpts, err := redis.ParseURL(cfg.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
		}
		opts = parsedOpts
		opts.PoolSize = cfg.PoolSize
		opts.MinIdleConns = cfg.MinIdleConns
	} else {
		opts.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	}

	client := redis.NewClient(opts)

	return &RedisCache{
		client:     client,
		prefix:     cfg.Prefix,
		defaultTTL: cfg.DefaultTTL,
		enabled:    true,
	}, nil
}

// Initialize initializes the cache connection.
func (c *RedisCache) Initialize(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.enabled {
		return nil
	}

	if c.initialized {
		return nil
	}

	// Ping to verify connection
	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	c.initialized = true
	return nil
}

// Close closes the cache connection.
func (c *RedisCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.enabled || c.client == nil {
		return nil
	}

	c.initialized = false
	return c.client.Close()
}

// Get retrieves a value from the cache.
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	if !c.isReady() {
		return nil, ErrCacheDisabled
	}

	if key == "" {
		return nil, ErrInvalidKey
	}

	val, err := c.client.Get(ctx, c.prefix+key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}
		return nil, err
	}

	return val, nil
}

// GetJSON retrieves and unmarshals a JSON value from the cache.
func (c *RedisCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// Set stores a value in the cache with the default TTL.
func (c *RedisCache) Set(ctx context.Context, key string, value []byte) error {
	return c.SetWithTTL(ctx, key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with a custom TTL.
func (c *RedisCache) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if !c.isReady() {
		return ErrCacheDisabled
	}

	if key == "" {
		return ErrInvalidKey
	}

	return c.client.Set(ctx, c.prefix+key, value, ttl).Err()
}

// SetJSON marshals and stores a value in the cache.
func (c *RedisCache) SetJSON(ctx context.Context, key string, value interface{}) error {
	return c.SetJSONWithTTL(ctx, key, value, c.defaultTTL)
}

// SetJSONWithTTL marshals and stores a value with a custom TTL.
func (c *RedisCache) SetJSONWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSerializeFailed, err)
	}

	return c.SetWithTTL(ctx, key, data, ttl)
}

// Delete removes a value from the cache.
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if !c.isReady() {
		return ErrCacheDisabled
	}

	if key == "" {
		return ErrInvalidKey
	}

	return c.client.Del(ctx, c.prefix+key).Err()
}

// DeletePattern removes all keys matching a pattern.
func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	if !c.isReady() {
		return ErrCacheDisabled
	}

	iter := c.client.Scan(ctx, 0, c.prefix+pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}

// Exists checks if a key exists in the cache.
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	if !c.isReady() {
		return false, ErrCacheDisabled
	}

	if key == "" {
		return false, ErrInvalidKey
	}

	n, err := c.client.Exists(ctx, c.prefix+key).Result()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}

// TTL returns the remaining TTL for a key.
func (c *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	if !c.isReady() {
		return 0, ErrCacheDisabled
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.TTL(ctx, c.prefix+key).Result()
}

// Expire sets a new TTL for an existing key.
func (c *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if !c.isReady() {
		return ErrCacheDisabled
	}

	if key == "" {
		return ErrInvalidKey
	}

	return c.client.Expire(ctx, c.prefix+key, ttl).Err()
}

// GetOrSet retrieves a value from cache or sets it using the provided function.
func (c *RedisCache) GetOrSet(ctx context.Context, key string, ttl time.Duration, fn func() ([]byte, error)) ([]byte, error) {
	// Try to get from cache first
	val, err := c.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	if !errors.Is(err, ErrCacheMiss) && !errors.Is(err, ErrCacheDisabled) {
		return nil, err
	}

	// Generate value
	val, err = fn()
	if err != nil {
		return nil, err
	}

	// Store in cache (ignore errors)
	_ = c.SetWithTTL(ctx, key, val, ttl)

	return val, nil
}

// GetOrSetJSON retrieves a JSON value from cache or sets it using the provided function.
func (c *RedisCache) GetOrSetJSON(ctx context.Context, key string, ttl time.Duration, dest interface{}, fn func() (interface{}, error)) error {
	// Try to get from cache first
	err := c.GetJSON(ctx, key, dest)
	if err == nil {
		return nil
	}

	if !errors.Is(err, ErrCacheMiss) && !errors.Is(err, ErrCacheDisabled) {
		return err
	}

	// Generate value
	val, err := fn()
	if err != nil {
		return err
	}

	// Store in cache (ignore errors)
	_ = c.SetJSONWithTTL(ctx, key, val, ttl)

	// Assign to dest
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// Increment increments an integer value.
func (c *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	if !c.isReady() {
		return 0, ErrCacheDisabled
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.Incr(ctx, c.prefix+key).Result()
}

// IncrementBy increments an integer value by a given amount.
func (c *RedisCache) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	if !c.isReady() {
		return 0, ErrCacheDisabled
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.IncrBy(ctx, c.prefix+key, value).Result()
}

// Decrement decrements an integer value.
func (c *RedisCache) Decrement(ctx context.Context, key string) (int64, error) {
	if !c.isReady() {
		return 0, ErrCacheDisabled
	}

	if key == "" {
		return 0, ErrInvalidKey
	}

	return c.client.Decr(ctx, c.prefix+key).Result()
}

// Client returns the underlying Redis client for advanced operations.
func (c *RedisCache) Client() *redis.Client {
	return c.client
}

// IsEnabled returns whether the cache is enabled.
func (c *RedisCache) IsEnabled() bool {
	return c.enabled
}

// IsReady returns whether the cache is ready for use.
func (c *RedisCache) IsReady() bool {
	return c.isReady()
}

func (c *RedisCache) isReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.enabled && c.initialized
}

// Stats returns cache statistics.
func (c *RedisCache) Stats(ctx context.Context) (map[string]interface{}, error) {
	if !c.isReady() {
		return nil, ErrCacheDisabled
	}

	info, err := c.client.Info(ctx, "memory", "stats", "keyspace").Result()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"info": info,
	}, nil
}
