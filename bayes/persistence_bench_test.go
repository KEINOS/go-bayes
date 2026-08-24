//go:build cgo

package bayes

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkSQLite_Train(b *testing.B) {
	ctx := context.Background()
	predictor, err := New(
		ctx,
		SQLiteStorage,
		0,
		WithSQLitePath(filepath.Join(b.TempDir(), "train.db")),
	)
	require.NoError(b, err)
	b.Cleanup(func() { require.NoError(b, predictor.Close()) })

	sequence := []string{"A", "B", "C", "D"}
	b.ResetTimer()

	for b.Loop() {
		require.NoError(b, predictor.Train(ctx, sequence))
	}
}

func BenchmarkSQLite_Predict(b *testing.B) {
	ctx := context.Background()
	predictor, err := New(
		ctx,
		SQLiteStorage,
		0,
		WithSQLitePath(filepath.Join(b.TempDir(), "predict.db")),
	)
	require.NoError(b, err)
	b.Cleanup(func() { require.NoError(b, predictor.Close()) })
	require.NoError(b, predictor.Train(ctx, []string{"A", "B", "C", "D"}))

	query := []string{"A", "B", "C"}
	b.ResetTimer()

	for b.Loop() {
		_, err = predictor.Predict(ctx, query)
		require.NoError(b, err)
	}
}

func BenchmarkSQLite_Save(b *testing.B) {
	ctx := context.Background()
	predictor, err := New(ctx, MemoryStorage, 0)
	require.NoError(b, err)
	b.Cleanup(func() { require.NoError(b, predictor.Close()) })
	require.NoError(b, predictor.Train(ctx, benchmarkScore()))

	path := filepath.Join(b.TempDir(), "portable.db")
	b.ResetTimer()

	for b.Loop() {
		require.NoError(b, predictor.Save(ctx, path))
	}
}

func BenchmarkSQLite_Load(b *testing.B) {
	ctx := context.Background()
	predictor, err := New(ctx, MemoryStorage, 0)
	require.NoError(b, err)
	b.Cleanup(func() { require.NoError(b, predictor.Close()) })
	require.NoError(b, predictor.Train(ctx, benchmarkScore()))

	path := filepath.Join(b.TempDir(), "portable.db")
	require.NoError(b, predictor.Save(ctx, path))
	b.ResetTimer()

	for b.Loop() {
		loaded, loadErr := Load(ctx, path)
		require.NoError(b, loadErr)
		require.NoError(b, loaded.Close())
	}
}
