package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
//  AnyToBytes()
// -----------------------------------------------------------------------------

func TestAnyToBytes_golen(t *testing.T) {
	for _, tt := range []struct {
		value  any
		expect []byte
	}{
		{int(1), []byte{0, 0, 0, 0, 0, 0, 0, 1}},
		{int8(1), []byte{0, 0, 0, 0, 0, 0, 0, 1}},
		{int16(1), []byte{0, 0, 0, 0, 0, 0, 0, 1}},
		{int32(1), []byte{0, 0, 0, 0, 0, 0, 0, 1}},
		{int64(1), []byte{0, 0, 0, 0, 0, 0, 0, 1}},
		{uint(1), []byte{0, 0, 0, 0, 0, 0, 0, 1}},
		{uint8(1), []byte{0, 0, 0, 0, 0, 0, 0, 1}},
		{uint16(1), []byte{0, 0, 0, 0, 0, 0, 0, 1}},
		{uint32(1), []byte{0, 0, 0, 0, 0, 0, 0, 1}},
		{uint64(1), []byte{0, 0, 0, 0, 0, 0, 0, 1}},
		{float32(1), []byte{0x3f, 0xf0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
		{float64(1), []byte{0x3f, 0xf0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
		{string("abcde"), []byte{0x61, 0x62, 0x63, 0x64, 0x65}},
		{[]byte{1}, []byte{1}},
		{true, []byte{255, 255, 255, 255, 255, 255, 255, 255}},
		{false, []byte{0, 0, 0, 0, 0, 0, 0, 0}},
	} {
		expect := tt.expect
		actual, err := AnyToBytes(tt.value)

		require.NoError(t, err)
		require.Equal(t, expect, actual,
			"Input type %T, got %v, want %v", tt.value, actual, expect)
	}
}

func TestAnyToBytes_unsupported_type(t *testing.T) {
	b, err := AnyToBytes(make(chan int))

	require.Error(t, err)
	require.Empty(t, b, "it should be empty on error")
	require.Contains(t, err.Error(), "failed to convert to byte slice")
	require.Contains(t, err.Error(), "Unsupported type:")
}

// -----------------------------------------------------------------------------
//  AnyToHash()
// -----------------------------------------------------------------------------

func TestAnyToHash_unsupported_type(t *testing.T) {
	{
		b, err := AnyToHash(make(chan int))

		require.Error(t, err)
		require.Empty(t, b, "it should be empty on error")
		require.Contains(t, err.Error(), "failed to hash any type value")
		require.Contains(t, err.Error(), "Unsupported type:")
	}
	{
		b, err := AnyToHash("foo", make(chan int))

		require.Error(t, err)
		require.Empty(t, b, "it should be empty on error")
		require.Contains(t, err.Error(), "failed to hash any type value")
		require.Contains(t, err.Error(), "Unsupported type:")
	}
}

func TestAnyToHash_arg_is_empty(t *testing.T) {
	{
		// Arg is empty.
		b, err := AnyToHash()

		require.Error(t, err)
		require.Empty(t, b, "it should be empty on error")
		require.Contains(t, err.Error(), "no items to hash. arg is empty")
	}
	{
		// Arg is nil.
		b, err := AnyToHash(nil)

		require.Error(t, err)
		require.Empty(t, b, "it should be empty on error")
		require.Contains(t, err.Error(), "no items to hash. arg is nil")
	}
	{
		// Arg contains nil.
		b, err := AnyToHash("foo", nil)

		require.Error(t, err)
		require.Empty(t, b, "it should be empty on error")
		require.Contains(t, err.Error(), "arg contains nil value")
	}
}

// -----------------------------------------------------------------------------
//  AnyToStateID()
// -----------------------------------------------------------------------------

func TestAnyToStateID_unsupported_type(t *testing.T) {
	b, err := AnyToStateID(make(chan int))

	require.Error(t, err)
	require.Empty(t, b, "it should be empty on error")
	require.Contains(t, err.Error(), "failed to hash any type value")
	require.Contains(t, err.Error(), "Unsupported type:")
}
