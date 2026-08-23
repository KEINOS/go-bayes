// Package xxhash3base provides the xxHash3 transition hasher implementation.
package xxhash3base

import (
	"encoding/binary"

	"github.com/KEINOS/go-bayes/bayes/hasher"
	"github.com/zeebo/xxh3"
)

const (
	bytesPerTransition       = 8
	inlineTransitionCapacity = 8
)

// Compile-time check that *Hasher satisfies hasher.TransitionHasher.
var _ hasher.TransitionHasher = (*Hasher)(nil)

// Hasher implements transition hashing with xxHash3.
type Hasher struct{}

// New returns a new xxHash3 transition hasher.
func New() *Hasher {
	return &Hasher{}
}

// HashTrans folds the ordered transition IDs into an xxHash3 context ID.
func (h *Hasher) HashTrans(transitions ...uint64) (uint64, error) {
	var inline [inlineTransitionCapacity * bytesPerTransition]byte

	encodedLength := len(transitions) * bytesPerTransition
	encoded := inline[:]

	if encodedLength > len(inline) {
		encoded = make([]byte, encodedLength)
	} else {
		encoded = encoded[:encodedLength]
	}

	for index, transition := range transitions {
		binary.LittleEndian.PutUint64(encoded[index*bytesPerTransition:], transition)
	}

	return xxh3.Hash(encoded), nil
}
