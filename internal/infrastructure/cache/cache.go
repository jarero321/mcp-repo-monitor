package cache

import (
	"sync"
	"time"
)

// Entry represents a cached item with expiration.
type Entry struct {
	Value     any
	ExpiresAt time.Time
}

// IsExpired checks if the entry has expired.
func (e *Entry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Cache is a thread-safe in-memory cache with TTL support.
type Cache struct {
	mu          sync.RWMutex
	entries     map[string]*Entry
	defaultTTL  time.Duration
	cleanupTick time.Duration
	stopCleanup chan struct{}
}

// Config holds cache configuration.
type Config struct {
	DefaultTTL  time.Duration // Default time-to-live for entries
	CleanupTick time.Duration // How often to run cleanup
}

// DefaultConfig returns sensible cache defaults.
func DefaultConfig() Config {
	return Config{
		DefaultTTL:  5 * time.Minute,
		CleanupTick: 1 * time.Minute,
	}
}

// New creates a new cache with the given configuration.
func New(cfg Config) *Cache {
	c := &Cache{
		entries:     make(map[string]*Entry),
		defaultTTL:  cfg.DefaultTTL,
		cleanupTick: cfg.CleanupTick,
		stopCleanup: make(chan struct{}),
	}

	go c.startCleanup()

	return c
}

// startCleanup periodically removes expired entries.
func (c *Cache) startCleanup() {
	ticker := time.NewTicker(c.cleanupTick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanup removes all expired entries.
func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.entries {
		if entry.IsExpired() {
			delete(c.entries, key)
		}
	}
}

// Stop stops the cleanup goroutine.
func (c *Cache) Stop() {
	close(c.stopCleanup)
}

// Set stores a value with the default TTL.
func (c *Cache) Set(key string, value any) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value with a custom TTL.
func (c *Cache) SetWithTTL(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &Entry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Get retrieves a value from the cache.
// Returns the value and true if found and not expired, nil and false otherwise.
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}

	if entry.IsExpired() {
		return nil, false
	}

	return entry.Value, true
}

// Delete removes an entry from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// Clear removes all entries from the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*Entry)
}

// Size returns the number of entries in the cache.
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

// Keys returns all non-expired keys in the cache.
func (c *Cache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.entries))
	for key, entry := range c.entries {
		if !entry.IsExpired() {
			keys = append(keys, key)
		}
	}

	return keys
}
