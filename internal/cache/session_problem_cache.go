package cache

import (
	"sync"
)

type SessionProblemCache struct {
	mu   sync.RWMutex
	data map[uint][]uint
}

func NewSessionProblemCache() *SessionProblemCache {
	return &SessionProblemCache{
		data: make(map[uint][]uint),
	}
}

func (c *SessionProblemCache) Get(sessionID uint) ([]uint, bool) {
	c.mu.RLock()
	ids, ok := c.data[sessionID]
	c.mu.RUnlock()
	return ids, ok
}

func (c *SessionProblemCache) Register(sessionID uint, problemIDs []uint) {
	c.mu.Lock()
	ids := make([]uint, len(problemIDs))
	copy(ids, problemIDs)
	c.data[sessionID] = ids
	c.mu.Unlock()
}

func (c *SessionProblemCache) Unregister(sessionID uint) {
	c.mu.Lock()
	delete(c.data, sessionID)
	c.mu.Unlock()
}
