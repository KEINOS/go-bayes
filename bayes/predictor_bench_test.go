package bayes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkPredictor_Train(b *testing.B) {
	score := benchmarkScore()

	for b.Loop() {
		predictor, err := NewPredictor(PredictorConfig{
			Storage: MemoryStorage,
			ScopeID: 0,
			Hasher:  nil,
		})
		require.NoError(b, err)

		err = predictor.Train(score)
		require.NoError(b, err)
	}
}

func BenchmarkPredictor_Predict(b *testing.B) {
	score := benchmarkScore()

	predictor, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  nil,
	})
	require.NoError(b, err)
	require.NoError(b, predictor.Train(score))

	quiz := []string{"So", "So", "La", "So", "Do", "Si"}

	for b.Loop() {
		_, err := predictor.Predict(quiz)
		require.NoError(b, err)
	}
}

func BenchmarkPredictor_HashTrans(b *testing.B) {
	predictor, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  nil,
	})
	require.NoError(b, err)

	context := []any{"So", int(-1), uint64(42), float64(1.5), true}

	for b.Loop() {
		_, err := predictor.HashTrans(context...)
		require.NoError(b, err)
	}
}

func benchmarkScore() []string {
	return []string{
		"So", "So", "La", "So", "Do", "Si",
		"So", "So", "La", "So", "Re", "Do",
		"So", "So", "So", "Mi", "Do", "Si", "La",
		"Fa", "Fa", "Mi", "Do", "Re", "Do",
	}
}
