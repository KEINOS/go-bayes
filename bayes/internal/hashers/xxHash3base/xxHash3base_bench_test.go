package xxhash3base

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zeebo/xxh3"
)

func BenchmarkHasher_Hash(b *testing.B) {
	for _, size := range []int{8, 48, 256, 2048} {
		input := make([]byte, size)

		b.Run(benchmarkName(size), func(b *testing.B) {
			hasher := New()
			for b.Loop() {
				hasher.Hash(input)
			}
		})
	}
}

func FuzzHasher_Hash(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0})
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
	f.Add([]byte{0xff, 0xff, 0xff, 0xff})

	f.Fuzz(func(t *testing.T, input []byte) {
		actual := New().Hash(input)
		require.Equal(t, xxh3.Hash(input), actual)

		if len(input) >= 8 {
			copyInput := append([]byte(nil), input...)
			binary.BigEndian.PutUint64(copyInput, binary.BigEndian.Uint64(copyInput)+1)
			require.NotEqual(t, actual, New().Hash(copyInput))
		}
	})
}

func benchmarkName(size int) string {
	switch size {
	case 8:
		return "bytes_8"
	case 48:
		return "bytes_48"
	case 256:
		return "bytes_256"
	case 2048:
		return "bytes_2048"
	default:
		panic("unsupported benchmark size")
	}
}
