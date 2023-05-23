package utils

import (
	"github.com/zeebo/blake3"
)

// BytesToHash converts the list of byte slices to a hash.
func BytesToHash(items ...[]byte) []byte {
	// Create hasher
	h := blake3.New()

	// Add items to hasher
	for _, v := range items {
		if _, err := h.Write(v); err != nil {
			return nil
		}
	}

	return h.Sum(nil) // Return hash
}
