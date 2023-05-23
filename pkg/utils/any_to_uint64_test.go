package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------------
//  AnyToUint64()
// ----------------------------------------------------------------------------

func TestAnyToUint64_nil_value(t *testing.T) {
	b, err := AnyToUint64(nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to convert the value to uint64")
	require.Contains(t, err.Error(), "failed to convert the value to bytes")
	require.Contains(t, err.Error(), "failed to gob encode the value")
	require.Zero(t, b, "it should be zero on error")
}

func TestAnyToUint64_unsupported_value(t *testing.T) {
	b, err := AnyToUint64(new(struct{}))

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid or unsupported value")
	require.Contains(t, err.Error(), "Input type: *struct {}")
	require.Zero(t, b, "it should be zero on error")
}

// ----------------------------------------------------------------------------
//  bytesToHash()
// ----------------------------------------------------------------------------

func Test_bytesToHash_nil_value(t *testing.T) {
	b, err := bytesToHash(nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to write the value to the hasher")
	require.Contains(t, err.Error(), "nil value given")
	require.Zero(t, b, "it should be zero on error")
}

// ----------------------------------------------------------------------------
//  bytesToUint64()
// ----------------------------------------------------------------------------

func Test_bytesToUint64(t *testing.T) {
	for _, tt := range []struct {
		value    []byte
		expected uint64
	}{
		{[]byte{}, 0x0},
		{[]byte{0}, 0x0},
		{[]byte{1}, 0x100000000000001},
		{[]byte{1, 0}, 0x100000000000002},
		{[]byte{0, 1}, 0x1000000000001},
		{[]byte{0, 1, 1}, 0x1010000000003},
		{[]byte{1, 0, 1}, 0x100010000000005},
		{[]byte{1, 1, 0}, 0x101000000000006},
		{[]byte{1, 1, 1}, 0x101010000000007},
		{[]byte{102, 111, 111}, 0x666f6f0000000028}, // == "foo" == []byte("foo")
	} {
		expect := tt.expected
		actual := bytesToUint64(tt.value)

		require.Equal(t, expect, actual,
			"value: %v",
			tt.value,
		)
	}
}
