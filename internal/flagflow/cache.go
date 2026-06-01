package flagflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"pulseroad/internal/pkg/redis"
)

type RedisStringStore interface {
	GetString(ctx context.Context, key string) (string, error)
	SetString(ctx context.Context, key string, value string, ttl time.Duration) error
	DeleteKey(ctx context.Context, key string) error
}

type RedisCache struct {
	store RedisStringStore
	ttl   time.Duration
}

func NewRedisCache(store RedisStringStore, ttl time.Duration) *RedisCache {
	return &RedisCache{store: store, ttl: ttl}
}

func (c *RedisCache) Get(ctx context.Context, productID uint, environment string, key string) (*FeatureFlag, error) {
	value, err := c.store.GetString(ctx, redisFlagKey(productID, environment, key))
	if errors.Is(err, redis.ErrKeyNotFound) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}

	var flag FeatureFlag
	if err := json.Unmarshal([]byte(value), &flag); err != nil {
		return nil, fmt.Errorf("decode cached flag: %w", err)
	}
	return &flag, nil
}

func (c *RedisCache) Set(ctx context.Context, flag FeatureFlag) error {
	body, err := json.Marshal(flag)
	if err != nil {
		return fmt.Errorf("encode cached flag: %w", err)
	}
	return c.store.SetString(ctx, redisFlagKey(flag.ProductID, flag.Environment, flag.Key), string(body), c.ttl)
}

func (c *RedisCache) Delete(ctx context.Context, productID uint, environment string, key string) error {
	return c.store.DeleteKey(ctx, redisFlagKey(productID, environment, key))
}

func redisFlagKey(productID uint, environment string, key string) string {
	return fmt.Sprintf("flagflow:flag:%d:%s:%s", productID, environment, key)
}
