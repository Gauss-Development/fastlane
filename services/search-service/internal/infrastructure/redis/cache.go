// Package redis is the search service's JSON cache for the expensive AI steps
// (spec extraction, query embedding, match explanations). Mirrors the
// api-gateway redis client (go-redis/v8).
package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
	client *redis.Client
}

func NewCache(addr, password string, db int) *Cache {
	return &Cache{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}

// GetJSON unmarshals the value at key into dst. The bool reports a cache hit;
// a miss (or unmarshal failure) returns (false, nil) so callers transparently
// fall through to recomputation.
func (c *Cache) GetJSON(ctx context.Context, key string, dst any) (bool, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal([]byte(val), dst); err != nil {
		return false, nil
	}
	return true, nil
}

func (c *Cache) SetJSON(ctx context.Context, key string, v any, ttl time.Duration) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, raw, ttl).Err()
}

func (c *Cache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Cache) Close() error {
	return c.client.Close()
}
