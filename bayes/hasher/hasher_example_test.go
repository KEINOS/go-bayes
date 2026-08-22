package hasher_test

import (
	"fmt"

	"github.com/KEINOS/go-bayes/bayes"
	"github.com/KEINOS/go-bayes/bayes/hasher"
)

// ExampleTransitionHasher demonstrates how to use the [TransitionHasher] interface
// with the default xxHash3 implementation.
func ExampleTransitionHasher() {
	var h hasher.TransitionHasher = bayes.NewDefaultHasher()

	// HashTrans returns a deterministic uint64 flow ID from the given transition IDs.
	flowID, err := h.HashTrans(100, 200, 300)
	if err != nil {
		panic(err)
	}

	fmt.Println(flowID)
	//
	// Output:
	// 8074401178316706309
}
