package tmdb

import (
	"container/list"
	"sync"
	"time"
)

// Cache implements a thread-safe LRU cache with TTL
type Cache struct {
	mu       sync.RWMutex
	items    map[string]*list.Element
	order    *list.List
	maxItems int
	memTTL   time.Duration
	diskTTL  time.Duration
}

type cacheEntry struct {
	key       string
	value     interface{}
	expiresAt time.Time
}

// NewCache creates a new LRU cache
func NewCache(maxItems int, memTTL, diskTTL time.Duration) *Cache {
	c := &Cache{
		items:    make(map[string]*list.Element),
		order:    list.New(),
		maxItems: maxItems,
		memTTL:   memTTL,
		diskTTL:  diskTTL,
	}

	// Start background cleanup
	go c.cleanup()

	return c
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	elem, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	entry := elem.Value.(*cacheEntry)

	// Check TTL
	if time.Now().After(entry.expiresAt) {
		c.mu.Lock()
		c.removeElement(elem)
		c.mu.Unlock()
		return nil, false
	}

	// Move to front (most recently used)
	c.mu.Lock()
	c.order.MoveToFront(elem)
	c.mu.Unlock()

	return entry.value, true
}

// Set adds or updates a value in the cache
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update existing
	if elem, exists := c.items[key]; exists {
		c.order.MoveToFront(elem)
		entry := elem.Value.(*cacheEntry)
		entry.value = value
		entry.expiresAt = time.Now().Add(c.memTTL)
		return
	}

	// Evict oldest if at capacity
	if c.order.Len() >= c.maxItems {
		oldest := c.order.Back()
		if oldest != nil {
			c.removeElement(oldest)
		}
	}

	// Add new entry
	entry := &cacheEntry{
		key:       key,
		value:     value,
		expiresAt: time.Now().Add(c.memTTL),
	}
	elem := c.order.PushFront(entry)
	c.items[key] = elem
}

// Delete removes a key from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		c.removeElement(elem)
	}
}

// Clear removes all entries
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.order.Init()
}

// Len returns the number of items in cache
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.order.Len()
}

func (c *Cache) removeElement(elem *list.Element) {
	c.order.Remove(elem)
	entry := elem.Value.(*cacheEntry)
	delete(c.items, entry.key)
}

// cleanup periodically removes expired entries
func (c *Cache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for elem := c.order.Back(); elem != nil; {
			entry := elem.Value.(*cacheEntry)
			if now.After(entry.expiresAt) {
				prev := elem.Prev()
				c.removeElement(elem)
				elem = prev
			} else {
				elem = elem.Prev()
			}
		}
		c.mu.Unlock()
	}
}
