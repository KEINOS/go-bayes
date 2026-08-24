//nolint:exhaustruct // tests set only fields relevant to each case.
package bayes

import (
	"context"
	"encoding/json"
	"errors"
	"iter"
	"math"
	"reflect"
	"testing"

	"github.com/KEINOS/go-bayes/bayes/internal/modelstores/mapstore"
	"github.com/KEINOS/go-bayes/bayes/modelstore"
	"github.com/stretchr/testify/require"
)

var (
	errTestApplyFailed   = errors.New("apply failed")
	errTestClassesFailed = errors.New("classes failed")
	errTestCloseFailed   = errors.New("close failed")
	errTestExportFailed  = errors.New("export failed")
	errTestOptionFailed  = errors.New("option failed")
	errTestResetFailed   = errors.New("reset failed")
	errTestStatsFailed   = errors.New("stats failed")
)

type fakeStore struct {
	scope      uint64
	applyErr   error
	stats      modelstore.Stats
	statsErr   error
	statsCalls int
	classes    []modelstore.Class
	classesErr error
	exportErr  error
	resetErr   error
	closeErr   error
	closed     bool
	lastBatch  modelstore.TrainingBatch
}

func (s *fakeStore) Apply(_ context.Context, batch modelstore.TrainingBatch) error {
	s.lastBatch = batch

	return s.applyErr
}

func (s *fakeStore) Classes(context.Context) ([]modelstore.Class, error) {
	return s.classes, s.classesErr
}

func (s *fakeStore) Close() error {
	s.closed = true

	return s.closeErr
}

func (s *fakeStore) ExportTransitions(context.Context, func(modelstore.TransitionCount) error) error {
	return s.exportErr
}

func (s *fakeStore) Reset(context.Context) error {
	return s.resetErr
}

func (s *fakeStore) ScopeID() uint64 {
	return s.scope
}

func (s *fakeStore) Stats(context.Context, uint64) (modelstore.Stats, error) {
	s.statsCalls++

	return s.stats, s.statsErr
}

func TestPredictor_memoryGoldenPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	predictor, err := New(ctx, MemoryStorage, 100)
	require.NoError(t, err)
	require.Equal(t, uint64(100), predictor.ID())

	require.NoError(t, predictor.Train(ctx, []string{"A", "B", "C", "D"}))
	predicted, err := predictor.Predict(ctx, []string{"A", "B", "C"})
	require.NoError(t, err)
	require.Equal(t, "D", predictor.GetClass(predicted))

	// Training records every suffix context.
	for _, input := range [][]string{{"A", "B", "C"}, {"B", "C"}, {"C"}} {
		predicted, err = predictor.Predict(ctx, input)
		require.NoError(t, err)
		require.Equal(t, "D", predictor.GetClass(predicted))
	}

	require.NoError(t, predictor.Reset(ctx))
	predicted, err = predictor.Predict(ctx, []string{"A"})
	require.NoError(t, err)
	require.Zero(t, predicted)
	require.Nil(t, predictor.GetClass(predicted))

	require.NoError(t, predictor.Close())
	require.NoError(t, predictor.Close())
	require.ErrorIs(t, predictor.Train(ctx, []string{"A", "B"}), ErrPredictorClosed)
	_, err = predictor.Predict(ctx, []string{"A"})
	require.ErrorIs(t, err, ErrPredictorClosed)
	require.ErrorIs(t, predictor.Reset(ctx), ErrPredictorClosed)
	require.ErrorIs(t, predictor.Save(ctx, "unused.db"), ErrPredictorClosed)
	_, err = predictor.HashTrans("A")
	require.NoError(t, err)
}

func TestPredictor_supportedClassTypes(t *testing.T) {
	t.Parallel()

	values := []any{
		false,
		true,
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

	for index, value := range values {
		ctx := context.Background()
		predictor, err := New(ctx, MemoryStorage, uint64(index))
		require.NoError(t, err)
		require.NoError(t, predictor.Train(ctx, []any{"input", value}))

		id, err := predictor.itemID(value)
		require.NoError(t, err)

		got := predictor.GetClass(id)
		require.Equal(t, reflect.TypeOf(value), reflect.TypeOf(got))

		switch typed := value.(type) {
		case float32:
			gotFloat, ok := got.(float32)
			require.True(t, ok)
			require.Equal(t, math.Float32bits(typed), math.Float32bits(gotFloat))
		case float64:
			gotFloat, ok := got.(float64)
			require.True(t, ok)
			require.Equal(t, math.Float64bits(typed), math.Float64bits(gotFloat))
		default:
			require.Equal(t, value, got)
		}
	}
}

func TestPredictor_predictReadsStatsOnceAndBreaksTiesByID(t *testing.T) {
	t.Parallel()

	store := &fakeStore{
		scope: 1,
		stats: modelstore.Stats{
			Total:     2,
			FromCount: 2,
			Candidates: []modelstore.CandidateStats{
				{ClassID: 9, ToCount: 1, PairCount: 1},
				{ClassID: 3, ToCount: 1, PairCount: 1},
			},
		},
	}
	predictor, err := NewPredictor(context.Background(), PredictorConfig{
		Storage: UnknownStorage, ScopeID: 1, ModelStore: store,
	})
	require.NoError(t, err)

	id, err := predictor.Predict(context.Background(), []string{"A"})
	require.NoError(t, err)
	require.Equal(t, uint64(3), id)
	require.Equal(t, 1, store.statsCalls)
}

func TestPredictor_trainFailureDoesNotChangeClassCache(t *testing.T) {
	t.Parallel()

	store := &fakeStore{scope: 1, applyErr: errTestApplyFailed}
	predictor, err := NewPredictor(context.Background(), PredictorConfig{
		Storage: UnknownStorage, ScopeID: 1, ModelStore: store,
	})
	require.NoError(t, err)

	err = predictor.Train(context.Background(), []string{"A", "B"})
	require.ErrorIs(t, err, errTestApplyFailed)
	require.Empty(t, predictor.classes)

	var deltas []modelstore.TransitionDelta
	for delta := range store.lastBatch.Transitions() {
		deltas = append(deltas, delta)
	}

	require.Len(t, deltas, 1)

	yielded := 0
	store.lastBatch.Transitions()(func(modelstore.TransitionDelta) bool {
		yielded++

		return false
	})
	require.Equal(t, 1, yielded, "the iterator must stop when its consumer stops")
}

func TestPredictor_propagatesStoreErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	classesStore := &fakeStore{scope: 1, classesErr: errTestClassesFailed}
	_, err := NewPredictor(ctx, PredictorConfig{
		Storage: UnknownStorage, ScopeID: 1, ModelStore: classesStore,
	})
	require.ErrorIs(t, err, errTestClassesFailed)
	require.False(t, classesStore.closed, "a rejected injected store remains caller-owned")

	invalidClassStore := &fakeStore{
		scope:   1,
		classes: []modelstore.Class{{TypeTag: 255}},
	}
	_, err = NewPredictor(ctx, PredictorConfig{
		Storage: UnknownStorage, ScopeID: 1, ModelStore: invalidClassStore,
	})
	require.ErrorIs(t, err, errInvalidStoredClass)
	require.False(t, invalidClassStore.closed, "a rejected injected store remains caller-owned")

	store := &fakeStore{
		scope:    1,
		statsErr: errTestStatsFailed,
		resetErr: errTestResetFailed,
		closeErr: errTestCloseFailed,
	}
	predictor, err := NewPredictor(ctx, PredictorConfig{
		Storage: UnknownStorage, ScopeID: 1, ModelStore: store,
	})
	require.NoError(t, err)

	_, err = predictor.Predict(ctx, []string{"A"})
	require.ErrorIs(t, err, errTestStatsFailed)
	require.ErrorIs(t, predictor.Reset(ctx), errTestResetFailed)
	require.ErrorIs(t, predictor.Close(), errTestCloseFailed)
	require.NoError(t, predictor.Close(), "Close remains idempotent after a store error")
}

func TestPredictor_errorPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	predictor, err := New(ctx, MemoryStorage, 1)
	require.NoError(t, err)

	for _, items := range []any{nil, "not-a-slice"} {
		require.Error(t, predictor.Train(ctx, items))
		_, err = predictor.Predict(ctx, items)
		require.Error(t, err)
	}

	require.Error(t, predictor.Train(ctx, []any{make(chan int)}))
	_, err = predictor.HashTrans(make(chan int))
	require.ErrorIs(t, err, errUnsupportedValueType)
	_, err = predictor.Predict(ctx, []any{make(chan int)})
	require.ErrorIs(t, err, errUnsupportedValueType)

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	require.ErrorIs(t, predictor.Train(canceled, []string{"A", "B"}), context.Canceled)

	predictor.hasher = nil
	_, err = predictor.HashTrans("A")
	require.ErrorIs(t, err, errPredictorHasherUninit)
	require.ErrorIs(t, predictor.Train(ctx, []string{"A"}), errPredictorHasherUninit)

	uninitialized := &Predictor{scopeID: 9, classes: make(map[uint64]classEntry)}
	require.Equal(t, uint64(9), uninitialized.ID())
	_, err = uninitialized.Predict(ctx, []string{"A"})
	require.ErrorIs(t, err, errPredictorNotInitialized)
	require.ErrorIs(t, uninitialized.Train(ctx, []string{"A"}), errPredictorNotInitialized)
	require.ErrorIs(t, uninitialized.Reset(ctx), errPredictorNotInitialized)
	require.NoError(t, uninitialized.Close())
}

func TestPredictor_JSONGuards(t *testing.T) {
	t.Parallel()

	predictor, err := New(context.Background(), MemoryStorage, 1)
	require.NoError(t, err)

	_, err = json.Marshal(predictor)
	require.ErrorIs(t, err, ErrJSONPersistenceUnsupported)
	require.ErrorIs(t, json.Unmarshal([]byte(`{}`), predictor), ErrJSONPersistenceUnsupported)
	require.Empty(t, predictor.classes)
}

func TestPredictor_TrainRejectsHashCollisionAtomically(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	predictor, err := NewPredictor(ctx, PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 1,
		Hasher:  stubHasher{name: "length"},
	})
	require.NoError(t, err)

	err = predictor.Train(ctx, []string{"A", "B", "C"})
	require.ErrorIs(t, err, ErrHashCollision)
	require.Nil(t, predictor.GetClass(predictor.hasher.Hash([]byte{itemDomain, tagString, 'B'})))

	require.NoError(t, predictor.Train(ctx, []string{"A", "B"}))
	err = predictor.Train(ctx, []string{"X", "C"})
	require.ErrorIs(t, err, ErrHashCollision)
}

func TestConstructorsAndOptions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	_, err := New(ctx, MemoryStorage, 1, nil)
	require.ErrorIs(t, err, errNewOptionNil)
	failedOption := func(*PredictorConfig) error { return errTestOptionFailed }
	_, err = New(ctx, MemoryStorage, 1, failedOption)
	require.ErrorIs(t, err, errTestOptionFailed)
	_, err = New(ctx, MemoryStorage, 1, WithHasher("unknown"))
	require.ErrorIs(t, err, errUnknownHasher)
	selected, err := New(ctx, MemoryStorage, 1, WithHasher("xxhash3"))
	require.NoError(t, err)
	require.Equal(t, "xxhash3", selected.hasher.Name())
	require.NoError(t, selected.Close())
	_, err = New(ctx, MemoryStorage, 1, WithSQLitePath(""))
	require.ErrorIs(t, err, errStorageConfig)
	_, err = New(ctx, MemoryStorage, 1, WithSQLitePath("model.db"))
	require.ErrorIs(t, err, errStorageConfig)
	_, err = New(ctx, MemoryStorage, 1, WithSQLiteCacheKiB(0))
	require.ErrorIs(t, err, errStorageConfig)
	_, err = New(ctx, MemoryStorage, 1, WithSQLiteSynchronous(SQLiteSynchronous(99)))
	require.ErrorIs(t, err, errStorageConfig)
	_, err = New(ctx, UnknownStorage, 1)
	require.ErrorIs(t, err, errStorageConfig)
	_, err = New(ctx, Storage(99), 1)
	require.ErrorIs(t, err, errStorageConfig)
	_, err = New(ctx, UnknownStorage, 1, WithModelStore(nil))
	require.ErrorIs(t, err, errStorageConfig)

	emptyName := stubHasher{name: ""}
	_, err = NewPredictor(ctx, PredictorConfig{Storage: MemoryStorage, Hasher: emptyName})
	require.ErrorIs(t, err, errPredictorHasherNameEmpty)

	canceled, cancel := context.WithCancel(ctx)
	cancel()

	_, err = New(canceled, MemoryStorage, 1)
	require.ErrorIs(t, err, context.Canceled)
}

func TestStorageTypes(t *testing.T) {
	t.Parallel()

	require.Equal(t, "unknown", UnknownStorage.Type())
	require.Equal(t, "in-memory", MemoryStorage.Type())
	require.Equal(t, "sqlite", SQLiteStorage.Type())
	require.Equal(t, "unknown", Storage(99).Type())
}

func TestConstructorRejectsInvalidInjectedStoreConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := mapstore.New(2)
	_, err := New(ctx, UnknownStorage, 1, WithModelStore(store))
	require.ErrorIs(t, err, errStorageConfig)
	require.NoError(t, store.Reset(ctx), "constructor failure must not take ownership")

	injected := &fakeStore{scope: 1}
	_, err = New(ctx, MemoryStorage, 1, WithModelStore(injected))
	require.ErrorIs(t, err, errStorageConfig)
	_, err = NewPredictor(ctx, PredictorConfig{
		Storage: UnknownStorage, ScopeID: 1, ModelStore: injected, SQLitePath: "model.db",
	})
	require.ErrorIs(t, err, errStorageConfig)
}

func TestClassDecoderRejectsMalformedRecords(t *testing.T) {
	t.Parallel()

	predictor, err := New(context.Background(), MemoryStorage, 1)
	require.NoError(t, err)

	tests := []modelstore.Class{
		{TypeTag: tagBool, Payload: nil},
		{TypeTag: tagBool, Payload: []byte{2}},
		{TypeTag: tagInt, Payload: []byte{1}},
		{TypeTag: tagInt16, Payload: []byte{0, 0, 0, 0, 0, 0, 0x80, 0}},
		{TypeTag: tagInt32, Payload: []byte{0, 0, 0, 0, 0x80, 0, 0, 0}},
		{TypeTag: tagUint16, Payload: []byte{0, 0, 0, 0, 0, 1, 0, 0}},
		{TypeTag: tagUint32, Payload: []byte{0, 0, 0, 1, 0, 0, 0, 0}},
		{TypeTag: tagFloat32, Payload: []byte{1}},
		{TypeTag: 255, Payload: nil},
	}
	for _, test := range tests {
		_, err = predictor.decodeClass(test)
		require.ErrorIs(t, err, errInvalidStoredClass)
	}

	valid, err := predictor.classRecord(0, "value")
	require.NoError(t, err)
	_, err = predictor.decodeClass(valid)
	require.ErrorIs(t, err, errInvalidStoredClass, "ID mismatch must fail")

	_, err = predictor.classRecord(0, make(chan int))
	require.ErrorIs(t, err, errUnsupportedValueType)
}

type stubHasher struct{ name string }

func (s stubHasher) Hash(data []byte) uint64 { return uint64(len(data)) }
func (s stubHasher) Name() string            { return s.name }

var _ Hasher = stubHasher{}
var _ modelstore.ModelStore = (*fakeStore)(nil)
var _ modelstore.TransitionIterator = func() iter.Seq[modelstore.TransitionDelta] { return nil }
