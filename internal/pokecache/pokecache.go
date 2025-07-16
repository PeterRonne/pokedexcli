package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	Storage map[string]cacheEntry `json:"storage"`
	Mu      sync.RWMutex          `json:"mux"`
}

func (c *Cache) Add(key string, val []byte) {

	c.Mu.Lock()
	c.Storage[key] = cacheEntry{
		CreatedAt: time.Now(),
		Val:       val,
	}
	c.Mu.Unlock()
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	entry, ok := c.Storage[key]
	if !ok {
		return nil, ok
	}

	return entry.Val, ok
}

type cacheEntry struct {
	CreatedAt time.Time `json:"created_at"`
	Val       []byte    `json:"val"`
}

func NewCache(interval time.Duration) *Cache {
	cache := Cache{
		Storage: make(map[string]cacheEntry),
		Mu:      sync.RWMutex{},
	}

	go reapLoop(interval, &cache)
	return &cache
}

func reapLoop(interval time.Duration, cache *Cache) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		cache.Mu.Lock()
		for key, entry := range cache.Storage {
			if entry.CreatedAt.Add(interval).Before(time.Now()) {
				delete(cache.Storage, key)
			}
		}
		cache.Mu.Unlock()
	}
}
