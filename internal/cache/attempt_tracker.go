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
	repo   *repository.ExamSessionRepo
}

func NewAttemptTracker(repo *repository.ExamSessionRepo) *AttemptTracker {
	return &AttemptTracker{
		counts: make(map[string]int),
		active: make(map[string]*activeSession),
		repo:   repo,
	}
}

func (t *AttemptTracker) LoadActiveSessions(ctx context.Context) error {
	now := time.Now()

	active, err := t.repo.ListActiveSessions(ctx, now)
	if err != nil {
		return fmt.Errorf("failed to load active sessions: %w", err)
	}

	t.mu.Lock()
	for _, a := range active {
		key := t.key(a.ExamID, a.StudentID)
		t.active[key] = &activeSession{ID: a.ID, EndTime: a.EndTime}
	}
	t.mu.Unlock()

	fmt.Printf("attempt tracker seeded: %d active sessions\n", len(active))
	return nil
}

func (t *AttemptTracker) WarmCompleted(ctx context.Context) error {
	now := time.Now()

	completed, err := t.repo.CountCompletedGroupBy(ctx, now)
	if err != nil {
		return fmt.Errorf("failed to warm completed counts: %w", err)
	}

	t.mu.Lock()
	for _, c := range completed {
		key := t.key(c.ExamID, c.StudentID)
		t.counts[key] = int(c.Count)
	}
	t.mu.Unlock()

	fmt.Printf("attempt tracker warmed: %d completed counts\n", len(completed))
	return nil
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

func (t *AttemptTracker) GetCount(ctx context.Context, examID uint, studentID string) int {
	key := t.key(examID, studentID)

	t.mu.RLock()
	c, ok := t.counts[key]
	t.mu.RUnlock()
	if ok {
		return c
	}

	count, err := t.repo.CountFinishedOrExpired(ctx, examID, studentID, time.Now())
	if err != nil {
		return 0
	}

	t.mu.Lock()
	if existing, loaded := t.counts[key]; loaded {
		t.mu.Unlock()
		return existing
	}
	t.counts[key] = int(count)
	t.mu.Unlock()
	return int(count)
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

func (t *AttemptTracker) key(examID uint, studentID string) string {
	return fmt.Sprintf("%d:%s", examID, studentID)
}
