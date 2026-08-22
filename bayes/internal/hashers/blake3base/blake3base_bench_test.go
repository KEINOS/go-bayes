package blake3base

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkHasher_HashTrans(b *testing.B) {
	hasher := New()
	input := []uint64{10, 11, 12, 13, 14, 15}

	for b.Loop() {
		_, err := hasher.HashTrans(input...)
		require.NoError(b, err)
	}
}
