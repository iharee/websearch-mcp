package cache

import (
	"container/list"
	"sync"
	"time"
)

type entry[V any] struct {
	key       string
	value     V
	expiresAt time.Time
	ele       *list.Element
}

// Cache is a generic LRU cache with TTL expiration. Safe for concurrent use.
type Cache[V any] struct {
	mu      sync.RWMutex
	m       map[string]*list.Element
	l       *list.List
	maxSize int
	ttl     time.Duration
}

// New creates a cache with the given max entries and TTL. Zero or negative
// maxSize means unlimited; zero or negative ttl means no expiration.
func New[V any](maxSize int, ttl time.Duration) *Cache[V] {
	if maxSize <= 0 {
		maxSize = 0 // sentinel for unlimited
	}
	return &Cache[V]{
		m:       make(map[string]*list.Element),
		l:       list.New(),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// Get returns the cached value and true if found and not expired. Expired
// entries are removed on access. Returns the zero value and false on miss.
func (c *Cache[V]) Get(key string) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ele, ok := c.m[key]
	if !ok {
		var zero V
		return zero, false
	}

	e := ele.Value.(*entry[V])

	if c.ttl > 0 && time.Now().After(e.expiresAt) {
		c.l.Remove(ele)
		delete(c.m, key)
		var zero V
		return zero, false
	}

	c.l.MoveToFront(ele)
	return e.value, true
}

// Put inserts or updates a key. If the cache is at capacity, the least
// recently used entry is evicted.
func (c *Cache[V]) Put(key string, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.m[key]; ok {
		e := ele.Value.(*entry[V])
		e.value = value
		e.expiresAt = time.Now().Add(c.ttl)
		c.l.MoveToFront(ele)
		return
	}

	if c.maxSize > 0 && c.l.Len() >= c.maxSize {
		c.evictOldest()
	}

	e := &entry[V]{
		key:       key,
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
	e.ele = c.l.PushFront(e)
	c.m[key] = e.ele
}

// Remove deletes a key from the cache.
func (c *Cache[V]) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.m[key]; ok {
		c.l.Remove(ele)
		delete(c.m, key)
	}
}

// Len returns the current number of cached entries.
func (c *Cache[V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.l.Len()
}

func (c *Cache[V]) evictOldest() {
	ele := c.l.Back()
	if ele == nil {
		return
	}
	e := ele.Value.(*entry[V])
	c.l.Remove(ele)
	delete(c.m, e.key)
}
