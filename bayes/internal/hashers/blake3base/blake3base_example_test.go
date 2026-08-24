package blake3base_test

import (
	"fmt"

	"github.com/KEINOS/go-bayes/bayes/internal/hashers/blake3base"
)

func ExampleHasher() {
	hasher := blake3base.New()

	fmt.Println(hasher.Name())
	fmt.Println(hasher.Hash([]byte("context")))
	// Output:
	// blake3
	// 14685906091650472651
}
