package obsolescence

import (
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

// Helper function to test Get and cache miss
func testCacheGet(t *testing.T, cache Cache) {
	cache.Add("key1", String("1234"))
	if v, ok := cache.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := cache.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

// Helper function to test removeOldest functionality
func testCacheRemoveOldest(t *testing.T, cache Cache) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"
	cap := len(k1 + k2 + v1 + v2)
	cache = NewLRUCache(int64(cap), nil)
	cache.Add(k1, String(v1))
	cache.Add(k2, String(v2))
	cache.Add(k3, String(v3))

	// Verify that the oldest entry (key1) was removed
	if _, ok := cache.Get("key1"); ok || cache.Len() != 2 {
		t.Fatalf("cache eviction failed")
	}
}

// -------------------- LRU 测试 --------------------

func TestLRUCache(t *testing.T) {
	lru := NewLRUCache(int64(0), nil)

	// Test Get
	t.Run("LRUGet", func(t *testing.T) {
		testCacheGet(t, lru)
	})

	// Test removeOldest
	t.Run("LRURemoveOldest", func(t *testing.T) {
		testCacheRemoveOldest(t, lru)
	})
}

// -------------------- LFU 测试 --------------------

func TestLFUCache(t *testing.T) {
	lfu := NewLFUCache(int64(0), nil)

	// Test Get
	t.Run("LFUGet", func(t *testing.T) {
		testCacheGet(t, lfu)
	})

	// Test removeOldest
	t.Run("LFURemoveOldest", func(t *testing.T) {
		testCacheRemoveOldest(t, lfu)
	})

}

// -------------------- FIFO 测试 --------------------

func TestFIFOCache(t *testing.T) {
	fifo := NewFIFOCache(int64(0), nil)

	// Test Get
	t.Run("FIFOGet", func(t *testing.T) {
		testCacheGet(t, fifo)
	})

	// Test removeOldest
	t.Run("FIFORemoveOldest", func(t *testing.T) {
		testCacheRemoveOldest(t, fifo)
	})
}
