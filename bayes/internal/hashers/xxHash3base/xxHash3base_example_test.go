package xxhash3base_test

import (
	"fmt"

	"github.com/KEINOS/go-bayes/bayes/internal/hashers/xxHash3base"
)

func ExampleHasher_Hash() {
	hasher := xxhash3base.New()

	fmt.Println(hasher.Hash([]byte("context")))
	// Output: 15840237918302701434
}
