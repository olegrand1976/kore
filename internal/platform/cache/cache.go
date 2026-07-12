package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kore/kore/pkg/kernel"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string, dest any) (found bool, err error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	GetOrLoad(ctx context.Context, key string, ttl time.Duration, load func(ctx context.Context) (any, error), dest any) error
}

type KeyBuilder interface {
	Key(tenant kernel.TenantID, module, name string, parts ...string) string
	PublicKey(module, name string, parts ...string) string
}

type keyBuilder struct {
	prefix string
}

func NewKeyBuilder(prefix string) KeyBuilder {
	return &keyBuilder{prefix: strings.TrimSpace(prefix)}
}

func (k *keyBuilder) Key(tenant kernel.TenantID, module, name string, parts ...string) string {
	segments := []string{k.prefix, tenant.String(), module, name}
	segments = append(segments, parts...)
	return strings.Join(segments, ":")
}

func (k *keyBuilder) PublicKey(module, name string, parts ...string) string {
	segments := []string{k.prefix, "public", module, name}
	segments = append(segments, parts...)
	return strings.Join(segments, ":")
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr, password string, useTLS bool) (*RedisCache, error) {
	opts := &redis.Options{
		Addr:     addr,
		Password: password,
	}
	if useTLS {
		opts.TLSConfig = nil // TLS config can be extended for prod
	}
	client := redis.NewClient(opts)
	return &RedisCache{client: client}, nil
}

func (r *RedisCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}

func (r *RedisCache) Get(ctx context.Context, key string, dest any) (bool, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(val, dest); err != nil {
		return false, err
	}
	return true, nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if ttl <= 0 {
		return fmt.Errorf("ttl must be positive")
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisCache) GetOrLoad(ctx context.Context, key string, ttl time.Duration, load func(ctx context.Context) (any, error), dest any) error {
	found, err := r.Get(ctx, key, dest)
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	value, err := load(ctx)
	if err != nil {
		return err
	}
	if err := r.Set(ctx, key, value, ttl); err != nil {
		// best effort cache
		_ = err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

type cacheEntry struct {
	value     []byte
	expiresAt time.Time
}

type InMemoryCache struct {
	mu    sync.RWMutex
	items map[string]cacheEntry
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{items: make(map[string]cacheEntry)}
}

func (m *InMemoryCache) Get(_ context.Context, key string, dest any) (bool, error) {
	m.mu.RLock()
	entry, ok := m.items[key]
	m.mu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		return false, nil
	}
	return true, json.Unmarshal(entry.value, dest)
}

func (m *InMemoryCache) Set(_ context.Context, key string, value any, ttl time.Duration) error {
	if ttl <= 0 {
		return fmt.Errorf("ttl must be positive")
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.items[key] = cacheEntry{value: data, expiresAt: time.Now().Add(ttl)}
	m.mu.Unlock()
	return nil
}

func (m *InMemoryCache) Delete(_ context.Context, keys ...string) error {
	m.mu.Lock()
	for _, key := range keys {
		delete(m.items, key)
	}
	m.mu.Unlock()
	return nil
}

func (m *InMemoryCache) GetOrLoad(ctx context.Context, key string, ttl time.Duration, load func(ctx context.Context) (any, error), dest any) error {
	found, err := m.Get(ctx, key, dest)
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	value, err := load(ctx)
	if err != nil {
		return err
	}
	if err := m.Set(ctx, key, value, ttl); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}
