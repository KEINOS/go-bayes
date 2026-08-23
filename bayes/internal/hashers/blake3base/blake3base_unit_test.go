package blake3base

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeebo/blake3"
)

func TestHasher_Hash(t *testing.T) {
	t.Parallel()

	for name, input := range map[string][]byte{
		"nil":   nil,
		"empty": {},
		"bytes": {0x00, 0x01, 0xff},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			digest := blake3.Sum256(input)
			require.Equal(t, binary.BigEndian.Uint64(digest[:]), New().Hash(input))
		})
	}
}

func TestHasher_Name(t *testing.T) {
	t.Parallel()

	require.Equal(t, "blake3", New().Name())
}
