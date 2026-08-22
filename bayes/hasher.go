package bayes

import (
	"github.com/KEINOS/go-bayes/bayes/hasher"
	"github.com/KEINOS/go-bayes/bayes/internal/hashers/blake3base"
)

// Hasher hashes transition IDs into a flow ID.
type Hasher = hasher.TransitionHasher

// NewDefaultHasher returns the current default hasher implementation.
//
//nolint:ireturn // returning interface is intentional for algorithm swapping.
func NewDefaultHasher() Hasher {
	return blake3base.New()
}
