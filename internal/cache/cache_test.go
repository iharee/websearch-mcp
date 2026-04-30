package cache

import (
	"testing"
	"time"
)

func TestCacheGetPut(t *testing.T) {
	c := New[string](10, time.Minute)

	_, ok := c.Get("a")
	if ok {
		t.Error("expected miss for non-existent key")
	}

	c.Put("a", "hello")
	val, ok := c.Get("a")
	if !ok {
		t.Error("expected hit after Put")
	}
	if val != "hello" {
		t.Errorf("got %q, want %q", val, "hello")
	}
}

func TestCacheTTLExpiry(t *testing.T) {
	c := New[string](10, 10*time.Millisecond)

	c.Put("a", "hello")
	time.Sleep(20 * time.Millisecond)

	_, ok := c.Get("a")
	if ok {
		t.Error("expected expiry after TTL")
	}
}

func TestCacheNoTTL(t *testing.T) {
	c := New[string](10, 0)

	c.Put("a", "hello")
	val, ok := c.Get("a")
	if !ok {
		t.Error("expected hit with no TTL")
	}
	if val != "hello" {
		t.Errorf("got %q, want %q", val, "hello")
	}
}

func TestCacheLRUEviction(t *testing.T) {
	c := New[string](2, time.Minute)

	c.Put("a", "1")
	c.Put("b", "2")
	// a is LRU, b is newest
	c.Put("c", "3") // evicts a

	if _, ok := c.Get("a"); ok {
		t.Error("a should have been evicted")
	}
	if v, ok := c.Get("b"); !ok || v != "2" {
		t.Error("b should still be present")
	}
	if v, ok := c.Get("c"); !ok || v != "3" {
		t.Error("c should be present")
	}
}

func TestCacheLRUAccessBumps(t *testing.T) {
	c := New[string](2, time.Minute)

	c.Put("a", "1")
	c.Put("b", "2")
	c.Get("a")   // a is now most recent, b is LRU
	c.Put("c", "3") // evicts b

	if v, ok := c.Get("a"); !ok || v != "1" {
		t.Error("a should remain after eviction")
	}
	if _, ok := c.Get("b"); ok {
		t.Error("b should have been evicted")
	}
	if v, ok := c.Get("c"); !ok || v != "3" {
		t.Error("c should be present")
	}
}

func TestCacheUpdateExisting(t *testing.T) {
	c := New[string](10, time.Minute)

	c.Put("a", "hello")
	c.Put("a", "world")

	val, _ := c.Get("a")
	if val != "world" {
		t.Errorf("got %q, want %q", val, "world")
	}
	if c.Len() != 1 {
		t.Errorf("Len = %d, want 1", c.Len())
	}
}

func TestCacheRemove(t *testing.T) {
	c := New[string](10, time.Minute)

	c.Put("a", "hello")
	c.Remove("a")

	if _, ok := c.Get("a"); ok {
		t.Error("expected miss after Remove")
	}
	if c.Len() != 0 {
		t.Errorf("Len = %d, want 0", c.Len())
	}
}

func TestCacheLen(t *testing.T) {
	c := New[string](10, time.Minute)

	if c.Len() != 0 {
		t.Errorf("Len = %d, want 0", c.Len())
	}

	c.Put("a", "1")
	c.Put("b", "2")

	if c.Len() != 2 {
		t.Errorf("Len = %d, want 2", c.Len())
	}
}

func TestCacheUnlimitedSize(t *testing.T) {
	c := New[string](0, time.Minute)

	for i := 0; i < 1000; i++ {
		c.Put("key"+string(rune(i)), "value")
	}
	if c.Len() != 1000 {
		t.Errorf("unlimited cache Len = %d, want 1000", c.Len())
	}
}
