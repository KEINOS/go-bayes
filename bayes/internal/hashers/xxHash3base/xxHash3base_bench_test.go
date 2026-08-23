package xxhash3base

import (
	"encoding/binary"
	"testing"

	"github.com/KEINOS/go-bayes/bayes/internal/hashers/blake3base"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/xxh3"
)

func BenchmarkHasher_HashTrans(b *testing.B) {
	for _, size := range []int{1, 6, 32, 256} {
		input := make([]uint64, size)
		for index := range input {
			input[index] = uint64(index) // #nosec G115 -- benchmark input is non-negative.
		}

		b.Run(benchmarkName(size), func(b *testing.B) {
			b.Run("original", func(b *testing.B) {
				hasher := blake3base.New()

				for b.Loop() {
					_, err := hasher.HashTrans(input...)
					if err != nil {
						b.Fatal(err)
					}
				}
			})

			b.Run("enhanced", func(b *testing.B) {
				hasher := New()

				for b.Loop() {
					_, err := hasher.HashTrans(input...)
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}

func FuzzHasher_HashTrans(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0})
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
	f.Add([]byte{0xff, 0xff, 0xff, 0xff})

	f.Fuzz(func(t *testing.T, input []byte) {
		transitions := make([]uint64, len(input))
		encoded := make([]byte, len(input)*8)

		for index, value := range input {
			transitions[index] = uint64(value)
			binary.LittleEndian.PutUint64(encoded[index*8:], uint64(value))
		}

		actual, err := New().HashTrans(transitions...)

		require.NoError(t, err)
		require.Equal(t, xxh3.Hash(encoded), actual)
	})
}

func benchmarkName(size int) string {
	switch size {
	case 1:
		return "transitions_1"
	case 6:
		return "transitions_6"
	case 32:
		return "transitions_32"
	case 256:
		return "transitions_256"
	default:
		panic("unsupported benchmark size")
	}
}
