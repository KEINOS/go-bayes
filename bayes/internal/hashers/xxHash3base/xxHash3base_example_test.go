package xxhash3base_test

import (
	"fmt"

	"github.com/KEINOS/go-bayes/bayes/internal/hashers/xxHash3base"
)

func ExampleHasher() {
	hasher := xxhash3base.New()

	fmt.Println(hasher.Name())
	fmt.Println(hasher.Hash([]byte("context")))
	// Output:
	// xxhash3
	// 15840237918302701434
}
