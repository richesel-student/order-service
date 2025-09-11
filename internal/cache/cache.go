package cache

import (
	"sync"
	"time"
	"yourmodule/internal/models"
)

type Cache struct {
	mu        sync.RWMutex
	items     map[string]models.Order
	createdAt time.Time
}

func New() *Cache {
	return &Cache{
		items:     make(map[string]models.Order),
		createdAt: time.Now(),
	}
}

func (c *Cache) Get(id string) (models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	o, ok := c.items[id]
	return o, ok
}

func (c *Cache) Set(id string, o models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[id] = o
}
