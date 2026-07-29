package cache

import (
	"context"
	"testing"
	"time"
)

func TestNewAttemptTracker(t *testing.T) {
	tracker := NewAttemptTracker(nil)
	if tracker == nil {
		t.Fatal("expected non-nil tracker")
	}
	if len(tracker.counts) != 0 {
		t.Errorf("expected empty counts, got %d", len(tracker.counts))
	}
	if len(tracker.active) != 0 {
		t.Errorf("expected empty active, got %d", len(tracker.active))
	}
}

func TestRegisterAndGetActive(t *testing.T) {
	tracker := NewAttemptTracker(nil)
	id, ok := tracker.GetActive(1, "student1")
	if ok {
		t.Error("expected no active session")
	}
	if id != 0 {
		t.Errorf("expected 0, got %d", id)
	}

	now := time.Now()
	tracker.RegisterActive(1, "student1", 42, now.Add(time.Hour))
	id, ok = tracker.GetActive(1, "student1")
	if !ok {
		t.Fatal("expected active session")
	}
	if id != 42 {
		t.Errorf("expected 42, got %d", id)
	}
}

func TestRefreshExpiresOldSession(t *testing.T) {
	tracker := NewAttemptTracker(nil)
	now := time.Now()

	tracker.RegisterActive(1, "student1", 42, now.Add(-time.Minute))

	tracker.Refresh(1, "student1", now)
	if count := tracker.GetCount(context.Background(), 1, "student1"); count != 1 {
		t.Errorf("expected count 1 after refresh, got %d", count)
	}
	if _, ok := tracker.GetActive(1, "student1"); ok {
		t.Error("expected active removed after refresh")
	}
}

func TestRefreshKeepsActiveSession(t *testing.T) {
	tracker := NewAttemptTracker(nil)
	now := time.Now()

	tracker.RegisterActive(1, "student1", 42, now.Add(time.Hour))
	tracker.Refresh(1, "student1", now)

	if _, ok := tracker.GetActive(1, "student1"); !ok {
		t.Error("expected active still present")
	}
}

func TestMarkFinished(t *testing.T) {
	tracker := NewAttemptTracker(nil)
	now := time.Now()

	tracker.RegisterActive(1, "student1", 42, now.Add(time.Hour))
	tracker.MarkFinished(1, "student1", 42)

	if count := tracker.GetCount(context.Background(), 1, "student1"); count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
	if _, ok := tracker.GetActive(1, "student1"); ok {
		t.Error("expected active removed after mark finished")
	}
}

func TestMarkFinishedNoActive(t *testing.T) {
	tracker := NewAttemptTracker(nil)
	tracker.MarkFinished(1, "student1", 99)
	if count := tracker.GetCount(context.Background(), 1, "student1"); count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
}

func TestMultipleKeys(t *testing.T) {
	tracker := NewAttemptTracker(nil)
	now := time.Now()

	tracker.RegisterActive(1, "alice", 10, now.Add(time.Hour))
	tracker.RegisterActive(2, "bob", 20, now.Add(time.Hour))

	tracker.MarkFinished(1, "alice", 10)
	tracker.Refresh(2, "bob", now.Add(2*time.Hour))

	if c := tracker.GetCount(context.Background(), 1, "alice"); c != 1 {
		t.Errorf("alice count: expected 1, got %d", c)
	}
	if c := tracker.GetCount(context.Background(), 2, "bob"); c != 1 {
		t.Errorf("bob count: expected 1, got %d", c)
	}
	if _, ok := tracker.GetActive(1, "alice"); ok {
		t.Error("alice active should be removed")
	}
	if _, ok := tracker.GetActive(2, "bob"); ok {
		t.Error("bob active should be removed")
	}
}
