package obsolescence

import "container/list"

type FIFOCache struct {
	maxBytes int64                    // max memory
	nBytes   int64                    // current memory
	ll       *list.List               // double linked list
	cache    map[string]*list.Element // map

	OnEvicted func(key string, value Value) // optional and executed when an lruEntry is purged.
}

func NewFIFOCache(maxBytes int64, onEvicted func(string, Value)) *FIFOCache {
	return &FIFOCache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

type fifoEntry struct {
	key   string
	value Value
}

func (c *FIFOCache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		kv := ele.Value.(*fifoEntry)
		return kv.value, true
	}
	return
}

func (c *FIFOCache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		kv := ele.Value.(*fifoEntry)
		kv.value = value
		c.ll.MoveToFront(ele)
		c.computeBytes(true, kv)
	} else {
		kv := &fifoEntry{key, value}
		ele := c.ll.PushFront(kv)
		c.cache[key] = ele
		c.computeBytes(true, kv)
	}
	for c.maxBytes != 0 && c.nBytes > c.maxBytes {
		c.RemoveOldest()
	}
}

func (c *FIFOCache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*fifoEntry)
		delete(c.cache, kv.key)
		c.computeBytes(false, kv) // 减少内存
	}
}

func (c *FIFOCache) Del(key string) {
	if ele, ok := c.cache[key]; ok {
		c.ll.Remove(ele)
		kv := ele.Value.(*fifoEntry)
		delete(c.cache, kv.key)
		c.computeBytes(false, kv)
	}
}

func (c *FIFOCache) computeBytes(isSum bool, kv *fifoEntry) {
	if isSum {
		c.nBytes += int64(len(kv.key)) + int64(kv.value.Len())
	}
	c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
}

func (c *FIFOCache) Len() int {
	return c.ll.Len()
}
