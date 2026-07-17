package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache is a nil-safe Redis-backed cache used for content-addressed caching
// across VLM, embedding, and wiki map layers. When rdb is nil (Lite mode)
// all operations are no-ops that report miss / no-op silently.
type Cache struct {
	rdb *redis.Client
}

// NewCache creates a new Cache. rdb may be nil (Lite mode).
func NewCache(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

func (c *Cache) Available() bool { return c.rdb != nil }

// Get returns the cached string value and true on hit.
func (c *Cache) Get(ctx context.Context, key string) (string, bool) {
	if c.rdb == nil {
		return "", false
	}
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", false
	}
	return val, true
}

// Set stores a string value. A zero ttl means no expiry.
func (c *Cache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if c.rdb == nil {
		return nil
	}
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

// GetJSON deserialises the cached JSON value into dest. Returns true on hit.
func (c *Cache) GetJSON(ctx context.Context, key string, dest any) (bool, error) {
	if c.rdb == nil {
		return false, nil
	}
	raw, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return false, nil
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return false, fmt.Errorf("cache deserialise %s: %w", key, err)
	}
	return true, nil
}

// SetJSON serialises val as JSON and stores it. Zero ttl means no expiry.
func (c *Cache) SetJSON(ctx context.Context, key string, val any, ttl time.Duration) error {
	if c.rdb == nil {
		return nil
	}
	raw, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("cache serialise %s: %w", key, err)
	}
	return c.rdb.Set(ctx, key, raw, ttl).Err()
}

// Del removes one or more keys.
func (c *Cache) Del(ctx context.Context, keys ...string) error {
	if c.rdb == nil || len(keys) == 0 {
		return nil
	}
	return c.rdb.Del(ctx, keys...).Err()
}

// DelByPattern removes all keys matching a glob pattern.
func (c *Cache) DelByPattern(ctx context.Context, pattern string) error {
	if c.rdb == nil {
		return nil
	}
	iter := c.rdb.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.rdb.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// Close closes the underlying redis connection.
func (c *Cache) Close() error {
	if c.rdb == nil {
		return nil
	}
	return c.rdb.Close()
}

// ---------------------------------------------------------------------------
// Cache key builders
// ---------------------------------------------------------------------------

const (
	vlmCachePrefix       = "vlm:"
	embeddingCachePrefix = "emb:"
	wikiMapCachePrefix   = "wiki:map:"
)

// VLMCacheKey builds the Redis key for a VLM result.
// Format: vlm:{imgHexHash}:{modelID}:{promptSHA}
func VLMCacheKey(imgHash, modelID, promptHash string) string {
	return vlmCachePrefix + imgHash + ":" + modelID + ":" + promptHash
}

// EmbeddingCacheKey builds the Redis key for an embedding vector.
// Format: emb:{contentSHA}:{modelID}:{dim}
func EmbeddingCacheKey(contentHash, modelID string, dims int) string {
	return embeddingCachePrefix + contentHash + ":" + modelID + ":" + strconv.Itoa(dims)
}

// WikiMapCacheKey builds the Redis key for a wiki per-doc map result.
// Format: wiki:map:{docContentSHA}:{granularity}:{modelID}:{promptSHA}
func WikiMapCacheKey(contentHash, granularity, modelID, promptHash string) string {
	return wikiMapCachePrefix + contentHash + ":" + granularity + ":" + modelID + ":" + promptHash
}
