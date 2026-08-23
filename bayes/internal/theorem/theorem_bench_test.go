package theorem

import "testing"

func BenchmarkBayes(b *testing.B) {
	priorPtoB := 0.3
	priorPfromAtoB := 0.4
	priorPNotFromAtoB := 0.5

	b.ResetTimer()

	for b.Loop() {
		_ = Bayes(priorPtoB, priorPfromAtoB, priorPNotFromAtoB)
	}
}
