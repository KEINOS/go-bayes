package bayes

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------------
//  Helper types and functions for testing
// ----------------------------------------------------------------------------

type stubHasher struct {
	got  [][]byte
	name string
	out  uint64
}

func (s *stubHasher) Hash(data []byte) uint64 {
	s.got = append(s.got, append([]byte(nil), data...))

	return s.out
}

func (s *stubHasher) Name() string { return s.name }

type namedHasher struct {
	name string
	base Hasher
}

func (h *namedHasher) Hash(data []byte) uint64 { return h.base.Hash(data) }

func (h *namedHasher) Name() string { return h.name }

type stubNodeLogger struct{}

func (stubNodeLogger) ID() uint64 { return 0 }

func (stubNodeLogger) Predict(_, _ uint64) float64 { return 0 }

func (stubNodeLogger) PriorProbTo(uint64) float64 { return 0 }

func (stubNodeLogger) PriorProbFromTo(_, _ uint64) float64 { return 0 }

func (stubNodeLogger) PriorProbNotFromTo(_, _ uint64) float64 { return 0 }

func (stubNodeLogger) Update(_, _ uint64) {}

// ----------------------------------------------------------------------------
//  NewPredictor
// ----------------------------------------------------------------------------

func TestNewPredictor_success(t *testing.T) {
	t.Parallel()

	instance, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  nil,
	})

	require.NoError(t, err)
	require.NotNil(t, instance)
	require.NotNil(t, instance.predictor)
	require.NotNil(t, instance.classes)
	require.NotNil(t, instance.hasher)
}

func TestNewPredictor_unknownStorage(t *testing.T) {
	t.Parallel()

	instance, err := NewPredictor(PredictorConfig{
		Storage: UnknownStorage,
		ScopeID: 0,
		Hasher:  nil,
	})

	require.Error(t, err)
	require.Nil(t, instance)
	require.ErrorContains(t, err, "failed to create predictor")
}

func TestNewPredictor_rejectsEmptyHasherName(t *testing.T) {
	t.Parallel()

	instance, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  &stubHasher{got: nil, name: "", out: 0},
	})

	require.ErrorIs(t, err, errPredictorHasherNameEmpty)
	require.Nil(t, instance)
}

// ----------------------------------------------------------------------------
//  Predictor
// ----------------------------------------------------------------------------

func TestPredictor_HashTrans_errorCases(t *testing.T) {
	t.Parallel()

	instance, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  nil,
	})
	require.NoError(t, err)

	_, err = instance.HashTrans(big.NewInt(1))
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to encode transition")
}

func TestPredictor_Predict_notInitialized(t *testing.T) {
	t.Parallel()

	instance := &Predictor{
		predictor: nil,
		classes:   nil,
		storage:   UnknownStorage,
		scopeID:   0,
		hasher:    nil,
		scratch:   codecScratch{bytes: nil, ids: nil},
	}

	actual, err := instance.Predict([]any{uint64(1)})
	require.Error(t, err)
	require.Zero(t, actual)
	require.ErrorContains(t, err, "predictor is not initialized")
}

func TestPredictor_TrainPredictReset_flow(t *testing.T) {
	t.Parallel()

	instance, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  nil,
	})
	require.NoError(t, err)

	require.NoError(t, instance.Train([]any{"a", "b", "a", "b"}))

	predicted, err := instance.Predict([]any{"a"})
	require.NoError(t, err)
	require.NotZero(t, predicted)
	require.Equal(t, "b", instance.GetClass(predicted))

	instance.SetStorage(UnknownStorage)
	err = instance.Reset()
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to reset predictor")

	instance.SetStorage(MemoryStorage)
	require.NoError(t, instance.Reset())
	require.Empty(t, instance.classes)
}

func TestPredictor_Train_errorCases(t *testing.T) {
	t.Parallel()

	instance := &Predictor{
		predictor: nil,
		storage:   UnknownStorage,
		scopeID:   0,
		hasher:    NewDefaultHasher(),
		classes:   map[uint64]classEntry{},
		scratch:   codecScratch{bytes: nil, ids: nil},
	}

	err := instance.Train([]any{"a", "b"})
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to initialize predictor")

	instance = &Predictor{
		predictor: nil,
		storage:   MemoryStorage,
		scopeID:   0,
		hasher:    nil,
		classes:   map[uint64]classEntry{},
		scratch:   codecScratch{bytes: nil, ids: nil},
	}

	err = instance.Train([]any{"a", "b"})
	require.Error(t, err)
	require.ErrorContains(t, err, "hasher is not initialized")

	instance, err = NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  nil,
	})
	require.NoError(t, err)

	err = instance.Train([]any{big.NewInt(1)})
	require.Error(t, err)
	require.ErrorContains(t, err, "failed during training iteration")

	err = instance.Train(nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "items must not be nil")

	err = instance.Train("not-a-slice")
	require.Error(t, err)
	require.ErrorContains(t, err, "items must be a slice")
}

func TestPredictor_ID_fallsBackToScopeID(t *testing.T) {
	t.Parallel()

	instance := &Predictor{
		predictor: nil,
		classes:   nil,
		storage:   UnknownStorage,
		scopeID:   77,
		hasher:    nil,
		scratch:   codecScratch{bytes: nil, ids: nil},
	}

	require.Equal(t, uint64(77), instance.ID())
}

func TestPredictor_instancesAreIsolated(t *testing.T) {
	t.Parallel()

	left, err := NewPredictor(PredictorConfig{Storage: MemoryStorage, ScopeID: 1, Hasher: nil})
	require.NoError(t, err)

	right, err := NewPredictor(PredictorConfig{Storage: MemoryStorage, ScopeID: 2, Hasher: nil})
	require.NoError(t, err)

	require.NoError(t, left.Train([]any{"x", "y"}))
	require.NoError(t, right.Train([]any{"u", "v"}))

	leftPred, err := left.Predict([]any{"x"})
	require.NoError(t, err)
	rightPred, err := right.Predict([]any{"u"})
	require.NoError(t, err)

	require.Equal(t, "y", left.GetClass(leftPred))
	require.Equal(t, "v", right.GetClass(rightPred))
}

func TestPredictor_addClass_defaultBranch(t *testing.T) {
	t.Parallel()

	instance := &Predictor{
		predictor: nil,
		classes:   map[uint64]classEntry{},
		storage:   UnknownStorage,
		scopeID:   0,
		hasher:    nil,
		scratch:   codecScratch{bytes: nil, ids: nil},
	}

	type custom struct{ Value string }
	instance.addClass(1, custom{Value: "x"})

	stored, ok := instance.classes[1]
	require.True(t, ok)
	require.Equal(t, custom{Value: "x"}, stored.Raw)
}

func TestPredictor_addClass_allSupportedTypes(t *testing.T) {
	t.Parallel()

	instance := &Predictor{
		predictor: nil,
		classes:   map[uint64]classEntry{},
		storage:   UnknownStorage,
		scopeID:   0,
		hasher:    nil,
		scratch:   codecScratch{bytes: nil, ids: nil},
	}

	inputs := []any{
		uint64(1),
		uint32(2),
		uint16(3),
		uint(4),
		int64(5),
		int32(6),
		int16(7),
		int(8),
		float64(9),
		float32(10),
		"foo",
		true,
	}

	for index, input := range inputs {
		classID := uint64(index + 1)
		instance.addClass(classID, input)

		stored, ok := instance.classes[classID]
		require.True(t, ok)
		require.Equal(t, input, stored.Raw)
		require.Equal(t, classID, stored.ID)
	}
}

func TestPredictor_SetStorage(t *testing.T) {
	t.Parallel()

	instance := &Predictor{
		predictor: nil,
		classes:   nil,
		storage:   UnknownStorage,
		scopeID:   0,
		hasher:    nil,
		scratch:   codecScratch{bytes: nil, ids: nil},
	}

	instance.SetStorage(MemoryStorage)
	require.Equal(t, MemoryStorage, instance.storage)
}

func TestPredictor_HashTrans_usesInjectedHasher(t *testing.T) {
	t.Parallel()

	stub := &stubHasher{got: nil, name: "stub", out: 0x99}
	instance, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  stub,
	})
	require.NoError(t, err)

	actual, err := instance.HashTrans("a", "b")
	require.NoError(t, err)
	require.Equal(t, uint64(0x99), actual)
	require.Len(t, stub.got, 3)
	require.Equal(t, []byte{itemDomain, tagString, 'a'}, stub.got[0])
	require.Equal(t, []byte{itemDomain, tagString, 'b'}, stub.got[1])

	wantContext := make([]byte, 0, 18)
	wantContext = append(wantContext, contextDomain, 0x02)
	wantContext = append(wantContext, eightBytes(0x99)...)
	wantContext = append(wantContext, eightBytes(0x99)...)
	require.Equal(t, wantContext, stub.got[2])
}

func TestPredictor_NewPredictor_usesConfigHasher(t *testing.T) {
	t.Parallel()

	stub := &stubHasher{got: nil, name: "stub", out: 0x55}

	instance, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  stub,
	})
	require.NoError(t, err)

	actual, err := instance.HashTrans("x")
	require.NoError(t, err)
	require.Equal(t, uint64(0x55), actual)
}

func TestPredictor_HashTrans_nilHasher(t *testing.T) {
	t.Parallel()

	instance, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  nil,
	})
	require.NoError(t, err)

	instance.hasher = nil

	_, err = instance.HashTrans(uint64(1))
	require.Error(t, err)
	require.ErrorContains(t, err, "hasher is not initialized")
}

func TestPredictor_Predict_hashFlowError(t *testing.T) {
	t.Parallel()

	instance, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  nil,
	})
	require.NoError(t, err)

	_, err = instance.Predict([]any{big.NewInt(1)})
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to hash the flow")

	_, err = instance.Predict(nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "items must not be nil")

	_, err = instance.Predict("not-a-slice")
	require.Error(t, err)
	require.ErrorContains(t, err, "items must be a slice")
}

func TestPredictor_Predict_noClass(t *testing.T) {
	t.Parallel()

	instance, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  nil,
	})
	require.NoError(t, err)

	actual, err := instance.Predict([]any{"x"})
	require.NoError(t, err)
	require.Zero(t, actual)
}

func TestPredictor_Train_nilClasses(t *testing.T) {
	t.Parallel()

	instance, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  nil,
	})
	require.NoError(t, err)

	instance.classes = nil

	err = instance.Train([]any{"a", "b"})
	require.NoError(t, err)
	require.NotNil(t, instance.classes)
}

func TestPredictor_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	for name, hasher := range map[string]Hasher{
		"xxhash3": NewXXHash3Hasher(),
		"blake3":  NewBlake3Hasher(),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			original, err := NewPredictor(PredictorConfig{
				Storage: MemoryStorage,
				ScopeID: 123,
				Hasher:  hasher,
			})
			require.NoError(t, err)
			require.NoError(t, original.Train([]any{"a", "b", "a", "b"}))

			raw, err := json.Marshal(original)
			require.NoError(t, err)
			require.Contains(t, string(raw), `"schemaVersion":1`)
			require.Contains(t, string(raw), `"hasher":"`+name+`"`)

			var restored Predictor
			require.NoError(t, json.Unmarshal(raw, &restored))
			require.Equal(t, name, restored.hasher.Name())

			actual, err := restored.Predict([]any{"a"})
			require.NoError(t, err)
			require.Equal(t, "b", restored.GetClass(actual))
		})
	}
}

func TestPredictor_JSONCustomHasher(t *testing.T) {
	t.Parallel()

	const customHasherName = "example-v1"

	custom := &namedHasher{name: customHasherName, base: NewXXHash3Hasher()}
	original, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 456,
		Hasher:  custom,
	})
	require.NoError(t, err)
	require.NoError(t, original.Train([]string{"a", "b", "a", "b"}))

	raw, err := json.Marshal(original)
	require.NoError(t, err)

	restored, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  &namedHasher{name: customHasherName, base: NewXXHash3Hasher()},
	})
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, restored))
	require.Equal(t, customHasherName, restored.hasher.Name())

	actual, err := restored.Predict([]string{"a"})
	require.NoError(t, err)
	require.Equal(t, "b", restored.GetClass(actual))

	var missing Predictor

	err = json.Unmarshal(raw, &missing)
	require.ErrorIs(t, err, errPredictorHasherMismatch)

	mismatch, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  &namedHasher{name: "different-v1", base: NewXXHash3Hasher()},
	})
	require.NoError(t, err)
	err = json.Unmarshal(raw, mismatch)
	require.ErrorIs(t, err, errPredictorHasherMismatch)
}

func TestPredictor_UnmarshalJSON_rejectsIncompatibleSchemas(t *testing.T) {
	t.Parallel()

	for name, raw := range map[string]string{
		"missing schema": `{"hasher":"xxhash3","storage":1}`,
		"unknown schema": `{"schemaVersion":2,"hasher":"xxhash3","storage":1}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var predictor Predictor

			err := json.Unmarshal([]byte(raw), &predictor)
			require.ErrorIs(t, err, errPredictorSchemaUnsupported)
		})
	}
}

func TestPredictor_UnmarshalJSON_rejectsMissingHasher(t *testing.T) {
	t.Parallel()

	var predictor Predictor

	err := json.Unmarshal([]byte(`{"schemaVersion":1,"storage":1}`), &predictor)
	require.ErrorIs(t, err, errPredictorHasherMismatch)
}

func TestPredictor_Reset_preservesHasher(t *testing.T) {
	t.Parallel()

	hasher := &namedHasher{name: "example-v1", base: NewXXHash3Hasher()}
	predictor, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  hasher,
	})
	require.NoError(t, err)

	require.NoError(t, predictor.Train([]string{"a", "b"}))
	require.NoError(t, predictor.Reset())
	require.Same(t, hasher, predictor.hasher)
}

func TestPredictor_MarshalJSON_withoutPredictor(t *testing.T) {
	t.Parallel()

	instance := &Predictor{
		predictor: nil,
		classes:   map[uint64]classEntry{},
		storage:   MemoryStorage,
		scopeID:   1,
		hasher:    NewDefaultHasher(),
		scratch:   codecScratch{bytes: nil, ids: nil},
	}

	_, err := json.Marshal(instance)
	require.NoError(t, err)
}

func TestPredictor_MarshalJSON_errorCases(t *testing.T) {
	t.Parallel()

	withoutHasher := &Predictor{
		predictor: nil,
		classes:   nil,
		storage:   MemoryStorage,
		scopeID:   0,
		hasher:    nil,
		scratch:   codecScratch{bytes: nil, ids: nil},
	}
	_, err := json.Marshal(withoutHasher)
	require.ErrorIs(t, err, errPredictorHasherUninit)

	emptyHasherName := &Predictor{
		predictor: nil,
		classes:   nil,
		storage:   MemoryStorage,
		scopeID:   0,
		hasher:    &stubHasher{got: nil, name: "", out: 0},
		scratch:   codecScratch{bytes: nil, ids: nil},
	}
	_, err = json.Marshal(emptyHasherName)
	require.ErrorIs(t, err, errPredictorHasherNameEmpty)

	instance := &Predictor{
		predictor: nil,
		classes: map[uint64]classEntry{
			1: {ID: 1, Raw: func() {}},
		},
		storage: MemoryStorage,
		scopeID: 1,
		hasher:  NewDefaultHasher(),
		scratch: codecScratch{bytes: nil, ids: nil},
	}

	_, err = json.Marshal(instance)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to marshal predictor JSON")

	initialized, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 1,
		Hasher:  nil,
	})
	require.NoError(t, err)

	initialized.classes = map[uint64]classEntry{1: {ID: 1, Raw: func() {}}}

	_, err = json.Marshal(initialized)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to marshal predictor JSON")
}

func TestPredictor_MarshalJSON_unsupportedPredictorImplementation(t *testing.T) {
	t.Parallel()

	instance := &Predictor{
		predictor: stubNodeLogger{},
		classes:   map[uint64]classEntry{},
		storage:   MemoryStorage,
		scopeID:   1,
		hasher:    NewDefaultHasher(),
		scratch:   codecScratch{bytes: nil, ids: nil},
	}

	_, err := json.Marshal(instance)
	require.Error(t, err)
	require.ErrorContains(t, err, "unsupported predictor implementation")
}

func TestPredictor_UnmarshalJSON_errorCases(t *testing.T) {
	t.Parallel()

	var instance Predictor

	err := json.Unmarshal([]byte(`{"storage":"x"}`), &instance)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to unmarshal predictor JSON")

	err = json.Unmarshal([]byte(`{"schemaVersion":1,"hasher":"xxhash3","storage":0,"scopeId":1}`), &instance)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to initialize predictor from JSON")
}

func TestPredictor_UnmarshalJSON_withoutNodeLog(t *testing.T) {
	t.Parallel()

	var instance Predictor

	raw := []byte(`{"schemaVersion":1,"hasher":"xxhash3","storage":1,"scopeId":9,"classes":null}`)
	err := json.Unmarshal(raw, &instance)
	require.NoError(t, err)
	require.NotNil(t, instance.predictor)
	require.NotNil(t, instance.classes)
}
