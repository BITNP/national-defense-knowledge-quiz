package cache

import (
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

func (c *ProblemCache) LoadAll(repo *repository.ProblemRepo) error {
	all, err := repo.GetAllActive()
	if err != nil {
		return fmt.Errorf("failed to load active problems: %w", err)
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
