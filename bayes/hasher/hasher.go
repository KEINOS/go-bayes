// Package hasher defines the extension interface for context-folding hashes.
package hasher

// TransitionHasher folds ordered item IDs into one deterministic, fixed-width
// context ID. Implementations do not guarantee collision-free IDs.
type TransitionHasher interface {
	HashTrans(transitions ...uint64) (uint64, error)
}
