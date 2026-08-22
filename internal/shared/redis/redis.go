package redis

import (
    "context"
    "fmt"
    "log"
    "strings"
    "time"

    "github.com/redis/go-redis/v9"
    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
)

// ParseURL is a wrapper around redis.ParseURL
func ParseURL(url string) (*redis.Options, error) {
    return redis.ParseURL(url)
}

// Client wraps the Redis client with context support
type Client struct {
    client *redis.Client
}

// NewClient creates a new Redis client
func NewClient(cfg *config.Config) (*Client, error) {
    redisURL := cfg.GetRedisURL()

    // Ensure scheme is present for go-redis URL parsing
    if !strings.HasPrefix(redisURL, "redis://") && !strings.HasPrefix(redisURL, "rediss://") {
        redisURL = "redis://" + redisURL
    }

    opts, err := redis.ParseURL(redisURL)
    if err != nil {
        return nil, fmt.Errorf("failed to parse Redis URL (%s): %w", redisURL, err)
    }

    client := redis.NewClient(opts)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := client.Ping(ctx).Err(); err != nil {
        return nil, err
    }

    log.Println("✅ Redis connection established successfully")
    return &Client{client: client}, nil
}

// GetClient returns the underlying Redis client
func (c *Client) GetClient() *redis.Client {
    return c.client
}

// Close closes the Redis connection
func (c *Client) Close() error {
    if c.client != nil {
        return c.client.Close()
    }
    return nil
}

// ============================================================
// REDIS OPERATIONS WITH CONTEXT
// ============================================================

// Set stores a value with TTL
func (c *Client) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
    return c.client.Set(ctx, key, value, ttl).Err()
}

// Get retrieves a value
func (c *Client) Get(ctx context.Context, key string) (string, bool, error) {
    val, err := c.client.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            return "", false, nil
        }
        return "", false, err
    }
    return val, true, nil
}

// Delete removes a key
func (c *Client) Delete(ctx context.Context, key string) error {
    return c.client.Del(ctx, key).Err()
}

// HSet stores a hash
func (c *Client) HSet(ctx context.Context, key string, values map[string]interface{}) error {
    return c.client.HSet(ctx, key, values).Err()
}

// HGetAll retrieves all fields from a hash
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, bool, error) {
    result, err := c.client.HGetAll(ctx, key).Result()
    if err != nil {
        return nil, false, err
    }
    if len(result) == 0 {
        return nil, false, nil
    }
    return result, true, nil
}

// Expire sets TTL on a key
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
    return c.client.Expire(ctx, key, ttl).Err()
}

// HGet gets a specific field from a hash
func (c *Client) HGet(ctx context.Context, key, field string) (string, bool, error) {
    val, err := c.client.HGet(ctx, key, field).Result()
    if err != nil {
        if err == redis.Nil {
            return "", false, nil
        }
        return "", false, err
    }
    return val, true, nil
}

// HSetField sets a single field in a hash
func (c *Client) HSetField(ctx context.Context, key, field string, value interface{}) error {
    return c.client.HSet(ctx, key, field, value).Err()
}