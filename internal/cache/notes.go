package cache

import (
	"os"
	"sync"
	"time"
)

type entry struct {
	content []byte
	modTime time.Time
}

type NoteCache struct {
	mu      sync.RWMutex
	items   map[string]*entry
	maxSize int
	order   []string
}

func NewNoteCache(maxSize int) *NoteCache {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &NoteCache{
		items:   make(map[string]*entry),
		maxSize: maxSize,
	}
}

func (c *NoteCache) ReadFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	c.mu.RLock()
	if e, ok := c.items[path]; ok && e.modTime.Equal(info.ModTime()) {
		c.mu.RUnlock()
		return e.content, nil
	}
	c.mu.RUnlock()

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.items[path]; ok && e.modTime.Equal(info.ModTime()) {
		return e.content, nil
	}

	c.items[path] = &entry{content: content, modTime: info.ModTime()}

	for i, p := range c.order {
		if p == path {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
	c.order = append(c.order, path)

	for len(c.items) > c.maxSize {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.items, oldest)
	}

	return content, nil
}

func (c *NoteCache) Invalidate(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, path)

	for i, p := range c.order {
		if p == path {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
}

func (c *NoteCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*entry)
	c.order = nil
}
