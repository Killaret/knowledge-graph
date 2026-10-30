package cachetest

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"knowledge-graph/internal/domain/cache"
)

// FakeCacheClient is an in-memory cache.CacheClient for unit tests.
type FakeCacheClient struct {
	mu   sync.RWMutex
	data map[string]cacheEntry
}

type cacheEntry struct {
	value string
	exp   time.Time
}

func NewFakeCacheClient() *FakeCacheClient {
	return &FakeCacheClient{data: make(map[string]cacheEntry)}
}

func (f *FakeCacheClient) Get(_ context.Context, key string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	entry, ok := f.data[key]
	if !ok || (!entry.exp.IsZero() && time.Now().After(entry.exp)) {
		return "", cache.ErrCacheMiss
	}
	return entry.value, nil
}

func (f *FakeCacheClient) Set(_ context.Context, key, value string, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	exp := time.Time{}
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	f.data[key] = cacheEntry{value: value, exp: exp}
	return nil
}

func (f *FakeCacheClient) Del(_ context.Context, keys ...string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, key := range keys {
		delete(f.data, key)
	}
	return nil
}

func (f *FakeCacheClient) Exists(_ context.Context, key string) (int64, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, ok := f.data[key]
	if ok {
		return 1, nil
	}
	return 0, nil
}

// TTL returns the remaining time-to-live for a key.
// It is not part of cache.CacheClient but useful for tests.
func (f *FakeCacheClient) TTL(_ context.Context, key string) (time.Duration, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	entry, ok := f.data[key]
	if !ok {
		return -1, nil
	}
	if entry.exp.IsZero() {
		return -1, nil
	}
	return time.Until(entry.exp), nil
}

func (f *FakeCacheClient) Expire(_ context.Context, key string, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	entry, ok := f.data[key]
	if !ok {
		return cache.ErrCacheMiss
	}
	entry.exp = time.Now().Add(ttl)
	f.data[key] = entry
	return nil
}

func (f *FakeCacheClient) Scan(_ context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var keys []string
	i := uint64(0)
	started := cursor == 0
	for key := range f.data {
		if started {
			if match == "" || true /* simple glob not implemented, return all */ {
				keys = append(keys, key)
				if int64(len(keys)) >= count {
					return keys, i + 1, nil
				}
			}
		}
		i++
		if i == cursor {
			started = true
		}
	}
	return keys, 0, nil
}

func (f *FakeCacheClient) Incr(_ context.Context, key string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	entry, ok := f.data[key]
	if !ok {
		f.data[key] = cacheEntry{value: "1"}
		return 1, nil
	}
	n, err := strconv.Atoi(entry.value)
	if err != nil {
		return 0, fmt.Errorf("cannot increment non-numeric value %q: %w", entry.value, err)
	}
	n++
	f.data[key] = cacheEntry{value: strconv.Itoa(n)}
	return int64(n), nil
}
