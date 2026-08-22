package bayes

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------------
//  addClass
// ----------------------------------------------------------------------------

func newTestPredictor(t *testing.T) *Predictor {
	t.Helper()

	instance, err := NewPredictor(PredictorConfig{
		Storage: MemoryStorage,
		ScopeID: 0,
		Hasher:  nil,
	})
	require.NoError(t, err)

	return instance
}

func Test_addClass(t *testing.T) {
	t.Parallel()

	instance := newTestPredictor(t)

	//nolint:varnamelen // tt is short but descriptive
	for i, tt := range []struct {
		input any
	}{
		{-1234},
		{12345},
		{uint(0xffffffffffffffff)},
		{uint64(1)},
		{uint32(1)},
		{uint16(1)},
		{uint(1)},
		{int64(1)},
		{int32(1)},
		{int16(1)},
		{int(0xff)},
		{float64(1.0)},
		{float32(1.0)},
		{"foobar"},
		{true},
		{false},
		{big.NewInt(9223372036854775807)},
	} {
		require.NotPanics(t, func() {
			instance.addClass(uint64(i), tt.input) // #nosec
		})

		c := instance.GetClass(uint64(i)) // #nosec

		require.Equal(t, tt.input, c)
	}
}

// ----------------------------------------------------------------------------
//  convAnyToUint64
// ----------------------------------------------------------------------------

func Test_convAnyToUint64_error_cases(t *testing.T) {
	t.Parallel()

	for _, tt := range []any{
		nil,
		big.NewInt(9223372036854775807),
		*big.NewInt(9223372036854775807),
	} {
		v, err := convAnyToUint64(tt)

		require.Error(t, err, "it should be an error if the input is nil")
		require.Zero(t, v, "it should be zero on error")

		assert.Contains(t, err.Error(), "unsupported type for conversion")
	}
}

func Test_convAnyToUint64_golden(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		input  any
		expect uint64
	}{
		{-1234, uint64(0xfffffffffffffb2e)},
		{12345, uint64(0x3039)},
		{uint(0xffffffffffffffff), uint64(0xffffffffffffffff)},
		{uint64(1), uint64(1)},
		{uint32(1), uint64(1)},
		{uint16(1), uint64(1)},
		{uint(1), uint64(1)},
		{int64(1), uint64(1)},
		{int32(1), uint64(1)},
		{int16(1), uint64(1)},
		{int(2147483647), uint64(0x7fffffff)},
		{int(-2147483648), uint64(0xffffffff80000000)},
		{int(0xff), uint64(0xff)},
		{float64(1.0), uint64(1)},
		{float32(1.0), uint64(1)},
		{"foobar", uint64(0xaa51dcd43d5c6c52)},
		{true, uint64(1)},
		{false, uint64(0)},
	} {
		v, err := convAnyToUint64(tt.input)

		require.NoError(t, err)
		assert.Equal(t, tt.expect, v, "input: %v", tt.input)
	}
}

func Test_convAnyToUint64_int_large_value(t *testing.T) {
	t.Parallel()

	v, err := convAnyToUint64(int(2147483648)) // maxInt32 + 1

	require.NoError(t, err)
	require.Equal(t, uint64(2147483648), v)
}

// ----------------------------------------------------------------------------
//  New
// ----------------------------------------------------------------------------

func TestNew_unknown_storage_type(t *testing.T) {
	t.Parallel()

	predictor, err := New(UnknownStorage, 0)

	require.Error(t, err, "unknown storage type should be an error")
	require.Nil(t, predictor, "it should be nil on error")

	assert.Contains(t, err.Error(), "unknown storage engine type")
}

func TestNew_returnsPredictorSupportingTypedSlices(t *testing.T) {
	t.Parallel()

	instance, err := New(MemoryStorage, 100)
	require.NoError(t, err)
	require.NotNil(t, instance)
	require.Equal(t, uint64(100), instance.ID())

	err = instance.Train([]string{"So", "So", "La", "So", "Do", "Si"})
	require.NoError(t, err)

	classID, err := instance.Predict([]string{"So", "So", "La", "So", "Do"})
	require.NoError(t, err)
	require.Equal(t, "Si", instance.GetClass(classID))
}

// ----------------------------------------------------------------------------
//  Predict
// ----------------------------------------------------------------------------

func TestPredict_slice_of_unsupported_type(t *testing.T) {
	t.Parallel()

	instance := newTestPredictor(t)

	classPredicted, err := instance.Predict([]any{
		*big.NewInt(9223372036854775807),
		*big.NewInt(9223372036854775807),
		*big.NewInt(9223372036854775807),
	})

	require.Error(t, err, "it should be an error if the input is a slice of unsupported type")
	require.Zero(t, classPredicted, "it should be zero on error")

	assert.Contains(t, err.Error(), "failed to hash the flow")
}

func TestPredict_not_initialized(t *testing.T) {
	t.Parallel()

	instance := &Predictor{
		predictor: nil,
		classes:   nil,
		storage:   UnknownStorage,
		scopeID:   0,
		hasher:    nil,
	}

	classPredicted, err := instance.Predict([]any{byte(0x10)})

	require.Error(t, err, "it should be an error if the predictor is not initialized")
	require.Zero(t, classPredicted, "it should be zero on error")

	assert.Contains(t, err.Error(), "predictor is not initialized")
}

// ----------------------------------------------------------------------------
//  Reset
// ----------------------------------------------------------------------------

func TestReset_panic(t *testing.T) {
	t.Parallel()

	instance := newTestPredictor(t)
	instance.SetStorage(UnknownStorage)

	assert.Panics(t, func() {
		err := instance.Reset()
		if err != nil {
			panic(err)
		}
	}, "it should panic if the storage is unknown")
}

// ----------------------------------------------------------------------------
//  Train
// ----------------------------------------------------------------------------

func TestTrain_not_initialized(t *testing.T) {
	t.Parallel()

	instance := newTestPredictor(t)
	instance.predictor = nil

	err := instance.Train([]any{"foo", "bar"})

	require.NoError(t, err)
}

func TestTrain_slice_of_unsupported_type(t *testing.T) {
	t.Parallel()

	instance := newTestPredictor(t)

	err := instance.Train([]any{
		*big.NewInt(9223372036854775807),
		*big.NewInt(9223372036854775807),
		*big.NewInt(9223372036854775807),
	})

	require.Error(t, err, "it should be an error if the input is a slice of unsupported type")
	assert.Contains(t, err.Error(), "failed during training iteration")
	assert.Contains(t, err.Error(), "unsupported type for conversion")
}
