package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeLoginFailureCounter struct {
	values     map[string]int64
	lastKey    string
	lastTTL    time.Duration
	deletedKey string
	err        error
}

func newFakeLoginFailureCounter() *fakeLoginFailureCounter {
	return &fakeLoginFailureCounter{values: make(map[string]int64)}
}

func (c *fakeLoginFailureCounter) Get(ctx context.Context, key string) (int64, error) {
	if c.err != nil {
		return 0, c.err
	}
	return c.values[key], nil
}

func (c *fakeLoginFailureCounter) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	if c.err != nil {
		return 0, c.err
	}
	c.lastKey = key
	c.lastTTL = ttl
	c.values[key]++
	return c.values[key], nil
}

func (c *fakeLoginFailureCounter) Delete(ctx context.Context, key string) error {
	if c.err != nil {
		return c.err
	}
	c.deletedKey = key
	delete(c.values, key)
	return nil
}

func TestRedisLoginLimiterBlocksWhenFailureLimitReached(t *testing.T) {
	counter := newFakeLoginFailureCounter()
	counter.values["auth:login_fail:ada@example.com"] = 5
	limiter := NewRedisLoginLimiter(counter, 5, 15*time.Minute)

	err := limiter.Check(context.Background(), "ada@example.com")
	if !errors.Is(err, ErrTooManyLoginAttempts) {
		t.Fatalf("expected ErrTooManyLoginAttempts, got %v", err)
	}
}

func TestRedisLoginLimiterRecordsFailureWithTTL(t *testing.T) {
	counter := newFakeLoginFailureCounter()
	limiter := NewRedisLoginLimiter(counter, 5, 15*time.Minute)

	if err := limiter.RecordFailure(context.Background(), "ada@example.com"); err != nil {
		t.Fatalf("record failure: %v", err)
	}

	if counter.values["auth:login_fail:ada@example.com"] != 1 {
		t.Fatalf("expected one failure, got %#v", counter.values)
	}
	if counter.lastTTL != 15*time.Minute {
		t.Fatalf("expected ttl 15m, got %s", counter.lastTTL)
	}
}

func TestRedisLoginLimiterResetClearsFailures(t *testing.T) {
	counter := newFakeLoginFailureCounter()
	counter.values["auth:login_fail:ada@example.com"] = 3
	limiter := NewRedisLoginLimiter(counter, 5, 15*time.Minute)

	if err := limiter.Reset(context.Background(), "ada@example.com"); err != nil {
		t.Fatalf("reset: %v", err)
	}

	if counter.deletedKey != "auth:login_fail:ada@example.com" {
		t.Fatalf("expected delete key, got %q", counter.deletedKey)
	}
	if counter.values["auth:login_fail:ada@example.com"] != 0 {
		t.Fatalf("expected failures cleared, got %#v", counter.values)
	}
}
