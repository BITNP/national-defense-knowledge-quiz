package cache

import (
	"sync"
	"testing"

	"national-defense-knowledge-quiz/internal/model"
)

func TestExamCache(t *testing.T) {
	c := NewExamCache()
	c.data[1] = &model.Exam{ID: 1, Title: "Test Exam", Active: true}

	exam, ok := c.Get(1)
	if !ok || exam.Title != "Test Exam" {
		t.Fatal("expected exam 1")
	}

	_, ok = c.Get(99)
	if ok {
		t.Fatal("expected false for nonexistent exam")
	}
}

func TestExamCacheConcurrent(t *testing.T) {
	c := NewExamCache()
	c.data[1] = &model.Exam{ID: 1, Title: "Test", Active: true}

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				_, _ = c.Get(1)
			}
		}()
	}
	wg.Wait()
}
