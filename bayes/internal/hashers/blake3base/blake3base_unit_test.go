package blake3base

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasher_HashTrans_golden(t *testing.T) {
	t.Parallel()

	hasher := New()

	for _, tt := range []struct {
		input  []uint64
		expect uint64
	}{
		{[]uint64{10, 11, 12, 13, 14, 15}, 0x6919623f91c5be2e},
		{[]uint64{10, 11, 12, 13, 15, 14}, 0x6e3603ca6af4590c},
		{[]uint64{1, 11, 12, 13, 14, 15}, 0xe95454663822c749},
		{[]uint64{1}, 0x1a0d1201d898958f},
	} {
		actual, err := hasher.HashTrans(tt.input...)
		require.NoError(t, err)
		require.Equal(t, tt.expect, actual)
	}
}

func TestHasher_HashTrans_edgeCases(t *testing.T) {
	t.Parallel()

	hasher := New()

	emptyFirst, err := hasher.HashTrans()
	require.NoError(t, err)

	emptySecond, err := hasher.HashTrans()
	require.NoError(t, err)
	require.Equal(t, emptyFirst, emptySecond)

	zero, err := hasher.HashTrans(0)
	require.NoError(t, err)

	maxUint64, err := hasher.HashTrans(^uint64(0))
	require.NoError(t, err)

	require.NotEqual(t, emptyFirst, zero)
	require.NotEqual(t, emptyFirst, maxUint64)
	require.NotEqual(t, zero, maxUint64)
}

func TestChopAndMergeBytes_missing(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name   string
		bytesA []byte
		bytesB []byte
	}{
		{"first input too short", []byte{0x01, 0x02, 0x03}, []byte{0x10, 0x20, 0x30, 0x40}},
		{"second input too short", []byte{0x01, 0x02, 0x03, 0x04}, []byte{0x10, 0x20, 0x30}},
		{"both inputs nil", nil, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := chopAndMergeBytes(tt.bytesA, tt.bytesB)

			require.ErrorIs(t, err, errInvalidBytesLength)
			require.Zero(t, actual)
		})
	}
}

func TestUint64ToByteArray(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		input  uint64
		expect []byte
	}{
		{0x0, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{0xffffffffffffffff, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{0xfffffffffffffff, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0f}},
	} {
		actual := uint64ToByteArray(tt.input)
		require.Equal(t, tt.expect, actual, "input: %v", tt.input)
	}
}
