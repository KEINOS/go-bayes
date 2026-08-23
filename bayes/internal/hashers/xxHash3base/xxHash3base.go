// Package xxhash3base provides the xxHash3 hasher implementation.
package xxhash3base

import "github.com/zeebo/xxh3"

// Hasher implements byte hashing with xxHash3.
type Hasher struct{}

// New returns a new xxHash3 hasher.
func New() *Hasher {
	return &Hasher{}
}

// Hash returns the xxHash3 digest of data.
func (h *Hasher) Hash(data []byte) uint64 { return xxh3.Hash(data) }

// Name returns the stable persistence name of this implementation.
func (h *Hasher) Name() string { return "xxhash3" }
