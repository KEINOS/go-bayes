package blake3base

import "testing"

func BenchmarkHasher_Hash(b *testing.B) {
	hasher := New()
	input := []byte("canonical context bytes")

	for b.Loop() {
		hasher.Hash(input)
	}
}
