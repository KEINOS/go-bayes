package hasher_test

import (
	"fmt"

	"github.com/KEINOS/go-bayes/bayes/hasher"
	"github.com/KEINOS/go-bayes/bayes/internal/hashers/blake3base"
)

// ExampleTransitionHasher demonstrates how to use the [TransitionHasher] interface
// with the default blake3base implementation.
func ExampleTransitionHasher() {
	// blake3base.New() returns a *Hasher that satisfies the TransitionHasher interface.
	var h hasher.TransitionHasher = blake3base.New()

	// HashTrans returns a deterministic uint64 flow ID from the given transition IDs.
	flowID, err := h.HashTrans(100, 200, 300)
	if err != nil {
		panic(err)
	}

	fmt.Println(flowID)
	//
	// Output:
	// 14580773609109142544
}
