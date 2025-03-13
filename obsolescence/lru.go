package obsolescence

import "container/list"

// LRUCache Cache is LRU cache. It is not safe for concurrent access.
type LRUCache struct {
	maxBytes int64                    // max memory
	nBytes   int64                    // current memory
	ll       *list.List               // double linked list
	cache    map[string]*list.Element // map

	OnEvicted func(key string, value Value) // optional and executed when an lruEntry is purged.
}

type lruEntry struct {
	key   string
	value Value
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

// NewLRUCache is the Constructor of LRUCache
func NewLRUCache(maxBytes int64, onEvicted func(string, Value)) *LRUCache {
	return &LRUCache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get look ups a key's value
func (c *LRUCache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*lruEntry)
		return kv.value, true
	}
	return
}

// RemoveOldest removes the oldest item
func (c *LRUCache) RemoveOldest() {
	// get the oldest element
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*lruEntry)
		delete(c.cache, kv.key)
		c.computeBytes(false, kv)
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add adds a value to the cache.
func (c *LRUCache) Add(key string, value Value) {
	// if the key exists, update the value and move to the front
	// if the key does not exist, add a new lruEntry to the front
	// if the memory exceeds the maxBytes, remove the oldest lruEntry
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*lruEntry)
		c.computeBytes(true, kv)
		kv.value = value
	} else {
		ele := c.ll.PushFront(&lruEntry{key, value})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.nBytes > c.maxBytes {
		c.RemoveOldest()
	}
}

func (c *LRUCache) computeBytes(isSum bool, kv *lruEntry) {
	if isSum {
		c.nBytes += int64(len(kv.key)) + int64(kv.value.Len())
	}
	c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
}

func (c *LRUCache) Del(key string) {
	if ele, ok := c.cache[key]; ok {
		c.ll.Remove(ele)
		kv := ele.Value.(*lruEntry)
		delete(c.cache, kv.key)
		c.computeBytes(false, kv)
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Len the number of cache entries
func (c *LRUCache) Len() int {
	return c.ll.Len()
}
