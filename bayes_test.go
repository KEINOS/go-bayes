package bayes

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------------
//  addClass
// ----------------------------------------------------------------------------

func Test_addClass(t *testing.T) {
	oldClasses := _classes
	defer func() { _classes = oldClasses }()

	for i, tt := range []struct {
		input interface{}
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
			addClass(uint64(i), tt.input)
		})

		c := GetClass(uint64(i))

		require.Equal(t, tt.input, c)
	}
}

// ----------------------------------------------------------------------------
//  convAnyToUint64
// ----------------------------------------------------------------------------

func Test_convAnyToUint64_error_cases(t *testing.T) {
	for _, tt := range []interface{}{
		nil,
		big.NewInt(9223372036854775807),
		*big.NewInt(9223372036854775807),
	} {
		v, err := convAnyToUint64(tt)

		require.Error(t, err, "it should be an error if the input is nil")
		require.Zero(t, v, "it should be zero on error")

		assert.Contains(t, err.Error(), "failed to convert to uint64")
		assert.Contains(t, err.Error(), "Unsupported type:")
	}
}

func Test_convAnyToUint64_golden(t *testing.T) {
	for _, tt := range []struct {
		input  interface{}
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

// ----------------------------------------------------------------------------
//  chopAndMergeBytes
// ----------------------------------------------------------------------------

func Test_chopAndMergeBytes_golden(t *testing.T) {
	a := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
	}
	b := []byte{
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
		0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20,
	}

	hashed, err := chopAndMergeBytes(a, b)
	require.NoError(t, err)

	expect := fmt.Sprintf("%016x", []byte{
		0x01, 0x02, 0x03, 0x04,
		0x11, 0x12, 0x13, 0x14,
	})
	actual := fmt.Sprintf("%016x", hashed)

	require.Equal(t, expect, actual)
}

func Test_chopAndMergeBytes_missing(t *testing.T) {
	a := []byte{
		0x01, 0x02, 0x03,
	}
	b := []byte{
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
		0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20,
	}

	hashed, err := chopAndMergeBytes(a, b)

	require.Error(t, err, "it should be an error if the input length is insufficient")
	assert.Contains(t, err.Error(), "Both of the input must be 4byte or more")
	require.Zero(t, hashed, "it should be zero on error")
}

// ----------------------------------------------------------------------------
//  getBlake3
// ----------------------------------------------------------------------------

func Test_getBlake3(t *testing.T) {
	for _, tt := range []struct {
		expect string
		input  []uint64
	}{
		{
			"71e0a99173564931c0b8acc52d2685a8e39c64dc52e3d02390fdac2a12b155cb",
			[]uint64{0x0},
		},
		{
			"73919af90e1fee9f2c6585e4534a6fa9e04931c0090b9c7ab9e631b16d8c8da0",
			[]uint64{0xffffffffffffffff},
		},
	} {
		hashed, err := getBlake3(tt.input...)

		require.NoError(t, err)

		expect := tt.expect
		actual := fmt.Sprintf("%032x", hashed)

		require.Equal(t, expect, actual, "input: %v", tt.input)
	}
}

// ----------------------------------------------------------------------------
//  getCRC32C
// ----------------------------------------------------------------------------

func Test_getCRC32C(t *testing.T) {
	for _, tt := range []struct {
		input  []byte
		expect []byte
	}{
		{
			[]byte{0x10},
			[]byte{0x42, 0x23, 0x94, 0x3E},
		},
		{
			[]byte{0x10, 0x20, 0x30, 0x40},
			[]byte{0x41, 0x72, 0x0E, 0xD1},
		},
		{
			[]byte{0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80},
			[]byte{0x1E, 0xE0, 0x0E, 0x76},
		},
	} {
		actual := getCRC32C(tt.input)
		expect := tt.expect

		require.Equal(t, expect, actual, "input: %v", tt.input)
	}
}

// ----------------------------------------------------------------------------
//  New
// ----------------------------------------------------------------------------

func TestNew_unknown_storage_type(t *testing.T) {
	trainer, err := New(UnknwonStorage, 0)

	require.Error(t, err, "unknown storage type should be an error")
	require.Nil(t, trainer, "it should be nil on error")

	assert.Contains(t, err.Error(), "unknown storage engine type")
}

// ----------------------------------------------------------------------------
//  Predict
// ----------------------------------------------------------------------------

func TestPredict_slice_of_unsupported_type(t *testing.T) {
	oldPredictor := _predictor
	defer func() {
		_predictor = oldPredictor // Recover object

		Reset() // Reset the predictor
	}()

	v, err := Predict([]big.Int{
		*big.NewInt(9223372036854775807),
		*big.NewInt(9223372036854775807),
		*big.NewInt(9223372036854775807),
	})

	require.Error(t, err, "it should be an error if the input is a slice of unsupported type")
	require.Zero(t, v, "it should be zero on error")

	assert.Contains(t, err.Error(), "failed to hash the flow")
}

func TestPredict_not_initialized(t *testing.T) {
	oldPredictor := _predictor
	defer func() {
		_predictor = oldPredictor
	}()

	// Mock the singleton predictor
	_predictor = nil

	v, err := Predict([]byte{0x10})

	require.Error(t, err, "it should be an error if the predictor is not initialized")
	require.Zero(t, v, "it should be zero on error")

	assert.Contains(t, err.Error(), "predictor is not initialized")
}

// ----------------------------------------------------------------------------
//  Reset
// ----------------------------------------------------------------------------

func TestReset_panic(t *testing.T) {
	oldStorage := _storage
	defer func() {
		_storage = oldStorage // Recover object
	}()

	// Mock the storage to unknown type
	SetStorage(UnknwonStorage)

	assert.Panics(t, func() {
		Reset()
	}, "it should panic if the storage is unknown")
}

// ----------------------------------------------------------------------------
//  Train
// ----------------------------------------------------------------------------

func TestTrain_not_initialized(t *testing.T) {
	oldPredictor := _predictor
	defer func() {
		_predictor = oldPredictor // Recover object

		Reset() // Reset the predictor
	}()

	// Mock the singleton predictor
	_predictor = nil

	err := Train([]string{"foo", "bar"})

	require.NoError(t, err)
}

func TestTrain_slice_of_unsupported_type(t *testing.T) {
	oldPredictor := _predictor
	defer func() {
		_predictor = oldPredictor // Recover object

		Reset() // Reset the predictor
	}()

	err := Train([]big.Int{
		*big.NewInt(9223372036854775807),
		*big.NewInt(9223372036854775807),
		*big.NewInt(9223372036854775807),
	})

	require.Error(t, err, "it should be an error if the input is a slice of unsupported type")
	assert.Contains(t, err.Error(), "failed during training iteration")
	assert.Contains(t, err.Error(), "Unsupported type")
}

// ----------------------------------------------------------------------------
//  uint64ToByteArray
// ----------------------------------------------------------------------------

func Test_uint64ToByteArray(t *testing.T) {
	for _, tt := range []struct {
		expect []byte
		input  uint64
	}{
		{
			[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			0x0,
		},
		{
			[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			0xffffffffffffffff,
		},
		{
			[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0f},
			0xfffffffffffffff,
		},
	} {
		actual := uint64ToByteArray(tt.input)
		expect := tt.expect

		require.Equal(t, expect, actual, "input: %v", tt.input)
	}
}
