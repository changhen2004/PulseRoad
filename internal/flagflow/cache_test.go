package flagflow

import (
	"context"
	"errors"
	"testing"
	"time"

	"pulseroad/internal/pkg/redis"
)

type fakeRedisStringStore struct {
	values  map[string]string
	deleted []string
	ttl     time.Duration
}

func newFakeRedisStringStore() *fakeRedisStringStore {
	return &fakeRedisStringStore{values: make(map[string]string)}
}

func (s *fakeRedisStringStore) GetString(_ context.Context, key string) (string, error) {
	value, ok := s.values[key]
	if !ok {
		return "", redis.ErrKeyNotFound
	}
	return value, nil
}

func (s *fakeRedisStringStore) SetString(_ context.Context, key string, value string, ttl time.Duration) error {
	s.values[key] = value
	s.ttl = ttl
	return nil
}

func (s *fakeRedisStringStore) DeleteKey(_ context.Context, key string) error {
	s.deleted = append(s.deleted, key)
	delete(s.values, key)
	return nil
}

func TestRedisCacheStoresFeatureFlag(t *testing.T) {
	store := newFakeRedisStringStore()
	cache := NewRedisCache(store, 5*time.Minute)
	flag := FeatureFlag{
		ID:                3,
		ProductID:         10,
		Key:               "new_dashboard",
		Name:              "New Dashboard",
		Environment:       "production",
		Enabled:           true,
		RolloutPercentage: 50,
	}

	if err := cache.Set(context.Background(), flag); err != nil {
		t.Fatalf("set cache: %v", err)
	}
	got, err := cache.Get(context.Background(), 10, "production", "new_dashboard")
	if err != nil {
		t.Fatalf("get cache: %v", err)
	}

	if got.ID != flag.ID || got.Key != flag.Key || !got.Enabled || got.RolloutPercentage != 50 {
		t.Fatalf("unexpected cached flag: %#v", got)
	}
	if store.ttl != 5*time.Minute {
		t.Fatalf("expected ttl 5m, got %s", store.ttl)
	}
}

func TestRedisCacheMapsMissingKeyToCacheMiss(t *testing.T) {
	cache := NewRedisCache(newFakeRedisStringStore(), time.Minute)

	_, err := cache.Get(context.Background(), 10, "production", "missing_flag")
	if !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("expected ErrCacheMiss, got %v", err)
	}
}

func TestRedisCacheDeletesFeatureFlagKey(t *testing.T) {
	store := newFakeRedisStringStore()
	cache := NewRedisCache(store, time.Minute)

	if err := cache.Delete(context.Background(), 10, "production", "new_dashboard"); err != nil {
		t.Fatalf("delete cache: %v", err)
	}

	if len(store.deleted) != 1 || store.deleted[0] != "flagflow:flag:10:production:new_dashboard" {
		t.Fatalf("unexpected deleted keys: %#v", store.deleted)
	}
}
