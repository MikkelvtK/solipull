package cache

import (
	"fmt"
	"sync"
)

type Cache[K comparable, V any] struct {
	cache map[K][]V
	mu    sync.Mutex
}

func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		cache: make(map[K][]V),
		mu:    sync.Mutex{},
	}
}

func (c *Cache[K, V]) Get(key K) ([]V, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	v, ok := c.cache[key]
	if !ok {
		return nil, fmt.Errorf("no values for key %v", key)
	}

	return v, nil
}

func (c *Cache[K, V]) GetAll() (map[K][]V, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil {
		return nil, fmt.Errorf("cache is empty")
	}

	return c.cache, nil
}

func (c *Cache[K, V]) Put(key K, val V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil {
		c.cache = make(map[K][]V)
	}

	if _, ok := c.cache[key]; !ok {
		c.cache[key] = make([]V, 0)
	}

	c.cache[key] = append(c.cache[key], val)
}
