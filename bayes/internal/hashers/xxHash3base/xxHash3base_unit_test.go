package xxhash3base

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeebo/xxh3"
)

func TestHasher_HashTrans_edgeCases(t *testing.T) {
	t.Parallel()

	tests := map[string][]uint64{
		"empty input":          {},
		"zero":                 {0},
		"maximum uint64":       {math.MaxUint64},
		"inline boundary":      {0, 1, 2, 3, 4, 5, 6, 7},
		"over inline boundary": {0, 1, 2, 3, 4, 5, 6, 7, 8},
	}

	for name, transitions := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			expectedInput := make([]byte, len(transitions)*8)
			for index, transition := range transitions {
				binary.LittleEndian.PutUint64(expectedInput[index*8:], transition)
			}

			actual, err := New().HashTrans(transitions...)

			require.NoError(t, err)
			require.Equal(t, xxh3.Hash(expectedInput), actual)
		})
	}
}

func TestHasher_HashTrans_isDeterministicAndOrderSensitive(t *testing.T) {
	t.Parallel()

	hasher := New()

	forward, err := hasher.HashTrans(1, 2, 3)
	require.NoError(t, err)

	repeated, err := hasher.HashTrans(1, 2, 3)
	require.NoError(t, err)

	reversed, err := hasher.HashTrans(3, 2, 1)
	require.NoError(t, err)

	require.Equal(t, forward, repeated)
	require.NotEqual(t, forward, reversed)
}
