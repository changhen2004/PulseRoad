package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	redisclient "github.com/redis/go-redis/v9"

	"pulseroad/internal/pkg/config"
)

var ErrKeyNotFound = errors.New("redis key not found")

type Client struct {
	rdb *redisclient.Client
}

func Init(cfg *config.RedisConfig) (*Client, error) {
	rdb := redisclient.NewClient(&redisclient.Options{Addr: cfg.Addr})
	client := &Client{rdb: rdb}
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("connect redis: %w", err)
	}
	return client, nil
}

func (c *Client) Close() error {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.Close()
}

func (c *Client) Get(ctx context.Context, key string) (int64, error) {
	value, err := c.rdb.Get(ctx, key).Int64()
	if errors.Is(err, redisclient.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (c *Client) GetString(ctx context.Context, key string) (string, error) {
	value, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, redisclient.Nil) {
		return "", ErrKeyNotFound
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

func (c *Client) SetString(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *Client) DeleteKey(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}

func (c *Client) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	count, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		if err := c.rdb.Expire(ctx, key, ttl).Err(); err != nil {
			return 0, err
		}
	}
	return count, nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}
