package cache

import (
	"fmt"
	"sync"

	"national-defense-knowledge-quiz/internal/model"
	"national-defense-knowledge-quiz/internal/repository"
)

type ExamCache struct {
	mu   sync.RWMutex
	data map[uint]*model.Exam
}

func NewExamCache() *ExamCache {
	return &ExamCache{
		data: make(map[uint]*model.Exam),
	}
}

func (c *ExamCache) LoadAll(repo *repository.ExamRepo) error {
	all, err := repo.GetAllActive()
	if err != nil {
		return fmt.Errorf("failed to load active exams: %w", err)
	}
	c.mu.Lock()
	for i := range all {
		c.data[all[i].ID] = all[i]
	}
	c.mu.Unlock()
	fmt.Printf("exam cache warmed: %d exams\n", len(all))
	return nil
}

func (c *ExamCache) Get(id uint) (*model.Exam, bool) {
	c.mu.RLock()
	exam, ok := c.data[id]
	c.mu.RUnlock()
	return exam, ok
}
