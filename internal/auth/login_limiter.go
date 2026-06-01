package auth

import (
	"context"
	"fmt"
	"time"
)

const loginFailureKeyPrefix = "auth:login_fail:"

type LoginFailureCounter interface {
	Get(ctx context.Context, key string) (int64, error)
	Increment(ctx context.Context, key string, ttl time.Duration) (int64, error)
	Delete(ctx context.Context, key string) error
}

type RedisLoginLimiter struct {
	counter LoginFailureCounter
	max     int64
	window  time.Duration
}

func NewRedisLoginLimiter(counter LoginFailureCounter, max int64, window time.Duration) *RedisLoginLimiter {
	return &RedisLoginLimiter{counter: counter, max: max, window: window}
}

func (l *RedisLoginLimiter) Check(ctx context.Context, email string) error {
	failures, err := l.counter.Get(ctx, l.key(email))
	if err != nil {
		return fmt.Errorf("check login failures: %w", err)
	}
	if failures >= l.max {
		return ErrTooManyLoginAttempts
	}
	return nil
}

func (l *RedisLoginLimiter) RecordFailure(ctx context.Context, email string) error {
	if _, err := l.counter.Increment(ctx, l.key(email), l.window); err != nil {
		return fmt.Errorf("record login failure: %w", err)
	}
	return nil
}

func (l *RedisLoginLimiter) Reset(ctx context.Context, email string) error {
	if err := l.counter.Delete(ctx, l.key(email)); err != nil {
		return fmt.Errorf("reset login failures: %w", err)
	}
	return nil
}

func (l *RedisLoginLimiter) key(email string) string {
	return loginFailureKeyPrefix + email
}
