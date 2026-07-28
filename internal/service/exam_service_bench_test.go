package service

import (
	"math/rand"
	"testing"

	"national-defense-knowledge-quiz/internal/model"
)

func makePrizes(n int) []model.Prize {
	prizes := make([]model.Prize, n)
	for i := 0; i < n; i++ {
		prizes[i] = model.Prize{ID: uint(i + 1), Remain: n - i}
	}
	return prizes
}

func BenchmarkWeightedRandom10(b *testing.B) {
	prizes := makePrizes(10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		weightedRandom(prizes)
	}
}

func BenchmarkWeightedRandom100(b *testing.B) {
	prizes := makePrizes(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		weightedRandom(prizes)
	}
}

func BenchmarkWeightedRandom1000(b *testing.B) {
	prizes := makePrizes(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		weightedRandom(prizes)
	}
}

func BenchmarkDataArray(b *testing.B) {
	p := model.Problem{Data: `["option A","option B","option C","option D"]`}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.DataArray()
	}
}

func init() {
	rand.Seed(1)
}
