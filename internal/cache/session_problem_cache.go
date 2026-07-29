package cache

import (
	"context"
	"fmt"
	"sync"

	"national-defense-knowledge-quiz/internal/repository"
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

func (c *SessionProblemCache) LoadAll(ctx context.Context, repo *repository.ExamSessionRepo) error {
	all, err := repo.GetAllProblemIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to load session problem IDs: %w", err)
	}
	c.mu.Lock()
	c.data = all
	c.mu.Unlock()
	fmt.Printf("session problem cache warmed: %d sessions\n", len(all))
	return nil
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
