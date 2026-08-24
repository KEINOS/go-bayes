//go:build cgo

//nolint:exhaustruct // tests set only fields relevant to each case.
package sqlitestore

import (
	"context"
	"database/sql"
	"errors"
	"iter"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/KEINOS/go-bayes/bayes/internal/modelstores/mapstore"
	"github.com/KEINOS/go-bayes/bayes/modelstore"
	"github.com/stretchr/testify/require"
)

var (
	errTestCommitFailed = errors.New("injected commit failure")
	errTestSinkFailed   = errors.New("sink failed")
)

//nolint:funlen // one lifecycle test keeps state transitions in order.
func TestStore_lifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "model.db")
	metadata := testMetadata()
	store, err := Create(ctx, path, metadata, OpenConfig{CacheKiB: 128})
	require.NoError(t, err)
	canonical, err := CanonicalPath(path)
	require.NoError(t, err)
	require.Equal(t, canonical, store.Path())
	require.Equal(t, metadata, store.Metadata())
	require.Equal(t, metadata.ScopeID, store.ScopeID())

	validated, err := store.Validate(ctx)
	require.NoError(t, err)
	require.Equal(t, metadata, validated)

	batch := sqliteBatch(
		[]modelstore.Class{
			{ID: math.MaxUint64, TypeTag: 2, Payload: []byte("last")},
			{ID: 2, TypeTag: 2, Payload: []byte("first")},
		},
		modelstore.TransitionDelta{FromID: math.MaxUint64, ToID: math.MaxUint64, Count: 2},
		modelstore.TransitionDelta{FromID: math.MaxUint64, ToID: math.MaxUint64, Count: 3},
		modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: 1},
	)
	require.NoError(t, store.Apply(ctx, batch))

	stats, err := store.Stats(ctx, math.MaxUint64)
	require.NoError(t, err)
	require.Equal(t, int64(6), stats.Total)
	require.Equal(t, int64(5), stats.FromCount)
	require.Equal(t, []modelstore.CandidateStats{
		{ClassID: 2, ToCount: 1, PairCount: 0},
		{ClassID: math.MaxUint64, ToCount: 5, PairCount: 5},
	}, stats.Candidates)

	classes, err := store.Classes(ctx)
	require.NoError(t, err)
	require.Equal(t, []uint64{2, math.MaxUint64}, []uint64{classes[0].ID, classes[1].ID})
	classes[0].Payload[0] = 'X'
	reloaded, err := store.Classes(ctx)
	require.NoError(t, err)
	require.Equal(t, []byte("first"), reloaded[0].Payload)

	var exported []modelstore.TransitionCount

	require.NoError(t, store.ExportTransitions(ctx, func(record modelstore.TransitionCount) error {
		exported = append(exported, record)

		return nil
	}))
	require.Equal(t, []modelstore.TransitionCount{
		{FromID: 1, ToID: 2, Count: 1},
		{FromID: math.MaxUint64, ToID: math.MaxUint64, Count: 5},
	}, exported)

	require.NoError(t, store.Close())
	require.NoError(t, store.Close())

	reopened, err := Open(ctx, path, OpenConfig{SynchronousNormal: true})
	require.NoError(t, err)
	stats, err = reopened.Stats(ctx, math.MaxUint64)
	require.NoError(t, err)
	require.Equal(t, int64(6), stats.Total)
	require.NoError(t, reopened.Reset(ctx))
	require.NoError(t, reopened.Close())

	portable, err := Open(ctx, path, OpenConfig{Portable: true})
	require.NoError(t, err)
	stats, err = portable.Stats(ctx, 0)
	require.NoError(t, err)
	require.Zero(t, stats.Total)
	require.NoError(t, portable.Close())
}

func TestStore_importAndFailures(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "import.db")
	store, err := Create(ctx, path, testMetadata(), OpenConfig{Portable: true})
	require.NoError(t, err)

	defer func() { require.NoError(t, store.Close()) }()

	require.ErrorIs(t, store.Apply(ctx, modelstore.TrainingBatch{}), modelstore.ErrInvalidBatch)
	require.ErrorIs(t, store.Import(ctx, nil, nil), modelstore.ErrInvalidBatch)
	require.ErrorIs(t, store.Apply(ctx, sqliteBatch(
		[]modelstore.Class{{ID: 2, TypeTag: 2, Payload: []byte("x")}},
		modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: 0},
	)), modelstore.ErrInvalidBatch)
	require.Error(t, store.Apply(ctx, sqliteBatch(
		nil,
		modelstore.TransitionDelta{FromID: 1, ToID: 9, Count: 1},
	)))
	require.ErrorIs(t, store.Apply(ctx, sqliteBatch(
		[]modelstore.Class{
			{ID: 8, TypeTag: 2, Payload: []byte("first")},
			{ID: 8, TypeTag: 2, Payload: []byte("second")},
		},
	)), modelstore.ErrClassConflict)

	source := mapstore.New(9)
	require.NoError(t, source.Apply(ctx, sqliteBatch(
		[]modelstore.Class{{ID: 4, TypeTag: 2, Payload: []byte("class")}},
		modelstore.TransitionDelta{FromID: 3, ToID: 4, Count: 7},
	)))
	classes, err := source.Classes(ctx)
	require.NoError(t, err)
	require.NoError(t, store.Import(ctx, classes, source))
	stats, err := store.Stats(ctx, 3)
	require.NoError(t, err)
	require.Equal(t, int64(7), stats.Total)

	require.ErrorIs(
		t,
		store.ExportTransitions(ctx, func(modelstore.TransitionCount) error { return errTestSinkFailed }),
		errTestSinkFailed,
	)
	require.ErrorIs(t, store.ExportTransitions(ctx, nil), modelstore.ErrInvalidBatch)

	canceled, cancel := context.WithCancel(ctx)
	cancel()

	_, err = store.Classes(canceled)
	require.ErrorIs(t, err, context.Canceled)
	_, err = store.Stats(canceled, 0)
	require.ErrorIs(t, err, context.Canceled)
	require.ErrorIs(t, store.Reset(canceled), context.Canceled)
	require.ErrorIs(t, store.Apply(canceled, sqliteBatch(nil)), context.Canceled)
}

func TestStore_overflowAndPoisonedStates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "overflow.db")
	store, err := Create(ctx, path, testMetadata(), OpenConfig{})
	require.NoError(t, err)

	require.NoError(t, store.Apply(ctx, sqliteBatch(
		[]modelstore.Class{{ID: 2, TypeTag: 2, Payload: []byte("x")}},
		modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: 1},
	)))

	for _, statement := range []string{
		"UPDATE metadata SET total_count = 9223372036854775807",
		"UPDATE from_a SET count = 9223372036854775807",
		"UPDATE to_b SET count = 9223372036854775807",
		"UPDATE from_a_to_b SET count = 9223372036854775807",
	} {
		_, err = store.conn.ExecContext(ctx, statement)
		require.NoError(t, err)
	}

	err = store.Apply(ctx, sqliteBatch(
		[]modelstore.Class{{ID: 2, TypeTag: 2, Payload: []byte("x")}},
		modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: 1},
	))
	require.ErrorIs(t, err, modelstore.ErrCountOverflow)

	store.poisoned = true
	require.ErrorIs(t, store.Apply(ctx, sqliteBatch(nil)), modelstore.ErrPoisoned)
	_, err = store.Classes(ctx)
	require.ErrorIs(t, err, modelstore.ErrPoisoned)
	_, err = store.Stats(ctx, 0)
	require.ErrorIs(t, err, modelstore.ErrPoisoned)
	require.ErrorIs(
		t,
		store.ExportTransitions(ctx, func(modelstore.TransitionCount) error { return nil }),
		modelstore.ErrPoisoned,
	)
	require.ErrorIs(t, store.Reset(ctx), modelstore.ErrPoisoned)
	require.NoError(t, store.Close())

	require.ErrorIs(t, store.Apply(ctx, sqliteBatch(nil)), modelstore.ErrClosed)
	_, err = store.Classes(ctx)
	require.ErrorIs(t, err, modelstore.ErrClosed)
	_, err = store.Stats(ctx, 0)
	require.ErrorIs(t, err, modelstore.ErrClosed)
	require.ErrorIs(
		t,
		store.ExportTransitions(ctx, func(modelstore.TransitionCount) error { return nil }),
		modelstore.ErrClosed,
	)
	require.ErrorIs(t, store.Reset(ctx), modelstore.ErrClosed)
}

func TestStore_commitFailurePoisonsStore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tests := map[string]func(context.Context, *Store) error{
		"apply": func(ctx context.Context, store *Store) error {
			return store.Apply(ctx, sqliteBatch(
				[]modelstore.Class{{ID: 2, TypeTag: 2, Payload: []byte("x")}},
				modelstore.TransitionDelta{FromID: 1, ToID: 2, Count: 1},
			))
		},
		"reset": func(ctx context.Context, store *Store) error {
			return store.Reset(ctx)
		},
	}

	for name, operation := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			store, err := Create(
				ctx,
				filepath.Join(t.TempDir(), "commit.db"),
				testMetadata(),
				OpenConfig{},
			)
			require.NoError(t, err)

			store.commit = func(*sql.Tx) error { return errTestCommitFailed }

			err = operation(ctx, store)
			require.ErrorIs(t, err, modelstore.ErrCommitIndeterminate)
			require.ErrorIs(t, err, errTestCommitFailed)
			_, err = store.Stats(ctx, 0)
			require.ErrorIs(t, err, modelstore.ErrPoisoned)
			require.NoError(t, store.Close())
		})
	}
}

//nolint:funlen // path and lock lifecycle checks share the same temporary model.
func TestPathLockAndCanonicalPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	directory := t.TempDir()
	path := filepath.Join(directory, "model.db")

	canonical, err := CanonicalPath(path)
	require.NoError(t, err)
	require.NotEmpty(t, canonical)

	_, err = CanonicalPath("")
	require.ErrorIs(t, err, ErrInvalidModel)
	_, err = CanonicalPath(filepath.Join(directory, "missing", "model.db"))
	require.Error(t, err)

	lock, err := AcquirePathLock(ctx, path)
	require.NoError(t, err)
	_, err = AcquirePathLock(ctx, path)
	require.ErrorIs(t, err, ErrLocked)
	require.NoError(t, lock.Close())
	require.NoError(t, (*PathLock)(nil).Close())

	canceled, cancel := context.WithCancel(ctx)
	cancel()

	_, err = AcquirePathLock(canceled, path)
	require.ErrorIs(t, err, context.Canceled)

	store, err := Create(ctx, path, testMetadata(), OpenConfig{})
	require.NoError(t, err)
	_, err = Create(ctx, path, testMetadata(), OpenConfig{})
	require.ErrorIs(t, err, ErrLocked)
	active, err := IsOpenAlias(path)
	require.NoError(t, err)
	require.True(t, active)

	alias := filepath.Join(directory, "alias.db")
	require.NoError(t, os.Link(path, alias))
	active, err = IsOpenAlias(alias)
	require.NoError(t, err)
	require.True(t, active)
	require.NoError(t, store.Close())
	_, err = Create(ctx, path, testMetadata(), OpenConfig{})
	require.ErrorIs(t, err, ErrInvalidModel)

	active, err = IsOpenAlias(path)
	require.NoError(t, err)
	require.False(t, active)

	_, err = IsOpenAlias(filepath.Join(directory, "missing", "nested", "model.db"))
	require.Error(t, err)
}

func TestValidateRejectsCorruptSchemaAndMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tests := map[string]string{
		"application id": "PRAGMA application_id = 1",
		"schema version": "PRAGMA user_version = 99",
		"extra table":    "CREATE TABLE extra (id INTEGER) STRICT",
		"extra index":    "CREATE INDEX extra_index ON metadata(total_count)",
		"extra trigger":  "CREATE TRIGGER extra_trigger AFTER UPDATE ON metadata BEGIN SELECT 1; END",
		"extra view":     "CREATE VIEW extra_view AS SELECT * FROM metadata",
		"counts":         "UPDATE metadata SET total_count = 1",
	}

	for name, mutation := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "invalid.db")
			store, err := Create(ctx, path, testMetadata(), OpenConfig{Portable: true})
			require.NoError(t, err)
			require.NoError(t, store.Close())

			db, err := sql.Open("sqlite3", path)
			require.NoError(t, err)
			_, err = db.ExecContext(ctx, mutation)
			require.NoError(t, err)
			require.NoError(t, db.Close())

			_, err = Open(ctx, path, OpenConfig{Portable: true})
			require.ErrorIs(t, err, ErrInvalidModel)
		})
	}
}

func sqliteBatch(classes []modelstore.Class, deltas ...modelstore.TransitionDelta) modelstore.TrainingBatch {
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

func testMetadata() Metadata {
	return Metadata{
		CodecVersion: 1,
		HasherName:   "test",
		ItemProbe:    math.MaxUint64,
		ContextProbe: 2,
		ScopeID:      math.MaxUint64,
	}
}
