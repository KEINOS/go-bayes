//nolint:exhaustruct // tests set only fields relevant to each case.
package mapstore

import (
	"context"
	"errors"
	"iter"
	"maps"
	"math"
	"testing"

	"github.com/KEINOS/go-bayes/bayes/modelstore"
	"github.com/stretchr/testify/require"
)

var errTestSinkStopped = errors.New("sink stopped")

func TestStore_returnsCopiesAndRejectsOperationsAfterClose(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := New(42)

	classPayload := []byte("answer")
	batch := batchOf(
		[]modelstore.Class{{ID: 2, TypeTag: 2, Payload: classPayload}},
		modelstore.TransitionDelta{FromID: 10, ToID: 2, Count: 2},
		modelstore.TransitionDelta{FromID: 10, ToID: 2, Count: 3},
	)
	require.NoError(t, store.Apply(ctx, batch))

	classPayload[0] = 'X'
	classes, err := store.Classes(ctx)
	require.NoError(t, err)
	require.Equal(t, []byte("answer"), classes[0].Payload)
	classes[0].Payload[0] = 'X'
	reloaded, err := store.Classes(ctx)
	require.NoError(t, err)
	require.Equal(t, []byte("answer"), reloaded[0].Payload)

	require.NoError(t, store.Close())
	require.NoError(t, store.Close())
	require.ErrorIs(t, store.Apply(ctx, batch), modelstore.ErrClosed)
	_, err = store.Stats(ctx, 0)
	require.ErrorIs(t, err, modelstore.ErrClosed)
	_, err = store.Classes(ctx)
	require.ErrorIs(t, err, modelstore.ErrClosed)
	require.ErrorIs(
		t,
		store.ExportTransitions(ctx, func(modelstore.TransitionCount) error { return nil }),
		modelstore.ErrClosed,
	)
	require.ErrorIs(t, store.Reset(ctx), modelstore.ErrClosed)
}

//nolint:funlen // table cases and post-validation state checks belong together.
func TestStore_validationAndAtomicity(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tests := map[string]struct {
		batch modelstore.TrainingBatch
		err   error
	}{
		"nil iterator": {
			batch: modelstore.TrainingBatch{},
			err:   modelstore.ErrInvalidBatch,
		},
		"zero delta": {
			batch: batchOf([]modelstore.Class{{ID: 2}}, modelstore.TransitionDelta{FromID: 1, ToID: 2}),
			err:   modelstore.ErrInvalidBatch,
		},
		"missing class": {
			batch: batchOf(nil, modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: 1}),
			err:   modelstore.ErrInvalidBatch,
		},
		"conflicting classes in batch": {
			batch: batchOf(
				[]modelstore.Class{
					{ID: 2, TypeTag: 1, Payload: []byte("A")},
					{ID: 2, TypeTag: 1, Payload: []byte("B")},
				},
			),
			err: modelstore.ErrClassConflict,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			store := New(1)
			require.ErrorIs(t, store.Apply(ctx, test.batch), test.err)
			stats, err := store.Stats(ctx, 1)
			require.NoError(t, err)
			require.Zero(t, stats.Total)
			require.Zero(t, stats.FromCount)
			require.Empty(t, stats.Candidates)
		})
	}

	canceled, cancel := context.WithCancel(ctx)
	cancel()

	store := New(1)
	require.ErrorIs(t, store.Apply(canceled, batchOf(nil)), context.Canceled)
	_, err := store.Classes(canceled)
	require.ErrorIs(t, err, context.Canceled)
	_, err = store.Stats(canceled, 0)
	require.ErrorIs(t, err, context.Canceled)
	require.ErrorIs(t, store.Reset(canceled), context.Canceled)
	require.ErrorIs(
		t,
		store.ExportTransitions(canceled, func(modelstore.TransitionCount) error { return nil }),
		context.Canceled,
	)
	require.ErrorIs(t, store.ExportTransitions(ctx, nil), modelstore.ErrInvalidBatch)

	require.NoError(t, store.Apply(ctx, batchOf(
		[]modelstore.Class{{ID: 2, TypeTag: 1, Payload: []byte("A")}},
		modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: 1},
	)))
	require.ErrorIs(t, store.Apply(ctx, batchOf(
		[]modelstore.Class{{ID: 2, TypeTag: 1, Payload: []byte("B")}},
		modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: 1},
	)), modelstore.ErrClassConflict)
	require.ErrorIs(
		t,
		store.ExportTransitions(ctx, func(modelstore.TransitionCount) error { return errTestSinkStopped }),
		errTestSinkStopped,
	)
}

func TestStore_overflowIsAtomic(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		prepare func(*Store)
		batch   modelstore.TrainingBatch
	}{
		"batch pair count": {
			batch: batchOf(
				[]modelstore.Class{{ID: 2}},
				modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: math.MaxInt64},
				modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: 1},
			),
		},
		"batch input count": {
			batch: batchOf(
				[]modelstore.Class{{ID: 2}, {ID: 3}},
				modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: math.MaxInt64},
				modelstore.TransitionDelta{FromID: 1, ToID: 3, Count: 1},
			),
		},
		"batch class count": {
			batch: batchOf(
				[]modelstore.Class{{ID: 2}},
				modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: math.MaxInt64},
				modelstore.TransitionDelta{FromID: 3, ToID: 2, Count: 1},
			),
		},
		"batch total count": {
			batch: batchOf(
				[]modelstore.Class{{ID: 2}, {ID: 4}},
				modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: math.MaxInt64},
				modelstore.TransitionDelta{FromID: 3, ToID: 4, Count: 1},
			),
		},
	}

	assertOverflowCasesAreAtomic(t, tests)
}

func TestStore_storedOverflowIsAtomic(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		prepare func(*Store)
		batch   modelstore.TrainingBatch
	}{
		"stored total count": {
			prepare: func(store *Store) { store.total = math.MaxInt64 },
			batch: batchOf(
				[]modelstore.Class{{ID: 2}},
				modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: 1},
			),
		},
		"stored input count": {
			prepare: func(store *Store) { store.from[1] = math.MaxInt64 },
			batch: batchOf(
				[]modelstore.Class{{ID: 2}},
				modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: 1},
			),
		},
		"stored class count": {
			prepare: func(store *Store) { store.to[2] = math.MaxInt64 },
			batch: batchOf(
				[]modelstore.Class{{ID: 2}},
				modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: 1},
			),
		},
		"stored pair count": {
			prepare: func(store *Store) { store.pairs[pair{from: 1, to: 2}] = math.MaxInt64 },
			batch: batchOf(
				[]modelstore.Class{{ID: 2}},
				modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: 1},
			),
		},
	}

	assertOverflowCasesAreAtomic(t, tests)
}

func assertOverflowCasesAreAtomic(t *testing.T, tests map[string]struct {
	prepare func(*Store)
	batch   modelstore.TrainingBatch
}) {
	t.Helper()

	ctx := context.Background()

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			store := New(1)
			if test.prepare != nil {
				test.prepare(store)
			}

			beforeTotal := store.total
			beforeFrom := maps.Clone(store.from)
			beforeTo := maps.Clone(store.to)
			beforePairs := maps.Clone(store.pairs)

			err := store.Apply(ctx, test.batch)
			require.ErrorIs(t, err, modelstore.ErrCountOverflow)
			require.Equal(t, beforeTotal, store.total)
			require.Equal(t, beforeFrom, store.from)
			require.Equal(t, beforeTo, store.to)
			require.Equal(t, beforePairs, store.pairs)
			require.Empty(t, store.classes)
		})
	}
}

func TestStore_ordersModelData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := New(1)
	require.NoError(t, store.Apply(ctx, batchOf(
		[]modelstore.Class{{ID: 9}, {ID: 1}, {ID: 5}},
		modelstore.TransitionDelta{FromID: 9, ToID: 9, Count: 1},
		modelstore.TransitionDelta{FromID: 1, ToID: 5, Count: 1},
		modelstore.TransitionDelta{FromID: 1, ToID: 1, Count: 1},
	)))

	classes, err := store.Classes(ctx)
	require.NoError(t, err)
	require.Equal(t, []uint64{1, 5, 9}, []uint64{classes[0].ID, classes[1].ID, classes[2].ID})

	stats, err := store.Stats(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, []uint64{1, 5, 9}, []uint64{
		stats.Candidates[0].ClassID,
		stats.Candidates[1].ClassID,
		stats.Candidates[2].ClassID,
	})

	var records []modelstore.TransitionCount
	require.NoError(t, store.ExportTransitions(ctx, func(record modelstore.TransitionCount) error {
		records = append(records, record)

		return nil
	}))
	require.Equal(t, []modelstore.TransitionCount{
		{FromID: 1, ToID: 1, Count: 1},
		{FromID: 1, ToID: 5, Count: 1},
		{FromID: 9, ToID: 9, Count: 1},
	}, records)
}

func TestStore_cancelsDuringStreamingOperations(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	store := New(1)
	batch := modelstore.TrainingBatch{
		Classes: []modelstore.Class{{ID: 2}, {ID: 3}},
		Transitions: func() iter.Seq[modelstore.TransitionDelta] {
			return func(yield func(modelstore.TransitionDelta) bool) {
				if !yield(modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: 1}) {
					return
				}

				cancel()
				yield(modelstore.TransitionDelta{FromID: 1, ToID: 3, Count: 1})
			}
		},
	}

	require.ErrorIs(t, store.Apply(ctx, batch), context.Canceled)
	require.Zero(t, store.total)
	require.Empty(t, store.classes)

	activeStore := New(2)
	require.NoError(t, activeStore.Apply(context.Background(), batchOf(
		[]modelstore.Class{{ID: 2}, {ID: 3}},
		modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: 1},
		modelstore.TransitionDelta{FromID: 1, ToID: 3, Count: 1},
	)))

	exportContext, stopExport := context.WithCancel(context.Background())
	exported := 0
	err := activeStore.ExportTransitions(exportContext, func(modelstore.TransitionCount) error {
		exported++
		stopExport()

		return nil
	})
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 1, exported)
}

func batchOf(classes []modelstore.Class, deltas ...modelstore.TransitionDelta) modelstore.TrainingBatch {
	return modelstore.TrainingBatch{
		Classes: classes,
		Transitions: func() iter.Seq[modelstore.TransitionDelta] {
			return func(yield func(modelstore.TransitionDelta) bool) {
				for _, delta := range deltas {
					if !yield(delta) {
						return
					}
				}
			}
		},
	}
}
