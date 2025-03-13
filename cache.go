package GoDistributedCache

import (
	"GoDistributedCache/obsolescence"
	"sync"
)

type cache struct {
	mu         sync.Mutex
	lru        *obsolescence.LRUCache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	// lazy initialization
	if c.lru == nil {
		c.lru = obsolescence.NewLRUCache(c.cacheBytes, nil)
	}
	c.mu.Lock()
	c.lru.Add(key, value)
	c.mu.Unlock()
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	if c.lru == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
