// Package hasher defines interfaces for transition hash algorithms.
package hasher

// TransitionHasher hashes transition IDs into a single flow ID.
type TransitionHasher interface {
	HashTrans(transitions ...uint64) (uint64, error)
}
