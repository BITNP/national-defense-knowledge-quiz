package cache

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"national-defense-knowledge-quiz/internal/model"
	"national-defense-knowledge-quiz/internal/repository"
)

type ProblemCache struct {
	mu   sync.RWMutex
	data map[uint][]model.Problem
}

func NewProblemCache() *ProblemCache {
	return &ProblemCache{
		data: make(map[uint][]model.Problem),
	}
}

func (c *ProblemCache) LoadAll(ctx context.Context, repo *repository.ProblemRepo) error {
	all, err := repo.GetAllActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to load active problems: %w", err)
	}

	for _, probs := range all {
		for i := range probs {
			probs[i].DataSlice = probs[i].DataArray()
		}
	}

	c.mu.Lock()
	c.data = all
	c.mu.Unlock()

	examCount := len(all)
	probCount := 0
	for _, probs := range all {
		probCount += len(probs)
	}
	fmt.Printf("problem cache warmed: %d exams, %d problems total\n", examCount, probCount)
	return nil
}

func (c *ProblemCache) Get(examID uint) ([]model.Problem, bool) {
	c.mu.RLock()
	probs, ok := c.data[examID]
	c.mu.RUnlock()
	return probs, ok
}

func (c *ProblemCache) GetByIDs(examID uint, ids []uint) ([]model.Problem, error) {
	c.mu.RLock()
	all, ok := c.data[examID]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no problems found for exam %d", examID)
	}
	byID := make(map[uint]model.Problem, len(all))
	for _, p := range all {
		byID[p.ID] = p
	}
	result := make([]model.Problem, len(ids))
	for i, id := range ids {
		p, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("problem %d not found in cache", id)
		}
		result[i] = p
	}
	return result, nil
}

func (c *ProblemCache) GetRandom(examID uint, n int) ([]model.Problem, error) {
	all, ok := c.Get(examID)
	if !ok {
		return nil, fmt.Errorf("no active problems found for exam %d", examID)
	}

	if n <= 0 || n >= len(all) {
		result := make([]model.Problem, len(all))
		copy(result, all)
		return result, nil
	}

	perm := rand.Perm(len(all))
	result := make([]model.Problem, n)
	for i := 0; i < n; i++ {
		result[i] = all[perm[i]]
	}
	return result, nil
}
