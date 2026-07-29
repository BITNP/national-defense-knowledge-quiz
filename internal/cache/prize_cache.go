package cache

import (
	"context"
	"fmt"
	"sync"

	"national-defense-knowledge-quiz/internal/model"
	"national-defense-knowledge-quiz/internal/repository"
)

type PrizeCache struct {
	mu   sync.RWMutex
	data map[uint][]model.Prize
}

func NewPrizeCache() *PrizeCache {
	return &PrizeCache{
		data: make(map[uint][]model.Prize),
	}
}

func (c *PrizeCache) LoadAll(ctx context.Context, repo *repository.PrizeRepo) error {
	all, err := repo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load prizes: %w", err)
	}
	c.mu.Lock()
	c.data = make(map[uint][]model.Prize, len(all))
	for i := range all {
		if all[i].Remain <= 0 {
			continue
		}
		c.data[all[i].ExamID] = append(c.data[all[i].ExamID], all[i])
	}
	c.mu.Unlock()
	fmt.Printf("prize cache warmed: %d prizes across %d exams\n", len(all), len(c.data))
	return nil
}

func (c *PrizeCache) GetByExamID(examID uint) []model.Prize {
	c.mu.RLock()
	defer c.mu.RUnlock()
	prizes := c.data[examID]
	if len(prizes) == 0 {
		return nil
	}
	result := make([]model.Prize, 0, len(prizes))
	for _, p := range prizes {
		if p.Remain > 0 {
			result = append(result, p)
		}
	}
	return result
}

func (c *PrizeCache) DecrementRemain(examID, prizeID uint) {
	c.mu.Lock()
	defer c.mu.Unlock()
	prizes := c.data[examID]
	for i := range prizes {
		if prizes[i].ID == prizeID {
			prizes[i].Remain--
			if prizes[i].Remain <= 0 {
				c.data[examID] = append(prizes[:i], prizes[i+1:]...)
			}
			return
		}
	}
}
