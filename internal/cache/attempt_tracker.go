package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"national-defense-knowledge-quiz/internal/repository"
)

type activeSession struct {
	ID      uint
	EndTime time.Time
}

type AttemptTracker struct {
	mu     sync.RWMutex
	counts map[string]int
	active map[string]*activeSession
}

func NewAttemptTracker() *AttemptTracker {
	return &AttemptTracker{
		counts: make(map[string]int),
		active: make(map[string]*activeSession),
	}
}

func (t *AttemptTracker) Seed(ctx context.Context, repo *repository.ExamSessionRepo) error {
	now := time.Now()

	completed, err := repo.CountCompletedGroupBy(ctx, now)
	if err != nil {
		return fmt.Errorf("failed to seed completed counts: %w", err)
	}

	t.mu.Lock()
	for _, c := range completed {
		key := t.key(c.ExamID, c.StudentID)
		t.counts[key] = int(c.Count)
	}
	t.mu.Unlock()

	active, err := repo.ListActiveSessions(ctx, now)
	if err != nil {
		return fmt.Errorf("failed to seed active sessions: %w", err)
	}

	t.mu.Lock()
	for _, a := range active {
		key := t.key(a.ExamID, a.StudentID)
		t.active[key] = &activeSession{ID: a.ID, EndTime: a.EndTime}
	}
	t.mu.Unlock()

	fmt.Printf("attempt tracker seeded: %d completed counts, %d active sessions\n", len(completed), len(active))
	return nil
}

func (t *AttemptTracker) key(examID uint, studentID string) string {
	return fmt.Sprintf("%d:%s", examID, studentID)
}

func (t *AttemptTracker) Refresh(examID uint, studentID string, now time.Time) {
	key := t.key(examID, studentID)
	t.mu.Lock()
	if a, ok := t.active[key]; ok && !a.EndTime.After(now) {
		t.counts[key]++
		delete(t.active, key)
	}
	t.mu.Unlock()
}

func (t *AttemptTracker) GetCount(examID uint, studentID string) int {
	key := t.key(examID, studentID)
	t.mu.RLock()
	c := t.counts[key]
	t.mu.RUnlock()
	return c
}

func (t *AttemptTracker) GetActive(examID uint, studentID string) (uint, bool) {
	key := t.key(examID, studentID)
	t.mu.RLock()
	a, ok := t.active[key]
	t.mu.RUnlock()
	if !ok {
		return 0, false
	}
	return a.ID, true
}

func (t *AttemptTracker) RegisterActive(examID uint, studentID string, sessionID uint, endTime time.Time) {
	key := t.key(examID, studentID)
	t.mu.Lock()
	t.active[key] = &activeSession{ID: sessionID, EndTime: endTime}
	t.mu.Unlock()
}

func (t *AttemptTracker) MarkFinished(examID uint, studentID string, sessionID uint) {
	key := t.key(examID, studentID)
	t.mu.Lock()
	if a, ok := t.active[key]; ok && a.ID == sessionID {
		t.counts[key]++
		delete(t.active, key)
	} else {
		t.counts[key]++
	}
	t.mu.Unlock()
}
