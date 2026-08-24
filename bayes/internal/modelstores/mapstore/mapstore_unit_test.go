//nolint:exhaustruct // tests set only fields relevant to each case.
package mapstore

import (
	"context"
	"errors"
	"iter"
	"math"
	"testing"

	"github.com/KEINOS/go-bayes/bayes/modelstore"
	"github.com/stretchr/testify/require"
)

var errTestSinkStopped = errors.New("sink stopped")

//nolint:funlen // one lifecycle test keeps state transitions in order.
func TestStore_lifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := New(42)
	require.Equal(t, uint64(42), store.ScopeID())

	stats, err := store.Stats(ctx, 10)
	require.NoError(t, err)
	require.Zero(t, stats.Total)
	require.Zero(t, stats.FromCount)
	require.Empty(t, stats.Candidates)

	classes, err := store.Classes(ctx)
	require.NoError(t, err)
	require.Empty(t, classes)

	classPayload := []byte("answer")
	batch := batchOf(
		[]modelstore.Class{{ID: 2, TypeTag: 2, Payload: classPayload}},
		modelstore.TransitionDelta{FromID: 10, ToID: 2, Count: 2},
		modelstore.TransitionDelta{FromID: 10, ToID: 2, Count: 3},
	)
	require.NoError(t, store.Apply(ctx, batch))

	classPayload[0] = 'X'

	stats, err = store.Stats(ctx, 10)
	require.NoError(t, err)
	require.Equal(t, modelstore.Stats{
		Total:     5,
		FromCount: 5,
		Candidates: []modelstore.CandidateStats{{
			ClassID: 2, ToCount: 5, PairCount: 5,
		}},
	}, stats)

	classes, err = store.Classes(ctx)
	require.NoError(t, err)
	require.Equal(t, []byte("answer"), classes[0].Payload)
	classes[0].Payload[0] = 'X'
	reloaded, err := store.Classes(ctx)
	require.NoError(t, err)
	require.Equal(t, []byte("answer"), reloaded[0].Payload)

	var exported []modelstore.TransitionCount

	require.NoError(t, store.ExportTransitions(ctx, func(record modelstore.TransitionCount) error {
		exported = append(exported, record)

		return nil
	}))
	require.Equal(t, []modelstore.TransitionCount{{FromID: 10, ToID: 2, Count: 5}}, exported)

	require.NoError(t, store.Reset(ctx))
	stats, err = store.Stats(ctx, 10)
	require.NoError(t, err)
	require.Zero(t, stats.Total)
	require.Zero(t, stats.FromCount)
	require.Empty(t, stats.Candidates)

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

	ctx := context.Background()
	key := pair{from: 1, to: 2}
	store := New(1)
	store.total = math.MaxInt64
	store.classes[2] = modelstore.Class{ID: 2}
	store.from[1] = math.MaxInt64
	store.to[2] = math.MaxInt64
	store.pairs[key] = math.MaxInt64

	err := store.Apply(ctx, batchOf(
		[]modelstore.Class{{ID: 2}},
		modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: 1},
	))
	require.ErrorIs(t, err, modelstore.ErrCountOverflow)
	require.Equal(t, int64(math.MaxInt64), store.total)
	require.Equal(t, int64(math.MaxInt64), store.pairs[key])

	overflowingBatch := batchOf(
		[]modelstore.Class{{ID: 3}},
		modelstore.TransitionDelta{FromID: 4, ToID: 3, Count: math.MaxInt64},
		modelstore.TransitionDelta{FromID: 4, ToID: 3, Count: 1},
	)
	empty := New(2)
	require.ErrorIs(t, empty.Apply(ctx, overflowingBatch), modelstore.ErrCountOverflow)
	require.Zero(t, empty.total)
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
