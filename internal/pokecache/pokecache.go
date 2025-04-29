package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	cacheMap map[string]cacheEntry
	mu       sync.Mutex
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		cacheMap: make(map[string]cacheEntry),
	}
	go c.reapLoop(interval)
	return c
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, v := range c.cacheMap {
			if now.Sub(v.createdAt) > interval {
				delete(c.cacheMap, k)
			}
		}
		c.mu.Unlock()
	}
}

func (c *Cache) Add(key string, val []byte) {
	currentTime := time.Now()
	x := cacheEntry{
		createdAt: currentTime,
		val:       val,
	}
	c.mu.Lock()
	c.cacheMap[key] = x
	c.mu.Unlock()
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	val, ok := c.cacheMap[key]
	c.mu.Unlock()

	if !ok {
		return nil, false
	}
	return val.val, true
}
