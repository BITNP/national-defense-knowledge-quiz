package cache

import (
	"sync"
	"testing"

	"national-defense-knowledge-quiz/internal/model"
)

func makeTestProblems(examID uint, n int) []model.Problem {
	probs := make([]model.Problem, n)
	for i := 0; i < n; i++ {
		probs[i] = model.Problem{ID: uint(i) + 1, ExamID: examID}
	}
	return probs
}

func TestGet(t *testing.T) {
	c := NewProblemCache()
	c.data[1] = makeTestProblems(1, 10)
	c.data[2] = makeTestProblems(2, 5)

	p, ok := c.Get(1)
	if !ok || len(p) != 10 {
		t.Fatal("expected 10 problems for exam 1")
	}

	p, ok = c.Get(2)
	if !ok || len(p) != 5 {
		t.Fatal("expected 5 problems for exam 2")
	}

	_, ok = c.Get(99)
	if ok {
		t.Fatal("expected false for nonexistent exam")
	}
}

func TestGetRandom(t *testing.T) {
	c := NewProblemCache()
	c.data[1] = makeTestProblems(1, 100)

	t.Run("n less than total", func(t *testing.T) {
		probs, err := c.GetRandom(1, 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(probs) != 10 {
			t.Fatalf("expected 10 problems, got %d", len(probs))
		}
		ids := make(map[uint]bool)
		for _, p := range probs {
			if ids[p.ID] {
				t.Fatal("duplicate ID in random selection")
			}
			ids[p.ID] = true
		}
	})

	t.Run("n equals total", func(t *testing.T) {
		probs, err := c.GetRandom(1, 100)
		if err != nil {
			t.Fatal(err)
		}
		if len(probs) != 100 {
			t.Fatalf("expected 100 problems, got %d", len(probs))
		}
	})

	t.Run("n greater than total", func(t *testing.T) {
		probs, err := c.GetRandom(1, 200)
		if err != nil {
			t.Fatal(err)
		}
		if len(probs) != 100 {
			t.Fatalf("expected 100 problems, got %d", len(probs))
		}
	})

	t.Run("n is zero", func(t *testing.T) {
		probs, err := c.GetRandom(1, 0)
		if err != nil {
			t.Fatal(err)
		}
		if len(probs) != 100 {
			t.Fatalf("expected 100 problems, got %d", len(probs))
		}
	})

	t.Run("nonexistent exam", func(t *testing.T) {
		_, err := c.GetRandom(99, 10)
		if err == nil {
			t.Fatal("expected error for nonexistent exam")
		}
	})
}

func TestConcurrent(t *testing.T) {
	c := NewProblemCache()
	c.data[1] = makeTestProblems(1, 100)

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				_, _ = c.GetRandom(1, 10)
			}
		}()
	}
	wg.Wait()
}

func BenchmarkGetRandom(b *testing.B) {
	c := NewProblemCache()
	c.data[1] = makeTestProblems(1, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.GetRandom(1, 100)
	}
}
