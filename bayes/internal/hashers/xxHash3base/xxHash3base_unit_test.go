package xxhash3base

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeebo/xxh3"
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

			require.Equal(t, xxh3.Hash(input), New().Hash(input))
		})
	}
}
