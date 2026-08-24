//go:build !cgo

package bayes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSQLitePublicAPI_withoutCGO(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	_, err := New(ctx, SQLiteStorage, 1, WithSQLitePath("model.db"))
	require.ErrorIs(t, err, ErrSQLiteUnavailable)
	_, err = Load(ctx, "model.db")
	require.ErrorIs(t, err, ErrSQLiteUnavailable)
	_, err = Open(ctx, "model.db")
	require.ErrorIs(t, err, ErrSQLiteUnavailable)

	predictor, err := New(ctx, MemoryStorage, 1)
	require.NoError(t, err)
	require.ErrorIs(t, predictor.Save(ctx, "model.db"), ErrSQLiteUnavailable)
}
