package obsolescence

import "container/list"

// LFUCache Cache is LRU cache. It is not safe for concurrent access.
type LFUCache struct {
	maxBytes int64                    // max memory
	nBytes   int64                    // current memory
	ll       *list.List               // double linked list
	cache    map[string]*list.Element // map

	OnEvicted func(key string, value Value) // optional and executed when an lruEntry is purged.
}

type lfuEntry struct {
	key   string
	value Value
	freq  int
}

// NewLFUCache is the Constructor of NewLFUCache
func NewLFUCache(maxBytes int64, onEvicted func(string, Value)) *LFUCache {
	return &LFUCache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *LFUCache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*lfuEntry)
		kv.freq++
		return kv.value, true
	}
	return
}

func (c *LFUCache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		kv := ele.Value.(*lfuEntry)
		kv.value = value
		c.ll.MoveToFront(ele)
		kv.freq++
		c.nBytes += int64(value.Len())
		c.computeBytes(true, kv)
	} else {
		kv := &lfuEntry{key, value, 1}
		ele := c.ll.PushFront(kv)
		c.cache[key] = ele
		c.nBytes += int64(len(key) + value.Len())
		c.computeBytes(true, kv)
	}
}

func (c *LFUCache) computeBytes(isSum bool, kv *lfuEntry) {
	if isSum {
		c.nBytes += int64(len(kv.key)) + int64(kv.value.Len())
	}
	c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
}

func (c *LFUCache) Del(key string) {
	if ele, ok := c.cache[key]; ok {
		c.ll.Remove(ele)
		kv := ele.Value.(*lfuEntry)
		delete(c.cache, kv.key)
		c.computeBytes(false, kv)
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *LFUCache) RemoveOldest() {
	c.Del(c.ll.Back().Value.(*lfuEntry).key)
}

func (c *LFUCache) Len() int {
	return c.ll.Len()

}
