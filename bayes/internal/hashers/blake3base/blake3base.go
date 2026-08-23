// Package blake3base provides the BLAKE3 hasher implementation.
package blake3base

import (
	"encoding/binary"

	"github.com/zeebo/blake3"
)

// Hasher implements byte hashing with BLAKE3.
type Hasher struct{}

// New returns a new BLAKE3 hasher.
func New() *Hasher {
	return &Hasher{}
}

// Hash returns the first 64 bits of the BLAKE3 digest in big-endian order.
func (h *Hasher) Hash(data []byte) uint64 {
	digest := blake3.Sum256(data)

	return binary.BigEndian.Uint64(digest[:])
}

// Name returns the stable persistence name of this implementation.
func (h *Hasher) Name() string { return "blake3" }
