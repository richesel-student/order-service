package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Cache struct {
	mu       sync.RWMutex
	items    map[string]Item
	ttl      time.Duration
	cleanup  time.Duration
	cancelGC context.CancelFunc
}

type Item struct {
	Value      interface{}
	Created    time.Time
	Expiration int64
}

func New(ttl, cleanupInterval time.Duration) *Cache {
	c := &Cache{
		items:   make(map[string]Item),
		ttl:     ttl,
		cleanup: cleanupInterval,
	}

	if cleanupInterval > 0 {
		ctx, cancel := context.WithCancel(context.Background())
		c.cancelGC = cancel
		go c.gc(ctx)
	}
	return c
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	if ttl == 0 {
		ttl = c.ttl
	}
	exp := int64(0)
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = Item{Value: value, Created: time.Now(), Expiration: exp}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.items[key]
	if !ok {
		return nil, false
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return nil, false
	}
	return item.Value, true
}

func (c *Cache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.items[key]; !ok {
		return errors.New("key not found")
	}
	delete(c.items, key)
	return nil
}

func (c *Cache) gc(ctx context.Context) {
	ticker := time.NewTicker(c.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.mu.Lock()
			for k, v := range c.items {
				if v.Expiration > 0 && time.Now().UnixNano() > v.Expiration {
					delete(c.items, k)
				}
			}
			c.mu.Unlock()
		}
	}
}

func (c *Cache) StopGC() {
	if c.cancelGC != nil {
		c.cancelGC()
	}
}
