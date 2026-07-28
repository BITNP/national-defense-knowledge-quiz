package cache

import (
	"context"
	"fmt"
	"sync"

	"github.com/bits-and-blooms/bloom/v3"

	"national-defense-knowledge-quiz/internal/db"
	"national-defense-knowledge-quiz/internal/model"
)

const DefaultBloomCapacity = 100_000

type AttemptTracker struct {
	mu     sync.RWMutex
	filter *bloom.BloomFilter
}

func NewAttemptTracker(capacity uint) *AttemptTracker {
	if capacity < 1 {
		capacity = 1
	}
	return &AttemptTracker{
		filter: bloom.NewWithEstimates(capacity, 0.01),
	}
}

func (t *AttemptTracker) HasAttempted(examID uint, studentID string) bool {
	key := []byte(fmt.Sprintf("%d:%s", examID, studentID))
	t.mu.RLock()
	ok := t.filter.Test(key)
	t.mu.RUnlock()
	return ok
}

func (t *AttemptTracker) MarkAttempted(examID uint, studentID string) {
	key := []byte(fmt.Sprintf("%d:%s", examID, studentID))
	t.mu.Lock()
	t.filter.Add(key)
	t.mu.Unlock()
}

func (t *AttemptTracker) Seed(ctx context.Context) error {
	type pair struct {
		ExamID    uint
		StudentID string
	}
	var pairs []pair
	if err := db.DB.WithContext(ctx).Model(&model.ExamSession{}).
		Select("DISTINCT exam_id, student_id").Find(&pairs).Error; err != nil {
		return fmt.Errorf("failed to seed attempt tracker: %w", err)
	}

	t.mu.Lock()
	for _, p := range pairs {
		t.filter.Add([]byte(fmt.Sprintf("%d:%s", p.ExamID, p.StudentID)))
	}
	t.mu.Unlock()
	fmt.Printf("attempt tracker seeded: %d distinct examinees\n", len(pairs))
	return nil
}
