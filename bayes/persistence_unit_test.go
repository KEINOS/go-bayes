//go:build cgo

//nolint:exhaustruct // tests set only fields relevant to each case.
package bayes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

//nolint:funlen // permission, locking, and replacement checks share one model.
func TestSQLite_SavePreservesPermissionsAndLocksActiveModels(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	directory := t.TempDir()
	modelPath := filepath.Join(directory, "model.db")
	copyPath := filepath.Join(directory, "copy.db")

	memory, err := New(ctx, MemoryStorage, 77, WithHasher("blake3"))
	require.NoError(t, err)
	require.NoError(t, memory.Train(ctx, []string{"A", "B", "C", "D"}))
	require.NoError(t, memory.Save(ctx, modelPath))

	info, err := os.Stat(modelPath)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	require.NoFileExists(t, modelPath+"-wal")
	require.NoFileExists(t, modelPath+"-shm")

	loaded, err := Load(ctx, modelPath)
	require.NoError(t, err)
	require.Equal(t, MemoryStorage, loaded.storage)
	require.Equal(t, uint64(77), loaded.ID())
	assertPrediction(t, loaded, []string{"A", "B", "C"}, "D")

	direct, err := Open(ctx, modelPath)
	require.NoError(t, err)
	require.Equal(t, SQLiteStorage, direct.storage)

	_, err = Open(ctx, modelPath)
	require.ErrorIs(t, err, ErrModelLocked)
	require.ErrorIs(t, direct.Save(ctx, modelPath), ErrModelLocked)

	require.NoError(t, direct.Train(ctx, []string{"X", "Y"}))
	require.NoError(t, direct.Save(ctx, copyPath))
	require.NoError(t, direct.Close())

	reopened, err := Open(ctx, modelPath)
	require.NoError(t, err)
	assertPrediction(t, reopened, []string{"X"}, "Y")
	require.NoError(t, reopened.Close())

	copyModel, err := Load(ctx, copyPath)
	require.NoError(t, err)
	assertPrediction(t, copyModel, []string{"A", "B", "C"}, "D")
	assertPrediction(t, copyModel, []string{"X"}, "Y")

	require.NoError(t, os.Chmod(copyPath, 0o640)) // #nosec G302 -- permission preservation is under test.
	require.NoError(t, memory.Save(ctx, copyPath))
	info, err = os.Stat(copyPath)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o640), info.Mode().Perm())
}

func TestSQLite_NewAndResetAreDurable(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "direct.db")
	direct, err := New(
		ctx,
		SQLiteStorage,
		11,
		WithSQLitePath(path),
		WithSQLiteCacheKiB(512),
		WithSQLiteSynchronous(SQLiteSynchronousNormal),
	)
	require.NoError(t, err)
	require.NoError(t, direct.Train(ctx, []int{1, 2, 3}))
	require.NoError(t, direct.Close())

	_, err = New(ctx, SQLiteStorage, 11, WithSQLitePath(path))
	require.ErrorIs(t, err, errStorageConfig)

	direct, err = Open(ctx, path)
	require.NoError(t, err)
	assertPrediction(t, direct, []int{1, 2}, 3)
	require.NoError(t, direct.Reset(ctx))
	require.NoError(t, direct.Close())

	loaded, err := Load(ctx, path)
	require.NoError(t, err)
	id, err := loaded.Predict(ctx, []int{1, 2})
	require.NoError(t, err)
	require.Zero(t, id)
	require.Empty(t, loaded.classes)
}

func TestSQLite_customHasherMustBeInjected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "custom.db")
	hasher := stableTestHasher{name: "custom-v1"}
	predictor, err := NewPredictor(ctx, PredictorConfig{
		Storage: MemoryStorage, ScopeID: 5, Hasher: hasher,
	})
	require.NoError(t, err)
	require.NoError(t, predictor.Train(ctx, []string{"A", "B"}))
	require.NoError(t, predictor.Save(ctx, path))

	_, err = Load(ctx, path)
	require.ErrorIs(t, err, ErrInvalidModel)
	loaded, err := Load(ctx, path, func(config *PredictorConfig) error {
		config.Hasher = hasher

		return nil
	})
	require.NoError(t, err)
	assertPrediction(t, loaded, []string{"A"}, "B")

	_, err = Load(ctx, path, func(config *PredictorConfig) error {
		config.Hasher = stableTestHasher{name: "custom-v1", salt: 1}

		return nil
	})
	require.ErrorIs(t, err, ErrInvalidModel)
}

func TestSQLite_LoadPreservesSupportedClassTypes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	values := []any{
		false,
		"\x00invalid-utf8-\xff",
		int(-1),
		int16(math.MinInt16),
		int32(math.MinInt32),
		int64(math.MinInt64),
		uint(1),
		uint16(math.MaxUint16),
		uint32(math.MaxUint32),
		uint64(math.MaxUint64),
		math.Float32frombits(0x7fc00001),
		math.Float64frombits(0x8000000000000000),
	}

	predictor, err := New(ctx, MemoryStorage, 17)
	require.NoError(t, err)

	for index, value := range values {
		require.NoError(t, predictor.Train(ctx, []any{fmt.Sprintf("input-%d", index), value}))
	}

	path := filepath.Join(t.TempDir(), "types.db")
	require.NoError(t, predictor.Save(ctx, path))
	restored, err := Load(ctx, path)
	require.NoError(t, err)

	for _, want := range values {
		id, idErr := restored.itemID(want)
		require.NoError(t, idErr)
		got := restored.GetClass(id)
		require.Equal(t, reflect.TypeOf(want), reflect.TypeOf(got))

		switch typed := want.(type) {
		case float32:
			gotFloat, ok := got.(float32)
			require.True(t, ok)
			require.Equal(t, math.Float32bits(typed), math.Float32bits(gotFloat))
		case float64:
			gotFloat, ok := got.(float64)
			require.True(t, ok)
			require.Equal(t, math.Float64bits(typed), math.Float64bits(gotFloat))
		default:
			require.Equal(t, want, got)
		}
	}
}

func TestSQLite_rejectsInvalidFilesAndOptions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	directory := t.TempDir()
	textPath := filepath.Join(directory, "not-model.db")
	require.NoError(t, os.WriteFile(textPath, []byte("not sqlite"), 0o600))
	_, err := Load(ctx, textPath)
	require.Error(t, err)

	_, err = New(ctx, SQLiteStorage, 1)
	require.ErrorIs(t, err, errStorageConfig)
	_, err = Open(ctx, textPath, WithSQLitePath("ignored.db"))
	require.ErrorIs(t, err, errStorageConfig)
	_, err = Open(ctx, textPath, WithModelStore(&fakeStore{}))
	require.ErrorIs(t, err, errStorageConfig)
	_, err = Load(ctx, textPath, WithSQLiteCacheKiB(1))
	require.ErrorIs(t, err, errStorageConfig)
	_, err = Load(ctx, textPath, WithSQLiteSynchronous(SQLiteSynchronousNormal))
	require.ErrorIs(t, err, errStorageConfig)

	validPath := filepath.Join(directory, "version.db")
	predictor, err := New(ctx, MemoryStorage, 1)
	require.NoError(t, err)
	require.NoError(t, predictor.Train(ctx, []string{"A", "B"}))
	require.NoError(t, predictor.Save(ctx, validPath))

	db, err := sql.Open("sqlite3", validPath)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "PRAGMA user_version = 99")
	require.NoError(t, err)
	require.NoError(t, db.Close())

	_, err = Load(ctx, validPath)
	require.ErrorIs(t, err, ErrInvalidModel)

	failedOption := func(*PredictorConfig) error { return errTestOptionFailed }
	_, err = Load(ctx, validPath, failedOption)
	require.ErrorIs(t, err, errTestOptionFailed)
	_, err = Open(ctx, validPath, failedOption)
	require.ErrorIs(t, err, errTestOptionFailed)
}

func TestSQLite_SaveErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	predictor, err := New(ctx, MemoryStorage, 1)
	require.NoError(t, err)
	require.Error(t, predictor.Save(ctx, ""))
	require.Error(t, predictor.Save(ctx, filepath.Join(t.TempDir(), "missing", "model.db")))

	uninitialized := &Predictor{}
	require.ErrorIs(t, saveModel(ctx, uninitialized, "model.db"), errPredictorNotInitialized)

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	require.ErrorIs(t, predictor.Save(canceled, filepath.Join(t.TempDir(), "model.db")), context.Canceled)

	classesFailure := &Predictor{
		hasher: NewDefaultHasher(),
		store:  &fakeStore{scope: 1, classesErr: errTestClassesFailed},
	}
	require.ErrorIs(
		t,
		saveModel(ctx, classesFailure, filepath.Join(t.TempDir(), "classes.db")),
		errTestClassesFailed,
	)

	exportFailure := &Predictor{
		hasher: NewDefaultHasher(),
		store:  &fakeStore{scope: 1, exportErr: errTestExportFailed},
	}
	require.ErrorIs(
		t,
		saveModel(ctx, exportFailure, filepath.Join(t.TempDir(), "export.db")),
		errTestExportFailed,
	)
}

func assertPrediction(t *testing.T, predictor *Predictor, input any, want any) {
	t.Helper()

	id, err := predictor.Predict(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, want, predictor.GetClass(id))
}

type stableTestHasher struct {
	name string
	salt uint64
}

func (h stableTestHasher) Hash(data []byte) uint64 {
	result := h.salt
	for _, value := range data {
		result = result*1099511628211 + uint64(value)
	}

	return result
}

func (h stableTestHasher) Name() string { return h.name }

var _ Hasher = stableTestHasher{}
var _ = errors.Is
