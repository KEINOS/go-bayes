package bayes

import (
	"encoding/json"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

var errHashFailed = errors.New("hash failed")

// ----------------------------------------------------------------------------
//  Helper types and functions for testing
// ----------------------------------------------------------------------------

type stubHasher struct {
	got []uint64
	out uint64
	err error
}

func (s *stubHasher) HashTrans(transitions ...uint64) (uint64, error) {
	s.got = append([]uint64{}, transitions...)

	if s.err != nil {
		return 0, s.err
	}

	return s.out, nil
}

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

// ----------------------------------------------------------------------------
//  Predictor
// ----------------------------------------------------------------------------

func TestPredictor_SetHasher_nil(t *testing.T) {
	t.Parallel()

	instance, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  nil,
	})
	require.NoError(t, err)

	err = instance.SetHasher(nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "hasher must not be nil")
}

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
	require.ErrorContains(t, err, "failed to convert transitions to uint64")

	stub := &stubHasher{got: nil, out: 0, err: errHashFailed}
	require.NoError(t, instance.SetHasher(stub))

	_, err = instance.HashTrans(uint64(1), uint64(2))
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to hash transitions")
}

func TestPredictor_Predict_notInitialized(t *testing.T) {
	t.Parallel()

	instance := &Predictor{
		predictor: nil,
		classes:   nil,
		storage:   UnknownStorage,
		scopeID:   0,
		hasher:    nil,
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
	}

	instance.SetStorage(MemoryStorage)
	require.Equal(t, MemoryStorage, instance.storage)
}

func TestPredictor_HashTrans_usesInjectedHasher(t *testing.T) {
	t.Parallel()

	instance, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  nil,
	})
	require.NoError(t, err)

	stub := &stubHasher{got: nil, out: 0x99, err: nil}
	require.NoError(t, instance.SetHasher(stub))

	actual, err := instance.HashTrans("a", "b")
	require.NoError(t, err)
	require.Equal(t, uint64(0x99), actual)
	require.Len(t, stub.got, 2)
}

func TestPredictor_NewPredictor_usesConfigHasher(t *testing.T) {
	t.Parallel()

	stub := &stubHasher{got: nil, out: 0x55, err: nil}

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

	original, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 123,
		Hasher:  nil,
	})
	require.NoError(t, err)
	require.NoError(t, original.Train([]any{"a", "b", "a", "b"}))

	raw, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Predictor
	require.NoError(t, json.Unmarshal(raw, &restored))

	actual, err := restored.Predict([]any{"a"})
	require.NoError(t, err)
	require.Equal(t, "b", restored.GetClass(actual))
}

func TestPredictor_MarshalJSON_withoutPredictor(t *testing.T) {
	t.Parallel()

	instance := &Predictor{
		predictor: nil,
		classes:   map[uint64]classEntry{},
		storage:   MemoryStorage,
		scopeID:   1,
		hasher:    NewDefaultHasher(),
	}

	_, err := json.Marshal(instance)
	require.NoError(t, err)
}

func TestPredictor_MarshalJSON_errorCases(t *testing.T) {
	t.Parallel()

	instance := &Predictor{
		predictor: nil,
		classes: map[uint64]classEntry{
			1: {ID: 1, Raw: func() {}},
		},
		storage: MemoryStorage,
		scopeID: 1,
		hasher:  NewDefaultHasher(),
	}

	_, err := json.Marshal(instance)
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

	err = json.Unmarshal([]byte(`{"storage":0,"scopeId":1}`), &instance)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to initialize predictor from JSON")
}

func TestPredictor_UnmarshalJSON_withoutNodeLog(t *testing.T) {
	t.Parallel()

	var instance Predictor

	err := json.Unmarshal([]byte(`{"storage":1,"scopeId":9,"classes":null}`), &instance)
	require.NoError(t, err)
	require.NotNil(t, instance.predictor)
	require.NotNil(t, instance.classes)
}
